// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewTestDbClient(setupScriptPath string) (*postgres.PostgresContainer, *sql.DB, error) {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase("sdncontroller"),
		postgres.WithInitScripts(setupScriptPath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		if err := postgresContainer.Terminate(ctx); err != nil {
			return nil, nil, err
		}
		return nil, nil, err
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		if err := postgresContainer.Terminate(ctx); err != nil {
			return nil, nil, err
		}
		return nil, nil, err
	}
	return postgresContainer, db, nil
}
