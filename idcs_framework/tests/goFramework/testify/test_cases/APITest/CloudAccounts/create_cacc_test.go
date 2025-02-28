//go:build All || Functional || CreateCloudAccounts || Create33 || Regression
// +build All Functional CreateCloudAccounts Create33 Regression

package CaAPITest

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/utils"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// IDC_1.0_CAcc_Create_SU
func (suite *CaAPITestSuite) Test_CAcc_create_SU() {
	logger.Log.Info("Creating a Standard User Cloud Account")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account(Standard User)")
}

// IDC_1.0_CAcc_Create_PU
func (suite *CaAPITestSuite) Test_CAcc_create_PU() {
	logger.Log.Info("Creating a Premium User Cloud Account")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account")
	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account(Premium user)")
}

// IDC_1.0_CAcc_Create_EU
func (suite *CaAPITestSuite) Test_CAcc_create_EU() {
	logger.Log.Info("Creating an Enterprise User Cloud Account")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_ENTERPRISE", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account")
	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account(Enterprise user)")
}

// IDC_1.0_CAcc_Create_IU
func (suite *CaAPITestSuite) Test_CAcc_create_IU() {
	logger.Log.Info("Creating an Intel User Cloud Account")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a intel user cloud account")
	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account(Intel user)")
}

// IDC_1.0_CAcc_Create_PU_Enroll
func (suite *CaAPITestSuite) Test_CAcc_create_PU_with_Enroll() {
	logger.Log.Info("Creating a Premium User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "premimum1-eid1"
	userName := utils.Get_UserName("Premium")
	get_CAcc_id, acc_type, jsonStr := cloudAccounts.CreateCAccwithEnroll("premium", tid, userName, enterpriseId, true, 200)
	//assert.Equal(suite.T(), "ENROLL_ACTION_COUPON_OR_CREDIT_CARD", gjson.Get(jsonStr, "action").String(), "Test Failed while validating type of cloud account")
	fmt.Println("ACC TYPE", acc_type)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_PREMIUM", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	//paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	//assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "IN", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	// Add coupon to the user
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
	assert.Equal(suite.T(), getret_value, true, "Failed: Get on Coupon Failed")

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
	couponData = gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	// Now enroll user
	get_CAcc_id, acc_type, jsonStr = cloudAccounts.CreateCAccwithEnroll("premium", tid, userName, enterpriseId, true, 200)
	assert.Equal(suite.T(), "ENROLL_ACTION_NONE", gjson.Get(jsonStr, "action").String(), "Test Failed while validating type of cloud account")
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_PREMIUM", "Test Failed while validating type of cloud account")
	_, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred = gjson.Get(responsePayload, "lowCredits").String()
	CountryCode = gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "true", paidSerAllowed, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "IN", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")
}

// IDC_1.0_CAcc_Create_SU_Enroll
func (suite *CaAPITestSuite) Test_CAcc_create_SU_with_Enroll() {
	logger.Log.Info("Creating a Standard User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "standard1-eid1"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account(Standard User)")
}

// IDC_1.0_CAcc_Create_IU_Enroll
//func (suite *CaAPITestSuite) Test_CAcc_create_IU_with_Enroll() {
//	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
//	tid := cloudAccounts.Rand_token_payload_gen()
//	enterpriseId := "inteleid-01"
//	userName := "testintel@intel.com"
//	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
//	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
//	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
//	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
//	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
//	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
//	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account(Intel User)")
//}

// IDC_1.0_CAcc_Create_EU_Enroll
// func (suite *CaAPITestSuite) Test_CAcc_create_EU_with_Enroll() {
// 	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	enterpriseId := "enterprise1-eid1"
// 	userName := "test.enterprise1@proton.me"
// 	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("enterprise", tid, userName, enterpriseId, true, 200)
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_ENTERPRISE", "Test Failed while validating type of cloud account")
// 	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
// 	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account(Enterprise User)")
// }
