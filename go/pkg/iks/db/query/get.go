// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/pointer"
	"strconv"
	"time"

	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	ilbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ilb_operator/api/v1alpha1"
	clusterv1alpha "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	iksOperator "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/kubernetes_provider/iks"
	kubeUtils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/utils"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"gopkg.in/yaml.v3"
)

const (
	GetClusterSecrets = `
		SELECT
			cluster_cacrt,
			cluster_cakey,
			encryptionkey_id,
			nonce
		FROM cluster_extraconfig c
		WHERE c.cluster_id = $1
	`

	GetDefaultConfigQuery = `
		SELECT value
		FROM  defaultconfig
		WHERE name = $1
	`

	GetNodeGroupCountByName = `
		SELECT count(nodegroup_id) as count
		FROM public.nodegroup n
		WHERE n.name = $2 and n.nodegrouptype_name = 'Worker' and cluster_id = (select cluster_id from public."cluster" c where c.unique_id = $1)
	`
	getClusterDetailsQuery = `
  SELECT c.cluster_id, c.name, c.description, c.clusterstate_name, c.region_name,
			c.networkservice_cidr, c.networkpod_cidr, c.cluster_dns, c.provider_name, n.k8sversion_name, n.runtime_name, c.created_date, c.storage_enable, c.clustertype
		FROM public.cluster c
			INNER JOIN nodegroup n
			ON c.cluster_id = n.cluster_id AND n.nodegrouptype_name = 'ControlPlane'
		WHERE c.unique_id = $1 AND c.cloudaccount_id = $2
	`

	getNodeGroupDetailsQuery = `
	SELECT n.nodegroup_id, n.unique_id, n.name, n.description, n.instancetype_name, n.nodegroupstate_name,
		n.nodecount, n.sshkey, n.osimageinstance_name, n.upgstrategydrainbefdel, n.upgstrategymaxnodes,
		n.nodegrouptype_name, n.k8sversion_name, n.runtime_name, c.provider_name ,n.networkinterface_name, n.vnets, n.createddate, n.userdata_webhook, n.nodegrouptype, c.clustertype
	  FROM public.nodegroup n
		INNER JOIN public.cluster c
		ON c.cluster_id = n.cluster_id
	WHERE c.unique_id = $1 AND n.unique_id = $2
	`
	getNodegroupNodesQuery = `
		SELECT
			n.k8snode_name,
			n.ip_address,
			COALESCE(n.dns_name, ''),
			COALESCE(n.osimageinstance_name, ''),
			n.k8snodestate_name,
			n.kubernetes_status,
			n.created_date,
			n.weka_storage_client_id,
			n.weka_storage_status,
			n.weka_storage_custom_status,
			n.weka_storage_message 
		FROM public.k8snode n
		WHERE n.nodegroup_id = $1
	`
	getAllClusterNodeGroupsUuidsQuery = `
		SELECT unique_id
		FROM public.nodegroup
		WHERE cluster_id = $1 AND nodegrouptype_name = 'Worker'
	`
	getClusterUuidsQuery = `
	  SELECT unique_id
		FROM public.cluster
		WHERE clusterstate_name != 'Deleted'
	`
	getAllClusterIds = `
	  SELECT cluster_id
		FROM public.cluster
	`
	getAllClusterUuids = `
	  SELECT unique_id
		FROM public.cluster
	`
	getAllActiveClusterUuids = `
	  SELECT unique_id
		FROM public.cluster c
		WHERE c.clusterstate_name != 'Deleted' AND c.cloudaccount_id =$1
	`
	getClusterIdByUuidQuery = `
	  SELECT cluster_id
		FROM public.cluster
		WHERE unique_id = $1
	`
	getMetadataK8sVersionQuery = `
		SELECT  ks.minor_version, ks.k8sversion_name, json_agg(distinct runtime_name) as compat
		FROM public.k8sversion ks
		INNER JOIN k8scompatibility kc
			ON ks.k8sversion_name = kc.k8sversion_name AND kc.provider_name = $1
		WHERE ks.lifecyclestate_id = (
			SELECT lifecyclestate_id 
			FROM lifecyclestate 
			WHERE name = 'Active') AND ks.test_version = false
		GROUP BY ks.k8sversion_name
	`
	getMetadataRuntimeQuery = `
		SELECT r.runtime_name, json_agg(distinct ks.minor_version) as compat, json_agg(distinct ks.k8sversion_name) as compat2
		FROM public.runtime r
		INNER JOIN k8scompatibility kc
			ON r.runtime_name = kc.runtime_name
		INNER JOIN k8sversion ks
			ON ks.k8sversion_name = kc.k8sversion_name AND kc.provider_name = $1
		WHERE ks.lifecyclestate_id = (
			SELECT lifecyclestate_id
			FROM lifecyclestate
			WHERE name = 'Active') AND ks.test_version = false
		GROUP BY r.runtime_name
	`
	getMetadataInstanceTypeQuery = `
		SELECT 
			i.instancetype_name,
			i.memory,
			i.cpu,
			i.storage,
			coalesce(i.displayname, i.instancetype_name),
			coalesce(i.description, ''),
			coalesce(i.instancecategory, '')
    FROM instancetype i
			WHERE i.nodeprovider_name = (SELECT nodeprovider_name FROM nodeprovider WHERE is_default = true) AND i.lifecyclestate_id = (
			SELECT lifecyclestate_id
			FROM lifecyclestate
			WHERE name = 'Active'
		)
	`

	getInstanceTypeQuery = `
	SELECT i.instancetype_name
	FROM instancetype i 
	WHERE i.nodeprovider_name = (SELECT nodeprovider_name from public.nodeprovider where is_default = true) 
	`
	GetNodegroupInstanceTypeQuery = `
	 SELECT n.instancetype_name
	 FROM nodegroup n
	 WHERE n.unique_id = $1
	`
	getActiveInstanceTypeQuery = `
	SELECT i.instancetype_name FROM instancetype i where i.lifecyclestate_id = (
		SELECT lifecyclestate_id
		FROM lifecyclestate
		WHERE name = 'Active') and i.nodeprovider_name = (SELECT nodeprovider_name from public.nodeprovider where is_default = true)
	`
	getNodeGroupUpgradeStrategyQuery = `
		SELECT upgstrategydrainbefdel, upgstrategymaxnodes
		FROM public.nodegroup n
			INNER JOIN public.cluster c
			ON c.cluster_id = n.cluster_id
		WHERE c.unique_id = $1 AND n.unique_id = $2
	`
	getClusterStatusInfoQuery = `
		SELECT c.name, c.clusterstate_name, ks.minor_version
		FROM public.cluster c
			INNER JOIN public.k8sversion ks
			ON ks.k8sversion_name = (
				SELECT k8sversion_name
				FROM public.nodegroup n
				WHERE n.cluster_id = c.cluster_id AND n.nodegrouptype_name = 'ControlPlane'
			)
		WHERE c.unique_id = $1
	`
	GetClusterIdQuery = `
	  SELECT cluster_id FROM public.cluster where unique_id = $1
	`
	GetClusterStatusQuery = `
		SELECT name, clusterstate_name, kubernetes_status FROM public.cluster  WHERE unique_id = $1 ORDER BY created_date DESC LIMIT 1
	`

	GetNodeGroupStatusQuery = `
	 SELECT name, nodecount, nodegroupstate_name, kubernetes_status FROM public.nodegroup WHERE unique_id = $1
	 `
	GetAdminKubeconfigQuery = `
	SELECT coalesce(admin_kubeconfig, '') as admin_kubeconfig FROM public.cluster where unique_id = $1`

	GetReadOnlyKubeconfigQuery = `
	SELECT coalesce(readonly_kubeconfig, '') as readonly_kubeconfig FROM public.cluster where unique_id = $1`

	GetVipsStatusByClusterQuery = `
		SELECT  v.vip_status
		FROM public.vip v
		WHERE v.cluster_id = $1 AND v.viptype_name = 'public' AND v.owner = 'system'
		`
	GetVipsQuery = `
	SELECT v.vip_id , d.vip_name, d.description , v.vipstate_name , v.vip_ip , d.port, d.pool_port, v.viptype_name, v.dns_aliases, v.vip_status, v.created_date
	FROM public.vip v
		INNER JOIN public.vipdetails d ON 
		v.vip_id = d.vip_id
	WHERE v.cluster_id = $1 AND v.vip_id = $2 AND v.owner = 'customer' 
	`

	GetVipidByClusterQuery = `
	SELECT vip_id FROM public.vip where cluster_id = $1 and viptype_name = 'public' and vipstate_name IN ('Active', 'Deleting')`

	GetvipidForFirewallQuery = `
	SELECT vip_id,viptype_name FROM public.vip where vip_ip = $1 AND cluster_id = $2`

	GetVipPortQuery = `
	SELECT port , pool_port, vip_name FROM public.vipdetails where vip_id = $1 and port = $2`

	GetSoureIpsQuery = `
	SELECT sourceips FROM public.vip where vip_id = $1`

	GetFirewallValuesQuery = `
	SELECT v.vip_ip , d.port
	FROM public.vip v
		INNER JOIN public.vipdetails d ON 
		v.vip_id = d.vip_id
	WHERE v.cluster_id = $1 AND v.vip_id = $2
	`
	GetPublicVipsForFirewallQuery = `
	SELECT v.vip_id, v.vip_ip , d.port, d.pool_port, v.sourceips, COALESCE(v.firewall_status,'Not Specified'), v.viptype_name ,d.protocol , d.vip_name
	FROM public.vip v
	    INNER JOIN public.vipdetails d ON
		v.vip_id = d.vip_id
	WHERE v.cluster_id = $1 AND v.vip_id = $2 AND v.viptype_name = 'public';
	`
	GetProtocolQuery = `
	SELECT vip_name,protocol from public.vipdetails WHERE vip_id = $1`

	GetVipMembersQuery = `
	SELECT ip_address FROM public.vipmembers where pool_id = (SELECT pool_id FROM public.vipdetails
		where vip_id = $1)`

	GetVipIdsQuery = `
	SELECT vip_id FROM public.vip where cluster_id = $1 and owner = 'customer'`

	GetVipNameQuery = `
	SELECT vip_name FROM public.vipdetails where vip_id = $1`

	GetMetadata = `
	SELECT tags, annotations FROM public.cluster where cluster_id = $1`

	GetNodeGroupCounts = `
	SELECT count(*) FROM public.nodegroup WHERE cluster_id = $1 and nodegrouptype_name = 'Worker'
	`
	GetClusterCounts = `
	SELECT count(*) FROM public.cluster WHERE cloudaccount_id = $1 AND clusterstate_name NOT IN ('Deleting','DeletePending','Deleted')
	`
	GetWorkerNodeGroups = `
	SELECT nodegroup_id,nodecount FROM public.nodegroup WHERE cluster_id = $1 AND nodegrouptype_name = 'Worker'
	`
	GetILBsCount = `
	SELECT count(*) FROM public.vip WHERE cluster_id = $1 and owner = 'customer'`
)

