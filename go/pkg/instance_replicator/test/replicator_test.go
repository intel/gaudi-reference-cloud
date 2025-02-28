// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/google/uuid"
	instanceoptest "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/test"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_replicator/replicator"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// A test environment that includes a replicator.
type ReplicatorTestEnv struct {
	BaseTestEnv
	ManagerStoppable *stoppable.Stoppable
}

func NewReplicatorTestEnv() *ReplicatorTestEnv {
	ctx := context.Background()

	listerWatcherTestEnv := NewBaseTestEnv()

	By("Creating manager")
	k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
		Scheme:  scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	By("Creating replicator")
	_, err = replicator.NewReplicator(ctx, k8sManager, listerWatcherTestEnv.InstancePrivateServiceClient)
	Expect(err).Should(Succeed())

	By("Starting manager")
	managerStoppable := stoppable.New(k8sManager.Start)
	managerStoppable.Start(ctx)

	return &ReplicatorTestEnv{
		BaseTestEnv:      *listerWatcherTestEnv,
		ManagerStoppable: managerStoppable,
	}
}

func (e *ReplicatorTestEnv) Stop() {
	ctx := context.Background()
	By("Stopping manager")
	Expect(e.ManagerStoppable.Stop(ctx)).Should(Succeed())
	By("Manager stopped")
}

func NewInstance(namespace string, instanceName string, resourceVersion string) *privatecloudv1alpha1.Instance {
	instance := instanceoptest.NewInstance(
		namespace,
		instanceName,
		"az1",
		"region1",
		instanceoptest.NewInstanceTypeSpec("tiny"),
		instanceoptest.NewSshPublicKeySpecs("ssh-rsa xxx"),
		instanceoptest.NewInterfaceSpecs(),
	)
	instance.ObjectMeta.ResourceVersion = resourceVersion
	return instance
}

var _ = Describe("Replicator", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	It("When replicator receives a new instance from GRPC Watch, it should create an instance in K8s", func() {
		testEnv := NewReplicatorTestEnv()
		defer testEnv.Stop()

		By("Get list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark("10")
		close(listResponseChan)

		By("Get watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending instance to watch response channel")
		namespace := uuid.NewString()
		instanceName := uuid.NewString()
		instance := NewInstance(namespace, instanceName, "11")
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, instance)

		By("Waiting for instance to be created in K8s")
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		instanceRef := &privatecloudv1alpha1.Instance{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, "10s").Should(Succeed())
	})

	It("When instance status is updated in K8s, replicator should call GRPC UpdateStatus", func() {
		testEnv := NewReplicatorTestEnv()
		defer testEnv.Stop()

		By("Get list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark("10")
		close(listResponseChan)

		By("Get watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending instance to watch response channel")
		namespace := uuid.NewString()
		instanceName := uuid.NewString()
		instance := NewInstance(namespace, instanceName, "11")
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, instance)

		By("Waiting for instance to be created in K8s")
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		instanceRef := &privatecloudv1alpha1.Instance{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Update instance status in K8s (simulate Instance Operator)")
		instanceRef.Status.Phase = privatecloudv1alpha1.PhaseReady
		instanceRef.Status.Message = privatecloudv1alpha1.InstanceMessageRunning
		instanceRef.Status.Interfaces = []privatecloudv1alpha1.InstanceInterfaceStatus{{
			Name:      "default",
			Addresses: []string{"1.2.3.4"},
		}}
		instanceRef.Status.Conditions = append(instanceRef.Status.Conditions, privatecloudv1alpha1.InstanceCondition{
			Type:               privatecloudv1alpha1.InstanceConditionRunning,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "reason",
			Message:            "message",
		})
		instanceRef.Status.SshProxy.ProxyUser = "guest"
		instanceRef.Status.SshProxy.ProxyAddress = "5.6.7.8"
		instanceRef.Status.SshProxy.ProxyPort = 22
		Expect(k8sClient.Status().Update(ctx, instanceRef)).Should(Succeed())

		By("Waiting for GRPC UpdateStatus to be called with updated status")
		Eventually(func(g Gomega) {
			value, ok := testEnv.UpdateStatusRecorder.Load(instanceName)
			g.Expect(ok).Should(BeTrue())
			req, ok := value.(*pb.InstanceUpdateStatusRequest)
			g.Expect(ok).Should(BeTrue())
			g.Expect(req.Status.Phase.String()).Should(Equal(string(instanceRef.Status.Phase)))
			g.Expect(req.Status.Message).Should(Equal(instanceRef.Status.Message))
			g.Expect(req.Status.Interfaces[0].Addresses[0]).Should(Equal(string(instanceRef.Status.Interfaces[0].Addresses[0])))
			g.Expect(req.Status.SshProxy.ProxyUser).Should(Equal(instanceRef.Status.SshProxy.ProxyUser))
			g.Expect(req.Status.SshProxy.ProxyAddress).Should(Equal(instanceRef.Status.SshProxy.ProxyAddress))
			g.Expect(int(req.Status.SshProxy.ProxyPort)).Should(Equal(instanceRef.Status.SshProxy.ProxyPort))
		}, "10s").Should(Succeed())

		close(watchResponseChan)
	})

	It("When replicator receives an instance with a deletionTimestamp, it should delete the instance from K8s and call RemoveFinalizer", func() {
		testEnv := NewReplicatorTestEnv()
		defer testEnv.Stop()

		By("Get list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark("10")
		close(listResponseChan)

		By("Get watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending instance to watch response channel")
		namespace := uuid.NewString()
		instanceName := uuid.NewString()
		instance := NewInstance(namespace, instanceName, "11")
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, instance)

		By("Waiting for instance to be created in K8s")
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		instanceRef := &privatecloudv1alpha1.Instance{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Sending instance with deletionTimestamp to watch response channel")
		now := metav1.Now()
		instance.ObjectMeta.DeletionTimestamp = &now
		instance.ObjectMeta.ResourceVersion = "12"
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, instance)

		By("Waiting for instance to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, instanceLookupKey, instanceRef))).Should(BeTrue())
		}, "10s").Should(Succeed())

		By("Waiting for GRPC RemoveFinalizer to be called")
		Eventually(func(g Gomega) {
			_, ok := testEnv.RemoveFinalizerRecorder.Load(instanceName)
			g.Expect(ok).Should(BeTrue())
		}, "10s").Should(Succeed())

		close(watchResponseChan)
	})
})
