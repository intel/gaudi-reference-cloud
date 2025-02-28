// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k8s_resource_patcher

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s_resource_patcher/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	toolsk8s "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

type KubernetesResourcePatcher struct {
	fleetAdminServiceClient pb.FleetAdminServiceClient
	cfg                     config.Config
	kubeConfig              *rest.Config
}

func NewKubernetesResourcePatcher(ctx context.Context, fleetAdminServiceClient pb.FleetAdminServiceClient, cfg config.Config) (*KubernetesResourcePatcher, error) {
	var kubeConfig *rest.Config
	// Check if ApplyRemoteKubeconfig is set to true. If so, apply the patches to a remote k8s cluster otherwise use Inclusterconfig for patching
	if cfg.ApplyRemoteKubeConfig {
		if cfg.RemoteKubeConfigFilePath == "" {
			return nil, fmt.Errorf("remoteKubeConfigPath cannot be empty when ApplyRemoteKubeConfig is set to true")
		}
		remoteClusterkubeConfig, err := toolsk8s.LoadKubeConfigFile(ctx, cfg.RemoteKubeConfigFilePath)
		if err != nil {
			return nil, fmt.Errorf("error encountered while loading remote kubeConfig file: %w", err)
		}

		kubeConfig = remoteClusterkubeConfig
	} else {
		// Load InCluster Configuration
		inClusterKubeConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, util.ErrConfigNotFound
		}
		kubeConfig = inClusterKubeConfig
	}
	return &KubernetesResourcePatcher{
		fleetAdminServiceClient: fleetAdminServiceClient,
		cfg:                     cfg,
		kubeConfig:              kubeConfig,
	}, nil
}

func (s *KubernetesResourcePatcher) Start(ctx context.Context) {
	log := log.FromContext(ctx).WithName("KubernetesResourcePatcher.Start")

	go func() {
		// Loop through periodically for every 'PatchApplyInterval' to get the resource patches that needs to be applied on the k8s cluster.
		ticker := time.NewTicker(s.cfg.PatchApplyInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := s.applyResourcePatches(ctx)
				if err != nil {
					log.Error(err, "ResourcePatcher Error")
				}
			}
		}
	}()
}

func (s *KubernetesResourcePatcher) applyResourcePatches(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("KubernetesResourcePatcher.applyResourcePatches")
	log.Info("Fetching Resource Patches from FleetAdminService", "config", s.cfg)

	req := &pb.GetResourcePatchesRequest{
		ClusterId:        s.cfg.ClusterId,
		Region:           s.cfg.Region,
		AvailabilityZone: s.cfg.AvailabilityZone,
	}
	resp, err := s.fleetAdminServiceClient.GetResourcePatches(ctx, req)
	if err != nil {
		return err
	}

	err = s.applyPatchToResource(ctx, s.kubeConfig, resp.ResourcePatches)
	if err != nil {
		return err
	}
	return nil
}

