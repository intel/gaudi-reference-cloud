// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	db_query "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/db_query_constants"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	clusterv1alpha "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type CreateClusterCrdResponse struct {
	ClusterId        int32
	CpProviderName   string
	CpK8sVersionName string
	Runtimename      string
}

const (
	superComputeGPNodegroupType = "supercompute-gp"
)

// func CreateSuperCompute Cluster and Nodegroup CRD
func CreateSuperComputeClusterAndNodegroupCRD(ctx context.Context, dbconn *sql.DB, tx *sql.Tx, record *pb.SuperComputeClusterCreateRequest, key []byte, keypub []byte, sshkey string, filepath string, computeClient pb.InstanceTypeServiceClient, vnetClient pb.VNetServiceClient) (*pb.ClusterCreateResponseForm, clusterv1alpha.Cluster, int32, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("IKS.SuperComputeCreateCluster").WithValues("name", record.Clusterspec.Name).Start()
	defer span.End()

	logger.Info("CreateSuperComputeClusterAndNodegroupCRD", "req", record)

	clusterCrdError := clusterv1alpha.Cluster{}
	returnError := &pb.ClusterCreateResponseForm{}
	returnValue := &pb.ClusterCreateResponseForm{
		Uuid:           "",
		Name:           record.Clusterspec.Name,
		Clusterstate:   "Pending",
		K8Sversionname: record.Clusterspec.K8Sversionname,
	}

	// Insert the Cluster Record Table and get the Cluster CRD with Cluster Information
	clusterResp, clusterCrd, createClusterCrdResp, err := CreateSuperComputeClusterCRD(ctx, dbconn, tx, record.Clusterspec, key, keypub, sshkey, filepath, record.CloudAccountId, record.Clustertype)
	logger.Info("CreateSuperComputeClusterAndNodegroupCRD", "Cluster Response", clusterResp)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, clusterCrdError, 0, utils.ErrorHandler(ctx, err, "CreateSuperComputeClusterCRD"+"TransactionRollbackError", "CreateSuperComputeClusterCRD Error")
		}
		return returnError, clusterCrdError, 0, err
	}

	// Insert the Nodegroup Record in Nodegroup Table and get the updated ClusterCRD with all nodegroup information
	if clusterResp != nil && clusterResp.Uuid != "" && createClusterCrdResp.ClusterId > 0 {
		for _, nodegroupReq := range record.Nodegroupspec {
			// Validate instance Types from compute
			instanceexists, err := GetComputeInstanceTypes(ctx, nodegroupReq.Instancetypeid, computeClient)
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return returnError, clusterCrdError, 0, utils.ErrorHandler(ctx, errtx, "GetComputeInstanceTypes"+"TransactionRollbackError", "Get compute instance types failed")
				}
				return returnError, clusterCrdError, 0, fmt.Errorf("get compute instance types failed")
			}
			if !instanceexists {
				if errtx := tx.Rollback(); errtx != nil {
					return returnError, clusterCrdError, 0, utils.ErrorHandler(ctx, errtx, "GetComputeInstanceTypes"+"TransactionRollbackError", "Instance types does not match to compute instance types")
				}
				return returnError, clusterCrdError, 0, fmt.Errorf("instance types does not match to compute instance types")
			}

			// Validate vnets are associated to cloud account
			if len(nodegroupReq.Vnets) == 0 {
				if errtx := tx.Rollback(); errtx != nil {
					return returnError, clusterCrdError, 0, utils.ErrorHandler(ctx, errtx, "Vnets Info"+"TransactionRollbackError", "Need at least one VNet to create a nodegroup")
				}
				return returnError, clusterCrdError, 0, fmt.Errorf("need at least one VNet to create a nodegroup")
			}
			for _, vnet := range nodegroupReq.Vnets {
				err := CheckCloudAccountVnets(ctx, record.CloudAccountId, vnet.Networkinterfacevnetname, vnet.Availabilityzonename, vnetClient)
				if err != nil {
					if errtx := tx.Rollback(); errtx != nil {
						return returnError, clusterCrdError, 0, utils.ErrorHandler(ctx, errtx, "GetComputeInstanceTypes"+"TransactionRollbackError", "GetComputeInstanceTypes Error")
					}
					return returnError, clusterCrdError, 0, err
				}
			}

			_, clusterCrd, err = CreateSuperComputeNodeGroupRecordCRD(ctx, dbconn, tx, nodegroupReq, clusterCrd, clusterResp.Uuid, record.CloudAccountId, record.Clustertype, createClusterCrdResp)
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return returnError, clusterCrdError, 0, utils.ErrorHandler(ctx, err, "CreateSuperComputeNodeGroupRecordCRD"+"TransactionRollbackError", "CreateSuperComputeNodeGroupRecordCRD Error")
				}
				return returnError, clusterCrdError, 0, err
			}
		}
	}

	if clusterResp != nil && clusterResp.Uuid != "" && createClusterCrdResp.ClusterId > 0 && record.Storagespec.Enablestorage && record.Storagespec.Storagesize != "" {
		logger.Info("CreateSuperComputeClusterAndNodegroupCRD", "Enabling Storage for super compute of size", record.Storagespec.Storagesize, "For Cluster ID:", clusterResp.Uuid)
		_, clusterCrd, err = EnableSuperComputeStorageCRD(ctx, dbconn, tx, record.Storagespec, clusterCrd, clusterResp.Uuid, record.CloudAccountId, createClusterCrdResp.ClusterId)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, clusterCrdError, 0, utils.ErrorHandler(ctx, err, "CreateSuperComputeNodeGroupRecordCRD"+"TransactionRollbackError", "CreateSuperComputeNodeGroupRecordCRD Error")
			}
			return returnError, clusterCrdError, 0, err
		}
		returnValue.Uuid = clusterResp.Uuid
		returnValue.Clusterstate = clusterResp.Clusterstate
	}

	return returnValue, clusterCrd, createClusterCrdResp.ClusterId, err
}

