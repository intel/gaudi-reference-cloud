//go:build Functional || Billing || Intel || IntelIntegration || Integration || MultiUserOPA
// +build Functional Billing Intel IntelIntegration Integration MultiUserOPA

package BillingAPITest

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"os"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func EnrollMember1(userType string, exp_acc_type string, countrycode string) (string, string, error) {
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

func (suite *BillingAPITestSuite) Test_Intel_Verify_Member_Cannot_SendInvite() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	invite_url := utils.Get_Base_Url1() + "/v1/cloudaccounts/invitations/create"
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Verify Send Invite flow
	memberEmail1, memcloudAccId, errEnroll := EnrollMember1("IntelMember", "ACCOUNT_TYPE_INTEL", "")
	assert.Equal(suite.T(), errEnroll, nil, "Test Failed in Enrolling Member")
	membertoken, _ := auth.Get_Azure_Bearer_Token(memberEmail1)
	membertoken = "Bearer " + membertoken

	memberEmail := "test@test.com"

	err := billing.AddMember(cloudAccId, memberEmail1, userToken, membertoken, authToken)
	assert.Equal(suite.T(), err, nil, "Test Failed in Adding Member with error :%s", err)

	// Check Member is not able to create otp

	create_otp_url := utils.Get_Base_Url1() + "/v1/otp/create"
	create_otp_payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"memberEmail": "%s"
	}`, cloudAccId, memberEmail)
	logger.Logf.Infof("Create OTP url :%s", create_otp_url)
	logger.Logf.Infof("Create OTP Payload :%s", create_otp_payload)
	code, body := financials.CreateOTP(create_otp_url, membertoken, create_otp_payload)
	logger.Logf.Infof("SendInvite response code  :  %d ", code)
	logger.Logf.Infof("SendInvite response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode := gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Check Member is not able to verify otp
	logger.Log.Info("Checking Member is not able to verify otp")
	verify_otp_payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"memberEmail": "%s",
		"otpCode": "%s"
	}`, cloudAccId, memberEmail, "1234567")
	logger.Logf.Infof("Verify OTP url :%s", base_url)
	logger.Logf.Infof("Verify OTP Payload :%s", verify_otp_payload)

	code, body = financials.VerifyOTP(base_url, membertoken, verify_otp_payload)
	logger.Logf.Infof("SendInvite response code  :  %d ", code)
	logger.Logf.Infof("SendInvite response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Check admin is not able to verify invite
	logger.Log.Info("Checking admin is not able to verify invite")
	payload := fmt.Sprintf(`{
		"adminCloudAccountId": "%s",
		"inviteCode": "%s",
		"memberEmail": "%s"
	}`, cloudAccId, "1234567", memberEmail)
	fmt.Println("Payload: ", payload)
	code, body = financials.VerifyInviteCode(base_url, userToken, payload)
	logger.Logf.Infof("VerifyInviteCode response code  :  %d ", code)
	logger.Logf.Infof("VerifyInviteCode response body :  %s ", body)
	logger.Logf.Infof("SendInvite response code  :  %d ", code)
	logger.Logf.Infof("SendInvite response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Check Member is not able to send invite
	logger.Log.Info("Check Member is not able to send invite")
	code, body = financials.CreateInviteCode(invite_url, membertoken, cloudAccId, memberEmail)
	logger.Logf.Infof("SendInvite response code  :  %d ", code)
	logger.Logf.Infof("SendInvite response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Member is not able to read admin cloud account invitations
	logger.Log.Info("Verify Member is not able to read admin cloud account invitations")
	code, body = financials.ReadInvitations(base_url, membertoken, cloudAccId, "")
	logger.Logf.Infof("ReadInvitations response code  :  %d ", code)
	logger.Logf.Infof("ReadInvitations response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Admin is not able to read member cloud account invitations
	logger.Log.Info("Verify Admin is not able to read member cloud account invitations")
	code, body = financials.ReadInvitations(base_url, userToken, memcloudAccId, "")
	logger.Logf.Infof("ReadInvitations response code  :  %d ", code)
	logger.Logf.Infof("ReadInvitations response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Member is not able to remove other members
	logger.Log.Info("Verify Member is not able to remove other members")
	code, body = financials.RemoveInvitation(base_url, membertoken, cloudAccId, "test@test.com")
	logger.Logf.Infof("RemoveMember response code  :  %d ", code)
	logger.Logf.Infof("RemoveMember response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Member is not able to remove self
	logger.Log.Info("Verify Member is not able to remove self")
	code, body = financials.RemoveInvitation(base_url, membertoken, cloudAccId, memberEmail)
	logger.Logf.Infof("RemoveMember response code  :  %d ", code)
	logger.Logf.Infof("RemoveMember response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Member can not reject other members invitations
	logger.Log.Info("Verify Member can not reject other members invitations")
	code, body = financials.Rejectinvitation(base_url, userToken, cloudAccId, "INVITE_STATE_ACCEPTED", memberEmail)
	logger.Logf.Infof("RemoveMember response code  :  %d ", code)
	logger.Logf.Infof("RemoveMember response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Get Members of admin with member token
	logger.Log.Info("Verify Get Members of admin with member token")
	code, body = financials.GetActiveMembers(base_url, membertoken, userName)
	logger.Logf.Infof("GetMembers response code  :  %d ", code)
	logger.Logf.Infof("GetMembers response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Get Members of member with admin token
	logger.Log.Info("Verify Get Members of member with admin token")
	code, body = financials.GetActiveMembers(base_url, userToken, memberEmail1)
	logger.Logf.Infof("GetMembers response code  :  %d ", code)
	logger.Logf.Infof("GetMembers response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify ResendInvitation with member token
	logger.Log.Info("Verify ResendInvitation with member token")
	payload = fmt.Sprintf(`{
		"adminAccountId": "%s",
		"memberEmail": "%s"
	}`, cloudAccId, memberEmail)
	resendUrl := base_url + "/v1/cloudaccounts/invitations/resend"
	code, body = financials.Resendinvitation(resendUrl, membertoken, payload)
	logger.Logf.Infof("Resendinvitation response code  :  %d ", code)
	logger.Logf.Infof("Resendinvitation response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Member is not able to access financial information of admin
	logger.Log.Info("Verify Member is not able to access financial information of admin")
	creditsUrl := utils.Get_Credits_Base_Url() + "/credit"
	code, body = financials.GetCredits(creditsUrl, membertoken, cloudAccId)
	logger.Logf.Infof("GetMembers response code  :  %d ", code)
	logger.Logf.Infof("GetMembers response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Member is not able to see usage of admin account
	logger.Log.Info("Verify Member is not able to see usage of admin account")
	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	code, body = financials.GetUsage(usage_url, membertoken)
	logger.Logf.Infof("GetMembers response code  :  %d ", code)
	logger.Logf.Infof("GetMembers response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Member is not able redeem coupon for admin
	logger.Log.Info("Verify Member is not able redeem coupon for admin")
	redeem_coupon_endpoint := utils.Get_Credits_Base_Url() + "/coupons/redeem"
	redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), "ABCD-EFGH-IJKL", cloudAccId)
	code, body = financials.RedeemCoupon(redeem_coupon_endpoint, membertoken, redeem_payload)
	logger.Logf.Infof("GetMembers response code  :  %d ", code)
	logger.Logf.Infof("GetMembers response body :  %s ", body)
	assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	errmsgCode = gjson.Get(body, "code").String()
	assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	// Verify Member is not able to delete members of admin
	// logger.Log.Info("Verify Member is not able to delete members of admin")
	// code, body = financials.DeleteMember(base_url, membertoken, cloudAccId, memberEmail)
	// logger.Logf.Infof("DeleteMember response code  :  %d ", code)
	// logger.Logf.Infof("DeleteMember response body :  %s ", body)
	// assert.Equal(suite.T(), 403, code, "Test Failed : Did not receive 403 status code")
	// assert.NotEqual(suite.T(), 200, code, "Test Failed : Should not receive 200 status code")
	// errmsgCode = gjson.Get(body, "code").String()
	// assert.Equal(suite.T(), "7", errmsgCode, "Test Failed : Code in error message is not correct")

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(memcloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")

}
