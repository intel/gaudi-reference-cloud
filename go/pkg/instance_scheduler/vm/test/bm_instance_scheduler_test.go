// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	bmenroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/server"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	toolsk8s "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
)

const (
	instanceType   = "test-baremetal"
	instanceTypeSC = "test-baremetal-sc"
)

type BmSchedulerOptions struct {
	Namespace string
	HostName  string
}

type BmSchedulerTestEnv struct {
	SchedulerServiceClient pb.InstanceSchedulingServiceClient
	SchedulingServer       *server.SchedulingServer
	SchedulerOpts          BmSchedulerOptions
	Ctx                    context.Context
	Cancel                 context.CancelFunc
}

var _ = Describe("BM Instance Scheduler", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("BM Instance Scheduler")
	_ = log

	It("Schedule instance should succeed for base deployment(No HBM, No GPU, No group)", func() {
		namespace := "test-ns-" + uuid.NewString()
		hostName := "test-host-1"
		instanceName := "test-instance-1"
		opts := BmSchedulerOptions{
			Namespace: namespace,
			HostName:  hostName}
		bmSchedulerTestEnv := NewBMSchedulerTestEnv(
			opts,
		)
		defer bmSchedulerTestEnv.Stop()
		bmSchedulerTestEnv.Start()
		req := newBmScheduleRequest(instanceName, instanceType)
		resp, err := bmSchedulerTestEnv.scheduleOneBmInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("testresp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal(namespace))
		Expect(resp.NodeId).Should(Equal(hostName))
	})

	It("Schedule instance should succeed for matching computeNodePools", func() {
		namespace := "test-ns-" + uuid.NewString()
		// test artifacts for computeNodePools
		hostName := "test-host-cnp-1"
		instanceName := "test-instance-cnp-1"
		opts := BmSchedulerOptions{
			Namespace: namespace,
			HostName:  hostName}
		bmSchedulerTestEnv := NewBMSchedulerTestEnv(
			opts,
		)
		defer bmSchedulerTestEnv.Stop()
		bmSchedulerTestEnv.Start()
		req := newBmScheduleRequest(instanceName, instanceType)
		req.Instances[0].Spec.ComputeNodePools = []string{"general"}
		resp, err := bmSchedulerTestEnv.scheduleOneBmInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("testresp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal(namespace))
		Expect(resp.NodeId).Should(Equal(hostName))
		Expect(resp.ComputeNodePools[0]).Should(Equal("general"))
	})

	It("Schedule instance should not succeed for wrong InstanceType", func() {
		namespace := "test-ns-" + uuid.NewString()
		hostName := "test-host-2"
		instanceName := "test-instance-2"
		opts := BmSchedulerOptions{
			Namespace: namespace,
			HostName:  hostName}
		bmSchedulerTestEnv := NewBMSchedulerTestEnv(
			opts,
		)
		defer bmSchedulerTestEnv.Stop()
		bmSchedulerTestEnv.Start()
		req := newBmScheduleRequest(instanceName, "bm-wrong-instancetype")
		resp, err := bmSchedulerTestEnv.scheduleOneBmInstance(ctx, req)
		Expect(err).To(HaveOccurred())
		log.Info("testresp", "resp", resp)
	})

	It("Schedule instance should not succeed when computeNodePools do not match", func() {
		namespace := "test-ns-" + uuid.NewString()
		// test artifacts for computeNodePools
		hostName := "test-host-cnp-2"
		instanceName := "test-instance-cnp-2"
		opts := BmSchedulerOptions{
			Namespace: namespace,
			HostName:  hostName}
		bmSchedulerTestEnv := NewBMSchedulerTestEnv(
			opts,
		)
		defer bmSchedulerTestEnv.Stop()
		bmSchedulerTestEnv.Start()
		req := newBmScheduleRequest(instanceName, instanceType)
		req.Instances[0].Spec.ComputeNodePools = []string{"general-not-match"}
		_, err := bmSchedulerTestEnv.scheduleOneBmInstance(ctx, req)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(MatchRegexp(".*" + server.InsufficientResourcesError))
	})

	It("Schedule single instances into cluster group if all single hosts are unavailable", func() {
		namespace := "test-ns-" + uuid.NewString()
		hostName := "single-host"
		instanceName := "test-instance"
		opts := BmSchedulerOptions{
			Namespace: namespace,
			HostName:  hostName,
		}
		bmSchedulerTestEnv := NewBMSchedulerTestEnv(
			opts,
		)
		defer bmSchedulerTestEnv.Stop()
		bmSchedulerTestEnv.Start()

		By("Creating a available cluster group with default network mode")
		hostGroupDefault := createBmHostGroup(ctx, namespace, "", 2)
		hostGroupVVX1 := createBmHostGroup(ctx, namespace, "vvx-1", 2,
			bmenroll.NetworkModeLabel, bmenroll.NetworkModeVVXStandalone)
		hostGroupVVX2 := createBmHostGroup(ctx, namespace, "vvx-2", 4,
			bmenroll.NetworkModeLabel, bmenroll.NetworkModeVVXStandalone)

		By("Scheduling an single instance")
		req := newBmScheduleRequest(instanceName, instanceType)
		resp, err := bmSchedulerTestEnv.scheduleOneBmInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("testresp", "resp", resp)

		By("Expecting to select a host that is not in a cluster group")
		Expect(resp.NetworkMode).Should(BeEmpty())
		Expect(resp.ClusterId).Should(Equal(namespace))
		Expect(resp.NodeId).Should(Equal(hostName))

		By("Having no single hosts available")
		for i := 0; i < 4; i++ {
			By("Scheduling a single instance")
			req = newBmScheduleRequest(instanceName, instanceType)
			resp, err = bmSchedulerTestEnv.scheduleOneBmInstance(ctx, req)
			Expect(err).Should(Succeed())
			log.Info("testresp", "resp", resp)

			By("Expecting to select hosts in a cluster group")
			expectedHosts := append(hostGroupDefault, hostGroupVVX1...)
			Expect(resp.NetworkMode).Should(BeElementOf("", bmenroll.NetworkModeVVXStandalone))
			Expect(resp.GroupId).Should(BeElementOf("", "vvx-1"))
			Expect(resp.ClusterId).Should(Equal(namespace))
			Expect(resp.NodeId).Should(EqualAnyHostNames(expectedHosts))
		}

		for i := 0; i < 4; i++ {
			By("Scheduling a single instance")
			req = newBmScheduleRequest(instanceName, instanceType)
			resp, err = bmSchedulerTestEnv.scheduleOneBmInstance(ctx, req)
			Expect(err).Should(Succeed())
			log.Info("testresp", "resp", resp)

			By("Expecting to select hosts in a cluster group")
			expectedHosts := append(hostGroupDefault, hostGroupVVX2...)
			Expect(resp.NetworkMode).Should(Equal(bmenroll.NetworkModeVVXStandalone))
			Expect(resp.GroupId).Should(Equal("vvx-2"))
			Expect(resp.ClusterId).Should(Equal(namespace))
			Expect(resp.NodeId).Should(EqualAnyHostNames(expectedHosts))
		}

		By("Having no hosts available")

		By("Scheduling a single instance")
		req = newBmScheduleRequest(instanceName, instanceType)
		resp, err = bmSchedulerTestEnv.scheduleOneBmInstance(ctx, req)

		By("Expecting an error due to insufficient capacity")
		Expect(err).Should(HaveOccurred())
		Expect(resp).Should(BeNil())
	})

	Describe("Scheduling a group of instances", Serial, func() {
		const (
			hostNamespace     = "test-namespace"
			instanceGroupName = "test-instance-group"

			clusterGroupDefault = ""
			clusterGroupVVX     = "vvx-1"

			clusterGroupVVV1 = "vvv-1"
			clusterGroupVVV2 = "vvv-2"

			clusterGroupXBX1 = "xbx-1"
			clusterGroupXBX2 = "xbx-2"
			clusterGroupXBX3 = "xbx-3"
			clusterGroupXBX4 = "xbx-4"
		)
		var (
			bmSchedulerTestEnv *BmSchedulerTestEnv
		)

		BeforeEach(func() {
			bmSchedulerTestEnv = NewBMSchedulerTestEnv(BmSchedulerOptions{})
			bmSchedulerTestEnv.Start()
			DeferCleanup(bmSchedulerTestEnv.Stop)
		})

		Context("Schedule for a new instance group", func() {

			It("Suggest a single cluster group using most allocated strategy", func() {
				By("Creating a cluster group with default network")
				hostGroupDefault := createBmHostGroup(ctx, hostNamespace, clusterGroupDefault, 2)

				By("Creating a cluster group with VVX network")
				hostGroupVVX := createBmHostGroup(ctx, hostNamespace, clusterGroupVVX, 4,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeVVXStandalone)

				By("Scheduling an instance group of size 2")
				req := newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 2)
				results, err := bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting hosts in cluster group with the least capacity")
				Expect(results).Should(HaveLen(2))
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(BeEmpty())
					Expect(r.GroupId).Should(Equal(clusterGroupDefault))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupDefault))
				}

				By("Creating a cluster group with VVV network")
				hostGroupVVV1 := createBmHostGroup(ctx, hostNamespace, clusterGroupVVV1, 2,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeVVV)

				By("Scheduling 3 more instance group of size 2")
				for i := 0; i < 3; i++ {
					req = newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 2)
					results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
					Expect(err).Should(Succeed())
					log.Info("scheduler response", "results", results)

					By("Expecting hosts in cluster group with default or VVV network")
					expectedHosts := hostGroupVVX
					expectedHosts = append(expectedHosts, hostGroupVVV1...)
					Expect(results).Should(HaveLen(2))
					Expect(results).Should(HaveOneGroupID())
					for _, r := range results {
						Expect(r.NetworkMode).Should(BeElementOf(bmenroll.NetworkModeVVXStandalone, bmenroll.NetworkModeVVV))
						Expect(r.GroupId).Should(BeElementOf(clusterGroupVVX, clusterGroupVVV1))
						Expect(r.ClusterId).Should(Equal(hostNamespace))
						Expect(r.NodeId).Should(EqualAnyHostNames(expectedHosts))
					}
				}

				By("Scheduling another instance group of size 2")
				req = newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 2)
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				log.Info("scheduler response", "results", results)

				By("Expecting an error due to insufficient capacity")
				Expect(err).Should(HaveOccurred())
				Expect(results).Should(HaveLen(0))
			})

			It("Suggest the optimized cluster groups with BGP network (exactly fit)", func() {
				By("Creating cluster groups with BGP network")
				hostGroupXBX1 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX1, 4,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)
				hostGroupXBX2 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX2, 4,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)
				hostGroupXBX3 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX3, 8,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)
				hostGroupXBX4 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX4, 16,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)

				By("Creating other cluster groups with non-BGP network")
				createBmHostGroup(ctx, hostNamespace, clusterGroupDefault, 16)
				createBmHostGroup(ctx, hostNamespace, clusterGroupVVX, 16,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeVVXStandalone)
				createBmHostGroup(ctx, hostNamespace, clusterGroupVVV1, 16,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeVVV)

				By("Scheduling an instance group of size 16")
				req := newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 16)
				results, err := bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting hosts in a single cluster group to be selected")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal(clusterGroupXBX4))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupXBX4))
				}

				By("Scheduling an instance group of size 16")
				req = newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 16)
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting hosts in multiple cluster groups with BGP network to be selected")
				expectedHosts := hostGroupXBX1
				expectedHosts = append(expectedHosts, hostGroupXBX2...)
				expectedHosts = append(expectedHosts, hostGroupXBX3...)
				Expect(results).ShouldNot(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(BeElementOf(clusterGroupXBX1, clusterGroupXBX2, clusterGroupXBX3))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(expectedHosts))
				}

				By("Scheduling an instance group of size 16")
				req = newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 16)
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				log.Info("scheduler response", "results", results)

				By("Expecting an error due to insufficient capacity")
				Expect(err).Should(HaveOccurred())
				Expect(results).Should(HaveLen(0))
			})

			It("Suggest the optimized cluster groups with BGP network (loosely fit)", func() {
				By("Creating cluster groups with BGP network")
				hostGroupXBX1 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX1, 4,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)
				hostGroupXBX2 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX2, 8,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)
				hostGroupXBX3 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX3, 10,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)
				hostGroupXBX4 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX4, 16,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)

				By("Scheduling an instance group of size 3")
				req := newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 3)
				results, err := bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting hosts in a single cluster group to be selected")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal(clusterGroupXBX1))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupXBX1))
				}

				By("Scheduling an instance group of size 14")
				req = newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 14)
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting hosts in multiple cluster groups with BGP network to be selected")
				expectedHosts := hostGroupXBX2
				expectedHosts = append(expectedHosts, hostGroupXBX3...)
				Expect(results).ShouldNot(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(BeElementOf(clusterGroupXBX2, clusterGroupXBX3))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(expectedHosts))
				}

				By("Scheduling an instance group of size 14")
				req = newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 14)
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting hosts in a single cluster group to be selected")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal(clusterGroupXBX4))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupXBX4))
				}
			})

			It("Suggest a supercompute group based on the requested size", func() {
				By("Creating target groups")
				hostGroupSC1 := createBmHostGroup(ctx, hostNamespace, "sc1-g1", 4,
					fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceTypeSC), "true",
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX,
					bmenroll.SuperComputeGroupID, "sc1")
				hostGroupSC2 := createBmHostGroup(ctx, hostNamespace, "sc2-g1", 10,
					fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceTypeSC), "true",
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX,
					bmenroll.SuperComputeGroupID, "sc2")

				By("Creating non-target groups")
				createBmHostGroup(ctx, hostNamespace, "non-sc", 4,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)

				By("Scheduling a supercompute instance group of size 4 (full cluster reservation)")
				req := newBMInstanceGroupScheduleRequest(instanceTypeSC, instanceGroupName, 4)
				results, err := bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting a single supercompute group to be selected")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal("sc1-g1"))
					Expect(r.SuperComputeGroupId).To(Equal("sc1"))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupSC1))
				}

				By("Scheduling a supercompute instance group of size 6")
				req = newBMInstanceGroupScheduleRequest(instanceTypeSC, instanceGroupName, 6)
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting a single supercompute group to be selected (partial cluster reservation)")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal("sc2-g1"))
					Expect(r.SuperComputeGroupId).To(Equal("sc2"))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupSC2))
				}

				By("Creating another target supercompute group")
				req = newBMInstanceGroupScheduleRequest(instanceTypeSC, instanceGroupName, 4)
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)

				By("Expecting an error due to insufficient capacity")
				Expect(err).Should(HaveOccurred())
				Expect(results).Should(HaveLen(0))
			})

			It("Suggest a supercompute group based on the specified group ID", func() {
				By("Creating supercompute groups")
				hostGroupSC1 := createBmHostGroup(ctx, hostNamespace, "sc1-g1", 4,
					fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceTypeSC), "true",
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX,
					bmenroll.SuperComputeGroupID, "sc1")
				hostGroupSC2 := createBmHostGroup(ctx, hostNamespace, "sc2-g1", 4,
					fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceTypeSC), "true",
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX,
					bmenroll.SuperComputeGroupID, "sc2")
				_ = createBmHostGroup(ctx, hostNamespace, "sc3-g1", 10,
					fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceTypeSC), "true",
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX,
					bmenroll.SuperComputeGroupID, "sc3")
				hostGroupSC4 := createBmHostGroup(ctx, hostNamespace, "sc4-g1", 8,
					fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceTypeSC), "true",
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX,
					bmenroll.SuperComputeGroupID, "sc4")

				By("Targeting a single SC group")

				By("Schduling an instance group of size 4 into sc1")
				instances := []*pb.InstancePrivate{}
				for i := 1; i <= 4; i++ {
					instanceName := fmt.Sprintf("instance-%d", i)
					instance := newBmInstance(instanceName, instanceTypeSC)
					instance.Spec.InstanceGroup = instanceGroupName
					instance.Spec.SuperComputeGroupId = "sc1"
					instance.Spec.InstanceGroupSize = 4
					instances = append(instances, instance)
				}
				req := &pb.ScheduleRequest{Instances: instances}
				results, err := bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal("sc1-g1"))
					Expect(r.SuperComputeGroupId).To(Equal("sc1"))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupSC1))
				}

				By("Schduling an instance group of size 4 into sc2")
				instances = []*pb.InstancePrivate{}
				for i := 1; i <= 4; i++ {
					instanceName := fmt.Sprintf("instance-%d", i)
					instance := newBmInstance(instanceName, instanceTypeSC)
					instance.Spec.InstanceGroup = instanceGroupName
					instance.Spec.SuperComputeGroupId = "sc2"
					instance.Spec.InstanceGroupSize = 4
					instances = append(instances, instance)
				}
				req = &pb.ScheduleRequest{Instances: instances}
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal("sc2-g1"))
					Expect(r.SuperComputeGroupId).To(Equal("sc2"))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupSC2))
				}

				By("Targeting multiple SC groups")

				By("Schduling an instance group of size 4 into sc3 or sc4")
				instances = []*pb.InstancePrivate{}
				for i := 1; i <= 4; i++ {
					instanceName := fmt.Sprintf("instance-%d", i)
					instance := newBmInstance(instanceName, instanceTypeSC)
					instance.Spec.InstanceGroup = instanceGroupName
					instance.Spec.SuperComputeGroupId = "sc3,sc4"
					instance.Spec.ClusterGroupId = ""
					instance.Spec.InstanceGroupSize = 4
					instances = append(instances, instance)
				}
				req = &pb.ScheduleRequest{Instances: instances}
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting sc4 to be selected as it has the least capacity")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal("sc4-g1"))
					Expect(r.SuperComputeGroupId).To(Equal("sc4"))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupSC4))
				}

				By("Creating non-target group sc4 so two groups (sc4 and sc5) have the same capacity")
				createBmHostGroup(ctx, hostNamespace, "sc5-g1", 4,
					fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceTypeSC), "true",
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX,
					bmenroll.SuperComputeGroupID, "sc5")

				By("Schduling another instance group of size 4 into sc3 or sc4")
				instances = []*pb.InstancePrivate{}
				for i := 1; i <= 4; i++ {
					instanceName := fmt.Sprintf("instance-%d", i)
					instance := newBmInstance(instanceName, instanceTypeSC)
					instance.Spec.InstanceGroup = instanceGroupName
					instance.Spec.SuperComputeGroupId = "sc3,sc4"
					instance.Spec.ClusterGroupId = "sc4-g1"
					instance.Spec.InstanceGroupSize = 4
					instances = append(instances, instance)
				}
				req = &pb.ScheduleRequest{Instances: instances}
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting sc4 to be selected again since it is currently assigned")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal("sc4-g1"))
					Expect(r.SuperComputeGroupId).To(Equal("sc4"))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupSC4))
				}
			})

			It("Schedule a large instance group", func() {
				Skip("This test is too large and should be run manually")
				requestedSize := 1024
				hostGroupSize := 8
				numOfHostGroups := requestedSize / hostGroupSize

				By(fmt.Sprintf("Creating %d of %d-node clusters. Total nodes: %d", numOfHostGroups, hostGroupSize, requestedSize))
				expectedHosts := []*baremetalv1alpha1.BareMetalHost{}
				expectedGroupIDs := []string{}
				for i := 0; i < numOfHostGroups; i++ {
					groupID := fmt.Sprintf("xbx-%d", i+1)
					hostGroup := createBmHostGroup(ctx, hostNamespace, groupID, hostGroupSize,
						bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)
					expectedHosts = append(expectedHosts, hostGroup...)
					expectedGroupIDs = append(expectedGroupIDs, groupID)
				}

				By("Scheduling an instance group")
				req := newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, requestedSize)
				start := time.Now()
				results, err := bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				elapsed := time.Since(start)
				By(fmt.Sprintf("Execution time: %s\n", elapsed))

				By("Expecting hosts in multiple groups to be selected")
				Expect(results).ShouldNot(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(BeElementOf(expectedGroupIDs))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(expectedHosts))
				}
			})
		})

		Context("Schedule for an existing instance group", func() {

			It("Schedule new instances into the same cluster group with non-BGP network", func() {
				By("Creating cluster groups: (4/4) (8/8)")
				hostGroupVVV1 := createBmHostGroup(ctx, hostNamespace, clusterGroupVVV1, 4,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeVVV)
				createBmHostGroup(ctx, hostNamespace, clusterGroupVVV2, 8,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeVVV)

				By("Scheduling an instance group of size 2")
				req := newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 2)
				results, err := bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expected capacity: (2/4) (8/8)")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeVVV))
					Expect(r.GroupId).Should(Equal(clusterGroupVVV1))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupVVV1))
				}

				By("Schduling 2 new instances to increase the group size to 4")
				instances := []*pb.InstancePrivate{}
				for i := 3; i <= 4; i++ {
					instanceName := fmt.Sprintf("instance-%d", i)
					instance := newBmInstance(instanceName, instanceType)
					instance.Spec.InstanceGroup = instanceGroupName
					instance.Spec.ClusterGroupId = clusterGroupVVV1
					instance.Spec.NetworkMode = bmenroll.NetworkModeVVV
					instance.Spec.InstanceGroupSize = 4
					instances = append(instances, instance)
				}
				req = &pb.ScheduleRequest{
					Instances: instances,
				}
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expected capacity: (0/4) (8/8)")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeVVV))
					Expect(r.GroupId).Should(Equal(clusterGroupVVV1))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupVVV1))
				}

				By("Schduling 2 new instances to increase the group size to 6")
				instances = []*pb.InstancePrivate{}
				for i := 5; i <= 6; i++ {
					instanceName := fmt.Sprintf("instance-%d", i)
					instance := newBmInstance(instanceName, instanceType)
					instance.Spec.InstanceGroup = instanceGroupName
					instance.Spec.ClusterGroupId = clusterGroupVVV1
					instance.Spec.NetworkMode = bmenroll.NetworkModeVVV
					instance.Spec.InstanceGroupSize = 6
					instances = append(instances, instance)
				}
				req = &pb.ScheduleRequest{
					Instances: instances,
				}
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				log.Info("scheduler response", "results", results)

				By("Expecting an error due to insufficient capacity")
				Expect(err).Should(HaveOccurred())
				Expect(results).Should(HaveLen(0))
			})

			It("Schedule new instances into the currently or newly assigned cluster groups with BGP network", func() {
				By("Creating target BGP cluster groups: (4/4) (6/6) (8/8) (8/8)")
				hostGroupXBX1 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX1, 4,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)
				hostGroupXBX2 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX2, 6,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)
				hostGroupXBX3 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX3, 8,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)
				hostGroupXBX4 := createBmHostGroup(ctx, hostNamespace, clusterGroupXBX4, 8,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX)

				By("Creating non-target cluster groups with non-BGP network")
				createBmHostGroup(ctx, hostNamespace, clusterGroupDefault, 4)
				createBmHostGroup(ctx, hostNamespace, clusterGroupVVX, 6,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeVVXStandalone)
				createBmHostGroup(ctx, hostNamespace, clusterGroupVVV1, 8,
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeVVV)

				By("Scheduling an instance group of size 4")
				req := newBMInstanceGroupScheduleRequest(instanceType, instanceGroupName, 4)
				results, err := bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expected capacity: (0/4) (6/6) (8/8) (8/8)")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal(clusterGroupXBX1))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupXBX1))
				}

				By("Schduling 6 new instances to increase the group size to 10")
				instances := []*pb.InstancePrivate{}
				for i := 5; i <= 10; i++ {
					instanceName := fmt.Sprintf("instance-%d", i)
					instance := newBmInstance(instanceName, instanceType)
					instance.Spec.InstanceGroup = instanceGroupName
					instance.Spec.ClusterGroupId = clusterGroupXBX1
					instance.Spec.NetworkMode = bmenroll.NetworkModeXBX
					instance.Spec.InstanceGroupSize = 10
					instances = append(instances, instance)
				}
				req = &pb.ScheduleRequest{
					Instances: instances,
				}
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expected capacity: (0/4) (0/6) (8/8) (8/8)")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal(clusterGroupXBX2))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupXBX2))
				}

				By("Schduling 16 new instances to increase the group size to 26")
				instances = []*pb.InstancePrivate{}
				for i := 11; i <= 26; i++ {
					instanceName := fmt.Sprintf("instance-%d", i)
					instance := newBmInstance(instanceName, instanceType)
					instance.Spec.InstanceGroup = instanceGroupName
					instance.Spec.ClusterGroupId = strings.Join([]string{clusterGroupXBX1, clusterGroupXBX2}, ",")
					instance.Spec.NetworkMode = bmenroll.NetworkModeXBX
					instance.Spec.InstanceGroupSize = 26
					instances = append(instances, instance)
				}
				req = &pb.ScheduleRequest{
					Instances: instances,
				}
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expected capacity: (0/4) (0/6) (0/8) (0/8)")
				expectedhosts := hostGroupXBX3
				expectedhosts = append(expectedhosts, hostGroupXBX4...)
				Expect(results).ShouldNot(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(BeElementOf(clusterGroupXBX3, clusterGroupXBX4))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(expectedhosts))
				}

				By("Schduling 6 new instances to increase the group size to 32")
				instances = []*pb.InstancePrivate{}
				for i := 26; i <= 32; i++ {
					instanceName := fmt.Sprintf("instance-%d", i)
					instance := newBmInstance(instanceName, instanceType)
					instance.Spec.InstanceGroup = instanceGroupName
					instance.Spec.ClusterGroupId = strings.Join([]string{clusterGroupXBX1, clusterGroupXBX2, clusterGroupXBX3, clusterGroupXBX4}, ",")
					instance.Spec.NetworkMode = bmenroll.NetworkModeXBX
					instance.Spec.InstanceGroupSize = 32
					instances = append(instances, instance)
				}
				req = &pb.ScheduleRequest{
					Instances: instances,
				}
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				log.Info("scheduler response", "results", results)

				By("Expecting an error due to insufficient capacity")
				Expect(err).Should(HaveOccurred())
				Expect(results).Should(HaveLen(0))
			})

			It("Schedule new instances into the currently or newly assigned supercompute groups", func() {
				By("Creating supercompute groups")
				hostGroupSC1 := createBmHostGroup(ctx, hostNamespace, "sc1-g1", 8,
					fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceTypeSC), "true",
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX,
					bmenroll.SuperComputeGroupID, "sc1")
				createBmHostGroup(ctx, hostNamespace, "sc2-g1", 4,
					fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceTypeSC), "true",
					bmenroll.NetworkModeLabel, bmenroll.NetworkModeXBX,
					bmenroll.SuperComputeGroupID, "sc2")

				By("Scheduling a supercompute instance group of size 6 (partial cluster reservation)")
				req := newBMInstanceGroupScheduleRequest(instanceTypeSC, instanceGroupName, 6)
				results, err := bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				By("Expecting a single supercompute group to be selected")
				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal("sc1-g1"))
					Expect(r.SuperComputeGroupId).To(Equal("sc1"))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupSC1))
				}

				By("Schduling new 2 instances into the assigned supercompute group sc-1")
				instances := []*pb.InstancePrivate{}
				for i := 7; i <= 8; i++ {
					instanceName := fmt.Sprintf("instance-%d", i)
					instance := newBmInstance(instanceName, instanceTypeSC)
					instance.Spec.InstanceGroup = instanceGroupName
					instance.Spec.SuperComputeGroupId = "sc1"
					instance.Spec.ClusterGroupId = "sc1-g1"
					instance.Spec.NetworkMode = bmenroll.NetworkModeXBX
					instance.Spec.InstanceGroupSize = 8
					instances = append(instances, instance)
				}
				req = &pb.ScheduleRequest{Instances: instances}
				results, err = bmSchedulerTestEnv.scheduleMultipleBmInstances(ctx, req)
				Expect(err).Should(Succeed())
				log.Info("scheduler response", "results", results)

				Expect(results).Should(HaveOneGroupID())
				for _, r := range results {
					Expect(r.NetworkMode).Should(Equal(bmenroll.NetworkModeXBX))
					Expect(r.GroupId).Should(Equal("sc1-g1"))
					Expect(r.SuperComputeGroupId).To(Equal("sc1"))
					Expect(r.ClusterId).Should(Equal(hostNamespace))
					Expect(r.NodeId).Should(EqualAnyHostNames(hostGroupSC1))
				}
			})
		})
	})
})

