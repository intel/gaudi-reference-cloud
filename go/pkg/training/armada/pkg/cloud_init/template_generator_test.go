// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloud_init

import (
	"context"
	"testing"

	"github.com/flosch/pongo2"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	ctx := context.Background()

	t.Run("file does not exist", func(t *testing.T) {
		_, err := ReadFile(ctx, "cloud-init-scripts/test-files/does-not-exist.yaml")
		assert.Error(t, err)
	})

	t.Run("file exists", func(t *testing.T) {
		contents, err := ReadFile(ctx, "cloud-init-scripts/test-files/basic-file.yaml")
		assert.NoError(t, err)
		assert.Equal(t, "hello: world\n", contents)
	})
}

func TestRenderFile(t *testing.T) {
	ctx := context.Background()

	t.Run("nothing to render", func(t *testing.T) {
		contents := "hello: world\n"
		rendered, err := RenderFile(ctx, contents, nil)
		assert.NoError(t, err)
		assert.Equal(t, contents, rendered)
	})

	t.Run("rendering works", func(t *testing.T) {
		contents := "hello: {{ world }}\n"
		rendered, err := RenderFile(ctx, contents, pongo2.Context{"world": "universe"})
		assert.NoError(t, err)
		assert.Equal(t, "hello: universe\n", rendered)
	})
}

func TestParseConnection(t *testing.T) {
	ctx := context.Background()

	t.Run("instance is nil", func(t *testing.T) {
		connection, err := ParseConnection(ctx, nil, "")
		assert.Error(t, err)
		assert.Nil(t, connection)
	})

	t.Run("no oneapi-instance-role label", func(t *testing.T) {
		instance := &v1.Instance{
			Metadata: &v1.InstanceMetadata{
				Labels: map[string]string{},
			},
		}
		connection, err := ParseConnection(ctx, instance, "")
		assert.Error(t, err)
		assert.Nil(t, connection)
	})

	t.Run("no network interfaces", func(t *testing.T) {
		instance := &v1.Instance{
			Metadata: &v1.InstanceMetadata{
				Labels: map[string]string{"oneapi-instance-role": "test-role"},
			},
			Status: &v1.InstanceStatus{
				Interfaces: []*v1.InstanceInterfaceStatus{},
			},
		}
		connection, err := ParseConnection(ctx, instance, "test-role")
		assert.Error(t, err)
		assert.Nil(t, connection)
	})

	t.Run("network interface with no addresses", func(t *testing.T) {
		instance := &v1.Instance{
			Metadata: &v1.InstanceMetadata{
				Labels: map[string]string{"oneapi-instance-role": "test-role"},
			},
			Status: &v1.InstanceStatus{
				Interfaces: []*v1.InstanceInterfaceStatus{
					{
						Addresses: []string{},
					},
				},
			},
		}
		connection, err := ParseConnection(ctx, instance, "test-role")
		assert.Error(t, err)
		assert.Nil(t, connection)
	})

	t.Run("incorrect role", func(t *testing.T) {
		instance := &v1.Instance{
			Metadata: &v1.InstanceMetadata{
				Labels: map[string]string{"oneapi-instance-role": "test-role"},
			},
			Status: &v1.InstanceStatus{
				Interfaces: []*v1.InstanceInterfaceStatus{
					{
						Addresses: []string{
							"127.0.0.1",
						},
					},
				},
			},
		}
		connection, err := ParseConnection(ctx, instance, "not-test-role")
		assert.Error(t, err)
		assert.Nil(t, connection)
	})

	t.Run("correctly parse connection", func(t *testing.T) {
		hostname := "test-hostname"
		address := "127.0.0.1"

		instance := &v1.Instance{
			Metadata: &v1.InstanceMetadata{
				Name:   hostname,
				Labels: map[string]string{"oneapi-instance-role": "test-role"},
			},
			Status: &v1.InstanceStatus{
				Interfaces: []*v1.InstanceInterfaceStatus{
					{
						Addresses: []string{
							address,
						},
					},
				},
			},
		}
		connection, err := ParseConnection(ctx, instance, "test-role")
		assert.NoError(t, err)
		assert.Equal(t, connection.Address, address)
		assert.Equal(t, connection.Hostname, hostname)
	})
}

