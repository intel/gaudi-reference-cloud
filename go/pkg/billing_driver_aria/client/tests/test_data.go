// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	ERROR_CODE_OK    = 0
	ERROR_MESSAGE_OK = "OK"
)

var (
	newCreditsExpirationDate            = time.Now().AddDate(0, 0, 100)
	DefaultCreditExpirationDate         = fmt.Sprintf("%04d-%02d-%02d", newCreditsExpirationDate.Year(), newCreditsExpirationDate.Month(), newCreditsExpirationDate.Day())
	preConfiguredPlansPromoCode         = "IDC_Plans"
	rate                        float64 = 0
	productName                         = "Test Product/Plan" + uuid.New().String()[:10]
	productPCQ                          = "Test PCQ"
	productFamilyName                   = "Test Service Plan" + uuid.New().String()[:10]
	productFamilyId                     = "Test Service Plan Id" + uuid.New().String()[:10]
	DefaultCloudCreditAmount            = float64(1000)
	DefaultUsageAmount                  = float64(50)
	DefaultUsageRecordId                = "1"
	DefaultCommentsForCredits           = "InitialCredits"
)

const kFixMeWeHaventImplementedReasonCodeYet int64 = 1

func GetClientAccountId() string {
	return client.GetAccountClientId(cloudaccount.MustNewId())
}

func GetProduct() *pb.Product {

	pbrate := []*pb.Rate{
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_PREMIUM,
			Rate:        ".05",
		},
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_ENTERPRISE,
			Rate:        ".01",
		},
		{
			AccountType: pb.AccountType_ACCOUNT_TYPE_INTEL,
			Rate:        ".01",
		},
	}

	product := &pb.Product{
		Name:  productName,
		Id:    uuid.New().String(),
		Pcq:   productPCQ,
		Rates: pbrate,
	}

	product.Metadata = make(map[string]string)
	product.Metadata["displayName"] = "Testing Service Name"
	return product
}

func GetProductFamily() *pb.ProductFamily {

	productFamily := &pb.ProductFamily{
		Name: productFamilyName,
		Id:   productFamilyId,
	}

	return productFamily
}

// todo: why do we use this - check with Amardeep.
func GetClientPlanId() string {
	return client.GetPlanClientId(uuid.NewString())
}

func GetDefaultClientPlanId() string {
	return "idc.master"
}

func GetCloudAcctIdFromClientAcctId(clientAccountId string) string {
	return clientAccountId[len(config.Cfg.ClientIdPrefix)+1:]
}

func GetProdIdFromClientPlanId(clientPlanId string) string {
	return clientPlanId[len(config.Cfg.ClientIdPrefix)+1:]
}