func NewBMSchedulerTestEnv(opts BmSchedulerOptions) *BmSchedulerTestEnv {
	// Create a context that will last throughout the lifetime of the test environment.
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	testEnv := &BmSchedulerTestEnv{
		SchedulerOpts: opts,
		Ctx:           ctx,
		Cancel:        cancel,
	}
	return testEnv
}

func (e *BmSchedulerTestEnv) Start() {
	ctx := e.Ctx
	log := log.FromContext(ctx).WithName("NewBMSchedulerTestEnv")
	log.Info("BEGIN")
	defer log.Info("END")

	_ = os.Setenv("ENABLE_VALIDATION_FEATURE", "true")

	By("WriteKubeConfigFiles")
	restConfigs := make(map[string]*rest.Config)
	restConfigs["bmConfig.hconf"] = k8sRestConfig
	kubeConfigDir, err := os.MkdirTemp("", "")
	Expect(err).ToNot(HaveOccurred())
	log.Info("kubeConfigDir", "kubeConfigDir", "kubeConfigDir")
	Expect(toolsk8s.WriteKubeConfigFiles(ctx, kubeConfigDir, restConfigs)).Should(Succeed())

	By("Creating manager")
	k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	Expect(err).ToNot(HaveOccurred())
	managerStoppable := stoppable.New(k8sManager.Start)

	By("Creating scheduling server")
	grpcServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	grpcListenPort := uint16(grpcServerListener.Addr().(*net.TCPAddr).Port)
	cfg := &privatecloudv1alpha1.VmInstanceSchedulerConfig{
		ListenPort:                               grpcListenPort,
		BmKubeConfigDir:                          kubeConfigDir,
		EnableBMaaSLocal:                         true,
		EnableBMaaSBinpack:                       true,
		BGPNetworkRequiredInstanceCountThreshold: 3, // use a small threshold for tests
		OvercommitConfig: privatecloudv1alpha1.OvercommitConfig{
			CPU:     100,
			Memory:  100,
			Storage: 100,
		},
	}
	mockController := gomock.NewController(GinkgoT())
	instanceTypeClient := pb.NewMockInstanceTypeServiceClient(mockController)
	e.SchedulingServer, err = server.NewSchedulingServer(ctx, cfg, k8sManager, grpcServerListener, instanceTypeClient)
	Expect(err).Should(Succeed())

	if e.SchedulerOpts.Namespace != "" {
		By("creating namespaces for BareMetalHosts")
		Expect(k8sClient.Create(ctx, newBmNamespace(e.SchedulerOpts.Namespace))).Should(Succeed())
	}

	if e.SchedulerOpts.HostName != "" {
		By("creating available BareMetalHosts")
		Expect(k8sClient.Create(ctx, newAvailableBmHost(e.SchedulerOpts.HostName, e.SchedulerOpts.Namespace))).Should(Succeed())
	}

	By("Starting manager")
	managerStoppable.Start(ctx)

	By("Creating GRPC client")
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	e.SchedulerServiceClient = pb.NewInstanceSchedulingServiceClient(clientConn)

	By("Waiting for service to become ready")
	Eventually(func(g Gomega) {
		_, err := e.SchedulerServiceClient.Ready(ctx, &emptypb.Empty{})
		g.Expect(err).Should(Succeed())
	}, "30s").Should(Succeed())
	By("Service is ready")
}

