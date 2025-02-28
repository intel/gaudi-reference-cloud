// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type QuotaUnitType string

const (
	QuotaCount   QuotaUnitType = "COUNT"
	QuotaReqSec  QuotaUnitType = "REQ_SEC"
	QuotaReqMin  QuotaUnitType = "REQ_MIN"
	QuotaReqHour QuotaUnitType = "REQ_HOUR"
)

type ScopeType string

const (
	ScopeAccountId   ScopeType = "QUOTA_ACCOUNT_ID"
	ScopeAccountType ScopeType = "QUOTA_ACCOUNT_TYPE"
)

type RegisteredService struct {
	ServiceId   string
	ServiceName string
	Region      string
	Resources   ServiceResource
}

type ServiceResource struct {
	ServiceId    string
	ResourceName string
	QuotaUnit    string
	MaxLimit     int64
}

type ServiceQuotaResource struct {
	ServiceId        string
	ResourceName     string
	RuleId           string
	QuotaUnit        string
	Limits           int64
	QuotaScope       string
	QuotaScopeValue  string
	Reason           string
	CreatedTimestamp time.Time
	UpdateTimestamp  time.Time
}

const (
	insertService = `
		INSERT INTO registered_services (service_id, service_name, region)
		VALUES ($1, $2, $3)
		RETURNING service_id, service_name, region
	`
	getService = `
		SELECT service_id, service_name, region
		FROM registered_services
		WHERE service_id = $1
	`
	getServiceByName = `
		SELECT service_id, service_name, region
		FROM registered_services
		WHERE service_name = $1
	`
	getAllServices = `
		SELECT service_id, service_name
		FROM registered_services
	`
	insertServiceResource = `
		INSERT INTO service_resources (service_id, resource_name, quota_unit, max_limit)
		VALUES ($1, $2, $3, $4)
		RETURNING service_id, resource_name, quota_unit, max_limit
	`
	getServiceResource = `
		SELECT service_id, resource_name, quota_unit, max_limit
		FROM service_resources
		WHERE service_id = $1 AND resource_name = $2
	`
	getServiceResourcesAll = `
		SELECT resource_name, quota_unit, max_limit
		FROM service_resources
		WHERE service_id = $1
	`
	updateServiceResource = `
		UPDATE service_resources 
		SET max_limit = $3
		WHERE service_id = $1 AND resource_name = $2
		RETURNING service_id, resource_name, quota_unit, max_limit
	`
	deleteService = `
		DELETE FROM registered_services 
		WHERE service_id = $1
	`
	insertServiceResourceQuota = `
		INSERT INTO service_resource_quotas 
		(service_id, resource_name, rule_id, limits, quota_unit,scope_type,scope_value, reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING service_id, resource_name, rule_id, limits, quota_unit,scope_type,scope_value, reason, created_timestamp, updated_timestamp
	`
	getServiceResourceQuota = `
		SELECT service_id, resource_name, rule_id, limits, quota_unit, scope_type, scope_value, reason, created_timestamp, updated_timestamp
		FROM service_resource_quotas
		WHERE service_id = $1 AND resource_name = $2
	`
	deleteAllServiceResources = `
		DELETE FROM service_resources 
		WHERE service_id = $1
	`
	getServiceResourceAllQuotas = `
		SELECT service_id, resource_name, rule_id, limits, quota_unit, scope_type,scope_value, reason, created_timestamp, updated_timestamp
		FROM service_resource_quotas
		WHERE service_id = $1 
	`
	updateServiceResourceQuota = `
		UPDATE service_resource_quotas
		SET limits = $4, reason = $5, updated_timestamp = $6
		WHERE service_id = $1 AND resource_name = $2 AND rule_id = $3
		RETURNING service_id, resource_name, rule_id, limits, quota_unit,scope_type,scope_value, reason, created_timestamp, updated_timestamp
	`
	deleteServiceResourceQuota = `
		DELETE FROM service_resource_quotas 
		WHERE service_id = $1 AND resource_name = $2 AND rule_id = $3
	`
	deleteAllServiceResourceQuotas = `
		DELETE FROM service_resource_quotas 
		WHERE service_id = $1 
	`
	getAllServiceResourceAllQuotas = `
		SELECT service_id, resource_name, rule_id, limits, quota_unit,scope_type,scope_value, reason, created_timestamp, updated_timestamp
		FROM service_resource_quotas
	`
	getServiceResourceQuotaPrivate = `
		SELECT resource_name, rule_id, limits, quota_unit, scope_type,scope_value, reason, created_timestamp, updated_timestamp
		FROM service_resource_quotas	
	`
	getServiceResourceQuotaForScope = `
		SELECT resource_name, rule_id, limits, quota_unit, scope_type,scope_value, reason, created_timestamp, updated_timestamp
		FROM service_resource_quotas 
		WHERE service_id = $1 AND resource_name = $2 and scope_value = $3
	`
)

