//go:build Functional || DailyRun
// +build Functional DailyRun

package StandardBillingAPITest

import (
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
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// var met_ret bool

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Create_Paid_Instance_Without_Credits() {
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
		time.Sleep(3 * time.Minute)
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Create_Paid_Instance_With_Coupon_Credits() {
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
		time.Sleep(3 * time.Minute)
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Launch_Paid_Instance_After_Using_More_Credits() {
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
		time.Sleep(3 * time.Minute)
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Redeem_Coupon_After_Using_All_Credits() {
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
		time.Sleep(3 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 404, delete_status, "Failed : Instance not deleted after all credits expired")
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

	time.Sleep(3 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(15), float64(15), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Redeem_Lesser_Coupon_After_Using_All_Credits() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(12), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "14925")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Validate credits

	credits_err := billing.ValidateCredits(cloudAccId, float64(12), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Redeem coupon

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Validate credits

	credits_err = billing.ValidateCredits(cloudAccId, float64(13), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Noew launch a paid instance, instance should not be launched

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(3 * time.Minute)
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Usage_More_Than_Credits_Redeem_Lesser_Value_Coupon() {
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
		time.Sleep(3 * time.Minute)
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Expire_Coupon_Launch_Paid_Instance() {
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
		time.Sleep(3 * time.Minute)
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Launch_Small_Instance_Validate_Usage() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	if vm_creation_error != nil {
		time.Sleep(3 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}
	time.Sleep(3 * time.Minute)
	usage_err := billing.ValidateUsageNotZero(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usages, validation failed with error : ", usage_err)
	time.Sleep(3 * time.Minute)
	credits_err := billing.ValidateCreditsNonZeroDepletion(cloudAccId, 20, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard User)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Enroll_New_User_Validate_cloudAcc_Attributes() {
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
	assert.Equal(suite.T(), emailNotification80, "", "Standard User received 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, true, "Cloud Account Listed in Deactivation List")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Enroll_New_User_Validate_cloudAcc_Attributes_After_adding_Credits() {
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

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// Validate Mail Notifications

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User received 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification ")
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Expire_Coupon_Validate_Instance_Deletion() {
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

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, true, "Cloud Account Not Listed in Deactivation List After Expired credits")

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

	// emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	// // assert.NotEqual(suite.T(), emailNotification100, "", "Standard User received 100% notification even when credits available")
	// assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	// emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	// // assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification even when credits available")
	// assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// var flag bool= true
	// if emailNotification100 == "" && emailNotificationexpire == ""{
	//     flag=false
	// }
	// assert.Equal(suite.T(), flag, true, "Standard User did not receive either 100 percent or credit expiry notification on credit expiry")
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Instance_Deletion_Upon_Usageof_AllCredits() {
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

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, true, "Cloud Account Not Listed in Deactivation List After Expired credits")

	if vm_creation_error != nil {
		time.Sleep(3 * time.Minute)
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
	// assert.NotEqual(suite.T(), emailNotification80, "", "Standard User received 80% notification ")
	// assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification100, "", "Standard User did not receive 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Gets_80_percent_Notification() {
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

	if vm_creation_error != nil {
		time.Sleep(3 * time.Minute)
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
	assert.NotEqual(suite.T(), emailNotification80, "", "Standard User did not receive 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	if vm_creation_error != nil {
		time.Sleep(3 * time.Minute)
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Adds_Credits_After_80_percent_Notification() {
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

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateUsageCreditsinRange(get_CAcc_id, float64(0.8), float64(1), authToken, float64(0.07), float64(0.2), float64(0.07), float64(0.2), float64(0))
	//credits_err := billing.ValidateCredits(get_CAcc_id, float64(1), authToken, float64(0.20), float64(-2.6), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : %s", credits_err)

	//Validate credits
	time.Sleep(2 * time.Minute)
	if vm_creation_error != nil {
		time.Sleep(3 * time.Minute)
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
	assert.NotEqual(suite.T(), emailNotification80, "", "Standard User did not receive 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// User adds credits

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(5), int64(2), get_CAcc_id)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	mailTime = time.Now().Add(5 * time.Minute)
	time.Sleep(10 * time.Minute)

	status, instanceListed = financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	emailNotification100, err1 = financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	if vm_creation_error != nil {
		time.Sleep(3 * time.Minute)
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Adds_Credits_After_100_percent_Notification() {
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
	//now1 := time.Now().UTC()

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
	time.Sleep(3 * time.Minute)
	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	initialTS := gjson.Get(responsePayload, "creditsDepleted").String()
	assert.Equal(suite.T(), initialTS, "1970-01-01T00:00:00Z", "Failed: Credit depletion time stamp not correct")

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

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, true, "Cloud Account Listed in Deactivation List")

	// Add credits again and check no new notifications are coming (TODO: Notification check)

	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User did not receive 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification100, "", "Standard User did not receive 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification ")
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

	status, instanceListed = financials.CheckCloudAccInDeactivationList(base_url, authToken, get_CAcc_id)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// Validate Mail Notification

	proxy_val = os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err = financials.GetMailFromInbox(inboxIdStandard, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Standard User did not receive 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 = financials.GetMailFromInbox(inboxIdStandard, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Standard User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 = financials.GetMailFromInbox(inboxIdStandard, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Standard User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	if vm_creation_error != nil {
		time.Sleep(3 * time.Minute)
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

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Upgrade_to_Premium_Using_Coupon_When_Account_Has_Credits_And_Usage() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

	// Redeem coupon for standard user before upgrade

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push some usage and let credit depletion happen

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "48000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsage(cloudAccId, float64(6), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(10), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")

	// Validate credit details

	usage_err = billing.ValidateZeroUsage(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(14), float64(14), float64(14))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Upgrade_to_Premium_Using_Credit_Card_Rest_When_Account_Has_Credits() {
	suite.T().Skip()
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	base_url := utils.Get_Base_Url1()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Redeem coupon for standard user before upgrade

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	logger.Logf.Infof("Redeem Error", coupon_err)
	logger.Logf.Infof("Cloud Acc ID before upgrade", cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Check cloud account attributes before upgrade

	creditCardDetails := financials.CreditCardDetails{
		CCNumber:      4111111111111111,
		CCExpireMonth: 11,
		CCExpireYear:  2045,
		CCV:           988,
	}

	upgrade_err := financials.UpgradeAccountWithCreditCard(base_url, cloudAccId, userToken, creditCardDetails, "Visa", "1", "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with credit card failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.Equal(suite.T(), "UPGRADE_COMPLETE", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Validate credit details

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(10), float64(10), float64(10))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	logger.Log.Info("Delete cloud acc response" + ret_value1)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Upgrade_to_Premium_Using_Credit_Card_Rest_When_Account_Has_Credits_And_Usage() {
	suite.T().Skip()
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Redeem coupon for standard user before upgrade

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	logger.Logf.Infof("coupon_err", coupon_err)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push some usage and let credit depletion happen

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "48000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsage(cloudAccId, float64(6), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	// cloud account creation
	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Check cloud account attributes before upgrade

	creditCardDetails := financials.CreditCardDetails{
		CCNumber:      4111111111111111,
		CCExpireMonth: 11,
		CCExpireYear:  2045,
		CCV:           988,
	}

	upgrade_err := financials.UpgradeAccountWithCreditCard(base_url, cloudAccId, userToken, creditCardDetails, "Visa", "1", "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with credit card failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.Equal(suite.T(), "UPGRADE_COMPLETE", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Validate credit details

	usage_err = billing.ValidateUsage(cloudAccId, float64(6), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(4), float64(4), float64(4))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

}

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Upgrade_to_Premium_Using_Credit_Card_Rest_When_Account_Has_No_Credits() {
	suite.T().Skip()
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Check cloud account attributes before upgrade

	creditCardDetails := financials.CreditCardDetails{
		CCNumber:      4111111111111111,
		CCExpireMonth: 11,
		CCExpireYear:  2045,
		CCV:           988,
	}

	upgrade_err := financials.UpgradeAccountWithCreditCard(base_url, cloudAccId, userToken, creditCardDetails, "Visa", "1", "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with credit card failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.Equal(suite.T(), "UPGRADE_COMPLETE", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Validate credit details

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// try to launch instance

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	if vm_creation_error != nil {
		time.Sleep(3 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	logger.Log.Info("Delete cloud acc response" + ret_value1)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Upgrade_to_Premium_Using_Coupon_When_Account_Has_No_Credits() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(10), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.Equal(suite.T(), "UPGRADE_COMPLETE", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Validate credit details

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(10), float64(10), float64(10))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Sanity_Validate_All_CloudAccount_Attributes() {
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
	now1 := time.Now().UTC()

	url := base_url + "/v1/cloudaccounts"

	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}

	apiKey := utils.GetMailSlurpKey()
	inboxIdStandard := utils.GetInboxIdStandard()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdStandard %s", inboxIdStandard)

	url = base_url + "/v1/cloudaccounts"
	cloudAccId, err = testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	cloudAccId = get_CAcc_id
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
	time.Sleep(3 * time.Minute)
	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	initialTS := gjson.Get(responsePayload, "creditsDepleted").String()
	assert.Equal(suite.T(), initialTS, "1970-01-01T00:00:00Z", "Failed: Credit depletion time stamp not correct")

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(2 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bm", "3000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)
	tamt := 1
	maxtamt := 4
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
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 2)
	amt := 1
	assert.Equal(suite.T(), float64(amt), usedAmount, "Failed to validate used credits")
	assert.Equal(suite.T(), remainingAmount, "0", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	unappliedCredits = testsetup.RoundFloat(unappliedCredits, 1)
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), unappliedCredits, float64(0), "Failed : Validation failed in Unapplied Cloud Credit")

	// Validate cloud credit data
	zeroamt := 0
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(zeroamt), totalRemainingAmount, "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	ret_value1, responsePayload = cloudAccounts.GetCAccById(get_CAcc_id, 200)
	ts := gjson.Get(responsePayload, "creditsDepleted").String()
	creditDepletedTimestamp, _ := time.Parse(time.RFC3339, ts)
	assert.Equal(suite.T(), true, creditDepletedTimestamp.After(now1), "Failed: Credit depletion time stamp not correct")

	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Validate Notifications, TODO from dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, cloudAccId)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// numberofNotifications := gjson.Get(responseBody, "numberOfNotifications").Int()
	// assert.Equal(suite.T(), 0, gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 on fresh user enrollment ")

	// Check Cloud account not in deactivation list

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// Add credits again and check no new notifications are coming (TODO: Notification check)

	// Create and redeem normal coupon

	coupon_err = billing.Create_Redeem_Coupon("Intel", int64(6), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	time.Sleep(6 * time.Minute)

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	initialTS = gjson.Get(responsePayload, "creditsDepleted").String()
	assert.Equal(suite.T(), initialTS, "1970-01-01T00:00:00Z", "Failed: Credit depletion time stamp not correct")

	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	now = time.Now().UTC()
	previousDate = now.Add(2 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload = financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "vm1", "300000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url = base_url + "/v1/meteringrecords"
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(6 * time.Minute)
	// Validate cloud credit data
	zeroamt = 0
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount = gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(zeroamt), totalRemainingAmount, "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	time.Sleep(6 * time.Minute)
	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ts = gjson.Get(responsePayload, "creditsDepleted").String()
	creditDepletedTimestamp1, _ := time.Parse(time.RFC3339, ts)
	assert.Equal(suite.T(), true, creditDepletedTimestamp1.After(creditDepletedTimestamp), "Failed: Credit depletion time stamp not updated")

	// Validate Notifications, TODO from dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, cloudAccId)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// numberofNotifications := gjson.Get(responseBody, "numberOfNotifications").Int()
	// assert.Equal(suite.T(), 0, gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 on fresh user enrollment ")

	// Check Cloud account not in deactivation list

	status, instanceListed = financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not deleted after all credits used")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	// Wait for scheduler run to check timestamp again

	time.Sleep(6 * time.Minute)
	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodeStandard(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ts = gjson.Get(responsePayload, "creditsDepleted").String()
	creditDepletedTimestamp2, _ := time.Parse(time.RFC3339, ts)
	assert.Equal(suite.T(), false, creditDepletedTimestamp2.After(creditDepletedTimestamp1), "Failed: Credit depletion time stamp did not remain same")
	assert.Equal(suite.T(), creditDepletedTimestamp1, creditDepletedTimestamp2, "Failed: Credit depletion time stamp did not remain same")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Standard user)")

}
