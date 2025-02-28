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

	_ "time"

	"github.com/stretchr/testify/assert"
)

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_Valid_Payload() {
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))
	response, err := billing.CreateMaasUsageRecords(cloudAccId, endTime, "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString(), authToken, 200)
	assert.Equal(suite.T(), response, "{}", "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_text_Valid_Payload() {
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))
	response, err := billing.CreateMaasUsageRecords(cloudAccId, endTime, "text", 7, "us-dev-1", startTime, timeStamp, uuid.NewString(), authToken, 200)
	assert.Equal(suite.T(), response, "{}", "Test Failed: Error Creating in MaaS Usage record")
	assert.Equal(suite.T(), err, nil, "Test Failed: Error Creating in MaaS Usage record")

}

// Tests with missing fields

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_CloudAccId_Missing() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	payload := fmt.Sprintf(`{		
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
	    }`, endTime, "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_EndTime_Missing() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))
	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",        
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, _ := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record") // Usages records always returns 200. Even if the payload is wrong. Validation is inside schedulers
	//assert.Contains(suite.T(), body, "invalid google.protobuf.Timestamp value", "Test Failed: Error Creating in MaaS Usage record")
}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_Properties_Missing() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))
	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",        
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, _ := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record") // Usages records always returns 200. Even if the payload is wrong. Validation is inside schedulers, only returns 400 for required attributes.
	//assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation", "Test Failed: Error Creating in MaaS Usage record")
}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_Region_Missing() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,      
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 7, startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation", "Test Failed: Error Creating in MaaS Usage record")
}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_StartTime_Missing() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

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
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 7, "us-dev-1", timeStamp, uuid.NewString())

	code, _ := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record") // Usages records always returns 200. Even if the payload is wrong. Validation is inside schedulers
	//assert.Contains(suite.T(), body, "invalid google.protobuf.Timestamp value", "Test Failed: Error Creating in MaaS Usage record")
}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_Timestamp_Missing() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

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
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 7, "us-dev-1", startTime, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation.", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_TransactionID_Missing() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

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
      
	    }`, cloudAccId, endTime, "image", 7, "us-dev-1", startTime, timeStamp)

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "unexpected token null", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_ServiceType_Missing() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {        
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, _ := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record") // Usages records always returns 200. Even if the payload is wrong. Validation is inside schedulers
	//assert.Equal(suite.T(), body, "invalid input arguments, ignoring product usage record creation", "Test Failed: Error Creating in MaaS Usage record")
}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_ProcessingType_Missing() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService"
        
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record") // Usages records always returns 200. Even if the payload is wrong. Validation is inside schedulers
	assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation", "Test Failed: Error Creating in MaaS Usage record")
}

// Empty Properties

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_CloudAccId_Empty() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	payload := fmt.Sprintf(`{
		"cloudAccountId": "",
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
	    }`, endTime, "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation.", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_EndTime_Empty() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "invalid google.protobuf.Timestamp value", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_Properties_Empty() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {        
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, _ := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record") // Usages records always returns 200. Even if the payload is wrong. Validation is inside schedulers
	//assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation", "Test Failed: Error Creating in MaaS Usage record")
}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_Region_Empty() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 7, startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_StartTime_Empty() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

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
      "startTime": "",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 7, "us-dev-1", timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "invalid google.protobuf.Timestamp value", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_Timestamp_Empty() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

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
      "timestamp": "",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 7, "us-dev-1", startTime, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "invalid google.protobuf.Timestamp value", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_TransactionID_Empty() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

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
      "transactionId": ""
	    }`, cloudAccId, endTime, "image", 7, "us-dev-1", startTime, timeStamp)

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_ServiceType_Empty() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_ProcessingType_Empty() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	cloudAccId := fmt.Sprintf("%012d", time.Now().UnixNano()/(1<<22))

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": ""
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, _ := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Usage Payload : %s ", payload)
	assert.Equal(suite.T(), code, 200, "Test Failed: Error Creating in MaaS Usage record")
	//assert.Contains(suite.T(), body, "invalid input arguments, ignoring product usage record creation", "Test Failed: Error Creating in MaaS Usage record")
}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_CloudAccId_With_Invalid_Data() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	payload := fmt.Sprintf(`{
		"cloudAccountId": %d,
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
	    }`, 123, endTime, "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	logger.Logf.Infof("Maas Usage Payload : ", payload)
	assert.Contains(suite.T(), body, "invalid value for string type", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_EndTime_With_Invalid_Data() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
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
	    }`, "123456789012", "test", "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	logger.Logf.Infof("Maas Usage Payload : ", payload)
	assert.Contains(suite.T(), body, "invalid google.protobuf.Timestamp value", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_Properties_With_Invalid_Data() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)
	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  "%s",
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, "123456789012", endTime, "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	logger.Logf.Infof("Maas Usage Payload : ", payload)
	assert.Contains(suite.T(), body, "proto: syntax error", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_Region_With_Invalid_Data() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": %t,
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, "123456789012", endTime, "image", 7, true, startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	logger.Logf.Infof("Maas Usage Payload : ", payload)
	assert.Contains(suite.T(), body, "invalid value for string type", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_StartTime_With_Invalid_Data() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": %d,
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, "123456789012", endTime, "image", 7, "us-dev-1", 123, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	logger.Logf.Infof("Maas Usage Payload : ", payload)
	assert.Contains(suite.T(), body, "proto: syntax error", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_Timestamp_With_Invalid_Data() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

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
	    }`, "123456789012", endTime, "image", 7, "us-dev-1", startTime, "test", uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	logger.Logf.Infof("Maas Usage Payload : ", payload)
	assert.Contains(suite.T(), body, "invalid google.protobuf.Timestamp value", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_TransactionID_With_Invalid_Data() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

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
      "transactionId": %d
	    }`, "123456789012", endTime, "image", 7, "us-dev-1", startTime, timeStamp, 123)

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	logger.Logf.Infof("Maas Usage Payload : ", payload)
	assert.Contains(suite.T(), body, "invalid value for string type", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_ServiceType_With_Invalid_Data() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": %t,
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, "123456789012", endTime, true, "image", 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	logger.Logf.Infof("Maas Usage Payload : ", payload)
	assert.Contains(suite.T(), body, "invalid value for string type", "Test Failed: Error Creating in MaaS Usage record")

}

func (suite *BillingAPITestSuite) Test_Create_MaaS_Usage_Record_image_ProcessingType_With_Invalid_Data() {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	logger.Log.Info("Starting MaaS Create Usage Record test")
	now := time.Now().UTC()
	timeStamp := now.Add(1 * time.Minute).Format(time.RFC3339)
	startTime := now.Add(10 * time.Minute).Format(time.RFC3339)
	endTime := now.Add(3 * time.Hour).Format(time.RFC3339)

	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": %d
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, "123456789012", endTime, 123, 7, "us-dev-1", startTime, timeStamp, uuid.NewString())

	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	assert.Equal(suite.T(), code, 400, "Test Failed: Error Creating in MaaS Usage record")
	logger.Logf.Infof("Maas Usage Payload : ", payload)
	assert.Contains(suite.T(), body, "invalid value for string type", "Test Failed: Error Creating in MaaS Usage record")

}
