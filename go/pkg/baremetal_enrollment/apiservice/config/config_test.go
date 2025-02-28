// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)
	expectedDefaultConfig := EnrollmentJobConfig{
		PlaybookImage:           "baremetal-enrollment:latest",
		Backofflimit:            2,
		ProvisioningDuration:    3600,
		DeprovisioningDuration:  3600,
		JobCleanupDelay:         30,
		VaultAddress:            "http://vault.vault.svc.cluster.local:8200",
		VaultAuthPath:           "auth/cluster-auth",
		VaultApproleSecretsPath: "controlplane/data/us-dev-1/baremetal/enrollment/approle",
		VaultAuthRole:           "us-dev-1-enrollment-role",
		NetboxAddress:           "netbox.idcs-enrollment.svc.cluster.local",
		DhcpProxy: DhcpProxyConfig{
			Url:     "http://172.18.0.2:50100",
			Enabled: true,
		},
		MenAndMice: MenAndMiceConfig{
			Url:          "https://1.2.3.5/",
			Enabled:      false,
			TftpServerIP: "1.1.1.1",
		},
		Region: "us-dev-1",
	}
	testcases := []struct {
		name          string
		configFile    string
		expectSuccess bool
	}{
		{
			name:          "get_config_with_config_file_present",
			configFile:    "./config_test.json",
			expectSuccess: true,
		},
		{
			name:          "get_config_with_config_file_missing",
			configFile:    "./config_test_tmp.json",
			expectSuccess: false,
		},
	}
	for _, test := range testcases {
		config, err := GetConfig(test.configFile)
		if test.expectSuccess {
			assert.Equal(expectedDefaultConfig, config, "Got expected default config")
		} else {
			assert.NotNil(err)
		}

	}
}