func InsertService(ctx context.Context, tx *sql.Tx, serviceId, serviceName, region string) (*RegisteredService, error) {
	logger := log.FromContext(ctx).WithName("InsertService")
	logger.Info("insert new service quota management into db")

	var service RegisteredService
	err := tx.QueryRowContext(
		ctx,
		insertService,
		serviceId,
		serviceName,
		region,
	).Scan(
		&service.ServiceId,
		&service.ServiceName,
		&service.Region,
	)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func GetRegisteredService(ctx context.Context, tx *sql.Tx, serviceId string) (*RegisteredService, error) {
	logger := log.FromContext(ctx).WithName("GetRegisteredService")
	logger.Info("get registered service")

	var service RegisteredService
	err := tx.QueryRowContext(
		ctx,
		getService,
		serviceId,
	).Scan(
		&service.ServiceId,
		&service.ServiceName,
		&service.Region,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "registered service not found")
		}
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	return &service, nil
}

func GetRegisteredServiceByName(ctx context.Context, tx *sql.Tx, serviceName string) (*RegisteredService, error) {
	logger := log.FromContext(ctx).WithName("GetRegisteredServiceByName")
	logger.Info("get registered service by name")

	var service RegisteredService
	err := tx.QueryRowContext(
		ctx,
		getServiceByName,
		serviceName,
	).Scan(
		&service.ServiceId,
		&service.ServiceName,
		&service.Region,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "registered service not found for given name")
		}
		logger.Info("RK=>", "get service by name error", err)
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	return &service, nil
}

func ListRegisteredServices(ctx context.Context, tx *sql.Tx) ([]*pb.ServiceDetail, error) {
	logger := log.FromContext(ctx).WithName("ListRegisteredServices")
	logger.Info("list all registered services")

	var servicesList []*pb.ServiceDetail
	rows, err := tx.QueryContext(
		ctx,
		getAllServices,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	defer rows.Close()

	for rows.Next() {
		var service pb.ServiceDetail
		err := rows.Scan(
			&service.ServiceId,
			&service.ServiceName,
		)
		if err != nil {
			return nil, err
		}
		servicesList = append(servicesList, &service)

	}

	return servicesList, nil
}

func ListServiceResources(ctx context.Context, tx *sql.Tx, serviceId string) ([]*pb.ServiceResource, error) {
	logger := log.FromContext(ctx).WithName("ListServiceResources")
	logger.Info("list all services resources from db")

	var serviceResources []*pb.ServiceResource
	rows, err := tx.QueryContext(
		ctx,
		getServiceResourcesAll,
		serviceId,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error(err, "error getting service resources")
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	defer rows.Close()

	for rows.Next() {
		var serviceResource pb.ServiceResource
		err := rows.Scan(
			&serviceResource.Name,
			&serviceResource.QuotaUnit,
			&serviceResource.MaxLimit,
		)
		if err != nil {
			return nil, err
		}
		serviceResources = append(serviceResources, &serviceResource)

	}

	return serviceResources, nil
}

func InsertServiceResource(ctx context.Context, tx *sql.Tx, serviceId, resourceName string, quotaUnit string, maxLimit int64) (*pb.ServiceResource, error) {
	logger := log.FromContext(ctx).WithName("InsertServiceResource")
	logger.Info("create new resource for a service, for quota management")

	var serviceResource pb.ServiceResource
	err := tx.QueryRowContext(
		ctx,
		insertServiceResource,
		serviceId,
		resourceName,
		quotaUnit,
		maxLimit,
	).Scan(
		&serviceId,
		&serviceResource.Name,
		&serviceResource.QuotaUnit,
		&serviceResource.MaxLimit,
	)

	if err != nil {
		return nil, err
	}
	return &serviceResource, nil
}

func GetServiceResource(ctx context.Context, tx *sql.Tx, serviceId, resourceType string) (*ServiceResource, error) {
	logger := log.FromContext(ctx).WithName("GetServiceResource")
	logger.Info("get resource for a service, for quota management")

	var serviceResource ServiceResource
	err := tx.QueryRowContext(
		ctx,
		getServiceResource,
		serviceId,
		resourceType,
	).Scan(
		&serviceResource.ServiceId,
		&serviceResource.ResourceName,
		&serviceResource.QuotaUnit,
		&serviceResource.MaxLimit,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "service resource not found")
		}
		return nil, status.Error(codes.Internal, "database transaction failed")
	}
	return &serviceResource, nil
}

