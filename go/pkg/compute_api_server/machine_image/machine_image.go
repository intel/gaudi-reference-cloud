// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package machine_image

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

type MachineImageService struct {
	pb.UnimplementedMachineImageServiceServer
	db             *sql.DB
	protoJsonTable *protodb.ProtoJsonTable
}

func NewMachineImageService(db *sql.DB) (*MachineImageService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &MachineImageService{
		db:             db,
		protoJsonTable: NewMachineImageProtoJsonTable(db),
	}, nil
}

func NewMachineImageProtoJsonTable(db *sql.DB) *protodb.ProtoJsonTable {
	return &protodb.ProtoJsonTable{
		Db:                 db,
		TableName:          "machine_image",
		KeyColumns:         []string{"name"},
		JsonDocumentColumn: "value",
		ConflictTarget:     "name",
		EmptyMessage:       &pb.MachineImage{},
		GetKeyValuesFunc: func(m proto.Message) ([]any, error) {
			var name string
			if msg, ok := m.(*pb.MachineImage); ok {
				name = msg.Metadata.Name
			} else if msg, ok := m.(*pb.MachineImageGetRequest); ok {
				name = msg.Metadata.Name
			} else if msg, ok := m.(*pb.MachineImageDeleteRequest); ok {
				name = msg.Metadata.Name
			} else {
				return nil, fmt.Errorf("unsupported type")
			}
			return []any{name}, nil
		},
		SearchFilterFunc: func(m proto.Message) (protodb.Flattened, error) {
			flattened := protodb.Flattened{}
			if msg, ok := m.(*pb.MachineImageSearchRequest); ok {
				if msg.Metadata == nil || msg.Metadata.InstanceType == "" {
					// no filter is passed, return all elements
					return flattened, nil
				} else {
					// search for entries which contain the specified instanceType in the array of values
					column := "value->'spec'->'instanceTypes'"
					flattened.Add(column, fmt.Sprintf("%q", msg.Metadata.InstanceType))
					return flattened, nil
				}
			} else {
				return flattened, fmt.Errorf("unsupported type")
			}
		},
	}
}

func (s *MachineImageService) Put(ctx context.Context, req *pb.MachineImage) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("MachineImageService.Put")
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

		// Validate VirtualMachine image name is not greater than 32 characters
		//TODO: revise the allowedList below once these images are no longer in use
		allowedList := map[string]bool{
			"iks-vm-u22-cd-cp-1-27-11-v20240227":             true,
			"iks-vm-u22-cd-cp-1-28-7-v20240227":              true,
			"iks-vm-u22-cd-wk-1-28-7-v20240227":              true,
			"iks-vm-u22-cd-cp-1-29-2-v20240227":              true,
			"iks-vm-u22-cd-wk-1-27-11-v20240227":             true,
			"iks-vm-u22-cd-wk-1-29-2-v20240227":              true,
			"iks-vm-u22-cd-cp-1-28-7-v20240813":              true,
			"iks-vm-u22-cd-cp-1-28-7-with-logging-v20240715": true,
		}

		for _, category := range req.Spec.InstanceCategories {
			if category == pb.InstanceCategory_VirtualMachine {
				if len(req.Metadata.Name) > 32 {
					if allowedList[req.Metadata.Name] {
						logger.Info("Name exceeds length but is in allowed list", "name", req.Metadata.Name)
						continue
					}
					return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "virtualMachine image name must be less than 32 characters")
				}
				break
			}
		}
		if err := s.protoJsonTable.Put(ctx, req); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *MachineImageService) Delete(ctx context.Context, req *pb.MachineImageDeleteRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("MachineImageService.Delete")
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

func (s *MachineImageService) Get(ctx context.Context, req *pb.MachineImageGetRequest) (*pb.MachineImage, error) {
	logger := log.FromContext(ctx).WithName("MachineImageService.Get")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.MachineImage, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		resp, err := s.protoJsonTable.Get(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp.(*pb.MachineImage), nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *MachineImageService) Search(ctx context.Context, req *pb.MachineImageSearchRequest) (*pb.MachineImageSearchResponse, error) {
	logger := log.FromContext(ctx).WithName("MachineImageService.Search")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.MachineImageSearchResponse, error) {
		items := make([]*pb.MachineImage, 0)
		handlerFunc := func(m proto.Message) error {
			if !m.(*pb.MachineImage).Spec.Hidden {
				items = append(items, m.(*pb.MachineImage))
			}
			return nil
		}
		if err := s.protoJsonTable.SearchContains(ctx, req, handlerFunc); err != nil {
			return nil, err
		}
		resp := &pb.MachineImageSearchResponse{
			Items: items,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *MachineImageService) SearchStream(req *pb.MachineImageSearchRequest, svc pb.MachineImageService_SearchStreamServer) error {
	ctx := svc.Context()
	logger := log.FromContext(ctx).WithName("MachineImageService.SearchStream")
	logger.Info("BEGIN", logkeys.Request, req)
	defer logger.Info("END")
	err := func() error {
		handlerFunc := func(m proto.Message) error {
			if !m.(*pb.MachineImage).Spec.Hidden {
				return svc.Send(m.(*pb.MachineImage))
			}
			return nil
		}
		if err := s.protoJsonTable.SearchContains(ctx, req, handlerFunc); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		logger.Error(err, logkeys.Error, logkeys.Request, req)
	}
	return utils.SanitizeError(err)
}
