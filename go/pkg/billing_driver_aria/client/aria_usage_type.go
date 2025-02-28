// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

type AriaUsageTypeClient struct {
	ariaAdminClient *AriaAdminClient
	ariaCredentials *AriaCredentials
}

// Todo: When and if we have a async implementation, the client behavior can be changed by adding a interface for client.

func NewAriaUsageTypeClient(ariaAdminClient *AriaAdminClient, ariaCredentials *AriaCredentials) *AriaUsageTypeClient {
	return &AriaUsageTypeClient{
		ariaAdminClient: ariaAdminClient,
		ariaCredentials: ariaCredentials,
	}
}

func (aUsageTypeC *AriaUsageTypeClient) CreateUsageType(ctx context.Context, usageTypeName string, usageTypeDesc string, usageUnitTypeNo int, usageTypeCode string) (*response.CreateUsageType, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageTypeClient.CreateUsageType").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	createUsageTypeRequest := request.CreateUsageType{
		AriaRequest: request.AriaRequest{
			RestCall: "create_usage_type_m"},
		OutputFormat:      "json",
		ClientNo:          aUsageTypeC.ariaCredentials.clientNo,
		AuthKey:           aUsageTypeC.ariaCredentials.authKey,
		UsageTypeName:     usageTypeName,
		UsageTypeDesc:     usageTypeDesc,
		UsageUnitTypeNo:   usageUnitTypeNo,
		UsageTypeCode:     usageTypeCode,
		UsageRatingTiming: USAGES_RATING_TIMING,
	}
	logger.Info("request", "createUsageTypeRequest", createUsageTypeRequest)
	return CallAriaAdmin[response.CreateUsageType](ctx, aUsageTypeC.ariaAdminClient, &createUsageTypeRequest, FailedToCreateUsageTypeError)
}

func (aUsageTypeC *AriaUsageTypeClient) UpdateUsageType(ctx context.Context, usageTypeName string, usageTypeDesc string, usageUnitTypeNo int, usageTypeCode string) (*response.CreateUsageType, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageTypeClient.UpdateUsageType").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	updateUsageTypeRequest := request.CreateUsageType{
		AriaRequest: request.AriaRequest{
			RestCall: "update_usage_type_m"},
		OutputFormat:    "json",
		ClientNo:        aUsageTypeC.ariaCredentials.clientNo,
		AuthKey:         aUsageTypeC.ariaCredentials.authKey,
		UsageTypeName:   usageTypeName,
		UsageTypeDesc:   usageTypeDesc,
		UsageUnitTypeNo: usageUnitTypeNo,
		UsageTypeCode:   usageTypeCode,
	}
	return CallAriaAdmin[response.CreateUsageType](ctx, aUsageTypeC.ariaAdminClient, &updateUsageTypeRequest, FailedToUpdateUsageTypeError)
}

func (aUsageTypeC *AriaUsageTypeClient) GetUsageUnitTypes(ctx context.Context) (*response.GetUsageUnitTypes, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageTypeClient.GetUsageUnitTypes").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getUsageUnitTypesRequest := request.GetUsageUnitTypes{
		AriaRequest: request.AriaRequest{
			RestCall: "get_usage_unit_types_m"},
		OutputFormat: "json",
		ClientNo:     aUsageTypeC.ariaCredentials.clientNo,
		AuthKey:      aUsageTypeC.ariaCredentials.authKey,
		AltCallerId:  AriaClientId,
	}

	return CallAriaAdmin[response.GetUsageUnitTypes](ctx, aUsageTypeC.ariaAdminClient, &getUsageUnitTypesRequest, FailedToGetUsageUnitTypesError)
}

func (aUsageTypeC *AriaUsageTypeClient) GetMinuteUsageUnitType(ctx context.Context) (*data.UsageUnitType, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageTypeClient.GetMinuteUsageUnitType").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	usageUnitTypes, err := aUsageTypeC.GetUsageUnitTypes(ctx)
	if err != nil {
		return nil, err
	}
	for _, usageUnitType := range usageUnitTypes.UsageUnitTypes {
		// todo: check with aria team if there can be multiple usage unit types with the same description.
		if usageUnitType.UsageUnitTypeDesc == USAGE_UNIT_TYPE_MINUTE_DESC {
			return &usageUnitType, nil
		}
	}
	return nil, nil
}

func (aUsageTypeC *AriaUsageTypeClient) GetUsageTypes(ctx context.Context) (*response.GetUsageTypes, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageTypeClient.GetUsageTypes").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getUsageTypesRequest := request.GetUsageTypes{
		AriaRequest: request.AriaRequest{
			RestCall: "get_usage_types_m"},
		OutputFormat: "json",
		ClientNo:     aUsageTypeC.ariaCredentials.clientNo,
		AuthKey:      aUsageTypeC.ariaCredentials.authKey,
		AltCallerId:  AriaClientId,
	}

	return CallAriaAdmin[response.GetUsageTypes](ctx, aUsageTypeC.ariaAdminClient, &getUsageTypesRequest, FailedToGetUsageTypesError)
}

