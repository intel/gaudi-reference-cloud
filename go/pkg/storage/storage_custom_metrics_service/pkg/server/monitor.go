// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	clusterTypeWeka  = "Weka"
	clusterTypeVast  = "VAST"
	clusterTypeMinio = "Minio"
	vastNS           = "vastns"
)

var (
	// store new metric values
	minioMetrics       []bucketMetricValues
	wekaMetrics        []wekaFileMetricValues
	vastMetrics        []vastFileMetricValues
	wekaClusterMetrics []clusterMetricValues
	vastClusterMetrics []clusterMetricValues
)

type bucketMetricValues struct {
	cluster storagecontroller.ClusterInfo
	bkInfo  map[bucketKey]bucketStat
	bkCount map[string]int
}

type clusterMetricValues struct {
	cluster storagecontroller.ClusterInfo
	nsList  []storagecontroller.Namespace
}

type wekaFileMetricValues struct {
	cluster storagecontroller.ClusterInfo
	iksInfo map[string]int
	fsInfo  map[string]int
	total   map[string]int
	files   []storagecontroller.Filesystem
}

type vastFileMetricValues struct {
	cluster storagecontroller.ClusterInfo
	fsInfo  map[string]int
	files   []*vastInfo
}

type bucketKey struct {
	ns string
	id string
}

type iksKey struct {
	ns   string
	name string
}

type bucketStat struct {
	Used int
	Size int
}

type vastInfo struct {
	Namespace string
	Name      string
	Size      int
}

type StorageCustomMetricService struct {
	strCntClient  *storagecontroller.StorageControllerClient
	kmsClient     pb.StorageKMSPrivateServiceClient
	storageClient pb.FilesystemPrivateServiceClient
	objectEnabled bool
}

// constructor for storage-custom-metrics-service
func NewStorageCustomMetricService(strClient *storagecontroller.StorageControllerClient, kmsClient pb.StorageKMSPrivateServiceClient, storageClient pb.FilesystemPrivateServiceClient, objectFlag bool) (*StorageCustomMetricService, error) {
	if strClient == nil {
		return nil, fmt.Errorf("storage client is required")
	}
	if kmsClient == nil {
		return nil, fmt.Errorf("kms client is required")
	}
	return &StorageCustomMetricService{
		strCntClient:  strClient,
		kmsClient:     kmsClient,
		storageClient: storageClient,
		objectEnabled: objectFlag,
	}, nil
}

