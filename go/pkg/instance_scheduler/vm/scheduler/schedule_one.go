// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubernetes 1.24 kube-scheduler (https://github.com/kubernetes/kubernetes/tree/73da4d3652771d6c6dfe904fe8fae594a1a72e2b/pkg/scheduler).
// To see changes made, run diff-kube-scheduler.sh.

/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	bmenroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	instanceoperatorutil "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/parallelize"
	frameworkruntime "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/metrics"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// SchedulerError is the reason recorded for events when an error occurs during scheduling a pod.
	SchedulerError = "SchedulerError"
	// Percentage of plugin metrics to be sampled.
	pluginMetricsSamplePercent = 10
	// minFeasibleNodesToFind is the minimum number of nodes that would be scored
	// in each scheduling cycle. This is a semi-arbitrary value to ensure that a
	// certain minimum of nodes are checked for feasibility. This in turn helps
	// ensure a minimum level of spreading.
	minFeasibleNodesToFind = 100
	// minFeasibleNodesPercentageToFind is the minimum percentage of nodes that
	// would be scored in each scheduling cycle. This is a semi-arbitrary value
	// to ensure that a certain minimum of nodes are checked for feasibility.
	// This in turn helps ensure a minimum level of spreading.
	minFeasibleNodesPercentageToFind = 5
)

var clearNominatedNode = &framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedNodeName: ""}

