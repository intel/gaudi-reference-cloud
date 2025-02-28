// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jinzhu/copier"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/crypto"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func airflowProtoToSql(data db.Airflow) (pb.DpaiAirflow, error) {
	state, err := utils.ConvertDpaiDeploymentStateToPbEnum(data.DeploymentStatusState)
	if err != nil {
		return pb.DpaiAirflow{}, err
	}

	tags, err := utils.ConvertBytesToTags(data.Tags)
	if err != nil {
		return pb.DpaiAirflow{}, err
	}
	return pb.DpaiAirflow{
		CloudAccountId: data.CloudAccountID,
		Id:             data.ID,
		WorkspaceName:  data.WorkspaceName,
		Name:           data.Name,
		Description:    data.Description.String,
		Version:        data.Version,
		Tags:           tags,
		StorageProperties: &pb.DpaiAirflowStorageProperties{
			BucketId:        data.BucketID.String,
			BucketPrincipal: data.BucketPrincipal.String,
			Path: &pb.DpaiAirflowPathProperties{
				DagFolderPath:    data.DagFolderPath.String,
				PluginFolderPath: data.PluginFolderPath.String,
				RequirementPath:  data.RequirementPath.String,
				LogFolder:        data.LogFolder.String,
			},
		},
		WebServerProperties: &pb.DpaiAirflowWebServerProperties{
			Endpoint:               data.Endpoint.String,
			WebserverAdminUsername: data.WebserverAdminUsername.String,
			WebserverAdminPassword: fmt.Sprintf("%d", data.WebserverAdminPasswordSecretID.Int32),
		},
		SizeProperties: &pb.DpaiAirflowSizeProperties{
			Size:               data.Size,
			NumberOfNodes:      data.NumberOfNodes.Int32,
			NumberOfWorkers:    data.NumberOfWorkers.Int32,
			NumberOfSchedulers: data.NumberOfSchedulers.Int32,
		},
		DeploymentMetadata: &pb.DpaiAirflowDeploymentMeta{
			DeploymentId:      data.DeploymentID,
			BackendDatabaseId: data.BackendDatabaseID,
			IksClusterId:      data.IksClusterID,
			WorkspaceId:       data.WorkspaceID,
			NodeGroupId:       data.NodeGroupID,
			DeploymentStatus: &pb.DpaiDeploymentStatus{
				State:       *state,
				DisplayName: data.DeploymentStatusDisplayName.String,
				Message:     data.DeploymentStatusMessage.String,
			},
		},
		Metadata: &pb.DpaiMeta{
			CreatedAt: timestamppb.New(data.CreatedAt.Time),
			CreatedBy: data.CreatedBy,
			UpdatedAt: timestamppb.New(data.UpdatedAt.Time),
			UpdatedBy: data.UpdatedBy.String,
		},
	}, nil
}

