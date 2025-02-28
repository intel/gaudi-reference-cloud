// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package provider

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/flosch/pongo2"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/armada/pkg/ansible"
	cloudGen "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/armada/pkg/cloud_init"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/armada/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/common"
	trainingConfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/database/query"
	idcComputeSvc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/idc_compute"
	idcStorageSvc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/idc_storage"
)

type ClusterProvisionScheduler struct {
	syncTicker    *time.Ticker
	db            *sql.DB
	globalCfg     *config.SchedulerGlobalConfig
	ComputeClient *idcComputeSvc.IDCServiceClient
	StorageClient *idcStorageSvc.IDCStorageServiceClient
}

// const (
// 	// TODO: Make this dynamic accepting it via the API body once it has been verified working.
// 	defaultVnet = "us-staging-1a-default"
// )

func NewClusterProvisionScheduler(ctx context.Context, db *sql.DB, cfg *config.Config) (*ClusterProvisionScheduler, error) {
	logger := log.FromContext(ctx)
	if db == nil {
		return nil, fmt.Errorf("db is requied")
	}
	idcConfig := trainingConfig.IdcConfig{
		ComputeGrpcAPIEndpoint:    cfg.CloudConfig.IDC.ComputeGrpcAPIEndpoint,
		FoundationGrpcAPIEndpoint: cfg.CloudConfig.IDC.FoundationGrpcAPIEndpoint,
		StorageGrpcAPIEndpoint:    cfg.CloudConfig.IDC.StorageGrpcAPIEndpoint,
		AvailabilityZone:          cfg.CloudConfig.IDC.AvailabilityZone,
		Region:                    cfg.CloudConfig.IDC.Region,
	}

	computeClient, err := idcComputeSvc.NewIDCComputeServiceClient(ctx, &idcConfig)
	if err != nil {
		logger.Error(err, "failed to initialize compute client")
		return nil, fmt.Errorf("error connecting to compute client")
	}

	storageClient, err := idcStorageSvc.NewIDCStorageServiceClient(ctx, &idcConfig)
	if err != nil {
		logger.Error(err, "failed to initialize storage client")
		return nil, fmt.Errorf("error connecting to storage client")
	}

	return &ClusterProvisionScheduler{
		syncTicker:    time.NewTicker(time.Duration(cfg.ClusterConfig.SchedulerInterval) * time.Second),
		db:            db,
		globalCfg:     &cfg.ClusterConfig,
		ComputeClient: computeClient,
		StorageClient: storageClient,
	}, nil
}

func (clusterSchd *ClusterProvisionScheduler) StartClusterProvisionScheduler(ctx context.Context) {
	log := log.FromContext(ctx).WithName("ClusterProvisionScheduler.StartClusterProvisionScheduler")
	log.Info("start cluster provisioning scheduler")
	clusterSchd.ProvisionLoop(ctx)
}

func (clusterSchd *ClusterProvisionScheduler) ProvisionLoop(ctx context.Context) {
	log := log.FromContext(ctx).WithName("ClusterProvisionScheduler.ProvisionLoop")
	log.Info("cluster provisioning")
	log.Info("debug", "global config", clusterSchd.globalCfg)
	for {
		clusterSchd.ProvisionCluster(ctx)
		tm := <-clusterSchd.syncTicker.C
		if tm.IsZero() {
			return
		}
	}
}

