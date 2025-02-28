// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package pbconvert

import (
	"fmt"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Convert between Protobuf messages.
type PbConverter struct {
}

func NewPbConverter() *PbConverter {
	return &PbConverter{}
}

// Convert a Protobuf message to another Protobuf message.
// This uses the binary Protobuf message as an intermediate message.
// Fields with common field numbers will be copied from the source to the target.
func (c *PbConverter) Transcode(source protoreflect.ProtoMessage, target protoreflect.ProtoMessage) error {
	marshalOptions := proto.MarshalOptions{}
	unmarshalOptions := proto.UnmarshalOptions{
		DiscardUnknown: true,
	}
	protoBytes, err := marshalOptions.Marshal(source)
	if err != nil {
		return fmt.Errorf("PbConverter.Transcode: Marshal: %w", err)
	}
	if err := unmarshalOptions.Unmarshal(protoBytes, target); err != nil {
		return fmt.Errorf("PbConverter.Transcode: Unmarshal: %w", err)
	}
	return nil
}

// Convert a Protobuf message to another Protobuf message.
// This uses JSON as an intermediate message.
// Fields with common field names will be copied from the source to the target.
func (c *PbConverter) TranscodeThroughJson(source protoreflect.ProtoMessage, target protoreflect.ProtoMessage) error {
	marshaler := &grpcruntime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			// Force fields with default values, including for enums.
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}
	jsonBytes, err := marshaler.Marshal(source)
	if err != nil {
		return fmt.Errorf("PbConverter.TranscodeThroughJson: unable to serialize to json: %w", err)
	}
	if err := marshaler.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("PbConverter.TranscodeThroughJson: unable to deserialize from json: %w", err)
	}
	return nil
}
