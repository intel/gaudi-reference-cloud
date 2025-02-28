// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	schedulerapi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins/names"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/scheduler"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/util"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	toolsk8s "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s"
	"google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	metal3ClientSet "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/clientset/versioned"
)

const (
	InsufficientResourcesError = "insufficient capacity to launch this instance"
	HarvesterAPIGroup          = "harvesterhci.io"
	HarvesterAPIVersion        = "v1beta1"
	HarvesterAPIResource       = "settings"
)

// SchedulingService provides:
//  1. Implementation of GRPC InstanceSchedulingService.
//  2. Scheduler:
//     - Monitors nodes for allocatable resources and workloads (pods) and updates the cache.
//     - Schedules a single instance (pod) based on information in the cache.
type SchedulingService struct {
	pb.UnimplementedInstanceSchedulingServiceServer
	Sched *scheduler.Scheduler
}

// Create a new scheduling service.
// Finds KubeConfig files in a directory and configures the controller manager so that it runs the informers on each identified cluster.
func NewSchedulingService(ctx context.Context, cfg *privatecloudv1alpha1.VmInstanceSchedulerConfig, mgr ctrl.Manager, instanceTypeServiceClient pb.InstanceTypeServiceClient) (*SchedulingService, error) {
	logger := log.FromContext(ctx).WithName("NewSchedulingService")
	logger.Info("BEGIN")
	// initialize Clusters
	var clusters []*scheduler.Cluster

	if cfg.EnableBMaaSLocal {
		// add BMaas cluster
		config, err := util.GetKubeRestConfig(cfg.BmKubeConfigDir)
		if err != nil {
			return nil, fmt.Errorf("unable to get K8s REST config: %v", err)
		}

		// Add BMaaS
		// TODO support multiple clusters
		client, err := metal3ClientSet.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("unable to get K8s REST config: %v", err)
		}
		informerFactory := scheduler.NewMetal3InformerFactory(client, 0)
		cluster := &scheduler.Cluster{
			ClusterId:       scheduler.BmaasLocalCluster,
			InformerFactory: informerFactory,
		}
		logger.Info("clusterInfo", logkeys.Cluster, cluster)
		clusters = append(clusters, cluster)
	}
	kubeConfigsDir := os.DirFS(cfg.VmClustersKubeConfigDir)

	harvesterKubeConfigs, err := toolsk8s.LoadKubeConfigFiles(ctx, kubeConfigsDir, util.HarvesterConfPattern)
	if err != nil {
		return nil, err
	}
	kubeVirtKubeConfigs, err := toolsk8s.LoadKubeConfigFiles(ctx, kubeConfigsDir, util.KubeVirtConfPattern)
	if err != nil {
		return nil, err
	}

	if err := addClusters(ctx, kubeVirtKubeConfigs, cfg, &clusters, false); err != nil {
		return nil, err
	}

	if err := addClusters(ctx, harvesterKubeConfigs, cfg, &clusters, true); err != nil {
		return nil, err
	}
	profile, err := getSchedulerProfile(cfg)
	if err != nil {
		return nil, err
	}
	logger.Info("NewSchedulingService profile", logkeys.ProfileName, profile)

	sched, err := scheduler.New(
		clusters,
		ctx.Done(),
		cfg,
		instanceTypeServiceClient,
		scheduler.WithProfiles(*profile),
		scheduler.WithDurationToExpireAssumedPod(cfg.DurationToExpireAssumedPod.Duration),
		scheduler.WithBmBinpack(cfg.EnableBMaaSBinpack),
		scheduler.WithBGPNetworkRequiredInstanceCountThreshold(cfg.BGPNetworkRequiredInstanceCountThreshold),
	)
	if err != nil {
		return nil, err
	}

	schedulingService := &SchedulingService{
		Sched: sched,
	}
	if err := mgr.Add(manager.RunnableFunc(schedulingService.Run)); err != nil {
		return nil, err
	}
	logger.Info("END")
	return schedulingService, nil
}

