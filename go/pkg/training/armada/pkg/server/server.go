// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	provisionConfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/armada/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/armada/pkg/provider"
)

type ClusterService struct {
	ManagedDb *manageddb.ManagedDb
}

func (svc *ClusterService) Init(ctx context.Context, cfg *provisionConfig.Config) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	log.Info("initializing IDC cluster provisioning...")

	if svc.ManagedDb == nil {
		var err error
		svc.ManagedDb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			log.Error(err, "error connecting to database")
			return err
		}
	}

	log.Info("successfully connected to the database")

	db, err := svc.ManagedDb.Open(ctx)
	if err != nil {
		log.Error(err, "error opening database connection")
		return err
	}
	log.Info("successfully opened database connection")

	clusterProvisionSched, err := provider.NewClusterProvisionScheduler(ctx, db, cfg)
	if err != nil {
		log.Error(err, "error starting cluster provisioning scheduler")
		return err
	}
	clusterProvisionSched.StartClusterProvisionScheduler(ctx)

	return nil
}
