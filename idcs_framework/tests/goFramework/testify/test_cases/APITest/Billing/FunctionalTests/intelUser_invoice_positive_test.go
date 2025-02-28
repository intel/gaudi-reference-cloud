//go:build Functional || Invoice || In || Billing10
// +build Functional Invoice In Billing10

package BillingAPITest

import (
	//"encoding/json"
	//"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/utils"

	//"goFramework/ginkGo/compute/compute_utils"
	//"goFramework/ginkGo/financials/financials_utils"
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

// func (suite *BillingAPITestSuite) TestEnrollIntelCustomerCheckInvoice() {
// 	// Enroll intel user
//     logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
// 	os.Setenv("intel_user_test", "True")
// 	//Create Cloud Account
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	enterpriseId := "testeid-01"
// 	userName := utils.Get_UserName("Intel")
// 	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
// 	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
// 	//paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
// 	///lowCred := gjson.Get(responsePayload, "lowCredits").String()
// 	CountryCode := gjson.Get(responsePayload, "countryCode").String()
//     //assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
// 	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
//     //assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

// 	// Get usage

// 	ret, invoiceData := billing.GetInvoice(get_CAcc_id, 200)
//     assert.Equal(suite.T(), ret, true, "Failed to fetch usage for intel account")
// 	result := gjson.Get(invoiceData, "invoices")
// 	result.ForEach(func(key, value gjson.Result) bool {
// 	    totalAmt := gjson.Get(value.String(), "total").String()
// 	    totalUsage := gjson.Get(value.String(), "paid").String()
// 	    assert.Equal(suite.T(), "0", totalAmt, "Total amount is not equal to 0 after enrollment")
// 	    assert.Equal(suite.T(), "0", totalUsage, "Total Usage is not equal to 0 after enrollment")
// 	    return true // keep iterating
//   })
// }

func (suite *BillingAPITestSuite) TestEnrollIntelCustomerCheckInvoiceHistory() {
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

	// Get usage
	//startTime, endTime := billing.GetCreationExpirationTime()
	ret, invoiceData := billing.GetInvoicesHistory(get_CAcc_id, "2023-06-19T19:41:41.851Z", "2023-06-19T19:41:41.851Z", 200)
	assert.Equal(suite.T(), ret, true, "Failed to fetch usage for intel account")
	result := gjson.Get(invoiceData, "invoices")

	result.ForEach(func(key, value gjson.Result) bool {
		totalAmt := gjson.Get(value.String(), "total").String()
		totalUsage := gjson.Get(value.String(), "paid").String()
		assert.Equal(suite.T(), "0", totalAmt, "Total amount is not equal to 0 after enrollment")
		assert.Equal(suite.T(), "0", totalUsage, "Total Usage is not equal to 0 after enrollment")
		return true // keep iterating
	})

	os.Unsetenv("intel_user_test")
}

// func (suite *BillingAPITestSuite) TestEnrollIntelCustomerCheckInvoiceById() {
// 	// Enroll intel user
//     logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
// 	os.Setenv("intel_user_test", "True")
// 	//Create Cloud Account
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	enterpriseId := "testeid-01"
// 	userName := utils.Get_UserName("Intel")
// 	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
// 	assert.Equal(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
// 	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
// 	//paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
// 	///lowCred := gjson.Get(responsePayload, "lowCredits").String()
// 	CountryCode := gjson.Get(responsePayload, "countryCode").String()
//     //assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
// 	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
//     //assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
// 	assert.Equal(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

// 	// Get usage

// 	ret, invoiceData := billing.GetInvoice(get_CAcc_id, 200)
//     assert.NotEqual(suite.T(), ret, true, "Failed to fetch usage for intel account")

// 	result := gjson.Get(invoiceData, "invoices")
// 	var id string
//     result.ForEach(func(key, value gjson.Result) bool {
// 	    id = gjson.Get(value.String(), "id").String()
// 	    return true // keep iterating
//   })
//   _, data := billing.GetInvoicesById(get_CAcc_id , id , 200)
//   result = gjson.Get(data, "invoices")
// 	var id2 string
//     result.ForEach(func(key, value gjson.Result) bool {
// 	    id2 = gjson.Get(value.String(), "id").String()
// 	    return true // keep iterating
//   })
//   assert.Equal(suite.T(), id, id2, "Invoice Id didnt match")
//   os.Unsetenv("intel_user_test")
// }
