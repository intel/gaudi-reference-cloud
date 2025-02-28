//go:build Functional || BillingAccountRest || Billing || Regression || Positive
// +build Functional BillingAccountRest Billing Regression Positive

package BillingAPITest

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/financials"
	"goFramework/testsetup"
	"goFramework/utils"
	"time"

	"github.com/stretchr/testify/assert"
)

func (suite *BillingAPITestSuite) TestCreatePremiumBillingAccountusingEnroll() {
	logger.Log.Info("Creating a Premium User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "premimum1-eid1"
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)

	computeUrl := utils.Get_Compute_Base_Url()
	logger.Log.Info("Compute Url" + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	url := base_url + "/v1/cloudaccounts"
	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}
	time.Sleep(1 * time.Minute)

	userToken = "Bearer " + userToken

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Now launch paid instance and see API throws 403 error

	upgrade_err := billing.Upgrade_to_Premium_with_coupon(get_CAcc_id, authToken, userToken, int64(300))
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)
}

func (suite *BillingAPITestSuite) TestCreateBillingAccountusingCloudAccountIdPremium() {
	logger.Log.Info("Starting Billing Account Create API Test for Premium")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account")
	ret_value, _ := billing.CreateBillingAccountWithSpecificCloudAccountId(get_CAcc_id, 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed: Starting Billing Account Create API Test")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// func (suite *BillingAPITestSuite) TestCreateBillingAccountusingCloudAccountIdEnterprise() {
// 	logger.Log.Info("Starting Billing Account Create API Test for Enterprise")
// 	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
// 	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
// 		true, true, "ACCOUNT_TYPE_ENTERPRISE", 200)
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account")
// 	ret_value, _ := billing.CreateBillingAccountWithSpecificCloudAccountId(get_CAcc_id, 200)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed: Starting Billing Account Create API Test")
// 	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
// 	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
// }

// func (suite *BillingAPITestSuite) TestCreateBillingAccountEnterprise() {
// 	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	enterpriseId := "enterprise1-eid1"
// 	userName := "test.enterprise1@proton.me"
// 	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("enterprise", tid, userName, enterpriseId, true, 200)
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_ENTERPRISE", "Test Failed while validating type of cloud account")
// 	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
// 	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account(Enterprise User)")
// }

// func (suite *BillingAPITestSuite) TestCreateBillingAccountIntel() {
// 	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	enterpriseId := "testeid-01"
// 	userName := "testacc@intel.com"
// 	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
// 	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
// 	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account(Intel User)")
// }
