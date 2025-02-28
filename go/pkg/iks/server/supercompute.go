// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"
	mathrand "math/rand"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/config"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	supercompute_query "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/supercompute_query"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type SuperComputeServer struct {
	pb.UnimplementedIksSuperComputeServer
	computeClient pb.InstanceTypeServiceClient
	sshkeyClient  pb.SshPublicKeyServiceClient
	vnetClient    pb.VNetServiceClient
	session       *sql.DB
	cfg           config.Config
}

// NewIksSuperComputeService Initializes DB connection
func NewIksSuperComputeService(session *sql.DB, computegrpcClient pb.InstanceTypeServiceClient, sshkeyClient pb.SshPublicKeyServiceClient, vnetClient pb.VNetServiceClient, cfg config.Config) (*SuperComputeServer, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}
	return &SuperComputeServer{
		session:       session,
		cfg:           cfg,
		computeClient: computegrpcClient,
		sshkeyClient:  sshkeyClient,
		vnetClient:    vnetClient,
	}, nil
}

// CreateNewSuperCluster will create new cluster entries
func (c *SuperComputeServer) SuperComputeCreateCluster(ctx context.Context, req *pb.SuperComputeClusterCreateRequest) (*pb.ClusterCreateResponseForm, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("IKS.SuperComputeCreateCluster").WithValues("name", req.Clusterspec.Name).Start()
	defer span.End()

	logger.Info("Request", "req", req)

	dbSession := c.session
	if dbSession == nil {
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("no database connection found")
	}
	err := req.ValidateAll()
	if err != nil {
		logger.Info("Error in validating grpc Supercompute Cluster Request", "error", err)
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("Supercompute Cluster Create Request validation failed. Cluster name accepts lower or upper case alphanumeric chars and dash,hyphens and can not exceed 63 bytes. Cluster description accepts zero or more lower or upper case alphanumeric chars and dash, hyphens, spaces and can not exceed 253 bytes.")
	}

	// Get default cloudaccount
	defaultvalues, err := utils.GetDefaultValues(ctx, dbSession)
	if err != nil {
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("get default cloud account failed %s ", err)
	}

	// Validate Super Compute Cluster Constraints Spec
	err = utils.ValidateSuperComputeRequestSpec(ctx, dbSession, req)
	if err != nil {
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("Validate Super compute spec failed: %s ", err)
	}

	// Generate SSH keys and upload to compute
	// generateSshKeyForCP()
	key, keypub, err := generateSshKeyForCP(ctx)
	if err != nil {
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("Ssh key generation for control plane failed %s ", err)
	}

	sshpubkey, err := c.UploadSshkeysForCP(ctx, keypub, defaultvalues["cp_cloudaccountid"], req.Clusterspec.Name)
	if err != nil {
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("Ssh key upload for control plane failed %s ", err)
	}

	// Start the transaction
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return &pb.ClusterCreateResponseForm{}, utils.ErrorHandler(ctx, err, "dbconn.BeginTx", "Error while initializing the DB transaction")
	}

	// Insert Cluster Record in Cluster Table and Get the Cluster Rev
	clusterResp, clusterCrd, clusterId, err := supercompute_query.CreateSuperComputeClusterAndNodegroupCRD(ctx, dbSession, tx, req, key, keypub, sshpubkey, c.cfg.EncryptionKeys, c.computeClient, c.vnetClient)
	if err != nil {
		return &pb.ClusterCreateResponseForm{}, err
	}

	// Insert into Cluster Rev Table
	if clusterId > 0 && clusterResp.Uuid != "" {
		err = supercompute_query.InsertClusterRevTable(ctx, dbSession, tx, clusterCrd, clusterResp.Uuid, clusterId)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return &pb.ClusterCreateResponseForm{}, utils.ErrorHandler(ctx, err, "InsertClusterRevTable"+"TransactionRollbackError", "InsertClusterRevTable Error")
			}
			return &pb.ClusterCreateResponseForm{}, err
		}
	} else {
		return &pb.ClusterCreateResponseForm{}, fmt.Errorf("Cluster ID or Cluster UUID is empty")
	}

	return clusterResp, nil
}

func (c *SuperComputeServer) UploadSshkeysForCP(ctx context.Context, pubkey []byte, cloudaccount string, clustername string) (string, error) {
	_, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("SshKeys").
		WithValues("clustername", clustername).Start()
	defer span.End()
	const charset = "abcdefgh12345"

	var seededRand *mathrand.Rand = mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	rndb := make([]byte, 4)

	for i := range rndb {
		rndb[i] = charset[seededRand.Intn(len(charset))]
	}

	var keyname string
	if len(clustername) > 13 {
		keyname = clustername[0:12] + "-" + cloudaccount + "-" + string(rndb)
	} else {
		keyname = clustername + "-" + cloudaccount + "-" + string(rndb)
	}

	getsshkeys, err := c.sshkeyClient.Search(ctx, &pb.SshPublicKeySearchRequest{
		Metadata: &pb.ResourceMetadataSearch{
			CloudAccountId: cloudaccount,
		},
	})
	if err != nil {
		return keyname, err
	}
	if len(getsshkeys.Items) != 0 {
		for i, _ := range getsshkeys.Items {
			if getsshkeys.Items[i].Metadata.Name == keyname {
				log.Info("Ssh key with the name ", keyname, "already exists.Creating cp with existing ssh key")
				return keyname, nil
			}
			continue
		}
	}
	sshkey, err := c.sshkeyClient.Create(ctx, &pb.SshPublicKeyCreateRequest{
		Metadata: &pb.ResourceMetadataCreate{
			CloudAccountId: cloudaccount,
			Name:           keyname,
		},
		Spec: &pb.SshPublicKeySpec{
			SshPublicKey: string(pubkey),
		},
	})

	if err != nil {
		log.Error(err, "\n .. Upload ssh key failed")
		return "", err
	}

	return sshkey.Metadata.Name, nil
}