func GetClusterRecord(ctx context.Context, dbconn *sql.DB, record *pb.ClusterID) (*pb.ClusterResponseForm, error) {
	friendlyMessage := "Could not get Cluster. Please try again"
	failedFunction := "GetClusterRecord."
	returnError := &pb.ClusterResponseForm{}
	returnValue := &pb.ClusterResponseForm{
		Uuid:             record.Clusteruuid,
		Name:             "",
		Description:      "",
		Clusterstate:     "",
		Clusterstatus:    &pb.ClusterStatus{},
		Createddate:      "",
		K8Sversion:       "",
		Upgradeavailable: false, // LOGIC IN DBv8
		Nodegroups:       []*pb.NodeGroupResponseForm{},
		Tags:             nil, // Need to convert
		Annotations:      nil, // Need to convert
		Network: &pb.Network{
			Region:      "",
			Servicecidr: nil,
			Clustercidr: nil,
			Clusterdns:  nil,
		},
		Clustertype: "",
		// Availabilityzonename: "",
	}

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}
	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* Get Cluster Response Information from Cluster table*/
	var (
		clusterProvider   string
		runtimeName       string
		clusterK8sVersion string
	)
	err = dbconn.QueryRowContext(ctx, getClusterDetailsQuery,
		record.Clusteruuid,
		record.CloudAccountId,
	).Scan(&clusterId, &returnValue.Name, &returnValue.Description,
		&returnValue.Clusterstate, &returnValue.Network.Region,
		&returnValue.Network.Servicecidr, &returnValue.Network.Clustercidr,
		&returnValue.Network.Clusterdns, &clusterProvider, &clusterK8sVersion, &runtimeName, &returnValue.Createddate, &returnValue.Storageenabled, &returnValue.Clustertype)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClusterDetailsQuery", friendlyMessage)
	}
	// Convert k8sversion to minor version
	returnValue.K8Sversion, err = utils.GetMinorVersion(ctx, dbconn, clusterK8sVersion)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetMinorVersion", friendlyMessage)
	}

	/* Get availabler K8sVersions for cluster */
	availablerVersions, err := utils.GetAvailableClusterVersionUpgrades(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetAvailableClusterVersionUpgrades", friendlyMessage)
	}

	returnValue.Upgradeavailable = len(availablerVersions) > 0
	returnValue.Upgradek8Sversionavailable = availablerVersions

	/* GET NODEGROUPS */
	nodeGroupsReq := &pb.GetNodeGroupsRequest{
		Clusteruuid:    record.Clusteruuid,
		CloudAccountId: record.CloudAccountId,
	}
	var getNodeGroupsResponse *pb.NodeGroupResponse
	getNodeGroupsResponse, err = GetNodeGroups(ctx, dbconn, nodeGroupsReq)
	returnValue.Nodegroups = getNodeGroupsResponse.Nodegroups

	/* GET VIPS */
	var getVipsResponse *pb.GetVipsResponse
	getVipsResponse, err = GetVips(ctx, dbconn, record)
	returnValue.Vips = getVipsResponse.Response

	/* GET ANNOTATIONS AND TAGS*/
	var tags []uint8
	var annotations []uint8
	var tagsParse []*pb.KeyValuePair
	var annotationsParse []*pb.Annotations
	err = dbconn.QueryRowContext(ctx, GetMetadata, clusterId).Scan(&tags, &annotations)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetMetadata", friendlyMessage)
	}
	err = json.Unmarshal([]byte(tags), &tagsParse)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetMetadata.unmarshall.tags", friendlyMessage)
	}
	err = json.Unmarshal([]byte(annotations), &annotationsParse)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetMetadata.unmarshall.annotations", friendlyMessage)
	}
	returnValue.Tags = tagsParse
	returnValue.Annotations = annotationsParse

	/* GET CLUSTER STORAGE */
	var getClustersStorageStatus []*pb.ClusterStorageStatus
	getClustersStorageStatus, err = utils.GetClusterStorageStatus(ctx, dbconn, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"utils.GetClusterStorageStatus", friendlyMessage)
	}
	returnValue.Storages = getClustersStorageStatus

	/* GET CLUSTER STATUS */
	getClusterStatusResponse, err := GetStatusRecord(ctx, dbconn, record)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetStatusRecord", friendlyMessage)
	}
	returnValue.Clusterstatus = getClusterStatusResponse
	return returnValue, nil
}

