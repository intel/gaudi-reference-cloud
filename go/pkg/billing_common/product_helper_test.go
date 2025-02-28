// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	cryptoRand "crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// Note The tests are using cloud account Id as random UUID although it is not which is perfectly fine.
// However, if we try to store the cloud account Id as a metering data which has a smaller size and is also fine because we won't store.

var (
	pmaxId int64 = 1_000_000_000_000
)

func getCloudAccountIdForUsageHelperTests(t *testing.T) string {
	intId, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(pmaxId))
	if err != nil {
		t.Fatal("failed to create cloud account id")
	}
	return fmt.Sprintf("%012d", intId)
}

func TestValidateProductsForInvalidProductFamily(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestValidateProductsForInvalidProductFamily")
	idcComputeProductFamilies := make([]*pb.ProductFamily, 0)
	idcVendors := make([]*pb.Vendor, 0)
	idcProducts := make([]*pb.Product, 0)
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	idcComputeProductFamily := GetIdcComputeProductFamily(idcComputeProductFamilyId)
	idcComputeProductFamilies = append(idcComputeProductFamilies, idcComputeProductFamily)
	vendor := GetIdcVendor(vendorId, idcComputeProductFamilies)
	productId := uuid.NewString()
	validProduct := GetProduct(computeProductVMSmallXeon3Name, productId, vendorId, idcComputeProductFamilyId, uuid.NewString(), computeProductVMSmallXeon3MatchExpr)
	invalidProduct := GetProduct(computeProductVMSmallXeon4Name, productId, vendorId, uuid.NewString(), uuid.NewString(), computeProductVMSmallXeon4MatchExpr)
	idcVendors = append(idcVendors, vendor)
	idcProducts = append(idcProducts, validProduct)
	idcProducts = append(idcProducts, invalidProduct)
	logger.Info("valid product", "id", validProduct.GetId())
	logger.Info("invalid product", "id", invalidProduct.GetId())
	validProducts, invalidProducts, err := ValidateProductsForProductFamily(idcVendors, idcProducts)
	if err != nil {
		t.Fatalf("failed to validate products for product family: %v", err)
	}
	if len(validProducts) != 1 {
		t.Fatalf("got more than one valid products instead of 1")
	}
	if len(invalidProducts) != 1 {
		t.Fatalf("got more than one invalid products instead of 1")
	}
	logger.Info("valid product returned", "id", validProducts[0].GetId())
	logger.Info("invalid product returned", "id", invalidProducts[0].GetId())
	if validProducts[0].Id != validProduct.Id {
		t.Fatalf("got wrong valid product")
	}
	if invalidProducts[0].Id != validProduct.Id {
		t.Fatalf("got wrong invalid product")
	}
	logger.Info("validated product helper for checking for invalid product family")
}

func TestValidateProductsForVendors(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestValidateProductsForVendors")
	idcComputeProductFamilies := make([]*pb.ProductFamily, 0)
	idcVendors := make([]*pb.Vendor, 0)
	idcProducts := make([]*pb.Product, 0)
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	idcComputeProductFamily := GetIdcComputeProductFamily(idcComputeProductFamilyId)
	idcComputeProductFamilies = append(idcComputeProductFamilies, idcComputeProductFamily)
	vendor := GetIdcVendor(vendorId, idcComputeProductFamilies)
	productId := uuid.NewString()
	validProduct := GetProduct(computeProductVMSmallXeon3Name, productId, vendorId, idcComputeProductFamilyId, uuid.NewString(), computeProductVMSmallXeon3MatchExpr)
	invalidProduct := GetProduct(computeProductVMSmallXeon4Name, productId, uuid.NewString(), idcComputeProductFamilyId, uuid.NewString(), computeProductVMSmallXeon4MatchExpr)
	idcVendors = append(idcVendors, vendor)
	idcProducts = append(idcProducts, validProduct)
	idcProducts = append(idcProducts, invalidProduct)
	logger.Info("valid product", "id", validProduct.GetId())
	logger.Info("invalid product", "id", invalidProduct.GetId())
	validProducts, invalidProducts, err := ValidateProductsForVendors(idcVendors, idcProducts)
	if err != nil {
		t.Fatalf("failed to validate products for vendor: %v", err)
	}
	if len(validProducts) != 1 {
		t.Fatalf("got more than one valid products instead of 1")
	}
	if len(invalidProducts) != 1 {
		t.Fatalf("got more than one invalid products instead of 1")
	}
	logger.Info("valid product returned", "id", validProducts[0].GetId())
	logger.Info("invalid product returned", "id", invalidProducts[0].GetId())
	if validProducts[0].Id != validProduct.Id {
		t.Fatalf("got wrong valid product")
	}
	if invalidProducts[0].Id != validProduct.Id {
		t.Fatalf("got wrong invalid product")
	}
	logger.Info("validated product helper for checking for invalid vendor")
}

