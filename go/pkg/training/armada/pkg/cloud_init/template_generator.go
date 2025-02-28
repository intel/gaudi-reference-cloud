// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloud_init

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	// Common cloud init configuration template files
	COMMON_WEKA_RUNCMD_TMPL_FILENAME  = "/training/cloud-init-scripts/common/weka/weka-cmd.yaml"
	COMMON_LDAP_FILES_TMPL_FILENAME   = "/training/cloud-init-scripts/common/ldap-client/ldap-files.yaml"
	COMMON_LDAP_RUNCMD_TMPL_FILENAME  = "/training/cloud-init-scripts/common/ldap-client/ldap-cmd.yaml"
	COMMON_MUNGE_FILES_TMPL_FILENAME  = "/training/cloud-init-scripts/common/slurm/munge/munge-files.yaml"
	COMMON_MUNGE_RUNCMD_TMPL_FILENAME = "/training/cloud-init-scripts/common/slurm/munge/munge-cmd.yaml"

	// Core cloud init configuration template files
	SLURMD_CLOUDINIT_TMPL_FILENAME     = "/training/cloud-init-scripts/compute-node.yaml"
	SLURMCTLD_CLOUDINIT_TMPL_FILENAME  = "/training/cloud-init-scripts/slurmctld-node.yaml"
	LOGIN_CLOUDINIT_TMPL_FILENAME      = "/training/cloud-init-scripts/login-node.yaml"
	JUPYTERHUB_CLOUDINIT_TMPL_FILENAME = "/training/cloud-init-scripts/jupyterhub-node.yaml"
)

func validateCloudConfig(osname string, userData string) (bool, error) {
	if !regexp.MustCompile("^#cloud-init\n").MatchString(userData) {
		return false, fmt.Errorf("Missing #cloud-init header...")
	}

	newCloudConfig, err := util.NewCloudConfig(osname, userData)
	if err != nil {
		return false, err
	}

	_, err = newCloudConfig.RenderYAML()
	if err != nil {
		return false, err
	}

	return true, nil
}

func ReadFile(ctx context.Context, templateFilename string) (string, error) {
	log := log.FromContext(context.Background()).WithName("TemplateGenerator.ReadFile")
	log.Info("Reading template file provided", "filename", templateFilename)

	template, err := os.ReadFile(templateFilename)
	if err != nil {
		return "", fmt.Errorf("Error cannot read file")
	}

	return string(template), nil
}

func RenderFile(ctx context.Context, rawString string, cloudConfig pongo2.Context) (string, error) {
	log := log.FromContext(context.Background()).WithName("TemplateGenerator.RenderFile")
	log.Info("Rendering template...")

	templateEngine, err := pongo2.FromString(rawString)
	if err != nil {
		return "", err
	}

	renderedTemplate, err := templateEngine.Execute(cloudConfig)
	if err != nil {
		return "", err
	}

	return renderedTemplate, nil
}

type InstanceConnect struct {
	Hostname string
	Address  string
}

func ParseConnection(ctx context.Context, instance *v1.Instance, roleName string) (*InstanceConnect, error) {
	log := log.FromContext(ctx).WithName("TemplateGenerator.parseConnection")

	if instance == nil {
		return nil, fmt.Errorf("Instance is nil")
	}

	connectOpt := InstanceConnect{}

	labels := instance.Metadata.GetLabels()
	if role, found := labels["oneapi-instance-role"]; found {
		interfaces := instance.Status.GetInterfaces()
		if len(interfaces) != 0 {
			addresses := interfaces[0].GetAddresses()
			if len(addresses) == 0 {
				log.Info("no network address found", "instanceId", instance.Metadata.GetResourceId())
				return nil, fmt.Errorf("no network address")
			}

			log.Info("debug", "interface", addresses, "role", role)

			if strings.EqualFold(role, roleName) {
				connectOpt.Hostname = instance.Metadata.Name
				connectOpt.Address = addresses[0]
			} else {
				log.Info("cannot find network interface", "role", roleName)
				return nil, fmt.Errorf("cannot find network interface")
			}
		} else {
			log.Info("no network interface found", "instanceId", instance.Metadata.GetResourceId())
			return nil, fmt.Errorf("no network interface")
		}
	} else {
		log.Info("cannot find oneapi-instance-role label", "instanceId", instance.Metadata.GetResourceId())
		return nil, fmt.Errorf("cannot find instance role label")
	}

	return &connectOpt, nil
}

