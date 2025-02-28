// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package services

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func postgresSizeProtoToSql(data db.PostgresSize) pb.DpaiPostgresSize {
	return pb.DpaiPostgresSize{
		Id:                       data.ID,
		Name:                     data.Name,
		Description:              data.Description.String,
		NumberOfInstancesDefault: data.NumberOfInstancesDefault,
		NumberOfInstancesMin:     data.NumberOfInstancesMin.Int32,
		NumberOfInstancesMax:     data.NumberOfInstancesMax.Int32,
		ResourceCpuLimit:         data.ResourceCpuLimit.String,
		ResourceMemoryLimit:      data.ResourceMemoryLimit.String,
		ResourceCpuRequest:       data.ResourceCpuRequest.String,
		ResourceMemoryRequest:    data.ResourceMemoryLimit.String,
		PgPoolProperties: &pb.DpaiPostgresSizePgPoolProperties{
			NumberOfInstancesDefault: data.NumberOfPgpoolInstancesDefault,
			NumberOfInstancesMin:     data.NumberOfPgpoolInstancesMin.Int32,
			NumberOfInstancesMax:     data.NumberOfPgpoolInstancesMax.Int32,
			ResourceCpuLimit:         data.ResourcePgpoolCpuLimit.String,
			ResourceMemoryLimit:      data.ResourcePgpoolMemoryLimit.String,
			ResourceCpuRequest:       data.ResourcePgpoolCpuRequest.String,
			ResourceMemoryRequest:    data.ResourcePgpoolMemoryLimit.String,
		},
		DiskProperties: &pb.DpaiPostgresSizeDiskProperties{
			DiskSizeInGbDefault: data.DiskSizeInGbDefault,
			DiskSizeInGbMin:     data.DiskSizeInGbMin.Int32,
			DiskSizeInGbMax:     data.DiskSizeInGbMax.Int32,
		},
		Metadata: &pb.DpaiMeta{
			CreatedAt: timestamppb.New(data.CreatedAt.Time),
			CreatedBy: data.CreatedBy,
			UpdatedAt: timestamppb.New(data.UpdatedAt.Time),
			UpdatedBy: data.UpdatedBy.String,
		},
	}
}

// TODO: Pagination
func (s *DpaiServer) DpaiPostgresSizeList(ctx context.Context, req *pb.DpaiPostgresSizeListRequest) (*pb.DpaiPostgresSizeListResponse, error) {
	log.Println("List PostgresSize...")
	records, err := s.Sql.ListPostgresSize(context.Background())
	if err != nil {
		log.Printf("Failed to list PostgresSize with error : %+v", err)
		return nil, err
	}

	var results []*pb.DpaiPostgresSize
	for _, record := range records {
		result := postgresSizeProtoToSql(record)
		results = append(results, &result)
	}
	return &pb.DpaiPostgresSizeListResponse{
		Data: results}, nil
}

