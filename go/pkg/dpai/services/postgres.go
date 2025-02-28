// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func postgresSqlToProto(postgres db.Postgres) (pb.DpaiPostgres, error) {
	state, err := utils.ConvertDpaiDeploymentStateToPbEnum(postgres.DeploymentStatusState)
	if err != nil {
		return pb.DpaiPostgres{}, err
	}
	advanceConf, err := utils.ConvertBytesToTags(postgres.AdvanceConfiguration)
	if err != nil {
		return pb.DpaiPostgres{}, err
	}
	tags, err := utils.ConvertBytesToTags(postgres.Tags)
	if err != nil {
		return pb.DpaiPostgres{}, err
	}
	secret, err := utils.ConvertBytesToSecretReference(postgres.AdminPasswordSecretReference)
	if err != nil {
		return pb.DpaiPostgres{}, err
	}
	return pb.DpaiPostgres{
		Id:          postgres.ID,
		WorkspaceId: postgres.WorkspaceID,
		Name:        postgres.Name,
		Description: postgres.Description.String,
		VersionId:   postgres.VersionID,
		SizeProperties: &pb.DpaiPostgresSizeProperties{
			SizeId:                  postgres.SizeID,
			NumberOfInstances:       postgres.NumberOfInstances.Int32,
			NumberOfPgPoolInstances: postgres.NumberOfPgpoolInstances.Int32,
			DiskSizeInGb:            postgres.DiskSizeInGb.Int32,
		},
		OptionalProperties: &pb.DpaiPostgresOptionalProperties{
			InitialDatabaseName: postgres.InitialDatabaseName.String,
		},
		AdminProperties: &pb.DpaiPostgresAdminProperties{
			AdminUsername:                postgres.AdminUsername,
			AdminPasswordSecretReference: secret,
		},
		AdvanceConfiguration: advanceConf,
		Tags:                 tags,
		ServerUrl:            postgres.ServerUrl.String,
		DeploymentMetadata: &pb.DpaiPostgresDeploymentMeta{
			DeploymentId: postgres.DeploymentID,
			DeploymentStatus: &pb.DpaiDeploymentStatus{
				State:       *state,
				DisplayName: "",
				Message:     "",
			},
		},
		Metadata: &pb.DpaiMeta{
			CreatedAt: timestamppb.New(postgres.CreatedAt.Time),
			CreatedBy: postgres.CreatedBy,
			UpdatedAt: timestamppb.New(postgres.UpdatedAt.Time),
			UpdatedBy: postgres.UpdatedBy.String,
		},
	}, nil
}

// TODO: Pagination
func (s *DpaiServer) DpaiPostgresList(ctx context.Context, req *pb.DpaiPostgresListRequest) (*pb.DpaiPostgresListResponse, error) {
	log.Println("List Postgres")
	params := db.ListPostgresParams{
		CloudAccountID: req.GetCloudAccountId(),
		WorkspaceID:    pgtype.Text{String: req.GetWorkspaceId(), Valid: req.GetWorkspaceId() != ""},
	}

	queryResults, err := s.Sql.ListPostgres(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("Failed to list Postgres with error : %+v", err)
	}

	var results []*pb.DpaiPostgres
	for _, queryResult := range queryResults {

		result, err := postgresSqlToProto(queryResult)
		if err != nil {
			return nil, err
		}
		results = append(results, &result)
	}
	return &pb.DpaiPostgresListResponse{
		Data:      results,
		HasMore:   false,
		PrevToken: "",
		NextToken: "",
	}, nil
}

func (s *DpaiServer) DpaiPostgresCreate(ctx context.Context, req *pb.DpaiPostgresCreateRequest) (*pb.DpaiDeploymentResponse, error) {

	log.Printf("Create Postgres : %s \n", req.GetName())
	uuidStr := uuid.New().String()
	deploymentId := fmt.Sprintf("postgres-%s-%s%s", req.GetName(), uuidStr[:8], uuidStr[len(uuidStr)-12:])

	// generate secret
	k8sClusterID, err := k8s.GetIksClusterID(s.SqlModel, req.WorkspaceId, "")
	if err != nil {
		return nil, err
	}
	k8sClient := k8s.K8sClient{
		ClusterID: k8sClusterID,
	}
	k8sClient.GetK8sClientSet()
	secretData := map[string][]byte{
		"password":        []byte(req.AdminProperties.GetAdminPassword()),
		"repmgr-password": []byte(req.AdminProperties.GetAdminPassword()),
		"admin-password":  []byte(req.AdminProperties.GetAdminPassword()),
	}
	err = k8sClient.CreateSecret("secrets", deploymentId, secretData)
	if err != nil {
		return nil, fmt.Errorf("failed to create the secret")
	}

	// mask user provided secret
	req.AdminProperties.AdminPassword = ""
	req.AdminProperties.AdminPasswordSecretReference = &pb.DpaiSecretReference{
		SecretName:    deploymentId,
		SecretKeyName: "password",
	}

	if req.OptionalProperties == nil {
		req.OptionalProperties = &pb.DpaiPostgresOptionalProperties{}
	}
	if req.GetOptionalProperties().GetInitialDatabaseName() == "" {
		req.GetOptionalProperties().InitialDatabaseName = "demo"
	}
	// convert input payload into the bytes
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		DeploymentId:    deploymentId,
		CloudAccountId:  k8sClusterID.CloudAccountId,
		WorkspaceId:     req.GetWorkspaceId(),
		ServiceType:     pb.DpaiServiceType_DPAI_POSTGRES,
		ChangeIndicator: pb.DpaiDeploymentChangeIndicator_DPAI_CREATE,
		CreatedBy:       req.GetCreatedBy(),
		Input:           jsonData,
	})

	if err != nil {
		return &pb.DpaiDeploymentResponse{
			Status: &pb.DpaiDeploymentStatus{
				State:       pb.DpaiDeploymentState_DPAI_FAILED,
				DisplayName: "Failed",
				Message:     "Failed to create the deployment",
			},
		}, fmt.Errorf("failed to create the deployment. Error message: %+v", err)
	}

	return &pb.DpaiDeploymentResponse{
		DeploymentId: deployment.GetDeploymentId(),
		Status: &pb.DpaiDeploymentStatus{
			State:       pb.DpaiDeploymentState_DPAI_ACCEPTED,
			DisplayName: "Accepted",
			Message:     "",
		},
	}, nil
}

