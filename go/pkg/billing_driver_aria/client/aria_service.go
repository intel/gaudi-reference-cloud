// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"encoding/json"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"

	"gopkg.in/yaml.v3"
)

// The gl_cd is string representation of number.
var ariaServiceRequestDefaults = []byte(`
output_format: json
service_type: Usage-Based
gl_cd: '1'
taxable_ind: 1
client_tax_group_id: SERVICES
allow_service_credits: Y
tax_group: 7
`)

// AriaServiceClient to store the defaults for creation and update of services and make Aria calls for service
// Adding the qualifier of Aria to make sure it is not confused with service :-) For other Aria calls, no qualifier
// should be needed.
// Need to make both calls - Admin using query string for CREATE, UPDATE, DELETE and non admin for GETs.
// Todo: Requested Aria team to start supporting JSON for admin APIs and when supported query strings will not be needed.
type AriaServiceClient struct {
	ariaServiceRequestJson []byte
	ariaAdminClient        *AriaAdminClient
	ariaCredentials        *AriaCredentials
}

// Todo: When and if we have a async implementation, the client behavior can be changed by adding a interface for client.

func NewAriaServiceClient(ariaAdminClient *AriaAdminClient, ariaCredentials *AriaCredentials) (*AriaServiceClient, error) {
	logger := log.FromContext(context.Background()).WithName("AriaServiceClient.NewAriaServiceClient")
	mm := map[any]any{}

	if err := yaml.Unmarshal(ariaServiceRequestDefaults, &mm); err != nil {
		logger.Error(err, "failed to unmarshal YAML default values for create aria service request")
		return nil, err
	}
	ariaServiceRequestJson, err := json.Marshal(mapKeysToString(mm))
	if err != nil {
		logger.Error(err, "failed to JSON marshal default values for create aria service request")
		return nil, err
	}

	return &AriaServiceClient{
			ariaServiceRequestJson: ariaServiceRequestJson,
			ariaAdminClient:        ariaAdminClient,
			ariaCredentials:        ariaCredentials,
		},
		nil
}

func (ariaServiceClient *AriaServiceClient) CreateAriaService(ctx context.Context, clientServiceId string,
	serviceName string, usageType int) (*response.CreateService, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaServiceClient.CreateAriaService").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	createServiceRequest, err := ariaServiceClient.GetRequestForService(ctx, "create_service_m", clientServiceId, serviceName, usageType)
	if err != nil {
		return nil, err
	}
	return CallAriaAdmin[response.CreateService](ctx, ariaServiceClient.ariaAdminClient, createServiceRequest, FailedToCreateAriaServiceError)
}

func (ariaServiceClient *AriaServiceClient) UpdateAriaService(ctx context.Context, clientServiceId string,
	serviceName string, usageType int) (*response.CreateService, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaServiceClient.UpdateAriaService").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	updateServiceRequest, err := ariaServiceClient.GetRequestForService(ctx, "update_service_m", clientServiceId, serviceName, usageType)
	if err != nil {
		return nil, err
	}
	return CallAriaAdmin[response.CreateService](ctx, ariaServiceClient.ariaAdminClient, updateServiceRequest, FailedToUpdateAriaServiceError)
}

func (ariaServiceClient *AriaServiceClient) GetRequestForService(ctx context.Context, restCall string, clientServiceId string,
	serviceName string, usageType int) (*request.CreateService, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaServiceClient.CreateOrUpdateAriaService").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	ariaServiceRequest := request.CreateService{}
	ariaServiceRequest.RestCall = restCall
	err := json.Unmarshal(ariaServiceClient.ariaServiceRequestJson, &ariaServiceRequest)

	if err != nil {
		logger.Error(err, "failed to JSON unmarshal default values for create or update aria service request")
		return nil, err
	}

	ariaServiceRequest.ClientNo = ariaServiceClient.ariaCredentials.clientNo
	ariaServiceRequest.AuthKey = ariaServiceClient.ariaCredentials.authKey
	ariaServiceRequest.ClientServiceId = clientServiceId
	ariaServiceRequest.UsageType = usageType
	ariaServiceRequest.ServiceName = serviceName
	ariaServiceRequest.AltCallerId = AriaClientId
	return &ariaServiceRequest, nil
}

func (ariaServiceClient *AriaServiceClient) GetServiceDetailsForServiceNo(ctx context.Context, serviceNo int) (*response.GetServiceDetails, error) {
	getServiceDetailsRequest := request.GetServiceDetails{
		AriaRequest: request.AriaRequest{
			RestCall: "get_service_details_m"},
		OutputFormat: "json",
		ClientNo:     ariaServiceClient.ariaCredentials.clientNo,
		AuthKey:      ariaServiceClient.ariaCredentials.authKey,
		AltCallerId:  AriaClientId,
		ServiceNo:    serviceNo,
	}

	return CallAriaAdmin[response.GetServiceDetails](ctx, ariaServiceClient.ariaAdminClient, &getServiceDetailsRequest, FailedToGetServiceDetailsError)
}