func UpdateServiceResource(ctx context.Context, tx *sql.Tx, serviceId, resourceName string, maxLimit int64) (*pb.ServiceResource, error) {
	logger := log.FromContext(ctx).WithName("UpdateServiceResource")
	logger.Info("updating a service resource")

	var serviceResource pb.ServiceResource
	err := tx.QueryRowContext(
		ctx,
		updateServiceResource,
		serviceId,
		resourceName,
		maxLimit,
	).Scan(
		&serviceId,
		&serviceResource.Name,
		&serviceResource.QuotaUnit,
		&serviceResource.MaxLimit,
	)

	if err != nil {
		return nil, err
	}
	return &serviceResource, nil
}

func DeleteService(ctx context.Context, tx *sql.Tx, serviceId string) error {
	logger := log.FromContext(ctx).WithName("DeleteService")
	logger.Info("delete a registered service, for quota management")
	rows, err := tx.ExecContext(ctx, deleteService, serviceId)
	if err != nil {
		return err
	}

	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		logger.Error(err, "error for RowsAffected, in DeleteSevice")
		return status.Errorf(codes.Internal, "%v", err)
	}

	if rowsAffected == 0 {
		return status.Errorf(codes.NotFound, "no services found to delete")
	}
	return nil
}

func DeleteAllServiceResources(ctx context.Context, tx *sql.Tx, serviceId string) error {
	logger := log.FromContext(ctx).WithName("DeleteAllServiceResources")
	logger.Info("delete all services resource for a service")
	rows, err := tx.ExecContext(ctx, deleteAllServiceResources, serviceId)
	if err != nil {
		return err
	}

	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		logger.Error(err, "error for RowsAffected, in DeleteSevice")
		return status.Errorf(codes.Internal, "%v", err)
	}

	if rowsAffected == 0 {
		return status.Errorf(codes.NotFound, "no services resources found to delete")
	}
	return nil
}

func GetServiceResourceQuotas(ctx context.Context, tx *sql.Tx, serviceId, resourceName string) ([]*ServiceQuotaResource, error) {
	logger := log.FromContext(ctx).WithName("GetServiceResourceQuotas")
	logger.Info("get resource quotas for a service, for quota management")

	var serviceQuotaResources []*ServiceQuotaResource
	var quotaQuery string
	var rows *sql.Rows
	var err error
	if resourceName == "" {
		quotaQuery = getServiceResourceAllQuotas
		rows, err = tx.QueryContext(
			ctx,
			quotaQuery,
			serviceId,
		)

	} else {
		quotaQuery = getServiceResourceQuota
		rows, err = tx.QueryContext(
			ctx,
			quotaQuery,
			serviceId,
			resourceName,
		)
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var serviceQuotaResource ServiceQuotaResource
		err := rows.Scan(
			&serviceQuotaResource.ServiceId,
			&serviceQuotaResource.ResourceName,
			&serviceQuotaResource.RuleId,
			&serviceQuotaResource.Limits,
			&serviceQuotaResource.QuotaUnit,
			&serviceQuotaResource.QuotaScope,
			&serviceQuotaResource.QuotaScopeValue,
			&serviceQuotaResource.Reason,
			&serviceQuotaResource.CreatedTimestamp,
			&serviceQuotaResource.UpdateTimestamp,
		)
		if err != nil {
			return nil, err
		}
		serviceQuotaResources = append(serviceQuotaResources, &serviceQuotaResource)
	}

	return serviceQuotaResources, nil
}

