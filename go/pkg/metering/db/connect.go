// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package db

import (
	"database/sql"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	"context"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func OpenDb(ctx context.Context, dbConnURL string) (*sql.DB, error) {
	log := log.FromContext(ctx).WithName("OpenDb")

	cfg, err := pgx.ParseConfig(dbConnURL)
	if err != nil {
		log.Error(err, "error parsing connection url")
		return nil, fmt.Errorf("error parsing connection url")
	}
	cfg.Tracer = otelpgx.NewTracer()

	db := stdlib.OpenDB(*cfg)

	err = db.Ping()
	if err != nil {
		log.Error(err, "error connecting to the database")
		return nil, err
	}

	return db, nil
}
