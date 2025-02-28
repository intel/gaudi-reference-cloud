// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quota_management/database/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	accountTypeMap = map[string]string{
		"ACCOUNT_TYPE_ENTERPRISE":         "ENTERPRISE",
		"ACCOUNT_TYPE_ENTERPRISE_PENDING": "ENTERPRISE_PENDING",
		"ACCOUNT_TYPE_INTEL":              "INTEL",
		"ACCOUNT_TYPE_PREMIUM":            "PREMIUM",
		"ACCOUNT_TYPE_STANDARD":           "STANDARD",
		"ACCOUNT_TYPE_MEMBER":             "MEMBER",
		"ACCOUNT_TYPE_UNSPECIFIED":        "UNSPECIFIED",
	}
)

// QuotaManagementServiceClient is used to implement pb.UnimplementedQuotaManagementServiceServer & pb.UnimplementedQuotaManagementServiceServer
type QuotaManagementServiceClient struct {
	pb.UnimplementedQuotaManagementServiceServer
	pb.UnimplementedQuotaManagementPrivateServiceServer

	session                   *sql.DB
	cloudAccountServiceClient pb.CloudAccountServiceClient
	selectedRegion            string
}

func NewQuotaManagementServiceClient(ctx context.Context, session *sql.DB, cloudAccountSvc pb.CloudAccountServiceClient, selectedRegion string) (*QuotaManagementServiceClient, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}

	qmSrv := QuotaManagementServiceClient{
		session:                   session,
		cloudAccountServiceClient: cloudAccountSvc,
		selectedRegion:            selectedRegion,
	}

	return &qmSrv, nil
}

func (s *QuotaManagementServiceClient) Register(ctx context.Context, req *pb.ServiceQuotaRegistrationRequest) (*pb.ServiceQuotaRegistrationResponse, error) {
	// Start a new transaction
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.registerService").Start()
	defer span.End()
	logger.Info("entering register service")
	defer logger.Info("returning from register service")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	defer tx.Rollback()

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}
	// generate unique service id for this service
	serviceId := s.generateUniqueServiceId()
	// Insert service into
	newRegisteredService, err := query.InsertService(ctx, tx, serviceId, req.ServiceName, req.Region)
	if err != nil {
		logger.Error(err, "failed to register new service")
		return nil, status.Error(codes.Internal, "service registration failed")
	}
	logger.Info("newly created service : ", "service in db : ", newRegisteredService)

	// for the newly registered service, add all the resource quotas
	var currentServiceResources []*pb.ServiceResource
	for _, serviceResource := range req.ServiceResources {
		logger.Info("adding service resource", "details:", serviceResource)
		newServiceResource, err := query.InsertServiceResource(ctx, tx, serviceId, serviceResource.Name, serviceResource.QuotaUnit, serviceResource.MaxLimit)
		if err != nil {
			logger.Error(err, "failed to register new service")
			return nil, status.Error(codes.Internal, "database transaction failed")
		}
		currentServiceResources = append(currentServiceResources, newServiceResource)
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	// Convert the quota to the protobuf message type
	pbService := &pb.ServiceQuotaRegistrationResponse{
		ServiceId:        newRegisteredService.ServiceId,
		ServiceName:      newRegisteredService.ServiceName,
		Region:           newRegisteredService.Region,
		ServiceResources: currentServiceResources,
	}
	logger.Info("response from register new service: ", "response from db : ", pbService)

	return pbService, nil
}

func (s *QuotaManagementServiceClient) ListRegisteredServices(ctx context.Context, req *pb.ServicesListRequest) (*pb.ServicesListResponse, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.listRegisteredServices").Start()
	defer span.End()
	logger.Info("entering registred services list")
	defer logger.Info("returning from registered services list")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	defer tx.Rollback()

	registeredServices, err := query.ListRegisteredServices(ctx, tx)
	if err != nil {
		logger.Error(err, "failed to get list of registered services")
		return nil, err
	}
	logger.Info("list of registered service retrieved: ", "service in db : ", registeredServices)

	return &pb.ServicesListResponse{
		Services: registeredServices,
	}, nil
}

