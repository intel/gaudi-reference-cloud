// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package db

import (
	"context"
	"embed"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

//go:embed migration/*.sql
var fsys embed.FS

func AutoMigrateDB(ctx context.Context, mdb *manageddb.ManagedDb) error {
	return mdb.Migrate(ctx, fsys, "migration")
}