// CreateSuperComputeClusterRecord to inset record into cluster table and return cluster rev
func CreateSuperComputeClusterCRD(ctx context.Context, dbconn *sql.DB, tx *sql.Tx, record *pb.ClusterSpec, key []byte, keypub []byte, sshkey string, filepath string, cloudAccountId string, clusterType string) (*pb.ClusterCreateResponseForm, clusterv1alpha.Cluster, CreateClusterCrdResponse, error) {
	var clusterCrdError clusterv1alpha.Cluster
	var clusterId int32
	var createClusterCrdResponseError CreateClusterCrdResponse
	friendlyMessage := "Could not create Super Compute Cluster. Please try again."
	failedFunction := "CreateSuperComputeClusterRecord."
	returnError := &pb.ClusterCreateResponseForm{}
	returnValue := &pb.ClusterCreateResponseForm{
		Uuid:           "",
		Name:           record.Name,
		Clusterstate:   "Pending",
		K8Sversionname: record.K8Sversionname,
	}

	/*Default Values for IKS */
	defaultvalues, err := utils.GetDefaultValues(ctx, dbconn)
	if err != nil {
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultvalues", friendlyMessage)
	}

	ilbenv, ilbusergroup, _, _, minActiveMembers, memberConnectionLimit, memberPriorityGroup, memberRatio, etcdport, etcdpool_port, apiport, apiserverpool_port, public_apiserverport, public_apiserverpool_port, konnectPort, konnectPoolPort, _, _, _, _, _, maxClusterCount, err := utils.ConvDefaultsToInt(ctx, defaultvalues)
	if err != nil {
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"convDefaultsToInt", friendlyMessage)
	}

	// Validate the Super Compute Create Cluster request
	cpProviderName, cpInstanceType, cpNodeProvier, cpOsImageName, err := utils.ValidateCreateCluster(ctx, dbconn, record.Name, cloudAccountId, maxClusterCount, failedFunction, friendlyMessage)
	if err != nil {
		return returnError, clusterCrdError, createClusterCrdResponseError, err
	}

	// Validate K8s Compatability and Return IMI Components
	var (
		cpOsImageInstanceName string // IMI for the control plane
		cpK8sVersionName      string // Control Plane Full k8sversion name (patch name)
		query                 string // Query used to determin if rke2 or iks
	)
	if cpProviderName != "" && cpProviderName == "rke2" {
		query = db_query.GetDefaultCpOsImageInstance + " AND ks.k8sversion_name = $5"
	} else {
		query = db_query.GetDefaultCpOsImageInstance + " AND ks.minor_version = $5"
	}
	err = dbconn.QueryRowContext(ctx, query,
		record.Runtimename,
		cpProviderName,
		cpInstanceType,
		cpOsImageName,
		record.K8Sversionname,
	).Scan(
		&cpOsImageInstanceName,
		&cpK8sVersionName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, clusterCrdError, createClusterCrdResponseError, status.Error(codes.InvalidArgument, "IMI version not compatible")
		}
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultCpOsImageInstance", friendlyMessage)
	}

	/* GET IMI ARTIFACT FOR CLUSTER REV DESIRED JSON */
	var imiartifact string
	err = dbconn.QueryRowContext(ctx, db_query.GetImiArtifactQuery, cpOsImageInstanceName).Scan(&imiartifact)
	if err != nil {
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"GetImiArtifactQuery.cpOsImageInstanceName", friendlyMessage)
	}

	/* GET default ADD Ons */
	// defaultaddonvalues, err := utils.GetDefaultAddons(ctx, dbconn, cpK8sVersionName)
	defaultaddonvalues, err := utils.GetAddons(ctx, dbconn, true, true, cpK8sVersionName, "kubernetes")
	if err != nil {
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultAddons", friendlyMessage)
	}

	// Start the transaction

	/* INSERT CLUSTER TABLE */
	// Insert cluster
	id, err := utils.GenerateUuid(ctx, dbconn, utils.ClusterUUIDType)
	if err != nil {
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"utils.GenerateUuid", friendlyMessage)
	}
	uniqueId := "cl-" + id

	var clusterCrd clusterv1alpha.Cluster
	clusterCrd.Name = uniqueId
	clusterCrd.Annotations = make(map[string]string, 0)
	clusterCrd.Spec = clusterv1alpha.ClusterSpec{
		KubernetesVersion:    cpK8sVersionName,
		InstanceType:         cpInstanceType,
		ClusterType:          clusterType,
		InstanceIMI:          imiartifact,
		ContainerRuntime:     record.Runtimename,
		SSHKey:               []string{sshkey},    // remove this for cp. it will be populated differently for each cp.
		ContainerRuntimeArgs: map[string]string{}, // ??? NEED DEFAULT OR USER INPUT
		KubernetesProvider:   cpProviderName,
		NodeProvider:         cpNodeProvier,
		Network: clusterv1alpha.Network{
			ServiceCIDR: defaultvalues["networkservice_cidr"],
			PodCIDR:     defaultvalues["networkpod_cidr"],
			ClusterDNS:  defaultvalues["cluster_cidr"],
			Region:      defaultvalues["region"],
		},
		Addons:            make([]clusterv1alpha.AddonTemplateSpec, 0), // ??? NEED TO IMPLEMENT DEFAULT LOGIC
		Nodegroups:        make([]clusterv1alpha.NodegroupTemplateSpec, 0),
		EtcdBackupEnabled: false,
		EtcdBackupConfig:  clusterv1alpha.EtcdBackupConfig{},
		AdvancedConfig:    clusterv1alpha.AdvancedConfig{}, // ??? NEED DEFAULT OR USER INPUT
		CloudAccountId:    defaultvalues["cp_cloudaccountid"],
		VNETS: []clusterv1alpha.VNET{
			{
				AvailabilityZone:     defaultvalues["availabilityzone"],
				NetworkInterfaceVnet: defaultvalues["vnet"],
			},
		},
		ILBS: make([]clusterv1alpha.ILBTemplateSpec, 0),
	}

	// Annotations
	if record.Annotations == nil {
		record.Annotations = make([]*pb.Annotations, 0)
	}
	var annotations []*pb.Annotations
	annotations = record.Annotations
	annotationsMap := make(map[string]string)
	for _, n := range annotations {
		annotationsMap[n.Key] = n.Value
	}
	clusterCrd.Annotations = annotationsMap

	// vnets
	vnets, err := json.Marshal(clusterCrd.Spec.VNETS)
	if err != nil {
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"marshal.clustercrd.spec.vnets", friendlyMessage)
	}

	// Addons
	for _, addon := range defaultaddonvalues {
		addonCrd := clusterv1alpha.AddonTemplateSpec{
			Name:     addon.Name,
			Type:     clusterv1alpha.AddonType(addon.Type),
			Artifact: addon.Artifact,
		}
		clusterCrd.Spec.Addons = append(clusterCrd.Spec.Addons, addonCrd)
	}

	vipCrdEtcd := clusterv1alpha.ILBTemplateSpec{
		Name:        defaultvalues["ilb_etcdservername"],
		Description: "Etcd members loadbalancer",
		Port:        etcdport,
		IPType:      defaultvalues["ilb_etcdiptype"],
		IPProtocol:  defaultvalues["ilb_ipprotocol"],
		Environment: ilbenv,
		Usergroup:   ilbusergroup,
		Persist:     defaultvalues["ilb_persist"],
		Owner:       defaultvalues["ilb_cp_owner"],
		Pool: clusterv1alpha.ILBPoolTemplateSpec{
			Name:                  defaultvalues["ilb_etcdservername"],
			Port:                  etcdpool_port, //change  2379
			LoadBalancingMode:     defaultvalues["ilb_loadbalancingmode"],
			MinActiveMembers:      minActiveMembers,
			Monitor:               defaultvalues["ilb_monitor"],
			MemberConnectionLimit: memberConnectionLimit,
			MemberPriorityGroup:   memberPriorityGroup,
			MemberRatio:           memberRatio,
			MemberAdminState:      defaultvalues["ilb_memberadminstate"],
		},
	}

	vipCrdApiServer := clusterv1alpha.ILBTemplateSpec{
		Name:        defaultvalues["ilb_apiservername"],
		Description: "Kube-apiserver members loadbalancer",
		Port:        apiport,
		IPType:      defaultvalues["ilb_apiserveriptype"],
		IPProtocol:  defaultvalues["ilb_ipprotocol"],
		Environment: ilbenv,
		Usergroup:   ilbusergroup,
		Persist:     defaultvalues["ilb_persist"],
		Owner:       defaultvalues["ilb_cp_owner"],
		Pool: clusterv1alpha.ILBPoolTemplateSpec{
			Name:                  defaultvalues["ilb_apiservername"],
			Port:                  apiserverpool_port,
			LoadBalancingMode:     defaultvalues["ilb_loadbalancingmode"],
			MinActiveMembers:      minActiveMembers,
			Monitor:               defaultvalues["ilb_monitor"],
			MemberConnectionLimit: memberConnectionLimit,
			MemberPriorityGroup:   memberPriorityGroup,
			MemberRatio:           memberRatio,
			MemberAdminState:      defaultvalues["ilb_memberadminstate"],
		},
	}
	// add one more VIP with type "public"
	vipCrdPublicApiServer := clusterv1alpha.ILBTemplateSpec{
		Name:        defaultvalues["ilb_public_apiservername"],
		Description: "Public Kube-apiserver members loadbalancer",
		Port:        public_apiserverport,
		IPType:      defaultvalues["ilb_public_apiserveriptype"],
		IPProtocol:  defaultvalues["ilb_ipprotocol"],
		Environment: ilbenv,
		Usergroup:   ilbusergroup,
		Persist:     defaultvalues["ilb_persist"],
		Owner:       defaultvalues["ilb_cp_owner"],
		Pool: clusterv1alpha.ILBPoolTemplateSpec{
			Name:                  defaultvalues["ilb_public_apiservername"],
			Port:                  public_apiserverpool_port,
			LoadBalancingMode:     defaultvalues["ilb_loadbalancingmode"],
			MinActiveMembers:      minActiveMembers,
			Monitor:               defaultvalues["ilb_monitor"],
			MemberConnectionLimit: memberConnectionLimit,
			MemberPriorityGroup:   memberPriorityGroup,
			MemberRatio:           memberRatio,
			MemberAdminState:      defaultvalues["ilb_memberadminstate"],
		},
	}

	vipCrdKonnectivity := clusterv1alpha.ILBTemplateSpec{
		Name:        defaultvalues["ilb_konnectivityname"],
		Description: "Konnectivity members loadbalancer",
		Port:        konnectPort, //change 443
		IPType:      defaultvalues["ilb_konnectivityiptype"],
		IPProtocol:  defaultvalues["ilb_ipprotocol"],
		Environment: ilbenv,
		Usergroup:   ilbusergroup,
		Persist:     defaultvalues["ilb_persist"],
		Owner:       defaultvalues["ilb_cp_owner"],
		Pool: clusterv1alpha.ILBPoolTemplateSpec{
			Name:                  defaultvalues["ilb_konnectivityname"],
			Port:                  konnectPoolPort, //change 8132
			LoadBalancingMode:     defaultvalues["ilb_loadbalancingmode"],
			MinActiveMembers:      minActiveMembers,
			Monitor:               defaultvalues["ilb_monitor"],
			MemberConnectionLimit: memberConnectionLimit,
			MemberPriorityGroup:   memberPriorityGroup,
			MemberRatio:           memberRatio,
			MemberAdminState:      defaultvalues["ilb_memberadminstate"],
		},
	}
	clusterCrd.Spec.ILBS = append(clusterCrd.Spec.ILBS, vipCrdEtcd, vipCrdApiServer, vipCrdPublicApiServer, vipCrdKonnectivity)

	description := ""
	if record.Description == nil {
		record.Description = &description
	}
	recordTags, err := json.Marshal(record.Tags)
	recordAnnotations, err := json.Marshal(record.Annotations)
	err = tx.QueryRowContext(ctx, db_query.InsertClusterRecordQuery,
		"Pending",
		record.Name,
		record.Description,
		clusterCrd.Spec.Network.Region,
		clusterCrd.Spec.Network.ServiceCIDR,
		clusterCrd.Spec.Network.ClusterDNS,
		clusterCrd.Spec.Network.PodCIDR,
		nil, //TEMP ENCRYPTION CONFIG
		nil, //TEMP ADVANCE CONFIG
		nil, //TEMP PROVIDER ARGS
		nil, //TEMP BACKUP ARGS
		cpProviderName,
		nil, // TEMP Labels
		recordTags,
		recordAnnotations,
		"S3", //TEMP BACKUPTYPE NAME
		time.Now(),
		uniqueId,
		cloudAccountId,
		`{"status":"cluster is being created"}`, // NEED DEFAULT LOGIC
		clusterType,
	).Scan(&clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"InsertClusterRecordQuery", friendlyMessage)
	}

	/* CREATE CONTROL PLANE NODEGROUP TABLE */
	var cpNodegroupId int32
	cpUniqueId := "cp-" + id
	var kubernetesstatus clusterv1alpha.NodegroupStatus
	kubernetesstatus.Name = cpUniqueId
	kubernetesstatus.State = "Updating"
	nodegroupstatus, err := json.Marshal(kubernetesstatus)
	if err != nil {
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"marshal.kubernetesstatus", friendlyMessage)
	}

	nonce, encryptionid, sskpkey, sshpubkey, err := controlplanesshencryption(ctx, key, keypub, filepath)
	if err != nil {
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"ControlPlaneSshEnc", friendlyMessage)
	}
	// Insert ssh key details into cluster_extraconfig table
	_, err = tx.QueryContext(ctx, db_query.InsertSshkeyQuery, clusterId, sskpkey, sshpubkey, sshkey, nonce, encryptionid)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"TransactionRollbackSshKeyError", friendlyMessage)
		}
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"InsertSshkeyQuery", friendlyMessage)
	}

	sshkeysarray := []*pb.SshKey{}
	sshkeyarray := &pb.SshKey{
		Sshkey: sshkey,
	}

	sshkeysarray = append(sshkeysarray, sshkeyarray)

	sshkeyjson, err := json.Marshal(sshkeysarray)
	if err != nil {
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"Marshal Ssh key", friendlyMessage)
	}
	err = tx.QueryRowContext(ctx, db_query.InsertControlPlaneQuery,
		clusterId,
		cpOsImageInstanceName,
		cpK8sVersionName,
		"Updating",
		"ControlPlane",
		cpUniqueId,
		"ControlPlane for: "+string(uniqueId),
		defaultvalues["networkinterfacename"], // TEMP
		cpInstanceType,
		3,
		sshkeyjson,
		record.Runtimename,
		nil,
		true,
		10,
		"Statedetails",
		time.Now(),
		cpUniqueId,
		1,
		nodegroupstatus,
		vnets,
		superComputeGPNodegroupType,
	).Scan(&cpNodegroupId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"InsertControlPlaneQuery", friendlyMessage)
	}

	// Load vip
	var vipid int32
	dnsalias := []string{}

	// provider is default
	var vipprovider string
	err = tx.QueryRowContext(ctx, db_query.GetVipProviderDefaults).Scan(&vipprovider)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"GetVipProviderDefaults", friendlyMessage)
	}

	for _, cpilb := range clusterCrd.Spec.ILBS {
		dnsAliasJson, err := json.Marshal(dnsalias)
		err = tx.QueryRowContext(ctx, db_query.InsertVipQuery,
			clusterId,
			dnsAliasJson,
			"",
			cpilb.IPType,
			`{"status":"Control Plane Vip is being created"}`,
			"Pending",
			cpilb.Owner,
			vipprovider, //default from provider table
			"",
		).Scan(&vipid)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"InsertVipQuery", friendlyMessage)
		}

		/*Insert VIP details table */
		var vipId int32
		err = tx.QueryRowContext(ctx, db_query.InsertVipDetailsQuery, vipid, cpilb.Name, cpilb.Description, cpilb.Port, cpilb.Pool.Name, cpilb.Port).Scan(&vipId)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, clusterCrdError, createClusterCrdResponseError, utils.ErrorHandler(ctx, err, failedFunction+"InsertVipDetailsQuery", friendlyMessage)
		}
	}

	returnValue.Uuid = uniqueId
	returnValue.Clusterstate = "Pending"

	createClusterCrdResponse := CreateClusterCrdResponse{
		ClusterId:        clusterId,
		CpProviderName:   cpProviderName,
		CpK8sVersionName: cpK8sVersionName,
		Runtimename:      record.Runtimename,
	}
	return returnValue, clusterCrd, createClusterCrdResponse, err
}

