//go:build Functional || Billing || PremiumIntegration || MeteringTest
// +build Functional Billing PremiumIntegration MeteringTest

package PremiumBillingAPITest

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/library/financials/metering"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) Test_Premium_Customer_Push_Invalid_Metering_Records_Validate_Usage() {
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

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	resourceId := uuid.NewString()
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	usage_err := billing.ValidateUsage(cloudAccId, float64(3.75), float64(0.0075), "vm-spr-sml", userToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err := billing.ValidateCredits(cloudAccId, float64(4), userToken, float64(11.25), float64(11.25), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Now push lesser usage

	previousDate = now.Add(10 * time.Hour).Format("2006-01-02T15:04:05.999999Z")

	transactionId := uuid.NewString()
	create_payload = financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		transactionId, resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "3000")
	fmt.Println("create_payload", create_payload)
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	search_payload := financials_utils.EnrichMmeteringSearchPayload(compute_utils.GetMeteringSearchPayload(), resourceId, cloudAccId)
	fmt.Println("search_payload", search_payload)
	response_status, response_body := financials.SearchAllMeteringRecords(metering_api_base_url, authToken, search_payload)

	result := gjson.Parse(response_body)
	arr := gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		if v.Get("transactionId").String() == transactionId {
			assert.Equal(suite.T(), v.Get("reported").String(), "true", "Reported not set to true for invalid record")
		}
		break
	}

	// Validate usage and credit data is not changed

	usage_err = billing.ValidateUsage(cloudAccId, float64(3.75), float64(0.0075), "vm-spr-sml", userToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(4), userToken, float64(11.25), float64(11.25), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	// Now push new metering data with same resourceId

	create_payload = financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), resourceId, cloudAccId, previousDate, "vm-spr-sml", "smallvm", "42000")
	fmt.Println("create_payload", create_payload)
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	usage_err = billing.ValidateUsage(cloudAccId, float64(3.75), float64(0.0075), "vm-spr-sml", userToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	time.Sleep(2 * time.Minute)

	credits_err = billing.ValidateCredits(cloudAccId, float64(4), userToken, float64(11.25), float64(11.25), float64(0))
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) Test_Premium_Customer_Push_Metering_Records_Validate_Search() {
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

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.AddDate(0, 0, -5).Format("2006-01-02") + "T00:00:01Z"
	fmt.Println("Metering Date", previousDate)
	medResourceId := uuid.NewString()
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), medResourceId, cloudAccId, previousDate, "vm-spr-med", "mediumvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
	endDate1 := now.AddDate(0, 0, -2).Format("2006-01-02") + "T00:00:01Z"
	endDate2 := now.AddDate(0, 0, -1).Format("2006-01-02") + "T00:00:01Z"
	endDate3 := now.AddDate(0, 0, -6).Format("2006-01-02") + "T00:00:01Z"

	now = time.Now().UTC()
	prevDate := now.Add(3 * time.Minute)
	date1 := prevDate.Format("2006-01-02") + "T00:00:01Z"
	date2 := prevDate.Format("2006-01-02") + "T23:00:01Z"
	smlResourceId := uuid.NewString()
	create_payload = financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), smlResourceId, cloudAccId, date1, "vm-spr-sml", "smlvm", "30000")
	fmt.Println("create_payload", create_payload)
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	filter := metering.UsageFilter{
		StartTime:      &previousDate,
		EndTime:        &endDate1,
		ResourceId:     &medResourceId,
		CloudAccountId: &cloudAccId,
	}
	resBody, responseCode := metering.Search_Metering_Records(filter, 200)
	assert.Equal(suite.T(), responseCode, 200, "Failed: Failed to create metering for medium vm data")
	result := gjson.Parse(resBody)
	logger.Logf.Info("Result of search", result)
	medcount := 0
	smlcount := 0
	arr := gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		if cloudAccId == v.Get("cloudAccountId").String() && medResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				medcount = medcount + 1
			}
		}

		if cloudAccId == v.Get("cloudAccountId").String() && smlResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				smlcount = smlcount + 1
			}
		}

	}

	assert.Equal(suite.T(), medcount, 1, "Failed: Failed to create metering for medium vm data")
	assert.Equal(suite.T(), smlcount, 0, "Failed: Failed to create metering for medium vm data")

	filter = metering.UsageFilter{
		StartTime:      &endDate1,
		EndTime:        &endDate2,
		ResourceId:     &smlResourceId,
		CloudAccountId: &cloudAccId,
	}
	resBody, responseCode = metering.Search_Metering_Records(filter, 200)

	result = gjson.Parse(resBody)
	logger.Logf.Info("Result of search", result)
	medcount = 0
	smlcount = 0
	arr = gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		if cloudAccId == v.Get("cloudAccountId").String() && medResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				medcount = medcount + 1
			}
		}

		if cloudAccId == v.Get("cloudAccountId").String() && smlResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				smlcount = smlcount + 1
			}
		}

	}

	assert.Equal(suite.T(), smlcount, 0, "Failed: Failed to create metering small data")
	assert.Equal(suite.T(), medcount, 0, "Failed: Failed to create metering for medium vm data")

	filter = metering.UsageFilter{
		StartTime:      &endDate2,
		EndTime:        &date1,
		ResourceId:     &smlResourceId,
		CloudAccountId: &cloudAccId,
	}
	resBody, responseCode = metering.Search_Metering_Records(filter, 200)

	result = gjson.Parse(resBody)
	logger.Logf.Info("Result of search", result)
	medcount = 0
	smlcount = 0
	arr = gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		if cloudAccId == v.Get("cloudAccountId").String() && medResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				medcount = medcount + 1
			}
		}

		if cloudAccId == v.Get("cloudAccountId").String() && smlResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				smlcount = smlcount + 1
			}
		}

	}

	assert.Equal(suite.T(), smlcount, 1, "Failed: Failed to create metering small data")
	assert.Equal(suite.T(), medcount, 0, "Failed: Failed to create metering for medium vm data")

	filter = metering.UsageFilter{
		StartTime:      &endDate3,
		EndTime:        &date2,
		ResourceId:     &smlResourceId,
		CloudAccountId: &cloudAccId,
	}
	resBody, responseCode = metering.Search_Metering_Records(filter, 200)

	result = gjson.Parse(resBody)
	logger.Logf.Info("Result of search", result)
	medcount = 0
	smlcount = 0
	arr = gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		if cloudAccId == v.Get("cloudAccountId").String() && medResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				medcount = medcount + 1
			}
		}

		if cloudAccId == v.Get("cloudAccountId").String() && smlResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				smlcount = smlcount + 1
			}
		}

	}

	filter = metering.UsageFilter{
		StartTime:      &endDate3,
		EndTime:        &date2,
		ResourceId:     &medResourceId,
		CloudAccountId: &cloudAccId,
	}
	resBody, responseCode = metering.Search_Metering_Records(filter, 200)

	result = gjson.Parse(resBody)
	logger.Logf.Info("Result of search", result)
	medcount = 0
	smlcount = 0
	arr = gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		if cloudAccId == v.Get("cloudAccountId").String() && medResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				medcount = medcount + 1
			}
		}

		if cloudAccId == v.Get("cloudAccountId").String() && smlResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				smlcount = smlcount + 1
			}
		}

	}

	assert.Equal(suite.T(), smlcount, 0, "Failed: Failed to create metering small data")
	assert.Equal(suite.T(), medcount, 1, "Failed: Failed to create metering for medium vm data")

	filter = metering.UsageFilter{
		StartTime:      &endDate3,
		EndTime:        &date2,
		CloudAccountId: &cloudAccId,
	}
	resBody, responseCode = metering.Search_Metering_Records(filter, 200)

	result = gjson.Parse(resBody)
	logger.Logf.Info("Result of search", result)
	medcount = 0
	smlcount = 0
	arr = gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		if cloudAccId == v.Get("cloudAccountId").String() && medResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				medcount = medcount + 1
			}
		}

		if cloudAccId == v.Get("cloudAccountId").String() && smlResourceId == v.Get("resourceId").String() {
			if v.Get("properties.runningSeconds").String() == "30000" {
				logger.Logf.Info("Found Search Result ", v)
				smlcount = smlcount + 1
			}
		}

	}

	assert.Equal(suite.T(), smlcount, 1, "Failed: Failed to create metering small data")
	assert.Equal(suite.T(), medcount, 1, "Failed: Failed to create metering for medium vm data")

	usage_err := billing.ValidateUsageDateRange(cloudAccId, previousDate, endDate1, float64(7.5), float64(0.015), "vm-spr-med", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsageDateRange(cloudAccId, endDate1, endDate2, float64(0), float64(0), "vm-spr-med", authToken)
	assert.NotEqual(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(7.5), float64(0.015), "vm-spr-med", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	usage_err = billing.ValidateUsage(cloudAccId, float64(3.75), float64(0.0075), "vm-spr-sml", authToken)
	assert.Equal(suite.T(), usage_err, nil, "Failed to validate usage, validation failed with error : ", usage_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")
}

func (suite *BillingAPITestSuite) Test_Premium_Customer_Push_Metering_Records_Different_Dates() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	//computeUrl := utils.Get_Compute_Base_Url()
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
	previousDate := now.AddDate(0, 0, -5).Format("2006-01-02") + "T00:00:01Z"
	fmt.Println("Metering Date", previousDate)
	medResourceId := uuid.NewString()
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), medResourceId, cloudAccId, previousDate, "vm-spr-med", "mediumvm", "30000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	now = time.Now().UTC()
	prevDate := now.Add(3 * time.Minute)
	date1 := prevDate.Format("2006-01-02") + "T00:00:01Z"

	create_payload = financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), medResourceId, cloudAccId, date1, "vm-spr-med", "mediumvm", "300000")
	fmt.Println("create_payload", create_payload)

	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// create metering data for small vm
	smlResourceId := uuid.NewString()
	create_payload = financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), smlResourceId, cloudAccId, date1, "vm-spr-sml", "smlvm", "30000")
	fmt.Println("create_payload", create_payload)
	response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)
	smlcount := 0
	medcount := 0

	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if gjson.Get(data, "productType").String() == "vm-spr-med" {
			medcount = 1
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			assert.Equal(suite.T(), actualAMount, float64(75), "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), 0.015, rateFactor, "Failed: Failed to validate rate amount")

			start := gjson.Get(data, "start").String()
			assert.Equal(suite.T(), previousDate, start, "Failed: Failed to validate start date")

			end := gjson.Get(data, "end").String()
			assert.Equal(suite.T(), date1, end, "Failed: Failed to validate end date")

			logger.Logf.Infof("Actual Usage : ", actualAMount)
			assert.Equal(suite.T(), actualAMount, float64(75), "Failed: Failed to validate usage amount")
		}

		if gjson.Get(data, "productType").String() == "vm-spr-sml" {
			smlcount = 1
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			assert.Equal(suite.T(), actualAMount, float64(3.75), "Failed: Failed to validate usage amount")

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

			start := gjson.Get(data, "start").String()
			assert.Equal(suite.T(), date1, start, "Failed: Failed to validate start date")

			end := gjson.Get(data, "end").String()
			assert.Equal(suite.T(), date1, end, "Failed: Failed to validate end date")

			logger.Logf.Infof("Actual Usage :    ", actualAMount)
			assert.Equal(suite.T(), actualAMount, float64(3.75), "Failed: Failed to validate usage amount")
		}

		return true // keep iterating
	})

	assert.Equal(suite.T(), smlcount, 1, "Failed: Failed to create metering for small vm data")
	assert.Equal(suite.T(), medcount, 1, "Failed: Failed to create metering for medium vm data")

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.Equal(suite.T(), total_amount_from_response, 78.75, "Failed: Failed to validate usage amount")

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")
}
