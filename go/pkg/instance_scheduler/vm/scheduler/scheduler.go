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
	"errors"
	"fmt"
	"sync"
	"time"

	schedulerapi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/parallelize"
	frameworkplugins "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins"
	frameworkruntime "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/runtime"
	internalcache "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/internal/cache"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/profile"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/util"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	metal3ClientSet "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/clientset/versioned"
	metal3Informerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/informers/externalversions"
)

const (
	// Fixed cluster ID used to indicate the K8s cluster with Metal3.
	BmaasLocalCluster = "BMaaS"
)

// ErrNoNodesAvailable is used to describe the error that no nodes available to schedule pods.
var ErrNoNodesAvailable = fmt.Errorf("no nodes available to schedule pods")

// Scheduler monitors nodes for allocatable resources and workloads (pods) and updates the cache.
// Schedules a single instance (pod) based on information in the cache.
// It attempts to find nodes that they fit on returns the recommended node.
type Scheduler struct {
	// It is expected that changes made via Cache will be observed
	// by NodeLister and Algorithm.
	Cache internalcache.Cache

	// Close this to shut down the scheduler.
	StopEverything <-chan struct{}

	// Profiles are the scheduling profiles.
	Profiles profile.Map

	nodeInfoSnapshot *internalcache.Snapshot

	percentageOfNodesToScore int32

	enableBinpack bool

	bgpNetworkRequiredInstanceCountThreshold int

	nextStartNodeIndex int

	clusters []*Cluster

	// Mutex to ensure only one pod is scheduled at a time.
	scheduleOneLock           sync.Mutex
	cfg                       *privatecloudv1alpha1.VmInstanceSchedulerConfig
	instanceTypeServiceClient pb.InstanceTypeServiceClient
}

type schedulerOptions struct {
	percentageOfNodesToScore                 int32
	profiles                                 []schedulerapi.KubeSchedulerProfile
	parallelism                              int32
	durationToExpireAssumedPod               time.Duration
	enableBinpack                            bool
	bgpNetworkRequiredInstanceCountThreshold int
}

// Option configures a Scheduler
type Option func(*schedulerOptions)

// ScheduleResult represents the result of scheduling a pod.
type ScheduleResult struct {
	// Name of the selected node.
	SuggestedHost string
	// Labels of the selected node.
	SuggestedHostLabels map[string]string
	// The number of nodes the scheduler evaluated the pod against in the filtering
	// phase and beyond.
	EvaluatedNodes int
	// The number of nodes out of the evaluated ones that fit the pod.
	FeasibleNodes int
	// If the pod had a non-empty TopologySpreadConstraints,
	// this will contain the partition that the pod must run on.
	Partition string

	// List of the pool labels configured with the Node
	ComputeNodePools []string
}

// WithProfiles sets profiles for Scheduler. By default, there is one profile
// with the name "default-scheduler".
func WithProfiles(p ...schedulerapi.KubeSchedulerProfile) Option {
	return func(o *schedulerOptions) {
		o.profiles = p
	}
}

// WithParallelism sets the parallelism for all scheduler algorithms. Default is 16.
func WithParallelism(threads int32) Option {
	return func(o *schedulerOptions) {
		o.parallelism = threads
	}
}

// WithPercentageOfNodesToScore sets percentageOfNodesToScore for Scheduler, the default value is 50
func WithPercentageOfNodesToScore(percentageOfNodesToScore int32) Option {
	return func(o *schedulerOptions) {
		o.percentageOfNodesToScore = percentageOfNodesToScore
	}
}

func WithDurationToExpireAssumedPod(durationToExpireAssumedPod time.Duration) Option {
	return func(o *schedulerOptions) {
		if durationToExpireAssumedPod != time.Duration(0) {
			o.durationToExpireAssumedPod = durationToExpireAssumedPod
		}
	}
}

func WithBmBinpack(enableBinpack bool) Option {
	return func(o *schedulerOptions) {
		o.enableBinpack = enableBinpack
	}
}

func WithBGPNetworkRequiredInstanceCountThreshold(instanceCount int) Option {
	return func(o *schedulerOptions) {
		o.bgpNetworkRequiredInstanceCountThreshold = instanceCount
	}
}

