// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	bmenroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	enroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	instanceoperatorutil "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_replicator/convert"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	commonutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	DefaultSchedulerName = "default"
	InstanceTypeLabel    = "instance-type.cloud.intel.com/"
)

type ResourceQuantity struct {
	CPU    resource.Quantity
	Memory resource.Quantity
}

// GetStatistics retrieves the statistics for the scheduler, including node statistics and resource allocations.
func (sched *Scheduler) GetStatistics(ctx context.Context) (*pb.SchedulerStatistics, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("Scheduler.GetStatistics").Start()
	defer span.End()
	log.Info("BEGIN", "region", sched.cfg.Region, "availabilityZone", sched.cfg.AvailabilityZone)
	defer log.Info("END", "region", sched.cfg.Region, "availabilityZone", sched.cfg.AvailabilityZone)

	// Prepare the response object
	response := &pb.SchedulerStatistics{
		SchedulerNodeStatistics: []*pb.SchedulerNodeStatistics{},
	}

	nodeInfos, err := sched.listNodesFromCache(ctx)
	if err != nil {
		log.Error(err, "Failed to list NodesFromCache")
		return nil, err
	}
	for _, nodeInfo := range nodeInfos {
		nodeStats, err := sched.getNodeStatistics(ctx, nodeInfo)
		if err != nil {
			log.Error(err, "Failed to get node statistics")
			return nil, err
		}
		response.SchedulerNodeStatistics = append(response.SchedulerNodeStatistics, nodeStats)
	}

	log.Info("final response", "response", response)
	return response, nil
}

