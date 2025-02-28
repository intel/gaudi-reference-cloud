// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func GetRates() []*pb.Rate {

	rates := []*pb.Rate{
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_STANDARD,
			Rate:        defaultStandardAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   defaultUsageMetricExpression,
		},
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_INTEL,
			Rate:        defaultInternalAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   defaultUsageMetricExpression,
		},
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_PREMIUM,
			Rate:        defaultPremiumAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   defaultUsageMetricExpression,
		},
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_ENTERPRISE,
			Rate:        defaultPremiumAccountRate,
			Unit:        defaultRateUnit,
			UsageExpr:   defaultUsageMetricExpression,
		},
	}
	return rates
}

func GetComputeMetadata(serviceName string, displayName string, region string, instanceType string) map[string]string {

	metadata := map[string]string{
		"instanceType": instanceType,
		"region":       region,
		"service":      serviceName,
		"category":     productCategory,
		"displayName":  displayName,
	}

	return metadata
}

func GetStorageMetadata(serviceName string, displayName string, region string) map[string]string {

	metadata := map[string]string{
		"region":      region,
		"service":     serviceName,
		"category":    productCategory,
		"displayName": displayName,
	}

	return metadata
}

func GetComputeProduct(name string, id string, vendorId string,
	familyId string, description string, matchExpr string,
	serviceName string, displayName string, region string, instanceType string) *pb.Product {

	product := &pb.Product{
		Name:        name,
		Id:          id,
		VendorId:    vendorId,
		FamilyId:    familyId,
		Description: description,
		Rates:       GetRates(),
		MatchExpr:   matchExpr,
		Metadata:    GetComputeMetadata(serviceName, displayName, region, instanceType),
	}
	return product
}

func GetStorageProduct(name string, id string, vendorId string,
	familyId string, description string, matchExpr string,
	serviceName string, displayName string, region string) *pb.Product {

	product := &pb.Product{
		Name:        name,
		Id:          id,
		VendorId:    vendorId,
		FamilyId:    familyId,
		Description: description,
		Rates:       GetRates(),
		MatchExpr:   matchExpr,
		Metadata:    GetStorageMetadata(serviceName, displayName, region),
	}
	return product
}

func GetIdcComputeProductFamily(id string) *pb.ProductFamily {
	productFamily := &pb.ProductFamily{
		Name:        idcComputeProductFamilyName,
		Id:          id,
		Description: idcComputeProductFamilyDescription,
	}
	return productFamily
}

func GetIdcStorageProductFamily(id string) *pb.ProductFamily {
	productFamily := &pb.ProductFamily{
		Name:        idcStorageProductFamilyName,
		Id:          id,
		Description: idcStorageProductFamilyDescription,
	}
	return productFamily
}

func GetIdcVendor(id string, productFamilies []*pb.ProductFamily) *pb.Vendor {
	vendor := &pb.Vendor{
		Name:        idcVendorName,
		Id:          id,
		Description: idcVendorDescription,
		Families:    productFamilies,
	}
	return vendor
}

func GetIdcProductsAndVendors() ([]*pb.Product, []*pb.Vendor) {
	idcComputeProductFamilies := make([]*pb.ProductFamily, 0)
	idcProducts := make([]*pb.Product, 0)
	idcVendors := make([]*pb.Vendor, 0)
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	idcComputeProductFamily := GetIdcComputeProductFamily(idcComputeProductFamilyId)
	idcComputeProductFamilies = append(idcComputeProductFamilies, idcComputeProductFamily)
	vendor := GetIdcVendor(vendorId, idcComputeProductFamilies)
	xeon3SmallProductId := uuid.NewString()
	xeon3SmallProduct := GetComputeProduct(computeProductVMSmallXeon3Name,
		xeon3SmallProductId, vendorId, idcComputeProductFamilyId, "someDescription", computeProductVMSmallXeon3MatchExpr,
		idcComputeServiceName, xeon3DisplayName, DefaultServiceRegion, xeon3SmallInstanceType)
	xeon3LargeProductId := uuid.NewString()
	xeon3LargeProduct := GetComputeProduct(computeProductVMLargeXeon3Name,
		xeon3LargeProductId, vendorId, idcComputeProductFamilyId, "someDescription", computeProductVMLargeXeon3MatchExpr,
		idcComputeServiceName, xeon3DisplayName, DefaultServiceRegion, xeon3LargeInstanceType)
	idcVendors = append(idcVendors, vendor)
	idcProducts = append(idcProducts, xeon3SmallProduct)
	idcProducts = append(idcProducts, xeon3LargeProduct)
	return idcProducts, idcVendors
}

func BuildStaticProductsAndVendors() ([]*pb.Product, []*pb.Vendor, []*pb.ProductFamily) {
	idcProductFamilies := make([]*pb.ProductFamily, 0)
	idcProducts := make([]*pb.Product, 0)
	idcVendors := make([]*pb.Vendor, 0)
	vendorId := "static-vendor-id"
	idcComputeProductFamilyId := "static-compute-product-family-id"
	idcStorageProductFamilyId := "static-storage-product-family-id"
	idcComputeProductFamily := GetIdcComputeProductFamily(idcComputeProductFamilyId)
	idcStorageProductFamily := GetIdcStorageProductFamily(idcStorageProductFamilyId)
	idcProductFamilies = append(idcProductFamilies, idcComputeProductFamily)
	idcProductFamilies = append(idcProductFamilies, idcStorageProductFamily)
	vendor := GetIdcVendor(vendorId, idcProductFamilies)
	//xeon3SmallProductId := "static-xeon3-small-productId"
	xeon3SmallProduct := GetComputeProduct(computeProductVMSmallXeon3Name,
		"xeon3-small-product-id", vendorId, idcComputeProductFamilyId, "someDescription", computeProductVMSmallXeon3MatchExpr,
		idcComputeServiceName, xeon3DisplayName, DefaultServiceRegion, xeon3SmallInstanceType)
	//xeon3LargeProductId := "static-xeon3-large-productId"
	xeon3LargeProduct := GetComputeProduct(computeProductVMLargeXeon3Name,
		"xeon3-large-product-id", vendorId, idcComputeProductFamilyId, "someDescription", computeProductVMLargeXeon3MatchExpr,
		idcComputeServiceName, xeon3DisplayName, DefaultServiceRegion, xeon3LargeInstanceType)
	fileStorageProduct := GetStorageProduct(fileStorageProductName,
		"file-storage-product-id", vendorId, idcStorageProductFamilyId, "someDescription", fileStorageProductMatchExpr,
		FileStorageServiceType, fileStorageDisplayName, DefaultServiceRegion)
	objectStorageProduct := GetStorageProduct(objectStorageProductName,
		"object-storage-product-id", vendorId, idcStorageProductFamilyId, "someDescription", objectStorageProductMatchExpr,
		ObjectStorageServiceType, objectStorageDisplayName, DefaultServiceRegion)
	idcVendors = append(idcVendors, vendor)
	idcProducts = append(idcProducts, xeon3SmallProduct)
	idcProducts = append(idcProducts, xeon3LargeProduct)
	idcProducts = append(idcProducts, fileStorageProduct)
	idcProducts = append(idcProducts, objectStorageProduct)
	return idcProducts, idcVendors, idcProductFamilies
}
