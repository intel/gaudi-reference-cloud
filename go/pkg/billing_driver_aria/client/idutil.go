// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"fmt"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"golang.org/x/exp/slices"
)

var storageUsageUnitTypes = []string{pb.RateUnit_RATE_UNIT_DOLLARS_PER_TB_PER_HOUR.Enum().String(), "per TB per Hour"}
var inferenceUsageUnitType = "RATE_UNIT_DOLLARS_PER_INFERENCE"
var tokenUsageUnitType = "RATE_UNIT_DOLLARS_PER_MILLION_TOKENS"

func GetClientIdPrefix() string {
	return config.Cfg.ClientIdPrefix
}

func GetPlanClientId(productId string) string {
	return config.Cfg.ClientIdPrefix + "." + productId
}

func GetMinsUsageTypeCode() string {
	return config.Cfg.ClientIdPrefix + "." + USAGE_TYPE_MINUTES_CODE_SUFFIX
}

func GetMinsUsageTypeDesc() string {
	return config.Cfg.ClientIdPrefix + "." + USAGE_TYPE_DESC
}

func GetPlanSetId() string {
	return strings.ToUpper(config.Cfg.ClientIdPrefix) + "_Plans"
}

func GetPlanSetName() string {
	return strings.ToUpper(config.Cfg.ClientIdPrefix) + " Plans"
}

func GetPromoCode() string {
	return config.Cfg.ClientIdPrefix + "_plans"
}

func GetDefaultPlanClientId() string {
	return config.Cfg.ClientIdPrefix + "." + "master"
}

func GetAccountClientId(cloudaccountId string) string {
	return config.Cfg.ClientIdPrefix + "." + cloudaccountId
}

func GetServiceClientId(id string) string {
	return config.Cfg.ClientIdPrefix + "." + id
}

func GetClientAcctGroupId() string {
	return config.Cfg.ClientAcctGroupId
}

func GetRateScheduleClientId(productId string, schedName string) string {
	return fmt.Sprintf("%s.%s.%s", config.Cfg.ClientIdPrefix, productId, schedName)
}

func GetClientMasterPlanInstanceId(cloudaccountId string, productId string) string {
	return config.Cfg.ClientIdPrefix + "." + cloudaccountId + "." + productId
}

func GetBillingGroupId(clientAccountID string) string {
	return clientAccountID + "." + "billing_group"
}

func GetBillingGroupName(clientPlanID string) string {
	return clientPlanID + "." + "billing_group"
}

func GetDunningGroupId(clientAccountID string) string {
	return clientAccountID + "." + "dunning_group"
}

func GetDunningGroupName(clientPlanID string) string {
	return clientPlanID + "." + "dunning_group"
}

func GetDefaultNotificationTemplateGroupId() string {
	return "B2C_US_Statement_Template"
}

func GetStorageUsageUnitTypeCode() string {
	return config.Cfg.ClientIdPrefix + "." + strings.ReplaceAll(config.Cfg.GetProductCatalogStorageUsageUnitType(), " ", "")
}

func GetTokenUsageUnitTypeCode() string {
	return config.Cfg.ClientIdPrefix + "." + strings.ReplaceAll(config.Cfg.GetProductCatalogTokenUsageUnitType(), " ", "")
}

func GetInferenceUsageUnitTypeCode() string {
	return config.Cfg.ClientIdPrefix + "." + strings.ReplaceAll(config.Cfg.GetProductCatalogInferenceUsageUnitType(), " ", "")
}

func GetStorageUsageUnitTypeDesc() string {
	return config.Cfg.ClientIdPrefix + "." + strings.ReplaceAll(config.Cfg.GetProductCatalogStorageUsageUnitType(), " ", "")

}

func GetTokenUsageUnitTypeDesc() string {
	return config.Cfg.ClientIdPrefix + "." + strings.ReplaceAll(config.Cfg.GetProductCatalogTokenUsageUnitType(), " ", "")

}

func GetInferenceUsageUnitTypeDesc() string {
	return config.Cfg.ClientIdPrefix + "." + strings.ReplaceAll(config.Cfg.GetProductCatalogInferenceUsageUnitType(), " ", "")

}

func GetProductStorageUsageUnitTypeName() string {
	return config.Cfg.GetProductCatalogStorageUsageUnitType()
}

func GetProductInferenceUsageUnitTypeName() string {
	return config.Cfg.GetProductCatalogInferenceUsageUnitType()
}

func GetProductTokenUsageUnitTypeName() string {
	return config.Cfg.GetProductCatalogTokenUsageUnitType()
}

func GetAriaSystemStorageUsageUnitTypeName() string {
	return config.Cfg.GetAriaSystemStorageUsageUnitType()
}

func GetAriaSystemTokenUsageUnitTypeName() string {
	return config.Cfg.GetAriaSystemTokenUsageUnitType()
}

func GetAriaSystemInferenceUsageUnitTypeName() string {
	return config.Cfg.GetAriaSystemInferenceUsageUnitType()
}

func GetUsageUnitTypeCode(usageType string) string {
	if slices.Contains(storageUsageUnitTypes, usageType) {
		return GetStorageUsageUnitTypeCode()
	}
	if inferenceUsageUnitType == usageType {
		return GetInferenceUsageUnitTypeCode()
	}
	if tokenUsageUnitType == usageType {
		return GetTokenUsageUnitTypeCode()
	}

	return GetMinsUsageTypeCode()
}
