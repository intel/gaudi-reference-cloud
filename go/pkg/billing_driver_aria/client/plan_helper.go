// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type ProductClientPlan struct {
	Product          *pb.Product
	ClientPlanDetail data.AllClientPlanDtl
	UpdateRequired   bool
	IsActive         bool
}

func DebugLogDifference(ctx context.Context, ariaKey string, ariaValue string, currentValue string) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ProductClientPlan.DebugLogDifference").Start()
	defer span.End()
	logger.V(1)
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info(fmt.Sprintf("aria response value %s for %s is different form current %s", ariaValue, ariaKey, currentValue))
}

func HasDiffPlanDetail(ctx context.Context, clientPlanDtl data.AllClientPlanDtl, product *pb.Product) bool {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ProductClientPlan.HasDiffPlanDetail").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	hasDiff := false
	clientPlanId := GetPlanClientId(product.GetId())
	if clientPlanDtl.PlanName != product.Name {
		hasDiff = true
		DebugLogDifference(ctx, "ClientPlanDtl.PlanName", clientPlanDtl.PlanName, product.Name)

	} else if clientPlanDtl.ClientPlanId != clientPlanId {
		hasDiff = true
		DebugLogDifference(ctx, "ClientPlanDtl.ClientPlanId", clientPlanDtl.ClientPlanId, clientPlanId)
	} else if clientPlanDtl.PlanDesc != product.Description {
		hasDiff = true
		DebugLogDifference(ctx, "ClientPlanDtl.PlanDesc", clientPlanDtl.PlanDesc, product.Description)
	}
	return hasDiff
}

func HasDiffPlanServiceRate(ctx context.Context, planServiceRates []data.PlanServiceRate, product *pb.Product) bool {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ProductClientPlan.HasDiffPlanServiceRate").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	rateMap := GetRateMap(product)
	hasDiff := false
	for _, planServiceRate := range planServiceRates {
		if rate, found := rateMap[planServiceRate.ClientRateScheduleId]; found {
			fromUnit := float32(TIERED_SCHEDULE_FROM_UNITS)
			if planServiceRate.FromUnit != fromUnit {
				hasDiff = true
				DebugLogDifference(ctx, "PlanServiceRate.FromUnit", strconv.FormatFloat(float64(planServiceRate.FromUnit), 'f', 6, 32), strconv.FormatFloat(float64(fromUnit), 'f', 6, 32))
				break
			}
			ratePerUnit, err := strconv.ParseFloat(rate, 64)
			if err != nil {
				logger.Error(err, "Error while parsing to float", rate)
				return false
			}
			if planServiceRate.RatePerUnit != ratePerUnit {
				hasDiff = true
				DebugLogDifference(ctx, "PlanServiceRate.RatePerUnit", strconv.FormatFloat(float64(planServiceRate.RatePerUnit), 'f', 6, 32), strconv.FormatFloat(float64(ratePerUnit), 'f', 6, 32))
				break
			}
		}
	}
	return hasDiff
}

func GetRateMap(product *pb.Product) map[string]string {
	rateMap := map[string]string{}
	for _, rate := range product.Rates {
		accountType := GetAccountType(rate)
		clientRateScheduleId := GetRateScheduleClientId(product.GetId(), accountType)
		rateMap[clientRateScheduleId] = rate.Rate
	}
	return rateMap
}

// TODO: pricing_rule, high_water, tax_inclusive_ind, taxable_ind, tax_group,
func HasDiffPlanService(ctx context.Context, planService data.PlanService, usageType *data.UsageType, product *pb.Product) bool {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ProductClientPlan.HasDiffPlanService").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	hasDiff := false
	clientServiceId := GetServiceClientId(product.Id)

	if planService.ClientServiceId != clientServiceId {
		hasDiff = true
		DebugLogDifference(ctx, "PlanService.ClientServiceId", planService.ClientServiceId, clientServiceId)
	} else if svcDesc, ok := product.Metadata["displayName"]; ok && planService.ServiceDesc != svcDesc {
		hasDiff = true
		DebugLogDifference(ctx, "PlanService.ServiceDesc", planService.ServiceDesc, product.Metadata["displayName"])
	} else if planService.UsageType != int64(usageType.UsageTypeNo) {
		hasDiff = true
		DebugLogDifference(ctx, "PlanService.UsageType", strconv.FormatInt(planService.UsageType, 10), strconv.Itoa(usageType.UsageTypeNo))
	} else if planService.UsageTypeName != usageType.UsageTypeName {
		hasDiff = true
		DebugLogDifference(ctx, "PlanService.UsageTypeName", planService.UsageTypeName, usageType.UsageTypeName)

	} else if planService.UsageTypeDesc != usageType.UsageTypeDesc {
		hasDiff = true
		DebugLogDifference(ctx, "PlanService.UsageTypeDesc", planService.UsageTypeDesc, usageType.UsageTypeDesc)

	} else if planService.UsageUnitLabel != usageType.UsageUnitType {
		hasDiff = true
		DebugLogDifference(ctx, "PlanService.UsageUnitLabel", planService.UsageUnitLabel, usageType.UsageUnitType)

	} else if planService.TaxableInd != 1 {
		hasDiff = true
		DebugLogDifference(ctx, "PlanService.TaxableInd", fmt.Sprintf("%d", planService.TaxableInd), "1")
	}
	return hasDiff
}

