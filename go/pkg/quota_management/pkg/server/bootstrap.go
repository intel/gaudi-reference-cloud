// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quota_management/database/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QuotaManagementBootstrapper struct {
	QWotaManagementServiceClient pb.QuotaManagementServiceClient `json:"quotaManagementServiceClient" yaml:"quotaManagementServiceClient"`
}

type BootstrappedService struct {
	QuotaUnit        string                             `json:"quotaUnit" yaml:"quotaUnit"`
	MaxLimits        map[string]map[string]int          `json:"maxLimits" yaml:"maxLimits"`
	QuotaAccountType map[string]QuotaAccountTypeDetails `json:"quotaAccountType" yaml:"quotaAccountType"`
}

type QuotaAccountTypeDetails struct {
	DefaultLimits map[string]map[string]int `json:"defaultLimits" yaml:"defaultLimits"`
}

func (s *QuotaManagementServiceClient) BootstrapQuotaManagementService(ctx context.Context, bootstrappedConfig []*BootstrappedService, region string) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.registerService").Start()
	defer span.End()
	logger.Info("entering bootstrapping of quota management service")
	defer logger.Info("returning from  bootstrapping of quota management service")

	// TODO: remove before merge
	//prettyPrintBootstrapRegistrations(ctx, bootstrappedConfig, region)

	err := bootstrapServicesCreateDefaultQuotas(ctx, bootstrappedConfig, s, region)
	if err != nil {
		return err
	}
	return nil
}

func bootstrapServicesCreateDefaultQuotas(ctx context.Context, bootstrapServices []*BootstrappedService, s *QuotaManagementServiceClient, region string) error {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.registerService").Start()
	defer span.End()
	for _, currentUnit := range bootstrapServices {
		unit := currentUnit.QuotaUnit
		// register services and max limits
		for serviceName, services := range currentUnit.MaxLimits {
			serviceExists, serviceId, err := s.serviceExists(ctx, serviceName)
			if err == nil {
				if !serviceExists {
					logger.Info("service is not registered", "creating new service", serviceName)
					regRequest := convertToBootstrappedServicesRequest(serviceName, region, unit, services)
					regResp, err := s.Register(ctx, regRequest)
					if err != nil {
						return err
					}
					logger.Info("Service registered successfully", "service name", regResp.ServiceName)
				} else {
					// service exists, only register the new resources leaving the existing ones untouched
					logger.Info("registered service exists will only add new service resources", "service", serviceName)
					regRequest := s.convertToBootstrappedServicesUpdateRequest(ctx, serviceId, unit, services)
					if len(regRequest.ServiceResources) > 0 {
						regResp, err := s.UpdateServiceRegistration(ctx, regRequest)
						if err != nil {
							logger.Error(err, "service update error in bootstrap")
							return err
						}
						logger.Info("registered service updated", "service name", regResp.ServiceName)
					} else {
						logger.Info("no new service resources found to update", "service name", serviceName)
					}
				}
			}
		}
		for accType, accDetails := range currentUnit.QuotaAccountType {
			// create default service resource quotas
			defaultLimits := accDetails.DefaultLimits
			for serviceName, services := range defaultLimits {
				serviceExists, serviceId, err := s.serviceExists(ctx, serviceName)
				if err == nil {
					if !serviceExists {
						logger.Info("service is not registered", "cannot create quotas for unregistered service", serviceName)
						return status.Error(codes.NotFound, "cannot create quotas for unregistered service")
					} else {
						// service exists, can create quota for this service
						for resource, limit := range services {
							serviceQuotaExists, existingQuota, err := s.serviceQuotaResourceExists(ctx, serviceId, resource, pb.QuotaScopeType_name[0], accType)
							if err == nil {
								if serviceQuotaExists {
									logger.Info("service resource quota exists and will be left untouched", "service name", serviceName, "resource type", resource, "rule ID", existingQuota.RuleId)
								} else {
									quotaRequest := convertToServiceResourceQuotaRequest(serviceId, resource, accType, unit, limit)
									quotaResp, err := s.CreateServiceQuota(ctx, quotaRequest)
									if err != nil {
										if maxLimitError(err) {
											logger.Error(err, "max limit conflict during bootstrapping, skipping create of", "service", serviceName, "resource", resource)
										} else {
											return err
										}
									} else {
										logger.Info("Service resource quota created successfully", "service name", quotaResp.ServiceName, "resource type", quotaResp.ServiceQuotaResource.ResourceType)
									}
								}
							}
						}
					}
				} else {
					return err
				}
			}
		}
	}
	return nil
}

func convertToBootstrappedServicesRequest(serviceName, region, unit string, resourceLimits map[string]int) *pb.ServiceQuotaRegistrationRequest {
	var currResourceList []*pb.ServiceResource
	serviceRegistration := &pb.ServiceQuotaRegistrationRequest{
		ServiceName: serviceName,
		Region:      region,
	}
	for resource, limit := range resourceLimits {
		res := &pb.ServiceResource{
			Name:      resource,
			QuotaUnit: unit,
			MaxLimit:  int64(limit),
		}
		currResourceList = append(currResourceList, res)
	}
	serviceRegistration.ServiceResources = currResourceList
	return serviceRegistration
}

