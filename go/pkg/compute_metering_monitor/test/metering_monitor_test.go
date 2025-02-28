// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_metering_monitor/metering_monitor"
	privatecloud "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/controllers"
	instanceoptest "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/test"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
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

func NewInstance(namespace string, instanceName string) *privatecloudv1alpha1.Instance {
	return instanceoptest.NewInstance(
		namespace,
		instanceName,
		"az1",
		"region1",
		instanceoptest.NewInstanceTypeSpec("tiny"),
		instanceoptest.NewSshPublicKeySpecs("ssh-rsa xxx"),
		instanceoptest.NewInterfaceSpecs(),
	)
}

func NewMockMeteringServiceClient() (pb.MeteringServiceClient, *sync.Map) {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	mockController := gomock.NewController(GinkgoT())
	meteringClient := pb.NewMockMeteringServiceClient(mockController)

	// Mock Create stores the instance record when instance enters ready state or is deleted.
	createMeteringRecorder := &sync.Map{}
	mockCreate := func(ctx context.Context, req *pb.UsageCreate, opts ...grpc.CallOption) (*emptypb.Empty, error) {
		createMeteringRecorder.Store(req.ResourceId, req)
		return &emptypb.Empty{}, nil
	}

	meteringClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(mockCreate).AnyTimes()
	return meteringClient, createMeteringRecorder
}

