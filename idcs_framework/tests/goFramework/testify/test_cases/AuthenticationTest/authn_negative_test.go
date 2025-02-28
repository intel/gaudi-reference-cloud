//go:build Functional || authn_regression || authn_negative
// +build Functional authn_regression authn_negative

package AuthenticationTest

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"
    "goFramework/framework/library/financials/metering"
	"goFramework/framework/library/vmaas/serviceapi"
	"goFramework/utils"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

func (suite *authNTestSuite) TestComputeAPIWithTaintedAuthN() {
	logger.Log.Info("Compute api GET call - negative flow with tainted token")

	//Fetching and populating the input from config file
	compute_endpoint_url := utils.GetBaseUrl() + gjson.Get(utils.GetComputeApiData(), "endpoint").String()

	//Get the all instance types supported with invalid token
	logger.Log.Info("Get instance types with tainted token via Compute API - GET method")
	get_response_body, get_response_status := serviceapi.GetAllInstanceType(compute_endpoint_url, 401)
	assert.Equal(suite.T(), get_response_status == 401, true, "Test should be failed with unauthorized response code")
	assert.Equal(suite.T(), strings.Contains(get_response_body, `Jwt header is an invalid JSON`), true,
		"Test should be failed with the message invalid JWT token")
	logger.Log.Info("Validated compute api with authentication successfully!!!")
}

func (suite *authNTestSuite) TestMeteringAPIWithTaintedAuthN() {
	logger.Log.Info("List previous metering records via Metering API - positive flow")

	//Fetching and populating the input from config file
	metering_endpoint_url := utils.GetBaseUrl() + gjson.Get(utils.GetMeteringApiData(), "endpoint").String()
	resourceId := gjson.Get(utils.GetMeteringApiData(), "resourceId").String()

	//Get the previous metering record with test resource id & invalid token
	logger.Log.Info("Retrieve previous metering records by resource id via Metering API - GET method")
	get_response_body, get_response_status := metering.GetMeteringRecordsWithParams(metering_endpoint_url, resourceId, 401)
	assert.Equal(suite.T(), get_response_status == 401, true, "Test should be failed with unauthorized response code")
	assert.Equal(suite.T(), strings.Contains(get_response_body, `Jwt header is an invalid JSON`), true,
		"Test should be failed with the message invalid JWT token")
	logger.Log.Info("Validated Metering api with authentication successfully!!!")

}

func (suite *authNTestSuite) TestProductCatalogAPIWithTaintedAuthN() {
	logger.Log.Info("List vendors API - positive flow")

	//Fetching and populating the input from config file
	productcatalog_endpoint_url := utils.GetBaseUrl() + gjson.Get(utils.GetProductCatalogApiData(), "endpoint").String()

	//Get vendors with params & invalid token
	logger.Log.Info("Retrieve vendors via ProductCatalog API - GET method")
	get_response_body, get_response_status := financials.GetVendorsWithParams(productcatalog_endpoint_url, "noFilters", 401)
	assert.Equal(suite.T(), get_response_status == 401, true, "Test should be failed with unauthorized response code")
	assert.Equal(suite.T(), strings.Contains(get_response_body, `Jwt header is an invalid JSON`), true,
		"Test should be failed with the message invalid JWT token")
	logger.Log.Info("Validated productCatalog api with authentication successfully!!!")

}

func (suite *authNTestSuite) TestCloudAccountWithTaintedAuthN() {
	logger.Log.Info("Get Cloud Accounts by type API - positive flow")

	//Fetching and populating the input from config file
	cloudaccount_endpoint_url := utils.GetBaseUrl() + gjson.Get(utils.GetCloudAccountApiData(), "endpoint").String()

	//Get Cloud Accounts by type & invalid token
	logger.Log.Info("Retrieve Cloud Accounts by type via get Cloud Accounts - GET method")
	get_response_body, get_response_status := financials.Get_CAcc_by_type(cloudaccount_endpoint_url, "ACCOUNT_TYPE_PREMIUM", 401)
	assert.Equal(suite.T(), get_response_status == 401, true, "Test should be failed with unauthorized response code")
	assert.Equal(suite.T(), strings.Contains(get_response_body, `Jwt header is an invalid JSON`), true,
		"Test should be failed with the message invalid JWT token")
	logger.Log.Info("Validated Metering api with authentication successfully!!!")

}

func (suite *authNTestSuite) TestBillingWithTaintedAuthN() {
	logger.Log.Info("Get Billing Coupons API - positive flow")

	//Fetching and populating the input from config file
	billing_endpoint_url := utils.GetBaseUrl() + gjson.Get(utils.GetBillingApiData(), "endpoint").String()

	//Get Billing coupons & invalid token
	logger.Log.Info("Retrieve all coupons using Billing - GET method")
	get_response_body, get_response_status := financials.GetBillingWithParams(billing_endpoint_url,"promo1", 401)
	assert.Equal(suite.T(), get_response_status == 401, true, "Test should be failed with unauthorized response code")
	assert.Equal(suite.T(), strings.Contains(get_response_body, `Jwt header is an invalid JSON`), true,
		"Test should be failed with the message invalid JWT token")
	logger.Log.Info("Validated Billing api with authentication successfully!!!")

}

func TestAuthNNegativeFlowSuite(t *testing.T) {
	suite.Run(t, new(authNTestSuite))
}
