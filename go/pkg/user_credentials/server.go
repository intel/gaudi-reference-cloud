// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package user_credentials

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cognitoutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/user_credentials/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Service struct {
	dbClient *sql.DB
	mdb      *manageddb.ManagedDb
}

func (svc *Service) Init(ctx context.Context, cfg *config.Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	logger := log.FromContext(ctx).WithName("Service.Init")
	logger.Info("service initialization")
	config.Cfg = cfg
	var err error
	if svc.mdb == nil {
		svc.mdb, err = manageddb.New(ctx, &cfg.CloudAccountDatabase)
		if err != nil {
			logger.Error(err, "manageddb.New failed")
			return err
		}
	}
	svc.dbClient, err = openDb(ctx, svc.mdb)
	if err != nil {
		logger.Error(err, "openDb failed")
		return err
	}
	awsCognitoRegion := config.Cfg.GetAWSCognitoRegion()
	awsCredentialsFile := config.Cfg.GetAWSCredentialsFile()
	awsUserPool := config.Cfg.GetAWSUserPool()
	logger.Info("AWS COGNITO", "awsCognitoRegion", awsCognitoRegion, "awsCredentialsFile", awsCredentialsFile, "awsUserPool", awsUserPool)
	cognitoUtil := cognitoutil.COGNITOUtil{}
	if err := cognitoUtil.Init(ctx, awsCognitoRegion, awsCredentialsFile, awsUserPool); err != nil {
		logger.Error(err, "couldn't init cognitoUtil client")
	}
	// Create cloudaccount client
	cloudAcctClient, err := initCloudAccountClient(ctx)
	if err != nil {
		logger.Error(err, "couldn't  init CloudAccountClient")
		return err
	}

	credentialsService, err := NewUserCredentialsService(&cognitoUtil, svc.dbClient, config.Cfg.GetCustomScope(), cloudAcctClient)
	if err != nil {
		logger.Error(err, "couldn't  start initService")
		return err
	}
	cloudAccountLocks = NewCloudAccountLocks(ctx)
	pb.RegisterUserCredentialsServiceServer(grpcServer, credentialsService)
	reflection.Register(grpcServer)
	return nil
}

func (*Service) Name() string {
	return "user-credentials"
}

func initCloudAccountClient(ctx context.Context) (pb.CloudAccountServiceClient, error) {
	resolver := &grpcutil.DnsResolver{}
	addr, err := resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		return nil, fmt.Errorf("error resolving cloudaccount client: %v", err)
	}

	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("error creating grpc client: %v", err)
	}
	return pb.NewCloudAccountServiceClient(conn), nil
}

func openDb(ctx context.Context, mdb *manageddb.ManagedDb) (*sql.DB, error) {
	log := log.FromContext(ctx)

	var err error
	db, err := mdb.Open(ctx)
	if err != nil {
		log.Error(err, "mdb.Open failed")
		return nil, err
	}
	return db, nil
}