type WekaStorageCloudConfigs struct {
	InstallWekaClient string
	WekaClusterAddr   string
	WekaNamespace     string
	User              string
	Password          string
	LocalMountDir     string
	RemoteMountPoint  string
}

type CommonCloudConfigs struct {
	ClusterId   string
	WekaStorage string
	NfsMount    string
	NfsRunCmd   string
	LdapFiles   string
	LdapRunCmd  string
	MungeFiles  string
	MungeRunCmd string
}

func (commonCloudCfg *CommonCloudConfigs) RenderCommonCloudConfigs(ctx context.Context, wekaStorages []*v1.FilesystemPrivate, localMountDirs, remoteMountDirs map[string]string) error {
	log := log.FromContext(ctx).WithName("TemplateGenerator.RenderCommonCloudConfigs")

	log.Info("initializing template rendering options...")
	if err := InitTemplateEngineOptions(); err != nil {
		return err
	}

	wekaStorageFS := []string{}
	for idx, fs := range wekaStorages {
		log.Info("start generating storage templates", "storage index", idx)

		log.Info("parsing common weka templates", "file", COMMON_WEKA_RUNCMD_TMPL_FILENAME)
		wekaRunCmdTemplate, err := ReadFile(ctx, COMMON_WEKA_RUNCMD_TMPL_FILENAME)
		if err != nil {
			return err
		}

		wekaMountInfo := fs.Status.GetMount()

		// Remove forward trailing slashes
		sanitizeRemoteMountDir := remoteMountDirs[fs.Metadata.GetResourceId()]
		sanitizeRemoteMountDir = strings.TrimLeft(sanitizeRemoteMountDir, "/")

		// Format weka mount url cli command (i.e. storage-system-name.us-dev-1.cloud.intel.com/test)
		sanitizeRemoteMountPoint := fmt.Sprintf("%s/%s", wekaMountInfo.GetClusterAddr(), sanitizeRemoteMountDir)

		commonWekaStorageValues := WekaStorageCloudConfigs{
			InstallWekaClient: fmt.Sprintf("%s/dist/v1/install", wekaMountInfo.GetClusterAddr()),
			WekaClusterAddr:   wekaMountInfo.GetClusterAddr(),
			WekaNamespace:     wekaMountInfo.GetNamespace(),
			User:              fs.Status.User.GetUser(),
			Password:          fs.Status.User.GetPassword(),
			LocalMountDir:     localMountDirs[fs.Metadata.GetResourceId()],
			RemoteMountPoint:  sanitizeRemoteMountPoint,
		}

		log.Info("rendering storage cloud configs", "storage index", idx)
		generatedWekaSetup, err := RenderFile(ctx, wekaRunCmdTemplate, pongo2.Context{"Values": commonWekaStorageValues})
		if err != nil {
			return err
		}

		wekaStorageFS = append(wekaStorageFS, generatedWekaSetup)
	}

	commonCloudCfg.WekaStorage = strings.Join(wekaStorageFS, "\n")

	log.Info("parsing common ldap-client templates", "file", COMMON_LDAP_RUNCMD_TMPL_FILENAME)
	ldapRunCommands, err := ReadFile(ctx, COMMON_LDAP_RUNCMD_TMPL_FILENAME)
	if err != nil {
		return err
	}

	log.Info("parsing common ldap-client templates", "file", COMMON_LDAP_FILES_TMPL_FILENAME)
	ldapFilesTemplate, err := ReadFile(ctx, COMMON_LDAP_FILES_TMPL_FILENAME)
	if err != nil {
		return err
	}

	log.Info("rendering ldap-client cloud configs")
	ldapFiles, err := RenderFile(ctx, ldapFilesTemplate, pongo2.Context{"Values": commonCloudCfg})
	if err != nil {
		return err
	}

	commonCloudCfg.LdapFiles = ldapFiles
	commonCloudCfg.LdapRunCmd = ldapRunCommands

	log.Info("parsing common munge templates", "file", COMMON_MUNGE_FILES_TMPL_FILENAME)
	mungeFiles, err := ReadFile(ctx, COMMON_MUNGE_FILES_TMPL_FILENAME)
	if err != nil {
		return err
	}

	log.Info("parsing common munge templates", "file", COMMON_MUNGE_RUNCMD_TMPL_FILENAME)
	mungeCommands, err := ReadFile(ctx, COMMON_MUNGE_RUNCMD_TMPL_FILENAME)
	if err != nil {
		return err
	}

	commonCloudCfg.MungeRunCmd = mungeCommands
	commonCloudCfg.MungeFiles = mungeFiles

	defer log.Info("successfully rendered common cloud configs...")
	return nil
}

