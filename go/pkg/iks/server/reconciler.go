// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	"database/sql"
	"fmt"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/reconciler_query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// Server is a cluster
type PrivateReconcilerServer struct {
	pb.UnimplementedIksPrivateReconcilerServer
	session *sql.DB
	cfg     config.Config
}

// NewIksService Initializes DB connection
func NewIksPrivateReconcilerService(session *sql.DB, cfg config.Config) (*PrivateReconcilerServer, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}
	return &PrivateReconcilerServer{
		session: session,
		cfg:     cfg,
	}, nil
}

// Get lastest rev
func (c *PrivateReconcilerServer) GetClustersReconciler(ctx context.Context, req *pb.ClusterReconcilerRequest) (*pb.ClusterReconcilerResponse, error) {
	log := log.FromContext(ctx).WithName("PrivateReconcilerServer.GetClustersReconciler")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &pb.ClusterReconcilerResponse{}, fmt.Errorf("no database connection found")
	}
	res, err := reconciler_query.GetReconcilerClusters(ctx, dbSession, req)
	if err != nil {
		return &pb.ClusterReconcilerResponse{}, err
	}
	return res, nil
}

func (c *PrivateReconcilerServer) PutClusterStateReconciler(ctx context.Context, req *pb.UpdateClusterStateRequest) (*empty.Empty, error) {
	log := log.FromContext(ctx).WithName("PrivateReconcilerServer.PutClusterStateReconciler")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}
	_, err := reconciler_query.PutClusterStateReconcilerQuery(ctx, dbSession, req)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

func (c *PrivateReconcilerServer) PutClusterChangeAppliedReconciler(ctx context.Context, req *pb.UpdateClusterChangeAppliedRequest) (*empty.Empty, error) {
	log := log.FromContext(ctx).WithName("PrivateReconcilerServer.PutClusterChangeAppliedReconciler")
	log.Info("Request", logkeys.Request, req)
	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}
	_, err := reconciler_query.PutClusterChangeAppliedReconcilerQuery(ctx, dbSession, req)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}
func (c *PrivateReconcilerServer) PutClusterStatusReconciler(ctx context.Context, req *pb.UpdateClusterStatusRequest) (*empty.Empty, error) {
	log := log.FromContext(ctx).WithName("PrivateReconcilerServer.PutClusterStatusReconciler")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}
	_, err := reconciler_query.PutClusterStatusReconcilerQuery(ctx, dbSession, req)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

func (c *PrivateReconcilerServer) PutClusterCertsReconciler(ctx context.Context, req *pb.UpdateClusterCertsRequest) (*empty.Empty, error) {
	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}
	_, err := reconciler_query.PutClusterCertsReconcilerQuery(ctx, dbSession, req, c.cfg.EncryptionKeys)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

func (c *PrivateReconcilerServer) DeleteClusterReconciler(ctx context.Context, req *pb.ClusterDeletionRequest) (*empty.Empty, error) {
	log := log.FromContext(ctx).WithName("PrivateReconcilerServer.DeleteClusterReconciler")
	log.Info("Request", logkeys.Request, req)

	dbSession := c.session
	if dbSession == nil {
		return &empty.Empty{}, fmt.Errorf("no database connection found")
	}
	_, err := reconciler_query.DeleteClusterReconcilerQuery(ctx, dbSession, req)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}
