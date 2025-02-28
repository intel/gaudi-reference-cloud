//go:build Functional || Coupons || Regression || Billing
// +build Functional Coupons Regression Billing

package BillingAPITest

import (
	"encoding/json"
	"fmt"
	_ "fmt"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"os"

	//"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/common/logger"
	//"strconv"
	"goFramework/utils"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) TestCreateCouponWithoutAmount() {
	logger.Log.Info("Starting Test : Create cloud Coupons Without Amount")
	createCoupon := utils.Get_Coupon_CreatePayload("WithoutAmount")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	assert.Equal(suite.T(), err_message, "amount should be greater than 0", "Failed: Create Coupon Failed to validate Missing Amount")
}

func (suite *BillingAPITestSuite) TestCreateCouponWithZeroAmount() {
	logger.Log.Info("Starting Test : Create cloud Coupons Without Amount")
	createCoupon := utils.Get_Coupon_CreatePayload("WithZeroAmount")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	assert.Equal(suite.T(), err_message, "amount should be greater than 0", "Failed: Create Coupon Failed to validate with zero Amount")
}

func (suite *BillingAPITestSuite) TestCreateCouponWithInvalidTypeAmount() {
	logger.Log.Info("Starting Test : Create cloud Coupons Invalid Type in  Amount")
	createCoupon := utils.Get_Coupon_CreatePayload("InvlidTypeAmount")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	assert.Contains(suite.T(), err_message, "invalid value for double type", "Failed: Create Coupon Failed to validate with invalid type for Amount")
}

func (suite *BillingAPITestSuite) TestCreateCouponWithInvalidTypeForCreator() {
	logger.Log.Info("Starting Test : Create cloud Coupons Without Amount")
	createCoupon := utils.Get_Coupon_CreatePayload("InvlidTypeCreator")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	assert.Contains(suite.T(), err_message, "invalid value for string type", "Failed: Create Coupon Failed to validate with invalid type for Amount")
}

func (suite *BillingAPITestSuite) TestCreateCouponWithInvalidTypeForExpiration() {
	logger.Log.Info("Starting Test : Create cloud Coupons Without Amount")
	createCoupon := utils.Get_Coupon_CreatePayload("InvlidTypeCreator")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	assert.Contains(suite.T(), err_message, "invalid value for string type", "Failed: Create Coupon Failed to validate with invalid type for Amount")
}

func (suite *BillingAPITestSuite) TestCreateCouponWithInvalidExpiration() {
	logger.Log.Info("Starting Test : Create cloud Coupons With Invalid Expiration")
	createCoupon := utils.Get_Coupon_CreatePayload("InvlidTypeExpires")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	assert.Contains(suite.T(), err_message, "unexpected token 10", "Failed: Create Coupon Failed to validate with invalid Expiration")
}

// func (suite *BillingAPITestSuite) TestCreateCouponWithoutProviidingExpiration() {
// 	logger.Log.Info("Starting Test : Create cloud Coupons Without Expiration")
// 	createCoupon := utils.Get_Coupon_CreatePayload("WithoutExpiration")
// 	req := []byte(createCoupon)
// 	_, data := billing.CreateCoupon(req, 200)
// 	creationTime := gjson.Get(data, "created").String()
// 	expirationTime := billing.GetExpirationTime(creationTime)
// 	assert.Equal(suite.T(), expirationTime, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate without Expiration")
// }

func (suite *BillingAPITestSuite) TestCreateCouponWithoutProviidingStartTime() {
	logger.Log.Info("Starting Test : Create cloud Coupons Without Start Time")
	createCoupon := utils.Get_Coupon_CreatePayload("WithoutStart")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 200)
	assert.Equal(suite.T(), gjson.Get(data, "created").String(), gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate without Start time")
}

func (suite *BillingAPITestSuite) TestCreateCouponWithInvalidStart() {
	logger.Log.Info("Starting Test : Create cloud Coupons With Invalid Start time")
	createCoupon := utils.Get_Coupon_CreatePayload("WithInvalidStart")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	assert.Contains(suite.T(), err_message, "unexpected token 10", "Failed: Create Coupon Failed to validate with invalid Start time")
}

