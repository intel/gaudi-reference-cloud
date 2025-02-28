// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"errors"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_replicator/pkg/config"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"

	"github.com/golang/mock/gomock"
	pc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

func NewVastRequest(delete bool) *pb.FilesystemRequestResponse {
	var version = new(string)
	*version = "v2"
	var dTime *timestamppb.Timestamp
	if delete {
		dTime = timestamppb.Now()
	}
	fsResp := &pb.FilesystemRequestResponse{
		Filesystem: &pb.FilesystemPrivate{
			Metadata: &pb.FilesystemMetadataPrivate{
				CloudAccountId:    "123456789012",
				Name:              "test1",
				ResourceId:        "6787226a-2a55-4d6f-bae9-fa2a2ca2450a",
				ResourceVersion:   "1",
				Description:       "Sample Filesystem",
				Labels:            map[string]string{"key": "value"},
				CreationTimestamp: &timestamppb.Timestamp{Seconds: 1637077200, Nanos: 0},
				DeletionTimestamp: dTime,
			}, //Meta
			Spec: &pb.FilesystemSpecPrivate{
				AvailabilityZone: "az1",
				Request: &pb.FilesystemCapacity{
					Storage: "2000000000",
				},
				MountProtocol: 0,
				Encrypted:     false,
				Scheduler: &pb.FilesystemSchedule{
					FilesystemName: "test",
					Cluster: &pb.AssignedCluster{
						ClusterName:    "1",
						ClusterAddr:    "1",
						ClusterUUID:    "1",
						ClusterVersion: version,
					},
					Namespace: &pb.AssignedNamespace{
						Name:            "123456789012",
						CredentialsPath: "/path/to/secret",
					},
				},

				SecurityGroup: &pb.VolumeSecurityGroup{
					NetworkFilterAllow: []*pb.VolumeNetworkGroup{
						{
							Subnet:       "127.0.0.1",
							PrefixLength: 24,
							Gateway:      "8.8.8.8",
						},
					},
				},
				StorageClass: pb.FilesystemStorageClass_GeneralPurposeStd,
			}, //Spec
			Status: &pb.FilesystemStatusPrivate{
				Phase:   pb.FilesystemPhase_FSProvisioning,
				Message: "Filesystem is being provisioned",
				// Add relevant fields from your status
			}, // Status
		}, //FSPrivate
	}
	return fsResp
}

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
		vastResource := schema.GroupVersionResource{Group: "private.cloud.intel.com", Version: "v1alpha1", Resource: "vaststorages"}
		vastFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, corev1.NamespaceAll, nil)
		vastInformer := vastFactory.ForResource(vastResource).Informer()
		mockClient = NewGRPCClient() // Initialize the mockClient here
		rService = &StorageReplicatorService{
			syncTicker:       time.NewTicker(time.Duration(cfg.SchedulerInterval) * time.Second),
			Cfg:              cfg,
			storageAPIClient: mockClient,
			k8sclient:        clusterClient,
			informerVast:     vastInformer,
		}
	})

	Context("When replicator receives a new VAST fs", func() {
		It("Should create a new VAST storage in K8s", func() {
			Expect(rService).NotTo(BeNil())

			By("Getting a VAST fs request")
			fsResp := NewVastRequest(false)
			Expect(fsResp).NotTo(BeNil())
			storageRef := &pc.VastStorage{}
			storageLookupKey := types.NamespacedName{Namespace: "123456789012", Name: "test1"}

			By("Passing VAST fs request to Replicator")
			resp := pb.NewMockFilesystemPrivateService_SearchFilesystemRequestsClient(gomock.NewController(GinkgoT()))

			resp.EXPECT().Recv().Return(fsResp, nil).Times(2)
			resp.EXPECT().Recv().Return(nil, io.EOF).Times(1)
			mockClient.EXPECT().SearchFilesystemRequests(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()

			latest := rService.Replicate(ctx, 0)
			Expect(latest).To(Equal(int64(1)))

			By("Waiting for VAST fs to be created in K8s")
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, storageLookupKey, storageRef)
				Expect(err).To(BeNil())
			}, timeout, interval).Should(Succeed())

			By("Trying to create VAST fs that already exist")
			err := rService.processFilesystemCreate(ctx, fsResp.Filesystem)
			Expect(err).To(BeNil())

		})
	})

	Context("When VAST fs status is updated in k8s", func() {
		It("Should call handleVASTResourceUpdate", func() {
			By("Getting a VAST fs request")
			fsResp := NewVastRequest(false)
			Expect(fsResp).NotTo(BeNil())
			storageRef := &pc.VastStorage{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "123456789012",
				},
				Spec: pc.VastStorageSpec{
					Networks: pc.Networks{
						SecurityGroups: pc.SecurityGroups{
							IPFilters: []pc.IPFilter{
								{
									Start: "192.168.1.1",
									End:   "192.168.1.255",
								},
							},
						},
					},
				},
			}
			storageLookupKey := types.NamespacedName{Namespace: "123456789012", Name: "test1"}

			By("Checking that VAST fs exist in K8s")
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, storageLookupKey, storageRef)
				Expect(err).To(BeNil())
			}, timeout, interval).Should(Succeed())

			By("Update VAST storage status in K8s (simulate Storage Operator)")
			runningCondition := pc.StorageCondition{
				Type:   pc.StorageConditionRunning,
				Status: corev1.ConditionTrue,
			}
			storageRef.Status.Conditions = append(storageRef.Status.Conditions, runningCondition)
			Expect(k8sClient.Status().Update(ctx, storageRef)).Should(Succeed())

			By("Calling handleVASTResourceUpdate")
			// Set Mock behavior
			mockClient.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			newObj := getStorageObject(ctx, fsResp.Filesystem)
			oldObj := getStorageObject(ctx, fsResp.Filesystem)

			err := rService.handleVASTResourceUpdate(ctx, oldObj, newObj)
			Expect(err).To(BeNil())
		})
	})

	Context("When replicator receives a new VAST fs with updateTimestamp", func() {
		It("Should update the VAST storage in K8s", func() {
			Expect(rService).NotTo(BeNil())

			By("Getting a VAST fs request with updateTimestamp")
			fsResp := NewVastRequest(false)
			fsResp.Filesystem.Metadata.UpdateTimestamp = timestamppb.New(time.Now())
			Expect(fsResp).NotTo(BeNil())
			storageRef := &pc.VastStorage{}
			storageLookupKey := types.NamespacedName{Namespace: "123456789012", Name: "test1"}

			By("Passing VAST fs request to Replicator")
			resp := pb.NewMockFilesystemPrivateService_SearchFilesystemRequestsClient(gomock.NewController(GinkgoT()))
			// Set expectations on the mock
			resp.EXPECT().Recv().Return(fsResp, nil).Times(2)
			resp.EXPECT().Recv().Return(nil, io.EOF).Times(1)
			mockClient.EXPECT().SearchFilesystemRequests(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)

			latest := rService.Replicate(ctx, 0)
			Expect(latest).To(Equal(int64(1)))

			By("Waiting for VAST fs to be created in K8s")
			Eventually(func(g Gomega) {
				err := k8sClient.Get(ctx, storageLookupKey, storageRef)
				Expect(err).To(BeNil())
			}, timeout, interval).Should(Succeed())

			By("Trying to update VAST fs that already exist")
			err := rService.processFilesystemUpdate(ctx, fsResp.Filesystem)
			Expect(err).To(BeNil())
		})
	})

	Context("When replicator receives VAST fs with deletionTimeStamp", func() {
		It("Should delete VAST fs from K8s and call removeFinalizers", func() {
			By("Getting a VAST fs request with DeletionTimeStamp")
			fsResp := NewVastRequest(true)
			storageRef := &pc.VastStorage{}
			storageLookupKey := types.NamespacedName{Namespace: "123456789012", Name: "test1"}

			By("Passing VAST fs request to Replicator")
			resp := pb.NewMockFilesystemPrivateService_SearchFilesystemRequestsClient(gomock.NewController(GinkgoT()))
			resp.EXPECT().Recv().Return(fsResp, nil).Times(2)
			resp.EXPECT().Recv().Return(nil, errors.New("error")).Times(1)
			mockClient.EXPECT().RemoveFinalizer(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
			mockClient.EXPECT().SearchFilesystemRequests(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()

			latest := rService.Replicate(ctx, 0)
			Expect(latest).To(Equal(int64(1)))

			By("Waiting for VAST fs to be not found so as to confirm its deletion")
			Eventually(func(g Gomega) {
				g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, storageLookupKey, storageRef))).Should(BeTrue())
			}, timeout, interval).Should(Succeed())

			By("Trying to delete VAST fs that does not exist")
			mockClient.EXPECT().RemoveFinalizer(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
			err := rService.processFilesystemDelete(ctx, fsResp.Filesystem)
			Expect(err).To(BeNil())
		})
	})

})

func strPtr(s string) *string {
	return &s
}
