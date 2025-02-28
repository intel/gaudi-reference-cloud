//go:build Functional || Coupons || Regression || Billing
// +build Functional Coupons Regression Billing

package BillingAPITest

import (
	"encoding/json"
	_ "fmt"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/utils"
	"os"

	//"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/common/logger"
	//"strconv"
	//"goFramework/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) TestCreateCoupon() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
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
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")

}

func (suite *BillingAPITestSuite) TestCreateCouponWithStartandEndTime() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	logger.Log.Info("Starting Billing Test : Create coupon without start Time and verify")
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
	//assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
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
}

func (suite *BillingAPITestSuite) TestCreateCouponWithoutCreator() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	logger.Log.Info("Starting Billing Test : Create coupon without start Time and verify")
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.CreateCouponStruct{
		Amount:  300,
		Expires: expirationtime,
		Start:   creation_time,
		NumUses: 2,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	message := gjson.Get(data, "message").String()
	assert.Equal(suite.T(), message, "creator email is required", "Failed: Create Coupon Passed with empty creator field")
	assert.Equal(suite.T(), ret_value, false, "Failed: Create Coupon Failed")
}

func (suite *BillingAPITestSuite) TestCreateCouponWithEmptyStringCreator() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	logger.Log.Info("Starting Billing Test : Create coupon without start Time and verify")
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.CreateCouponStruct{
		Amount:  300,
		Creator: "",
		Expires: expirationtime,
		Start:   creation_time,
		NumUses: 2,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	message := gjson.Get(data, "message").String()
	assert.Equal(suite.T(), message, "creator email is required", "Failed: Create Coupon Passed with empty creator field")
	assert.Equal(suite.T(), ret_value, false, "Failed: Create Coupon Failed")
}

func (suite *BillingAPITestSuite) Test_Create_Coupon_Standard_with_Nouses_morethan_500() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("standard_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    610,
		IsStandard: true,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 400)
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	assert.Equal(suite.T(), gjson.Get(data, "message").String(), "Coupon number of uses 610 exceeds the limit of 500 for standard coupons", "Failed: Create Coupon Failed")

}

func (suite *BillingAPITestSuite) Test_Create_Coupon_Standard_with_Nouses_Equal_500() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("standard_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    500,
		IsStandard: true,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	assert.Equal(suite.T(), gjson.Get(data, "numUses").String(), "500", "Failed: Create Coupon Failed")

}

func (suite *BillingAPITestSuite) Test_Create_Coupon_Standard_with_Nouses_lessthan_500() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("standard_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    50,
		IsStandard: true,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	assert.Equal(suite.T(), gjson.Get(data, "numUses").String(), "50", "Failed: Create Coupon Failed")

}

func (suite *BillingAPITestSuite) Test_Create_Coupon_NonStandard_with_Nouses_morethan_50() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("standard_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    610,
		IsStandard: false,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 400)
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	assert.Contains(suite.T(), gjson.Get(data, "message").String(), "Coupon number of uses 610 exceeds the limit of 50 for non-standard coupons", "Failed: Create Coupon Failed")

}

func (suite *BillingAPITestSuite) Test_Create_Coupon_NonStandard_with_Nouses_Equal_50() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("standard_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    50,
		IsStandard: false,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	assert.Equal(suite.T(), gjson.Get(data, "numUses").String(), "50", "Failed: Create Coupon Failed")

}

func (suite *BillingAPITestSuite) Test_Create_Coupon_NonStandard_with_Nouses_LessThan_50() {
	logger.Log.Info("Starting Test : Create cloud Coupons")
	os.Setenv("standard_user_test", "True")
	//Create Clloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")

	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    10,
		IsStandard: true,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	assert.Equal(suite.T(), gjson.Get(data, "numUses").String(), "10", "Failed: Create Coupon Failed")

}