func GetClustersRecord(ctx context.Context, dbconn *sql.DB, record *pb.IksCloudAccountId) (*pb.ClustersResponse, error) {
	/* SET UP VARS */
	friendlyMessage := "Could not get Clusters. Please try again"
	failedFunction := "GetClustersRecord."
	returnError := &pb.ClustersResponse{}
	returnValue := &pb.ClustersResponse{
		Clusters:       []*pb.ClusterResponseForm{},
		Resourcelimits: &pb.ResourceLimits{},
	}
	/* GET CLUSTERS */
	rows, err := dbconn.QueryContext(ctx, getAllActiveClusterUuids, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getAllActiveClusterUuids", friendlyMessage)
	}
	defer rows.Close()
	/* GET CLUSTER INFORMATION */
	for rows.Next() {
		clusterIdReq := &pb.ClusterID{
			Clusteruuid:    "",
			CloudAccountId: record.CloudAccountId,
		}
		err = rows.Scan(&clusterIdReq.Clusteruuid)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getAllActiveClusterUuids.rows.scan", friendlyMessage)
		}
		getClusterResponse, err := GetClusterRecord(ctx, dbconn, clusterIdReq)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterRecord", friendlyMessage)
		}
		returnValue.Clusters = append(returnValue.Clusters, getClusterResponse)
	}
	defaultvalues, err := utils.GetDefaultValues(ctx, dbconn)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultValues", friendlyMessage)
	}
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, maxIlbsPerCluster, maxNodegroupsPerCluster, maxNodesPerNodegroup, _, _, cloudAccountMaxClusters, err := convDefaultsToInt(ctx, defaultvalues)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"convDefaultsToInt", friendlyMessage)
	}
	maxClustersPerCloudAccount, maxNodegroupsPerClusterPerAccount, cloudAccountmaxIlbsPerCluster, _, maxNodesPerCloudAccount, err := utils.GetCloudAccountMaxValues(ctx, dbconn, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetCloudAccountMaxValues", friendlyMessage)
	}
	if maxClustersPerCloudAccount != -1 {
		cloudAccountMaxClusters = maxClustersPerCloudAccount
	}
	if maxNodegroupsPerClusterPerAccount != -1 {
		maxNodegroupsPerCluster = maxNodegroupsPerClusterPerAccount
	}
	if cloudAccountmaxIlbsPerCluster != -1 {
		maxIlbsPerCluster = cloudAccountmaxIlbsPerCluster
	}
	if maxNodesPerCloudAccount != -1 {
		maxNodesPerNodegroup = maxNodesPerCloudAccount
	}
	limits := &pb.ResourceLimits{
		Maxclusterpercloudaccount: int32(cloudAccountMaxClusters),
		Maxnodegroupspercluster:   int32(maxNodegroupsPerCluster),
		Maxvipspercluster:         int32(maxIlbsPerCluster),
		Maxnodespernodegroup:      int32(maxNodesPerNodegroup),
	}
	returnValue.Resourcelimits = limits
	return returnValue, nil
}

func GetStatusRecord(ctx context.Context, dbconn *sql.DB, record *pb.ClusterID) (*pb.ClusterStatus, error) {
	friendlyMessage := "Could not get Cluster Status. Please try again"
	failedFunction := "GetStatusRecord."
	returnError := &pb.ClusterStatus{}
	returnValue := &pb.ClusterStatus{}

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}
	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* GET CLUSTER STATUS INFORMATION */
	returnValue.Clusteruuid = record.Clusteruuid
	var clusterStatus []byte
	var clusterStatusStruct clusterv1alpha.ClusterStatus
	err = dbconn.QueryRowContext(ctx, GetClusterStatusQuery, record.Clusteruuid).Scan(
		&returnValue.Name,
		&returnValue.State,
		&clusterStatus,
	)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterStatusQuery", friendlyMessage)
	}
	// Parse kubernetes status to get lastupdate, reason, and message
	err = json.Unmarshal(clusterStatus, &clusterStatusStruct)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"json.unmarshal.clusterStatus", friendlyMessage)
	}
	returnValue.Message = clusterStatusStruct.Message
	returnValue.Reason = clusterStatusStruct.Reason
	returnValue.Lastupdate = clusterStatusStruct.LastUpdate.String()
	if returnValue.Message != "" {
		parsedMessage, errorCode, err := utils.ParseOperatorMessage(ctx, []byte(returnValue.Message))
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"utils.ParseOperatorMessage", friendlyMessage)
		}
		returnValue.Message = parsedMessage
		returnValue.Errorcode = errorCode
	}

	return returnValue, nil
}

