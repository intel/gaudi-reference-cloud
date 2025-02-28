// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

var (
	errHostAlreadyDeleting = errors.New("host is already being deleted")
	errHostInUse           = errors.New("host is in use")
	errHostNotFound        = errors.New("host is not found or already deleted")
)

type DisenrollmentTask struct {
	deviceData    *DeviceData
	netBox        dcim.DCIM
	vault         secrets.SecretManager
	clientSet     kubernetes.Interface
	dynamicClient dynamic.Interface
}

func NewDisenrollmentTask(ctx context.Context) (*DisenrollmentTask, error) {
	log := log.FromContext(ctx).WithName("NewDisenrollmentTask")
	log.Info("Initializing new disenrollment task")

	deviceData, err := getDeviceData(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get device information: %v", err)
	}

	vault, err := GetVaultClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	netBox, err := GetNetBoxClient(ctx, vault, deviceData.Region)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize NetBox client: %v", err)
	}

	clientSet, dynamicClient, err := getK8sClients(ctx)
	if err != nil {
		err := fmt.Errorf("unable to initialize K8s Client: %v", err)
		updateDeviceStatus(ctx, netBox, deviceData, dcim.BMDisenrollmentFailed, err.Error())
		return nil, err
	}

	task := &DisenrollmentTask{
		deviceData:    deviceData,
		netBox:        netBox,
		vault:         vault,
		clientSet:     clientSet,
		dynamicClient: dynamicClient,
	}

	return task, nil
}

func (t *DisenrollmentTask) Run(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("DisenrollmentTask.Run")
	log.Info("Disenrollment task started")

	if err := t.disenroll(ctx); err != nil {
		if err := t.updateDeviceStatus(ctx, dcim.BMDisenrollmentFailed, fmt.Sprintf("Disenrollment failed: %v", err)); err != nil {
			return err
		}
		return fmt.Errorf("unable to disenroll device: %v", err)
	}

	if err := t.updateDeviceStatus(ctx, dcim.BMDisenrolled, "Disenrollment is complete"); err != nil {
		return err
	}

	log.Info("Disenrollment task completed!!")

	return nil
}

func (t *DisenrollmentTask) disenroll(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("DisenrollmentTask.disenroll")

	if err := t.updateDeviceStatus(ctx, dcim.BMDisenrolling, "Disenrollment is in progress"); err != nil {
		return err
	}

	bmh, err := t.getHost(ctx)
	if errors.Is(err, errHostNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to get BMHost: %v", err)
	}

	if err := t.checkHost(bmh); err != nil {
		if errors.Is(err, errHostAlreadyDeleting) {
			log.Info("BMHost is already being deleted")
			if err := t.waitForHostToBeDeleted(ctx, bmh.GetName(), bmh.GetNamespace()); err != nil {
				return fmt.Errorf("error waiting for host to be deleted: %v", err)
			}
			log.Info("BMHost is deleted", "name", bmh.GetName(), "namespace", bmh.GetNamespace())
			return nil
		}
		return err
	}

	if err := t.deleteHost(ctx, bmh.GetName(), bmh.GetNamespace()); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("BMHost already deleted", "name", t.deviceData.Name)
			return nil
		}
		return fmt.Errorf("unable to delete BMHost: %v", err)
	}

	if err := t.waitForHostToBeDeleted(ctx, bmh.GetName(), bmh.GetNamespace()); err != nil {
		return fmt.Errorf("error waiting for host to be deleted: %v", err)
	}

	log.Info("BMHost is deleted", "name", bmh.GetName(), "namespace", bmh.GetNamespace())

	return nil
}

func (t *DisenrollmentTask) updateDeviceStatus(ctx context.Context, status dcim.BMEnrollmentStatus, comment string) error {
	if err := t.netBox.UpdateDeviceCustomFields(ctx, t.deviceData.Name, t.deviceData.ID, &dcim.DeviceCustomFields{
		BMEnrollmentStatus:  status,
		BMEnrollmentComment: comment,
	}); err != nil {
		return fmt.Errorf("unable to update NetBox device status: %v", err)
	}
	return nil
}

func (t *DisenrollmentTask) getHost(ctx context.Context) (*baremetalv1alpha1.BareMetalHost, error) {
	log := log.FromContext(ctx).WithName("DisenrollmentTask.getHost")
	log.Info("Retrieving the BareMetalHost from Metal3 namespaces")

	namespaces, err := getMetal3Namespaces(ctx, t.clientSet)
	if err != nil {
		return nil, err
	}

	for _, namespace := range namespaces {
		list, err := t.dynamicClient.Resource(bmHostGVR).Namespace(namespace.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to list BMHosts: %v", err)
		}

		for _, obj := range list.Items {
			if obj.GetName() == t.deviceData.Name {
				log.Info("Found BareMetalHost", "name", obj.GetName(), "namespace", obj.GetNamespace())

				bmh := &baremetalv1alpha1.BareMetalHost{}
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), bmh); err != nil {
					return nil, fmt.Errorf("unable to decode BareMetalHost object")
				}
				return bmh, nil
			}
		}
	}

	return nil, errHostNotFound
}

func (t *DisenrollmentTask) checkHost(bmh *baremetalv1alpha1.BareMetalHost) error {
	if bmh.DeletionTimestamp != nil {
		return errHostAlreadyDeleting
	}
	if bmh.Spec.ConsumerRef != nil {
		return errHostInUse
	}

	return nil
}

func (t *DisenrollmentTask) deleteHost(ctx context.Context, name string, namespace string) error {
	log := log.FromContext(ctx).WithName("DisenrollmentTask.deleteHost")
	log.Info("Deleting BMHost", "name", name, "namespace", namespace)
	//start by skipping clean mode to complete this task faster
	patch := []byte(`{"spec":{"automatedCleaningMode": "disabled"}}`)
	_, err := t.dynamicClient.Resource(bmHostGVR).
		Namespace(namespace).
		Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	//now call delete
	if err := t.dynamicClient.Resource(bmHostGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (t *DisenrollmentTask) waitForHostToBeDeleted(ctx context.Context, name string, namespace string) error {
	log := log.FromContext(ctx).WithName("DisenrollmentTask.waitForHostToBeDeleted")
	log.Info("Waiting for BMHost to be deleted", "name", name, "namespace", namespace)

	for {
		if _, err := t.dynamicClient.Resource(bmHostGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("unable to get host: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
}
