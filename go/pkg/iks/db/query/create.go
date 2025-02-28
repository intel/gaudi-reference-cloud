// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"strconv"
	"strings"
	"time"

	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	clusterv1alpha "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	// Validation Queries
	// APPENDS ADDITIONAL 'AND' ON CREATE CLUSTER
	GetDefaultCpOsImageInstance = `
		SELECT kc.cp_osimageinstance_name, kc.k8sversion_name
		FROM public.k8scompatibility kc
			INNER JOIN public.k8sversion ks
			ON kc.k8sversion_name = ks.k8sversion_name AND kc.provider_name = $2
		WHERE kc.runtime_name = $1 AND NOT ks.test_version AND kc.instancetype_name = $3 AND kc.osimage_name = $4
		AND ks.lifecyclestate_id = (SELECT l.lifecyclestate_id FROM lifecyclestate l WHERE l.name = 'Active')
	`
	// APPENDS ADDITIONAL 'AND' ON CREATE NODEGROUP
	GetDefaultWrkOsImageInstance = `
		SELECT kc.wrk_osimageinstance_name
		FROM public.k8scompatibility kc
			INNER JOIN public.k8sversion ks
			ON kc.k8sversion_name = ks.k8sversion_name AND kc.provider_name = $2
		WHERE kc.runtime_name = $1 AND NOT ks.test_version AND kc.instancetype_name = $3 AND kc.osimage_name = $4
	`
)

const (
	superComputeClusterType     = "supercompute"
	generalPurposeClusterType   = "generalpurpose"
	superComputeGPNodegroupType = "supercompute-gp"
	generalPurposeNodeGroupType = "gp"
)

const (
	GetDefaultCpInstanceTypeAndNodeProvider = `
		SELECT instancetype_name , nodeprovider_name
		FROM public.instancetype
		WHERE is_default = true and nodeprovider_name = (SELECT nodeprovider_name from public.nodeprovider where is_default = true)
	`
	GetDefaultCpOsImageQuery = `
		SELECT osimage_name
		FROM osimage
		WHERE cp_default='true' AND lifecyclestate_id=(SELECT lifecyclestate_id FROM lifecyclestate WHERE name='Active')
	`
	GetCloudAccountProvider = `
		SELECT coalesce(
			(	SELECT provider_name
				FROM public.cloudaccountextraspec c
				WHERE cloudaccount_id = $1
			),'iks') AS provider_name
	`
	clusterExistanceQuery = `
		SELECT cluster_id
		FROM public.cluster
		WHERE unique_id = $1 and clusterstate_name != 'Deleted'
	`
	nodeGroupExistanceQuery = `
		SELECT n.cluster_id
	 	FROM public.nodegroup n
		WHERE n.unique_id = $1 AND n.cluster_id = (
			SELECT cluster_id
			FROM public.cluster  c
			WHERE c.unique_id = $2 AND c.clusterstate_name != 'Deleted')
	`
	nodeGroupExistanceAndPermissionQuery = `
		SELECT count(*)
	 	FROM public.nodegroup n
		WHERE n.unique_id = $1 AND n.cluster_id = (
				SELECT c.cluster_id
				FROM cluster c
				WHERE c.unique_id = $2 AND c.cloudaccount_id = $3
		)
	`
	// insert queries
	InsertClusterRecordQuery = `
 		INSERT INTO public.cluster (clusterstate_name, name, description, region_name, networkservice_cidr,
			cluster_dns, networkpod_cidr, encryptionconig, advanceconfigs, provider_args, backup_args, provider_name,
			labels, tags, annotations, backuptype_name, created_date, unique_id, cloudaccount_id, kubernetes_status, clustertype)
 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING cluster_id
 	`
	InsertControlPlaneQuery = `
		INSERT INTO public.nodegroup (cluster_id, osimageinstance_name, k8sversion_name,
			nodegroupstate_name, nodegrouptype_name, name, description,
			networkinterface_name, instancetype_name, nodecount, sshkey, runtime_name,
			tags, upgstrategydrainbefdel, upgstrategymaxnodes, statedetails,
			createddate, unique_id, lifecyclestate_id, kubernetes_status, vnets, nodegrouptype)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING nodegroup_id
		`
	InsertNodeGroupQuery = `
		INSERT INTO public.nodegroup (cluster_id, osimageinstance_name, k8sversion_name,
			nodegroupstate_name, nodegrouptype_name, name, description,
			networkinterface_name, instancetype_name, nodecount, sshkey, runtime_name,
			tags, upgstrategydrainbefdel, upgstrategymaxnodes, statedetails,
			createddate, unique_id, lifecyclestate_id ,annotations, kubernetes_status, vnets, userdata_webhook, nodegrouptype)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22, $23, $24)
		RETURNING nodegroup_id
		`
	InsertRevQuery = `
 		INSERT INTO public.clusterrev (cluster_id, currentspec_json, desiredspec_json, component_typegrp,
			component_typename, currentdata, desireddata, timestamp, change_applied)
 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING clusterrev_id
 	`
	InsertProvisioningQuery = `
	INSERT INTO public.provisioninglog (cluster_id, logentry, loglevel_name,logobject,logtimestamp)
	VALUES ($1,$2,$3,$4,$5)
	`
	InsertVipQuery = `
	INSERT INTO public.vip (cluster_id, dns_aliases, vip_dns,viptype_name, vip_status, vipstate_name, owner,vipprovider_name,vipinstance_id)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING vip_id`

	InsertVipDetailsQuery = `
		INSERT INTO public.vipdetails (vip_id, vip_name, description, port, pool_name, pool_port)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING vip_id
	`

	UpdateVipQuery = `
	UPDATE public.vip set sourceips= $1 , firewall_status = $2 where vip_id = $3`

	UpdateVipDetailsQuery = `
	UPDATE public.vipdetails set protocol = $1 where vip_id = $2`

	InsertSshkeyQuery = `
	    INSERT INTO public.cluster_extraconfig (cluster_id,cluster_ssh_key,cluster_ssh_pub_key,cluster_ssh_key_name,nonce,encryptionkey_id) VALUES ($1,$2,$3,$4,$5,$6)
	`
	// Get queries
	GetClustersStatesByName = `
		SELECT unique_id, clusterstate_name
	 	FROM public.cluster
		WHERE name = $1 and cloudaccount_id = $2
	`

	GetVipByName = `
	SELECT count(*)
	FROM public.vip v
	JOIN public.vipdetails vd ON v.vip_id = vd.vip_id
	WHERE v.cluster_id = $1 and vd.vip_name = $2
	`

	GetNodeGroupCrdQuery = `
		SELECT name ,k8sversion_name, instancetype_name, osimageinstance_name, nodecount, upgstrategymaxnodes, upgstrategydrainbefdel
		FROM public.nodegroup where cluster_id = $1 and unique_id = $2
	`
	GetAddOncrdQuery = `
		SELECT addonversion_name, addonargs
		FROM public.clusteraddonversion where cluster_id = $1
	`
	getControlPlaneDefaults = `
		SELECT c.cluster_id, n.runtime_name, n.k8sversion_name
		FROM public.cluster c
			INNER JOIN nodegroup n
			ON c.cluster_id = n.cluster_id AND n.nodegrouptype_name = 'ControlPlane'
		WHERE c.unique_id = $1
	`
	GetInstanceTypeOverrideQuery = `
		SELECT imi_override
		FROM instancetype
		WHERE instancetype_name = $1 AND lifecyclestate_id = (SELECT lifecyclestate_id FROM lifecyclestate l WHERE l.name = 'Active' )
	`

	GetVipProviderDefaults = `
	    SELECT vipprovider_name FROM public.vipprovider where is_default = 'true'
	`

	getLatestClusterRevIDQuery = `
	SELECT clusterrev_id FROM  public.clusterrev WHERE cluster_id = $1
	ORDER  BY timestamp DESC NULLS LAST
    LIMIT  1;`

	GetImiArtifactQuery = `
	SELECT imiartifact FROM public.osimageinstance WHERE osimageinstance_name = $1
	`

	LockVIPTable = `
	LOCK TABLE public.vip IN EXCLUSIVE MODE`

	LockClusterTable = `
	LOCK TABLE public.cluster IN EXCLUSIVE MODE`

	LockNodegroupTable = `
	LOCK TABLE public.cluster IN EXCLUSIVE MODE`
)

