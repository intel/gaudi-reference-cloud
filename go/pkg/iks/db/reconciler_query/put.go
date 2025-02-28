// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package reconciler_query

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"strconv"
	"strings"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	fwv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	query "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/query"
	ilbv1alpha "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ilb_operator/api/v1alpha1"
	clusterv1alpha "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (

	// Select
	GetClusterStorages = `
		SELECT storage_id, storagestate_name, storageprovider_name
		FROM storage
		WHERE cluster_id = $1
	`
	GetEncryptionKeyAndNonce = `
		SELECT encryptionkey_id, nonce
		FROM cluster_extraconfig
		WHERE cluster_id = (
			SELECT cluster_id FROM public.cluster WHERE unique_id = $1
		)
	`
	GetClusterRevChangeAppliedQuery = `
		SELECT change_applied
		FROM public.clusterrev cr
			inner join public.cluster c 
			on c.cluster_id = cr.cluster_id 
		WHERE c.unique_id  = $1
		ORDER BY clusterrev_id desc
		LIMIT 1
	`
	GetClusterState = `
		SELECT clusterstate_name FROM public.cluster WHERE cluster_id = $1;
	`
	CheckUuid                = `SELECT cluster_id FROM cluster WHERE unique_id = $1`
	CheckUuidAndClusterrevId = `SELECT cluster_id FROM clusterrev WHERE clusterrev_id = $1`
	CheckClusterExtraConfig  = `
		SELECT count(*)
		FROM cluster_extraconfig
		WHERE cluster_id = $1
	`

	GetNodeGroupId = `
		SELECT nodegroup_id
		FROM nodegroup 
		WHERE unique_id = $1
	`
	GetNode = `
		SELECT 
			nodegroup.nodegroup_id,
			nodegroup.k8sversion_name,
			k8snode.k8snode_name,
			k8snode.ip_address,
			COALESCE(k8snode.dns_name, ''),
			COALESCE(k8snode.osimageinstance_name, ''),
			k8snode.k8snodestate_name,
			k8snode.created_date
		FROM k8snode
		INNER JOIN nodegroup ON k8snode.cluster_id = nodegroup.cluster_id AND k8snode.nodegroup_id = nodegroup.nodegroup_id
		WHERE k8snode.cluster_id = $1
	`
	GetOsImageInstanceFromImiArtifact = `
		SELECT o.osimageinstance_name
		FROM osimageinstance o
		WHERE imiartifact = $1 
	`
	GetNodesbyNodegroup = `
		SELECT k8snode.k8snode_name, k8snode.ip_address, nodegroup.osimageinstance_name, 
		nodegroup.k8sversion_name, k8snode.k8snodestate_name, k8snode.created_date 
		FROM k8snode
		INNER JOIN nodegroup ON k8snode.cluster_id = nodegroup.cluster_id AND k8snode.nodegroup_id = nodegroup.nodegroup_id
		WHERE nodegroup.unique_id = $1 AND nodegroup.cluster_id = $2
	`
	GetNodegroupProvider = `
		SELECT i.nodeprovider_name 
		FROM instancetype i
		WHERE i.instancetype_name = (
			SELECT instancetype_name
			FROM nodegroup n
			WHERE n.nodegroup_id = $1
		)
	`
	GetNodeGroups = `
		SELECT nodegroup.unique_id, nodegroup.nodecount, nodegroup.nodegrouptype_name,
		nodegroup.nodegroupstate_name FROM nodegroup 
   		WHERE cluster_id = $1
		`
	GetClusterAddonVersions = `
		SELECT addonversion_name, lastchangetimestamp, clusteraddonstate_name 
		FROM clusteraddonversion WHERE cluster_id = $1
	`
	GetClusterILBVersions = `
		SELECT vip.vip_id, vip.vipstate_name, vipdetails.vip_name 
		FROM vip
		INNER JOIN vipdetails ON vip.vip_id = vipdetails.vip_id
		WHERE cluster_id = $1
	`
	GetClusterILBIdName = `
		SELECT vip.vip_id, vipdetails.vip_name 
        FROM vip 
        INNER JOIN vipdetails ON vip.vip_id = vipdetails.vip_id
        WHERE vip.cluster_id = $1
	`
	//Update
	PutClusterState = `
		UPDATE cluster SET clusterstate_name = $1  WHERE cluster_id = $2;
	`
	PutClusterrevChangeApplied = `
		UPDATE clusterrev SET change_applied = $1  WHERE clusterrev_id = $2;
	`
	PutClusterStatus = `
		UPDATE cluster SET kubernetes_status = $1, 
		clusterstate_name = $2 WHERE cluster_id = $3
	`
	PutNodeGroupStatus = `
		UPDATE nodegroup
		SET kubernetes_status = $1, nodegroupstate_name = $2
		WHERE unique_id = $3 AND cluster_id = $4
	`
	PutStorageStatus = `
		UPDATE storage SET kubernetes_status = $3, storagestate_name= $4
		WHERE cluster_id = $1 AND storageprovider_name=$2
	`
	PutStorageState = `
		UPDATE storage SET storagestate_name= $3
		WHERE cluster_id = $1 AND storageprovider_name=$2
	`
	PutClusterAddonVersion = `
		UPDATE clusteraddonversion SET kubernetes_status = $1, clusteraddonstate_name = $2
	 	WHERE cluster_id = $3 AND addonversion_name =$4
	 `
	PutClusterVIPStatus = `
		UPDATE vip SET vip_status = $1, vipstate_name = $2, vip_ip = $3
	 	WHERE vip_id = $4 AND cluster_id = $5
	 `
	PutClusterVIPStatusNoVip = `
		UPDATE vip SET vip_status = $1, vipstate_name = $2
	 	WHERE vip_id = $3 AND cluster_id = $4
	 `
	PutClusterVIPDetailsStatus = `
		UPDATE vipdetails SET pool_id = $1
	 	WHERE vip_name = $2 AND vip_id = $3
	 `
	PutNodeStatus = `
		UPDATE k8snode SET 
		kubernetes_status = $1, 
		nodeprovider_name = $2, 
		k8snodestate_name = $3, 
		created_date = $4,
		weka_storage_client_id = $7,
		weka_storage_status = $8,
		weka_storage_custom_status = $9,
		weka_storage_message = $10 
		WHERE cluster_id = $5 AND k8snode_name = $6
	`
	PutFwStatus = `
		UPDATE vip SET firewall_status = $3 where cluster_id = $1 and vip_ip = $2
	`

	PutClusterCerts = `
		UPDATE cluster_extraconfig
		SET cluster_cacrt = $2,
			cluster_cakey = $3,
			cluster_etcd_cacrt = $4,
			cluster_etcd_cakey = $5,
			cluster_etcd_rotation_keys = $6,
			cluster_sa_pub = $7,
			cluster_sa_key = $8,
			cluster_cp_reg_cmd = $9,
			cluster_wk_reg_cmd = $10
		WHERE cluster_id = $1;
	`

	//Insert
	Insertk8sNode = `
		INSERT INTO k8snode (
			k8snode_name,
			cluster_id,
			k8snodestate_name,
			nodegroup_id,
			idc_instance_id,
			ip_address,
			nodeprovider_name,
			kubernetes_status,
			created_date,
			dns_name,
			osimageinstance_name,
			weka_storage_client_id,
			weka_storage_status,
			weka_storage_custom_status,
			weka_storage_message)
		VALUES ($1,$2,$3,$4,$5,$6,$7, $8, $9, $10, $11, $12, $13, $14, $15);
	`

	// Delete
	Deletek8sNode = `
		DELETE FROM k8snode
		WHERE  ip_address = $1;
	`
	Deletek8sNodegroup = `
		DELETE FROM nodegroup
		WHERE  unique_id = $1;
	`
	DeleteStorageStatus = `
		DELETE FROM storage
		WHERE  cluster_id=$1 AND storageprovider_name=$2;
	`
	Deletek8sAddons = `
		DELETE FROM clusteraddonversion
		WHERE  addonversion_name = $1;
	`
	Deletek8sILB = `
		DELETE FROM vip
		WHERE  vip_id = $1;
	`
	Deletek8sILBDetails = `
		DELETE FROM vipdetails
		WHERE  vip_id = $1;
	`
	DeleteFWRule = `
	update vip set sourceips = '[]'::JSONB where vip_ip = $1
	`

	DeleteProtocol = `
	update vipdetails set protocol = '[]'::JSONB where vip_id = (select vip_id from vip where vip_ip = $1 AND firewall_status = 'Deleting')
	`

	GetClusterFirewall = `
	SELECT COALESCE(v.firewall_status,'Not Specified'), COALESCE(v.vip_ip,'0.0.0.0'), v.sourceips, v.vip_id, d.port, COALESCE(d.protocol,'["TCP"]'::json) 
	FROM vip v INNER JOIN vipdetails d ON v.vip_id = d.vip_id
	WHERE v.cluster_id = $1 AND v.viptype_name = 'public' AND v.firewall_status != 'Pending'
`

	GetFirewallRuleStateClusterQuery = ` 
	SELECT COALESCE(firewall_status, 'Not Specified') FROM vip WHERE cluster_id = $1 AND vip_ip = $2
	`
)