func (e *BmSchedulerTestEnv) Stop() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("BmSchedulerTestEnv.Stop")
	log.Info("BEGIN")
	defer log.Info("END")

	By("Cancelling context")
	e.Cancel()

	By("Cleanup test env")
	_ = os.Unsetenv("ENABLE_VALIDATION_FEATURE")

	By("Wait for BareMetalHosts in all namespaces to be deleted")
	// This is required because the scheduler will watch BareMetalHosts in all namespaces.
	Eventually(func(g Gomega) {
		bareMetalHostList := &baremetalv1alpha1.BareMetalHostList{}
		g.Expect(k8sClient.List(ctx, bareMetalHostList)).Should(Succeed())
		for _, bmh := range bareMetalHostList.Items {
			By("Deleting " + bmh.Namespace + "/" + bmh.Name)
			_ = k8sClient.Delete(ctx, &bmh)
		}
		g.Expect(len(bareMetalHostList.Items)).Should(Equal(0))
	}, "60s").Should(Succeed())

	By("Deleting namespace " + e.SchedulerOpts.Namespace)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: e.SchedulerOpts.Namespace,
		},
	}
	_ = k8sClient.Delete(ctx, ns)
}

func newBmNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				bmenroll.Metal3NamespaceSelectorKey: "true",
			},
		},
	}
}