func GetNodeGroup(ctx context.Context, dbconn *sql.DB, record *pb.GetNodeGroupRequest) (*pb.NodeGroupResponseForm, error) {
	friendlyMessage := "Could not get Node Group. Please try again"
	failedFunction := "GetNodeGroup."
	returnError := &pb.NodeGroupResponseForm{}
	returnValue := &pb.NodeGroupResponseForm{
		Nodegroupuuid:    "",
		Clusteruuid:      record.Clusteruuid,
		Name:             "",
		Createddate:      "",
		Description:      nil,
		Instancetypeid:   "",
		Nodegroupstate:   "",
		Nodegroupstatus:  &pb.Nodegroupstatus{},
		Count:            0,
		Sshkeyname:       []*pb.SshKey{}, // In DBv8
		Imiid:            "",
		Upgradeimiid:     []string{}, // In DBv8
		Upgradeavailable: false,      // In DBv8
		Upgradestrategy: &pb.UpgradeStrategy{
			Drainnodes:               false,
			Maxunavailablepercentage: 0,
		},
		Networkinterfacename: "",
		Vnets:                []*pb.Vnet{},
		Nodes:                []*pb.NodeStatus{},
		Userdataurl:          nil,
		Nodegrouptype:        nil,
		Clustertype:          "",
	}
	/* VALIDATE CLUSTER AND NODEGROUP EXISTANCE */
	clusterId, nodeGroupId, err := utils.ValidateNodeGroupExistance(ctx, dbconn, record.Clusteruuid, record.Nodegroupuuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateNodeGroupExistance", friendlyMessage)
	}
	if clusterId == -1 || nodeGroupId == -1 {
		return returnError, status.Errorf(codes.NotFound, "NodeGroup not found in Cluster: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* GET NODE GROUP DETAILS */
	var (
		sshKeyNameJson  []byte
		clusterProvider string
		runtimeName     string
		k8sVersion      string
		nodegroupType   string
		vnet            []byte
	)
	err = dbconn.QueryRowContext(ctx, getNodeGroupDetailsQuery,
		record.Clusteruuid,
		record.Nodegroupuuid,
	).Scan(&nodeGroupId, &returnValue.Nodegroupuuid, &returnValue.Name, &returnValue.Description,
		&returnValue.Instancetypeid, &returnValue.Nodegroupstate,
		&returnValue.Count, &sshKeyNameJson, &returnValue.Imiid,
		&returnValue.Upgradestrategy.Drainnodes, &returnValue.Upgradestrategy.Maxunavailablepercentage,
		&nodegroupType, &k8sVersion, &runtimeName, &clusterProvider, &returnValue.Networkinterfacename, &vnet, &returnValue.Createddate,
		&returnValue.Userdataurl, &returnValue.Nodegrouptype, &returnValue.Clustertype,
	)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getNodeGroupDetailsQuery", friendlyMessage)
	}

	err = json.Unmarshal(sshKeyNameJson, &returnValue.Sshkeyname)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getNodeGroupDetailsQuery.Unmarshall.sshKeyNameJson", friendlyMessage)
	}

	/* GET VNETS */
	if len(vnet) != 0 {
		err = json.Unmarshal(vnet, &returnValue.Vnets)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getNodeGroupDetailsQuery.Unmarshall.vnets", friendlyMessage)
		}
	}

	/* GET CURRENT IMIS AVAILABLE TO UPGRADE*/
	_, availableVersions, err := utils.GetAvailableWorkerImiUpgrades(ctx, dbconn, record.Clusteruuid, record.Nodegroupuuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetAvailableWorkerImiUpgrades", friendlyMessage)
	}
	returnValue.Upgradeavailable = len(availableVersions) > 0
	returnValue.Upgradeimiid = availableVersions

	/* GET NODEGROUP STATUS */
	getNodeGroupStatusReq := &pb.NodeGroupid{
		Clusteruuid:    record.Clusteruuid,
		Nodegroupuuid:  record.Nodegroupuuid,
		CloudAccountId: record.CloudAccountId,
	}
	getNodegroupStatusResponse, err := GetNodeGroupStatusRecord(ctx, dbconn, getNodeGroupStatusReq)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetStatusRecord", friendlyMessage)
	}
	returnValue.Nodegroupstatus = getNodegroupStatusResponse

	summary := &pb.NodegroupSummary{}
	/* GET NODES IF ASKED*/
	if record.Nodes != nil && *record.Nodes {
		// Get Nodes of the nodegroup nodes
		rows, err := dbconn.QueryContext(ctx, getNodegroupNodesQuery, nodeGroupId)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getNodegroupNodesQuery", friendlyMessage)
		}
		defer rows.Close()
		for rows.Next() {
			// Scan into Node Status return
			node := &pb.NodeStatus{
				WekaStorage: &pb.WekaStorageStatus{},
			}
			var nodeStatus []byte
			var nodeStatusStruct clusterv1alpha.NodeStatus
			var wekaStorageClientId string
			var wekaStorageStatus string
			var wekaStorageCustomStatus string
			var wekaStorageMessage string

			if err = rows.Scan(
				&node.Name,
				&node.Ipaddress,
				&node.Dnsname,
				&node.Imi,
				&node.State,
				&nodeStatus,
				&node.Createddate,
				&wekaStorageClientId,
				&wekaStorageStatus,
				&wekaStorageCustomStatus,
				&wekaStorageMessage); err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getNodegroupNodesQuery.rows.scan", friendlyMessage)
			}

			node.WekaStorage.ClientId = wekaStorageClientId
			node.WekaStorage.Status = wekaStorageStatus
			node.WekaStorage.CustomStatus = wekaStorageCustomStatus
			node.WekaStorage.Message = wekaStorageMessage

			//Edit Status into json
			err = json.Unmarshal(nodeStatus, &nodeStatusStruct)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"json.unmarshal.nodeStatus", friendlyMessage)
			}
			node.Message = nodeStatusStruct.Message
			node.Reason = nodeStatusStruct.Reason
			node.Instanceimi = nodeStatusStruct.InstanceIMI
			node.Unschedulable = nodeStatusStruct.Unschedulable

			if node.Message != "" {
				parsedMessage, errorCode, err := utils.ParseOperatorMessage(ctx, []byte(node.Message))
				if err != nil {
					return returnError, utils.ErrorHandler(ctx, err, failedFunction+"utils.ParseOperatorMessage", friendlyMessage)
				}
				node.Message = parsedMessage
				node.Errorcode = errorCode
			}

			returnValue.Nodes = append(returnValue.Nodes, node)

			switch node.State {
			case "Active":
				summary.Activenodes++
			case "Creating", "Updating":
				summary.Provisioningnodes++
			case "Error":
				summary.Errornodes++
			case "Deleting":
				summary.Deletingnodes++
			default:
				summary.Errornodes++
			}
		}
	}
	returnValue.Nodegroupstatus.Nodegroupsummary = summary

	return returnValue, nil
}

