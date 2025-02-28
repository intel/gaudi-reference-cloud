// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	// Make sure the rates for the tiers are different.
	defaultInternalAccountRate   = "2"
	defaultPremiumAccountRate    = "3"
	defaultEnterpriseAccountRate = "4"

	// Only one usage expression supported which is the default.
	defaultUsageMetricExpression = "time – previous.time"

	// Only one rate unit supported which is the default.
	defaultRateUnit                        = pb.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE
	idcComputeProductFamilyName            = "idc-compute"
	idcComputeProductFamilyDescription     = "Intel developer cloud compute services"
	idcVendorName                          = "intel"
	idcVendorDescription                   = "Intel developer cloud provider"
	computeProductVMSmallMatchExpr         = "billUsage && service == compute && (instanceType == vm.xeon3.small || instanceType == vm.xeon4.small)"
	computeProductVMLargeMatchExpr         = "billUsage && service == compute && (instanceType == vm.xeon3.large || instanceType == vm.xeon4.large)"
	computeProductVMSmallXeon3Name         = "compute-xeon3-small-vm"
	computeProductVMSmallXeon3MatchExpr    = "billUsage && service == compute && instanceType == vm.xeon3.small"
	computeProductVMSmallXeon4Name         = "compute-xeon4-small-vm"
	computeProductVMSmallXeon4MatchExpr    = "billUsage && service == compute && instanceType == vm.xeon4.small"
	defaultServiceRegion                   = "hillsboro-1"
	idcComputeServiceName                  = "compute"
	xeon3SmallInstanceType                 = "vm.xeon3.small"
	requestRegion                          = "us-east"
	productCatalogBMaaSMatchExpr           = "service == \"BMaaS\" && cpu == \"xeon3\" && mem == \"256G\""
	invalidProductCatalogBMaaSMatchExpr    = "service == \"Unknown\" && cpu == \"Unknown\" && mem == "
	productCatalogVMaaSMatchExpr           = "service == \"VMaaS\" && cpu == \"xeon4\" && mem == \"8G\""
	productCategory                        = "Released"
	cpuSockets                             = "2"
	diskSize                               = "2TB"
	displayName                            = "Intel® Max Series GPU (PVC) on 4th Gen Intel® Xeon® processors - 1100 series (1x)"
	memorySize                             = "256GB"
	processorType                          = "GPU"
	IdcFileStorageProductName              = "storage-file"
	idcFileStorageProductFamilyName        = "storage-file"
	idcFileStorageProductFamilyDescription = "Storage Service - Filesystem"
	IdcFileStroageProductMatchExpr         = "serviceType == FileStorageAsAService"
	idcFileStorageInstanceType             = "storage-file"
	idcFileStorageServiceName              = "Storage Service - File"
	idcFileStorageDisplayName              = "Filesystem Storage Service"
	releaseStatus                          = "Released"
)

func GetRates() []*pb.Rate {

	rates := []*pb.Rate{
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

func GetMetadata() map[string]string {

	metadata := map[string]string{
		"instanceType": xeon3SmallInstanceType,
		"region":       requestRegion,
		"service":      idcComputeServiceName,
		"category":     productCategory,
		"cpu.sockets":  cpuSockets,
		"disks.size":   diskSize,
		"displayName":  displayName,
		"memory.size":  memorySize,
		"processor":    processorType,
	}

	return metadata
}

func GetInvalidMetadata() map[string]string {

	// Invalid metadata because "service" and "instanceType" key is missing which are mandatory
	// Also, none of the value should be empty, here "region" key has empty value
	metadata := map[string]string{
		"region":      "",
		"category":    productCategory,
		"cpu.sockets": cpuSockets,
		"disks.size":  diskSize,
		"displayName": displayName,
		"memory.size": memorySize,
		"processor":   processorType,
	}

	return metadata
}

func GetProduct(name string, id string, vendorId string, familyId string, description string, matchExpr string) *pb.Product {

	product := &pb.Product{
		Name:        name,
		Id:          id,
		VendorId:    vendorId,
		FamilyId:    familyId,
		Description: description,
		Rates:       GetRates(),
		MatchExpr:   matchExpr,
		Metadata:    GetMetadata(),
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

func GetIdcVendor(id string, productFamilies []*pb.ProductFamily) *pb.Vendor {
	vendor := &pb.Vendor{
		Name:        idcVendorName,
		Id:          id,
		Description: idcVendorDescription,
		Families:    productFamilies,
	}
	return vendor
}

func GetFileStorageMetadata() map[string]string {

	metadata := map[string]string{
		"instanceType":    idcFileStorageInstanceType,
		"region":          requestRegion,
		"service":         idcFileStorageServiceName,
		"releaseStatus":   releaseStatus,
		"volume.size.min": "5",
		"volume.size.max": "2000",
		"displayName":     idcFileStorageDisplayName,
		"processor":       processorType,
	}

	return metadata
}

func GetIdcFileStorageProduct(name string, id string, vendorId string, familyId string, description string, matchExpr string) *pb.Product {
	product := &pb.Product{
		Name:        name,
		Id:          id,
		VendorId:    vendorId,
		FamilyId:    familyId,
		Description: description,
		Rates:       GetRates(),
		MatchExpr:   IdcFileStroageProductMatchExpr,
		Metadata:    GetFileStorageMetadata(),
	}
	return product
}

func GetIdcStorageProductFamily(id string) *pb.ProductFamily {
	productFamily := &pb.ProductFamily{
		Name:        idcFileStorageProductFamilyName,
		Id:          id,
		Description: idcFileStorageProductFamilyDescription,
	}
	return productFamily
}