// Collects and returns statistics for a specific node, including CPU, memory, and GPU resources.
func (sched *Scheduler) getNodeStatistics(ctx context.Context, nodeInfo *framework.NodeInfo) (*pb.SchedulerNodeStatistics, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("Scheduler.getNodeStatistics").Start()
	defer span.End()
	log.Info("BEGIN", "node", nodeInfo.Node().Name)

	// calculate and get stats data
	freeMilliCPU := nodeInfo.Allocatable.MilliCPU - nodeInfo.Requested.MilliCPU
	usedMilliCPU := nodeInfo.Requested.MilliCPU
	freeMemoryBytes := nodeInfo.Allocatable.Memory - nodeInfo.Requested.Memory
	usedMemoryBytes := nodeInfo.Requested.Memory
	instanceCategory, err := sched.getNodeInstanceCategory(nodeInfo)
	if err != nil {
		return nil, err
	}

	/*
		Node names follow different conventions depending on the instance type.

		For Bare Metal hosts, nodeInfo.Node().Name is formatted as "namespace/node_name",
		and the cluster ID is a fixed value: BMaaS.

		For Virtual Machines, nodeInfo.Node().Name is formatted as "clusterId/node_name",
		and the namespace remains empty.

		This logic is subject to change with the Bare Metal MultiCluster implementation
		tracked by PR#9449.  This code should be reviewed and updated once that PR is
		ready.
	*/
	nodeNamePrefix, nodeName, err := extractClusterFromNodeName(nodeInfo.Node().Name)
	if err != nil {
		return nil, err
	}
	namespace, clusterID := "", nodeNamePrefix
	if instanceCategory == pb.InstanceCategory_BareMetalHost {
		namespace, clusterID = nodeNamePrefix, BmaasLocalCluster
	}

	partition := nodeInfo.Node().Labels[instanceoperatorutil.TopologySpreadTopologyKey]
	networkMode := nodeInfo.Node().Labels[instanceoperatorutil.NetworkModeKey]
	clusterGroup := nodeInfo.Node().Labels[bmenroll.ClusterGroup]

	freeGPU, usedGPU, err := sched.getGPUStatistics(ctx, nodeInfo)
	if err != nil {
		return nil, err
	}

	// TODO: evaluate the suitability for caching
	instanceTypeInfo, err := sched.getInstanceTypeInfo(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: evaluate the suitability for other affinities (NodePools, TopologySpreadTopologyKey (cloud.intel.com/partition))
	runningInstanceMap, err := sched.getRunningInstancesByType(ctx, nodeInfo)
	if err != nil {
		return nil, err
	}

	computeNodePoolsIds := getNodePoolsIds(nodeInfo.Node().Labels)
	log.Info("Compute NodePools", "computeNodePoolsIds", computeNodePoolsIds)

	sourceGvr, err := sched.getGroupVersionResource(nodeInfo)
	if err != nil {
		return nil, err
	}

	schedulerNodeStatistics := &pb.SchedulerNodeStatistics{
		SchedulerNode: &pb.SchedulerNode{
			Region:           sched.cfg.Region,
			AvailabilityZone: sched.cfg.AvailabilityZone,
			ClusterId:        clusterID,
			Namespace:        namespace,
			NodeName:         nodeName,
			Partition:        partition,
			ClusterGroup:     clusterGroup,
			NetworkMode:      networkMode,
			SourceGvr:        sourceGvr,
			ComputeNodePools: computeNodePoolsIds,
		},
		NodeResources: &pb.NodeResources{
			FreeMilliCPU:    freeMilliCPU,
			UsedMilliCPU:    usedMilliCPU,
			FreeMemoryBytes: freeMemoryBytes,
			UsedMemoryBytes: usedMemoryBytes,
			FreeGPU:         freeGPU,
			UsedGPU:         usedGPU,
		},
	}

	instanceTypeStatistics, err := sched.GetInstanceTypeStatistics(ctx, nodeInfo, schedulerNodeStatistics, instanceTypeInfo, runningInstanceMap)
	if err != nil {
		return nil, err
	}
	schedulerNodeStatistics.InstanceTypeStatistics = instanceTypeStatistics

	return schedulerNodeStatistics, nil
}

// Get the Node API resource by group, version, and resource type.
func (sched *Scheduler) getGroupVersionResource(nodeInfo *framework.NodeInfo) (*pb.GroupVersionResource, error) {
	instanceCategory, err := sched.getNodeInstanceCategory(nodeInfo)
	if err != nil {
		return nil, err
	}

	switch instanceCategory {
	case pb.InstanceCategory_BareMetalHost:
		return &pb.GroupVersionResource{
			Group:    "metal3.io",
			Version:  "v1alpha1",
			Resource: "baremetalhosts",
		}, nil
	case pb.InstanceCategory_VirtualMachine:
		return &pb.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "nodes",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported k8s node: %s", nodeInfo.Node().Name)
	}
}

// determines the category of a node (VM or BM) based on its labels
func (sched *Scheduler) getNodeInstanceCategory(nodeInfo *framework.NodeInfo) (pb.InstanceCategory, error) {
	for label := range nodeInfo.Node().Labels {
		if strings.Contains(label, "kubevirt.io") {
			return pb.InstanceCategory_VirtualMachine, nil
		} else if strings.Contains(label, "cloud.intel.com/host-memory-size") {
			return pb.InstanceCategory_BareMetalHost, nil
		}
	}
	return -1, fmt.Errorf("InstanceCategory cannot be determined for Node: %s", nodeInfo.Node().Name)
}

func (sched *Scheduler) GetInstanceTypeStatistics(ctx context.Context, nodeInfo *framework.NodeInfo,
	schedulerNodeStatistics *pb.SchedulerNodeStatistics, instanceTypeInfo map[string]*pb.InstanceType,
	instanceTypeToRunningCountMap map[string]int) ([]*pb.InstanceTypeStatistics, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("Scheduler.GetInstanceTypeStatistics").Start()
	defer span.End()
	log.Info("BEGIN", "node", nodeInfo.Node().Name, "instanceTypeToRunningCountMap", instanceTypeToRunningCountMap)
	defer log.Info("END", "node", nodeInfo.Node().Name)

	// Prepare the response object
	instanceTypeStatistics := []*pb.InstanceTypeStatistics{}
	var instanceTypeName string
	for label, value := range nodeInfo.Node().Labels {
		if strings.HasPrefix(label, InstanceTypeLabel) && value == "true" {
			log.Info("Node label", "label", label)
			substrings := strings.Split(label, InstanceTypeLabel)
			if substrings[1] != "" {
				instanceTypeName = substrings[1]
				instanceType := instanceTypeInfo[instanceTypeName]
				if instanceType == nil {
					log.Info("Missing instanceType", "instanceTypeName", instanceTypeName)
					continue
				}
				log.Info("InstanceType from node label", "instanceTypeName", instanceTypeName, "instanceType", instanceType)
				resourceQuantity, err := sched.calculateResourceQuantities(instanceType.Spec)
				if err != nil {
					return nil, err
				}
				maxNewInstances := calculateMaxNewInstances(schedulerNodeStatistics, resourceQuantity, instanceType)

				instanceCategory, err := sched.getNodeInstanceCategory(nodeInfo)
				if err != nil {
					return nil, err
				}
				stats := &pb.InstanceTypeStatistics{
					InstanceType:     instanceTypeName,
					RunningInstances: int32(instanceTypeToRunningCountMap[instanceTypeName]),
					MaxNewInstances:  int32(maxNewInstances),
					InstanceCategory: instanceCategory.String(),
				}
				instanceTypeStatistics = append(instanceTypeStatistics, stats)
			}
		}
	}

	log.Info("InstanceType Statistics", "instanceTypeStatistics", instanceTypeStatistics)
	return instanceTypeStatistics, nil
}

func calculateMaxNewInstances(nodeStatistics *pb.SchedulerNodeStatistics, resourceQuantity *ResourceQuantity, instanceType *pb.InstanceType) int64 {
	// Calculate how many additional instances can be supported by the CPU resources available
	freeCpuInstances := nodeStatistics.NodeResources.FreeMilliCPU / resourceQuantity.CPU.MilliValue()

	// Calculate how many additional instances can be supported by the memory resources available
	freeMemoryInstances := nodeStatistics.NodeResources.FreeMemoryBytes / resourceQuantity.Memory.Value()

	// Determine the maximum new instances based on the lesser of CPU or memory capacity
	maxInstances := freeCpuInstances
	if freeMemoryInstances < freeCpuInstances {
		maxInstances = freeMemoryInstances
	}

	// If GPUs are present, adjust the maximum new instances based on GPU availability
	if instanceType.Spec.Gpu != nil && instanceType.Spec.Gpu.Count > 0 {
		freeGpuInstances := int64(nodeStatistics.NodeResources.FreeGPU / instanceType.Spec.Gpu.Count)
		if freeGpuInstances < maxInstances {
			maxInstances = freeGpuInstances
		}
	}
	return maxInstances
}

func (sched *Scheduler) calculateResourceQuantities(instanceTypeSpec *pb.InstanceTypeSpec) (*ResourceQuantity, error) {
	milliCpuQty := int64(instanceTypeSpec.Cpu.Cores*1000) * int64(100) / int64(sched.cfg.OvercommitConfig.CPU)
	cpuQtyReq := *resource.NewMilliQuantity(milliCpuQty, resource.DecimalSI)
	memoryQty, err := resource.ParseQuantity(instanceTypeSpec.Memory.Size)
	if err != nil {
		return nil, fmt.Errorf("unable to parse memory size: %v", err)
	}
	return &ResourceQuantity{
		CPU:    cpuQtyReq,
		Memory: memoryQty,
	}, nil
}

func (sched *Scheduler) getRunningInstancesByType(ctx context.Context, nodeInfo *framework.NodeInfo) (map[string]int, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("Scheduler.getRunningInstancesByType").Start()
	defer span.End()
	log.Info("BEGIN", "node", nodeInfo.Node().Name)
	defer log.Info("END", "node", nodeInfo.Node().Name)

	instanceTypeFrequencyMap := make(map[string]int)

	// Function to increment the map count
	incrementMap := func(instanceTypeName string, count int) {
		instanceTypeFrequencyMap[instanceTypeName] += count
	}
	nodeInstanceCategory, err := sched.getNodeInstanceCategory(nodeInfo)
	if err != nil {
		return nil, err
	}
	if nodeInstanceCategory == pb.InstanceCategory_BareMetalHost {
		// Handling for Bare Metal Hosts (BMaaS)
		for label := range nodeInfo.Node().Labels {
			if strings.HasPrefix(label, InstanceTypeLabel) && nodeInfo.Node().Labels[label] == "true" {
				instanceTypeName := strings.TrimPrefix(label, InstanceTypeLabel)
				incrementMap(instanceTypeName, len(nodeInfo.Pods))
			}
		}
	} else if nodeInstanceCategory == pb.InstanceCategory_VirtualMachine {
		// Handling for Virtual Machine (VMaaS)
		for _, pod := range nodeInfo.Pods {
			// Filtering to obtain vmaas-created PODs. This will be removed once backward compatibility support is discontinued.
			if pod.Pod.Labels["harvesterhci.io/creator"] == "harvester-vmaas" {
				log.Info("vmaas created pod", "podLabels", pod.Pod.Labels)
				processed := false
				for label := range pod.Pod.Labels {
					if strings.HasPrefix(label, InstanceTypeLabel) && pod.Pod.Labels[label] == "true" {
						instanceTypeName := strings.TrimPrefix(label, InstanceTypeLabel)
						incrementMap(instanceTypeName, 1)
						processed = true
						log.Info("instance-type using pod label", "instanceType", instanceTypeName)
						break
					}
				}
				// Backward compatibility
				if !processed {
					log.Info("backward compatibility ", "podLabels", pod.Pod.Labels)
					for _, nodeSelectorTerm := range pod.Pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
						for _, matchExpression := range nodeSelectorTerm.MatchExpressions {
							log.Info("node selector", "matchExpression", matchExpression)
							if strings.HasPrefix(matchExpression.Key, InstanceTypeLabel) &&
								len(matchExpression.Values) > 0 &&
								commonutils.ContainsTrue(matchExpression.Values) {
								instanceTypeName := strings.TrimPrefix(matchExpression.Key, InstanceTypeLabel)
								incrementMap(instanceTypeName, 1)
								processed = true
								log.Info("instance-type using affinity", "instanceType", instanceTypeName)
								break
							}
						}
					}
				}
			}
		}
	} else {
		log.Info("unsupported instanceCategory", "nodeInstanceCategory", nodeInstanceCategory)
	}

	return instanceTypeFrequencyMap, nil
}

