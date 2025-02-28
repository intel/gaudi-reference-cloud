// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"bytes"
	"context"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"strings"
	"time"
)

// Config holds the required information for the operator and its controllers.
type Config struct {
	MonitorPeriodicity               time.Duration            `json:"monitorPeriodicity"`
	MonitorGracePeriod               time.Duration            `json:"monitorGracePeriod"`
	MonitorGracePeriodByInstanceType map[string]time.Duration `json:"monitorGracePeriodByInstanceType"`
	CertExpirations                  CertExpiration           `json:"certExpirations"`
	KubernetesProviders              KubernetesProviders      `json:"kubernetesProviders"`
	NodeProviders                    NodeProviders            `json:"nodeProviders"`
	S3Addons                         S3Addons                 `json:"s3Addons"`
	S3Snapshots                      S3Snapshots              `json:"s3Snapshots"`
	ClusterMaxConcurrentReconciles   int                      `json:"clusterMaxConcurrentReconciles"`
	NodegroupMaxConcurrentReconciles int                      `json:"nodegroupMaxConcurrentReconciles"`
	AddonMaxConcurrentReconciles     int                      `json:"addonMaxConcurrentReconciles"`
	IPTables                         IPTables                 `json:"iptables"`
	Logging                          Logging                  `json:"logging,omitempty"`
	Monitoring                       *Monitoring              `json:"monitoring,omitempty"`
	Metrics                          *Metrics                 `json:"metrics,omitempty"`
	IdcGrpcUrl                       string                   `json:"idcGrpcUrl"`
	Weka                             Weka                     `json:"weka"`
}

type Metrics struct {
	SystemMetrics  *SystemMetrics  `json:"systemMetrics,omitempty"`
	EndUserMetrics *EndUserMetrics `json:"endUserMetrics,omitempty"`
}

type EndUserMetrics struct {
	PrometheusRemoteWrite *PrometheusRemoteWrite `json:"prometheusRemoteWrite,omitempty"`
}

type SystemMetrics struct {
	PrometheusRemoteWrite *PrometheusRemoteWrite `json:"prometheusRemoteWrite,omitempty"`
}

type PrometheusRemoteWrite struct {
	Url         string     `json:"url"`
	BearerToken string     `json:"bearerToken,omitempty"`
	BasicAuth   *BasicAuth `json:"basicAuth,omitempty"`
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type IPTables struct {
	Enabled          bool   `json:"enabled"`
	ControlplaneCIDR string `json:"controlplaneCIDR"`
}

// Deprecated: Use Metrics instead.
type Monitoring struct {
	Enabled        bool   `json:"enabled"`
	RemoteWriteURL string `json:"remoteWriteURL"`
	Username       string `json:"username"`
	Password       string `json:"password"`
}

type Logging struct {
	Enabled  bool   `json:"enabled,omitempty"`
	Host     string `json:"host,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

type KubernetesProviders struct {
	RKE2 RKE2KubernetesProvider `json:"rke2"`
	IKS  IKSKubernetesProvider  `json:"iks"`
}

type RKE2KubernetesProvider struct {
	URL                         string `json:"url"`
	AccessKey                   string `json:"accessKey"`
	SecretKey                   string `json:"secretKey"`
	ControlplaneBootstrapScript string `json:"controlplaneBootstrapScript"`
	WorkerBootstrapScript       string `json:"workerBootstrapScript"`
}

type IKSKubernetesProvider struct {
	ControlplaneBootstrapScript string `json:"controlplaneBootstrapScript"`
	WorkerBootstrapScript       string `json:"workerBootstrapScript"`
}

type NodeProviders struct {
	Harvester HarvesterNodeProvider `json:"harvester"`
	Compute   ComputeNodeProvider   `json:"compute"`
}

type HarvesterNodeProvider struct {
	URL       string `json:"url"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type ComputeNodeProvider struct {
	URL string `json:"url"`
}

type S3Snapshots struct {
	Enabled              bool          `json:"enabled"`
	SnapshotsPeriodicity time.Duration `json:"snapshotsPeriodicity"`
	SnapshotsFolder      string        `json:"snapshotsFolder"`
	URL                  string        `json:"url"`
	AccessKey            string        `json:"accessKey"`
	SecretKey            string        `json:"secretKey"`
	UseSSL               bool          `json:"useSSL"`
	BucketName           string        `json:"bucketName"`
	ContentType          string        `json:"contentType"`
	S3Path               string        `json:"s3Path"`
}

type S3Addons struct {
	URL        string `json:"url"`
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	UseSSL     bool   `json:"useSSL"`
	BucketName string `json:"bucketName"`
	S3Path     string `json:"s3Path"`
}

type CertExpiration struct {
	CaCertExpirationPeriod           time.Duration `json:"caCertExpirationPeriod"`
	ControlPlaneCertExpirationPeriod time.Duration `json:"controlPlaneCertExpirationPeriod"`
	ControllerCertExpirationPeriod   time.Duration `json:"controllerCertExpirationPeriod"`
}

type Weka struct {
	SoftwareVersion         string                      `json:"softwareVersion"`
	ClusterUrl              string                      `json:"clusterUrl"`
	ClusterPort             string                      `json:"clusterPort"`
	Scheme                  string                      `json:"scheme"`
	FilesystemGroupName     string                      `json:"filesystemGroupName"`
	InitialFilesystemSizeGB string                      `json:"initialFilesystemSizeGB"`
	ReclaimPolicy           string                      `json:"reclaimPolicy"`
	HelmchartRepoUrl        string                      `json:"helmchartRepoUrl"`
	HelmchartName           string                      `json:"helmchartName"`
	InstanceTypes           map[string]WekaInstanceType `json:"instanceTypes"`
	RecreateGracePeriod     time.Duration               `json:"recreateGracePeriod"`
}

type WekaInstanceType struct {
	Mode     string `json:"mode"`
	NumCores string `json:"numCores"`
}

// EnrichString convert map[string]string to comma seperated key=value string
func (l *Logging) EnrichString(ctx context.Context, data map[string]string) string {
	var b bytes.Buffer
	for k, v := range data {
		if _, err := fmt.Fprintf(&b, "%s=%s,", k, v); err != nil {
			log.FromContext(ctx).
				WithName("logging.EnrichString").
				Error(err, "failed to load logging enrichment context configs")
		}
	}
	return strings.TrimSuffix(b.String(), ",") // remove last comma
}

// ConvertMonitoringToSystemMetrics transfers data from the deprecated Monitoring field to the SystemMetrics field in Metrics.
func (c *Config) ConvertMonitoringToSystemMetrics() {
	if c.Monitoring == nil {
		return
	}

	// metrics configuration in use already
	if c.Metrics != nil && c.Metrics.SystemMetrics != nil {
		return
	}

	if c.Monitoring.Enabled {
		if c.Metrics == nil {
			c.Metrics = &Metrics{}
		}

		if c.Metrics.SystemMetrics == nil {
			c.Metrics.SystemMetrics = &SystemMetrics{}
		}

		c.Metrics.SystemMetrics.PrometheusRemoteWrite = &PrometheusRemoteWrite{
			Url: c.Monitoring.RemoteWriteURL,
			BasicAuth: &BasicAuth{
				Username: c.Monitoring.Username,
				Password: c.Monitoring.Password,
			},
		}
	}

}