func (e *BmSchedulerTestEnv) scheduleOneBmInstance(ctx context.Context, req *pb.ScheduleRequest) (*pb.ScheduleInstanceResult, error) {
	resp, err := e.SchedulerServiceClient.Schedule(ctx, req)
	if err != nil {
		return nil, err
	}
	Expect(len(resp.InstanceResults)).Should(Equal(1))
	return resp.InstanceResults[0], err
}

func (e *BmSchedulerTestEnv) scheduleMultipleBmInstances(ctx context.Context, req *pb.ScheduleRequest) ([]*pb.ScheduleInstanceResult, error) {
	resp, err := e.SchedulerServiceClient.Schedule(ctx, req)
	if err != nil {
		return nil, err
	}
	Expect(len(resp.InstanceResults)).Should(Equal(len(req.Instances)))
	return resp.InstanceResults, err
}

func newBmScheduleRequest(name string, instanceType string) *pb.ScheduleRequest {
	req := &pb.ScheduleRequest{
		Instances: []*pb.InstancePrivate{
			newBmInstance(name, instanceType),
		},
	}
	return req
}

func newBMInstanceGroupScheduleRequest(instanceType string, instanceGroupName string, instanceGroupSize int) *pb.ScheduleRequest {
	req := &pb.ScheduleRequest{
		Instances: []*pb.InstancePrivate{},
	}
	for i := 1; i <= instanceGroupSize; i++ {
		instanceName := fmt.Sprintf("instance-%d", i)
		instance := newBmInstance(instanceName, instanceType)
		instance.Spec.InstanceGroup = instanceGroupName
		instance.Spec.InstanceGroupSize = int32(instanceGroupSize)
		instance.Spec.ComputeNodePools = []string{"general", "habana"}
		req.Instances = append(req.Instances, instance)
	}
	return req
}

