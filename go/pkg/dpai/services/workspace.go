// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	// "github.com/google/uuid"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jinzhu/copier"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func workspaceSqlToProto(data db.Workspace) (pb.DpaiWorkspace, error) {
	state, err := utils.ConvertDpaiDeploymentStateToPbEnum(data.DeploymentStatusState)
	if err != nil {
		return pb.DpaiWorkspace{}, err
	}
	tags, err := utils.ConvertBytesToTags(data.Tags)
	if err != nil {
		return pb.DpaiWorkspace{}, err
	}

	return pb.DpaiWorkspace{
		Id:             data.ID,
		CloudAccountId: data.CloudAccountID,
		Name:           data.Name,
		Region:         data.Region.String,
		Description:    data.Description.String,
		Tags:           tags,
		DeploymentMetadata: &pb.DpaiWorkspaceDeploymentMeta{
			DeploymentId: data.DeploymentID,
			DeploymentStatus: &pb.DpaiDeploymentStatus{
				State:       *state,
				DisplayName: data.DeploymentStatusDisplayName.String,
				Message:     data.DeploymentStatusMessage.String,
			},
			IksId:                 data.IksID.String,
			ManagementNodeGroupId: data.ManagementNodegroupID.String,
		},

		Metadata: &pb.DpaiMeta{
			CreatedAt: timestamppb.New(data.CreatedAt.Time),
			CreatedBy: data.CreatedBy,
			UpdatedAt: timestamppb.New(data.UpdatedAt.Time),
			UpdatedBy: data.UpdatedBy.String,
		},
	}, nil
}

// TODO: Pagination
func (s *DpaiServer) DpaiWorkspaceList(ctx context.Context, req *pb.DpaiWorkspaceListRequest) (*pb.DpaiWorkspaceListResponse, error) {
	log.Println("List Workspaces")

	if req.GetCloudAccountId() == "" {
		log.Println("Cloud Account Id must be provided.")
		return nil, fmt.Errorf("missing cloud account id")
	}

	if req.GetLimit() == 0 {
		req.Limit = 10
	}

	params := db.ListWorkspacesParams{
		CloudAccountID:    req.GetCloudAccountId(),
		Name:              pgtype.Text{String: req.GetName(), Valid: req.GetName() != ""},
		StatusDisplayName: pgtype.Text{String: req.GetStatus(), Valid: req.GetStatus() != ""},
	}

	queryResults, err := s.SqlModel.ListWorkspaces(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("Failed to list Workspaces with error : %+v", err)
	}

	var results []*pb.DpaiWorkspace
	for _, record := range queryResults {
		var workspace db.Workspace
		err = copier.Copy(&workspace, record)
		if err != nil {
			return nil, err
		}

		result, err := workspaceSqlToProto(workspace)
		if err != nil {
			return nil, err
		}
		results = append(results, &result)
	}
	var previousOffset, nextOffset, lastOffset, totalCount int64

	// Count the total number of records
	var countParams db.CountListWorkspacesParams
	err = copier.Copy(&countParams, params)
	if err != nil {
		return nil, err
	}

	totalCount, err = s.Sql.CountListWorkspaces(context.Background(), countParams)
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

	return &pb.DpaiWorkspaceListResponse{
		Data:         results,
		Limit:        req.GetLimit(),
		CurrOffset:   req.GetOffset(),
		LastOffset:   lastOffset,
		NextOffset:   nextOffset,
		PrevOffset:   previousOffset,
		TotalRecords: totalCount,
	}, nil
}

