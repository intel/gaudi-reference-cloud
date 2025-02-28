//go:build Functional || Coupons || Regression || Billing
// +build Functional Coupons Regression Billing

package BillingAPITest

import (
	"encoding/json"
	_ "fmt"
	"goFramework/framework/library/financials/billing"

	//"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/common/logger"
	//"strconv"
	//"goFramework/utils"
	"github.com/stretchr/testify/assert"
	//"github.com/tidwall/gjson"
)

// func (suite *BillingAPITestSuite) TestDisableCouponTwiceandValidate() {
// 	logger.Log.Info("Starting Test : Create cloud Coupons")
// 	createCoupon := billing.CreateCouponStruct{
// 		Amount:  300,
// 		Creator: "idc_billing@intel.com",
// 		Expires: "2024-05-09T09:13:44Z",
// 		NumUses: 0,
// 		Start:   "2024-05-09T09:13:44Z",
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
// 	assert.Equal(suite.T(), "", gjson.Get(getdata, "result.disabled").String(), "Failed: Create Coupon Failed to validate disabled")
// 	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(getdata, "result.start").String(), "Failed: Create Coupon Failed to validate numUses")
// 	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")

// 	//disable coupon
// 	creation_time, _ := billing.GetCreationExpirationTime()
// 	disableCoupon := billing.DisableCouponStruct{
// 		Code:     couponCode,
// 		Disabled: creation_time,
// 	}

// 	jsonPayload, _ = json.Marshal(disableCoupon)
// 	req = []byte(jsonPayload)
// 	ret_value, _ = billing.DisableCoupon(req, 200)
// 	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed to disable coupon")

// 	//Get and validate the coupon details

// 	getret_value, getdata = billing.GetCoupons(couponCode, 200)
// 	assert.Equal(suite.T(), creation_time, gjson.Get(getdata, "result.disabled").String(), "Failed: Validate Disable Coupon")

// 	//Disable coupon again

// 	ret_value, _ = billing.DisableCoupon(req, 400)
// 	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed to disable coupon")
// }

func (suite *BillingAPITestSuite) TestDisableUsingWrongCouponCode() {
	logger.Log.Info("Starting Test : Create cloud Coupons")

	//disable coupon
	creation_time, _ := billing.GetCreationExpirationTime()
	disableCoupon := billing.DisableCouponStruct{
		Code:     "ABCD-EFGH-IIJKL",
		Disabled: creation_time,
	}
	jsonPayload, _ := json.Marshal(disableCoupon)
	req := []byte(jsonPayload)
	ret_value, _ := billing.DisableCoupon(req, 400)
	logger.Logf.Infof("Response from Disable COupon  :", ret_value)
	//assert.Equal(suite.T(), true, ret_value, "Failed: Response code is not correct")
}

func (suite *BillingAPITestSuite) TestDisableUsingWithoutCouponCode() {
	logger.Log.Info("Starting Test : Create cloud Coupons")

	//disable coupon
	creation_time, _ := billing.GetCreationExpirationTime()
	disableCoupon := billing.DisableCouponStruct{
		Disabled: creation_time,
	}
	jsonPayload, _ := json.Marshal(disableCoupon)
	req := []byte(jsonPayload)
	ret_value, _ := billing.DisableCoupon(req, 400)
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed to disable coupon")
}

func (suite *BillingAPITestSuite) TestDisableUsingWithoutDisabled() {
	logger.Log.Info("Starting Test : Create cloud Coupons")

	//disable coupon
	creation_time, _ := billing.GetCreationExpirationTime()
	disableCoupon := billing.DisableCouponStruct{
		Disabled: creation_time,
	}
	jsonPayload, _ := json.Marshal(disableCoupon)
	req := []byte(jsonPayload)
	ret_value, _ := billing.DisableCoupon(req, 400)
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed to disable coupon")
}
