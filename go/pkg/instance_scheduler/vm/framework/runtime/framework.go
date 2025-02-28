// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubernetes 1.24 kube-scheduler (https://github.com/kubernetes/kubernetes/tree/73da4d3652771d6c6dfe904fe8fae594a1a72e2b/pkg/scheduler).
// To see changes made, run diff-kube-scheduler.sh.

/*
Copyright 2019 The Kubernetes Authors.

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

package runtime

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/parallelize"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// Filter is the name of the filter extension point.
	Filter                      = "Filter"
	preFilter                   = "PreFilter"
	preFilterExtensionAddPod    = "PreFilterExtensionAddPod"
	preFilterExtensionRemovePod = "PreFilterExtensionRemovePod"
	postFilter                  = "PostFilter"
	preScore                    = "PreScore"
	score                       = "Score"
	scoreExtensionNormalize     = "ScoreExtensionNormalize"
	reserve                     = "Reserve"
	unreserve                   = "Unreserve"
)

// frameworkImpl is the component responsible for initializing and running scheduler
// plugins.
type frameworkImpl struct {
	registry             Registry
	snapshotSharedLister framework.SharedLister
	scorePluginWeight    map[string]int
	preFilterPlugins     []framework.PreFilterPlugin
	filterPlugins        []framework.FilterPlugin
	postFilterPlugins    []framework.PostFilterPlugin
	preScorePlugins      []framework.PreScorePlugin
	scorePlugins         []framework.ScorePlugin
	reservePlugins       []framework.ReservePlugin

	metricsRecorder *metricsRecorder
	profileName     string

	parallelizer parallelize.Parallelizer
}

// extensionPoint encapsulates desired and applied set of plugins at a specific extension
// point. This is used to simplify iterating over all extension points supported by the
// frameworkImpl.
type extensionPoint struct {
	// the set of plugins to be configured at this extension point.
	plugins *config.PluginSet
	// a pointer to the slice storing plugins implementations that will run at this
	// extension point.
	slicePtr interface{}
}

func (f *frameworkImpl) getExtensionPoints(plugins *config.Plugins) []extensionPoint {
	return []extensionPoint{
		{&plugins.PreFilter, &f.preFilterPlugins},
		{&plugins.Filter, &f.filterPlugins},
		{&plugins.PostFilter, &f.postFilterPlugins},
		{&plugins.Reserve, &f.reservePlugins},
		{&plugins.PreScore, &f.preScorePlugins},
		{&plugins.Score, &f.scorePlugins},
	}
}

type frameworkOptions struct {
	snapshotSharedLister framework.SharedLister
	metricsRecorder      *metricsRecorder
	parallelizer         parallelize.Parallelizer
}

// Option for the frameworkImpl.
type Option func(*frameworkOptions)

// WithSnapshotSharedLister sets the SharedLister of the snapshot.
func WithSnapshotSharedLister(snapshotSharedLister framework.SharedLister) Option {
	return func(o *frameworkOptions) {
		o.snapshotSharedLister = snapshotSharedLister
	}
}

// WithParallelism sets parallelism for the scheduling frameworkImpl.
func WithParallelism(parallelism int) Option {
	return func(o *frameworkOptions) {
		o.parallelizer = parallelize.NewParallelizer(parallelism)
	}
}

func defaultFrameworkOptions() frameworkOptions {
	return frameworkOptions{
		metricsRecorder: newMetricsRecorder(1000, time.Second),
		parallelizer:    parallelize.NewParallelizer(parallelize.DefaultParallelism),
	}
}

var _ framework.Framework = &frameworkImpl{}

// NewFramework initializes plugins given the configuration and the registry.
func NewFramework(r Registry, profile *config.KubeSchedulerProfile, opts ...Option) (framework.Framework, error) {
	ctx := context.Background()
	options := defaultFrameworkOptions()
	for _, opt := range opts {
		opt(&options)
	}

	f := &frameworkImpl{
		registry:             r,
		snapshotSharedLister: options.snapshotSharedLister,
		scorePluginWeight:    make(map[string]int),
		metricsRecorder:      options.metricsRecorder,
		parallelizer:         options.parallelizer,
	}

	if profile == nil {
		return f, nil
	}

	f.profileName = profile.SchedulerName
	if profile.Plugins == nil {
		return f, nil
	}

	// get needed plugins from config
	pg := f.pluginsNeeded(profile.Plugins)

	pluginConfig := make(map[string]runtime.Object, len(profile.PluginConfig))
	for i := range profile.PluginConfig {
		name := profile.PluginConfig[i].Name
		if _, ok := pluginConfig[name]; ok {
			return nil, fmt.Errorf("repeated config for plugin %s", name)
		}
		pluginConfig[name] = profile.PluginConfig[i].Args
	}
	outputProfile := config.KubeSchedulerProfile{
		SchedulerName: f.profileName,
		Plugins:       profile.Plugins,
		PluginConfig:  make([]config.PluginConfig, 0, len(pg)),
	}

	pluginsMap := make(map[string]framework.Plugin)
	for name, factory := range r {
		// initialize only needed plugins.
		if _, ok := pg[name]; !ok {
			continue
		}

		args := pluginConfig[name]
		if args != nil {
			outputProfile.PluginConfig = append(outputProfile.PluginConfig, config.PluginConfig{
				Name: name,
				Args: args,
			})
		}
		p, err := factory(args, f)
		if err != nil {
			return nil, fmt.Errorf("initializing plugin %q: %w", name, err)
		}
		pluginsMap[name] = p
	}

	// initialize plugins per individual extension points
	for _, e := range f.getExtensionPoints(profile.Plugins) {
		if err := updatePluginList(e.slicePtr, *e.plugins, pluginsMap); err != nil {
			return nil, err
		}
	}

	// initialize multiPoint plugins to their expanded extension points
	if len(profile.Plugins.MultiPoint.Enabled) > 0 {
		if err := f.expandMultiPointPlugins(ctx, profile, pluginsMap); err != nil {
			return nil, err
		}
	}

	if err := getScoreWeights(f, pluginsMap, append(profile.Plugins.Score.Enabled, profile.Plugins.MultiPoint.Enabled...)); err != nil {
		return nil, err
	}

	// Verifying the score weights again since Plugin.Name() could return a different
	// value from the one used in the configuration.
	for _, scorePlugin := range f.scorePlugins {
		if f.scorePluginWeight[scorePlugin.Name()] == 0 {
			return nil, fmt.Errorf("score plugin %q is not configured with weight", scorePlugin.Name())
		}
	}

	return f, nil
}

// getScoreWeights makes sure that, between MultiPoint-Score plugin weights and individual Score
// plugin weights there is not an overflow of MaxTotalScore.
func getScoreWeights(f *frameworkImpl, pluginsMap map[string]framework.Plugin, plugins []config.Plugin) error {
	var totalPriority int64
	scorePlugins := reflect.ValueOf(&f.scorePlugins).Elem()
	pluginType := scorePlugins.Type().Elem()
	for _, e := range plugins {
		pg := pluginsMap[e.Name]
		if !reflect.TypeOf(pg).Implements(pluginType) {
			continue
		}

		// We append MultiPoint plugins to the list of Score plugins. So if this plugin has already been
		// encountered, let the individual Score weight take precedence.
		if _, ok := f.scorePluginWeight[e.Name]; ok {
			continue
		}
		// a weight of zero is not permitted, plugins can be disabled explicitly
		// when configured.
		f.scorePluginWeight[e.Name] = int(e.Weight)
		if f.scorePluginWeight[e.Name] == 0 {
			f.scorePluginWeight[e.Name] = 1
		}

		// Checks totalPriority against MaxTotalScore to avoid overflow
		if int64(f.scorePluginWeight[e.Name])*framework.MaxNodeScore > framework.MaxTotalScore-totalPriority {
			return fmt.Errorf("total score of Score plugins could overflow")
		}
		totalPriority += int64(f.scorePluginWeight[e.Name]) * framework.MaxNodeScore
	}
	return nil
}

type orderedSet struct {
	set         map[string]int
	list        []string
	deletionCnt int
}

func newOrderedSet() *orderedSet {
	return &orderedSet{set: make(map[string]int)}
}

func (os *orderedSet) insert(s string) {
	if os.has(s) {
		return
	}
	os.set[s] = len(os.list)
	os.list = append(os.list, s)
}

func (os *orderedSet) has(s string) bool {
	_, found := os.set[s]
	return found
}

func (os *orderedSet) delete(s string) {
	if i, found := os.set[s]; found {
		delete(os.set, s)
		os.list = append(os.list[:i-os.deletionCnt], os.list[i+1-os.deletionCnt:]...)
		os.deletionCnt++
	}
}

func (f *frameworkImpl) expandMultiPointPlugins(ctx context.Context, profile *config.KubeSchedulerProfile, pluginsMap map[string]framework.Plugin) error {
	log := log.FromContext(ctx).WithName("frameworkImpl.expandMultiPointPlugins")
	// initialize MultiPoint plugins
	for _, e := range f.getExtensionPoints(profile.Plugins) {
		plugins := reflect.ValueOf(e.slicePtr).Elem()
		pluginType := plugins.Type().Elem()
		// build enabledSet of plugins already registered via normal extension points
		// to check double registration
		enabledSet := newOrderedSet()
		for _, plugin := range e.plugins.Enabled {
			enabledSet.insert(plugin.Name)
		}

		disabledSet := sets.NewString()
		for _, disabledPlugin := range e.plugins.Disabled {
			disabledSet.Insert(disabledPlugin.Name)
		}
		if disabledSet.Has("*") {
			log.V(3).Info("all plugins disabled for extension point, skipping MultiPoint expansion", logkeys.Extension, pluginType)
			continue
		}

		// track plugins enabled via multipoint separately from those enabled by specific extensions,
		// so that we can distinguish between double-registration and explicit overrides
		multiPointEnabled := newOrderedSet()
		overridePlugins := newOrderedSet()
		for _, ep := range profile.Plugins.MultiPoint.Enabled {
			pg, ok := pluginsMap[ep.Name]
			if !ok {
				return fmt.Errorf("%s %q does not exist", pluginType.Name(), ep.Name)
			}

			// if this plugin doesn't implement the type for the current extension we're trying to expand, skip
			if !reflect.TypeOf(pg).Implements(pluginType) {
				continue
			}

			// a plugin that's enabled via MultiPoint can still be disabled for specific extension points
			if disabledSet.Has(ep.Name) {
				log.V(3).Info("plugin disabled for extension point", logkeys.Plugin, ep.Name, logkeys.Extension, pluginType)
				continue
			}

			// if this plugin has already been enabled by the specific extension point,
			// the user intent is to override the default plugin or make some other explicit setting.
			// Either way, discard the MultiPoint value for this plugin.
			// This maintains expected behavior for overriding default plugins (see https://github.com/kubernetes/kubernetes/pull/99582)
			if enabledSet.has(ep.Name) {
				overridePlugins.insert(ep.Name)
				log.Info("MultiPoint plugin is explicitly re-configured; overriding", logkeys.Plugin, ep.Name)
				continue
			}

			// if this plugin is already registered via MultiPoint, then this is
			// a double registration and an error in the config.
			if multiPointEnabled.has(ep.Name) {
				return fmt.Errorf("plugin %q already registered as %q", ep.Name, pluginType.Name())
			}

			// we only need to update the multipoint set, since we already have the specific extension set from above
			multiPointEnabled.insert(ep.Name)
		}

		// Reorder plugins. Here is the expected order:
		// - part 1: overridePlugins. Their order stay intact as how they're specified in regular extension point.
		// - part 2: multiPointEnabled - i.e., plugin defined in multipoint but not in regular extension point.
		// - part 3: other plugins (excluded by part 1 & 2) in regular extension point.
		newPlugins := reflect.New(reflect.TypeOf(e.slicePtr).Elem()).Elem()
		// part 1
		for _, name := range enabledSet.list {
			if overridePlugins.has(name) {
				newPlugins = reflect.Append(newPlugins, reflect.ValueOf(pluginsMap[name]))
				enabledSet.delete(name)
			}
		}
		// part 2
		for _, name := range multiPointEnabled.list {
			newPlugins = reflect.Append(newPlugins, reflect.ValueOf(pluginsMap[name]))
		}
		// part 3
		for _, name := range enabledSet.list {
			newPlugins = reflect.Append(newPlugins, reflect.ValueOf(pluginsMap[name]))
		}
		plugins.Set(newPlugins)
	}
	return nil
}

func updatePluginList(pluginList interface{}, pluginSet config.PluginSet, pluginsMap map[string]framework.Plugin) error {
	plugins := reflect.ValueOf(pluginList).Elem()
	pluginType := plugins.Type().Elem()
	set := sets.NewString()
	for _, ep := range pluginSet.Enabled {
		pg, ok := pluginsMap[ep.Name]
		if !ok {
			return fmt.Errorf("%s %q does not exist", pluginType.Name(), ep.Name)
		}

		if !reflect.TypeOf(pg).Implements(pluginType) {
			return fmt.Errorf("plugin %q does not extend %s plugin", ep.Name, pluginType.Name())
		}

		if set.Has(ep.Name) {
			return fmt.Errorf("plugin %q already registered as %q", ep.Name, pluginType.Name())
		}

		set.Insert(ep.Name)

		newPlugins := reflect.Append(plugins, reflect.ValueOf(pg))
		plugins.Set(newPlugins)
	}
	return nil
}

// RunPreFilterPlugins runs the set of configured PreFilter plugins. It returns
// *Status and its code is set to non-success if any of the plugins returns
// anything but Success. If a non-success status is returned, then the scheduling
// cycle is aborted.
func (f *frameworkImpl) RunPreFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod) (_ *framework.PreFilterResult, status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(preFilter, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	var result *framework.PreFilterResult
	var pluginsWithNodes []string
	for _, pl := range f.preFilterPlugins {
		r, s := f.runPreFilterPlugin(ctx, pl, state, pod)
		if !s.IsSuccess() {
			s.SetFailedPlugin(pl.Name())
			if s.IsUnschedulable() {
				return nil, s
			}
			return nil, framework.AsStatus(fmt.Errorf("running PreFilter plugin %q: %w", pl.Name(), status.AsError())).WithFailedPlugin(pl.Name())
		}
		if !r.AllNodes() {
			pluginsWithNodes = append(pluginsWithNodes, pl.Name())
		}
		result = result.Merge(r)
		if !result.AllNodes() && len(result.NodeNames) == 0 {
			msg := fmt.Sprintf("node(s) didn't satisfy plugin(s) %v simultaneously", pluginsWithNodes)
			if len(pluginsWithNodes) == 1 {
				msg = fmt.Sprintf("node(s) didn't satisfy plugin %v", pluginsWithNodes[0])
			}
			return nil, framework.NewStatus(framework.Unschedulable, msg)
		}

	}
	return result, nil
}

func (f *frameworkImpl) runPreFilterPlugin(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState, pod *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreFilter(ctx, state, pod)
	}
	startTime := time.Now()
	result, status := pl.PreFilter(ctx, state, pod)
	f.metricsRecorder.observePluginDurationAsync(preFilter, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return result, status
}

// RunPreFilterExtensionAddPod calls the AddPod interface for the set of configured
// PreFilter plugins. It returns directly if any of the plugins return any
// status other than Success.
func (f *frameworkImpl) RunPreFilterExtensionAddPod(
	ctx context.Context,
	state *framework.CycleState,
	podToSchedule *v1.Pod,
	podInfoToAdd *framework.PodInfo,
	nodeInfo *framework.NodeInfo,
) (status *framework.Status) {
	log := log.FromContext(ctx).WithName("frameworkImpl.RunPreFilterExtensionAddPod")

	for _, pl := range f.preFilterPlugins {
		if pl.PreFilterExtensions() == nil {
			continue
		}
		status = f.runPreFilterExtensionAddPod(ctx, pl, state, podToSchedule, podInfoToAdd, nodeInfo)
		if !status.IsSuccess() {
			err := status.AsError()
			log.Error(err, "Failed running AddPod on PreFilter plugin", logkeys.Plugin, pl.Name(), logkeys.Pod, podToSchedule)
			return framework.AsStatus(fmt.Errorf("running AddPod on PreFilter plugin %q: %w", pl.Name(), err))
		}
	}

	return nil
}

func (f *frameworkImpl) runPreFilterExtensionAddPod(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState, podToSchedule *v1.Pod, podInfoToAdd *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreFilterExtensions().AddPod(ctx, state, podToSchedule, podInfoToAdd, nodeInfo)
	}
	startTime := time.Now()
	status := pl.PreFilterExtensions().AddPod(ctx, state, podToSchedule, podInfoToAdd, nodeInfo)
	f.metricsRecorder.observePluginDurationAsync(preFilterExtensionAddPod, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunPreFilterExtensionRemovePod calls the RemovePod interface for the set of configured
// PreFilter plugins. It returns directly if any of the plugins return any
// status other than Success.
func (f *frameworkImpl) RunPreFilterExtensionRemovePod(
	ctx context.Context,
	state *framework.CycleState,
	podToSchedule *v1.Pod,
	podInfoToRemove *framework.PodInfo,
	nodeInfo *framework.NodeInfo,
) (status *framework.Status) {
	log := log.FromContext(ctx).WithName("frameworkImpl.RunPreFilterExtensionRemovePod")
	for _, pl := range f.preFilterPlugins {
		if pl.PreFilterExtensions() == nil {
			continue
		}
		status = f.runPreFilterExtensionRemovePod(ctx, pl, state, podToSchedule, podInfoToRemove, nodeInfo)
		if !status.IsSuccess() {
			err := status.AsError()
			log.Error(err, "Failed running RemovePod on PreFilter plugin", logkeys.Plugin, pl.Name(), logkeys.Pod, podToSchedule)
			return framework.AsStatus(fmt.Errorf("running RemovePod on PreFilter plugin %q: %w", pl.Name(), err))
		}
	}

	return nil
}

func (f *frameworkImpl) runPreFilterExtensionRemovePod(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState, podToSchedule *v1.Pod, podInfoToRemove *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreFilterExtensions().RemovePod(ctx, state, podToSchedule, podInfoToRemove, nodeInfo)
	}
	startTime := time.Now()
	status := pl.PreFilterExtensions().RemovePod(ctx, state, podToSchedule, podInfoToRemove, nodeInfo)
	f.metricsRecorder.observePluginDurationAsync(preFilterExtensionRemovePod, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunFilterPlugins runs the set of configured Filter plugins for pod on
// the given node. If any of these plugins doesn't return "Success", the
// given node is not suitable for running pod.
// Meanwhile, the failure message and status are set for the given node.
func (f *frameworkImpl) RunFilterPlugins(
	ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	nodeInfo *framework.NodeInfo,
) framework.PluginToStatus {
	statuses := make(framework.PluginToStatus)
	for _, pl := range f.filterPlugins {
		pluginStatus := f.runFilterPlugin(ctx, pl, state, pod, nodeInfo)
		if !pluginStatus.IsSuccess() {
			if !pluginStatus.IsUnschedulable() {
				// Filter plugins are not supposed to return any status other than
				// Success or Unschedulable.
				pluginStatus = framework.AsStatus(fmt.Errorf("running %q filter plugin: %w", pl.Name(), pluginStatus.AsError()))
			}
			pluginStatus.SetFailedPlugin(pl.Name())
			return map[string]*framework.Status{pl.Name(): pluginStatus}
		}
	}

	return statuses
}

func (f *frameworkImpl) runFilterPlugin(ctx context.Context, pl framework.FilterPlugin, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.Filter(ctx, state, pod, nodeInfo)
	}
	startTime := time.Now()
	status := pl.Filter(ctx, state, pod, nodeInfo)
	f.metricsRecorder.observePluginDurationAsync(Filter, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunPostFilterPlugins runs the set of configured PostFilter plugins until the first
// Success or Error is met, otherwise continues to execute all plugins.
func (f *frameworkImpl) RunPostFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (_ *framework.PostFilterResult, status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(postFilter, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()

	statuses := make(framework.PluginToStatus)
	// `result` records the last meaningful(non-noop) PostFilterResult.
	var result *framework.PostFilterResult
	for _, pl := range f.postFilterPlugins {
		r, s := f.runPostFilterPlugin(ctx, pl, state, pod, filteredNodeStatusMap)
		if s.IsSuccess() {
			return r, s
		} else if !s.IsUnschedulable() {
			// Any status other than Success or Unschedulable is Error.
			return nil, framework.AsStatus(s.AsError())
		} else if r != nil && r.Mode() != framework.ModeNoop {
			result = r
		}
		statuses[pl.Name()] = s
	}

	return result, statuses.Merge()
}

func (f *frameworkImpl) runPostFilterPlugin(ctx context.Context, pl framework.PostFilterPlugin, state *framework.CycleState, pod *v1.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PostFilter(ctx, state, pod, filteredNodeStatusMap)
	}
	startTime := time.Now()
	r, s := pl.PostFilter(ctx, state, pod, filteredNodeStatusMap)
	f.metricsRecorder.observePluginDurationAsync(postFilter, pl.Name(), s, metrics.SinceInSeconds(startTime))
	return r, s
}

// RunFilterPluginsWithNominatedPods runs the set of configured filter plugins
// for nominated pod on the given node.
// This function is called from two different places: Schedule and Preempt.
// When it is called from Schedule, we want to test whether the pod is
// schedulable on the node with all the existing pods on the node plus higher
// and equal priority pods nominated to run on the node.
// When it is called from Preempt, we should remove the victims of preemption
// and add the nominated pods. Removal of the victims is done by
// SelectVictimsOnNode(). Preempt removes victims from PreFilter state and
// NodeInfo before calling this function.
func (f *frameworkImpl) RunFilterPluginsWithNominatedPods(ctx context.Context, state *framework.CycleState, pod *v1.Pod, info *framework.NodeInfo) *framework.Status {
	var status *framework.Status

	podsAdded := false
	// We run filters twice in some cases. If the node has greater or equal priority
	// nominated pods, we run them when those pods are added to PreFilter state and nodeInfo.
	// If all filters succeed in this pass, we run them again when these
	// nominated pods are not added. This second pass is necessary because some
	// filters such as inter-pod affinity may not pass without the nominated pods.
	// If there are no nominated pods for the node or if the first run of the
	// filters fail, we don't run the second pass.
	// We consider only equal or higher priority pods in the first pass, because
	// those are the current "pod" must yield to them and not take a space opened
	// for running them. It is ok if the current "pod" take resources freed for
	// lower priority pods.
	// Requiring that the new pod is schedulable in both circumstances ensures that
	// we are making a conservative decision: filters like resources and inter-pod
	// anti-affinity are more likely to fail when the nominated pods are treated
	// as running, while filters like pod affinity are more likely to fail when
	// the nominated pods are treated as not running. We can't just assume the
	// nominated pods are running because they are not running right now and in fact,
	// they may end up getting scheduled to a different node.
	for i := 0; i < 2; i++ {
		stateToUse := state
		nodeInfoToUse := info
		if i == 0 {
			var err error
			podsAdded, stateToUse, nodeInfoToUse, err = addNominatedPods(ctx, f, pod, state, info)
			if err != nil {
				return framework.AsStatus(err)
			}
		} else if !podsAdded || !status.IsSuccess() {
			break
		}

		statusMap := f.RunFilterPlugins(ctx, stateToUse, pod, nodeInfoToUse)
		status = statusMap.Merge()
		if !status.IsSuccess() && !status.IsUnschedulable() {
			return status
		}
	}

	return status
}

// addNominatedPods adds pods with equal or greater priority which are nominated
// to run on the node. It returns 1) whether any pod was added, 2) augmented cycleState,
// 3) augmented nodeInfo.
func addNominatedPods(ctx context.Context, fh framework.Handle, pod *v1.Pod, state *framework.CycleState, nodeInfo *framework.NodeInfo) (bool, *framework.CycleState, *framework.NodeInfo, error) {
	return false, state, nodeInfo, nil
}

// RunPreScorePlugins runs the set of configured pre-score plugins. If any
// of these plugins returns any status other than "Success", the given pod is rejected.
func (f *frameworkImpl) RunPreScorePlugins(
	ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	nodes []*v1.Node,
) (status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(preScore, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	for _, pl := range f.preScorePlugins {
		status = f.runPreScorePlugin(ctx, pl, state, pod, nodes)
		if !status.IsSuccess() {
			return framework.AsStatus(fmt.Errorf("running PreScore plugin %q: %w", pl.Name(), status.AsError()))
		}
	}

	return nil
}

func (f *frameworkImpl) runPreScorePlugin(ctx context.Context, pl framework.PreScorePlugin, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreScore(ctx, state, pod, nodes)
	}
	startTime := time.Now()
	status := pl.PreScore(ctx, state, pod, nodes)
	f.metricsRecorder.observePluginDurationAsync(preScore, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunScorePlugins runs the set of configured scoring plugins. It returns a list that
// stores for each scoring plugin name the corresponding NodeScoreList(s).
// It also returns *Status, which is set to non-success if any of the plugins returns
// a non-success status.
func (f *frameworkImpl) RunScorePlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) (ps framework.PluginToNodeScores, status *framework.Status) {
	log := log.FromContext(ctx).WithName("frameworkImpl.RunScorePlugins")
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(score, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	pluginToNodeScores := make(framework.PluginToNodeScores, len(f.scorePlugins))
	for _, pl := range f.scorePlugins {
		pluginToNodeScores[pl.Name()] = make(framework.NodeScoreList, len(nodes))
	}
	ctx, cancel := context.WithCancel(ctx)
	errCh := parallelize.NewErrorChannel()

	// Run Score method for each node in parallel.
	f.Parallelizer().Until(ctx, len(nodes), func(index int) {
		for _, pl := range f.scorePlugins {
			nodeName := nodes[index].Name
			s, status := f.runScorePlugin(ctx, pl, state, pod, nodeName)
			if !status.IsSuccess() {
				err := fmt.Errorf("plugin %q failed with: %w", pl.Name(), status.AsError())
				errCh.SendErrorWithCancel(err, cancel)
				return
			}
			pluginToNodeScores[pl.Name()][index] = framework.NodeScore{
				Name:  nodeName,
				Score: s,
			}
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		return nil, framework.AsStatus(fmt.Errorf("running Score plugins: %w", err))
	}

	logCurrentEnv(ctx, pluginToNodeScores, "before-normalize-score")
	// Run NormalizeScore method for each ScorePlugin in parallel.
	f.Parallelizer().Until(ctx, len(f.scorePlugins), func(index int) {
		pl := f.scorePlugins[index]
		nodeScoreList := pluginToNodeScores[pl.Name()]
		if pl.ScoreExtensions() == nil {
			return
		}
		status := f.runScoreExtension(ctx, pl, state, pod, nodeScoreList)
		if !status.IsSuccess() {
			err := fmt.Errorf("plugin %q failed with: %w", pl.Name(), status.AsError())
			errCh.SendErrorWithCancel(err, cancel)
			return
		}
	})
	logCurrentEnv(ctx, pluginToNodeScores, "after-normalize-beforewights-score")
	if err := errCh.ReceiveError(); err != nil {
		return nil, framework.AsStatus(fmt.Errorf("running Normalize on Score plugins: %w", err))
	}

	// Apply score defaultWeights for each ScorePlugin in parallel.
	f.Parallelizer().Until(ctx, len(f.scorePlugins), func(index int) {
		pl := f.scorePlugins[index]
		// Score plugins' weight has been checked when they are initialized.
		weight := f.scorePluginWeight[pl.Name()]
		nodeScoreList := pluginToNodeScores[pl.Name()]

		for i, nodeScore := range nodeScoreList {
			// return error if score plugin returns invalid score.
			if nodeScore.Score > framework.MaxNodeScore || nodeScore.Score < framework.MinNodeScore {
				err := fmt.Errorf("plugin %q returns an invalid score %v for node %q, it should in the range of [%v, %v] after normalizing", pl.Name(), nodeScore.Score, nodeScore.Name, framework.MinNodeScore, framework.MaxNodeScore)
				log.Error(err, "Invalid Score")

				// TODO: Temporarily, we are rounding scores up to the maximum when they exceed 100, and down to the minimum if they are below zero. We will implement a permanent solution as soon as it is ready.
				//errCh.SendErrorWithCancel(err, cancel)
				//return
				logCurrentEnv(ctx, pluginToNodeScores, "invalid-score")
				if nodeScore.Score > framework.MaxNodeScore {
					nodeScore.Score = framework.MaxNodeScore
				} else {
					nodeScore.Score = framework.MinNodeScore
				}

			}
			nodeScoreList[i].Score = nodeScore.Score * int64(weight)
		}
	})
	logCurrentEnv(ctx, pluginToNodeScores, "after-weights-score")
	if err := errCh.ReceiveError(); err != nil {
		return nil, framework.AsStatus(fmt.Errorf("applying score defaultWeights on Score plugins: %w", err))
	}

	return pluginToNodeScores, nil
}

func (f *frameworkImpl) runScorePlugin(ctx context.Context, pl framework.ScorePlugin, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	if !state.ShouldRecordPluginMetrics() {
		return pl.Score(ctx, state, pod, nodeName)
	}
	startTime := time.Now()
	s, status := pl.Score(ctx, state, pod, nodeName)
	f.metricsRecorder.observePluginDurationAsync(score, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return s, status
}

func (f *frameworkImpl) runScoreExtension(ctx context.Context, pl framework.ScorePlugin, state *framework.CycleState, pod *v1.Pod, nodeScoreList framework.NodeScoreList) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.ScoreExtensions().NormalizeScore(ctx, state, pod, nodeScoreList)
	}
	startTime := time.Now()
	status := pl.ScoreExtensions().NormalizeScore(ctx, state, pod, nodeScoreList)
	f.metricsRecorder.observePluginDurationAsync(scoreExtensionNormalize, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunReservePluginsReserve runs the Reserve method in the set of configured
// reserve plugins. If any of these plugins returns an error, it does not
// continue running the remaining ones and returns the error. In such a case,
// the pod will not be scheduled and the caller will be expected to call
// RunReservePluginsUnreserve.
func (f *frameworkImpl) RunReservePluginsReserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (status *framework.Status) {
	log := log.FromContext(ctx).WithName("frameworkImpl.RunReservePluginsReserve")
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(reserve, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	for _, pl := range f.reservePlugins {
		status = f.runReservePluginReserve(ctx, pl, state, pod, nodeName)
		if !status.IsSuccess() {
			err := status.AsError()
			log.Error(err, "Failed running Reserve plugin", logkeys.Plugin, pl.Name(), logkeys.Pod, pod)
			return framework.AsStatus(fmt.Errorf("running Reserve plugin %q: %w", pl.Name(), err))
		}
	}
	return nil
}

func (f *frameworkImpl) runReservePluginReserve(ctx context.Context, pl framework.ReservePlugin, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.Reserve(ctx, state, pod, nodeName)
	}
	startTime := time.Now()
	status := pl.Reserve(ctx, state, pod, nodeName)
	f.metricsRecorder.observePluginDurationAsync(reserve, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunReservePluginsUnreserve runs the Unreserve method in the set of
// configured reserve plugins.
func (f *frameworkImpl) RunReservePluginsUnreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(unreserve, framework.Success.String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	// Execute the Unreserve operation of each reserve plugin in the
	// *reverse* order in which the Reserve operation was executed.
	for i := len(f.reservePlugins) - 1; i >= 0; i-- {
		f.runReservePluginUnreserve(ctx, f.reservePlugins[i], state, pod, nodeName)
	}
}

func (f *frameworkImpl) runReservePluginUnreserve(ctx context.Context, pl framework.ReservePlugin, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	if !state.ShouldRecordPluginMetrics() {
		pl.Unreserve(ctx, state, pod, nodeName)
		return
	}
	startTime := time.Now()
	pl.Unreserve(ctx, state, pod, nodeName)
	f.metricsRecorder.observePluginDurationAsync(unreserve, pl.Name(), nil, metrics.SinceInSeconds(startTime))
}

// SnapshotSharedLister returns the scheduler's SharedLister of the latest NodeInfo
// snapshot. The snapshot is taken at the beginning of a scheduling cycle and remains
// unchanged until a pod finishes "Reserve". There is no guarantee that the information
// remains unchanged after "Reserve".
func (f *frameworkImpl) SnapshotSharedLister() framework.SharedLister {
	return f.snapshotSharedLister
}

// HasFilterPlugins returns true if at least one filter plugin is defined.
func (f *frameworkImpl) HasFilterPlugins() bool {
	return len(f.filterPlugins) > 0
}

// HasPostFilterPlugins returns true if at least one postFilter plugin is defined.
func (f *frameworkImpl) HasPostFilterPlugins() bool {
	return len(f.postFilterPlugins) > 0
}

// HasScorePlugins returns true if at least one score plugin is defined.
func (f *frameworkImpl) HasScorePlugins() bool {
	return len(f.scorePlugins) > 0
}

// ListPlugins returns a map of extension point name to plugin names configured at each extension
// point. Returns nil if no plugins where configured.
func (f *frameworkImpl) ListPlugins() *config.Plugins {
	m := config.Plugins{}

	for _, e := range f.getExtensionPoints(&m) {
		plugins := reflect.ValueOf(e.slicePtr).Elem()
		extName := plugins.Type().Elem().Name()
		var cfgs []config.Plugin
		for i := 0; i < plugins.Len(); i++ {
			name := plugins.Index(i).Interface().(framework.Plugin).Name()
			p := config.Plugin{Name: name}
			if extName == "ScorePlugin" {
				// Weights apply only to score plugins.
				p.Weight = int32(f.scorePluginWeight[name])
			}
			cfgs = append(cfgs, p)
		}
		if len(cfgs) > 0 {
			e.plugins.Enabled = cfgs
		}
	}
	return &m
}

func (f *frameworkImpl) pluginsNeeded(plugins *config.Plugins) sets.String {
	pgSet := sets.String{}

	if plugins == nil {
		return pgSet
	}

	find := func(pgs *config.PluginSet) {
		for _, pg := range pgs.Enabled {
			pgSet.Insert(pg.Name)
		}
	}

	for _, e := range f.getExtensionPoints(plugins) {
		find(e.plugins)
	}
	// Parse MultiPoint separately since they are not returned by f.getExtensionPoints()
	find(&plugins.MultiPoint)

	return pgSet
}

// ProfileName returns the profile name associated to this framework.
func (f *frameworkImpl) ProfileName() string {
	return f.profileName
}

// Parallelizer returns a parallelizer holding parallelism for scheduler.
func (f *frameworkImpl) Parallelizer() parallelize.Parallelizer {
	return f.parallelizer
}

func logCurrentEnv(ctx context.Context, pluginToNodeScores framework.PluginToNodeScores, step string) {
	log := log.FromContext(ctx).WithName("framework.logCurrentEnv")
	for pluginName, nodeScores := range pluginToNodeScores {
		for _, score := range nodeScores {
			log.V(3).Info(step, logkeys.Plugin, pluginName, logkeys.NodeName, score.Name, logkeys.NodeScore, score.Score)
		}
	}
}
