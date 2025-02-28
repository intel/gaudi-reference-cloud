// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/google/uuid"
	lbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	loadbalanceroptest "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/test"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_replicator/replicator"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	_, err = replicator.NewLoadBalancerReplicator(ctx, k8sManager, listerWatcherTestEnv.LoadBalancerPrivateServiceClient, "us-dev-1", "us-dev-1a")
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

func NewLoadBalancer(namespace string, loadbalancerName string, resourceVersion string) *lbv1alpha1.Loadbalancer {
	lb := loadbalanceroptest.NewLoadbalancer(loadbalancerName, namespace)
	lb.ObjectMeta.ResourceVersion = resourceVersion
	return lb
}

var _ = Describe("Replicator", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	It("When replicator receives a new loadbalancer from GRPC Watch, it should create a loadbalancer in K8s", func() {
		testEnv := NewReplicatorTestEnv()
		defer testEnv.Stop()

		By("Get list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark("10")
		close(listResponseChan)

		By("Get watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending loadbalancer to watch response channel")
		namespace := uuid.NewString()
		loadbalancerName := uuid.NewString()
		loadbalancer := NewLoadBalancer(namespace, loadbalancerName, "11")
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer)

		By("Waiting for loadbalancer to be created in K8s")
		loadbalancerLookupKey := types.NamespacedName{Namespace: namespace, Name: loadbalancerName}
		loadbalancerRef := &lbv1alpha1.Loadbalancer{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, loadbalancerLookupKey, loadbalancerRef)).Should(Succeed())
		}, "10s").Should(Succeed())
	})

	It("When loadbalancer status is updated in K8s, replicator should call GRPC UpdateStatus", func() {
		testEnv := NewReplicatorTestEnv()
		defer testEnv.Stop()

		By("Get list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark("10")
		close(listResponseChan)

		By("Get watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending loadbalancer to watch response channel")
		namespace := uuid.NewString()
		loadbalancerName := uuid.NewString()
		loadbalancer := NewLoadBalancer(namespace, loadbalancerName, "11")
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer)

		By("Waiting for loadbalancer to be created in K8s")
		loadbalancerLookupKey := types.NamespacedName{Namespace: namespace, Name: loadbalancerName}
		loadbalancerRef := &lbv1alpha1.Loadbalancer{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, loadbalancerLookupKey, loadbalancerRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Update loadbalancer status in K8s (simulate loadbalancer Operator)")
		loadbalancerRef.Status.State = lbv1alpha1.READY
		loadbalancerRef.Status.Vip = "1.2.3.4"
		loadbalancerRef.Status.Conditions = lbv1alpha1.ConditionsStatus{
			Listeners: []lbv1alpha1.ConditionsListenerStatus{{
				Port:          9090,
				PoolCreated:   true,
				VIPPoolLinked: true,
				VIPCreated:    true,
			}},
			FirewallRuleCreated: true,
		}
		loadbalancerRef.Status.Listeners = []lbv1alpha1.ListenerStatus{{
			Port:    9090,
			Name:    "mylb",
			Message: "current status is foo",
			PoolMembers: []lbv1alpha1.PoolStatusMember{{
				InstanceResourceId: "a-b-c-d",
				IPAddress:          "1.1.1.1",
			}, {
				InstanceResourceId: "c-d-e-f",
				IPAddress:          "2.2.2.2",
			}},
			PoolID: 11,
			VipID:  12,
		}}

		Expect(k8sClient.Status().Update(ctx, loadbalancerRef)).Should(Succeed())

		By("Waiting for GRPC UpdateStatus to be called with updated status")
		Eventually(func(g Gomega) {
			value, ok := testEnv.UpdateStatusRecorder.Load(loadbalancerName)
			g.Expect(ok).Should(BeTrue())
			req, ok := value.(*pb.LoadBalancerUpdateStatusRequest)
			g.Expect(ok).Should(BeTrue())
			g.Expect(req.Status.State).Should(Equal(string(loadbalancerRef.Status.State)))
			g.Expect(req.Status.Vip).Should(Equal(string(loadbalancerRef.Status.Vip)))
			g.Expect(len(req.Status.Listeners)).Should(Equal(len(loadbalancerRef.Status.Listeners)))
			g.Expect(req.Status.Conditions.FirewallRuleCreated).Should(Equal(loadbalancerRef.Status.Conditions.FirewallRuleCreated))
			g.Expect(len(req.Status.Conditions.Listeners)).Should(Equal(len(loadbalancerRef.Status.Conditions.Listeners)))
		}, "10s").Should(Succeed())

		close(watchResponseChan)
	})

	It("When replicator receives a loadbalancer with a deletionTimestamp, it should delete the loadbalancer from K8s and call RemoveFinalizer", func() {
		testEnv := NewReplicatorTestEnv()
		defer testEnv.Stop()

		By("Get list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark("10")
		close(listResponseChan)

		By("Get watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending loadbalancer to watch response channel")
		namespace := uuid.NewString()
		loadbalancerName := uuid.NewString()
		loadbalancer := NewLoadBalancer(namespace, loadbalancerName, "11")
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer)

		By("Waiting for loadbalancer to be created in K8s")
		loadbalancerLookupKey := types.NamespacedName{Namespace: namespace, Name: loadbalancerName}
		loadbalancerRef := &lbv1alpha1.Loadbalancer{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, loadbalancerLookupKey, loadbalancerRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Sending loadbalancer with deletionTimestamp to watch response channel")
		now := metav1.Now()
		loadbalancer.ObjectMeta.DeletionTimestamp = &now
		loadbalancer.ObjectMeta.ResourceVersion = "12"
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer)

		By("Waiting for loadbalancer to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, loadbalancerLookupKey, loadbalancerRef))).Should(BeTrue())
		}, "10s").Should(Succeed())

		By("Waiting for GRPC RemoveFinalizer to be called")
		Eventually(func(g Gomega) {
			_, ok := testEnv.RemoveFinalizerRecorder.Load(loadbalancerName)
			g.Expect(ok).Should(BeTrue())
		}, "10s").Should(Succeed())

		close(watchResponseChan)
	})
})