func newBmHost(name, namespace string) *baremetalv1alpha1.BareMetalHost {
	bmh := &baremetalv1alpha1.BareMetalHost{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "metal3.io/v1alpha1",
			Kind:       "BareMetalHost",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				bmenroll.CPUIDLabel:                                   "0x00001",
				bmenroll.CPUCountLabel:                                "1",
				bmenroll.GPUModelNameLabel:                            "",
				bmenroll.GPUCountLabel:                                "0",
				bmenroll.HBMModeLabel:                                 "",
				fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceType): "true",
				bmenroll.VerifiedLabel:                                "true",
				bmenroll.NetworkModeLabel:                             "",
				"pool.cloud.intel.com/general":                        "true",
				"pool.cloud.intel.com/habana":                         "true",
			},
		},
		Spec: baremetalv1alpha1.BareMetalHostSpec{
			Online:         true,
			BootMode:       baremetalv1alpha1.Legacy,
			BootMACAddress: "a1:b2:c3:d4:f5:e6",
			BMC: baremetalv1alpha1.BMCDetails{
				Address:                        "redfish+http://10.11.12.13:8001/redfish/v1/Systems/1",
				CredentialsName:                "secret-1",
				DisableCertificateVerification: true,
			},
			RootDeviceHints: &baremetalv1alpha1.RootDeviceHints{
				DeviceName: "/dev/vda",
			},
		},
		Status: baremetalv1alpha1.BareMetalHostStatus{
			Provisioning: baremetalv1alpha1.ProvisionStatus{
				State: baremetalv1alpha1.StateAvailable,
			},
		},
	}
	return bmh
}

