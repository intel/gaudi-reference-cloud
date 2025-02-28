// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package instance_type

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type InstanceTypeService struct {
	pb.UnimplementedInstanceTypeServiceServer
	db             *sql.DB
	protoJsonTable *protodb.ProtoJsonTable
}

func NewInstanceTypeService(db *sql.DB) (*InstanceTypeService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &InstanceTypeService{
		db:             db,
		protoJsonTable: NewInstanceTypeProtoJsonTable(db),
	}, nil
}

func NewInstanceTypeProtoJsonTable(db *sql.DB) *protodb.ProtoJsonTable {
	return &protodb.ProtoJsonTable{
		Db:                 db,
		TableName:          "instance_type",
		KeyColumns:         []string{"name"},
		JsonDocumentColumn: "value",
		ConflictTarget:     "name",
		EmptyMessage:       &pb.InstanceType{},
		GetKeyValuesFunc: func(m proto.Message) ([]any, error) {
			var name string
			if msg, ok := m.(*pb.InstanceType); ok {
				name = msg.Metadata.Name
			} else if msg, ok := m.(*pb.InstanceTypeGetRequest); ok {
				name = msg.Metadata.Name
			} else if msg, ok := m.(*pb.InstanceTypeDeleteRequest); ok {
				name = msg.Metadata.Name
			} else {
				return nil, fmt.Errorf("unsupported type")
			}
			return []any{name}, nil
		},
	}
}

func (s *InstanceTypeService) Put(ctx context.Context, req *pb.InstanceType) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("InstanceTypeService.Put")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if req.Metadata.Name == "" {
			return nil, status.Error(codes.InvalidArgument, "missing metadata.name")
		}
		if req.Spec == nil {
			return nil, status.Error(codes.InvalidArgument, "missing spec")
		}
		if err := s.protoJsonTable.Put(ctx, req); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *InstanceTypeService) Delete(ctx context.Context, req *pb.InstanceTypeDeleteRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("InstanceTypeService.Delete")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if err := s.protoJsonTable.Delete(ctx, req); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *InstanceTypeService) Get(ctx context.Context, req *pb.InstanceTypeGetRequest) (*pb.InstanceType, error) {
	logger := log.FromContext(ctx).WithName("InstanceTypeService.Get")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.InstanceType, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		resp, err := s.protoJsonTable.Get(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.(*pb.InstanceType), nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *InstanceTypeService) Search(ctx context.Context, req *pb.InstanceTypeSearchRequest) (*pb.InstanceTypeSearchResponse, error) {
	logger := log.FromContext(ctx).WithName("InstanceTypeService.Search")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.InstanceTypeSearchResponse, error) {
		items := make([]*pb.InstanceType, 0)
		handlerFunc := func(m proto.Message) error {
			items = append(items, m.(*pb.InstanceType))
			return nil
		}
		if err := s.protoJsonTable.Search(ctx, req, handlerFunc); err != nil {
			return nil, err
		}
		resp := &pb.InstanceTypeSearchResponse{
			Items: items,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *InstanceTypeService) SearchStream(req *pb.InstanceTypeSearchRequest, svc pb.InstanceTypeService_SearchStreamServer) error {
	ctx := svc.Context()
	logger := log.FromContext(ctx).WithName("InstanceTypeService.SearchStream")
	logger.Info("BEGIN", logkeys.Request, req)
	defer logger.Info("END")
	err := func() error {
		handlerFunc := func(m proto.Message) error {
			return svc.Send(m.(*pb.InstanceType))
		}
		if err := s.protoJsonTable.Search(ctx, req, handlerFunc); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		logger.Error(err, logkeys.Error, logkeys.Request, req)
	}
	return utils.SanitizeError(err)
}