var defaultSchedulerOptions = schedulerOptions{
	percentageOfNodesToScore:   schedulerapi.DefaultPercentageOfNodesToScore,
	parallelism:                int32(parallelize.DefaultParallelism),
	durationToExpireAssumedPod: util.DurationToExpireAssumedPod,
}

// New returns a Scheduler
func New(
	clusters []*Cluster,
	stopCh <-chan struct{},
	cfg *privatecloudv1alpha1.VmInstanceSchedulerConfig,
	instanceTypeServiceClient pb.InstanceTypeServiceClient,
	opts ...Option,
) (*Scheduler, error) {

	stopEverything := stopCh
	if stopEverything == nil {
		stopEverything = wait.NeverStop
	}

	options := defaultSchedulerOptions
	for _, opt := range opts {
		opt(&options)
	}

	registry := frameworkplugins.NewInTreeRegistry()

	metrics.Register()

	// The nominator will be passed all the way to framework instantiation.
	snapshot := internalcache.NewEmptySnapshot()

	profiles, err := profile.NewMap(options.profiles, registry,
		frameworkruntime.WithSnapshotSharedLister(snapshot),
		frameworkruntime.WithParallelism(int(options.parallelism)),
	)
	if err != nil {
		return nil, fmt.Errorf("initializing profiles: %v", err)
	}

	if len(profiles) == 0 {
		return nil, errors.New("at least one profile is required")
	}

	schedulerCache := internalcache.New(options.durationToExpireAssumedPod, stopEverything)

	sched := &Scheduler{
		Cache:                                    schedulerCache,
		StopEverything:                           stopEverything,
		Profiles:                                 profiles,
		nodeInfoSnapshot:                         snapshot,
		percentageOfNodesToScore:                 options.percentageOfNodesToScore,
		enableBinpack:                            options.enableBinpack,
		bgpNetworkRequiredInstanceCountThreshold: options.bgpNetworkRequiredInstanceCountThreshold,
		clusters:                                 clusters,
		scheduleOneLock:                          sync.Mutex{},
		cfg:                                      cfg,
		instanceTypeServiceClient:                instanceTypeServiceClient,
	}

	for _, cluster := range sched.clusters {
		cluster.informersStarted = make(chan struct{})
		if cluster.ClusterId == BmaasLocalCluster {
			addAllEventHandlers(sched, cluster.InformerFactory.(metal3Informerfactory.SharedInformerFactory), cluster.ClusterId)
		} else {
			addAllEventHandlers(sched, cluster.InformerFactory.(informers.SharedInformerFactory), cluster.ClusterId)
		}
	}

	return sched, nil
}

func (sched *Scheduler) StartInformers(ctx context.Context) {
	for _, cluster := range sched.clusters {
		if cluster.ClusterId == BmaasLocalCluster {
			cluster.InformerFactory.(metal3Informerfactory.SharedInformerFactory).Start(ctx.Done())
		} else {
			cluster.InformerFactory.(informers.SharedInformerFactory).Start(ctx.Done())
		}
		close(cluster.informersStarted)
	}
}

func NewMetal3InformerFactory(cs metal3ClientSet.Interface, resyncPeriod time.Duration) metal3Informerfactory.SharedInformerFactory {

	informerFactory := metal3Informerfactory.NewSharedInformerFactory(cs, resyncPeriod)
	return informerFactory
}

// NewInformerFactory creates a SharedInformerFactory and initializes a scheduler specific
// in-place podInformer.
func NewInformerFactory(cs clientset.Interface, resyncPeriod time.Duration) informers.SharedInformerFactory {
	informerFactory := informers.NewSharedInformerFactory(cs, resyncPeriod)
	informerFactory.InformerFor(&v1.Pod{}, newPodInformer)
	return informerFactory
}

// newPodInformer creates a shared index informer that returns only non-terminal pods.
func newPodInformer(cs clientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	selector := fmt.Sprintf("status.phase!=%v,status.phase!=%v", v1.PodSucceeded, v1.PodFailed)
	tweakListOptions := func(options *metav1.ListOptions) {
		options.FieldSelector = selector
	}
	return coreinformers.NewFilteredPodInformer(cs, metav1.NamespaceAll, resyncPeriod, nil, tweakListOptions)
}
