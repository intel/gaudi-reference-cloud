// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"unicode"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var maxId int64 = 1_000_000_000_000

// Public for testing purposes
func NewCloudAcctId() (string, error) {
	intId, err := rand.Int(rand.Reader, big.NewInt(maxId))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%012d", intId), nil
}

func MustNewCloudAcctId() string {
	id, err := NewCloudAcctId()
	if err != nil {
		panic(err)
	}
	return id
}

func IsValidCloudAcctId(id string) bool {
	if len(id) != 12 {
		return false
	}

	for _, ch := range id {
		if !unicode.IsDigit(ch) {
			return false
		}
	}
	return true
}

func CheckValidCloudAcctId(id string) error {
	if !IsValidCloudAcctId(id) {
		//CloudAccountId should be string with exactly 12 digits
		return status.Error(codes.InvalidArgument, "invalid CloudAccountId")
	}
	return nil
}

const (
	// Make sure the rates for the tiers are different.
	defaultStandardAccountRate   = "2"
	defaultInternalAccountRate   = "2"
	defaultPremiumAccountRate    = "3"
	defaultEnterpriseAccountRate = "4"

	// Only one usage expression supported which is the default.
	defaultUsageMetricExpression = "time – previous.time"

	// Only one rate unit supported which is the default.
	defaultRateUnit                     = pb.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE
	idcComputeProductFamilyName         = "idc-compute"
	idcStorageProductFamilyName         = "idc-storage"
	idcComputeProductFamilyDescription  = "Intel developer cloud compute services"
	idcStorageProductFamilyDescription  = "Intel developer cloud storage services"
	idcVendorName                       = "intel"
	idcVendorDescription                = "Intel developer cloud provider"
	computeProductVMSmallXeon3Name      = "compute-xeon3-small-vm"
	fileStorageProductName              = "file-storage-product"
	objectStorageProductName            = "object-storage-product"
	computeProductVMLargeXeon3Name      = "compute-xeon3-large-vm"
	computeProductVMSmallXeon3MatchExpr = "service == compute && instanceType == vm.xeon3.small"
	computeProductVMLargeXeon3MatchExpr = "service == compute && instanceType == vm.xeon3.large"
	fileStorageProductMatchExpr         = "service == " + FileStorageServiceType
	objectStorageProductMatchExpr       = "service == " + ObjectStorageServiceType
	DefaultServiceRegion                = "us-west-1"
	idcComputeServiceName               = "compute"
	xeon3SmallInstanceType              = "vm.xeon3.small"
	xeon3LargeInstanceType              = "vm.xeon3.large"
	productCategory                     = "Released"
	xeon3DisplayName                    = "Intel® Max Series GPU (PVC) on 4th Gen Intel® Xeon® processors - 1100 series (1x)"
	fileStorageDisplayName              = "Intel® File storage"
	objectStorageDisplayName            = "Intel® Object storage"
	idcFileStorageServiceName           = "Storage Service - Filesystem"
	idcFileStorageServiceType           = "FileStorageServiceType"
	idcFileStorageInstanceType          = "storage-file"
)

func GetXeon3SmallInstanceType() string {
	return xeon3SmallInstanceType
}

func GetIdcComputeServiceName() string {
	return idcComputeServiceName
}

func GetIdcFileStorageInstanceType() string {
	return idcFileStorageInstanceType
}

func GetIdcFileStoragServiceType() string {
	return idcFileStorageServiceType
}

func GetIdcFileStorageServiceName() string {
	return idcFileStorageServiceName
}