func (aUsageTypeC *AriaUsageTypeClient) GetUsageTypeDetails(ctx context.Context, usageTypeCode string) (*response.GetUsageTypeDetails, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageTypeClient.GetUsageTypeDetails").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getUsageTypeDetailsRequest := request.GetUsageTypeDetails{
		AriaRequest: request.AriaRequest{
			RestCall: "get_usage_type_details_m"},
		OutputFormat:  "json",
		ClientNo:      aUsageTypeC.ariaCredentials.clientNo,
		AuthKey:       aUsageTypeC.ariaCredentials.authKey,
		AltCallerId:   AriaClientId,
		UsageTypeCode: usageTypeCode,
	}

	return CallAriaAdmin[response.GetUsageTypeDetails](ctx, aUsageTypeC.ariaAdminClient, &getUsageTypeDetailsRequest, FailedToGetUsageTypeDetailsError)
}

func (aUsageTypeC *AriaUsageTypeClient) GetMinutesUsageType(ctx context.Context) (*data.UsageType, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageTypeClient.GetMinutesUsageType").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	usageTypeDetails, err := aUsageTypeC.GetUsageTypeDetails(ctx, GetMinsUsageTypeCode())
	if err != nil {
		return nil, err
	}
	return &data.UsageType{
		UsageTypeNo:   usageTypeDetails.UsageTypeNo,
		UsageTypeDesc: usageTypeDetails.UsageTypeDesc,
		UsageUnitType: usageTypeDetails.UsageUnitType,
		UsageTypeName: usageTypeDetails.UsageTypeName,
		IsEditable:    usageTypeDetails.IsEditable,
		UsageTypeCode: usageTypeDetails.UsageTypeCode,
	}, nil
}

func (aUsageTypeC *AriaUsageTypeClient) GetStorageUsageUnitType(ctx context.Context) (*data.UsageUnitType, error) {
	usageUnitTypes, err := aUsageTypeC.GetUsageUnitTypes(ctx)
	if err != nil {
		return nil, err
	}
	for _, usageUnitType := range usageUnitTypes.UsageUnitTypes {
		if usageUnitType.UsageUnitTypeDesc == GetAriaSystemStorageUsageUnitTypeName() {
			return &usageUnitType, nil
		}
	}
	return nil, nil
}

func (aUsageTypeC *AriaUsageTypeClient) GetStorageUsageType(ctx context.Context) (*data.UsageType, error) {
	usageTypeDetails, err := aUsageTypeC.GetUsageTypeDetails(ctx, GetStorageUsageUnitTypeCode())
	if err != nil {
		return nil, err
	}
	return &data.UsageType{
		UsageTypeNo:   usageTypeDetails.UsageTypeNo,
		UsageTypeDesc: usageTypeDetails.UsageTypeDesc,
		UsageUnitType: usageTypeDetails.UsageUnitType,
		UsageTypeName: usageTypeDetails.UsageTypeName,
		IsEditable:    usageTypeDetails.IsEditable,
		UsageTypeCode: usageTypeDetails.UsageTypeCode,
	}, nil
}

func (aUsageTypeC *AriaUsageTypeClient) GetInferenceUsageUnitType(ctx context.Context) (*data.UsageUnitType, error) {
	usageUnitTypes, err := aUsageTypeC.GetUsageUnitTypes(ctx)
	if err != nil {
		return nil, err
	}
	for _, usageUnitType := range usageUnitTypes.UsageUnitTypes {
		if usageUnitType.UsageUnitTypeDesc == GetAriaSystemInferenceUsageUnitTypeName() {
			return &usageUnitType, nil
		}
	}
	return nil, nil
}

func (aUsageTypeC *AriaUsageTypeClient) GetInferenceUsageType(ctx context.Context) (*data.UsageType, error) {
	usageTypeDetails, err := aUsageTypeC.GetUsageTypeDetails(ctx, GetInferenceUsageUnitTypeCode())
	if err != nil {
		return nil, err
	}
	return &data.UsageType{
		UsageTypeNo:   usageTypeDetails.UsageTypeNo,
		UsageTypeDesc: usageTypeDetails.UsageTypeDesc,
		UsageUnitType: usageTypeDetails.UsageUnitType,
		UsageTypeName: usageTypeDetails.UsageTypeName,
		IsEditable:    usageTypeDetails.IsEditable,
		UsageTypeCode: usageTypeDetails.UsageTypeCode,
	}, nil
}

func (aUsageTypeC *AriaUsageTypeClient) GetTokenUsageUnitType(ctx context.Context) (*data.UsageUnitType, error) {
	usageUnitTypes, err := aUsageTypeC.GetUsageUnitTypes(ctx)
	if err != nil {
		return nil, err
	}
	for _, usageUnitType := range usageUnitTypes.UsageUnitTypes {
		if usageUnitType.UsageUnitTypeDesc == GetAriaSystemTokenUsageUnitTypeName() {
			return &usageUnitType, nil
		}
	}
	return nil, nil
}

func (aUsageTypeC *AriaUsageTypeClient) GetTokenUsageType(ctx context.Context) (*data.UsageType, error) {
	usageTypeDetails, err := aUsageTypeC.GetUsageTypeDetails(ctx, GetTokenUsageUnitTypeCode())
	if err != nil {
		return nil, err
	}
	return &data.UsageType{
		UsageTypeNo:   usageTypeDetails.UsageTypeNo,
		UsageTypeDesc: usageTypeDetails.UsageTypeDesc,
		UsageUnitType: usageTypeDetails.UsageUnitType,
		UsageTypeName: usageTypeDetails.UsageTypeName,
		IsEditable:    usageTypeDetails.IsEditable,
		UsageTypeCode: usageTypeDetails.UsageTypeCode,
	}, nil
}
