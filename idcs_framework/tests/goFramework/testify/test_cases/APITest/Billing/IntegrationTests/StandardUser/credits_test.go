//go:build Functional || Billing || Standard || StandardIntegration || Integration || CreditsTest
// +build Functional Billing Standard StandardIntegration Integration CreditsTest

package StandardBillingAPITest

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// var met_ret bool

func (suite *BillingAPITestSuite) Test_Standard_Add_Credit_And_Check_Credit_Response() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	logger.Log.Info("Starting Test : Create cloud Coupons")
	logger.Log.Info("Starting Billing Test : Create coupon without start Time and verify")
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

	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: cloudAccId,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), cloudAccId, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}

	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCreditsByHistory(baseUrl, userToken, cloudAccId, "true")
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(300), totalRemainingAmount, "Failed: Failed to validate Total Remaining Amount")

	totalUsedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	assert.Equal(suite.T(), float64(0), totalUsedAmount, "Failed: Failed to validate Total Used Amount")

	totalUnappliedAmount := gjson.Get(responseBody, "totalUnAppliedAmount").Float()
	assert.Equal(suite.T(), float64(0), totalUnappliedAmount, "Failed: Failed to validate Total Unapplied  Amount")

	expirationDate := gjson.Get(responseBody, "expirationDate").String()
	// expirationDateTimeStamp1, _ := time.Parse(time.RFC3339, expirationDate)
	// expirationDateTimeStamp := expirationDateTimeStamp1.Format("2024-08-17T11:48:54Z")
	assert.Equal(suite.T(), expirationDate, expirationtime, "Failed: Failed to validate Credit expiration")

	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)
	found := false
	logger.Logf.Info("credit responseBody: ", responseBody)
	result := gjson.Parse(responseBody)
	arr := gjson.Get(result.String(), "credits")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Credits Data : ", data)
		if gjson.Get(data, "couponCode").String() == couponCode {
			found = true
			created := gjson.Get(data, "created").String()
			redeemedTimeStamp1, _ := time.Parse(time.RFC3339, created)
			//redeemedTimeStamp := redeemedTimeStamp1.Format("2006-01-02")
			startTimeStamp1, _ := time.Parse(time.RFC3339, creation_time)
			//startTimeStamp := startTimeStamp1.Format("2006-01-02")
			assert.Equal(suite.T(), true, redeemedTimeStamp1.After(startTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//timediff := redeemedTimeStamp1.Sub(startTimeStamp1)
			//d, _ := time.ParseDuration(timediff)
			//duration := timediff.Minutes()
			now := time.Now()
			assert.Equal(suite.T(), true, now.After(redeemedTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//assert.Less(suite.T(), duration, 5*time.Minutes(), "Coupon Redemption difference is greater than 5 minutes")

			expiration_time := gjson.Get(data, "expiration").String()
			// expirationTimeStamp1, _ := time.Parse(time.RFC3339, expirationtime)
			// expirationTimeStamp := expirationTimeStamp1.Format("2024-08-17T11:48:54Z")
			assert.Equal(suite.T(), expirationtime, expiration_time, "Credit Expiration did not match")

			cloudAcc := gjson.Get(data, "cloudAccountId").String()
			assert.Equal(suite.T(), cloudAcc, cloudAccId, "Cloud Account Id did not match in credit response")

			reason := gjson.Get(data, "reason").String()
			assert.Equal(suite.T(), reason, financials_utils.GetCreditReason(), "Cloud Account Id did not match in credit reason")

			coupon_code := gjson.Get(data, "couponCode").String()
			assert.Equal(suite.T(), coupon_code, couponCode, "Coupon Code did not match")

			originalAmount := gjson.Get(data, "originalAmount").Float()
			assert.Equal(suite.T(), originalAmount, float64(300), "originalAmount did not match")

			remainingAmount := gjson.Get(data, "remainingAmount").Float()
			assert.Equal(suite.T(), remainingAmount, float64(300), "remainingAmount did not match")

			amountUsed := gjson.Get(data, "amountUsed").Float()
			assert.Equal(suite.T(), amountUsed, float64(0), "amountUsed did not match")
		}
		return true // keep iterating
	})
	assert.Equal(suite.T(), found, true, "Test Failed while validating credits data, coupon Code not found in response for (Standard user)")

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Add_Expired_Credit_And_Check_Credit_Response() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Now launch paid instance and see API throws 403 error

	logger.Log.Info("Starting Test : Create cloud Coupons")
	logger.Log.Info("Starting Billing Test : Create coupon without start Time and verify")

	creation_time, expirationtime := billing.GetExpirationInThreeMinute()
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

	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: cloudAccId,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), cloudAccId, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}

	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCreditsByHistory(baseUrl, userToken, cloudAccId, "true")
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(300), totalRemainingAmount, "Failed: Failed to validate Total Remaining Amount")

	totalUsedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	assert.Equal(suite.T(), float64(0), totalUsedAmount, "Failed: Failed to validate Total Used Amount")

	totalUnappliedAmount := gjson.Get(responseBody, "totalUnAppliedAmount").Float()
	assert.Equal(suite.T(), float64(0), totalUnappliedAmount, "Failed: Failed to validate Total Unapplied  Amount")

	expirationDate := gjson.Get(responseBody, "expirationDate").String()
	// expirationDateTimeStamp1, _ := time.Parse(time.RFC3339, expirationDate)
	// expirationDateTimeStamp := expirationDateTimeStamp1.Format("2024-08-17T11:48:54Z")
	assert.Equal(suite.T(), expirationDate, expirationtime, "Failed: Failed to validate Credit expiration")

	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)
	found := false
	logger.Logf.Info("credit responseBody: ", responseBody)
	result := gjson.Parse(responseBody)
	arr := gjson.Get(result.String(), "credits")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Credits Data : ", data)
		if gjson.Get(data, "couponCode").String() == couponCode {
			found = true
			created := gjson.Get(data, "created").String()
			redeemedTimeStamp1, _ := time.Parse(time.RFC3339, created)
			//redeemedTimeStamp := redeemedTimeStamp1.Format("2006-01-02")
			startTimeStamp1, _ := time.Parse(time.RFC3339, creation_time)
			//startTimeStamp := startTimeStamp1.Format("2006-01-02")
			assert.Equal(suite.T(), true, redeemedTimeStamp1.After(startTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//timediff := redeemedTimeStamp1.Sub(startTimeStamp1)
			//d, _ := time.ParseDuration(timediff)
			//duration := timediff.Minutes()
			now := time.Now()
			assert.Equal(suite.T(), true, now.After(redeemedTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//assert.Less(suite.T(), duration, 5*time.Minutes(), "Coupon Redemption difference is greater than 5 minutes")

			expiration_time := gjson.Get(data, "expiration").String()
			expTimeS, _ := time.Parse(time.RFC3339, expiration_time)
			testExpiry := time.Now().Add(3 * time.Minute)
			assert.Equal(suite.T(), true, testExpiry.After(expTimeS), "Failed: Validation failed for expiration time stamp")
			// expirationTimeStamp1, _ := time.Parse(time.RFC3339, expirationtime)
			// expirationTimeStamp := expirationTimeStamp1.Format("2024-08-17T11:48:54Z")
			assert.Equal(suite.T(), expirationtime, expiration_time, "Credit Expiration did not match")

			cloudAcc := gjson.Get(data, "cloudAccountId").String()
			assert.Equal(suite.T(), cloudAcc, cloudAccId, "Cloud Account Id did not match in credit response")

			reason := gjson.Get(data, "reason").String()
			assert.Equal(suite.T(), reason, financials_utils.GetCreditReason(), "Cloud Account Id did not match in credit reason")

			coupon_code := gjson.Get(data, "couponCode").String()
			assert.Equal(suite.T(), coupon_code, couponCode, "Coupon Code did not match")

			originalAmount := gjson.Get(data, "originalAmount").Float()
			assert.Equal(suite.T(), originalAmount, float64(300), "originalAmount did not match")

			remainingAmount := gjson.Get(data, "remainingAmount").Float()
			assert.Equal(suite.T(), remainingAmount, float64(300), "remainingAmount did not match")

			amountUsed := gjson.Get(data, "amountUsed").Float()
			assert.Equal(suite.T(), amountUsed, float64(0), "amountUsed did not match")
		}
		return true // keep iterating
	})
	assert.Equal(suite.T(), found, true, "Test Failed while validating credits data, coupon Code not found in response for (Standard user)")

	// Wait for coupon to Expire

	time.Sleep(4 * time.Minute)

	response_status, responseBody = financials.GetCreditsByHistory(baseUrl, userToken, cloudAccId, "true")
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount = gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(0), totalRemainingAmount, "Failed: Failed to validate Total Remaining Amount")

	totalUsedAmount = gjson.Get(responseBody, "totalUsedAmount").Float()
	assert.Equal(suite.T(), float64(0), totalUsedAmount, "Failed: Failed to validate Total Used Amount")

	totalUnappliedAmount = gjson.Get(responseBody, "totalUnAppliedAmount").Float()
	assert.Equal(suite.T(), float64(0), totalUnappliedAmount, "Failed: Failed to validate Total Unapplied  Amount")

	expirationDate = gjson.Get(responseBody, "expirationDate").String()
	// expirationDateTimeStamp1, _ := time.Parse(time.RFC3339, expirationDate)
	// expirationDateTimeStamp := expirationDateTimeStamp1.Format("2024-08-17T11:48:54Z")
	assert.Equal(suite.T(), expirationDate, expirationtime, "Failed: Failed to validate Credit expiration")

	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)
	found = false
	logger.Logf.Info("credit responseBody: ", responseBody)
	result = gjson.Parse(responseBody)
	arr = gjson.Get(result.String(), "credits")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Credits Data : ", data)
		if gjson.Get(data, "couponCode").String() == couponCode {
			found = true
			created := gjson.Get(data, "created").String()
			redeemedTimeStamp1, _ := time.Parse(time.RFC3339, created)
			//redeemedTimeStamp := redeemedTimeStamp1.Format("2006-01-02")
			startTimeStamp1, _ := time.Parse(time.RFC3339, creation_time)
			//startTimeStamp := startTimeStamp1.Format("2006-01-02")
			assert.Equal(suite.T(), true, redeemedTimeStamp1.After(startTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//timediff := redeemedTimeStamp1.Sub(startTimeStamp1)
			//d, _ := time.ParseDuration(timediff)
			//duration := timediff.Minutes()
			now := time.Now()
			assert.Equal(suite.T(), true, now.After(redeemedTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//assert.Less(suite.T(), duration, 5*time.Minutes(), "Coupon Redemption difference is greater than 5 minutes")

			expiration_time := gjson.Get(data, "expiration").String()
			expTimeS, _ := time.Parse(time.RFC3339, expiration_time)
			testExpiry := time.Now().Add(3 * time.Minute)
			assert.Equal(suite.T(), true, testExpiry.After(expTimeS), "Failed: Validation failed for expiration time stamp")
			// expirationTimeStamp1, _ := time.Parse(time.RFC3339, expirationtime)
			// expirationTimeStamp := expirationTimeStamp1.Format("2024-08-17T11:48:54Z")
			assert.Equal(suite.T(), expirationtime, expiration_time, "Credit Expiration did not match")

			cloudAcc := gjson.Get(data, "cloudAccountId").String()
			assert.Equal(suite.T(), cloudAcc, cloudAccId, "Cloud Account Id did not match in credit response")

			reason := gjson.Get(data, "reason").String()
			assert.Equal(suite.T(), reason, financials_utils.GetCreditReason(), "Cloud Account Id did not match in credit reason")

			coupon_code := gjson.Get(data, "couponCode").String()
			assert.Equal(suite.T(), coupon_code, couponCode, "Coupon Code did not match")

			originalAmount := gjson.Get(data, "originalAmount").Float()
			assert.Equal(suite.T(), originalAmount, float64(300), "originalAmount did not match")

			remainingAmount := gjson.Get(data, "remainingAmount").Float()
			assert.Equal(suite.T(), remainingAmount, float64(300), "remainingAmount did not match")

			amountUsed := gjson.Get(data, "amountUsed").Float()
			assert.Equal(suite.T(), amountUsed, float64(0), "amountUsed did not match")
		}
		return true // keep iterating
	})

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Add_Multiple_Credit_And_Check_Credit_Response() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Now launch paid instance and see API throws 403 error

	logger.Log.Info("Starting Test : Create cloud Coupons")
	logger.Log.Info("Starting Billing Test : Create coupon without start Time and verify")
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     35,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    2,
		IsStandard: true,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode1 := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	//assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode1, 200)
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

	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: cloudAccId,
		Code:           couponCode1,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode1, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), cloudAccId, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode1, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}

	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "240000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire
	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsage(cloudAccId, float64(30), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCreditsByHistory(baseUrl, userToken, cloudAccId, "true")
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(5), totalRemainingAmount, "Failed: Failed to validate Total Remaining Amount")

	totalUsedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	assert.Equal(suite.T(), float64(30), totalUsedAmount, "Failed: Failed to validate Total Used Amount")

	totalUnappliedAmount := gjson.Get(responseBody, "totalUnAppliedAmount").Float()
	assert.Equal(suite.T(), float64(0), totalUnappliedAmount, "Failed: Failed to validate Total Unapplied  Amount")

	expirationDate := gjson.Get(responseBody, "expirationDate").String()
	// expirationDateTimeStamp1, _ := time.Parse(time.RFC3339, expirationDate)
	// expirationDateTimeStamp := expirationDateTimeStamp1.Format("2024-08-17T11:48:54Z")
	assert.Equal(suite.T(), expirationDate, expirationtime, "Failed: Failed to validate Credit expiration")

	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)
	found := false
	logger.Logf.Info("credit responseBody: ", responseBody)
	result := gjson.Parse(responseBody)
	arr := gjson.Get(result.String(), "credits")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Credits Data : ", data)
		if gjson.Get(data, "couponCode").String() == couponCode1 {
			found = true
			created := gjson.Get(data, "created").String()
			redeemedTimeStamp1, _ := time.Parse(time.RFC3339, created)
			startTimeStamp1, _ := time.Parse(time.RFC3339, creation_time)
			assert.Equal(suite.T(), true, redeemedTimeStamp1.After(startTimeStamp1), "Failed: Redeemed time is not coming after creation time")

			now := time.Now()
			assert.Equal(suite.T(), true, now.After(redeemedTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//assert.Less(suite.T(), duration, 5*time.Minutes(), "Coupon Redemption difference is greater than 5 minutes")

			expiration_time := gjson.Get(data, "expiration").String()
			assert.Equal(suite.T(), expirationtime, expiration_time, "Credit Expiration did not match")

			cloudAcc := gjson.Get(data, "cloudAccountId").String()
			assert.Equal(suite.T(), cloudAcc, cloudAccId, "Cloud Account Id did not match in credit response")

			reason := gjson.Get(data, "reason").String()
			assert.Equal(suite.T(), reason, financials_utils.GetCreditReason(), "Cloud Account Id did not match in credit reason")

			coupon_code := gjson.Get(data, "couponCode").String()
			assert.Equal(suite.T(), coupon_code, couponCode1, "Coupon Code did not match")

			originalAmount := gjson.Get(data, "originalAmount").Float()
			assert.Equal(suite.T(), originalAmount, float64(35), "originalAmount did not match")

			remainingAmount := gjson.Get(data, "remainingAmount").Float()
			assert.Equal(suite.T(), remainingAmount, float64(5), "remainingAmount did not match")

			amountUsed := gjson.Get(data, "amountUsed").Float()
			assert.Equal(suite.T(), amountUsed, float64(30), "amountUsed did not match")
		}
		return true // keep iterating
	})
	assert.Equal(suite.T(), found, true, "Test Failed while validating credits data, coupon Code not found in response for (Standard user)")

	// Now redeem another coupon and check credit response

	creation_time1, expirationtime1 := billing.GetCreationExpirationTime()
	createCoupon = billing.StandardCreateCouponStruct{
		Amount:     15,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime1,
		Start:      creation_time1,
		NumUses:    2,
		IsStandard: true,
	}
	jsonPayload, _ = json.Marshal(createCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.CreateCoupon(req, 200)
	couponCode2 := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	//assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode2, 200)
	couponData = gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")

	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon = billing.RedeemCouponStruct{
		CloudAccountID: cloudAccId,
		Code:           couponCode2,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode2, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions = gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), cloudAccId, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode2, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}

	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	// Now check credit response

	response_status, responseBody = financials.GetCreditsByHistory(baseUrl, userToken, cloudAccId, "true")
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount = gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(20), totalRemainingAmount, "Failed: Failed to validate Total Remaining Amount")

	totalUsedAmount = gjson.Get(responseBody, "totalUsedAmount").Float()
	assert.Equal(suite.T(), float64(30), totalUsedAmount, "Failed: Failed to validate Total Used Amount")

	totalUnappliedAmount = gjson.Get(responseBody, "totalUnAppliedAmount").Float()
	assert.Equal(suite.T(), float64(0), totalUnappliedAmount, "Failed: Failed to validate Total Unapplied  Amount")

	expirationDate = gjson.Get(responseBody, "expirationDate").String()
	assert.Equal(suite.T(), expirationDate, expirationtime1, "Failed: Failed to validate Credit expiration")

	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)
	found1 := false
	found2 := false
	logger.Logf.Info("credit responseBody: ", responseBody)
	result = gjson.Parse(responseBody)
	arr = gjson.Get(result.String(), "credits")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Credits Data : ", data)
		if gjson.Get(data, "couponCode").String() == couponCode1 {
			found1 = true
			created := gjson.Get(data, "created").String()
			redeemedTimeStamp1, _ := time.Parse(time.RFC3339, created)
			startTimeStamp1, _ := time.Parse(time.RFC3339, creation_time)
			assert.Equal(suite.T(), true, redeemedTimeStamp1.After(startTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			now := time.Now()
			assert.Equal(suite.T(), true, now.After(redeemedTimeStamp1), "Failed: Redeemed time is not coming after creation time")

			expiration_time := gjson.Get(data, "expiration").String()
			assert.Equal(suite.T(), expirationtime, expiration_time, "Credit Expiration did not match")

			cloudAcc := gjson.Get(data, "cloudAccountId").String()
			assert.Equal(suite.T(), cloudAcc, cloudAccId, "Cloud Account Id did not match in credit response")

			reason := gjson.Get(data, "reason").String()
			assert.Equal(suite.T(), reason, financials_utils.GetCreditReason(), "Cloud Account Id did not match in credit reason")

			coupon_code := gjson.Get(data, "couponCode").String()
			assert.Equal(suite.T(), coupon_code, couponCode1, "Coupon Code did not match")

			originalAmount := gjson.Get(data, "originalAmount").Float()
			assert.Equal(suite.T(), originalAmount, float64(35), "originalAmount did not match")

			remainingAmount := gjson.Get(data, "remainingAmount").Float()
			assert.Equal(suite.T(), remainingAmount, float64(5), "remainingAmount did not match")

			amountUsed := gjson.Get(data, "amountUsed").Float()
			assert.Equal(suite.T(), amountUsed, float64(30), "amountUsed did not match")
		}

		if gjson.Get(data, "couponCode").String() == couponCode2 {
			found2 = true
			created := gjson.Get(data, "created").String()
			redeemedTimeStamp1, _ := time.Parse(time.RFC3339, created)
			//redeemedTimeStamp := redeemedTimeStamp1.Format("2006-01-02")
			startTimeStamp1, _ := time.Parse(time.RFC3339, creation_time1)
			assert.Equal(suite.T(), true, redeemedTimeStamp1.After(startTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			now := time.Now()
			assert.Equal(suite.T(), true, now.After(redeemedTimeStamp1), "Failed: Redeemed time is not coming after creation time")

			expiration_time := gjson.Get(data, "expiration").String()

			assert.Equal(suite.T(), expirationtime1, expiration_time, "Credit Expiration did not match")

			cloudAcc := gjson.Get(data, "cloudAccountId").String()
			assert.Equal(suite.T(), cloudAcc, cloudAccId, "Cloud Account Id did not match in credit response")

			reason := gjson.Get(data, "reason").String()
			assert.Equal(suite.T(), reason, financials_utils.GetCreditReason(), "Cloud Account Id did not match in credit reason")

			coupon_code := gjson.Get(data, "couponCode").String()
			assert.Equal(suite.T(), coupon_code, couponCode2, "Coupon Code did not match")

			originalAmount := gjson.Get(data, "originalAmount").Float()
			assert.Equal(suite.T(), originalAmount, float64(15), "originalAmount did not match")

			remainingAmount := gjson.Get(data, "remainingAmount").Float()
			assert.Equal(suite.T(), remainingAmount, float64(15), "remainingAmount did not match")

			amountUsed := gjson.Get(data, "amountUsed").Float()
			assert.Equal(suite.T(), amountUsed, float64(0), "amountUsed did not match")
		}
		return true // keep iterating
	})

	assert.Equal(suite.T(), found1, true, "Test Failed while validating credits data, coupon Code not found in response for (Standard user)")
	assert.Equal(suite.T(), found2, true, "Test Failed while validating credits data, coupon Code not found in response for (Standard user)")

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Add_Expired_Credit_And_Check_Credit_Response_After_Adding_More_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Now launch paid instance and see API throws 403 error

	logger.Log.Info("Starting Test : Create cloud Coupons")
	logger.Log.Info("Starting Billing Test : Create coupon without start Time and verify")
	creation_time, expirationtime := billing.GetExpirationInTime(5)
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     35,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    2,
		IsStandard: true,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode1 := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	//assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode1, 200)
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

	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: cloudAccId,
		Code:           couponCode1,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode1, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), cloudAccId, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode1, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}

	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "240000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire
	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsage(cloudAccId, float64(30), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCreditsByHistory(baseUrl, userToken, cloudAccId, "true")
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(0), totalRemainingAmount, "Failed: Failed to validate Total Remaining Amount")

	totalUsedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	assert.Equal(suite.T(), float64(30), totalUsedAmount, "Failed: Failed to validate Total Used Amount")

	totalUnappliedAmount := gjson.Get(responseBody, "totalUnAppliedAmount").Float()
	assert.Equal(suite.T(), float64(0), totalUnappliedAmount, "Failed: Failed to validate Total Unapplied  Amount")

	expirationDate := gjson.Get(responseBody, "expirationDate").String()
	// expirationDateTimeStamp1, _ := time.Parse(time.RFC3339, expirationDate)
	// expirationDateTimeStamp := expirationDateTimeStamp1.Format("2024-08-17T11:48:54Z")
	assert.Equal(suite.T(), expirationDate, expirationtime, "Failed: Failed to validate Credit expiration")

	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)
	found := false
	logger.Logf.Info("credit responseBody: ", responseBody)
	result := gjson.Parse(responseBody)
	arr := gjson.Get(result.String(), "credits")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Credits Data : ", data)
		if gjson.Get(data, "couponCode").String() == couponCode1 {
			found = true
			created := gjson.Get(data, "created").String()
			redeemedTimeStamp1, _ := time.Parse(time.RFC3339, created)
			//redeemedTimeStamp := redeemedTimeStamp1.Format("2006-01-02")
			startTimeStamp1, _ := time.Parse(time.RFC3339, creation_time)
			//startTimeStamp := startTimeStamp1.Format("2006-01-02")
			assert.Equal(suite.T(), true, redeemedTimeStamp1.After(startTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//timediff := redeemedTimeStamp1.Sub(startTimeStamp1)
			//d, _ := time.ParseDuration(timediff)
			//duration := timediff.Minutes()
			now := time.Now()
			assert.Equal(suite.T(), true, now.After(redeemedTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//assert.Less(suite.T(), duration, 5*time.Minutes(), "Coupon Redemption difference is greater than 5 minutes")

			expiration_time := gjson.Get(data, "expiration").String()
			// expirationTimeStamp1, _ := time.Parse(time.RFC3339, expirationtime)
			// expirationTimeStamp := expirationTimeStamp1.Format("2024-08-17T11:48:54Z")
			assert.Equal(suite.T(), expirationtime, expiration_time, "Credit Expiration did not match")

			cloudAcc := gjson.Get(data, "cloudAccountId").String()
			assert.Equal(suite.T(), cloudAcc, cloudAccId, "Cloud Account Id did not match in credit response")

			reason := gjson.Get(data, "reason").String()
			assert.Equal(suite.T(), reason, financials_utils.GetCreditReason(), "Cloud Account Id did not match in credit reason")

			coupon_code := gjson.Get(data, "couponCode").String()
			assert.Equal(suite.T(), coupon_code, couponCode1, "Coupon Code did not match")

			originalAmount := gjson.Get(data, "originalAmount").Float()
			assert.Equal(suite.T(), originalAmount, float64(35), "originalAmount did not match")

			remainingAmount := gjson.Get(data, "remainingAmount").Float()
			assert.Equal(suite.T(), remainingAmount, float64(5), "remainingAmount did not match")

			amountUsed := gjson.Get(data, "amountUsed").Float()
			assert.Equal(suite.T(), amountUsed, float64(30), "amountUsed did not match")
		}
		return true // keep iterating
	})
	assert.Equal(suite.T(), found, true, "Test Failed while validating credits data, coupon Code not found in response for (Standard user)")

	// Now redeem another coupon and check credit response

	creation_time1, expirationtime1 := billing.GetCreationExpirationTime()
	createCoupon = billing.StandardCreateCouponStruct{
		Amount:     15,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime1,
		Start:      creation_time1,
		NumUses:    2,
		IsStandard: true,
	}
	jsonPayload, _ = json.Marshal(createCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.CreateCoupon(req, 200)
	couponCode2 := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	//assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode2, 200)
	couponData = gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")

	logger.Log.Info("Starting Billing Test : Redeem coupon for Premium cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon = billing.RedeemCouponStruct{
		CloudAccountID: cloudAccId,
		Code:           couponCode2,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode2, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions = gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), cloudAccId, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode2, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}

	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	// Now check credit response

	response_status, responseBody = financials.GetCreditsByHistory(baseUrl, userToken, cloudAccId, "true")
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount = gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(15), totalRemainingAmount, "Failed: Failed to validate Total Remaining Amount")

	totalUsedAmount = gjson.Get(responseBody, "totalUsedAmount").Float()
	assert.Equal(suite.T(), float64(30), totalUsedAmount, "Failed: Failed to validate Total Used Amount")

	totalUnappliedAmount = gjson.Get(responseBody, "totalUnAppliedAmount").Float()
	assert.Equal(suite.T(), float64(0), totalUnappliedAmount, "Failed: Failed to validate Total Unapplied  Amount")

	expirationDate = gjson.Get(responseBody, "expirationDate").String()
	// expirationDateTimeStamp1, _ := time.Parse(time.RFC3339, expirationDate)
	// expirationDateTimeStamp := expirationDateTimeStamp1.Format("2024-08-17T11:48:54Z")
	assert.Equal(suite.T(), expirationDate, expirationtime1, "Failed: Failed to validate Credit expiration")

	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)
	found1 := false
	found2 := false
	logger.Logf.Info("credit responseBody: ", responseBody)
	result = gjson.Parse(responseBody)
	arr = gjson.Get(result.String(), "credits")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Credits Data : ", data)
		if gjson.Get(data, "couponCode").String() == couponCode1 {
			found1 = true
			created := gjson.Get(data, "created").String()
			redeemedTimeStamp1, _ := time.Parse(time.RFC3339, created)
			//redeemedTimeStamp := redeemedTimeStamp1.Format("2006-01-02")
			startTimeStamp1, _ := time.Parse(time.RFC3339, creation_time)
			//startTimeStamp := startTimeStamp1.Format("2006-01-02")
			assert.Equal(suite.T(), true, redeemedTimeStamp1.After(startTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//timediff := redeemedTimeStamp1.Sub(startTimeStamp1)
			//d, _ := time.ParseDuration(timediff)
			//duration := timediff.Minutes()
			now := time.Now()
			assert.Equal(suite.T(), true, now.After(redeemedTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//assert.Less(suite.T(), duration, 5*time.Minutes(), "Coupon Redemption difference is greater than 5 minutes")

			expiration_time := gjson.Get(data, "expiration").String()
			// expirationTimeStamp1, _ := time.Parse(time.RFC3339, expirationtime)
			// expirationTimeStamp := expirationTimeStamp1.Format("2024-08-17T11:48:54Z")
			assert.Equal(suite.T(), expirationtime, expiration_time, "Credit Expiration did not match")

			cloudAcc := gjson.Get(data, "cloudAccountId").String()
			assert.Equal(suite.T(), cloudAcc, cloudAccId, "Cloud Account Id did not match in credit response")

			reason := gjson.Get(data, "reason").String()
			assert.Equal(suite.T(), reason, financials_utils.GetCreditReason(), "Cloud Account Id did not match in credit reason")

			coupon_code := gjson.Get(data, "couponCode").String()
			assert.Equal(suite.T(), coupon_code, couponCode1, "Coupon Code did not match")

			originalAmount := gjson.Get(data, "originalAmount").Float()
			assert.Equal(suite.T(), originalAmount, float64(35), "originalAmount did not match")

			remainingAmount := gjson.Get(data, "remainingAmount").Float()
			assert.Equal(suite.T(), remainingAmount, float64(5), "remainingAmount did not match")

			amountUsed := gjson.Get(data, "amountUsed").Float()
			assert.Equal(suite.T(), amountUsed, float64(30), "amountUsed did not match")
		}

		if gjson.Get(data, "couponCode").String() == couponCode2 {
			found2 = true
			created := gjson.Get(data, "created").String()
			redeemedTimeStamp1, _ := time.Parse(time.RFC3339, created)
			//redeemedTimeStamp := redeemedTimeStamp1.Format("2006-01-02")
			startTimeStamp1, _ := time.Parse(time.RFC3339, creation_time1)
			//startTimeStamp := startTimeStamp1.Format("2006-01-02")
			assert.Equal(suite.T(), true, redeemedTimeStamp1.After(startTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//timediff := redeemedTimeStamp1.Sub(startTimeStamp1)
			//d, _ := time.ParseDuration(timediff)
			//duration := timediff.Minutes()
			now := time.Now()
			assert.Equal(suite.T(), true, now.After(redeemedTimeStamp1), "Failed: Redeemed time is not coming after creation time")
			//assert.Less(suite.T(), duration, 5*time.Minutes(), "Coupon Redemption difference is greater than 5 minutes")

			expiration_time := gjson.Get(data, "expiration").String()
			// expirationTimeStamp1, _ := time.Parse(time.RFC3339, expirationtime)
			// expirationTimeStamp := expirationTimeStamp1.Format("2024-08-17T11:48:54Z")
			assert.Equal(suite.T(), expirationtime1, expiration_time, "Credit Expiration did not match")

			cloudAcc := gjson.Get(data, "cloudAccountId").String()
			assert.Equal(suite.T(), cloudAcc, cloudAccId, "Cloud Account Id did not match in credit response")

			reason := gjson.Get(data, "reason").String()
			assert.Equal(suite.T(), reason, financials_utils.GetCreditReason(), "Cloud Account Id did not match in credit reason")

			coupon_code := gjson.Get(data, "couponCode").String()
			assert.Equal(suite.T(), coupon_code, couponCode2, "Coupon Code did not match")

			originalAmount := gjson.Get(data, "originalAmount").Float()
			assert.Equal(suite.T(), originalAmount, float64(15), "originalAmount did not match")

			remainingAmount := gjson.Get(data, "remainingAmount").Float()
			assert.Equal(suite.T(), remainingAmount, float64(15), "remainingAmount did not match")

			amountUsed := gjson.Get(data, "amountUsed").Float()
			assert.Equal(suite.T(), amountUsed, float64(0), "amountUsed did not match")
		}
		return true // keep iterating
	})

	assert.Equal(suite.T(), found1, true, "Test Failed while validating credits data, coupon Code not found in response for (Standard user)")
	assert.Equal(suite.T(), found2, true, "Test Failed while validating credits data, coupon Code not found in response for (Standard user)")

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Create_Free_Instance_Without_Credits() {
	suite.T().Skip()
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()

	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Now launch paid instance and see API throws 403 error

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-tny", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Create_Paid_Instance_Without_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Now launch paid instance and see API throws 403 error

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Create_Free_Instance_With_Coupon_Credits() {
	suite.T().Skip()
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-tny", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Create_Paid_Instance_With_Coupon_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Launch_Paid_Instance_After_Using_All_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "largevm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Launch_Free_Instance_After_Using_All_Credits() {
	suite.T().Skip()
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-tny", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Launch_Paid_Instance_After_Using_Less_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "14926")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Now launch paid instance and see API throws 403 error

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	// Validate credits
	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(5), float64(5), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Launch_Paid_Instance_After_Using_More_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Add coupon to the user

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "14926")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Now launch paid instance and see API throws 403 error

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)
	// Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(10), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Launch_Free_Instance_After_Using_More_Credits() {
	suite.T().Skip()
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Add coupon to the user

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "14926")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Now launch free instance

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-tny", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	// Validate credits

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(10), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Redeem_Coupon_After_Using_Less_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "14926")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	// Validate credits

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(5), float64(5), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Redeem coupon again

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	time.Sleep(8 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(15), float64(15), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Redeem_Coupon_After_Using_More_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Add coupon to the user

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "14926")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Validate credits

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(10), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Redeem coupon

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	time.Sleep(8 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(5), float64(5), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Redeem_Coupon_After_Using_All_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Add coupon to the user

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "14926")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	// Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Redeem coupon

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Validate credits

	time.Sleep(8 * time.Minute)
	vm_creation_error, create_response_body, skip_val = billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	credits_err = billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(15), float64(15), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Redeem_Lesser_Coupon_After_Using_All_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "14925")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Validate credits

	credits_err := billing.ValidateUsageCreditsinRange(cloudAccId, float64(14.8), float64(15), authToken, float64(0), float64(8), float64(0), float64(0.5), float64(0.2))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : %s", credits_err)

	// Redeem coupon

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Validate credits

	credits_err = billing.ValidateUsageCreditsinRange(cloudAccId, float64(14.8), float64(15), authToken, float64(0.8), float64(1.5), float64(0.8), float64(1.5), float64(0.2))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : %s", credits_err)

	// Noew launch a paid instance, instance should not be launched

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		//assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Redeem_Two_Coupons_And_Verify_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Add coupon to the user

	logger.Log.Info("Create and redeem coupon worth 15$")
	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	logger.Log.Info("Create and redeem coupon worth 10$")
	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Validate credits

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(25), float64(25), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Usage_More_Than_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	// Add coupon to the user

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsageinRange(cloudAccId, float64(0.0603), "bm-spr", authToken, float64(30), float64(31))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
}

func (suite *BillingAPITestSuite) Test_Standard_Usage_More_Than_Credits_Redeem_Lesser_Value_Coupon() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	// Add coupon to the user

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "29852")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsageinRange(cloudAccId, float64(0.0603), "bm-spr", authToken, float64(30), float64(31))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Redeem coupon with lesser value than unapplied credits

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	credits_err = billing.ValidateCredits(cloudAccId, float64(25), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Check instance should not be launched

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 404, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
}