func (s *DpaiServer) DpaiWorkspaceCreate(ctx context.Context, req *pb.DpaiWorkspaceCreateRequest) (*pb.DpaiWorkspace, error) {

	fmt.Printf("Create Workspace : %s \n", req.GetName())

	// convert input payload into the bytes
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("create workflow failed with the error %s", err)
	}

	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		CloudAccountId:  req.GetCloudAccountId(),
		ServiceType:     pb.DpaiServiceType_DPAI_WORKSPACE,
		ChangeIndicator: pb.DpaiDeploymentChangeIndicator_DPAI_CREATE,
		CreatedBy:       req.GetCreatedBy(),
		Input:           jsonData,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create the deployment. Error message: %+v", err)
	}

	tags, err := utils.ConvertTagsToBytes(req.GetTags())
	if err != nil {
		return nil, err
	}
	workspace, err := s.SqlModel.CreateWorkspace(context.Background(), db.CreateWorkspaceParams{
		ID:                    uuid.New().String(),
		CloudAccountID:        req.GetCloudAccountId(),
		Name:                  req.GetName(),
		Description:           pgtype.Text{String: req.GetDescription(), Valid: true},
		Tags:                  tags,
		DeploymentID:          deployment.DeploymentId,
		DeploymentStatusState: pb.DpaiDeploymentState_DPAI_PENDING.String(),
		CreatedBy:             req.GetCreatedBy(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create the workspace. Error message: %+v", err)
	}
	result, err := workspaceSqlToProto(workspace)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *DpaiServer) DpaiWorkspaceGet(ctx context.Context, req *pb.DpaiWorkspaceGetRequest) (*pb.DpaiWorkspace, error) {
	id := req.GetId()

	params := db.GetWorkspaceParams{
		CloudAccountID: pgtype.Text{String: req.GetCloudAccountId(), Valid: true},
		WorkspaceID:    req.GetId(),
	}

	fmt.Printf("Get Workspace By Id: %s", id)
	queryResult, err := s.SqlModel.GetWorkspace(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("Failed to get Workspace with error : %+v", err)
	}

	var workspace db.Workspace
	err = copier.Copy(&workspace, queryResult)
	if err != nil {
		return nil, err
	}
	result, err := workspaceSqlToProto(workspace)
	if err != nil {
		return nil, err
	}

	listServices, err := s.Sql.ListWorkspaceServices(context.Background(), db.ListWorkspaceServicesParams{
		CloudAccountID: pgtype.Text{String: req.GetCloudAccountId(), Valid: true},
		WorkspaceID:    pgtype.Text{String: req.GetId(), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Workspace Services with error : %+v", err)
	}

	var services []*pb.DpaiWorkspaceServices
	for _, result := range listServices {
		state, err := utils.ConvertDpaiDeploymentStateToPbEnum(result.StatusState)
		if err != nil {
			return nil, err
		}
		service := pb.DpaiWorkspaceServices{
			Id:        result.ID,
			Type:      result.ServiceType,
			Name:      result.Name,
			VersionId: result.Version,
			DeploymentStatus: &pb.DpaiDeploymentStatus{
				State:       *state,
				DisplayName: result.StatusDisplayName.String,
				Message:     result.StatusMessage.String,
			},
			UpdatedAt: timestamppb.New(result.ServiceUpdatedAt.Time),
		}
		services = append(services, &service)
	}
	result.Services = services
	return &result, nil
}

func (s *DpaiServer) DpaiWorkspaceUpdate(ctx context.Context, req *pb.DpaiWorkspaceUpdateRequest) (*pb.DpaiWorkspace, error) {
	id := req.GetId()
	fmt.Printf("Update Workspace By Id: %s", id)
	temp, err := s.SqlModel.GetUpdateWorkspace(context.Background(), id)
	if err != nil {
		log.Printf("No record found for id : %s Error: %+v", id, err)
		return nil, err
	}

	description := req.GetDescription()
	if description == "" {
		description = temp.Description.String
	}

	var tags []byte
	if len(req.GetTags()) == 0 {
		tags = temp.Tags
	} else {
		tags, err = utils.ConvertTagsToBytes(req.GetTags())
		if err != nil {
			return nil, err
		}
	}

	updateUserParams := db.UpdateWorkspaceParams{
		ID:          id,
		Tags:        tags,
		Description: pgtype.Text{String: description, Valid: true},
	}

	queryResult, err := s.SqlModel.UpdateWorkspace(context.Background(), updateUserParams)
	if err != nil {
		log.Printf("Failed to update workspace with error : %+v", err)
		return nil, err
	}

	var resultTags map[string]string
	err = json.Unmarshal(queryResult.Tags, &resultTags)
	if err != nil {
		log.Printf("Failed to unmarshal Workspace Update tags with error : %+v", err)
		return nil, err
	}

	result, err := workspaceSqlToProto(queryResult)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *DpaiServer) DpaiWorkspaceDelete(ctx context.Context, req *pb.DpaiWorkspaceDeleteRequest) (*pb.DpaiDeploymentResponse, error) {
	id := req.GetId()

	fmt.Printf("Delete Workspace by Id: %s", id)
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		ServiceId:       req.GetId(),
		ServiceType:     pb.DpaiServiceType_DPAI_WORKSPACE,
		ChangeIndicator: pb.DpaiDeploymentChangeIndicator_DPAI_DELETE,
		Input:           jsonData,
	})

	if err != nil {
		return &pb.DpaiDeploymentResponse{
			Status: &pb.DpaiDeploymentStatus{
				State:       pb.DpaiDeploymentState_DPAI_FAILED,
				DisplayName: "Failed",
				Message:     "Failed to delete",
			},
		}, fmt.Errorf("failed to create the deployment. Error message: %+v", err)
	}

	return &pb.DpaiDeploymentResponse{
		DeploymentId: deployment.GetDeploymentId(),
		Status: &pb.DpaiDeploymentStatus{
			State:       pb.DpaiDeploymentState_DPAI_ACCEPTED,
			DisplayName: "Accepted",
			Message:     "Accepted the Delete request",
		},
	}, nil

}
