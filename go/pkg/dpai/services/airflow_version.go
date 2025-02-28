// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package services

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/jackc/pgx/v5/pgtype"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func airflowVersionSqlToProto(data db.AirflowVersion) (pb.DpaiAirflowVersion, error) {
	chart, err := utils.ConvertBytesToChartReference(data.ChartReference)
	if err != nil {
		return pb.DpaiAirflowVersion{}, err
	}
	image, err := utils.ConvertBytesToImageReference(data.ImageReference)
	if err != nil {
		return pb.DpaiAirflowVersion{}, err
	}
	return pb.DpaiAirflowVersion{
		Id:                       data.ID,
		Name:                     data.Name,
		Version:                  data.Version,
		BackendDatabaseVersionId: data.BackendDatabaseVersionID,
		AirflowVersionProperties: &pb.DpaiAirflowVersionProperties{
			AirflowVersion:  data.AirflowVersion,
			PythonVersion:   data.PythonVersion,
			PostgresVersion: data.PostgresVersion,
			RedisVersion:    data.RedisVersion.String,
		},
		ExecutorType:           data.ExecutorType.String,
		ImageReference:         image,
		ChartReference:         chart,
		Description:            data.Description.String,
		BackwardCompatibleFrom: data.BackwardCompatibleFrom.String,
		Metadata: &pb.DpaiMeta{
			CreatedAt: timestamppb.New(data.CreatedAt.Time),
			CreatedBy: data.CreatedBy,
			UpdatedAt: timestamppb.New(data.UpdatedAt.Time),
			UpdatedBy: data.UpdatedBy.String,
		},
	}, nil
}

func (s *DpaiServer) DpaiAirflowVersionList(ctx context.Context, req *pb.DpaiAirflowVersionListRequest) (*pb.DpaiAirflowVersionListResponse, error) {
	log.Println("List Airflow Versions...")
	records, err := s.Sql.ListAirflowVersion(context.Background())
	if err != nil {
		log.Printf("Failed to list airflow version with error : %+v", err)
		return nil, err
	}

	var prototypes []*pb.DpaiAirflowVersion
	for _, record := range records {
		prototype, err := airflowVersionSqlToProto(record)
		if err != nil {
			return nil, err
		}
		prototypes = append(prototypes, &prototype)
	}
	return &pb.DpaiAirflowVersionListResponse{
		Data: prototypes}, nil
}

