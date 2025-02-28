// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package testminio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

type TestMinIO struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func New() *TestMinIO {
	return &TestMinIO{}
}

func (t *TestMinIO) Start(ctx context.Context, bucketName string) (*url.URL, error) {
	var err error

	t.pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("dockertest.newPool: %w", err)
	}
	if err := t.pool.Client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping docker server: %w", err)
	}

	username := "minioadmin"
	password := uuid.New().String()
	if err := os.Setenv("AWS_ACCESS_KEY_ID", username); err != nil {
		return nil, fmt.Errorf("failed to set AWS_ACCESS_KEY_ID environment variable: %w", err)
	}
	if err := os.Setenv("AWS_SECRET_ACCESS_KEY", password); err != nil {
		return nil, fmt.Errorf("failed to set AWS_ACCESS_KEY_ID environment variable: %w", err)
	}

	// See https://github.com/ory/dockertest/
	t.resource, err = t.pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Entrypoint: []string{"/bin/sh"},
		// Pre-create bucket
		Cmd: []string{"-c", fmt.Sprintf("mkdir -p /data/%s && /opt/bin/minio server /data", bucketName)},
		Env: []string{
			"MINIO_ROOT_PASSWORD=" + password,
			"MINIO_ROOT_USER=" + username,
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, fmt.Errorf("RunWithOptions: %w", err)
	}
	// Tell docker to hard kill the container in 120 seconds
	err = t.resource.Expire(120)
	if err != nil {
		return nil, err
	}

	hostAndPort := t.resource.GetHostPort("9000/tcp")
	minioUrl, err := url.Parse(fmt.Sprintf("http://%s", hostAndPort))
	if err != nil {
		return nil, err
	}

	// exponential backoff-retry, because the application in the container might not be ready
	// to accept connections yet
	t.pool.MaxWait = 120 * time.Second
	minioLivenessEndpoint := "minio/health/live"
	err = t.pool.Retry(func() error {
		resp, err := http.Get(fmt.Sprintf("%s/%s", minioUrl.String(), minioLivenessEndpoint))
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("minio liveness check failed")
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to minio within timeout period: %w", err)
	}
	return minioUrl, nil
}

func (t *TestMinIO) Stop(ctx context.Context) error {
	if err := t.pool.Purge(t.resource); err != nil {
		return fmt.Errorf("could not purge resource: %w", err)
	}
	return nil
}
