// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package db

import "embed"

//go:embed migrations/*.sql
var MigrationsFs embed.FS
var MigrationsDir string = "migrations"