func InsertServiceResourceQuota(ctx context.Context, tx *sql.Tx, serviceId, resourceName, ruleId string, limits int64,
	quotaUnit, scopeType, scopeValue, reason string) (*ServiceQuotaResource, error) {
	logger := log.FromContext(ctx).WithName("InsertServiceResourceQuota")
	logger.Info("create new resource for a service, for quota management")

	var serviceQuotaResource ServiceQuotaResource
	err := tx.QueryRowContext(
		ctx,
		insertServiceResourceQuota,
		serviceId,
		resourceName,
		ruleId,
		limits,
		quotaUnit,
		scopeType,
		scopeValue,
		reason,
	).Scan(
		&serviceQuotaResource.ServiceId,
		&serviceQuotaResource.ResourceName,
		&serviceQuotaResource.RuleId,
		&serviceQuotaResource.Limits,
		&serviceQuotaResource.QuotaUnit,
		&serviceQuotaResource.QuotaScope,
		&serviceQuotaResource.QuotaScopeValue,
		&serviceQuotaResource.Reason,
		&serviceQuotaResource.CreatedTimestamp,
		&serviceQuotaResource.UpdateTimestamp,
	)

	if err != nil {
		return nil, err
	}
	return &serviceQuotaResource, nil
}

func UpdateServiceResourceQuota(ctx context.Context, tx *sql.Tx, serviceId, resourceName, ruleId, reason string, limits int64) (*ServiceQuotaResource, error) {
	logger := log.FromContext(ctx).WithName("UpdateServiceResourceQuota")
	logger.Info("update service resource quota, for quota management")

	var serviceQuotaResource ServiceQuotaResource
	err := tx.QueryRowContext(
		ctx,
		updateServiceResourceQuota,
		serviceId,
		resourceName,
		ruleId,
		limits,
		reason,
		time.Now(),
	).Scan(
		&serviceQuotaResource.ServiceId,
		&serviceQuotaResource.ResourceName,
		&serviceQuotaResource.RuleId,
		&serviceQuotaResource.Limits,
		&serviceQuotaResource.QuotaUnit,
		&serviceQuotaResource.QuotaScope,
		&serviceQuotaResource.QuotaScopeValue,
		&serviceQuotaResource.Reason,
		&serviceQuotaResource.CreatedTimestamp,
		&serviceQuotaResource.UpdateTimestamp,
	)

	if err != nil {
		return nil, err
	}
	return &serviceQuotaResource, nil
}

func DeleteServiceResourceQuota(ctx context.Context, tx *sql.Tx, serviceId, resourceName, ruleId string) error {
	logger := log.FromContext(ctx).WithName("CreateServiceResource")
	logger.Info("create new resource for a service, for quota management")

	//rows, err := tx.QueryContext(ctx, deleteServiceResourceQuota, service_id, resource_name)
	rows, err := tx.ExecContext(ctx, deleteServiceResourceQuota, serviceId, resourceName, ruleId)
	if err != nil {
		return err
	}

	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		logger.Error(err, "error for RowsAffected, in DeleteSeviceResourceQuota")
		return status.Errorf(codes.Internal, "%v", err)
	}

	if rowsAffected == 0 {
		return status.Errorf(codes.NotFound, "no services resource quotas found to delete")
	}
	return nil
}

func DeleteAllServiceResourceQuota(ctx context.Context, tx *sql.Tx, serviceId string) error {
	logger := log.FromContext(ctx).WithName("DeleteAllServiceResourceQuota")
	logger.Info("delete all resource quotas for a service")

	rows, err := tx.ExecContext(ctx, deleteAllServiceResourceQuotas, serviceId)
	if err != nil {
		return err
	}

	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		logger.Error(err, "error for RowsAffected, in DeleteSeviceResourceQuota")
		return status.Errorf(codes.Internal, "%v", err)
	}

	if rowsAffected == 0 {
		return status.Errorf(codes.NotFound, "no services resource quotas found to delete")
	}
	return nil
}

func GetAllServiceResourceQuotas(ctx context.Context, tx *sql.Tx) ([]*ServiceQuotaResource, error) {
	logger := log.FromContext(ctx).WithName("GetServiceResourceQuotas")
	logger.Info("get resource quotas for a service, for quota management")

	var serviceQuotaResources []*ServiceQuotaResource
	rows, err := tx.QueryContext(
		ctx,
		getAllServiceResourceAllQuotas,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var serviceQuotaResource ServiceQuotaResource
		err := rows.Scan(
			&serviceQuotaResource.ServiceId,
			&serviceQuotaResource.ResourceName,
			&serviceQuotaResource.RuleId,
			&serviceQuotaResource.Limits,
			&serviceQuotaResource.QuotaUnit,
			&serviceQuotaResource.QuotaScope,
			&serviceQuotaResource.QuotaScopeValue,
			&serviceQuotaResource.Reason,
			&serviceQuotaResource.CreatedTimestamp,
			&serviceQuotaResource.UpdateTimestamp,
		)
		if err != nil {
			return nil, err
		}
		serviceQuotaResources = append(serviceQuotaResources, &serviceQuotaResource)
	}

	return serviceQuotaResources, nil
}