// central logic for retrieving require info need for metrics
func (sm *StorageCustomMetricService) scanClusters(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.scanClusters")
	wekaCount, vastCount, minioCount := 0, 0, 0
	logger.Info("Begin retrieving cluster UUIDs from sds")
	minioClusters, wekaClusters, vastClusters, clusterIds, err := sm.filterClusters(ctx)
	if err != nil {
		logger.Error(err, "error getting cluster info from sds")
	}

	// Minio
	if sm.objectEnabled { //if false, skip bucket metric creation
		logger.Info("Start update Bucket metrics")
		logger.Info("Begin retrieving bucket data from minio clusters")
		minioCount = len(minioClusters)
		for _, cluster := range minioClusters {
			logger.Info("clusterInfo", logkeys.Cluster, cluster)
			//get all buckets for cluster
			bkInfo, bkCount, err := sm.fetchBuckets(ctx, cluster.UUID)
			if err != nil {
				logger.Info("error counting number buckets in cluster")
			}
			minioMetrics = append(minioMetrics, bucketMetricValues{
				cluster: cluster,
				bkInfo:  bkInfo,
				bkCount: bkCount,
			})
		}
	} else {
		logger.Info("Skipping bucket metric update")
	}

	// Weka
	logger.Info("Start update Weka metrics")
	logger.Info("Begin retrieving org data from weka clusters")
	wekaCount = len(wekaClusters)
	for _, cluster := range wekaClusters {
		logger.Info("weka clusterInfo", logkeys.Cluster, cluster)
		//get all ns for cluster
		nsList, err := sm.fetchNamespaces(ctx, cluster.UUID)
		if err != nil {
			logger.Error(err, "error fetching ns from cluster")
		}
		//get all fs for cluster
		iksInfo, fsInfo, total, files, err := sm.fetchFilesytems(ctx, cluster.UUID, nsList)
		if err != nil {
			logger.Error(err, "error fetching fs from cluster")
		}
		wekaClusterMetrics = append(wekaClusterMetrics, clusterMetricValues{
			cluster: cluster,
			nsList:  nsList,
		})
		wekaMetrics = append(wekaMetrics, wekaFileMetricValues{
			cluster: cluster,
			iksInfo: iksInfo,
			fsInfo:  fsInfo,
			total:   total,
			files:   files,
		})
	}

	// Vast
	vastCount = len(vastClusters)
	if vastCount > 0 {
		logger.Info("Start update Vast metrics")
		for _, cluster := range vastClusters {
			logger.Info("vast clusterInfo", logkeys.Cluster, cluster)
			//get all ns for cluster
			nsList, err := sm.fetchNamespaces(ctx, cluster.UUID)
			if err != nil {
				logger.Error(err, "error fetching ns from cluster")
			}
			//get all vast-fs for cluster
			fsInfo, files, err := sm.fetchVastFilesytems(ctx, cluster.UUID, nsList)
			if err != nil {
				logger.Error(err, "error fetching fs from vast cluster")
			}
			vastClusterMetrics = append(vastClusterMetrics, clusterMetricValues{
				cluster: cluster,
				nsList:  nsList,
			})
			vastMetrics = append(vastMetrics, vastFileMetricValues{
				cluster: cluster,
				fsInfo:  fsInfo,
				files:   files,
			})
		}
	}
	// Clean stale metrics
	sm.cleanMetrics(ctx, clusterIds)
	// Update metrics
	updateMetrics(ctx)

	logger.Info("cluster_count", logkeys.ClusterType, "weka", logkeys.TotalCount, wekaCount)
	metricClusterCount.WithLabelValues("weka").Set(float64(wekaCount))
	logger.Info("cluster_count", logkeys.ClusterType, "vast", logkeys.TotalCount, vastCount)
	metricClusterCount.WithLabelValues("vast").Set(float64(vastCount))
	logger.Info("cluster_count", logkeys.ClusterType, "minio", logkeys.TotalCount, minioCount)
	metricClusterCount.WithLabelValues("minio").Set(float64(minioCount))
	logger.Info("metrics update finished")

}

func updateMetrics(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.scanClusters")
	logger.Info("Updating all metrics")
	// Update Minio Metrics
	for _, v := range minioMetrics {
		updateBKMetric(ctx, v.cluster, v.bkInfo, v.bkCount)
	}
	// Update weka metrics
	for _, v := range wekaClusterMetrics {
		updateWekaNSMetric(ctx, v.cluster, v.nsList)
	}
	for _, v := range wekaMetrics {
		updateWekaFSMetric(ctx, v.cluster, v.iksInfo, v.fsInfo, v.total, v.files)
	}
	// Update vast metrics
	for _, v := range vastClusterMetrics {
		updateVastNSMetric(ctx, v.cluster, v.nsList)
	}
	for _, v := range vastMetrics {
		updateVastFSMetric(ctx, v.cluster, v.fsInfo, v.files)
	}
}

// helper function to filter weka clusters
func (sm *StorageCustomMetricService) filterClusters(ctx context.Context) ([]storagecontroller.ClusterInfo, []storagecontroller.ClusterInfo, []storagecontroller.ClusterInfo, []string, error) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.filterClusters")
	logger.Info("Begin filter clusters")
	defer logger.Info("End filter clusters")
	// get all clusters
	var minioClusters []storagecontroller.ClusterInfo
	var wekaClusters []storagecontroller.ClusterInfo
	var vastClusters []storagecontroller.ClusterInfo
	var clusterIds []string
	clusters, err := sm.strCntClient.GetClusters(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// filter clusters by type
	for _, cluster := range clusters {
		if cluster.Type == clusterTypeMinio {
			minioClusters = append(minioClusters, cluster)
		}
		if cluster.Type == clusterTypeWeka {
			wekaClusters = append(wekaClusters, cluster)
		}
		if cluster.Type == clusterTypeVast {
			vastClusters = append(vastClusters, cluster)
		}
		// keep track of all cluster uuids
		clusterIds = append(clusterIds, cluster.UUID)
	}

	return minioClusters, wekaClusters, vastClusters, clusterIds, nil
}