func (suite *BillingAPITestSuite) Test_Standard_Usage_More_Than_Credits_Redeem_Lesser_Value_Coupon_After_Scheduler_Run() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	// Add coupon to the user

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsageinRange(cloudAccId, float64(0.0603), "bm-spr", authToken, float64(30), float64(31))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Redeem coupon with lesser value than unapplied credits

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Check instance should not be launched after scheduler run

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 404, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
}

func (suite *BillingAPITestSuite) Test_Standard_Expire_Coupon_Launch_Paid_Instance() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Standard", int64(15), int64(2), cloudAccId, time.Duration(5))
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 404, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	// Validate cloud credit data
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	zeroamt := 0
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(zeroamt), totalRemainingAmount, "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard User)")

}

func (suite *BillingAPITestSuite) Test_Standard_Expire_Coupon_Validate_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Standard", int64(15), int64(2), cloudAccId, time.Duration(10))
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "240000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire
	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsage(cloudAccId, float64(30), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	//Now Create a paid instance

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 404, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard User)")

}

func (suite *BillingAPITestSuite) Test_Standard_Expire_Coupon_Validate_Credits1() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	coupon_err = billing.Create_Redeem_Coupon_With_Shrt_Expiry("Standard", int64(40), int64(2), cloudAccId, time.Duration(5))
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "240000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire
	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsage(cloudAccId, float64(30), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(30), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard User)")

}