// CreateClusterRecord to inset record into cluster table
func CreateClusterRecord(ctx context.Context, dbconn *sql.DB, record *pb.ClusterRequest, key []byte, keypub []byte, sshkey string, filepath string) (*pb.ClusterCreateResponseForm, error) {
	friendlyMessage := "Could not create Cluster. Please try again."
	functionName := "CreateClusterRecord."
	returnError := &pb.ClusterCreateResponseForm{}
	returnValue := &pb.ClusterCreateResponseForm{
		Uuid:           "",
		Name:           record.Name,
		Clusterstate:   "Pending",
		K8Sversionname: record.K8Sversionname,
	}

	log := log.FromContext(ctx).WithName("CreateNewVip")

	/* VALIDATE CLUSTER TYPE */
	var clusterType string
	var nodegroupType string
	if record.Clustertype != nil && *record.Clustertype != "" && *record.Clustertype == superComputeClusterType {
		return returnError, utils.ErrorHandlerWithGrpcCode(ctx, nil, functionName, "Please use supercompute create cluster API to create a supercompute cluster.", codes.InvalidArgument)
	} else {
		clusterType = generalPurposeClusterType
		nodegroupType = generalPurposeNodeGroupType
	}

	/* VALIDATE CLOUD ACCOUNT RESTRICTIONS*/
	activeAccount, err := utils.ValidateClusterCloudRestrictions(ctx, dbconn, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"Utils.ValidateClusterCloudRestrictions", friendlyMessage)
	}
	if !activeAccount {
		return returnError, status.Error(codes.PermissionDenied, "Due to restrictions, we are currently not allowing non-approved users to provision clusters.")
	}

	/* VALIDATE CLUSTER NAME UNIQUENESS*/
	rows, err := dbconn.QueryContext(ctx, GetClustersStatesByName, record.Name, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetClustersStatesbyName", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			clusterState string
			clusterUuid  string
		)
		err = rows.Scan(&clusterUuid, &clusterState)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, functionName+"GetClustersStatesbyName.rows.scan", friendlyMessage)
		}
		if clusterState != "Deleting" && clusterState != "Deleted" {
			return returnError, status.Error(codes.AlreadyExists, "Cluster name already in use")
		}
	}

	/* CHECK CLOUDACCOUNT DATA FOR CUSTOM NODE PROVIDER*/
	var cpProviderName string
	err = dbconn.QueryRowContext(ctx, GetCloudAccountProvider,
		record.CloudAccountId,
	).Scan(&cpProviderName)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetCloudAccountProvider", friendlyMessage)
	}

	/*Default Values for IKS */
	defaultvalues, err := utils.GetDefaultValues(ctx, dbconn)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetDefaultvalues", friendlyMessage)
	}

	ilbenv, ilbusergroup, _, _, minActiveMembers, memberConnectionLimit, memberPriorityGroup, memberRatio, etcdport, etcdpool_port, apiport, apiserverpool_port, public_apiserverport, public_apiserverpool_port, konnectPort, konnectPoolPort, _, _, _, _, _, maxClusterCount, err := convDefaultsToInt(ctx, defaultvalues)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"convDefaultsToInt", friendlyMessage)
	}

	/* Max Cluster Count for Cloud Account*/
	cloudAccountMaxClusters := -1
	cloudAccountMaxClusters, _, _, _, _, err = utils.GetCloudAccountMaxValues(ctx, dbconn, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetCloudAccountMaxValues", friendlyMessage)
	}
	if cloudAccountMaxClusters > -1 {
		maxClusterCount = cloudAccountMaxClusters
	}

	/* VALIDATE CONTROL PLANE COMPATABILITY TABLES*/
	// Get the default instance type of CP nodes
	var (
		cpInstanceType string // Default control plane instance type
		cpNodeProvier  string // Default control plane node provider
	)
	err = dbconn.QueryRowContext(ctx, GetDefaultCpInstanceTypeAndNodeProvider).Scan(&cpInstanceType, &cpNodeProvier)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetDefaultCpInstanceTypeAndNodeProvider", friendlyMessage)
	}

	// **TEMP** Obtain the OS Image name for CPs (22.04). Will change once customers can select OS Images
	var (
		cpOsImageName string
	)
	err = dbconn.QueryRowContext(ctx, GetDefaultCpOsImageQuery).Scan(&cpOsImageName)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetOsImageName", friendlyMessage)
	}

	// Validate K8s Compatability and Return IMI Components
	var (
		cpOsImageInstanceName string // IMI for the control plane
		cpK8sVersionName      string // Control Plane Full k8sversion name (patch name)
		query                 string // Query used to determin if rke2 or iks
	)
	if cpProviderName != "" && cpProviderName == "rke2" {
		query = GetDefaultCpOsImageInstance + " AND ks.k8sversion_name = $5"
	} else {
		query = GetDefaultCpOsImageInstance + " AND ks.minor_version = $5"
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
		if errors.Is(err, sql.ErrNoRows) {
			return returnError, status.Error(codes.FailedPrecondition, "IMI version not compatible")
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetDefaultCpOsImageInstance", friendlyMessage)
	}

	/* GET IMI ARTIFACT FOR CLUSTER REV DESIRED JSON */
	var imiartifact string
	err = dbconn.QueryRowContext(ctx, GetImiArtifactQuery, cpOsImageInstanceName).Scan(&imiartifact)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetImiArtifactQuery.cpOsImageInstanceName", friendlyMessage)
	}

	/* GET default ADD Ons */
	//defaultaddonvalues, err := utils.GetDefaultAddons(ctx, dbconn, cpK8sVersionName)
	defaultaddonvalues, err := utils.GetAddons(ctx, dbconn, true, true, cpK8sVersionName, "kubernetes")
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetDefaultAddons", friendlyMessage)
	}

	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"dbconn.BeginTx", friendlyMessage)
	}

	_, err = tx.ExecContext(ctx, LockClusterTable)
	log.Info(fmt.Sprintf("%s started create transaction for cluster %s", functionName, record.Name))

	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"LockClusterTable", friendlyMessage)
	}

	/* Validate Cluster Count */
	var clustercount int
	err = dbconn.QueryRowContext(ctx, GetClusterCounts,
		record.CloudAccountId,
	).Scan(&clustercount)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetClusterCounts", friendlyMessage)
	}
	if clustercount >= maxClusterCount {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, status.Errorf(codes.PermissionDenied, "Can not create more than %d clusters for this cloud account", maxClusterCount)
	}

	/* INSERT CLUSTER TABLE */
	// Insert cluster
	var clusterId int32
	id, err := utils.GenerateUuid(ctx, dbconn, utils.ClusterUUIDType)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"utils.GenerateUuid", friendlyMessage)
	}
	uniqueId := "cl-" + id

	var clusterCrd clusterv1alpha.Cluster
	clusterCrd.Name = uniqueId
	clusterCrd.Annotations = make(map[string]string, 0)
	clusterCrd.Spec = clusterv1alpha.ClusterSpec{
		KubernetesVersion:    cpK8sVersionName,
		InstanceType:         cpInstanceType,
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
		Addons:                 make([]clusterv1alpha.AddonTemplateSpec, 0), // ??? NEED TO IMPLEMENT DEFAULT LOGIC
		Nodegroups:             make([]clusterv1alpha.NodegroupTemplateSpec, 0),
		EtcdBackupEnabled:      false,
		EtcdBackupConfig:       clusterv1alpha.EtcdBackupConfig{},
		AdvancedConfig:         clusterv1alpha.AdvancedConfig{}, // ??? NEED DEFAULT OR USER INPUT
		CustomerCloudAccountId: record.CloudAccountId,
		CloudAccountId:         defaultvalues["cp_cloudaccountid"],
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

	// Add default values for cloudmonitorEnable
	if ok := defaultvalues["cloudmonitorEnable"]; ok == "true" {
		annotationsMap["cloudmonitorEnable"] = defaultvalues["cloudmonitorEnable"]
	} else {
		annotationsMap["cloudmonitorEnable"] = "false"
	}

	clusterCrd.Annotations = annotationsMap

	// vnets
	vnets, err := json.Marshal(clusterCrd.Spec.VNETS)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"marshal.clustercrd.spec.vnets", friendlyMessage)
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

	// Add default value "CloudmonitoringEnable" to annotations - IDCK8S-999
	if cloudmonitorEnable, ok := defaultvalues["cloudmonitorEnable"]; ok {
		record.Annotations = append(record.Annotations, &pb.Annotations{Key: "cloudmonitorEnable", Value: cloudmonitorEnable})
	}

	recordAnnotations, err := json.Marshal(record.Annotations)

	err = tx.QueryRowContext(ctx, InsertClusterRecordQuery,
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
		record.CloudAccountId,
		`{"status":"cluster is being created"}`, // NEED DEFAULT LOGIC
		clusterType,
	).Scan(&clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertClusterRecordQuery", friendlyMessage)
	}

	/* CREATE CONTROL PLANE NODEGROUP TABLE */
	var cpNodegroupId int32
	cpUniqueId := "cp-" + id
	var kubernetesstatus clusterv1alpha.NodegroupStatus
	kubernetesstatus.Name = cpUniqueId
	kubernetesstatus.State = "Updating"
	nodegroupstatus, err := json.Marshal(kubernetesstatus)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"marshal.kubernetesstatus", friendlyMessage)
	}

	nonce, encryptionid, sskpkey, sshpubkey, err := controlplanesshencryption(ctx, key, keypub, filepath)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"ControlPlaneSshEnc", friendlyMessage)
	}
	// Insert ssh key details into cluster_extraconfig table
	_, err = tx.QueryContext(ctx, InsertSshkeyQuery, clusterId, sskpkey, sshpubkey, sshkey, nonce, encryptionid)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return returnError, utils.ErrorHandler(ctx, err, functionName+"TransactionRollbackSshKeyError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertSshkeyQuery", friendlyMessage)
	}

	sshkeysarray := []*pb.SshKey{}
	sshkeyarray := &pb.SshKey{
		Sshkey: sshkey,
	}

	sshkeysarray = append(sshkeysarray, sshkeyarray)

	sshkeyjson, err := json.Marshal(sshkeysarray)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"Marshal Ssh key", friendlyMessage)
	}
	err = tx.QueryRowContext(ctx, InsertControlPlaneQuery,
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
		nodegroupType,
	).Scan(&cpNodegroupId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertControlPlaneQuery", friendlyMessage)
	}

	// Load vip
	var vipid int32
	dnsalias := []string{}

	// provider is default
	var vipprovider string
	err = tx.QueryRowContext(ctx, GetVipProviderDefaults).Scan(&vipprovider)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetVipProviderDefaults", friendlyMessage)
	}

	for _, cpilb := range clusterCrd.Spec.ILBS {
		dnsAliasJson, err := json.Marshal(dnsalias)
		err = tx.QueryRowContext(ctx, InsertVipQuery,
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
				return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertVipQuery", friendlyMessage)
		}

		/*Insert VIP details table */
		var vipId int32
		err = tx.QueryRowContext(ctx, InsertVipDetailsQuery, vipid, cpilb.Name, cpilb.Description, cpilb.Port, cpilb.Pool.Name, cpilb.Port).Scan(&vipId)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertVipDetailsQuery", friendlyMessage)
		}
	}

	/* CREATE CLUSTER REV TABLE */
	var revversion string
	clusterCrdJson, err := json.Marshal(clusterCrd)
	err = tx.QueryRowContext(ctx, InsertRevQuery,
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
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertRevQuery", friendlyMessage)
	}

	/* INSERT PROVISIONING LOG */
	_, err = tx.QueryContext(ctx,
		InsertProvisioningQuery,
		clusterId,
		"cluster create pending",
		"INFO",
		"cluster create",
		time.Now(),
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"insertProvisioningQuery", friendlyMessage)
	}

	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"tx.commit", friendlyMessage)
	}

	log.Info(fmt.Sprintf("%s finished create transaction for cluster %s", functionName, record.Name))

	returnValue.Uuid = uniqueId
	returnValue.Clusterstate = "Pending"
	return returnValue, err
}

