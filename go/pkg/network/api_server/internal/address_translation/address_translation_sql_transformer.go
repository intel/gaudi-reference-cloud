// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package address_translation

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

// Transforms an AddressTranslation to a form that can be written to a SQL database.
// Also performs the inverse, reading from sql.Rows and creating an AddressTranslation.
// This uses the JSON serializer from the GRPC Gateway.
type AddressTranslationSqlTransformer struct {
	marshaler *runtime.JSONPb
}

func NewAddressTranslationSqlTransformer() *AddressTranslationSqlTransformer {
	return &AddressTranslationSqlTransformer{
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
func (s *AddressTranslationSqlTransformer) Flatten(ctx context.Context, addressTranslation *pb.AddressTranslationPrivate) (*protodb.Flattened, error) {
	flattened := &protodb.Flattened{}
	jsonAddressTranslation, err := s.marshaler.Marshal(addressTranslation)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize to json: %w", err)
	}

	flattened.Add("value", jsonAddressTranslation)
	return flattened, nil
}

// Returns a Flattened object that can be used to construct a SQL Query for list operations.
func (s *AddressTranslationSqlTransformer) FlattenForList(ctx context.Context, listReq *pb.AddressTranslationListPrivateRequest) *protodb.Flattened {
	spec := listReq.Spec
	flattened := &protodb.Flattened{}

	if spec.PortId != "" {
		flattened.Add("value->'spec'->>'portId'", spec.PortId)
	}
	if spec.TranslationType != "" {
		flattened.Add("value->'spec'->>'translationType'", spec.TranslationType)
	}
	if spec.ProfileId != "" {
		flattened.Add("value->'spec'->>'profileId'", spec.ProfileId)
	}
	if spec.IpAddress != "" {
		flattened.Add("value->'spec'->>'ipAddress'", spec.IpAddress)
	}
	if spec.MacAddress != "" {
		flattened.Add("value->'spec'->>'macAddress'", spec.MacAddress)
	}
	flattened.Add("deleted_timestamp", common.TimestampInfinityStr)

	return flattened
}

// Read a database row into an AddressTranslation.
func (s *AddressTranslationSqlTransformer) FromRow(ctx context.Context, rows *sql.Rows) (*pb.AddressTranslationPrivate, error) {
	log := log.FromContext(ctx).WithName("AddressTranslationSqlTransformer.FromRow")
	metadata := &pb.AddressTranslationMetadataPrivate{}
	var deletedTimestamp string
	var resourceJson []byte
	if err := rows.Scan(&metadata.ResourceId, &metadata.CloudAccountId, &deletedTimestamp, &metadata.ResourceVersion, &resourceJson); err != nil {
		return nil, fmt.Errorf("RowToAddressTranslationPrivate: Scan: %w", err)
	}
	log.V(9).Info("scanned", logkeys.ResourceId, metadata.ResourceId, logkeys.ResourceJson, string(resourceJson))
	addressTranslation := &pb.AddressTranslationPrivate{}
	if err := s.marshaler.Unmarshal(resourceJson, &addressTranslation); err != nil {
		return nil, err
	}
	log.V(9).Info("decoded", logkeys.ADDRESS_TRANSLATION, addressTranslation)
	// Copy fields directly in the row to the addressTranslation.
	addressTranslation.Metadata.CloudAccountId = metadata.CloudAccountId
	addressTranslation.Metadata.ResourceId = metadata.ResourceId
	addressTranslation.Metadata.ResourceVersion = metadata.ResourceVersion

	var err error
	addressTranslation.Metadata.DeletionTimestamp, err = timestampStrToPbTimestamp(deletedTimestamp)
	if err != nil {
		return nil, err
	}
	return addressTranslation, nil
}

// Read a database row into an AddressTranslationPrivateWatchResponse.
// This encodes the Spec & Status as json blobs to allow informer to handle the proto style resources.
func (s *AddressTranslationSqlTransformer) FromRowWatchResponse(ctx context.Context, rows *sql.Rows) (*pb.AddressTranslationPrivateWatchResponse, error) {
	log := log.FromContext(ctx).WithName("AddressTranslationSqlTransformer.FromRowWatchResponse")
	metadata := &pb.AddressTranslationMetadataPrivate{}
	var deletedTimestamp string
	var resourceJson []byte
	if err := rows.Scan(&metadata.ResourceId, &metadata.CloudAccountId, &deletedTimestamp, &metadata.ResourceVersion, &resourceJson); err != nil {
		return nil, fmt.Errorf("FromRowWatchResponse: Scan: %w", err)
	}
	log.V(9).Info("scanned", logkeys.ResourceId, metadata.ResourceId, logkeys.ResourceJson, string(resourceJson))

	// Unmarshal into a AddressTranslationPrivate
	addressTranslationPrivate := &pb.AddressTranslationPrivate{}
	if err := s.marshaler.Unmarshal(resourceJson, &addressTranslationPrivate); err != nil {
		return nil, err
	}

	// Convert the AddressTranslationPrivate into an AddressTranslationPrivateWatchResponse
	spec, err := s.marshaler.Marshal(addressTranslationPrivate.Spec)
	if err != nil {
		return nil, err
	}

	status, err := s.marshaler.Marshal(addressTranslationPrivate.Status)
	if err != nil {
		return nil, err
	}

	addressTranslation := &pb.AddressTranslationPrivateWatchResponse{
		Metadata: addressTranslationPrivate.Metadata,
		Spec:     string(spec),
		Status:   string(status),
	}

	log.V(9).Info("decoded", logkeys.ADDRESS_TRANSLATION, addressTranslation)
	// Copy fields directly in the row to the address translation.
	addressTranslation.Metadata.CloudAccountId = metadata.CloudAccountId
	addressTranslation.Metadata.ResourceId = metadata.ResourceId
	addressTranslation.Metadata.ResourceVersion = metadata.ResourceVersion
	addressTranslation.Metadata.DeletionTimestamp = addressTranslationPrivate.Metadata.DeletionTimestamp
	addressTranslation.Metadata.DeletedTimestamp, err = timestampStrToPbTimestamp(deletedTimestamp)
	if err != nil {
		return nil, err
	}

	return addressTranslation, nil
}

// When using FromRow, the SQL SELECT query must select these columns.
func (s *AddressTranslationSqlTransformer) ColumnsForFromRow() string {
	cols := []string{"resource_id", "cloud_account_id", "deleted_timestamp", "resource_version", "value"}
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