func GetNodeGroups(ctx context.Context, dbconn *sql.DB, record *pb.GetNodeGroupsRequest) (*pb.NodeGroupResponse, error) {
	friendlyMessage := "Could not get Node Groups. Please try again"
	failedFunction := "GetNodeGroups."
	returnError := &pb.NodeGroupResponse{}
	returnValue := &pb.NodeGroupResponse{
		Nodegroups: []*pb.NodeGroupResponseForm{},
	}

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/*GET NODE GROUP UUIDS*/
	rows, err := dbconn.QueryContext(ctx, getAllClusterNodeGroupsUuidsQuery, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getAllClusterNodeGroupsUuidsQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		nodeGroupReq := &pb.GetNodeGroupRequest{
			Clusteruuid:    record.Clusteruuid,
			Nodegroupuuid:  "",
			Nodes:          record.Nodes,
			CloudAccountId: record.CloudAccountId,
		}
		err = rows.Scan(&nodeGroupReq.Nodegroupuuid)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getAllClusterNodeGroupsUuidsQuery.rows.scan", friendlyMessage)
		}
		nodeGroupResponse, err := GetNodeGroup(ctx, dbconn, nodeGroupReq)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetNodeGroup", friendlyMessage)
		}
		returnValue.Nodegroups = append(returnValue.Nodegroups, nodeGroupResponse)
	}
	return returnValue, nil
}

func GetNodeGroupStatusRecord(ctx context.Context, dbconn *sql.DB, record *pb.NodeGroupid) (*pb.Nodegroupstatus, error) {
	friendlyMessage := "Could not get Node Group Status. Please try again"
	failedFunction := "GetNodeGroupStatusRecord."
	returnError := &pb.Nodegroupstatus{}
	returnValue := &pb.Nodegroupstatus{}

	/* VALIDATE CLUSTER AND NODEGROUP EXISTANCE */
	clusterId, nodeGroupId, err := utils.ValidateNodeGroupExistance(ctx, dbconn, record.Clusteruuid, record.Nodegroupuuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateNodeGroupExistance", friendlyMessage)
	}
	if clusterId == -1 || nodeGroupId == -1 {
		return returnError, status.Errorf(codes.NotFound, "NodeGroup not found in Cluster: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* GET NODEGORUP STATUS INFORMATION*/
	returnValue.Nodegroupuuid = record.Nodegroupuuid
	returnValue.Clusteruuid = record.Clusteruuid
	var nodegroupStatus []byte
	var nodegroupStatusStruct clusterv1alpha.NodegroupStatus
	err = dbconn.QueryRowContext(ctx, GetNodeGroupStatusQuery, record.Nodegroupuuid).Scan(
		&returnValue.Name,
		&returnValue.Count,
		&returnValue.State,
		&nodegroupStatus,
	)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetNodeGroupStatusQuery", friendlyMessage)
	}
	// parse kubernetes status to get reason and message
	err = json.Unmarshal(nodegroupStatus, &nodegroupStatusStruct)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"json.Unmarshall.nodegroupStatus", friendlyMessage)
	}
	returnValue.Reason = nodegroupStatusStruct.Reason
	returnValue.Message = nodegroupStatusStruct.Message
	if returnValue.Message != "" {
		parsedMessage, errorCode, err := utils.ParseOperatorMessage(ctx, []byte(returnValue.Message))
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"utils.ParseOperatorMessage", friendlyMessage)
		}
		returnValue.Message = parsedMessage
		returnValue.Errorcode = errorCode
	}

	return returnValue, nil
}

func GetPublicAllK8sVersionsRecords(ctx context.Context, dbconn *sql.DB, record *pb.IksCloudAccountId) (*pb.GetPublicAllK8SversionResponse, error) {
	// Set up return values
	friendlyMessage := "Could not get available K8s Versions. Please try again"
	failedFunction := "GetPublicAllK8sVersionsRecords."
	returnError := &pb.GetPublicAllK8SversionResponse{}
	returnValue := &pb.GetPublicAllK8SversionResponse{
		K8Sversions: []*pb.GetPublicK8SversionResponse{},
	}
	var providerName string
	err := dbconn.QueryRowContext(ctx, GetCloudAccountProvider,
		record.CloudAccountId,
	).Scan(&providerName)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetCloudAccountProvider", friendlyMessage)
	}
	// Query to get K8s Versions
	rows, err := dbconn.QueryContext(ctx, getMetadataK8sVersionQuery, providerName)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getMetadataK8sVersionQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		// Scan into struct
		k8sVersion := &pb.GetPublicK8SversionResponse{
			K8Sversionname: "",
			Runtimename:    []string{},
		}
		var (
			k8sversion_name string
			compatRuntimes  string
		)
		err = rows.Scan(&k8sVersion.K8Sversionname, &k8sversion_name, &compatRuntimes)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getMetadataK8sVersionQuery.rows.scan", friendlyMessage)
		}
		if providerName != "" && providerName == "rke2" {
			k8sVersion.K8Sversionname = k8sversion_name
		}
		// convert json into array of strings
		err = json.Unmarshal([]byte(compatRuntimes), &k8sVersion.Runtimename)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.compatRuntimes", friendlyMessage)
		}
		// append to return value
		returnValue.K8Sversions = append(returnValue.K8Sversions, k8sVersion)
	}
	return returnValue, nil
}

