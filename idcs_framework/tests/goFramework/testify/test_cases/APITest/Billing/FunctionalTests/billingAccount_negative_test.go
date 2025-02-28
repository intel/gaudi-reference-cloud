//go:build Functional || BillingAccountRest || Regression || Billing || Negative
// +build Functional BillingAccountRest Regression Billing Negative

package BillingAPITest

import (
	//"encoding/json"
	//"time"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"

	"github.com/stretchr/testify/assert"

	//"github.com/tidwall/gjson"
	//"github.com/tidwall/gjson"
	"goFramework/framework/common/logger"
	//"goFramework/utils"
)

// func (suite *BillingAPITestSuite) TestCreateBillingAccountusingCloudAccountIdStandard() {
// 	logger.Log.Info("Starting Billing Account Create API Test for Standard")
// 	ret_value, err := billing.CreateBillingAccountWithCloudAccountId("ACCOUNT_TYPE_STANDARD", 400)
// 	logger.Log.Info("Error :  " + err)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed: Billing Account Created for Standard Customer")
// }

// func (suite *BillingAPITestSuite) TestCreatBillingAccountusingCloudAccountIdIntel() {
// 	logger.Log.Info("Starting Billing Account Create API Test for Intel")
// 	ret_value, err := billing.CreateBillingAccountWithCloudAccountId("ACCOUNT_TYPE_INTEL", 400)
// 	logger.Log.Info("Error :  " + err)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed: Starting Billing Account Create for Intel Customer")
// }

// func (suite *BillingAPITestSuite) TestCreateCloudAccountusingExistingAccountsCloudAccountId() {
// 	logger.Log.Info("Starting Billing Account Create for existing billing Account ID")
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	enterpriseId := "premimum1-eid1"
// 	userName := utils.Get_UserName("Premium")
// 	get_CAcc_id, acc_type, jsonStr := cloudAccounts.CreateCAccwithEnroll("premium", tid, userName, enterpriseId, true, 200)
// 	assert.Equal(suite.T(), "ENROLL_ACTION_COUPON_OR_CREDIT_CARD", gjson.Get(jsonStr, "action").String(), "Test Failed while validating type of cloud account")
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_PREMIUM", "Test Failed while validating type of cloud account")
// 	// Add coupon to the user
// 	creation_time, expirationtime := billing.GetCreationExpirationTime()
// 	createCoupon := billing.CreateCouponStruct{
// 		Amount:  300,
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
// 	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(getdata, "result.amount").Int(), "Failed: Create Coupon Failed to validate Amount")
// 	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(getdata, "result.creator").String(), "Failed: Create Coupon Failed to validate Creator")
// 	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(getdata, "result.expires").String(), "Failed: Create Coupon Failed to validate Expires")
// 	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(getdata, "result.numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
// 	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(getdata, "result.start").String(), "Failed: Create Coupon Failed to validate numUses")
// 	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")
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
// 	redemptions := gjson.Get(getdata,"result.redemptions")
// 	for _, val := range redemptions.Array() {
// 		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
// 	    assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
// 	    assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
// 	}
// 	assert.Equal(suite.T(), "1", gjson.Get(getdata, "result.numRedeemed").String(), "Failed: Validation Failed in Coupon Redemption")

// 	// Now enroll user
// 	get_CAcc_id, acc_type, jsonStr = cloudAccounts.CreateCAccwithEnroll("premium", tid, userName, enterpriseId, true, 200)
// 	assert.Equal(suite.T(), "ENROLL_ACTION_NONE", gjson.Get(jsonStr, "action").String(), "Test Failed while validating type of cloud account")
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_PREMIUM", "Test Failed while validating type of cloud account")
// 	ret, err := billing.CreateBillingAccountWithSpecificCloudAccountId(get_CAcc_id, 500)
// 	assert.Equal(suite.T(), ret, true, "Test Failed: Starting Billing Account Create for existing billing Account ID")
// 	assert.Contains(suite.T(), err, "aria client api error cloudAccountId:", "Failed : Test :Apply Cloud Credits Negative Test : InvalidEMail")
// }

func (suite *BillingAPITestSuite) TestCreateCloudAccountusingWrongCloudAccountId() {
	logger.Log.Info("Starting Billing Account Create API Test for Wrong Account ID")
	ret_value, err := billing.CreateBillingAccountWithCloudAccountId("CloudAccId", 500)
	logger.Log.Info("Error :  " + err)
	assert.Contains(suite.T(), err, "FAILED_TO_CREATE_BILLING_ACCT", "Test Failed: TestCreateCloudAccountusingWrongCloudAccountId")
	assert.Equal(suite.T(), ret_value, true, "Test Failed: Starting Billing Account Create for Wrong Account ID")
}

func (suite *BillingAPITestSuite) TestCreateCloudAccountusingWrongCloudAccountIdType() {
	logger.Log.Info("Starting Billing Account Create API Test for Wrong Cloud Account ID type ")
	ret_value, _ := billing.CreateBillingAccountWithCloudAccountId("WrongTypeCloudAccId", 400)
	assert.Equal(suite.T(), ret_value, true, "Test Failed: Starting Billing Account Create  for Wrong Cloud Account ID type")
}

func (suite *BillingAPITestSuite) Test_Check_Aria_Account_not_created_for_Standard_User() {
	logger.Log.Info("Ensure no related Aria account is created for standard user")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	check_Billing, _ := cloudAccounts.CAcc_check_BillingAcc(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), check_Billing, "False", "Test Failed while checking billing account for Cloud account")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_IU
func (suite *BillingAPITestSuite) Test_Check_Aria_Account_not_created_for_Intel_User() {
	logger.Log.Info("Ensure no related Aria account is created for intel user")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	check_Billing, _ := cloudAccounts.CAcc_check_BillingAcc(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), check_Billing, "False", "Test Failed while checking billing account for Cloud account")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// func (suite *BillingAPITestSuite) TestCreateCloudAccountWithInvalidName() {
// 	logger.Log.Info("Starting Billing Account Create API Test : Invalidtid")
// 	_, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
// 	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(10, oid, owner, parentid, tid, false, false, false, false, false,
// 		true, false, "ACCOUNT_TYPE_INTEL", 400)
// 	assert.Equal(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account with invalid details : Invalid tid")
// }

// func (suite *BillingAPITestSuite) TestCreateCloudAccountWithInvalidOid() {
// 	logger.Log.Info("Starting Billing Account Create API Test : InvalidOid")
// 	name, _, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
// 	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, 10, owner, parentid, tid, false, false, false, false, false,
// 		true, false, "ACCOUNT_TYPE_INTEL", 400)
// 	assert.Equal(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account with invalid details : Invalid Oid")
// }

// func (suite *BillingAPITestSuite) TestCreateCloudAccountWithInvalidOwner() {
// 	logger.Log.Info("Starting Billing Account Create API Test : InvalidOwner")
// 	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
// 	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, 10, parentid, tid, false, false, false, false, false,
// 		true, false, "ACCOUNT_TYPE_INTEL", 400)
// 	assert.Equal(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account with invalid details : Invalid Owner")
// }

// func (suite *BillingAPITestSuite) TestCreateCloudAccountWithInvalidParentId() {
// 	logger.Log.Info("Starting Billing Account Create API Test : InvalidParentId")
// 	name, oid, owner, _, tid := cloudAccounts.CAcc_RandomPayload_gen()
// 	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, 10, tid, false, false, false, false, false,
// 		true, false, "ACCOUNT_TYPE_INTEL", 400)
// 	assert.Equal(suite.T(), ret_value, false, "Test Failed: Starting Billing Account Create API Test : InvalidParentId")
// }
