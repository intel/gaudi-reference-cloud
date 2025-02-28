// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"time"

	"github.com/google/uuid"
	lbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_replicator/convert"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_replicator/replicator"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	toolscache "k8s.io/client-go/tools/cache"
)

// A test environment that includes a ListerWatcher and an Informer.
type InformerTestEnv struct {
	BaseTestEnv
	Cancel   context.CancelFunc
	Lw       *replicator.LoadBalancerListerWatcher
	Informer toolscache.SharedIndexInformer
}

func NewInformerTestEnv(timeout time.Duration) *InformerTestEnv {
	ctx := context.Background()
	listWatcherTestEnv := NewBaseTestEnv()

	converter, err := convert.NewLoadBalancerConverter("us-dev-1", "us-dev-1a")
	Expect(err).Should(Succeed())

	lw, err := replicator.NewLoadBalancerListerWatcher(listWatcherTestEnv.LoadBalancerPrivateServiceClient, timeout, converter)
	Expect(err).Should(Succeed())

	informer := toolscache.NewSharedIndexInformer(lw, &lbv1alpha1.Loadbalancer{}, 0, toolscache.Indexers{})
	ctx, cancel := context.WithCancel(ctx)
	By("Starting informer")
	go informer.Run(ctx.Done())
	testEnv := &InformerTestEnv{
		BaseTestEnv: *listWatcherTestEnv,
		Cancel:      cancel,
		Lw:          lw,
		Informer:    informer,
	}
	return testEnv
}

func EnsureInformerHasResourceVersion(g Gomega, informer toolscache.SharedIndexInformer, namespace string, loadbalancerName string, resourceVersion string) {
	store := informer.GetStore()
	storedItem, exists, err := store.GetByKey(namespace + "/" + loadbalancerName)
	g.Expect(err).Should(Succeed())
	g.Expect(exists).Should(BeTrue())
	storedLoadBalancer := storedItem.(*lbv1alpha1.Loadbalancer)
	g.Expect(storedLoadBalancer.ResourceVersion).Should(Equal(resourceVersion))
}

