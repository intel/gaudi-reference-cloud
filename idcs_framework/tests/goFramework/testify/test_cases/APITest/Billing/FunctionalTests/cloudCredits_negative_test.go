//go:build Functional || CloudCreditsREST || Billing || Negative || Regression
// +build Functional CloudCreditsREST Billing Negative Regression

package BillingAPITest

import (
	//"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/billing"

	"github.com/stretchr/testify/assert"
	//"goFramework/framework/library/financials/cloudAccounts"
	//"github.com/tidwall/gjson"
	//"time"
	//"encoding/json"
	//"goFramework/framework/library/auth"
	//"goFramework/utils"
	//"goFramework/testsetup"
)

func (suite *BillingAPITestSuite) TestAssignCloudCreditsNonexistingCloudAccId() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : NonexistingCloudAccId")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "nonexistingCloudAccId", 404)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : NonexistingCloudAccId")
	assert.Contains(suite.T(), err, "cloud account 123456789012 not found", "Failed : Test :Apply Cloud Credits Negative Test : NonexistingCloudAccId")
}

// func (suite *BillingAPITestSuite) TestAssignCloudCreditsNegativeAmount() {
// 	logger.Log.Info("Starting Apply Cloud Credits Negative Test : NonexistingCloudAccId")
// 	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "negativeAmount", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : NonexistingCloudAccId")
// 	assert.Contains(suite.T(), err, "invalid CloudAccountId", "Failed : Test :Apply Cloud Credits Negative Test : NonexistingCloudAccId")
// }

func (suite *BillingAPITestSuite) TestAssignCloudCreditsInvalidCloudAccIdwithLesserLength() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : InvalidCloudAccIdLessLength")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "InvalidCloudAccIdLessLength", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : InvalidCloudAccIdLessLength")
	assert.Contains(suite.T(), err, "invalid CloudAccountId", "Failed : Test :Apply Cloud Credits Negative Test : NonexistingCloudAccId")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsInvalidCloudAccId() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : InvalidCloudAccId")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "InvalidCloudAccId", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : InvalidCloudAccId")
	assert.Contains(suite.T(), err, "invalid value for string type:", "Failed : Test :Apply Cloud Credits Negative Test : InvalidCloudAccId")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsInvalidExpiration() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : InvalidExpiration")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "InvalidExpiration", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : InvalidExpiration")
	assert.Contains(suite.T(), err, "syntax error", "Failed : Test :Apply Cloud Credits Negative Test : InvalidExpiration")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsWrongReason() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : WrongReason")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "WrongReason", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : WrongReason")
	assert.Contains(suite.T(), err, "invalid value for enum type", "Failed : Test :Apply Cloud Credits Negative Test : WrongReason")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsInvalidReason() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : InvalidReason")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "InvalidReason", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : InvalidReason") // Suites are returning 200 - Potentially a bug
	assert.Contains(suite.T(), err, "INVALID_BILLING_CREDIT_REASON", "Failed : Test :Apply Cloud Credits Negative Test : InvalidReason")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsInvalidAmount() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : InvalidAmount")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "InvalidAmount", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : InvalidAmount")
	assert.Contains(suite.T(), err, "invalid value for double type:", "Failed : Test :Apply Cloud Credits Negative Test : InvalidAmount")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsPreviousExpirationDate() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : PreviousExpirationDate")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "PreviousExpirationDate", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : PreviousExpirationDate")
	assert.Contains(suite.T(), err, "cannot be smaller than current time", "Failed : Test :Apply Cloud Credits Negative Test : PreviousExpirationDate")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsZeroAmount() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : ZeroAmount")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "ZeroAmount", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : ZeroAmount")
	assert.Contains(suite.T(), err, "INVALID_BILLING_CREDIT_AMOUNT", "Failed : Test :Apply Cloud Credits Negative Test : ZeroAmount")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsInvalidamountUsed() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : InvalidamountUsed")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "InvalidamountUsed", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : InvalidamountUsed")
	assert.Contains(suite.T(), err, "invalid value for double type", "Failed : Test :Apply Cloud Credits Negative Test : InvalidamountUsed")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsInvalidremainingAmount() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : InvalidremainingAmount")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "InvalidremainingAmount", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : InvalidremainingAmount")
	assert.Contains(suite.T(), err, "invalid value for double type", "Failed : Test :Apply Cloud Credits Negative Test : InvalidremainingAmount")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsInvalidCouponCode() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : InvalidCouponCode")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "InvalidCouponCode", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : InvalidCouponCode")
	assert.Contains(suite.T(), err, "invalid value for string type", "Failed : Test :Apply Cloud Credits Negative Test : InvalidCouponCode")
}

// func (suite *BillingAPITestSuite) TestAssignCloudCreditsMissingamountUsed() {
// 	logger.Log.Info("Starting Apply Cloud Credits Negative Test : MissingamountUsed")
// 	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "MissingamountUsed", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : MissingamountUsed")
// 	assert.Contains(suite.T(), err, "INVALID_BILLING_CREDIT_AMOUNT", "Failed : Test :Apply Cloud Credits Negative Test : MissingamountUsed")
// }

