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

// TODO: Pagination
func (s *DpaiServer) DpaiDeploymentTaskList(ctx context.Context, req *pb.DpaiDeploymentTaskListRequest) (*pb.DpaiDeploymentTaskListResponse, error) {
	fmt.Println("List DeploymentTasks")
	deployments, err := s.SqlModel.ListDeploymentTasks(context.Background(), req.GetDeploymentId())
	if err != nil {
		log.Printf("Failed to list deployments with error : %+v", err)
		return nil, err
	}

	var results []*pb.DpaiDeploymentTask
	for _, deployment := range deployments {
		state, err := utils.ConvertDpaiDeploymentStateToPbEnum(deployment.StatusState)
		if err != nil {
			return nil, err
		}
		result := pb.DpaiDeploymentTask{
			Id:           deployment.ID,
			DeploymentId: deployment.DeploymentID,
			Name:         deployment.Name,
			Description:  deployment.Description.String,
			Status: &pb.DpaiDeploymentStatus{
				State:       *state,
				DisplayName: deployment.StatusDisplayName.String,
				Message:     deployment.StatusMessage.String,
			},
			ErrorMessage: deployment.ErrorMessage.String,
			Metadata: &pb.DpaiDeploymentTaskMeta{
				CreatedAt: timestamppb.New(deployment.CreatedAt.Time),
				UpdatedAt: timestamppb.New(deployment.UpdatedAt.Time),
			},
		}
		results = append(results, &result)
	}
	return &pb.DpaiDeploymentTaskListResponse{
		Data: results}, nil
}

func (s *DpaiServer) DpaiDeploymentTaskCreate(ctx context.Context, req *pb.DpaiDeploymentTaskCreateRequest) (*pb.DpaiDeploymentTaskCreateResponse, error) {

	fmt.Printf("Create DeploymentTask")

	create, err := s.SqlModel.CreateDeploymentTask(ctx, db.CreateDeploymentTaskParams{
		ID:           uuid.New().String(),
		DeploymentID: req.GetDeploymentId(),
		Name:         req.GetName(),
		Description:  pgtype.Text{String: req.GetDescription(), Valid: true},
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	return &pb.DpaiDeploymentTaskCreateResponse{
		DpaiDeploymentTaskId: create.ID,
	}, nil
}

func (s *DpaiServer) DpaiDeploymentTaskGet(ctx context.Context, req *pb.DpaiDeploymentTaskGetRequest) (*pb.DpaiDeploymentTaskGetResponse, error) {
	id := req.GetId()
	fmt.Printf("Get DeploymentTask By Id: %s", id)
	deployment, err := s.SqlModel.GetDeploymentTask(context.Background(), id)
	if err != nil {
		log.Printf("Failed to get deployment with error : %+v", err)
		return nil, fmt.Errorf("failed to get deployment with error : %+v", err)
	}
	state, err := utils.ConvertDpaiDeploymentStateToPbEnum(deployment.StatusState)
	if err != nil {
		return nil, err
	}
	result := pb.DpaiDeploymentTask{
		Id:           deployment.ID,
		DeploymentId: deployment.DeploymentID,
		Name:         deployment.Name,
		Description:  deployment.Description.String,
		Input:        deployment.InputPayload,
		Output:       deployment.OutputPayload,
		Status: &pb.DpaiDeploymentStatus{
			State:       *state,
			DisplayName: deployment.StatusDisplayName.String,
			Message:     deployment.StatusMessage.String,
		},
		ErrorMessage: deployment.ErrorMessage.String,
		Metadata: &pb.DpaiDeploymentTaskMeta{
			CreatedAt: timestamppb.New(deployment.CreatedAt.Time),
			UpdatedAt: timestamppb.New(deployment.UpdatedAt.Time),
		},
	}

	return &pb.DpaiDeploymentTaskGetResponse{
		DpaiDeploymentTask: &result}, nil
}

func (s *DpaiServer) DpaiDeploymentTaskUpdate(ctx context.Context, req *pb.DpaiDeploymentTaskUpdateRequest) (*pb.DpaiDeploymentTaskUpdateResponse, error) {
	id := req.GetId()
	fmt.Printf("Update DeploymentTask By Id: %s", id)
	temp, err := s.SqlModel.GetUpdateDeploymentTask(context.Background(), id)
	if err != nil {
		log.Printf("No record found for id : %s Error: %+v", id, err)
		return nil, err
	}

	errorMessage := req.GetErrorMessage()
	if errorMessage == "" {
		errorMessage = temp.ErrorMessage.String
	}
	if req.GetStatus() == nil {
		req.Status = &pb.DpaiDeploymentStatus{}
	}
	updateUserParams := db.UpdateDeploymentTaskParams{
		ID:                id,
		StatusState:       req.Status.GetState().String(),
		StatusDisplayName: pgtype.Text{String: req.Status.GetDisplayName(), Valid: req.Status.GetDisplayName() != ""},
		StatusMessage:     pgtype.Text{String: req.Status.GetMessage(), Valid: req.Status.GetMessage() != ""},
		ErrorMessage:      pgtype.Text{String: errorMessage, Valid: true},
	}

	deployment, err := s.SqlModel.UpdateDeploymentTask(context.Background(), updateUserParams)
	if err != nil {
		log.Printf("Failed to update deployment task with error : %+v", err)
		return nil, err
	}
	state, err := utils.ConvertDpaiDeploymentStateToPbEnum(deployment.StatusState)
	if err != nil {
		return nil, err
	}
	result := pb.DpaiDeploymentTask{
		Id:           deployment.ID,
		DeploymentId: deployment.DeploymentID,
		Name:         deployment.Name,
		Description:  deployment.Description.String,
		Status: &pb.DpaiDeploymentStatus{
			State:       *state,
			DisplayName: deployment.StatusDisplayName.String,
			Message:     deployment.StatusMessage.String,
		},
		ErrorMessage: deployment.ErrorMessage.String,
		Metadata: &pb.DpaiDeploymentTaskMeta{
			CreatedAt: timestamppb.New(deployment.CreatedAt.Time),
			UpdatedAt: timestamppb.New(deployment.UpdatedAt.Time),
		},
	}

	return &pb.DpaiDeploymentTaskUpdateResponse{DpaiDeploymentTask: &result}, nil
}

func (s *DpaiServer) DpaiDeploymentTaskDelete(ctx context.Context, req *pb.DpaiDeploymentTaskDeleteRequest) (*pb.DpaiDeploymentTaskDeleteResponse, error) {
	id := req.GetId()

	status := true
	fmt.Printf("Delete DeploymentTask by Id: %s", id)
	err := s.SqlModel.DeleteDeploymentTask(context.Background(), id)
	if err != nil {
		log.Printf("Failed to delete deployment with error : %+v", err)
		status = false
	}

	return &pb.DpaiDeploymentTaskDeleteResponse{Success: status}, nil
}