func (s *QuotaManagementServiceClient) ListServiceResources(ctx context.Context, req *pb.ServiceResourcesListRequest) (*pb.ServiceResourcesListResponse, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.ListServiceResources").Start()
	defer span.End()
	logger.Info("entering registred list service resources")
	defer logger.Info("returning from list service resources")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	defer tx.Rollback()

	registeredService, err := query.GetRegisteredService(ctx, tx, req.ServiceId)
	if err != nil {
		logger.Error(err, "failed to get registered service")
		return nil, err
	}
	logger.Info("registered service retrieved: ", "service in db : ", registeredService)

	serviceResourcesList, err := query.ListServiceResources(ctx, tx, req.ServiceId)
	if err != nil {
		logger.Error(err, "failed to get list of service resources")
		return nil, err
	}
	logger.Info("list of service resources retrieved: ", "service in db : ", serviceResourcesList)

	return &pb.ServiceResourcesListResponse{
		ServiceId:        registeredService.ServiceId,
		ServiceName:      registeredService.ServiceName,
		ServiceResources: serviceResourcesList,
	}, nil
}

func (s *QuotaManagementServiceClient) UpdateServiceRegistration(ctx context.Context, req *pb.UpdateServiceRegistrationRequest) (*pb.ServiceQuotaRegistrationResponse, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.updateServiceRegistration").Start()
	defer span.End()
	logger.Info("entering update service registration")
	defer logger.Info("returning from update service registration")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	defer tx.Rollback()

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}

	logger.Info("update service registration params", "request details:", req)
	registeredService, err := query.GetRegisteredService(ctx, tx, req.ServiceId)
	if err != nil {
		logger.Error(err, "failed to get registered service")
		return nil, err
	}
	logger.Info("registered service retrieved: ", "service in db : ", registeredService)

	// TODO enable this check if required. First release version has it disabled
	// validate if the new max resource limits being set for the service
	// is NOT less than those being used by any of the existing quotas for that service
	//allQuotasForService, err := query.GetServiceResourceQuotas(ctx, tx, req.ServiceId, "")
	//if err != nil {
	//	logger.Error(err, "failed to get service resource quotas for service")
	//	return nil, status.Error(codes.Internal, "database transaction failed")
	//}

	// for each resource type in the service, store the minimum limit value,
	// which will be later used to check when service registration values are being updated
	//perResourceMinQuota := s.getMinLimitsResource(allQuotasForService)

	var updatedServiceResources []*pb.ServiceResource
	for _, resource := range req.ServiceResources {
		logger.Info("current service resource", "service id", req.ServiceId, "resource name", resource.Name, "quota unit", resource.QuotaUnit, "max limit", resource.MaxLimit)
		serviceResource, err := query.GetServiceResource(ctx, tx, req.ServiceId, resource.Name)
		if err != nil {
			logger.Info("resource type not found for update", "skipping updation of this resource", resource.Name, "error: ", err)
			newServiceResource, err := query.InsertServiceResource(ctx, tx, req.ServiceId, resource.Name, resource.QuotaUnit, resource.MaxLimit)
			if err != nil {
				logger.Error(err, "failed to update service with new service resource")
				return nil, status.Error(codes.Internal, "database transaction failed")
			}
			updatedServiceResources = append(updatedServiceResources, newServiceResource)
		} else {
			logger.Info("service resource found", "details", serviceResource)
			// TODO enable this check if required. First release version has it disabled
			//if minLimit, ok := perResourceMinQuota[resource.Name]; ok {
			//	if resource.MaxLimit < minLimit {
			//		logger.Info("failed to update of service registration quotas: max limit is less than what has already been assigned")
			//		return nil, status.Errorf(codes.InvalidArgument, "not allowed to set max limit to less than that of one of the quotas already created")
			//	}
			//}
			logger.Info("update query params", "service id", req.ServiceId, "resource name", resource.Name, "quota unit", resource.QuotaUnit, "max limit", resource.MaxLimit)
			updatedServiceResource, err := query.UpdateServiceResource(ctx, tx, req.ServiceId, resource.Name, resource.MaxLimit)
			if err != nil {
				logger.Error(err, "database transaction for updating service resource failed")
				return nil, status.Error(codes.InvalidArgument, "database transaction for updating service resource failed")
			}

			logger.Info("response from update of service resource: ", "response from db : ", updatedServiceResource)
			updatedServiceResources = append(updatedServiceResources, updatedServiceResource)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	// Convert the quota to the protobuf message type
	updatedService := &pb.ServiceQuotaRegistrationResponse{
		ServiceId:        registeredService.ServiceId,
		ServiceName:      registeredService.ServiceName,
		Region:           registeredService.Region,
		ServiceResources: updatedServiceResources,
	}
	logger.Info("response from update of registered service: ", "response from db : ", updatedService)
	return updatedService, nil
}

func (s *QuotaManagementServiceClient) DeleteService(ctx context.Context, req *pb.DeleteServiceRequest) (*empty.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.deleteServiceRegistration").Start()
	defer span.End()
	logger.Info("entering delete service")
	defer logger.Info("returning from delete service")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	defer tx.Rollback()

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}

	registeredService, err := query.GetRegisteredService(ctx, tx, req.ServiceId)
	if err != nil {
		logger.Error(err, "failed to get registered service")
		return nil, err
	}
	logger.Info("found service to be deleted along with resources", "service ID", registeredService.ServiceId)
	// delete service
	err = query.DeleteService(ctx, tx, registeredService.ServiceId)
	if err != nil {
		logger.Error(err, "failed to delete registered service")
		return nil, status.Error(codes.Internal, "failed to delete registered service")
	}

	// delete service resources
	err = query.DeleteAllServiceResources(ctx, tx, registeredService.ServiceId)
	if err != nil {
		logger.Error(err, "failed to delete service resources")
		return nil, status.Error(codes.Internal, "failed to delete all resources for a service")
	}

	// delete service resource quotas
	err = query.DeleteAllServiceResourceQuota(ctx, tx, registeredService.ServiceId)
	if err != nil {
		if notFoundError(err) {
			logger.Info("no service resource quotas exist skipping deletion of service quotas")
		} else {
			logger.Error(err, "failed to delete service resource quotas")
			return nil, status.Error(codes.Internal, "failed to delete all resource quotas for a service")
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	return &emptypb.Empty{}, nil
}

func (s *QuotaManagementServiceClient) GetServiceResource(ctx context.Context, req *pb.ServiceResourceRequest) (*pb.ServiceResourceResponse, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.GetServiceResource").Start()
	defer span.End()
	logger.Info("entering get service resource")
	defer logger.Info("returning from get service resource")
	logger.Info("input to get service resource: ", "serviceid:", req.ServiceId, "resourceType", req.ResourceName)

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}

	registeredService, err := query.GetRegisteredService(ctx, tx, req.ServiceId)
	if err != nil {
		logger.Error(err, "failed to get registered service")
		return nil, err
	}
	serviceResource, err := query.GetServiceResource(ctx, tx, req.ServiceId, req.ResourceName)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get service resource")
	}

	// Convert the quota to the protobuf message type
	pbService := &pb.ServiceResourceResponse{
		ServiceId:   serviceResource.ServiceId,
		ServiceName: registeredService.ServiceName,
		ServiceResources: &pb.ServiceResource{
			Name:      serviceResource.ResourceName,
			QuotaUnit: serviceResource.QuotaUnit,
			MaxLimit:  serviceResource.MaxLimit,
		},
	}
	logger.Info("response from get service resource: ", "response from db : ", pbService)

	return pbService, nil
}