// func (suite *BillingAPITestSuite) TestAssignCloudCreditsMissingremainingAmount() {
// 	logger.Log.Info("Starting Apply Cloud Credits Negative Test : MissingremainingAmount")
// 	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "MissingremainingAmount", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : MissingremainingAmount")
// 	assert.Contains(suite.T(), err, "INVALID_BILLING_CREDIT_AMOUNT", "Failed : Test :Apply Cloud Credits Negative Test : MissingremainingAmount")
// }

func (suite *BillingAPITestSuite) TestAssignCloudCreditsMissingCouponCode() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : MissingCouponCode")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "MissingCouponCode", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : MissingCouponCode")
	assert.Contains(suite.T(), err, "invalid value for double type:", "Failed : Test :Apply Cloud Credits Negative Test : MissingCouponCode")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsMissingCloudAccId() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : MissingCloudAccId")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "MissingCloudAccId", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : MissingCloudAccId")
	assert.Contains(suite.T(), err, "invalid CloudAccountId", "Failed : Test :Apply Cloud Credits Negative Test : MissingCloudAccId")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsMissingExpiration() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : MissingExpiration")
	ret_value, _ := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "MissingExpiration", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : MissingExpiration")
	//assert.Contains(suite.T(), err, "cloud account", "Failed : Test :Apply Cloud Credits Negative Test : MissingExpiration")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsMissingReason() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : MissingReason")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "MissingReason", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : MissingReason") // Potentially a bug
	assert.Contains(suite.T(), err, "INVALID_BILLING_CREDIT_REASON", "Failed : Test :Apply Cloud Credits Negative Test : MissingReason")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsMissingAmount() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : MissingAmount")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "MissingAmount", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : MissingAmount")
	assert.Contains(suite.T(), err, "INVALID_BILLING_CREDIT_AMOUNT", "Failed : Test :Apply Cloud Credits Negative Test : MissingAmount")
}

func (suite *BillingAPITestSuite) TestAssignCloudCreditsEmptyPayload() {
	logger.Log.Info("Starting Apply Cloud Credits Negative Test : EmptyPayload")
	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "EmptyPayload", 400)
	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : EmptyPayload")
	assert.Contains(suite.T(), err, "Expiration time cannot be empty or nil", "Failed : Test :Apply Cloud Credits Negative Test : EmptyPayload")
}

// func (suite *BillingAPITestSuite) TestAssignCloudCreditsextraFiled() {
// 	logger.Log.Info("Starting Apply Cloud Credits Negative Test : extraFiled")
// 	ret_value, err := billing.CreateCloudCredits("ACCOUNT_TYPE_PREMIUM", "extraFiled", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Failed : Test :Apply Cloud Credits Negative Test : extraFiled")
// 	assert.Contains(suite.T(), err, "unknown field", "Failed : Test :Apply Cloud Credits Negative Test : extraFiled")
// }

func (suite *BillingAPITestSuite) TestGetUnappliedCloudCreditsWrongCustomerId() {
	logger.Log.Info("Starting Get Unapplied Cloud Credits Negative Test : WrongCustomerId")
	ret_value := billing.GetUnappliedCloudCreditsNegative("123456789012", 401)
	logger.Log.Info("ret_value : " + ret_value)
	//assert.Equal(suite.T(), ret_value, true, "Failed : Test :Get Unapplied Cloud Credits Negative Test : WrongCustomerId")
	//assert.Contains(suite.T(), ret_value, "cloud account 123456789012 not found", "Failed : Test :Get Unapplied Cloud Credits Negative Test : WrongCustomerId")
}

func (suite *BillingAPITestSuite) TestGetUnappliedCloudCreditsEmptyCustomerId() {
	logger.Log.Info("Starting Get Unapplied Cloud Credits Negative Test : EmptyCustomerId")
	ret_value := billing.GetUnappliedCloudCreditsNegative("", 403)
	logger.Log.Info("ret_value : " + ret_value)

}

func (suite *BillingAPITestSuite) TestGetUnappliedCloudCreditsInvalidCustomerId() {
	logger.Log.Info("Starting Get Unapplied Cloud Credits Negative Test : InvalidCustomerId")
	ret_value := billing.GetUnappliedCloudCreditsNegative("-10", 403)
	logger.Log.Info("ret_value : " + ret_value)
}

