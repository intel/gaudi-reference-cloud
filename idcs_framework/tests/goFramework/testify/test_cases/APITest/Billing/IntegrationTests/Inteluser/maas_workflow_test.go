//go:build Functional || MaaS || IntelIntegration
// +build Functional MaaS IntelIntegration

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
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) Test_Intel_Create_MaaS_Usage_Record_And_Search_Product1() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	//computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"productName": "maas-model-mistral-7b-v0.1",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-mistral-7b-v0.1",
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

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Create_MaaS_Usage_Record_And_Search_Product_all_Text_Products() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	//computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()
	transactionid3 := uuid.NewString()

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"productName": "maas-model-mistral-7b-v0.1",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-mistral-7b-v0.1",
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
		"productName": "maas-model-llama-3.1-70b",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-llama-3.1-70b",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "text", 70, "us-dev-1", startTime, timeStamp, transactionid3)

	code, body = financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	payload = fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"productName": "maas-model-llama-3.1-8b",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-llama-3.1-8b",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "text", 70, "us-dev-1", startTime, timeStamp, transactionid2)

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
	found3 := false
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

		if transactionId == transactionid3 {
			found3 = true
		}

	}

	assert.Equal(suite.T(), count, 3, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, true, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")
	assert.Equal(suite.T(), found3, true, "Test Failed: Search failed in finding transaction id3")

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Create_MaaS_Usage_Record_Validate_Usage_For_Text() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	//computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(5), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"productName": "maas-model-llama-3.1-8b",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-llama-3.1-8b",
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

	estimateUsage := financials_utils.GetIntelTextRate() * float64(10)
	estimateRemaining := float64(5) - float64(estimateUsage)

	logger.Logf.Info("Text Rate for intel : ", financials_utils.GetIntelTextRate())

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimateUsage), financials_utils.GetIntelTextRate(), "maas-model-llama-3.1-8b", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(estimateUsage), authToken, float64(estimateRemaining), float64(estimateRemaining), float64(estimateRemaining))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Create_MaaS_Usage_Record_Validate_Usage_For_Image() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	//computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(5), int64(2), cloudAccId)
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

	estimateUsage := financials_utils.GetIntelTextRate() * float64(10)
	estimateRemaining := float64(5) - float64(estimateUsage)

	logger.Logf.Info("Text Rate for intel : ", financials_utils.GetIntelTextRate())

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimateUsage), financials_utils.GetIntelTextRate(), "modelasaservice-image-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(estimateUsage), authToken, float64(estimateRemaining), float64(estimateRemaining), float64(estimateRemaining))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_Create_MaaS_Usage_Record_Validate_Usage_For_Image_And_Text() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()

	//computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(15), int64(2), cloudAccId)
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
		"productName": "maas-model-llama-3.1-70b",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-llama-3.1-70b",
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

	estimatetextUsage := financials_utils.GetIntelTextRate() * float64(200)
	estimateimageUsage := financials_utils.GetIntelImageRate() * float64(100)
	totalUsage := estimatetextUsage + estimateimageUsage
	estimateRemaining := float64(15) - float64(totalUsage)

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimatetextUsage), financials_utils.GetIntelTextRate(), "maas-model-llama-3.1-70b", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(estimateimageUsage), financials_utils.GetIntelImageRate(), "modelasaservice-image-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(totalUsage), authToken, float64(estimateRemaining), float64(estimateRemaining), float64(estimateRemaining))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_User_Uses_All_Credits_And_Tries_to_Launch_instance() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(15), int64(2), cloudAccId)
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
		"productName": "maas-model-mistral-7b-v0.1",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-mistral-7b-v0.1",
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

	estimatetextUsage := financials_utils.GetIntelTextRate() * float64(200)
	estimateimageUsage := financials_utils.GetIntelImageRate() * float64(100)
	totalUsage := estimatetextUsage + estimateimageUsage
	estimateRemaining := float64(15) - float64(totalUsage)

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimatetextUsage), financials_utils.GetIntelTextRate(), "maas-model-mistral-7b-v0.1", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(estimateimageUsage), financials_utils.GetIntelImageRate(), "modelasaservice-image-processing", authToken)
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

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_User_Uses_Less_Credits_And_Tries_to_Launch_instance() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(25), int64(2), cloudAccId)
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
		"productName": "maas-model-mistral-7b-v0.1",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-mistral-7b-v0.1",
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

	estimatetextUsage := financials_utils.GetIntelTextRate() * float64(200)
	estimateimageUsage := financials_utils.GetIntelImageRate() * float64(100)
	totalUsage := estimatetextUsage + estimateimageUsage
	estimateRemaining := float64(25) - float64(totalUsage)

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimatetextUsage), financials_utils.GetIntelTextRate(), "maas-model-mistral-7b-v0.1", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(estimateimageUsage), financials_utils.GetIntelImageRate(), "modelasaservice-image-processing", authToken)
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

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Intel_User_Uses_Adds_Credits_After_100_percent_And_Tries_to_Launch_instance() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(15), int64(2), cloudAccId)
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
	    }`, cloudAccId, endTime, "image", 200, "us-dev-1", startTime, timeStamp, transactionid1)

	maas_usage_url := base_url + "/v1/usages/records/products"
	code, body := financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	payload = fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"productName": "maas-model-mistral-7b-v0.1",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-mistral-7b-v0.1",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "text", 300, "us-dev-1", startTime, timeStamp, transactionid2)

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

	estimatetextUsage := financials_utils.GetIntelTextRate() * float64(300)
	estimateimageUsage := financials_utils.GetIntelImageRate() * float64(200)
	// totalUsage := estimatetextUsage + estimateimageUsage
	// estimateRemaining := float64(15) - float64(totalUsage)

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimatetextUsage), financials_utils.GetIntelTextRate(), "maas-model-mistral-7b-v0.1", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(estimateimageUsage), financials_utils.GetIntelImageRate(), "modelasaservice-image-processing", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	coupon_err = billing.Create_Redeem_Coupon("Intel", int64(15), int64(2), cloudAccId)
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

func (suite *BillingAPITestSuite) TestIntelUsageCurrentMonth() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	//computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid2 := uuid.NewString()

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(5), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"productName": "maas-model-llama-3.1-8b",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-llama-3.1-8b",
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

	estimateUsage := financials_utils.GetIntelTextRate() * float64(10)
	estimateRemaining := float64(5) - float64(estimateUsage)

	logger.Logf.Info("Text Rate for intel : ", financials_utils.GetIntelTextRate())

	usage_err := billing.ValidateUsage(cloudAccId, float64(estimateUsage), financials_utils.GetIntelTextRate(), "maas-model-llama-3.1-8b", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(estimateUsage), authToken, float64(estimateRemaining), float64(estimateRemaining), float64(estimateRemaining))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) TestIntelUsagePreviousMonth() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	maas_usage_url := base_url + "/v1/usages/records/products"
	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	//zeroamt := 0
	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	timeStamp := now.AddDate(0, 0, -36).Format(time.RFC3339)
	startTime := now.AddDate(0, 0, -36).Format(time.RFC3339)
	endTime := now.AddDate(0, 0, -35).Format(time.RFC3339)
	transactionid2 := uuid.NewString()

	create_payload := fmt.Sprintf(`{
			"cloudAccountId": "%s",
			"productName": "maas-model-llama-3.1-8b",
			"endTime": "%s",
			"properties":  {
			"serviceType": "ModelAsAService",
			"modelType": "maas-model-llama-3.1-8b",
			"processingType": "%s"
			},
		  "quantity": %d,
		  "region": "%s",
		  "startTime": "%s",
		  "timestamp": "%s",
		  "transactionId": "%s"
			}`, cloudAccId, endTime, "text", 100, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body := financials.CreateMaasUsageRecords(maas_usage_url, authToken, create_payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", create_payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	flag := false
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if gjson.Get(data, "productType").String() == "maas-model-llama-3.1-8b" {
			flag = true
		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.Equal(suite.T(), float64(0), total_amount_from_response, "Failed: Failed to validate total amount")

	assert.Equal(suite.T(), false, flag, "Failed: Usage details visible for MaSS maas-model-llama-3.1-8b")

	total_usage_from_response := gjson.Get(usage_response_body, "totalUsage").Float()
	assert.Equal(suite.T(), float64(0), total_usage_from_response, "Failed: Failed to validate total usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	assert.Equal(suite.T(), usedAmount, float64(5), "Failed to validate used credits")
	assert.Equal(suite.T(), remainingAmount, "10", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), unappliedCredits, "10", "Failed : Unapplied cloud credit did not become zero")

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")
}

func (suite *BillingAPITestSuite) TestIntelUsageDateRange() {
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	maas_usage_url := base_url + "/v1/usages/records/products"
	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Intel", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	//zeroamt := 0
	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	timeStamp := now.AddDate(0, 0, -36).Format(time.RFC3339)
	startTime := now.AddDate(0, 0, -36).Format(time.RFC3339)
	endTime := now.AddDate(0, 0, -35).Format(time.RFC3339)
	transactionid2 := uuid.NewString()

	create_payload := fmt.Sprintf(`{
			"cloudAccountId": "%s",
			"productName": "maas-model-mistral-7b-v0.1",
			"endTime": "%s",
			"properties":  {
			"serviceType": "ModelAsAService",
			"modelType": "maas-model-mistral-7b-v0.1",
			"processingType": "%s"
			},
		  "quantity": %d,
		  "region": "%s",
		  "startTime": "%s",
		  "timestamp": "%s",
		  "transactionId": "%s"
			}`, cloudAccId, endTime, "text", 200, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body := financials.CreateMaasUsageRecords(maas_usage_url, authToken, create_payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", create_payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	endDate1 := now.AddDate(0, 0, -30).Format("2006-01-02") + "T00:00:01Z"
	endDate2 := now.AddDate(0, 0, -25).Format("2006-01-02") + "T00:00:01Z"

	now = time.Now().UTC()
	timeStamp1 := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime1 := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime1 := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid3 := uuid.NewString()

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"productName": "maas-model-llama-3.1-8b",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
		"modelType": "maas-model-llama-3.1-8b",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime1, "text", 100, "us-dev-1", startTime1, timeStamp1, transactionid3)

	code, body = financials.CreateMaasUsageRecords(maas_usage_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	now = time.Now().UTC()
	prevDate := now.Add(3 * time.Minute)
	date1 := prevDate.Format("2006-01-02") + "T00:00:01Z"
	create_payload = financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, date1, "vm-spr-sml", "smlvm", "75000")
	fmt.Println("create_payload", create_payload)

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	//time.Sleep(10 * time.Minute)

	usage_err := billing.ValidateUsageDateRange(cloudAccId, startTime, endDate1, float64(10), financials_utils.GetIntelTextRate(), "maas-model-mistral-7b-v0.1", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsageDateRange(cloudAccId, endDate1, endDate2, float64(0), float64(0), "maas-model-mistral-7b-v0.1", authToken)
	assert.NotEqual(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsageDateRange(cloudAccId, endDate1, endDate2, float64(0), float64(0), "maas-model-llama-3.1-8b", authToken)
	assert.NotEqual(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(0), float64(0), "maas-model-mistral-7b-v0.1", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(5), financials_utils.GetIntelTextRate(), "maas-model-llama-3.1-8b", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")
}
