// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ddi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	helper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
)

const (
	enrollmentEnvNamespace     = "JOBS_NAMESPACE"
	EnrollmentDefaultNamespace = "idcs-enrollment"
	DeviceInfoLabelKey         = "device-info"
	EnrollmentJobNamePrefix    = "bmaas-enrollment"
	DisenrollmentJobNamePrefix = "bmaas-disenrollment"
)

var (
	EnrollmentNamespace string = helper.GetEnv(enrollmentEnvNamespace, EnrollmentDefaultNamespace)
)

type BMaaSEnrollmentData struct {
	AvailabilityZone string `json:"site" binding:"required"`
	DeviceID         int64  `json:"id" binding:"required"`
	DeviceName       string `json:"name" binding:"required"`
	RackName         string `json:"rack" binding:"required"`
	ClusterName      string `json:"cluster" binding:"required"`
	EnrollmentStatus string `json:"bm_enrollment_status" binding:"required"`
}

// Note: Region is added to get the netbox token.
// Once the netbox token is received, enrollment use the
// GetClusterRegion method to get the region from the netbox.

type EnrollmentJobConfig struct {
	PlaybookImage           string                    `json:"jobplaybookimage"`
	Backofflimit            int32                     `json:"jobbackofflimit"`
	ProvisioningDuration    int                       `json:"provisioningDuration"`
	DeprovisioningDuration  int                       `json:"deprovisioningDuration"`
	JobCleanupDelay         int32                     `json:"jobCleanupDelay"`
	VaultAddress            string                    `json:"vaultAddress"`
	VaultAuthPath           string                    `json:"vaultAuthPath"`
	VaultApproleSecretsPath string                    `json:"vaultApproleSecretsPath"`
	VaultAuthRole           string                    `json:"vaultAuthRole"`
	NetboxAddress           string                    `json:"netboxaddress"`
	NetboxSkipTlsVerify     bool                      `json:"netboxSkipTlsVerify"`
	SetBiosPassword         bool                      `json:"setBiosPassword"`
	DhcpProxy               DhcpProxyConfig           `json:"dhcpProxy"`
	MenAndMice              MenAndMiceConfig          `json:"menAndMice"`
	Region                  string                    `json:"enrollmentRegion"`
	ImagePullSecrets        []v1.LocalObjectReference `json:"imagePullSecrets"`
	ComputeApiServerAddress string                    `json:"computeApiServerAddr"`
	SecretCaPemPath         string                    `json:"secretCaPemPath"`
	SecretCertKeyPemPath    string                    `json:"secretCertKeyPemPath"`
}