// func (suite *BillingAPITestSuite) TestInstanceCreationForPremiumUserwithExpiredCredits() {
// 	logger.Log.Info("Starting Test: Apply cloud Credits to billing account and verify in Aria")
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	enterpriseId := "premimum1-eid1"
// 	userName := utils.Get_UserName("Premium")
// 	get_CAcc_id, acc_type, jsonStr := cloudAccounts.CreateCAccwithEnroll("premium", tid, userName, enterpriseId, true, 200)
// 	assert.Equal(suite.T(), "ENROLL_ACTION_COUPON_OR_CREDIT_CARD", gjson.Get(jsonStr, "action").String(), "Test Failed while validating type of cloud account")
// 	fmt.Println("ACC TYPE", acc_type)
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_PREMIUM", "Test Failed while validating type of cloud account")
// 	// Add coupon to the user
// 	creation_time, expirationtime := billing.GetCreationExpirationTime()
// 	createCoupon := billing.CreateCouponStruct{
// 		Amount:  1,
// 		Creator: "idc_billing@intel.com",
// 		Expires: expirationtime,
// 		Start:   creation_time,
// 		NumUses: 2,
// 	}
// 	jsonPayload, _ := json.Marshal(createCoupon)
// 	req := []byte(jsonPayload)
// 	ret_value, data := billing.CreateCoupon(req, 200)
// 	couponCode := gjson.Get(data, "code").String()
// 	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
// 	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
// 	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
// 	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
// 	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
// 	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
// 	// Get coupon and validate
// 	getret_value, getdata := billing.GetCoupons(couponCode, 200)

//     couponData := gjson.Get(getdata,"coupons")
//     for _, val := range couponData.Array() {
// 	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
// 	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
// 	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
// 	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
// 	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
// 	assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
// 	}
// 	assert.Equal(suite.T(), getret_value, true, "Failed: Get on Coupon Failed")

// 	//Redeem coupon
// 	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
//     time.Sleep(1 * time.Minute)
// 	redeemCoupon := billing.RedeemCouponStruct{
// 		CloudAccountID: get_CAcc_id,
// 		Code:           couponCode,
// 	}
// 	jsonPayload, _ = json.Marshal(redeemCoupon)
// 	req = []byte(jsonPayload)
// 	ret_value, data = billing.RedeemCoupon(req, 200)

// 	// Get coupon and validate
// 	getret_value, getdata = billing.GetCoupons(couponCode, 200)
// 	couponData = gjson.Get(getdata,"coupons")
// 	redemptions := gjson.Get(getdata,"result.redemptions")
// 	for _, val := range redemptions.Array() {
// 		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
// 	    assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
// 	    assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
// 	}
// 	for _, val := range couponData.Array() {
// 		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
// 		}

// 	// Now enroll user
// 	get_CAcc_id, acc_type, jsonStr = cloudAccounts.CreateCAccwithEnroll("premium", tid, userName, enterpriseId, true, 200)
// 	assert.Equal(suite.T(), "ENROLL_ACTION_NONE", gjson.Get(jsonStr, "action").String(), "Test Failed while validating type of cloud account")
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_PREMIUM", "Test Failed while validating type of cloud account")
// 	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_PREMIUM", "Test Failed while validating type of cloud account")
// 	logger.Log.Info("Starting Billing Test : Apply cloud Credits to billing account and verify in Aria")

// 	creation_time, expirationtime = billing.GetExpirationInOneMinute()
// 	createCloudCredits := billing.CreateCloudCreditsStruct{
// 		CloudAccountID:  get_CAcc_id,
// 		Expiration:      expirationtime,
// 		Reason:          financials_utils.GetCreditReason(),
// 		OriginalAmount:  200,
// 		AmountUsed:      0,
// 		CouponCode:      utils.GenerateString(6),
// 		Created:         "2024-03-16T14:53:29Z",
// 		RemainingAmount: 0,
// 	}
// 	ret_value, _ = billing.ApplyCloudCreditsToBillingAccount(createCloudCredits, 200, "3000")
// 	assert.Equal(suite.T(), ret_value, true, "Failed : Test : Apply cloud Credits to billing account and verify in Aria")

// 	// //token := utils.Get_intel_Token()
// 	computeUrl := utils.Get_Compute_Base_Url()
// 	logger.Log.Info("Compute Url"+ computeUrl)
//     baseUrl := utils.Get_Base_Url1()
// 	// Create an ssh key  for the user
// 	authToken:= "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

// 	sshKeyData := utils.Get_ssh_Key_Payload("premium.user.charlie@proton.me", "testpremiumuserkey")
// 	logger.Log.Info("SSH Key data " + sshKeyData)
// 	err := testsetup.CreateSSHKey(sshKeyData , computeUrl, baseUrl,  authToken)
//     assert.Equal(suite.T(), nil, err, "Failed: SSH key creation failed for intel user")

// 	vnetData := utils.Get_Vet_Payload("premium.user.charlie@proton.me", "us-dev-1a-default")
// 	logger.Log.Info("VNet data "+ vnetData)
// 	err = testsetup.CreateVnet(vnetData , computeUrl, baseUrl, authToken)
//     assert.Equal(suite.T(), nil, err, "Failed: Vnet creation failed for intel user")

// 	instanceData := utils.Get_Instance_Payload("premium.user.charlie@proton.me", "testintelsmall")
// 	logger.Log.Info("Instance data "+ instanceData)
// 	//instanceName := gjson.Get(instanceData, "name").String()
// 	err = testsetup.CreateVmInstance(instanceData, computeUrl, baseUrl, authToken)
// 	logger.Logf.Infof("Error message from create Instance with expired credit", err)
//     assert.NotEqual(suite.T(), nil, err, "Failed: Instance creation failed for premium user")
// 	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")
// }
