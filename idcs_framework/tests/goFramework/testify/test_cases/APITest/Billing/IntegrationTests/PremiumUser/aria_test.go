//go:build Functional || Billing || Premium || PremiumIntegration || AriaTest
// +build Functional Billing Premium PremiumIntegration AriaTest

package PremiumBillingAPITest

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) Test_Premium_Validate_Master_Plan_Creation() {
	// Standard user is already enrolled, so start upgrade to premium
	userName := utils.Get_UserName("Premium")
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
	cloudAccId, _ = testsetup.GetCloudAccountId(userName, base_url, authToken)

	ariaclientId, ariaAuth := utils.Get_Aria_Config()
	logger.Logf.Infof("Aria details, ariaclientId", ariaclientId)
	logger.Logf.Infof("Aria details, ariaAuth", ariaAuth)

	client_acct_id := "idc." + cloudAccId
	response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(cloudAccId, ariaclientId, ariaAuth)
	client_acc_id1 := gjson.Get(responseBody, "client_acct_id").String()
	master_plan_count := gjson.Get(responseBody, "master_plan_count").String()
	logger.Logf.Infof("Aria Response ", responseBody)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Details from Aria")
	assert.Equal(suite.T(), client_acc_id1, client_acct_id, "Validation failed fetching billing account details from aria, Billing Account number did not match")
	assert.Equal(suite.T(), master_plan_count, "1", "Validation failed fetching billing account details from aria")
	//Expect(strings.Contains(responseBody, `"master_plan_count" : 1`)).To(BeTrue(), "Validation failed fetching billing account details from aria")
	//Expect(strings.Contains(responseBody, `"error_msg" : "OK"`)).To(BeTrue(), "Validation failed fetching billing account details from aria")

}

func (suite *BillingAPITestSuite) Test_Premium_Validate_Vm_Plan_Creation() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Premium")
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
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(15), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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

	ariaclientId, ariaAuth := utils.Get_Aria_Config()
	logger.Logf.Infof("Aria details, ariaclientId", ariaclientId)
	logger.Logf.Infof("Aria details, ariaAuth", ariaAuth)

	client_acct_id := "idc." + cloudAccId
	response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(cloudAccId, ariaclientId, ariaAuth)
	logger.Logf.Infof("Account Details response  %s ", responseBody)
	client_acc_id1 := gjson.Get(responseBody, "client_acct_id").String()
	master_plan_count := gjson.Get(responseBody, "master_plan_count").String()
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Details from Aria")
	assert.Equal(suite.T(), client_acc_id1, client_acct_id, "Validation failed fetching billing account details from aria, Billing Account number did not match")
	assert.Equal(suite.T(), master_plan_count, "1", "Validation failed fetching billing account details from aria")

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), utils.GenerateString(10), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "240000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire
	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())

	//time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	time.Sleep(20 * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	usage_err := billing.ValidateUsage(cloudAccId, float64(30), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	response_status, responseBody = financials.GetAriaAccountDetailsAllForClientId(cloudAccId, ariaclientId, ariaAuth)
	logger.Logf.Infof("Account Details response  %s ", responseBody)
	client_acc_id1 = gjson.Get(responseBody, "client_acct_id").String()
	master_plan_count = gjson.Get(responseBody, "master_plan_count").String()
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Details from Aria")
	assert.Equal(suite.T(), client_acc_id1, client_acct_id, "Validation failed fetching billing account details from aria, Billing Account number did not match")
	assert.Equal(suite.T(), master_plan_count, "2", "Validation failed fetching billing account details from aria")

	_, responseBody1 := financials.Get_Client_Plans(ariaclientId, ariaAuth, cloudAccId)
	logger.Logf.Infof("Aria Plan Response %s ", responseBody1)
	now = time.Now().UTC()
	todatDate := now.Format("2006-01-02")
	billDate := now.AddDate(0, 0, 30).Format("2006-01-02")
	result := gjson.Parse(responseBody1)
	master_plan_count1 := 0
	vm_plan_count := 0
	arr := gjson.Get(result.String(), "acct_plans_m")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if gjson.Get(data, "plan_name").String() == "IDC Master Plan" {
			master_plan_count1 = master_plan_count1 + 1
			planDate := gjson.Get(data, "plan_date").String()
			billThroughDate := gjson.Get(data, "bill_thru_date").String()
			assignmentDate := gjson.Get(data, "plan_assignment_date").String()

			assert.Equal(suite.T(), planDate, todatDate, "Master Plan date did not match")
			assert.Equal(suite.T(), billDate, billThroughDate, "Master Plan Bill Through date did not match")
			assert.Equal(suite.T(), todatDate, assignmentDate, "Master Plan Assignment date did not match")
			assert.Equal(suite.T(), todatDate, planDate, "Master Plan date did not match")
		}

		if gjson.Get(data, "plan_name").String() == "vm-spr-sml" {
			vm_plan_count = vm_plan_count + 1
			planDate := gjson.Get(data, "plan_date").String()
			billThroughDate := gjson.Get(data, "bill_thru_date").String()
			assignmentDate := gjson.Get(data, "plan_assignment_date").String()
			logger.Logf.Infof("planDate : %s", planDate)
			logger.Logf.Infof("billThroughDate : %s", billThroughDate)
			logger.Logf.Infof("assignmentDate : %s", assignmentDate)
			// assert.Equal(suite.T(), planDate, todatDate, "Vm Plan  date did not match")
			// assert.Equal(suite.T(), billDate, billThroughDate, "Vm Plan Bill Through date did not match")
			// assert.Equal(suite.T(), todatDate, assignmentDate, "Vm Plan Assignment date did not match")
			// assert.Equal(suite.T(), todatDate, planDate, "Vm Plan date did not match")
		}
		return true // keep iterating
	})

	assert.Equal(suite.T(), master_plan_count1, 1, "Master Plan id not found in response")
	assert.Equal(suite.T(), vm_plan_count, 1, "Vm  Plan id not found in response")

}
