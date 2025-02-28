//go:build Functional || DailyRun
// +build Functional DailyRun

package BillingAPITest

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
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Create_Paid_Instance_Without_Credits() {
	userName := utils.Get_UserName("Intel")
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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Expire_Coupon_Launch_Paid_Instance() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Intel", int64(15), int64(2), cloudAccId, time.Duration(5))
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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel User)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Create_Paid_Instance_With_Coupon_Credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(1), int64(2), cloudAccId)
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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Redeem_Coupon_After_Using_Less_Credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-lrg", "largevm", "75000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-med", userToken, 200)
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

	coupon_err = billing.Create_Redeem_Coupon("Intel", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	time.Sleep(8 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(15), float64(15), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Redeem_Coupon_After_Using_More_Credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Add coupon to the user

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-lrg", "largevm", "75000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-med", userToken, 403)
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

	coupon_err = billing.Create_Redeem_Coupon("Intel", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	time.Sleep(8 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(5), float64(5), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Redeem_Lesser_Coupon_After_Using_All_Credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-lrg", "largevm", "90000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Validate credits

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Redeem coupon

	coupon_err = billing.Create_Redeem_Coupon("Intel", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Validate credits

	credits_err = billing.ValidateCredits(cloudAccId, float64(16), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Now launch a paid instance, instance should not be launched

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-tny", userToken, 403)
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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Expire_Coupon_Validate_Credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Intel", int64(15), int64(2), cloudAccId, time.Duration(3))
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Wait for the coupon to expire
	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsage(cloudAccId, float64(0), financials_utils.GetIntelSmlVmRate(), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	//Now Create a paid instance

	time.Sleep(2 * time.Minute)

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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel User)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Enroll_New_User_Validate_cloudAcc_Attributes() {
	//pass
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Validate Notifications ToDo From DevloudAccId

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, cloudAccId)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// assert.Equal(suite.T(), int64(0), gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 on fresh user enrollment ")

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Enroll_New_User_Validate_cloudAcc_Attributes_After_adding_Credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Add coupon to the user
	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(300), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// Validate Notifications TODO: Pending dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, cloudAccId)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// assert.Equal(suite.T(), int64(0), gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 ")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Expire_Coupon_Validate_Instance_Deletion() {
	// pass
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	// Create and redeem normal coupon

	coupon_expire_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Intel", int64(5), int64(2), cloudAccId, time.Duration(3))
	assert.Equal(suite.T(), coupon_expire_err, nil, "Failed to create coupon with shorter expiry, failed with error : %s", coupon_expire_err)

	// Now launch paid instance and see API throws 403 error

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Validate Cloud Account attributes

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Not Listed in Deactivation List After Expired credits")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), remainingAmount, float64(0), "Failed to validate remaining credits")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Instance_Deletion_Upon_Usageof_AllCredits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// Push Metering Data to use all Credits

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "400000")
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
			assert.Equal(suite.T(), financials_utils.GetIntelSmlVmRate(), rateFactor, "Failed: Failed to validate rate amount")

			logger.Logf.Infof("Actual Usage ", actualAMount)
			assert.GreaterOrEqual(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")
		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.GreaterOrEqual(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	amt := 20
	assert.GreaterOrEqual(suite.T(), usedAmount, float64(amt), "Failed to validate used credits")
	assert.Equal(suite.T(), remainingAmount, "0", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	unappliedCredits = testsetup.RoundFloat(unappliedCredits, 1)
	assert.Equal(suite.T(), unappliedCredits, float64(0), "Failed : Unapplied cloud credit did not become zero")

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Not Listed in Deactivation List After Expired credits")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not deleted after all credits used")
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	// Validate cloud credit data
	zeroamt := 0
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(zeroamt), totalRemainingAmount, "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Instance_Termination_Usage_More_Than_Credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()

	// Create and redeem normal coupon

	// Create and redeem  coupon

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// Now launch paid instance

	// Push Metering Data to use all Credits

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "600000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsageinRange(cloudAccId, financials_utils.GetIntelSmlVmRate(), "vm-spr-sml", authToken, float64(30), float64(31))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : %s", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(20), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : %s", credits_err)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not deleted after all credits used")
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_User_Gets_80_percent_Notification() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()

	// Create and redeem normal coupon

	// Create and redeem  coupon
	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	// Now launch paid instance

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bm", "2000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	usage_err := billing.ValidateUsageNotZero(cloudAccId, financials_utils.GetIntelBMRate(), "bm-spr", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : %s", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateUsageCreditsinRange(cloudAccId, float64(0.8), float64(1), authToken, float64(0.07), float64(0.2), float64(0.07), float64(0.2), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : %s", credits_err)

	//Validate credits
	time.Sleep(2 * time.Minute)
	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance deleted at 80 percent")
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_User_Gets_100_percent_Notification() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()

	// Create and redeem  coupon
	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Now launch paid instance

	// Push Metering Data to use all Credits

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	now := time.Now().UTC()
	previousDate := now.Add(2 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "20000")
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
	tamt := 1.02
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s", data)
		if gjson.Get(data, "productType").String() == "vm-spr-sml" {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			actualAMount = testsetup.RoundFloat(actualAMount, 2)
			assert.GreaterOrEqual(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), financials_utils.GetIntelSmlVmRate(), rateFactor, "Failed: Failed to validate rate amount")

			logger.Logf.Infof("Actual Usage ", actualAMount)
			assert.GreaterOrEqual(suite.T(), actualAMount, float64(tamt), "Failed: Failed to validate usage amount")
		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.GreaterOrEqual(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 2)
	assert.GreaterOrEqual(suite.T(), usedAmount, float64(1), "Failed to validate used credits")
	assert.Equal(suite.T(), remainingAmount, "0", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	unappliedCredits = testsetup.RoundFloat(unappliedCredits, 1)
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), unappliedCredits, float64(0), "Failed : Validation failed in Unapplied Cloud Credit")

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not deleted after all credits used")
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	// Validate cloud credit data
	zeroamt := 0
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Equal(suite.T(), float64(zeroamt), totalRemainingAmount, "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_User_Adds_Credits_After_100_percent_Notification() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	now1 := time.Now().UTC()
	// Create and redeem normal coupon
	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	initialTS := gjson.Get(responsePayload, "creditsDepleted").String()
	assert.Equal(suite.T(), initialTS, "1970-01-01T00:00:00Z", "Failed: Credit depletion time stamp not correct")

	// Now launch paid instance

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
			assert.Equal(suite.T(), financials_utils.GetIntelBMRate(), rateFactor, "Failed: Failed to validate rate amount")

		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.Greater(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")
	assert.Less(suite.T(), total_amount_from_response, float64(maxtamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
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

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	ts := gjson.Get(responsePayload, "creditsDepleted").String()
	creditDepletedTimestamp, _ := time.Parse(time.RFC3339, ts)
	assert.Equal(suite.T(), true, creditDepletedTimestamp.After(now1), "Failed: Credit depletion time stamp not correct")

	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	coupon_err = billing.Create_Redeem_Coupon("Intel", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	time.Sleep(6 * time.Minute)

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Validate_all_Usages() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()

	// Create and redeem normal coupon
	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(100), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	//token := utils.Get_Standard_Token()
	logger.Log.Info("Compute Url : " + computeUrl)
	baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)

	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Premium User")

	skip_small := false
	skip_med := false
	skip_large := false
	skip_bm := false
	skip_gaudi := false

	var med_count int
	var small_count int
	var bm_count int
	var large_count int
	var gaudi_count int

	var VMName string
	var instance_id_created1 string
	var instance_id_created2 string
	var instance_id_created3 string
	var instance_id_created4 string
	var instance_id_created5 string

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Failed to create instance of type small due to high demand %s : ", create_response_body)
		skip_small = true

	} else {
		instance_id_created1 = gjson.Get(create_response_body, "metadata.resourceId").String()
		VMName = gjson.Get(create_response_body, "metadata.name").String()
		assert.Equal(suite.T(), create_response_status, 200, "Failed: Vm Instance creation failed")
		assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	}

	vm_name = "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload = compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-med", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Failed to create instance of type medium due to high demand %s : ", create_response_body)
		skip_med = true

	} else {
		instance_id_created2 = gjson.Get(create_response_body, "metadata.resourceId").String()
		VMName = gjson.Get(create_response_body, "metadata.name").String()
		assert.Equal(suite.T(), create_response_status, 200, "Failed: Vm Instance creation failed")
		assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	}

	vm_name = "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload = compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-lrg", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Failed to create instance of type large due to high demand %s : ", create_response_body)
		skip_large = true
	} else {
		instance_id_created3 = gjson.Get(create_response_body, "metadata.resourceId").String()
		VMName = gjson.Get(create_response_body, "metadata.name").String()
		assert.Equal(suite.T(), create_response_status, 200, "Failed: Vm Instance creation failed")
		assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	}

	vm_name = "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload = compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "bm-spr", "ubuntu-22.04-spr-metal-cloudimg-amd64-v20240115", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Failed to create instance of type BM due to high demand %s : ", create_response_body)
		skip_bm = true
	} else {
		instance_id_created4 = gjson.Get(create_response_body, "metadata.resourceId").String()
		VMName = gjson.Get(create_response_body, "metadata.name").String()
		assert.Equal(suite.T(), create_response_status, 200, "Failed: Vm Instance creation failed")
		assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	}

	vm_name = "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload = compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "bm-icp-gaudi2", "ubuntu-20.04-gaudi-metal-cloudimg-amd64-v20231013", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Failed to create instance of type BM due to high demand %s : ", create_response_body)
		skip_gaudi = true
	} else {
		instance_id_created5 = gjson.Get(create_response_body, "metadata.resourceId").String()
		VMName = gjson.Get(create_response_body, "metadata.name").String()
		assert.Equal(suite.T(), create_response_status, 200, "Failed: Vm Instance creation failed")
		assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	}

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(25 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "400020")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire
	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)
	tamt := 20
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if skip_small != true {
			if gjson.Get(data, "productType").String() == "vm-spr-sml" {
				small_count = 1
				Amount := gjson.Get(data, "amount").String()
				actualAMount, _ := strconv.ParseFloat(Amount, 64)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")

				minsUsed := gjson.Get(data, "minsUsed").String()
				minsFactor, _ := strconv.ParseFloat(minsUsed, 64)
				assert.Greater(suite.T(), minsFactor, float64(0), "Failed: Failed to validate minsUsed")

				rate := gjson.Get(data, "rate").String()
				rateFactor, _ := strconv.ParseFloat(rate, 64)
				assert.Equal(suite.T(), financials_utils.GetIntelSmlVmRate(), rateFactor, "Failed: Failed to validate rate amount")

				logger.Logf.Infof("Actual Usage :    ", actualAMount)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")
				assert.Equal(suite.T(), small_count, 1, "Failed: Small Instance Usage not displayed")
			}

		}

		if skip_med != true {
			if gjson.Get(data, "productType").String() == "vm-spr-med" {
				med_count = 1
				Amount := gjson.Get(data, "amount").String()
				actualAMount, _ := strconv.ParseFloat(Amount, 64)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")

				minsUsed := gjson.Get(data, "minsUsed").String()
				minsFactor, _ := strconv.ParseFloat(minsUsed, 64)
				assert.Greater(suite.T(), minsFactor, float64(0), "Failed: Failed to validate minsUsed")

				rate := gjson.Get(data, "rate").String()
				rateFactor, _ := strconv.ParseFloat(rate, 64)
				assert.Equal(suite.T(), financials_utils.GetIntelMedVmRate(), rateFactor, "Failed: Failed to validate rate amount")
				assert.Equal(suite.T(), med_count, 1, "Failed: Medium Usage not displayed")

			}

		}

		if skip_large != true {
			if gjson.Get(data, "productType").String() == "vm-spr-lrg" {
				large_count = 1
				Amount := gjson.Get(data, "amount").String()
				actualAMount, _ := strconv.ParseFloat(Amount, 64)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")

				minsUsed := gjson.Get(data, "minsUsed").String()
				minsFactor, _ := strconv.ParseFloat(minsUsed, 64)
				assert.Greater(suite.T(), minsFactor, float64(0), "Failed: Failed to validate minsUsed")

				rate := gjson.Get(data, "rate").String()
				rateFactor, _ := strconv.ParseFloat(rate, 64)
				assert.Equal(suite.T(), financials_utils.GetIntelLrgVmRate(), rateFactor, "Failed: Failed to validate rate amount")
				assert.Equal(suite.T(), large_count, 1, "Failed: Large Usage not displayed")

			}

		}

		if skip_bm != true {
			if gjson.Get(data, "productType").String() == "bm-spr" {
				bm_count = 1
				Amount := gjson.Get(data, "amount").String()
				actualAMount, _ := strconv.ParseFloat(Amount, 64)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")

				rate := gjson.Get(data, "rate").String()
				rateFactor, _ := strconv.ParseFloat(rate, 64)
				assert.Equal(suite.T(), financials_utils.GetIntelBMRate(), rateFactor, "Failed: Failed to validate rate amount")

				minsUsed := gjson.Get(data, "minsUsed").String()
				minsFactor, _ := strconv.ParseFloat(minsUsed, 64)
				assert.Greater(suite.T(), minsFactor, float64(0), "Failed: Failed to validate minsUsed")

				logger.Logf.Infof("Actual Usage :    ", actualAMount)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")
				assert.Equal(suite.T(), bm_count, 1, "Failed: Bm Usage not displayed")
			}

		}

		if skip_gaudi != true {
			if gjson.Get(data, "productType").String() == "bm-icp-gaudi2" {
				gaudi_count = 1
				Amount := gjson.Get(data, "amount").String()
				actualAMount, _ := strconv.ParseFloat(Amount, 64)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")

				rate := gjson.Get(data, "rate").String()
				rateFactor, _ := strconv.ParseFloat(rate, 64)
				assert.Equal(suite.T(), financials_utils.GetIntelGaudi2Rate(), rateFactor, "Failed: Failed to validate rate amount")

				minsUsed := gjson.Get(data, "minsUsed").String()
				minsFactor, _ := strconv.ParseFloat(minsUsed, 64)
				assert.Greater(suite.T(), minsFactor, float64(0), "Failed: Failed to validate minsUsed")

				logger.Logf.Infof("Actual Usage :    ", actualAMount)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")
				assert.Equal(suite.T(), gaudi_count, 1, "Failed: Gaudi Usage not displayed")
			}

		}

		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.GreaterOrEqual(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl = utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	amt := 20
	assert.GreaterOrEqual(suite.T(), usedAmount, float64(amt), "Failed to validate used credits")
	assert.Greater(suite.T(), remainingAmount, "50", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Greater(suite.T(), unappliedCredits, "0", "Failed : Unapplied cloud credit did not become zero")

	if create_response_status == 200 {
		time.Sleep(5 * time.Minute)
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		if skip_small != true {
			get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created1)
			logger.Logf.Info("get_response_status: ", get_response_status)
			logger.Logf.Info("get_response_body: ", get_response_body)
			assert.Equal(suite.T(), 200, get_response_status, "Failed : Small Instance not deleted after all credits used")
		}
		if skip_med != true {
			get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created2)
			logger.Logf.Info("get_response_status: ", get_response_status)
			logger.Logf.Info("get_response_body: ", get_response_body)
			assert.Equal(suite.T(), 200, get_response_status, "Failed : Medium Instance not deleted after all credits used")
		}

		if skip_large != true {
			get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created3)
			logger.Logf.Info("get_response_status: ", get_response_status)
			logger.Logf.Info("get_response_body: ", get_response_body)
			assert.Equal(suite.T(), 200, get_response_status, "Failed : Large Instance not deleted after all credits used")
		}

		if skip_bm != true {
			get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created4)
			logger.Logf.Info("get_response_status: ", get_response_status)
			logger.Logf.Info("get_response_body: ", get_response_body)
			assert.Equal(suite.T(), 200, get_response_status, "Failed : BM Instance not deleted after all credits used")
		}

		if skip_gaudi != true {
			get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created5)
			logger.Logf.Info("get_response_status: ", get_response_status)
			logger.Logf.Info("get_response_body: ", get_response_body)
			assert.Equal(suite.T(), 200, get_response_status, "Failed : Gaudi Instance not deleted after all credits used")
		}
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	// Validate cloud credit data
	zeroamt := 0
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Greater(suite.T(), totalRemainingAmount, float64(zeroamt), "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)
	assert.Less(suite.T(), totalRemainingAmount, float64(100), "Test Failed while deleting the cloud account(Premium User)")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium User)")

}

func (suite *BillingAPITestSuite) Test_Intel_Sanity_Validate_All_CloudAccount_Attributes() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	now1 := time.Now().UTC()
	// Create and redeem normal coupon
	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	initialTS := gjson.Get(responsePayload, "creditsDepleted").String()
	assert.Equal(suite.T(), initialTS, "1970-01-01T00:00:00Z", "Failed: Credit depletion time stamp not correct")

	// Now launch paid instance

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
			assert.Equal(suite.T(), financials_utils.GetIntelBMRate(), rateFactor, "Failed: Failed to validate rate amount")

		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.Greater(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")
	assert.Less(suite.T(), total_amount_from_response, float64(maxtamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
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

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	ts := gjson.Get(responsePayload, "creditsDepleted").String()
	creditDepletedTimestamp, _ := time.Parse(time.RFC3339, ts)
	assert.Equal(suite.T(), true, creditDepletedTimestamp.After(now1), "Failed: Credit depletion time stamp not correct")

	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	coupon_err = billing.Create_Redeem_Coupon("Intel", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	time.Sleep(6 * time.Minute)

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Intel user)")

}
