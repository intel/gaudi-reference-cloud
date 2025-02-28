//go:build Functional || DailyRun
// +build Functional DailyRun

package PremiumBillingAPITest

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

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Create_Paid_Instance_With_Coupon_Credits() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(1), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Redeem_Coupon_After_Using_Less_Credits() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(20), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-lrg", "largevm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

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

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	time.Sleep(8 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(15), float64(15), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Redeem_Coupon_After_Using_More_Credits() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
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

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-lrg", "largevm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

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

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	time.Sleep(8 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(5), float64(5), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Redeem_Coupon_After_Using_All_Credits() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
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

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-lrg", "largevm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

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

	// Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Redeem coupon

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Validate credits

	time.Sleep(8 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(15), float64(15), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Usage_More_Than_Credits_Redeem_Lesser_Value_Coupon() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
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

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-icp-gaudi2", "gaudi2vm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	usage_err := billing.ValidateUsageinRange(cloudAccId, float64(0.1733), "bm-icp-gaudi2", authToken, float64(86.2), float64(87))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : %s", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Redeem coupon with lesser value than unapplied credits

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(10), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Check instance should not be launched

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-med", userToken, 403)
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

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Generate_Invoice_Using_Coupon() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ariaclientId, ariaAuth := utils.Get_Aria_Config()
	logger.Logf.Infof("Aria details, ariaclientId", ariaclientId)
	logger.Logf.Infof("Aria details, ariaAuth", ariaAuth)

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(1000), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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

	time.Sleep(10 * time.Minute)

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
			assert.Equal(suite.T(), value.Get("status").String(), "Paid", "Due amount in invoice did not match")
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

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Enroll_New_User_Validate_cloudAcc_Attributes() {
	// pass
	logger.Log.Info("Starting Test : Launch Paid instance without coupon or credit card")

	apiKey := utils.GetMailSlurpKey()
	inboxIdPremium := utils.GetInboxIdPremium()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdPremium %s", inboxIdPremium)
	mailTime := time.Now().Add(5 * time.Minute)

	// Enroll customer
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
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(1), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodePremium(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Validate Notifications ToDo From Dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, cloudAccId)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// assert.Equal(suite.T(), int64(0), gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 on fresh user enrollment ")

	// Validate Mail Notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdPremium, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Premium user did not receive 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdPremium, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Premium User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdPremium, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Premium User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Enroll_New_User_Validate_cloudAcc_Attributes_After_adding_CreditCard() {
	// pass
	suite.T().Skip()
	apiKey := utils.GetMailSlurpKey()
	inboxIdPremium := utils.GetInboxIdPremium()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdPremium %s", inboxIdPremium)
	mailTime := time.Now().Add(5 * time.Minute)

	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
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

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// Validate Mail Notifications

	time.Sleep(time.Duration(10 * time.Minute))

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdPremium, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Premium user did not receive 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdPremium, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Premium User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdPremium, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Premium User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// Validate Notifications TODO: Pending dev

	// response_status, responseBody := financials.GetNotificationsShortPoll(base_url, authToken, cloudAccId)
	// fmt.Println("Response Status to get notifications", response_status)
	// fmt.Println("Response Body to get notifications", responseBody)
	// assert.Equal(suite.T(), int64(0), gjson.Get(responseBody, "numberOfNotifications").Int(), "Validation failed on Notifications, User Notifications should be 0 ")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Validate_Instance_Not_Deleted_With_Credit_Card() {
	// pass
	suite.T().Skip()
	apiKey := utils.GetMailSlurpKey()
	inboxIdPremium := utils.GetInboxIdPremium()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdPremium %s", inboxIdPremium)
	mailTime := time.Now().Add(5 * time.Minute)

	//token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
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

	// Now launch paid instance and see API throws 403 error

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	// Push Metering Data to use all Credits

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	// Validate Cloud Account attributes

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Check Cloud account not in deactivation list

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation With Credit Card")

	//Validate credits

	// Check Cloud account not in deactivation list

	status, instanceListed = financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// Validate Mail Notifications

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdPremium, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Premium user did not receive 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdPremium, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Premium User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdPremium, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Premium User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// Delete instance
	computeUrl := utils.Get_Compute_Base_Url()
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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Instance_Termination_Usage_More_Than_Credits() {
	logger.Log.Info("Starting Test : Launch Paid instance without coupon or credit card")
	// Enroll customer

	apiKey := utils.GetMailSlurpKey()
	inboxIdPremium := utils.GetInboxIdPremium()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdPremium %s", inboxIdPremium)

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
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(20), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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

	// Now launch paid instance
	mailTime := time.Now().Add(5 * time.Minute)
	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
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
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "240000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	usage_err := billing.ValidateUsageinRange(cloudAccId, float64(0.0075), "vm-spr-sml", authToken, float64(30), float64(31))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : %s", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateCredits(cloudAccId, float64(20), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : %s", credits_err)
	computeUrl := utils.Get_Compute_Base_Url()
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

	// Validate Mail Notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	// emailNotification80, err := financials.GetMailFromInbox(inboxIdPremium, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	// assert.NotEqual(suite.T(), emailNotification80, "", "Premium User received 80% notification ")
	// assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdPremium, "You have used 100", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification100, "", "Premium user did not receive 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdPremium, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Premium User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_User_Gets_80_percent_Notification() {
	logger.Log.Info("Starting Test : Launch Paid instance without coupon or credit card")
	// Enroll customer

	apiKey := utils.GetMailSlurpKey()
	inboxIdPremium := utils.GetInboxIdPremium()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdPremium %s", inboxIdPremium)
	mailTime := time.Now().Add(5 * time.Minute)

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
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(1), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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

	// Now launch paid instance

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
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
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bm", "840")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	time.Sleep(3 * time.Minute)
	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	usage_err := billing.ValidateUsageNotZero(cloudAccId, float64(0.0603), "bm-spr", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : %s", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateUsageCreditsinRange(cloudAccId, float64(0.8), float64(1), authToken, float64(0.07), float64(0.2), float64(0.07), float64(0.2), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : %s", credits_err)
	computeUrl := utils.Get_Compute_Base_Url()
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

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodePremium(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	// Validate Mail Notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdPremium, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification80, "", "Premium user did not receive 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdPremium, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Premium User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdPremium, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Premium User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

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
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_User_Adds_Credits_After_80_percent_Notification() {
	logger.Log.Info("Starting Test : Launch Paid instance without coupon or credit card")
	// Enroll customer
	computeUrl := utils.Get_Compute_Base_Url()
	apiKey := utils.GetMailSlurpKey()
	inboxIdPremium := utils.GetInboxIdPremium()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdPremium %s", inboxIdPremium)
	mailTime := time.Now().Add(1 * time.Minute)

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
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(1), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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

	// Now launch paid instance

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
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
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "6420")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	usage_err := billing.ValidateUsageinRange(cloudAccId, float64(0.0075), "vm-spr-sml", authToken, float64(0.79), float64(0.90))
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : %s", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)
	credits_err := billing.ValidateUsageCreditsinRange(cloudAccId, float64(0.8), float64(1), authToken, float64(0.07), float64(0.2), float64(0.07), float64(0.2), float64(0))
	//credits_err := billing.ValidateCredits(cloudAccId, float64(1), authToken, float64(0.20), float64(-2.6), float64(0))
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

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodePremium(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	// Validate Mail Notification

	time.Sleep(20 * time.Second)
	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdPremium, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification80, "", "Premium user did not receive 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdPremium, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Premium User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdPremium, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Premium User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// User adds credits

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(5), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	mailTime = time.Now().Add(5 * time.Minute)
	time.Sleep(10 * time.Minute)

	status, instanceListed = financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	emailNotification100, err1 = financials.GetMailFromInbox(inboxIdPremium, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Premium User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

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
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_User_Adds_Credits_After_100_percent_Notification() {
	logger.Log.Info("Starting Test : Launch Paid instance without coupon or credit card")
	// Enroll customer

	apiKey := utils.GetMailSlurpKey()
	inboxIdPremium := utils.GetInboxIdPremium()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdPremium %s", inboxIdPremium)
	mailTime := time.Now().Add(1 * time.Minute)

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
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(1), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bm", "1200")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	time.Sleep(3 * time.Minute)

	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

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
			assert.Equal(suite.T(), 0.0603, rateFactor, "Failed: Failed to validate rate amount")

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
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodePremium(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	time.Sleep(3 * time.Minute)
	status, instanceListed := financials.CheckCloudAccInDeactivationList(base_url, authToken, cloudAccId)
	assert.Equal(suite.T(), status, 200, "Get Deactivation API Did not return 200 response code")
	assert.Equal(suite.T(), instanceListed, false, "Cloud Account Listed in Deactivation List")

	// Add credits again and check no new notifications are coming (TODO: Notification check)

	proxy_val := os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err := financials.GetMailFromInbox(inboxIdPremium, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Premium user received 80% notification")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 := financials.GetMailFromInbox(inboxIdPremium, "You have used 100", apiKey, mailTime, true)
	assert.NotEqual(suite.T(), emailNotification100, "", "Premium User did not receive 100% notification")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 := financials.GetMailFromInbox(inboxIdPremium, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Premium User received expire notification")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(1), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	mailTime = time.Now().Add(1 * time.Minute)
	time.Sleep(6 * time.Minute)
	// Validate cloud credit data
	zeroamt = 1
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount = gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.GreaterOrEqual(suite.T(), float64(zeroamt), totalRemainingAmount, "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodePremium(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

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

	// Validate Mail Notification

	proxy_val = os.Getenv("https_proxy")

	os.Setenv("https_proxy", "http://internal-placeholder.com:912")
	emailNotification80, err = financials.GetMailFromInbox(inboxIdPremium, "Your Cloud Credits are about to run out", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification80, "", "Premium user did not receive 80% notification ")
	assert.Equal(suite.T(), err, nil, "Error while accessing the inbox")

	emailNotification100, err1 = financials.GetMailFromInbox(inboxIdPremium, "You have used 100", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotification100, "", "Premium User received 100% notification ")
	assert.Equal(suite.T(), err1, nil, "Error while accessing the inbox")

	emailNotificationexpire, err2 = financials.GetMailFromInbox(inboxIdPremium, "Credits Expired Notification", apiKey, mailTime, true)
	assert.Equal(suite.T(), emailNotificationexpire, "", "Premium User received expire notification ")
	assert.Equal(suite.T(), err2, nil, "Error while accessing the inbox")
	// Check Cloud account not in deactivation list
	os.Setenv("https_proxy", proxy_val)
	computeUrl := utils.Get_Compute_Base_Url()
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
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Validate_Master_Plan_Creation() {
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
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(1), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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
	client_acc_id1 := gjson.Get(responseBody, "client_acct_id").String()
	master_plan_count := gjson.Get(responseBody, "master_plan_count").String()
	logger.Logf.Infof("Aria Response ", responseBody)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Details from Aria")
	assert.Equal(suite.T(), client_acc_id1, client_acct_id, "Validation failed fetching billing account details from aria, Billing Account number did not match")
	assert.Equal(suite.T(), master_plan_count, "1", "Validation failed fetching billing account details from aria")
	//Expect(strings.Contains(responseBody, `"master_plan_count" : 1`)).To(BeTrue(), "Validation failed fetching billing account details from aria")
	//Expect(strings.Contains(responseBody, `"error_msg" : "OK"`)).To(BeTrue(), "Validation failed fetching billing account details from aria")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Validate_Vm_Plan_Creation() {
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
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(2), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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

	coupon_err := billing.Create_Redeem_Coupon_With_Shrt_Expiry("Premium", int64(15), int64(2), cloudAccId, time.Duration(10))
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "240000")
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
			assert.Equal(suite.T(), planDate, todatDate, "Vm Plan  date did not match")
			assert.Equal(suite.T(), billDate, billThroughDate, "Vm Plan Bill Through date did not match")
			assert.Equal(suite.T(), todatDate, assignmentDate, "Vm Plan Assignment date did not match")
			assert.Equal(suite.T(), todatDate, planDate, "Vm Plan date did not match")
		}
		return true // keep iterating
	})

	assert.Equal(suite.T(), master_plan_count1, 1, "Master Plan id not found in response")
	assert.Equal(suite.T(), vm_plan_count, 1, "Vm  Plan id not found in response")

}

func (suite *BillingAPITestSuite) Test_Premium_Sanity_Validate_all_Usages() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(100), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
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
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "180000")
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
				assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

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
				assert.Equal(suite.T(), 0.015, rateFactor, "Failed: Failed to validate rate amount")
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
				assert.Equal(suite.T(), 0.03, rateFactor, "Failed: Failed to validate rate amount")
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
				assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

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
				assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

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
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium User)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Sanity_Validate_All_CloudAccount_Attributes() {
	logger.Log.Info("Starting Test : Launch Paid instance without coupon or credit card")
	// Enroll customer

	apiKey := utils.GetMailSlurpKey()
	inboxIdPremium := utils.GetInboxIdPremium()
	logger.Logf.Infof("Mail Slurp key %s", apiKey)
	logger.Logf.Infof("inboxIdPremium %s", inboxIdPremium)

	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	now1 := time.Now().UTC()
	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(1), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade
	time.Sleep(3 * time.Minute)
	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	initialTS := gjson.Get(responsePayload, "creditsDepleted").String()
	assert.Equal(suite.T(), initialTS, "1970-01-01T00:00:00Z", "Failed: Credit depletion time stamp not correct")

	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.Equal(suite.T(), "UPGRADE_COMPLETE", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

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
	assert.Equal(suite.T(), utils.GetCountryCodePremium(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(6), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	time.Sleep(6 * time.Minute)

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), utils.GetCountryCodePremium(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	initialTS = gjson.Get(responsePayload, "creditsDepleted").String()
	assert.Equal(suite.T(), initialTS, "1970-01-01T00:00:00Z", "Failed: Credit depletion time stamp not correct")

	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	now = time.Now().UTC()
	previousDate = now.Add(5 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload = financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "vm1", "300000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url = base_url + "/v1/meteringrecords"
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
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
	assert.Equal(suite.T(), utils.GetCountryCodePremium(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
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
	assert.Equal(suite.T(), utils.GetCountryCodePremium(), gjson.Get(responsePayload, "countryCode").String(), "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ts = gjson.Get(responsePayload, "creditsDepleted").String()
	creditDepletedTimestamp2, _ := time.Parse(time.RFC3339, ts)
	assert.Equal(suite.T(), false, creditDepletedTimestamp2.After(creditDepletedTimestamp1), "Failed: Credit depletion time stamp did not remain same")
	assert.Equal(suite.T(), creditDepletedTimestamp1, creditDepletedTimestamp2, "Failed: Credit depletion time stamp did not remain same")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")
}
