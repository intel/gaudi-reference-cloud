//go:build Functional || Billing || Standard || StandardIntegration || Integration || MockUsage
// +build Functional Billing Standard StandardIntegration Integration MockUsage

package StandardBillingAPITest

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
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) Test_Standard_Usage_Free_VM() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-tny", "tinyvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsage(cloudAccId, float64(0), float64(0), "vm-spr-tny", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(0), authToken, float64(15), float64(15), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Usage_Small_VM() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	//zeroamt := 0
	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsage(cloudAccId, float64(3.75), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(4), authToken, float64(11.25), float64(11.25), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
}

func (suite *BillingAPITestSuite) Test_Standard_Usage_Current_Month() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_err := billing.ValidateUsage(cloudAccId, float64(30.15), float64(0.0603), "bm-spr", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	//Validate credits

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(15), authToken, float64(0), float64(0), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) Test_Standard_Usage_Previous_Month() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	//zeroamt := 0
	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.AddDate(0, 0, -35).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("previousDate", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

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
		if gjson.Get(data, "productType").String() == "vm-spr-sml" {
			flag = true
		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.Equal(suite.T(), float64(0), total_amount_from_response, "Failed: Failed to validate total amount")

	assert.Equal(suite.T(), false, flag, "Failed: Usage details visible for vm")

	total_usage_from_response := gjson.Get(usage_response_body, "totalUsaget").Float()
	assert.Equal(suite.T(), float64(0), total_usage_from_response, "Failed: Failed to validate total usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	amt := 3.75
	assert.GreaterOrEqual(suite.T(), usedAmount, float64(amt), "Failed to validate used credits")
	assert.Less(suite.T(), usedAmount, float64(4.15), "Failed to validate used credits")
	assert.Equal(suite.T(), remainingAmount, "11.25", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), unappliedCredits, "11.25", "Failed : Unapplied cloud credit did not become zero")

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
}

func (suite *BillingAPITestSuite) Test_Standard_Usage_Date_Range() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Create and redeem normal coupon

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.AddDate(0, 0, -35).Format("2006-01-02") + "T00:00:01Z"
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "mediumvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
	endDate1 := now.AddDate(0, 0, -30).Format("2006-01-02") + "T00:00:01Z"
	endDate2 := now.AddDate(0, 0, -25).Format("2006-01-02") + "T00:00:01Z"

	now = time.Now().UTC()
	prevDate := now.Add(3 * time.Minute)
	date1 := prevDate.Format("2006-01-02") + "T00:00:01Z"
	create_payload = financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, date1, "vm-spr-sml", "smlvm", "30000")
	fmt.Println("create_payload", create_payload)
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	//time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	time.Sleep(10 * time.Minute)

	usage_err := billing.ValidateUsageDateRange(cloudAccId, previousDate, endDate1, float64(30.15), float64(0.0603), "bm-spr", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsageDateRange(cloudAccId, endDate1, endDate2, float64(0), float64(0), "bm-spr", authToken)
	assert.NotEqual(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(0), float64(0.0603), "bm-spr", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(3.75), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
}

func (suite *BillingAPITestSuite) Test_Enroll_Standard_Customer_Check_Usage() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Get usage

	ret, usageData := billing.GetUsage(cloudAccId, 200)
	assert.Equal(suite.T(), ret, true, "Failed to fetch usage for intel account")
	totalAmt := gjson.Get(usageData, "totalAmount").String()
	totalUsage := gjson.Get(usageData, "totalUsage").String()
	assert.Equal(suite.T(), "0", totalAmt, "Total amount is not equal to 0 after enrollment")
	assert.Equal(suite.T(), "0", totalUsage, "Total Usage is not equal to 0 after enrollment")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) Test_Enroll_Standard_Customer_Check_Usage_History() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Get usage
	//startTime, endTime := billing.GetCreationExpirationTime()
	ret, usageData := billing.GetUsageHistory(cloudAccId, "2023-06-19T19:41:41.851Z", "2023-06-19T19:41:41.851Z", 200)
	assert.Equal(suite.T(), ret, true, "Failed to fetch usage for intel account")
	totalAmt := gjson.Get(usageData, "totalAmount").String()
	totalUsage := gjson.Get(usageData, "totalUsage").String()
	assert.Equal(suite.T(), "0", totalAmt, "Total amount is not equal to 0 after enrollment")
	assert.Equal(suite.T(), "0", totalUsage, "Total Usage is not equal to 0 after enrollment")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) Test_Enroll_Standard_Customer_Check_Usage_History_Start_Date() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	// Get usage
	//startTime, endTime := billing.GetCreationExpirationTime()
	ret, _ := billing.GetUsageHistory(cloudAccId, "2023-06-19T19:41:41.851Z", "", 400)
	assert.NotEqual(suite.T(), ret, false, "Failed to fetch usage for intel account")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) Test_Enroll_Standard_Customer_Check_Usage_History_EndDate() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Get usage
	//startTime, endTime := billing.GetCreationExpirationTime()
	ret, _ := billing.GetUsageHistory(cloudAccId, "", "2023-06-19T19:41:41.851Z", 400)
	assert.NotEqual(suite.T(), ret, false, "Failed to fetch usage for intel account")
	os.Unsetenv("intel_user_test")
}

func (suite *BillingAPITestSuite) Test_Standard_Usage_BM() {
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)
	// Add coupon to the user

	coupon_err := billing.Create_Redeem_Coupon("Standard", int64(15), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : %s", coupon_err)

	//zeroamt := 0
	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(30 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "bm-spr", "bmvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if gjson.Get(data, "productType").String() == "bm-spr" {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			assert.GreaterOrEqual(suite.T(), actualAMount, 30.149999618530273, "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), 0.0603, rateFactor, "Failed: Failed to validate rate amount")

			logger.Logf.Infof("Actual Usage :    ", actualAMount)
			assert.GreaterOrEqual(suite.T(), actualAMount, 30.149999618530273, "Failed: Failed to validate usage amount")
		}
		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.GreaterOrEqual(suite.T(), total_amount_from_response, 30.149999618530273, "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl := utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	amt := 15
	assert.Equal(suite.T(), usedAmount, float64(amt), "Failed to validate used credits")
	assert.Equal(suite.T(), remainingAmount, "0", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	unappliedCredits = testsetup.RoundFloat(unappliedCredits, 0)
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Equal(suite.T(), unappliedCredits, float64(0), "Failed : Unapplied cloud credit did not become zero")

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")
}