func (s *QuotaManagementServiceClient) CreateServiceQuota(ctx context.Context, req *pb.CreateServiceQuotaRequest) (*pb.CreateServiceQuotaResponse, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.CreateServiceQuota").Start()
	defer span.End()
	logger.Info("entering create service quota")
	defer logger.Info("returning from create service quota")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	defer tx.Rollback()

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}

	// check if the service resource already exists
	referenceResource, err := query.GetServiceResource(ctx, tx, req.ServiceId, req.ServiceQuotaResource.ResourceType)
	if err != nil {
		logger.Error(err, "failed to get registered resource type", "resource name", req.ServiceQuotaResource.ResourceType, "service id", req.ServiceId)
		return nil, status.Errorf(codes.Internal, "failed to create quota, add resource types that are not registered")
	}

	// check if requested quota limit exceeds max limit in service registration
	if req.ServiceQuotaResource.QuotaConfig.Limits > referenceResource.MaxLimit {
		return nil, status.Errorf(codes.InvalidArgument, "not allowed to create/update resource limit to more than that specified in service registration")
	}

	// fetch name and region for service
	registeredService, err := query.GetRegisteredService(ctx, tx, req.ServiceId)
	if err != nil {
		logger.Error(err, "failed to get registered service")
		return nil, err
	}

	logger.Info("registered service retrieved: ", "service in db : ", registeredService)

	// generate unique rule id for this service
	ruleId := s.generateUniqueServiceId()

	serviceQuotaResourceInserted, err := query.InsertServiceResourceQuota(ctx, tx, registeredService.ServiceId, req.ServiceQuotaResource.ResourceType,
		ruleId, req.ServiceQuotaResource.QuotaConfig.Limits, req.ServiceQuotaResource.QuotaConfig.QuotaUnit,
		req.ServiceQuotaResource.Scope.ScopeType, req.ServiceQuotaResource.Scope.ScopeValue,
		req.ServiceQuotaResource.Reason)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	serviceQuotaResponse := &pb.CreateServiceQuotaResponse{
		ServiceId:   registeredService.ServiceId,
		ServiceName: registeredService.ServiceName,
		Region:      registeredService.Region,
		ServiceQuotaResource: &pb.ServiceQuotaResource{
			ResourceType: serviceQuotaResourceInserted.ResourceName,
			QuotaConfig: &pb.QuotaConfig{
				Limits:    serviceQuotaResourceInserted.Limits,
				QuotaUnit: serviceQuotaResourceInserted.QuotaUnit,
			},
			Scope: &pb.QuotaScope{
				ScopeType:  serviceQuotaResourceInserted.QuotaScope,
				ScopeValue: serviceQuotaResourceInserted.QuotaScopeValue,
			},
			RuleId:      serviceQuotaResourceInserted.RuleId,
			Reason:      serviceQuotaResourceInserted.Reason,
			CreatedTime: timestamppb.New(serviceQuotaResourceInserted.CreatedTimestamp),
			UpdatedTime: timestamppb.New(serviceQuotaResourceInserted.UpdateTimestamp),
		},
	}
	logger.Info("new service quota inserted: ", "service quota : ", serviceQuotaResponse)

	return serviceQuotaResponse, nil

}

