// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	clusterv1alpha "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	UpdateClusterAdminKubeconfig = `
 		UPDATE public.cluster
		SET admin_kubeconfig = $2
		WHERE cluster_id = $1
 	`
	UpdateClusterReadonlyKubeconfig = `
 		UPDATE public.cluster
		SET readonly_kubeconfig = $2
		WHERE cluster_id = $1
 	`

	updateNodegroupImiQuery = `
		UPDATE public.nodegroup
		SET k8sversion_name = $2, osimageinstance_name = $3
		WHERE unique_id = $1
	`
	getNodegroupQuery = `
	 SELECT nodegroup_id, nodegrouptype_name, name, instancetype_name, osimage_name, sshkey,
	 	nodecount, upgstrategydrainbefdel, upgstrategymaxnodes
	 FROM public.nodegroup
	 WHERE cluster_id = $1
	 `
	getNodeGroupStateQuery = `
		SELECT n.nodegroupstate_name
		FROM  public.nodegroup n
		WHERE n.unique_id = $1
	`
	getAddOnQuery = `
	SELECT addonversion_name, addonargs FROM public.clusteraddonversion where cluster_id = $4
	`
	getClusterMetadataQuery = `
		SELECT
			c.name,
			c.description,
			c.tags,
			c.annotations
		FROM public.cluster c
		WHERE c.unique_id = $1
	`
	getNodeGroupMetadataQuery = `
		SELECT
			n.name,
			n.description,
			n.tags,
			n.annotations,
			n.nodecount,
			n.upgstrategydrainbefdel,
			n.upgstrategymaxnodes
		FROM public.nodegroup n
		WHERE n.unique_id = $1
	`
	getLatestClusterRev = `
		SELECT desiredspec_json
		FROM public.clusterrev
		WHERE cluster_id = $1
		ORDER BY clusterrev_id desc
		LIMIT 1
	`
	GetClusterCheckNameQuery = `
	SELECT name, clusterstate_name, unique_id
	 FROM public.cluster where name = $1
	`
	updateNodeGroupQuery = `
		UPDATE public.nodegroup
		SET nodegroupstate_name = $2,
				nodecount = COALESCE($3, nodecount),
				name = COALESCE($4, name),
				description = COALESCE($5, description),
				annotations = COALESCE($6, annotations),
				tags = COALESCE($7, tags),
				upgstrategydrainbefdel = $8,
				upgstrategymaxnodes = $9
		WHERE unique_id = $1
 	`
	updateClusterValidationQuery = `
		SELECT c.clusterstate_name, n.runtime_name
	 	FROM public.cluster c
			INNER JOIN nodegroup n
			ON c.cluster_id = n.cluster_id AND nodegrouptype_name = 'ControlPlane'
		WHERE c.unique_id = $1
	`
	updateClusterTableQuery = `
		UPDATE public.cluster
		SET (name, description, annotations, tags) = (
			$2,
			$3,
			COALESCE($4, annotations),
			COALESCE($5, tags)
		)
		WHERE unique_id = $1
		RETURNING cluster_id
	`
	updateClusterControlPlaneQuery = `
		UPDATE public.nodegroup
		SET (k8sversion_name, osimageinstance_name) = ($2, $3)
		WHERE cluster_id = $1 AND nodegrouptype_name = 'ControlPlane'
		RETURNING nodegroup_id
	`
	UpdateClusterStateQuery = `
		UPDATE public.cluster
		SET clusterstate_name = $2
		WHERE cluster_id = $1
	`
	UpdateClusterStorageDataQuery = `
		UPDATE public.cluster
		SET storage_enable = $2
		WHERE cluster_id = $1
	`
	UpdateNodegroupStateQuery = `
		UPDATE public.nodegroup
		SET nodegroupstate_name = $3
		WHERE unique_id = $2 AND cluster_id = (SELECT cluster_id FROM public.cluster WHERE unique_id = $1)
	`
	GetDefaultStorageProvider = `
		SELECT storageprovider_name
		FROM storageprovider
		WHERE is_default=true
	`
	InsertStorageTableQuery = `
		INSERT INTO public.storage (cluster_id, storageprovider_name, storagestate_name, size) values ($1, $2, $3, $4)
	`

	UpdateStorageTableQuery = `
		INSERT INTO public.storage (cluster_id, storageprovider_name, storagestate_name, size) values ($1, $2, $3, $4)
	`

	UpdateCloudAccountStorageQuery = `
		UPDATE public.cloudaccountextraspec
		SET total_storage_size = total_storage_size + $2
		WHERE cloudaccount_id = $1
	`

	PutVipQuery = `
	UPDATE public.vip set sourceips= $1 , firewall_status = $2 where vip_id = $3`

	GetVipStatebyVipIdQuery = `
	Select vipstate_name from vip where vip_id = $1`

	GetFirewallRuleStateQuery = `
	select COALESCE(firewall_status,'Not Specified') from vip where vip_id = $1`

	GetFirewallRuleQuery = `
	SELECT v.sourceips FROM public.vip v INNER JOIN public.vipdetails d ON v.vip_id = d.vip_id
	WHERE v.cluster_id = $1 AND v.vip_id = $2 AND v.vip_ip = $3 AND d.port = $4 AND jsonb(d.protocol) = $5
	`

	UpdateStorageSizeTableQuery = `
	UPDATE public.storage SET (storagestate_name, size) = ($2, $3) WHERE cluster_id = $1
    `
)

func PatchRecord(ctx context.Context, dbconn *sql.DB, record *pb.UpdateClusterRequest) (*pb.ClusterCreateResponseForm, error) {
	// Start the transaction

	friendlyMessage := "Could not Update Cluster. Please try again."
	failedFunction := "PatchRecord."
	returnError := &pb.ClusterCreateResponseForm{}
	returnValue := &pb.ClusterCreateResponseForm{}

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

	/* VALIDATE CLUSTER IS ACTIONABLE*/
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return returnError, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	/* GET CURRENT CLUSTER METADATA INFOMRATION */
	var (
		clusterName        string
		clusterDescription string
		clusterTags        []byte
		clusterAnnotations []byte
	)
	crdChange := false
	err = dbconn.QueryRowContext(ctx, getClusterMetadataQuery,
		record.Clusteruuid,
	).Scan(&clusterName, &clusterDescription, &clusterTags, &clusterAnnotations)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClusterMetadataQuery", friendlyMessage)
	}

	/* CHECK FOR CLUSTER NAME CHANGES*/
	if record.Name != nil {
		rows, err := dbconn.QueryContext(ctx, GetClustersStatesByName, record.Name, record.CloudAccountId)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterStatesByName", friendlyMessage)
		}
		defer rows.Close()
		for rows.Next() {
			var (
				clusterState string
				clusterUuid  string
			)
			err = rows.Scan(&clusterUuid, &clusterState)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterStatesByName.rows.scan", friendlyMessage)
			}
			if clusterState != "Deleting" && clusterState != "Deleted" && clusterUuid != record.Clusteruuid {
				return returnError, status.Error(codes.AlreadyExists, "Cluster name already in use")
			}
		}
		returnValue.Name = *record.Name
		clusterName = *record.Name
	}
	/* CHECK FOR CLUSTER DESCRIPTION CHANGES */
	if record.Description != nil {
		clusterDescription = *record.Description
	}

	/* CHECK FOR ANNOTATION CHANGES*/
	dbAnnotations, err := utils.ParseKeyValuePairIntoMap(ctx, clusterAnnotations)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ParseKeyValuePairIntoMap", friendlyMessage)
	}
	recordAnnotations := make(map[string]string)
	for _, annotation := range record.Annotations {
		if recordAnnotations[annotation.Key] != "" {
			return returnError, status.Error(codes.InvalidArgument, "Annotations sent are duplicated. Correct and try again")
		}
		recordAnnotations[annotation.Key] = annotation.Value
	}
	if reflect.DeepEqual(dbAnnotations, recordAnnotations) {
		record.Annotations = nil
	} else {
		crdChange = true
	}

	/* CHECK FOR TAGS CHANGES*/
	dbTags, err := utils.ParseKeyValuePairIntoMap(ctx, clusterTags)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ParseKeyValuePairIntoMap", friendlyMessage)
	}
	recordTags := make(map[string]string)
	for _, tag := range record.Tags {
		if recordTags[tag.Key] != "" {
			return returnError, status.Error(codes.InvalidArgument, "Tags sent are duplicated. Correct and try again")
		}
		recordTags[tag.Key] = tag.Value
	}
	if reflect.DeepEqual(dbTags, recordTags) {
		record.Tags = nil
	}

	/* START TRANSACTION */
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage)
	}

	/* UPDATE CLUSTER TABLE */
	err = tx.QueryRowContext(ctx, updateClusterTableQuery,
		record.Clusteruuid,
		clusterName,
		clusterDescription,
		record.Annotations,
		record.Tags,
	).Scan(&clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateClusterTableQuery", friendlyMessage)
	}
	/* UPDATE CLUSTER REV TABLE*/
	if crdChange {
		// Get current json from REV table
		currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage)
		}
		// Update the clusterCrd
		clusterCrd.Annotations = recordAnnotations

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
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
		}
		/* UPDATE CLUSTER STATE TO PENDING*/
		_, err = tx.QueryContext(ctx, UpdateClusterStateQuery,
			clusterId,
			"Pending",
		)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
		}
	}

	// Insert Provisioning Log
	_, err = tx.QueryContext(ctx, InsertProvisioningQuery,
		clusterId,
		"cluster update pending",
		"INFO",
		"cluster update",
		time.Now(),
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertProvisioningQuery", friendlyMessage)
	}

	/* Set Values*/
	returnValue.Uuid = record.Clusteruuid
	err = tx.QueryRowContext(ctx, getClusterStatusInfoQuery,
		record.Clusteruuid,
	).Scan(
		&returnValue.Name,
		&returnValue.Clusterstate,
		&returnValue.K8Sversionname,
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClusterStatueInfoQuery", friendlyMessage)
	}
	// commit Transaction
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}

	return returnValue, nil
}

func UpgradeCluster(ctx context.Context, dbconn *sql.DB, record *pb.UpgradeClusterRequest) (*pb.ClusterStatus, error) {
	friendlyMessage := "Could not updgrade Cluster. Please try again"
	failedFunction := "GetClusterRecord."
	returnError := &pb.ClusterStatus{}
	returnValue := &pb.ClusterStatus{}

	/* VALIDATIONS */
	// Validate Cluster Existance
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}
	// Validate Cluster Cloud Account Permissions
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateCLusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}
	// Validate Cluster is in Actionable State
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return returnError, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	/* GET INFO NEEDED TO UPDATE THE CONTROL PLANE VRESION [Latest available IMI, Latest IMI's Artiface]*/
	// Get Cluster Compatability Info [Provider, Runtime, InstanceType, Os Image]
	var (
		clusterProvider     string
		clusterRuntime      string
		clusterInstanceType string
		clusterOsImage      string
	)
	clusterProvider, _, clusterRuntime, clusterInstanceType, clusterOsImage, err = utils.GetClusterCompatDetails(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterCompatDetails", friendlyMessage)
	}
	// Get the Available Minor Version Upgrades for the cluster
	availableVersions, err := utils.GetAvailableClusterVersionUpgrades(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetAvailableClusterVersionUpgrades", friendlyMessage)
	}
	if len(availableVersions) == 0 {
		return returnError, status.Error(codes.FailedPrecondition, "No Upgrades Available")
	}
	// Set the Query to get IMI for either IKS(Default) or RKE2
	lastK8sVersion := availableVersions[len(availableVersions)-1]
	var query string
	if clusterProvider != "" && clusterProvider == "rke2" {
		query = GetDefaultCpOsImageInstance + " AND ks.k8sversion_name = $5"
		lastK8sVersion = "v" + availableVersions[len(availableVersions)-1]
		if record.K8Sversionname != nil {
			lastK8sVersion = "v" + *record.K8Sversionname
		}
	} else {
		query = GetDefaultCpOsImageInstance + " AND ks.minor_version = $5"
	}
	// Get the IMI for the latest available K8sVersion
	var newCpOsImageInstanceName string
	var newK8sVersionName string
	err = dbconn.QueryRowContext(ctx, query,
		clusterRuntime,
		clusterProvider,
		clusterInstanceType,
		clusterOsImage,
		lastK8sVersion,
	).Scan(&newCpOsImageInstanceName, &newK8sVersionName)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, status.Error(codes.FailedPrecondition, "IMI version not compatible")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultCpOsImageInstance.query", friendlyMessage)
	}
	// Get the CP IMI artifact for the cluster CRD
	var newCpImiArtifact string
	err = dbconn.QueryRowContext(ctx, GetImiArtifactQuery, newCpOsImageInstanceName).Scan(&newCpImiArtifact)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetImiArtifactQuery.osImageInstanceNamw", friendlyMessage)
	}

	defaultAddons, err := utils.GetDefaultAddons(ctx, dbconn, newK8sVersionName)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultAddons", friendlyMessage)
	}

	/* UPDATE THE CLUSTER'S CONTROL PLANE DB ENTRY AND CRD*/
	// Begin the Transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage)
	}
	// Update the Cluster's ControlPlane with new K8sversion
	var nodegroupId int32
	err = tx.QueryRowContext(ctx, updateClusterControlPlaneQuery,
		clusterId,
		newK8sVersionName,
		newCpOsImageInstanceName,
	).Scan(&nodegroupId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateClusterControlPlaneQuery", friendlyMessage)
	}
	// Get current json from REV table
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage)
	}
	// Update Cluster CRD
	clusterCrd.Spec.KubernetesVersion = newK8sVersionName
	clusterCrd.Spec.InstanceIMI = newCpImiArtifact

	/* UPDATE ALL WORKER NODEGROUPS IN THAT CLUSTER*/
	// Get all worker nodegroups for the cluster
	nodeGroupReq := &pb.GetNodeGroupsRequest{
		Clusteruuid:    record.Clusteruuid,
		CloudAccountId: record.CloudAccountId,
	}
	var getNodeGroupsResponse *pb.NodeGroupResponse
	getNodeGroupsResponse, err = GetNodeGroups(ctx, dbconn, nodeGroupReq)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetNodeGroups", friendlyMessage)
	}
	// Update all worker nodes
	for _, dbNodeGroup := range getNodeGroupsResponse.Nodegroups {
		if dbNodeGroup.Nodegroupstate == "Updating" || dbNodeGroup.Nodegroupstate == "Creating" || dbNodeGroup.Nodegroupstate == "Deleting" {
			return returnError, status.Error(codes.FailedPrecondition, "Unable to upgrade cluster. Nodegroups is updating.")
		}
		// Get worker compat details [Cluster Runtime, Cluster Provider, New k8sverion, Nodegroup Os Image, Nodegroup Instane Type]
		ngProvider, _, ngRuntime, _, _, ngInstanceType, ngOsImage, err := utils.GetNodeGroupCompatDetails(ctx, dbconn, dbNodeGroup.Nodegroupuuid)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetNodeGroupCompatDetails", friendlyMessage)
		}
		// Get New Worker IMI for new K8s Version
		var (
			nodeGroupQuery         string
			wrkOsImageInstanceName string
		)
		nodeGroupQuery = GetDefaultWrkOsImageInstance + " AND ks.k8sversion_name = $5"
		err = tx.QueryRowContext(ctx, nodeGroupQuery,
			ngRuntime,
			ngProvider,
			ngInstanceType,
			ngOsImage,
			newK8sVersionName,
		).Scan(&wrkOsImageInstanceName)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			if err == sql.ErrNoRows {
				return returnError, status.Error(codes.FailedPrecondition, "IMI version not compatible")
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"nodeGroupQuery", friendlyMessage)
		}
		// Get New Worker IMI artifact
		var wrkImiArtifact string
		err = tx.QueryRowContext(ctx, GetImiArtifactQuery, wrkOsImageInstanceName).Scan(&wrkImiArtifact)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetImiArtifactQuery.wrkOsImageInstanceNamw", friendlyMessage)
		}
		// Update the DB entry
		_, err = tx.Exec(updateNodegroupImiQuery,
			dbNodeGroup.Nodegroupuuid,
			newK8sVersionName,
			wrkOsImageInstanceName,
		)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateNodegroupImiQuery", friendlyMessage)
		}
		// Update the CRD for nodegroups
		for i, crdNodeGroup := range clusterCrd.Spec.Nodegroups {
			if crdNodeGroup.Name == dbNodeGroup.Nodegroupuuid {
				clusterCrd.Spec.Nodegroups[i].InstanceIMI = wrkImiArtifact
				clusterCrd.Spec.Nodegroups[i].KubernetesVersion = newK8sVersionName
			}
		}

	}
	// Update the CRD for Addons
	var newAddons []clusterv1alpha.AddonTemplateSpec
	for _, addon := range defaultAddons {
		addonCrd := clusterv1alpha.AddonTemplateSpec{
			Name:     addon.Name,
			Type:     clusterv1alpha.AddonType(addon.Type),
			Artifact: addon.Artifact,
		}
		newAddons = append(newAddons, addonCrd)
	}
	clusterCrd.Spec.Addons = newAddons

	/* INSERT THE LATEST CLUSTER REV VERSION */
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
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	}

	/* INSERT NEW PROVISIONING LOG*/
	_, err = tx.QueryContext(ctx, InsertProvisioningQuery,
		clusterId,
		"Ccluster Upgrade Pending",
		"INFO",
		"Cluster Upgrade",
		time.Now(),
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertProvisioningQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, UpdateClusterStateQuery, clusterId, "Pending")
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
	}

	/* COMMIT TRANSACTION */
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}

	/* GET CLUSTER STATUS */
	getClusterStatusReq := &pb.ClusterID{
		Clusteruuid:    record.Clusteruuid,
		CloudAccountId: record.CloudAccountId,
	}
	getClusterStatusResponse, err := GetStatusRecord(ctx, dbconn, getClusterStatusReq)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetStatusRecord", friendlyMessage)
	}
	returnValue = getClusterStatusResponse

	return returnValue, nil
}

func EnableClusterStorage(ctx context.Context, dbconn *sql.DB, record *pb.ClusterStorageRequest) (*pb.ClusterStorageStatus, error) {
	friendlyMessage := "Could not enable Cluster Storage. Please try again"
	failedFunction := "GetClusterRecord."
	returnError := &pb.ClusterStorageStatus{}
	returnValue := &pb.ClusterStorageStatus{}

	/* VALIDATIONS */
	var clusterId int32
	// Validate cluster existance
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}
	// Validate Cluster Cloud Account Permissions
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateCLusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}
	// Validate Cloudaccount storage permission
	storageAllowed, err := utils.ValidateStorageRestrictions(ctx, dbconn, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"Utils.ValidateStorageRestrictions", friendlyMessage)
	}
	if !storageAllowed {
		return returnError, status.Error(codes.PermissionDenied, "Due to storage restrictions, we are currently not allowing non-approved users to use storage.")
	}
	// Validate Cluster is in Actionable State
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return returnError, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	// Validate cluster storage is not already enabled
	isEnabled, err := utils.ValidateStorageEnabled(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateStorageEnabled", friendlyMessage)
	}
	if isEnabled {
		return returnError, status.Error(codes.FailedPrecondition, "Cluster storage is already enabled")
	}

	// Validated request is not "enable: false"
	if !record.Enablestorage {
		return returnError, status.Error(codes.FailedPrecondition, "Cluster Storage temporarily cannot be disabled")
	}

	/* GET THE CLUSER CRD */
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage)
	}

	/* Get the Provider */
	var storageProvider string
	err = dbconn.QueryRowContext(ctx, GetDefaultStorageProvider).Scan(&storageProvider)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultStorageProvider", friendlyMessage)
	}

	/* ADD THE ADDONS */
	addons, err := utils.GetAddons(ctx, dbconn, true, false, clusterCrd.Spec.KubernetesVersion, storageProvider)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"Utils.Addons."+storageProvider, friendlyMessage)
	}

	/* ADD THE VALUES TO THE CRD */
	clusterCrd.Spec.CustomerCloudAccountId = record.CloudAccountId
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

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BegintTx", friendlyMessage)
	}
	/* UPDATE STORAGE TABLE */

	// storageSize := utils.ParseFileSize(record.Storagesize)
	// if storageSize == -1 {
	// 	return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ParseFileSize", "Invalid Storage Size")
	// }

	storageSize := utils.ParseFileSize(record.Storagesize)

	if storageSize == -1 {
		return returnError, errors.New("Invalid Storage Size")
	}

	_, err = tx.ExecContext(ctx, InsertStorageTableQuery, clusterId, storageProvider, "Updating", storageSize)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.Exec.InsertStorageTableQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER TABLE WITH STORAGE INFO*/
	_, err = tx.ExecContext(ctx, UpdateClusterStorageDataQuery, clusterId, record.Enablestorage)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.Exec.UpdateClusterStorageDataQuery", friendlyMessage)
	}

	/* UPDATE CLOUD ACCOUNT TABLE*/
	_, err = tx.ExecContext(ctx, UpdateCloudAccountStorageQuery, record.CloudAccountId, storageSize)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.Exec.UpdateCloudAccountStorageQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER REV TABLE*/
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
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	}
	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, UpdateClusterStateQuery,
		clusterId,
		"Pending",
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
	}

	// Insert Provisioning Log
	_, err = tx.QueryContext(ctx, InsertProvisioningQuery,
		clusterId,
		"Cluster to Pending - Storage",
		"INFO",
		"Cluster to Pending - Storage",
		time.Now(),
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertProvisioningQuery", friendlyMessage)
	}
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}

	/* GET CLUSTER STATUS */
	clusterStatuses, err := utils.GetClusterStorageStatus(ctx, dbconn, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetStatusRecord", friendlyMessage)
	}
	for _, status := range clusterStatuses {
		if status.Storageprovider == storageProvider {
			returnValue = status
		}
	}
	// commit Transaction
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	return returnValue, nil
}

// Update Cluster Storage size in CRD and store in DB
func UpdateClusterStorage(ctx context.Context, dbconn *sql.DB, record *pb.ClusterStorageUpdateRequest) (*pb.ClusterStorageStatus, error) {
	friendlyMessage := "Could not update Cluster Storage. Please try again"
	failedFunction := "UpdateClusterRecord."
	returnError := &pb.ClusterStorageStatus{}
	returnValue := &pb.ClusterStorageStatus{}

	/* VALIDATIONS */
	var clusterId int32
	// Validate cluster existance
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, errors.New("Cluster not found: " + record.Clusteruuid)
	}
	// Validate Cluster Cloud Account Permissions
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateCLusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}
	// Validate Cloudaccount storage permission
	storageAllowed, err := utils.ValidateStorageRestrictions(ctx, dbconn, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"Utils.ValidateStorageRestrictions", friendlyMessage)
	}
	if !storageAllowed {
		return returnError, errors.New("Due to storage restrictions, we are currently not allowing non-approved users to use storage.")
	}
	// Validate Cluster is in Actionable State
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return returnError, errors.New("Cluster not in actionable state")
	}

	// Validate cluster storage is already enabled
	isEnabled, err := utils.ValidateStorageEnabled(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateStorageEnabled", friendlyMessage)
	}
	if !isEnabled {
		return returnError, errors.New("Cluster storage is not already enabled")
	}

	// parse storage size
	storageSize := utils.ParseFileSize(record.Storagesize)

	if storageSize == -1 {
		return returnError, errors.New("Invalid Storage Size")
	}

	/* GET THE CLUSER CRD */
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage)
	}

	// Update the storage size in the Cluster CRD
	clusterCrd.Spec.Storage[0].Size = record.Storagesize

	/* Get the Provider */
	var storageProvider string
	err = dbconn.QueryRowContext(ctx, GetDefaultStorageProvider).Scan(&storageProvider)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultStorageProvider", friendlyMessage)
	}

	/* Get Cluster Stroage Status to check if weka or VAST */
	clusterStatuses, err := utils.GetClusterStorageStatus(ctx, dbconn, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetStatusRecord", friendlyMessage)
	}
	for _, status := range clusterStatuses {
		if status.Storageprovider != storageProvider {
			return returnError, errors.New("Weka Storage is disabled, cannot update Weka Storage. Please create new cluster with VAST storage.")
		}
	}

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BegintTx", friendlyMessage)
	}

	/* UPDATE STORAGE TABLE */
	_, err = tx.ExecContext(ctx, UpdateStorageSizeTableQuery, clusterId, "Updating", storageSize)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.Exec.UpdateStorageSizeTableQuery", friendlyMessage)
	}

	/* UPDATE CLOUD ACCOUNT TABLE*/
	_, err = tx.ExecContext(ctx, UpdateCloudAccountStorageQuery, record.CloudAccountId, storageSize)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.Exec.UpdateCloudAccountStorageQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER REV TABLE*/
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
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	}
	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, UpdateClusterStateQuery,
		clusterId,
		"Pending",
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
	}

	// ///[TODO] Is this required?
	// Insert Provisioning Log
	_, err = tx.QueryContext(ctx, InsertProvisioningQuery,
		clusterId,
		"Cluster to Pending - Storage",
		"INFO",
		"Cluster to Pending - Storage",
		time.Now(),
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertProvisioningQuery", friendlyMessage)
	}

	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}

	/* GET CLUSTER STATUS */
	clusterStatuses, err = utils.GetClusterStorageStatus(ctx, dbconn, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetStatusRecord", friendlyMessage)
	}
	for _, status := range clusterStatuses {
		if status.Storageprovider == storageProvider {
			returnValue = status
		}
	}
	// commit Transaction
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	return returnValue, nil
}

func PutNodeGroup(ctx context.Context, dbconn *sql.DB, record *pb.UpdateNodeGroupRequest) (*pb.Nodegroupstatus, error) {
	friendlyMessage := "Could Update Nodegroup. Please try again"
	failedFunction := "GetClusterRecord."

	returnError := &pb.Nodegroupstatus{}
	returnValue := &pb.Nodegroupstatus{
		Name:          "",
		Clusteruuid:   record.Clusteruuid,
		Nodegroupuuid: record.Nodegroupuuid,
		Count:         0,
		State:         "",
	}

	/* VALIDATE SUPER COMPUTE NODEGROUP TYPE TO MAKE SURE WE ARE UPDATING THE CORRECT GP NODEGROUP TYPE AND ITS INSTANCE TYPE*/
	isSuperComputeCluster, err := utils.ValidateSuperComputeClusterType(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, err
	}
	if isSuperComputeCluster {
		_, err := utils.ValidateSuperComputeGPNodegroupAndInstanceType(ctx, dbconn, record.Clusteruuid, record.Nodegroupuuid)
		if err != nil {
			return returnError, err
		}
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
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClsuterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}

	/* VALIDATE CLUSTER IS ACTIONABLE*/
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return returnError, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	/*Default Values for IKS */
	defaultvalues, err := utils.GetDefaultValues(ctx, dbconn)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultValues", friendlyMessage)
	}

	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, max_nodes, _, _, _, err := convDefaultsToInt(ctx, defaultvalues)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"convDefaultsToInt", friendlyMessage)
	}

	/* GET CURRENT INFORMATION */
	var (
		nodeGroupName        string
		nodeGroupDescription string
		nodeGroupTags        []byte
		nodeGroupAnnotations []byte
		nodeGroupCount       int32
		nodeGroupDrain       bool
		nodeGroupMaxNodes    int32
	)
	err = dbconn.QueryRowContext(ctx, getNodeGroupMetadataQuery,
		record.Nodegroupuuid,
	).Scan(
		&nodeGroupName,
		&nodeGroupDescription,
		&nodeGroupTags,
		&nodeGroupAnnotations,
		&nodeGroupCount,
		&nodeGroupDrain,
		&nodeGroupMaxNodes,
	)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getNodeGoupMetadataQuery", friendlyMessage)
	}

	crdChange := false
	/* CHECK NODEGROUP NAME */
	if record.Name != nil {
		nodeGroupName = *record.Name
	}

	/* CHECK NODEGROUP DESCRIPTION */
	if record.Description != nil {
		nodeGroupDescription = *record.Description
	}

	/* CHECK NODEGROUP COUNT */
	if record.Count != nil && *record.Count != nodeGroupCount {
		if *record.Count > int32(max_nodes) {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"iksnodegroupCount", "Can not create more than 10 nodes for this nodegroup for this cluster")
		}
		nodeGroupCount = *record.Count
		crdChange = true
	}

	/* CHECK NODEGROUP UPGRADE STRATEGY */
	if record.Upgradestrategy != nil && (record.Upgradestrategy.Drainnodes != nodeGroupDrain || record.Upgradestrategy.Maxunavailablepercentage != nodeGroupMaxNodes) {
		nodeGroupDrain = record.Upgradestrategy.Drainnodes
		nodeGroupMaxNodes = record.Upgradestrategy.Maxunavailablepercentage
		crdChange = true
	}
	/* CHECK FOR ANNOTATION CHANGES*/
	dbAnnotations, err := utils.ParseKeyValuePairIntoMap(ctx, nodeGroupAnnotations)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ParseKeyValuePairIntoMap.nodeGroupAnnotations", friendlyMessage)
	}
	recordAnnotations := make(map[string]string)
	for _, annotation := range record.Annotations {
		if recordAnnotations[annotation.Key] != "" {
			return returnError, status.Error(codes.InvalidArgument, "Annotations sent are duplicated. Correct and try again")
		}
		recordAnnotations[annotation.Key] = annotation.Value
	}
	if reflect.DeepEqual(dbAnnotations, recordAnnotations) {
		record.Annotations = nil
	} else {
		crdChange = true
	}

	/* CHECK TAGS */
	dbTags, err := utils.ParseKeyValuePairIntoMap(ctx, nodeGroupTags)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ParseKeyValuePairIntoMap.nodeGroupTags", friendlyMessage)
	}
	recordTags := make(map[string]string)
	for _, tag := range record.Tags {
		if recordTags[tag.Key] != "" {
			return returnError, status.Error(codes.InvalidArgument, "Tags sent are duplicated. Correct and try again")
		}
		recordTags[tag.Key] = tag.Value
	}
	if reflect.DeepEqual(dbTags, recordTags) {
		record.Tags = nil
	}

	/* SET CLUSTER */
	var state string
	err = dbconn.QueryRowContext(ctx, getNodeGroupStateQuery, record.Nodegroupuuid).Scan(&state)
	if crdChange {
		state = "Updating"
	}

	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BegintTx", friendlyMessage)
	}

	/* UPDATE ENTRY IN NODEGROUP TABLE */
	_, err = tx.ExecContext(ctx, updateNodeGroupQuery,
		record.Nodegroupuuid,
		state,
		nodeGroupCount,
		nodeGroupName,
		nodeGroupDescription,
		record.Annotations,
		record.Tags,
		nodeGroupDrain,
		nodeGroupMaxNodes,
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateNodeGroupQuery", friendlyMessage)
	}

	if crdChange {
		/* UPDATE CLUSTER STATE TO PENDING*/
		_, err = tx.QueryContext(ctx, UpdateClusterStateQuery,
			clusterId,
			"Pending",
		)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
		}
		/* UPDATE CLUSTER REV TABLE*/
		// Get current json from REV table
		currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage)
		}

		// Update Nodegroup CRD and append to clusterCrd
		for i, nodeGroup := range clusterCrd.Spec.Nodegroups {
			if nodeGroup.Name == record.Nodegroupuuid {
				// Check if nodegroup is Instance Group and if the Node Count is more than 1
				isInstanceGroup := utils.IsInstanceGroup(clusterCrd.Spec.Nodegroups[i].InstanceType)
				if isInstanceGroup && nodeGroupCount > 1 {
					if errtx := tx.Rollback(); errtx != nil {
						return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
					}
					return returnError, status.Error(codes.InvalidArgument, "Only one instance group is allowed per Nodegroup")
				}

				// Update CRD
				clusterCrd.Spec.Nodegroups[i].Count = int(nodeGroupCount)
				clusterCrd.Spec.Nodegroups[i].UpgradeStrategy.DrainBefore = nodeGroupDrain
				clusterCrd.Spec.Nodegroups[i].UpgradeStrategy.MaxUnavailablePercent = int(nodeGroupMaxNodes)
				clusterCrd.Spec.Nodegroups[i].Annotations = recordAnnotations
			}
		}
		// Create new cluster rev table entry
		_, err = tx.ExecContext(ctx, InsertRevQuery,
			clusterId,
			currentJson,
			clusterCrd,
			"test", // ?? DEFAULT VALUES
			"test", // ?? DEFAULT VALUES
			"test", // ?? DEFAULT VALUES
			"test", // ?? DEFAULT VALUES
			time.Now(),
			false,
		)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
		}
	}

	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}

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
	returnValue = getNodegroupStatusResponse

	return returnValue, nil
}

func UpgradeNodeGroup(ctx context.Context, dbconn *sql.DB, record *pb.NodeGroupid) (*pb.Nodegroupstatus, error) {
	friendlyMessage := "Could not get Cluster. Please try again"
	failedFunction := "GetClusterRecord."
	returnError := &pb.Nodegroupstatus{}
	returnValue := &pb.Nodegroupstatus{}

	/*VALIDATTIONS*/
	// Validate cluster and nodegroup existance
	clusterId, nodeGroupId, err := utils.ValidateNodeGroupExistance(ctx, dbconn, record.Clusteruuid, record.Nodegroupuuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateNodeGroupExistance", friendlyMessage)
	}
	if clusterId == -1 || nodeGroupId == -1 {
		return returnError, status.Errorf(codes.NotFound, "NodeGroup not found in Cluster: %s", record.Clusteruuid)
	}
	// Validate cluster cloud account permissions
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}
	// Validate cluster is actionable
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return returnError, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	/* GET AVAILABLE UPGRADE VERSIONS */
	availableVersionsKeys, availableVersionsImi, err := utils.GetAvailableWorkerImiUpgrades(ctx, dbconn, record.Clusteruuid, record.Nodegroupuuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetAvailableWorkerImiUpgrades", friendlyMessage)
	}
	if len(availableVersionsImi) == 0 {
		return returnError, status.Error(codes.FailedPrecondition, "No Upgrades Available")
	}

	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage)
	}
	_, err = tx.Exec(updateNodegroupImiQuery,
		record.Nodegroupuuid,
		availableVersionsKeys[len(availableVersionsKeys)-1],
		availableVersionsImi[len(availableVersionsImi)-1],
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateNodegroupImiQuery", friendlyMessage)
	}
	var imiartifact string
	/* GET IMI ARTIFACT FOR CLUSTER REV DESIRED JSON */
	err = tx.QueryRowContext(ctx, GetImiArtifactQuery, availableVersionsImi[len(availableVersionsImi)-1]).Scan(&imiartifact)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetImiArtifactQuery", friendlyMessage)
	}
	/* UPDATE CLUSTER REV TABLE*/
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage)
	}
	for i, nodeGroup := range clusterCrd.Spec.Nodegroups {
		if nodeGroup.Name == record.Nodegroupuuid {
			clusterCrd.Spec.Nodegroups[i].InstanceIMI = imiartifact
		}
	}
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
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	}
	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, UpdateClusterStateQuery,
		clusterId,
		"Pending",
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
	}
	/* UPDATE NODEGROUP STATE TO PENDING */
	_, err = tx.QueryContext(ctx, UpdateNodegroupStateQuery,
		record.Clusteruuid,
		record.Nodegroupuuid,
		"Updating",
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateNodegroupStateQuery", friendlyMessage)
	}

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
	returnValue = getNodegroupStatusResponse

	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage)
	}

	return returnValue, nil
}

func UpdateFirewallRule(ctx context.Context, dbconn *sql.DB, record *pb.UpdateFirewallRuleRequest) (*pb.FirewallRuleResponse, error) {
	friendlyMessage := "Could not update security rules. Please try again"
	failedFunction := "UpdateFirewallRule."
	returnError := &pb.FirewallRuleResponse{}
	returnValue := &pb.FirewallRuleResponse{
		Destinationip: "",
		State:         "",
		Sourceip:      []string{},
		Vipid:         0,
		Port:          0,
		Vipname:       "",
		Viptype:       "",
		Protocol:      []string{},
		Internalport:  0,
	}

	/* VALIDATIONS */
	// Validate Cluster Existance
	var cluster_id int32
	cluster_id, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if cluster_id == -1 {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}
	// Validate Cluster Cloud Account Permissions
	isOwner, err := utils.ValidateClusterCloudAccount(ctx, dbconn, record.Clusteruuid, record.CloudAccountId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateCLusterCloudAccount", friendlyMessage)
	}
	if !isOwner {
		return returnError, status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid) // return 404 to avoid leaking cluster existence
	}
	// Validate Cluster is in Actionable State
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage)
	}
	if !actionableState {
		return returnError, status.Error(codes.FailedPrecondition, "Cluster not in actionable state")
	}

	// Get Protocol from default config table
	defaultvalues, err := utils.GetDefaultValues(ctx, dbconn)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultvalues", friendlyMessage)
	}

	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage)
	}

	// Check if vip ip exists for cluster_id
	var vip_id int32
	var viptype string
	err = tx.QueryRowContext(ctx, GetvipidForFirewallQuery, record.Internalip, cluster_id).Scan(&vip_id, &viptype)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetvipidForFirewallQuery", "No Vip ip information found")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetvipidForFirewallQuery", friendlyMessage)
	}

	// Check if port exists for vip_id
	var port int32
	var poolport int32
	var protocol []string
	var vipname string
	err = tx.QueryRowContext(ctx, GetVipPortQuery, vip_id, record.Port).Scan(&port, &poolport, &vipname)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetVipPortQuery", "No vip port information found")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetVipPortQuery", friendlyMessage)
	}

	if record.Protocol != nil {
		protocol = record.Protocol
	} else {
		protocol = []string{defaultvalues["firewall_protocol"]}
	}

	// check if vip is active
	var isActive string
	err = tx.QueryRowContext(ctx, GetVipStatebyVipIdQuery, vip_id).Scan(&isActive)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetVipStatebyVipIdQuery", friendlyMessage)
	}

	if isActive != "Active" {
		return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, failedFunction+"GetVipStatebyVipIdQuery", "Wait for Vip to get active before updating Firewall rule", codes.FailedPrecondition)
	}

	// check count of source ips if greater than 20 not accepted
	max_rules, err := strconv.Atoi(defaultvalues["max_source_ips"])
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"convDefaultsToInt", friendlyMessage)
	}

	if record.Sourceip != nil {
		len_source_ips := len(record.Sourceip)
		if len_source_ips > max_rules {
			return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, failedFunction+"Validate Count Security rules ", "Security rule cannot exceed "+defaultvalues["max_source_ips"], codes.PermissionDenied)
		}
	}

	// Check if firewall rule exists for vip and if yes check the statust o be "Active"
	var firewall_status string
	err = tx.QueryRowContext(ctx, GetFirewallRuleStateQuery, vip_id).Scan(&firewall_status)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetFirewallRuleStateQuery", friendlyMessage)
	}
	if firewall_status == "Reconciling" || firewall_status == "Deleting" {
		return returnError, utils.ErrorHandlerWithGrpcCode(ctx, err, failedFunction+"GetFirewallRuleStateQuery", "Security rule can only be updated for Active Rules", codes.FailedPrecondition)
	}

	// Check if the rule exists already
	var sourceips []byte
	var sourceipsresult []string
	ruleExists, err := utils.ValidateSecRuleExistance(ctx, dbconn, vip_id)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateSecRuleExistance", friendlyMessage)
	}
	if ruleExists {
		/*compare each source ip*/
		err = tx.QueryRowContext(ctx, GetFirewallRuleQuery, cluster_id, vip_id, record.Internalip, record.Port, protocol).Scan(&sourceips)
		if err != sql.ErrNoRows && err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetFirewallRuleQuery", friendlyMessage)
		}
		if sourceips != nil {
			err = json.Unmarshal(sourceips, &sourceipsresult)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetMetadata.unmarshall.souceip", friendlyMessage)
			}
		}
		var items = make(map[string]bool)
		rulefound := true
		existingRulecount := len(sourceipsresult)
		inputRulecount := len(record.Sourceip)

		if existingRulecount != inputRulecount {
			rulefound = false
		} else {
			for _, n := range sourceipsresult {
				items[n] = true
			}
			for _, n := range record.Sourceip {
				if _, found := items[n]; !found {
					rulefound = false
				}
			}
		}
		if rulefound {
			return returnError, nil
		}
	}
	// update vip details table // In future this can be values from user or more than one value for protocol
	_, err = tx.QueryContext(ctx, UpdateVipDetailsQuery,
		protocol,
		vip_id,
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateVipDetailsQuery", friendlyMessage)
	}

	_, err = tx.QueryContext(ctx, PutVipQuery,
		record.Sourceip,
		"Pending",
		vip_id,
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateVipQuery", friendlyMessage)
	}

	returnValue.Sourceip = record.Sourceip
	returnValue.Port = port
	returnValue.Internalport = poolport
	returnValue.Vipid = vip_id
	returnValue.Destinationip = record.Internalip
	returnValue.State = "Pending"
	returnValue.Viptype = viptype
	returnValue.Vipname = vipname
	returnValue.Protocol = protocol

	// fwspec = append(fwspec, firewallspec)

	/* UPDATE CLUSTER REV TABLE*/
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage)
	}

	var fwIndex = -1 //Required when the delete call is made again and vip is not present in CRD
	if clusterCrd.Spec.Firewall != nil {
		for i, fw := range clusterCrd.Spec.Firewall {
			if fw.DestinationIp == record.Internalip && fw.Port == int(port) {
				fwIndex = i
				break
			}
		}
		if fwIndex != -1 {
			clusterCrd.Spec.Firewall[fwIndex].DestinationIp = record.Internalip
			clusterCrd.Spec.Firewall[fwIndex].Port = int(port)
			clusterCrd.Spec.Firewall[fwIndex].Protocol = protocol[0]
			clusterCrd.Spec.Firewall[fwIndex].SourceIps = record.Sourceip
		}
		if fwIndex == -1 {
			firewallspec := clusterv1alpha.FirewallSpec{
				DestinationIp: record.Internalip, //TBD
				Port:          int(port),
				Protocol:      protocol[0], //TBD
				SourceIps:     record.Sourceip,
			}
			clusterCrd.Spec.Firewall = append(clusterCrd.Spec.Firewall, firewallspec)
		}
	} else {
		firewallspec := clusterv1alpha.FirewallSpec{
			DestinationIp: record.Internalip, //TBD
			Port:          int(port),
			Protocol:      protocol[0], //TBD
			SourceIps:     record.Sourceip,
		}
		clusterCrd.Spec.Firewall = append(clusterCrd.Spec.Firewall, firewallspec)
	}

	// Create new cluster rev table entry
	var revversion string
	err = tx.QueryRowContext(ctx, InsertRevQuery,
		cluster_id,
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
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertRevQuery", friendlyMessage)
	}

	/* UPDATE CLUSTER STATE TO PENDING*/
	_, err = tx.QueryContext(ctx, UpdateClusterStateQuery, cluster_id, "Pending")
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"UpdateClusterStateQuery", friendlyMessage)
	}

	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage)
	}

	return returnValue, nil
}