func newAvailableBmHost(name, namespace string) *baremetalv1alpha1.BareMetalHost {
	bmh := newBmHost(name, namespace)
	bmh.Status = baremetalv1alpha1.BareMetalHostStatus{
		Provisioning: baremetalv1alpha1.ProvisionStatus{
			State: baremetalv1alpha1.StateAvailable,
		},
		OperationalStatus: baremetalv1alpha1.OperationalStatusOK,
		HardwareDetails: &baremetalv1alpha1.HardwareDetails{
			NIC: []baremetalv1alpha1.NIC{
				{
					PXE:    true,
					MAC:    "a1:b2:c3:d4:f5:e6",
					Model:  "test-nic",
					Name:   "eno0",
					VLANID: 0,
				},
			},
		},
	}
	return bmh
}

func createBmHostGroup(ctx context.Context, namespace string, groupID string, groupSize int, labelKeyValues ...any) []*baremetalv1alpha1.BareMetalHost {
	By(fmt.Sprintf("Creating hosts in cluster group %q with size %d", groupID, groupSize))

	hosts := []*baremetalv1alpha1.BareMetalHost{}
	err := k8sClient.Get(ctx, client.ObjectKey{Name: namespace}, &corev1.Namespace{})
	if errors.IsNotFound(err) {
		Expect(k8sClient.Create(ctx, newBmNamespace(namespace))).Should(Succeed())
	}

	for i := 1; i <= groupSize; i++ {
		hostNamespace := namespace
		hostname := fmt.Sprintf("group-%s-host-%d", groupID, i)
		if groupID == "" {
			hostname = fmt.Sprintf("group-host-%d", i)
		}
		host := newAvailableBmHost(hostname, hostNamespace)
		host.Labels[bmenroll.ClusterGroupID] = groupID
		host.Labels[bmenroll.VerifiedLabel] = "true"
		// add extra labels
		Expect(len(labelKeyValues) % 2).Should(BeZero())
		for i := 0; i < len(labelKeyValues); i += 2 {
			key := fmt.Sprintf("%v", labelKeyValues[i])
			value := fmt.Sprintf("%v", labelKeyValues[i+1])
			host.Labels[key] = value
		}
		// create host
		hostStatus := host.Status
		Expect(k8sClient.Create(ctx, host)).Should(Succeed())
		// update host status
		host.Status = hostStatus
		Expect(k8sClient.Status().Update(ctx, host)).Should(Succeed())
		hosts = append(hosts, host)
	}
	return hosts
}