func TestRenderCommonCloudConfigs(t *testing.T) {
	ctx := context.Background()

	t.Run("no weka storages", func(t *testing.T) {
		// inject filenames
		COMMON_LDAP_FILES_TMPL_FILENAME = "cloud-init-scripts/test-files/ldap-files.yaml"
		COMMON_LDAP_RUNCMD_TMPL_FILENAME = "cloud-init-scripts/test-files/ldap-cmd.yaml"
		COMMON_MUNGE_FILES_TMPL_FILENAME = "cloud-init-scripts/test-files/munge-files.yaml"
		COMMON_MUNGE_RUNCMD_TMPL_FILENAME = "cloud-init-scripts/test-files/munge-cmd.yaml"

		commonCloudConfig := CommonCloudConfigs{
			ClusterId: "test-cluster-id",
		}
		wekaStorages := []*v1.FilesystemPrivate{}
		localMountDirs := make(map[string]string)
		remoteMoundDirs := make(map[string]string)

		// call the function under test
		err := commonCloudConfig.RenderCommonCloudConfigs(ctx, wekaStorages, localMountDirs, remoteMoundDirs)
		assert.NoError(t, err)

		// verify weka storage is empty
		assert.Equal(t, commonCloudConfig.WekaStorage, "")

		// verify the rendered cloud configs
		assert.Equal(t, commonCloudConfig.LdapFiles, "ldap: files\n")
		assert.Equal(t, commonCloudConfig.LdapRunCmd, "ldap: cmd\n")
		assert.Equal(t, commonCloudConfig.MungeFiles, "munge: files\n")
		assert.Equal(t, commonCloudConfig.MungeRunCmd, "munge: cmd\n")
	})

	t.Run("one weka storage provided", func(t *testing.T) {
		// inject filenames
		COMMON_WEKA_RUNCMD_TMPL_FILENAME = "cloud-init-scripts/test-files/weka-cmd.yaml"
		COMMON_LDAP_FILES_TMPL_FILENAME = "cloud-init-scripts/test-files/ldap-files.yaml"
		COMMON_LDAP_RUNCMD_TMPL_FILENAME = "cloud-init-scripts/test-files/ldap-cmd.yaml"
		COMMON_MUNGE_FILES_TMPL_FILENAME = "cloud-init-scripts/test-files/munge-files.yaml"
		COMMON_MUNGE_RUNCMD_TMPL_FILENAME = "cloud-init-scripts/test-files/munge-cmd.yaml"

		commonCloudConfig := CommonCloudConfigs{
			ClusterId: "test-cluster-id",
		}
		wekaStorages := []*v1.FilesystemPrivate{
			{
				Metadata: &v1.FilesystemMetadataPrivate{
					ResourceId: "test-resource-id",
				},
				Spec: &v1.FilesystemSpecPrivate{},
				Status: &v1.FilesystemStatusPrivate{
					Mount: &v1.FilesystemMountStatusPrivate{
						ClusterAddr: "test-cluster-addr",
						Namespace:   "test-namespace",
					},
					User: &v1.FilesystemUserStatusPrivate{
						User:     "test-user",
						Password: "test-password",
					},
				},
			},
		}
		localMountDirs := make(map[string]string)
		remoteMoundDirs := make(map[string]string)

		localMountDirs["test-resource-id"] = "test-local-mount-dir"
		remoteMoundDirs["test-resource-id"] = "test-remote-mount-dir"

		// call the function under test
		err := commonCloudConfig.RenderCommonCloudConfigs(ctx, wekaStorages, localMountDirs, remoteMoundDirs)
		assert.NoError(t, err)

		// verify weka storage contains storage values
		assert.Contains(t, commonCloudConfig.WekaStorage, "installWekaClient: test-cluster-addr/dist/v1/install\n")
		assert.Contains(t, commonCloudConfig.WekaStorage, "wekaClusterAddr: test-cluster-addr\n")
		assert.Contains(t, commonCloudConfig.WekaStorage, "wekaNamespace: test-namespace\n")
		assert.Contains(t, commonCloudConfig.WekaStorage, "user: test-user\n")
		assert.Contains(t, commonCloudConfig.WekaStorage, "password: test-password\n")
		assert.Contains(t, commonCloudConfig.WekaStorage, "localMountDir: test-local-mount-dir\n")
		assert.Contains(t, commonCloudConfig.WekaStorage, "remoteMountPoint: test-cluster-addr/test-remote-mount-dir\n")

		// verify the rendered cloud configs
		assert.Equal(t, commonCloudConfig.LdapFiles, "ldap: files\n")
		assert.Equal(t, commonCloudConfig.LdapRunCmd, "ldap: cmd\n")
		assert.Equal(t, commonCloudConfig.MungeFiles, "munge: files\n")
		assert.Equal(t, commonCloudConfig.MungeRunCmd, "munge: cmd\n")
	})
}

func TestRenderCloudConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid cloud config", func(t *testing.T) {
		pongoContext := pongo2.Context{}
		rendered, err := RenderCloudConfig(ctx, "cloud-init-scripts/test-files/invalid-cloud-config.yaml", pongoContext)
		assert.Error(t, err)
		assert.Equal(t, "invalid: cloud-config\n", rendered)
	})

	t.Run("valid cloud config", func(t *testing.T) {
		pongoContext := pongo2.Context{"world": "universe"}
		rendered, err := RenderCloudConfig(ctx, "cloud-init-scripts/test-files/valid-cloud-config.yaml", pongoContext)
		assert.NoError(t, err)
		assert.Equal(t, "#cloud-init\npackages:\n  - nginx\n", rendered)
	})
}

func TestRenderSlurmctldCloudConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("no slurm instances", func(t *testing.T) {
		// inject filename
		SLURMCTLD_CLOUDINIT_TMPL_FILENAME = "cloud-init-scripts/test-files/slurmctld-cloudinit.yaml"

		config := SlurmctldCloudConfig{
			Common: &CommonCloudConfigs{
				ClusterId: "test-cluster-id",
			},
			NodeName:      "test-node-name",
			PartitionName: "test-partition-name",
		}
		slurmInstances := []*v1.Instance{}

		// call the function under test
		rendered, err := RenderSlurmctldCloudConfig(ctx, &config, slurmInstances)
		assert.NoError(t, err)
		assert.Contains(t, rendered, "computeHostInstance: \n")
	})

	t.Run("parse connection error, incorrect label", func(t *testing.T) {
		// inject filename
		SLURMCTLD_CLOUDINIT_TMPL_FILENAME = "cloud-init-scripts/test-files/slurmctld-cloudinit.yaml"

		config := SlurmctldCloudConfig{
			Common: &CommonCloudConfigs{
				ClusterId: "test-cluster-id",
			},
			NodeName:      "test-node-name",
			PartitionName: "test-partition-name",
		}
		slurmInstances := []*v1.Instance{{
			Metadata: &v1.InstanceMetadata{
				Name:   "test-hostname",
				Labels: map[string]string{"oneapi-instance-role": "incorrect-role-name"},
			},
			Status: &v1.InstanceStatus{
				Interfaces: []*v1.InstanceInterfaceStatus{
					{
						Addresses: []string{
							"127.0.0.1",
						},
					},
				},
			},
		}}

		// call the function under test
		rendered, err := RenderSlurmctldCloudConfig(ctx, &config, slurmInstances)
		assert.Error(t, err)
		assert.Equal(t, "", rendered)
	})

	t.Run("provided slurm instances", func(t *testing.T) {
		// inject filename
		SLURMCTLD_CLOUDINIT_TMPL_FILENAME = "cloud-init-scripts/test-files/slurmctld-cloudinit.yaml"

		config := SlurmctldCloudConfig{
			Common: &CommonCloudConfigs{
				ClusterId: "test-cluster-id",
			},
			NodeName:      "test-node-name",
			PartitionName: "test-partition-name",
		}
		slurmInstances := []*v1.Instance{{
			Metadata: &v1.InstanceMetadata{
				Name:   "test-hostname",
				Labels: map[string]string{"oneapi-instance-role": "slurm-compute-node"},
			},
			Status: &v1.InstanceStatus{
				Interfaces: []*v1.InstanceInterfaceStatus{
					{
						Addresses: []string{
							"127.0.0.1",
						},
					},
				},
			},
		}}

		// call the function under test
		rendered, err := RenderSlurmctldCloudConfig(ctx, &config, slurmInstances)
		assert.NoError(t, err)
		assert.Contains(t, rendered, "computeHostInstance: 127.0.0.1 test-hostname\n")
	})
}
