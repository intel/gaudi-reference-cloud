//go:build Functional || Billing || Standard || StandardIntegration || Integration || MeteringTest
// +build Functional Billing Standard StandardIntegration Integration MeteringTest

package StandardBillingAPITest

import (
	"fmt"
	_ "fmt"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/metering"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"time"

	"github.com/google/uuid"
)

func (suite *BillingAPITestSuite) Test_Create_Usage_Records_With_NonExisting_cloudAccId() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	resourceId := uuid.NewString()
	create_payload := financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		uuid.NewString(), resourceId, "123456789012", previousDate, "vm-spr-sml", "smallvm", "30000", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	time.Sleep(2 * time.Minute)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
	accId := "123456789012"
	filter := metering.SearchInvalidMeteringRecord{
		CloudAccountId: &accId,
	}
	flag, _ := metering.Get_Invalid_Metering_Record(filter, 200, 1, "FAILED_TO_GET_PRODUCT_RATE")
	assert.Equal(suite.T(), flag, true, "Test Failed : CloudAccount Id not found in invalid records")
}

func (suite *BillingAPITestSuite) Test_Create_Usage_Records_With_Empty_cloudAccId() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	resourceId := uuid.NewString()
	create_payload := financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		uuid.NewString(), resourceId, "", previousDate, "vm-spr-sml", "smallvm", "30000", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	time.Sleep(2 * time.Minute)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
	accId := ""
	filter := metering.SearchInvalidMeteringRecord{
		CloudAccountId: &accId,
	}
	flag, _ := metering.Get_Invalid_Metering_Record(filter, 200, 1, "MISSING_CLOUD_ACCOUNT_ID")
	assert.Equal(suite.T(), flag, true, "Test Failed : CloudAccount Id not found in invalid records")
}

func (suite *BillingAPITestSuite) Test_Create_Records_With_Invalid_Product() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	resourceId := uuid.NewString()
	create_payload := financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		uuid.NewString(), resourceId, cloudAccId, previousDate, "vm-spr-tny", "smallvm", "30000", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	time.Sleep(2 * time.Minute)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
	filter := metering.SearchInvalidMeteringRecord{
		CloudAccountId: &cloudAccId,
	}
	flag, _ := metering.Get_Invalid_Metering_Record(filter, 200, 1, "NO_MATCHING_PRODUCT")
	assert.Equal(suite.T(), flag, true, "Test Failed : CloudAccount Id not found in invalid records")
}

func (suite *BillingAPITestSuite) Test_Create_Records_With_Duplicate_TransactionId() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	resourceId := uuid.NewString()
	transactionId := uuid.NewString()
	create_payload := financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		transactionId, resourceId, cloudAccId, previousDate, "test-prod", "smallvm", "30000", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	time.Sleep(2 * time.Minute)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	create_payload = financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		transactionId, resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "300000", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	filter := metering.SearchInvalidMeteringRecord{
		CloudAccountId: &cloudAccId,
	}
	flag, _ := metering.Get_Invalid_Metering_Record(filter, 200, 1, "DUPLICATE_TRANSACTION_ID")
	assert.Equal(suite.T(), flag, true, "Test Failed : CloudAccount Id not found in invalid records")
}

func (suite *BillingAPITestSuite) Test_Create_Records_With_Invalid_Run_Seconds() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	resourceId := uuid.NewString()

	create_payload := financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		uuid.NewString(), resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "300", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	create_payload = financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		uuid.NewString(), resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "3000", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	create_payload = financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		uuid.NewString(), resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "30", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(2 * time.Minute)

	filter := metering.SearchInvalidMeteringRecord{
		CloudAccountId: &cloudAccId,
	}
	flag, jsonStr := metering.Get_Invalid_Metering_Record(filter, 200, 1, "INVALID_METERING_QTY")
	metering_qty := gjson.Get(jsonStr, "result.properties.runningSeconds").String()
	assert.Equal(suite.T(), metering_qty, "30", "Run Seconds did not match")
	assert.Equal(suite.T(), flag, true, "Test Failed : CloudAccount Id not found in invalid records")
}

func (suite *BillingAPITestSuite) Test_Create_Records_With_Missing_Cloud_Account_Id() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	resourceId := uuid.NewString()
	transactionId := uuid.NewString()
	create_payload := financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		transactionId, resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "30000", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	create_payload, _ = sjson.Delete(create_payload, "cloudAccountId")
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	time.Sleep(2 * time.Minute)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
	filter := metering.SearchInvalidMeteringRecord{
		TransactionId: &transactionId,
	}
	flag, _ := metering.Get_Invalid_Metering_Record(filter, 200, 1, "MISSING_CLOUD_ACCOUNT_ID")
	assert.Equal(suite.T(), flag, true, "Test Failed : CloudAccount Id not found in invalid records")
}

func (suite *BillingAPITestSuite) Test_Create_Records_With_Missing_TransactionId() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	resourceId := uuid.NewString()
	transactionId := uuid.NewString()
	create_payload := financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		transactionId, resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "30000", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	create_payload, _ = sjson.Delete(create_payload, "transactionId")
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	time.Sleep(2 * time.Minute)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
	filter := metering.SearchInvalidMeteringRecord{
		CloudAccountId: &cloudAccId,
	}
	flag, _ := metering.Get_Invalid_Metering_Record(filter, 200, 1, "MISSING_TRANSACTION_ID")
	assert.Equal(suite.T(), flag, true, "Test Failed : CloudAccount Id not found in invalid records")
}

func (suite *BillingAPITestSuite) Test_Create_Records_With_Missing_RespurceId() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	resourceId := uuid.NewString()
	transactionId := uuid.NewString()
	create_payload := financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		transactionId, resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "30000", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	create_payload, _ = sjson.Delete(create_payload, "resourceId")
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	time.Sleep(2 * time.Minute)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
	filter := metering.SearchInvalidMeteringRecord{
		CloudAccountId: &cloudAccId,
	}
	flag, _ := metering.Get_Invalid_Metering_Record(filter, 200, 1, "MISSING_RESOURCE_NAME")
	assert.Equal(suite.T(), flag, true, "Test Failed : CloudAccount Id not found in invalid records")
}

func (suite *BillingAPITestSuite) Test_Create_Records_With_Invalid_run_seconds() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	resourceId := uuid.NewString()
	transactionId := uuid.NewString()
	create_payload := financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
		transactionId, resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "test", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
	fmt.Println("create_payload", create_payload)
	create_payload, _ = sjson.Delete(create_payload, "resourceId")
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	time.Sleep(2 * time.Minute)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
	filter := metering.SearchInvalidMeteringRecord{
		CloudAccountId: &cloudAccId,
	}
	flag, _ := metering.Get_Invalid_Metering_Record(filter, 200, 1, "FAILED_TO_CALCULATE_QTY")
	assert.Equal(suite.T(), flag, true, "Test Failed : CloudAccount Id not found in invalid records")
}
