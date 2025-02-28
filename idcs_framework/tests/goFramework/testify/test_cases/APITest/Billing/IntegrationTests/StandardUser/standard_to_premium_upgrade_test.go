//go:build Functional || Billing || Standard || Integration || UpgradeCredit || Upgrade || StandardIntegration
// +build Functional Billing Standard Integration UpgradeCredit Upgrade StandardIntegration

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
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon_When_Account_Has_No_Credits() {
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

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon_When_Account_Has_Credits() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	base_url := utils.Get_Base_Url1()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

	// Redeem coupon for standard user before upgrade

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(10), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "failed to migrate credits for cloud account with error : ", upgrade_err)

	// Migrate Credits

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

	usage_err := billing.ValidateZeroUsage(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(20), float64(20), float64(20))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon_When_Account_Has_Credits_Check_Credit_Migration_With_Coupon_Redemption() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	base_url := utils.Get_Base_Url1()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

	// Redeem coupon for standard user before upgrade

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(10), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "failed to migrate credits for cloud account with error : ", upgrade_err)
	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)
	// Migrate Credits

	coupon_err = billing.Create_Redeem_Coupon("Premium", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")

	// Validate credit details

	usage_err := billing.ValidateZeroUsage(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(30), float64(30), float64(30))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon_When_Account_Has_Credits_And_Usage() {
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

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon_When_Account_Has_Credits_And_Instances() {
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
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

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

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

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

	usage_err := billing.ValidateZeroUsage(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	assert.GreaterOrEqual(suite.T(), usedAmount, float64(0), "Failed to validate used credits")
	assert.LessOrEqual(suite.T(), remainingAmount, float64(20), "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.LessOrEqual(suite.T(), unappliedCredits, "20", "Failed : Unapplied cloud credit did not become zero")

	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
	get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
	logger.Logf.Info("get_response_status: ", get_response_status)
	logger.Logf.Info("get_response_body: ", get_response_body)
	assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not Found after Standard to premium upgrade")

	if vm_creation_error == nil {
		time.Sleep(2 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not running")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon_When_Account_Has_Credits_And_Instances_Validate_Usage() {
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
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

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

	time.Sleep(3 * time.Minute)

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

	time.Sleep(2 * time.Minute)
	// Check instance is running and can report usages

	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
	get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
	logger.Logf.Info("get_response_status: ", get_response_status)
	logger.Logf.Info("get_response_body: ", get_response_body)
	assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not Found after Standard to premium upgrade")

	// Wait for some time to get usages

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	usage_err := billing.ValidateUsageNotZero(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usages, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCreditsNonZeroDepletion(cloudAccId, 20, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	if vm_creation_error == nil {
		time.Sleep(2 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not running")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Multiple_Coupon_When_Account_Has_Credits_And_Instances_Validate_Usage() {
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
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

	// Redeem coupon for standard user before upgrade

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Create instance and upgrade

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with with error : ", vm_creation_error)

	time.Sleep(3 * time.Minute)

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

	time.Sleep(2 * time.Minute)
	// Check instance is running and can report usages

	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
	get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
	logger.Logf.Info("get_response_status: ", get_response_status)
	logger.Logf.Info("get_response_body: ", get_response_body)
	assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not Found after Standard to premium upgrade")

	// Wait for some time to get usages

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	usage_err := billing.ValidateUsageNotZero(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usages, validation failed with error : ", usage_err)
	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCreditsNonZeroDepletion(cloudAccId, 20, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	if vm_creation_error == nil {
		time.Sleep(2 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not running")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Validate_Credit_Migration_After_Adding_Coupon() {
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
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

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

	time.Sleep(3 * time.Minute)

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(10), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	// migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	// assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")

	// Validate credit details

	time.Sleep(2 * time.Minute)
	// Check instance is running and can report usages

	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
	get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
	logger.Logf.Info("get_response_status: ", get_response_status)
	logger.Logf.Info("get_response_body: ", get_response_body)
	assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not Found after Standard to premium upgrade")

	// Wait for some time to get usages

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	usage_err := billing.ValidateUsageNotZero(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usages, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCreditsNonZeroDepletion(cloudAccId, 10, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Add new coupon and check credits migrated

	coupon_err = billing.Create_Redeem_Coupon("Premium", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	time.Sleep(2 * time.Minute)
	credits_err = billing.ValidateCreditsNonZeroDepletion(cloudAccId, 30, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	if vm_creation_error == nil {
		time.Sleep(2 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not running")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon_When_Account_Has_Deleted_Instances_Validate_Usage() {
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
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

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

	vm_creation_error1, create_response_body1, skip_val1 := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val1 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error1, nil, "Failed to create vm with with error : ", vm_creation_error)

	time.Sleep(10 * time.Minute) // Delete the instance

	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	time.Sleep(10 * time.Second)
	instance_id_created := gjson.Get(create_response_body1, "metadata.resourceId").String()
	// delete the instance created
	_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
	time.Sleep(10 * time.Second)

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(10), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")

	// Validate credit details

	time.Sleep(10 * time.Minute)
	// Check instance is running and can report usages

	instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
	get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
	logger.Logf.Info("get_response_status: ", get_response_status)
	logger.Logf.Info("get_response_body: ", get_response_body)
	assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not Found after Standard to premium upgrade")

	instance_id_created = gjson.Get(create_response_body1, "metadata.resourceId").String()
	// delete the instance created
	instance_endpoint = computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	get_response_status, get_response_body = frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
	logger.Logf.Info("get_response_status: ", get_response_status)
	logger.Logf.Info("get_response_body: ", get_response_body)
	assert.Equal(suite.T(), 404, get_response_status, "Failed : Instance not found")

	// Wait for some time to get usages

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	usage_err := billing.ValidateUsageNotZero(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usages, validation failed with error : ", usage_err)
	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCreditsNonZeroDepletion(cloudAccId, 20, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	if vm_creation_error == nil {
		time.Sleep(5 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not running")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon_When_Usage_is_more_than_credits1() {
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
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

	// Redeem coupon for standard user before upgrade

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Create instance and upgrade

	// Push some usage and let credit depletion happen
	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with with error : ", vm_creation_error)


	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smlvm", "240000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
	

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	usage_err := billing.ValidateUsageinRange(cloudAccId, float64(0.0075), "vm-spr-sml", authToken, float64(30), float64(31))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	credits_err := billing.ValidateCredits(cloudAccId, float64(10), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

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
	time.Sleep(2 * time.Minute)
	// Check instance is running and can report usages

	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
	get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
	logger.Logf.Info("get_response_status: ", get_response_status)
	logger.Logf.Info("get_response_body: ", get_response_body)
	assert.Equal(suite.T(), 404, get_response_status, "Failed : Instance not Found after Standard to premium upgrade")

	// Wait for some time to get usages

	//time.Sleep(3 * time.Minute)

	usage_err = billing.ValidateUsage(cloudAccId, float64(0), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(0), float64(0.015), "vm-spr-med", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(10), float64(10), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	if vm_creation_error == nil {
		time.Sleep(2 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 404, get_response_status, "Failed : Instance not running")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon_When_Usage_is_more_than_credits() {
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
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

	// Redeem coupon for standard user before upgrade

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Create instance and upgrade

	// Push some usage and let credit depletion happen

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smlvm", "240000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with with error : ", vm_creation_error)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsageinRange(cloudAccId, float64(0.0075), "vm-spr-sml", authToken, float64(30), float64(31))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	credits_err := billing.ValidateCredits(cloudAccId, float64(10), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// User redeems a new coupon before upgrade

	coupon_err = billing.Create_Redeem_Coupon("Standard", int64(45), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

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

	time.Sleep(2 * time.Minute)
	// Check instance is running and can report usages

	// Wait for some time to get usages
	usage_err = billing.ValidateUsage(cloudAccId, float64(0), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)
	credits_err = billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(35), float64(35), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	if vm_creation_error == nil {
		time.Sleep(2 * time.Minute)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("get_response_status: ", get_response_status)
		logger.Logf.Info("get_response_body: ", get_response_body)
		assert.Equal(suite.T(), 200, get_response_status, "Failed : Instance not running")
		del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("del_response_status: ", del_response_status)
		logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Standard_User_Upgrade_to_Premium_Using_Coupon_Generate_Invoice() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	base_url := utils.Get_Base_Url1()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	ariaclientId, ariaAuth := utils.Get_Aria_Config()
	logger.Logf.Infof("Aria details, ariaclientId", ariaclientId)
	logger.Logf.Infof("Aria details, ariaAuth", ariaAuth)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account Type")

	// Redeem coupon for standard user before upgrade

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(10), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Push some usage and let credit depletion happen

	auto_app_response_status, auto_app_response_body := financials.SetAutoApprovalToFalse(cloudAccId, ariaclientId, ariaAuth)
	logger.Logf.Infof("Response code auto approval : ", auto_app_response_status)
	logger.Logf.Infof("Response body auto approval : ", auto_app_response_body)

	now := time.Now().UTC()
	previousDate := now.AddDate(0, 0, -25).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-med", "medvm", "180000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	time.Sleep(6 * time.Minute)

	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	response_status, response_body := financials.GetAriaPendingInvoiceNumberForClientId(cloudAccId, ariaclientId, ariaAuth)
	logger.Logf.Infof("Aria details, response_body", response_body)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve pending invoice number")
	json := gjson.Parse(response_body)
	pendingInvoice := json.Get("pending_invoice")
	var directive int64 = 2
	pendingInvoice.ForEach(func(_, value gjson.Result) bool {
		invoiceNo := value.Get("invoice_no").String()
		fmt.Println("Discarding pending Invoice No:", invoiceNo)
		response_status, response_body = financials.ManageAriaPendingInvoiceForClientId(cloudAccId, invoiceNo, ariaclientId, ariaAuth, directive)
		logger.Logf.Infof("Response code get pending invoices : ", response_status)
		logger.Logf.Infof("Response body get pending invoices : ", response_body)
		assert.Equal(suite.T(), response_status, 200, "Failed to discard pending invoice number")
		return true
	})

	response_status, response_body = financials.GenerateAriaInvoiceForClientId(cloudAccId, ariaclientId, ariaAuth)
	assert.Equal(suite.T(), response_status, 200, "Failed to Generate Invoice")
	json = gjson.Parse(response_body)
	pendingInvoice = json.Get("out_invoices")
	logger.Logf.Infof("Pending invoices ", pendingInvoice)
	var directive1 int64 = 1
	var medInvoiceNo string
	pendingInvoice.ForEach(func(_, value gjson.Result) bool {
		medInvoiceNo = value.Get("invoice_no").String()
		logger.Logf.Infof("Approving pending Invoice No:", medInvoiceNo)
		response_status, response_body = financials.ManageAriaPendingInvoiceForClientId(cloudAccId, medInvoiceNo, ariaclientId, ariaAuth, directive1)
		logger.Logf.Infof("Response code generate invoices : ", response_status)
		logger.Logf.Infof("Response body generate invoices : ", response_body)
		assert.Equal(suite.T(), response_status, 200, "Failed to Approving pending Invoice")
		return true
	})

	logger.Logf.Infof("Get billing invoice for clientId")
	url := base_url + "/v1/billing/invoices"
	respCode, invoices := financials.GetInvoice(url, userToken, cloudAccId)
	assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice for clientId")
	logger.Logf.Infof("invoices in account :", invoices)

	jsonInvoices := gjson.Parse(invoices).Get("invoices")
	flag := false
	jsonInvoices.ForEach(func(_, value gjson.Result) bool {
		invoiceNo := value.Get("id").String()
		logger.Logf.Infof(" Processing invoiceNo : ", invoiceNo)
		if invoiceNo == medInvoiceNo {
			assert.Equal(suite.T(), value.Get("total").String(), "45", "Total amount in invoice did not match")
			assert.Equal(suite.T(), value.Get("paid").String(), "45", "Total amount in invoice did not match")
			assert.Equal(suite.T(), value.Get("due").String(), "0", "Due amount in invoice did not match")
			assert.Equal(suite.T(), value.Get("status").String(), "paid", "Due amount in invoice did not match")
			flag = true
		}
		// Bug is open for download link
		// downloadLink := value.Get("downloadLink").String()
		//Expect(downloadLink).NotTo(BeNil(), "Invoice download link unavailable nil.")

		//invoice details
		url := base_url + "/v1/billing/invoices/detail"
		//TOdo invoiceNo
		respCode, detail := financials.GetInvoicewithInvoiceId(url, userToken, cloudAccId, invoiceNo)
		assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice details for clientId")

		logger.Logf.Infof("Invoice details : ", detail) // Empty Response

		// invoices statement
		url = base_url + "/v1/billing/invoices/statement"
		//TOdo invoiceNo
		respCode, statement := financials.GetInvoicewithInvoiceId(url, userToken, cloudAccId, invoiceNo)

		assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice statement for clientId")
		logger.Logf.Infof(" Invoice statement", statement)
		return true
	})

	//invoices unbilled
	url = base_url + "/v1/billing/invoices/unbilled"
	respCode, resp := financials.GetInvoice(url, userToken, cloudAccId)
	logger.Logf.Infof(" Processing unbilled invoices  : ", resp)
	assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice for clientId")
	assert.Equal(suite.T(), flag, true, "Can not get invoice in user account with number ", medInvoiceNo)

	// Check cloud account attributes after upgrade

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}
