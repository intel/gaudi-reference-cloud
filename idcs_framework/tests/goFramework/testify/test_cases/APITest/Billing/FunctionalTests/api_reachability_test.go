//go:build Functional || Coupons || Regression || Billing
// +build Functional Coupons Regression Billing

package BillingAPITest

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/financials"
	"goFramework/testsetup"
	"goFramework/utils"
	"os"
	"time"

	//"strconv"
	//"goFramework/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) Test_Check_deativate_instances_API() {
	logger.Log.Info("Check Deactivation List API")
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	get_response_status, _ := financials.GetInstanceDeactivationList(base_url, authToken)
	assert.Equal(suite.T(), get_response_status, 200, "Get Deactivation API Did not return 200 response code")
}

func (suite *BillingAPITestSuite) Test_Check_StaaS_deativate_API() {
	logger.Log.Info("Check Deactivation List API")
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	get_response_status, _ := financials.GetStaaSDeactivationList(base_url, authToken)
	assert.Equal(suite.T(), get_response_status, 403, "Get StaaS Deactivation API Did not return 200 response code") // 403 is expected as this API is not allowed for admin users
}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	cloudAccId, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), cloudAccId, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")

	cloudAccId, _ = testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(10), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.Equal(suite.T(), "UPGRADE_COMPLETE", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Validate credit details

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(10), float64(10), float64(10))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Credit_Card_Visa_Card() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	cloudAccId, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), cloudAccId, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")

	cloudAccId, _ = testsetup.GetCloudAccountId(userName, base_url, authToken)

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Check cloud account attributes before upgrade

	creditCardDetails := financials.CreditCardDetails{
		CCNumber:      4012888888881881,
		CCExpireMonth: 11,
		CCExpireYear:  2030,
		CCV:           123,
	}

	upgrade_err := financials.UpgradeAccountWithCreditCard(base_url, cloudAccId, userToken, creditCardDetails, "Visa", "1", "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with credit card failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.Equal(suite.T(), "UPGRADE_COMPLETE", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Validate credit details

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	logger.Log.Info("Delete cloud acc response" + ret_value1)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Get_Billing_Rates() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("intel_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	base_url := utils.Get_Base_Url1()
	get_response_status, body := financials.GetProductRate(base_url, userToken, get_CAcc_id, "3bc52387-da79-4947-a562-04aad42e1db7")
	assert.Equal(suite.T(), get_response_status, 200, "Get Product Rate Did not return 200 response code")
	assert.NotEqual(suite.T(), body, "", "Get Product Rate Did not return any response")
}
