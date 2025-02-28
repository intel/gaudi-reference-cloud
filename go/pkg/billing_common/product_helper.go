// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"errors"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	ProductInvalidProductFamily   string = "invalid product family"
	ProductInvalidVendor          string = "invalid vendor"
	ProductInvalidUsageExpression string = "missing usage expression"
	ProductInvalidUsage           string = "invalid usage"
	ProductCatalogBMaaSService    string = "BMaaS"
	ProductCatalogVMaaSService    string = "VMaaS"
	ProductCatalogCPUXeon3        string = "xeon3"
	ProductCatalogCPUXeon4        string = "xeon4"
)

type ProductValidationError struct {
	Product *pb.Product
	Err     error
}

// Add all helper methods that are needed across billing and drivers related to product catalog.
func ValidateProductsForProductFamily(inVendors []*pb.Vendor, inProducts []*pb.Product) ([]*pb.Product, []*pb.Product, error) {
	var validProducts []*pb.Product
	var invalidProducts []*pb.Product
	vendorMap := make(map[string]*pb.ProductFamily)
	for _, vendor := range inVendors {
		for _, productFamily := range vendor.GetFamilies() {
			vendorMap[productFamily.GetId()] = productFamily
		}
	}
	for _, product := range inProducts {
		_, productEntryIsValid := vendorMap[product.GetFamilyId()]
		if !productEntryIsValid {
			invalidProducts = append(invalidProducts, product)
		} else {
			validProducts = append(validProducts, product)
		}
	}
	return validProducts, invalidProducts, nil
}

func ValidateProductsForVendors(inVendors []*pb.Vendor, inProducts []*pb.Product) ([]*pb.Product, []*pb.Product, error) {
	if inVendors == nil {
		return nil, nil, errors.New("vendors cannot be nil")
	}
	if inProducts == nil {
		return nil, nil, errors.New("products cannot be nil")
	}
	var validProducts []*pb.Product
	var invalidProducts []*pb.Product
	for _, product := range inProducts {
		vendorCount := 0
		for _, vendor := range inVendors {
			if product.GetVendorId() == vendor.Id {
				vendorCount += 1
			}
		}
		if vendorCount == 1 {
			validProducts = append(validProducts, product)
		} else {
			invalidProducts = append(invalidProducts, product)
		}
	}
	return validProducts, invalidProducts, nil
}

func CheckPremiumOrEnterprise(inProducts []*pb.Product) ([]*pb.Product, error) {
	if inProducts == nil {
		return nil, errors.New("products cannot be nil")
	}
	var premiumOrEnterpriseProducts []*pb.Product
	for _, product := range inProducts {
		rates := product.GetRates()
		for _, rate := range rates {
			// Check if the product has rates for enterprise or premium
			if (rate.AccountType == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE) ||
				(rate.AccountType == pb.AccountType_ACCOUNT_TYPE_PREMIUM) {
				premiumOrEnterpriseProducts = append(premiumOrEnterpriseProducts, product)
				// Todo: Identify how not to have to break in a iteration for this.
				break
			}
		}
	}
	return premiumOrEnterpriseProducts, nil
}

func ValidateProductForMetadata(inProducts []*pb.Product) ([]*pb.Product, []*pb.Product, error) {
	logger := log.FromContext(context.Background()).WithName("ValidateProductForMetadata")
	var validProducts []*pb.Product
	var invalidProducts []*pb.Product

	for _, product := range inProducts {
		validMetric := true

		if _, exists := product.Metadata["service"]; !exists {
			validMetric = false
			err := errors.New(" \"service\" key does not exists in Product's Metadata")
			logger.Error(err, "")
		}

		if _, exists := product.Metadata["instanceType"]; !exists {
			validMetric = false
			err := errors.New(" \"instanceType\" key does not exists in Product's Metadata")
			logger.Error(err, "")
		}

		for key, metadata := range product.Metadata {
			// Currently its checking only for the emptiness of metadata. This can be improvized for checking specific value in the metadata.
			// Non empty metadata -> valid product
			// Empty metadata -> invalid product
			if metadata == "" {
				validMetric = false
				err := "Value of key: \"" + key + "\" is empty"
				logger.Error(errors.New(err), "")
			}
		}
		if validMetric {
			validProducts = append(validProducts, product)
		} else {
			invalidProducts = append(invalidProducts, product)
		}
	}

	return validProducts, invalidProducts, nil
}

func ValidateProductsForUsageMetricType(inProducts []*pb.Product) ([]*pb.Product, []*pb.Product, error) {
	if inProducts == nil {
		return nil, nil, errors.New("products cannot be nil")
	}
	var validProducts []*pb.Product
	var invalidProducts []*pb.Product
	for _, product := range inProducts {
		// This can lead to false positives and hence it is important to check for every condition that will set it to false in the if conditions.
		validMetric := true
		for _, rate := range product.Rates {
			if rate.UsageExpr != SupportedMetricType {
				validMetric = false
			}
		}
		if validMetric {
			validProducts = append(validProducts, product)
		} else {
			invalidProducts = append(invalidProducts, product)
		}
	}
	return validProducts, invalidProducts, nil
}

