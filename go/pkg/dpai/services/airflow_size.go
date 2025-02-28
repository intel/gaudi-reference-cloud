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

func airflowSizeProtoToSql(data db.AirflowSize) pb.DpaiAirflowSize {
	return pb.DpaiAirflowSize{
		Id:                    data.ID,
		Name:                  data.Name,
		Description:           data.Description.String,
		NumberOfNodesDefault:  data.NumberOfNodesDefault,
		NodeSizeId:            data.NodeSizeID,
		BackendDatabaseSizeId: data.BackendDatabaseSizeID.String,
		AirflowSizeWebserverProperties: &pb.DpaiAirflowSizeWebServerProperties{
			WebserverCount:         data.WebserverCount.Int32,
			WebserverCpuLimit:      data.WebserverCpuLimit.String,
			WebserverMemoryLimit:   data.WebserverMemoryLimit.String,
			WebserverCpuRequest:    data.WebserverCpuRequest.String,
			WebserverMemoryRequest: data.WebserverMemoryRequest.String,
		},
		LogDirectoryDiskSize: data.LogDirectoryDiskSize.String,
		RedisDiskSize:        data.RedisDiskSize.String,

		AirflowSchedulerProperties: &pb.DpaiAirflowSchedulerProperties{
			SchedularCountDefault:  data.SchedularCountDefault.Int32,
			SchedulerCountMin:      data.SchedulerCountMin.Int32,
			SchedulerCountMax:      data.SchedulerCountMax.Int32,
			SchedulerCpuLimit:      data.SchedulerCpuLimit.String,
			SchedulerMemoryLimit:   data.SchedulerMemoryLimit.String,
			SchedulerMemoryRequest: data.SchedulerMemoryRequest.String,
			SchedulerCpuRequest:    data.SchedulerCpuRequest.String,
		},

		AirflowWorkerProperties: &pb.DpaiAirflowWorkerProperties{
			WorkerCountDefault:  data.WorkerCountDefault.Int32,
			WorkerCountMin:      data.WorkerCountMin.Int32,
			WorkerCountMax:      data.WorkerCountMax.Int32,
			WorkerMemoryLimit:   data.WorkerMemoryLimit.String,
			WorkerMemoryRequest: data.WorkerMemoryRequest.String,
			WorkerCpuLimit:      data.WorkerCpuLimit.String,
			WorkerCpuRequest:    data.WorkerCpuRequest.String,
		},

		AirflowTriggerProperties: &pb.DpaiAirflowTriggerProperties{
			TriggerCount:         data.TriggerCount.Int32,
			TriggerMemoryLimit:   data.TriggerMemoryLimit.String,
			TriggerMemoryRequest: data.TriggerMemoryRequest.String,
			TriggerCpuLimit:      data.TriggerCpuLimit.String,
			TriggerCpuRequest:    data.TriggerCpuRequest.String,
		},

		Metadata: &pb.DpaiMeta{
			CreatedAt: timestamppb.New(data.CreatedAt.Time),
			CreatedBy: data.CreatedBy,
			UpdatedAt: timestamppb.New(data.UpdatedAt.Time),
			UpdatedBy: data.UpdatedBy.String,
		},
	}
}

func (s *DpaiServer) DpaiAirflowSizeList(ctx context.Context, req *pb.DpaiAirflowSizeListRequest) (*pb.DpaiAirflowSizeListResponse, error) {
	log.Println("List Airflow Sizes...")
	records, err := s.Sql.ListAirflowSize(context.Background())
	if err != nil {
		log.Printf("Failed to list airflow size with error : %+v", err)
		return nil, err
	}

	var prototypes []*pb.DpaiAirflowSize
	for _, record := range records {
		prototype := airflowSizeProtoToSql(record)
		prototypes = append(prototypes, &prototype)
	}
	return &pb.DpaiAirflowSizeListResponse{
		Data: prototypes}, nil
}

