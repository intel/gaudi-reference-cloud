// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"errors"
	"fmt"
	"net/url"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type CasbinRule struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Ptype string `gorm:"size:512;uniqueIndex:unique_index"`
	V0    string `gorm:"size:512;uniqueIndex:unique_index"`
	V1    string `gorm:"size:512;uniqueIndex:unique_index"`
	V2    string `gorm:"size:512;uniqueIndex:unique_index"`
	V3    string `gorm:"size:512;uniqueIndex:unique_index"`
	V4    string `gorm:"size:512;uniqueIndex:unique_index"`
	V5    string `gorm:"size:512;uniqueIndex:unique_index"`
}

func NewAdapter(u *url.URL) (*gormadapter.Adapter, error) {
	if u == nil {
		err := errors.New("url is required")
		logger.Error(err, "url is required")
		return nil, err
	}

	logger.Info("casbinEngine database configuration username", "username:", string(u.User.Username()))

	password, _ := u.User.Password()
	query := u.Query()
	sslmode := query.Get("sslmode")
	sslrootcert := query.Get("sslrootcert")
	dsn := "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslrootcert=%s"
	dsn = fmt.Sprintf(dsn,
		u.Hostname(), u.Port(), u.User.Username(), password, u.Path[1:], sslmode, sslrootcert)

	// Initialize the Gorm adapter for PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error(err, "failed to initialize db session")
		return nil, err
	}
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(db, &CasbinRule{})
	if err != nil {
		logger.Error(err, "failed to create new adapter db")
		return nil, err
	}
	return adapter, nil
}
