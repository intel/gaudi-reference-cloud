// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	insertReleaseMDQuery = `
		INSERT INTO k8s_release 
			(version, vendor, license, PURL, release_timestamp, eos_timestamp, eol_timestamp, properties)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT(version, vendor) DO NOTHING
	`

	insertReleaseComponentQuery = `
		WITH cid AS(
			SELECT id 
			FROM k8s_release
			WHERE version = $1 AND VENDOR = $2
		)
		INSERT INTO k8s_release_components 
			(name, version, sha256, url, license, release_id, type, release_timestamp)
		VALUES ($3, $4, $5, $6, $7, (select id from cid), $8, $9)
		ON CONFLICT (release_id,name,version) DO NOTHING
	`

	getReleaseMDQuery = `
		SELECT version, vendor, license, PURL, release_timestamp, eos_timestamp, eol_timestamp, properties 
		FROM k8s_release 
		WHERE version = $1
	`

	getAllReleasesQuery = `
		SELECT r.version, r.vendor, r.license, r.PURL, r.release_timestamp, r.eos_timestamp, r.eol_timestamp, r.properties 
		FROM k8s_release as r
		ORDER BY r.version DESC
	`

	getReleaseComponents = `
		SELECT c.name, c.version, c.sha256, c.url, c.license, c.release_timestamp, c.type
		FROM k8s_release as r, k8s_release_components as c
		WHERE  r.id = c.release_id AND
		r.version = $1 AND r.vendor = $2

	`
)

func StoreReleaseMetadata(ctx context.Context, dbconn *sql.DB, in *v1.K8SReleaseMD) (int64, error) {
	logger := log.FromContext(ctx).WithName("StoreReleaseMetadata()")
	var idx int64

	res, err := dbconn.ExecContext(ctx,
		insertReleaseMDQuery,
		in.ReleaseId,
		common.MapOSSVendorTypeToSQL(in.Vendor),
		in.License,
		in.Purl,
		in.ReleaseTimestamp.AsTime(),
		in.EosTimestamp.AsTime(),
		in.EolTimestamp.AsTime(),
		in.Properties,
	)

	if err != nil {
		logger.Error(err, "error updating k8s_release state in db")
		return idx, status.Errorf(codes.Internal, "release record insertion failed")
	}

	if res == nil {
		logger.Error(err, "empty result received")
		return idx, status.Errorf(codes.Internal, "empty result received")
	}

	idx, err = res.LastInsertId()
	if err != nil {
		logger.Error(err, "error inserting record")
	}
	logger.Info("record inserted", "idx", idx)

	for _, c := range in.Components {
		_, err = dbconn.ExecContext(ctx,
			insertReleaseComponentQuery,
			in.ReleaseId,
			common.MapOSSVendorTypeToSQL(in.Vendor),
			c.Name,
			c.Version,
			c.Sha256,
			c.Purl,
			c.License,
			common.MapComponentTypeToSQL(c.GetType()),
			c.ReleaseTime.AsTime(),
		)
		if err != nil {
			logger.Error(err, "error updating k8s_release_component in db")
			break
			// return idx, status.Errorf(codes.Internal, "release component record insertion failed")
		}
	}

	return idx, nil
}

func GetReleaseMetadata(ctx context.Context, dbconn *sql.DB, in *v1.GetReleaseRequest) (*v1.K8SReleaseMD, error) {
	logger := log.FromContext(ctx).WithName("GetReleaseMetadata()")

	res := v1.K8SReleaseMD{}
	row := dbconn.QueryRowContext(ctx,
		getReleaseMDQuery,
		in.ReleaseId)

	var releaseTS, eosTS, eolTS time.Time
	props := []byte{}
	vendorStr := ""
	switch err := row.Scan(&res.ReleaseId, &vendorStr, &res.License,
		&res.Purl, &releaseTS, &eosTS, &eolTS, &props); err {
	case sql.ErrNoRows:
		logger.Info("no records found ", " version ", in.ReleaseId)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		res.ReleaseTimestamp = timestamppb.New(releaseTS)
		res.EolTimestamp = timestamppb.New(eolTS)
		res.EosTimestamp = timestamppb.New(eosTS)
		res.Vendor = common.MapOSSVendorTypeToPB(vendorStr)
		logger.Info("vendor mapped ", " vendor ", res.Vendor)
		if props != nil {
			err := json.Unmarshal([]byte(props), &res.Properties)
			if err != nil {
				logger.Error(err, "Error Unmarshalling JSON:")
			}
		}
		res.Components, err = getAllReleaseComponents(ctx, dbconn, res.GetReleaseId(), vendorStr)
		if err != nil {
			logger.Error(err, "Error fetching release components")
		}
	default:
		logger.Error(err, "error searching release record in db")
		return nil, status.Errorf(codes.Internal, "release record find failed")
	}

	return &res, nil
}

func GetAllReleases(ctx context.Context, topK int, component pb.ReleaseComponent, dbconn *sql.DB) ([]*v1.K8SReleaseMD, error) {
	logger := log.FromContext(ctx).WithName("GetAllReleases()")

	rows, err := dbconn.QueryContext(ctx, getAllReleasesQuery)
	if err != nil {
		logger.Error(err, "error searching release record in db")
		return nil, status.Errorf(codes.Internal, "release record find failed")
	}
	allReleases := false
	// set topk to 0 to return all releases versions (needed for internal API calls)
	if topK == 0 {
		allReleases = true
	}
	defer rows.Close()
	releases := []*v1.K8SReleaseMD{}
	for rows.Next() {
		if !allReleases && topK == 0 {
			logger.Info("filter threashold reached", "topK", topK)
			break
		}
		res := v1.K8SReleaseMD{}
		var eos, eol, release time.Time
		props := []byte{}
		vendorStr := ""
		if err := rows.Scan(&res.ReleaseId, &vendorStr, &res.License,
			&res.Purl, &release, &eos, &eol, &props); err != nil {
			logger.Error(err, "error reading result row, continue...")
		}
		// component filtering logic
		if component == pb.ReleaseComponent_KUBERNETES && !strings.HasPrefix(res.ReleaseId, "v1") {
			continue
		} else if component == pb.ReleaseComponent_CALICO && !strings.HasPrefix(res.ReleaseId, "v3") {
			continue
		}
		res.EolTimestamp = timestamppb.New(eol)
		res.EosTimestamp = timestamppb.New(eos)
		res.ReleaseTimestamp = timestamppb.New(release)
		res.Vendor = common.MapOSSVendorTypeToPB(vendorStr)
		if props != nil {
			err := json.Unmarshal([]byte(props), &res.Properties)
			if err != nil {
				logger.Error(err, "Error Unmarshalling JSON")
			}
		}
		res.Components, err = getAllReleaseComponents(ctx, dbconn, res.GetReleaseId(), vendorStr)
		if err != nil {
			logger.Error(err, "Error fetching release components")
		}
		releases = append(releases, &res)
		if topK != 0 {
			topK--
		}
	}
	return releases, nil
}

func getAllReleaseComponents(ctx context.Context, dbconn *sql.DB, releaseId, vendor string) ([]*v1.ReleaseComponents, error) {
	logger := log.FromContext(ctx).WithName("getAllReleaseComponents()")
	components := []*v1.ReleaseComponents{}
	rows, err := dbconn.QueryContext(ctx, getReleaseComponents, releaseId, vendor)
	if err != nil {
		logger.Error(err, "error searching release components record in db")
		return nil, status.Errorf(codes.Internal, "release components record find failed")
	}

	defer rows.Close()
	for rows.Next() {
		comp := v1.ReleaseComponents{}
		var compType string
		releaseTs := time.Time{}
		if err := rows.Scan(&comp.Name, &comp.Version, &comp.Sha256, &comp.Purl, &comp.License, &releaseTs, &compType); err != nil {
			logger.Error(err, "error reading result row, continue...")
		}
		comp.ReleaseId = fmt.Sprintf("%s:%s", comp.Name, comp.Version)
		comp.Vendor = common.MapOSSVendorTypeToPB("oss")
		comp.ReleaseTime = timestamppb.New(releaseTs)
		comp.Type = common.MapComponentTypeToPB(compType)

		components = append(components, &comp)
	}
	return components, nil
}

func StoreReleaseComponentsMetadata(ctx context.Context, dbconn *sql.DB, in *v1.ReleaseComponents) (int64, error) {
	logger := log.FromContext(ctx).WithName("StoreReleaseMetadata()")
	var idx int64

	_, err := dbconn.ExecContext(ctx,
		insertReleaseComponentQuery,
		in.ReleaseId,
		common.MapOSSVendorTypeToSQL(in.Vendor),
		in.Name,
		in.Version,
		in.Sha256,
		in.Purl,
		in.License,
		common.MapComponentTypeToSQL(in.GetType()),
		in.ReleaseTime.AsTime(),
	)
	if err != nil {
		logger.Error(err, "error updating k8s_release_component in db")
		return idx, status.Errorf(codes.Internal, "release component record insertion failed")
	}
	return idx, nil
}
