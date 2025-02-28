// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package utils

import (
	"errors"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/jackc/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func sanitizeDBError(pgErr *pgconn.PgError) error {
	if pgErr.Code == manageddb.KErrResourceNotFoundViolation {
		return status.Error(codes.NotFound, "resource not found")
	}
	if pgErr.Code == manageddb.KErrUniqueViolation {
		return status.Error(codes.AlreadyExists, "unique resource constraint: resource already exists")
	} else {
		return status.Error(codes.Internal, "internal server error")
	}
}

func sanitizeUncategorizedError(err error) error {
	//handle if errors are pg database error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return sanitizeDBError(pgErr)
	}
	// handle all uncategorized errors
	return status.Error(codes.Unknown, "an unknown error occurred")
}

func SanitizeError(err error) error {
	if err == nil {
		return err
	}

	if st, ok := status.FromError(err); ok {
		// Handle specific error codes
		switch st.Code() {
		case codes.InvalidArgument:
			return err
		case codes.NotFound:
			return status.Error(codes.NotFound, "resource not found")
		case codes.AlreadyExists:
			return status.Error(codes.AlreadyExists, "resource already exists")
		case codes.Internal:
			return status.Error(codes.Internal, "internal server error")
		case codes.Unknown:
			return status.Error(codes.Unknown, "an unknown server error occurred")
		case codes.PermissionDenied:
			return err
		case codes.FailedPrecondition:
			return err
		case codes.ResourceExhausted:
			return err
		case codes.OutOfRange:
			return err
		default:
			return sanitizeUncategorizedError(err)
		}
	}

	return sanitizeUncategorizedError(err)
}
