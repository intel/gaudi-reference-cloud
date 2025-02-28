// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	v1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	privatecloud "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/object_store_operator/controllers"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ObjectStore tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	log.Info("Begin")

	var (
		managerStoppable *stoppable.Stoppable
		strclient        *sc.StorageControllerClient
	)

	BeforeEach(func() {
		By("Creating manager")
		k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
			Scheme:  scheme,
			Metrics: metricsserver.Options{BindAddress: "0"},
		})
		Expect(err).ToNot(HaveOccurred())

		By("Creating Mock Storage Controller client")
		strclient = NewMockStorageControllerClient()
		Expect(strclient).NotTo(BeNil())

		By("Creating ObjectStore controller")
		client, err := privatecloud.NewObjectStoreOperator(ctx, k8sManager, strclient)
		Expect(err).Should(Succeed())
		Expect(client).NotTo(BeNil())

		By("Starting manager")
		managerStoppable = stoppable.New(k8sManager.Start)
		managerStoppable.Start(ctx)
	})
	AfterEach(func() {
		By("Stopping manager")
		Expect(managerStoppable.Stop(ctx)).Should(Succeed())
		By("Manager stopped")
	})
	Context("ObjectStore tests", func() {
		//Object references
		namespace := "123456789012"
		bucketName := "test-bucket"
		bucket := NewBucket(namespace, bucketName)
		bucketLookupKey := types.NamespacedName{Namespace: namespace, Name: bucketName}
		bucketRef := &v1alpha1.ObjectStore{}
		nsObject := NewNamespace(namespace)
		It("Should reconcile bucket for creation", func() {
			By("Create namespace")
			Expect(nsObject).NotTo(BeNil())
			Expect(bucketLookupKey).NotTo(BeNil())
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

			By("Create the bucket")
			Expect(k8sClient.Create(ctx, bucket)).Should(Succeed())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Waiting for bucket to be created in K8s")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Checking finalizer have been added to the bucket CRD")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
				g.Expect(bucketRef.GetFinalizers()).Should(ConsistOf(privatecloud.BucketMeteringMonitorFinalizer, privatecloud.ObjectFinalizer))
			}, timeout, interval).Should(Succeed())
			k8sClient.Get(ctx, bucketLookupKey, bucketRef)

			By("Checking the bucket condition")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
				cond := privatecloud.FindStatusCondition(bucketRef.Status.Conditions, v1alpha1.ObjectStoreConditionAccepted)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())

			By("Checking the bucket phase")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
				log.Info("Object Store Reconciler should set Phase to Ready", logkeys.StatusPhase, bucketRef.Status.Phase,
					logkeys.StatusMessage, bucketRef.Status.Message, logkeys.StatusConditions, bucketRef.Status.Conditions)
				g.Expect(bucketRef.Status.Phase).Should(Equal(v1alpha1.ObjectStorePhasePhaseReady))
			}, timeout, interval).Should(Succeed())

			//Delete bucket
			By("Setting deletion timestamp to mark object for deletion")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Delete(ctx, bucketRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Removing finalizers so object can be deleted")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
				bucketRef.SetFinalizers(nil)
				g.Expect(k8sClient.Update(ctx, bucketRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Waiting for storage to be not found so as to confirm its deletion")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, bucketLookupKey, bucketRef)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
		It("Should reconcile bucket for deletion", func() {
			namespace := "123456789013"
			bucketName := "test-bucket2"
			bucket := NewBucket(namespace, bucketName)
			bucketLookupKey := types.NamespacedName{Namespace: namespace, Name: bucketName}
			bucketRef := &v1alpha1.ObjectStore{}
			nsObject := NewNamespace(namespace)
			Expect(nsObject).NotTo(BeNil())
			Expect(bucketLookupKey).NotTo(BeNil())
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

			By("Create the bucket")
			Expect(k8sClient.Create(ctx, bucket)).Should(Succeed())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Adding finalizer to the bucket CRD")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
				g.Expect(bucketRef.GetFinalizers()).Should(ConsistOf(privatecloud.BucketMeteringMonitorFinalizer, privatecloud.ObjectFinalizer))
			}, timeout, interval).Should(Succeed())

			By("Checking the bucket condition")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
				cond := privatecloud.FindStatusCondition(bucketRef.Status.Conditions, v1alpha1.ObjectStoreConditionAccepted)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
				log.Info("Object Store Reconciler should set Phase to Ready", "bucketRef.Status.Phase", bucketRef.Status.Phase,
					"bucketRef.Status.Message", bucketRef.Status.Message, "conditions", bucketRef.Status.Conditions)
				g.Expect(bucketRef.Status.Phase).Should(Equal(v1alpha1.ObjectStorePhasePhaseReady))
			}, timeout, interval).Should(Succeed())

			By("Setting deletion timestamp to mark object for deletion")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Delete(ctx, bucketRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Removing finalizers so object can be deleted")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
				bucketRef.SetFinalizers(nil)
				g.Expect(k8sClient.Update(ctx, bucketRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Waiting for storage to be not found so as to confirm its deletion")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, bucketLookupKey, bucketRef)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

		})

		ns := ""
		bName := ""
		bk := NewBucket(ns, bName)
		Expect(bk).NotTo(BeNil())
		bKey := types.NamespacedName{Namespace: ns, Name: bName}
		bRef := &v1alpha1.ObjectStore{}
		nsObject2 := NewNamespace(ns)
		It("Should give error while creating namespace", func() {
			Expect(k8sClient.Create(ctx, nsObject2)).ShouldNot(Succeed())
		})
		It("Should give error when creating bucket", func() {
			Expect(k8sClient.Create(ctx, nsObject2)).ShouldNot(Succeed())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bKey, bRef)).ShouldNot(Succeed())
			}, timeout, interval).Should(Succeed())
		})
	})
})