func (s *KubernetesResourcePatcher) applyPatchToResource(ctx context.Context, kubeConfigFile *rest.Config, resourcePatches []*pb.ResourcePatch) error {
	log := log.FromContext(ctx).WithName("KubernetesResourcePatcher.applyPatchToResource")

	// Create a dynamic client for interacting with the 'resourcetype'
	dynamicClient, err := dynamic.NewForConfig(rest.AddUserAgent(kubeConfigFile, "k8s-resource-patcher"))
	if err != nil {
		return fmt.Errorf("error occuured while creating dynamic clientset: %v", err)
	}

	if len(resourcePatches) == 0 {
		log.Info("ResourcePatch list is empty. Nothing to patch.")
		return nil
	}

	var failToApplyPatch bool
	for _, resourcePatch := range resourcePatches {
		log.Info("Applying resourcePatch", "resourcePatch", resourcePatch)
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Ensure that ownedLabelsRegex equals allowedLabelsPrefixes + “/.*”.
			err = s.CheckAllowedLabelRegex(resourcePatch.OwnedLabelsRegex)
			if err != nil {
				return fmt.Errorf("error occurred in checking allowed label regex for the patch with nodeName %s and namespace %s: %w", resourcePatch.NodeName, resourcePatch.Namespace, err)
			}

			gvr := schema.GroupVersionResource{
				Group:    resourcePatch.Gvr.Group,
				Version:  resourcePatch.Gvr.Version,
				Resource: resourcePatch.Gvr.Resource,
			}

			var dynamicResource dynamic.ResourceInterface
			if resourcePatch.Namespace != "" {
				dynamicResource = dynamicClient.Resource(gvr).Namespace(resourcePatch.Namespace)
			} else {
				dynamicResource = dynamicClient.Resource(gvr)
			}

			obj, err := dynamicResource.Get(ctx, resourcePatch.NodeName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error occurred while getting resourcetype %s for resourcePatch with nodeName %s: %w", resourcePatch.Gvr.Resource, resourcePatch.NodeName, err)
			}

			labels := obj.GetLabels()
			if labels == nil {
				labels = make(map[string]string)
			}
			// Check whether any labels need to be changed. If not, no need to call Update
			updateLabels, err := s.LabelsNeedsToBeUpdated(labels, resourcePatch.Labels, resourcePatch.OwnedLabelsRegex)
			if err != nil {
				return fmt.Errorf("failed to evaluate patch update for resource %s (nodeName: %s): %w", resourcePatch.Gvr.Resource, resourcePatch.NodeName, err)
			}

			if !updateLabels {
				log.Info("No label changes detected. Nothing to patch.", "NodeName", resourcePatch.NodeName)
				return nil
			}

			// Remove the owned labels in memory that match the specified regular expression
			labelsToBeDeleted, err := MatchLabelsByRegex(labels, resourcePatch.OwnedLabelsRegex)
			if err != nil {
				return err
			}
			for labelToDelete := range labelsToBeDeleted {
				delete(labels, labelToDelete)
			}

			// Apply patches in memory
			for key, value := range resourcePatch.Labels {
				labels[key] = value
			}
			obj.SetLabels(labels)

			// Update the object in k8s API
			_, err = dynamicResource.Update(ctx, obj, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("error occurred while updating object for patch with nodeName %s: %w", resourcePatch.NodeName, err)
			}
			log.Info("Resource Updated", "NodeName", resourcePatch.NodeName, logkeys.UpdatedLabels, resourcePatch.Labels)
			return nil
		})
		if err != nil {
			failToApplyPatch = true
			log.Error(err, "failed to apply patch", "NodeName", resourcePatch.NodeName)
			continue
		}

	}
	if failToApplyPatch {
		return fmt.Errorf("1 or more patches could not be applied")
	}
	log.Info("resources patched completed")
	return nil
}

func (s *KubernetesResourcePatcher) CheckAllowedLabelRegex(ownedLabelsRegex []string) error {
	allowedRegexMap := make(map[string]bool)
	for _, allowedLabelPrefix := range s.cfg.AllowedLabelPrefixes {
		allowedRegex := allowedLabelPrefix + "/.*"
		allowedRegexMap[allowedRegex] = true
	}

	for _, ownedLabelRegex := range ownedLabelsRegex {
		if _, exists := allowedRegexMap[ownedLabelRegex]; !exists {
			return fmt.Errorf("regex %s is not allowed", ownedLabelRegex)
		}
	}
	return nil
}

func (s *KubernetesResourcePatcher) LabelsNeedsToBeUpdated(labels map[string]string, labelsToBePatched map[string]string, regexPatterns []string) (bool, error) {
	// Check for labels addition or updation
	for labelToBePatched, valueToBePatched := range labelsToBePatched {
		if val, found := labels[labelToBePatched]; !found || valueToBePatched != val {
			return true, nil
		}
	}

	// Check for labels deletion
	existingPoolInstanTypeLabels, err := MatchLabelsByRegex(labels, regexPatterns)
	if err != nil {
		return false, err
	}
	for key := range existingPoolInstanTypeLabels {
		if _, exists := labelsToBePatched[key]; !exists {
			return true, nil
		}
	}
	return false, nil
}

// Match labels by regex
func MatchLabelsByRegex(labels map[string]string, regexPatterns []string) (map[string]string, error) {
	matchingLabels := make(map[string]string)
	for _, pattern := range regexPatterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile regex pattern (%s): %w", pattern, err)
		}
		for label, value := range labels {
			if regex.MatchString(label) {
				matchingLabels[label] = value
			}
		}
	}
	return matchingLabels, nil
}
