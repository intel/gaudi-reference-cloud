//go:build Functional || authn_regression || authn_positive
// +build Functional authn_regression authn_positive

package AuthenticationTest

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/metering"
	"goFramework/framework/library/financials"
	"goFramework/framework/library/vmaas/serviceapi"
	"goFramework/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

func (suite *authNTestSuite) TestComputeAPIWithAuthN() {
	logger.Log.Info("Compute api GET call - positive flow with valid token")

	//Fetching and populating the input from config file
	compute_endpoint_url := utils.GetBaseUrl() + gjson.Get(utils.GetComputeApiData(), "endpoint").String()

	//Get the all instance types supported
	logger.Log.Info("Get instance types with valid token via Compute API - GET method")
	_, get_response_status := serviceapi.GetAllInstanceType(compute_endpoint_url, 200)
	assert.Equal(suite.T(), get_response_status == 200, true, "Test Failed to perform GET api with AuthN enabled on compute api")
	logger.Log.Info("Compute API service is validated with Authentication successfully!!!")
}

func (suite *authNTestSuite) TestMeteringServiceWithAuthN() {
	logger.Log.Info("Metering api GET call - positive flow with valid token")

	//Fetching and populating the input from config file
	metering_endpoint_url := utils.GetBaseUrl() + gjson.Get(utils.GetMeteringApiData(), "endpoint").String()
	resourceId := gjson.Get(utils.GetMeteringApiData(), "resourceId").String()

	//Get the previous metering record with test resource id
	logger.Log.Info("Get instance types with valid token via Compute API - GET method")
	_, get_response_status := metering.GetMeteringRecordsWithParams(metering_endpoint_url, resourceId, 200)
	assert.Equal(suite.T(), get_response_status == 200, true, "Test Failed to perform GET api with AuthN enabled on metering api")
	logger.Log.Info("Metering API service is validated with Authentication successfully!!!")
}

func (suite *authNTestSuite) TestProductCatalogServiceWithAuthN() {
	logger.Log.Info("ProductCatalog api GET vendors call - positive flow with valid token")

	//Fetching and populating the input from config file
	productcatalog_endpoint_url := utils.GetBaseUrl() + gjson.Get(utils.GetProductCatalogApiData(), "endpoint").String()

	//Get available vendors
	logger.Log.Info("Get vendors with valid token via API - GET method")
	_, get_response_status := financials.GetVendorsWithParams(productcatalog_endpoint_url,"noFilters", 200)
	assert.Equal(suite.T(), get_response_status == 200, true, "Test Failed to perform GET api with AuthN enabled on Product Catalog api")
	logger.Log.Info("Product Catalog API service is validated with Authentication successfully!!!")
}

func (suite *authNTestSuite) TestCloudAccountServiceWithAuthN() {
	logger.Log.Info("CloudAccount api GET call - positive flow with valid token")

	//Fetching and populating the input from config file
	cloudaccount_endpoint_url := utils.GetBaseUrl() + gjson.Get(utils.GetCloudAccountApiData(), "endpoint").String()

	//Get the cloud accounts by type
	logger.Log.Info("Get CloudAccount by type with valid token via cloudaccounts API - GET method")
	_, get_response_status := financials.Get_CAcc_by_type(cloudaccount_endpoint_url,"ACCOUNT_TYPE_PREMIUM", 200)
	assert.Equal(suite.T(), get_response_status == 200, true, "Test Failed to perform GET api with AuthN enabled on CloudAccount api")
	logger.Log.Info("Cloud Account API service is validated with Authentication successfully!!!")
}

func (suite *authNTestSuite) TestBillingServiceWithAuthN() {
	logger.Log.Info("Billing api GET coupons call - positive flow with valid token")

	//Fetching and populating the input from config file
	billing_endpoint_url := utils.GetBaseUrl() + gjson.Get(utils.GetBillingApiData(), "endpoint").String()

	//Get billing coupons
	logger.Log.Info("Get coupons with valid token via API - GET method")
	_, get_response_status := financials.GetBillingWithParams(billing_endpoint_url,"promo1", 200)
	assert.Equal(suite.T(), get_response_status == 200, true, "Test Failed to perform GET api with AuthN enabled on Billing api")
	logger.Log.Info("Billing API service is validated with Authentication successfully!!!")
}

func TestAuthNPositiveFlowSuite(t *testing.T) {
	suite.Run(t, new(authNTestSuite))
}
