//go:build Functional || Coupons || Regression || Billing
// +build Functional Coupons Regression Billing

package BillingAPITest

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"os"
	"time"

	//"strconv"
	//"goFramework/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	//"goFramework/framework/library/auth"
	//"goFramework/testsetup"
)

// New Validation: We can't create coupons with No Uses
// We catch this excenario in createCoupons Tests
/*
func (suite *BillingAPITestSuite) TestRedeemCouponPremiumwithNouses() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
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
	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
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

}
*/
func (suite *BillingAPITestSuite) Test_RedeemCoupon_Premium_with_No_uses_More_Than_50() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
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
		NumUses: 61,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	_, data := billing.CreateCoupon(req, 400)
	assert.Contains(suite.T(), data, "Coupon number of uses 61 exceeds the limit of 50", "Test Failed in validation")

}

func (suite *BillingAPITestSuite) Test_RedeemCoupon_Standard_coupon_to_Intel() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    1,
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
		assert.Equal(suite.T(), "", gjson.Get(val.String(), "disabled").String(), "Failed: Create Coupon Failed to validate disabled")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")
	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 500)
	assert.Contains(suite.T(), data, "can be redeemed only for standard customers", "Failed: Redeem Coupon Test Failed")

	// Get coupon and validate
	_, getdata = billing.GetCoupons(couponCode, 200)
	couponData = gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

}

func (suite *BillingAPITestSuite) TestRedeemExpiredCouponIntel() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")

	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
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

	ret_value6, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value6, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) Test_RedeemCoupon_Standard_coupon_to_Premium() {
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

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    1,
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
		assert.Equal(suite.T(), "", gjson.Get(val.String(), "disabled").String(), "Failed: Create Coupon Failed to validate disabled")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")
	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 500)
	assert.Contains(suite.T(), data, "can be redeemed only for standard customers", "Failed: Redeem Coupon Test Failed")

	// Get coupon and validate
	_, getdata = billing.GetCoupons(couponCode, 200)
	couponData = gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

}

func (suite *BillingAPITestSuite) TestRedeemCouponRedeembySameCloudAccountId() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
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
	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
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
	logger.Log.Info("Starting Billing Test : Redeem coupon for the second time Premium cloud account")

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

}

func (suite *BillingAPITestSuite) TestRedeemCouponWithInvalidCloudAccId() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	//Create Clloud Account

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
	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")

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

}

func (suite *BillingAPITestSuite) TestRedeemCouponWithInvalidCouponId() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	//Create Clloud Account
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
	assert.Contains(suite.T(), data, "Invalid coupon code format", "Failed: Validation Failed in Coupon Redemption")

}

func (suite *BillingAPITestSuite) TestRedeemCouponPremiumSetExpiry2Days() {
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
	ret_value6, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value6, "False", "Test Failed while deleting the cloud account(Premium user)")
	creation_time, expirationtime := billing.GetLesserExpirationTime()
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
	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)
	fmt.Println("ret_val", ret_value)
	fmt.Println("data", data)
}

