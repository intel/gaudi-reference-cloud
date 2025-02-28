// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package subnet

import (
	"context"
	"database/sql"
	"fmt"
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

// Transforms an Subnet to a form that can be written to a SQL database.
// Also performs the inverse, reading from sql.Rows and creating an Subnet.
// This uses the JSON serializer from the GRPC Gateway.
type SubnetSQLTransformer struct {
	marshaler *runtime.JSONPb
}

func NewSubnetSQLTransformer() *SubnetSQLTransformer {
	return &SubnetSQLTransformer{
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
func (s *SubnetSQLTransformer) Flatten(ctx context.Context, subnet *pb.SubnetPrivate) (*protodb.Flattened, error) {
	flattened := &protodb.Flattened{}
	jsonSubnet, err := s.marshaler.Marshal(subnet)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize to json: %w", err)
	}
	flattened.Add("value", jsonSubnet)
	return flattened, nil
}

// Read a database row into an Subnet.
func (s *SubnetSQLTransformer) FromRow(ctx context.Context, rows *sql.Rows) (*pb.SubnetPrivate, error) {
	log := log.FromContext(ctx).WithName("SubnetSQLTransformer.FromRow")
	metadata := &pb.SubnetMetadataPrivate{}
	var deletedTimestamp string
	var resourceJson []byte
	if err := rows.Scan(&metadata.CloudAccountId, &metadata.ResourceId, &metadata.Name, &deletedTimestamp, &metadata.ResourceVersion, &resourceJson); err != nil {
		return nil, fmt.Errorf("RowToSubnetPrivate: Scan: %w", err)
	}
	log.V(9).Info("scanned", logkeys.ResourceId, metadata.ResourceId, logkeys.ResourceJson, string(resourceJson))
	subnet := &pb.SubnetPrivate{}
	if err := s.marshaler.Unmarshal(resourceJson, &subnet); err != nil {
		return nil, err
	}
	log.V(9).Info("decoded", logkeys.SUBNET, subnet)
	// Copy fields directly in the row to the subnet.
	subnet.Metadata.CloudAccountId = metadata.CloudAccountId
	subnet.Metadata.ResourceId = metadata.ResourceId
	subnet.Metadata.ResourceVersion = metadata.ResourceVersion
	return subnet, nil
}

// Read a database row into an SubnetPrivateWatchResponse.
// This encodes the Spec & Status as json blobs to allow informer to handle the proto style resources.
func (s *SubnetSQLTransformer) FromRowWatchResponse(ctx context.Context, rows *sql.Rows) (*pb.SubnetPrivateWatchResponse, error) {
	log := log.FromContext(ctx).WithName("SubnetSQLTransformer.FromRowWatchResponse")
	metadata := &pb.SubnetMetadataPrivate{}
	var deletedTimestamp string
	var resourceJson []byte
	if err := rows.Scan(&metadata.CloudAccountId, &metadata.ResourceId, &metadata.Name, &deletedTimestamp, &metadata.ResourceVersion, &resourceJson); err != nil {
		return nil, fmt.Errorf("FromRowWatchResponse: Scan: %w", err)
	}
	log.V(9).Info("scanned", logkeys.ResourceId, metadata.ResourceId, logkeys.ResourceJson, string(resourceJson))

	// Unmarshal into a SubnetPrivate
	subnetPrivate := &pb.SubnetPrivate{}
	if err := s.marshaler.Unmarshal(resourceJson, &subnetPrivate); err != nil {
		return nil, err
	}

	// Convert the SubnetPrivate into a SubnetPrivateWatchResponse
	spec, err := s.marshaler.Marshal(subnetPrivate.Spec)
	if err != nil {
		return nil, err
	}

	status, err := s.marshaler.Marshal(subnetPrivate.Status)
	if err != nil {
		return nil, err
	}

	subnet := &pb.SubnetPrivateWatchResponse{
		Metadata: subnetPrivate.Metadata,
		Spec:     string(spec),
		Status:   string(status),
	}

	log.V(9).Info("decoded", logkeys.SUBNET, subnet)
	// Copy fields directly in the row to the subnet.
	subnet.Metadata.CloudAccountId = metadata.CloudAccountId
	subnet.Metadata.ResourceId = metadata.ResourceId
	subnet.Metadata.Name = metadata.Name
	subnet.Metadata.ResourceVersion = metadata.ResourceVersion
	subnet.Metadata.DeletionTimestamp = subnetPrivate.Metadata.DeletionTimestamp
	subnet.Metadata.DeletedTimestamp, err = timestampStrToPbTimestamp(deletedTimestamp)
	if err != nil {
		return nil, err
	}
	return subnet, nil
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