type DhcpProxyConfig struct {
	Url     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type MenAndMiceConfig struct {
	Url                string `json:"url"`
	Enabled            bool   `json:"enabled"`
	TftpServerIP       string `json:"tftpServerIp"`
	ServerAddress      string `json:"serverAddress"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}

func (j EnrollmentJobConfig) getEnvVar(deviceInfo BMaaSEnrollmentData) []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name:  dcim.DeviceNameEnvVar,
			Value: deviceInfo.DeviceName,
		},
		{
			Name:  dcim.DeviceIdEnvVar,
			Value: strconv.FormatInt(deviceInfo.DeviceID, 10),
		},
		{
			Name:  dcim.RackNameEnvVar,
			Value: deviceInfo.RackName,
		},
		{
			Name:  dcim.ClusterNameEnvVar,
			Value: deviceInfo.ClusterName,
		},
		{
			Name:  secrets.VaultAddressEnvVar,
			Value: j.VaultAddress,
		},
		{
			Name:  dcim.NetBoxAddressEnvVar,
			Value: j.NetboxAddress,
		},
		{
			Name:  dcim.RegionEnvVar,
			Value: j.Region,
		},
		{
			Name:  dcim.AvailabilityZoneEnvVar,
			Value: deviceInfo.AvailabilityZone,
		},
		{
			Name:  tasks.ComputeApiServerAddrEnvVar,
			Value: j.ComputeApiServerAddress,
		},
	}
}

func (j EnrollmentJobConfig) getAnnotations(deviceInfo BMaaSEnrollmentData) map[string]string {
	secretIdTemplate := fmt.Sprintf("{{- with secret \"%s\" -}}{{ .Data.data.secret_id }}{{- end }}", j.VaultApproleSecretsPath)
	roleIdTemplate := fmt.Sprintf("{{- with secret \"%s\" -}}{{ .Data.data.role_id }}{{- end }}", j.VaultApproleSecretsPath)
	templateCaPemTemplate := fmt.Sprintf("{{- with secret \"%s\" -}}{{ .Data.certificate }}{{- end }}", j.SecretCaPemPath)
	templateCertKeyPemTemplate := fmt.Sprintf(`
	{{- with pkiCert "%s-ca/issue/%s-baremetal-enrollment-task" "common_name=%s-baremetal-enrollment-task.idcs-system.svc.cluster.local" "ttl=24h" -}}
	{{ .Data.Cert }}
	{{ .Data.CA }}
	{{ .Data.Key }}
	{{ .Data.Cert | writeToFile "/vault/secrets/cert.pem" "vault" "vault" "0644" }}
	{{ .Data.CA | writeToFile "/vault/secrets/cert.pem" "vault" "vault" "0644" "append" }}
	{{ .Data.Key | writeToFile "/vault/secrets/cert.key" "vault" "vault" "0644" }}
	{{- end }}`, deviceInfo.AvailabilityZone, deviceInfo.AvailabilityZone, deviceInfo.AvailabilityZone)

	return map[string]string{
		"vault.hashicorp.com/agent-init-first":                  "true",
		"vault.hashicorp.com/agent-inject":                      "true",
		"vault.hashicorp.com/agent-inject-status":               "update",
		"vault.hashicorp.com/agent-inject-secret-secret-id":     j.VaultApproleSecretsPath,
		"vault.hashicorp.com/agent-inject-secret-role-id":       j.VaultApproleSecretsPath,
		"vault.hashicorp.com/role":                              j.VaultAuthRole,
		"vault.hashicorp.com/service":                           j.VaultAddress,  // TODO: This should be set by Vault Agent Injector.
		"vault.hashicorp.com/auth-path":                         j.VaultAuthPath, // TODO: This should be set by Vault Agent Injector.
		"vault.hashicorp.com/agent-pre-populate-only":           "true",
		"vault.hashicorp.com/agent-inject-template-secret-id":   secretIdTemplate,
		"vault.hashicorp.com/agent-inject-template-role-id":     roleIdTemplate,
		"vault.hashicorp.com/agent-inject-secret-ca.pem":        j.SecretCaPemPath,
		"vault.hashicorp.com/agent-inject-secret-certkey.pem":   j.SecretCertKeyPemPath,
		"vault.hashicorp.com/agent-inject-template-ca.pem":      templateCaPemTemplate,
		"vault.hashicorp.com/agent-inject-template-certkey.pem": templateCertKeyPemTemplate,
	}
}

func (j EnrollmentJobConfig) CreateEnrollmentJobSpec(deviceInfo BMaaSEnrollmentData) *batchv1.Job {
	jobName := fmt.Sprintf("%s-%s", EnrollmentJobNamePrefix, deviceInfo.DeviceName)

	// env variables
	envVariables := j.getEnvVar(deviceInfo)

	// set dhcp proxy variable if present
	if j.DhcpProxy.Enabled {
		envVariables = append(envVariables, v1.EnvVar{Name: tasks.DhcpProxyUrlEnvVar, Value: j.DhcpProxy.Url})
	}
	// set men and mice variable if present
	if j.MenAndMice.Enabled {
		envVariables = append(envVariables, v1.EnvVar{Name: tasks.MenAndMiceUrlEnvVar, Value: j.MenAndMice.Url})
		envVariables = append(envVariables, v1.EnvVar{Name: tasks.MenAndMiceServerAddressEnvVar, Value: j.MenAndMice.ServerAddress})
		envVariables = append(envVariables, v1.EnvVar{Name: tasks.TftpServerIPEnvVar, Value: j.MenAndMice.TftpServerIP})

		// check skip tls verify for MenAndMice
		if !j.MenAndMice.InsecureSkipVerify {
			envVariables = append(envVariables, v1.EnvVar{Name: ddi.InsecureSkipVerifyEnvVar, Value: "false"})
		}
	}

	// enabled Bios master password
	if j.SetBiosPassword {
		envVariables = append(envVariables, v1.EnvVar{Name: tasks.SetBiosPasswordEnvVar, Value: "true"})
	}
	// check skip tls verify for Netbox
	if !j.NetboxSkipTlsVerify {
		envVariables = append(envVariables, v1.EnvVar{Name: dcim.InsecureSkipVerifyEnvVar, Value: "false"})
	}

	// Add different stages duration and timeouts
	// Calculate time available for inspection

	//Adjust Total Duration and add 10 minutes required for system prep before registering with metal3
	totalDuration := int64(j.ProvisioningDuration + j.DeprovisioningDuration + 600)
	envVariables = append(envVariables, v1.EnvVar{Name: helper.ProvisioningTimeoutVar, Value: strconv.Itoa(j.ProvisioningDuration)})
	envVariables = append(envVariables, v1.EnvVar{Name: helper.DeprovisionTimeoutVar, Value: strconv.Itoa(j.DeprovisioningDuration)})
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: EnrollmentNamespace,
			Labels: map[string]string{
				DeviceInfoLabelKey: jobName,
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					ImagePullSecrets: j.ImagePullSecrets,
					Containers: []v1.Container{
						{
							Name:            EnrollmentJobNamePrefix,
							Image:           j.PlaybookImage,
							ImagePullPolicy: v1.PullAlways,
							Env:             envVariables,
							Args:            []string{"enroll"},
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: fmt.Sprintf("%s-baremetal-enrollment-task", deviceInfo.AvailabilityZone),
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: j.getAnnotations(deviceInfo),
				},
			},
			BackoffLimit:            &j.Backofflimit,
			ActiveDeadlineSeconds:   &totalDuration,
			TTLSecondsAfterFinished: &j.JobCleanupDelay,
		},
	}
}

func (j EnrollmentJobConfig) CreateDisenrollmentJobSpec(deviceInfo BMaaSEnrollmentData) *batchv1.Job {
	jobName := fmt.Sprintf("%s-%s", DisenrollmentJobNamePrefix, deviceInfo.DeviceName)
	var deprovisioningDuration int64
	deprovisioningDuration = int64(j.DeprovisioningDuration)
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: EnrollmentNamespace,
			Labels: map[string]string{
				DeviceInfoLabelKey: jobName,
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					ImagePullSecrets: j.ImagePullSecrets,
					Containers: []v1.Container{
						{
							Name:            DisenrollmentJobNamePrefix,
							Image:           j.PlaybookImage,
							ImagePullPolicy: v1.PullAlways,
							Env:             j.getEnvVar(deviceInfo),
							Args:            []string{"disenroll"},
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: fmt.Sprintf("%s-baremetal-enrollment-task", deviceInfo.AvailabilityZone),
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: j.getAnnotations(deviceInfo),
				},
			},
			BackoffLimit:            &j.Backofflimit,
			ActiveDeadlineSeconds:   &deprovisioningDuration,
			TTLSecondsAfterFinished: &j.JobCleanupDelay,
		},
	}
}

func GetConfig(configFile string) (EnrollmentJobConfig, error) {
	content, err := os.ReadFile(configFile)
	if err != nil {
		return EnrollmentJobConfig{}, err
	}

	var cfg EnrollmentJobConfig
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return EnrollmentJobConfig{}, err
	}

	return cfg, nil

}