func getAirflowWithDeploymentStatus(s *DpaiServer, airflow db.Airflow) (*pb.DpaiAirflow, error) {

	deployment, err := s.Sql.GetDeployment(context.Background(), db.GetDeploymentParams{
		ID:             airflow.DeploymentID,
		CloudAccountID: pgtype.Text{String: airflow.CloudAccountID, Valid: true},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get deployment with error : %+v", err)
	}

	airflow.DeploymentStatusState = deployment.StatusState
	airflow.DeploymentStatusDisplayName = deployment.StatusDisplayName
	airflow.DeploymentStatusMessage = deployment.StatusMessage

	result, err := airflowProtoToSql(airflow)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *DpaiServer) DpaiAirflowList(ctx context.Context, req *pb.DpaiAirflowListRequest) (*pb.DpaiAirflowListResponse, error) {
	log.Println("List Airflow")

	if req.GetLimit() == 0 {
		req.Limit = 20
	}

	params := db.ListAirflowParams{
		CloudAccountID: req.GetCloudAccountId(),
		Name:           pgtype.Text{String: req.GetName(), Valid: req.GetName() != ""},
		WorkspaceID:    pgtype.Text{String: req.GetWorkspaceId(), Valid: req.GetWorkspaceId() != ""},
		Limit:          int64(req.GetLimit()),
		Offset:         int64(req.GetOffset()),
	}

	records, err := s.Sql.ListAirflow(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("failed to list Airflow with error : %+v", err)
	}

	var results []*pb.DpaiAirflow
	for _, record := range records {

		result, err := getAirflowWithDeploymentStatus(s, record)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	var previousOffset, nextOffset, lastOffset, totalCount int64

	// Count the total number of records
	var countParams db.CountListAirflowParams
	err = copier.Copy(&countParams, params)
	if err != nil {
		return nil, err
	}

	totalCount, err = s.Sql.CountListAirflow(context.Background(), countParams)
	if err != nil {
		log.Printf("Failed to count SchemaRequests with error : %+v", err)
		return nil, err
	}

	// test debug to ensure the iks az clientset is set
	if s.IKSAzClientSet != nil {
		log.Printf("IKSAzClientSet is set %+v", s.IKSAzClientSet)
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

	return &pb.DpaiAirflowListResponse{
		Data:         results,
		Limit:        req.GetLimit(),
		CurrOffset:   req.GetOffset(),
		LastOffset:   lastOffset,
		NextOffset:   nextOffset,
		PrevOffset:   previousOffset,
		TotalRecords: totalCount,
	}, nil
}

func (s *DpaiServer) DpaiAirflowCreate(ctx context.Context, req *pb.DpaiAirflowCreateRequest) (*pb.DpaiAirflow, error) {

	log.Printf("Creating Airflow : %s \n", req.GetName())

	deploymentId, err := utils.GenerateUniqueServiceId(s.Sql, pb.DpaiServiceType_DPAI_AIRFLOW, true)
	if err != nil {
		return nil, err
	}
	airflowId := deploymentId[4:]
	log.Printf("Airflow : %s \n", airflowId)

	// proto validation
	err = req.ValidateAll()
	if err != nil {
		return nil, fmt.Errorf("input validation error: %s", err)
	}

	count, err := s.Sql.CheckUniqueAirflow(context.Background(), db.CheckUniqueAirflowParams{
		CloudAccountID: req.GetCloudAccountId(),
		Name:           req.GetName(),
		WorkspaceName:  req.GetWorkspaceName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check unique airflow with error : %+v", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("input error. Please provide a unique combination of CloudAccountId, name, workspaceName")
	}

	log.Printf("Validation Completed")
	// storage
	// bucketId := req.StorageProperties.BucketId
	// st := storage.Storage{}
	// log.Printf("Storage Config: %+v, CloudAccountID: %s ", s.Config, req.CloudAccountId)
	// err = st.GetStorageClient(s.Config, req.CloudAccountId)
	// if err != nil {
	// 	return nil, fmt.Errorf("error in GetStorageClient : %s", err)
	// }

	// isBucketExists, err := st.IsValidAirflowBucket(bucketId)
	// if err != nil {
	// 	return nil, err
	// }
	// log.Printf("isBucketExists : %v", isBucketExists)
	// if !isBucketExists {
	// 	return nil, fmt.Errorf("no bucket found for the given id %s", bucketId)
	// }

	// Version
	_, err = s.Sql.GetAirflowVersionByName(context.Background(), req.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the version: %s. Error message: %s", req.Version, err)
	}

	// Size
	size, err := s.Sql.GetAirflowSizeByName(context.Background(), req.GetSizeProperties().Size)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the Size: %s. Error message: %s", req.GetSizeProperties().Size, err)
	}

	if req.SizeProperties.GetNumberOfNodes() == 0 {
		req.SizeProperties.NumberOfNodes = size.NumberOfNodesDefault
	}

	// generate secret ToDO
	// k8sClient := s.IksClient.   //GetIksClient().K8sClient{
	// 	ClusterID: k8sClusterID,
	// }
	// k8sClient.GetIksClient(s.Config)

	// err = k8sClient.CreateSecret("secrets", deploymentId, secretData)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create the secret")
	// }
	// Secret

	secret, err := crypto.EncryptPassword(s.Sql, *s.Config, req.GetWebServerProperties().GetWebserverAdminPassword())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt the password. Error message: %+v", err)
	}

	req.WebServerProperties.WebserverAdminPassword = fmt.Sprintf("%d", secret)
	// convert input payload into the bytes
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	//TODO
	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		DeploymentId:    deploymentId,
		CloudAccountId:  req.GetCloudAccountId(),
		ServiceType:     pb.DpaiServiceType_DPAI_AIRFLOW,
		ServiceId:       airflowId,
		ChangeIndicator: pb.DpaiDeploymentChangeIndicator_DPAI_CREATE,
		CreatedBy:       "CallerOfTheAPI",
		Input:           jsonData,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create the deployment. Error message: %+v", err)
	}

	tags, err := utils.ConvertTagsToBytes(req.GetTags())
	if err != nil {
		return nil, err
	}
	airflow, err := s.SqlModel.CreateAirflow(context.Background(), db.CreateAirflowParams{
		ID:                             airflowId,
		CloudAccountID:                 req.GetCloudAccountId(),
		Name:                           req.GetName(),
		Description:                    pgtype.Text{String: req.GetDescription(), Valid: true},
		Tags:                           tags,
		Version:                        req.GetVersion(),
		Size:                           req.SizeProperties.GetSize(),
		NumberOfNodes:                  pgtype.Int4{Int32: req.SizeProperties.GetNumberOfNodes(), Valid: true},
		NumberOfSchedulers:             pgtype.Int4{Int32: req.SizeProperties.GetNumberOfSchedulers(), Valid: true},
		NumberOfWorkers:                pgtype.Int4{Int32: req.SizeProperties.GetNumberOfWorkers(), Valid: true},
		WorkspaceName:                  req.GetWorkspaceName(),
		WebserverAdminUsername:         pgtype.Text{String: req.GetWebServerProperties().GetWebserverAdminUsername(), Valid: true},
		WebserverAdminPasswordSecretID: pgtype.Int4{Int32: secret, Valid: true},

		BucketID:         pgtype.Text{String: req.StorageProperties.GetBucketId(), Valid: true},
		DagFolderPath:    pgtype.Text{String: req.GetStorageProperties().GetPath().GetDagFolderPath(), Valid: true},
		PluginFolderPath: pgtype.Text{String: req.GetStorageProperties().GetPath().GetPluginFolderPath(), Valid: true},
		RequirementPath:  pgtype.Text{String: req.GetStorageProperties().GetPath().GetRequirementPath(), Valid: true},
		LogFolder:        pgtype.Text{String: req.GetStorageProperties().GetPath().GetLogFolder(), Valid: true},

		DeploymentID:                deployment.DeploymentId,
		DeploymentStatusState:       pb.DpaiDeploymentState_DPAI_PENDING.String(),
		DeploymentStatusDisplayName: pgtype.Text{String: "Pending", Valid: true},
		DeploymentStatusMessage:     pgtype.Text{String: "Deployment is in pending state.", Valid: true},
		CreatedBy:                   "CallerOfTheAPI",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create the workspace. Error message: %+v", err)
	}
	result, err := getAirflowWithDeploymentStatus(s, airflow)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DpaiServer) DpaiAirflowGetById(ctx context.Context, req *pb.DpaiAirflowGetByIdRequest) (*pb.DpaiAirflow, error) {

	log.Printf("Get Airflow By Id: %s", req.GetId())
	queryResult, err := s.Sql.GetAirflowById(context.Background(), db.GetAirflowByIdParams{
		ID:             req.GetId(),
		CloudAccountID: req.GetCloudAccountId(),
	})
	if err != nil {
		log.Printf("Failed to get Airflow with error : %+v", err)
		return nil, err
	}

	result, err := getAirflowWithDeploymentStatus(s, queryResult)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DpaiServer) DpaiAirflowGetByName(ctx context.Context, req *pb.DpaiAirflowGetByNameRequest) (*pb.DpaiAirflow, error) {
	log.Printf("Get Airflow By Name: %s", req.GetName())
	queryResult, err := s.Sql.GetAirflowByName(context.Background(), db.GetAirflowByNameParams{
		CloudAccountID: req.CloudAccountId,
		Name:           req.GetName(),
	})
	if err != nil {
		log.Printf("Failed to get Airflow with error : %+v", err)
		return nil, err
	}

	result, err := getAirflowWithDeploymentStatus(s, queryResult)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DpaiServer) DpaiAirflowUpdate(ctx context.Context, req *pb.DpaiAirflowUpdateRequest) (*pb.DpaiAirflow, error) {
	id := req.GetId()
	log.Printf("Update Airflow By Id: %s", id)
	temp, err := s.Sql.GetUpdateAirflow(context.Background(), id)
	if err != nil {
		log.Printf("No record found for id : %s Error: %+v", id, err)
		return nil, err
	}

	if req.GetDescription() == "" {
		req.Description = temp.Description.String
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

	updateParams := db.UpdateAirflowParams{
		ID:          id,
		Tags:        tags,
		Description: pgtype.Text{String: req.Description, Valid: true},
	}

	queryResult, err := s.Sql.UpdateAirflow(context.Background(), updateParams)
	if err != nil {
		log.Printf("Failed to update Airflow with error : %+v", err)
		return nil, err
	}

	var resultTags map[string]string
	json.Unmarshal(queryResult.Tags, &resultTags)

	result, err := getAirflowWithDeploymentStatus(s, queryResult)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DpaiServer) DpaiAirflowDelete(ctx context.Context, req *pb.DpaiAirflowDeleteRequest) (*pb.DpaiDeploymentResponse, error) {

	log.Printf("Delete Airflow by Id: %s", req.GetId())
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("Not able to parse the input request to Json data. Error: %+v", err)
		return nil, err
	}

	data, err := s.Sql.GetAirflowById(context.Background(), db.GetAirflowByIdParams{
		ID:             req.GetId(),
		CloudAccountID: req.GetCloudAccountId(),
	})
	if err != nil {
		return nil, err
	}

	//Removing Condition on deleteing instance
	// if data.DeploymentStatusState != pb.DpaiDeploymentState_DPAI_SUCCESS.String() {
	// 	return &pb.DpaiDeploymentResponse{
	// 		Status: &pb.DpaiDeploymentStatus{
	// 			State:       pb.DpaiDeploymentState_DPAI_FAILED,
	// 			DisplayName: "Failed",
	// 			Message:     fmt.Sprintf("Only airflow in %s state can be deleted. Current State: %s", pb.DpaiDeploymentState_DPAI_SUCCESS.String(), data.DeploymentStatusState),
	// 		},
	// 	}, fmt.Errorf("failed to delete the deployment. Error message: Deployment is in success state")
	// }

	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		CloudAccountId:  req.GetCloudAccountId(),
		WorkspaceId:     data.WorkspaceID,
		ServiceId:       req.GetId(),
		ServiceType:     pb.DpaiServiceType_DPAI_AIRFLOW,
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

// TODO: Refector the below methods to bring the deployment status from the deployment table
func (s *DpaiServer) DpaiAirflowRestart(ctx context.Context, req *pb.DpaiAirflowRestartRequest) (*pb.DpaiDeploymentResponse, error) {
	id := req.GetId()

	log.Printf("Restart Airflow by Id: %s", id)
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("Not able to parse the input request to Json data. Error: %+v", err)
		return nil, err
	}
	data, err := s.Sql.GetAirflowById(context.Background(), db.GetAirflowByIdParams{
		ID: id,
	})
	if err != nil {
		return nil, err
	}
	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		WorkspaceId:     data.WorkspaceID,
		ServiceId:       req.GetId(),
		ServiceType:     pb.DpaiServiceType_DPAI_AIRFLOW,
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

// TODO: Refector the below methods to bring the deployment status from the deployment table
func (s *DpaiServer) DpaiAirflowResize(ctx context.Context, req *pb.DpaiAirflowResizeRequest) (*pb.DpaiDeploymentResponse, error) {
	id := req.GetId()

	log.Printf("Resize Airflow by Id: %s", id)
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("Not able to parse the input request to Json data. Error: %+v", err)
		return nil, err
	}

	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		ServiceId:       req.GetId(),
		ServiceType:     pb.DpaiServiceType_DPAI_AIRFLOW,
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

func (s *DpaiServer) DpaiAirflowListUpgrade(ctx context.Context, req *pb.DpaiAirflowListUpgradeRequest) (*pb.DpaiAirflowVersionListResponse, error) {
	id := req.GetId()
	log.Printf("List Airflow versions available for Airflow instance: %s", id)

	data, err := s.Sql.GetAirflowById(context.Background(), db.GetAirflowByIdParams{
		ID: id,
	})
	if err != nil {
		log.Printf("No Airflow instance found with the id: %s : %+v", req.GetId(), err)
		return nil, err
	}

	queryResults, err := s.Sql.ListAirflowVersionUpgrades(context.Background(), pgtype.Text{String: data.Version, Valid: true})
	if err != nil {
		log.Printf("Failed to list Airflow version upgrades with error : %+v", err)
		return nil, err
	}

	var results []*pb.DpaiAirflowVersion
	for _, queryResult := range queryResults {

		result, err := airflowVersionSqlToProto(queryResult)
		if err != nil {
			return nil, err
		}
		results = append(results, &result)
	}
	return &pb.DpaiAirflowVersionListResponse{
		Data: results}, nil
}

// TODO: Refector the below methods to bring the deployment status from the deployment table
func (s *DpaiServer) DpaiAirflowUpgrade(ctx context.Context, req *pb.DpaiAirflowUpgradeRequest) (*pb.DpaiDeploymentResponse, error) {

	if req.GetVersion() == "" || req.GetId() == "" {
		return nil, fmt.Errorf("missing input: Id and VersionId are mandatory for the upgrade operation")
	}

	log.Printf("Upgrade the Airflow with Id: %s to version: %s", req.GetId(), req.GetVersion())
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("Not able to parse the input request to Json data. Error: %+v", err)
		return nil, err
	}

	deployment, err := s.DpaiDeploymentCreate(ctx, &pb.DpaiDeploymentCreateRequest{
		ServiceId:       req.GetId(),
		ServiceType:     pb.DpaiServiceType_DPAI_AIRFLOW,
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
