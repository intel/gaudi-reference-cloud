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

func airflowConfProtoToSql(sqlData db.AirflowConf) pb.DpaiAirflowConf {
	return pb.DpaiAirflowConf{
		Id:        sqlData.ID,
		AirflowId: sqlData.AirflowID,
		Key:       sqlData.Key,
		Value:     sqlData.Value.String,
		Metadata: &pb.DpaiMeta{
			CreatedAt: timestamppb.New(sqlData.CreatedAt.Time),
			CreatedBy: sqlData.CreatedBy,
			UpdatedAt: timestamppb.New(sqlData.UpdatedAt.Time),
			UpdatedBy: sqlData.UpdatedBy.String,
		},
	}
}

func (s *DpaiServer) DpaiAirflowConfList(ctx context.Context, req *pb.DpaiAirflowConfListRequest) (*pb.DpaiAirflowConfListResponse, error) {
	log.Println("List AirflowConf...")
	records, err := s.Sql.ListAirflowConf(context.Background(), req.GetAirflowId())
	if err != nil {
		log.Printf("Failed to list Airflow conf with error : %+v", err)
		return nil, err
	}

	var prototypes []*pb.DpaiAirflowConf
	for _, record := range records {
		prototype := airflowConfProtoToSql(record)
		prototypes = append(prototypes, &prototype)
	}
	return &pb.DpaiAirflowConfListResponse{
		Data: prototypes}, nil
}

func (s *DpaiServer) DpaiAirflowConfCreate(ctx context.Context, req *pb.DpaiAirflowConfCreateRequest) (*pb.DpaiAirflowConf, error) {

	log.Println("Create AirflowConf:")

	create, err := s.Sql.CreateAirflowConf(ctx, db.CreateAirflowConfParams{
		ID:        uuid.New().String(),
		AirflowID: req.GetAirflowId(),
		Key:       req.GetKey(),
		Value:     pgtype.Text{String: req.GetValue(), Valid: true},
		CreatedBy: "CallerOfTheAPI",
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	resp := airflowConfProtoToSql(create)
	return &resp, nil
}

func (s *DpaiServer) DpaiAirflowConfGetById(ctx context.Context, req *pb.DpaiAirflowConfGetByIdRequest) (*pb.DpaiAirflowConf, error) {

	log.Printf("Get AirflowConf By Id: %s", req.GetId())
	data, err := s.Sql.GetAirflowConfById(context.Background(), req.GetId())
	if err != nil {
		log.Printf("Failed to get the AirflowConf with error : %+v", err)
		return nil, err
	}

	resp := airflowConfProtoToSql(data)

	return &resp, nil
}

func (s *DpaiServer) DpaiAirflowConfUpdate(ctx context.Context, req *pb.DpaiAirflowConfUpdateRequest) (*pb.DpaiAirflowConf, error) {
	id := req.GetId()
	log.Printf("Update AirflowConf By Id: %s", id)
	temp, err := s.Sql.GetAirflowConfById(context.Background(), id)
	if err != nil {
		log.Printf("No record found for id : %s Error: %+v", id, err)
		return nil, err
	}

	if req.GetKey() == "" {
		req.Key = temp.Key
	}
	if req.GetValue() == "" {
		req.Value = temp.Value.String
	}

	updateUserParams := db.UpdateAirflowConfParams{
		ID:    id,
		Key:   req.GetKey(),
		Value: pgtype.Text{String: req.GetValue(), Valid: true},
	}

	data, err := s.Sql.UpdateAirflowConf(context.Background(), updateUserParams)
	if err != nil {
		log.Printf("Failed to update AirflowConf with error : %+v", err)
		return nil, err
	}
	resp := airflowConfProtoToSql(data)

	return &resp, nil
}

func (s *DpaiServer) DpaiAirflowConfDelete(ctx context.Context, req *pb.DpaiAirflowConfDeleteRequest) (*pb.DpaiAirflowConfDeleteResponse, error) {
	id := req.GetId()

	status := true
	var errorMessage string
	log.Printf("Delete AirflowConf by Id: %s", id)
	err := s.Sql.DeleteAirflowConf(context.Background(), id)
	if err != nil {
		log.Printf("Failed to delete AirflowConf with error : %+v", err)
		status = false
		errorMessage = fmt.Sprintf("Failed to delete AirflowConf with error : %+v", err)
	}

	return &pb.DpaiAirflowConfDeleteResponse{Success: status, ErrorMessage: errorMessage}, nil
}

func (s *DpaiServer) DpaiAirflowConfDeleteByAirflowId(ctx context.Context, req *pb.DpaiAirflowConfDeleteByAirflowIdRequest) (*pb.DpaiAirflowConfDeleteResponse, error) {
	id := req.GetId()

	status := true
	var errorMessage string
	log.Printf("Delete AirflowConf by AirflowId: %s", id)
	err := s.Sql.DeleteAirflowConfByAirflowId(context.Background(), id)
	if err != nil {
		log.Printf("Failed to delete AirflowConf with error : %+v", err)
		status = false
		errorMessage = fmt.Sprintf("Failed to delete AirflowConf with error : %+v", err)
	}

	return &pb.DpaiAirflowConfDeleteResponse{Success: status, ErrorMessage: errorMessage}, nil
}