func (s *DpaiServer) DpaiAirflowSizeCreate(ctx context.Context, req *pb.DpaiAirflowSizeCreateRequest) (*pb.DpaiAirflowSize, error) {

	log.Println("Create AirflowSize:")

	create, err := s.Sql.CreateAirflowSize(ctx, db.CreateAirflowSizeParams{
		ID:                    uuid.New().String(),
		Name:                  req.GetName(),
		Description:           pgtype.Text{String: req.GetDescription(), Valid: true},
		NumberOfNodesDefault:  req.GetNumberOfNodesDefault(),
		NodeSizeID:            req.GetNodeSizeId(),
		BackendDatabaseSizeID: pgtype.Text{String: req.GetBackendDatabaseSizeId(), Valid: true},

		WebserverCount:         pgtype.Int4{Int32: req.AirflowSizeWebserverProperties.GetWebserverCount(), Valid: true},
		WebserverCpuLimit:      pgtype.Text{String: req.AirflowSizeWebserverProperties.GetWebserverCpuLimit(), Valid: true},
		WebserverMemoryLimit:   pgtype.Text{String: req.AirflowSizeWebserverProperties.GetWebserverMemoryLimit(), Valid: true},
		WebserverCpuRequest:    pgtype.Text{String: req.AirflowSizeWebserverProperties.GetWebserverCpuRequest(), Valid: true},
		WebserverMemoryRequest: pgtype.Text{String: req.AirflowSizeWebserverProperties.GetWebserverMemoryRequest(), Valid: true},

		LogDirectoryDiskSize: pgtype.Text{String: req.GetLogDirectoryDiskSize(), Valid: true},
		RedisDiskSize:        pgtype.Text{String: req.GetRedisDiskSize(), Valid: true},

		SchedularCountDefault:  pgtype.Int4{Int32: req.AirflowSchedulerProperties.GetSchedularCountDefault(), Valid: true},
		SchedulerCountMin:      pgtype.Int4{Int32: req.AirflowSchedulerProperties.GetSchedulerCountMin(), Valid: true},
		SchedulerCountMax:      pgtype.Int4{Int32: req.AirflowSchedulerProperties.GetSchedulerCountMax(), Valid: true},
		SchedulerCpuLimit:      pgtype.Text{String: req.AirflowSchedulerProperties.GetSchedulerCpuLimit(), Valid: true},
		SchedulerMemoryLimit:   pgtype.Text{String: req.AirflowSchedulerProperties.GetSchedulerMemoryLimit(), Valid: true},
		SchedulerMemoryRequest: pgtype.Text{String: req.AirflowSchedulerProperties.GetSchedulerMemoryRequest(), Valid: true},
		SchedulerCpuRequest:    pgtype.Text{String: req.AirflowSchedulerProperties.GetSchedulerCpuRequest(), Valid: true},

		WorkerCountDefault:  pgtype.Int4{Int32: req.AirflowWorkerProperties.GetWorkerCountDefault(), Valid: true},
		WorkerCountMin:      pgtype.Int4{Int32: req.AirflowWorkerProperties.GetWorkerCountMin(), Valid: true},
		WorkerCountMax:      pgtype.Int4{Int32: req.AirflowWorkerProperties.GetWorkerCountMax(), Valid: true},
		WorkerMemoryLimit:   pgtype.Text{String: req.AirflowWorkerProperties.GetWorkerMemoryLimit(), Valid: true},
		WorkerMemoryRequest: pgtype.Text{String: req.AirflowWorkerProperties.GetWorkerMemoryRequest(), Valid: true},
		WorkerCpuLimit:      pgtype.Text{String: req.AirflowWorkerProperties.GetWorkerCpuLimit(), Valid: true},
		WorkerCpuRequest:    pgtype.Text{String: req.AirflowWorkerProperties.GetWorkerCpuRequest(), Valid: true},

		TriggerCount:         pgtype.Int4{Int32: req.AirflowTriggerProperties.GetTriggerCount(), Valid: true},
		TriggerMemoryLimit:   pgtype.Text{String: req.AirflowTriggerProperties.GetTriggerMemoryLimit(), Valid: true},
		TriggerMemoryRequest: pgtype.Text{String: req.AirflowTriggerProperties.GetTriggerMemoryRequest(), Valid: true},
		TriggerCpuLimit:      pgtype.Text{String: req.AirflowTriggerProperties.GetTriggerCpuLimit(), Valid: true},
		TriggerCpuRequest:    pgtype.Text{String: req.AirflowTriggerProperties.GetTriggerCpuRequest(), Valid: true},

		CreatedBy: "CallerOfTheAPI",
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	resp := airflowSizeProtoToSql(create)
	return &resp, nil
}

func (s *DpaiServer) DpaiAirflowSizeGetById(ctx context.Context, req *pb.DpaiAirflowSizeGetByIdRequest) (*pb.DpaiAirflowSize, error) {

	log.Printf("Get Workspaces By Id: %s", req.GetId())
	data, err := s.Sql.GetAirflowSizeById(context.Background(), req.GetId())
	if err != nil {
		log.Printf("Failed to list AirflowSize with error : %+v", err)
		return nil, err
	}

	resp := airflowSizeProtoToSql(data)

	return &resp, nil
}

func (s *DpaiServer) DpaiAirflowSizeGetByName(ctx context.Context, req *pb.DpaiAirflowSizeGetByNameRequest) (*pb.DpaiAirflowSize, error) {
	name := req.GetName()
	log.Printf("Get Workspaces By Name: %s", name)
	data, err := s.Sql.GetAirflowSizeByName(context.Background(), name)
	if err != nil {
		log.Printf("Failed to list airflow size with error : %+v", err)
		return nil, err
	}

	resp := airflowSizeProtoToSql(data)

	return &resp, nil
}

// ToDo: DpaiAirflowSizeUpdate
func (s *DpaiServer) DpaiAirflowSizeUpdate(ctx context.Context, req *pb.DpaiAirflowSizeUpdateRequest) (*pb.DpaiAirflowSize, error) {
	id := req.GetId()
	log.Printf("Update Workspaces By Id: %s", id)
	temp, err := s.Sql.GetAirflowSizeById(context.Background(), id)
	if err != nil {
		log.Printf("No record found for id : %s Error: %+v", id, err)
		return nil, err
	}

	description := req.GetDescription()
	if description == "" {
		description = temp.Description.String
	}
	numberOfNodesDefault := req.GetNumberOfNodesDefault()
	if numberOfNodesDefault == 0 {
		numberOfNodesDefault = temp.NumberOfNodesDefault
	}
	nodeSizeID := req.GetNodeSizeId()
	if nodeSizeID == "" {
		nodeSizeID = temp.NodeSizeID
	}
	if req.GetBackendDatabaseSizeId() == "" {
		req.BackendDatabaseSizeId = temp.BackendDatabaseSizeID.String
	}

	webserverAirflowCount := req.AirflowSizeWebserverProperties.GetWebserverCount()
	if webserverAirflowCount == 0 {
		webserverAirflowCount = temp.WebserverCount.Int32
	}
	webserverAirflowCpuLimit := req.AirflowSizeWebserverProperties.GetWebserverCpuLimit()
	if webserverAirflowCpuLimit == "" {
		webserverAirflowCpuLimit = temp.WebserverCpuLimit.String
	}
	webserverAirflowMemoryLimit := req.AirflowSizeWebserverProperties.GetWebserverMemoryLimit()
	if webserverAirflowMemoryLimit == "" {
		webserverAirflowMemoryLimit = temp.WebserverMemoryLimit.String
	}
	webserverAirflowCpuRequest := req.AirflowSizeWebserverProperties.GetWebserverCpuRequest()
	if webserverAirflowCpuRequest == "" {
		webserverAirflowCpuRequest = temp.WebserverCpuRequest.String
	}
	webserverAirflowMemoryRequest := req.AirflowSizeWebserverProperties.GetWebserverMemoryRequest()
	if webserverAirflowMemoryRequest == "" {
		webserverAirflowMemoryRequest = temp.WebserverMemoryRequest.String
	}

	logDirectoryDiskSize := req.GetLogDirectoryDiskSize()
	if logDirectoryDiskSize == "" {
		logDirectoryDiskSize = temp.LogDirectoryDiskSize.String
	}

	redisDiskSize := req.GetRedisDiskSize()
	if redisDiskSize == "" {
		redisDiskSize = temp.RedisDiskSize.String
	}

	schedularAirflowCountDefault := req.AirflowSchedulerProperties.GetSchedularCountDefault()
	if schedularAirflowCountDefault == 0 {
		schedularAirflowCountDefault = temp.SchedularCountDefault.Int32
	}
	schedulerAirflowCountMin := req.AirflowSchedulerProperties.GetSchedulerCountMin()
	if schedulerAirflowCountMin == 0 {
		schedulerAirflowCountMin = temp.SchedulerCountMin.Int32
	}
	schedulerAirflowCountMax := req.AirflowSchedulerProperties.GetSchedulerCountMax()
	if schedulerAirflowCountMax == 0 {
		schedulerAirflowCountMax = temp.SchedulerCountMax.Int32
	}
	schedulerAirflowCpuLimit := req.AirflowSchedulerProperties.GetSchedulerCpuLimit()
	if schedulerAirflowCpuLimit == "" {
		schedulerAirflowCpuLimit = temp.SchedulerCpuLimit.String
	}
	schedulerAirflowMemoryLimit := req.AirflowSchedulerProperties.GetSchedulerMemoryLimit()
	if schedulerAirflowMemoryLimit == "" {
		schedulerAirflowMemoryLimit = temp.SchedulerMemoryLimit.String
	}
	schedulerAirflowMemoryRequest := req.AirflowSchedulerProperties.GetSchedulerMemoryRequest()
	if schedulerAirflowMemoryRequest == "" {
		schedulerAirflowMemoryRequest = temp.SchedulerMemoryRequest.String
	}
	schedulerAirflowCpuRequest := req.AirflowSchedulerProperties.GetSchedulerCpuRequest()
	if schedulerAirflowCpuRequest == "" {
		schedulerAirflowCpuRequest = temp.SchedulerCpuRequest.String
	}

	workerAirflowCountDefault := req.AirflowWorkerProperties.GetWorkerCountDefault()
	if workerAirflowCountDefault == 0 {
		workerAirflowCountDefault = temp.WorkerCountDefault.Int32
	}
	workerAirflowCountMin := req.AirflowWorkerProperties.GetWorkerCountMin()
	if workerAirflowCountMin == 0 {
		workerAirflowCountMin = temp.WorkerCountMax.Int32
	}
	workerAirflowCountMax := req.AirflowWorkerProperties.GetWorkerCountMax()
	if workerAirflowCountMax == 0 {
		workerAirflowCountMax = temp.WorkerCountMax.Int32
	}
	workerAirflowMemoryLimit := req.AirflowWorkerProperties.GetWorkerMemoryLimit()
	if workerAirflowMemoryLimit == "" {
		workerAirflowMemoryLimit = temp.WorkerMemoryLimit.String
	}
	workerAirflowMemoryRequest := req.AirflowWorkerProperties.GetWorkerMemoryRequest()
	if workerAirflowMemoryRequest == "" {
		workerAirflowMemoryRequest = temp.WorkerMemoryRequest.String
	}
	workerAirflowCpuLimit := req.AirflowWorkerProperties.GetWorkerCpuLimit()
	if workerAirflowCpuLimit == "" {
		workerAirflowCpuLimit = temp.WorkerCpuLimit.String
	}
	workerAirflowCpuRequest := req.AirflowWorkerProperties.GetWorkerCpuRequest()
	if workerAirflowCpuRequest == "" {
		workerAirflowCpuRequest = temp.WorkerCpuRequest.String
	}

	triggerAirflowCount := req.AirflowTriggerProperties.GetTriggerCount()
	if triggerAirflowCount == 0 {
		triggerAirflowCount = temp.TriggerCount.Int32
	}
	triggerAirflowMemoryLimit := req.AirflowTriggerProperties.GetTriggerMemoryLimit()
	if triggerAirflowMemoryLimit == "" {
		triggerAirflowMemoryLimit = temp.TriggerMemoryLimit.String
	}
	triggerAirflowMemoryRequest := req.AirflowTriggerProperties.GetTriggerMemoryRequest()
	if triggerAirflowMemoryRequest == "" {
		triggerAirflowMemoryRequest = temp.TriggerMemoryRequest.String
	}
	triggerAirflowCpuLimit := req.AirflowTriggerProperties.GetTriggerCpuLimit()
	if triggerAirflowCpuLimit == "" {
		triggerAirflowCpuLimit = temp.TriggerCpuLimit.String
	}
	triggerAirflowCpuRequest := req.AirflowTriggerProperties.GetTriggerCpuRequest()
	if triggerAirflowCpuRequest == "" {
		triggerAirflowCpuRequest = temp.TriggerCpuRequest.String
	}

	updateUserParams := db.UpdateAirflowSizeParams{
		ID:                     id,
		Description:            pgtype.Text{String: description, Valid: true},
		NumberOfNodesDefault:   numberOfNodesDefault,
		NodeSizeID:             nodeSizeID,
		BackendDatabaseSizeID:  pgtype.Text{String: req.BackendDatabaseSizeId, Valid: true},
		WebserverCount:         pgtype.Int4{Int32: webserverAirflowCount, Valid: true},
		WebserverCpuLimit:      pgtype.Text{String: webserverAirflowCpuLimit, Valid: true},
		WebserverMemoryLimit:   pgtype.Text{String: webserverAirflowMemoryLimit, Valid: true},
		WebserverCpuRequest:    pgtype.Text{String: webserverAirflowCpuRequest, Valid: true},
		WebserverMemoryRequest: pgtype.Text{String: webserverAirflowMemoryRequest, Valid: true},
		LogDirectoryDiskSize:   pgtype.Text{String: logDirectoryDiskSize, Valid: true},
		RedisDiskSize:          pgtype.Text{String: redisDiskSize, Valid: true},
		SchedularCountDefault:  pgtype.Int4{Int32: schedularAirflowCountDefault, Valid: true},
		SchedulerCountMin:      pgtype.Int4{Int32: schedulerAirflowCountMin, Valid: true},
		SchedulerCountMax:      pgtype.Int4{Int32: schedulerAirflowCountMax, Valid: true},
		SchedulerCpuLimit:      pgtype.Text{String: schedulerAirflowCpuLimit, Valid: true},
		SchedulerMemoryLimit:   pgtype.Text{String: schedulerAirflowMemoryLimit, Valid: true},
		SchedulerMemoryRequest: pgtype.Text{String: schedulerAirflowMemoryRequest, Valid: true},
		SchedulerCpuRequest:    pgtype.Text{String: schedulerAirflowCpuRequest, Valid: true},
		WorkerCountDefault:     pgtype.Int4{Int32: workerAirflowCountDefault, Valid: true},
		WorkerCountMin:         pgtype.Int4{Int32: workerAirflowCountMin, Valid: true},
		WorkerCountMax:         pgtype.Int4{Int32: workerAirflowCountMax, Valid: true},
		WorkerMemoryLimit:      pgtype.Text{String: workerAirflowMemoryLimit, Valid: true},
		WorkerMemoryRequest:    pgtype.Text{String: workerAirflowMemoryRequest, Valid: true},
		WorkerCpuLimit:         pgtype.Text{String: workerAirflowCpuLimit, Valid: true},
		WorkerCpuRequest:       pgtype.Text{String: workerAirflowCpuRequest, Valid: true},
		TriggerCount:           pgtype.Int4{Int32: triggerAirflowCount, Valid: true},
		TriggerMemoryLimit:     pgtype.Text{String: triggerAirflowMemoryLimit, Valid: true},
		TriggerMemoryRequest:   pgtype.Text{String: triggerAirflowMemoryRequest, Valid: true},
		TriggerCpuLimit:        pgtype.Text{String: triggerAirflowCpuLimit, Valid: true},
		TriggerCpuRequest:      pgtype.Text{String: triggerAirflowCpuRequest, Valid: true},
	}

	data, err := s.Sql.UpdateAirflowSize(context.Background(), updateUserParams)
	if err != nil {
		log.Printf("Failed to update Airflow Size with error : %+v", err)
		return nil, err
	}
	resp := airflowSizeProtoToSql(data)

	return &resp, nil
}

func (s *DpaiServer) DpaiAirflowSizeDelete(ctx context.Context, req *pb.DpaiAirflowSizeDeleteRequest) (*pb.DpaiAirflowSizeDeleteResponse, error) {
	id := req.GetId()

	status := true
	var errorMessage string
	log.Printf("Delete AirflowSize by Id: %s", id)
	err := s.Sql.DeleteAirflowSize(context.Background(), id)
	if err != nil {
		log.Printf("Failed to delete AirflowSize with error : %+v", err)
		status = false
		errorMessage = fmt.Sprintf("Failed to delete Airflow Size with error : %+v", err)
	}

	return &pb.DpaiAirflowSizeDeleteResponse{Success: status, ErrorMessage: errorMessage}, nil
}
