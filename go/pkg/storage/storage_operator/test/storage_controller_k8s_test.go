// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/google/uuid"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	str "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_operator/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var _ = Describe("Storage Operator", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	var (
		managerStoppable *stoppable.Stoppable
	)
	BeforeEach(func() {
		By("Creating manager")
		k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
			Scheme:  scheme,
			Metrics: metricsserver.Options{BindAddress: "0"},
		})
		Expect(err).ToNot(HaveOccurred())

		client, kms_client := NewMockOperatorServiceClient(true)
		Expect(client).NotTo(BeNil())
		Expect(kms_client).NotTo(BeNil())

		_, err = str.NewStorageOperator(ctx, k8sManager, client, kms_client)

		By("Starting manager")
		managerStoppable = stoppable.New(k8sManager.Start)
		managerStoppable.Start(ctx)
	})

	AfterEach(func() {
		By("Stopping manager")
		Expect(managerStoppable.Stop(ctx)).Should(Succeed())
		By("Manager stopped")
	})
	Context("Failure Operator tests for IKS k8s flow", func() {
		namespace := uuid.NewString()
		filesystemName := uuid.NewString()
		filesystemType := "ComputeKubernetes"
		storage := NewStorage(namespace, filesystemName, filesystemType)
		storageLookupKey := types.NamespacedName{Namespace: namespace, Name: filesystemName}
		storageRef := &privatecloudv1alpha1.Storage{}
		nsObject := NewNamespace(namespace)
		Expect(storage).NotTo(BeNil())
		It("Should create namespace", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create storage successfully for kubernetes flow K8s object", func() {
			Expect(k8sClient.Create(ctx, storage)).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
		})

		It("When a Storage object for is created for k8s flow", func() {
			namespace := uuid.NewString()
			filesystemName := uuid.NewString()
			filesystemType := "ComputeKubernetes"
			storage := NewStorage(namespace, filesystemName, filesystemType)

			Expect(storage).NotTo(BeNil())
		})

		It("Reconciler should add finalizer to storages CRD", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
				g.Expect(storageRef.GetFinalizers()).Should(ConsistOf(str.StorageFinalizer, str.StorageMeteringMonitorFinalizer))
			}, timeout, interval).Should(Succeed())
		})

		It("Storage Reconciler should set Phase to Failed for failed cases for kubernetes flow", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
				log.Info("Storage Reconciler should set Phase to Provisioning", "storageRef.Status.Phase", storageRef.Status.Phase,
					"storageRef.Status.Message", storageRef.Status.Message, "conditions", storageRef.Status.Conditions)
				g.Expect(storageRef.Status.Phase).Should(Equal(privatecloudv1alpha1.FilesystemPhaseFailed))
				g.Expect(storageRef.Status.Message).Should(MatchRegexp(".*%s.*", privatecloudv1alpha1.StorageMessageFailed))
			}, timeout, interval).Should(Succeed())
		})

		It("Waiting for storage to be deleted from K8s", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Delete(ctx, storageRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

		})

		namespace2 := ""
		filesystemName2 := ""
		filesystemType = "ComputeKubernetes"
		storage2 := NewStorage(namespace2, filesystemName, filesystemType)
		storageLookupKey2 := types.NamespacedName{Namespace: namespace2, Name: filesystemName2}
		storageRef2 := &privatecloudv1alpha1.Storage{}
		nsObject2 := NewNamespace(namespace)
		Expect(storage2).NotTo(BeNil())
		It("Should give error while creating namespace", func() {
			Expect(k8sClient.Create(ctx, nsObject2)).ShouldNot(Succeed())
		})

		It("Should give error while creating storage", func() {
			Expect(k8sClient.Create(ctx, nsObject2)).ShouldNot(Succeed())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey2, storageRef2)).ShouldNot(Succeed())
			}, timeout, interval).Should(Succeed())
		})
	})

	It("Delete kubernetes flow and clean resources on the backend", func() {
		namespace := uuid.NewString()
		filesystemName := uuid.NewString()
		filesystemType := "ComputeKubernetes"
		storage := NewStorage(namespace, filesystemName, filesystemType)
		storageLookupKey := types.NamespacedName{Namespace: namespace, Name: filesystemName}
		storageRef := &privatecloudv1alpha1.Storage{}
		nsObject := NewNamespace(namespace)

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("Creating Storage successfully")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Create(ctx, storage)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for storage to be created in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Updating storage status in K8s (simulate Storage Operator)")
		runningCondition := privatecloudv1alpha1.StorageCondition{
			Type:   privatecloudv1alpha1.StorageConditionRunning,
			Status: v1.ConditionTrue,
		}
		storageRef.Status.Conditions = append(storageRef.Status.Conditions, runningCondition)
		Expect(k8sClient.Status().Update(ctx, storageRef)).Should(Succeed())

		By("Waiting for storage to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Delete(ctx, storageRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for storage to be not found so as to confirm its deletion")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, storageLookupKey, storageRef))).Should(BeTrue())
		}, timeout, interval).Should(Succeed())

	})
})
