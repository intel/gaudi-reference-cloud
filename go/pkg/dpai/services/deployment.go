// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package services

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment/airflow"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment/workspace"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TriggerDeployment(id string, serviceType pb.DpaiServiceType, changeIndicator pb.DpaiDeploymentChangeIndicator, server DpaiServer) (bool, error) {

	deploymentInputContext := deployment.DeploymentInputContext{
		ID:      id,
		SqlPool: server.SqlPool,
		Conf:    server.Config,
	}

	switch serviceType {
	// Workspace
	case pb.DpaiServiceType_DPAI_WORKSPACE:
		switch changeIndicator {
		case pb.DpaiDeploymentChangeIndicator_DPAI_CREATE:
			go workspace.CreateDeployment(deploymentInputContext)
		case pb.DpaiDeploymentChangeIndicator_DPAI_DELETE:
			go workspace.DeleteDeployment(deploymentInputContext)
		default:
			return false, fmt.Errorf("provided service type %s doesnt have a deployment for change indicator `%s`", serviceType, changeIndicator)
		}
	// Airflow
	case pb.DpaiServiceType_DPAI_AIRFLOW:
		switch changeIndicator {
		case pb.DpaiDeploymentChangeIndicator_DPAI_CREATE:
			go airflow.CreateDeployment(deploymentInputContext)
		case pb.DpaiDeploymentChangeIndicator_DPAI_DELETE:
			go workspace.DeleteDeployment(deploymentInputContext)
		case pb.DpaiDeploymentChangeIndicator_DPAI_UPGRADE:
			go airflow.UpgradeDeployment(deploymentInputContext)
		case pb.DpaiDeploymentChangeIndicator_DPAI_RESIZE:
			go airflow.ResizeDeployment(deploymentInputContext)
		case pb.DpaiDeploymentChangeIndicator_DPAI_RESTART:
			go airflow.RestartDeployment(deploymentInputContext)
		default:
			return false, fmt.Errorf("provided service type %s doesnt have a deployment for change indicator `%s`", serviceType, changeIndicator)
		}
	// // HMS
	// case pb.DpaiServiceType_DPAI_HMS:
	// 	switch changeIndicator {
	// 	case pb.DpaiDeploymentChangeIndicator_DPAI_CREATE:
	// 		go hms.CreateDeployment(deploymentInputContext)
	// 	case pb.DpaiDeploymentChangeIndicator_DPAI_DELETE:
	// 		go hms.DeleteDeployment(deploymentInputContext)
	// 	case pb.DpaiDeploymentChangeIndicator_DPAI_UPGRADE:
	// 		go hms.UpgradeDeployment(deploymentInputContext)
	// 	case pb.DpaiDeploymentChangeIndicator_DPAI_RESIZE:
	// 		go hms.ResizeDeployment(deploymentInputContext)
	// 	case pb.DpaiDeploymentChangeIndicator_DPAI_RESTART:
	// 		go hms.RestartDeployment(deploymentInputContext)
	// 	default:
	// 		return false, fmt.Errorf("provided service type %s doesnt have a deployment for change indicator `%s`", serviceType, changeIndicator)
	// 	}
	default:
		return false, fmt.Errorf("provided service type doesnt have a workflow")
	}
	return true, nil
}

func convertDeploymentSqlToPb(deployment db.Deployment) (pb.DpaiDeployment, error) {

	serviceType, err := utils.ConvertDpaiServiceTypeToPbEnum(deployment.ServiceType)
	if err != nil {
		return pb.DpaiDeployment{}, err
	}
	changeIndicator, err := utils.ConvertDpaiDeploymentChangeIndicatorToPbEnum(deployment.ChangeIndicator)
	if err != nil {
		return pb.DpaiDeployment{}, err
	}
	state, err := utils.ConvertDpaiDeploymentStateToPbEnum(deployment.StatusState)
	if err != nil {
		return pb.DpaiDeployment{}, err
	}
	return pb.DpaiDeployment{
		Id:              deployment.ID,
		ServiceId:       deployment.ServiceID.String,
		ServiceType:     *serviceType,
		ChangeIndicator: *changeIndicator,
		Status: &pb.DpaiDeploymentStatus{
			State:       *state,
			DisplayName: deployment.StatusDisplayName.String,
			Message:     deployment.StatusMessage.String,
		},
		ErrorMessage: deployment.ErrorMessage.String,
		Metadata: &pb.DpaiDeploymentMeta{
			CreatedAt: timestamppb.New(deployment.CreatedAt.Time),
			UpdatedAt: timestamppb.New(deployment.UpdatedAt.Time),
			CreatedBy: deployment.CreatedBy,
		},
	}, nil
}