func (s *DpaiServer) DpaiPostgresGetById(ctx context.Context, req *pb.DpaiPostgresGetByIdRequest) (*pb.DpaiPostgres, error) {
	id := req.GetId()
	log.Printf("Get Postgres By Id: %s", id)
	queryResult, err := s.Sql.GetPostgresById(context.Background(), id)
	if err != nil {
		log.Printf("Failed to get Postgres with error : %+v", err)
		return nil, err
	}

	result, err := postgresSqlToProto(queryResult)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *DpaiServer) DpaiPostgresGetByName(ctx context.Context, req *pb.DpaiPostgresGetByNameRequest) (*pb.DpaiPostgres, error) {
	log.Printf("Get Postgres By Name: %s", req.GetName())
	queryResult, err := s.Sql.GetPostgresByName(context.Background(), db.GetPostgresByNameParams{
		WorkspaceID: req.GetWorkspaceId(),
		Name:        req.GetName(),
	})
	if err != nil {
		log.Printf("Failed to get Postgres with error : %+v", err)
		return nil, err
	}

	result, err := postgresSqlToProto(queryResult)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *DpaiServer) DpaiPostgresUpdate(ctx context.Context, req *pb.DpaiPostgresUpdateRequest) (*pb.DpaiPostgres, error) {
	id := req.GetId()
	log.Printf("Update Postgres By Id: %s", id)
	temp, err := s.Sql.GetUpdatePostgres(context.Background(), id)
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

	updateParams := db.UpdatePostgresParams{
		ID:          id,
		Tags:        tags,
		Description: pgtype.Text{String: description, Valid: true},
	}

	queryResult, err := s.Sql.UpdatePostgres(context.Background(), updateParams)
	if err != nil {
		log.Printf("Failed to update Postgres with error : %+v", err)
		return nil, err
	}

	var resultTags map[string]string
	json.Unmarshal(queryResult.Tags, &resultTags)

	result, err := postgresSqlToProto(queryResult)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *DpaiServer) DpaiPostgresDelete(ctx context.Context, req *pb.DpaiPostgresDeleteRequest) (*pb.DpaiDeploymentResponse, error) {
	id := req.GetId()

	log.Printf("Delete Postgres by Id: %s", id)
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("Not able to parse the input request to Json data. Error: %+v", err)
		return nil, err
	}

	pg, err := s.Sql.GetPostgresById(context.Background(), id)
	if err != nil {
		return nil, err
	}

	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		WorkspaceId:     pg.WorkspaceID,
		ServiceId:       req.GetId(),
		ServiceType:     pb.DpaiServiceType_DPAI_POSTGRES,
		ChangeIndicator: pb.DpaiDeploymentChangeIndicator_DPAI_DELETE,
		Input:           jsonData,
	})

	if err != nil {
		return &pb.DpaiDeploymentResponse{
			Status: &pb.DpaiDeploymentStatus{
				State:       pb.DpaiDeploymentState_DPAI_FAILED,
				DisplayName: "Failed",
				Message:     "Failed to create the deployment",
			},
		}, fmt.Errorf("failed to create the deployment. Error message: %+v", err)
	}

	return &pb.DpaiDeploymentResponse{
		DeploymentId: deployment.GetDeploymentId(),
		Status: &pb.DpaiDeploymentStatus{
			State:       pb.DpaiDeploymentState_DPAI_ACCEPTED,
			DisplayName: "Accepted",
		},
	}, nil

}