// Get the scheduler profile that defines the kube-scheduler plugins that will be used.
func getSchedulerProfile(cfg *privatecloudv1alpha1.VmInstanceSchedulerConfig) (*schedulerapi.KubeSchedulerProfile, error) {
	profile := &schedulerapi.KubeSchedulerProfile{
		SchedulerName: scheduler.DefaultSchedulerName,
		Plugins: &schedulerapi.Plugins{
			MultiPoint: schedulerapi.PluginSet{
				Enabled: []schedulerapi.Plugin{
					{
						// Filter nodes with unschedule flag.
						Name: names.NodeUnschedulable,
					},
					{
						// Filter tainted nodes (avoid nodes with full disks).
						Name: names.TaintToleration,
					},
					{
						// Filter and score nodes based on CPU and memory resources.
						Name: names.NodeResourcesFit,
					},
					{
						// Filter and score nodes based on node affinity rules (instance type).
						Name: names.NodeAffinity,
					},
					{
						// Filter and score nodes based on pod affinity rules (e.g. rack anti-affinity).
						Name: names.InterPodAffinity,
					},
					{
						// Filter and score nodes based on pod topology spread.
						Name: names.PodTopologySpread,
					},
				},
			},
		},
		PluginConfig: []schedulerapi.PluginConfig{
			{
				Name: names.NodeResourcesFit,
				Args: &schedulerapi.NodeResourcesFitArgs{
					ScoringStrategy: &schedulerapi.ScoringStrategy{
						Type: schedulerapi.LeastAllocated,
						Resources: []schedulerapi.ResourceSpec{
							// Relative weights used for scoring nodes based on resources.
							{
								Name:   string(corev1.ResourcePods),
								Weight: 1,
							},
							{
								Name:   string(corev1.ResourceCPU),
								Weight: 1,
							},
							{
								Name:   string(corev1.ResourceMemory),
								Weight: 1,
							},
						},
					},
				},
			},
			{
				Name: names.NodeAffinity,
				Args: &schedulerapi.NodeAffinityArgs{},
			},
			{
				Name: names.InterPodAffinity,
				Args: &schedulerapi.InterPodAffinityArgs{},
			},
			{
				Name: names.PodTopologySpread,
				Args: &schedulerapi.PodTopologySpreadArgs{},
			},
		},
	}
	return profile, nil
}

// Run informers for all clusters.
// Blocks until context is cancelled.
func (s *SchedulingService) Run(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("SchedulingService.Run")
	logger.Info("BEGIN")
	defer logger.Info("END")
	defer utilruntime.HandleCrash()
	go s.Sched.StartInformers(ctx)
	if err := s.Sched.WaitForCacheSync(ctx); err != nil {
		return err
	}
	logger.Info("Running")
	<-ctx.Done()
	return nil
}

// Implements the GRPC Schedule function.
func (s *SchedulingService) Schedule(ctx context.Context, req *pb.ScheduleRequest) (*pb.ScheduleResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("SchedulingService.Schedule").Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.ScheduleResponse, error) {
		return s.Sched.Schedule(ctx, req.Instances, req.DryRun)
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, sanitizeScheduleError(err)
}

// Implements the GRPC GetStatistics function.
func (s *SchedulingService) GetStatistics(ctx context.Context, req *emptypb.Empty) (*pb.SchedulerStatistics, error) {
	logger := log.FromContext(ctx).WithName("SchedulingService.GetStatistics")
	logger.Info("SchedulingService.GetStatistics BEGIN")
	defer logger.Info("SchedulingService.GetStatistics END")
	return s.Sched.GetStatistics(ctx)
}

// Implements the GRPC Ready function.
func (sched *SchedulingService) Ready(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("SchedulingService.Ready")
	logger.Info("Ready")
	return &emptypb.Empty{}, nil
}

func sanitizeScheduleError(err error) error {
	if err == nil {
		return err
	}
	switch err.(type) {
	case *framework.FitError:
		return fmt.Errorf(InsufficientResourcesError)
	default:
		return err
	}
}

