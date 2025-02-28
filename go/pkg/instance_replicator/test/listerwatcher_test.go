// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_replicator/replicator"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
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
	Lw       *replicator.ListerWatcher
	Informer toolscache.SharedIndexInformer
}

func NewInformerTestEnv(timeout time.Duration) *InformerTestEnv {
	ctx := context.Background()
	listWatcherTestEnv := NewBaseTestEnv()
	lw := replicator.NewListerWatcher(listWatcherTestEnv.InstancePrivateServiceClient, timeout)
	informer := toolscache.NewSharedIndexInformer(lw, &privatecloudv1alpha1.Instance{}, 0, toolscache.Indexers{})
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

func EnsureInformerHasResourceVersion(g Gomega, informer toolscache.SharedIndexInformer, namespace string, instanceName string, resourceVersion string) {
	store := informer.GetStore()
	storedItem, exists, err := store.GetByKey(namespace + "/" + instanceName)
	g.Expect(err).Should(Succeed())
	g.Expect(exists).Should(BeTrue())
	storedInstance := storedItem.(*privatecloudv1alpha1.Instance)
	g.Expect(storedInstance.ResourceVersion).Should(Equal(resourceVersion))
}

var _ = Describe("ListerWatcher Unit Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	It("When list returns no instances, informer store should contain no instances", func() {
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

	It("When list returns a single instance, informer store should contain the instance", func() {
		testEnv := NewInformerTestEnv(1 * time.Hour)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending initial instances to watch response channel")
		namespace := uuid.NewString()
		instanceName := uuid.NewString()
		instance := NewInstance(namespace, instanceName, "11")
		listResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, instance)

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(instance.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "1s").Should(Succeed())

		store := testEnv.Informer.GetStore()
		storedItem, exists, err := store.GetByKey(namespace + "/" + instanceName)
		Expect(err).Should(Succeed())
		Expect(exists).Should(BeTrue())
		storedInstance := storedItem.(*privatecloudv1alpha1.Instance)
		Expect(storedInstance.ResourceVersion).Should(Equal(instance.ResourceVersion))
	})

	It("When watch returns an updated instance, informer store should eventually contain the updated instance", func() {
		testEnv := NewInformerTestEnv(1 * time.Hour)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending initial instances to watch response channel")
		namespace := uuid.NewString()
		instanceName := uuid.NewString()
		instance := NewInstance(namespace, instanceName, "11")
		listResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, instance)

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(instance.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "1s").Should(Succeed())

		By("Create watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending updated instance to watch response channel")
		instance.ResourceVersion = "12"
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, instance)

		By("Waiting for informer store to contain the updated instance (new ResourceVersion)")
		Eventually(func(g Gomega) {
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, instanceName, instance.ResourceVersion)
		}, "1s").Should(Succeed())

		By("Closing watch response channel")
		close(watchResponseChan)
	})

	It("When watch returns an error, informer should recover by calling List and replacing the cache with the List results", func() {
		testEnv := NewInformerTestEnv(1 * time.Hour)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending instance1 to list response channel")
		namespace := uuid.NewString()
		instance1Name := "instance1-" + uuid.NewString()
		instance1 := NewInstance(namespace, instance1Name, "11")
		listResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, instance1)

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(instance1.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "1s").Should(Succeed())

		By("Create watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending updated instance to watch response channel")
		instance1.ResourceVersion = "12"
		watchResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, instance1)

		By("Waiting for informer store to contain the updated instance")
		Eventually(func(g Gomega) {
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, instance1Name, instance1.ResourceVersion)
		}, "1s").Should(Succeed())

		By("Injecting failure so that Watch returns an error")
		watchResponseChan <- NewWatchResponseInjectFailure()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan = testEnv.NextListResponseChannel()

		By("Simulate that instance1 was deleted and instance2 was created during failure recovery. Send only instance2 to list response channel.")
		instance2Name := "instance2-" + uuid.NewString()
		instance2 := NewInstance(namespace, instance2Name, "13")
		listResponseChan <- NewWatchResponse(pb.WatchDeltaType_Updated, instance2)

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(instance2.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer store to contain only instance2")
		Eventually(func(g Gomega) {
			keys := testEnv.Informer.GetStore().ListKeys()
			g.Expect(keys).Should(Equal([]string{namespace + "/" + instance2Name}))
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, instance2Name, instance2.ResourceVersion)
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

		By("Simulate that instance was created during failure recovery.")
		namespace := uuid.NewString()
		instanceName := uuid.NewString()
		instance := NewInstance(namespace, instanceName, "12")
		listResponseChan2 <- NewWatchResponse(pb.WatchDeltaType_Updated, instance)

		By("Sending Bookmark to list response channel")
		listResponseChan2 <- NewWatchResponseBookmark(instance.ResourceVersion)
		close(listResponseChan2)

		By("Waiting for informer store to contain instance")
		Eventually(func(g Gomega) {
			keys := testEnv.Informer.GetStore().ListKeys()
			g.Expect(keys).Should(Equal([]string{namespace + "/" + instanceName}))
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, instanceName, instance.ResourceVersion)
		}, "10s").Should(Succeed())

		By("Closing watch response channel")
		close(watchResponseChan)
	})
})
