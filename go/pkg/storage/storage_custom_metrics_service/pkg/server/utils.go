// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"strconv"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/prometheus/client_golang/prometheus"
)

const gb int64 = 1000000000

// Update weka org related metrics
func updateWekaNSMetric(ctx context.Context, cl storagecontroller.ClusterInfo, nsList []storagecontroller.Namespace) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricServiceUtil.updateWekaNSMetric")
	// Check if the custom metric for the cluster exists
	logger.Info("Begin update Weka NS metric")
	clusterId := cl.UUID
	clusterUsageRatio := float64(cl.TotalCapacity-cl.AvailableCapacity) / float64(cl.TotalCapacity)
	totalCapacity := cl.TotalCapacity / gb
	availCapacity := cl.AvailableCapacity / gb
	orgs := uint(cl.TotalNamespace - cl.AvailableNamespace)
	ratio := float64(orgs) / float64(cl.TotalNamespace)
	// Update the metric with the used ns for the cluster
	logger.Info("weka_cluster_org_used", logkeys.ClusterId, clusterId, logkeys.FilesystemCount, uint(orgs))
	metricNSUsed.WithLabelValues(clusterId).Set(float64(orgs))
	logger.Info("------------------------------------------------------------")
	// Update the metric with the ratio for ns consumption
	logger.Info("weka_cluster_org_used_ratio", logkeys.ClusterId, clusterId, logkeys.Ratio, ratio)
	metricNSRatio.WithLabelValues(clusterId).Set(ratio)
	logger.Info("------------------------------------------------------------")
	// Update the metric with the ratio for cluster space consumption
	logger.Info("cluster_space_usage", logkeys.ClusterId, clusterId, logkeys.ClusterType, "weka", logkeys.Ratio, clusterUsageRatio)
	metricClusterUsage.WithLabelValues(clusterId, "weka").Set(clusterUsageRatio)
	logger.Info("cluster_space_total", logkeys.ClusterId, clusterId, logkeys.ClusterType, "weka", logkeys.ClusterSpaceTotal, totalCapacity)
	metricClusterSpaceTotal.WithLabelValues(clusterId, "weka").Set(float64(totalCapacity))
	logger.Info("cluster_space_available", logkeys.ClusterId, logkeys.ClusterType, "weka", clusterId, logkeys.ClusterSpaceAvailable, availCapacity)
	metricClusterSpaceAvailable.WithLabelValues(clusterId, "weka").Set(float64(availCapacity))
	logger.Info("------------------------------------------------------------")
	// Update the metric with NS space usage
	for _, ns := range nsList {
		usage, err := strconv.ParseFloat(ns.Properties.Quota, 64)
		if err != nil {
			logger.Error(err, "error converting string to float64")
		}
		usage = float64(int(usage) / int(gb)) //convert to GB
		logger.Info("weka_org_space_usage in Gb", logkeys.ClusterId, clusterId, logkeys.Namespace, ns.Metadata.Name, logkeys.Usage, usage)
		metricNSUsage.With(prometheus.Labels{"cluster": clusterId, "namespace": ns.Metadata.Name}).Set(usage)
	}
	logger.Info("NS Metrics update complete")
}

