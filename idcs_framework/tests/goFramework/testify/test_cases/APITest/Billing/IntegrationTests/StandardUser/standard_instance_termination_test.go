//go:build Functional || InstanceTerminationStandard || StandardIntegration
// +build Functional InstanceTerminationStandard StandardIntegration

package StandardBillingAPITest

import (
	"fmt"
	_ "fmt"
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
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) Test_Standard_Enroll_New_User_Validate_cloudAcc_Attributes() {
	//pass
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a Standard User Cloud Account using Enroll API")
	base_url := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)

	logger.Logf.Infof(" Billing base Url :  %s", base_url)
	mailTime := time.Now().Add(1 * time.Minute)

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
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Validate Notifications ToDo From Dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, get_CAcc_id)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// assert.Equal(suite.T(), int64(0), gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 on fresh user enrollment ")

	// Validate Mail Notification

	time.Sleep(20 * time.Second)

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User received 80% notification even when credits available")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, true, "Cloud Account Listed in Deactivation List")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Enroll_New_User_Validate_cloudAcc_Attributes_After_adding_Credits() {
	//pass
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a Standard User Cloud Account using Enroll API")
	base_url := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)

	logger.Logf.Infof(" Billing base Url :  %s", base_url)
	mailTime := time.Now().Add(1 * time.Minute)

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
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Add coupon to the user
	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(300), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon")

	time.Sleep(time.Duration(5 * time.Minute))

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// Validate Mail Notifications

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User received 80% notification even when credits available")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// Validate Notifications TODO: Pending dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, get_CAcc_id)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// assert.Equal(suite.T(), int64(0), gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 ")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Expire_Coupon_Validate_Instance_Deletion() {
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

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)

	logger.Logf.Infof(" Billing base Url :  %s", base_url)
	mailTime := time.Now().Add(1 * time.Minute)

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
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create and redeem normal coupon

	coupon_expire_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Standard", int64(5), int64(2), get_CAcc_id, time.Duration(2))
	assert.Equal(suite.T(), coupon_expire_err, nil, "Failed to create coupon with shorter expiry, failed with error : %s", coupon_expire_err)

	// Now launch paid instance and see API throws 403 error

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(get_CAcc_id, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Push Metering Data to use all Credits

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Validate Cloud Account attributes

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
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

	// Validate email notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User received 80% notification even when credits available")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

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

func (suite *BillingAPITestSuite) Test_Standard_Expire_Coupon_Validate_Instance_Runs_When_Credits_Available() {
	// pass
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

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)

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
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// // Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(35), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// coupon_expire_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry( "Standard", int64(15), int64(2), get_CAcc_id, time.Duration(10))
	// assert.Equal(suite.T(), coupon_expire_err, nil, "Failed to create coupon with shorter expiry, failed with error : %s", coupon_expire_err)

	// // Now launch paid instance

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(get_CAcc_id, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	mailTime := time.Now().Add(1 * time.Minute)

	// Push Metering data
	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	//previousDate = previousDate.Add(2 * time.Minute)
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), get_CAcc_id, previousDate, "vm-spr-sml", "smallvm", "180000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + get_CAcc_id
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: %s ", usage_response_body)
	tamt := 22.5
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s", data)
		if gjson.Get(data, "productType").String() == "vm-spr-sml" {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			assert.GreaterOrEqual(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

			logger.Logf.Infof("Actual Usage :  ", actualAMount)
			assert.GreaterOrEqual(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")
		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.GreaterOrEqual(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	zeroamt := 0
	baseUrl = utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	amt := 22.5
	assert.GreaterOrEqual(suite.T(), usedAmount, float64(amt), "Failed to validate used credits")
	assert.GreaterOrEqual(suite.T(), remainingAmount, float64(10), "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	logger.Logf.Info("Unapplied credits After credits1 coupon: %s ", unappliedCredits)
	assert.Greater(suite.T(), unappliedCredits, float64(9), "Failed : Unapplied cloud credit did not become zero")

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, false, "Cloud Account Not Listed in Deactivation List After Expired credits")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), delete_status, 200, "Failed : Instance got deleted after credits expired")
		time.Sleep(10 * time.Second)
	}

	// Validate Mail Notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User received 80% notification even when credits available")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// Validate cloud credit data
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	logger.Logf.Info("Credit Response: %s ", responseBody)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Greater(suite.T(), totalRemainingAmount, float64(zeroamt), "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Expire_Coupon_Validate_Instance_Runs_When_Credits_Available_Redeem_Expire_First() {
	// pass
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

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)

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
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create and redeem  coupon
	coupon_expire_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Standard", int64(15), int64(2), get_CAcc_id, time.Duration(7))
	assert.Equal(suite.T(), coupon_expire_err, nil, "Failed to create coupon with shorter expiry, failed with error : %s", coupon_expire_err)

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// Now launch paid instance

	mailTime := time.Now().Add(10 * time.Second)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(get_CAcc_id, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Wait for the coupon to expire

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	//previousDate = previousDate.Add(2 * time.Minute)
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), get_CAcc_id, previousDate, "vm-spr-sml", "smallvm", "180000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + get_CAcc_id
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: %s ", usage_response_body)
	tamt := 22.5
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s", data)
		if gjson.Get(data, "productType").String() == "vm-spr-sml" {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			assert.GreaterOrEqual(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

			logger.Logf.Infof("Actual Usage :  ", actualAMount)
			assert.GreaterOrEqual(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")
		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.GreaterOrEqual(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	zeroamt := 0
	baseUrl = utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	amt := 22.5
	assert.GreaterOrEqual(suite.T(), usedAmount, float64(amt), "Failed to validate used credits")
	assert.Greater(suite.T(), remainingAmount, float64(zeroamt), "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	logger.Logf.Info("Unapplied credits After credits1 coupon: %s ", unappliedCredits)
	assert.Greater(suite.T(), unappliedCredits, float64(zeroamt), "Failed : Unapplied cloud credit did not become zero")

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, false, "Cloud Account Not Listed in Deactivation List After Expired credits")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), delete_status, 200, "Failed : Instance got deleted after credits expired")
		time.Sleep(10 * time.Second)
	}

	// Validate Mail Notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User received 80% notification even when credits available")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// Validate cloud credit data
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	logger.Logf.Info("Credit Response: %s ", responseBody)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Greater(suite.T(), totalRemainingAmount, float64(zeroamt), "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Instance_Deletion_Upon_Usageof_AllCredits() {
	// pass
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a Standard User Cloud Account using Enroll API")
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	computeUrl := utils.Get_Compute_Base_Url()
	base_url := utils.Get_Base_Url1()
	baseUrl := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)

	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(20), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// Now launch paid instance
	mailTime := time.Now().Add(1 * time.Minute)
	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(get_CAcc_id, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), get_CAcc_id, previousDate, "vm-spr-sml", "smallvm", "180000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + get_CAcc_id
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)
	tamt := 20
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s", data)
		if gjson.Get(data, "productType").String() == "vm-spr-sml" {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			assert.GreaterOrEqual(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

			logger.Logf.Infof("Actual Usage ", actualAMount)
			assert.GreaterOrEqual(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")
		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.GreaterOrEqual(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl = utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	amt := 20
	assert.GreaterOrEqual(suite.T(), usedAmount, float64(amt), "Failed to validate used credits")
	assert.Equal(suite.T(), remainingAmount, "0", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	unappliedCredits = testsetup.RoundFloat(unappliedCredits, 1)
	assert.Equal(suite.T(), unappliedCredits, float64(0), "Failed : Unapplied cloud credit did not become zero")

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, true, "Cloud Account Not Listed in Deactivation List After Expired credits")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 404, get_response_status, "Failed : Instance not deleted after all credits used")
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
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

	// Validate Mail Notification

	// Validate Mail Notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	// emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	// assert.NotEqual(suite.T(), emailNotification80, "", "Standard User received 80% notification even when credits available")
	// assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification100, "", "Standard User did not receive 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) TestStandardUsageMoreThanCredits() {
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a Standard User Cloud Account using Enroll API")
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	computeUrl := utils.Get_Compute_Base_Url()
	baseUrl := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)

	// Create and redeem normal coupon

	// Create and redeem  coupon

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(20), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// Now launch paid instance

	// Push Metering Data to use all Credits

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(get_CAcc_id, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	mailTime := time.Now().Add(10 * time.Second)
	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), get_CAcc_id, previousDate, "vm-spr-sml", "smallvm", "240000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsageinRange(get_CAcc_id, float64(0.0075), "vm-spr-sml", authToken, float64(30), float64(31))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(get_CAcc_id, float64(20), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : %s", credits_err)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 404, get_response_status, "Failed : Instance not deleted after all credits used")
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	// Validate Mail Notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	// emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	// assert.NotEqual(suite.T(), emailNotification80, "", "Standard User received 80% notification even when credits available")
	// assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification100, "", "Standard User did not receive 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Gets_100_percent_Notification() {
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a Standard User Cloud Account using Enroll API")
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	computeUrl := utils.Get_Compute_Base_Url()
	base_url := utils.Get_Base_Url1()
	baseUrl := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"

	url := base_url + "/v1/cloudaccounts"

	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)
	mailTime := time.Now().Add(1 * time.Minute)

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create and redeem  coupon
	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(1), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// Now launch paid instance

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(2 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), get_CAcc_id, previousDate, "vm-spr-sml", "smallvm", "8820")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(get_CAcc_id, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + get_CAcc_id
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)
	tamt := 0.5
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s", data)
		if gjson.Get(data, "productType").String() == "vm-spr-sml" {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			actualAMount = testsetup.RoundFloat(actualAMount, 2)
			assert.Greater(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

			logger.Logf.Infof("Actual Usage ", actualAMount)
			assert.GreaterOrEqual(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")
		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.GreaterOrEqual(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl = utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 2)
	assert.GreaterOrEqual(suite.T(), usedAmount, float64(1), "Failed to validate used credits")
	assert.Equal(suite.T(), remainingAmount, "0", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	unappliedCredits = testsetup.RoundFloat(unappliedCredits, 1)
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), unappliedCredits, float64(0), "Failed : Validation failed in Unapplied Cloud Credit")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 404, get_response_status, "Failed : Instance not deleted after all credits used")
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
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

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Validate Notifications, TODO from dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, get_CAcc_id)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// numberofNotifications := gjson.Get(responseBody, "numberOfNotifications").Int()
	// assert.Equal(suite.T(), 0, gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 on fresh user enrollment ")

	// Check Cloud account not in deactivation list

	// Validate Mail Notification

	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User did not receive 80% notification even when credits available")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, true, "Cloud Account Listed in Deactivation List")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Gets_80_percent_Notification() {
	// pass
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a Standard User Cloud Account using Enroll API")
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	computeUrl := utils.Get_Compute_Base_Url()
	//base_url := utils.Get_Base_Url1()
	baseUrl := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)
	mailTime := time.Now().Add(1 * time.Minute)

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create and redeem normal coupon

	// Create and redeem  coupon
	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(1), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// Now launch paid instance

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(get_CAcc_id, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), get_CAcc_id, previousDate, "bm-spr", "bm", "840")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsageNotZero(get_CAcc_id, float64(0.0603), "bm-spr", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : %s", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCreditsNonZeroDepletion(get_CAcc_id, 1, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	//Validate credits
	time.Sleep(4 * time.Minute)
	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance deleted at 80 percent")
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Validate Notifications, TODO from dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, get_CAcc_id)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// numberofNotifications := gjson.Get(responseBody, "numberOfNotifications").Int()
	// assert.Equal(suite.T(), 0, gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 on fresh user enrollment ")

	// Check Cloud account not in deactivation list

	// Validate Mail Notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification80, "", "Standard User did not receive 80% notification even when credits available")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not deleted after all credits used")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Adds_Credits_After_80_percent_Notification() {
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a Standard User Cloud Account using Enroll API")
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	computeUrl := utils.Get_Compute_Base_Url()
	base_url := utils.Get_Base_Url1()
	baseUrl := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"

	url := base_url + "/v1/cloudaccounts"

	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)
	mailTime := time.Now().Add(1 * time.Minute)

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create and redeem normal coupon

	// Create and redeem  coupon
	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(1), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), get_CAcc_id, previousDate, "vm-spr-sml", "smallvm", "6420")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Now launch paid instance

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(get_CAcc_id, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	usage_err := billing.ValidateUsage(get_CAcc_id, float64(0.8025), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : %s", usage_err)

	//Validate credits

	time.Sleep(3 * time.Minute)
	credits_err := billing.ValidateUsageCreditsinRange(get_CAcc_id, float64(0.8), float64(1), authToken, float64(0.07), float64(0.2), float64(0.07), float64(0.2), float64(0))
	//credits_err := billing.ValidateCredits(get_CAcc_id, float64(1), authToken, float64(0.20), float64(-2.6), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : %s", credits_err)

	//Validate credits
	time.Sleep(4 * time.Minute)
	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance deleted at 80 percent")
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Validate Notifications, TODO from dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, get_CAcc_id)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// numberofNotifications := gjson.Get(responseBody, "numberOfNotifications").Int()
	// assert.Equal(suite.T(), 0, gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 on fresh user enrollment ")

	// Check Cloud account not in deactivation list

	// Validate Mail Notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification80, "", "Standard User did not receive 80% notification even when credits available")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// User adds credits

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(5), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	mailTime = time.Now().Add(5 * time.Minute)
	time.Sleep(10 * time.Minute)

	// status, instanceListed = financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	emailNotification100, err1 = financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not deleted after all credits used")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Adds_Credits_After_100_percent_Notification() {
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a Standard User Cloud Account using Enroll API")
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	computeUrl := utils.Get_Compute_Base_Url()
	base_url := utils.Get_Base_Url1()
	baseUrl := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"

	url := base_url + "/v1/cloudaccounts"

	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)
	mailTime := time.Now().Add(2 * time.Minute)

	url = base_url + "/v1/cloudaccounts"
	cloudAccId, err = testsetup.GetCloudAccountId(userName, base_url, authToken)
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
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Create and redeem normal coupon
	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(1), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(2 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), get_CAcc_id, previousDate, "bm-spr", "bm", "1200")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(get_CAcc_id, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	time.Sleep(3 * time.Minute)

	// Now launch paid instance

	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + get_CAcc_id
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)
	tamt := 1
	maxtamt := 2
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s", data)
		if gjson.Get(data, "productType").String() == "bm-spr" {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			actualAMount = testsetup.RoundFloat(actualAMount, 2)
			assert.Greater(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")
			assert.Less(suite.T(), actualAMount, float64(maxtamt), "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), 0.0603, rateFactor, "Failed: Failed to validate rate amount")

		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.Greater(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")
	assert.Less(suite.T(), total_amount_from_response, float64(maxtamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl = utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 2)
	amt := 1
	assert.Equal(suite.T(), float64(amt), usedAmount, "Failed to validate used credits")
	assert.Equal(suite.T(), remainingAmount, "0", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(get_CAcc_id, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	unappliedCredits = testsetup.RoundFloat(unappliedCredits, 1)
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), unappliedCredits, float64(0), "Failed : Validation failed in Unapplied Cloud Credit")

	// Validate cloud credit data
	zeroamt := 0
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(zeroamt), totalRemainingAmount, "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Validate Notifications, TODO from dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, get_CAcc_id)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// numberofNotifications := gjson.Get(responseBody, "numberOfNotifications").Int()
	// assert.Equal(suite.T(), 0, gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 on fresh user enrollment ")

	// Check Cloud account not in deactivation list

	// status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, true, "Cloud Account Listed in Deactivation List")

	// Add credits again and check no new notifications are coming (TODO: Notification check)

	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User did not receive 80% notification even when credits available")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// Create and redeem normal coupon

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(1), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	mailTime = time.Now().Add(1 * time.Minute)
	time.Sleep(6 * time.Minute)
	// Validate cloud credit data
	zeroamt = 1
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, get_CAcc_id)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount = gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.GreaterOrEqual(suite.T(), float64(zeroamt), totalRemainingAmount, "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Validate Notifications, TODO from dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, get_CAcc_id)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// numberofNotifications := gjson.Get(responseBody, "numberOfNotifications").Int()
	// assert.Equal(suite.T(), 0, gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 on fresh user enrollment ")

	// Check Cloud account not in deactivation list

	// status, instanceListed = financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	// assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	// assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// Validate Mail Notification

	proxy_val = os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err = financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User did not receive 80% notification even when credits available")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 = financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 = financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not deleted after all credits used")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Standard user)")

}