// getInstanceTypes retrieves instance types.
func (sched *Scheduler) getInstanceTypeInfo(ctx context.Context) (map[string]*pb.InstanceType, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("Scheduler.getInstanceTypeInfo").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	instanceTypesSearchResponse, err := sched.instanceTypeServiceClient.Search(ctx, &pb.InstanceTypeSearchRequest{})
	if err != nil {
		return nil, fmt.Errorf("unable to get instanceTypeList information from InstanceTypeClient: %v", err)
	}
	instanceTypes := make(map[string]*pb.InstanceType)
	for _, instanceType := range instanceTypesSearchResponse.Items {
		log.Info("InstanceType Info", "instanceType", instanceType.Metadata.Name)
		instanceTypes[instanceType.Metadata.Name] = instanceType
	}
	return instanceTypes, nil
}

// getGPUStatistics calculates and returns the free and used GPU resources for a node.
func (sched *Scheduler) getGPUStatistics(ctx context.Context, nodeInfo *framework.NodeInfo) (int32, int32, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("Scheduler.getGPUStatistics").Start()
	defer span.End()
	log.Info("BEGIN", "nodeName", nodeInfo.Node().Name)
	defer log.Info("END", "nodeName", nodeInfo.Node().Name)

	var freeGPU, usedGPU int32
	for resource, value := range nodeInfo.Allocatable.ScalarResources {
		// Assumption: A single node cannot have multiple types of allocatable GPUs
		if value > 0 {
			if utils.IsScalarResourceName(resource) {
				totalGPU := int32(value)
				usedGPU = int32(nodeInfo.Requested.ScalarResources[resource])
				freeGPU = totalGPU - usedGPU
				break
			} else {
				log.Info("unknown scalarResourceName", "scalarResourceName", resource)
			}
		}
	}

	return freeGPU, usedGPU, nil
}