func GetPublicAllRuntimesRecords(ctx context.Context, dbconn *sql.DB, record *pb.IksCloudAccountId) (*pb.GetPublicAllRuntimeResponse, error) {
	// Set up return values
	friendlyMessage := "Could not get available Runtime Versions. Please try again"
	failedFunction := "GetPublicAllRuntimesRecords."
	returnError := &pb.GetPublicAllRuntimeResponse{}
	returnValue := &pb.GetPublicAllRuntimeResponse{
		Runtimes: []*pb.GetPublicRuntimeResponse{},
	}
	var providerName string
	err := dbconn.QueryRowContext(ctx, GetCloudAccountProvider,
		record.CloudAccountId,
	).Scan(&providerName)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetCloudAccountProvider", friendlyMessage)
	}
	// Query to get runtimes
	rows, err := dbconn.QueryContext(ctx, getMetadataRuntimeQuery, providerName)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getMetadataRuntimeQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		// Scan into struct
		runtime := &pb.GetPublicRuntimeResponse{
			Runtimename:    "",
			K8Sversionname: []string{},
		}
		var compatK8s string
		var compatK8s2 string
		err = rows.Scan(&runtime.Runtimename, &compatK8s, &compatK8s2)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getMetadataRuntimeQuery.rows.scan", friendlyMessage)
		}
		// convert json into array of strings
		if providerName != "" && providerName == "rke2" {
			err = json.Unmarshal([]byte(compatK8s2), &runtime.K8Sversionname)
		} else {
			err = json.Unmarshal([]byte(compatK8s), &runtime.K8Sversionname)
		}
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.compatk8s."+providerName, friendlyMessage)
		}
		// append to return value
		returnValue.Runtimes = append(returnValue.Runtimes, runtime)
	}

	return returnValue, nil
}

func GetPublicAllInstancetypeRecords(ctx context.Context, dbconn *sql.DB, record *pb.IksCloudAccountId) (*pb.GetPublicAllInstancetypeResponse, error) {
	friendlyMessage := "Could not get Instance Types. Please try again"
	failedFunction := "GetPublicAllInstancetypeRecords."
	returnError := &pb.GetPublicAllInstancetypeResponse{}
	returnValue := &pb.GetPublicAllInstancetypeResponse{
		Instancetypes: []*pb.GetPublicInstancetypeResponse{},
	}
	// Query to get instance types  ??? SUBJECT TO CHANGE BASED ON UPSTREAM
	rows, err := dbconn.QueryContext(ctx, getMetadataInstanceTypeQuery)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getMetadataInstanceTypeQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		// Scan into struct
		instanceType := &pb.GetPublicInstancetypeResponse{
			Instancetypename: "",
			Displayname:      "",
			Description:      "",
			Instancecategory: "",
			Memory:           0,
			Cpu:              0,
			Storage:          0,
		}
		err = rows.Scan(&instanceType.Instancetypename, &instanceType.Memory, &instanceType.Cpu, &instanceType.Storage,
			&instanceType.Displayname, &instanceType.Description, &instanceType.Instancecategory,
		)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getMetadataInstanceTypeQuery.rows.scan", friendlyMessage)
		}
		// append to return value
		returnValue.Instancetypes = append(returnValue.Instancetypes, instanceType)
	}

	return returnValue, nil
}

func GetKubeConfig(ctx context.Context, dbconn *sql.DB, record *pb.GetKubeconfigRequest, filePath string) (*pb.GetKubeconfigResponse, error) {
	friendlyMessage := "Could not get Cluster Kube Config. Please try again"
	failedFunction := "GetKubeConfig."
	returnError := &pb.GetKubeconfigResponse{}

	readonly := pointer.BoolDeref(record.Readonly, false)

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getMetadataInstanceTypeQuery", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* GET DEFAULT EXPIRATION FROM DB */
	var expirationDays int
	err = dbconn.QueryRowContext(ctx, GetDefaultConfigQuery, "kubecfg_cert_expiration_days").Scan(&expirationDays)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultConfigQuery.experation_days", friendlyMessage)
	}

	/* VALIDATE CREATED TIME */
	var kubeconfig string

	var getKubeconfigQuery string
	if readonly {
		getKubeconfigQuery = GetReadOnlyKubeconfigQuery
	} else {
		getKubeconfigQuery = GetAdminKubeconfigQuery
	}

	err = dbconn.QueryRowContext(ctx, getKubeconfigQuery, record.Clusteruuid).Scan(&kubeconfig)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetAdminKubeconfigQuery", friendlyMessage)
	}

	// Check if kubeconfig is about to expire, if not return kubeconfig
	if kubeconfig != "" {
		aboutToExpire, err := isKubeconfigAboutToExpire(ctx, dbconn, kubeconfig, expirationDays)
		if err != nil {
			return returnError, err
		}
		if !aboutToExpire {
			return &pb.GetKubeconfigResponse{
				Clusterid:  clusterId,
				Kubeconfig: kubeconfig,
			}, nil
		}
	}

	/* GET API SERVER FROM VIP STATES */
	publicApiserverLB, publicApiserverLBPort, err := getApiServerDetails(ctx, dbconn, clusterId, record)
	if err != nil {
		return returnError, err
	}

	// get decoded ca keys and certs required to sign cert to create kubeconfig
	caCrtByte, caKeyByte, encodedCaCrtByte, err := getDecodesCerts(ctx, dbconn, clusterId, filePath)
	if err != nil {
		return returnError, err
	}

	caCrtx509, err := kubeUtils.ParseCert(caCrtByte)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, "kubeUtils.ParseAdminCertificate.caCrt", "Could not parse CA certificate")
	}
	caKeyx509, err := kubeUtils.ParsePrivateKey(caKeyByte)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, "kubeUtils.ParsePKCS1PrivateKey", "Could not parse CA key")
	}

	var certConfig kubeUtils.CertConfig
	var contextUser string
	if readonly {
		certConfig = kubeUtils.CertConfig{
			CommonName: "readonly",
			Organizations: []string{
				"system:users",
			},
		}
		contextUser = "default-readonly"
	} else {
		certConfig = kubeUtils.CertConfig{
			CommonName: "admin",
			Organizations: []string{
				"system:masters",
			},
		}
		contextUser = "default-admin"
	}

	/* CREATE AND SIGN CERTS */
	expirationDate := time.Now().AddDate(0, 0, expirationDays)
	certPEM, privateKeyPEM, err := kubeUtils.CreateAndSignCert(caCrtx509, caKeyx509, certConfig, &expirationDate)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"kubeUtils.CreateAndSignCert", friendlyMessage)
	}

	/* CREATE NEW KUBECONFIG */
	resultKubeconfig := &bytes.Buffer{}
	if err := iksOperator.KubeconfigTemplate.Execute(resultKubeconfig, &iksOperator.KubeconfigTemplateConfig{
		ClusterName: record.Clusteruuid,
		Server:      publicApiserverLB,
		Port:        publicApiserverLBPort,
		User:        contextUser,
		Cert:        base64.StdEncoding.EncodeToString(certPEM),
		Key:         base64.StdEncoding.EncodeToString(privateKeyPEM),
		Ca:          string(encodedCaCrtByte),
	}); err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"kubeconfigTemplate.Execute", friendlyMessage)
	}

	var updateClusterKubeconfigQuery string
	if readonly {
		updateClusterKubeconfigQuery = UpdateClusterReadonlyKubeconfig
	} else {
		updateClusterKubeconfigQuery = UpdateClusterAdminKubeconfig
	}

	/* INSERT KUBECONFIG TO CLUSTER */
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage)
	}
	_, err = tx.ExecContext(ctx, updateClusterKubeconfigQuery, clusterId, resultKubeconfig.String())
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterKubeconfig", friendlyMessage)
	}
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}

	return &pb.GetKubeconfigResponse{
		Clusterid:  clusterId,
		Kubeconfig: resultKubeconfig.String(),
	}, nil
}

