// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	insertReleasSBOMQuery = `
	WITH cid AS(
		SELECT id 
		FROM k8s_release
		WHERE version = $1
	)
		INSERT INTO k8s_release_sbom 
			(release_id, create_timestamp, format, sbom)
		VALUES ((select id from cid), $2, $3, $4)
	`

	readReleaseSBOMQuery = `
		select sbom.create_timestamp, sbom.sbom, sbom.format
		FROM k8s_release_sbom as sbom, k8s_release as rel
		WHERE rel.version = $1 AND
		rel.id = sbom.release_id
	`
)

func StoreSBOM(ctx context.Context, dbconn *sql.DB, in *v1.ReleaseSBOM) error {
	logger := log.FromContext(ctx).WithName("StoreSBOM")

	_, err := dbconn.ExecContext(ctx,
		insertReleasSBOMQuery,
		in.ReleaseId,
		in.Sbom.CreateTimestamp.AsTime(), in.Sbom.Format, in.Sbom.Sbom,
	)

	if err != nil {
		logger.Error(err, "error updating k8s_release_sbom state in db")
		return status.Errorf(codes.Internal, "release sbom record insertion failed")
	}
	logger.Info("release sbom stored successfully", "releaseId", in.ReleaseId)
	return nil
}

func ReadSBOM(ctx context.Context, dbconn *sql.DB, in *v1.GetReleaseRequest) (*v1.SBOM, error) {
	logger := log.FromContext(ctx).WithName("ReadSBOM")

	row := dbconn.QueryRowContext(ctx,
		readReleaseSBOMQuery,
		in.ReleaseId)

	sbomRes := v1.SBOM{}

	var createTS time.Time
	var format string
	switch err := row.Scan(&createTS, &sbomRes.Sbom, &format); err {
	case sql.ErrNoRows:
		logger.Info("no records found ", " version ", in.ReleaseId)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		sbomRes.CreateTimestamp = timestamppb.New(createTS)
		sbomRes.Format = marshalFormatToPb(format)
		// if sbomBuf != nil {
		// 	sbomRes.Sbom = sbomBuf
		// 	// if err != nil {
		// 	// 	logger.Error(err, "Error Unmarshalling JSON:")
		// 	// 	return nil, status.Errorf(codes.Internal, "sbom format error")
		// 	// }
		// }
	default:
		logger.Error(err, "error searching sbom record in db")
		return nil, status.Errorf(codes.Internal, "sbom record find failed")
	}

	return &sbomRes, nil
}

func marshalFormatToPb(format string) v1.ValidSBOMFormats {
	switch format {
	case "SPDX_FORMAT":
		return v1.ValidSBOMFormats_SPDX_FORMAT
	case "CDX_FORMAT":
		return v1.ValidSBOMFormats_CYCLONEDX_FORMAT
	default:
		return v1.ValidSBOMFormats_UNSPECIFIED_FORMAT
	}
}