// calculates bucket stats for given cluster
func (sm *StorageCustomMetricService) fetchBuckets(ctx context.Context, clusterId string) (map[bucketKey]bucketStat, map[string]int, error) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.fetchBuckets")
	logger.Info("BEGIN")
	defer logger.Info("END")
	bkUsage := make(map[bucketKey]bucketStat)
	bkCount := make(map[string]int)
	// get all buckets from cluster
	res, err := sm.strCntClient.GetAllBuckets(ctx, storagecontroller.BucketFilter{ClusterId: clusterId})
	if err != nil {
		return bkUsage, bkCount, err
	} else {
		for _, bk := range res {
			name := bk.Metadata.BucketId
			// Check if bucket is prefixed by cloudaccount
			if len(name) > 13 && err == nil {
				cloudaccount := name[:12]
				_, err := strconv.Atoi(cloudaccount)
				if err != nil {
					continue
				}
				// calculate bytes used
				currentUsage := 0
				if bk.Spec.AvailableBytes != 0 {
					currentUsage = int(bk.Spec.Totalbytes - bk.Spec.AvailableBytes)
				}
				stats := bucketStat{Used: int(currentUsage), Size: int(bk.Spec.Totalbytes)}
				bkUsage[bucketKey{cloudaccount, name}] = stats
				bkCount[cloudaccount]++
			} else {
				continue
			}
		}
	}
	return bkUsage, bkCount, nil
}

func (sm *StorageCustomMetricService) fetchIKSPVs(ctx context.Context, req storagecontroller.FilesystemMetadata, files []storagecontroller.Filesystem) (map[iksKey]storagecontroller.Filesystem, error) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.fetchIKSPVs")
	fsKubernetes := make(map[iksKey]storagecontroller.Filesystem) // imitate set structure in golang
	searchReq := pb.FilesystemSearchStreamPrivateRequest{ResourceVersion: "0", AvailabilityZone: "az1"}
	fsStream, err := sm.storageClient.SearchFilesystemRequests(ctx, &searchReq)
	if err != nil {
		logger.Error(err, "error reading requests")
	}
	var fsReq *pb.FilesystemRequestResponse
	for {
		fsReq, err = fsStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error(err, "error reading from stream")
			break
		}
		if fsReq == nil {
			logger.Info("received empty response")
			break
		}
		logger.Info("handle filesystem request", "request", fsReq, "resourceVersion", fsReq.Filesystem.Metadata.ResourceVersion)
		cloudAcc := fsReq.Filesystem.Metadata.CloudAccountId
		if fsReq.Filesystem.Spec.FilesystemType == pb.FilesystemType_ComputeKubernetes {
			pfx := fsReq.Filesystem.Spec.Prefix
			if pfx != "" {
				// Filter list of all filesystems for names with matching prefix
				iksPVs := filterByPrefix(ctx, files, pfx)
				for _, pv := range iksPVs {
					key := iksKey{ns: cloudAcc, name: pv.Metadata.FileSystemName}
					// Add PV to list if not present
					if _, ok := fsKubernetes[key]; !ok {
						fsKubernetes[key] = pv
					}
				}
			}
		}
	}
	return fsKubernetes, nil
}

