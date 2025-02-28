//go:build Functional || UsageTest || Intel
// +build Functional UsageTest Intel

package BillingAPITest

import (
	// "encoding/json"
	// "fmt"
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/testsetup"
	"goFramework/utils"
	"time"

	//"goFramework/framework/library/auth"

	//"strconv"
	"os"
	//"goFramework/utils"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	//"goFramework/testsetup"
	//"goFramework/framework/library/financials/metering"
	//"time"
)

func (suite *BillingAPITestSuite) TestEnrollIntelCustomerCheckUsage() {
	// Enroll intel user
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
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
	assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Get usage

	ret, usageData := billing.GetUsage(get_CAcc_id, 200)
	assert.Equal(suite.T(), ret, true, "Failed to fetch usage for intel account")
	totalAmt := gjson.Get(usageData, "totalAmount").String()
	totalUsage := gjson.Get(usageData, "totalUsage").String()
	assert.Equal(suite.T(), "0", totalAmt, "Total amount is not equal to 0 after enrollment")
	assert.Equal(suite.T(), "0", totalUsage, "Total Usage is not equal to 0 after enrollment")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestEnrollIntelCustomerCheckUsageHistory() {
	// Enroll intel user
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
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
	assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Get usage
	//startTime, endTime := billing.GetCreationExpirationTime()
	ret, usageData := billing.GetUsageHistory(get_CAcc_id, "2023-06-19T19:41:41.851Z", "2023-06-19T19:41:41.851Z", 200)
	assert.Equal(suite.T(), ret, true, "Failed to fetch usage for intel account")
	totalAmt := gjson.Get(usageData, "totalAmount").String()
	totalUsage := gjson.Get(usageData, "totalUsage").String()
	assert.Equal(suite.T(), "0", totalAmt, "Total amount is not equal to 0 after enrollment")
	assert.Equal(suite.T(), "0", totalUsage, "Total Usage is not equal to 0 after enrollment")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestEnrollIntelCustomerCheckUsageHistoryStartDate() {
	// Enroll intel user
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
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
	assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Get usage
	//startTime, endTime := billing.GetCreationExpirationTime()
	ret, _ := billing.GetUsageHistory(get_CAcc_id, "2023-06-19T19:41:41.851Z", "", 400)
	assert.NotEqual(suite.T(), ret, false, "Failed to fetch usage for intel account")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestEnrollIntelCustomerCheckUsageHistoryEndDate() {
	// Enroll intel user
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
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
	assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Get usage
	//startTime, endTime := billing.GetCreationExpirationTime()
	ret, _ := billing.GetUsageHistory(get_CAcc_id, "", "2023-06-19T19:41:41.851Z", 400)
	assert.NotEqual(suite.T(), ret, false, "Failed to fetch usage for intel account")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestGetAndValidateUsageFreeAndPaidProducts() {
	// Enroll intel user
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
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
	assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	//token := utils.Get_intel_Token()
	computeUrl := utils.Get_Compute_Base_Url()
	logger.Log.Info("Compute Url" + computeUrl)
	baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	sshKeyData := utils.Get_ssh_Key_Payload("visali.vasiraju@intel.com", "testinteluserkey")
	logger.Log.Info("SSH Key data " + sshKeyData)
	err := testsetup.CreateSSHKey(sshKeyData, computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: SSH key creation failed for intel user")

	vnetData := utils.Get_Vet_Payload("visali.vasiraju@intel.com", "autotestvnet8990")
	logger.Log.Info("VNet data " + vnetData)
	err = testsetup.CreateVnet(vnetData, computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: Vnet creation failed for intel user")

	instanceData := utils.Get_Instance_Payload("visali.vasiraju@intel.com", "testintelfree")
	logger.Log.Info("Instance data " + instanceData)
	//instanceName := gjson.Get(instanceData, "name").String()
	err = testsetup.CreateVmInstance(instanceData, computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: Vnet creation failed for intel user")

	time.Sleep(3 * time.Minute)

	_, err1 := testsetup.GetUsageAndValidateTotalUsage("visali.vasiraju@intel.com", "vm-spr-tny", baseUrl, authToken, computeUrl)
	assert.Equal(suite.T(), nil, err1, "Failed: Validating Usage failed for intel user for free product")

	//Cleanup

	err := testsetup.DeleteSSHKey("testinteluserkey", "visali.vasiraju@intel.com", computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: SSH Key  deletion failed")
	err = testsetup.DeleteVnet("autotestvnet8990", "visali.vasiraju@intel.com", computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: VNet  deletion failed")
	err = testsetup.DeleteInstance("testintelfree", "visali.vasiraju@intel.com", computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: Instance deletion failed")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account(Intel User)")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestGetAndValidateCreditDepletion() {
	// Enroll intel user
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	//paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	///lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	//assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	//token := utils.Get_intel_Token()
	computeUrl := utils.Get_Compute_Base_Url()
	logger.Log.Info("Compute Url" + computeUrl)
	baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	// 	sshKeyData := utils.Get_ssh_Key_Payload("visali.vasiraju@intel.com", "testinteluserkey")
	// 	logger.Log.Info("SSH Key data " + sshKeyData)
	// 	err := testsetup.CreateSSHKey(sshKeyData , computeUrl, baseUrl,  authToken)
	//     assert.Equal(suite.T(), nil, err, "Failed: SSH key creation failed for intel user")

	// 	vnetData := utils.Get_Vet_Payload("visali.vasiraju@intel.com", "autotestvnet8990")
	// 	logger.Log.Info("VNet data "+ vnetData)
	// 	err = testsetup.CreateVnet(vnetData , computeUrl, baseUrl, authToken)
	//     assert.Equal(suite.T(), nil, err, "Failed: Vnet creation failed for intel user")

	// 	instanceData := utils.Get_Instance_Payload("visali.vasiraju@intel.com", "testintelfree")
	// 	logger.Log.Info("Instance data "+ instanceData)
	// 	//instanceName := gjson.Get(instanceData, "name").String()
	// 	err = testsetup.CreateVmInstance(instanceData, computeUrl, baseUrl, authToken)
	//     assert.Equal(suite.T(), nil, err, "Failed: Vnet creation failed for intel user")

	//    time.Sleep(3*time.Minute)

	_, err1 := testsetup.GetUsageAndValidateTotalUsage("visali.vasiraju@intel.com", "vm-spr-tny", baseUrl, authToken, computeUrl)
	assert.Equal(suite.T(), nil, err1, "Failed: Validating Usage failed for intel user for all product")

	//    err:= testsetup.ValidateCreditsDeduction("visali.vasiraju@intel.com", "vm-spr-tny" , baseUrl, authToken, computeUrl)
	//    assert.Equal(suite.T(), nil, err, "Failed: Validating Credit depletion failed for intel user")
	//Cleanup

	//    err := testsetup.DeleteSSHKey("testinteluserkey", "visali.vasiraju@intel.com", computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: SSH Key  deletion failed")
	//    err = testsetup.DeleteVnet("autotestvnet8990", "visali.vasiraju@intel.com",computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: VNet  deletion failed")
	//    err=testsetup.DeleteInstance("testintelfree" , "visali.vasiraju@intel.com",computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: Instance deletion failed")
	//    ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	// 	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account(Intel User)")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestGetUsageIntelUserPaidProduct() {
	// Enroll intel user
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	///lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create a Coupon

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
	logger.Log.Info("Starting Billing Test : Redeem coupon for Intel cloud account")

	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	time.Sleep(1 * time.Minute)
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

	// //token := utils.Get_intel_Token()
	computeUrl := utils.Get_Compute_Base_Url()
	logger.Log.Info("Compute Url" + computeUrl)
	baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	sshKeyData := utils.Get_ssh_Key_Payload("visali.vasiraju@intel.com", "testinteluserkey")
	logger.Log.Info("SSH Key data " + sshKeyData)
	err := testsetup.CreateSSHKey(sshKeyData, computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: SSH key creation failed for intel user")

	vnetData := utils.Get_Vet_Payload("visali.vasiraju@intel.com", "us-dev-1a-default")
	logger.Log.Info("VNet data " + vnetData)
	err = testsetup.CreateVnet(vnetData, computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: Vnet creation failed for intel user")

	instanceData := utils.Get_Instance_Payload("visali.vasiraju@intel.com", "testintelsmall")
	logger.Log.Info("Instance data " + instanceData)
	//instanceName := gjson.Get(instanceData, "name").String()
	err = testsetup.CreateVmInstance(instanceData, computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: Instance creation failed for intel user")

	time.Sleep(3 * time.Minute)
	logger.Log.Info("Triggering s chedules ....")
	time.Sleep(3 * time.Minute)

	usage, err1 := testsetup.GetUsageAndValidateTotalUsage("visali.vasiraju@intel.com", "vm-spr-sml", baseUrl, authToken, computeUrl)
	assert.Equal(suite.T(), nil, err1, "Failed: Validating Usage failed for intel user for free product")
	fmt.Println("Intel user usage data", usage)

	// Cleanup

	err = testsetup.DeleteSSHKey("testinteluserkey", "visali.vasiraju@intel.com", computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: SSH Key  deletion failed")
	//    err = testsetup.DeleteVnet("us-dev-1a-default", "visali.vasiraju@intel.com",computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: VNet  deletion failed")
	err = testsetup.DeleteInstance("testintelsmall", "visali.vasiraju@intel.com", computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: Instance deletion failed")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account(Intel User)")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestGetUsageIntelUserPaidProductAfterDeletingInstance() {
	// Enroll intel user
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	///lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create a Coupon

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
	logger.Log.Info("Starting Billing Test : Redeem coupon for Intel cloud account")

	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	time.Sleep(1 * time.Minute)
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

	// //token := utils.Get_intel_Token()
	computeUrl := utils.Get_Compute_Base_Url()
	logger.Log.Info("Compute Url" + computeUrl)
	baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	sshKeyData := utils.Get_ssh_Key_Payload("visali.vasiraju@intel.com", "testinteluserkey")
	logger.Log.Info("SSH Key data " + sshKeyData)
	err := testsetup.CreateSSHKey(sshKeyData, computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: SSH key creation failed for intel user")

	vnetData := utils.Get_Vet_Payload("visali.vasiraju@intel.com", "us-dev-1a-default")
	logger.Log.Info("VNet data " + vnetData)
	err = testsetup.CreateVnet(vnetData, computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: Vnet creation failed for intel user")

	instanceData := utils.Get_Instance_Payload("visali.vasiraju@intel.com", "testintelsmall")
	logger.Log.Info("Instance data " + instanceData)
	//instanceName := gjson.Get(instanceData, "name").String()
	err = testsetup.CreateVmInstance(instanceData, computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: Instance creation failed for intel user")

	time.Sleep(3 * time.Minute)
	logger.Log.Info("Triggering schedules ....")
	time.Sleep(6 * time.Minute)

	usage, err1 := testsetup.GetUsageAndValidateTotalUsage("visali.vasiraju@intel.com", "vm-spr-sml", baseUrl, authToken, computeUrl)
	assert.Equal(suite.T(), nil, err1, "Failed: Validating Usage failed for intel user for free product")
	fmt.Println("Intel user usage data", usage)

	//Cleanup

	err = testsetup.DeleteSSHKey("testinteluserkey", "visali.vasiraju@intel.com", computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: SSH Key  deletion failed")
	//    err = testsetup.DeleteVnet("us-dev-1a-default", "visali.vasiraju@intel.com",computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: VNet  deletion failed")
	err = testsetup.DeleteInstance("testintelsmall", "visali.vasiraju@intel.com", computeUrl, baseUrl, authToken)
	assert.Equal(suite.T(), nil, err, "Failed: Instance deletion failed")
	logger.Log.Info("Waiting for sometime after deleting instance....")

	time.Sleep(5 * time.Minute)

	usage1, err1 := testsetup.GetUsageAndValidateTotalUsage("visali.vasiraju@intel.com", "vm-spr-sml", baseUrl, authToken, computeUrl)
	assert.Equal(suite.T(), nil, err1, "Failed: Validating Usage failed for intel user for free product")
	fmt.Println("Intel user usage data", usage)

	time.Sleep(6 * time.Minute)

	usage2, err1 := testsetup.GetUsageAndValidateTotalUsage("visali.vasiraju@intel.com", "vm-spr-sml", baseUrl, authToken, computeUrl)
	assert.Equal(suite.T(), nil, err1, "Failed: Validating Usage failed for intel user for free product")
	assert.Equal(suite.T(), usage1, usage2, "Failed: Validating Usage failed for intel user for free product, usages did not match")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account(Intel User)")
	logger.Logf.Infof("Intel user usage data Initial", usage)
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestGetUsageIntelUserBothFreePaidProduct() {
	// Enroll intel user
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	///lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create a Coupon

	// creation_time, expirationtime := billing.GetCreationExpirationTime()
	// createCoupon := billing.CreateCouponStruct{
	// 	Amount:  300,
	// 	Creator: "idc_billing@intel.com",
	// 	Expires: expirationtime,
	// 	Start:   creation_time,
	// 	NumUses: 2,
	// }
	// jsonPayload, _ := json.Marshal(createCoupon)
	// req := []byte(jsonPayload)
	// ret_value, data := billing.CreateCoupon(req, 200)
	// couponCode := gjson.Get(data, "code").String()
	// assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	// assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	// assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	// assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	// assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	// assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// // Get coupon and validate
	// getret_value, getdata := billing.GetCoupons(couponCode, 200)
	// couponData := gjson.Get(getdata, "coupons")
	// for _, val := range couponData.Array() {
	// 	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	// 	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	// 	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	// 	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	// 	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
	// 	assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	// }
	// assert.Equal(suite.T(), getret_value, true, "Failed: Get Coupon Failed to validate start")
	// //Redeem coupon
	// logger.Log.Info("Starting Billing Test : Redeem coupon for Intel cloud account")

	// redeemCoupon := billing.RedeemCouponStruct{
	// 	CloudAccountID: get_CAcc_id,
	// 	Code:           couponCode,
	// }
	// time.Sleep(1 * time.Minute)
	// jsonPayload, _ = json.Marshal(redeemCoupon)
	// req = []byte(jsonPayload)
	// ret_value, data = billing.RedeemCoupon(req, 200)

	// // Get coupon and validate
	// getret_value, getdata = billing.GetCoupons(couponCode, 200)
	// redemptions := gjson.Get(getdata, "result.redemptions")
	// for _, val := range redemptions.Array() {
	// 	assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
	// 	assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
	// 	assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	// }
	// couponData = gjson.Get(getdata, "coupons")
	// for _, val := range couponData.Array() {
	// 	assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	// }

	//token := utils.Get_intel_Token()
	computeUrl := utils.Get_Compute_Base_Url()
	logger.Log.Info("Compute Url" + computeUrl)
	baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	// 	sshKeyData := utils.Get_ssh_Key_Payload("visali.vasiraju@intel.com", "testinteluserkey")
	// 	logger.Log.Info("SSH Key data " + sshKeyData)
	// 	err := testsetup.CreateSSHKey(sshKeyData , computeUrl, baseUrl,  authToken)
	//     assert.Equal(suite.T(), nil, err, "Failed: SSH key creation failed for intel user")

	// 	vnetData := utils.Get_Vet_Payload("visali.vasiraju@intel.com", "autotestvnet8990")
	// 	logger.Log.Info("VNet data "+ vnetData)
	// 	err = testsetup.CreateVnet(vnetData , computeUrl, baseUrl, authToken)
	//     assert.Equal(suite.T(), nil, err, "Failed: Vnet creation failed for intel user")

	// 	instanceData := utils.Get_Instance_Payload("visali.vasiraju@intel.com", "testintelfree")
	// 	logger.Log.Info("Instance data "+ instanceData)
	// 	//instanceName := gjson.Get(instanceData, "name").String()
	// 	err = testsetup.CreateVmInstance(instanceData, computeUrl, baseUrl, authToken)
	//     assert.Equal(suite.T(), nil, err, "Failed: Instance creation failed for intel user")

	// 	instanceData = utils.Get_Instance_Payload("visali.vasiraju@intel.com", "testintelsmall")
	// 	logger.Log.Info("Instance data "+ instanceData)
	// 	//instanceName := gjson.Get(instanceData, "name").String()
	// 	err = testsetup.CreateVmInstance(instanceData, computeUrl, baseUrl, authToken)
	//     assert.Equal(suite.T(), nil, err, "Failed: Instance creation failed for intel user")

	//    time.Sleep(3*time.Minute)

	usage, err1 := testsetup.GetUsageAndValidateTotalUsage("visali.vasiraju@intel.com", "vm-spr-sml", baseUrl, authToken, computeUrl)
	assert.Equal(suite.T(), nil, err1, "Failed: Validating Usage failed for intel user for free product")
	fmt.Println("Intel user usage data", usage)

	//Cleanup

	//    err = testsetup.DeleteSSHKey("testinteluserkey", "visali.vasiraju@intel.com", computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: SSH Key  deletion failed")
	//    err = testsetup.DeleteVnet("autotestvnet8990", "visali.vasiraju@intel.com",computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: VNet  deletion failed")
	//    err=testsetup.DeleteInstance("testintelfree" , "visali.vasiraju@intel.com",computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: Instance deletion failed")
	//    ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	// 	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account(Intel User)")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) TestGetUsageDateRange() {
	// Enroll intel user
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	//paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	///lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	//assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	//token := utils.Get_intel_Token()
	computeUrl := utils.Get_Compute_Base_Url()
	logger.Log.Info("Compute Url" + computeUrl)
	baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	// 	sshKeyData := utils.Get_ssh_Key_Payload("visali.vasiraju@intel.com", "testinteluserkey")
	// 	logger.Log.Info("SSH Key data " + sshKeyData)
	// 	err := testsetup.CreateSSHKey(sshKeyData , computeUrl, baseUrl,  authToken)
	//     assert.Equal(suite.T(), nil, err, "Failed: SSH key creation failed for intel user")

	// 	vnetData := utils.Get_Vet_Payload("visali.vasiraju@intel.com", "autotestvnet8990")
	// 	logger.Log.Info("VNet data "+ vnetData)
	// 	err = testsetup.CreateVnet(vnetData , computeUrl, baseUrl, authToken)
	//     assert.Equal(suite.T(), nil, err, "Failed: Vnet creation failed for intel user")

	// 	instanceData := utils.Get_Instance_Payload("visali.vasiraju@intel.com", "testintelfree")
	// 	logger.Log.Info("Instance data "+ instanceData)
	// 	//instanceName := gjson.Get(instanceData, "name").String()
	// 	err = testsetup.CreateVmInstance(instanceData, computeUrl, baseUrl, authToken)
	//     assert.Equal(suite.T(), nil, err, "Failed: Vnet creation failed for intel user")

	//    time.Sleep(3*time.Minute)

	//    _, err1 := testsetup.GetUsageAndValidateTotalUsage("visali.vasiraju@intel.com", "vm-spr-tny" , baseUrl, authToken, computeUrl)
	//    assert.Equal(suite.T(), nil, err1, "Failed: Validating Usage failed for intel user for free product")

	//Cleanup

	//    err := testsetup.DeleteSSHKey("testinteluserkey", "visali.vasiraju@intel.com", computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: SSH Key  deletion failed")
	//    err = testsetup.DeleteVnet("autotestvnet8990", "visali.vasiraju@intel.com",computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: VNet  deletion failed")
	//    err=testsetup.DeleteInstance("testintelfree" , "visali.vasiraju@intel.com",computeUrl, baseUrl ,authToken)
	//    assert.Equal(suite.T(), nil, err, "Failed: Instance deletion failed")
	//    ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	// 	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account(Intel User)")
	os.Unsetenv("intel_user_test")
}