// CreateNodeGroupRecord to insert record into nodegroup table
func CreateNodeGroupRecord(ctx context.Context, dbconn *sql.DB, record *pb.CreateNodeGroupRequest) (*pb.NodeGroupResponseForm, error) {
	friendlyMessage := "Could not create Node Group. Please try again."
	functionName := "CreateNodeGroupRecord."
	returnError := &pb.NodeGroupResponseForm{}
	log := log.FromContext(ctx).WithName("CreateNewNodegroup")

	/*Validate IKS instance types */
	var instancetype, clusterNodeGroupType string
	var activeinstances []string
	var instances []string
	var instanceexistsiks bool
	var instanceactive bool
	instanceexistsiks = false
	instanceactive = false

	/*Default Values for IKS */
	defaultvalues, err := utils.GetDefaultValues(ctx, dbconn)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetDefaultValues", friendlyMessage)
	}

	/* Validate All IKS Instance Types */
	rows, err := dbconn.QueryContext(ctx, getInstanceTypeQuery)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetIKSInstanceTypesGetQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&instancetype)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, functionName+"GetIKSInstanceTypesRollback", "Get Iks instance types Transaction rollback")
		}
		instances = append(instances, instancetype)
	}

	for _, ins := range instances {
		if ins == record.Instancetypeid {
			instanceexistsiks = true
		}

	}

	if !instanceexistsiks {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetIKSDBInstanceTypes", "Instance type is not supported by IKS")
	}

	/*Validate Active instance types */
	rows, err = dbconn.QueryContext(ctx, getActiveInstanceTypeQuery)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetIKSInstanceTypesActive", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&instancetype)
		if err != nil {
			return returnError, err
		}
		activeinstances = append(activeinstances, instancetype)
	}

	for _, activeIns := range activeinstances {
		if activeIns == record.Instancetypeid {
			instanceactive = true
			break
		}
	}

	if !instanceactive {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"ValidateInstanceTypesActiveState", "Instance Type is not supported by iks")
	}

	/* VALIDATE INSTANCE GROUP */
	isInstanceGroup := utils.IsInstanceGroup(record.Instancetypeid)
	if isInstanceGroup && record.Count > 1 {
		return returnError, status.Error(codes.InvalidArgument, "Only one instance group is allowed per Nodegroup")
	}

	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, max_nodegroups, max_nodegroup_vm, _, _, _, err := convDefaultsToInt(ctx, defaultvalues)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"convDefaultsToInt", friendlyMessage)
	}

	cloudAccountmaxNgPerCluster := -1
	maxNodesPerNodegroup := -1
	_, cloudAccountmaxNgPerCluster, _, _, maxNodesPerNodegroup, err = utils.GetCloudAccountMaxValues(ctx, dbconn, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetCloudAccountMaxValues", friendlyMessage)
	}
	if cloudAccountmaxNgPerCluster > -1 {
		max_nodegroups = cloudAccountmaxNgPerCluster
	}
	if maxNodesPerNodegroup > -1 && maxNodesPerNodegroup > max_nodegroup_vm {
		max_nodegroup_vm = maxNodesPerNodegroup
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
		Clusteruuid:    record.Clusteruuid,
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
		Nodegrouptype:        &clusterNodeGroupType,
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

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err = utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	aiNodegroupTypeList := strings.Split(defaultvalues["sc_ai_nodegrouptype"], ",")

	/* GET CLUSTER TYPE */
	isSuperComputeCluster, err := utils.ValidateSuperComputeClusterType(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, err
	}
	if isSuperComputeCluster {
		// check if request has nodegroup type and ensure its supercompute-gp
		nodeGroupType := record.Nodegrouptype
		if nodeGroupType == nil {
			return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, functionName, "SuperCompute Cluster: NodeGroupType cannot be empty.", codes.InvalidArgument)
		} else if *nodeGroupType != superComputeGPNodegroupType {
			return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, functionName, "SuperCompute Cluster: Invalid NodeGroup type", codes.InvalidArgument)
		}

		/* Validate for supercompute cluster if the instance type is Gaudi */
		isValidSCAIInstanceType := utils.ValidateSCAiNodegroupInstanceType(record.Instancetypeid, aiNodegroupTypeList)
		if isValidSCAIInstanceType {
			return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, functionName, "SuperCompute Cluster: Cannot create Gaudi Instance type", codes.InvalidArgument)
		}

		clusterNodeGroupType = *nodeGroupType

	} else {
		// update the nodegroup type as "gp"
		clusterNodeGroupType = generalPurposeNodeGroupType
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* VALIDATE CLUSTER IS ACTIONABLE*/
	actionableState := false
	actionableState, err = utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return returnError, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	/* VALIDATE NODEGROUP UNIQUE NAME */
	count := 0
	err = dbconn.QueryRowContext(ctx, GetNodeGroupCountByName, record.Clusteruuid, record.Name).Scan(&count)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetNodeGroupCountByName", friendlyMessage)
	}
	if count != 0 {
		return returnError, status.Error(codes.AlreadyExists, "Nodegroup name already in use")
	}

	/* VALIDATE COMPATABILITY TABLES*/
	// Get information from the Control Plane
	clusterProvider, clusterK8sVersion, clusterRuntime, _, _, err := utils.GetClusterCompatDetails(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetClustersCompatDetails", friendlyMessage)
	}

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
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetOsImageName", friendlyMessage)
	}

	// Get Default worker IMI
	query := GetDefaultWrkOsImageInstance + " AND ks.k8sversion_name = $5"
	var wrkOsImageInstanceName string
	err = dbconn.QueryRowContext(ctx, query,
		clusterRuntime,
		clusterProvider,
		record.Instancetypeid,
		wrkOsImageName,
		clusterK8sVersion,
	).Scan(&wrkOsImageInstanceName)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, status.Error(codes.FailedPrecondition, "IMI version not compatiblePlease contact support regarding this issue.")
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"getDefaultWrkOsImageInstance.query", friendlyMessage)
	}

	// Get IMI artifact
	imiartifact, err := utils.GetImiArtifact(ctx, dbconn, wrkOsImageInstanceName)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, status.Error(codes.FailedPrecondition, "IMI version not compatible. Please contact support regarding this issue.")
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetImiArtifact", friendlyMessage)
	}

	/* Get cluster id from cluster  uuid */
	var clusterid int
	err = dbconn.QueryRowContext(ctx, GetClusterIdQuery,
		record.Clusteruuid,
	).Scan(&clusterid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"Getclusteridfromuuid", friendlyMessage)
	}

	/* CREATE ENTRY IN NODEGROUP TABLE */
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"dbconn.BeginTx", friendlyMessage)
	}

	log.Info(fmt.Sprintf("%s started create transaction for nodegroup %s", functionName, record.Name))

	_, err = tx.ExecContext(ctx, LockNodegroupTable)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"LockNodegroupTable", friendlyMessage)
	}

	/* Validate Nodegroup Count */
	var nodegroupcount int
	err = dbconn.QueryRowContext(ctx, GetNodeGroupCounts,
		clusterid,
	).Scan(&nodegroupcount)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetnodegroupCount", friendlyMessage)
	}

	if nodegroupcount >= max_nodegroups {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, functionName+"iksnodegroupCount", fmt.Sprintf("Can not create more than %d nodegroups for this cluster", max_nodegroups), codes.PermissionDenied)
	}

	/* Validate Nodegroup nodes Count for non cluster instance types*/
	if !isInstanceGroup && int(record.Count) > max_nodegroup_vm {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, functionName+"iksnodegroupCount", fmt.Sprintf("Can not create more than %d nodes for this nodegroup", max_nodegroup_vm), codes.PermissionDenied)
	}

	var nodegroupid int32
	id, err := utils.GenerateUuid(ctx, dbconn, utils.NodegroupUUIDType)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"utils.GenerateUuid.NG", friendlyMessage)
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
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"marshal.kubernetesstatus", friendlyMessage)
	}

	vnets, err := json.Marshal(record.Vnets)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"marshal.vnets", friendlyMessage)
	}

	err = tx.QueryRowContext(ctx, InsertNodeGroupQuery,
		clusterId,
		wrkOsImageInstanceName,
		clusterK8sVersion,
		"Updating",
		"Worker",
		record.Name,
		record.Description,
		defaultvalues["networkinterfacename"],
		record.Instancetypeid,
		record.Count,
		record.Sshkeyname,
		clusterRuntime,
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
		clusterNodeGroupType,
	).Scan(&nodegroupid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertNodeGroupQuery", friendlyMessage)
	}
	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, UpdateClusterStateQuery,
		clusterId,
		"Pending",
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"UpdateClusterStateQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER REV TABLE*/
	// Get current json from REV table
	// Create Nodegroup CRD and append to clusterCrd
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetLatestClusterRev", friendlyMessage)
	}
	nodeGroupCrd := clusterv1alpha.NodegroupTemplateSpec{
		Name:              ngUniqueId,
		KubernetesVersion: clusterK8sVersion,
		InstanceType:      record.Instancetypeid,
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
		CloudAccountId: record.CloudAccountId,
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
		clusterCrd.Spec.SSHKey = append(clusterCrd.Spec.SSHKey, keys.Sshkey)
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

	// Create new cluster rev table entry
	var revversion string
	err = tx.QueryRowContext(ctx, InsertRevQuery,
		clusterId,
		currentJson,
		clusterCrd,
		"test",
		"test",
		"test",
		"test",
		time.Now(),
		false,
	).Scan(&revversion)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertRevQuery", friendlyMessage)
	}

	/* COMMIT TRANSACTION */
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"tx.Commit", friendlyMessage)
	}

	log.Info(fmt.Sprintf("%s finished create transaction for nodegroup %s", functionName, record.Name))

	/* GET NODEGROUP STATUS */
	getNodeGroupStatusReq := &pb.NodeGroupid{
		Clusteruuid:    record.Clusteruuid,
		Nodegroupuuid:  ngUniqueId,
		CloudAccountId: record.CloudAccountId,
	}
	getNodegroupStatusResponse, err := GetNodeGroupStatusRecord(ctx, dbconn, getNodeGroupStatusReq)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetStatusRecord", friendlyMessage)
	}
	returnValue.Nodegroupstatus = getNodegroupStatusResponse

	return returnValue, nil

}