func (clusterSchd *ClusterProvisionScheduler) ProvisionCluster(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("ClusterProvisionScheduler.ProvisionCluster")
	logger.Info("entering a new cluster provisioning request scan")

	clusterReq, err := query.GetNextClusterRequest(ctx, clusterSchd.db)
	if err != nil {
		logger.Error(err, "error reading cluster request")
		return
	}
	if clusterReq == nil {
		logger.Info("cluster request not found, returning from reconciler")
		return
	}

	logger.Info("new cluster request", "configurations", clusterReq)

	// Tracker
	instanceCreated := []idcComputeSvc.InstanceCreateRequest{}

	clusterReqCounts := map[pb.NodeRole]int{
		pb.NodeRole_JUPYTERHUB_NODE: 0,
		pb.NodeRole_COMPUTE_NODE:    0,
		pb.NodeRole_CONTROLLER_NODE: 0,
		pb.NodeRole_LOGIN_NODE:      0,
	}
	for _, node := range clusterReq.GetNodes() {
		clusterReqCounts[node.Role]++
	}

	if err := query.UpdateClusterState(ctx, clusterSchd.db, clusterReq.ClusterId, clusterReq.CloudAccountId, "ACCEPTED"); err != nil {
		logger.Error(err, "error updating cluster state to 'ACCEPTED', skipping this cluster request")
		return
	}

	// Check if vnet exists, if not then create new
	vnetExists, err := clusterSchd.ComputeClient.IsVNetExists(ctx, clusterReq.GetCloudAccountId(), clusterReq.GetName())
	if err != nil {
		logger.Error(err, "error checking vnet ", "vnet", clusterReq.GetName())
		return
	}
	if !vnetExists {
		logger.Info("vnet not found, creating new vnet", "vnet", clusterReq.GetName())
		_, err := clusterSchd.ComputeClient.CreateVNet(ctx, clusterReq.ClusterId, clusterReq.CloudAccountId, clusterReq.GetName(), clusterReq.GetSpec())
		if err != nil {
			logger.Error(err, "error creating vnet")
			return
		}
		logger.Info("vnet created successfully", "vnet", clusterReq.GetName())
	} else {
		logger.Info("vnet found, reusing existing one", "vnet", clusterReq.GetName())
	}

	// Create shhkeypair for ansible executor
	priKeyfile := fmt.Sprintf("/training/sshkey-%s", clusterReq.ClusterId)
	pubKeyFile := fmt.Sprintf("/training/sshkey-%s.pub", clusterReq.ClusterId)

	if err = common.GenerateSSHKeyPair(ctx, priKeyfile, pubKeyFile); err != nil {
		logger.Error(err, "sshkey generation failed")
	}

	pubKey, err := os.ReadFile(pubKeyFile)
	if err != nil {
		logger.Error(err, "cannot read public sshkey")
	}

	sshkey, err := clusterSchd.ComputeClient.CreateSSHKey(ctx, clusterReq.ClusterId, clusterReq.CloudAccountId, pubKey)
	if err != nil {
		logger.Error(err, "error createSSHKey")
	}

	// Delete the ssh key-pair for the ansible executor after successfully completing cluster setup or on faults
	defer clusterSchd.ComputeClient.DeleteSSHKey(ctx, clusterReq.CloudAccountId, priKeyfile, pubKeyFile, sshkey)

	sshKeys := []string{sshkey.Metadata.Name}
	if clusterReq.GetSSHKeyName() != nil {
		sshKeys = append(sshKeys, clusterReq.GetSSHKeyName()...)
	}

	if err := query.UpdateClusterState(ctx, clusterSchd.db, clusterReq.ClusterId, clusterReq.CloudAccountId, "PROVISIONING"); err != nil {
		logger.Error(err, "error updating cluster state to 'PROVISIONING', skipping this cluster request")
		return
	}

	// Create storage
	localStorageMountDirs := map[string]string{}
	remoteStorageMountDirs := map[string]string{}
	storageReq := []idcStorageSvc.FilesystemCreateRequest{}

	if clusterReq.GetStorageNodes() != nil {
		for _, storageNode := range clusterReq.GetStorageNodes() {
			storageReq = append(storageReq, getSlurmStorageNodeSpecs(storageNode, clusterReq.ClusterId, clusterReq.CloudAccountId, clusterSchd.StorageClient.AvailabilityZone))

			localStorageMountDirs[storageNode.GetFsResourceId()] = storageNode.GetLocalMountDir()
			remoteStorageMountDirs[storageNode.GetFsResourceId()] = storageNode.GetRemoteMountDir()
		}
	}

	logger.Info("creating storage instances", "requested", storageReq)
	if err := clusterSchd.CreateClusterStorageFilesystems(ctx, storageReq); err != nil {
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}

	storageFilesystems := []*pb.FilesystemPrivate{}
	for idx, fs := range storageReq {
		readyStorageFS, err := clusterSchd.StorageClient.GetStorageStatusByName(ctx, fs)
		if err != nil {
			logger.Info("error getting storage instance state", "error", err)
			logger.Info("error getting storage state", "error on storage index", idx)
			clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
			return
		}

		storageFilesystems = append(storageFilesystems, readyStorageFS)
	}

	commonCloudCfg := cloudGen.CommonCloudConfigs{
		ClusterId: clusterReq.ClusterId,
	}

	if err := commonCloudCfg.RenderCommonCloudConfigs(ctx, storageFilesystems, localStorageMountDirs, remoteStorageMountDirs); err != nil {
		logger.Error(err, "error rendering weka storage cloud config")
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}

	slurmctldNodeConfig := cloudGen.SlurmctldCloudConfig{
		Common:        &commonCloudCfg,
		NodeName:      fmt.Sprintf("%s-slurmd-[1-%d]", clusterReq.ClusterId, clusterReqCounts[pb.NodeRole_COMPUTE_NODE]),
		PartitionName: clusterReq.ClusterId,
	}
	serializedSlurmctldUserData, err := cloudGen.RenderSlurmctldCloudConfig(ctx, &slurmctldNodeConfig, []*pb.Instance{})
	if err != nil {
		logger.Error(err, "could not generate slurmctld cloud config template")
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}

	instanceReq := []idcComputeSvc.InstanceCreateRequest{}
	instIdx := 1

	for _, node := range clusterReq.GetNodes() {
		if node.Role == pb.NodeRole_CONTROLLER_NODE {
			instanceReq = append(instanceReq, getSlurmControllerNodeSpecs(node, clusterReq.GetName(), clusterReq.ClusterId, clusterReq.CloudAccountId, instIdx, sshKeys, serializedSlurmctldUserData))
			instIdx++
		}
	}

	logger.Info("instance request", "creating slurmctld instance", instanceReq)
	if err := clusterSchd.CreateClusterInstances(ctx, instanceReq); err != nil {
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}
	instanceCreated = append(instanceCreated, instanceReq...)

	// It takes a while for slurmctld to setup because of the ldap server and mounting weka can take a while too
	logger.Info("Wait 2-minutes for slurmctld node to get configured...")
	time.Sleep(120 * time.Second)

	jupyterHubNodeConfig := cloudGen.JupyterHubCloudConfig{Common: &commonCloudCfg}
	serializedJupyterHubUserData, err := cloudGen.RenderCloudConfig(ctx, cloudGen.JUPYTERHUB_CLOUDINIT_TMPL_FILENAME, pongo2.Context{"Values": jupyterHubNodeConfig})
	if err != nil {
		logger.Error(err, "error reading and creating jupyterhub node cloud config")
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}

	slurmdNodeConfig := cloudGen.SlurmdCloudConfig{Common: &commonCloudCfg}
	serializedSlurmdUserData, err := cloudGen.RenderCloudConfig(ctx, cloudGen.SLURMD_CLOUDINIT_TMPL_FILENAME, pongo2.Context{"Values": slurmdNodeConfig})
	if err != nil {
		logger.Error(err, "could not generate slurmd cloud config template")
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}

	instanceReq = []idcComputeSvc.InstanceCreateRequest{}
	for _, node := range clusterReq.GetNodes() {
		if node.Role == pb.NodeRole_JUPYTERHUB_NODE {
			instanceReq = append(instanceReq, getSlurmJupyterhubNodeSpecs(node, clusterReq.GetName(), clusterReq.ClusterId, clusterReq.CloudAccountId, instIdx, sshKeys, serializedJupyterHubUserData))
			instIdx++
		}
	}

	logger.Info("instance request", "creating admin instance", instanceReq)
	if err := clusterSchd.CreateClusterInstances(ctx, instanceReq); err != nil {
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}
	instanceCreated = append(instanceCreated, instanceReq...)

	jupyterHubReadyInstance, err := clusterSchd.ComputeClient.GetInstanceStateByName(ctx, instanceReq[0])
	if err != nil {
		logger.Info("error getting jupyterhub instance state", "error", err)
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}

	if err := ansible.SetupJupyterHub(ctx, clusterReq.ClusterId, priKeyfile, jupyterHubReadyInstance); err != nil {
		logger.Info("error running jupyterhub setup, cluster setup failed")
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}

	instanceReq = []idcComputeSvc.InstanceCreateRequest{}
	computeInstanceIdx := 1

	for _, node := range clusterReq.GetNodes() {
		if node.Role == pb.NodeRole_COMPUTE_NODE {
			instanceReq = append(instanceReq, getSlurmComputeNodeSpecs(node, clusterReq.GetName(), clusterReq.ClusterId, clusterReq.CloudAccountId, computeInstanceIdx, sshKeys, serializedSlurmdUserData))
			computeInstanceIdx++
		}
	}

	logger.Info("instance request", "creating slurmd instance", instanceReq)
	if err := clusterSchd.CreateClusterInstances(ctx, instanceReq); err != nil {
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}
	instanceCreated = append(instanceCreated, instanceReq...)

	slurmdReadyInstances := []*pb.Instance{}
	for _, instance := range instanceReq {
		readyInstance, err := clusterSchd.ComputeClient.GetInstanceStateByName(ctx, instance)
		if err != nil {
			logger.Info("error getting slurmd instance state", "error", err)
			clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
			return
		}
		slurmdReadyInstances = append(slurmdReadyInstances, readyInstance)
	}

	loginNodeConfig := cloudGen.LoginCloudConfig{Common: &commonCloudCfg}
	serializedLoginUserData, err := cloudGen.RenderCloudConfig(ctx, cloudGen.LOGIN_CLOUDINIT_TMPL_FILENAME, pongo2.Context{"Values": loginNodeConfig})
	if err != nil {
		logger.Error(err, "could not generate login cloud config template")
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}

	instanceReq = []idcComputeSvc.InstanceCreateRequest{}

	// Create the rest of the cluster instances
	for _, node := range clusterReq.GetNodes() {
		if node.Role == pb.NodeRole_LOGIN_NODE {
			instanceReq = append(instanceReq, getSlurmLoginNodeSpecs(node, clusterReq.GetName(), clusterReq.ClusterId, clusterReq.CloudAccountId, instIdx, sshKeys, serializedLoginUserData))
			instIdx++
		}
	}

	logger.Info("instance request", "creating slurmctld and login instances", instanceReq)
	if err := clusterSchd.CreateClusterInstances(ctx, instanceReq); err != nil {
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}
	instanceCreated = append(instanceCreated, instanceReq...)

	if err := query.UpdateClusterState(ctx, clusterSchd.db, clusterReq.ClusterId, clusterReq.CloudAccountId, "READY"); err != nil {
		logger.Error(err, "error updating cluster state status to 'READY'")
		clusterSchd.processProvisionerTeardown(ctx, clusterReq, instanceCreated, storageReq)
		return
	}

	logger.Info("cluster setup completed successfully ")
	defer logger.Info("returning from new cluster provisioning request scan")
}