// Schedule resources (1 or more instances).
func (sched *Scheduler) Schedule(ctx context.Context, instances []*pb.InstancePrivate, dryRun bool) (*pb.ScheduleResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Scheduler.Schedule").Start()
	defer span.End()
	log.Info("Request", logkeys.InstanceCount, len(instances))

	instanceCount := len(instances)
	if instanceCount == 0 {
		return &pb.ScheduleResponse{}, nil
	}

	netInfo, err := sched.getNetworkInfo(ctx, instances)
	if err != nil {
		return nil, err
	}

	assumedPods := []*corev1.Pod{}
	if dryRun {
		defer func() { sched.unassumePods(ctx, assumedPods) }()
	}

	resp, err := func() (*pb.ScheduleResponse, error) {
		resp := &pb.ScheduleResponse{}
		for instanceIndex, instance := range instances {
			requestedSize := len(instances) - instanceIndex
			pod, err := sched.InstanceToPod(ctx, instance, netInfo)
			if err != nil {
				return nil, err
			}
			log.Info("podToSchedule request", logkeys.InstanceIndex, instanceIndex, logkeys.Pod, pod)
			scheduleResult, err := sched.scheduleOnePod(ctx, pod, requestedSize)
			if err != nil {
				return nil, err
			}
			log.Info("podToSchedule result", logkeys.InstanceIndex, instanceIndex, logkeys.Result, scheduleResult)
			pod.Spec.NodeName = scheduleResult.SuggestedHost
			assumedPods = append(assumedPods, pod)
			clusterId, nodeId, err := extractClusterFromNodeName(scheduleResult.SuggestedHost)
			if err != nil {
				return nil, err
			}
			groupId := ""
			if value, exists := scheduleResult.SuggestedHostLabels[enroll.ClusterGroupID]; exists {
				groupId = value
			}
			superComputeGroupId := ""
			if value, exists := scheduleResult.SuggestedHostLabels[enroll.SuperComputeGroupID]; exists {
				superComputeGroupId = value
			}
			networkMode := ""
			if value, exists := scheduleResult.SuggestedHostLabels[enroll.NetworkModeLabel]; exists {
				networkMode = value
			}

			// Intersection of compute node pools in request and those in recommended node
			computeNodePools := commonutils.Intersect(instance.Spec.ComputeNodePools, getNodePoolsIds(scheduleResult.SuggestedHostLabels))

			result := &pb.ScheduleInstanceResult{
				ClusterId:           clusterId,
				NodeId:              nodeId,
				Partition:           scheduleResult.Partition,
				GroupId:             groupId,
				ComputeNodePools:    computeNodePools,
				NetworkMode:         networkMode,
				SuperComputeGroupId: superComputeGroupId,
			}
			resp.InstanceResults = append(resp.InstanceResults, result)
		}
		return resp, nil
	}()
	if err != nil {
		// An error occurred. Unreserve all pods that were reserved by this function.
		sched.unassumePods(ctx, assumedPods)
		if len(instances) > 1 {
			err = fmt.Errorf("scheduling instance %d of %d: %w", len(assumedPods), len(instances), err)
		}
		return nil, err
	}
	return resp, nil
}

