//go:build Functional || CloudCreditsREST || Billing || Negative || Regression
// +build Functional CloudCreditsREST Billing Negative Regression

package BillingAPITest

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/utils"
	"time"

	"goFramework/framework/service_api/financials"
	"goFramework/testsetup"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	//"strconv"
)

func (suite *BillingAPITestSuite) TestAssignCloudCreditsPremiumUser() {
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

	ret_value1, _ = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	logger.Log.Info("Starting Billing Test : Apply cloud Credits to billing account and verify in Aria")

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      expirationtime,
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(7),
		Created:         creation_time,
		RemainingAmount: 0,
	}
	ret_value, _ := billing.ApplyCloudCreditsToBillingAccount(createCloudCredits, 200, "500")
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply cloud Credits to billing account and verify in Aria")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) TestApplyMoreCreditsPremiumUser() {
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

	ret_value1, _ = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	logger.Log.Info("Starting Billing Test : Apply cloud Credits to billing account and verify in Aria")
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      expirationtime,
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(6),
		Created:         creation_time,
		RemainingAmount: 0,
	}
	ret_value, _ := billing.ApplyCloudCreditsToBillingAccount(createCloudCredits, 200, "500")
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply cloud Credits to billing account and verify in Aria")
	morecreateCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      expirationtime,
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(6),
		Created:         creation_time,
		RemainingAmount: 200,
	}
	ret_value, _ = billing.ApplyCloudCreditsToBillingAccount(morecreateCloudCredits, 200, "700")
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply more cloud Credits to billing account and verify in Aria")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) TestGetCloudCreditsWithCloudAccountId() {
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

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("premium", tid, userName, enterpriseId, false, 200)
	/////
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Now launch paid instance and see API throws 403 error

	upgrade_err := billing.Upgrade_to_Premium_with_coupon(get_CAcc_id, authToken, userToken, int64(300))
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)
	ret_value1, _ = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	logger.Log.Info("Starting Billing Test : Apply cloud Credits to billing account and verify in Aria")
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      expirationtime,
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(7),
		Created:         creation_time,
		RemainingAmount: 0,
	}
	ret_value, _ := billing.ApplyCloudCreditsToBillingAccount(createCloudCredits, 200, "500")
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply cloud Credits to billing account and verify in Aria")
	logger.Log.Info("Performing Get on Cloud Credits using CloudAccountId")
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	logger.Log.Info("Output from Get cloud Credits : %s" + responseBody)
	assert.Equal(suite.T(), 200, response_status, "Failed : Test : Get cloud Credits Failed")
	//validate the data
	//fail because coupon code is coming as empty in response
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) TestGetUnappliedCloudCredits() {
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
	ret_value1, _ = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	logger.Log.Info("Starting Billing Test : Apply cloud Credits to billing account and verify in Aria")
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      expirationtime,
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(6),
		Created:         creation_time,
		RemainingAmount: 0,
	}

	ret_value, _ := billing.ApplyCloudCreditsToBillingAccount(createCloudCredits, 200, "500")
	res := billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply cloud Credits to billing account and verify in Aria")
	morecreateCloudCredits := billing.CreateCloudCreditsStruct{
		CloudAccountID:  get_CAcc_id,
		Expiration:      expirationtime,
		Reason:          financials_utils.GetCreditReason(),
		OriginalAmount:  200,
		AmountUsed:      0,
		CouponCode:      utils.GenerateString(6),
		Created:         creation_time,
		RemainingAmount: 200,
	}
	ret_value, _ = billing.ApplyCloudCreditsToBillingAccount(morecreateCloudCredits, 200, "700")
	res = billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits = gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits2 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply more cloud Credits to billing account and verify in Aria")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}
