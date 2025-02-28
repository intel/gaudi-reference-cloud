// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func verifyCheckIfProductMapsToMeteringProperties(t *testing.T, matchExpression string, meteringProperties map[string]string) {
	logger := log.FromContext(context.Background()).WithName("verifyCheckIfProductMapsToMeteringProperties")
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	productId := uuid.NewString()
	matchingExpressionProduct := &pb.Product{
		Name:        "computeProductVMSmallXeon3Name",
		Id:          productId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       GetRates(),
		MatchExpr:   matchExpression,
	}

	productMatches, err := CheckIfProductMapsToProperties(matchingExpressionProduct, meteringProperties)

	if err != nil {
		t.Fatalf("failed to match product to expression: %v", err)
	}
	if !productMatches {
		t.Fatalf("failed to match product")
	}
	logger.Info("Verified match expression type 1", "expression", meteringProperties)

}

func TestCheckProductMatchesMetering(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCheckProductMatchesMetering")
	tests := []struct {
		enabled                            bool
		expressionType                     string
		matchingExpression                 string
		matchedProperties                  map[string]string
		verifyCheckIfProductMapsToMetering func(t *testing.T, matchExpression string, usageProperties map[string]string)
	}{
		{
			enabled:            true,
			expressionType:     "evalTrue",
			matchingExpression: "billUsage",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalNotTrue",
			matchingExpression: "!complianceAdded",
			matchedProperties: map[string]string{
				"region":          DefaultServiceRegion,
				"service":         idcComputeServiceName,
				"billUsage":       "true",
				"complianceAdded": "false",
				"instanceType":    xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "equals",
			matchingExpression: fmt.Sprintf("service == \"%s\"", idcComputeServiceName),
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "equalsEmpty",
			matchingExpression: "service == \"\"",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      "",
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "equalsEmptyForNumber",
			matchingExpression: "cpuCount == \"\"",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      "",
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
				"cpuCount":     "0",
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "equalsEmptyNotPresent",
			matchingExpression: "service == \"\"",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "equals0NotPresent",
			matchingExpression: "cpuCount == 0",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      "",
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "notequals",
			matchingExpression: fmt.Sprintf("service != \"%s\"", idcComputeServiceName),
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      uuid.NewString(),
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "notequalsEmpty",
			matchingExpression: "service != \"\"",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      uuid.NewString(),
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "notequalsEmptyForNumber",
			matchingExpression: "cpuCount != \"\"",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      "",
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
				"cpuCount":     "10",
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalTrue&&equals",
			matchingExpression: fmt.Sprintf("billUsage && service == \"%s\" && instanceType == \"%s\"", idcComputeServiceName, xeon3SmallInstanceType),
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalTrue&&orequals",
			matchingExpression: fmt.Sprintf("billUsage && (instanceType == \"%s\" || instanceType == someInstanceType)", xeon3SmallInstanceType),
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalTrue&&evalNotTrue&&equals",
			matchingExpression: fmt.Sprintf("billUsage && !complianceAdded && service == \"%s\" && instanceType == \"%s\"", idcComputeServiceName, xeon3SmallInstanceType),
			matchedProperties: map[string]string{
				"region":          DefaultServiceRegion,
				"service":         idcComputeServiceName,
				"billUsage":       "true",
				"complianceAdded": "false",
				"instanceType":    xeon3SmallInstanceType,
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalTrue&&evalNotTrue&&equals&&notequals",
			matchingExpression: fmt.Sprintf("billUsage && !complianceAdded && service == \"%s\" && instanceType == \"%s\" && instanceName != \"someName\"", idcComputeServiceName, xeon3SmallInstanceType),
			matchedProperties: map[string]string{
				"region":          DefaultServiceRegion,
				"service":         idcComputeServiceName,
				"billUsage":       "true",
				"complianceAdded": "false",
				"instanceType":    xeon3SmallInstanceType,
				"instanceName":    uuid.NewString(),
			},
			verifyCheckIfProductMapsToMetering: verifyCheckIfProductMapsToMeteringProperties,
		},
	}
	for index := range tests {
		if tests[index].enabled {
			logger.Info("Running tests for expression of", "expression type", tests[index].expressionType)
			tests[index].verifyCheckIfProductMapsToMetering(t, tests[index].matchingExpression, tests[index].matchedProperties)
		} else {
			logger.Info("Enable the test for expression of", "expression type", tests[index].expressionType)
		}
	}

	logger.Info("All verifications done")
}

func verifyErrorsCheckIfProductMapsToMeteringProperties(t *testing.T, matchExpression string, meteringProperties map[string]string) {
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	productId := uuid.NewString()
	matchingExpressionProduct := &pb.Product{
		Name:        "computeProductVMSmallXeon3Name",
		Id:          productId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       GetRates(),
		MatchExpr:   matchExpression,
	}

	_, err := CheckIfProductMapsToProperties(matchingExpressionProduct, meteringProperties)
	if err == nil {
		t.Fatalf("failed to generate error when error was expected")
	}
}

func TestCheckProductMatchesMeteringErrors(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCheckProductMatchesMeteringErrors")
	tests := []struct {
		enabled                                  bool
		expressionType                           string
		matchingExpression                       string
		matchedProperties                        map[string]string
		verifyErrorsCheckIfProductMapsToMetering func(t *testing.T, matchExpression string, meteringProperties map[string]string)
	}{
		{
			enabled:            true,
			expressionType:     "evalTrueWrongFormat",
			matchingExpression: "billUsage$$",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyErrorsCheckIfProductMapsToMetering: verifyErrorsCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalTrueWrongFormat",
			matchingExpression: "billUsage&&",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyErrorsCheckIfProductMapsToMetering: verifyErrorsCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalTrueWrongValueInProperties",
			matchingExpression: "billUsage",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "10",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyErrorsCheckIfProductMapsToMetering: verifyErrorsCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalNotTrueWrongFormat",
			matchingExpression: "!complianceAdded$$",
			matchedProperties: map[string]string{
				"region":          DefaultServiceRegion,
				"service":         idcComputeServiceName,
				"billUsage":       "true",
				"complianceAdded": "false",
				"instanceType":    xeon3SmallInstanceType,
			},
			verifyErrorsCheckIfProductMapsToMetering: verifyErrorsCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalNotTrueWrongValueInProperties",
			matchingExpression: "!complianceAdded",
			matchedProperties: map[string]string{
				"region":          DefaultServiceRegion,
				"service":         idcComputeServiceName,
				"billUsage":       "true",
				"complianceAdded": "10",
				"instanceType":    xeon3SmallInstanceType,
			},
			verifyErrorsCheckIfProductMapsToMetering: verifyErrorsCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "equalsWrongFormat",
			matchingExpression: fmt.Sprintf("service$$ == \"%s\"", idcComputeServiceName),
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyErrorsCheckIfProductMapsToMetering: verifyErrorsCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "notequalsWrongFormat",
			matchingExpression: fmt.Sprintf("service$$ != \"%s\"", idcComputeServiceName),
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      uuid.NewString(),
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyErrorsCheckIfProductMapsToMetering: verifyErrorsCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalTrue&&equalsNotEquals",
			matchingExpression: fmt.Sprintf("billUsage && service ==!= \"%s\" && instanceType == \"%s\"", idcComputeServiceName, xeon3SmallInstanceType),
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyErrorsCheckIfProductMapsToMetering: verifyErrorsCheckIfProductMapsToMeteringProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalTrue&&orequals",
			matchingExpression: fmt.Sprintf("billUsage && (instanceType == \"%s\" && instanceType == someInstanceType)", xeon3SmallInstanceType),
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyErrorsCheckIfProductMapsToMetering: verifyErrorsCheckIfProductMapsToMeteringProperties,
		},
	}
	for index := range tests {
		if tests[index].enabled {
			logger.Info("Running tests for expression of", "expression type", tests[index].expressionType)
			tests[index].verifyErrorsCheckIfProductMapsToMetering(t, tests[index].matchingExpression, tests[index].matchedProperties)
		} else {
			logger.Info("Enable the test for expression of", "expression type", tests[index].expressionType)
		}
	}

	logger.Info("All verifications done")
}

func verifyCheckIfProductDoesNotMapToProperties(t *testing.T, matchExpression string, meteringProperties map[string]string) {
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	productId := uuid.NewString()
	matchingExpressionProduct := &pb.Product{
		Name:        "computeProductVMSmallXeon3Name",
		Id:          productId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       GetRates(),
		MatchExpr:   matchExpression,
	}
	productMatches, err := CheckIfProductMapsToProperties(matchingExpressionProduct, meteringProperties)
	if err != nil {
		t.Fatalf("failed to check if product matches expression: %v", err)
	}
	if productMatches {
		t.Fatalf("product matched when it is not supposed to")
	}
}

func TestCheckProductDoesNotMatchMetering(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCheckProductDoesNotMatchMetering")
	tests := []struct {
		enabled                                  bool
		expressionType                           string
		matchingExpression                       string
		matchedProperties                        map[string]string
		verifyCheckIfProductDoesNotMapToMetering func(t *testing.T, matchExpression string, meteringProperties map[string]string)
	}{
		{
			enabled:            true,
			expressionType:     "evalTrueNoMatch",
			matchingExpression: "billUsage",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "false",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductDoesNotMapToMetering: verifyCheckIfProductDoesNotMapToProperties,
		},
		{
			enabled:            true,
			expressionType:     "evalNotTrueNoMatch",
			matchingExpression: "!complianceAdded",
			matchedProperties: map[string]string{
				"region":          DefaultServiceRegion,
				"service":         idcComputeServiceName,
				"billUsage":       "true",
				"complianceAdded": "true",
				"instanceType":    xeon3SmallInstanceType,
			},
			verifyCheckIfProductDoesNotMapToMetering: verifyCheckIfProductDoesNotMapToProperties,
		},
		{
			enabled:            true,
			expressionType:     "equalsNoMatch",
			matchingExpression: fmt.Sprintf("service == \"%s\"", idcComputeServiceName),
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      uuid.NewString(),
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductDoesNotMapToMetering: verifyCheckIfProductDoesNotMapToProperties,
		},
		{
			enabled:            true,
			expressionType:     "equalsEmptyNoMatch",
			matchingExpression: "service == \"\"",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      uuid.NewString(),
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductDoesNotMapToMetering: verifyCheckIfProductDoesNotMapToProperties,
		},
		{
			enabled:            true,
			expressionType:     "equalsEmptyForNumberNoMatch",
			matchingExpression: "cpuCount == \"\"",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      "",
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
				"cpuCount":     "10",
			},
			verifyCheckIfProductDoesNotMapToMetering: verifyCheckIfProductDoesNotMapToProperties,
		},
		{
			enabled:            true,
			expressionType:     "notequalsNoMatch",
			matchingExpression: fmt.Sprintf("service != \"%s\"", idcComputeServiceName),
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      idcComputeServiceName,
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductDoesNotMapToMetering: verifyCheckIfProductDoesNotMapToProperties,
		},
		{
			enabled:            true,
			expressionType:     "notequalsEmptyNoMatch",
			matchingExpression: "service != \"\"",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      "",
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
			},
			verifyCheckIfProductDoesNotMapToMetering: verifyCheckIfProductDoesNotMapToProperties,
		},
		{
			enabled:            true,
			expressionType:     "notequalsEmptyForNumberNoMatch",
			matchingExpression: "cpuCount != \"\"",
			matchedProperties: map[string]string{
				"region":       DefaultServiceRegion,
				"service":      "",
				"billUsage":    "true",
				"instanceType": xeon3SmallInstanceType,
				"cpuCount":     "0",
			},
			verifyCheckIfProductDoesNotMapToMetering: verifyCheckIfProductDoesNotMapToProperties,
		},
	}
	for index := range tests {
		if tests[index].enabled {
			logger.Info("Running tests for expression of", "expression type", tests[index].expressionType)
			tests[index].verifyCheckIfProductDoesNotMapToMetering(t, tests[index].matchingExpression, tests[index].matchedProperties)
		} else {
			logger.Info("Enable the test for expression of", "expression type", tests[index].expressionType)
		}
	}

	logger.Info("All verifications done")
}
