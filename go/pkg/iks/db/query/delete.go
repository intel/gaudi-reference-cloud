// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	clusterv1alpha "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	SetClusterState = `
		UPDATE public.cluster
		SET clusterstate_name = $1
		WHERE cluster_id = $2
	`
	SetAllClusterNodegroupsState = `
		UPDATE public.nodegroup
		SET nodegroupstate_name = $1
		WHERE cluster_id = $2
	`
	SetAllClusterVipsState = `
		UPDATE public.vip
		SET vipstate_name= $1
		WHERE cluster_id = $2
	`
	SetAllClusterAddonsState = `
		UPDATE public.clusteraddonversion
		SET clusteraddonstate_name = $1
		WHERE cluster_id = $2
	`
	DeleteVipRecordQuery = `
	UPDATE public.vip SET vip_status = $3 , vipstate_name = 'Deleting' where vip_id = $1 AND cluster_id = $2 AND owner = 'customer';
	`
	DeleteFirewallRuleQuery = `
	UPDATE public.vip SET firewall_status = 'Deleting' where vip_id = $1;
	`

	GetdestipQuery = `
	Select vip_ip from public.vip where vip_id = $1`
)

func DeleteRecord(ctx context.Context, dbconn *sql.DB, record *pb.ClusterID) (*empty.Empty, error) {

	friendlyMessage := "Could not delete Cluster. Please try again."
	failedFunction := "DeleteRecord."
	emptyReturn := &empty.Empty{}

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return emptyReturn, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return emptyReturn, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) //returning not found to avoid leaking cluster existence
	}

	/* VALIDATE CLUSTER DELETING STATE */
	deletingState, err := utils.ValidateClusterDelete(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidaterClusterDelete", friendlyMessage)
	}
	if deletingState {
		return emptyReturn, status.Error(codes.FailedPrecondition, "Cluster can not be deleted , it is currently in deleting state")
	}

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage)
	}
	/* SET CLUSTER OBJECTS TO DELETING*/
	// Set cluster's nodegroups to deleting
	_, err = tx.QueryContext(ctx, SetAllClusterNodegroupsState, "Deleting", clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"SetAllClusterNodegroupsState", friendlyMessage)
	}
	// Set cluster's vips to deleting
	_, err = tx.QueryContext(ctx, SetAllClusterVipsState, "Deleting", clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"SetAllClusterVipsState", friendlyMessage)
	}
	// Set cluster's addons to deleting
	_, err = tx.QueryContext(ctx, SetAllClusterAddonsState, "Deleting", clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"SetAllClusterAddonsState", friendlyMessage)
	}
	// Set cluster to delete pending
	_, err = tx.QueryContext(ctx, SetClusterState, "DeletePending", clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"DetClusterState", friendlyMessage)
	}

	// Insert Provisioning Log
	_, err = tx.QueryContext(ctx, InsertProvisioningQuery,
		clusterId,
		"cluster delete pending",
		"INFO",
		"cluster delete",
		time.Now(),
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"InsertProvisioningQuery", friendlyMessage)
	}
	// commit Transaction
	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}
	return emptyReturn, nil
}

func DeleteNodeGroupRecord(ctx context.Context, dbconn *sql.DB, record *pb.NodeGroupid) (*empty.Empty, error) {
	friendlyMessage := "Could not delete Cluster Node Group. Please try again."
	failedFunction := "DeleteNodeGroupRecord."
	emptyReturn := &empty.Empty{}

	/* VALIDATE CLUSTER AND NODEGROUP EXISTANCE */
	clusterId, nodeGroupId, err := utils.ValidateNodeGroupExistance(ctx, dbconn, record.Clusteruuid, record.Nodegroupuuid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateNodeGroupExistance", friendlyMessage)
	}
	if clusterId == -1 || nodeGroupId == -1 {
		return emptyReturn, status.Errorf(codes.NotFound, "NodeGroup not found in Cluster: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return emptyReturn, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* VALIDATE CLUSTER IS ACTIONABLE*/
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidaterClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return emptyReturn, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	/* VALIDATE SUPER COMPUTE NODEGROUP TYPE TO MAKE SURE WE ARE DELETING GP NODEGROUP TYPE */
	isSuperComputeCluster, err := utils.ValidateSuperComputeClusterType(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return emptyReturn, err
	}
	if isSuperComputeCluster {
		_, err := utils.ValidateSuperComputeGPNodegroupType(ctx, dbconn, record.Clusteruuid, record.Nodegroupuuid)
		if err != nil {
			return emptyReturn, err
		}
	}

	/* VALIDATE IF NODEGROUP IS ALREADY IN DELETING STATE*/
	deletingState, err := utils.ValidateNodegroupDelete(ctx, dbconn, record.Nodegroupuuid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateNodegroupDelete", friendlyMessage)
	}
	if deletingState {
		return emptyReturn, status.Error(codes.FailedPrecondition, "Cannot delete nodegroup in deleting state")
	}
	/* START THE DB TRANSACTION */
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage)
	}

	/* MARK NODEGROUP AS 'DELETING' IN THE DATABSE */
	_, err = tx.QueryContext(ctx, UpdateNodegroupStateQuery, record.Clusteruuid, record.Nodegroupuuid, "Deleting")
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"UpdateNodegroupStateQuery", friendlyMessage)
	}
	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, UpdateClusterStateQuery, clusterId, "Pending")
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
	}
	/* UPDATE CLUSTER REV TABLE*/
	// Get current json from REV table
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage)
	}

	// Conditions
	// nodegroup with length 1 {1} if record found remove record return empty
	// deleting element in between {1,2,3,4} if found, return start to index-1 , index+1 to last
	// element first{1,2} if index = 0 , return index+1 to last
	// element last{1,2}  if index = (length-1) return start to index-1

	// Create Nodegroup CRD and append to clusterCrd
	var deleteNodeIndex = -1 //Required when the delete call is made again and nodegroup is not present in CRD
	for i, nodeGroup := range clusterCrd.Spec.Nodegroups {
		if nodeGroup.Name == record.Nodegroupuuid {
			deleteNodeIndex = i
			break
		}
	}
	var ngcrdlength = len(clusterCrd.Spec.Nodegroups)
	if ngcrdlength > 1 { // Check if length of ng crd is greater than 1
		if deleteNodeIndex != -1 {
			if (deleteNodeIndex + 1) == ngcrdlength { // if index is last element
				clusterCrd.Spec.Nodegroups = clusterCrd.Spec.Nodegroups[:deleteNodeIndex]
			} else if deleteNodeIndex == 0 { // if index is first element
				clusterCrd.Spec.Nodegroups = clusterCrd.Spec.Nodegroups[deleteNodeIndex+1:]
			} else {
				clusterCrd.Spec.Nodegroups = append(clusterCrd.Spec.Nodegroups[:deleteNodeIndex], clusterCrd.Spec.Nodegroups[deleteNodeIndex+1:]...)
			}
		}
	} else if ngcrdlength == 1 { //there is only one nodegroup left and it is getting deleted
		if deleteNodeIndex == 0 {
			clusterCrd.Spec.Nodegroups = []clusterv1alpha.NodegroupTemplateSpec{}
		}
	}

	if deleteNodeIndex == -1 { //Required when the delete call is made again and nodegroup is not present in CRD
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, nil
	}
	// Create new cluster rev table entry
	var revversion string
	clusterCrdJson, err := json.Marshal(clusterCrd)
	err = tx.QueryRowContext(ctx, InsertRevQuery,
		clusterId,
		currentJson,
		clusterCrdJson,
		"test",
		"test",
		"test",
		"test",
		time.Now(),
		false,
	).Scan(&revversion)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	}

	/* COMMIT TRANSACTION */
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}
	return emptyReturn, nil
}

func DeleteVip(ctx context.Context, dbconn *sql.DB, record *pb.VipId) (*empty.Empty, error) {
	friendlyMessage := "Could not delete Cluster VIP. Please try again"
	failedFunction := "DeleteVip."
	emptyReturn := &empty.Empty{}

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return emptyReturn, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return emptyReturn, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* VALIDATE CLUSTER AND VIP EXISTANCE */
	clusterId, vipId, err := utils.ValidateVipExistance(ctx, dbconn, record.Clusteruuid, record.Vipid, true)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateVipExistance", friendlyMessage)
	}
	if clusterId == -1 || vipId == -1 {
		return emptyReturn, status.Errorf(codes.NotFound, "Vip not found in Cluster, Cluster: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER IS ACTIONABLE*/
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return emptyReturn, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	/* VALIDATE VIP DELETING STATE */
	deletingState, err := utils.ValidateVipDelete(ctx, dbconn, record.Vipid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidaterVipDelete", friendlyMessage)
	}
	if deletingState {
		return emptyReturn, status.Errorf(codes.FailedPrecondition, "Vip can not be deleted, it is currently in deleting state")
	}

	/* VALIDATE FW RULE IS ACTIVE */
	// If security rule exists and not active it cannot be deleted
	SecState, err := utils.ValidateSecActive(ctx, dbconn, record.Vipid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateSecurityDelete", friendlyMessage)
	}
	if SecState {
		return emptyReturn, status.Error(codes.FailedPrecondition, "Vip can not be deleted, there is security rule with non active state")
	}

	/* START THE DB TRANSACTION */
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage)
	}
	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, UpdateClusterStateQuery,
		clusterId,
		"Pending",
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
	}

	//update vip table with status as deleting
	_, err = tx.QueryContext(ctx, DeleteVipRecordQuery, record.Vipid, clusterId, `{"status":"Vip is being deleted"}`)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteVipRecordQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER REV TABLE*/
	// Get current json from REV table
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage)
	}
	// Get Vip name
	var vipname string
	err = tx.QueryRowContext(ctx, GetVipNameQuery, record.Vipid).Scan(&vipname)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"getVipNameQuery", friendlyMessage)
	}

	// Conditions
	// deleting element in between {1,2,3,4} if found, return start to index-1 , index+1 to last
	// element first{1,2} if index = 0 , return index+1 to last
	// element last{1,2}  if index = (length-1) return start to index-1

	var vipcrdlength = len(clusterCrd.Spec.ILBS)

	var deletevipIndex = -1 //Required when the delete call is made again and vip is not present in CRD
	for i, ilb := range clusterCrd.Spec.ILBS {
		if vipname == ilb.Name {
			deletevipIndex = i
			break
		}
	}
	if deletevipIndex != -1 {
		if (deletevipIndex + 1) == vipcrdlength { //last element is index
			clusterCrd.Spec.ILBS = clusterCrd.Spec.ILBS[:deletevipIndex]
		} else if deletevipIndex == 0 { // index is first element
			clusterCrd.Spec.ILBS = clusterCrd.Spec.ILBS[deletevipIndex+1:]
		} else {
			clusterCrd.Spec.ILBS = append(clusterCrd.Spec.ILBS[:deletevipIndex], clusterCrd.Spec.ILBS[deletevipIndex+1:]...)
		}
	}

	if deletevipIndex == -1 { //Required when the delete call is made again and vip is not present in CRD
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, nil
	}

	//if secruleExists {
	var destip string
	err = tx.QueryRowContext(ctx, GetdestipQuery, record.Vipid).Scan(&destip)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetdestipQuery", friendlyMessage)
	}
	// update vip table for firewall rule deletion
	_, err = tx.QueryContext(ctx, DeleteFirewallRuleQuery, record.Vipid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"DeletefirewallRulequery", friendlyMessage)
	}

	var fwlength = len(clusterCrd.Spec.Firewall)

	var deletefwIndex = -1 //Required when the delete call is made again and vip is not present in CRD
	for i, fw := range clusterCrd.Spec.Firewall {
		if destip == fw.DestinationIp {
			deletefwIndex = i
			break
		}
	}
	if deletefwIndex != -1 {
		if (deletefwIndex + 1) == fwlength { //last element is index
			clusterCrd.Spec.Firewall = clusterCrd.Spec.Firewall[:deletefwIndex]
		} else if deletefwIndex == 0 { // index is first element
			clusterCrd.Spec.Firewall = clusterCrd.Spec.Firewall[deletefwIndex+1:]
		} else {
			clusterCrd.Spec.Firewall = append(clusterCrd.Spec.Firewall[:deletefwIndex], clusterCrd.Spec.Firewall[deletefwIndex+1:]...)
		}
	}
	// Create new cluster rev table entry
	var revversion string
	clusterCrdJson, err := json.Marshal(clusterCrd)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"marshal cluster crd error", friendlyMessage)
	}

	err = tx.QueryRowContext(ctx, InsertRevQuery,
		clusterId,
		currentJson,
		clusterCrdJson,
		"test",
		"test",
		"test",
		"test",
		time.Now(),
		false,
	).Scan(&revversion)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	}
	//}
	/* COMMIT TRANSACTION */
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}

	return emptyReturn, nil
}

func DeleteFirewallRule(ctx context.Context, dbconn *sql.DB, record *pb.DeleteFirewallRuleRequest) (*empty.Empty, error) {
	friendlyMessage := "Could not delete security rule Please try again."
	failedFunction := "DeleteRecord."
	emptyReturn := &empty.Empty{}

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return emptyReturn, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return emptyReturn, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* Validate if vip is public-apiserver */
	var vipname string
	err = dbconn.QueryRowContext(ctx, GetVipNameQuery, record.Vipid).Scan(&vipname)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"getVipNameQuery", friendlyMessage)
	}

	if vipname == "public-apiserver" {
		/* VALIDATE CLUSTER AND VIP EXISTANCE */
		clusterId, vipId, err := utils.ValidateVipExistance(ctx, dbconn, record.Clusteruuid, record.Vipid, false)
		if err != nil {
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateVipExistance", friendlyMessage)
		}
		if clusterId == -1 || vipId == -1 {
			return emptyReturn, status.Errorf(codes.NotFound, "Public apiserver vip not found in Cluster, Cluster: %s", record.Clusteruuid)
		}
	} else {
		clusterId, vipId, err := utils.ValidateVipExistance(ctx, dbconn, record.Clusteruuid, record.Vipid, true)
		if err != nil {
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateVipExistance", friendlyMessage)
		}
		if clusterId == -1 || vipId == -1 {
			return emptyReturn, status.Errorf(codes.NotFound, "Vip not found in Cluster, Cluster: %s", record.Clusteruuid)
		}
	}

	/* VALIDATE IF VIP HAS SECURITY RULE */
	ruleExists, err := utils.ValidateSecRuleExistance(ctx, dbconn, record.Vipid)
	if !ruleExists {
		return emptyReturn, status.Error(codes.NotFound, "Security rules does not exists for this vip")
	}

	/* VALIDATE CLUSTER IS ACTIONABLE*/
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return emptyReturn, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	/* VALIDATE IF VIP IS NOT IN DELETING STATE*/
	deletingState, err := utils.ValidateVipDelete(ctx, dbconn, record.Vipid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidaterVipDelete", friendlyMessage)
	}
	if deletingState {
		return emptyReturn, status.Error(codes.FailedPrecondition, "Security rule can not be deleted , vip is currently in deleting state")
	}

	/* VALIDATE IF SECURITY RULE IS NOT IN DELETING STATE*/
	SecState, err := utils.ValidateSecActive(ctx, dbconn, record.Vipid)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"ValidateSecurityDelete", friendlyMessage)
	}
	if SecState {
		return emptyReturn, status.Error(codes.FailedPrecondition, "Only Active security rules can be deleted")
	}

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage)
	}

	// delete from vip table
	_, err = tx.QueryContext(ctx, DeleteFirewallRuleQuery, record.Vipid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, UpdateClusterStateQuery, clusterId, "Pending")
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
	}
	/* UPDATE CLUSTER REV TABLE*/
	// Get current json from REV table
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage)
	}

	// Get Vip name
	var destip string
	err = tx.QueryRowContext(ctx, GetdestipQuery, record.Vipid).Scan(&destip)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"GetdestipQuery", friendlyMessage)
	}
	// Conditions
	// deleting element in between {1,2,3,4} if found, return start to index-1 , index+1 to last
	// element first{1,2} if index = 0 , return index+1 to last
	// element last{1,2}  if index = (length-1) return start to index-1

	var fwlength = len(clusterCrd.Spec.Firewall)

	var deletefwIndex = -1 //Required when the delete call is made again and vip is not present in CRD
	for i, fw := range clusterCrd.Spec.Firewall {
		if destip == fw.DestinationIp {
			deletefwIndex = i
			break
		}
	}
	if deletefwIndex != -1 {
		if (deletefwIndex + 1) == fwlength { //last element is index
			clusterCrd.Spec.Firewall = clusterCrd.Spec.Firewall[:deletefwIndex]
		} else if deletefwIndex == 0 { // index is first element
			clusterCrd.Spec.Firewall = clusterCrd.Spec.Firewall[deletefwIndex+1:]
		} else {
			clusterCrd.Spec.Firewall = append(clusterCrd.Spec.Firewall[:deletefwIndex], clusterCrd.Spec.Firewall[deletefwIndex+1:]...)
		}
	}

	if deletefwIndex == -1 { //Required when the delete call is made again and vip is not present in CRD
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, nil
	}

	// Create new cluster rev table entry
	var revversion string
	clusterCrdJson, err := json.Marshal(clusterCrd)
	err = tx.QueryRowContext(ctx, InsertRevQuery,
		clusterId,
		currentJson,
		clusterCrdJson,
		"test",
		"test",
		"test",
		"test",
		time.Now(),
		false,
	).Scan(&revversion)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	}

	/* COMMIT TRANSACTION */
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}

	return emptyReturn, nil

}
