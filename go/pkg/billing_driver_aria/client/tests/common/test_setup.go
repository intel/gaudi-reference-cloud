// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package common

import (
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	ariaClient          *client.AriaClient
	ariaAdminClient     *client.AriaAdminClient
	ariaCredentials     *client.AriaCredentials
	ariaServiceClient   *client.AriaServiceClient
	ariaAccountClient   *client.AriaAccountClient
	serviceCreditClient *client.ServiceCreditClient
	ariaPlanClient      *client.AriaPlanClient
	ariaPaymentClient   *client.AriaPaymentClient
	ariaUsageTypeClient *client.AriaUsageTypeClient
	ariaUsageClient     *client.AriaUsageClient
	promoClient         *client.PromoClient
	ariaInvoiceClient   *client.AriaInvoiceClient
	product             *pb.Product
	productFamily       *pb.ProductFamily
)

var insecureSsl = false

// The purpose of this method is to do exactly that the server would do when it comes up.
func Init() error {
	err := config.InitTestConfig()
	if err != nil || config.Cfg.AriaSystem.AuthKey == "" {
		// We can get rid of this environment variable check once
		// all the tests pass with the mock server
		useMock, ok := os.LookupEnv("IDC_ARIA_USE_MOCK_SERVER")
		if ok && useMock == "true" {
			err := client.MockServer()
			if err != nil {
				return err
			}
			err = nil
		}
	}
	if err != nil {
		return err
	}

	// The values for configuring the Aria admin client will be loaded from configuration.
	ariaAdminClient = client.NewAriaAdminClient(config.Cfg.GetAriaSystemServerUrlAdminToolsApi(), insecureSsl)
	// The values for configuring the Aria client will be loaded from configuration.
	ariaClient = client.NewAriaClient(config.Cfg.GetAriaSystemServerUrlCoreApi(), config.Cfg.GetAriaSystemCoreApiSuffix(), insecureSsl)
	// Credentials will come from the secrets service.
	ariaCredentials = client.NewAriaCredentials(config.Cfg.GetAriaSystemClientNo(), config.Cfg.GetAriaSystemAuthKey())
	// All the clients for Aria will be initialized using the HTTP clients. Hence, all aspects with respect to
	// healthy clients and configured clients is applied once to all Aria API clients!!!
	// Same goes for credentials - Credentials are not distributed everywhere. Credentials everywhere is a nightmare.
	// It does however put the question of how to rotate credentials :-) It is simple, just like credentials are
	// initialized at the time of creation of these clients. When rotated, these clients will provide methods to update
	// the credentials with new credentials.
	// Alternatively, when calls fail due to bad credentials as called by a caller, the caller will call get new
	// credentials and reapply!! Simple!
	ariaServiceClient, err = client.NewAriaServiceClient(ariaAdminClient, ariaCredentials)
	if err != nil {
		return err
	}

	ariaPlanClient = client.NewAriaPlanClient(config.Cfg, ariaAdminClient, ariaClient, ariaCredentials)
	ariaAccountClient = client.NewAriaAccountClient(ariaClient, ariaCredentials)
	serviceCreditClient = client.NewServiceCreditClient(ariaClient, ariaCredentials)
	ariaPaymentClient = client.NewAriaPaymentClient(ariaClient, ariaCredentials)
	ariaUsageTypeClient = client.NewAriaUsageTypeClient(ariaAdminClient, ariaCredentials)
	ariaUsageClient = client.NewAriaUsageClient(ariaClient, ariaCredentials)
	promoClient = client.NewPromoClient(ariaAdminClient, ariaCredentials)
	ariaInvoiceClient = client.NewAriaInvoiceClient(ariaClient, ariaCredentials)
	return nil
}

func GetAriaClient() *client.AriaClient {
	return ariaClient
}

func GetAriaAdminClient() *client.AriaAdminClient {
	return ariaAdminClient
}

func GetAriaCredentials() *client.AriaCredentials {
	return ariaCredentials
}

func GetAriaServiceClient() *client.AriaServiceClient {
	return ariaServiceClient
}

func GetAriaAccountClient() *client.AriaAccountClient {
	return ariaAccountClient
}

func GetServiceCreditClient() *client.ServiceCreditClient {
	return serviceCreditClient
}

func GetAriaUsageTypeClient() *client.AriaUsageTypeClient {
	return ariaUsageTypeClient
}

func GetAriaPlanClient() *client.AriaPlanClient {
	return ariaPlanClient
}

func GetTestClientId(id string) string {
	return config.Cfg.ClientIdPrefix + "." + id
}

func GetAriaPaymentClient() *client.AriaPaymentClient {
	return ariaPaymentClient
}

func GetPromoClient() *client.PromoClient {
	return promoClient
}

func GetAriaInvoiceClient() *client.AriaInvoiceClient {
	return ariaInvoiceClient
}

func GetUsageClient() *client.AriaUsageClient {
	return ariaUsageClient
}
