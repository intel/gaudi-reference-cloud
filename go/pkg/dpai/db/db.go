// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package db

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func getDBUrl(cfg *config.Config) (string, error) {
	log := log.FromContext(context.Background()).WithName("DpaiDb.getDBUrl")

	managedDb, err := manageddb.New(context.Background(), &cfg.Database)

	if err != nil {
		log.Error(err, "Error fetching database URL...")
		return "", err
	}

	return managedDb.DatabaseURL.String(), nil

}

func InitDB(cfg *config.Config) (*pgxpool.Pool, error) {
	log := log.FromContext(context.Background()).WithName("DpaiDb.InitDB")

	database_url, err := getDBUrl(cfg)
	if err != nil {
		return nil, err
	}

	poolConfig, err := pgxpool.ParseConfig(database_url)
	if err != nil {
		return nil, err
	}

	// TODO: Template this as per the capcity of the connection pooler
	poolConfig.MaxConns = 27
	poolConfig.MinConns = 5
	poolConfig.ConnConfig.ConnectTimeout = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Error(err, "Error connecting to the database...")
		return nil, err
	}

	log.Info("Database connection pool initialization succeeded...")

	return pool, nil
}

func AcquireConn(pool *pgxpool.Pool) (*pgxpool.Conn, error) {
	log := log.FromContext(context.Background()).WithName("DpaiDb.AcquireConn")
	if err := pool.Ping(context.Background()); err != nil {
		fmt.Printf("Error: %+v", err)
	} else {
		log.Info("Ping success \n")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Error(err, "Error connecting to the database...")
	}

	return conn, err
}

//go:embed migrations/*.sql
var fs embed.FS

//go:embed migrations/*.sql
var MigrationsFs embed.FS
var MigrationsDir string = "migrations"

func Migrate(cfg *config.Config, pool *pgxpool.Pool) error {
	log := log.FromContext(context.Background()).WithName("DpaiDb.Migrate")
	log.Info("Creating and migrating database")
	database_url, err := getDBUrl(cfg)
	if err != nil {
		return err
	}
	fmt.Println("Database url: ", database_url)
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("unable to initialize the filesystem: %w", err)
	}
	fmt.Println("initialization completed")
	m, err := migrate.NewWithSourceInstance("iofs", d, database_url)
	if err != nil {
		return fmt.Errorf("unable to create postgres migrate instance: %w", err)
	}

	// m.Up() // or m.Step(2) if you want to explicitly set the number of migrations to run
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate database: %w", err)
	}
	log.Info("Database migrated successfully")
	return nil
}