type NodeStatusID struct {
	JsonStatus *clusterv1alpha.NodeStatus
	NodegrouID int32
}
type ILBNameID struct {
	VipID   int32
	VipName string
}

const (
	nameLength = 8
)

// Utils
var emptyReturn = &empty.Empty{}

func getDeletedNodes(nodesInStatus []clusterv1alpha.NodeStatus, nodesInDb []clusterv1alpha.NodeStatus) []clusterv1alpha.NodeStatus {
	var deletedNodes []clusterv1alpha.NodeStatus
	var nodesInStatusMap = make(map[string]bool)

	for _, n := range nodesInStatus {
		nodesInStatusMap[n.IpAddress] = true
	}

	for _, n := range nodesInDb {
		if _, found := nodesInStatusMap[n.IpAddress]; !found {
			deletedNodes = append(deletedNodes, n)
		}
	}

	return deletedNodes
}

func getNewAndExistingNodes(nodesInStatus []clusterv1alpha.NodeStatus, nodesInDb []clusterv1alpha.NodeStatus) ([]clusterv1alpha.NodeStatus, []clusterv1alpha.NodeStatus) {
	var existingNodes []clusterv1alpha.NodeStatus
	var newNodes []clusterv1alpha.NodeStatus
	var nodesInDBMap = make(map[string]bool)

	for _, n := range nodesInDb {
		nodesInDBMap[n.IpAddress] = true
	}

	// The node that comes from status should have the updated information
	// so we append that one, so that it can be updated in DB.
	for _, n := range nodesInStatus {
		_, found := nodesInDBMap[n.IpAddress]

		if found {
			existingNodes = append(existingNodes, n)
		} else {
			newNodes = append(newNodes, n)
		}
	}

	return newNodes, existingNodes
}

func ExtractSuffix(fullString, prefix string) (suffix string, err error) {
	if strings.HasPrefix(fullString, prefix) {
		suffix = strings.TrimPrefix(fullString, prefix)
		if strings.HasPrefix(suffix, "-") {
			suffix = strings.TrimPrefix(suffix, "-")
		}
		return suffix, nil
	}
	return "", errors.New("Prefix not found in the input string")
}