// scheduleOnePod does the entire scheduling workflow for a single pod.
func (sched *Scheduler) scheduleOnePod(ctx context.Context, pod *v1.Pod, requestedSize int) (ScheduleResult, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("Scheduler.scheduleOnePod").Start()
	defer span.End()
	// Get mutex so that scheduleOnePod only runs one at a time.
	sched.scheduleOneLock.Lock()
	defer sched.scheduleOneLock.Unlock()

	podInfo := &framework.QueuedPodInfo{
		PodInfo: &framework.PodInfo{
			Pod: pod,
		},
		Timestamp:               time.Now(),
		Attempts:                1,
		InitialAttemptTimestamp: time.Now(),
		UnschedulablePlugins:    map[string]sets.Empty{},
	}
	fwk, err := sched.frameworkForPod(ctx, pod)
	if err != nil {
		// This shouldn't happen, because we only accept for scheduling the pods
		// which specify a scheduler name that matches one of the profiles.
		log.Error(err, "Error occurred")
		return ScheduleResult{}, err
	}

	log.V(3).Info("Attempting to schedule pod", logkeys.Pod, pod)

	// Synchronously attempt to find a fit for the pod.
	start := time.Now()
	state := framework.NewCycleState()
	state.SetRecordPluginMetrics(rand.Intn(100) < pluginMetricsSamplePercent)

	schedulingCycleCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer sched.DumpToLog(ctx)
	scheduleResult, err := sched.SchedulePod(schedulingCycleCtx, fwk, state, pod, requestedSize)
	if err != nil {
		log.Error(err, "Error while selecting node for pod", logkeys.Pod, pod)
		var nominatingInfo *framework.NominatingInfo
		reason := v1.PodReasonUnschedulable
		if fitError, ok := err.(*framework.FitError); ok {
			if !fwk.HasPostFilterPlugins() {
				log.V(3).Info("No PostFilter plugins are registered, so no preemption will be performed")
			} else {
				// Run PostFilter plugins to try to make the pod schedulable in a future scheduling cycle.
				result, status := fwk.RunPostFilterPlugins(ctx, state, pod, fitError.Diagnosis.NodeToStatusMap)
				if status.Code() == framework.Error {
					log.Error(nil, "Status after running PostFilter plugins for pod", logkeys.Pod, pod, logkeys.Status, status)
				} else {
					fitError.Diagnosis.PostFilterMsg = status.Message()
					log.V(3).Info("Status after running PostFilter plugins for pod", logkeys.Pod, pod, logkeys.Status, status)
				}
				if result != nil {
					log.V(3).Info("Result after running PostFilter plugins for pod", logkeys.Pod, pod, logkeys.Result, result)
					nominatingInfo = result.NominatingInfo
				}
			}
			// Pod did not fit anywhere, so it is counted as a failure. If preemption
			// succeeds, the pod should get counted as a success the next time we try to
			// schedule it. (hopefully)
			metrics.PodUnschedulable(fwk.ProfileName(), metrics.SinceInSeconds(start))
		} else if err == ErrNoNodesAvailable {
			nominatingInfo = clearNominatedNode
			// No nodes available is counted as unschedulable rather than an error.
			metrics.PodUnschedulable(fwk.ProfileName(), metrics.SinceInSeconds(start))
		} else {
			nominatingInfo = clearNominatedNode
			log.Error(err, "Error selecting node for pod", logkeys.Pod, pod)
			metrics.PodScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
			reason = SchedulerError
		}
		sched.handleSchedulingFailure(ctx, fwk, podInfo, err, reason, nominatingInfo)
		return ScheduleResult{}, err
	}
	metrics.SchedulingAlgorithmLatency.Observe(metrics.SinceInSeconds(start))
	// Tell the cache to assume that a pod now is running on a given node, even though it hasn't been bound yet.
	// This allows us to keep scheduling without waiting on binding to occur.
	assumedPodInfo := podInfo.DeepCopy()
	assumedPod := assumedPodInfo.Pod
	// assume modifies `assumedPod` by setting NodeName=scheduleResult.SuggestedHost
	err = sched.assume(ctx, assumedPod, scheduleResult.SuggestedHost)
	if err != nil {
		log.Error(err, "Error while assumeing node for pod", logkeys.Pod, pod)
		metrics.PodScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
		// This is most probably result of a BUG in retrying logic.
		// We report an error here so that pod scheduling can be retried.
		// This relies on the fact that Error will check if the pod has been bound
		// to a node and if so will not add it back to the unscheduled pods queue
		// (otherwise this would cause an infinite loop).
		sched.handleSchedulingFailure(ctx, fwk, assumedPodInfo, err, SchedulerError, clearNominatedNode)
		return ScheduleResult{}, err
	}

	// Run the Reserve method of reserve plugins.
	if sts := fwk.RunReservePluginsReserve(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost); !sts.IsSuccess() {
		metrics.PodScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
		// trigger un-reserve to clean up state associated with the reserved Pod
		fwk.RunReservePluginsUnreserve(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if forgetErr := sched.Cache.ForgetPod(ctx, assumedPod); forgetErr != nil {
			log.Error(forgetErr, "Scheduler cache ForgetPod failed")
		}
		sched.handleSchedulingFailure(ctx, fwk, assumedPodInfo, sts.AsError(), SchedulerError, clearNominatedNode)
		return ScheduleResult{}, sts.AsError()
	}

	// Finish binding to allow assumed pod to timeout.
	sched.finishBinding(ctx, assumedPod)

	// Calculating nodeResourceString can be heavy. Avoid it if log verbosity is below 2.
	log.V(2).Info("Successfully bound pod to node", logkeys.Pod, pod, logkeys.Node, scheduleResult.SuggestedHost, logkeys.EvaluatedNodes, scheduleResult.EvaluatedNodes, logkeys.FeasibleNodes, scheduleResult.FeasibleNodes)
	metrics.PodScheduled(fwk.ProfileName(), metrics.SinceInSeconds(start))
	metrics.PodSchedulingAttempts.Observe(float64(podInfo.Attempts))
	metrics.PodSchedulingDuration.WithLabelValues(getAttemptsLabel(podInfo)).Observe(metrics.SinceInSeconds(podInfo.InitialAttemptTimestamp))

	return scheduleResult, nil
}

func (sched *Scheduler) frameworkForPod(ctx context.Context, pod *v1.Pod) (framework.Framework, error) {
	log := log.FromContext(ctx).WithName("Scheduler.frameworkForPod")
	fwk, ok := sched.Profiles[pod.Spec.SchedulerName]
	if !ok {
		return nil, fmt.Errorf("profile not found for scheduler name %q", pod.Spec.SchedulerName)
	}
	log.V(3).Info("Framework for pod", logkeys.ProfileName, fwk.ProfileName(), logkeys.SchedulerName, pod.Spec.SchedulerName)
	return fwk, nil
}

// schedulePod tries to schedule the given pod to one of the nodes in the node list.
// If it succeeds, it will return the name of the node.
// If it fails, it will return a FitError with reasons.
func (sched *Scheduler) SchedulePod(ctx context.Context, fwk framework.Framework, state *framework.CycleState, pod *v1.Pod, requestedSize int) (result ScheduleResult, err error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("Scheduler.SchedulePod").Start()
	defer span.End()
	if err := sched.Cache.UpdateSnapshot(ctx, sched.nodeInfoSnapshot); err != nil {
		return result, err
	}

	sched.DumpToLog(ctx)

	if sched.nodeInfoSnapshot.NumNodes() == 0 {
		return result, ErrNoNodesAvailable
	}

	feasibleNodes, diagnosis, err := sched.findNodesThatFitPod(ctx, fwk, state, pod)
	if err != nil {
		return result, err
	}

	if len(feasibleNodes) == 0 {
		return result, &framework.FitError{
			Pod:         pod,
			NumAllNodes: sched.nodeInfoSnapshot.NumNodes(),
			Diagnosis:   diagnosis,
		}
	}

	if sched.enableBinpack {
		feasibleNodes = binpack(ctx, pod, feasibleNodes, requestedSize)
	}

	for _, node := range feasibleNodes {
		log.Info("Feasible node for pod", logkeys.Node, node, logkeys.Pod, pod)
	}

	host := ""
	hostLabels := map[string]string{}
	if len(feasibleNodes) == 1 {
		// When only one node after predicate, just use it.
		host = feasibleNodes[0].Name
		hostLabels = feasibleNodes[0].Labels
	} else {
		// When more than one node is feasible, score nodes.
		priorityList, err := prioritizeNodes(ctx, fwk, state, pod, feasibleNodes)
		if err != nil {
			return result, err
		}
		host, err = selectHost(priorityList)
		if err != nil {
			return result, err
		}
		for i := range feasibleNodes {
			if host == feasibleNodes[i].Name {
				hostLabels = feasibleNodes[i].Labels
				break
			}
		}
	}

	// Include required partition only if topology spread constraints were provided.
	partition := ""
	if len(pod.Spec.TopologySpreadConstraints) > 0 {
		partition = getNodePartition(findNodeWithName(host, feasibleNodes))
	}

	return ScheduleResult{
		SuggestedHost:       host,
		SuggestedHostLabels: hostLabels,
		EvaluatedNodes:      len(feasibleNodes) + len(diagnosis.NodeToStatusMap),
		FeasibleNodes:       len(feasibleNodes),
		Partition:           partition,
	}, err
}

// Filters the nodes to find the ones that fit the pod based on the framework
// filter plugins and filter extenders.
func (sched *Scheduler) findNodesThatFitPod(ctx context.Context, fwk framework.Framework, state *framework.CycleState, pod *v1.Pod) ([]*v1.Node, framework.Diagnosis, error) {
	log := log.FromContext(ctx).WithName("Scheduler.findNodesThatFitPod")
	diagnosis := framework.Diagnosis{
		NodeToStatusMap:      make(framework.NodeToStatusMap),
		UnschedulablePlugins: sets.NewString(),
	}

	// Run "prefilter" plugins.
	preRes, s := fwk.RunPreFilterPlugins(ctx, state, pod)
	allNodes, err := sched.nodeInfoSnapshot.NodeInfos().List()
	if err != nil {
		return nil, diagnosis, err
	}
	if !s.IsSuccess() {
		if !s.IsUnschedulable() {
			return nil, diagnosis, s.AsError()
		}
		// All nodes will have the same status. Some non trivial refactoring is
		// needed to avoid this copy.
		for _, n := range allNodes {
			diagnosis.NodeToStatusMap[n.Node().Name] = s
		}
		// Status satisfying IsUnschedulable() gets injected into diagnosis.UnschedulablePlugins.
		if s.FailedPlugin() != "" {
			diagnosis.UnschedulablePlugins.Insert(s.FailedPlugin())
		}
		return nil, diagnosis, nil
	}

	// "NominatedNodeName" can potentially be set in a previous scheduling cycle as a result of preemption.
	// This node is likely the only candidate that will fit the pod, and hence we try it first before iterating over all nodes.
	if len(pod.Status.NominatedNodeName) > 0 {
		feasibleNodes, err := sched.evaluateNominatedNode(ctx, pod, fwk, state, diagnosis)
		if err != nil {
			log.Error(err, "Evaluation failed on nominated node", logkeys.Pod, pod, logkeys.Node, pod.Status.NominatedNodeName)
		}
		// Nominated node passes all the filters, scheduler is good to assign this node to the pod.
		if len(feasibleNodes) != 0 {
			return feasibleNodes, diagnosis, nil
		}
	}

	nodes := allNodes
	if !preRes.AllNodes() {
		nodes = make([]*framework.NodeInfo, 0, len(preRes.NodeNames))
		for n := range preRes.NodeNames {
			nInfo, err := sched.nodeInfoSnapshot.NodeInfos().Get(n)
			if err != nil {
				return nil, diagnosis, err
			}
			nodes = append(nodes, nInfo)
		}
	}
	feasibleNodes, err := sched.findNodesThatPassFilters(ctx, fwk, state, pod, diagnosis, nodes)
	if err != nil {
		return nil, diagnosis, err
	}

	return feasibleNodes, diagnosis, nil
}

func (sched *Scheduler) evaluateNominatedNode(ctx context.Context, pod *v1.Pod, fwk framework.Framework, state *framework.CycleState, diagnosis framework.Diagnosis) ([]*v1.Node, error) {
	nnn := pod.Status.NominatedNodeName
	nodeInfo, err := sched.nodeInfoSnapshot.Get(nnn)
	if err != nil {
		return nil, err
	}
	node := []*framework.NodeInfo{nodeInfo}
	feasibleNodes, err := sched.findNodesThatPassFilters(ctx, fwk, state, pod, diagnosis, node)
	if err != nil {
		return nil, err
	}

	return feasibleNodes, nil
}

// findNodesThatPassFilters finds the nodes that fit the filter plugins.
func (sched *Scheduler) findNodesThatPassFilters(
	ctx context.Context,
	fwk framework.Framework,
	state *framework.CycleState,
	pod *v1.Pod,
	diagnosis framework.Diagnosis,
	nodes []*framework.NodeInfo) ([]*v1.Node, error) {
	numNodesToFind := sched.numFeasibleNodesToFind(int32(len(nodes)))

	// Create feasible list with enough space to avoid growing it
	// and allow assigning.
	feasibleNodes := make([]*v1.Node, numNodesToFind)

	if !fwk.HasFilterPlugins() {
		length := len(nodes)
		for i := range feasibleNodes {
			feasibleNodes[i] = nodes[(sched.nextStartNodeIndex+i)%length].Node()
		}
		sched.nextStartNodeIndex = (sched.nextStartNodeIndex + len(feasibleNodes)) % length
		return feasibleNodes, nil
	}

	errCh := parallelize.NewErrorChannel()
	var statusesLock sync.Mutex
	var feasibleNodesLen int32
	ctx, cancel := context.WithCancel(ctx)
	checkNode := func(i int) {
		// We check the nodes starting from where we left off in the previous scheduling cycle,
		// this is to make sure all nodes have the same chance of being examined across pods.
		nodeInfo := nodes[(sched.nextStartNodeIndex+i)%len(nodes)]
		status := fwk.RunFilterPluginsWithNominatedPods(ctx, state, pod, nodeInfo)
		if status.Code() == framework.Error {
			errCh.SendErrorWithCancel(status.AsError(), cancel)
			return
		}
		if status.IsSuccess() {
			length := atomic.AddInt32(&feasibleNodesLen, 1)
			if length > numNodesToFind {
				cancel()
				atomic.AddInt32(&feasibleNodesLen, -1)
			} else {
				feasibleNodes[length-1] = nodeInfo.Node()
			}
		} else {
			statusesLock.Lock()
			diagnosis.NodeToStatusMap[nodeInfo.Node().Name] = status
			diagnosis.UnschedulablePlugins.Insert(status.FailedPlugin())
			statusesLock.Unlock()
		}
	}

	beginCheckNode := time.Now()
	statusCode := framework.Success
	defer func() {
		// We record Filter extension point latency here instead of in framework.go because framework.RunFilterPlugins
		// function is called for each node, whereas we want to have an overall latency for all nodes per scheduling cycle.
		// Note that this latency also includes latency for `addNominatedPods`, which calls framework.RunPreFilterAddPod.
		metrics.FrameworkExtensionPointDuration.WithLabelValues(frameworkruntime.Filter, statusCode.String(), fwk.ProfileName()).Observe(metrics.SinceInSeconds(beginCheckNode))
	}()

	// Stops searching for more nodes once the configured number of feasible nodes
	// are found.
	fwk.Parallelizer().Until(ctx, len(nodes), checkNode)
	processedNodes := int(feasibleNodesLen) + len(diagnosis.NodeToStatusMap)
	sched.nextStartNodeIndex = (sched.nextStartNodeIndex + processedNodes) % len(nodes)

	feasibleNodes = feasibleNodes[:feasibleNodesLen]
	if err := errCh.ReceiveError(); err != nil {
		statusCode = framework.Error
		return nil, err
	}
	return feasibleNodes, nil
}

// numFeasibleNodesToFind returns the number of feasible nodes that once found, the scheduler stops
// its search for more feasible nodes.
func (sched *Scheduler) numFeasibleNodesToFind(numAllNodes int32) (numNodes int32) {
	if numAllNodes < minFeasibleNodesToFind || sched.percentageOfNodesToScore >= 100 {
		return numAllNodes
	}

	adaptivePercentage := sched.percentageOfNodesToScore
	if adaptivePercentage <= 0 {
		basePercentageOfNodesToScore := int32(50)
		adaptivePercentage = basePercentageOfNodesToScore - numAllNodes/125
		if adaptivePercentage < minFeasibleNodesPercentageToFind {
			adaptivePercentage = minFeasibleNodesPercentageToFind
		}
	}

	numNodes = numAllNodes * adaptivePercentage / 100
	if numNodes < minFeasibleNodesToFind {
		return minFeasibleNodesToFind
	}

	return numNodes
}

// prioritizeNodes prioritizes the nodes by running the score plugins,
// which return a score for each node from the call to RunScorePlugins().
// The scores from each plugin are added together to make the score for that node, then
// any extenders are run as well.
// All scores are finally combined (added) to get the total weighted scores of all nodes
func prioritizeNodes(
	ctx context.Context,
	fwk framework.Framework,
	state *framework.CycleState,
	pod *v1.Pod,
	nodes []*v1.Node,
) (framework.NodeScoreList, error) {
	log := log.FromContext(ctx).WithName("Scheduler.prioritizeNodes")
	// If no priority configs are provided, then all nodes will have a score of one.
	// This is required to generate the priority list in the required format
	if !fwk.HasScorePlugins() {
		result := make(framework.NodeScoreList, 0, len(nodes))
		for i := range nodes {
			result = append(result, framework.NodeScore{
				Name:  nodes[i].Name,
				Score: 1,
			})
		}
		return result, nil
	}

	// Run PreScore plugins.
	preScoreStatus := fwk.RunPreScorePlugins(ctx, state, pod, nodes)
	if !preScoreStatus.IsSuccess() {
		return nil, preScoreStatus.AsError()
	}

	// Run the Score plugins.
	scoresMap, scoreStatus := fwk.RunScorePlugins(ctx, state, pod, nodes)
	if !scoreStatus.IsSuccess() {
		return nil, scoreStatus.AsError()
	}

	// Additional details logged at level 10 if enabled.
	logV := log.V(3)
	if logV.Enabled() {
		for plugin, nodeScoreList := range scoresMap {
			for _, nodeScore := range nodeScoreList {
				logV.Info("Plugin scored node for pod", "pod", pod, "plugin", plugin, "node", nodeScore.Name, "score", nodeScore.Score)
			}
		}
	}

	// Summarize all scores.
	result := make(framework.NodeScoreList, 0, len(nodes))

	for i := range nodes {
		result = append(result, framework.NodeScore{Name: nodes[i].Name, Score: 0})
		for j := range scoresMap {
			result[i].Score += scoresMap[j][i].Score
		}
	}

	if logV.Enabled() {
		for i := range result {
			logV.Info("Calculated node's final score for pod", "pod", pod, "node", result[i].Name, "score", result[i].Score)
		}
	}
	return result, nil
}

// selectHost takes a prioritized list of nodes and then picks one
// in a reservoir sampling manner from the nodes that had the highest score.
func selectHost(nodeScoreList framework.NodeScoreList) (string, error) {
	if len(nodeScoreList) == 0 {
		return "", fmt.Errorf("empty priorityList")
	}
	maxScore := nodeScoreList[0].Score
	selected := nodeScoreList[0].Name
	cntOfMaxScore := 1
	for _, ns := range nodeScoreList[1:] {
		if ns.Score > maxScore {
			maxScore = ns.Score
			selected = ns.Name
			cntOfMaxScore = 1
		} else if ns.Score == maxScore {
			cntOfMaxScore++
			if rand.Intn(cntOfMaxScore) == 0 {
				// Replace the candidate with probability of 1/cntOfMaxScore
				selected = ns.Name
			}
		}
	}
	return selected, nil
}

// Resources are reserved for a period of time after a scheduling request is processed successfully.
// This is referred as "assuming" a pod.
// The rest of the process (Compute API Server, Instance Replicator, VM Instance Operator, Kubevirt)
// has this much time to create the pod with guaranteed resources on the recommended node.
// If a pod is not created by this time, the resources will be unreserved and made available to other instances.
// assume signals to the cache that a pod is already in the cache, so that binding can be asynchronous.
// assume modifies `assumed`.
func (sched *Scheduler) assume(ctx context.Context, assumed *v1.Pod, host string) error {
	log := log.FromContext(ctx).WithName("Scheduler.assume")
	assumed.Spec.NodeName = host

	if err := sched.Cache.AssumePod(ctx, assumed); err != nil {
		log.Error(err, "Scheduler cache AssumePod failed")
		return err
	}

	return nil
}

func (sched *Scheduler) unassumePods(ctx context.Context, assumed []*corev1.Pod) {
	log := log.FromContext(ctx).WithName("Scheduler.unassumePods")
	log.Info("Unreserving assumed pods", "count", len(assumed))

	for _, pod := range assumed {
		if err := sched.Cache.ForgetPod(ctx, pod); err != nil {
			log.Error(err, "Scheduler cache ForgetPod failed")
		}
	}
}

func (sched *Scheduler) finishBinding(ctx context.Context, assumed *v1.Pod) {
	log := log.FromContext(ctx).WithName("Scheduler.finishBinding")
	if err := sched.Cache.FinishBinding(ctx, assumed); err != nil {
		log.Error(err, "Scheduler cache FinishBinding failed")
	}
}

func getAttemptsLabel(p *framework.QueuedPodInfo) string {
	// We breakdown the pod scheduling duration by attempts capped to a limit
	// to avoid ending up with a high cardinality metric.
	if p.Attempts >= 15 {
		return "15+"
	}
	return strconv.Itoa(p.Attempts)
}

// handleSchedulingFailure records an event for the pod that indicates the
// pod has failed to schedule. Also, update the pod condition and nominated node name if set.
func (sched *Scheduler) handleSchedulingFailure(ctx context.Context, fwk framework.Framework, podInfo *framework.QueuedPodInfo, err error, reason string, nominatingInfo *framework.NominatingInfo) {
	// No need to take any action here. The error will be returned to the caller of the Schedule RPC.
	log := log.FromContext(ctx).WithName("Scheduler.handleSchedulingFailure")
	log.Error(err, "scheduling failure")
}

func (sched *Scheduler) DumpToLog(ctx context.Context) {
	sched.Cache.DumpToLog(ctx)
}

func getNodePartition(node *v1.Node) string {
	if node == nil || node.Labels == nil {
		return ""
	}
	return node.Labels[instanceoperatorutil.TopologySpreadTopologyKey]
}

func findNodeWithName(name string, nodes []*v1.Node) *v1.Node {
	for _, node := range nodes {
		if node.Name == name {
			return node
		}
	}
	return nil
}

// binpack picks a node based on the pod's requirements and the nodes' capacities.
// The nodes are expected to have been filterred by scheduling plugins (e.g. pod affinities).
//
// For group scheduling, it selects a node within a group that has:
//  1. a sufficient capacity for the requested size.
//  2. the least number of available nodes. One is randomly selected in case
//     there are multiple groups with the least number of available nodes.
//
// For single scheduling, the single nodes are preferred if available.
// Otherwise, the group scheduling logics will be applied to select a node from a group.
func binpack(ctx context.Context, pod *v1.Pod, nodes []*v1.Node, requestedSize int) []*v1.Node {
	log := log.FromContext(ctx).WithName("binpack")
	// apply to BM nodes only
	if v, ok := pod.Labels["instance-category"]; ok &&
		v != string(cloudv1alpha1.InstanceCategoryBareMetalHost) {
		return nodes
	}

	for _, node := range nodes {
		log.V(3).Info("Feasible node for pod", logkeys.Node, node, logkeys.Pod, pod)
	}

	if v, ok := pod.Labels[bmenroll.NetworkModeLabel]; ok && v == bmenroll.NetworkModeXBX {
		return nodes
	}

	// get single node info
	var singles []*v1.Node
	for _, node := range nodes {
		if _, ok := node.Labels[bmenroll.ClusterGroupID]; !ok {
			singles = append(singles, node)
		}
	}

	// single nodes are found only when scheduling a non-group instance
	// in that case, they will be selected first.
	if len(singles) > 0 {
		return singles
	}

	// get cluster group info
	groupInfos := NewClusterGroupInfos(nodes,
		WithGroupFilters(FilterGroupsWithMinimumCurrentCap(requestedSize)),
	)
	groups := groupInfos.groups

	for _, group := range groups {
		log.V(3).Info("Feasible group", group.GetLogKeyValues()...)
	}

	selectedNodes := []*v1.Node{}
	if len(groups) > 0 {
		selectedGroup, err := groupInfos.FindGroupWithLeastCap([]string{})
		if err != nil {
			return selectedNodes
		}
		for _, node := range nodes {
			if id, ok := node.Labels[bmenroll.ClusterGroupID]; ok {
				if id == selectedGroup.id {
					selectedNodes = append(selectedNodes, node)
				}
			}
		}
		return selectedNodes
	}

	return selectedNodes
}