func GetServiceResourceQuotasPrivate(ctx context.Context, tx *sql.Tx, serviceId, serviceName, resourceName, cloudAccountId string, cloudAccountSvc pb.CloudAccountServiceClient) (*pb.ServiceQuotasPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetServiceResourceQuotasPrivate")
	logger.Info("get resource quotas private")

	var queryBuilder strings.Builder
	var serviceQuotaResourcesCustom, serviceQuotaResourcesAccountType []*pb.ServiceQuotaResource
	var err error
	foundRows := false

	queryBuilder, params, err := getQueryForResourceQuota(ctx, serviceId, resourceName, cloudAccountId, "")
	if err != nil {
		logger.Error(err, "failed to build query for get quotas")
		return nil, err
	}
	serviceQuotaResourcesCustom, foundRows, err = getServiceQuotasFromDb(ctx, tx, queryBuilder.String(), params)
	if err != nil {
		logger.Error(err, "failed to query for resource quota using resource type and/or cloudaccount id")
		return nil, status.Errorf(codes.Internal, "quota look up failed")
	}

	// no rows found using the given search params but cloudaccount ID has been provided
	if !foundRows && cloudAccountId != "" {
		// retrieve the account type (STANDARD, PREMIUM, INTEL etc) if cloud account is valid
		cloudAccount, valid := getValidCloudAccount(ctx, cloudAccountId, cloudAccountSvc)
		if valid {
			accountType := cloudAccount.GetType()
			logger.Info("looking up quota using account type", "cloud account type", accountType, "resource type", resourceName)
			queryBuilderAccType, params, err := getQueryForResourceQuota(ctx, serviceId, resourceName, "", accountType.String())
			if err != nil {
				logger.Error(err, "failed to build query for get quotas")
				return nil, err
			}

			serviceQuotaResourcesAccountType, foundRows, err = getServiceQuotasFromDb(ctx, tx, queryBuilderAccType.String(), params)
			if err != nil {
				logger.Error(err, "failed to query for resource quota")
				return nil, status.Errorf(codes.Internal, "quota look up failed")
			}
		} else {
			return nil, status.Errorf(codes.NotFound, "failed to query for resource quota because of invalid cloudaccount id")
		}
	} else if !foundRows && cloudAccountId == "" {
		return nil, status.Errorf(codes.NotFound, "quota look up failed with cloudaccount ID not provided")
	} else if foundRows && cloudAccountId != "" {
		// custom quotas already found, if cloudaccount was supplied also try getting quota based on it's type
		cloudAccount, valid := getValidCloudAccount(ctx, cloudAccountId, cloudAccountSvc)
		if valid {
			accountType := cloudAccount.GetType()
			logger.Info("looking up quota using account type", "cloud account type", accountType, "resource type", resourceName)
			queryBuilderAccType, params, err := getQueryForResourceQuota(ctx, serviceId, resourceName, "", accountType.String())
			if err != nil {
				logger.Error(err, "failed to build query for get quotas")
				return nil, err
			}

			serviceQuotaResourcesAccountType, foundRows, err = getServiceQuotasFromDb(ctx, tx, queryBuilderAccType.String(), params)
			if err != nil {
				logger.Error(err, "failed to query for resource quota")
				return nil, status.Errorf(codes.Internal, "quota look up failed")
			}
		}
	} else if foundRows && cloudAccountId == "" {
		//already found quotas that user requested w/o cloudaccount ID, assign same to default type
		serviceQuotaResourcesAccountType = serviceQuotaResourcesCustom
	}

	// if no quotas are found using cloudaccount ID or cloudaccount type
	if !foundRows {
		return nil, status.Errorf(codes.Internal, "quota look up failed with cloudaccount ID and/or cloudaccount type")
	}

	return &pb.ServiceQuotasPrivate{
		CustomQuota: &pb.ServiceQuotaPrivate{
			ServiceName:      serviceName,
			ServiceResources: serviceQuotaResourcesCustom,
		},
		DefaultQuota: &pb.ServiceQuotaPrivate{
			ServiceName:      serviceName,
			ServiceResources: serviceQuotaResourcesAccountType,
		}}, nil
}

