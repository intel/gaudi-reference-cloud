package testutils

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

type TestDB interface {
	Start(ctx context.Context) *manageddb.TestDb
	Stop(ctx context.Context) *manageddb.TestDb
}

type SimDB struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func getHostPort(resource *dockertest.Resource, id string) string {
	dockerURL := os.Getenv("DOCKER_HOST")
	if dockerURL == "" {
		return resource.GetHostPort(id)
	}
	u, err := url.Parse(dockerURL)
	if err != nil {
		panic(err)
	}
	return u.Hostname() + ":" + resource.GetPort(id)
}

func (test *SimDB) Start(ctx context.Context) (*manageddb.ManagedDb, error) {
	fmt.Printf("STARTING......")
	var err error

	test.pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("dockertest.newPool: %w", err)
	}

	username := "user"
	password := uuid.New().String()
	dbName := "test"

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
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, fmt.Errorf("RunWithOptions: %w", err)
	}
	// Tell docker to hard kill the container in 120 seconds
	err = test.resource.Expire(120)
	if err != nil {
		return nil, err
	}

	hostAndPort := test.resource.GetHostPort("5432/tcp")
	fmt.Println("HOST.....", hostAndPort)
	databaseRawURL := fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", username, password, hostAndPort, dbName)
	databaseURL, err := url.Parse(databaseRawURL)
	if err != nil {
		return nil, err
	}
	managedDb := &manageddb.ManagedDb{
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

func (test *SimDB) Stop(ctx context.Context) error {
	if err := test.pool.Purge(test.resource); err != nil {
		return fmt.Errorf("could not purge resource: %w", err)
	}
	return nil
}