func (s *DpaiServer) DpaiPostgresRestart(ctx context.Context, req *pb.DpaiPostgresRestartRequest) (*pb.DpaiDeploymentResponse, error) {
	id := req.GetId()

	log.Printf("Restart Postgres by Id: %s", id)
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("Not able to parse the input request to Json data. Error: %+v", err)
		return nil, err
	}
	pg, err := s.Sql.GetPostgresById(context.Background(), id)
	if err != nil {
		return nil, err
	}
	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		WorkspaceId:     pg.WorkspaceID,
		ServiceId:       req.GetId(),
		ServiceType:     pb.DpaiServiceType_DPAI_POSTGRES,
		ChangeIndicator: pb.DpaiDeploymentChangeIndicator_DPAI_RESTART,
		Input:           jsonData,
	})

	if err != nil {
		return &pb.DpaiDeploymentResponse{
			Status: &pb.DpaiDeploymentStatus{
				State:       pb.DpaiDeploymentState_DPAI_FAILED,
				DisplayName: "Failed",
				Message:     "Failed to create the deployment",
			},
		}, fmt.Errorf("failed to create the deployment. Error message: %+v", err)
	}

	return &pb.DpaiDeploymentResponse{
		DeploymentId: deployment.GetDeploymentId(),
		Status: &pb.DpaiDeploymentStatus{
			State:       pb.DpaiDeploymentState_DPAI_ACCEPTED,
			DisplayName: "Accepted",
		},
	}, nil

}

func (s *DpaiServer) DpaiPostgresResize(ctx context.Context, req *pb.DpaiPostgresResizeRequest) (*pb.DpaiDeploymentResponse, error) {
	id := req.GetId()

	log.Printf("Resize Postgres by Id: %s", id)
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("Not able to parse the input request to Json data. Error: %+v", err)
		return nil, err
	}

	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		ServiceId:       req.GetId(),
		ServiceType:     pb.DpaiServiceType_DPAI_POSTGRES,
		ChangeIndicator: pb.DpaiDeploymentChangeIndicator_DPAI_RESIZE,
		Input:           jsonData,
	})

	if err != nil {
		return &pb.DpaiDeploymentResponse{
			Status: &pb.DpaiDeploymentStatus{
				State:       pb.DpaiDeploymentState_DPAI_FAILED,
				DisplayName: "Failed",
				Message:     "Failed to create the deployment",
			},
		}, fmt.Errorf("failed to create the deployment. Error message: %+v", err)
	}

	return &pb.DpaiDeploymentResponse{
		DeploymentId: deployment.GetDeploymentId(),
		Status: &pb.DpaiDeploymentStatus{
			State:       pb.DpaiDeploymentState_DPAI_ACCEPTED,
			DisplayName: "Accepted",
		},
	}, nil

}

func (s *DpaiServer) DpaiPostgresListUpgrade(ctx context.Context, req *pb.DpaiPostgresListUpgradeRequest) (*pb.DpaiPostgresVersionListResponse, error) {
	log.Printf("List Postgres versions available for postgres instance: %s", req.GetId())

	postgresData, err := s.Sql.GetPostgresById(context.Background(), req.GetId())
	if err != nil {
		log.Printf("No postgres instance found with the id: %s : %+v", req.GetId(), err)
		return nil, err
	}

	queryResults, err := s.Sql.ListPostgresVersionUpgrades(context.Background(), pgtype.Text{String: postgresData.VersionID, Valid: true})
	if err != nil {
		log.Printf("Failed to list Postgres version upgrades with error : %+v", err)
		return nil, err
	}

	var results []*pb.DpaiPostgresVersion
	for _, queryResult := range queryResults {

		result, err := postgresVersionSqlToProto(queryResult)
		if err != nil {
			return nil, err
		}
		results = append(results, &result)
	}
	return &pb.DpaiPostgresVersionListResponse{
		Data: results}, nil
}

func (s *DpaiServer) DpaiPostgresUpgrade(ctx context.Context, req *pb.DpaiPostgresUpgradeRequest) (*pb.DpaiDeploymentResponse, error) {

	if req.GetVersionId() == "" || req.GetId() == "" {
		return nil, fmt.Errorf("missing input: Id and VersionId are mandatory for the upgrade operation")
	}

	log.Printf("Upgrade the Postgres with Id: %s to version: %s", req.GetId(), req.GetVersionId())
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("Not able to parse the input request to Json data. Error: %+v", err)
		return nil, err
	}

	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		ServiceId:       req.GetId(),
		ServiceType:     pb.DpaiServiceType_DPAI_POSTGRES,
		ChangeIndicator: pb.DpaiDeploymentChangeIndicator_DPAI_UPGRADE,
		Input:           jsonData,
	})

	if err != nil {
		return &pb.DpaiDeploymentResponse{
			Status: &pb.DpaiDeploymentStatus{
				State:       pb.DpaiDeploymentState_DPAI_FAILED,
				DisplayName: "Failed",
				Message:     "Failed to create the deployment",
			},
		}, fmt.Errorf("failed to create the deployment. Error message: %+v", err)
	}

	return &pb.DpaiDeploymentResponse{
		DeploymentId: deployment.GetDeploymentId(),
		Status: &pb.DpaiDeploymentStatus{
			State:       pb.DpaiDeploymentState_DPAI_ACCEPTED,
			DisplayName: "Accepted",
		},
	}, nil

}