func findVipID(vipList []*ILBNameID, targetName string) (int32, error) {
	for _, vip := range vipList {
		if vip == nil {
			return -1, nil
		}
		if vip.VipName == targetName { // vip name from list == vip name from above
			return vip.VipID, nil // Vip id might be null
		}
	}
	return 0, errors.New("VipName " + targetName + " not found")
}

func differencesStorages(list1 []*clusterv1alpha.StorageStatus, list2 []*clusterv1alpha.StorageStatus) []*clusterv1alpha.StorageStatus {
	var difference []*clusterv1alpha.StorageStatus

	set := make(map[string]bool)
	for _, item := range list1 {
		set[item.Provider] = true
	}

	for _, item := range list2 {
		if _, exists := set[item.Provider]; !exists {
			difference = append(difference, item)
		}
	}

	return difference
}

func differencesNodeGroups(list1 []*clusterv1alpha.NodegroupStatus, list2 []*clusterv1alpha.NodegroupStatus) []*clusterv1alpha.NodegroupStatus {
	var difference []*clusterv1alpha.NodegroupStatus

	set := make(map[string]bool)
	for _, item := range list1 {
		set[item.Name] = true
	}

	for _, item := range list2 {
		if _, exists := set[item.Name]; !exists {
			difference = append(difference, item)
		}
	}

	return difference
}

func differencesAddons(list1 []*clusterv1alpha.AddonStatus, list2 []*clusterv1alpha.AddonStatus) []*clusterv1alpha.AddonStatus {
	var difference []*clusterv1alpha.AddonStatus

	set := make(map[string]bool)
	for _, item := range list1 {
		set[item.Name] = true
	}

	for _, item := range list2 {
		if _, exists := set[item.Name]; !exists {
			difference = append(difference, item)
		}
	}

	return difference
}
func differencesILB(list1 []*ilbv1alpha.IlbStatus, list2 []*ilbv1alpha.IlbStatus) []*ilbv1alpha.IlbStatus {
	var difference []*ilbv1alpha.IlbStatus

	set := make(map[string]bool)
	for _, item := range list1 {
		set[strconv.Itoa(item.VipID)] = true
	}

	for _, item := range list2 {
		if _, exists := set[strconv.Itoa(item.VipID)]; !exists {
			difference = append(difference, item)
		}
	}

	return difference
}

func differenceFW(list1 []*clusterv1alpha.FirewallStatus, list2 []*clusterv1alpha.FirewallStatus) []*clusterv1alpha.FirewallStatus {
	var difference []*clusterv1alpha.FirewallStatus

	set := make(map[string]bool)
	for _, item := range list1 {
		set[item.DestinationIp] = true
	}

	for _, item := range list2 {
		if _, exists := set[item.DestinationIp]; !exists {
			difference = append(difference, item)
		}
	}

	return difference
}

func transformToNodeStatus(nodeList []*pb.NodeStatusRequest) []clusterv1alpha.NodeStatus {
	var k8snodes []clusterv1alpha.NodeStatus
	for _, node := range nodeList {
		var transformNode clusterv1alpha.NodeStatus
		nodejson, err := json.Marshal(node)
		if err != nil {
			return nil
		}
		err = json.Unmarshal(nodejson, &transformNode)
		if err != nil {
			return nil
		}
		k8snodes = append(k8snodes, transformNode)
	}
	return k8snodes
}

func getDBNodesbyGroup(nodes []*NodeStatusID, nodeGroupId int32) []clusterv1alpha.NodeStatus {
	var k8snode []clusterv1alpha.NodeStatus
	for _, node := range nodes {
		if nodeGroupId == node.NodegrouID {
			k8snode = append(k8snode, *node.JsonStatus)
		}
	}
	return k8snode
}

// APIS
func PutClusterStateReconcilerQuery(ctx context.Context, dbconn *sql.DB, req *pb.UpdateClusterStateRequest) (*empty.Empty, error) {
	friendlyMessage := "PutClusterStateReconciler.UnexpectedError"
	failedFunction := "PutClusterStateReconcilerQuery."

	//Check if cluster exists
	var clusterId int32
	err := dbconn.QueryRowContext(ctx, CheckUuid, req.Uuid).Scan(&clusterId)
	if err != nil {
		if err == sql.ErrNoRows {
			return emptyReturn, errors.New("No cluster found")
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"CheckUuid", friendlyMessage+err.Error())
	}

	// Begin transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	// Update Cluster State
	_, err = tx.QueryContext(ctx, PutClusterState, req.State, clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	// Commit
	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	return emptyReturn, nil
}

func PutClusterChangeAppliedReconcilerQuery(ctx context.Context, dbconn *sql.DB, req *pb.UpdateClusterChangeAppliedRequest) (*empty.Empty, error) {
	friendlyMessage := "PutClusterStateReconciler.UnexpectedError"
	failedFunction := "PutClusterStateReconcilerQuery."

	//Check if cluster id and clusterev id exists
	var clusterId int32
	err := dbconn.QueryRowContext(ctx, CheckUuidAndClusterrevId, req.ClusterrevId).Scan(&clusterId)
	if err != nil {
		if err == sql.ErrNoRows {
			return emptyReturn, errors.New("No match between cluster and clusterrev, or clusterrev doesn't exist")
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"CheckUuidAndClusterrevId", friendlyMessage+err.Error())
	}

	// Begin Transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.ErrorHandler", friendlyMessage+err.Error())
	}

	// Update Change Apllied
	_, err = tx.QueryContext(ctx, PutClusterrevChangeApplied, req.ChangeApplied, req.ClusterrevId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"PutClusterrevChangeApplied", friendlyMessage+err.Error())
	}

	// Commit
	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}
	return emptyReturn, nil
}