func CreateNewVip(ctx context.Context, dbconn *sql.DB, record *pb.VipCreateRequest) (*pb.VipResponse, error) {
	log := log.FromContext(ctx).WithName("CreateNewVip")

	friendlyMessage := "Could not create VIP. Please try again."
	functionName := "CreateNewVip."
	returnError := &pb.VipResponse{}
	returnValue := &pb.VipResponse{
		Vipid:       0,
		Name:        record.Name,
		Description: record.Description,
		Vipstate:    "Pending",
		Port:        record.Port,
		Poolport:    record.Port,
		Viptype:     record.Viptype,
		Dnsalias:    []string{},
	}

	/*Default Values for IKS */
	defaultvalues, err := utils.GetDefaultValues(ctx, dbconn)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetDefaultValues", friendlyMessage)
	}

	_, _, ilbcustomerenv, ilbcustomerusergroup, minActiveMembers, memberConnectionLimit, memberPriorityGroup, memberRatio, _, _, _, _, _, _, _, _, maxvips, _, _, ilbportone, ilbporttwo, _, err := convDefaultsToInt(ctx, defaultvalues)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"convDefaultsToInt", friendlyMessage)
	}
	cloudAccountmaxIlbsPerCluster := -1
	_, _, cloudAccountmaxIlbsPerCluster, _, _, err = utils.GetCloudAccountMaxValues(ctx, dbconn, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetCloudAccountMaxValues", friendlyMessage)
	}
	if cloudAccountmaxIlbsPerCluster > -1 {
		maxvips = cloudAccountmaxIlbsPerCluster
	}

	/* VALIDATE VIP TYPE */
	if record.GetViptype() != "public" {
		if record.GetViptype() != "private" {
			return returnError, status.Error(codes.InvalidArgument, "Vip type should be 'public' or 'private'")
		}
	}

	/* Validate input ports */
	if record.GetPort() != int32(ilbportone) && record.GetPort() != int32(ilbporttwo) {
		return returnError, status.Error(codes.InvalidArgument, "Vip ports are not correct")
	}

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err = utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}
	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}
	/* VALIDATE CLUSTER IS ACTIONABLE*/
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return returnError, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	/* Get cluster id from cluster  uuid */
	var clusterid int
	err = dbconn.QueryRowContext(ctx, GetClusterIdQuery,
		record.Clusteruuid,
	).Scan(&clusterid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"Getclusteridfromuuid", friendlyMessage)
	}

	/* Validate Nodegroup Count */
	var nodegroupcount int
	err = dbconn.QueryRowContext(ctx, GetNodeGroupCounts,
		clusterid,
	).Scan(&nodegroupcount)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetnodegroupCount", friendlyMessage)
	}

	if nodegroupcount < 1 {
		return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, functionName+"iksnodegroupCount", "Can not create a VIP without at least 1 nodegroup", codes.FailedPrecondition)
	}

	/* START THE TRANSACTION */
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"dbconn.BeginTx", friendlyMessage)
	}
	log.Info(fmt.Sprintf("%s started create transaction for vip %s", functionName, record.Name))

	// Lock VIP tables
	_, err = tx.ExecContext(ctx, LockVIPTable)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"LockVipTable", friendlyMessage)
	}

	/* VALIDATE VIP NAME UNIQUENESS*/
	var vipNameCount int
	err = dbconn.QueryRowContext(ctx, GetVipByName, clusterid, record.Name).Scan(&vipNameCount)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetVipByName", friendlyMessage)
	}
	if vipNameCount >= 1 {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, functionName+"Vip name already in use", "Vip name already in use", codes.FailedPrecondition)
	}

	/* Validate ILB counts */
	var vipcounts int
	err = dbconn.QueryRowContext(ctx, GetILBsCount,
		clusterid,
	).Scan(&vipcounts)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GeVipCount", friendlyMessage)
	}

	if vipcounts >= maxvips {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, functionName+"iksvipCount", fmt.Sprintf("Can not create more than %d User Vips for this cluster", maxvips), codes.PermissionDenied)
	}

	/* LOAD THE VIP */
	var vipid int32
	dnsalias := []string{}
	// provider is default
	var vipprovider string
	err = tx.QueryRowContext(ctx, GetVipProviderDefaults).Scan(&vipprovider)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetVipProviderDefaults", friendlyMessage)
	}
	err = tx.QueryRowContext(ctx, InsertVipQuery,
		clusterId,
		dnsalias,
		"",
		record.Viptype,
		`{"status":"Vip is being created"}`,
		"Pending",
		"customer",
		vipprovider, //default from provider table
		"",
	).Scan(&vipid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertVipQuery", friendlyMessage)
	}
	returnValue.Vipid = vipid

	/*Insert VIP details table */
	var vipId int32
	err = tx.QueryRowContext(ctx, InsertVipDetailsQuery, vipid, record.Name, record.Description, record.Port, record.Name, record.Port).Scan(&vipId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertVipDetailsQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, UpdateClusterStateQuery,
		clusterId,
		"Pending",
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"UpdateClusterStateQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER REV TABLE*/
	// Get current json from REV table
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"GetLatestClusterRev", friendlyMessage)
	}

	vipCrd := clusterv1alpha.ILBTemplateSpec{
		Name:        record.Name,
		Description: record.Description,
		Port:        int(record.Port),
		IPType:      record.Viptype,
		IPProtocol:  defaultvalues["ilb_ipprotocol"],
		Environment: ilbcustomerenv,
		Usergroup:   ilbcustomerusergroup,
		Persist:     defaultvalues["ilb_persist"],
		Owner:       "customer",
		Pool: clusterv1alpha.ILBPoolTemplateSpec{
			Name:                  record.Name,
			Port:                  int(record.Port),
			LoadBalancingMode:     defaultvalues["ilb_loadbalancingmode"],
			MinActiveMembers:      minActiveMembers,
			Monitor:               defaultvalues["ilb_monitor"],
			MemberConnectionLimit: memberConnectionLimit,
			MemberPriorityGroup:   memberPriorityGroup,
			MemberRatio:           memberRatio,
			MemberAdminState:      defaultvalues["ilb_memberadminstate"],
		},
	}

	clusterCrd.Spec.ILBS = append(clusterCrd.Spec.ILBS, vipCrd)

	// Create new cluster rev table entry
	var revversion string
	err = tx.QueryRowContext(ctx, InsertRevQuery,
		clusterId,
		currentJson,
		clusterCrd,
		"test", // ?? DEFAULT VALUES
		"test", // ?? DEFAULT VALUES
		"test", // ?? DEFAULT VALUES
		"test", // ?? DEFAULT VALUES
		time.Now(),
		false,
	).Scan(&revversion)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, functionName+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, functionName+"InsertRevQuery", friendlyMessage)
	}

	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, functionName+"tx.commit", friendlyMessage)
	}
	log.Info(fmt.Sprintf("%s finished create transaction for vip %s", functionName, record.Name))

	return returnValue, nil
}