// Converts an instance into a pod that is expected to be similar to the pod that will ultimately be created
// by Compute API Server, Instance Replicator, VM Instance Operator, and Kubevirt.
// This virtual pod will be used by this scheduler to select the best cluster and node.
func (sched *Scheduler) InstanceToPod(ctx context.Context, instance *pb.InstancePrivate, net *instanceoperatorutil.InstanceNetworkInfo) (*corev1.Pod, error) {
	log := log.FromContext(ctx).WithName("Scheduler.InstanceToPod")
	log.Info("BEGIN")
	defer log.Info("END")

	if instance == nil ||
		instance.Metadata == nil ||
		instance.Metadata.ResourceId == "" ||
		instance.Metadata.CloudAccountId == "" ||
		instance.Spec == nil ||
		instance.Spec.InstanceTypeSpec == nil ||
		instance.Spec.InstanceTypeSpec.Cpu == nil ||
		instance.Spec.InstanceTypeSpec.Memory == nil {
		return nil, fmt.Errorf("Scheduler.InstanceToPod: incomplete instance")
	}
	resourceQuantity, err := sched.calculateResourceQuantities(instance.Spec.InstanceTypeSpec)
	if err != nil {
		return nil, err
	}

	memoryQty, err := resource.ParseQuantity(instance.Spec.InstanceTypeSpec.Memory.Size)
	if err != nil {
		return nil, fmt.Errorf("Scheduler.InstanceToPod: unable to parse memory size: %w", err)
	}
	if instance.Spec.InstanceTypeSpec.InstanceCategory == pb.InstanceCategory_BareMetalHost {
		if instance.Metadata.Labels == nil {
			instance.Metadata.Labels = make(map[string]string)
		}
		if net.NetworkMode != "" {
			instance.Metadata.Labels[enroll.NetworkModeLabel] = net.NetworkMode
		}
		if instance.Spec.ClusterGroupId != "" {
			if len(strings.Split(instance.Spec.ClusterGroupId, ",")) > 1 {
				// the scheduler will decide the final group ID
				instance.Spec.ClusterGroupId = ""
			}
			if net.NetworkMode != enroll.NetworkModeXBX {
				// non-bgp network mode requires a single group
				net.ClusterGroupIDs = []string{instance.Spec.ClusterGroupId}
			}
		}
		if instance.Spec.SuperComputeGroupId != "" {
			if len(strings.Split(instance.Spec.SuperComputeGroupId, ",")) > 1 {
				// the scheduler will decide the final group ID
				instance.Spec.SuperComputeGroupId = ""
			}
		}
	}
	converter := convert.NewInstanceConverter()
	k8sInstance, err := converter.PbToK8s(instance)
	if err != nil {
		return nil, err
	}
	affinity, err := AffinityForInstanceScheduling(k8sInstance, sched.enableBinpack, *net)
	if err != nil {
		return nil, err
	}
	topologySpreadConstraints, err := TopologySpreadConstraintsForInstanceScheduling(k8sInstance)
	if err != nil {
		return nil, err
	}
	labels := map[string]string{
		framework.ResourceIdPodLabel: instance.Metadata.ResourceId,
	}
	for k, v := range instance.Metadata.Labels {
		labels[k] = v
	}
	for k, v := range k8sInstance.Labels {
		labels[k] = v
	}
	podName := fmt.Sprintf("virt-launcher-%s-abcde", instance.Metadata.ResourceId)
	if instance.Spec.InstanceTypeSpec.InstanceCategory == pb.InstanceCategory_BareMetalHost {
		podName = instance.Metadata.ResourceId
	}
	if instance.Spec.ClusterGroupId != "" {
		labels[enroll.ClusterGroupID] = instance.Spec.ClusterGroupId
	}

	// Handle VM overhead memory
	if instance.Spec.InstanceTypeSpec.InstanceCategory == pb.InstanceCategory_VirtualMachine {
		gpuCount := int32(0)
		if gpu := instance.Spec.InstanceTypeSpec.Gpu; gpu != nil {
			gpuCount = gpu.Count
		}
		log.Info("memoryQty", "WithoutOverhead", memoryQty.Value())

		// Currently, only consider VM overhead memory for non-GPU instance types during scheduling
		if gpuCount == 0 {
			// Get and add the VM overhead memory if applicable
			if vmOverheadMemoryQty := instanceoperatorutil.GetVmOverheadMemory(ctx, memoryQty, gpuCount); vmOverheadMemoryQty.Value() > 0 {
				memoryQty.Add(vmOverheadMemoryQty)
				log.Info("memoryQty", "WithOverhead", memoryQty.Value())
			}
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: instance.Metadata.CloudAccountId,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Affinity: affinity,
			Containers: []corev1.Container{{
				Resources: corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU:    resourceQuantity.CPU,
						corev1.ResourceMemory: resourceQuantity.Memory,
					},
				},
				// Value of required field does not matter.
				Image: "image0",
				// Value of required field does not matter.
				Name: "container0",
			}},
			SchedulerName:             DefaultSchedulerName,
			TopologySpreadConstraints: topologySpreadConstraints,
		},
	}
	if instance.Spec.InstanceTypeSpec.InstanceCategory == pb.InstanceCategory_VirtualMachine &&
		instance.Spec.InstanceTypeSpec.Gpu != nil &&
		instance.Spec.InstanceTypeSpec.Gpu.Count > 0 {
		addGpuResources(ctx, pod, instance)
	}
	return pod, nil
}