func (s *DpaiServer) DpaiAirflowVersionCreate(ctx context.Context, req *pb.DpaiAirflowVersionCreateRequest) (*pb.DpaiAirflowVersion, error) {

	log.Println("Create AirflowVersion:")

	chart, err := utils.ConvertChartReferenceToBytes(req.GetChartReference())
	if err != nil {
		return nil, err
	}
	image, err := utils.ConvertImageReferenceToBytes(req.GetImageReference())
	if err != nil {
		return nil, err
	}
	create, err := s.Sql.CreateAirflowVersion(ctx, db.CreateAirflowVersionParams{
		ID:                       uuid.New().String(),
		Name:                     req.GetName(),
		Version:                  req.GetVersion(),
		BackendDatabaseVersionID: req.GetBackendDatabaseVersionId(),
		AirflowVersion:           req.AirflowVersionProperties.GetAirflowVersion(),
		PythonVersion:            req.AirflowVersionProperties.GetPythonVersion(),
		PostgresVersion:          req.AirflowVersionProperties.GetPostgresVersion(),
		RedisVersion:             pgtype.Text{String: req.AirflowVersionProperties.GetRedisVersion(), Valid: true},
		ExecutorType:             pgtype.Text{String: req.GetExecutorType(), Valid: true},
		ImageReference:           image,
		ChartReference:           chart,
		Description:              pgtype.Text{String: req.GetDescription(), Valid: true},
		BackwardCompatibleFrom:   pgtype.Text{String: req.GetBackwardCompatibleFrom(), Valid: true},
		CreatedBy:                "CallerOfTheAPI",
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	resp, err := airflowVersionSqlToProto(create)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *DpaiServer) DpaiAirflowVersionGetById(ctx context.Context, req *pb.DpaiAirflowVersionGetByIdRequest) (*pb.DpaiAirflowVersion, error) {

	log.Printf("Get AirflowVersion By Id: %s", req.GetId())
	data, err := s.Sql.GetAirflowVersionById(context.Background(), req.GetId())
	if err != nil {
		log.Printf("Failed to list AirflowVersion with error : %+v", err)
		return nil, err
	}

	resp, err := airflowVersionSqlToProto(data)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *DpaiServer) DpaiAirflowVersionGetByName(ctx context.Context, req *pb.DpaiAirflowVersionGetByNameRequest) (*pb.DpaiAirflowVersion, error) {
	name := req.GetName()
	log.Printf("Get AirflowVersion By Name: %s", name)
	data, err := s.Sql.GetAirflowVersionByName(context.Background(), name)
	if err != nil {
		log.Printf("Failed to list AirflowVersion with error : %+v", err)
		return nil, err
	}

	resp, err := airflowVersionSqlToProto(data)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *DpaiServer) DpaiAirflowVersionUpdate(ctx context.Context, req *pb.DpaiAirflowVersionUpdateRequest) (*pb.DpaiAirflowVersion, error) {
	id := req.GetId()
	log.Printf("Update AirflowVersion By Id: %s", id)
	temp, err := s.Sql.GetAirflowVersionById(context.Background(), id)
	if err != nil {
		log.Printf("No record found for id : %s Error: %+v", id, err)
		return nil, err
	}

	if req.GetDescription() == "" {
		req.Description = temp.Description.String
	}

	if req.GetBackwardCompatibleFrom() == "" {
		req.BackwardCompatibleFrom = temp.BackwardCompatibleFrom.String
	}

	existingImageReference, err := utils.ConvertBytesToImageReference(temp.ImageReference)
	if err != nil {
		return nil, err
	}
	mergedImageReference := utils.MergeImageReference(existingImageReference, req.GetImageReference())

	existingChartReference, err := utils.ConvertBytesToChartReference(temp.ChartReference)
	if err != nil {
		return nil, err
	}
	mergedChartReference := utils.MergeChartReference(existingChartReference, req.GetChartReference())
	chart, err := utils.ConvertChartReferenceToBytes(mergedChartReference)
	if err != nil {
		return nil, err
	}
	image, err := utils.ConvertImageReferenceToBytes(mergedImageReference)
	if err != nil {
		return nil, err
	}
	updateUserParams := db.UpdateAirflowVersionParams{
		ID:                     id,
		Version:                req.Version,
		ImageReference:         image,
		ChartReference:         chart,
		Description:            pgtype.Text{String: req.Description, Valid: true},
		BackwardCompatibleFrom: pgtype.Text{String: req.BackwardCompatibleFrom, Valid: true},
	}

	data, err := s.Sql.UpdateAirflowVersion(context.Background(), updateUserParams)
	if err != nil {
		log.Printf("Failed to update AirflowVersion with error : %+v", err)
		return nil, err
	}
	resp, err := airflowVersionSqlToProto(data)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *DpaiServer) DpaiAirflowVersionDelete(ctx context.Context, req *pb.DpaiAirflowVersionDeleteRequest) (*pb.DpaiAirflowVersionDeleteResponse, error) {
	id := req.GetId()

	status := true
	var errorMessage string
	log.Printf("Delete AirflowVersion by Id: %s", id)
	err := s.Sql.DeleteAirflowVersion(context.Background(), id)
	if err != nil {
		log.Printf("Failed to delete AirflowVersion with error : %+v", err)
		status = false
		errorMessage = fmt.Sprintf("Failed to delete AirflowVersion with error : %+v", err)
	}

	return &pb.DpaiAirflowVersionDeleteResponse{Success: status, ErrorMessage: errorMessage}, nil
}