func (s *QuotaManagementServiceClient) GetServiceQuotaResource(ctx context.Context, req *pb.ServiceQuotaResourceRequest) (*pb.ServiceQuotaResourceResponse, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.updateServiceRegistration").Start()
	defer span.End()
	logger.Info("entering get service quota")
	defer logger.Info("returning from get service quota")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	defer tx.Rollback()

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}

	serviceId := req.ServiceId
	resourceType := req.ResourceType
	registeredService, err := query.GetRegisteredService(ctx, tx, serviceId)
	if err != nil {
		logger.Error(err, "failed to get registered service")
		return nil, err
	}

	serviceQuotaResourceResponse := &pb.ServiceQuotaResourceResponse{
		ServiceId:            registeredService.ServiceId,
		ServiceName:          registeredService.ServiceName,
		ServiceQuotaResource: []*pb.ServiceQuotaResource{},
	}

	// get quotas for service for specified resource type
	logger.Info("input to get resource quotas: ", "serviceid:", serviceId, "resourceType", resourceType)
	allQuotasForService, err := query.GetServiceResourceQuotas(ctx, tx, serviceId, resourceType)
	if err != nil {
		logger.Error(err, "failed to get service resource quotas")
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	for _, currQuota := range allQuotasForService {
		qRes := &pb.ServiceQuotaResource{
			ResourceType: currQuota.ResourceName,
			QuotaConfig: &pb.QuotaConfig{
				Limits:    currQuota.Limits,
				QuotaUnit: currQuota.QuotaUnit,
			},
			Scope: &pb.QuotaScope{
				ScopeType:  currQuota.QuotaScope,
				ScopeValue: convertToUserFriendlyAccount(ctx, currQuota.QuotaScopeValue),
			},
			RuleId:      currQuota.RuleId,
			Reason:      currQuota.Reason,
			CreatedTime: timestamppb.New(currQuota.CreatedTimestamp),
			UpdatedTime: timestamppb.New(currQuota.UpdateTimestamp),
		}
		serviceQuotaResourceResponse.ServiceQuotaResource = append(serviceQuotaResourceResponse.ServiceQuotaResource, qRes)
	}
	logger.Info("retrieved quotas for service: ", "service in db : ", serviceQuotaResourceResponse)

	return serviceQuotaResourceResponse, nil

}

