package controller

import (
	"context"
	"errors"
	"time"

	"github.com/golang/mock/gomock"
	pc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/bucket_replicator/pkg/config"

	//corev1 "k8s.io/api/core/v1"
	k8sv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BucketReplicator", func() {

	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	var (
		mockClient *pb.MockObjectStorageServicePrivateClient
		brService  *BucketReplicatorService
	)

	BeforeEach(func() {
		// Create fake client for testing
		clusterClient, err := dynamic.NewForConfig(k8sRestConfig)
		Expect(err).To(BeNil())

		cfg := config.NewDefaultConfig()
		resource := schema.GroupVersionResource{Group: "private.cloud.intel.com", Version: "v1alpha1", Resource: "objectstores"}
		factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, k8sv1.NamespaceAll, nil)
		informer := factory.ForResource(resource).Informer()
		mockClient = NewGRPCClient()
		brService = &BucketReplicatorService{
			syncTicker:      time.NewTicker(time.Duration(cfg.SchedulerInterval) * time.Second),
			Cfg:             cfg,
			bucketAPIClient: mockClient,
			k8sclient:       clusterClient,
			informer:        informer,
		}
		Expect(brService).NotTo(BeNil())
	})

	Context("NewBucketReplicatorService", func() {
		It("Should return error", func() {
			cfg := config.NewDefaultConfig()
			By("Setting gRPC address to be nil in the config")
			res1, err := NewBucketReplicatorService(ctx, cfg, k8sRestConfig)
			Expect(err).NotTo(BeNil())
			Expect(res1).To(BeNil())

			By("Setting k8sconfig to be nil")
			res2, err := NewBucketReplicatorService(ctx, cfg, nil)
			Expect(err).NotTo(BeNil())
			Expect(res2).To(BeNil())

			By("Setting invalid gRPC endpoint in cfg")
			cfg.IDCServiceConfig.StorageAPIGrpcEndpoint = "xyz"
			res3, err := NewBucketReplicatorService(ctx, cfg, k8sRestConfig)
			Expect(err).NotTo(BeNil())
			Expect(res3).To(BeNil())
		})
		It("Should succeed and, return NewBucketReplicatorService", func() {
			By("Setting valid config, with invalid grpc endpoint")
			cfg := config.NewDefaultConfig()
			cfg.IDCServiceConfig.StorageAPIGrpcEndpoint = "localhost"
			res, err := NewBucketReplicatorService(ctx, cfg, k8sRestConfig)
			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})
	})

	Context("When replicator recieves a new bucket", func() {
		It("Should create a new object in K8s", func() {

			By("Getting a bucket request")
			bkResp := NewRequest(false)
			Expect(bkResp).NotTo(BeNil())
			objectRef := &pc.ObjectStore{}
			objectLookupKey := types.NamespacedName{Namespace: "123456789012", Name: "test"}

			By("Passing bucket request to Replicator")
			resp := pb.NewMockObjectStorageServicePrivate_SearchBucketPrivateClient(gomock.NewController(GinkgoT()))
			// Set expectations on the mock
			mockClient.EXPECT().SearchBucketPrivate(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			resp.EXPECT().Recv().Return(bkResp, nil).Times(1)
			resp.EXPECT().Recv().Return(nil, errors.New("error")).Times(1)
			latest := brService.Replicate(ctx, 0)
			Expect(latest).To(Equal(int64(1)))

			By("Waiting for bucket to be created in K8s")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, objectLookupKey, objectRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Trying to create bucket that already exist")
			err := brService.processBucketCreate(ctx, bkResp.Bucket)
			Expect(err).To(BeNil())

		}) //It
	})

	Context("When bucket status is updated in k8s", func() {
		It("Should create a new object in K8s", func() {

			By("Getting a bucket request")
			bkResp := NewRequest(false)
			Expect(bkResp).NotTo(BeNil())
			objectRef := &pc.ObjectStore{}
			objectLookupKey := types.NamespacedName{Namespace: "123456789012", Name: "test"}

			By("Checking that bucket exist in K8s")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, objectLookupKey, objectRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Update object status in K8s")
			runningCondition := pc.ObjectStoreCondition{
				Type:   pc.ObjectStoreConditionRunning,
				Status: k8sv1.ConditionTrue,
			}
			objectRef.Status.Conditions = append(objectRef.Status.Conditions, runningCondition)
			Expect(k8sClient.Status().Update(ctx, objectRef)).Should(Succeed())

			mockClient.EXPECT().UpdateBucketStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)

			newObj := getBucketObject(bkResp.Bucket)
			oldObj := getBucketObject(bkResp.Bucket)

			err := brService.handleResourceUpdate(ctx, oldObj, newObj)
			Expect(err).To(BeNil())

		}) //It
	})

	Context("When bucket status is deleted in k8s", func() {
		It("Should create a new object in K8s", func() {

			By("Getting a bucket request")
			bkResp := NewRequest(true)
			Expect(bkResp).NotTo(BeNil())
			objectRef := &pc.ObjectStore{}
			objectLookupKey := types.NamespacedName{Namespace: "123456789012", Name: "test"}

			By("Passing bucket request to Replicator")
			resp := pb.NewMockObjectStorageServicePrivate_SearchBucketPrivateClient(gomock.NewController(GinkgoT()))
			// Set expectations on the mock
			mockClient.EXPECT().RemoveBucketFinalizer(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
			mockClient.EXPECT().SearchBucketPrivate(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			resp.EXPECT().Recv().Return(bkResp, nil).Times(1)
			resp.EXPECT().Recv().Return(nil, errors.New("error")).Times(1)
			latest := brService.Replicate(ctx, 0)
			Expect(latest).To(Equal(int64(1)))

			By("Waiting for bucket to be not found so as to confirm its deletion")
			Eventually(func(g Gomega) {
				g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, objectLookupKey, objectRef))).Should(BeTrue())
			}, timeout, interval).Should(Succeed())

			By("Trying to create bucket that already exist")
			err := brService.processBucketDelete(ctx, bkResp.Bucket)
			Expect(err).To(BeNil())

		}) //It
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
		It("Should return corresponding pb.BucketPhase", func() {
			//Provisioning
			out := mapResPhaseFromK8sToPB(pc.ObjectStorePhasePhaseProvisioning)
			Expect(out).To(Equal(pb.BucketPhase_BucketProvisioning))
			//Ready
			out = mapResPhaseFromK8sToPB(pc.ObjectStorePhasePhaseReady)
			Expect(out).To(Equal(pb.BucketPhase_BucketReady))
			//Failed
			out = mapResPhaseFromK8sToPB(pc.ObjectStorePhasePhaseFailed)
			Expect(out).To(Equal(pb.BucketPhase_BucketFailed))
			//Terminating
			out = mapResPhaseFromK8sToPB(pc.ObjectStorePhasePhaseTerminating)
			Expect(out).To(Equal(pb.BucketPhase_BucketDeleting))

		})
	})

	It("Should create a namespace if it does not exist", func() {
		By("Setting valid namespace")
		err := brService.createNamespaceIfNeeded(ctx, "111111111111")
		Expect(err).To(BeNil())
	})

	It("Should return error", func() {
		By("Supplying invalid namespace")
		err := brService.createNamespaceIfNeeded(ctx, "")
		Expect(err).NotTo(BeNil())
	})

})
