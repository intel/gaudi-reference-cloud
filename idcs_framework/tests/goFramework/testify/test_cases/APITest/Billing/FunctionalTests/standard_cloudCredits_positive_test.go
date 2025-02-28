//go:build Functional || BillingAccountRest || Billing || Positive || Standard || Regression
// +build Functional BillingAccountRest Billing Positive Standard Regression

package BillingAPITest

import (
	//"fmt"
	//"encoding/json"
	//"time"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/utils"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	//"strconv"
	"os"
)

func (suite *BillingAPITestSuite) TestGetCloudCreditsWithCloudAccountIdStandard() {
	logger.Log.Info("Starting Test : Performing test Get on Cloud Credits using CloudAccountId")
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	logger.Log.Info("Starting Billing Test : Apply cloud Credits to cloud account and verify")

	createCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      "2032-06-16T14:53:29Z",
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(7),
		Created:         "2024-03-16T14:53:29Z",
		RemainingAmount: 200,
	}
	ret_value, _ := billing.ApplyCloudCreditsToBillingAccount(createCloudCredits, 200, "200")
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply cloud Credits to cloud account and verify")
	logger.Log.Info("Performing Get on Cloud Credits using CloudAccountId")
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	logger.Log.Info("Output from Get cloud Credits : %s" + responseBody)
	assert.Equal(suite.T(), 200, response_status, "Failed : Test : Get cloud Credits Failed")
	//validate the data
	//fail because coupon code is coming as empty in response
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
	os.Unsetenv("standard_user_test")
}

func (suite *BillingAPITestSuite) TestGetUnappliedCloudCreditsStandarduserStandard() {
	logger.Log.Info("Starting Test: Get Unapplied Cloud Credits")
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	logger.Log.Info("Starting Billing Test : Apply cloud Credits to cloud account and verify")

	createCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      "2032-06-16T14:53:29Z",
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(6),
		Created:         "2024-03-16T14:53:29Z",
		RemainingAmount: 200,
	}

	ret_value, _ := billing.ApplyCloudCreditsToBillingAccount(createCloudCredits, 200, "200")
	res := billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply cloud Credits to cloud account and verify")
	morecreateCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      "2032-06-16T14:53:29Z",
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(6),
		Created:         "2024-03-16T14:53:29Z",
		RemainingAmount: 200,
	}
	ret_value, _ = billing.ApplyCloudCreditsToBillingAccount(morecreateCloudCredits, 200, "400")
	res = billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits = gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits2 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply more cloud Credits to cloud account and verify")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
	os.Unsetenv("standard_user_test")

}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsStandardUserStandard() {
	logger.Log.Info("Starting Test : Assign cloud Credits")
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	logger.Log.Info("Starting Billing Test : Apply cloud Credits to cloud account ")

	createCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      "2032-06-16T14:53:29Z",
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(7),
		Created:         "2024-03-16T14:53:29Z",
		RemainingAmount: 200,
	}
	ret_value, _ := billing.ApplyCloudCreditsToBillingAccount(createCloudCredits, 200, "200")
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply cloud Credits to cloud account ")
	logger.Log.Info("Performing Get on Cloud Credits using CloudAccountId")
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	logger.Log.Info("Output from Get cloud Credits : %s" + responseBody)
	assert.Equal(suite.T(), 200, response_status, "Failed : Test : Get cloud Credits Failed")
	//validate the data
	//fail because coupon code is coming as empty in response
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
	os.Unsetenv("standard_user_test")
}

func (suite *BillingAPITestSuite) TestGetUnappliedCloudCreditsStandard() {
	logger.Log.Info("Starting Test: Get Unapplied Cloud Credits")
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	logger.Log.Info("Starting Billing Test : Apply cloud Credits to cloud account ")

	createCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      "2032-06-16T14:53:29Z",
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(6),
		Created:         "2024-03-16T14:53:29Z",
		RemainingAmount: 200,
	}

	ret_value, _ := billing.ApplyCloudCreditsToBillingAccount(createCloudCredits, 200, "200")
	res := billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply cloud Credits to cloud account ")
	morecreateCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      "2032-06-16T14:53:29Z",
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  300,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(6),
		Created:         "2024-03-16T14:53:29Z",
		RemainingAmount: 300,
	}
	ret_value, _ = billing.ApplyCloudCreditsToBillingAccount(morecreateCloudCredits, 200, "500")
	res = billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits = gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits2 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply more cloud Credits to cloud account ")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
	os.Unsetenv("standard_user_test")
}

func (suite *BillingAPITestSuite) TestTApplyMoreCreditsStandard1User() {
	logger.Log.Info("Starting Test: Apply cloud Credits to cloud account and verify")
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	logger.Log.Info("Starting Billing Test : Apply cloud Credits to cloud account and verify")

	// res11 := billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	// unappliedCredits11 := gjson.Get(res11, "unappliedAmount").String()
	// logger.Logf.Info("Unapplied credits before coupon redemption: ", unappliedCredits11)
	createCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      "2032-06-16T14:53:29Z",
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(6),
		Created:         "2024-03-16T14:53:29Z",
		RemainingAmount: 200,
	}
	ret_value, _ := billing.ApplyCloudCreditsToBillingAccount(createCloudCredits, 200, "200")
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply cloud Credits to cloud account and verify")
	morecreateCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      "2032-06-16T14:53:29Z",
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(6),
		Created:         "2024-03-16T14:53:29Z",
		RemainingAmount: 200,
	}
	ret_value, _ = billing.ApplyCloudCreditsToBillingAccount(morecreateCloudCredits, 200, "400")
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply more cloud Credits to cloud account and verify")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
	os.Unsetenv("standard_user_test")
}