func (suite *BillingAPITestSuite) Test_Standard_ExpireCoupon_Validate_Instance_Runs_When_Credits_Available() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon
	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	coupon_err1 := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Standard", int64(15), int64(2), cloudAccId, time.Duration(10))
	assert.Equal(suite.T(), coupon_err1, nil, "Failed to create coupon, failed with error : ", coupon_err1)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	//previousDate = previousDate.Add(2 * time.Minute)
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "180000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsageinRange(cloudAccId, float64(0.0075), "vm-spr-sml", authToken, float64(22.4), float64(23))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(23), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard User)")

}

func (suite *BillingAPITestSuite) Test_Standard_Redeem_ExpireCoupon_First_Validate_Instance_Runs_When_Credits_Available() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Standard", int64(15), int64(2), cloudAccId, time.Duration(10))
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	coupon_err1 := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err1, nil, "Failed to create coupon, failed with error : ", coupon_err1)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	//previousDate = previousDate.Add(2 * time.Minute)
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "180000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsageinRange(cloudAccId, float64(0.0075), "vm-spr-sml", authToken, float64(22.4), float64(23))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateUsageCreditsinRange(cloudAccId, float64(22.5), float64(23.5), authToken, float64(6), float64(8), float64(6), float64(8), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : %s", credits_err)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard User)")
}

