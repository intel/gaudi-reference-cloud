// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package manageddb

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

type TestDb struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func (test *TestDb) Start(ctx context.Context) (*ManagedDb, error) {
	var err error

	test.pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("dockertest.newPool: %w", err)
	}

	username := "user"
	password := uuid.New().String()
	dbName := "test"

	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}

	// See https://github.com/ory/dockertest/
	test.resource, err = test.pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_USER=" + username,
			"POSTGRES_DB=" + dbName,
			"listen_addresses = '*'",
		},
		Labels: map[string]string{
			"createdByExecutable": executable,
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, fmt.Errorf("RunWithOptions: %w", err)
	}
	// Tell docker to hard kill the container after some period.
	// This should be greater than the test duration.
	var killAfterSeconds uint = 15 * 60
	err = test.resource.Expire(killAfterSeconds)
	if err != nil {
		return nil, err
	}

	hostAndPort := test.resource.GetHostPort("5432/tcp")
	databaseRawURL := fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", username, password, hostAndPort, dbName)
	databaseURL, err := url.Parse(databaseRawURL)
	if err != nil {
		return nil, err
	}
	managedDb := &ManagedDb{
		DatabaseURL: databaseURL,
		Driver:      "pgx",
	}

	// exponential backoff-retry, because the application in the container might not be ready
	// to accept connections yet
	test.pool.MaxWait = 120 * time.Second
	err = test.pool.Retry(func() error {
		_, err = managedDb.Open(ctx)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	return managedDb, nil
}

func (test *TestDb) Stop(ctx context.Context) error {
	if err := test.pool.Purge(test.resource); err != nil {
		// Ignore some errors.
		match, matchErr := regexp.MatchString(".*removal of container .* is already in progress.*", err.Error())
		match2, matchErr2 := regexp.MatchString(".*No such container.*", err.Error())
		if matchErr == nil && match || matchErr2 == nil && match2 {
			return nil
		}
		return fmt.Errorf("could not purge resource: %w", err)
	}
	return nil
}
