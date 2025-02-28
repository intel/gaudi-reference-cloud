// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package loadbalancer

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

// Transforms a LoadBalancerPrivate to a form that can be written to a SQL database.
// Also performs the inverse, reading from sql.Rows and creating an LoadBalancerPrivate.
// This uses the JSON serializer from the GRPC Gateway.
type LoadBalancerSqlTransformer struct {
	marshaler *runtime.JSONPb
}

func NewLoadBalancerSqlTransformer() *LoadBalancerSqlTransformer {
	return &LoadBalancerSqlTransformer{
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
func (s *LoadBalancerSqlTransformer) Flatten(ctx context.Context, loadbalancer *pb.LoadBalancerPrivate) (*protodb.Flattened, error) {
	flattened := &protodb.Flattened{}
	jsonLoadBalancer, err := s.marshaler.Marshal(loadbalancer)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize to json: %w", err)
	}
	flattened.Add("value", jsonLoadBalancer)
	return flattened, nil
}

// Read a database row into an LoadBalancerPrivate. Used for private APIs.
func (s *LoadBalancerSqlTransformer) FromRow(ctx context.Context, rows *sql.Rows) (*pb.LoadBalancerPrivate, error) {
	log := log.FromContext(ctx).WithName("LoadBalancerService.RowToLoadBalancerPrivate")
	metadata := &pb.LoadBalancerMetadataPrivate{}
	var deletedTimestamp string
	var resourceJson []byte
	if err := rows.Scan(&metadata.CloudAccountId, &metadata.ResourceId, &metadata.Name, &deletedTimestamp, &metadata.ResourceVersion, &resourceJson); err != nil {
		return nil, fmt.Errorf("RowToLoadBalancerPrivate: Scan: %w", err)
	}
	log.V(9).Info("scanned", logkeys.ResourceId, metadata.ResourceId, logkeys.ResourceJson, string(resourceJson))
	loadbalancer := &pb.LoadBalancerPrivate{}
	if err := s.marshaler.Unmarshal(resourceJson, &loadbalancer); err != nil {
		return nil, err
	}
	log.V(9).Info("decoded", logkeys.LoadBalancer, loadbalancer)

	// Copy fields directly in the row to the loadbalancer.
	loadbalancer.Metadata.CloudAccountId = metadata.CloudAccountId
	loadbalancer.Metadata.ResourceId = metadata.ResourceId
	loadbalancer.Metadata.Name = metadata.Name
	loadbalancer.Metadata.ResourceVersion = metadata.ResourceVersion

	return loadbalancer, nil
}

// When using FromRow, the SQL SELECT query must select these columns.
func (s *LoadBalancerSqlTransformer) ColumnsForFromRow() string {
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
