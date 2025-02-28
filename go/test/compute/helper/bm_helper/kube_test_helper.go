// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bm_helper

import (
	"context"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
	baremetalv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	// pollingInterval            = 1 * time.Second
	metal3NamespaceSelectorKey = "cloud.intel.com/bmaas-metal3-namespace"
)

func GetNamespace(namespace string) (*v1.Namespace, error) {
	ctx := context.Background()
	config, err := util.GetRESTConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	// _, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}

	return clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
}

func GetNamespaceListWithLabel(selector string) (*v1.NamespaceList, error) {
	ctx := context.Background()
	config, err := util.GetRESTConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	// _, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}
	namespaces, err := clientset.CoreV1().
		Namespaces().List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return &v1.NamespaceList{}, fmt.Errorf("unable to list metal3 namespaces: %v", err)
	}
	if len(namespaces.Items) == 0 {
		return &v1.NamespaceList{}, fmt.Errorf("no metal3 namespace found")
	}

	return namespaces, nil
}

func GetBmhByConsumer(consumerName string) (*baremetalv1alpha1.BareMetalHost, error) {
	config, err := util.GetRESTConfig()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	resource := baremetalv1alpha1.SchemeBuilder.GroupVersion.WithResource("baremetalhosts")
	// struct for bmh
	bmh := &baremetalv1alpha1.BareMetalHost{}

	// use the name to convert the right object
	selector := fmt.Sprintf("%s=true", metal3NamespaceSelectorKey)
	namespaces, err := GetNamespaceListWithLabel(selector)
	if err != nil {
		return nil, err
	}
	for _, ns := range namespaces.Items {
		hosts, err := client.Resource(resource).Namespace(ns.Name).List(context.Background(), metav1.ListOptions{})
		if err == nil {
			for _, host := range hosts.Items {
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(host.UnstructuredContent(), bmh); err == nil {
					if bmh.Spec.ConsumerRef != nil {
						if bmh.Spec.ConsumerRef.Name == consumerName {
							return bmh, nil
						}
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("Unable to get BMH resource, error: %v", err)
}

func GetBmhByName(devicename string) (*baremetalv1alpha1.BareMetalHost, error) {
	config, err := util.GetRESTConfig()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	resource := baremetalv1alpha1.SchemeBuilder.GroupVersion.WithResource("baremetalhosts")
	// struct for bmh
	bmh := &baremetalv1alpha1.BareMetalHost{}

	// use the name to convert the right object
	selector := fmt.Sprintf("%s=true", metal3NamespaceSelectorKey)
	namespaces, err := GetNamespaceListWithLabel(selector)
	if err != nil {
		return nil, err
	}
	for _, ns := range namespaces.Items {
		host, err := client.Resource(resource).Namespace(ns.Name).Get(context.Background(), devicename, metav1.GetOptions{})
		if err != nil {
			continue
		} else {
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(host.UnstructuredContent(), bmh); err != nil {
				return nil, fmt.Errorf("Unable to decode BareMetalHost object")
			} else {
				return bmh, nil
			}
		}
	}
	return nil, fmt.Errorf("Unable find BMH with device name %s", devicename)
}

func CheckBMHState(system string, state string, timeout int) (bool, error) {
	stateReached := make(chan bool)

	var to time.Duration = (time.Duration(timeout) * time.Second)
	var reached bool
	var err error

	go func() {
		checkState(stateReached, system, state)
		<-stateReached
	}()
	select {
	case res := <-stateReached:
		if res {
			reached = true
			err = nil
		}
	case <-time.After(to):
		reached = false
		err = fmt.Errorf("timeout waiting for %s to reach state: %s", system, state)
	}

	return reached, err
}

func checkState(done chan bool, system string, state string) {
	for {
		bmh, err := GetBmhByName(system)
		if err != nil {
			fmt.Printf("Unable to get BMH: %s, waiting... \n", system)
			time.Sleep(time.Duration(pollingInterval))
			continue
		}
		currState := fmt.Sprint(bmh.Status.Provisioning.State)
		fmt.Printf("->-> System: %s is %s, waiting to reach %s state\n", system, currState, state)
		if fmt.Sprint(bmh.Status.Provisioning.State) == state {
			fmt.Printf("***System: %s is in state %s\n", system, state)
			break
		}
		time.Sleep(time.Duration(pollingInterval))
	}
	done <- true
}

func GetPodList(namespace string) (*v1.PodList, error) {
	return listPods(namespace, metav1.ListOptions{})
}

func GetPodListByFieldSelector(namespace string, fieldselector string) (*v1.PodList, error) {
	return listPods(namespace, metav1.ListOptions{FieldSelector: fieldselector})
}

func listPods(namespace string, opts metav1.ListOptions) (*v1.PodList, error) {
	ctx := context.Background()
	config, err := util.GetRESTConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, opts)

	if err != nil {
		return nil, err
	}

	return pods, nil
}

func DeletePod(pod string, namespace string) error {
	ctx := context.Background()
	config, err := util.GetRESTConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return clientset.CoreV1().Pods(namespace).Delete(ctx, pod, metav1.DeleteOptions{})
}