func GetVips(ctx context.Context, dbconn *sql.DB, record *pb.ClusterID) (*pb.GetVipsResponse, error) {

	friendlyMessage := "Could not get VIPs. Please try again"
	failedFunction := "GetVips."
	returnError := &pb.GetVipsResponse{}
	returnValue := &pb.GetVipsResponse{
		Response: []*pb.GetVipResponse{},
	}
	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* GET VIP IDs */
	rows, err := dbconn.QueryContext(ctx, GetVipIdsQuery, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetVipIdsQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		vipId := &pb.VipId{
			Clusteruuid:    record.Clusteruuid,
			Vipid:          0,
			CloudAccountId: record.CloudAccountId,
		}
		err = rows.Scan(&vipId.Vipid)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetVipIdsQuery.rows.scan", friendlyMessage)
		}
		vipresponse, err := GetVip(ctx, dbconn, vipId)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetVip", friendlyMessage)
		}
		returnValue.Response = append(returnValue.Response, vipresponse)
	}
	return returnValue, nil
}

func GetVip(ctx context.Context, dbconn *sql.DB, record *pb.VipId) (*pb.GetVipResponse, error) {
	friendlyMessage := "Could not get VIP. Please try again"
	failedFunction := "GetVip."

	returnError := &pb.GetVipResponse{}
	returnValue := &pb.GetVipResponse{
		Vipid:       0,
		Name:        "",
		Vipstate:    "",
		VipIp:       nil,
		Port:        0,
		Poolport:    0,
		Viptype:     "",
		Createddate: "",
		Dnsalias:    []string{},
		Members:     []*pb.Members{},
		Vipstatus:   &pb.VipStatus{},
	}

	/* VALIDATE CLUSTER AND VIP EXISTANCE */
	clusterId, vipId, err := utils.ValidateVipExistance(ctx, dbconn, record.Clusteruuid, record.Vipid, true)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateVipExistance", friendlyMessage)
	}
	if clusterId == -1 || vipId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Vip not found in Cluster: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* GET VIPS */
	var vipStatus []byte
	rows, err := dbconn.QueryContext(ctx, GetVipsQuery, clusterId, record.Vipid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetVipsQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		var dnsalias string
		err = rows.Scan(&returnValue.Vipid, &returnValue.Name, &returnValue.Description, &returnValue.Vipstate,
			&returnValue.VipIp, &returnValue.Port, &returnValue.Poolport,
			&returnValue.Viptype, &dnsalias, &vipStatus, &returnValue.Createddate,
		)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetVipsQuery.rows.query", friendlyMessage)
		}

		returnValue.Dnsalias = append(returnValue.Dnsalias, dnsalias)
	}

	/* GET VIP MEMBERS*/
	rows, err = dbconn.QueryContext(ctx, GetVipMembersQuery, record.Vipid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetVipsMembersQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		var ipaddress string
		member := &pb.Members{
			Ipaddresses: []string{},
		}
		err = rows.Scan(&ipaddress)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetVipsMembersQuery.rows.scan", friendlyMessage)
		}

		member.Ipaddresses = append(member.Ipaddresses, ipaddress)
		returnValue.Members = append(returnValue.Members, member)
	}

	/* PARSE VIP STATUS */
	var vipStatusStruct ilbv1alpha1.IlbStatus
	err = json.Unmarshal(vipStatus, &vipStatusStruct)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"json.Unmarshall.vipStatus", friendlyMessage)
	}
	returnValue.Vipstatus.Name = vipStatusStruct.Name
	returnValue.Vipstatus.Vipstate = string(vipStatusStruct.State)
	returnValue.Vipstatus.Poolid = int32(vipStatusStruct.PoolID)
	returnValue.Vipstatus.Vipid = strconv.Itoa(vipStatusStruct.VipID)
	returnValue.Vipstatus.Message = vipStatusStruct.Message
	if returnValue.Vipstatus.Message != "" {
		parsedMessage, errorCode, err := utils.ParseOperatorMessage(ctx, []byte(returnValue.Vipstatus.Message))
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"utils.ParseOperatorMessage", friendlyMessage)
		}
		returnValue.Vipstatus.Message = parsedMessage
		returnValue.Vipstatus.Errorcode = errorCode
	}

	return returnValue, nil
}

func GetFirewallRule(ctx context.Context, dbconn *sql.DB, record *pb.ClusterID) (*pb.GetFirewallRuleResponse, error) {
	friendlyMessage := "Could not get Security Rules. Please try again"
	failedFunction := "GetSecurityRule."

	returnError := &pb.GetFirewallRuleResponse{}
	returnValue := &pb.GetFirewallRuleResponse{
		Getfirewallresponse: []*pb.FirewallRuleResponse{},
	}
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	var vip_id int32

	// Get vip id from clusteruuid
	rows, err := dbconn.QueryContext(ctx, GetVipidByClusterQuery, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"Getvipidbycluster", friendlyMessage)
	}

	defer rows.Close()

	for rows.Next() {
		var sourceips []byte
		var protocol []byte
		var sourceipsresult []string
		var protocolresult []string

		fwresponse := &pb.FirewallRuleResponse{
			Destinationip: "",
			State:         "Not Specified",
			Sourceip:      []string{},
			Vipid:         0,
			Port:          0,
			Vipname:       "",
			Viptype:       "",
			Protocol:      []string{},
			Internalport:  0,
		}
		err = rows.Scan(&vip_id)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"scanvipidsfromClusterid.rows.scan", friendlyMessage)
		}
		err = dbconn.QueryRowContext(ctx, GetPublicVipsForFirewallQuery, clusterId, vip_id).Scan(&fwresponse.Vipid, &fwresponse.Destinationip, &fwresponse.Port, &fwresponse.Internalport, &sourceips, &fwresponse.State, &fwresponse.Viptype, &protocol, &fwresponse.Vipname)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetPublicVipsForFirewallQuery", friendlyMessage)
		}
		if sourceips != nil {
			err = json.Unmarshal(sourceips, &sourceipsresult)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetMetadata.unmarshall.souceip", friendlyMessage)
			}
			fwresponse.Sourceip = sourceipsresult
		} else {
			fwresponse.Sourceip = []string{}
		}
		if protocol != nil {
			err = json.Unmarshal(protocol, &protocolresult)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetMetadata.unmarshall.protocol", friendlyMessage)
			}
			fwresponse.Protocol = protocolresult
		}
		returnValue.Getfirewallresponse = append(returnValue.Getfirewallresponse, fwresponse)
	}
	return returnValue, nil
}

