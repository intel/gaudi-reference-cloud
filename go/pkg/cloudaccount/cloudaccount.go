// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"database/sql"
	"embed"

	authz "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Service struct {
	mdb *manageddb.ManagedDb
}

func (svc *Service) Init(ctx context.Context, cfg *config.Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	config.Cfg = cfg
	if svc.mdb == nil {
		var err error
		svc.mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			return err
		}
	}
	if err := openDb(ctx, svc.mdb); err != nil {
		return err
	}

	logger := log.FromContext(ctx).WithName("Service.Init")
	logger.Info("service initialization")
	notificationClient, err := NewNotificationClient(ctx, resolver, cfg)
	if err != nil {
		logger.Error(err, "failed to initialize notification client")
		return err
	}

	var authzClient *authz.AuthzClient
	if cfg.Authz.Enabled {
		authzClient, err = authz.NewAuthzClient(ctx, resolver)
		if err != nil {
			logger.Error(err, "failed to initialize authz client")
			return err
		}
	}

	userCredentialsClient, err := NewUserCredentialsClient(ctx, resolver, cfg)
	if err != nil {
		logger.Error(err, "failed to initialize user credentials client")
		return err
	}

	pb.RegisterCloudAccountServiceServer(grpcServer, &CloudAccountService{authzClient: authzClient, cfg: cfg})
	pb.RegisterCloudAccountMemberServiceServer(grpcServer, &CloudAccountMemberService{})
	pb.RegisterOtpServiceServer(grpcServer, &OtpService{notificationClient: notificationClient})
	pb.RegisterCloudAccountInvitationServiceServer(grpcServer, &InvitationService{session: db, notificationClient: notificationClient, authzClient: authzClient, userCredentialsClient: userCredentialsClient, cfg: cfg})
	pb.RegisterCloudAccountInvitationMemberServiceServer(grpcServer, &InvitationMemberService{session: db})
	pb.RegisterCloudAccountAppClientServiceServer(grpcServer, &CloudAccountAppClientService{})
	reflection.Register(grpcServer)

	if cfg.RunSchedulers && cfg.Features.InvitationsExpiryScheduler {
		invitationExpiryScheduler := NewInvitationsExpiryScheduler(cfg, db, notificationClient)

		startInvitationsExpiryScheduler(invitationExpiryScheduler, ctx)
	}

	return nil
}

func (*Service) Name() string {
	return "cloudaccount"
}

var db *sql.DB

//go:embed sql/*.sql
var fs embed.FS

func openDb(ctx context.Context, mdb *manageddb.ManagedDb) error {
	log := log.FromContext(ctx)
	if err := mdb.Migrate(ctx, fs, "sql"); err != nil {
		log.Error(err, "migrate:")
		return err
	}

	var err error
	db, err = mdb.Open(ctx)
	if err != nil {
		log.Error(err, "mdb.Open failed")
		return err
	}
	return nil
}