func PutClusterCertsReconcilerQuery(ctx context.Context, dbconn *sql.DB, req *pb.UpdateClusterCertsRequest, filePath string) (*empty.Empty, error) {
	friendlyMessage := "PutClusterCertsQuery.UnexpectedError"
	failedFunction := "PutClusterCertsQuery."
	emptyReturn := &empty.Empty{}

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	err := dbconn.QueryRowContext(ctx, CheckUuid, req.Uuid).Scan(&clusterId)
	if err != nil {
		if err == sql.ErrNoRows {
			return emptyReturn, errors.New("No cluster found")
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"CheckUuid", friendlyMessage+err.Error())
	}

	/* GET NONCE AND ENCRYPTION KEY ID FROM THE DB */
	var nonce string
	var encryptionKeyId int32
	err = dbconn.QueryRowContext(ctx, GetEncryptionKeyAndNonce, req.Uuid).Scan(&encryptionKeyId, &nonce)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetEncryptionKeyAndNonce", friendlyMessage+err.Error())
	}
	decodedNonce, err := utils.Base64DecodeString(ctx, nonce)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetEncryptionKeyAndNonce.decode", friendlyMessage+err.Error())
	}

	// GET SPECIFIC ENCRYPTION KEY FROM VAULT
	encryptionKeyByte, err := utils.GetSpecificEncryptionKey(ctx, filePath, encryptionKeyId)

	// ENCRYPT INCOMING DATA (Incoming Data is base64 encoded)
	caCert, err := utils.AesEncryptSecret(ctx, req.CaCert, encryptionKeyByte, decodedNonce) // Returns a B64 Encoded encrypted string
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"AesEncryptSecrets.caCert", friendlyMessage+err.Error())
	}
	caKey, err := utils.AesEncryptSecret(ctx, req.CaKey, encryptionKeyByte, decodedNonce)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"AesEncryptSecrets.caCert", friendlyMessage+err.Error())
	}
	etcdCaCert, err := utils.AesEncryptSecret(ctx, req.EtcdCaCert, encryptionKeyByte, decodedNonce)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"AesEncryptSecrets.caCert", friendlyMessage+err.Error())
	}
	etcdCaKey, err := utils.AesEncryptSecret(ctx, req.EtcdCaKey, encryptionKeyByte, decodedNonce)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"AesEncryptSecrets.caCert", friendlyMessage+err.Error())
	}
	etcdCaRotationKey, err := utils.AesEncryptSecret(ctx, req.EtcdCaRotationKey, encryptionKeyByte, decodedNonce)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"AesEncryptSecrets.caCert", friendlyMessage+err.Error())
	}
	saPub, err := utils.AesEncryptSecret(ctx, req.SaPub, encryptionKeyByte, decodedNonce)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"AesEncryptSecrets.caCert", friendlyMessage+err.Error())
	}
	saKey, err := utils.AesEncryptSecret(ctx, req.SaKey, encryptionKeyByte, decodedNonce)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"AesEncryptSecrets.caCert", friendlyMessage+err.Error())
	}
	cpRegistrationCmd, err := utils.AesEncryptSecret(ctx, req.CpRegistrationCmd, encryptionKeyByte, decodedNonce)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"AesEncryptSecrets.caCert", friendlyMessage+err.Error())
	}
	wrkRegistrationCmd, err := utils.AesEncryptSecret(ctx, req.WrkRegistrationCmd, encryptionKeyByte, decodedNonce)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"AesEncryptSecrets.caCert", friendlyMessage+err.Error())
	}

	/* START TRANSACTION */
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}
	/* INSERT INTO DB */
	_, err = tx.ExecContext(ctx, PutClusterCerts, clusterId,
		caCert,
		caKey,
		etcdCaCert,
		etcdCaKey,
		etcdCaRotationKey,
		saPub,
		saKey,
		cpRegistrationCmd,
		wrkRegistrationCmd,
	)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterCerts", friendlyMessage+err.Error())
	}

	/* COMMIT */
	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	return emptyReturn, nil
}