func (s *QuotaManagementServiceClient) DeleteServiceQuotaResource(ctx context.Context, req *pb.DeleteServiceQuotaRequest) (*empty.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.deleteServiceRegistration").Start()
	defer span.End()
	logger.Info("entering delete service quota resource")
	defer logger.Info("returning from delete service quota resource")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	defer tx.Rollback()

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}

	err = query.DeleteServiceResourceQuota(ctx, tx, req.ServiceId, req.ResourceType, req.RuleId)
	if err != nil {
		logger.Error(err, "deleting service resource quota")
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	return &emptypb.Empty{}, nil

}

func (s *QuotaManagementServiceClient) UpdateServiceQuotaResource(ctx context.Context, req *pb.UpdateQuotaServiceRequest) (*pb.UpdateQuotaServiceResponse, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.updateServiceQuotaResource").Start()
	defer span.End()
	logger.Info("entering update service resource quota registration")
	defer logger.Info("returning from update service resource quota registration")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	defer tx.Rollback()

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}

	logger.Info("update service quota resource", "params", req)

	// check if entry to be updated exists
	allQuotasForService, err := query.GetServiceResourceQuotas(ctx, tx, req.ServiceId, req.ResourceType)
	if err != nil {
		logger.Error(err, "failed to get service resource quota")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	serviceResource, err := query.GetServiceResource(ctx, tx, req.ServiceId, req.ResourceType)
	if err != nil {
		logger.Error(err, "update service request", "resource type not found for update, skipping updation of this resource", req.ResourceType)
	} else {
		logger.Info("service resource retrived", "values", serviceResource)
		if req.QuotaConfig.Limits > serviceResource.MaxLimit {
			return nil, status.Error(codes.InvalidArgument, "not allowed to create/update resource limit to more than that specified in service registration")
		}
	}
	ruleIdExists := false
	for _, currQuota := range allQuotasForService {
		if currQuota.RuleId == req.RuleId {
			ruleIdExists = true
			logger.Info("matching ruleID found", "ruleId:", currQuota.RuleId)
		}
	}

	if !ruleIdExists {
		return nil, status.Errorf(codes.Internal, "service resource updation failed due to incorrect input")
	}

	updateServiceQuotaResource, err := query.UpdateServiceResourceQuota(ctx, tx, req.ServiceId, req.ResourceType, req.RuleId, req.Reason, req.QuotaConfig.Limits)
	if err != nil {
		logger.Error(err, "update of service resource quota failed")
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	updateQuotaServiceResponse := &pb.UpdateQuotaServiceResponse{
		ServiceId: updateServiceQuotaResource.ServiceId,
		ServiceQuotaResource: &pb.ServiceQuotaResource{
			ResourceType: updateServiceQuotaResource.ResourceName,
			QuotaConfig: &pb.QuotaConfig{
				Limits:    updateServiceQuotaResource.Limits,
				QuotaUnit: updateServiceQuotaResource.QuotaUnit,
			},
			Scope: &pb.QuotaScope{
				ScopeType:  updateServiceQuotaResource.QuotaScope,
				ScopeValue: updateServiceQuotaResource.QuotaScopeValue,
			},
			RuleId:      updateServiceQuotaResource.RuleId,
			Reason:      updateServiceQuotaResource.Reason,
			CreatedTime: timestamppb.New(updateServiceQuotaResource.CreatedTimestamp),
			UpdatedTime: timestamppb.New(updateServiceQuotaResource.UpdateTimestamp),
		},
	}
	return updateQuotaServiceResponse, nil

}

func (s *QuotaManagementServiceClient) ListServiceQuota(ctx context.Context, req *pb.ListServiceQuotaRequest) (*pb.ListServiceQuotaResponse, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.ListServiceQuota").Start()
	defer span.End()
	logger.Info("entering list service quota")

	defer logger.Info("returning from list service quota resource")
	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database transaction begin failed:")
	}

	defer tx.Rollback()

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}

	serviceId := req.ServiceId
	registeredService, err := query.GetRegisteredService(ctx, tx, serviceId)
	if err != nil {
		if registeredService == nil {
			return &pb.ListServiceQuotaResponse{}, nil
		}
		logger.Error(err, "failed to get registered service")
		return nil, err
	}

	listServiceQuotasResponse := &pb.ListServiceQuotaResponse{
		ServiceId:                registeredService.ServiceId,
		ServiceName:              registeredService.ServiceName,
		ServiceQuotaAllResources: []*pb.ServiceQuotaResource{},
	}

	// get all quotas for service
	allQuotasForService, err := query.GetServiceResourceQuotas(ctx, tx, serviceId, "")
	if err != nil {
		logger.Error(err, "failed to get service resource quotas for service")
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	for _, currQuota := range allQuotasForService {
		qRes := &pb.ServiceQuotaResource{
			ResourceType: currQuota.ResourceName,
			QuotaConfig: &pb.QuotaConfig{
				Limits:    currQuota.Limits,
				QuotaUnit: currQuota.QuotaUnit,
			},
			Scope: &pb.QuotaScope{
				ScopeType:  currQuota.QuotaScope,
				ScopeValue: convertToUserFriendlyAccount(ctx, currQuota.QuotaScopeValue),
			},
			RuleId:      currQuota.RuleId,
			Reason:      currQuota.Reason,
			CreatedTime: timestamppb.New(currQuota.CreatedTimestamp),
			UpdatedTime: timestamppb.New(currQuota.UpdateTimestamp),
		}
		listServiceQuotasResponse.ServiceQuotaAllResources = append(listServiceQuotasResponse.ServiceQuotaAllResources, qRes)
	}
	logger.Info("retrieved quotas for service: ", "service in db : ", listServiceQuotasResponse)

	return listServiceQuotasResponse, nil
}

