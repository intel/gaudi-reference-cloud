// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package instance

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

// Transforms an InstancePrivate to a form that can be written to a SQL database.
// Also performs the inverse, reading from sql.Rows and creating an InstancePrivate.
// This uses the JSON serializer from the GRPC Gateway.
type InstanceSqlTransformer struct {
	marshaler *runtime.JSONPb
}

func NewInstanceSqlTransformer() *InstanceSqlTransformer {
	return &InstanceSqlTransformer{
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
func (s *InstanceSqlTransformer) Flatten(ctx context.Context, instance *pb.InstancePrivate) (*protodb.Flattened, error) {
	flattened := &protodb.Flattened{}
	jsonInstance, err := s.marshaler.Marshal(instance)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize to json: %w", err)
	}
	flattened.Add("value", jsonInstance)
	return flattened, nil
}

// Read a database row into an InstancePrivate. Used for private APIs.
func (s *InstanceSqlTransformer) FromRow(ctx context.Context, rows *sql.Rows) (*pb.InstancePrivate, error) {
	log := log.FromContext(ctx).WithName("InstanceSqlTransformer.FromRow")
	metadata := &pb.InstanceMetadataPrivate{}
	var deletedTimestamp string
	var resourceJson []byte
	if err := rows.Scan(&metadata.CloudAccountId, &metadata.ResourceId, &metadata.Name, &deletedTimestamp, &metadata.ResourceVersion, &resourceJson); err != nil {
		return nil, fmt.Errorf("RowToInstancePrivate: Scan: %w", err)
	}
	log.V(9).Info("scanned", logkeys.ResourceId, metadata.ResourceId, logkeys.ResourceJson, string(resourceJson))
	instance := &pb.InstancePrivate{}
	if err := s.marshaler.Unmarshal(resourceJson, &instance); err != nil {
		return nil, err
	}
	log.V(9).Info("decoded", logkeys.Instance, instance)
	// Copy fields directly in the row to the instance.
	instance.Metadata.CloudAccountId = metadata.CloudAccountId
	instance.Metadata.ResourceId = metadata.ResourceId
	instance.Metadata.Name = metadata.Name
	instance.Metadata.ResourceVersion = metadata.ResourceVersion
	var err error
	instance.Metadata.DeletedTimestamp, err = timestampStrToPbTimestamp(deletedTimestamp)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// When using FromRow, the SQL SELECT query must select these columns.
func (s *InstanceSqlTransformer) ColumnsForFromRow() string {
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
