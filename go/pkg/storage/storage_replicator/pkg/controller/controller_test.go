// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"errors"
	"time"

	"github.com/golang/mock/gomock"
	pc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_replicator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

var _ = Describe("Replicator", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	var (
		mockClient *pb.MockFilesystemPrivateServiceClient
		rService   *StorageReplicatorService
	)

	BeforeEach(func() {
		// Create fake client for testing
		clusterClient, err := dynamic.NewForConfig(k8sRestConfig)
		Expect(err).To(BeNil())

		cfg := config.NewDefaultConfig()
		resource := schema.GroupVersionResource{Group: "private.cloud.intel.com", Version: "v1alpha1", Resource: "storages"}
		factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, corev1.NamespaceAll, nil)
		informer := factory.ForResource(resource).Informer()
		mockClient = NewGRPCClient()
		rService = &StorageReplicatorService{
			syncTicker:       time.NewTicker(time.Duration(cfg.SchedulerInterval) * time.Second),
			Cfg:              cfg,
			storageAPIClient: mockClient,
			k8sclient:        clusterClient,
			informer:         informer,
		}
	})
	Context("NewStorageReplicatorService", func() {
		It("Should return error", func() {
			cfg := config.NewDefaultConfig()
			By("Setting gRPC address to be nil in the config")
			res1, err := NewStorageReplicatorService(ctx, cfg, k8sRestConfig)
			Expect(err).NotTo(BeNil())
			Expect(res1).To(BeNil())

			By("Setting k8sconfig to be nil")
			res2, err := NewStorageReplicatorService(ctx, cfg, nil)
			Expect(err).NotTo(BeNil())
			Expect(res2).To(BeNil())

			By("Setting invalid gRPC endpoint in cfg")
			cfg.IDCServiceConfig.StorageAPIGrpcEndpoint = "xyz"
			res3, err := NewStorageReplicatorService(ctx, cfg, k8sRestConfig)
			Expect(err).NotTo(BeNil())
			Expect(res3).To(BeNil())
		})
		It("Should succeed and, return NewStorageReplicatorService", func() {
			By("Setting valid config, with invalid grpc endpoint")
			cfg := config.NewDefaultConfig()
			cfg.IDCServiceConfig.StorageAPIGrpcEndpoint = "localhost"
			res, err := NewStorageReplicatorService(ctx, cfg, k8sRestConfig)
			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})
	})

	Context("When replicator recieves a new fs", func() {
		It("Should create a new storage in K8s", func() {
			Expect(rService).NotTo(BeNil())

			By("Getting a fs request")
			fsResp := NewRequest(false)
			Expect(fsResp).NotTo(BeNil())
			storageRef := &pc.Storage{}
			storageLookupKey := types.NamespacedName{Namespace: "123456789012", Name: "test"}

			By("Passing fs request to Replicator")
			resp := pb.NewMockFilesystemPrivateService_SearchFilesystemRequestsClient(gomock.NewController(GinkgoT()))
			// Set expectations on the mock
			mockClient.EXPECT().SearchFilesystemRequests(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			resp.EXPECT().Recv().Return(fsResp, nil).Times(1)
			resp.EXPECT().Recv().Return(nil, errors.New("error")).Times(1)
			latest := rService.Replicate(ctx, 0)
			Expect(latest).To(Equal(int64(1)))

			By("Waiting for fs to be created in K8s")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Trying to create fs that already exist")
			err := rService.processFilesystemCreate(ctx, fsResp.Filesystem)
			Expect(err).To(BeNil())

		}) //It
	})

	Context("When fs status is updated in k8s", func() {
		It("Should call handleResourceUpdate", func() {
			By("Getting a fs request")
			fsResp := NewRequest(false)
			Expect(fsResp).NotTo(BeNil())
			storageRef := &pc.Storage{}
			storageLookupKey := types.NamespacedName{Namespace: "123456789012", Name: "test"}

			By("Checking that fs exist in K8s")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Update storage status in K8s (simulate Storage Operator)")
			runningCondition := pc.StorageCondition{
				Type:   pc.StorageConditionRunning,
				Status: corev1.ConditionTrue,
			}
			storageRef.Status.Conditions = append(storageRef.Status.Conditions, runningCondition)
			Expect(k8sClient.Status().Update(ctx, storageRef)).Should(Succeed())

			By("Calling handResourceUpdate")
			// Set Mock behavior
			mockClient.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			newObj := getStorageObject(ctx, fsResp.Filesystem)
			oldObj := getStorageObject(ctx, fsResp.Filesystem)

			err := rService.handleResourceUpdate(ctx, oldObj, newObj)
			Expect(err).To(BeNil())

		}) //It
	}) //Context

	Context("When replicator recieves fs with deletionTimeStamp", func() {
		It("Should delete fs from K8s and call removeFinalizers", func() {
			By("Getting a fs request with DeletionTimeStamp")
			fsResp := NewRequest(true)
			storageRef := &pc.Storage{}
			storageLookupKey := types.NamespacedName{Namespace: "123456789012", Name: "test"}

			By("Passing fs request to Replicator")
			resp := pb.NewMockFilesystemPrivateService_SearchFilesystemRequestsClient(gomock.NewController(GinkgoT()))
			// Set expectations on the mock
			mockClient.EXPECT().SearchFilesystemRequests(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			resp.EXPECT().Recv().Return(fsResp, nil).Times(1)
			resp.EXPECT().Recv().Return(nil, errors.New("error")).Times(1)
			latest := rService.Replicate(ctx, 0)
			Expect(latest).To(Equal(int64(1)))

			By("Waiting for fs to be not found so as to confirm its deletion")
			Eventually(func(g Gomega) {
				g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, storageLookupKey, storageRef))).Should(BeTrue())
			}, timeout, interval).Should(Succeed())

			By("Trying to delete fs that does not exist")
			mockClient.EXPECT().RemoveFinalizer(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			err := rService.processFilesystemDelete(ctx, fsResp.Filesystem)
			Expect(err).To(BeNil())

		}) //It
	}) //Context

	Context("NewDefaultConfig", func() {
		It("Should return config", func() {
			By("calling the default constructor ")
			cfg := config.NewDefaultConfig()
			Expect(cfg).NotTo(BeNil())
		})
	})

	Context("When suppplied corresponding Error, the checking functions", func() {
		It("Should return True", func() {
			By("calling the AlreadyExistError function ")
			alreadyExist := &apierrors.StatusError{
				ErrStatus: metav1.Status{
					Reason: metav1.StatusReasonAlreadyExists,
				},
			}
			res := IsResourceAlreadyExistsError(alreadyExist)
			Expect(res).To(BeTrue())
			// Test false case
			err := errors.New("Error")
			res = IsResourceAlreadyExistsError(err)
			Expect(res).NotTo(BeTrue())

			By("Calling the IsResourceNotFoundError function ")
			notFound := &apierrors.StatusError{
				ErrStatus: metav1.Status{
					Reason: metav1.StatusReasonNotFound,
				},
			}
			res = IsResourceNotFoundError(notFound)
			Expect(res).To(BeTrue())
			// Test false case
			err = errors.New("Error")
			res = IsResourceNotFoundError(err)
			Expect(res).NotTo(BeTrue())

		})

	})

	Context("mapResPhaseFromk8sToPB", func() {
		It("Should return corresponding pb.FilesystemPhase", func() {
			//Provisioning
			out := mapResPhaseFromK8sToPB(pc.FilesystemPhaseProvisioning)
			Expect(out).To(Equal(pb.FilesystemPhase_FSProvisioning))
			//Ready
			out = mapResPhaseFromK8sToPB(pc.FilesystemPhaseReady)
			Expect(out).To(Equal(pb.FilesystemPhase_FSReady))
			//Failed
			out = mapResPhaseFromK8sToPB(pc.FilesystemPhaseFailed)
			Expect(out).To(Equal(pb.FilesystemPhase_FSFailed))
			//Terminating
			out = mapResPhaseFromK8sToPB(pc.FilesystemPhaseDeleting)
			Expect(out).To(Equal(pb.FilesystemPhase_FSDeleting))

		})
	})

	It("Should create a namespace if it does not exist", func() {
		By("Setting valid namespace")
		err := rService.createNamespaceIfNeeded(ctx, "111111111111")
		Expect(err).To(BeNil())
	})

	It("Should return error", func() {
		By("Supplying invalid namespace")
		err := rService.createNamespaceIfNeeded(ctx, "")
		Expect(err).NotTo(BeNil())
	})

})