func (s *QuotaManagementServiceClient) ListAllServiceQuotas(ctx context.Context, req *pb.ListAllServiceQuotaRequest) (*pb.ListAllServiceQuotaResponse, error) {
	//Unimplemented for now
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.ListAllServiceQuotas").Start()
	defer span.End()
	logger.Info("entering list all service quotas")
	defer logger.Info("returning from list all service quotas")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}

	listAllServiceQuotasResponse := &pb.ListAllServiceQuotaResponse{
		AllServicesQuotaResponse: []*pb.ServiceQuotaResourceResponse{},
	}

	// get all quotas for service
	allQuotasForService, err := query.GetAllServiceResourceQuotas(ctx, tx)
	if err != nil {
		logger.Error(err, "failed to list all service quotas for services")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	for _, currQuota := range allQuotasForService {
		// TODO: consider avoiding this call, tradeoff is adding an extra column to the service_resource_quotas table
		registeredService, err := query.GetRegisteredService(ctx, tx, currQuota.ServiceId)
		if err != nil {
			logger.Error(err, "failed to get registered service")
			return nil, err
		}
		qSvc := &pb.ServiceQuotaResourceResponse{
			ServiceId:            currQuota.ServiceId,
			ServiceName:          registeredService.ServiceName,
			ServiceQuotaResource: []*pb.ServiceQuotaResource{},
		}
		qRes := &pb.ServiceQuotaResource{
			ResourceType: currQuota.ResourceName,
			QuotaConfig: &pb.QuotaConfig{
				Limits:    currQuota.Limits,
				QuotaUnit: currQuota.QuotaUnit,
			},
			Scope: &pb.QuotaScope{
				ScopeType:  currQuota.QuotaScope,
				ScopeValue: convertToUserFriendlyAccount(ctx, currQuota.QuotaScopeValue),
			},
			RuleId:      currQuota.RuleId,
			Reason:      currQuota.Reason,
			CreatedTime: timestamppb.New(currQuota.CreatedTimestamp),
			UpdatedTime: timestamppb.New(currQuota.UpdateTimestamp),
		}
		qSvc.ServiceQuotaResource = append(qSvc.ServiceQuotaResource, qRes)
		listAllServiceQuotasResponse.AllServicesQuotaResponse = append(listAllServiceQuotasResponse.AllServicesQuotaResponse, qSvc)
	}
	return listAllServiceQuotasResponse, nil
}