var _ = Describe("ListerWatcher Unit Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	It("When list returns no loadbalancers, informer store should contain no loadbalancers", func() {
		testEnv := NewInformerTestEnv(1 * time.Hour)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark("11")
		close(listResponseChan)

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "1s").Should(Succeed())

		store := testEnv.Informer.GetStore()
		keys := store.ListKeys()
		Expect(keys).Should(Equal([]string{}))
	})

	It("When list returns a single loadbalancer, informer store should contain the loadbalancer", func() {
		testEnv := NewInformerTestEnv(1 * time.Hour)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending initial loadbalancers to watch response channel")
		namespace := uuid.NewString()
		loadbalancerName := uuid.NewString()
		loadbalancer := NewLoadBalancer(namespace, loadbalancerName, "11")
		listResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer)

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(loadbalancer.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "1s").Should(Succeed())

		store := testEnv.Informer.GetStore()
		storedItem, exists, err := store.GetByKey(namespace + "/" + loadbalancerName)
		Expect(err).Should(Succeed())
		Expect(exists).Should(BeTrue())
		storedLoadbalancer := storedItem.(*lbv1alpha1.Loadbalancer)
		Expect(storedLoadbalancer.ResourceVersion).Should(Equal(loadbalancer.ResourceVersion))
	})

	It("When watch returns an updated loadbalancer, informer store should eventually contain the updated loadbalancer", func() {
		testEnv := NewInformerTestEnv(1 * time.Hour)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending initial loadbalancers to watch response channel")
		namespace := uuid.NewString()
		loadbalancerName := uuid.NewString()
		loadbalancer := NewLoadBalancer(namespace, loadbalancerName, "11")
		listResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer)

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(loadbalancer.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "1s").Should(Succeed())

		By("Create watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending updated loadbalancer to watch response channel")
		loadbalancer.ResourceVersion = "12"
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer)

		By("Waiting for informer store to contain the updated loadbalancer (new ResourceVersion)")
		Eventually(func(g Gomega) {
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, loadbalancerName, loadbalancer.ResourceVersion)
		}, "1s").Should(Succeed())

		By("Closing watch response channel")
		close(watchResponseChan)
	})

	It("When watch returns an error, informer should recover by calling List and replacing the cache with the List results", func() {
		testEnv := NewInformerTestEnv(1 * time.Hour)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending loadbalancer1 to list response channel")
		namespace := uuid.NewString()
		loadbalancer1Name := "loadbalancer1-" + uuid.NewString()
		loadbalancer1 := NewLoadBalancer(namespace, loadbalancer1Name, "11")
		listResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer1)

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(loadbalancer1.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "1s").Should(Succeed())

		By("Create watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending updated loadbalancer to watch response channel")
		loadbalancer1.ResourceVersion = "12"
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer1)

		By("Waiting for informer store to contain the updated loadbalancer")
		Eventually(func(g Gomega) {
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, loadbalancer1Name, loadbalancer1.ResourceVersion)
		}, "1s").Should(Succeed())

		By("Injecting failure so that Watch returns an error")
		watchResponseChan <- NewWatchResponseInjectFailure()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan = testEnv.NextListResponseChannel()

		By("Simulate that loadbalancer1 was deleted and loadbalancer2 was created during failure recovery. Send only loadbalancer2 to list response channel.")
		loadbalancer2Name := "loadbalancer2-" + uuid.NewString()
		loadbalancer2 := NewLoadBalancer(namespace, loadbalancer2Name, "13")
		listResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer2)

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(loadbalancer2.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer store to contain only loadbalancer2")
		Eventually(func(g Gomega) {
			keys := testEnv.Informer.GetStore().ListKeys()
			g.Expect(keys).Should(Equal([]string{namespace + "/" + loadbalancer2Name}))
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, loadbalancer2Name, loadbalancer2.ResourceVersion)
		}, "10s").Should(Succeed())

		By("Closing watch response channel")
		close(watchResponseChan)
	})

	It("When GRPC SearchStreamPrivate does not respond within the timeout, informer should retry", func() {
		testEnv := NewInformerTestEnv(1 * time.Second)
		defer testEnv.Stop()

		By("Simulate a SearchStreamPrivate RPC that never returns from Recv")
		listResponseChan1 := testEnv.NextListResponseChannel()
		_ = listResponseChan1

		By("Ensure that informer is not synced")
		Consistently(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeFalse())
		}, "1s").Should(Succeed())

		By("Create 2nd list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan2 := testEnv.NextListResponseChannel()

		By("Sending Bookmark to 2nd list response channel")
		listResponseChan2 <- NewWatchResponseBookmark("11")
		close(listResponseChan2)

		By("Create watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		_ = testEnv.NextWatchResponseChannel()

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "60s").Should(Succeed())

		close(listResponseChan1)
	})

	It("When GRPC Watch is idle for too long, informer should retry", func() {
		testEnv := NewInformerTestEnv(1 * time.Second)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark("11")
		close(listResponseChan)

		By("Create watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending Bookmark to watch response channel")
		watchResponseChan <- NewWatchResponseBookmark("11")

		By("Create 2nd list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan2 := testEnv.NextListResponseChannel()

		By("Simulate that loadbalancer was created during failure recovery.")
		namespace := uuid.NewString()
		loadbalancerName := uuid.NewString()
		loadbalancer := NewLoadBalancer(namespace, loadbalancerName, "12")
		listResponseChan2 <- NewWatchResponse(pb.WatchDeltaType_Updated, loadbalancer)

		By("Sending Bookmark to list response channel")
		listResponseChan2 <- NewWatchResponseBookmark(loadbalancer.ResourceVersion)
		close(listResponseChan2)

		By("Waiting for informer store to contain loadbalancer")
		Eventually(func(g Gomega) {
			keys := testEnv.Informer.GetStore().ListKeys()
			g.Expect(keys).Should(Equal([]string{namespace + "/" + loadbalancerName}))
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, loadbalancerName, loadbalancer.ResourceVersion)
		}, "10s").Should(Succeed())

		By("Closing watch response channel")
		close(watchResponseChan)
	})
})