// func (suite *BillingAPITestSuite) TestCreateCouponWithoutStartandExpires() {
// 	logger.Log.Info("Starting Test : Create cloud Coupons Without Start & Expiration")
// 	createCoupon := utils.Get_Coupon_CreatePayload("WithoutStartExpire")
// 	req := []byte(createCoupon)
// 	_, data := billing.CreateCoupon(req, 200)
// 	creationTime := gjson.Get(data, "created").String()
// 	expirationTime := billing.GetExpirationTime(creationTime)
// 	assert.Equal(suite.T(), expirationTime, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate without Expiration")
// 	assert.Equal(suite.T(), creationTime, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate without Start time")
// }

func (suite *BillingAPITestSuite) TestCreateCouponWithInvalidNumberofuses() {
	logger.Log.Info("Starting Test : Create cloud Coupons With Invalid numUses")
	createCoupon := utils.Get_Coupon_CreatePayload("WithInvalidnumUses")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	assert.Contains(suite.T(), err_message, "invalid value for uint32 type:", "Failed: Create Coupon Failed to validate with invalid numUses")
}

// func (suite *BillingAPITestSuite) TestCreateCouponWithPreviousStartTime() {
// 	logger.Log.Info("Starting Test : Create cloud Coupons With Previous Start time")
// 	createCoupon := utils.Get_Coupon_CreatePayload("PreviousStart")
// 	req := []byte(createCoupon)
// 	_, data := billing.CreateCoupon(req, 400)
// 	err_message := gjson.Get(data, "message").String()
// 	assert.Contains(suite.T(), err_message, "cannot be smaller than current time", "Failed: Create Coupon Failed to validate with Previous Start Time")
// }

func (suite *BillingAPITestSuite) TestCreateCouponWithPreviousExpiratioinTime() {
	logger.Log.Info("Starting Test : Create cloud Coupons With Previous Start time")
	createCoupon := utils.Get_Coupon_CreatePayload("PreviousExpire")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	assert.Contains(suite.T(), err_message, "cannot be smaller than current time", "Failed: Create Coupon Failed to validate with Previous Expires Time")
}

func (suite *BillingAPITestSuite) TestCreateCouponWithoutNumUses() {
	logger.Log.Info("Starting Test : Create cloud Coupons Without Num Uses")
	createCoupon := utils.Get_Coupon_CreatePayload("WithoutNumUses")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	assert.Equal(suite.T(), true, strings.Contains(err_message, "number of uses should be atleast 1"), "Failed: Create Coupon Failed to validate without numUses")
}

func (suite *BillingAPITestSuite) TestCreateCouponWithStartTimeGreaterthanExpires() {
	logger.Log.Info("Starting Test : Create cloud Coupons With Start time Greater then expires")
	createCoupon := utils.Get_Coupon_CreatePayload("Time1")
	req := []byte(createCoupon)
	_, data := billing.CreateCoupon(req, 400)
	err_message := gjson.Get(data, "message").String()
	fmt.Println(err_message)
	//assert.Contains(suite.T(), err_message, "cannot be smaller than start time", "Failed: Create Coupon Failed to validate with Start time Greater then expires")
}

func (suite *BillingAPITestSuite) Test_Create_Coupon_NonStandard_with_Nouses_set_to_zero() {
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
		NumUses:    0,
		IsStandard: false,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 400)
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	assert.Equal(suite.T(), gjson.Get(data, "message").String(), "number of uses should be atleast 1", "Failed: Create Coupon Failed")

}

func (suite *BillingAPITestSuite) Test_Create_Coupon_Invalid_Value_for_numUses() {
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
		NumUses:    -10,
		IsStandard: false,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 400)
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	assert.Contains(suite.T(), gjson.Get(data, "message").String(), "invalid value for uint32 type: -10", "Failed: Create Coupon Failed")

}
