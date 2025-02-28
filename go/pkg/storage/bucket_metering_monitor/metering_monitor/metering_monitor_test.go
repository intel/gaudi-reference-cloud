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
	v1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"

	privatecloud "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/object_store_operator/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	k8sv1 "k8s.io/api/core/v1"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func NewStorage(namespace string, bucketName string) *v1alpha1.ObjectStore {
	return NewObjectStoreInit(
		namespace,
		bucketName,
		"az1",
		"",
	)
}

func NewMockMeteringServiceClient() (pb.MeteringServiceClient, *sync.Map) {
	mockController := gomock.NewController(GinkgoT())
	meteringClient := pb.NewMockMeteringServiceClient(mockController)

	// Mock Create stores the metering record when bucket enters ready state or is deleted.
	createMeteringRecorder := &sync.Map{}
	mockCreate := func(ctx context.Context, req *pb.UsageCreate, opts ...grpc.CallOption) (*emptypb.Empty, error) {
		createMeteringRecorder.Store(req.CloudAccountId, req)
		return &emptypb.Empty{}, nil
	}

	meteringClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(mockCreate).AnyTimes()
	return meteringClient, createMeteringRecorder
}

func NewMockStorageControllerClient() *sc.StorageControllerClient {
	ctrl := gomock.NewController(GinkgoT())
	mockClient := mocks.NewMockS3ServiceClient(ctrl)
	buckets := []*api.Bucket{}
	bucket := &api.Bucket{
		Id: &api.BucketIdentifier{
			ClusterId: &api.ClusterIdentifier{
				Uuid: "ab03e000-9a4a-48b4-b9be-f9a0ce8f9e84",
			},
			Id: "test-bucket",
		},
		Name:      "test-bucket",
		Versioned: false,
		Capacity: &api.Bucket_Capacity{
			TotalBytes: 10000000000,
		},
	}
	buckets = append(buckets, bucket)
	mockClient.EXPECT().ListBuckets(gomock.Any(), gomock.Any()).Return(&api.ListBucketsResponse{
		Buckets: buckets}, nil).AnyTimes()
	client := &sc.StorageControllerClient{
		S3ServiceClient: mockClient,
	}
	return client
}