func (s *QuotaManagementServiceClient) GetResourceQuotaPrivate(ctx context.Context, req *pb.ServiceQuotaResourceRequestPrivate) (*pb.ServiceQuotasPrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.GetResourceQuotaPrivate").Start()
	defer span.End()
	logger.Info("entering get resource quota private")
	defer logger.Info("returning get resource quota private")

	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}

	if err := s.validateServiceRequest(ctx, tx, req); err != nil {
		return nil, err
	}

	logger.Info("get quota private params", "request details:", req)
	registeredService, err := query.GetRegisteredServiceByName(ctx, tx, req.ServiceName)
	if err != nil {
		logger.Error(err, "failed to get registered service")
		return nil, err
	}
	logger.Info("registered service retrieved by name: ", "service in db : ", registeredService.ServiceName)

	resourceType, cloudaccountId := req.ResourceType, req.CloudAccountId

	serviceQuotas, err := query.GetServiceResourceQuotasPrivate(ctx, tx, registeredService.ServiceId, registeredService.ServiceName, resourceType, cloudaccountId, s.cloudAccountServiceClient)
	if err != nil {
		return nil, err
	}
	logger.Info("quota private retrieved", "quota", serviceQuotas)
	return serviceQuotas, nil
}

func (s *QuotaManagementServiceClient) PingPrivate(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("QuotaManagementServicePrivate.Ping")
	logger.Info("entering quota management service private Ping")
	defer logger.Info("returning quota management service private Ping")

	return &emptypb.Empty{}, nil
}

func (s *QuotaManagementServiceClient) generateUniqueServiceId() string {
	//TODO: return a more user friendly name if required
	return uuid.NewString()
}

