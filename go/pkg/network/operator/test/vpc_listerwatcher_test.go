// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	vpcv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/internal/controller/vpc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	toolscache "k8s.io/client-go/tools/cache"
)

// A test environment that includes a ListerWatcher and an Informer.
type InformerTestEnv struct {
	BaseTestEnv
	Cancel   context.CancelFunc
	Lw       *vpc.VPCListerWatcher
	Informer toolscache.SharedIndexInformer
}

func NewInformerTestEnv(timeout time.Duration) *InformerTestEnv {
	ctx := context.Background()
	listWatcherTestEnv := NewBaseTestEnv()
	lw := vpc.NewVPCListerWatcher(listWatcherTestEnv.VPCPrivateServiceClient, timeout)
	informer := toolscache.NewSharedIndexInformer(lw, &vpcv1alpha1.VPC{}, 0, toolscache.Indexers{})
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

func EnsureInformerHasResourceVersion(g Gomega, informer toolscache.SharedIndexInformer, namespace string, vpcName string, resourceVersion string) {
	store := informer.GetStore()
	storedItem, exists, err := store.GetByKey(namespace + "/" + vpcName)
	g.Expect(err).Should(Succeed())
	g.Expect(exists).Should(BeTrue())
	storedVPC := storedItem.(*vpcv1alpha1.VPC)
	g.Expect(storedVPC.ResourceVersion).Should(Equal(resourceVersion))
}

var _ = Describe("ListerWatcher Unit Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	It("When list returns no vpcs, informer store should contain no vpc", func() {
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

	It("When list returns a single vpc, informer store should contain the vpc", func() {
		testEnv := NewInformerTestEnv(1 * time.Hour)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending initial vpc to watch response channel")
		cloudaccountId, err := NewCloudAcctId()
		Expect(err).Should(Succeed())
		vpcName := uuid.NewString()
		vpc, err := NewVPC(vpcName, cloudaccountId, "10.0.0.1/16")
		Expect(err).Should(Succeed())
		resp, err := NewVPCWatchResponse(pb.WatchDeltaType_Updated, vpc)
		Expect(err).Should(Succeed())
		listResponseChan <- resp

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(vpc.Metadata.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "10s").Should(Succeed())

		store := testEnv.Informer.GetStore()
		storedItem, exists, err := store.GetByKey(cloudaccountId + "/" + vpcName)
		Expect(err).Should(Succeed())

		fmt.Println("--- keys: ", store.ListKeys())

		Expect(exists).Should(BeTrue())
		storedVPC := storedItem.(*vpcv1alpha1.VPC)
		Expect(storedVPC.ResourceVersion).Should(Equal(vpc.Metadata.ResourceVersion))
	})

	It("When watch returns an updated vpc, informer store should eventually contain the updated vpc", func() {
		testEnv := NewInformerTestEnv(1 * time.Hour)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending initial vpc to watch response channel")
		namespace := uuid.NewString()
		vpcName := uuid.NewString()
		vpc, err := NewVPC(vpcName, namespace, "10.0.0.1/16")
		resp, err := NewVPCWatchResponse(pb.WatchDeltaType_Updated, vpc)
		Expect(err).Should(Succeed())
		listResponseChan <- resp

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(vpc.Metadata.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "1s").Should(Succeed())

		By("Create watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending updated vpc to watch response channel")
		vpc.Metadata.ResourceVersion = "12"
		resp, err = NewVPCWatchResponse(pb.WatchDeltaType_Updated, vpc)
		Expect(err).Should(Succeed())
		watchResponseChan <- resp

		By("Waiting for informer store to contain the updated vpc (new ResourceVersion)")
		Eventually(func(g Gomega) {
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, vpcName, vpc.Metadata.ResourceVersion)
		}, "1s").Should(Succeed())

		By("Closing watch response channel")
		close(watchResponseChan)
	})

	It("When watch returns an error, informer should recover by calling List and replacing the cache with the List results", func() {
		testEnv := NewInformerTestEnv(1 * time.Hour)
		defer testEnv.Stop()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan := testEnv.NextListResponseChannel()

		By("Sending vpc1 to list response channel")
		namespace := uuid.NewString()
		vpc1Name := "vpc1-" + uuid.NewString()
		vpc1, err := NewVPC(vpc1Name, namespace, "10.0.0.1/16")
		resp, err := NewVPCWatchResponse(pb.WatchDeltaType_Updated, vpc1)
		Expect(err).Should(Succeed())
		listResponseChan <- resp

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(vpc1.Metadata.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer to be synced")
		Eventually(func(g Gomega) {
			g.Expect(testEnv.Informer.HasSynced()).Should(BeTrue())
		}, "1s").Should(Succeed())

		By("Create watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
		watchResponseChan := testEnv.NextWatchResponseChannel()

		By("Sending updated vpc to watch response channel")
		vpc1.Metadata.ResourceVersion = "12"
		resp, err = NewVPCWatchResponse(pb.WatchDeltaType_Updated, vpc1)
		Expect(err).Should(Succeed())
		watchResponseChan <- resp

		By("Waiting for informer store to contain the updated vpc")
		Eventually(func(g Gomega) {
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, vpc1Name, vpc1.Metadata.ResourceVersion)
		}, "1s").Should(Succeed())

		By("Injecting failure so that Watch returns an error")
		watchResponseChan <- NewWatchResponseInjectFailure()

		By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
		listResponseChan = testEnv.NextListResponseChannel()

		By("Simulate that vpc1 was deleted and vpc2 was created during failure recovery. Send only vpc2 to list response channel.")
		vpc2Name := "vpc2-" + uuid.NewString()
		vpc2, err := NewVPC(vpc2Name, namespace, "10.0.0.1/16")
		resp, err = NewVPCWatchResponse(pb.WatchDeltaType_Updated, vpc2)
		Expect(err).Should(Succeed())
		listResponseChan <- resp

		By("Sending Bookmark to list response channel")
		listResponseChan <- NewWatchResponseBookmark(vpc2.Metadata.ResourceVersion)
		close(listResponseChan)

		By("Waiting for informer store to contain only vpc2")
		Eventually(func(g Gomega) {
			keys := testEnv.Informer.GetStore().ListKeys()
			g.Expect(keys).Should(Equal([]string{namespace + "/" + vpc2Name}))
			EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, vpc2Name, vpc2.Metadata.ResourceVersion)
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

	//It("When GRPC Watch is idle for too long, informer should retry", func() {
	//	testEnv := NewInformerTestEnv(1 * time.Second)
	//	defer testEnv.Stop()
	//
	//	By("Create list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
	//	listResponseChan := testEnv.NextListResponseChannel()
	//
	//	By("Sending Bookmark to list response channel")
	//	listResponseChan <- NewWatchResponseBookmark("11")
	//	close(listResponseChan)
	//
	//	By("Create watch response channel that will mock Watch RPC responses to ListerWatcher.Watch")
	//	watchResponseChan := testEnv.NextWatchResponseChannel()
	//
	//	By("Sending Bookmark to watch response channel")
	//	watchResponseChan <- NewWatchResponseBookmark("11")
	//
	//	By("Create 2nd list response channel that will mock SearchStreamPrivate RPC responses to ListerWatcher.List")
	//	listResponseChan2 := testEnv.NextListResponseChannel()
	//
	//	By("Simulate that vpc was created during failure recovery.")
	//	namespace := uuid.NewString()
	//	vpcName := uuid.NewString()
	//	vpc, err := NewVPC(vpcName, namespace, "10.0.0.1/16")
	//	resp, err := NewVPCWatchResponse(pb.WatchDeltaType_Updated, vpc)
	//	Expect(err).Should(Succeed())
	//	listResponseChan2 <- resp
	//
	//	By("Sending Bookmark to list response channel")
	//	listResponseChan2 <- NewWatchResponseBookmark(vpc.Metadata.ResourceVersion)
	//	close(listResponseChan2)
	//
	//	By("Waiting for informer store to contain vpc")
	//	Eventually(func(g Gomega) {
	//		keys := testEnv.Informer.GetStore().ListKeys()
	//		g.Expect(keys).Should(Equal([]string{namespace + "/" + vpcName}))
	//		EnsureInformerHasResourceVersion(g, testEnv.Informer, namespace, vpcName, vpc.Metadata.ResourceVersion)
	//	}, "10s").Should(Succeed())
	//
	//	By("Closing watch response channel")
	//	close(watchResponseChan)
	//})
})