var _ = Describe("Compute Metering Monitor", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	var createMeteringRecorder *sync.Map
	var managerStoppable *stoppable.Stoppable

	BeforeEach(func() {
		By("Creating manager")
		k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
			Scheme: scheme,
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
		})
		Expect(err).ToNot(HaveOccurred())

		By("Creating mock metering Client")
		var meteringClient pb.MeteringServiceClient
		meteringClient, createMeteringRecorder = NewMockMeteringServiceClient()

		By("Creating metering monitor")
		_, err = metering_monitor.NewMeteringMonitor(ctx, k8sManager, meteringClient, maxRequeueTimeMillliseconds)
		Expect(err).Should(Succeed())

		By("Starting manager")
		managerStoppable = stoppable.New(k8sManager.Start)
		managerStoppable.Start(ctx)
	})

	AfterEach(func() {
		By("Stopping manager")
		Expect(managerStoppable.Stop(ctx)).Should(Succeed())
		By("Manager stopped")
	})

	It("When an Instance is created, compute metering montior should create a record in metering db", func() {
		namespace := uuid.NewString()
		instanceName := uuid.NewString()
		instance := NewInstance(namespace, instanceName)
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		instanceRef := &privatecloudv1alpha1.Instance{}
		nsObject := instanceoptest.NewNamespace(namespace)

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("Creating Instance successfully")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for instance to be created in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Updating instance status in K8s (simulate Instance Operator)")
		startupCompleteCondition := privatecloudv1alpha1.InstanceCondition{
			Type:   privatecloudv1alpha1.InstanceConditionStartupComplete,
			Status: k8sv1.ConditionTrue,
		}
		instanceRef.Status.Conditions = append(instanceRef.Status.Conditions, startupCompleteCondition)
		Expect(k8sClient.Status().Update(ctx, instanceRef)).Should(Succeed())

		By("Confirming StartupComplete condition to be true")
		startupCompleteCond := metering_monitor.FindStatusCondition(instanceRef.Status.Conditions, privatecloudv1alpha1.InstanceConditionStartupComplete)
		Expect(startupCompleteCond).ShouldNot(BeNil())
		Expect(startupCompleteCond.Status).Should(Equal(k8sv1.ConditionTrue))

		By("Waiting for metering record to be recorded")
		Eventually(func(g Gomega) {
			value, ok := createMeteringRecorder.Load(instanceName)
			g.Expect(ok).Should(BeTrue())
			req, ok := value.(*pb.UsageCreate)
			g.Expect(ok).Should(BeTrue())
			g.Expect(req.Properties["firstReadyTimestamp"]).Should(Equal(startupCompleteCond.LastTransitionTime.Format(time.RFC3339)))
			g.Expect(req.Properties["region"]).Should(Equal("region1"))
			g.Expect(req.Properties["deleted"]).Should(Equal("false"))
		}, timeout, interval).Should(Succeed())

		By("Waiting for instance to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Delete(ctx, instanceRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())
	})

	It("When metering montior receives an instance with a deletionTimestamp, it should create a record in metering db before deleting the instance", func() {
		namespace := uuid.NewString()
		instanceName := uuid.NewString()
		instance := NewInstance(namespace, instanceName)
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		instanceRef := &privatecloudv1alpha1.Instance{}
		nsObject := instanceoptest.NewNamespace(namespace)
		finalizers := []string{privatecloud.InstanceMeteringMonitorFinalizer}

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("Creating Instance successfully")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for instance to be created in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Updating instance status in K8s (simulate Instance Operator)")
		startupCompleteCondition := privatecloudv1alpha1.InstanceCondition{
			Type:   privatecloudv1alpha1.InstanceConditionStartupComplete,
			Status: k8sv1.ConditionTrue,
		}
		instanceRef.Status.Conditions = append(instanceRef.Status.Conditions, startupCompleteCondition)
		Expect(k8sClient.Status().Update(ctx, instanceRef)).Should(Succeed())

		By("Confirming StartupComplete condition to be true")
		startupCompleteCond := metering_monitor.FindStatusCondition(instanceRef.Status.Conditions, privatecloudv1alpha1.InstanceConditionStartupComplete)
		Expect(startupCompleteCond).ShouldNot(BeNil())
		Expect(startupCompleteCond.Status).Should(Equal(k8sv1.ConditionTrue))

		By("Sending instance with deletionTimestamp and metering monitor finalizer (simulate Instance Operator) to Metering Monitor")
		now := metav1.Now()
		instance.ObjectMeta.DeletionTimestamp = &now
		instanceRef.SetFinalizers(finalizers)
		err := k8sClient.Update(ctx, instanceRef)
		Expect(err).Should(Succeed(), fmt.Errorf("error occured while updating instance metadata: %w", err))

		By("Waiting for instance to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Delete(ctx, instanceRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for instance to be not found so as to confirm its deletion")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, instanceLookupKey, instanceRef))).Should(BeTrue())
		}, timeout, interval).Should(Succeed())

		By("Creating a Record in the metering db with deleted property set to true")
		Eventually(func(g Gomega) {
			value, ok := createMeteringRecorder.Load(instanceName)
			g.Expect(ok).Should(BeTrue())
			req, ok := value.(*pb.UsageCreate)
			g.Expect(ok).Should(BeTrue())
			g.Expect(req.Properties["firstReadyTimestamp"]).Should(Equal(startupCompleteCond.LastTransitionTime.Format(time.RFC3339)))
			g.Expect(req.Properties["deleted"]).Should(Equal("true"))
		}, timeout, interval).Should(Succeed())
	})

	It("When an Instance is created, compute metering montior should create at least three records in metering db when requeue interval is 500ms", func() {
		namespace := uuid.NewString()
		instanceName := uuid.NewString()
		instance := NewInstance(namespace, instanceName)
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		instanceRef := &privatecloudv1alpha1.Instance{}
		nsObject := instanceoptest.NewNamespace(namespace)

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("Creating Instance successfully")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for instance to be created in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Updating instance status in K8s (simulate Instance Operator)")
		startupCompleteCondition := privatecloudv1alpha1.InstanceCondition{
			Type:   privatecloudv1alpha1.InstanceConditionStartupComplete,
			Status: k8sv1.ConditionTrue,
		}
		instanceRef.Status.Conditions = append(instanceRef.Status.Conditions, startupCompleteCondition)
		Expect(k8sClient.Status().Update(ctx, instanceRef)).Should(Succeed())

		By("Waiting for at least three metering records to be recorded within four seconds of timeout")
		var numberOfRecords int
		Eventually(func(g Gomega) {
			createMeteringRecorder.Range(func(key any, value any) bool {
				numberOfRecords += 1
				return true
			})
			g.Expect(numberOfRecords).Should(BeNumerically(">=", 3))
		}, "4s", interval).Should(Succeed())
	})

	It("When an Instance is created, compute metering montior should create a record in metering db with correct instanceGroupSize", func() {
		namespace := uuid.NewString()
		nsObject := instanceoptest.NewNamespace(namespace)

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		var instanceGroupSize int = 2
		for i := 1; i <= instanceGroupSize; i++ {
			instanceName := uuid.NewString()
			instance := NewInstance(namespace, instanceName)
			instance.Spec.InstanceGroupSize = int32(instanceGroupSize)
			instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
			instanceRef := &privatecloudv1alpha1.Instance{}

			By("Creating Instance successfully")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Waiting for instance to be created in K8s")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Updating instance status in K8s (simulate Instance Operator)")
			startupCompleteCondition := privatecloudv1alpha1.InstanceCondition{
				Type:   privatecloudv1alpha1.InstanceConditionStartupComplete,
				Status: k8sv1.ConditionTrue,
			}
			instanceRef.Status.Conditions = append(instanceRef.Status.Conditions, startupCompleteCondition)
			Expect(k8sClient.Status().Update(ctx, instanceRef)).Should(Succeed())

			By("Confirming StartupComplete condition to be true")
			startupCompleteCond := metering_monitor.FindStatusCondition(instanceRef.Status.Conditions, privatecloudv1alpha1.InstanceConditionStartupComplete)
			Expect(startupCompleteCond).ShouldNot(BeNil())
			Expect(startupCompleteCond.Status).Should(Equal(k8sv1.ConditionTrue))

			By("Waiting for metering record to be recorded")
			Eventually(func(g Gomega) {
				value, ok := createMeteringRecorder.Load(instanceName)
				g.Expect(ok).Should(BeTrue())
				req, ok := value.(*pb.UsageCreate)
				g.Expect(ok).Should(BeTrue())
				g.Expect(req.Properties["firstReadyTimestamp"]).Should(Equal(startupCompleteCond.LastTransitionTime.Format(time.RFC3339)))
				g.Expect(req.Properties["region"]).Should(Equal("region1"))
				g.Expect(req.Properties["deleted"]).Should(Equal("false"))
				g.Expect(req.Properties["instanceGroupSize"]).Should(Equal(strconv.FormatInt(int64(instanceGroupSize), 10)))
			}, timeout, interval).Should(Succeed())

			By("Waiting for instance to be deleted from K8s")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Delete(ctx, instanceRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
		}
	})

})