func GetProductsOfUsageMetricType(inProducts []*pb.Product, usageMetricExpression string) ([]*pb.Product, error) {
	if inProducts == nil {
		return nil, errors.New("products cannot be nil")
	}
	var productsOfUsageMetricType []*pb.Product
	for _, product := range inProducts {
		for _, rates := range product.Rates {
			// Todo: When matches the metric, the product matches.
			if rates.UsageExpr == usageMetricExpression {
				productsOfUsageMetricType = append(productsOfUsageMetricType, product)
				break
			}
		}
	}
	return productsOfUsageMetricType, nil
}

func ValidateProducts(ctx context.Context, vendors []*pb.Vendor, products []*pb.Product) ([]*pb.Product, []*ProductValidationError, error) {
	logger := log.FromContext(ctx).WithName("ProductHelper.getValidProducts")
	// check the products for the product family
	validProductsForProdFamily, invalidProductsForProdFamily, err := ValidateProductsForProductFamily(vendors, products)
	// fail fast if cannot validate products for product family
	if err != nil {
		logger.Error(err, "failed to validate products for product family")
		return nil, nil, err
	}
	var productValidationErrors []*ProductValidationError
	// populate errors for invalid products for product family
	for _, invalidProduct := range invalidProductsForProdFamily {
		logger.Info("invalid product", "name", invalidProduct.Name)
		productValidationErrors = append(productValidationErrors, &ProductValidationError{Product: invalidProduct, Err: errors.New(ProductInvalidProductFamily)})
	}
	// check the products for vendor
	validProductsForVendor, invalidProductsForVendor, err := ValidateProductsForVendors(vendors, validProductsForProdFamily)
	// fail fast if cannot validate products for vendors
	if err != nil {
		logger.Error(err, "failed to validate products for vendors")
		return nil, productValidationErrors, err
	}

	// populate errors for invalid products for product family
	for _, invalidProduct := range invalidProductsForVendor {
		logger.Info("invalid product", "name", invalidProduct.Name)
		productValidationErrors = append(productValidationErrors, &ProductValidationError{Product: invalidProduct, Err: errors.New(ProductInvalidVendor)})
	}
	// check the products for usage expression
	//productsWithValidUsageExpression, productsWithInvalidUsageExpression, err := ValidateProductsForUsageMetricType(validProductsForVendor)
	//if err != nil {
	//	logger.Error(err, "failed to validate products for usage expression")
	//	return nil, productValidationErrors, err
	//}
	//for _, productWithInvalidUsageExpression := range productsWithInvalidUsageExpression {
	//	productValidationErrors = append(productValidationErrors, &ProductValidationError{Product: productWithInvalidUsageExpression, Err: errors.New(ProductInvalidUsageExpression)})
	//}
	// To get products for enterprise or premium tiers.
	//enterpriseOrPremiumProducts, err := CheckPremiumOrEnterprise(productsWithValidUsageExpression)
	// fail fast if cannot check for product tiers.
	//if err != nil {
	//	logger.Error(err, "failed to check for product tiers")
	//	return nil, productValidationErrors, err
	//}

	// Todo: "time - previous.time needs to be a configurable parameter"
	// Todo: Use the returned products of a metric type to evaluate usage based on the metric type.
	//_, err = GetProductsOfUsageMetricType(productsWithValidUsageExpression, SupportedMetricType)
	// fail fast if cannot check for product tiers.
	//if err != nil {
	//	logger.Error(err, "failed to check for product of usage metric type of used mins")
	//	return nil, productValidationErrors, err
	//}

	// check the products metadata
	// Checking if Product's Metadata has "service and instanceType" keys and other keys values for non emptyness
	// For now, it's just logging an error but not failing.

	// TODO: Fail if metadata is not valid (If service or instanceType key is missing and if values for any metadata keys are empty)
	//validProductsMetadata, invalidProductsMetadata, err := ValidateProductForMetadata(validProductsForProdFamily)
	//if err != nil {
	//	logger.Error(err, "failed to validate products metadata")
	//}

	//if len(validProductsMetadata) > 0 {
	//	for key, _ := range validProductsMetadata {
	//		logger.Info("Valid product returned with metadata : ", "Metadata", validProductsMetadata[key].Metadata)
	//	}
	//}
	//if len(invalidProductsMetadata) > 0 {
	//	for key, _ := range invalidProductsMetadata {
	//		logger.Info("Invalid product returned with metadata : ", "Metadata", invalidProductsMetadata[key].Metadata)
	//	}
	//}

	return validProductsForVendor, productValidationErrors, nil

}
