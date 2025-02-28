// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iprm

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

// Transforms an Port to a form that can be written to a SQL database.
// Also performs the inverse, reading from sql.Rows and creating Port.
// This uses the JSON serializer from the GRPC Gateway.
type PortSQLTransformer struct {
	marshaler *runtime.JSONPb
}

func NewPortSQLTransformer() *PortSQLTransformer {
	return &PortSQLTransformer{
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
func (s *PortSQLTransformer) Flatten(ctx context.Context, port *pb.PortPrivate) (*protodb.Flattened, error) {
	flattened := &protodb.Flattened{}
	json, err := s.marshaler.Marshal(port)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize to json: %w", err)
	}
	flattened.Add("value", json)
	return flattened, nil
}

// Read a database row into an port.
func (s *PortSQLTransformer) FromRow(ctx context.Context, rows *sql.Rows) (*pb.PortPrivate, error) {
	log := log.FromContext(ctx).WithName("PortSQLTransformer.FromRow")
	metadata := &pb.PortMetadataPrivate{}
	var deletedTimestamp string
	var resourceJson []byte
	if err := rows.Scan(&metadata.CloudAccountId, &metadata.ResourceId, &metadata.Name, &deletedTimestamp, &metadata.ResourceVersion, &resourceJson); err != nil {
		return nil, fmt.Errorf("FromRow: Scan: %w", err)
	}
	log.V(9).Info("scanned", logkeys.ResourceId, metadata.ResourceId, logkeys.ResourceJson, string(resourceJson))
	port := &pb.PortPrivate{}
	if err := s.marshaler.Unmarshal(resourceJson, &port); err != nil {
		return nil, err
	}
	log.V(9).Info("decoded", logkeys.PORT, port)
	// Copy fields directly in the row to the port.
	port.Metadata.CloudAccountId = metadata.CloudAccountId
	port.Metadata.ResourceId = metadata.ResourceId
	port.Metadata.ResourceVersion = metadata.ResourceVersion
	return port, nil
}

// Read a database row into an PortPrivateWatchResponse.
// This encodes the Spec & Status as json blobs to allow informer to handle the proto style resources.
func (s *PortSQLTransformer) FromRowWatchResponse(ctx context.Context, rows *sql.Rows) (*pb.PortPrivateWatchResponse, error) {
	log := log.FromContext(ctx).WithName("PortSQLTransformer.FromRowWatchResponse")
	metadata := &pb.PortMetadataPrivate{}
	var deletedTimestamp string
	var resourceJson []byte
	if err := rows.Scan(&metadata.CloudAccountId, &metadata.ResourceId, &metadata.Name, &deletedTimestamp, &metadata.ResourceVersion, &resourceJson); err != nil {
		return nil, fmt.Errorf("FromRowWatchResponse: Scan: %w", err)
	}
	log.V(9).Info("scanned", logkeys.ResourceId, metadata.ResourceId, logkeys.ResourceJson, string(resourceJson))

	// Unmarshal into a PortPrivate
	portPrivate := &pb.PortPrivate{}
	if err := s.marshaler.Unmarshal(resourceJson, &portPrivate); err != nil {
		return nil, err
	}

	// Convert the PortPrivate into a PortPrivateWatchResponse
	spec, err := s.marshaler.Marshal(portPrivate.Spec)
	if err != nil {
		return nil, err
	}

	status, err := s.marshaler.Marshal(portPrivate.Status)
	if err != nil {
		return nil, err
	}

	port := &pb.PortPrivateWatchResponse{
		Metadata: portPrivate.Metadata,
		Spec:     string(spec),
		Status:   string(status),
	}

	log.V(9).Info("decoded", logkeys.PORT, port)
	// Copy fields directly in the row to the port.
	port.Metadata.CloudAccountId = metadata.CloudAccountId
	port.Metadata.ResourceId = metadata.ResourceId
	//port.Metadata.Name = metadata.Name
	port.Metadata.ResourceVersion = metadata.ResourceVersion
	port.Metadata.DeletionTimestamp = portPrivate.Metadata.DeletionTimestamp
	port.Metadata.DeletedTimestamp, err = timestampStrToPbTimestamp(deletedTimestamp)
	if err != nil {
		return nil, err
	}
	return port, nil
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