func (s *QuotaManagementServiceClient) validateServiceRequest(ctx context.Context, tx *sql.Tx, obj interface{}) error {
	if req, ok := obj.(*pb.ServiceQuotaRegistrationRequest); ok {
		if req.ServiceName == "" || req.Region == "" {
			return status.Errorf(codes.InvalidArgument, "service registration failed due to invalid request")
		}
		for _, resource := range req.ServiceResources {
			if resource.Name == "" || resource.QuotaUnit == "" || resource.MaxLimit < 0 {
				return status.Errorf(codes.InvalidArgument, "service registration failed due to invalid request")
			}
		}
		regService, err := query.GetRegisteredServiceByName(ctx, tx, req.ServiceName)
		if err == nil && regService != nil {
			return status.Errorf(codes.InvalidArgument, "service with name %s is already registered", req.ServiceName)
		}
	}
	if req, ok := obj.(*pb.UpdateServiceRegistrationRequest); ok {
		if req.ServiceId == "" || len(req.ServiceResources) == 0 {
			return status.Errorf(codes.InvalidArgument, "service registration update failed due to invalid request")
		}

		for _, resource := range req.ServiceResources {
			if resource.Name == "" || resource.MaxLimit < 0 {
				return status.Errorf(codes.InvalidArgument, "service registration update failed due to invalid request")
			}
		}
	}
	if req, ok := obj.(*pb.ServiceResourceRequest); ok {
		if req.ServiceId == "" || req.ResourceName == "" {
			return status.Errorf(codes.InvalidArgument, "service resource could not be fetched due to invalid request")
		}
	}
	if req, ok := obj.(*pb.CreateServiceQuotaRequest); ok {
		if req.ServiceId == "" || req.ServiceQuotaResource == nil || req.ServiceQuotaResource.Reason == "" {
			return status.Errorf(codes.InvalidArgument, "creation of service resource quota failed due to invalid request")
		}
	}
	if req, ok := obj.(*pb.DeleteServiceQuotaRequest); ok {
		if req.ServiceId == "" || req.ResourceType == "" {
			return status.Errorf(codes.InvalidArgument, "service resource quota could not be deleted due to invalid request")
		}
	}
	if req, ok := obj.(*pb.DeleteServiceRequest); ok {
		if req.ServiceId == "" {
			return status.Errorf(codes.InvalidArgument, "service could not be deleted due to invalid request")
		}
	}
	if req, ok := obj.(*pb.UpdateQuotaServiceRequest); ok {
		if req.ServiceId == "" || req.ResourceType == "" || req.RuleId == "" || req.QuotaConfig.Limits < 0 || req.Reason == "" {
			return status.Errorf(codes.InvalidArgument, "service resource quota could not be updated due to invalid request")
		}
	}
	if req, ok := obj.(*pb.ListServiceQuotaRequest); ok {
		if req.ServiceId == "" {
			return status.Errorf(codes.InvalidArgument, "service resource quotas could not be listed for the service due to invalid request")
		}
	}
	if req, ok := obj.(*pb.ServiceQuotaResourceRequestPrivate); ok {
		if req.ServiceName == "" {
			return status.Errorf(codes.InvalidArgument, "private service quota could not be fetched due to invalid request")
		}
	}
	return nil
}

//func (s *QuotaManagementServiceClient) getMinLimitsResource(allQuotasForService []*query.ServiceQuotaResource) map[string]int64 {
//	serviceResourceMinLimits := make(map[string]int64)
//	for _, quota := range allQuotasForService {
//		limit, ok := serviceResourceMinLimits[quota.ResourceName]
//		if !ok {
//			serviceResourceMinLimits[quota.ResourceName] = quota.Limits
//		} else {
//			if limit < quota.Limits {
//				serviceResourceMinLimits[quota.ResourceName] = limit
//			}
//		}
//	}
//	return serviceResourceMinLimits
//}

func convertToUserFriendlyAccount(ctx context.Context, accountType string) string {
	logger := log.FromContext(ctx).WithName("convertToUserFriendly")
	logger.Info("entering convert to user friendly account", "acc type", accountType)
	userFriendlyAccount, ok := accountTypeMap[accountType]
	if ok {
		logger.Info("found  user friendly account", "user friendly acc type", userFriendlyAccount)
		return userFriendlyAccount
	}
	logger.Info("not found  user friendly account", "returning acc type", accountType)
	return accountType
}
