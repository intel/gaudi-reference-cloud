//go:build Functional || Billing || Intel || IntelIntegration || Integration || MultiUser
// +build Functional Billing Intel IntelIntegration Integration MultiUser

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
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// var met_ret bool

func EnrollMember(userType string, exp_acc_type string, countrycode string) (string, string, error) {
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	//Create Cloud Account
	userName := utils.Get_UserName(userType)
	memberType := utils.Get_MemberType(userType)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	base_url := utils.Get_Base_Url1()
	url := base_url + "/v1/cloudaccounts"
	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}
	time.Sleep(1 * time.Minute)

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll(memberType, tid, userName, enterpriseId, false, 200)
	if get_CAcc_id == "False" {
		return userName, get_CAcc_id, fmt.Errorf("Enroll API Failed")
	}
	if acc_type != exp_acc_type {
		return userName, get_CAcc_id, fmt.Errorf("Enroll API Failed to get proper account type, expected : %s, got :%s", exp_acc_type, acc_type)
	}

	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)

	if ret_value1 == "False" {
		return userName, get_CAcc_id, fmt.Errorf("Get on cloud account failed ")
	}
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	if countrycode != CountryCode {
		return userName, get_CAcc_id, fmt.Errorf("Get on cloud account failed get proper CountryCode, expected : %s, got :%s", countrycode, CountryCode)
	}

	return userName, get_CAcc_id, nil
}