func newBmInstance(name string, instanceType string) *pb.InstancePrivate {
	resourceId := uuid.NewString()
	return &pb.InstancePrivate{
		Metadata: &pb.InstanceMetadataPrivate{
			CloudAccountId: uuid.NewString(),
			Name:           fmt.Sprintf("%s-instance-%s", name, resourceId),
			ResourceId:     resourceId,
			Labels: map[string]string{
				fmt.Sprintf(bmenroll.InstanceTypeLabel, instanceType): "true",
			},
		},
		Spec: &pb.InstanceSpecPrivate{
			AvailabilityZone: "us-dev-1a",
			ComputeNodePools: []string{"general", "habana"},
			InstanceType:     instanceType,
			InstanceTypeSpec: &pb.InstanceTypeSpec{
				Name:             instanceType,
				InstanceCategory: pb.InstanceCategory_BareMetalHost,
				Cpu: &pb.CpuSpec{
					Cores:   1,
					Sockets: 1,
					Threads: 1,
					Id:      "0x00001",
				},
				Memory: &pb.MemorySpec{
					Size: "8Gi",
				},
				Disks: []*pb.DiskSpec{{
					Size: "10Gi",
				}},
			},
		},
	}
}

func EqualAnyHostNames(hosts []*baremetalv1alpha1.BareMetalHost) gomegatypes.GomegaMatcher {
	return WithTransform(func(hostName string) bool {
		for _, host := range hosts {
			if host.Name == hostName {
				return true
			}
		}
		return false
	}, BeTrue())
}

func HaveOneGroupID() gomegatypes.GomegaMatcher {
	return WithTransform(func(results []*pb.ScheduleInstanceResult) bool {
		if len(results) == 0 {
			return false
		}
		for _, r := range results {
			if r.GroupId != results[0].GroupId {
				return false
			}
		}
		return true
	}, BeTrue())
}
