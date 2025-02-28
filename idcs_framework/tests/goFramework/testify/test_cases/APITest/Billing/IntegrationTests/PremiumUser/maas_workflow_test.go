//go:build Functional || MaaS || PremiumIntegration
// +build Functional MaaS PremiumIntegration

package PremiumBillingAPITest

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) Test_Premium_Create_MaaS_Usage_Record_And_Search() {
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

	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "text", 7, "us-dev-1", startTime, timeStamp, transactionid1)

	maas_usage_url := base_url + "/v1/usages/records/products"
	code, body := financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("cloudAccId : %s ", cloudAccId)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	payload = fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 70, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"cloudAccountId": "%s"
	    }`, cloudAccId)

	found1 := false
	found2 := false
	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1
		assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			found1 = true
		}

		if transactionId == transactionid2 {
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 2, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, true, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_Create_MaaS_Usage_Record_Validate_Usage_For_Text() {

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

	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(5), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "text", 10, "us-dev-1", startTime, timeStamp, transactionid2)

	maas_usage_url := base_url + "/v1/usages/records/products"
	code, body := financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"cloudAccountId": "%s"
	    }`, cloudAccId)

	found2 := false
	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1
		assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid2 {
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 1, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	estimateUsage := financials_utils.GetPremiumTextRate() * float64(10)
	estimateRemaining := float64(5) - float64(estimateUsage)

	logger.Logf.Info("Text Rate for intel : ", financials_utils.GetPremiumTextRate())

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimateUsage), financials_utils.GetPremiumTextRate(), "modelasaservice-text-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(estimateUsage), authToken, float64(estimateRemaining), float64(estimateRemaining), float64(estimateRemaining))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_Create_MaaS_Usage_Record_Validate_Usage_For_Image() {
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

	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(5), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 10, "us-dev-1", startTime, timeStamp, transactionid2)

	maas_usage_url := base_url + "/v1/usages/records/products"
	code, body := financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"cloudAccountId": "%s"
	    }`, cloudAccId)

	found2 := false
	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1
		assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid2 {
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 1, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	estimateUsage := financials_utils.GetPremiumTextRate() * float64(10)
	estimateRemaining := float64(5) - float64(estimateUsage)

	logger.Logf.Info("Text Rate for intel : ", financials_utils.GetPremiumTextRate())

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimateUsage), financials_utils.GetPremiumTextRate(), "modelasaservice-image-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(estimateUsage), authToken, float64(estimateRemaining), float64(estimateRemaining), float64(estimateRemaining))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_Create_MaaS_Usage_Record_Validate_Usage_For_Image_And_Text() {
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
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 10, "us-dev-1", startTime, timeStamp, transactionid1)

	maas_usage_url := base_url + "/v1/usages/records/products"
	code, body := financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	payload = fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "text", 20, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"cloudAccountId": "%s"
	    }`, cloudAccId)

	found1 := false
	found2 := false
	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1
		assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			found1 = true
		}

		if transactionId == transactionid2 {
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 2, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, true, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")

	// Validate Usage & credit depletion

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	estimatetextUsage := financials_utils.GetPremiumTextRate() * float64(200)
	estimateimageUsage := financials_utils.GetPremiumImageRate() * float64(100)
	totalUsage := estimatetextUsage + estimateimageUsage
	estimateRemaining := float64(15) - float64(totalUsage)

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimatetextUsage), financials_utils.GetPremiumTextRate(), "modelasaservice-text-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(estimateimageUsage), financials_utils.GetPremiumImageRate(), "modelasaservice-image-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(totalUsage), authToken, float64(estimateRemaining), float64(estimateRemaining), float64(estimateRemaining))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Uses_All_Credits_And_Tries_to_Launch_instance() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()

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
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 100, "us-dev-1", startTime, timeStamp, transactionid1)

	maas_usage_url := base_url + "/v1/usages/records/products"
	code, body := financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	payload = fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "text", 200, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"cloudAccountId": "%s"
	    }`, cloudAccId)

	found1 := false
	found2 := false
	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1
		assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			found1 = true
		}

		if transactionId == transactionid2 {
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 2, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, true, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")

	// Validate Usage & credit depletion

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	estimatetextUsage := financials_utils.GetPremiumTextRate() * float64(200)
	estimateimageUsage := financials_utils.GetPremiumImageRate() * float64(100)
	totalUsage := estimatetextUsage + estimateimageUsage
	estimateRemaining := float64(15) - float64(totalUsage)

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimatetextUsage), financials_utils.GetPremiumTextRate(), "modelasaservice-text-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(estimateimageUsage), financials_utils.GetPremiumImageRate(), "modelasaservice-image-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(totalUsage), authToken, float64(estimateRemaining), float64(estimateRemaining), float64(estimateRemaining))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

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

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_User_Uses_Less_Credits_And_Tries_to_Launch_instance() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()

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
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(25), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 100, "us-dev-1", startTime, timeStamp, transactionid1)

	maas_usage_url := base_url + "/v1/usages/records/products"
	code, body := financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	payload = fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "text", 200, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"cloudAccountId": "%s"
	    }`, cloudAccId)

	found1 := false
	found2 := false
	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1
		assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			found1 = true
		}

		if transactionId == transactionid2 {
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 2, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, true, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")

	// Validate Usage & credit depletion

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	estimatetextUsage := financials_utils.GetPremiumTextRate() * float64(200)
	estimateimageUsage := financials_utils.GetPremiumImageRate() * float64(100)
	totalUsage := estimatetextUsage + estimateimageUsage
	estimateRemaining := float64(25) - float64(totalUsage)

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimatetextUsage), financials_utils.GetPremiumTextRate(), "modelasaservice-text-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(estimateimageUsage), financials_utils.GetPremiumImageRate(), "modelasaservice-image-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(totalUsage), authToken, float64(estimateRemaining), float64(estimateRemaining), float64(estimateRemaining))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

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

func (suite *BillingAPITestSuite) Test_Premium_User_Uses_Adds_Credits_After_100_percent_And_Tries_to_Launch_instance() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	computeUrl := utils.Get_Compute_Base_Url()

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
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 100, "us-dev-1", startTime, timeStamp, transactionid1)

	maas_usage_url := base_url + "/v1/usages/records/products"
	code, body := financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	payload = fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "text", 200, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"cloudAccountId": "%s"
	    }`, cloudAccId)

	found1 := false
	found2 := false
	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1
		assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			found1 = true
		}

		if transactionId == transactionid2 {
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 2, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, true, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")

	// Validate Usage & credit depletion

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	estimatetextUsage := financials_utils.GetPremiumTextRate() * float64(200)
	estimateimageUsage := financials_utils.GetPremiumImageRate() * float64(100)
	totalUsage := estimatetextUsage + estimateimageUsage
	estimateRemaining := float64(15) - float64(totalUsage)

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimatetextUsage), financials_utils.GetPremiumTextRate(), "modelasaservice-text-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(estimateimageUsage), financials_utils.GetPremiumImageRate(), "modelasaservice-image-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(totalUsage), authToken, float64(estimateRemaining), float64(estimateRemaining), float64(estimateRemaining))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	coupon_err = billing.Create_Redeem_Coupon("Premium", int64(15), int64(2), cloudAccId)
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

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}