// TODO: Pagination
func (s *DpaiServer) DpaiDeploymentList(ctx context.Context, req *pb.DpaiDeploymentListRequest) (*pb.DpaiDeploymentListResponse, error) {
	fmt.Println("List Deployments")

	if req.GetCloudAccountId() == "" {
		log.Println("Cloud Account Id must be provided.")
		return nil, fmt.Errorf("missing cloud account id")
	}

	fmt.Printf("Cloud Id is: %s", req.GetCloudAccountId())

	if req.GetLimit() == 0 {
		req.Limit = 100
	}

	var deployments []db.Deployment
	var err error

	params := db.ListDeploymentsParams{
		CloudAccountID:     pgtype.Text{String: req.GetCloudAccountId(), Valid: true},
		WorkspaceID:        pgtype.Text{String: req.GetWorkspaceId(), Valid: req.GetWorkspaceId() != ""},
		ServiceID:          pgtype.Text{String: req.GetServiceId(), Valid: req.GetServiceId() != ""},
		ServiceType:        pgtype.Text{String: req.GetServiceType().String(), Valid: req.GetServiceType().String() != pb.DpaiServiceType_DPAI_SERVICE_TYPE_UNSPECIFIED.String()},
		ChangeIndicator:    pgtype.Text{String: req.GetChangeIndicator().String(), Valid: req.GetChangeIndicator().String() != pb.DpaiDeploymentChangeIndicator_DPAI_DEPLOYMENT_CHANGE_INDICATOR_UNSPECIFIED.String()},
		ParentDeploymentID: pgtype.Text{String: req.GetParentDeploymentId(), Valid: req.GetParentDeploymentId() != ""},
		Limit:              req.GetLimit(),
		Offset:             req.GetOffset(),
	}

	deployments, err = s.SqlModel.ListDeployments(context.Background(), params)

	if err != nil {
		return nil, fmt.Errorf("failed to list deployments. Error: %+v", err)
	}

	var results []*pb.DpaiDeployment
	for _, deployment := range deployments {
		result, err := convertDeploymentSqlToPb(deployment)
		if err != nil {
			return nil, fmt.Errorf("failed to convert deployment to pb. Error: %+v", err)
		}
		results = append(results, &result)
	}

	var previousOffset, nextOffset, lastOffset, totalCount int64

	// Count the total number of records
	var countParams db.CountListDeploymentsParams
	err = copier.Copy(&countParams, params)
	if err != nil {
		return nil, err
	}

	totalCount, err = s.Sql.CountListDeployments(context.Background(), countParams)
	if err != nil {
		log.Printf("Failed to count Deployments with error : %+v", err)
		return nil, err
	}

	// Calculate the previous and next offsets
	if req.GetOffset() > 0 {
		previousOffset = req.GetOffset() - req.GetLimit()
		if previousOffset < 0 {
			previousOffset = 0
		}
	}
	nextOffset = req.GetOffset() + req.GetLimit()
	if nextOffset >= totalCount {
		nextOffset = -1
	}
	lastOffset = ((totalCount - 1) / req.GetLimit()) * req.GetLimit()

	return &pb.DpaiDeploymentListResponse{
		Data:         results,
		Limit:        req.GetLimit(),
		CurrOffset:   req.GetOffset(),
		LastOffset:   lastOffset,
		NextOffset:   nextOffset,
		PrevOffset:   previousOffset,
		TotalRecords: totalCount,
	}, nil
}