func getQueryForResourceQuota(ctx context.Context, serviceId, resourceName, cloudAccountId, cloudAccountype string) (strings.Builder, []interface{}, error) {
	logger := log.FromContext(ctx).WithName("GetServiceResourceQuotasPrivate")
	var queryBuilder strings.Builder
	_, err := queryBuilder.WriteString(getServiceResourceQuotaPrivate)
	if err != nil {
		logger.Error(err, "failed to build query for get quotas")
		return strings.Builder{}, nil, err
	}

	if _, err := queryBuilder.WriteString(" WHERE 1=1"); err != nil {
		logger.Error(err, "failed to build query for get quotas")
		return strings.Builder{}, nil, err
	}

	params := []interface{}{}
	if serviceId != "" {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND service_id=$%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append service id to query to get quotas")
			return strings.Builder{}, nil, err
		}
		params = append(params, serviceId)
	}
	if resourceName != "" {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND resource_name = $%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append resource type to query to get quotas")
			return strings.Builder{}, nil, err
		}
		params = append(params, resourceName)
	}
	// cloud account is passed so checking it's validity before proceeding
	if cloudAccountId != "" {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND scope_value = $%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append scope value to query to get quotas")
			return strings.Builder{}, nil, err
		}
		params = append(params, cloudAccountId)
	}

	if cloudAccountype != "" {
		_, err := queryBuilder.WriteString(fmt.Sprintf(" AND scope_value = $%d ", len(params)+1))
		if err != nil {
			logger.Error(err, "failed to append resource type to query to get quotas")
			return strings.Builder{}, nil, err
		}
		params = append(params, cloudAccountype)
	}
	return queryBuilder, params, nil

}

func getServiceQuotasFromDb(ctx context.Context, tx *sql.Tx, query string, params []interface{}) ([]*pb.ServiceQuotaResource, bool, error) {
	logger := log.FromContext(ctx).WithName("getServiceQuotasFromDb")
	foundRows := false
	var serviceQuotaResourcesPrivate []*pb.ServiceQuotaResource

	rows, err := tx.QueryContext(ctx, query, params...)
	if err != nil {
		logger.Error(err, "failed to query for resource quota")
		return nil, false, status.Errorf(codes.Internal, "quota look up failed")
	}

	defer rows.Close()
	for rows.Next() {
		foundRows = true
		serviceQuotaResourcePrivate := pb.ServiceQuotaResource{QuotaConfig: &pb.QuotaConfig{}, Scope: &pb.QuotaScope{}}
		var createdTime, updatedTime time.Time
		err := rows.Scan(
			&serviceQuotaResourcePrivate.ResourceType,
			&serviceQuotaResourcePrivate.RuleId,
			&serviceQuotaResourcePrivate.QuotaConfig.Limits,
			&serviceQuotaResourcePrivate.QuotaConfig.QuotaUnit,
			&serviceQuotaResourcePrivate.Scope.ScopeType,
			&serviceQuotaResourcePrivate.Scope.ScopeValue,
			&serviceQuotaResourcePrivate.Reason,
			&createdTime,
			&updatedTime,
		)
		if err != nil {
			logger.Error(err, "failed to get resource quota error")
			return nil, false, status.Errorf(codes.Internal, "resource quota look up failed")
		}
		serviceQuotaResourcePrivate.CreatedTime = timestamppb.New(createdTime)
		serviceQuotaResourcePrivate.UpdatedTime = timestamppb.New(updatedTime)
		serviceQuotaResourcesPrivate = append(serviceQuotaResourcesPrivate, &serviceQuotaResourcePrivate)
	}
	return serviceQuotaResourcesPrivate, foundRows, nil

}

func getValidCloudAccount(ctx context.Context, cloudAccountId string, cloudAccountServiceClient pb.CloudAccountServiceClient) (*pb.CloudAccount, bool) {
	logger := log.FromContext(ctx).WithName("GetServiceResourceQuotasPrivate")
	cloudAccount, err := cloudAccountServiceClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil || cloudAccount.Id != cloudAccountId {
		logger.Info("failed to get cloudaccount")
		return nil, false
	}
	return cloudAccount, true
}