func (sched *Scheduler) getNetworkInfo(ctx context.Context, instances []*pb.InstancePrivate) (*instanceoperatorutil.InstanceNetworkInfo, error) {
	log := log.FromContext(ctx).WithName("Scheduler.getNetworkInfo")
	instance := instances[0]

	if instance.Spec.InstanceTypeSpec.InstanceCategory != pb.InstanceCategory_BareMetalHost {
		return &instanceoperatorutil.InstanceNetworkInfo{SuperComputeGroupIDs: []string{}, ClusterGroupIDs: []string{}}, nil
	}

	net, err := getInstanceAssignedNetworkInfo(instances)
	if err != nil {
		return nil, err
	}
	log.Info("Assigned network", "info", net)

	nodeInfos, err := sched.listNodesFromCache(ctx)
	if err != nil {
		return nil, err
	}

	// determine the target network mode to allocate from
	if instance.Spec.InstanceGroup != "" {
		if len(net.SuperComputeGroupIDs) > 0 {
			log.Info("Scheduling for supercompute instance group")
			net.NetworkMode = bmenroll.NetworkModeXBX
		} else if len(net.ClusterGroupIDs) == 0 {
			log.Info("Scheduling for new instance group")
			if sched.bgpNetworkRequiredInstanceCountThreshold > 0 &&
				instance.Spec.InstanceGroupSize >= int32(sched.bgpNetworkRequiredInstanceCountThreshold) {
				net.NetworkMode = bmenroll.NetworkModeXBX
			} else {
				net.NetworkMode = bmenroll.NetworkModeVVV
			}
		}
	} else {
		log.Info("Scheduling a single instance")
		net.NetworkMode = bmenroll.NetworkModeVVV
	}

	if net.NetworkMode == bmenroll.NetworkModeXBX {
		net.SuperComputeGroupIDs, net.ClusterGroupIDs, err = sched.suggestGroupIDs(ctx, nodeInfos, instances, net)
		if err != nil {
			return nil, err
		}
	}
	log.Info("Target network", "info", net)

	return net, nil
}

func (sched *Scheduler) listNodesFromCache(ctx context.Context) ([]*framework.NodeInfo, error) {
	if err := sched.Cache.UpdateSnapshot(ctx, sched.nodeInfoSnapshot); err != nil {
		return nil, err
	}
	nodeInfos, err := sched.nodeInfoSnapshot.NodeInfos().List()
	if err != nil {
		return nil, err
	}
	return nodeInfos, nil
}

func (sched *Scheduler) suggestGroupIDs(ctx context.Context, nodeInfos []*framework.NodeInfo, instances []*pb.InstancePrivate, assigned *instanceoperatorutil.InstanceNetworkInfo) (suggestedSuperComputeGroupIDs, suggestedClusterGroupIDs []string, err error) {
	log := log.FromContext(ctx).WithName("Scheduler.suggestGroupIDs")
	feasibleSuperComputeGroups, feasibleClusterGroups, err := sched.getFeasibleGroups(ctx, nodeInfos, instances, assigned)
	if err != nil {
		return nil, nil, err
	}

	// supercompute is preferred when feasible.
	if len(assigned.SuperComputeGroupIDs) > 0 || len(feasibleSuperComputeGroups.groups) > 0 {
		// Any presence of assigned SC suggestedSC IDs or SC groups implies that SC instance type is being requested.
		// Find the group that is most-allocated and has the capacity to accommodate the instances.
		suggestedSC, err := feasibleSuperComputeGroups.FindGroupWithLeastCap(assigned.SuperComputeGroupIDs)
		if err != nil {
			return nil, nil, err
		}
		suggestedSuperComputeGroupIDs = []string{suggestedSC.id}

		// allocate any available groups within SC
		subGroups := NewClusterGroupInfos([]*v1.Node{})
		subGroups.groups = suggestedSC.subGroups
		suggestedClusterGroupIDs, err = subGroups.FindGroups(len(instances), assigned.ClusterGroupIDs)
		if err != nil {
			return nil, nil, err
		}

	} else if len(assigned.ClusterGroupIDs) > 0 || len(feasibleClusterGroups.groups) > 0 {
		// remove any group that has parent in SC
		// allocate non-SC groups
		suggestedClusterGroupIDs, err = feasibleClusterGroups.FindGroups(len(instances), assigned.ClusterGroupIDs)
		if err != nil {
			return nil, nil, err
		}
	}

	if len(suggestedSuperComputeGroupIDs) == 0 && len(suggestedClusterGroupIDs) == 0 {
		return nil, nil, fmt.Errorf("no suggested groups found")
	}

	log.Info("Suggested Groups", "superComputeGroupIds", suggestedSuperComputeGroupIDs, "groupIds", suggestedClusterGroupIDs)

	return suggestedSuperComputeGroupIDs, suggestedClusterGroupIDs, nil
}

