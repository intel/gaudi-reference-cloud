// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	fmt.Println("Running postgres_deploy_migrations.go")
	var (
		host    = os.Args[2]
		port    = 5432
		user    = "postgres"
		dbname  = "main"
		sslmode = "disable"
	)
	if os.Args[1] == "dbaas" {
		host = "100.64.17.217"
		user = "psqliks_admin"
		dbname = "main"
		sslmode = "require"
	}
	password := os.Getenv("post_pass")
	conn_string := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", user, password, host, port, dbname, sslmode)

	fmt.Println("Running migration on DB")
	m, err := migrate.New(
		"file:///tmp/migrations/",
		conn_string)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Running M.Up")
	if err := m.Up(); err != nil {
		log.Fatal(err)
	}
}