// CreateNodeGroupRecord to insert record into nodegroup table
func CreateSuperComputeNodeGroupRecordCRD(ctx context.Context, dbconn *sql.DB, tx *sql.Tx, record *pb.NodegroupSpec, clusterCrd clusterv1alpha.Cluster, clusteruuid string, cloudAccountId string, clusterType string, createClusterCrdResp CreateClusterCrdResponse) (*pb.NodeGroupResponseForm, clusterv1alpha.Cluster, error) {
	clusterCrdError := clusterv1alpha.Cluster{}
	friendlyMessage := "Could not create Node Group. Please try again."
	failedFunction := "CreateNodeGroupRecord."
	returnError := &pb.NodeGroupResponseForm{}

	/*Default Values for IKS */
	defaultvalues, err := utils.GetDefaultValues(ctx, dbconn)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultValues", friendlyMessage)
	}

	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, max_nodegroups, _, _, _, _, err := utils.ConvDefaultsToInt(ctx, defaultvalues)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"convDefaultsToInt", friendlyMessage)
	}

	// Validate Create Nodegroup Request SPEC for Cluster
	_, err = utils.ValidateCreateNodegroup(ctx, dbconn, record.Instancetypeid, record.Count, cloudAccountId, clusteruuid, record.Name, clusterType, max_nodegroups, failedFunction, friendlyMessage)
	if err != nil {
		return returnError, clusterCrdError, err
	}

	cloudAccountmaxNgPerCluster := -1
	_, cloudAccountmaxNgPerCluster, _, _, _, err = utils.GetCloudAccountMaxValues(ctx, dbconn, cloudAccountId)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"GetCloudAccountMaxValues", friendlyMessage)
	}
	if cloudAccountmaxNgPerCluster > -1 {
		max_nodegroups = cloudAccountmaxNgPerCluster
	}

	description := ""
	if record.Description == nil {
		record.Description = &description
	}

	if record.Userdataurl == nil {
		record.Userdataurl = new(string)
	}

	returnValue := &pb.NodeGroupResponseForm{
		Nodegroupuuid:  "",
		Clusteruuid:    clusteruuid,
		Name:           record.Name,
		Description:    record.Description,
		Instancetypeid: record.Instancetypeid,
		Nodegroupstate: "Pending",
		Createddate:    time.Now().String(),
		Upgradestrategy: &pb.UpgradeStrategy{
			Drainnodes:               true,
			Maxunavailablepercentage: 10,
		},
		Tags:                 record.Tags,
		Nodegroupstatus:      &pb.Nodegroupstatus{},
		Vnets:                []*pb.Vnet{},
		Networkinterfacename: defaultvalues["networkinterfacename"],
		Userdataurl:          record.Userdataurl,
	}
	if record.Upgradestrategy != nil {
		returnValue.Upgradestrategy.Drainnodes = record.Upgradestrategy.Drainnodes
		returnValue.Upgradestrategy.Maxunavailablepercentage = record.Upgradestrategy.Maxunavailablepercentage
	}
	if record.Sshkeyname == nil {
		record.Sshkeyname = make([]*pb.SshKey, 0)
	}
	if record.Annotations == nil {
		record.Annotations = make([]*pb.Annotations, 0)
	}

	/* VALIDATE COMPATABILITY TABLES*/
	// Get information from the Control Plane
	// clusterProvider, clusterK8sVersion, clusterRuntime, _, _, err := utils.GetClusterCompatDetails(ctx, dbconn, clusteruuid)
	// if err != nil {
	// 	return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"GetClustersCompatDetails", friendlyMessage)
	// }

	// **TEMP** Obtain the OS Image name for WRKs (22.04). Will change once customers can select OS Images
	GetDefaultWrkOsImageQuery := `
		SELECT osimage_name
		FROM osimage
		WHERE wrk_default='true' AND lifecyclestate_id=(SELECT lifecyclestate_id FROM lifecyclestate WHERE name='Active')
	`
	var (
		wrkOsImageName string // *TEMP* Default OS image
	)
	err = dbconn.QueryRowContext(ctx, GetDefaultWrkOsImageQuery).Scan(&wrkOsImageName)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageName", friendlyMessage)
	}

	// Get Default worker IMI
	query := db_query.GetDefaultWrkOsImageInstance + " AND ks.k8sversion_name = $5"
	var wrkOsImageInstanceName string
	err = dbconn.QueryRowContext(ctx, query,
		createClusterCrdResp.Runtimename,
		createClusterCrdResp.CpProviderName,
		record.Instancetypeid,
		wrkOsImageName,
		createClusterCrdResp.CpK8sVersionName,
	).Scan(&wrkOsImageInstanceName)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, clusterCrdError, status.Error(codes.InvalidArgument, "IMI version not compatible. Please contact support regarding this issue.")
		}
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"getDefaultWrkOsImageInstance.query", friendlyMessage)
	}

	// Get IMI artifact
	imiartifact, err := utils.GetImiArtifact(ctx, dbconn, wrkOsImageInstanceName)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, clusterCrdError, status.Error(codes.InvalidArgument, "IMI version not compatible. Please contact support regarding this issue.")
		}
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"GetImiArtifact", friendlyMessage)
	}

	/* CREATE ENTRY IN NODEGROUP TABLE */
	var nodegroupid int32
	id, err := utils.GenerateUuid(ctx, dbconn, utils.NodegroupUUIDType)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, clusterCrdError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"utils.GenerateUuid.NG", friendlyMessage)
	}
	ngUniqueId := "ng-" + id
	returnValue.Nodegroupuuid = ngUniqueId
	returnValue.Vnets = record.Vnets
	returnValue.Sshkeyname = record.Sshkeyname
	returnValue.Count = record.Count

	var kubernetesstatus clusterv1alpha.NodegroupStatus
	kubernetesstatus.Name = ngUniqueId
	kubernetesstatus.State = "creating"
	nodegroupstatus, err := json.Marshal(kubernetesstatus)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"marshal.kubernetesstatus", friendlyMessage)
	}

	vnets, err := json.Marshal(record.Vnets)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"marshal.vnets", friendlyMessage)
	}

	err = tx.QueryRowContext(ctx, db_query.InsertNodeGroupQuery,
		createClusterCrdResp.ClusterId,
		wrkOsImageInstanceName,
		createClusterCrdResp.CpK8sVersionName,
		"Updating",
		"Worker",
		record.Name,
		record.Description,
		defaultvalues["networkinterfacename"],
		record.Instancetypeid,
		record.Count,
		record.Sshkeyname,
		createClusterCrdResp.Runtimename,
		record.Tags,
		returnValue.Upgradestrategy.Drainnodes,
		returnValue.Upgradestrategy.Maxunavailablepercentage,
		"statedetails",
		time.Now(),
		ngUniqueId,
		1,
		record.Annotations,
		nodegroupstatus,
		vnets,
		record.Userdataurl,
		record.Nodegrouptype,
	).Scan(&nodegroupid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, clusterCrdError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"InsertNodeGroupQuery", friendlyMessage)
	}
	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, db_query.UpdateClusterStateQuery,
		createClusterCrdResp.ClusterId,
		"Pending",
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, clusterCrdError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
	}

	// Create Nodegroup CRD and append to clusterCrd that was already created
	nodeGroupCrd := clusterv1alpha.NodegroupTemplateSpec{
		Name:              ngUniqueId,
		KubernetesVersion: createClusterCrdResp.CpK8sVersionName,
		InstanceType:      record.Instancetypeid,
		ClusterType:       clusterType,
		InstanceIMI:       imiartifact,
		SSHKey:            []string{}, // NEEDS TO IMPLEMENT LOGIC
		Count:             int(record.Count),
		UpgradeStrategy: clusterv1alpha.UpgradeStrategy{
			DrainBefore:           returnValue.Upgradestrategy.Drainnodes,
			MaxUnavailablePercent: int(returnValue.Upgradestrategy.Maxunavailablepercentage),
		},
		Taints:         map[string]string{},     //NEED TO EITHER FIX IN IKS PROTO OR CONVERT
		Labels:         map[string]string{},     //NEED TO EITHER FIX IN IKS PROTO OR CONVERT
		Annotations:    map[string]string{},     //NEED TO EITHER FIX IN IKS PROTO OR CONVERT
		VNETS:          []clusterv1alpha.VNET{}, // TBD updated
		CloudAccountId: cloudAccountId,
	}

	if record.Userdataurl != nil && *record.Userdataurl != "" {
		nodeGroupCrd.UserDataURL = *record.Userdataurl
	}

	var annotations []*pb.Annotations
	annotations = record.Annotations
	annotationsMap := make(map[string]string)
	for _, n := range annotations {
		annotationsMap[n.Key] = n.Value
	}
	nodeGroupCrd.Annotations = annotationsMap

	var sshkeys []*pb.SshKey
	sshkeys = record.Sshkeyname
	for _, keys := range sshkeys {
		nodeGroupCrd.SSHKey = append(nodeGroupCrd.SSHKey, keys.Sshkey)
	}

	var vnetcrd clusterv1alpha.VNET
	var vnetscrd []clusterv1alpha.VNET
	Vnets := record.Vnets
	for _, n := range Vnets {
		vnetcrd.AvailabilityZone = n.Availabilityzonename
		vnetcrd.NetworkInterfaceVnet = n.Networkinterfacevnetname

		vnetscrd = append(vnetscrd, vnetcrd)
	}
	nodeGroupCrd.VNETS = vnetscrd

	clusterCrd.Spec.Nodegroups = append(clusterCrd.Spec.Nodegroups, nodeGroupCrd)

	// // Create new cluster rev table entry
	// var revversion string
	// err = tx.QueryRowContext(ctx, InsertRevQuery,
	// 	clusterId,
	// 	currentJson,
	// 	clusterCrd,
	// 	"test",
	// 	"test",
	// 	"test",
	// 	"test",
	// 	time.Now(),
	// 	false,
	// ).Scan(&revversion)
	// if err != nil {
	// 	if errtx := tx.Rollback(); errtx != nil {
	// 		return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
	// 	}
	// 	return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	// }

	/* GET NODEGROUP STATUS */
	// getNodeGroupStatusReq := &pb.NodeGroupid{
	// 	Clusteruuid:    clusteruuid,
	// 	Nodegroupuuid:  ngUniqueId,
	// 	CloudAccountId: cloudAccountId,
	// }
	// getNodegroupStatusResponse, err := iks_query.GetNodeGroupStatusRecord(ctx, dbconn, getNodeGroupStatusReq)
	// if err != nil {
	// 	return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"GetStatusRecord", friendlyMessage)
	// }
	// returnValue.Nodegroupstatus = getNodegroupStatusResponse

	return returnValue, clusterCrd, nil
}

