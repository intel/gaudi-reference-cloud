// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vpc

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Transforms an VPC to a form that can be written to a SQL database.
// Also performs the inverse, reading from sql.Rows and creating an VPC.
// This uses the JSON serializer from the GRPC Gateway.
type VPCSqlTransformer struct {
	marshaler *runtime.JSONPb
}

func NewVPCSqlTransformer() *VPCSqlTransformer {
	return &VPCSqlTransformer{
		marshaler: &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				// When writing JSON, emit fields that have default values, including for enums.
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				// When reading JSON, ignore fields with unknown names.
				DiscardUnknown: true,
			},
		},
	}
}

// Returns a Flattened object that can be used to construct a SQL INSERT or UPDATE statement.
// The Flattened object omits columns that are never updated, such as the primary key columns.
func (s *VPCSqlTransformer) Flatten(ctx context.Context, vpc *pb.VPCPrivate) (*protodb.Flattened, error) {
	flattened := &protodb.Flattened{}
	jsonVPC, err := s.marshaler.Marshal(vpc)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize to json: %w", err)
	}
	flattened.Add("value", jsonVPC)
	return flattened, nil
}

// Read a database row into an VPC.
func (s *VPCSqlTransformer) FromRow(ctx context.Context, rows *sql.Rows) (*pb.VPCPrivate, error) {
	log := log.FromContext(ctx).WithName("VPCSqlTransformer.FromRow")
	metadata := &pb.VPCMetadataPrivate{}
	var deletedTimestamp string
	var resourceJson []byte
	if err := rows.Scan(&metadata.CloudAccountId, &metadata.ResourceId, &metadata.Name, &deletedTimestamp, &metadata.ResourceVersion, &resourceJson); err != nil {
		return nil, fmt.Errorf("RowToVPCPrivate: Scan: %w", err)
	}
	log.V(9).Info("scanned", logkeys.ResourceId, metadata.ResourceId, logkeys.ResourceJson, string(resourceJson))
	vpc := &pb.VPCPrivate{}
	if err := s.marshaler.Unmarshal(resourceJson, &vpc); err != nil {
		return nil, err
	}
	log.V(9).Info("decoded", logkeys.VPC, vpc)
	// Copy fields directly in the row to the vpc.
	vpc.Metadata.CloudAccountId = metadata.CloudAccountId
	vpc.Metadata.ResourceId = metadata.ResourceId
	vpc.Metadata.Name = metadata.Name
	vpc.Metadata.ResourceVersion = metadata.ResourceVersion

	return vpc, nil
}

// Read a database row into an VPCPrivateWatchResponse.
// This encodes the Spec & Status as json blobs to allow informer to handle the proto style resources.
func (s *VPCSqlTransformer) FromRowWatchResponse(ctx context.Context, rows *sql.Rows) (*pb.VPCPrivateWatchResponse, error) {
	log := log.FromContext(ctx).WithName("VPCSqlTransformer.FromRowWatchResponse")
	metadata := &pb.VPCMetadataPrivate{}
	var deletedTimestamp string
	var resourceJson []byte
	if err := rows.Scan(&metadata.CloudAccountId, &metadata.ResourceId, &metadata.Name, &deletedTimestamp, &metadata.ResourceVersion, &resourceJson); err != nil {
		return nil, fmt.Errorf("FromRowWatchResponse: Scan: %w", err)
	}
	log.V(9).Info("scanned", logkeys.ResourceId, metadata.ResourceId, logkeys.ResourceJson, string(resourceJson))

	// Unmarshal into a VPCPrivate
	vpcPrivate := &pb.VPCPrivate{}
	if err := s.marshaler.Unmarshal(resourceJson, &vpcPrivate); err != nil {
		return nil, err
	}

	// Convert the VPCPrivate into a VPCPrivateWatchResponse
	spec, err := s.marshaler.Marshal(vpcPrivate.Spec)
	if err != nil {
		return nil, err
	}

	status, err := s.marshaler.Marshal(vpcPrivate.Status)
	if err != nil {
		return nil, err
	}

	vpc := &pb.VPCPrivateWatchResponse{
		Metadata: vpcPrivate.Metadata,
		Spec:     string(spec),
		Status:   string(status),
	}

	log.V(9).Info("decoded", logkeys.VPC, vpc)
	// Copy fields directly in the row to the vpc.
	vpc.Metadata.CloudAccountId = metadata.CloudAccountId
	vpc.Metadata.ResourceId = metadata.ResourceId
	vpc.Metadata.Name = metadata.Name
	vpc.Metadata.ResourceVersion = metadata.ResourceVersion
	vpc.Metadata.DeletionTimestamp = vpcPrivate.Metadata.DeletionTimestamp
	vpc.Metadata.DeletedTimestamp, err = timestampStrToPbTimestamp(deletedTimestamp)
	if err != nil {
		return nil, err
	}

	return vpc, nil
}

// When using FromRow, the SQL SELECT query must select these columns.
func (s *VPCSqlTransformer) ColumnsForFromRow() string {
	cols := []string{"cloud_account_id", "resource_id", "name", "deleted_timestamp", "resource_version", "value"}
	return strings.Join(cols, ", ")
}

// Convert a timestamp from Postgres format to Protobuf.
// The special time "infinity" is returned as (nil, nil).
func timestampStrToPbTimestamp(ts string) (*timestamppb.Timestamp, error) {
	if ts == common.TimestampInfinityStr {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return nil, err
	}
	return timestamppb.New(t), nil
}