// getFeasibleGroups identifies groups that are feasible for the instances.
func (sched *Scheduler) getFeasibleGroups(ctx context.Context, nodeInfos []*framework.NodeInfo, instances []*pb.InstancePrivate, net *instanceoperatorutil.InstanceNetworkInfo) (*ClusterGroupInfos, *ClusterGroupInfos, error) {
	log := log.FromContext(ctx).WithName("Scheduler.getFeasibleGroups")
	instance := instances[0]
	log.Info("Searching groups", logkeys.InstanceType, instance.Spec.InstanceType, logkeys.NetworkMode, net.NetworkMode, "AssignedSuperComputeGroupIds", net.SuperComputeGroupIDs, logkeys.AssignedGroupIds, net.ClusterGroupIDs)

	assignedMode := net.NetworkMode
	assignedSuperComputeGroups := sets.NewString(net.SuperComputeGroupIDs...)
	assignedGroups := sets.NewString(net.ClusterGroupIDs...)

	allNodes := []*v1.Node{}
	for _, info := range nodeInfos {
		node := info.Node()
		if len(info.Pods) > 0 {
			node.Labels[nodeAssignedLabel] = "true"
		}
		allNodes = append(allNodes, node)
	}

	nodeSelectorTerms := getNodeSelectorTerms(instance, net)
	nodeSelector := &v1.NodeSelector{NodeSelectorTerms: nodeSelectorTerms}

	groupFilters := []ClusterGroupInfosFilter{FilterGroupsWithNetworkMode(assignedMode)}
	groupMinimumCapFilter := FilterGroupsWithMinimumCurrentCap(len(instances))
	if len(assignedSuperComputeGroups) > 0 {
		groupFilters = append(groupFilters, groupMinimumCapFilter)
	}

	opts := []ClusterGroupInfosOption{
		WithNodeSelector(nodeSelector),
		WithGroupFilters(groupFilters...),
		WithGroupIdentifier(SuperComputeGroupIdentifier),
	}

	allGroups := NewClusterGroupInfos(allNodes, opts...)
	superComputeGroups := NewClusterGroupInfos([]*v1.Node{}, opts...)
	clusterGroups := NewClusterGroupInfos([]*v1.Node{}, opts...)

	for id, group := range allGroups.groups {
		log.Info("Evaulating group", group.GetLogKeyValues()...)
		if group.currentCap > 0 {
			// Split groups by type since they are handled differently.
			// Generally, feasible groups are that ones currently assigned or explicitly requested.
			// The groups are selected based on capacity and multi-tenancy requirements.
			switch group.groupType {
			case SuperComputeGroupType:
				// Exclude if the supercompute group is not targeted.
				if assignedSuperComputeGroups.Len() == 0 && assignedGroups.Len() > 0 {
					break
				}
				// An SC group is multi-tenant, but its subgroups are single-tenant.
				if assignedSuperComputeGroups.Has(id) || groupMinimumCapFilter(group) {
					for _, subGroup := range group.subGroups {
						// Exclude any subgroups that are reserved by other tenents unless they are explicitly requested.
						if !assignedGroups.Has(subGroup.id) && subGroup.assigned {
							allGroups.deleteSubGroup(group, subGroup.id)
						}
					}
					// Reevaluate the group capacity after filtering subgroups.
					if group.currentCap > 0 {
						superComputeGroups.groups[id] = group
					}
				}
			case ClusterGroupType:
				// Exclude if the group belongs to a supercompute group.
				if group.parentGroup != nil && group.parentGroup.groupType == SuperComputeGroupType {
					break
				}
				if assignedGroups.Has(id) || !group.assigned {
					clusterGroups.groups[id] = group
				}
			}
		}
	}

	if len(superComputeGroups.groups) == 0 && len(clusterGroups.groups) == 0 {
		return nil, nil, fmt.Errorf("no feasible groups found")
	}

	for _, group := range superComputeGroups.groups {
		log.Info("Feasible group", group.GetLogKeyValues()...)
	}
	for _, group := range clusterGroups.groups {
		log.Info("Feasible group", group.GetLogKeyValues()...)
	}

	return superComputeGroups, clusterGroups, nil
}

func getNodeSelectorTerms(instance *pb.InstancePrivate, net *instanceoperatorutil.InstanceNetworkInfo) []v1.NodeSelectorTerm {
	requirements := []v1.NodeSelectorRequirement{}

	// add requirement to select the specified instance type
	if instance.Spec != nil &&
		instance.Spec.InstanceTypeSpec != nil &&
		instance.Spec.InstanceTypeSpec.Name != "" {
		requirements = append(requirements, v1.NodeSelectorRequirement{
			Key:      fmt.Sprintf(bmenroll.InstanceTypeLabel, strings.ToLower(instance.Spec.InstanceTypeSpec.Name)),
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"true"},
		})
	}

	// add requirement to select a node with the supported firmware versions
	if firmwareVersions, ok := instance.Metadata.Labels[bmenroll.FWVersionLabel]; ok && firmwareVersions != "" {
		versions := strings.Split(firmwareVersions, "_")
		requirements = append(requirements, v1.NodeSelectorRequirement{
			Key:      bmenroll.FWVersionLabel,
			Operator: v1.NodeSelectorOpIn,
			Values:   versions,
		})
	}

	// add requirement to select a node from the specified supercompute group
	if len(net.SuperComputeGroupIDs) > 0 {
		requirements = append(requirements, v1.NodeSelectorRequirement{
			Key:      enroll.SuperComputeGroupID,
			Operator: v1.NodeSelectorOpIn,
			Values:   net.SuperComputeGroupIDs,
		})
	}

	// add requirement to select a node from the allowed node pools
	nodePoolRequirements := []v1.NodeSelectorRequirement{}
	for _, pool := range instance.Spec.ComputeNodePools {
		nodePoolRequirements = append(nodePoolRequirements, v1.NodeSelectorRequirement{
			Key:      instanceoperatorutil.LabelKeyForComputeNodePools(pool),
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"true"},
		})
	}

	terms := []v1.NodeSelectorTerm{{MatchExpressions: requirements}}
	terms = distributeRequirements(nodePoolRequirements, terms)
	return terms
}