// update weka filesystem relation metrics
func updateWekaFSMetric(ctx context.Context, cl storagecontroller.ClusterInfo, iksList map[string]int, fsList map[string]int, total map[string]int, files []storagecontroller.Filesystem) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricServiceUtil.updateWekaFSMetric")
	clusterId := cl.UUID
	// Check if the custom metric for the cluster exists
	logger.Info("Begin update WekaFS metrics")
	fsCount := len(files)
	// Update total # fs per cluster
	metricFSCount.With(prometheus.Labels{"cluster": clusterId}).Set(float64(fsCount))
	logger.Info("weka_cluster_filesystems_count", logkeys.ClusterId, clusterId, logkeys.FilesystemCount, fsCount)
	metricFSRatio.With(prometheus.Labels{"cluster": clusterId}).Set(float64(fsCount) / 1024)
	logger.Info("weka_cluster_filesystems_ratio", logkeys.ClusterId, clusterId, logkeys.Ratio, float64(fsCount)/1024)
	logger.Info("------------------------------------------------------------")
	for ns, iks := range iksList {
		metricIKS.With(prometheus.Labels{"cluster": clusterId, "namespace": ns}).Set(float64(iks))
		logger.Info("weka_cluster_iks_filesystems_per_org", logkeys.ClusterId, clusterId, logkeys.Namespace, ns, logkeys.IksCount, iks)
	}
	logger.Info("Done IKS Update")
	for ns, fs := range fsList {
		metricFS.With(prometheus.Labels{"cluster": clusterId, "namespace": ns}).Set(float64(fs))
		logger.Info("weka_cluster_filesystems_per_org", logkeys.ClusterId, clusterId, logkeys.Namespace, ns, logkeys.FilesystemCount, fs)
	}
	logger.Info("Done FS Update")
	for ns, fs := range total {
		metricTotalFS.With(prometheus.Labels{"cluster": clusterId, "namespace": ns}).Set(float64(fs))
		logger.Info("weka_cluster_total_filesystems_per_org", logkeys.ClusterId, clusterId, logkeys.Namespace, ns, logkeys.TotalCount, fs)
	}
	logger.Info("Done Total FS Update")
	for _, fs := range files {
		size, err := strconv.Atoi(fs.Properties.FileSystemCapacity)
		if err != nil {
			continue
		}
		size /= int(gb) // convert to GB
		metricFSUsage.With(prometheus.Labels{"cluster": clusterId, "cloudaccount": fs.Metadata.NamespaceName, "name": fs.Metadata.FileSystemName}).Set(float64(size))
		logger.Info("weka_filesystem_size", logkeys.ClusterId, clusterId, logkeys.Namespace, fs.Metadata.NamespaceName, logkeys.FilesystemName, fs.Metadata.FileSystemName, logkeys.Size, float64(size))
	}
	logger.Info("Done FS Usage Update")
	logger.Info("FS Metrics update complete")
}

// Update vast org related metrics
func updateVastNSMetric(ctx context.Context, cl storagecontroller.ClusterInfo, nsList []storagecontroller.Namespace) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricServiceUtil.updateVastNSMetric")
	// Check if the custom metric for the cluster exists
	logger.Info("Begin update Vast NS metric")
	clusterId := cl.UUID
	clusterUsageRatio := float64(cl.TotalCapacity-cl.AvailableCapacity) / float64(cl.TotalCapacity)
	totalCapacity := cl.TotalCapacity / gb
	availCapacity := cl.AvailableCapacity / gb
	orgs := len(nsList)
	logger.Info("used orgs", "count", orgs)
	ratio := float64(orgs) / float64(1024)
	// Update the metric with the used ns for the cluster
	logger.Info("vast_cluster_org_used", logkeys.ClusterId, clusterId, logkeys.FilesystemCount, uint(orgs))
	metricVastNSUsed.WithLabelValues(clusterId).Set(float64(orgs))
	logger.Info("------------------------------------------------------------")
	// Update the metric with the ratio for ns consumption
	logger.Info("vast_cluster_org_used_ratio", logkeys.ClusterId, clusterId, logkeys.Ratio, ratio)
	metricVastNSRatio.WithLabelValues(clusterId).Set(ratio)
	logger.Info("------------------------------------------------------------")
	// Update the metric with the ratio for cluster space consumption
	logger.Info("cluster_space_usage", logkeys.ClusterId, clusterId, logkeys.ClusterType, "vast", logkeys.Ratio, clusterUsageRatio)
	metricClusterUsage.WithLabelValues(clusterId, "vast").Set(clusterUsageRatio)
	logger.Info("cluster_space_total", logkeys.ClusterId, clusterId, logkeys.ClusterType, "vast", logkeys.ClusterSpaceTotal, totalCapacity)
	metricClusterSpaceTotal.WithLabelValues(clusterId, "vast").Set(float64(totalCapacity))
	logger.Info("cluster_space_available", logkeys.ClusterId, logkeys.ClusterType, "vast", clusterId, logkeys.ClusterSpaceAvailable, availCapacity)
	metricClusterSpaceAvailable.WithLabelValues(clusterId, "vast").Set(float64(availCapacity))
	logger.Info("------------------------------------------------------------")
	// Update the metric with NS space usage
	for _, ns := range nsList {
		usage, err := strconv.ParseFloat(ns.Properties.Quota, 64)
		if err != nil {
			logger.Error(err, "error converting string to float64")
		}
		usage = float64(int(usage) / int(gb)) //convert to GB
		logger.Info("vast_org_space_usage in Gb", logkeys.ClusterId, clusterId, logkeys.Namespace, ns.Metadata.Name, logkeys.Usage, usage)
		metricVastNSUsage.With(prometheus.Labels{"cluster": clusterId, "namespace": ns.Metadata.Name}).Set(usage)
	}
	logger.Info("Vast NS Metrics update complete")
}