var _ = Describe("Bucket Metering Monitor", Serial, func() {
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

		By("Creating mock metering client")
		var meteringClient pb.MeteringServiceClient
		meteringClient, createMeteringRecorder = NewMockMeteringServiceClient()

		By("Create mock storage controller client")
		strclient := NewMockStorageControllerClient()

		By("Creating metering monitor")
		// Convert time.Duration to metav1.Duration
		cfg := &v1alpha1.BucketMeteringMonitorConfig{
			MaxUsageRecordSendIntervalMinutes: 1,
		}
		monitor, err = NewMeteringMonitor(ctx, k8sManager, meteringClient, cfg, strclient)
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

	It("When a bucket is created and in Ready state, bucket metering montior should create a record in metering db", func() {
		namespace := uuid.NewString()
		bucketName := uuid.NewString()
		bucket := NewStorage(namespace, bucketName)
		bucketLookupKey := types.NamespacedName{Namespace: namespace, Name: bucketName}
		bucketRef := &v1alpha1.ObjectStore{}
		nsObject := NewNamespace(namespace)

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("Creating Bucket successfully")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Create(ctx, bucket)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for storage to be created in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Updating storage status in K8s (simulate Storage Operator)")
		runningCondition := v1alpha1.ObjectStoreCondition{
			Type:   v1alpha1.ObjectStoreConditionType(v1alpha1.ObjectStoreConditionAccepted),
			Status: k8sv1.ConditionTrue,
		}
		bucketRef.Status.Conditions = append(bucketRef.Status.Conditions, runningCondition)
		Expect(k8sClient.Status().Update(ctx, bucketRef)).Should(Succeed())

		By("Confirming Accepted condition to be true")
		runningCond := FindStatusCondition(bucketRef.Status.Conditions, v1alpha1.ObjectStoreConditionAccepted)
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

		By("Waiting for bucket to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Delete(ctx, bucketRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())
	})

	It("When metering montior reconciles a bucket with a deletionTimestamp, it should create a record in metering db before deleting the bucket", func() {
		namespace := uuid.NewString()
		bucketName := uuid.NewString()
		bucket := NewStorage(namespace, bucketName)
		bucketLookupKey := types.NamespacedName{Namespace: namespace, Name: bucketName}
		bucketRef := &v1alpha1.ObjectStore{}
		nsObject := NewNamespace(namespace)
		finalizers := []string{privatecloud.BucketMeteringMonitorFinalizer}

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("Creating bucket successfully")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Create(ctx, bucket)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for bucket to be created in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Updating bucket status in K8s (simulate Object Store Operator)")
		acceptedCondition := v1alpha1.ObjectStoreCondition{
			Type:   v1alpha1.ObjectStoreConditionAccepted,
			Status: k8sv1.ConditionTrue,
		}
		bucketRef.Status.Conditions = append(bucketRef.Status.Conditions, acceptedCondition)
		Expect(k8sClient.Status().Update(ctx, bucketRef)).Should(Succeed())

		By("Confirming StartupComplete condition to be true")
		acceptedCond := FindStatusCondition(bucketRef.Status.Conditions, v1alpha1.ObjectStoreConditionAccepted)
		Expect(acceptedCond).ShouldNot(BeNil())
		Expect(acceptedCond.Status).Should(Equal(k8sv1.ConditionTrue))

		By("Sending bucket with deletionTimestamp and metering monitor finalizer (simulate Object Store Operator) to Metering Monitor")
		now := metav1.Now()
		bucket.ObjectMeta.DeletionTimestamp = &now
		bucketRef.SetFinalizers(finalizers)
		err := k8sClient.Update(ctx, bucketRef)
		Expect(err).Should(Succeed(), fmt.Errorf("error occured while updating bucket metadata: %w", err))

		By("Waiting for bucket to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Delete(ctx, bucketRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for bucket to be not found so as to confirm its deletion")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, bucketLookupKey, bucketRef))).Should(BeTrue())
		}, timeout, interval).Should(Succeed())

		By("Creating a Record in the metering db with deleted property set to true")
		Eventually(func(g Gomega) {
			value, ok := createMeteringRecorder.Load(namespace)
			g.Expect(ok).Should(BeTrue())
			req, ok := value.(*pb.UsageCreate)
			g.Expect(ok).Should(BeTrue())
			g.Expect(req.Properties["firstReadyTimestamp"]).Should(Equal(acceptedCond.LastTransitionTime.Format(time.RFC3339)))
			g.Expect(req.Properties["deleted"]).Should(Equal("true"))
		}, timeout, interval).Should(Succeed())
	})

	It("When an Bucket is created, Bucket metering montior should create a record in metering db when requeue interval is 60s", func() {
		namespace := uuid.NewString()
		bucketName := uuid.NewString()
		bucket := NewStorage(namespace, bucketName)
		bucketLookupKey := types.NamespacedName{Namespace: namespace, Name: bucketName}
		bucketRef := &v1alpha1.ObjectStore{}
		nsObject := NewNamespace(namespace)

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("Creating Bucket successfully")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Create(ctx, bucket)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for bucket to be created in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, bucketLookupKey, bucketRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Updating bucket status in K8s (simulate Storage Operator)")
		acceptedCondition := v1alpha1.ObjectStoreCondition{
			Type:   v1alpha1.ObjectStoreConditionAccepted,
			Status: k8sv1.ConditionTrue,
		}
		bucketRef.Status.Conditions = append(bucketRef.Status.Conditions, acceptedCondition)
		Expect(k8sClient.Status().Update(ctx, bucketRef)).Should(Succeed())

		By("Waiting for a metering record to be recorded within timeout")
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
			cfg := &v1alpha1.BucketMeteringMonitorConfig{
				MaxUsageRecordSendIntervalMinutes: 1,
				ServiceType:                       "ObjectStorageAsAService",
				Region:                            "test",
			}
			resp, err := NewMeteringMonitor(ctx, k8sManager, meteringClient, cfg, nil)
			Expect(resp).To(BeNil())
			Expect(err).NotTo(BeNil())
		})
	})

	Context("When find status condition fails", func() {
		It("Should return nil", func() {
			By("Creating a bucket object")
			bucketRef := &v1alpha1.ObjectStore{}
			By("Setting condition values")
			acceptedCondition := v1alpha1.ObjectStoreCondition{
				Type:   v1alpha1.ObjectStoreConditionAccepted,
				Status: k8sv1.ConditionTrue,
			}
			bucketRef.Status.Conditions = append(bucketRef.Status.Conditions, acceptedCondition)
			By("Confirming Accepted condition to be true")
			failedCond := FindStatusCondition(bucketRef.Status.Conditions, v1alpha1.ObjectStoreConditionFailed)
			Expect(failedCond).To(BeNil())
		})
	})
	Context("Create Metering Record", func() {
		It("Should succeed ", func() {
			namespace := uuid.NewString()
			bucketName := uuid.NewString()
			bucket := NewStorage(namespace, bucketName)
			//bucket.Status.Conditions[0].Type = privatecloudv1alpha1.StorageConditionRunning
			time2 := metav1.Time{Time: time.Now()}
			time1 := time2.Sub(time.Now())
			err := monitor.CreateRecordInDB(context.Background(), bucket, time1, time2, false)
			Expect(err).To(BeNil())
		})
	})
	Context("processObjectStorage on getting bucket object with failed state", func() {
		It("Should skip creating metering record1", func() {
			namespace := uuid.NewString()
			bucketName := uuid.NewString()
			bucket := NewStorage(namespace, bucketName)
			bucket.Status.Conditions[0].Type = v1alpha1.ObjectStoreConditionFailed
			res, err := monitor.processObjectStorage(ctx, bucket)
			Expect(res).NotTo(BeNil())
			Expect(err).To(BeNil())

		})
	})

	Context("processDeleteObjectStorage on getting storage object with failed state", func() {
		It("Should skip creating metering record2", func() {
			namespace := uuid.NewString()
			bucketName := uuid.NewString()
			bucket := NewStorage(namespace, bucketName)
			Expect(bucket).NotTo(BeNil())
			bucketRef := &v1alpha1.ObjectStore{
				Status: v1alpha1.ObjectStoreStatus{
					Bucket: v1alpha1.ObjectStoreBucket{
						Name: bucketName,
					},
				},
			}
			failedCondition := v1alpha1.ObjectStoreCondition{
				Type:   v1alpha1.ObjectStoreConditionFailed,
				Status: k8sv1.ConditionTrue,
			}
			bucketRef.Status.Conditions = append(bucketRef.Status.Conditions, failedCondition)
			//bucket.Status.Conditions[0].Type = privatecloudv1alpha1.StorageConditionFailed
			res, err := monitor.processDeleteObjectStorage(ctx, bucketRef)
			Expect(res).NotTo(BeNil())
			Expect(err).NotTo(BeNil())

		})
	})
})
