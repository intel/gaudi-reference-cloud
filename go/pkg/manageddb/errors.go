// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package manageddb

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	KErrUniqueViolation           = "23505"
	KErrResourceNotFoundViolation = "404"
)

func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	pgErr := &pgconn.PgError{}
	return errors.As(err, &pgErr) && pgErr.Code == KErrUniqueViolation
}