// distributeRequirements distributes the disjunct requirements over the existing terms
func distributeRequirements(reqs []v1.NodeSelectorRequirement, terms []v1.NodeSelectorTerm) []v1.NodeSelectorTerm {
	if len(reqs) == 0 || len(terms) == 0 {
		return terms
	}
	newTerms := make([]v1.NodeSelectorTerm, 0, len(reqs)*len(terms))
	for _, req := range reqs {
		for _, term := range terms {
			term.MatchExpressions = append(term.MatchExpressions, req)
			newTerms = append(newTerms, term)
		}
	}
	return newTerms
}

func getInstanceAssignedNetworkInfo(instances []*pb.InstancePrivate) (*instanceoperatorutil.InstanceNetworkInfo, error) {
	instanceTypes := sets.NewString()
	networkModes := sets.NewString()
	clusterGroupIDs := sets.NewString()
	superComputeGroupIDs := sets.NewString()

	for _, instance := range instances {
		instanceTypes.Insert(instance.Spec.InstanceTypeSpec.Name)
		networkModes.Insert(instance.Spec.NetworkMode)
		if instance.Spec.SuperComputeGroupId != "" {
			superComputeGroupIDs.Insert(strings.Split(instance.Spec.SuperComputeGroupId, ",")...)
		}
		if instance.Spec.ClusterGroupId != "" {
			clusterGroupIDs.Insert(strings.Split(instance.Spec.ClusterGroupId, ",")...)
		}
	}

	if instanceTypes.Len() != 1 {
		return nil, fmt.Errorf("instance type must be the same for the instance group")
	}
	if networkModes.Len() != 1 {
		return nil, fmt.Errorf("network mode must be the same for the instance group")
	}

	return &instanceoperatorutil.InstanceNetworkInfo{
		NetworkMode:          networkModes.List()[0],
		SuperComputeGroupIDs: superComputeGroupIDs.List(),
		ClusterGroupIDs:      clusterGroupIDs.List(),
	}, nil
}

// checks the instance for a specific GPU model and adds GPU resources to the pod's container if needed.
func addGpuResources(ctx context.Context, pod *corev1.Pod, instance *pb.InstancePrivate) {
	log := log.FromContext(ctx).WithName("Scheduler.addGpuResources")

	if len(pod.Spec.Containers) == 0 {
		log.Info("No containers in the Pod; nothing to do")
		return
	}

	gpuName, exists := utils.GpuModelToResourceName[instance.Spec.InstanceTypeSpec.Gpu.ModelName]
	if !exists {
		log.Info("GPU model name not found; no resources added")
		return
	}

	gpuResourceName := corev1.ResourceName(gpuName)
	gpuQuantity := resource.MustParse(strconv.FormatInt(int64(instance.Spec.InstanceTypeSpec.Gpu.Count), 10))

	// Directly access the first container assuming there's always exactly one, for simplicity.
	container := &pod.Spec.Containers[0]

	// Initialize the maps if they're nil to avoid nil map assignment error
	if container.Resources.Requests == nil {
		container.Resources.Requests = corev1.ResourceList{}
	}
	if container.Resources.Limits == nil {
		container.Resources.Limits = corev1.ResourceList{}
	}

	// Add or update the GPU resource requests and limits
	container.Resources.Requests[gpuResourceName] = gpuQuantity
	container.Resources.Limits[gpuResourceName] = gpuQuantity

	log.Info("GPU resources added to the container", logkeys.ResourceName, gpuResourceName, logkeys.Quantity, gpuQuantity)
}

// Return the poolIds from the node pool labels whose values are true.
func getNodePoolsIds(nodePoolLabels map[string]string) []string {
	var computePools []string
	for key, value := range nodePoolLabels {
		if strings.HasPrefix(key, instanceoperatorutil.ComputeNodePoolLabelPrefix) && value == "true" {
			// Extract the poolId from the label
			poolId := key[len(instanceoperatorutil.ComputeNodePoolLabelPrefix):]
			computePools = append(computePools, poolId)
		}
	}

	return computePools
}