func (suite *BillingAPITestSuite) Test_Intel_Send_Invite_Intel_Member() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Send_Invite_Standard_Member() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("StandardMember", "ACCOUNT_TYPE_STANDARD", utils.GetCountryCode("StandardMember"))
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Send_Invite_Premium_Member() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("PremiumMember", "ACCOUNT_TYPE_STANDARD", utils.GetCountryCode("PremiumMember"))
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	upgrade_err := billing.Upgrade_to_Premium_with_coupon(memcloudAccId, authToken, membertoken, int64(1))
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Send_Invite_Member_With_No_CloudAccount() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	memberEmail := utils.Get_UserName("MemberNoCloudAccount")

	// Verify Send Invite flow
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Is_Not_Able_To_Launch_Instance_when_Admin_has_no_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	// Add Standard Member

	// Verify Send Invite flow
	memberEmail1, _, errEnroll1 := EnrollMember("StandardMember", "ACCOUNT_TYPE_STANDARD", utils.GetCountryCode("StandardMember"))
	assert.Equal(suite.T(), errEnroll1, nil, "Test Failed in Enrolling Member")
	membertoken1, _ := auth.Get_Azure_Bearer_Token(memberEmail1)
	membertoken1 = "Bearer " + membertoken1

	err1 := billing.AddMember(cloudAccId, memberEmail1, userToken, membertoken1, authToken)
	assert.Equal(suite.T(), err1, nil, "Test Failed in Adding Member with error :%s", err)

	// Verify Send Invite flow
	memberEmail2, memcloudAccId2, errEnroll2 := EnrollMember("PremiumMember", "ACCOUNT_TYPE_STANDARD", utils.GetCountryCode("PremiumMember"))
	assert.Equal(suite.T(), errEnroll2, nil, "Test Failed in Enrolling Member")
	membertoken2, _ := auth.Get_Azure_Bearer_Token(memberEmail2)
	membertoken2 = "Bearer " + membertoken2

	upgrade_err := billing.Upgrade_to_Premium_with_coupon(memcloudAccId2, authToken, membertoken2, int64(1))
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	err = billing.AddMember(cloudAccId, memberEmail2, userToken, membertoken2, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	vm_creation_error1, create_response_body1, skip_val1 := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken1, 403)
	if skip_val1 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error1)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error1, nil, "Failed to create vm with error : %s", vm_creation_error1)

	vm_creation_error2, create_response_body2, skip_val2 := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken2, 403)
	if skip_val2 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error2)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error2, nil, "Failed to create vm with error : %s", vm_creation_error2)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	if vm_creation_error1 != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body1, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	if vm_creation_error2 != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body2, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Is_Able_To_Launch_Instance_when_Admin_has_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	// Add Standard Member

	// Verify Send Invite flow
	memberEmail1, _, errEnroll1 := EnrollMember("StandardMember", "ACCOUNT_TYPE_STANDARD", utils.GetCountryCode("StandardMember"))
	assert.Equal(suite.T(), errEnroll1, nil, "Test Failed in Enrolling Member")
	membertoken1, _ := auth.Get_Azure_Bearer_Token(memberEmail1)
	membertoken1 = "Bearer " + membertoken1

	err1 := billing.AddMember(cloudAccId, memberEmail1, userToken, membertoken1, authToken)
	assert.Equal(suite.T(), err1, nil, "Test Failed in Adding Member with error :%s", err)

	// Verify Send Invite flow
	memberEmail2, memcloudAccId2, errEnroll2 := EnrollMember("PremiumMember", "ACCOUNT_TYPE_STANDARD", utils.GetCountryCode("PremiumMember"))
	assert.Equal(suite.T(), errEnroll2, nil, "Test Failed in Enrolling Member")
	membertoken2, _ := auth.Get_Azure_Bearer_Token(memberEmail2)
	membertoken2 = "Bearer " + membertoken2

	upgrade_err := billing.Upgrade_to_Premium_with_coupon(memcloudAccId2, authToken, membertoken2, int64(1))
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	err = billing.AddMember(cloudAccId, memberEmail2, userToken, membertoken2, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	vm_creation_error1, create_response_body1, skip_val1 := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken1, 200)
	if skip_val1 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error1)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error1, nil, "Failed to create vm with error : %s", vm_creation_error1)

	vm_creation_error2, create_response_body2, skip_val2 := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken2, 200)
	if skip_val2 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error2)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error2, nil, "Failed to create vm with error : %s", vm_creation_error2)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	if vm_creation_error1 != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body1, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	if vm_creation_error2 != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body2, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Is_Not_Able_To_Launch_Instance_when_Admin_has_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 200)
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

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Is_Able_To_Launch_Instance_when_Admin_has_credits_Validate_Usages() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	coupon_err1 := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), memcloudAccId)
	assert.Equal(suite.T(), coupon_err1, nil, "Failed to create coupon, failed with error : ", coupon_err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	usage_err := billing.ValidateUsageNotZero(cloudAccId, financials_utils.GetIntelSmlVmRate(), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usages, validation failed with error : ", usage_err)
	time.Sleep(5 * time.Minute)
	credits_err := billing.ValidateCreditsNonZeroDepletion(cloudAccId, 20, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	usage_err = billing.ValidateZeroUsage(memcloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)
	credits_err = billing.ValidateCredits(memcloudAccId, float64(0), authToken, float64(20), float64(20), float64(20))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Is_Not_Able_To_Launch_Instance_when_Admin_Uses_All_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(15), int64(2), cloudAccId)
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

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-med", membertoken, 403)
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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Is_Not_Able_To_Launch_Instance_when_Admin_Uses_Less_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

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

	// Now launch paid instance and see API throws 403 error

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-med", membertoken, 200)
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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Is_Not_Able_To_Launch_Instance_when_Admin_has_expired_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	coupon_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Intel", int64(15), int64(2), cloudAccId, time.Duration(5))
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 403)
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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Removed_Member_Is_Not_Able_To_Launch_Instance_in_Admin() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Remove member invitation

	_, err1 := billing.RemoveMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err1, nil, "Test Failed in Removing Member with error :%s", err)

	// Now try to launch instance

	vm_creation_error1, create_response_body1, skip_val1 := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 403)
	if skip_val1 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.NotEqual(suite.T(), vm_creation_error1, "failed to create ssh key", "Failed to create vm with error : %s", vm_creation_error1)
	assert.Contains(suite.T(), create_response_body1, "failed to create ssh key", "Removed memeber could create ssh keys")

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

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Removed_Add_Member_Is_Able_To_Launch_Instance_in_Admin() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	// Remove member invitation

	_, err1 := billing.RemoveMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err1, nil, "Test Failed in Removing Member with error :%s", err)

	// Now try to launch instance

	vm_creation_error1, create_response_body1, skip_val1 := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 403)
	if skip_val1 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.NotEqual(suite.T(), vm_creation_error1, nil, "Failed to create vm with error : %s", vm_creation_error1)
	assert.Contains(suite.T(), create_response_body1, "failed to create ssh key", "Removed memeber could create ssh keys")

	err = billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	vm_creation_error, create_response_body, skip_val = billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 200)
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

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Usage_and_Credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	coupon_err1 := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), memcloudAccId)
	assert.Equal(suite.T(), coupon_err1, nil, "Failed to create coupon, failed with error : ", coupon_err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(memcloudAccId, "vm-spr-sml", membertoken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + memcloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	usage_err := billing.ValidateUsageNotZero(memcloudAccId, financials_utils.GetIntelSmlVmRate(), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usages, validation failed with error : ", usage_err)
	time.Sleep(5 * time.Minute)
	credits_err := billing.ValidateCreditsNonZeroDepletion(memcloudAccId, 20, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	usage_err = billing.ValidateZeroUsage(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)
	credits_err = billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(20), float64(20), float64(20))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Is_Able_To_Launch_Instance_In_Members_Account_when_Admin_Uses_All_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	// Verify Send Invite flow
	memberEmail, memcloudAccId, errEnroll := EnrollMember("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	coupon_err = billing.Create_Redeem_Coupon("Intel", int64(14), int64(2), memcloudAccId)
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

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-med", membertoken, 403)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	vm_creation_error1, create_response_body1, skip_val1 := billing.Create_Vm_Instance(memcloudAccId, "vm-spr-med", membertoken, 200)
	if skip_val1 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error1)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error1, nil, "Failed to create vm with error : %s", vm_creation_error1)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	if vm_creation_error1 != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + memcloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body1, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, membertoken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}

///  Member with no cloud account tests

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Without_Account_Is_Not_Able_To_Launch_Instance_when_Admin_has_no_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	memberEmail := utils.Get_UserName("MemberNoCloudAccount")

	// Verify Send Invite flow
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	// Add Standard Member

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 403)
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

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_With_No_Cloud_Acc_Is_Able_To_Launch_Instance_when_Admin_has_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	memberEmail := utils.Get_UserName("MemberNoCloudAccount")

	// Verify Send Invite flow
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 200)
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

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Without_CloudAccount_Is_Able_To_Launch_Instance_when_Admin_has_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	memberEmail := utils.Get_UserName("MemberNoCloudAccount")

	// Verify Send Invite flow
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 200)
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

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Without_CloudAcc_Id_Is_Able_To_Launch_Instance_when_Admin_has_credits_Validate_Usages() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	memberEmail := utils.Get_UserName("MemberNoCloudAccount")

	// Verify Send Invite flow
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(20), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", membertoken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	if vm_creation_error != nil {
		time.Sleep(5 * time.Minute)
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
		time.Sleep(10 * time.Second)
	}

	usage_err := billing.ValidateUsageNotZero(cloudAccId, financials_utils.GetIntelSmlVmRate(), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usages, validation failed with error : ", usage_err)
	time.Sleep(5 * time.Minute)
	credits_err := billing.ValidateCreditsNonZeroDepletion(cloudAccId, 20, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Check_Member_Without_Cloud_Acc_Is_Not_Able_To_Launch_Instance_when_Admin_Uses_All_credits() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()
	memberEmail := utils.Get_UserName("MemberNoCloudAccount")

	// Verify Send Invite flow
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail)
	membertoken = "Bearer " + membertoken

	err := billing.AddMember(cloudAccId, memberEmail, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(15), int64(2), cloudAccId)
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

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-med", membertoken, 403)
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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}