func TestCheckPremiumOrEnterprise(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCheckPremiumOrEnterprise")
	idcProducts := make([]*pb.Product, 0)
	internalRates := []*pb.Rate{
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_INTEL,
			Rate:        defaultInternalAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   defaultUsageMetricExpression,
		},
	}
	premiumRates := []*pb.Rate{
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_PREMIUM,
			Rate:        defaultPremiumAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   defaultUsageMetricExpression,
		},
	}
	enterpriseRates := []*pb.Rate{
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_ENTERPRISE,
			Rate:        defaultEnterpriseAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   defaultUsageMetricExpression,
		},
	}
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	internalProductId := uuid.NewString()
	premiumProductId := uuid.NewString()
	enterpriseProductId := uuid.NewString()
	internalProduct := &pb.Product{
		Name:        computeProductVMSmallXeon3Name,
		Id:          internalProductId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       internalRates,
		MatchExpr:   computeProductVMSmallXeon3MatchExpr,
	}
	premiumProduct := &pb.Product{
		Name:        computeProductVMSmallXeon3Name,
		Id:          premiumProductId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       premiumRates,
		MatchExpr:   computeProductVMSmallXeon3MatchExpr,
	}
	enterpriseProduct := &pb.Product{
		Name:        computeProductVMSmallXeon3Name,
		Id:          enterpriseProductId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       enterpriseRates,
		MatchExpr:   computeProductVMSmallXeon3MatchExpr,
	}

	idcProducts = append(idcProducts, internalProduct)
	idcProducts = append(idcProducts, premiumProduct)
	idcProducts = append(idcProducts, enterpriseProduct)
	logger.Info("internal product id", "id", internalProduct.GetId())
	logger.Info("premium product", "id", premiumProduct.GetId())
	logger.Info("enterprise product", "id", enterpriseProduct.GetId())
	premiumOrEnterpriseProducts, err := CheckPremiumOrEnterprise(idcProducts)
	if err != nil {
		t.Fatalf("failed to check for premium or enterprise: %v", err)
	}
	if len(premiumOrEnterpriseProducts) != 2 {
		t.Fatalf("got not 2 premium or enterprise products")
	}
	logger.Info("validated product helper for checking for premium or enterprise")
}

func TestValidateProductsForMetaData(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestValidateProductsForMetaData")

	idcProducts := make([]*pb.Product, 0)
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	productId := uuid.NewString()

	// Product with valid metedata
	validProduct := GetProduct(computeProductVMSmallXeon3Name, productId, vendorId, idcComputeProductFamilyId, uuid.NewString(), computeProductVMSmallXeon3MatchExpr)

	// Product with invalid metedata
	invalidProduct := GetProduct(computeProductVMSmallXeon3Name, productId, vendorId, idcComputeProductFamilyId, uuid.NewString(), computeProductVMSmallXeon3MatchExpr)
	invalidProduct.Metadata = GetInvalidMetadata()

	idcProducts = append(idcProducts, validProduct)
	idcProducts = append(idcProducts, invalidProduct)

	// idcProducts contains valid and invalid product metadata
	validProducts, invalidProducts, err := ValidateProductForMetadata(idcProducts)
	if err != nil {
		t.Fatalf("failed to validate products for metadata: %v", err)
	}

	if len(validProducts) > 0 {
		for key, _ := range validProducts {
			logger.Info("Valid product returned with metadata : ", "Metadata", validProducts[key].Metadata)
		}
	}
	if len(invalidProducts) > 0 {
		for key, _ := range invalidProducts {
			logger.Info("Invalid product returned with metadata : ", "Metadata", invalidProducts[key].Metadata)
		}
	}

}