func (suite *BillingAPITestSuite) Test_Standard_Delete_Instance_And_Validate_Usage() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	base_url := utils.Get_Base_Url1()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")

	// Redeem coupon for standard user before upgrade

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Create instance and upgrade

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with with error : ", vm_creation_error)

	time.Sleep(15 * time.Minute)
	if vm_creation_error == nil {
		time.Sleep(5 * time.Minute)
		instance_id_created1 := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created1)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not running")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created1)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	time.Sleep(5 * time.Minute)

	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)

	var usage1 float64
	var usage2 float64
	var minsUsed1 float64
	var minsUsed2 float64

	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if gjson.Get(data, "productType").String() == "vm-spr-sml" {
			Amount := gjson.Get(data, "amount").String()
			usage1, _ = strconv.ParseFloat(Amount, 64)
			assert.Greater(suite.T(), usage1, float64(0), "Failed: Failed to validate usage amount")

			Mins := gjson.Get(data, "minsUsed").String()
			minsUsed1, _ = strconv.ParseFloat(Mins, 64)
			assert.Greater(suite.T(), minsUsed1, float64(0), "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

			logger.Logf.Infof("Actual Usage :    ", usage1)
			assert.Greater(suite.T(), usage1, float64(0), "Failed: Failed to validate usage amount")
		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.Greater(suite.T(), total_amount_from_response, float64(0), "Failed: Failed to validate usage amount")

	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	zeroamt := 0
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount1 := gjson.Get(responseBody, "totalRemainingAmount").Float()
	totalUsedAmount1 := gjson.Get(responseBody, "totalUsedAmount").Float()
	assert.Greater(suite.T(), totalRemainingAmount1, float64(zeroamt), "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount1)

	time.Sleep(2 * time.Minute)
	// Check instance is running and can report usages

	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
	get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
	logger.Logf.Info("get_response_status: ", get_response_status)
	logger.Logf.Info("get_response_body: ", get_response_body)
	assert.Equal(suite.T(), 404, get_response_status, "Failed : Instance not Found after Standard to premium upgrade")

	// Wait for some time to get usages

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	usage_response_status, usage_response_body = financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)
	result = gjson.Parse(usage_response_body)
	arr = gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if gjson.Get(data, "productType").String() == "vm-spr-sml" {
			Amount := gjson.Get(data, "amount").String()
			usage2, _ = strconv.ParseFloat(Amount, 64)
			assert.Greater(suite.T(), usage2, float64(0), "Failed: Failed to validate usage amount")

			Mins := gjson.Get(data, "minsUsed").String()
			minsUsed2, _ = strconv.ParseFloat(Mins, 64)
			assert.Equal(suite.T(), minsUsed2, minsUsed1, "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

			logger.Logf.Infof("Actual Usage :    ", usage2)
			assert.Greater(suite.T(), usage2, float64(0), "Failed: Failed to validate usage amount")

		}
		return true // keep iterating
	})

	total_amount_from_response = gjson.Get(usage_response_body, "totalAmount").Float()
	assert.Greater(suite.T(), total_amount_from_response, float64(0), "Failed: Failed to validate usage amount")
	assert.Equal(suite.T(), usage2, usage1, "Failed: Failed to validate usage amount")
	//time.Sleep(15 * time.Minute)
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount2 := gjson.Get(responseBody, "totalRemainingAmount").Float()
	totalUsedAmount2 := gjson.Get(responseBody, "totalUsedAmount").Float()
	assert.Equal(suite.T(), totalRemainingAmount2, totalRemainingAmount1, "Failed: Failed to validate expired credits")
	assert.Equal(suite.T(), totalUsedAmount1, totalUsedAmount2, "Failed: Failed to validate expired credits")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")
}
