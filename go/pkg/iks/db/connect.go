// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package db

import (
	"database/sql"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	"context"
	"fmt"

	"embed"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var MigrationsFs embed.FS
var MigrationsDir string = "migrations"

//go:embed migrations_unit/*.sql
var MigrationsFsUnit embed.FS
var MigrationsDirUnit string = "migrations_unit"

// OpenDb will start the db connection for postgres db
func OpenDb(ctx context.Context, dbUrl string) (*sql.DB, error) {

	log := log.FromContext(ctx).WithName("OpenDb")
	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		// DO not pass or print `err` to calling function as the
		// error message might contain complete URL wth user/password
		return nil, fmt.Errorf("error opening the database connection")
	}

	err = db.Ping()
	if err != nil {
		log.Error(err, "error connecting to the database")
		return nil, err
	}

	return db, nil
}