func (clusterSchd *ClusterProvisionScheduler) processProvisionerTeardown(ctx context.Context, clusterReq *pb.Cluster, instances []idcComputeSvc.InstanceCreateRequest, storages []idcStorageSvc.FilesystemCreateRequest) {
	logger := log.FromContext(ctx).WithName("ClusterProvisionScheduler.processProvisionerTeardown")

	logger.Info("PROVISIONER TEARDOWN START", "skipping the rest of the cluster setup", "REASON: faults/error during cluster provisioning process")
	defer logger.Info("PROVISIONER TEARDOWN END")

	logger.Info("Updating cluster state to 'FAILED'...")
	if err := query.UpdateClusterState(ctx, clusterSchd.db, clusterReq.ClusterId, clusterReq.CloudAccountId, "FAILED"); err != nil {
		logger.Error(err, "error updating cluster state to 'FAILED' during reconcile cluster teardown")
		logger.Info("manually update cluster state to 'FAILED' if needed", "cluster-id", clusterReq.ClusterId)
	}

	logger.Info("Deleting storage filesystems from cluster...")
	for _, storage := range storages {
		if err := clusterSchd.StorageClient.DeleteIDCStorageFilesystem(ctx, storage); err != nil {
			logger.Info("error occurred, manually delete storage if needed", "storage name", storage.Name)
		}

		if err := query.DeleteStorageNodePrivate(ctx, clusterSchd.db, storage.Name, clusterReq.CloudAccountId); err != nil {
			logger.Info("error occurred, manually delete storage nodes from the database if needed", "storage name", storage.Name)
		}
	}

	// FUTURE IMPROVEMENT: Remember state and retry from last successful point.
	logger.Info("Deleting compute instances from cluster...")
	for _, inst := range instances {
		if err := clusterSchd.ComputeClient.DeleteIDCComputeInstance(ctx, inst); err != nil {
			logger.Info("error occurred, manually delete instance if needed", "instance name", inst.Name)
		}
	}

	if err := query.DeleteAllClusterNodeInstancesPrivate(ctx, clusterSchd.db, clusterReq.ClusterId, clusterReq.CloudAccountId); err != nil {
		logger.Info("error occurred, manually delete nodes from the database if needed")
	}

	return
}