func (s *DpaiServer) DpaiDeploymentCreate(ctx context.Context, req *pb.DpaiDeploymentCreateRequest) (*pb.DpaiDeploymentCreateResponse, error) {

	fmt.Printf("Create Deployment...\n")
	if req.GetDeploymentId() == "" {
		req.DeploymentId = uuid.New().String()
	}

	create, err := s.SqlModel.CreateDeployment(ctx, db.CreateDeploymentParams{
		CloudAccountID:  pgtype.Text{String: req.GetCloudAccountId(), Valid: true},
		ID:              req.DeploymentId,
		WorkspaceID:     pgtype.Text{String: req.GetWorkspaceId(), Valid: true},
		ServiceID:       pgtype.Text{String: req.GetServiceId(), Valid: true},
		ServiceType:     req.GetServiceType().String(),
		ChangeIndicator: req.GetChangeIndicator().String(),
		CreatedBy:       req.GetCreatedBy(),
		InputPayload:    req.GetInput(),
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	_, err = TriggerDeployment(create.ID, req.GetServiceType(), req.GetChangeIndicator(), *s)
	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}
	return &pb.DpaiDeploymentCreateResponse{
		DeploymentId: create.ID,
	}, nil
}

func (s *DpaiServer) DpaiDeploymentGet(ctx context.Context, req *pb.DpaiDeploymentGetRequest) (*pb.DpaiDeployment, error) {

	fmt.Printf("Get Deployment By Id: %s", req.GetId())
	deployment, err := s.SqlModel.GetDeployment(context.Background(), db.GetDeploymentParams{
		ID:             req.GetId(),
		CloudAccountID: pgtype.Text{String: req.GetCloudAccountId(), Valid: true},
	})
	if err != nil {
		log.Printf("No deployment found for id : %s Error: %+v", req.GetId(), err)
		return nil, err
	}

	result, err := convertDeploymentSqlToPb(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to convert deployment to pb. Error: %+v", err)
	}

	return &result, nil
}

// func (s *DpaiServer) DpaiDeploymentUpdate(ctx context.Context, req *pb.DpaiDeploymentUpdateRequest) (*pb.DpaiDeployment, error) {
// 	id := req.GetId()
// 	fmt.Printf("Update Deployment By Id: %s", id)
// 	temp, err := s.SqlModel.GetUpdateDeployment(context.Background(), id)
// 	if err != nil {
// 		log.Printf("No deployment found for id : %s Error: %+v", id, err)
// 		return nil, err
// 	}

// 	errorMessage := req.GetErrorMessage()
// 	if errorMessage == "" {
// 		errorMessage = temp.ErrorMessage.String
// 	}
// 	status := req.GetStatus().String()
// 	if status == "" {
// 		status = temp.StatusState
// 	}

// 	updateUserParams := db.UpdateDeploymentParams{
// 		ID:           id,
// 		StatusState:  status,
// 		ErrorMessage: pgtype.Text{String: errorMessage, Valid: true},
// 	}

// 	deployment, err := s.SqlModel.UpdateDeployment(context.Background(), updateUserParams)
// 	if err != nil {
// 		log.Printf("Failed to update workspace_size with error : %+v", err)
// 		return nil, err
// 	}

// 	result, err := convertDeploymentSqlToPb(deployment)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to convert deployment to pb. Error: %+v", err)
// 	}

// 	return &result, nil
// }

func (s *DpaiServer) DpaiDeploymentDelete(ctx context.Context, req *pb.DpaiDeploymentDeleteRequest) (*pb.DpaiDeploymentDeleteResponse, error) {
	id := req.GetId()

	status := true
	fmt.Printf("Delete Deployment by Id: %s", id)
	err := s.SqlModel.DeleteDeployment(context.Background(), id)
	if err != nil {
		log.Printf("Failed to delete deployment with error : %+v", err)
		status = false
	}

	return &pb.DpaiDeploymentDeleteResponse{Success: status}, nil
}

func (s *DpaiServer) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	// log := log.FromContext(ctx).WithName("DpaiDeployment.Ping")
	log.Print("Ping")
	return &emptypb.Empty{}, nil
}