func (s *DpaiServer) DpaiPostgresSizeCreate(ctx context.Context, req *pb.DpaiPostgresSizeCreateRequest) (*pb.DpaiPostgresSize, error) {

	log.Println("Create PostgresSize:")

	create, err := s.Sql.CreatePostgresSize(ctx, db.CreatePostgresSizeParams{
		ID:                       uuid.New().String(),
		Name:                     req.GetName(),
		Description:              pgtype.Text{String: req.GetDescription(), Valid: true},
		NumberOfInstancesDefault: req.GetNumberOfInstancesDefault(),
		NumberOfInstancesMin:     pgtype.Int4{Int32: req.GetNumberOfInstancesMin(), Valid: true},
		NumberOfInstancesMax:     pgtype.Int4{Int32: req.GetNumberOfInstancesMax(), Valid: true},
		ResourceCpuLimit:         pgtype.Text{String: req.GetResourceCpuLimit(), Valid: true},
		ResourceCpuRequest:       pgtype.Text{String: req.GetResourceCpuRequest(), Valid: true},
		ResourceMemoryLimit:      pgtype.Text{String: req.GetResourceMemoryLimit(), Valid: true},
		ResourceMemoryRequest:    pgtype.Text{String: req.GetResourceMemoryRequest(), Valid: true},

		NumberOfPgpoolInstancesDefault: req.PgPoolProperties.GetNumberOfInstancesDefault(),
		NumberOfPgpoolInstancesMin:     pgtype.Int4{Int32: req.PgPoolProperties.GetNumberOfInstancesMin(), Valid: true},
		NumberOfPgpoolInstancesMax:     pgtype.Int4{Int32: req.PgPoolProperties.GetNumberOfInstancesMax(), Valid: true},
		ResourcePgpoolCpuLimit:         pgtype.Text{String: req.PgPoolProperties.GetResourceCpuLimit(), Valid: true},
		ResourcePgpoolCpuRequest:       pgtype.Text{String: req.PgPoolProperties.GetResourceCpuRequest(), Valid: true},
		ResourcePgpoolMemoryLimit:      pgtype.Text{String: req.PgPoolProperties.GetResourceMemoryLimit(), Valid: true},
		ResourcePgpoolMemoryRequest:    pgtype.Text{String: req.PgPoolProperties.GetResourceMemoryRequest(), Valid: true},

		DiskSizeInGbDefault: req.DiskProperties.GetDiskSizeInGbDefault(),
		DiskSizeInGbMin:     pgtype.Int4{Int32: req.DiskProperties.GetDiskSizeInGbMin(), Valid: true},
		DiskSizeInGbMax:     pgtype.Int4{Int32: req.DiskProperties.GetDiskSizeInGbMax(), Valid: true},

		CreatedBy: "CallerOfTheAPI",
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	resp := postgresSizeProtoToSql(create)
	return &resp, nil
}

func (s *DpaiServer) DpaiPostgresSizeGetById(ctx context.Context, req *pb.DpaiPostgresSizeGetByIdRequest) (*pb.DpaiPostgresSize, error) {

	log.Printf("Get PostgresSize By Id: %s", req.GetId())
	data, err := s.Sql.GetPostgresSizeById(context.Background(), req.GetId())
	if err != nil {
		log.Printf("Failed to list PostgresSize with error : %+v", err)
		return nil, err
	}

	resp := postgresSizeProtoToSql(data)

	return &resp, nil
}

func (s *DpaiServer) DpaiPostgresSizeGetByName(ctx context.Context, req *pb.DpaiPostgresSizeGetByNameRequest) (*pb.DpaiPostgresSize, error) {
	name := req.GetName()
	log.Printf("Get PostgresSize By Name: %s", name)
	data, err := s.Sql.GetPostgresSizeByName(context.Background(), name)
	if err != nil {
		log.Printf("Failed to list PostgresSize with error : %+v", err)
		return nil, err
	}

	resp := postgresSizeProtoToSql(data)

	return &resp, nil
}

func (s *DpaiServer) DpaiPostgresSizeUpdate(ctx context.Context, req *pb.DpaiPostgresSizeUpdateRequest) (*pb.DpaiPostgresSize, error) {
	id := req.GetId()
	log.Printf("Update PostgresSize By Id: %s", id)
	temp, err := s.Sql.GetPostgresSizeById(context.Background(), id)
	if err != nil {
		log.Printf("No record found for id : %s Error: %+v", id, err)
		return nil, err
	}

	description := req.GetDescription()
	if description == "" {
		description = temp.Description.String
	}

	numberOfInstancesDefault := req.GetNumberOfInstancesDefault()
	if numberOfInstancesDefault == 0 {
		numberOfInstancesDefault = temp.NumberOfInstancesDefault
	}
	numberOfInstancesMin := req.GetNumberOfInstancesMin()
	if numberOfInstancesMin == 0 {
		numberOfInstancesMin = temp.NumberOfInstancesMin.Int32
	}
	numberOfInstancesMax := req.GetNumberOfInstancesMax()
	if numberOfInstancesMax == 0 {
		numberOfInstancesMax = temp.NumberOfInstancesMax.Int32
	}
	resourceCpuLimit := req.GetResourceCpuLimit()
	if resourceCpuLimit == "" {
		resourceCpuLimit = temp.ResourceCpuLimit.String
	}
	resourceCpuRequest := req.GetResourceCpuRequest()
	if resourceCpuRequest == "" {
		resourceCpuRequest = temp.ResourceCpuRequest.String
	}
	resourceMemoryLimit := req.GetResourceMemoryLimit()
	if resourceMemoryLimit == "" {
		resourceMemoryLimit = temp.ResourceMemoryLimit.String
	}
	resourceMemoryRequest := req.GetResourceMemoryRequest()
	if resourceMemoryRequest == "" {
		resourceMemoryRequest = temp.ResourceMemoryRequest.String
	}

	numberOfPgpoolInstancesDefault := req.PgPoolProperties.GetNumberOfInstancesDefault()
	if numberOfPgpoolInstancesDefault == 0 {
		numberOfPgpoolInstancesDefault = temp.NumberOfPgpoolInstancesDefault
	}
	numberOfPgpoolInstancesMin := req.PgPoolProperties.GetNumberOfInstancesMin()
	if numberOfPgpoolInstancesMin == 0 {
		numberOfPgpoolInstancesMin = temp.NumberOfPgpoolInstancesMin.Int32
	}
	numberOfPgpoolInstancesMax := req.PgPoolProperties.GetNumberOfInstancesMax()
	if numberOfPgpoolInstancesMax == 0 {
		numberOfPgpoolInstancesMax = temp.NumberOfPgpoolInstancesMax.Int32
	}
	resourcePgpoolCpuLimit := req.PgPoolProperties.GetResourceCpuLimit()
	if resourcePgpoolCpuLimit == "" {
		resourcePgpoolCpuLimit = temp.ResourceCpuLimit.String
	}
	resourcePgpoolCpuRequest := req.PgPoolProperties.GetResourceCpuRequest()
	if resourcePgpoolCpuRequest == "" {
		resourcePgpoolCpuRequest = temp.ResourceCpuRequest.String
	}
	resourcePgpoolMemoryLimit := req.PgPoolProperties.GetResourceMemoryLimit()
	if resourcePgpoolMemoryLimit == "" {
		resourcePgpoolMemoryLimit = temp.ResourceMemoryLimit.String
	}
	resourcePgpoolMemoryRequest := req.PgPoolProperties.GetResourceMemoryRequest()
	if resourcePgpoolMemoryRequest == "" {
		resourcePgpoolMemoryRequest = temp.ResourceMemoryRequest.String
	}

	diskSizeInGbDefault := req.DiskProperties.GetDiskSizeInGbDefault()
	if diskSizeInGbDefault == 0 {
		diskSizeInGbDefault = temp.DiskSizeInGbDefault
	}
	diskSizeInGbMax := req.DiskProperties.GetDiskSizeInGbMax()
	if diskSizeInGbMax == 0 {
		diskSizeInGbMax = temp.DiskSizeInGbMax.Int32
	}
	diskSizeInGbMin := req.DiskProperties.GetDiskSizeInGbMin()
	if diskSizeInGbMin == 0 {
		diskSizeInGbMin = temp.DiskSizeInGbMin.Int32
	}
	storageClassName := req.DiskProperties.GetStorageClassName()
	if storageClassName == "" {
		storageClassName = temp.StorageClassName.String
	}

	updateUserParams := db.UpdatePostgresSizeParams{
		ID:                             id,
		Description:                    pgtype.Text{String: description, Valid: true},
		NumberOfInstancesDefault:       numberOfInstancesDefault,
		NumberOfInstancesMin:           pgtype.Int4{Int32: numberOfInstancesMin, Valid: true},
		NumberOfInstancesMax:           pgtype.Int4{Int32: numberOfInstancesMax, Valid: true},
		ResourceCpuLimit:               pgtype.Text{String: resourceCpuLimit, Valid: true},
		ResourceCpuRequest:             pgtype.Text{String: resourceCpuRequest, Valid: true},
		ResourceMemoryLimit:            pgtype.Text{String: resourceMemoryLimit, Valid: true},
		ResourceMemoryRequest:          pgtype.Text{String: resourceMemoryRequest, Valid: true},
		NumberOfPgpoolInstancesDefault: numberOfPgpoolInstancesDefault,
		NumberOfPgpoolInstancesMin:     pgtype.Int4{Int32: numberOfPgpoolInstancesMin, Valid: true},
		NumberOfPgpoolInstancesMax:     pgtype.Int4{Int32: numberOfPgpoolInstancesMax, Valid: true},
		ResourcePgpoolCpuLimit:         pgtype.Text{String: resourcePgpoolCpuLimit, Valid: true},
		ResourcePgpoolCpuRequest:       pgtype.Text{String: resourcePgpoolCpuRequest, Valid: true},
		ResourcePgpoolMemoryLimit:      pgtype.Text{String: resourcePgpoolMemoryLimit, Valid: true},
		ResourcePgpoolMemoryRequest:    pgtype.Text{String: resourcePgpoolMemoryRequest, Valid: true},
		DiskSizeInGbDefault:            diskSizeInGbDefault,
		DiskSizeInGbMin:                pgtype.Int4{Int32: diskSizeInGbMin, Valid: true},
		DiskSizeInGbMax:                pgtype.Int4{Int32: diskSizeInGbMax, Valid: true},
		StorageClassName:               pgtype.Text{String: storageClassName, Valid: true},
	}

	data, err := s.Sql.UpdatePostgresSize(context.Background(), updateUserParams)
	if err != nil {
		log.Printf("Failed to update PostgresSize with error : %+v", err)
		return nil, err
	}
	resp := postgresSizeProtoToSql(data)

	return &resp, nil
}

func (s *DpaiServer) DpaiPostgresSizeDelete(ctx context.Context, req *pb.DpaiPostgresSizeDeleteRequest) (*pb.DpaiPostgresSizeDeleteResponse, error) {
	id := req.GetId()

	status := true
	var errorMessage string
	log.Printf("Delete PostgresSize by Id: %s", id)
	err := s.Sql.DeletePostgresSize(context.Background(), id)
	if err != nil {
		log.Printf("Failed to delete PostgresSize with error : %+v", err)
		status = false
		errorMessage = fmt.Sprintf("Failed to delete PostgresSize with error : %+v", err)
	}

	return &pb.DpaiPostgresSizeDeleteResponse{Success: status, ErrorMessage: errorMessage}, nil
}