// fetch all fs per namespace for a cluster
func (sm *StorageCustomMetricService) fetchFilesytems(ctx context.Context, clusterId string, nsList []storagecontroller.Namespace) (map[string]int, map[string]int, map[string]int, []storagecontroller.Filesystem, error) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.fetchFilesystems")
	logger.Info("BEGIN")
	defer logger.Info("END")
	fsCount := 0
	totalFsList := make(map[string]int)
	fsList := make(map[string]int)
	iksList := make(map[string]int)
	files := []storagecontroller.Filesystem{}

	//---------------------------------------------------------------
	for _, ns := range nsList {
		// form credential path
		name := ns.Metadata.Name
		// check if valid namespace for kms
		if len(name) < 12 {
			totalFsList[name] = 0
			continue
		}
		cloudaccount := name[len(name)-12:]
		path := utils.GenerateKMSPath(cloudaccount, clusterId, false)
		// fetch credentials
		user, pass, found, err := readSecretsFromStorageKMS(ctx, sm.kmsClient, path)
		if err != nil || !found {
			continue
		}
		//form request
		req := storagecontroller.FilesystemMetadata{
			User:          user,
			Password:      pass,
			UUID:          clusterId,
			NamespaceName: ns.Metadata.Name,
		}
		// fetch fs for namespace
		filesystems, found, err := sm.strCntClient.GetAllFileSystems(ctx, req)
		if err != nil {
			logger.Info("failed sds call")
		}
		// fetch iks fs for this namespace
		iksPVs, err := sm.fetchIKSPVs(ctx, req, filesystems)
		if err != nil {
			logger.Info("failed to fetch iks fs")
		} else {
			iksList[name] = len(iksPVs)
		}
		if found {
			totalFsList[name] = len(filesystems)
			fsList[name] = len(filesystems) - len(iksPVs)
			fsCount += len(filesystems)
			files = append(files, filesystems...)
		} else {
			totalFsList[name] = 0
		}
		// sleep logic to avoid sds api rate limit
		time.Sleep(5 * time.Second)
	}
	return iksList, fsList, totalFsList, files, nil
}

// fetch all fs per namespace for a cluster
func (sm *StorageCustomMetricService) fetchVastFilesytems(ctx context.Context, clusterId string, nsList []storagecontroller.Namespace) (map[string]int, []*vastInfo, error) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.fetchVastFilesystems")
	logger.Info("BEGIN")
	defer logger.Info("END")
	fsCount := 0
	fsList := make(map[string]int)
	files := []*vastInfo{}

	//---------------------------------------------------------------
	for _, ns := range nsList {
		name := ns.Metadata.Name
		//form request
		req := &storagecontroller.ListFilesystemsParams{
			ClusterID:   clusterId,
			NamespaceID: ns.Metadata.Id,
		}
		// fetch fs for namespace
		filesystems, err := sm.strCntClient.ListVastFilesystems(ctx, req)
		if err != nil {
			logger.Info("failed sds call", "response", err)
		}
		if len(filesystems) > 0 {
			fsList[name] = len(filesystems)
			fsCount += len(filesystems)
			for _, f := range filesystems {
				file := &vastInfo{
					Namespace: name,
					Name:      f.Name,
					Size:      int(f.Capacity.TotalBytes),
				}
				files = append(files, file)
			}

		}
		// sleep logic to avoid sds api rate limit
		logger.Info("vast fs count", "count", fsCount)
		time.Sleep(3 * time.Second)
	}
	return fsList, files, nil
}

// fetch all ns for a cluster
func (sm *StorageCustomMetricService) fetchNamespaces(ctx context.Context, clusterId string) ([]storagecontroller.Namespace, error) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.fetchNamespaces")
	logger.Info("BEGIN")
	defer logger.Info("END")
	nsList, found, err := sm.strCntClient.GetAllFileSystemOrgs(ctx, clusterId)
	if err != nil {
		return nsList, err
	} else if !found {
		logger.Info("no namespaces found for cluster")
		return nsList, nil
	}
	return nsList, nil
}

// fetch all ns for a cluster
func (sm *StorageCustomMetricService) fetchVastNamespaces(ctx context.Context, clusterId string) ([]storagecontroller.Namespace, error) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.fetchVastNamespaces")
	logger.Info("BEGIN")
	defer logger.Info("END")
	nsList, found, err := sm.strCntClient.GetAllFileSystemOrgs(ctx, clusterId)
	if err != nil {
		return nsList, err
	} else if !found {
		logger.Info("no namespaces found for cluster")
		return nsList, nil
	}
	return nsList, nil
}

func (sm *StorageCustomMetricService) StartMetricUpdater(ctx context.Context, interval time.Duration, reg *prometheus.Registry) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.StartMetricUpdater")
	logger.Info("Start Storage Custom Metrics Service")
	// Periodically update the metric
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.scanClusters(ctx)
		}
	}
}