func (sched *SchedulingService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("SchedulingService.Ping")
	logger.Info("Ping")
	return &emptypb.Empty{}, nil
}

func patchHarvesterOvercommitConfig(ctx context.Context, kubeConfig *restclient.Config, cfg *privatecloudv1alpha1.VmInstanceSchedulerConfig) error {
	logger := log.FromContext(ctx).WithName("patchHarvesterOvercommitConfig")

	// GroupVersionResource for the Harvester 'Settings' Custom Resource
	gvr := schema.GroupVersionResource{
		Group:    HarvesterAPIGroup,
		Version:  HarvesterAPIVersion,
		Resource: HarvesterAPIResource,
	}

	// Create a dynamic client for intearcting with Custom Resources
	dynamicClient, err := dynamic.NewForConfig(restclient.AddUserAgent(kubeConfig, "vm_instance_scheduler"))
	if err != nil {
		return err
	}
	obj, err := dynamicClient.Resource(gvr).Get(ctx, "overcommit-config", metav1.GetOptions{})
	if err != nil {
		return err
	}

	var currentOverCommitConfigSetting map[string]interface{}
	var updatedOvercommitConfigSetting map[string]interface{}

	currentOverCommitConfigString, ok := obj.UnstructuredContent()["value"].(string)
	if !ok {
		return fmt.Errorf("type assertion failed while converting interface to string for current overcommit config settings")
	}
	updatedOvercommitConfigString := fmt.Sprintf(`{"cpu":%d,"memory":%d,"storage":%d}`, cfg.OvercommitConfig.CPU, cfg.OvercommitConfig.Memory, cfg.OvercommitConfig.Storage)

	if err := json.Unmarshal([]byte(currentOverCommitConfigString), &currentOverCommitConfigSetting); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(updatedOvercommitConfigString), &updatedOvercommitConfigSetting); err != nil {
		return err
	}

	if !reflect.DeepEqual(currentOverCommitConfigSetting, updatedOvercommitConfigSetting) {
		patch := []interface{}{
			map[string]interface{}{
				"op":    "replace",
				"path":  "/value",
				"value": updatedOvercommitConfigString,
			},
		}
		payload, err := json.Marshal(patch)
		if err != nil {
			return err
		}

		logger.Info("Patching updated overcommit config setting on harvester", "updatedOvercommitConfigString", updatedOvercommitConfigString)
		_, err = dynamicClient.Resource(gvr).Patch(ctx, "overcommit-config", types.JSONPatchType, payload, metav1.PatchOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
func addClusters(ctx context.Context, kubeConfigs map[string]*restclient.Config, cfg *privatecloudv1alpha1.VmInstanceSchedulerConfig, clusters *[]*scheduler.Cluster, isHarvester bool) error {
	logger := log.FromContext(ctx).WithName("addClusters")
	logger.Info("BEGIN")
	for kubeConfigFilename, kubeConfig := range kubeConfigs {
		logger.Info("kubeConfigs", logkeys.FileName, kubeConfigFilename, logkeys.Configuration, kubeConfig)
		// strip off the file extension
		clusterId := strings.TrimSuffix(kubeConfigFilename, filepath.Ext(kubeConfigFilename))
		client, err := clientset.NewForConfig(restclient.AddUserAgent(kubeConfig, "vm_instance_scheduler"))
		if err != nil {
			return fmt.Errorf("unable to create clientset for %s: %w", kubeConfigFilename, err)
		}

		if isHarvester {
			// Patch Harvester overcommit config settings
			err = patchHarvesterOvercommitConfig(ctx, kubeConfig, cfg)
			if err != nil {
				return err
			}
		}

		informerFactory := scheduler.NewInformerFactory(client, 0)
		cluster := &scheduler.Cluster{
			ClusterId:       clusterId,
			InformerFactory: informerFactory,
		}
		logger.Info("clusterInfo", logkeys.Cluster, cluster, logkeys.ConfigFile, kubeConfigFilename)
		*clusters = append(*clusters, cluster)
	}
	return nil
}