type JupyterHubCloudConfig struct {
	Common *CommonCloudConfigs
}

type SlurmdCloudConfig struct {
	Common *CommonCloudConfigs
}

type SlurmctldCloudConfig struct {
	Common              *CommonCloudConfigs
	NodeName            string
	PartitionName       string
	ComputeHostInstance string
}

type LoginCloudConfig struct {
	Common *CommonCloudConfigs
}

func RenderCloudConfig(ctx context.Context, cloudConfigFilename string, cloudConfig pongo2.Context) (string, error) {
	log := log.FromContext(context.Background()).WithName("TemplateGenerator.RenderCloudConfig")
	defer log.Info("Exiting cloud-init configuration generation...", "file", cloudConfigFilename)

	renderedCloudConfigTemplate := ""

	rawStringTemplate, err := ReadFile(ctx, cloudConfigFilename)
	if err != nil {
		return renderedCloudConfigTemplate, err
	}

	renderedCloudConfigTemplate, err = RenderFile(ctx, rawStringTemplate, cloudConfig)
	if err != nil {
		return renderedCloudConfigTemplate, err
	}

	_, err = validateCloudConfig("ubuntu", renderedCloudConfigTemplate)
	if err != nil {
		return renderedCloudConfigTemplate, err
	}

	return renderedCloudConfigTemplate, nil
}

func RenderSlurmctldCloudConfig(ctx context.Context, slurmctldCloudConfig *SlurmctldCloudConfig, slurmdInstances []*v1.Instance) (string, error) {
	log := log.FromContext(context.Background()).WithName("TemplateGenerator.RenderSlurmctldCloudConfig")
	defer log.Info("Exiting cloud-init configuration generation...", "file", SLURMCTLD_CLOUDINIT_TMPL_FILENAME)

	log.Info("Being parsing all slurmd ready instances for hostname and ip addresses...")
	slurmdEtcHost := ""
	for _, instance := range slurmdInstances {
		slurmdConnection, err := ParseConnection(ctx, instance, "slurm-compute-node")
		if err != nil {
			return "", err
		}

		slurmdEtcHost += fmt.Sprintf("%s %s\n", slurmdConnection.Address, slurmdConnection.Hostname)
	}

	slurmctldCloudConfig.ComputeHostInstance = slurmdEtcHost

	serializedUserData, err := RenderCloudConfig(ctx, SLURMCTLD_CLOUDINIT_TMPL_FILENAME, pongo2.Context{"Values": slurmctldCloudConfig})
	if err != nil {
		return "", err
	}

	return serializedUserData, nil
}
