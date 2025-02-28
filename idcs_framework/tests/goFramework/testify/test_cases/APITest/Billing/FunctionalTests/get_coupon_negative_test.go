//go:build Functional || Coupons || Regression || Billing
// +build Functional Coupons Regression Billing

package BillingAPITest

import (
	"fmt"
	//"goFramework/framework/common/http_client"
	"goFramework/framework/library/financials/billing"
	//"goFramework/utils"
	//"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/common/logger"
	//"github.com/stretchr/testify/assert"
	//"strconv"
)

func (suite *BillingAPITestSuite) TestGetCouponWithWrongCode() {
	logger.Log.Info("Starting Test : Create cloud Coupons Without Amount")
	_, err_message := billing.GetCoupons("ABCD", 404)
	fmt.Println("Error Message ", err_message)
	//assert.Contains(suite.T(), "not found", err_message, "Failed: Create Coupon Failed to validate Missing Amount")
}

func (suite *BillingAPITestSuite) TestGetCouponWithInvalidCode() {
	logger.Log.Info("Starting Test : Create cloud Coupons Without Amount")
	_, err_message := billing.GetCoupons("10", 404)
	fmt.Println("Error Message ", err_message)
	//assert.Contains(suite.T(), "not found", err_message, "Failed: Create Coupon Failed to validate Missing Amount")
}

func (suite *BillingAPITestSuite) TestGetCouponWithEmptyCode() {
	logger.Log.Info("Starting Test : Create cloud Coupons Without Amount")
	_, err_message := billing.GetCoupons("", 400)
	fmt.Println("Error Message ", err_message)
	//assert.Contains(suite.T(), "not found", err_message, "Failed: Create Coupon Failed to validate Missing Amount")
}

// func (suite *BillingAPITestSuite) TestGetCouponWithWrongFilter() {
// 	logger.Log.Info("Starting Test : Create cloud Coupons Without Amount")
// 	url := utils.Get_Billing_Base_Url() + "/coupons" + "?test=10"
// 	logger.Logf.Info("Get Coupons URL  : ", url)
// 	jsonStr, respCode := http_client.Get(url, 200)
// 	logger.Logf.Info("Get Coupons Output  : ", jsonStr)
// 	assert.Equal(suite.T(), 404, respCode, "Failed: Create Coupon Failed to validate with wrong filter")
// }