// update weka filesystem relation metrics
func updateVastFSMetric(ctx context.Context, cl storagecontroller.ClusterInfo, fsList map[string]int, files []*vastInfo) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricServiceUtil.updateVastFSMetric")
	clusterId := cl.UUID
	// Check if the custom metric for the cluster exists
	logger.Info("Begin update Vast FS metrics")
	fsCount := len(files)
	// Update total # fs per cluster
	metricVastFSCount.With(prometheus.Labels{"cluster": clusterId}).Set(float64(fsCount))
	logger.Info("vast_cluster_filesystems_count", logkeys.ClusterId, clusterId, logkeys.FilesystemCount, fsCount)
	metricVastFSRatio.With(prometheus.Labels{"cluster": clusterId}).Set(float64(fsCount) / 1024)
	logger.Info("vast_cluster_filesystems_ratio", logkeys.ClusterId, clusterId, logkeys.Ratio, float64(fsCount)/1024)
	logger.Info("------------------------------------------------------------")
	for ns, fs := range fsList {
		metricVastFS.With(prometheus.Labels{"cluster": clusterId, "namespace": ns}).Set(float64(fs))
		logger.Info("vast_cluster_filesystems_per_org", logkeys.ClusterId, clusterId, logkeys.Namespace, ns, logkeys.FilesystemCount, fs)
	}
	logger.Info("Done FS Update")
	for _, fs := range files {
		size := fs.Size / int(gb) // convert to GB
		ns := fs.Namespace
		if !strings.HasPrefix(ns, vastNS) {
			continue
		}
		cloudacc := ns[len(ns)-12:]
		metricVastFSUsage.With(prometheus.Labels{"cluster": clusterId, "cloudaccount": cloudacc, "name": fs.Name}).Set(float64(size))
		logger.Info("vast_filesystem_size", logkeys.ClusterId, clusterId, logkeys.CloudAccount, cloudacc, logkeys.FilesystemName, fs.Name, logkeys.Size, float64(size))
	}
	logger.Info("Done Vast FS Usage Update")
	logger.Info("Vast FS Metrics update complete")
}

