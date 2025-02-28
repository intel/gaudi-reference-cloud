//go:build Functional || BillingAccountRest || Billing || Positive || Intel
// +build Functional BillingAccountRest Billing Positive Intel

package BillingAPITest

import (
	"encoding/json"
	//"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/utils"
	"os"
	"time"

	//"strconv"
	//"goFramework/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) TestRedeemCouponIntelwithNouses1() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("intel_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.CreateCouponStruct{
		Amount:  300,
		Creator: "idc_billing@intel.com",
		Expires: expirationtime,
		Start:   creation_time,
		NumUses: 1,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode, 200)
	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "", gjson.Get(val.String(), "disabled").String(), "Failed: Create Coupon Failed to validate disabled")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")
	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Intel cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	_, getdata = billing.GetCoupons(couponCode, 400)
	couponData = gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestRedeemCouponRedeembySameCloudAccountIdIntel() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("intel_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an intel user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.CreateCouponStruct{
		Amount:  300,
		Creator: "idc_billing@intel.com",
		Expires: expirationtime,
		Start:   creation_time,
		NumUses: 2,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode, 200)
	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "", gjson.Get(val.String(), "disabled").String(), "Failed: Create Coupon Failed to validate disabled")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")
	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Intel cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode, 200)
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}
	couponData = gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	logger.Log.Info("Starting Billing Test : Redeem coupon for the second time Intel cloud account")

	redeemCoupon = billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 409)

	// Validate
	assert.Contains(suite.T(), gjson.Get(data, "message").String(), "has already been redeemed", "Coupon redeemed twice for (Intel user)")

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode, 200)
	redemptions = gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}
	couponData = gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	redeemCoupon = billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 409)

	// Validate
	assert.Contains(suite.T(), gjson.Get(data, "message").String(), "has already been redeemed", "Coupon redeemed twice for (Intel user)")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account(Intel User)")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestRedeemCouponWithInvalidCloudAccIdIntel() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	//Create Clloud Account
	os.Setenv("intel_user_test", "True")
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.CreateCouponStruct{
		Amount:  300,
		Creator: "idc_billing@intel.com",
		Expires: expirationtime,
		Start:   creation_time,
		NumUses: 2,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode, 200)
	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "", gjson.Get(val.String(), "disabled").String(), "Failed: Create Coupon Failed to validate disabled")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")
	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Intel cloud account")

	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: "123456789543",
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 400)

	// // Get coupon and validate
	_, getdata = billing.GetCoupons(couponCode, 200)
	couponData = gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestRedeemCouponWithInvalidCouponIdIntel() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	//Create Clloud Account
	os.Setenv("intel_user_test", "True")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           "ABCD_EFGH_IJKL",
	}
	jsonPayload, _ := json.Marshal(redeemCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.RedeemCoupon(req, 404)
	assert.Equal(suite.T(), true, ret_value, "Failed: Validation Failed in Coupon Redemption")
	assert.Contains(suite.T(), data, "does not exists. please provide a valid coupon code", "Failed: Validation Failed in Coupon Redemption")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestRedeemExpiredCoupon() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("intel_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	creation_time, expirationtime := billing.GetExpirationInOneMinute()
	createCoupon := billing.CreateCouponStruct{
		Amount:  300,
		Creator: "idc_billing@intel.com",
		Expires: expirationtime,
		Start:   creation_time,
		NumUses: 2,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode, 200)
	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")
	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Intel cloud account")
	time.Sleep(3 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode, 500)
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}
	couponData = gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestRedeemStandardCoupontoIntel() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("intel_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	creation_time, expirationtime := billing.GetExpirationInOneMinute()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    2,
		IsStandard: true,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode, 200)
	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")
	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Intel cloud account")
	time.Sleep(3 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 500)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode, 500)
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "false", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}
	couponData = gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	os.Unsetenv("intel_user_test")
}