// isKubeconfigAboutToExpire checks if the kubeconfig is about to expire
func isKubeconfigAboutToExpire(ctx context.Context, dbconn *sql.DB, kubeconfig string, expirationDays int) (bool, error) {
	var kubeConfStruct *iksOperator.ClusterTemplateStruct
	err := yaml.Unmarshal([]byte(kubeconfig), &kubeConfStruct)
	if err != nil {
		return false, utils.ErrorHandler(ctx, err, "isKubeconfigAboutToExpire.yaml.Unmarshal", "Could not unmarshal kubeconfig")
	}
	if len(kubeConfStruct.Users) < 1 {
		return false, utils.ErrorHandler(ctx, err, "isKubeconfigAboutToExpire.kubeConfStruct.Users", "No users found in kubeconfig")
	}
	crt := kubeConfStruct.Users[0].User.ClientCertificateData
	crtByte, err := utils.Base64DecodeString(ctx, crt)
	if err != nil {
		return false, utils.ErrorHandler(ctx, err, "isKubeconfigAboutToExpire.Base64DecodeString", "Could not decode certificate")
	}
	caCrtx509, err := kubeUtils.ParseCert([]byte(crtByte))
	if err != nil {
		return false, utils.ErrorHandler(ctx, err, "isKubeconfigAboutToExpire.ParseCert", "Could not parse certificate")
	}

	if expirationDays < 1 {
		expirationDays = 1
	}
	today := time.Now()
	timeDiff := caCrtx509.NotAfter.Sub(today).Seconds()

	// time to expiration is less than configured expiration check
	return int(timeDiff) <= int((time.Hour).Seconds())*(expirationDays*24), nil
}

// getDecodesCerts gets ca certs and key from db, decrypts and decodes the CA certificate and key, returning the decoded bytes for caCert, caKey, and encodedCaCrt
func getDecodesCerts(ctx context.Context, dbconn *sql.DB, clusterId int32, filePath string) ([]byte, []byte, []byte, error) {
	var caCrt, caKey, nonce string
	var encryptionKeyId int32
	err := dbconn.QueryRowContext(ctx, GetClusterSecrets, clusterId).Scan(&caCrt, &caKey, &encryptionKeyId, &nonce)
	if err != nil {
		return nil, nil, nil, utils.ErrorHandler(ctx, err, "getDecodesCerts.GetClusterSecrets", "Could not get cluster secrets")
	}
	decodedNonce, err := utils.Base64DecodeString(ctx, nonce)
	if err != nil {
		return nil, nil, nil, utils.ErrorHandler(ctx, err, "getDecodesCerts.DecodeNonce", "Could not decode nonce")
	}

	encryptionKeyBytes, err := utils.GetSpecificEncryptionKey(ctx, filePath, encryptionKeyId)
	if err != nil {
		return nil, nil, nil, utils.ErrorHandler(ctx, err, "getDecodesCerts.utils.GetSpecificEncryptionKey", "Could not get specific encryption key")
	}
	encodedCaCrtByte, err := utils.AesDecryptSecret(ctx, caCrt, encryptionKeyBytes, decodedNonce)
	if err != nil {
		return nil, nil, nil, utils.ErrorHandler(ctx, err, "getDecodesCerts.utils.AesDecryptSecret.caCrt", "Could not decrypt CA certificate")
	}
	encodedCaKeyByte, err := utils.AesDecryptSecret(ctx, caKey, encryptionKeyBytes, decodedNonce)
	if err != nil {
		return nil, nil, nil, utils.ErrorHandler(ctx, err, "getDecodesCerts.utils.AesDecryptSecret.caKey", "Could not decrypt CA key")
	}
	caCrtByte, err := utils.Base64DecodeString(ctx, encodedCaCrtByte)
	if err != nil {
		return nil, nil, nil, utils.ErrorHandler(ctx, err, "getDecodesCerts.utils.Base64DecodeString.caCrt", "Could not decode CA certificate")
	}
	caKeyByte, err := utils.Base64DecodeString(ctx, encodedCaKeyByte)
	if err != nil {
		return nil, nil, nil, utils.ErrorHandler(ctx, err, "getDecodesCerts.utils.Base64DecodeString.caKey", "Could not decode CA key")
	}
	return caCrtByte, caKeyByte, []byte(encodedCaCrtByte), nil
}

// getApiServerDetails returns kubernetes-api server  public IP and port
func getApiServerDetails(ctx context.Context, dbconn *sql.DB, clusterId int32, record *pb.GetKubeconfigRequest) (string, string, error) {
	var publicApiserverLB string
	var vipStatus []byte
	err := dbconn.QueryRowContext(ctx, GetVipsStatusByClusterQuery, clusterId).Scan(&vipStatus)
	if err != nil {
		return "", "", utils.ErrorHandler(ctx, err, "getApiServerDetails.getNodegroupNodesQuery", "Could not get VIP status")
	}
	var vipStatusJson ilbv1alpha1.IlbStatus
	err = json.Unmarshal(vipStatus, &vipStatusJson)
	if err != nil {
		return "", "", utils.ErrorHandler(ctx, err, "getApiServerDetails.json.Unmarshal.vipStatus", "Could not unmarshal VIP status")
	}
	if vipStatusJson.Vip == "" {
		return "", "", status.Error(codes.NotFound, "Kubeconfig not ready. Please try again later.")
	}
	publicApiserverLB = vipStatusJson.Vip

	if publicApiserverLB == "" {
		return "", "", utils.ErrorHandler(ctx, err, "getApiServerDetails.find.publicApiserverLBPort", "Could not find public API server LB port")
	}

	_, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return "", "", utils.ErrorHandler(ctx, err, "getApiServerDetails.utils.GetLatestClusterRev", "Could not get latest cluster revision")
	}
	var publicApiserverLBPort string
	for _, ilb := range clusterCrd.Spec.ILBS {
		if ilb.Name == "public-apiserver" {
			publicApiserverLBPort = strconv.Itoa(ilb.Port)
			break
		}
	}
	return publicApiserverLB, publicApiserverLBPort, nil
}
