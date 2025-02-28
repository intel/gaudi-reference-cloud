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
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func postgresVersionSqlToProto(data db.PostgresVersion) (pb.DpaiPostgresVersion, error) {
	chart, err := utils.ConvertBytesToChartReference(data.ChartReference)
	if err != nil {
		return pb.DpaiPostgresVersion{}, err
	}
	return pb.DpaiPostgresVersion{
		Id:                     data.ID,
		Name:                   data.Name,
		Description:            data.Description.String,
		Version:                data.Version,
		PostgresVersion:        data.PostgresVersion,
		ImageReference:         data.ImageReference.String,
		ChartReference:         chart,
		BackwardCompatibleFrom: data.BackwardCompatibleFrom.String,
		Metadata: &pb.DpaiMeta{
			CreatedAt: timestamppb.New(data.CreatedAt.Time),
			CreatedBy: data.CreatedBy,
			UpdatedAt: timestamppb.New(data.UpdatedAt.Time),
			UpdatedBy: data.UpdatedBy.String,
		},
	}, nil
}

// TODO: Pagination
func (s *DpaiServer) DpaiPostgresVersionList(ctx context.Context, req *pb.DpaiPostgresVersionListRequest) (*pb.DpaiPostgresVersionListResponse, error) {
	log.Println("List PostgresVersion...")
	records, err := s.Sql.ListPostgresVersion(context.Background())
	if err != nil {
		log.Printf("Failed to list PostgresVersion with error : %+v", err)
		return nil, err
	}

	var results []*pb.DpaiPostgresVersion
	for _, record := range records {
		result, err := postgresVersionSqlToProto(record)
		if err != nil {
			return nil, err
		}
		results = append(results, &result)
	}
	return &pb.DpaiPostgresVersionListResponse{
		Data: results}, nil
}

func (s *DpaiServer) DpaiPostgresVersionCreate(ctx context.Context, req *pb.DpaiPostgresVersionCreateRequest) (*pb.DpaiPostgresVersion, error) {

	log.Println("Create PostgresVersion:")
	chart, err := utils.ConvertChartReferenceToBytes(req.GetChartReference())
	if err != nil {
		return nil, err
	}

	create, err := s.Sql.CreatePostgresVersion(ctx, db.CreatePostgresVersionParams{
		ID:                     uuid.New().String(),
		Name:                   req.GetName(),
		Description:            pgtype.Text{String: req.GetDescription(), Valid: true},
		Version:                req.GetVersion(),
		PostgresVersion:        req.GetPostgresVersion(),
		ImageReference:         pgtype.Text{String: req.GetImageReference(), Valid: true},
		ChartReference:         chart,
		BackwardCompatibleFrom: pgtype.Text{String: req.GetBackwardCompatibleFrom(), Valid: true},
		CreatedBy:              "CallerOfTheAPI",
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	resp, err := postgresVersionSqlToProto(create)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *DpaiServer) DpaiPostgresVersionGetById(ctx context.Context, req *pb.DpaiPostgresVersionGetByIdRequest) (*pb.DpaiPostgresVersion, error) {

	log.Printf("Get PostgresVersion By Id: %s", req.GetId())
	data, err := s.Sql.GetPostgresVersionById(context.Background(), req.GetId())
	if err != nil {
		log.Printf("Failed to list PostgresVersion with error : %+v", err)
		return nil, err
	}

	resp, err := postgresVersionSqlToProto(data)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *DpaiServer) DpaiPostgresVersionGetByName(ctx context.Context, req *pb.DpaiPostgresVersionGetByNameRequest) (*pb.DpaiPostgresVersion, error) {
	name := req.GetName()
	log.Printf("Get PostgresVersion By Name: %s", name)
	data, err := s.Sql.GetPostgresVersionByName(context.Background(), name)
	if err != nil {
		log.Printf("Failed to list PostgresVersion with error : %+v", err)
		return nil, err
	}

	resp, err := postgresVersionSqlToProto(data)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *DpaiServer) DpaiPostgresVersionUpdate(ctx context.Context, req *pb.DpaiPostgresVersionUpdateRequest) (*pb.DpaiPostgresVersion, error) {
	id := req.GetId()
	log.Printf("Update PostgresVersion By Id: %s", id)
	temp, err := s.Sql.GetPostgresVersionById(context.Background(), id)
	if err != nil {
		log.Printf("No record found for id : %s Error: %+v", id, err)
		return nil, err
	}

	if req.GetDescription() == "" {
		req.Description = temp.Description.String
	}
	if req.GetImageReference() == "" {
		req.ImageReference = temp.ImageReference.String
	}
	if req.GetBackwardCompatibleFrom() == "" {
		req.BackwardCompatibleFrom = temp.BackwardCompatibleFrom.String
	}

	existingChartReference, err := utils.ConvertBytesToChartReference(temp.ChartReference)
	if err != nil {
		return nil, err
	}
	mergedChartReference := utils.MergeChartReference(existingChartReference, req.GetChartReference())
	chart, err := utils.ConvertChartReferenceToBytes(mergedChartReference)
	if err != nil {
		return nil, err
	}
	updateUserParams := db.UpdatePostgresVersionParams{
		ID:                     id,
		Description:            pgtype.Text{String: req.Description, Valid: true},
		ImageReference:         pgtype.Text{String: req.ImageReference, Valid: true},
		ChartReference:         chart,
		BackwardCompatibleFrom: pgtype.Text{String: req.BackwardCompatibleFrom, Valid: true},
	}

	data, err := s.Sql.UpdatePostgresVersion(context.Background(), updateUserParams)
	if err != nil {
		log.Printf("Failed to update PostgresVersion with error : %+v", err)
		return nil, err
	}
	resp, err := postgresVersionSqlToProto(data)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *DpaiServer) DpaiPostgresVersionDelete(ctx context.Context, req *pb.DpaiPostgresVersionDeleteRequest) (*pb.DpaiPostgresVersionDeleteResponse, error) {
	id := req.GetId()

	status := true
	var errorMessage string
	log.Printf("Delete PostgresVersion by Id: %s", id)
	err := s.Sql.DeletePostgresVersion(context.Background(), id)
	if err != nil {
		log.Printf("Failed to delete PostgresVersion with error : %+v", err)
		status = false
		errorMessage = fmt.Sprintf("Failed to delete PostgresVersion with error : %+v", err)
	}

	return &pb.DpaiPostgresVersionDeleteResponse{Success: status, ErrorMessage: errorMessage}, nil
}
