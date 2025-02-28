// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"slices"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/conf"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/server/helpers"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClusterHandler struct {
	Backends map[string]backend.Interface
	Clusters []*conf.Cluster
}

// GetCluster implements v1.ClusterServiceServer.
func (h *ClusterHandler) GetCluster(ctx context.Context, r *v1.GetClusterRequest) (*v1.GetClusterResponse, error) {
	var cluster *conf.Cluster

	for _, v := range h.Clusters {
		if v.UUID == r.GetClusterId().GetUuid() {
			cluster = v
		}
	}
	if cluster == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster does not exists")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	status, err := h.Backends[cluster.UUID].GetStatus(ctx)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Str("uuid", r.GetClusterId().GetUuid()).Msg("Could not obtain cluster status")
		return nil, err
	}

	return &v1.GetClusterResponse{
		Cluster: intoProtoCluster(status, cluster),
	}, nil
}

// ListClusters implements v1.ClusterServiceServer.
func (h *ClusterHandler) ListClusters(ctx context.Context, r *v1.ListClustersRequest) (*v1.ListClustersResponse, error) {
	clusters := make([]*v1.Cluster, 0)

	for _, cluster := range filterClusters(h.Clusters, r.GetFilter()) {
		status, err := h.Backends[cluster.UUID].GetStatus(ctx)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Str("uuid", cluster.UUID).Msg("Could not obtain cluster status")
		} else {
			clusters = append(clusters, intoProtoCluster(status, cluster))
		}
	}

	return &v1.ListClustersResponse{
		Clusters: clusters,
	}, nil
}

func intoProtoCluster(status *backend.ClusterStatus, cluster *conf.Cluster) *v1.Cluster {
	// Cluster labels are overridden by actual status labels
	for k, v := range status.Labels {
		cluster.Labels[k] = v
	}

	return &v1.Cluster{
		Id: &v1.ClusterIdentifier{
			Uuid: cluster.UUID,
		},
		Name:        cluster.Name,
		Location:    cluster.Location,
		Type:        intoProtoClusterType(cluster.Type),
		SupportsApi: intoSupportsAPI(cluster.SupportsAPI),
		Labels:      cluster.Labels,
		Health: &v1.Cluster_Health{
			Status: v1.Cluster_Health_Status(status.HealthStatus), // Such conversation can be dangerous if enums description change
		},
		Capacity: &v1.Cluster_Capacity{
			Storage: &v1.Cluster_Capacity_Storage{
				TotalBytes:     status.TotalBytes,
				AvailableBytes: status.AvailableBytes,
			},
			Namespaces: &v1.Cluster_Capacity_Namespaces{
				TotalCount:     status.NamespacesLimit,
				AvailableCount: status.NamespacesAvailable,
			},
		},
	}
}

func intoProtoClusterType(t conf.ClusterType) v1.Cluster_Type {
	if t == conf.Weka {
		return v1.Cluster_TYPE_WEKA
	} else if t == conf.MinIO {
		return v1.Cluster_TYPE_MINIO
	} else if t == conf.Vast {
		return v1.Cluster_TYPE_VAST
	}

	return v1.Cluster_TYPE_UNSPECIFIED
}

func intoSupportsAPI(apis []conf.SupportsAPI) []v1.Cluster_ApiType {
	result := []v1.Cluster_ApiType{}
	for _, api := range apis {
		if api == conf.ObjectStore {
			result = append(result, *v1.Cluster_API_TYPE_OBJECT_STORE.Enum())
		} else if api == conf.WekaFilesystem {
			result = append(result, *v1.Cluster_API_TYPE_WEKA_FILESYSTEM.Enum())
		} else if api == conf.VastView {
			result = append(result, *v1.Cluster_API_TYPE_VAST_VIEW.Enum())
		}
	}

	return result
}

func filterClusters(clusters []*conf.Cluster, filter *v1.ListClustersRequest_Filter) []*conf.Cluster {
	if filter == nil {
		return clusters
	}

	result := []*conf.Cluster{}
	for _, cluster := range clusters {
		if len(filter.GetTypes()) > 0 {
			if cluster.Type == conf.Weka && !slices.Contains(filter.GetTypes(), v1.Cluster_TYPE_WEKA) {
				continue
			}
		}
		if len(filter.GetLocations()) > 0 && !slices.Contains(filter.GetLocations(), cluster.Location) {
			continue
		}
		if len(filter.GetNames()) > 0 && !slices.Contains(filter.GetNames(), cluster.Name) {
			continue
		}
		if len(filter.GetLabels()) > 0 && !helpers.IsAllKeyValueExists(cluster.Labels, filter.Labels) {
			continue
		}

		result = append(result, cluster)
	}
	return result
}