func HasDiffPlanSupplField(ctx context.Context, supplementalObjectFields []data.PlanSuppField, product *pb.Product) bool {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ProductClientPlan.HasDiffPlanSupplField").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	hasDiff := false
	for _, supplementalObjectField := range supplementalObjectFields {
		if supplementalObjectField.PlanSuppFieldName == PLANSUPPFIELDPCQID &&
			supplementalObjectField.PlanSuppFieldValue != product.Pcq {
			hasDiff = true
			DebugLogDifference(ctx, "PlanSuppField.PlanSuppFieldValue", supplementalObjectField.PlanSuppFieldValue, product.Pcq)
			break
		}
		businessUnit := GetBusinessUnit(product)
		if supplementalObjectField.PlanSuppFieldName == PLAN_SUPP_FIELD_BUSINESS_UNIT &&
			supplementalObjectField.PlanSuppFieldValue != businessUnit {
			hasDiff = true
			DebugLogDifference(ctx, "PlanSuppField.PlanSuppFieldValue", supplementalObjectField.PlanSuppFieldValue, businessUnit)
			break
		}
		businessUnitContact := GetBusinessUnitContact(product)
		if supplementalObjectField.PlanSuppFieldName == PLAN_SUPP_FIELD_BUSINESS_UNIT_CONTACT &&
			supplementalObjectField.PlanSuppFieldValue != businessUnitContact {
			hasDiff = true
			DebugLogDifference(ctx, "PlanSuppField.PlanSuppFieldValue", supplementalObjectField.PlanSuppFieldValue, businessUnitContact)
			break
		}
		glNumber := GetGLNumber(product)
		if supplementalObjectField.PlanSuppFieldName == PLAN_SUPP_FIELD_GL_NUMBER &&
			supplementalObjectField.PlanSuppFieldValue != glNumber {
			hasDiff = true
			DebugLogDifference(ctx, "PlanSuppField.PlanSuppFieldValue", supplementalObjectField.PlanSuppFieldValue, glNumber)
			break
		}
		legalEntity := GetLegalEntity(product)
		if supplementalObjectField.PlanSuppFieldName == PLAN_SUPP_FIELD_LEGAL_ENTITY &&
			supplementalObjectField.PlanSuppFieldValue != legalEntity {
			hasDiff = true
			DebugLogDifference(ctx, "PlanSuppField.PlanSuppFieldValue", supplementalObjectField.PlanSuppFieldValue, legalEntity)
			break
		}
		productLine := GetProductLine(product)
		if supplementalObjectField.PlanSuppFieldName == PLAN_SUPP_FIELD_PRODUCT_LINE &&
			supplementalObjectField.PlanSuppFieldValue != GetProductLine(product) {
			hasDiff = true
			DebugLogDifference(ctx, "PlanSuppField.PlanSuppFieldValue", supplementalObjectField.PlanSuppFieldValue, productLine)
			break
		}
		profitCenter := GetProfitCenter(product)
		if supplementalObjectField.PlanSuppFieldName == PLAN_SUPP_FIELD_PROFIT_CENTER &&
			supplementalObjectField.PlanSuppFieldValue != profitCenter {
			hasDiff = true
			DebugLogDifference(ctx, "PlanSuppField.PlanSuppFieldValue", supplementalObjectField.PlanSuppFieldValue, profitCenter)
			break
		}
		superGrp := GetSuberGroup(product)
		if supplementalObjectField.PlanSuppFieldName == PLAN_SUPP_FIELD_SUPER_GROUP &&
			supplementalObjectField.PlanSuppFieldValue != superGrp {
			hasDiff = true
			DebugLogDifference(ctx, "PlanSuppField.PlanSuppFieldValue", supplementalObjectField.PlanSuppFieldValue, superGrp)
			break
		}
	}
	return hasDiff
}

func MapResponseToClientPlanDetail(resp *response.GetPlanDetailResponse, planServices []data.PlanService) *data.AllClientPlanDtl {
	clientPlanDetail := &data.AllClientPlanDtl{}
	clientPlanDetail.PlanNo = int64(resp.PlanNo)
	clientPlanDetail.ClientPlanId = resp.ClientPlanId
	clientPlanDetail.PlanName = resp.PlanName
	clientPlanDetail.PlanDesc = resp.PlanDesc
	clientPlanDetail.PlanServices = planServices
	clientPlanDetail.PlanSuppFields = resp.PlanSuppFields
	return clientPlanDetail
}

func GetConfigTrimedId(ariaId string) string {
	return ariaId[len(config.Cfg.ClientIdPrefix):]
}

func CreateProductClientPlanMap(ctx context.Context, clientPlansDetail []data.AllClientPlanDtl) map[string]*ProductClientPlan {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ProductClientPlan.CreateProductClientPlanMap").Start()
	defer span.End()
	logger.V(1)
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	planMap := map[string]*ProductClientPlan{}
	for _, clientPlanDetail := range clientPlansDetail {
		logger.Info("client plan id", "ClientPlanId", clientPlanDetail.ClientPlanId)
		if !filterInvalidPlan(ctx, clientPlanDetail) {
			planMap[clientPlanDetail.ClientPlanId] = &ProductClientPlan{UpdateRequired: false,
				IsActive: false, ClientPlanDetail: clientPlanDetail, Product: &pb.Product{}}
		}
	}
	return planMap
}

func filterInvalidPlan(ctx context.Context, clientPlanDetail data.AllClientPlanDtl) bool {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ProductClientPlan.filterInvalidPlan").Start()
	defer span.End()
	logger.V(1)
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	if strings.HasPrefix(clientPlanDetail.ClientPlanId, config.Cfg.ClientIdPrefix) {
		return false
	}
	logger.Info("invalid client plan id", "ClientPlanId", clientPlanDetail.ClientPlanId)
	return true
}