func TestValidateProductsForUsageMetric(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestValidateProductsForUsageMetric")
	idcProducts := make([]*pb.Product, 0)
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	ratesMatchingUsageMetric := []*pb.Rate{
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_PREMIUM,
			Rate:        defaultPremiumAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   defaultUsageMetricExpression,
		},
	}
	ratesNotMatchingUsageMetric := []*pb.Rate{
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_PREMIUM,
			Rate:        defaultPremiumAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   "SomeUnsupportedExpression",
		},
	}
	matchingExpressionProductId := uuid.NewString()
	notMatchingExpressionProductId := uuid.NewString()
	matchingUsageExpressionProduct := &pb.Product{
		Name:        computeProductVMSmallXeon3Name,
		Id:          matchingExpressionProductId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       ratesMatchingUsageMetric,
		MatchExpr:   computeProductVMSmallXeon3MatchExpr,
	}
	notMatchingUsageExpressionProduct := &pb.Product{
		Name:        computeProductVMSmallXeon3Name,
		Id:          notMatchingExpressionProductId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       ratesNotMatchingUsageMetric,
		MatchExpr:   computeProductVMSmallXeon3MatchExpr,
	}
	idcProducts = append(idcProducts, matchingUsageExpressionProduct)
	idcProducts = append(idcProducts, notMatchingUsageExpressionProduct)
	logger.Info("matching usage expression product id", "id", matchingUsageExpressionProduct.GetId())
	logger.Info("not matching usage expression product id", "id", notMatchingUsageExpressionProduct.GetId())
	validProducts, invalidProducts, err := ValidateProductsForUsageMetricType(idcProducts)
	if err != nil {
		t.Fatalf("failed to validate products for usage metric type: %v", err)
	}
	if len(validProducts) != 1 {
		t.Fatalf("got more than one valid products instead of 1")
	}
	if len(invalidProducts) != 1 {
		t.Fatalf("got more than one invalid products instead of 1")
	}
	logger.Info("valid product returned", "id", validProducts[0].GetId())
	logger.Info("invalid product returned", "id", invalidProducts[0].GetId())
	if validProducts[0].Id != matchingUsageExpressionProduct.Id {
		t.Fatalf("got wrong valid product")
	}
	if invalidProducts[0].Id != notMatchingUsageExpressionProduct.Id {
		t.Fatalf("got wrong invalid product")
	}
	logger.Info("validated product helper for checking for usage metric type")
}

func TestProductsOfUsageMetricType(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestProductsOfUsageMetricType")
	idcProducts := make([]*pb.Product, 0)
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	ratesMatchingUsageMetric := []*pb.Rate{
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_PREMIUM,
			Rate:        defaultPremiumAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   defaultUsageMetricExpression,
		},
	}
	ratesNotMatchingUsageMetric := []*pb.Rate{
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_PREMIUM,
			Rate:        defaultPremiumAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   "SomeUnsupportedExpression",
		},
	}
	matchingExpressionProductId := uuid.NewString()
	notMatchingExpressionProductId := uuid.NewString()
	matchingUsageExpressionProduct := &pb.Product{
		Name:        computeProductVMSmallXeon3Name,
		Id:          matchingExpressionProductId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       ratesMatchingUsageMetric,
		MatchExpr:   computeProductVMSmallXeon3MatchExpr,
	}
	notMatchingUsageExpressionProduct := &pb.Product{
		Name:        computeProductVMSmallXeon3Name,
		Id:          notMatchingExpressionProductId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       ratesNotMatchingUsageMetric,
		MatchExpr:   computeProductVMSmallXeon3MatchExpr,
	}
	idcProducts = append(idcProducts, matchingUsageExpressionProduct)
	idcProducts = append(idcProducts, notMatchingUsageExpressionProduct)
	logger.Info("matching usage expression product id", "id", matchingUsageExpressionProduct.GetId())
	logger.Info("not matching usage expression product id", "id", notMatchingUsageExpressionProduct.GetId())
	productsMatchingUsageExpr, err := GetProductsOfUsageMetricType(idcProducts, "time – previous.time")
	if err != nil {
		t.Fatalf("failed to get products matching usage expression: %v", err)
	}
	if len(productsMatchingUsageExpr) != 1 {
		t.Fatalf("got more than one product matching the expression instead of 1")
	}
	logger.Info("matching product returned", "id", productsMatchingUsageExpr[0].GetId())
	if productsMatchingUsageExpr[0].Id != matchingUsageExpressionProduct.Id {
		t.Fatalf("Got wrong valid product")
	}
	logger.Info("validated product helper for getting the product matching the metric type")
}