func PutClusterStatusReconcilerQuery(ctx context.Context, dbconn *sql.DB, req *pb.UpdateClusterStatusRequest) (*empty.Empty, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("PutClusterStatusReconcilerQuery")

	friendlyMessage := "PutClusterStatusReconcilerQuery.UnexpectedError"
	failedFunction := "PutClusterStatusReconcilerQuery."

	var nodegroup_id int32
	var cluster_id int32
	var nodeprovider_name string
	var dbNodesName []*NodeStatusID
	var dbNodeGroupsName []*clusterv1alpha.NodegroupStatus
	var dbAddonStatus []*clusterv1alpha.AddonStatus
	var dbIlbStatus []*ilbv1alpha.IlbStatus
	var reqNodes []*clusterv1alpha.NodegroupStatus

	// Check if cluster exists
	err := dbconn.QueryRowContext(ctx, CheckUuid, req.Uuid).Scan(&cluster_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return emptyReturn, errors.New("No cluster found")
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"CheckUuid", friendlyMessage+err.Error())
	}

	// Parse Cluster Status
	var putClusterStatus clusterv1alpha.ClusterStatus
	clusterjson, err := json.Marshal(req.ClusterStatus)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"marshal.req.ClusterStatus", friendlyMessage+err.Error())
	}
	err = json.Unmarshal(clusterjson, &putClusterStatus)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.putClusterStatus", friendlyMessage+err.Error())
	}
	putClusterStatus.Nodegroups = nil
	putClusterStatus.Addons = nil

	// Begin Transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	// Get Nodes from DB k8snodes
	rows, err := tx.QueryContext(ctx, GetNode, cluster_id)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetNode", friendlyMessage+err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var nodegroupId int32
		var kubeletVersion string
		var name string
		var ipAddress string
		var dnsName string
		var instanceIMI string
		var state clusterv1alpha.NodegroupState
		var creationTime string

		// Scan Nodes
		err = rows.Scan(&nodegroupId, &kubeletVersion, &name, &ipAddress, &dnsName, &instanceIMI, &state, &creationTime)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetNode.rows.scan", friendlyMessage+err.Error())
		}

		status := &clusterv1alpha.NodeStatus{
			Name:             name,
			IpAddress:        ipAddress,
			InstanceIMI:      instanceIMI,
			KubeletVersion:   kubeletVersion,
			KubeProxyVersion: kubeletVersion,
			State:            state,
		}
		dbNodesName = append(dbNodesName, &NodeStatusID{
			JsonStatus: status,
			NodegrouID: nodegroupId,
		})
	}

	// Get nodegroup and nodegroup state from db
	rows, err = tx.QueryContext(ctx, GetNodeGroups, cluster_id)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetNodeGroups", friendlyMessage+err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var count int
		var typeng clusterv1alpha.NodegroupType
		var state clusterv1alpha.NodegroupState
		// Scan Node Group information
		err = rows.Scan(&name, &count, &typeng, &state)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetNodeGroups.rows.scan", friendlyMessage+err.Error())
		}

		status := &clusterv1alpha.NodegroupStatus{
			Name:  name,
			Count: count,
			Type:  typeng,
			State: state,
		}
		dbNodeGroupsName = append(dbNodeGroupsName, status)
	}

	// Get Nodegroup Status from request that is from cluster status source of truth from operator
	for _, nodeGroups := range req.ClusterStatus.Nodegroups {
		var putNodeGroupStatus clusterv1alpha.NodegroupStatus
		nodegroupjson, err := json.Marshal(nodeGroups)
		if err != nil {
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"marshal.nodegroupjson", friendlyMessage+err.Error())
		}
		// Parse NodeGroup Status
		err = json.Unmarshal(nodegroupjson, &putNodeGroupStatus)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.nodegroupjson", friendlyMessage+err.Error())
		}
		reqNodes = append(reqNodes, &putNodeGroupStatus)

		// check if the number of nodegroup count atches to number of nodes
		if int(nodeGroups.Count) != len(nodeGroups.Nodes) {
			return emptyReturn, errors.New("The number of nodes doesn't match the count number of nodes. [" + nodeGroups.Name + "]")
		}

		err = tx.QueryRowContext(ctx, GetNodeGroupId, nodeGroups.Name).Scan(&nodegroup_id)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			if err == sql.ErrNoRows {
				return emptyReturn, errors.New("No cluster nodegroup found")
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetNodeGroupId", friendlyMessage+err.Error())
		}
		/* Validation check if nodegroup state is deleting from DB, do not take any action fir update status */

		/* VALIDATE IF NODEGROUP IS ALREADY IN DELETING STATE*/
		deletingState, err := utils.ValidateNodegroupDelete(ctx, dbconn, nodeGroups.Name)
		if err != nil {
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateNodegroupDelete", friendlyMessage)
		}
		if !deletingState {
			_, err = tx.QueryContext(ctx, PutNodeGroupStatus, putNodeGroupStatus, nodeGroups.State, nodeGroups.Name, cluster_id)
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
				}
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"putNodeGroupStatus", friendlyMessage+err.Error())
			}

			dbNodes := getDBNodesbyGroup(dbNodesName, nodegroup_id) // get node status from db nodes
			nodesInStatus := transformToNodeStatus(nodeGroups.Nodes)
			deletedNodesInStatus := getDeletedNodes(nodesInStatus, dbNodes)
			newNodesInStatus, existingNodesInStatus := getNewAndExistingNodes(nodesInStatus, dbNodes)

			//Node difference
			err = tx.QueryRowContext(ctx, GetNodegroupProvider, nodegroup_id).Scan(&nodeprovider_name)
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
				}
				if err == sql.ErrNoRows {
					return emptyReturn, errors.New("No nodeprovider found")
				}
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetNodegroupProvider", friendlyMessage+err.Error())
			}

			if len(newNodesInStatus) > 0 {
				for _, node := range newNodesInStatus {
					// Get Os Image from the compatability information
					var osImageInstanceName string
					err = dbconn.QueryRowContext(ctx, GetOsImageInstanceFromImiArtifact, node.InstanceIMI).Scan(&osImageInstanceName)
					if err != nil {
						if errtx := tx.Rollback(); errtx != nil {
							return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
						}
						return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageInstanceFromImiArtifact", friendlyMessage+err.Error())
					}
					if node.IpAddress != "" {
						_, err = tx.QueryContext(ctx, Insertk8sNode,
							node.Name,
							cluster_id,
							node.State,
							nodegroup_id,
							0, // TEMP IDC INSTANCE ID
							node.IpAddress,
							nodeprovider_name,
							node,
							node.CreationTime.Time,
							"", // TEMP DNS NAME
							osImageInstanceName,
							node.WekaStorageStatus.ClientId,
							node.WekaStorageStatus.Status,
							node.WekaStorageStatus.CustomStatus,
							node.WekaStorageStatus.Message,
						)
						if err != nil {
							if errtx := tx.Rollback(); errtx != nil {
								return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
							}
							return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"InsertK8sNode", friendlyMessage+err.Error())
						}
					}
				}
			}

			if len(deletedNodesInStatus) > 0 {
				for _, node := range deletedNodesInStatus {
					_, err = tx.QueryContext(ctx, Deletek8sNode, node.IpAddress)
					if err != nil {
						if errtx := tx.Rollback(); errtx != nil {
							return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
						}
						return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"nodeDifference.Deletek8sNode", friendlyMessage+err.Error())
					}
				}
			}

			for _, node := range existingNodesInStatus {
				if _, err = tx.QueryContext(
					ctx,
					PutNodeStatus,
					node,
					nodeprovider_name,
					node.State,
					node.CreationTime.Time,
					cluster_id,
					node.Name,
					node.WekaStorageStatus.ClientId,
					node.WekaStorageStatus.Status,
					node.WekaStorageStatus.CustomStatus,
					node.WekaStorageStatus.Message); err != nil {
					if errtx := tx.Rollback(); errtx != nil {
						return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
					}
					return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"PutNodeStatus", friendlyMessage+err.Error())
				}
			}
		}
	}
	nodegrouplistDiff := differencesNodeGroups(reqNodes, dbNodeGroupsName)

	for _, nodegroup := range nodegrouplistDiff {
		if nodegroup.State == "Deleting" {
			var nodeGroupNodedb []*clusterv1alpha.NodeStatus
			rows, err := tx.QueryContext(ctx, GetNodesbyNodegroup, nodegroup.Name, cluster_id)
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
				}
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetNodesbyNodegroup", friendlyMessage+err.Error())
			}
			defer rows.Close()
			for rows.Next() {
				var name string
				var ipAddress string
				var instanceIMI string
				var kubeletVersion string
				var state clusterv1alpha.NodegroupState
				var creationTime string

				err = rows.Scan(&name, &ipAddress, &instanceIMI, &kubeletVersion, &state, &creationTime)
				if err != nil {
					if errtx := tx.Rollback(); errtx != nil {
						return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
					}
					return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetNodesbyNodegroup.rows.scan", friendlyMessage+err.Error())
				}

				var parsedCreationTime time.Time
				parsedCreationTime, err := time.Parse(time.RFC3339, creationTime)
				if err != nil {
					return emptyReturn, err
				}

				status := &clusterv1alpha.NodeStatus{
					Name:             name,
					IpAddress:        ipAddress,
					InstanceIMI:      instanceIMI,
					KubeletVersion:   kubeletVersion,
					KubeProxyVersion: kubeletVersion,
					State:            state,
					CreationTime:     metav1.Time{Time: parsedCreationTime},
				}

				nodeGroupNodedb = append(nodeGroupNodedb, status)
			}
			for _, node := range nodeGroupNodedb {
				_, err = tx.QueryContext(ctx, Deletek8sNode, node.IpAddress)
				if err != nil {
					if errtx := tx.Rollback(); errtx != nil {
						return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
					}
					return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"nodeGroupNodedb.Deletek8sNode", friendlyMessage+err.Error())
				}
			}

			_, err = tx.QueryContext(ctx, Deletek8sNodegroup, nodegroup.Name)
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
				}
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"nodeGroupNodedb.Deletek8sNodegroup", friendlyMessage+err.Error())
			}
		}
	}

	// Addon Status
	rows, err = tx.QueryContext(ctx, GetClusterAddonVersions, cluster_id)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterAddonVersions", friendlyMessage+err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var lastUpdate string
		var state clusterv1alpha.AddonState

		err = rows.Scan(&name, &lastUpdate, &state)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterAddonVersions.rows.scan", friendlyMessage+err.Error())
		}
		status := &clusterv1alpha.AddonStatus{
			Name:  name,
			State: state,
		}

		dbAddonStatus = append(dbAddonStatus, status)
	}

	// Request
	addonsList := []*clusterv1alpha.AddonStatus{}
	for _, addons := range req.ClusterStatus.Addons {
		var putAddonsStatus clusterv1alpha.AddonStatus
		addonjson, err := json.Marshal(addons)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"marshal.addons", friendlyMessage+err.Error())
		}
		err = json.Unmarshal(addonjson, &putAddonsStatus)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.addonsjson", friendlyMessage+err.Error())
		}
		putAddonsStatus.State = clusterv1alpha.AddonState(addons.State)
		addonsList = append(addonsList, &putAddonsStatus)
	}

	for _, addon := range addonsList {
		_, err = tx.QueryContext(ctx, PutClusterAddonVersion, addon, addon.State, cluster_id, addon.Name)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"PutClusterAddonVersion", friendlyMessage+err.Error())
		}
	}

	addonListDiff := differencesAddons(addonsList, dbAddonStatus)

	for _, addonDiff := range addonListDiff {
		if addonDiff.State == "Deleting" {
			_, err = tx.QueryContext(ctx, Deletek8sAddons, addonDiff.Name)
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
				}
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"Deletek8sAddons", friendlyMessage+err.Error())
			}
		}
	}

	/* VIP STATUS */
	// Database
	rows, err = tx.QueryContext(ctx, GetClusterILBVersions, cluster_id)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterILBVersions", friendlyMessage+err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var state ilbv1alpha.State
		var vipId int32
		var name string

		err = rows.Scan(&vipId, &state, &name)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterILBVersions.rows.scan", friendlyMessage+err.Error())
		}
		status := &ilbv1alpha.IlbStatus{
			State: state,
			VipID: int(vipId),
			Name:  name,
		}

		dbIlbStatus = append(dbIlbStatus, status)
	}
	ilbList := []*ilbv1alpha.IlbStatus{}
	var ilbListNameId []*ILBNameID

	rows, err = tx.QueryContext(ctx, GetClusterILBIdName, cluster_id)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterILBIdName", friendlyMessage+err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var vipId int32
		var name string

		err = rows.Scan(&vipId, &name)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterILBIdName.rows.scan", friendlyMessage+err.Error())
		}
		ilbList := &ILBNameID{
			VipID:   int32(vipId),
			VipName: name,
		}

		ilbListNameId = append(ilbListNameId, ilbList)
	}

	ilbList = []*ilbv1alpha.IlbStatus{}
	for _, ilbs := range req.ClusterStatus.Ilbs {
		var putIlbListStatus ilbv1alpha.IlbStatus
		ilbJson, err := json.Marshal(ilbs)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"marshal.ilbs", friendlyMessage+err.Error())
		}
		err = json.Unmarshal(ilbJson, &putIlbListStatus)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.ilbJson", friendlyMessage+err.Error())
		}
		putIlbListStatus.State = ilbv1alpha.State(ilbs.State)
		ilbList = append(ilbList, &putIlbListStatus)
	}

	for _, ilb := range ilbList {
		if ilb.Name != "" {
			ilbName, err := ExtractSuffix(ilb.Name, req.Uuid)
			if err != nil {
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ExtractSuffix", friendlyMessage+err.Error())
			}
			vipId, err := findVipID(ilbListNameId, ilbName)
			if err != nil {
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"findVipID", friendlyMessage+err.Error())
			}
			/* VALIDATE VIP DELETING STATE */
			deletingState, err := utils.ValidateVipDelete(ctx, dbconn, vipId)
			if err != nil {
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidaterVipDelete", friendlyMessage)
			}
			if !deletingState {
				// Validate if vip state is Deleting, if yes do nothing and just check the CRD state = "Deleting" ref line 1020
				if vipId != -1 {
					if ilb.Vip != "" {
						_, err = tx.QueryContext(ctx, PutClusterVIPStatus, ilb, ilb.State, ilb.Vip, vipId, cluster_id)
						if err != nil {
							if errtx := tx.Rollback(); errtx != nil {
								return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
							}
							return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"PutClusterVIPStatus", friendlyMessage+err.Error())
						}
					} else {
						_, err = tx.QueryContext(ctx, PutClusterVIPStatusNoVip, ilb, ilb.State, vipId, cluster_id)
						if err != nil {
							if errtx := tx.Rollback(); errtx != nil {
								return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
							}
							return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"PutClusterVIPStatusNoVip", friendlyMessage+err.Error())
						}
					}
					if ilb.PoolID != 0 {
						_, err = tx.QueryContext(ctx, PutClusterVIPDetailsStatus, strconv.Itoa(ilb.PoolID), ilbName, vipId)
						if err != nil {
							if errtx := tx.Rollback(); errtx != nil {
								return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
							}
							return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"PutClusterVIPDetailsStatus", friendlyMessage+err.Error())
						}
					}
				}
			}
		}
	}
	vipListDiff := differencesILB(ilbList, dbIlbStatus)

	for _, vipDiff := range vipListDiff {
		if vipDiff.State == "Deleting" {
			ruleExists, err := utils.ValidateSecRuleExistance(ctx, dbconn, int32(vipDiff.VipID))
			if err != nil {
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateSecRuleExistance", friendlyMessage)
			}
			if ruleExists {
				var firewall_status string
				err = tx.QueryRowContext(ctx, query.GetFirewallRuleStateQuery, int32(vipDiff.VipID)).Scan(&firewall_status)
				if err != nil {
					return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetFirewallRuleStateQuery", friendlyMessage)
				}
				if firewall_status == "Not Specified" {
					if err := deleteILB(ctx, tx, int32(vipDiff.VipID), failedFunction, friendlyMessage); err != nil {
						return emptyReturn, err
					}
				}
			} else {
				if err := deleteILB(ctx, tx, int32(vipDiff.VipID), failedFunction, friendlyMessage); err != nil {
					return emptyReturn, err
				}
			}
		}
	}

	/* Storage Status */
	// Get Current DB Storage List in clusterv1 format
	dbStorageList := []*clusterv1alpha.StorageStatus{}
	rows, err = tx.QueryContext(ctx, GetClusterStorages, cluster_id)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetNodeGroups", friendlyMessage+err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var storageId string
		storageStatus := &clusterv1alpha.StorageStatus{}
		err = rows.Scan(&storageId, &storageStatus.State, &storageStatus.Provider)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterStorages.rows.scan", friendlyMessage+err.Error())
		}
		dbStorageList = append(dbStorageList, storageStatus)
	}

	// Get Storage Status List from Request in clusterv1 format
	reqStorageList := []*clusterv1alpha.StorageStatus{}
	for _, storageStatus := range req.ClusterStatus.Storages {
		var storageStatusCrd clusterv1alpha.StorageStatus
		storageStatusJson, err := json.Marshal(storageStatus)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"marshal.storageSttus", friendlyMessage+err.Error())
		}
		err = json.Unmarshal(storageStatusJson, &storageStatusCrd)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.storageSttus", friendlyMessage+err.Error())
		}
		storageStatusCrd.State = clusterv1alpha.StorageState(storageStatus.StorageState)
		reqStorageList = append(reqStorageList, &storageStatusCrd)
	}

	// Update DB with latest info from request (reqStorageList)
	for _, reqStorage := range reqStorageList {
		_, err = tx.QueryContext(ctx, PutStorageStatus, cluster_id, reqStorage.Provider, reqStorage, reqStorage.State)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"PutClusterStorage", friendlyMessage+err.Error())
		}
	}

	// Get the DB storages to remove (Exists in DB list, but not in Req list)
	storageListDelete := differencesStorages(reqStorageList, dbStorageList)
	for _, storageDelete := range storageListDelete {
		if storageDelete.State == "Deleting" {
			_, err = tx.QueryContext(ctx, PutStorageStatus, cluster_id, storageDelete.Provider, "Deleted")
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
				}
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"DeleteStorageStatus", friendlyMessage+err.Error())
			}
		}
	}

	// update firewall status
	dbfwList := []*clusterv1alpha.FirewallStatus{}
	rowsFW, err := tx.QueryContext(ctx, GetClusterFirewall, cluster_id)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetFirewalls", friendlyMessage+err.Error())
	}
	defer rowsFW.Close()
	for rowsFW.Next() {
		var vip_id int32
		var sourceips []byte
		var sourceipsresult []string
		firewallStatus := &clusterv1alpha.FirewallStatus{}
		err = rowsFW.Scan(&firewallStatus.Firewallrulestatus.State, &firewallStatus.DestinationIp, &sourceips, &vip_id, &firewallStatus.Port, &firewallStatus.Protocol)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetFirewallStatus.rowsFW.scan", friendlyMessage+err.Error())
		}
		if sourceips != nil {
			err = json.Unmarshal(sourceips, &sourceipsresult)
			if err != nil {
				return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetMetadata.unmarshall.souceip", friendlyMessage+err.Error())
			}
			firewallStatus.SourceIps = sourceipsresult
		} else {
			firewallStatus.SourceIps = []string{}
		}
		dbfwList = append(dbfwList, firewallStatus)
	}

	// Get FW status List from Request in clusterv1 format
	reqfwList := []*clusterv1alpha.FirewallStatus{}
	for _, fwStatus := range req.ClusterStatus.Firewall {
		var fwstatusCrd clusterv1alpha.FirewallStatus
		fwStatusJson, err := json.Marshal(fwStatus)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"marshal.fwstatus", friendlyMessage+err.Error())
		}

		err = json.Unmarshal(fwStatusJson, &fwstatusCrd)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.fwstatus", friendlyMessage+err.Error())
		}
		if fwStatus.Firewallstate == "" {
			fwstatusCrd.Firewallrulestatus.State = "Pending"
		} else {
			fwstatusCrd.Firewallrulestatus.State = fwv1alpha1.State(fwStatus.Firewallstate)
		}
		reqfwList = append(reqfwList, &fwstatusCrd)
	}

	// Update DB with latest info from request (reqFWList)
	for _, reqfw := range reqfwList {
		var firewall_status string
		err = tx.QueryRowContext(ctx, GetFirewallRuleStateClusterQuery, cluster_id, reqfw.DestinationIp).Scan(&firewall_status)
		if err != nil {
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetFirewallRuleStateQuery", friendlyMessage)
		}
		// if db state is "Deleting" but cluster CRD firewall state is still not updated to "Deleting",Do not update the DB
		// except if it's already progressed to "Deleted" state
		if firewall_status == "Deleting" && reqfw.Firewallrulestatus.State != "Deleted" {
			continue
		} else {
			err = setFwStatus(ctx, tx, cluster_id, reqfw, failedFunction, friendlyMessage)
			if err != nil {
				return emptyReturn, err

			}
		}
	}

	// Get the DB fw to remove (Exists in DB list, but not in Req list)
	fwListDelete := differenceFW(reqfwList, dbfwList)
	for _, fwDelete := range fwListDelete {
		_, err = tx.QueryContext(ctx, PutFwStatus, cluster_id, fwDelete.DestinationIp, "Not Specified")
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"DeleteFirewalleStatusNotSpecified", friendlyMessage+err.Error())
		}
		_, err = tx.QueryContext(ctx, DeleteFWRule, fwDelete.DestinationIp)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"DeleteFirewallrule", friendlyMessage+err.Error())
		}
		_, err = tx.QueryContext(ctx, DeleteProtocol, fwDelete.DestinationIp)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"DeleteFirewallProtocol", friendlyMessage+err.Error())
		}
	}

	// Cluster Status
	_, err = tx.QueryContext(ctx, PutClusterStatus, putClusterStatus, putClusterStatus.State, cluster_id)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"PutClusterStatus", friendlyMessage+err.Error())
	}

	/* VERIFY THAT CLUSTER IS IN COMMITABLE STATE */
	// Get Cluster state
	var currentClusterState string
	err = dbconn.QueryRowContext(ctx, GetClusterState, cluster_id).Scan(&currentClusterState)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterState", friendlyMessage+err.Error())
	}
	if currentClusterState == "Pending" || currentClusterState == "DeletePending" || currentClusterState == "Deleting" || currentClusterState == "Deleted" {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		log.Info("[WARNING] GetClusterStates Cluster is updating, not commiting changes", logkeys.Function, failedFunction+"GetClusterStates", logkeys.CurrentClusterState, currentClusterState)
		return emptyReturn, nil
	}
	// Get Cluster rev applied
	var changeApplied bool
	err = dbconn.QueryRowContext(ctx, GetClusterRevChangeAppliedQuery, req.Uuid).Scan(&changeApplied)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterRevChangeApplied", friendlyMessage+err.Error())
	}
	if !changeApplied {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		log.Info("[WARNING] GetClusterStates Cluster is updating, not commiting changes", logkeys.Function, failedFunction+"GetClusterRevChangeApplied", logkeys.ChangeApplied, changeApplied)
		return emptyReturn, nil
	}

	/* COMMIT CHANGES */
	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	return emptyReturn, nil
}

// setFwStatus sets the firewall status in the database according to one in the request
func setFwStatus(ctx context.Context, tx *sql.Tx, cluster_id int32, reqfw *clusterv1alpha.FirewallStatus, failedFunction string, friendlyMessage string) error {
	_, err := tx.QueryContext(ctx, PutFwStatus, cluster_id, reqfw.DestinationIp, reqfw.Firewallrulestatus.State)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return utils.ErrorHandler(ctx, err, failedFunction+"PutClusterfirewallStates", friendlyMessage+err.Error())
	}
	return nil
}

// deleteILB deletes the ILB from the database
func deleteILB(ctx context.Context, tx *sql.Tx, vipID int32, failedFunction string, friendlyMessage string) error {
	if _, err := tx.QueryContext(ctx, Deletek8sILBDetails, vipID); err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return utils.ErrorHandler(ctx, err, failedFunction+"Deletek8sILBDetails", friendlyMessage+err.Error())
	}
	if _, err := tx.QueryContext(ctx, Deletek8sILB, vipID); err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return utils.ErrorHandler(ctx, err, failedFunction+"Deletek8sILB", friendlyMessage+err.Error())
	}
	return nil
}