func EnableSuperComputeStorageCRD(ctx context.Context, dbconn *sql.DB, tx *sql.Tx, record *pb.StorageSpec, clusterCrd clusterv1alpha.Cluster, clusteruuid string, cloudAccountId string, clusterId int32) (*pb.ClusterStorageStatus, clusterv1alpha.Cluster, error) {
	var clusterCrdError clusterv1alpha.Cluster
	friendlyMessage := "Could not enable Cluster Storage. Please try again"
	failedFunction := "GetClusterRecord."
	returnError := &pb.ClusterStorageStatus{}
	returnValue := &pb.ClusterStorageStatus{}

	/* VALIDATIONS */
	// Validate Cloudaccount storage permission
	storageAllowed, err := utils.ValidateStorageRestrictions(ctx, dbconn, cloudAccountId)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"Utils.ValidateStorageRestrictions", friendlyMessage)
	}
	if !storageAllowed {
		return returnError, clusterCrdError, status.Error(codes.PermissionDenied, "Due to storage restrictions, we are currently not allowing non-approved users to use storage.")
	}

	// Validated request is not "enable: false"
	if !record.Enablestorage {
		return returnError, clusterCrdError, status.Error(codes.FailedPrecondition, "Cluster Storage temporarily cannot be disabled.")
	}

	/* Get the Provider */
	var storageProvider string
	err = dbconn.QueryRowContext(ctx, db_query.GetDefaultStorageProvider).Scan(&storageProvider)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultStorageProvider", friendlyMessage)
	}

	/* ADD THE ADDONS */
	addons, err := utils.GetAddons(ctx, dbconn, true, false, clusterCrd.Spec.KubernetesVersion, storageProvider)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"Utils.Addons."+storageProvider, friendlyMessage)
	}

	/* ADD THE VALUES TO THE CRD */
	clusterCrd.Spec.CustomerCloudAccountId = cloudAccountId
	clusterCrd.Spec.Storage = []clusterv1alpha.Storage{}
	newStorage := clusterv1alpha.Storage{
		Provider: storageProvider,
		Size:     record.Storagesize,
	}
	clusterCrd.Spec.Storage = append(clusterCrd.Spec.Storage, newStorage)
	for _, addon := range addons {
		addonCrd := clusterv1alpha.AddonTemplateSpec{
			Name:     addon.Name,
			Type:     clusterv1alpha.AddonType(addon.Type),
			Artifact: addon.Artifact,
		}
		clusterCrd.Spec.Addons = append(clusterCrd.Spec.Addons, addonCrd)
	}

	/* UPDATE STORAGE TABLE */
	storageSize := utils.ParseFileSize(record.Storagesize)
	if storageSize == -1 {
		return returnError, clusterCrdError, status.Error(codes.InvalidArgument, "invalid Storage Size")
	}
	_, err = tx.ExecContext(ctx, db_query.InsertStorageTableQuery, clusterId, storageProvider, "Updating", storageSize)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.Exec.InsertStorageTableQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER TABLE WITH STORAGE INFO*/
	_, err = tx.ExecContext(ctx, db_query.UpdateClusterStorageDataQuery, clusterId, record.Enablestorage)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.Exec.UpdateClusterStorageDataQuery", friendlyMessage)
	}

	/* UPDATE CLOUD ACCOUNT TABLE*/
	_, err = tx.ExecContext(ctx, db_query.UpdateCloudAccountStorageQuery, cloudAccountId, storageSize)
	if err != nil {
		return returnError, clusterCrdError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.Exec.UpdateCloudAccountStorageQuery", friendlyMessage)
	}

	return returnValue, clusterCrd, nil
}

