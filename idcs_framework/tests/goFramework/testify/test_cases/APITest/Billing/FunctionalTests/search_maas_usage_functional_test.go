//go:build Functional || MaaS || Regression
// +build Functional MaaS Regression

package BillingAPITest

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/service_api/financials"
	"goFramework/utils"
	_ "log"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"

	//"golang.org/x/exp/rand"

	_ "time"

	"github.com/stretchr/testify/assert"
)

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_CloudAccId() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

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
	    }`, cloudAccId, endTime, "image", 7, "us-dev-1", startTime, timeStamp, transactionid1)

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
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
	    }`, cloudAccId, endTime, "text", 70, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(base_url, authToken, payload)
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

}

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_Region() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

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
	    }`, cloudAccId, endTime, "image", 7, "us-dev-2", startTime, timeStamp, transactionid1)

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
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
	    }`, cloudAccId, endTime, "text", 70, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"region": "%s"      
	    }`, "us-dev-1")

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
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			count = count + 1
			found1 = true
		}

		if transactionId == transactionid2 {
			assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
			count = count + 1
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 1, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, false, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")

}

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_EndTime() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime1 := now.Add(3 * time.Hour).Format(time.RFC3339)
	endTime2 := now.Add(4 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

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
	    }`, cloudAccId, endTime1, "image", 7, "us-dev-2", startTime, timeStamp, transactionid1)

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
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
	    }`, cloudAccId, endTime2, "text", 70, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"endTime": "%s"      
	    }`, endTime1)

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
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			count = count + 1
			found1 = true
		}

		if transactionId == transactionid2 {
			assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
			count = count + 1
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 1, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, true, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, false, "Test Failed: Search failed in finding transaction id2")
}

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_StartTime() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime1 := now.Add(10 * time.Minute).Format(time.RFC3339)
	startTime2 := now.Add(40 * time.Minute).Format(time.RFC3339)
	endTime1 := now.Add(3 * time.Hour).Format(time.RFC3339)
	endTime2 := now.Add(4 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

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
	    }`, cloudAccId, endTime1, "image", 7, "us-dev-2", startTime1, timeStamp, transactionid1)

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
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
	    }`, cloudAccId, endTime2, "text", 70, "us-dev-1", startTime2, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"startTime": "%s"      
	    }`, startTime2)

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
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			count = count + 1
			found1 = true
		}

		if transactionId == transactionid2 {
			assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
			count = count + 1
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 1, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, false, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")

}

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_Transactionid() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime1 := now.Add(10 * time.Minute).Format(time.RFC3339)
	startTime2 := now.Add(40 * time.Minute).Format(time.RFC3339)
	endTime1 := now.Add(3 * time.Hour).Format(time.RFC3339)
	endTime2 := now.Add(4 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

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
	    }`, cloudAccId, endTime1, "image", 7, "us-dev-2", startTime1, timeStamp, transactionid1)

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
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
	    }`, cloudAccId, endTime2, "text", 70, "us-dev-1", startTime2, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"transactionId": "%s"      
	    }`, transactionid2)

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
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			count = count + 1
			found1 = true
		}

		if transactionId == transactionid2 {
			assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
			count = count + 1
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 1, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, false, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")

}

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_CloudAccId_Region() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

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
	    }`, cloudAccId, endTime, "image", 7, "us-dev-2", startTime, timeStamp, transactionid1)

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
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
	    }`, cloudAccId, endTime, "text", 70, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"region": "%s",
		"cloudAccountId": "%s"      
	    }`, "us-dev-1", cloudAccId)

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
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			count = count + 1
			found1 = true
		}

		if transactionId == transactionid2 {
			assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
			count = count + 1
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 1, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, false, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")

}

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_CloudAccId_Region_TransactionId() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	transactionid1 := uuid.NewString()
	transactionid2 := uuid.NewString()
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

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
	    }`, cloudAccId, endTime, "image", 7, "us-dev-2", startTime, timeStamp, transactionid1)

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
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
	    }`, cloudAccId, endTime, "text", 70, "us-dev-1", startTime, timeStamp, transactionid2)

	code, body = financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), body, "{}", "Test Failed: Error Creating in MaaS Usage record")

	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"region": "%s",
		"cloudAccountId": "%s",
		"transactionId": "%s"     
	    }`, "us-dev-1", cloudAccId, transactionid2)

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
		transactionId := v.Get("transactionId").String()
		if transactionId == transactionid1 {
			count = count + 1
			found1 = true
		}

		if transactionId == transactionid2 {
			assert.Equal(suite.T(), cloudAccId, v.Get("cloudAccountId").String(), "Test Failed: Cloud Account Id did not match in search result")
			count = count + 1
			found2 = true
		}

	}

	assert.Equal(suite.T(), count, 1, "Test Failed: Number of records did not match in search")
	assert.Equal(suite.T(), found1, false, "Test Failed: Search failed in finding transaction id1")
	assert.Equal(suite.T(), found2, true, "Test Failed: Search failed in finding transaction id2")

}

// Search with wrong data

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_Invalid_CloudAccId() {
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	// Search with cloud account id

	searchpayload := fmt.Sprintf(`{
		"cloudAccountId": "%s"      
	    }`, "999999999999")

	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1

	}

	assert.Equal(suite.T(), count, 0, "Test Failed: Number of records did not match in search")

}

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_Invalid_TransActionId() {
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")

	searchpayload := fmt.Sprintf(`{
		"transactionId": "%s"      
	    }`, "999999999999")

	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1

	}

	assert.Equal(suite.T(), count, 0, "Test Failed: Number of records did not match in search")

}

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_Invalid_Region() {
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	searchpayload := fmt.Sprintf(`{
		"region": "%s"      
	    }`, "test-region")

	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1

	}

	assert.Equal(suite.T(), count, 0, "Test Failed: Number of records did not match in search")

}

func (suite *BillingAPITestSuite) Test_Search_MaaS_Usage_Record_image_Invalid_Field() {
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	searchpayload := fmt.Sprintf(`{
		"test": "%s"      
	    }`, "test-field")

	logger.Logf.Infof("Maas Search Payload : %s", searchpayload)
	response, err := billing.SearchMaasUsageRecords(searchpayload, authToken, 200)
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")
	result := gjson.Parse(response)
	arr := gjson.Get(result.String(), "..#.result")
	count := 0
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1

	}

	assert.Equal(suite.T(), count, 16, "Test Failed: Number of records did not match in search")

}
