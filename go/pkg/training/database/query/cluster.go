// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	createClusterState = `
		INSERT INTO cluster (cluster_id, cloud_account_id, name, description, ssh_key_names, status) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	addClusterNodes = `
		INSERT INTO node (node_id, cloud_account_id, node_role, labels, machine_type, image, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	addClusterSpec = `
        INSERT INTO vnet_spec (cloud_account_id,cluster_id, region, availabilityZone, prefixLength) 
        VALUES ($1, $2, $3, $4, $5)
    `

	addClusterStorageNodes = `
		INSERT INTO storage (fs_resource_id, cloud_account_id, name, description, capacity, access, mount, local_mount_dir, remote_mount_dir)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
		WHERE NOT EXISTS (
			SELECT 1 FROM storage
			WHERE name = $3
		)
	`

	addClusterNodeMapping = `
		WITH nodeid AS (
			SELECT id 
			FROM node 
			WHERE node_id = $1
		),
		clusterid AS (
			SELECT id 
			FROM cluster
			WHERE cluster_id = $2
		)
		INSERT INTO cluster_node_mapping(node_id, cluster_id) 
		VALUES ((SELECT id from nodeid), (SELECT id from clusterid))
	`

	addClusterStorageNodeMapping = `
		WITH fsresourceid AS (
			SELECT id
			FROM storage
			WHERE fs_resource_id = $1
		),
		clusterid AS (
			SELECT id
			FROM cluster
			WHERE cluster_id = $2
		)
		INSERT INTO cluster_storage_mapping(fs_resource_id, cluster_id)
		VALUES ((SELECT id from fsresourceid), (SELECT id from clusterid))
	`

	updateClusterState = `
		UPDATE cluster SET status = $1, updated_at = $2
		WHERE cluster_id = $3 AND cloud_account_id = $4
	`

	getClusterVnetSpec = `
		SELECT vs.region, vs.availabilityZone, vs.prefixLength
		FROM vnet_spec as vs, cluster as c
		WHERE vs.cluster_id = c.cluster_id AND
		c.cluster_id = $1
	`

	getClusterNodes = `
		SELECT n.node_id, n.name, n.node_role, n.labels, n.machine_type, n.image, n.status
		FROM node as n, cluster_node_mapping as cn, cluster as c
		WHERE n.id = cn.node_id AND
		cn.cluster_id = c.id AND
		c.cluster_id = $1
	`

	getClusterStorageNodes = `
		SELECT s.fs_resource_id, s.name, s.description, s.capacity, s.access, s.mount, s.local_mount_dir, s.remote_mount_dir
		FROM storage as s, cluster_storage_mapping as cs, cluster as c
		WHERE s.id = cs.fs_resource_id AND
		cs.cluster_id = c.id AND
		c.cluster_id = $1
	`

	getClusterNodesWithCount = `
		SELECT n.node_id, n.name, n.node_role, n.labels, n.machine_type, n.image, n.status, COUNT(*) as node_count
		FROM node as n, cluster_node_mapping as cn, cluster as c
		WHERE n.id = cn.node_id AND
		cn.cluster_id = c.id AND
		c.cluster_id = $1
		GROUP BY n.node_id, n.name, n.node_role, n.labels, n.machine_type, n.image, n.status
	`

	getClusterByCloudAccountId = `
		SELECT name, cluster_id, description
		FROM cluster
		WHERE cloud_account_id = $1
	`

	getClusterById = `
		SELECT name, description, ssh_key_names
		FROM cluster
		WHERE cluster_id = $1 AND cloud_account_id = $2
	`

	getNextClusterRequest = `
		SELECT c.cluster_id, c.cloud_account_id, c.name, c.ssh_key_names, c.requested_at, c.updated_at
		FROM cluster as c
		WHERE  c.status = 'REQUESTED'
		LIMIT 1;
	`
)

const (
	STATE_REQUESTED    = "REQUESTED"
	STATE_PROVISIONING = "PROVISIONING"
	STATE_FAILED       = "FAILED"
	STATE_READY        = "READY"
	STATE_SLURM        = "SLURMSTATIC"
)

func CreateClusterState(ctx context.Context, dbconn *sql.DB, clusterId string, cluster *v1.SlurmClusterCreateRequest) error {
	logger := log.FromContext(ctx).WithName("CreateClusterState")
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "error begin transaction data")
		return fmt.Errorf("error storing cluster request. Error beginning database transaction.")
	}
	defer tx.Rollback()

	sshKeys := cluster.Cluster.GetSSHKeyName()
	serializedSshKeys, err := json.Marshal(sshKeys)
	if err != nil {
		return fmt.Errorf("error marshaling sshKeys")
	}
	serializedSshKeys, err = json.MarshalIndent(sshKeys, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshal indenting sshKeys")
	}

	_, err = tx.ExecContext(ctx, createClusterState, clusterId, cluster.GetCloudAccountId(), cluster.Cluster.GetName(),
		cluster.Cluster.GetDescription(), string(serializedSshKeys), STATE_REQUESTED)
	logger.Info(string(serializedSshKeys))
	if err != nil {
		logger.Error(err, "error inserting cluster state")
		return fmt.Errorf("error storing cluster request")
	}

	spec := cluster.Cluster.GetSpec()

	_, err = tx.ExecContext(ctx, addClusterSpec, cluster.GetCloudAccountId(), clusterId, spec.GetRegion(), spec.GetAvailabilityZone(), spec.GetPrefixLength())

	if err != nil {
		logger.Error(err, "error inserting vnet spec")
		return fmt.Errorf("error storing cluster request for vnet spec")
	}

	for _, node := range cluster.Cluster.Nodes {
		for count := 0; count < int(node.Count); count++ {
			nodeId := uuid.NewString()

			nodeLabels := node.GetLabels()
			serializedNodeLabels, err := json.Marshal(node.GetLabels())
			if err != nil {
				return fmt.Errorf("error marshaling node label")
			}
			serializedNodeLabels, err = json.MarshalIndent(nodeLabels, "", "  ")
			if err != nil {
				return fmt.Errorf("error marshal indenting node labels")
			}

			_, err = tx.ExecContext(ctx, addClusterNodes, nodeId, cluster.GetCloudAccountId(),
				mapNodeRole(node.GetRole()), string(serializedNodeLabels),
				mapMachineType(node.GetMachineType()), node.GetImageType(), STATE_REQUESTED)

			if err != nil {
				logger.Error(err, "error inserting node state")
				return fmt.Errorf("error storing cluster request for nodes")
			}

			_, err = tx.ExecContext(ctx, addClusterNodeMapping, nodeId, clusterId)
			if err != nil {
				logger.Error(err, "error inserting cluster-node mapping")
				return fmt.Errorf("error storing cluster request for nodes map")
			}
		}
	}

	for _, fs := range cluster.Cluster.StorageNodes {
		fsResourceId := uuid.NewString()

		duplicateNameResults, err := tx.ExecContext(ctx, checkDuplicateStorageName, fs.GetName(), cluster.GetCloudAccountId())
		if err != nil {
			return err
		}
		rowsAffected, err := duplicateNameResults.RowsAffected()
		if rowsAffected > 0 {
			return fmt.Errorf("error duplicate storage names are not allowed")
		}

		_, err = tx.ExecContext(ctx, addClusterStorageNodes, fsResourceId, cluster.GetCloudAccountId(), fs.GetName(), fs.GetDescription(), fs.GetCapacity(),
			mapAccessModeType(fs.GetAccessMode()), mapMountType(fs.GetMount()), fs.GetLocalMountDir(), fs.GetRemoteMountDir())

		if err != nil {
			logger.Error(err, "error inserting storage node")
			return fmt.Errorf("error storing cluster request for storage nodes")
		}

		_, err = tx.ExecContext(ctx, addClusterStorageNodeMapping, fsResourceId, clusterId)
		if err != nil {
			logger.Error(err, "error inserting cluster-storage-node mapping")
			return fmt.Errorf("error storing cluster request for storage nodes map")
		}
	}
	// tx.Commit()
	err = tx.Commit()
	if err != nil {
		logger.Error(err, "error committing transaction")
		return err
	}
	return nil
}

func mapAccessModeType(accessMode pb.StorageAccessModeType) string {
	switch accessMode {
	case pb.StorageAccessModeType_STORAGE_READ_WRITE:
		return "STORAGE_READ_WRITE"
	case pb.StorageAccessModeType_STORAGE_READ_ONLY:
		return "STORAGE_READ_ONLY"
	case pb.StorageAccessModeType_STORAGE_READ_WRITE_ONCE:
		return "STORAGE_READ_WRITE_ONCE"
	default:
		return "UNKNOWN"
	}
}

func mapMountType(mountType pb.StorageMountType) string {
	switch mountType {
	case pb.StorageMountType_STORAGE_WEKA:
		return "STORAGE_WEKA"
	default:
		return "UNKNOWN"
	}
}

func marshalAccessModeToPb(accessMode string) pb.StorageAccessModeType {
	switch accessMode {
	case "STORAGE_READ_WRITE":
		return pb.StorageAccessModeType_STORAGE_READ_WRITE
	case "STORAGE_READ_ONLY":
		return pb.StorageAccessModeType_STORAGE_READ_ONLY
	case "STORAGE_READ_WRITE_ONCE":
		return pb.StorageAccessModeType_STORAGE_READ_WRITE_ONCE
	default:
		return pb.StorageAccessModeType_STORAGE_ACCESS_UNKNOWN
	}
}

func marshalMountProtocolToPb(mount string) pb.StorageMountType {
	switch mount {
	case "STORAGE_WEKA":
		return pb.StorageMountType_STORAGE_WEKA
	default:
		return pb.StorageMountType_STORAGE_MOUNT_UNKNOWN
	}
}

func mapNodeRole(nodeRole pb.NodeRole) string {
	switch nodeRole {
	case pb.NodeRole_JUPYTERHUB_NODE:
		return "JUPYTERHUB_NODE"
	case pb.NodeRole_COMPUTE_NODE:
		return "SLURM_COMPUTE_NODE"
	case pb.NodeRole_CONTROLLER_NODE:
		return "SLURM_CONTROLLER_NODE"
	case pb.NodeRole_LOGIN_NODE:
		return "LOGIN_NODE"
	default:
		return "UNKNOWN"
	}
}

func marshalRoleToPb(role string) pb.NodeRole {
	switch role {
	case "JUPYTERHUB_NODE":
		return pb.NodeRole_JUPYTERHUB_NODE
	case "SLURM_COMPUTE_NODE":
		return pb.NodeRole_COMPUTE_NODE
	case "SLURM_CONTROLLER_NODE":
		return pb.NodeRole_CONTROLLER_NODE
	case "LOGIN_NODE":
		return pb.NodeRole_LOGIN_NODE
	default:
		return pb.NodeRole_UNKNOWN_ROLE
	}
}

func marshalMachineTypeToPb(machineType string) pb.MachineType {
	switch machineType {
	case "SML_VM_TYPE":
		return pb.MachineType_SML_VM_TYPE
	case "MED_VM_TYPE":
		return pb.MachineType_MED_VM_TYPE
	case "LRG_VM_TYPE":
		return pb.MachineType_LRG_VM_TYPE
	case "PVC_BM_1100_4":
		return pb.MachineType_PVC_BM_1100_4
	case "PVC_BM_1100_8":
		return pb.MachineType_PVC_BM_1100_8
	case "PVC_BM_1550_8":
		return pb.MachineType_PVC_BM_1550_8
	case "GAUDI_BM_TYPE":
		return pb.MachineType_GAUDI_BM_TYPE
	default:
		return pb.MachineType_UNKNOWN_TYPE
	}
}

func mapMachineType(macType pb.MachineType) string {
	switch macType {
	case pb.MachineType_GAUDI_BM_TYPE:
		return "GAUDI_BM_TYPE"
	case pb.MachineType_LRG_VM_TYPE:
		return "LRG_VM_TYPE"
	case pb.MachineType_MED_VM_TYPE:
		return "MED_VM_TYPE"
	case pb.MachineType_PVC_BM_1100_4:
		return "PVC_BM_1100_4"
	case pb.MachineType_PVC_BM_1100_8:
		return "PVC_BM_1100_8"
	case pb.MachineType_PVC_BM_1550_8:
		return "PVC_BM_1550_8"
	case pb.MachineType_SML_VM_TYPE:
		return "SML_VM_TYPE"
	default:
		return "UNKNOWN"
	}
}

func UpdateClusterState(ctx context.Context, dbconn *sql.DB, clusterId, accountId, status string) error {
	currTime := time.Now()
	logger := log.FromContext(ctx).WithName("UpdateClusterState")
	_, err := dbconn.ExecContext(ctx, updateClusterState, status, currTime, clusterId, accountId)
	if err != nil {
		logger.Error(err, "error updating cluster state")
		return errors.New("record update failed")
	}
	return nil
}

func GetNextClusterRequest(ctx context.Context, dbconn *sql.DB) (*v1.Cluster, error) {
	logger := log.FromContext(ctx).WithName("GetNextClusterRequest")
	cl := v1.Cluster{}

	row := dbconn.QueryRowContext(ctx, getNextClusterRequest)
	reqTime := time.Time{}
	updateTime := time.Time{}
	sshKeyNames := json.RawMessage{}

	switch err := row.Scan(&cl.ClusterId, &cl.CloudAccountId, &cl.Name, &sshKeyNames, &reqTime, &updateTime); err {
	case sql.ErrNoRows:
		logger.Info("no records found ")
		return nil, nil
	case nil:
		logger.Error(err, "error querying cluster state")
		return nil, fmt.Errorf("error querying cluster request")
	default:
		if err := json.Unmarshal(sshKeyNames, &cl.SSHKeyName); err != nil {
			logger.Error(err, "error unmarshaling cluster state sshKeyName")
			return nil, fmt.Errorf("error getting cluster sshKeyName")
		}
	}

	specRows, err := dbconn.QueryContext(ctx, getClusterVnetSpec, cl.ClusterId)
	if err != nil {
		logger.Error(err, "error reading vnet spec")
		return nil, errors.New("error querying vnet spec data")
	}
	for specRows.Next() {
		spec := pb.VNetSpec{}

		if err := specRows.Scan(&spec.Region, &spec.AvailabilityZone, &spec.PrefixLength); err != nil {
			logger.Error(err, "error scaning vnet spec record in db")
			return nil, errors.New("record search failed")
		}
		cl.Spec = &spec
	}

	nodeRows, err := dbconn.QueryContext(ctx, getClusterNodes, cl.ClusterId)
	if err != nil {
		logger.Error(err, "error reading nodes")
		return nil, errors.New("error querying nodes data")
	}

	for nodeRows.Next() {
		node := pb.ClusterNode{}
		var role, machineType, status string
		labels := json.RawMessage{}
		if err := nodeRows.Scan(&node.NodeId, &node.Name, &role, &labels, &machineType, &node.ImageType, &status); err != nil {
			logger.Error(err, "error scaning node record in db")
			return nil, errors.New("record search failed")
		}
		node.Role = marshalRoleToPb(role)
		node.MachineType = marshalMachineTypeToPb(machineType)
		cl.Nodes = append(cl.Nodes, &node)
	}

	storageRows, err := dbconn.QueryContext(ctx, getClusterStorageNodes, cl.ClusterId)
	if err != nil {
		logger.Error(err, "error reading storage nodes")
		return nil, errors.New("error querying storage nodes data")
	}

	for storageRows.Next() {
		storageFS := pb.StorageNode{}
		var access, mount string
		if err := storageRows.Scan(&storageFS.FsResourceId, &storageFS.Name, &storageFS.Description, &storageFS.Capacity, &access, &mount, &storageFS.LocalMountDir, &storageFS.RemoteMountDir); err != nil {
			logger.Error(err, "error scanning storage node record in database")
			return nil, errors.New("record search failed")
		}

		storageFS.AccessMode = marshalAccessModeToPb(access)
		storageFS.Mount = marshalMountProtocolToPb(mount)
		cl.StorageNodes = append(cl.StorageNodes, &storageFS)
	}

	return &cl, nil
}

func GetClusterByID(ctx context.Context, dbconn *sql.DB, req *v1.SlurmClusterRequest) (*v1.Cluster, error) {
	logger := log.FromContext(ctx).WithName("GetClusterByID")
	cl := v1.Cluster{}
	sshKeyNames := json.RawMessage{}

	row := dbconn.QueryRowContext(ctx, getClusterById, req.GetClusterId(), req.GetCloudAccountId())

	switch err := row.Scan(&cl.Name, &cl.Description, &sshKeyNames); {
	case err == sql.ErrNoRows:
		logger.Info("no records found", "clusterId and cloudAccountId", req)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case err != nil:
		logger.Error(err, "error querying cluster by IDs")
		return nil, status.Errorf(codes.Internal, "error querying cluster request by cluster and cloudAccount ID")
	default:
		if err := json.Unmarshal(sshKeyNames, &cl.SSHKeyName); err != nil {
			logger.Error(err, "error unmarshaling cluster sshKeyName")
			return nil, status.Errorf(codes.Internal, "error unmarshaling cluster sshKeyName")
		}

		cl.CloudAccountId = req.GetCloudAccountId()
		cl.ClusterId = req.GetClusterId()
	}

	specRows, err := dbconn.QueryContext(ctx, getClusterVnetSpec, cl.ClusterId)
	if err != nil {
		logger.Error(err, "error reading vnet spec")
		return nil, errors.New("error querying vnet spec data")
	}
	for specRows.Next() {
		spec := pb.VNetSpec{}

		if err := specRows.Scan(&spec.Region, &spec.AvailabilityZone, &spec.PrefixLength); err != nil {
			logger.Error(err, "error scaning vnet spec record in db")
			return nil, errors.New("record search failed")
		}
		cl.Spec = &spec
	}

	nodeRows, err := dbconn.QueryContext(ctx, getClusterNodesWithCount, req.GetClusterId())
	if err != nil {
		logger.Error(err, "error reading cluster nodes")
		return nil, status.Errorf(codes.Internal, "error querying cluster nodes data")
	}

	for nodeRows.Next() {
		node := pb.ClusterNode{}
		var role, machineType, nodeStatus string
		labels := json.RawMessage{}

		switch err := nodeRows.Scan(&node.NodeId, &node.Name, &role, &labels, &machineType, &node.ImageType, &nodeStatus, &node.Count); {
		case err == sql.ErrNoRows:
			logger.Info("no records found", "cluster nodes", req.GetClusterId())
			return nil, status.Errorf(codes.NotFound, "no records found for cluster nodes")
		case err != nil:
			logger.Error(err, "error scaning node record in database")
			return nil, status.Errorf(codes.Internal, "error querying node record in database")
		default:
			if err := json.Unmarshal(labels, &node.Labels); err != nil {
				logger.Error(err, "error unmarshaling node cluster labels")
				return nil, status.Errorf(codes.Internal, "error unmarshaling node cluster labels")
			}
		}

		node.Role = marshalRoleToPb(role)
		node.MachineType = marshalMachineTypeToPb(machineType)
		cl.Nodes = append(cl.Nodes, &node)
	}

	storageRows, err := dbconn.QueryContext(ctx, getClusterStorageNodes, cl.ClusterId)
	if err != nil {
		logger.Error(err, "error reading storage nodes")
		return nil, errors.New("error querying storage nodes data")
	}

	for storageRows.Next() {
		storageFS := pb.StorageNode{}
		var access, mount string
		if err := storageRows.Scan(&storageFS.FsResourceId, &storageFS.Name, &storageFS.Description, &storageFS.Capacity, &access, &mount, &storageFS.LocalMountDir, &storageFS.RemoteMountDir); err != nil {
			logger.Error(err, "error scanning storage node record in database")
			return nil, errors.New("record search failed")
		}

		storageFS.AccessMode = marshalAccessModeToPb(access)
		storageFS.Mount = marshalMountProtocolToPb(mount)
		cl.StorageNodes = append(cl.StorageNodes, &storageFS)
	}

	return &cl, nil
}

func GetClustersByCloudAccount(ctx context.Context, dbconn *sql.DB, req *v1.ClusterListOption) (*v1.SlurmClusterResponse, error) {
	logger := log.FromContext(ctx).WithName("GetClustersByCloudAccount")
	clusters := v1.SlurmClusterResponse{}

	rows, err := dbconn.QueryContext(ctx, getClusterByCloudAccountId, req.GetCloudAccountId())
	if err != nil {
		logger.Error(err, "error reading cluster with cloudAccountId")
		return nil, status.Errorf(codes.Internal, "error reading cluster with cloudAccountId")
	}

	for rows.Next() {
		cl := v1.ClusterInfo{}

		switch err := rows.Scan(&cl.Name, &cl.ClusterId, &cl.Description); {
		case err == sql.ErrNoRows:
			logger.Info("no records found", "cloudAccountId", req)
			return nil, status.Errorf(codes.NotFound, "no matching records found")
		case err != nil:
			logger.Error(err, "error querying cluster by cloud account")
			return nil, status.Errorf(codes.Internal, "error querying cluster by cloudAccountId")
		default:
			cl.CloudAccountId = req.GetCloudAccountId()
			clusters.Clusters = append(clusters.Clusters, &cl)
		}
	}

	return &clusters, nil
}