func InsertClusterRevTable(ctx context.Context, dbconn *sql.DB, tx *sql.Tx, clusterCrd clusterv1alpha.Cluster, clusteruuid string, clusterId int32) error {
	friendlyMessage := "Could not Insert Cluster CRD. Please try again."
	failedFunction := "InsertClusterRevTable."

	/* VALIDATE CLUSTER EXISTANCE */
	/* var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, clusteruuid)
	if err != nil {
		return utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return errors.New("Cluster not found: " + clusteruuid)
	} */

	// Start the transaction
	/* CREATE CLUSTER REV TABLE */
	var revversion string
	clusterCrdJson, err := json.Marshal(clusterCrd)
	err = tx.QueryRowContext(ctx, db_query.InsertRevQuery,
		clusterId,
		`{}`,
		clusterCrdJson,
		"test", // ?? DEFAULT VALUES
		"test", // ?? DEFAULT VALUES
		"test", // ?? DEFAULT VALUES
		"test", // ?? DEFAULT VALUES
		time.Now(),
		false,
	).Scan(&revversion)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	}

	/* INSERT PROVISIONING LOG */
	_, err = tx.QueryContext(ctx,
		db_query.InsertProvisioningQuery,
		clusterId,
		"cluster create pending",
		"INFO",
		"cluster create",
		time.Now(),
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return utils.ErrorHandler(ctx, err, failedFunction+"insertProvisioningQuery", friendlyMessage)
	}

	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	err = tx.Commit()
	if err != nil {
		return utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}

	return nil
}

