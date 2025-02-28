<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
## IDC Managed Database Library

This library provides a simple interface for applications to manage the schema of a database.
To define a schema, a developer writes standard SQL files such as the following.

```sql
create table foo (
    id uuid primary key,
    name varchar(63) not null
);
```

These files are named with a timestamp prefix that defines the order in which these SQL files should be executed.
When an application starts a ManagedDb object, it will create the database if needed.
Then it applies the SQL files in order, if needed.
It remembers the SQL files that need to run by tracking the last-applied timestamp in a database table named `schema_migrations`.

[golang-migrate](https://github.com/golang-migrate/migrate) is used to create the database schema and upgrade it as needed.

## Using in applications

```go
import "embed"

//go:embed sql/*.sql
var fs embed.FS

    managedDb, err := manageddb.New(ctx, &cfg.Database)
    if err != nil {
        return err
    }
	if err := managedDb.Migrate(ctx, fs, "sql"); err != nil {
		return err
	}
	db, err := managedDb.Open(ctx)
	if err != nil {
		return err
	}
```

In the example, //go:embed is used to embed the migrations in the application. It's possible
to pass in a filesystem representing local storage using os.DirFS, but //go:embed is preferred.

## Using a database during testing

The type TestDb provides a ManagedDb object that is backed by a Postgres container that it starts in Docker.
This allows integration tests to use a database without the user having to start a database separately.

```go
	managedDb, err = testDb.Start(ctx)
	Expect(err).Should(Succeed())
	Expect(managedDb).ShouldNot(BeNil())
	Expect(managedDb.Migrate(ctx, fs, "sql")).Should(Succeed())
```

## Local development

### Generate secrets

```bash
cd $(git rev-parse --show-toplevel) && make secrets
```

### Install the Postgres client to allow convenient access to Postgres servers.

```bash
sudo apt install postgresql-client
```

### Start Postgres server

```bash
./db_start.sh
```

### Run Postgres client

```bash
./db_client.sh
```

## Using golang-migrate

[golang-migrate](https://github.com/golang-migrate/migrate) is used to create the database schema and upgrade it as needed.

### Install Golang Migrate

```bash
cd go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### To create a new schema change script
 
```bash
MIGRATIONS_DIR=../compute_api_server/db/migrations
migrate create -dir ${MIGRATIONS_DIR} -ext sql  "add_column_foo_to_table_bar"
```

### To apply changes on a running cluster

```bash
make gazelle
DEPLOY_ALL_IN_KIND_APPLICATIONS_TO_DELETE=".*-compute-db|.*-compute-api-server" \
make upgrade-all-in-kind-v2 |& ts | ts -i | ts -s | tee local/upgrade-all-in-kind-v2.log
```

**NOTE**: Replace `commpute-db` and `compute-api-server` with the db and the corresponding service which is responsible for applying the updated schema.

### To manually apply schema changes

```
export DBNAME=postgres
MIGRATIONS_DIR=../compute_api_server/db/migrations
./db_migrate.sh ${MIGRATIONS_DIR}
```

### Run Postgres Client in Kubernetes

For development, the `postgres` user password is not rotated by Vault.
It can be used with the Postgres client in the *-db-postgresql-0 pod.

```bash
PGPASSWORD=$POSTGRES_POSTGRES_PASSWORD psql -U postgres -d ${POSTGRES_DB}
```
