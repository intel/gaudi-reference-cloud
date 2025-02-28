// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/casbin/casbin/v2"
	"github.com/go-logr/logr"
	config "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Service struct {
	mdb                *manageddb.ManagedDb
	casbinEngine       *casbin.Enforcer
	resourceRepository *ResourceRepository
	auditLogging       *AuditLogging
}

func (svc *Service) Init(ctx context.Context, cfg *config.Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	log.SetDefaultLogger()
	logger = log.FromContext(ctx).WithName("AuthorizationService.Init")
	logger.Info("initializing IDC Authorization Service...")

	// Check if cfg is nil
	if cfg == nil {
		err := errors.New("config is nil")
		logger.Error(err, "config is nil")
		return err
	}

	logger.Info("policies startup sync feature", "enabled", cfg.Features.PoliciesStartupSync)
	logger.Info("audit logging feature", "enabled", cfg.Features.AuditLogging.Enabled)

	config.Cfg = cfg
	if svc.mdb == nil {
		var err error
		svc.mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			logger.Error(err, "error creating manageddb")
			return err
		}
	}
	if err := openDb(ctx, svc.mdb); err != nil {
		return err
	}

	// Audit Logging
	if svc.auditLogging == nil {
		auditLogging, err := NewAuditLogging(db, cfg.Features.AuditLogging.Enabled)
		svc.auditLogging = auditLogging
		if err != nil {
			logger.Error(err, "error initializing audit Logging")
			return err
		}
	}

	// Permission Repository
	cloudAccountRoleRepository, err := NewCloudAccountRoleRepository(db)
	if err != nil {
		logger.Error(err, "error initializing cloud account role repository")
		return err
	}

	// Resources Repository Setup
	if svc.resourceRepository == nil {
		svc.resourceRepository, err = NewResourceRepository(cfg.ResourcesFilePath)
		if err != nil {
			logger.Error(err, "error initializing resources file")
			return err
		}
	}

	// Permission Engine
	permissionEngine, err := NewPermissionEngine(config.Cfg, cloudAccountRoleRepository, svc.resourceRepository, svc.auditLogging)
	if err != nil {
		logger.Error(err, "error initializing permission engine")
		return err
	}

	// Casbin engine Setup
	if svc.casbinEngine == nil {
		svc.casbinEngine, err = NewSyncedEnforcer(svc.mdb, cfg)
		if err != nil {
			logger.Error(err, "error initializing casbin engine")
			return err
		}
	}
	checkPermissionsWithCtx := func(args ...interface{}) (interface{}, error) {
		return permissionEngine.CheckPermissions(ctx, args...)
	}
	svc.casbinEngine.EnableAutoBuildRoleLinks(true)
	svc.casbinEngine.EnableLog(false)
	svc.casbinEngine.AddFunction("checkPermissions", checkPermissionsWithCtx)
	svc.casbinEngine.AddFunction("keyMatchAuthz", KeyMatchAuthzFunc)
	svc.casbinEngine.AddFunction("keyGetAuthz", KeyGetAuthzFunc)

	pb.RegisterAuthzServiceServer(grpcServer, &AuthorizationService{casbinEngine: svc.casbinEngine, permissionEngine: permissionEngine, resourceRepository: svc.resourceRepository, auditLogging: svc.auditLogging})
	reflection.Register(grpcServer)

	if cfg.Features.AuditLogging.CleanupSchedulerEnabled {
		auditLoggingCleanupScheduler := NewAuditLoggingCleanupScheduler(cfg, db)

		startAuditLoggingCleanupScheduler(auditLoggingCleanupScheduler, ctx)
	}

	return nil
}

func (*Service) Name() string {
	return "authz"
}

var db *sql.DB
var logger logr.Logger

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
