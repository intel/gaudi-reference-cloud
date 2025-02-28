// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package manageddb

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"strings"

	"github.com/exaring/otelpgx"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// Configuration for a ManagedDb.
type Config struct {
	// Database URL such as: postgres://127.0.0.1:5432/compute?sslmode=disable
	// See details in: https://www.postgresql.org/docs/current/libpq-connect.html
	// Do not include the username or password, as these will be replaced.
	URL string `koanf:"url"`
	// Path to the file containing the username. It should not containing a trailing new-line.
	UsernameFile string `koanf:"usernameFile"`
	// Path to the file containing the password. It should not containing a trailing new-line.
	PasswordFile string `koanf:"passwordFile"`
	// If non-zero, set the maximum number of idle connections.
	MaxIdleConnectionCount int `koanf:"maxIdleConnectionCount"`
}

type ManagedDb struct {
	DatabaseURL            *url.URL
	Driver                 string
	MaxIdleConnectionCount int
}

func New(ctx context.Context, config *Config) (*ManagedDb, error) {
	log := log.FromContext(ctx).WithName("ManagedDb.New")
	log.Info("Database configuration", "config", config)
	databaseURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}
	username, err := os.ReadFile(config.UsernameFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read username file %s: %v", config.UsernameFile, err)
	}
	log.Info("Database configuration", "username", string(username))

	password, err := os.ReadFile(config.PasswordFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read password file %s: %v", config.PasswordFile, err)
	}
	// if username and password are empty allow userinfo from the URL
	// to be used. This is helpful for developer testing.
	if string(username) != "" && string(password) != "" {
		databaseURL.User = url.UserPassword(string(username), strings.TrimSpace(string(password)))
	}
	return &ManagedDb{
		DatabaseURL:            databaseURL,
		Driver:                 "pgx",
		MaxIdleConnectionCount: config.MaxIdleConnectionCount,
	}, nil
}

func (m *ManagedDb) Migrate(ctx context.Context, fsys fs.FS, subdir string) error {
	log := log.FromContext(ctx).WithName("ManagedDb.Migrate")
	log.Info("Creating and migrating database")
	dir, err := iofs.New(fsys, subdir)
	if err != nil {
		return fmt.Errorf("create iofs for migrate: %w", err)
	}
	mm, err := migrate.NewWithSourceInstance("iofs", dir, m.DatabaseURL.String())
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	if err = mm.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate database: %w", err)
	}
	log.Info("Database migrated successfully")
	return nil
}

func (m *ManagedDb) Open(ctx context.Context) (*sql.DB, error) {
	log := log.FromContext(ctx).WithName("ManagedDb.Open")
	log.Info("Opening database")
	cfg, err := pgx.ParseConfig(m.DatabaseURL.String())
	if err != nil {
		return nil, fmt.Errorf("unable to parse database configuration: %w", err)
	}
	// pgx ensures TLSConfig is initialized if a secure connection is required
	if cfg.TLSConfig != nil {
		// Set the minimum TLS version to TLS 1.3
		log.Info("Applying enhanced security configurations")
		cfg.TLSConfig.MinVersion = tls.VersionTLS13
	} else {
		// Default connection settings
		log.Info("Proceeding with default connection settings")
	}

	cfg.Tracer = otelpgx.NewTracer()
	db := stdlib.OpenDB(*cfg)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	log.Info("Connected to database")

	// Update the maximum idle connection count. This can be increased to improve performance.
	if m.MaxIdleConnectionCount != 0 {
		log.Info("Setting database MaxIdleConnections configuration", "MaxIdleConnectionCount", m.MaxIdleConnectionCount)
		db.SetMaxIdleConns(m.MaxIdleConnectionCount)
	}

	return db, nil
}
