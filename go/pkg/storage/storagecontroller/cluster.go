// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package storagecontroller

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
)

const (
	clusterLabelWekaBackend     = "wekaBackend"
	clusterLabelVASTBackend     = "vastBackend"
	clusterLabelVastVMSEndpoint = "vastVMSEndpoint"
	clusterLabelCategory        = "category"
	clusterLabels3Endpoint      = "s3Endpoints"
	clusterTypeLabel            = "type"
)

type ClusterInfo struct {
	Name               string
	Addr               string
	UUID               string
	DCName             string
	DCLocation         string
	AvailableCapacity  int64
	TotalCapacity      int64
	AvailableNamespace int32
	TotalNamespace     int32
	Status             string
	Category           string
	S3Endpoint         string
	WekaBackend        string
	VASTBackend        string
	VastVMSEndpoint    string
	Type               string
	ClusterLabelType   string
}

// Gets the list of clusters with UUIDs
func (client *StorageControllerClient) GetClusters(ctx context.Context) ([]ClusterInfo, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.GetClusters").Start()
	defer span.End()
	logger.Info("starting list clusters query")

	clusters := []ClusterInfo{}
	resp, err := client.ClusterSvcClient.ListClusters(ctx, &stcnt_api.ListClustersRequest{})
	if err != nil {
		logger.Error(err, "error quering cluster information")
		return nil, fmt.Errorf("error querying cluster information")
	}
	if resp != nil {
		for _, cl := range resp.Clusters {
			logger.Info("cluster status information", logkeys.Cluster, cl)
			cluster := intoClusterInfo(cl)
			logger.Info("adding cluster information", logkeys.Cluster, cluster)
			clusters = append(clusters, cluster)
		}
	} else {
		logger.Info("no clusters returned", logkeys.Response, resp)
	}

	return clusters, nil
}

// Gets the list of clusters with UUIDs
func (client *StorageControllerClient) GetClusterStatus(ctx context.Context, clusterUUID string) (ClusterInfo, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageControllerClient.GetClusterStatus").Start()
	defer span.End()
	logger.Info("starting cluster status query")
	var status ClusterInfo
	resp, err := client.ClusterSvcClient.GetCluster(ctx, &stcnt_api.GetClusterRequest{
		ClusterId: &stcnt_api.ClusterIdentifier{Uuid: clusterUUID},
	})
	if err != nil {
		logger.Error(err, "error quering cluster state", logkeys.ClusterId, clusterUUID)
		return status, fmt.Errorf("error quering cluster status")
	}
	status = intoClusterInfo(resp.Cluster)
	logger.Info("cluster status information", logkeys.ClusterStatus, status)
	return status, nil
}

func intoClusterInfo(cl *stcnt_api.Cluster) ClusterInfo {
	var status string
	switch cl.Health.Status {
	case stcnt_api.Cluster_Health_STATUS_HEALTHY:
		status = "Online"
	case stcnt_api.Cluster_Health_STATUS_DEGRADED:
		status = "Degraded"
	default:
		status = "Offline"
	}
	backendAddr := ""
	wekaBackend := ""
	if val, found := cl.Labels[clusterLabelWekaBackend]; found {
		wekaBackend = val
		backendAddr = val
	}

	vastBackend := ""
	if val, found := cl.Labels[clusterLabelVASTBackend]; found {
		vastBackend = val
		backendAddr = val
	}
	vastVMSEndpoint := ""
	if val, found := cl.Labels[clusterLabelVastVMSEndpoint]; found {
		vastVMSEndpoint = val
	}

	s3Endpoint := ""
	if val, found := cl.Labels[clusterLabels3Endpoint]; found {
		s3Endpoint = val
	}
	category := ""
	if val, found := cl.Labels[clusterLabelCategory]; found {
		category = val
	}
	typeOfCluster := ""
	if label, found := cl.Labels[clusterTypeLabel]; found {
		typeOfCluster = label
	}
	var clusterType string
	switch cl.Type {
	case stcnt_api.Cluster_TYPE_WEKA:
		clusterType = "Weka"
	case stcnt_api.Cluster_TYPE_MINIO:
		clusterType = "Minio"
	case stcnt_api.Cluster_TYPE_VAST:
		clusterType = "VAST"
	default:
		clusterType = "Unspecified"
	}
	return ClusterInfo{
		Name:               cl.Name,
		Addr:               backendAddr,
		UUID:               cl.Id.Uuid,
		DCName:             cl.Location, // this field is redundant, location already encoded dc name
		DCLocation:         cl.Location,
		AvailableCapacity:  int64(cl.Capacity.Storage.AvailableBytes),
		TotalCapacity:      int64(cl.Capacity.Storage.TotalBytes),
		AvailableNamespace: cl.Capacity.Namespaces.AvailableCount,
		TotalNamespace:     cl.Capacity.Namespaces.TotalCount,
		Status:             status,
		Category:           category,
		S3Endpoint:         s3Endpoint,
		WekaBackend:        wekaBackend,
		VASTBackend:        vastBackend,
		VastVMSEndpoint:    vastVMSEndpoint,
		Type:               clusterType,
		ClusterLabelType:   typeOfCluster,
	}
}