func (s *QuotaManagementServiceClient) convertToBootstrappedServicesUpdateRequest(ctx context.Context, serviceId, unit string, resourceLimits map[string]int) *pb.UpdateServiceRegistrationRequest {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.convertToBootstrappedServicesUpdateRequest").Start()
	defer span.End()
	var currResourceList []*pb.ServiceResource
	serviceUpdate := &pb.UpdateServiceRegistrationRequest{
		ServiceId: serviceId,
	}
	for resource, limit := range resourceLimits {
		serviceResourceExists, err := s.serviceResourceExists(ctx, serviceId, resource)
		if err == nil && !serviceResourceExists {
			logger.Info("found new service resource to be added", "resource", resource, "limit", limit)
			res := &pb.ServiceResource{
				Name:      resource,
				QuotaUnit: unit,
				MaxLimit:  int64(limit),
			}
			currResourceList = append(currResourceList, res)
		} else {
			logger.Info("service resource already exists", "resource", resource, "limit", limit)
		}
	}
	serviceUpdate.ServiceResources = currResourceList
	return serviceUpdate
}

func convertToServiceResourceQuotaRequest(serviceId, resource, accType, unit string, limit int) *pb.CreateServiceQuotaRequest {
	createServiceQuotaRequest := &pb.CreateServiceQuotaRequest{
		ServiceId: serviceId,
		ServiceQuotaResource: &pb.ServiceQuotaResource{
			ResourceType: resource,
			QuotaConfig: &pb.QuotaConfig{
				QuotaUnit: unit,
				Limits:    int64(limit),
			},
			Scope: &pb.QuotaScope{
				ScopeType:  pb.QuotaScopeType_name[0],
				ScopeValue: accType,
			},
			Reason: "default quota created during bootstrap", // TODO write function to get appropriate reason based on service name and resource type
		},
	}
	return createServiceQuotaRequest
}

func convertToServiceResourceQuotaUpdateRequest(serviceId, resourceType, ruleId, unit string, limit int) *pb.UpdateQuotaServiceRequest {
	updateServiceQuotaRequest := &pb.UpdateQuotaServiceRequest{
		ServiceId:    serviceId,
		RuleId:       ruleId,
		ResourceType: resourceType,
		QuotaConfig: &pb.QuotaConfig{
			QuotaUnit: unit,
			Limits:    int64(limit),
		},
		Reason: "default quota created during bootstrap", // TODO write function to get appropriate reason based on service name and resource type
	}
	return updateServiceQuotaRequest
}

func (s *QuotaManagementServiceClient) serviceExists(ctx context.Context, serviceName string) (bool, string, error) {
	_, logger, _ := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.serviceExists").Start()
	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "service exists error")
		return false, "", status.Error(codes.Internal, "database transaction failed")
	}

	defer tx.Rollback()
	registeredService, err := query.GetRegisteredServiceByName(ctx, tx, serviceName)
	if err != nil {
		if notFoundError(err) {
			return false, "", nil
		} else {
			return false, "", err
		}
	}
	return true, registeredService.ServiceId, nil
}

func (s *QuotaManagementServiceClient) serviceQuotaResourceExists(ctx context.Context, serviceId, resourceName, scope, scopeValue string) (bool, *query.ServiceQuotaResource, error) {
	_, logger, _ := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.serviceQuotaResourceExists").Start()
	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "service quota exists")
		return false, nil, status.Error(codes.Internal, "database transaction failed")
	}

	defer tx.Rollback()
	resourceQuotas, err := query.GetServiceResourceQuotas(ctx, tx, serviceId, resourceName)
	if err != nil {
		if notFoundError(err) {
			return false, nil, nil
		} else {
			return false, nil, err
		}
	}
	// quotas found for resource
	for _, res := range resourceQuotas {
		if res.ResourceName == resourceName && res.QuotaScope == scope && res.QuotaScopeValue == scopeValue {
			// quota found for given service/resource type/quotscope combo, so update it
			logger.Info("found existing service resource quota", "resource name", res.ResourceName, "scope type", res.QuotaScope, "scope value", res.QuotaScopeValue)
			return true, res, nil
		}
	}
	return false, nil, nil
}

func (s *QuotaManagementServiceClient) serviceResourceExists(ctx context.Context, serviceId, resourceName string) (bool, error) {
	_, logger, _ := obs.LogAndSpanFromContext(ctx).WithName("QuotaManagementService.serviceResourceExists").Start()
	tx, err := s.session.BeginTx(ctx, nil)
	if err != nil {
		logger.Info("service resource exists ", "error", err)
		return false, status.Error(codes.Internal, "database transaction failed")
	}

	defer tx.Rollback()
	_, err = query.GetServiceResource(ctx, tx, serviceId, resourceName)
	if err != nil {
		if notFoundError(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func notFoundError(err error) bool {
	if st, ok := status.FromError(err); ok {
		return st.Code() == codes.NotFound
	}
	return false
}

func maxLimitError(err error) bool {
	if st, ok := status.FromError(err); ok {
		return (st.Code() == codes.InvalidArgument && st.Message() == "not allowed to create/update resource limit to more than that specified in service registration")
	}
	return false
}