func (suite *BillingAPITestSuite) Test_Standard_Expire_Coupon_Validate_Instance_Creation() {
	// Actual failure
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a Standard User Cloud Account using Enroll API")
	computeUrl := utils.Get_Compute_Base_Url()
	base_url := utils.Get_Base_Url1()
	baseUrl := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	logger.Logf.Infof(" Billing base Url :  %s", base_url)
	url := base_url + "/v1/cloudaccounts"

	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	//CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	//assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create and redeem normal coupon

	coupon_expire_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Standard", int64(5), int64(2), get_CAcc_id, time.Duration(2))
	assert.Equal(suite.T(), coupon_expire_err, nil, "Failed to create coupon with shorter expiry, failed with error : %s", coupon_expire_err)

	// Now launch paid instance and see API throws 403 error

	time.Sleep(8 * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(get_CAcc_id, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Validate Cloud Account attributes

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, true, "Cloud Account Not Listed in Deactivation List After Expired credits")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl = utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), remainingAmount, float64(0), "Failed to validate remaining credits")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 404, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	// Validate cloud credit data
	zeroamt := 0
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(zeroamt), totalRemainingAmount, "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Upgrade_with_expired_coupon() {
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
	logger.Log.Info("Starting Billing Test : Redeem coupon for Standard cloud account")
	time.Sleep(5 * time.Minute)
	base_url = utils.Get_Base_Url1() + "/v1/cloudaccounts/upgrade"
	Coupon_api_payload := financials_utils.EnrichUpgradeCouponPayload(financials_utils.GetUpgradeCouponPayload(), get_CAcc_id, "ACCOUNT_TYPE_PREMIUM", couponCode)
	logger.Logf.Infof("Upgrade coupon payload : %s ", Coupon_api_payload)
	response_status, responseBody := financials.UpgradeWithCoupon(base_url, userToken, Coupon_api_payload)
	logger.Logf.Infof("Upgrade responseBody : %s ", responseBody)
	assert.NotEqual(suite.T(), 200, response_status, "Test Failed while upgradeing to premium with expired coupon")

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
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	ret_value6, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value6, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Expire_Coupon_Validate_Instance_Creation() {
	// pass
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	logger.Log.Info("Starting Test : Create cloud Coupons")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")

	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	// Create and redeem normal coupon

	coupon_expire_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Intel", int64(5), int64(2), get_CAcc_id, time.Duration(2))
	assert.Equal(suite.T(), coupon_expire_err, nil, "Failed to create coupon with shorter expiry, failed with error : %s", coupon_expire_err)

	time.Sleep(5 * time.Minute)

	// Now launch paid instance and see API throws 403 error

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Wait for the coupon to expire

	// Validate Cloud Account attributes
	/* Schedulers do not react that fast to reflect the changes in the cloud account attributes using a short expiry */
	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, false, "Cloud Account Not Listed in Deactivation List After Expired credits")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), remainingAmount, float64(0), "Failed to validate remaining credits")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) TestRedeemExpiredCouponPremium() {
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
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	upgrade_err := billing.Upgrade_to_Premium_with_coupon(get_CAcc_id, authToken, userToken, int64(300))
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)
	ret_value1, _ = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	creation_time, expirationtime = billing.GetExpirationInOneMinute()
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

	ret_value6, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value6, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) TestUpgradeExpiredCouponPremium() {
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
	logger.Log.Info("Starting Billing Test : Redeem coupon for Standard cloud account")
	time.Sleep(5 * time.Minute)
	base_url = utils.Get_Base_Url1() + "/v1/cloudaccounts/upgrade"
	Coupon_api_payload := financials_utils.EnrichUpgradeCouponPayload(financials_utils.GetUpgradeCouponPayload(), get_CAcc_id, "ACCOUNT_TYPE_PREMIUM", couponCode)
	logger.Logf.Infof("Upgrade coupon payload : %s ", Coupon_api_payload)
	response_status, responseBody := financials.UpgradeWithCoupon(base_url, userToken, Coupon_api_payload)
	logger.Logf.Infof("Upgrade responseBody : %s ", responseBody)
	assert.NotEqual(suite.T(), 200, response_status, "Test Failed while upgradeing to premium with expired coupon")

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
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	ret_value6, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value6, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) TestUpgradeWithStandardCoupon() {
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
	creation_time, expirationtime := billing.GetCreationExpirationTime()
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
	logger.Log.Info("Starting Billing Test : Redeem coupon for Standard cloud account")
	time.Sleep(1 * time.Minute)
	base_url = utils.Get_Base_Url1() + "/v1/cloudaccounts/upgrade"
	Coupon_api_payload := financials_utils.EnrichUpgradeCouponPayload(financials_utils.GetUpgradeCouponPayload(), get_CAcc_id, "ACCOUNT_TYPE_PREMIUM", couponCode)
	logger.Logf.Infof("Upgrade coupon payload : %s ", Coupon_api_payload)
	response_status, responseBody := financials.UpgradeWithCoupon(base_url, userToken, Coupon_api_payload)
	logger.Logf.Infof("Upgrade responseBody : %s ", responseBody)
	assert.NotEqual(suite.T(), 200, response_status, "Test Failed while upgradeing to premium with expired coupon")

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
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	ret_value6, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value6, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) TestRedeemCouponWithSpecialCharacter() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
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
	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
	time.Sleep(1 * time.Minute)

	couponCode = couponCode + "\u0000"
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	token := "Bearer " + userToken
	redeem_coupon_endpoint := utils.Get_Credits_Base_Url() + "/coupons/redeem"
	redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), couponCode, get_CAcc_id)
	coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
	assert.Equal(suite.T(), coupon_redeem_status, 400, "Test Failed: Redeemed empty coupon ")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account(Intel User)")

}

func (suite *BillingAPITestSuite) Test_Redeem_Empty_CouponCode() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")

	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	token := "Bearer " + userToken
	redeem_coupon_endpoint := utils.Get_Credits_Base_Url() + "/coupons/redeem"
	redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), "", get_CAcc_id)
	coupon_redeem_status, body := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
	assert.Equal(suite.T(), coupon_redeem_status, 400, "Test Failed: Redeemed empty coupon ")
	assert.Contains(suite.T(), body, "missing coupon code", "Test Failed: Redeemed empty coupon ")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account(Intel User)")

}

func (suite *BillingAPITestSuite) Test_Redeem_CouponCode_with_spaces() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")

	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	token := "Bearer " + userToken
	redeem_coupon_endpoint := utils.Get_Credits_Base_Url() + "/coupons/redeem"
	redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), "  ", get_CAcc_id)
	logger.Log.Info("Negative Coupon Payload: " + redeem_payload)
	coupon_redeem_status, body := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
	assert.Equal(suite.T(), coupon_redeem_status, 500, "Test Failed: Redeemed empty coupon ")
	assert.Contains(suite.T(), body, "Invalid coupon code format", "Test Failed: Redeemed empty coupon ")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account(Intel User)")

}