func convDefaultsToInt(ctx context.Context, defaultvalues map[string]string) (int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, error) {
	failedFunction := "convDefaultsToIn"
	friendlyMessage := "Could not conver defaults to int"

	/* convert string to Int */
	ilbenv, err := strconv.Atoi(defaultvalues["ilb_environment"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_environment", friendlyMessage)
	}

	ilbusergroup, err := strconv.Atoi(defaultvalues["ilb_usergroup"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_usergroup", friendlyMessage)
	}

	ilbcustomerenv, err := strconv.Atoi(defaultvalues["ilb_customer_environment"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_customerenvironment", friendlyMessage)
	}

	ilbcustomerusergroup, err := strconv.Atoi(defaultvalues["ilb_customer_usergroup"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_customerusergroup", friendlyMessage)
	}

	minActiveMembers, err := strconv.Atoi(defaultvalues["ilb_minactivemembers"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_minactivemembers", friendlyMessage)
	}

	memberConnectionLimit, err := strconv.Atoi(defaultvalues["ilb_memberConnectionLimit"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues._ilb_memberConnectionLimit", friendlyMessage)
	}

	memberPriorityGroup, err := strconv.Atoi(defaultvalues["ilb_memberPriorityGroup"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_memberPriorityGroup", friendlyMessage)
	}

	memberRatio, err := strconv.Atoi(defaultvalues["ilb_memberratio"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_memberratio", friendlyMessage)
	}

	etcdport, err := strconv.Atoi(defaultvalues["ilb_etcdport"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_etcdport", friendlyMessage)
	}

	etcdpool_port, err := strconv.Atoi(defaultvalues["ilb_etcdpool_port"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_etcdpoolport", friendlyMessage)
	}

	apiport, err := strconv.Atoi(defaultvalues["ilb_apiserverport"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_apiport", friendlyMessage)
	}

	apiserverpool_port, err := strconv.Atoi(defaultvalues["ilb_apiserverpool_port"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_apiserverpoolport", friendlyMessage)
	}

	public_apiserverport, err := strconv.Atoi(defaultvalues["ilb_public_apiserverport"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_publicapiserverport", friendlyMessage)
	}

	public_apiserverpool_port, err := strconv.Atoi(defaultvalues["ilb_public_apiserverpool_port"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_publicapiserverpoolport", friendlyMessage)
	}

	konnectPort, err := strconv.Atoi(defaultvalues["ilb_konnectivityport"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_kinnectivityport", friendlyMessage)
	}

	konnectPoolPort, err := strconv.Atoi(defaultvalues["ilb_konnectivitypool_port"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_kinnectivitpoolyport", friendlyMessage)
	}

	max_cust_cluster_ilb, err := strconv.Atoi(defaultvalues["max_cust_cluster_ilb"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.max_cust_cluster_ilb", friendlyMessage)
	}

	max_cluster_ng, err := strconv.Atoi(defaultvalues["max_cluster_ng"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.max_cust_cluster_ilb", friendlyMessage)
	}

	max_nodegroup_vm, err := strconv.Atoi(defaultvalues["max_nodegroup_vm"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.max_nodegroup_vm", friendlyMessage)
	}

	split := strings.Split(defaultvalues["ilb_allowed_ports"], ",")

	var ilbportone int
	var ilbporttwo int
	for range split {
		ilbportone, err = strconv.Atoi(string(split[0]))
		if err != nil {
			return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilballowedport", friendlyMessage)
		}

		ilbporttwo, err = strconv.Atoi(split[1])
		if err != nil {
			return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilballowedport", friendlyMessage)
		}
	}

	max_cluster, err := strconv.Atoi(defaultvalues["max_cluster"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, utils.ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.max_nodegroup_vm", friendlyMessage)
	}

	return ilbenv, ilbusergroup, ilbcustomerenv, ilbcustomerusergroup, minActiveMembers, memberConnectionLimit, memberPriorityGroup, memberRatio, etcdport, etcdpool_port, apiport, apiserverpool_port, public_apiserverport, public_apiserverpool_port, konnectPort, konnectPoolPort, max_cust_cluster_ilb, max_cluster_ng, max_nodegroup_vm, ilbportone, ilbporttwo, max_cluster, nil
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