func controlplanesshencryption(ctx context.Context, sshprivatekey []byte, sshpubkey []byte, filepath string) (string, int, string, string, error) {
	/*ENCRYPT INCOMING DATA */
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", 0, "", "", err
	}
	encodedNonce, err := utils.Base64EncodeString(ctx, nonce)
	if err != nil {
		return "", 0, "", "", err
	}

	encryptionKeyByte, encryptionKeyId, err := utils.GetLatestEncryptionKey(ctx, filepath)
	if err != nil {
		return "", 0, "", "", err
	}
	sshprivateEnc, err := utils.AesEncryptSecret(ctx, string(sshprivatekey), encryptionKeyByte, nonce)
	if err != nil {
		return "", 0, "", "", err
	}
	sshpublicEnc, err := utils.AesEncryptSecret(ctx, string(sshpubkey), encryptionKeyByte, nonce)
	if err != nil {
		return "", 0, "", "", err
	}
	return encodedNonce, encryptionKeyId, sshprivateEnc, sshpublicEnc, nil
}

func CheckCloudAccountVnets(ctx context.Context, cloudAccount string, vnetName string, availabilityZoneName string, vnetClient pb.VNetServiceClient) error {

	keyName := &pb.VNetGetRequest_Metadata_Name{
		Name: vnetName,
	}

	/* CHECK IF CURRENT VNET CONNECTION EXISTS */
	getReq := &pb.VNetGetRequest{
		Metadata: &pb.VNetGetRequest_Metadata{
			NameOrId:       keyName,
			CloudAccountId: cloudAccount,
		},
	}
	vnet, err := vnetClient.Get(ctx, getReq)
	if err != nil || vnet == nil {
		return fmt.Errorf("Unable to get VNet with name %q", vnetName) // This is a friendly message. Do we want specific?
	}
	if vnet.Spec.AvailabilityZone != availabilityZoneName {
		return fmt.Errorf("VNet %q is in a different availability zone", vnetName)
	}

	return nil
}

func GetComputeInstanceTypes(ctx context.Context, instancetype string, computeClient pb.InstanceTypeServiceClient) (bool, error) {
	_, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ComputeInstanceTypes").
		WithValues("instancetype", instancetype).Start()

	defer span.End()
	var instanceexists bool
	identifier := "cluster"

	if strings.Contains(instancetype, identifier) {
		igInstanceTypeSplit := strings.Split(instancetype, "-")
		instancetype = strings.Join(igInstanceTypeSplit[:len(igInstanceTypeSplit)-2], "-")
	}

	instances, err := computeClient.Search(ctx, &pb.InstanceTypeSearchRequest{})
	if err != nil {
		log.Error(err, "\n .. Get compute instance types failed")
		return false, err
	}

	for i, _ := range instances.Items {
		if instancetype == instances.Items[i].Spec.Name {
			instanceexists = true
			break
		}
	}

	return instanceexists, nil
}
