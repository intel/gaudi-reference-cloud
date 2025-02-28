// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package metering_monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	privatecloud "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/vast_storage_operator/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	k8sv1 "k8s.io/api/core/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func NewStorage(namespace string, storageName string) *privatecloudv1alpha1.VastStorage {
	return NewStorageInit(
		namespace,
		storageName,
		"az1",
		storageName,
	)
}

func NewMockMeteringServiceClient() (pb.MeteringServiceClient, *sync.Map) {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	mockController := gomock.NewController(GinkgoT())
	meteringClient := pb.NewMockMeteringServiceClient(mockController)

	// Mock Create stores the storage record when storage enters ready state or is deleted.
	createMeteringRecorder := &sync.Map{}
	mockCreate := func(ctx context.Context, req *pb.UsageCreate, opts ...grpc.CallOption) (*emptypb.Empty, error) {
		createMeteringRecorder.Store(req.CloudAccountId, req)
		return &emptypb.Empty{}, nil
	}

	meteringClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(mockCreate).AnyTimes()
	return meteringClient, createMeteringRecorder
}

var _ = Describe("Storage Metering Monitor", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	var createMeteringRecorder *sync.Map
	var managerStoppable *stoppable.Stoppable

	BeforeEach(func() {
		By("Creating manager")
		k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
			Scheme:  scheme,
			Metrics: metricsserver.Options{BindAddress: "0"},
		})
		Expect(err).ToNot(HaveOccurred())

		By("Creating mock metering Client")
		var meteringClient pb.MeteringServiceClient
		meteringClient, createMeteringRecorder = NewMockMeteringServiceClient()

		By("Creating metering monitor")
		config := &privatecloudv1alpha1.VastMeteringMonitorConfig{
			MaxUsageRecordSendIntervalMinutes: 1,
		}
		monitor, err = NewMeteringMonitor(ctx, k8sManager, meteringClient, config)
		Expect(err).Should(Succeed())
		Expect(monitor).NotTo(BeNil())

		By("Starting manager")
		managerStoppable = stoppable.New(k8sManager.Start)
		managerStoppable.Start(ctx)
	})

	AfterEach(func() {
		By("Stopping manager")
		Expect(managerStoppable.Stop(ctx)).Should(Succeed())
		By("Manager stopped")
	})

	It("When a Storage is created and in Running Condition, storage metering montior should create a record in metering db", func() {
		namespace := uuid.NewString()
		filesystemName := uuid.NewString()
		storage := NewStorage(namespace, filesystemName)
		storageLookupKey := types.NamespacedName{Namespace: namespace, Name: filesystemName}
		storageRef := &privatecloudv1alpha1.VastStorage{}
		nsObject := NewNamespace(namespace)

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("creating vast storage successfully")
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
			Status: k8sv1.ConditionTrue,
		}
		storageRef.Status.Conditions = append(storageRef.Status.Conditions, runningCondition)
		Expect(k8sClient.Status().Update(ctx, storageRef)).Should(Succeed())

		By("Confirming Running condition to be true")
		runningCond := FindStatusCondition(storageRef.Status.Conditions, privatecloudv1alpha1.StorageConditionRunning)
		Expect(runningCond).ShouldNot(BeNil())
		Expect(runningCond.Status).Should(Equal(k8sv1.ConditionTrue))

		By("Waiting for metering record to be recorded")
		Eventually(func(g Gomega) {
			value, ok := createMeteringRecorder.Load(namespace)
			g.Expect(ok).Should(BeTrue())
			req, ok := value.(*pb.UsageCreate)
			g.Expect(ok).Should(BeTrue())
			g.Expect(req.Properties["firstReadyTimestamp"]).Should(Equal(runningCond.LastTransitionTime.Format(time.RFC3339)))
			g.Expect(req.Properties["deleted"]).Should(Equal("false"))
		}, timeout, interval).Should(Succeed())

		By("Waiting for storage to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Delete(ctx, storageRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())
	})

	It("When metering montior receives an storage with a deletionTimestamp, it should create a record in metering db before deleting the storage", func() {
		namespace := uuid.NewString()
		filesystemName := uuid.NewString()
		storage := NewStorage(namespace, filesystemName)
		storageLookupKey := types.NamespacedName{Namespace: namespace, Name: filesystemName}
		storageRef := &privatecloudv1alpha1.VastStorage{}
		nsObject := NewNamespace(namespace)
		finalizers := []string{privatecloud.VastStorageMeteringMonitorFinalizer}

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("creating vast storage successfully")
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
			Status: k8sv1.ConditionTrue,
		}
		storageRef.Status.Conditions = append(storageRef.Status.Conditions, runningCondition)
		Expect(k8sClient.Status().Update(ctx, storageRef)).Should(Succeed())

		By("Confirming StartupComplete condition to be true")
		runningCond := FindStatusCondition(storageRef.Status.Conditions, privatecloudv1alpha1.StorageConditionRunning)
		Expect(runningCond).ShouldNot(BeNil())
		Expect(runningCond.Status).Should(Equal(k8sv1.ConditionTrue))

		By("Sending storage with deletionTimestamp and metering monitor finalizer (simulate Storage Operator) to Metering Monitor")
		now := metav1.Now()
		storage.ObjectMeta.DeletionTimestamp = &now
		storageRef.SetFinalizers(finalizers)
		err := k8sClient.Update(ctx, storageRef)
		Expect(err).Should(Succeed(), fmt.Errorf("error occured while updating storage metadata: %w", err))

		By("Waiting for storage to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Delete(ctx, storageRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for storage to be not found so as to confirm its deletion")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, storageLookupKey, storageRef))).Should(BeTrue())
		}, timeout, interval).Should(Succeed())

		By("Creating a Record in the metering db with deleted property set to true")
		Eventually(func(g Gomega) {
			value, ok := createMeteringRecorder.Load(namespace)
			g.Expect(ok).Should(BeTrue())
			req, ok := value.(*pb.UsageCreate)
			g.Expect(ok).Should(BeTrue())
			g.Expect(req.Properties["firstReadyTimestamp"]).Should(Equal(runningCond.LastTransitionTime.Format(time.RFC3339)))
			g.Expect(req.Properties["deleted"]).Should(Equal("true"))
		}, timeout, interval).Should(Succeed())
	})

	It("When an Storage is created, Storage metering montior should create a record in metering db when requeue interval is 60s", func() {
		namespace := uuid.NewString()
		storageName := uuid.NewString()
		storage := NewStorage(namespace, storageName)
		storageLookupKey := types.NamespacedName{Namespace: namespace, Name: storageName}
		storageRef := &privatecloudv1alpha1.VastStorage{}
		nsObject := NewNamespace(namespace)

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("creating vast storage successfully")
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
			Status: k8sv1.ConditionTrue,
		}
		storageRef.Status.Conditions = append(storageRef.Status.Conditions, runningCondition)
		Expect(k8sClient.Status().Update(ctx, storageRef)).Should(Succeed())

		By("Waiting for a metering records to be recorded within timeout")
		var numberOfRecords int
		Eventually(func(g Gomega) {
			createMeteringRecorder.Range(func(key any, value any) bool {
				numberOfRecords += 1
				return true
			})
			g.Expect(numberOfRecords).Should(BeNumerically(">=", 1))
		}, "70s", interval).Should(Succeed())
	})
	Context("When supplying invalid agruments to metering monitor constructor", func() {
		It("should fail and return an error", func() {
			By("Creating manager")
			//malform the k8sManager
			k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
				//Scheme:             scheme,
				Metrics: metricsserver.Options{BindAddress: "0"},
			})
			By("Creating mock metering Client")
			var meteringClient pb.MeteringServiceClient
			meteringClient, _ = NewMockMeteringServiceClient()
			cfg := &privatecloudv1alpha1.VastMeteringMonitorConfig{
				MaxUsageRecordSendIntervalMinutes: 1,
				ServiceType:                       "FileStorageAsAService",
				Region:                            "test",
			}
			resp, err := NewMeteringMonitor(ctx, k8sManager, meteringClient, cfg)
			Expect(resp).To(BeNil())
			Expect(err).NotTo(BeNil())
		})
	})

	Context("When find status condition fails", func() {
		It("Should return nil", func() {
			By("Creating a vast storage object")
			storageRef := &privatecloudv1alpha1.VastStorage{}
			By("Setting condition values")
			runningCondition := privatecloudv1alpha1.StorageCondition{
				Type:   privatecloudv1alpha1.StorageConditionRunning,
				Status: k8sv1.ConditionTrue,
			}
			storageRef.Status.Conditions = append(storageRef.Status.Conditions, runningCondition)
			By("Confirming Running condition to be true")
			failedCond := FindStatusCondition(storageRef.Status.Conditions, privatecloudv1alpha1.StorageConditionFailed)
			Expect(failedCond).To(BeNil())
		})
	})
	Context("Create Metering Record", func() {
		It("Should succeed ", func() {
			namespace := uuid.NewString()
			filesystemName := uuid.NewString()
			storage := NewStorage(namespace, filesystemName)
			//storage.Status.Conditions[0].Type = privatecloudv1alpha1.StorageConditionRunning
			time2 := metav1.Time{Time: time.Now()}
			time1 := time2.Sub(time.Now())
			err := monitor.CreateRecordInDB(context.Background(), storage, time1, time2, false)
			Expect(err).To(BeNil())
		})
	})
	Context("processStorage on getting vast storage object with failed state", func() {
		It("Should skip creating metering record", func() {
			namespace := uuid.NewString()
			filesystemName := uuid.NewString()
			storage := NewStorage(namespace, filesystemName)
			storage.Status.Conditions[0].Type = privatecloudv1alpha1.StorageConditionFailed
			res, err := monitor.processStorage(ctx, storage)
			Expect(res).NotTo(BeNil())
			Expect(err).To(BeNil())

		})
	})

	Context("processDeleteStorage on getting vast storage object with failed state", func() {
		It("Should skip creating metering record", func() {
			namespace := uuid.NewString()
			filesystemName := uuid.NewString()
			storage := NewStorage(namespace, filesystemName)
			Expect(storage).NotTo(BeNil())
			storageRef := &privatecloudv1alpha1.VastStorage{}
			failedCondition := privatecloudv1alpha1.StorageCondition{
				Type:   privatecloudv1alpha1.StorageConditionFailed,
				Status: k8sv1.ConditionTrue,
			}
			storageRef.Status.Conditions = append(storageRef.Status.Conditions, failedCondition)
			//storage.Status.Conditions[0].Type = privatecloudv1alpha1.StorageConditionFailed
			res, err := monitor.processDeleteStorage(ctx, storageRef)
			Expect(res).NotTo(BeNil())
			Expect(err).NotTo(BeNil())

		})
	})
})