// update minio metrics
func updateBKMetric(ctx context.Context, cl storagecontroller.ClusterInfo, bkUsage map[bucketKey]bucketStat, bkCount map[string]int) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricServiceUtil.updateBKMetric")
	// Check if the custom metric for the cluster exists
	logger.Info("Begin update BK metric")
	gbytes := int(gb)
	clusterId := cl.UUID
	clusterUsageRatio := float64(cl.TotalCapacity-cl.AvailableCapacity) / float64(cl.TotalCapacity)
	totalCapacity := cl.TotalCapacity / gb
	availCapacity := cl.AvailableCapacity / gb
	// Update the metric with the ratio for cluster space consumption
	logger.Info("cluster_space_usage", logkeys.ClusterId, clusterId, logkeys.ClusterType, "minio", logkeys.Ratio, clusterUsageRatio)
	metricClusterUsage.WithLabelValues(clusterId, "minio").Set(clusterUsageRatio)
	logger.Info("cluster_space_total", logkeys.ClusterId, clusterId, logkeys.ClusterType, "minio", logkeys.ClusterSpaceTotal, totalCapacity)
	metricClusterSpaceTotal.WithLabelValues(clusterId, "minio").Set(float64(totalCapacity))
	logger.Info("cluster_space_available", logkeys.ClusterId, clusterId, logkeys.ClusterType, "minio", logkeys.ClusterSpaceAvailable, availCapacity)
	metricClusterSpaceAvailable.WithLabelValues(clusterId, "minio").Set(float64(availCapacity))
	logger.Info("------------------------------------------------------------")
	totalSize, totalCount := 0, 0

	for bk, stats := range bkUsage {
		metricBkUsage.With(prometheus.Labels{"cluster": clusterId, "cloudaccount": bk.ns, "bucketname": bk.id}).Set(float64(stats.Used / gbytes))
		logger.Info("minio_cluster_bucket_usage", logkeys.ClusterId, clusterId, logkeys.CloudAccountId, bk.ns, logkeys.BucketId, bk.id, logkeys.BucketUsage, stats.Used/gbytes)
		metricBkSize.With(prometheus.Labels{"cluster": clusterId, "cloudaccount": bk.ns, "bucketname": bk.id}).Set(float64(stats.Size / gbytes))
		logger.Info("minio_cluster_bucket_size", logkeys.ClusterId, clusterId, logkeys.CloudAccountId, bk.ns, logkeys.BucketId, bk.id, logkeys.BucketUsage, stats.Size/gbytes)
		totalSize += stats.Size
	}
	// update bucket count metric
	for account, count := range bkCount {
		metricBkCount.With(prometheus.Labels{"cluster": clusterId, "cloudaccount": account}).Set(float64(count))
		logger.Info("minio_cluster_buckets_per_org", logkeys.ClusterId, clusterId, logkeys.CloudAccountId, account, logkeys.BucketCount, count)
		totalCount += count
	}
	metricMinioAllocated.With(prometheus.Labels{"cluster": clusterId}).Set(float64(totalSize / gbytes))
	logger.Info("minio_cluster_space_allocated", logkeys.ClusterId, clusterId, logkeys.Size, totalSize/gbytes)
	metricTotalBK.With(prometheus.Labels{"cluster": clusterId}).Set(float64(totalCount))
	logger.Info("minio_cluster_bucket_total_count", logkeys.ClusterId, clusterId, logkeys.BucketCount, totalCount)
	logger.Info("BK Metrics update complete")
}

// utility function to filter volume names by iks prefix
func filterByPrefix(ctx context.Context, fsList []storagecontroller.Filesystem, prefix string) []storagecontroller.Filesystem {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricServiceUtil.filterPrefix")
	iksVolumes := []storagecontroller.Filesystem{}
	for _, fs := range fsList {
		if strings.HasPrefix(fs.Metadata.FileSystemName, prefix) {
			iksVolumes = append(iksVolumes, fs)
		}
	}
	logger.Info("iks", "volumes", iksVolumes)
	return iksVolumes

}

// retrieve namespace credentials from vault
func readSecretsFromStorageKMS(ctx context.Context, kmsClient pb.StorageKMSPrivateServiceClient, secretKeyPath string) (string, string, bool, error) {
	// initialize default return
	user, pass := "", ""
	foundCreds := false
	request := pb.GetSecretRequest{
		KeyPath: secretKeyPath,
	}
	nsCreds, err := kmsClient.Get(ctx, &request)
	// check for errors
	if err != nil {
		return "", "", false, nil
	} else {
		// check if user and pass not empty
		if usr, found := nsCreds.Secrets["username"]; found {
			user = usr
		}
		if password, found := nsCreds.Secrets["password"]; found {
			pass = password
		}
		if user != "" && pass != "" {
			foundCreds = true
		}
	}
	return user, pass, foundCreds, nil
}

// Utility function to delete stale metrics with partial prometheus label matches
func (sm *StorageCustomMetricService) cleanMetrics(ctx context.Context, clusters []string) {
	logger := log.FromContext(ctx).WithName("StorageCustomMetricService.cleanMetrics")
	logger.Info("Begin cleaning stale metrics")
	defer logger.Info("Finished cleaning stale metrics")
	for _, m := range metrics {
		for _, id := range clusters {
			res := m.DeletePartialMatch(prometheus.Labels{"cluster": id})
			if res != 0 {
				logger.Info("deleted metrics", "cluster", id, "count", res)
			}
		}
	}
}
