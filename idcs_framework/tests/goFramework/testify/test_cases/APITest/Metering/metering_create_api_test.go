//go:build Functional || CreateUsageRecordREST || Metering || CreateM || Regression
// +build Functional CreateUsageRecordREST Metering CreateM Regression

package MeteringAPITest

import (
	_ "fmt"

	"github.com/stretchr/testify/assert"

	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/metering"
)

// var met_ret bool

func (suite *MeteringAPITestSuite) TestCreateMeteringRecords() {
	logger.Log.Info("Starting Metering Create API Test")
	ret_value, err := metering.Create_Usage_Record("validPayload", 200)
	assert.Equal(suite.T(), ret_value, true, "Starting Metering Create API Test")
	assert.Contains(suite.T(), err, "{}", "Records created")
}

func (suite *MeteringAPITestSuite) TestCreateMeteringRecordsWithoutProperties() {
	logger.Log.Info("Starting Metering Create API Test without properties")
	ret_value, _ := metering.Create_Usage_Record("createWithoutProperties", 200)
	assert.Equal(suite.T(), ret_value, true, "Starting Metering Create API Test without properties")
	//assert.Contains(suite.T(), err, "nvalid input arguments, ignoring record creation", "Test Failed :Starting Metering Create API Test without properties")
}
func (suite *MeteringAPITestSuite) TestCreateRecordsWithOutTransActionId() {
	logger.Log.Info("Starting Metering Create API Test without Transaction Id")
	ret_value, _ := metering.Create_Usage_Record("missingFields.transactionId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test without Transaction Id")
	//assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test without CloudAccountId")
}

func (suite *MeteringAPITestSuite) TestCreateRecordsWithOutCloudAccountId() {
	logger.Log.Info("Starting Metering Create API Test without CloudAccountId")
	ret_value, _ := metering.Create_Usage_Record("missingFields.cloudAccountId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test without CloudAccountId")
	//assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test without CloudAccountId")
}

func (suite *MeteringAPITestSuite) TestCreateRecordsandValidateCreation() {
	logger.Log.Info("Starting Metering Create API Test")
	ret_value, _ := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret_value, "", "Test Failed: Starting Metering Create API Test")
}

func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidcloudAccountId() {
	logger.Log.Info("Starting Metering Create API Test Invalid Cloud Account Id")
	ret_value, err := metering.Create_Usage_Record("invalid_values.cloudAccountId", 400)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test wit invalid  CloudAccountId")
	assert.Contains(suite.T(), err, "invalid value for string type:", "Test Failed : Starting Metering Create API Test without CloudAccountId")
}

func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidResourceId() {
	logger.Log.Info("Starting Metering Create API Test with Invalid Resource Id")
	ret_value, err := metering.Create_Usage_Record("invalid_values.resourceId", 400)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test wit invalid  resourceId")
	assert.Contains(suite.T(), err, "invalid value for string type:", "Test Failed : Starting Metering Create API Test without resourceId")
}

func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidTransactionId() {
	logger.Log.Info("Starting Metering Create API Test with Invalid transactionId")
	ret_value, err := metering.Create_Usage_Record("invalid_values.transactionId", 400)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test wit invalid  transactionId")
	assert.Contains(suite.T(), err, "invalid value for string type:", "Test Failed : Starting Metering Create API Test without transactionId")
}

func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidtimestamp() {
	logger.Log.Info("Starting Metering Create API Test with Invalid timestamp")
	ret_value, err := metering.Create_Usage_Record("invalid_values.timestamp", 400)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test wit invalid  timestamp")
	assert.Contains(suite.T(), err, "invalid google.protobuf.Timestamp value", "Test Failed : Starting Metering Create API Test wit invalid  timestamp")

}

func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingResourceId() {
	logger.Log.Info("Starting Metering Create API Test with missing resourceId")
	ret_value, _ := metering.Create_Usage_Record("missingFields.resourceId", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test without resourceId")
	//assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test without resourceId")
}

func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingtimestamp() {
	logger.Log.Info("Starting Metering Create API Test with missing timestamp")
	ret_value, _ := metering.Create_Usage_Record("missingFields.timestamp", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test without timestamp")
	//assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test without timestamp")
}

func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyPayload() {
	logger.Log.Info("Starting Metering Create API Test with empty payload")
	ret_value, _ := metering.Create_Usage_Record("missingFields.empty", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty  payload")
	//assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test with empty payload")
}

func (suite *MeteringAPITestSuite) TestCreateRecordsWithSameIDs() {
	logger.Log.Info("Starting Metering Create API Test with same transactionID and resourceID")
	ret_value, err := metering.Create_Usage_Record("sameids", 200)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with same transactionID and resourceID")
	assert.Contains(suite.T(), err, "{}", "Test Failed : Starting Metering Create API Test with same transactionID and resourceID")
}

// // Negative tests on properties, commented these tests

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidAvailabilityZone() {
// 	logger.Log.Info("Starting Metering Create API Test with invalid availabilityZone")
// 	ret_value, err := metering.Create_Usage_Record("invalidPropertyFields.availabilityZone", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with invalid availabilityZone")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with invalid availabilityZone")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidclusterId() {
// 	logger.Log.Info("Starting Metering Create API Test with invalid clusterId")
// 	ret_value, err := metering.Create_Usage_Record("invalidPropertyFields.clusterId", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with invalid clusterId")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with invalid clusterId")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidDeleted() {
// 	logger.Log.Info("Starting Metering Create API Test with invalid deleted")
// 	ret_value, err := metering.Create_Usage_Record("invalidPropertyFields.deleted", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with invalid deleted")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with invalid deleted")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidinstanceName() {
// 	logger.Log.Info("Starting Metering Create API Test with invalid instanceName")
// 	ret_value, err := metering.Create_Usage_Record("invalidPropertyFields.instanceName", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with invalid instanceName")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with invalid instanceName")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidinstanceType() {
// 	logger.Log.Info("Starting Metering Create API Test with invalid instanceType")
// 	ret_value, err := metering.Create_Usage_Record("invalidPropertyFields.instanceType", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with invalid instanceType")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with invalid instanceType")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidRegion() {
// 	logger.Log.Info("Starting Metering Create API Test with invalid region")
// 	ret_value, err := metering.Create_Usage_Record("invalidPropertyFields.region", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with invalid region")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with invalid region")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidRunningSeconds() {
// 	logger.Log.Info("Starting Metering Create API Test with invalid runningSeconds")
// 	ret_value, err := metering.Create_Usage_Record("invalidPropertyFields.runningSeconds", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with invalid runningSeconds")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with invalid runningSeconds")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidRunningSeconds1() {
// 	logger.Log.Info("Starting Metering Create API Test with invalid runningSeconds1")
// 	ret_value, err := metering.Create_Usage_Record("invalidPropertyFields.runningSeconds1", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with invalid runningSeconds1")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with invalid runningSeconds1")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidServiceType() {
// 	logger.Log.Info("Starting Metering Create API Test with invalid serviceType")
// 	ret_value, err := metering.Create_Usage_Record("invalidPropertyFields.serviceType", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with invalid serviceType")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with invalid serviceType")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithInvalidServiceType1() {
// 	logger.Log.Info("Starting Metering Create API Test with invalid serviceType1")
// 	ret_value, err := metering.Create_Usage_Record("invalidPropertyFields.serviceType1", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with invalid serviceType1")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with invalid serviceType1")
// }

// // Empty Values to the property fields

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyAvailabilityZone() {
// 	logger.Log.Info("Starting Metering Create API Test with empty availabilityZone")
// 	ret_value, err := metering.Create_Usage_Record("emptyPropertyFields.availabilityZone", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty availabilityZone")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with empty availabilityZone")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyclusterId() {
// 	logger.Log.Info("Starting Metering Create API Test with empty clusterId")
// 	ret_value, err := metering.Create_Usage_Record("emptyPropertyFields.clusterId", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty clusterId")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with empty clusterId")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyDeleted() {
// 	logger.Log.Info("Starting Metering Create API Test with empty deleted")
// 	ret_value, err := metering.Create_Usage_Record("emptyPropertyFields.deleted", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty deleted")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with empty deleted")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyinstanceName() {
// 	logger.Log.Info("Starting Metering Create API Test with empty instanceName")
// 	ret_value, err := metering.Create_Usage_Record("emptyPropertyFields.instanceName", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty instanceName")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with empty instanceName")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyinstanceType() {
// 	logger.Log.Info("Starting Metering Create API Test with empty instanceType")
// 	ret_value, err := metering.Create_Usage_Record("emptyPropertyFields.instanceType", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty instanceType")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with empty instanceType")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyRegion() {
// 	logger.Log.Info("Starting Metering Create API Test with empty region")
// 	ret_value, err := metering.Create_Usage_Record("emptyPropertyFields.region", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty region")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with empty region")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyRunningSeconds() {
// 	logger.Log.Info("Starting Metering Create API Test with empty runningSeconds")
// 	ret_value, err := metering.Create_Usage_Record("emptyPropertyFields.runningSeconds", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty runningSeconds")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with empty runningSeconds")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyRunningSeconds1() {
// 	logger.Log.Info("Starting Metering Create API Test with empty runningSeconds1")
// 	ret_value, err := metering.Create_Usage_Record("emptyPropertyFields.runningSeconds1", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty runningSeconds1")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with empty runningSeconds1")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyServiceType() {
// 	logger.Log.Info("Starting Metering Create API Test with empty serviceType")
// 	ret_value, err := metering.Create_Usage_Record("emptyPropertyFields.serviceType", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty serviceType")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with empty serviceType")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyServiceType1() {
// 	logger.Log.Info("Starting Metering Create API Test with empty serviceType1")
// 	ret_value, err := metering.Create_Usage_Record("emptyPropertyFields.serviceType1", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty serviceType1")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with empty serviceType1")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingAvailabilityZone() {
// 	logger.Log.Info("Starting Metering Create API Test with missing availabilityZone")
// 	ret_value, err := metering.Create_Usage_Record("missingPropertiesFields.availabilityZone", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with missing availabilityZone")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with missing availabilityZone")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingclusterId() {
// 	logger.Log.Info("Starting Metering Create API Test with missing clusterId")
// 	ret_value, err := metering.Create_Usage_Record("missingPropertiesFields.clusterId", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with missing clusterId")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with missing clusterId")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingDeleted() {
// 	logger.Log.Info("Starting Metering Create API Test with missing deleted")
// 	ret_value, err := metering.Create_Usage_Record("missingPropertiesFields.deleted", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with missing deleted")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with missing deleted")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissinginstanceName() {
// 	logger.Log.Info("Starting Metering Create API Test with missing instanceName")
// 	ret_value, err := metering.Create_Usage_Record("missingPropertiesFields.instanceName", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with missing instanceName")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with missing instanceName")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissinginstanceType() {
// 	logger.Log.Info("Starting Metering Create API Test with missing instanceType")
// 	ret_value, err := metering.Create_Usage_Record("missingPropertiesFields.instanceType", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with missing instanceType")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with missing instanceType")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingRegion() {
// 	logger.Log.Info("Starting Metering Create API Test with missing region")
// 	ret_value, err := metering.Create_Usage_Record("missingPropertiesFields.region", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with missing region")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with missing region")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingRunningSeconds() {
// 	logger.Log.Info("Starting Metering Create API Test with missing runningSeconds")
// 	ret_value, err := metering.Create_Usage_Record("missingPropertiesFields.runningSeconds", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with missing runningSeconds")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with missing runningSeconds")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingRunningSeconds1() {
// 	logger.Log.Info("Starting Metering Create API Test with missing runningSeconds1")
// 	ret_value, err := metering.Create_Usage_Record("missingPropertiesFields.runningSeconds1", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with missing runningSeconds1")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with missing runningSeconds1")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingServiceType() {
// 	logger.Log.Info("Starting Metering Create API Test with missing serviceType")
// 	ret_value, err := metering.Create_Usage_Record("missingPropertiesFields.serviceType", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with missing serviceType")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with missing serviceType")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingServiceType1() {
// 	logger.Log.Info("Starting Metering Create API Test with missing serviceType1")
// 	ret_value, err := metering.Create_Usage_Record("missingPropertiesFields.serviceType1", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with missing serviceType1")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with missing serviceType1")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithMissingProperties() {
// 	logger.Log.Info("Starting Metering Create API Test with missing Properties")
// 	ret_value, err := metering.Create_Usage_Record("missingProperties.Properties", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with with missing Properties")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with with missing Properties")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithEmptyProperties() {
// 	logger.Log.Info("Starting Metering Create API Test with empty Properties")
// 	ret_value, err := metering.Create_Usage_Record("emptyProperties.Properties", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with with empty Properties")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with with empty Properties")
// }

// // Non Existing

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithNonExistingAvailabilityZone() {
// 	logger.Log.Info("Starting Metering Create API Test with non existing availabilityZone")
// 	ret_value, err := metering.Create_Usage_Record("nonExistingPropertyFields.availabilityZone", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with with non existing availabilityZone")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with non existing availabilityZone")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithNonExistingclusterId() {
// 	logger.Log.Info("Starting Metering Create API Test with non existing clusterId")
// 	ret_value, err := metering.Create_Usage_Record("nonExistingPropertyFields.clusterId", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with with non existing clusterId")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with non existing clusterId")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithNonExistingcloudAccId() {
// 	os.Setenv("intel_user_test", "True")
// 	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	enterpriseId := "testeid-01"
// 	//Create Cloud Account
// 	userName := utils.Get_UserName("Standard")
// 	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
// 	userToken = "Bearer " + userToken
// 	base_url := utils.Get_Base_Url1()
// 	url := base_url + "/v1/cloudaccounts"
// 	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
// 	if err == nil {
// 		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
// 	}
// 	time.Sleep(1 * time.Minute)
// 	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
// 	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
// 	CountryCode := gjson.Get(responsePayload, "countryCode").String()
// 	assert.Equal(suite.T(), "IN", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

// 	now := time.Now().UTC()
// 	previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
// 	fmt.Println("Metering Date", previousDate)
// 	resourceId := uuid.NewString()
// 	create_payload := financials_utils.EnrichInvalidMeteringCreatePayload(compute_utils.GetInvalidMeteringCreatePayload(),
// 		uuid.NewString(), resourceId, "123456789012", previousDate, "vm-spr-sml", "smallvm", "30000", "us-dev-1a", "harvester1", "false", "us-dev-1", "ComputeAsAService")
// 	fmt.Println("create_payload", create_payload)
// 	metering_api_base_url := base_url + "/v1/meteringrecords"
// 	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
// 	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")
// 	accId := "123456789012"
// 	filter := metering.SearchInvalidMeteringRecord{
// 		CloudAccountId: &accId,
// 	}
// 	flag, _ := metering.Get_Invalid_Metering_Record(filter, 200, 1, "DEFAULT_INVALIDITY_REASON")
// 	assert.Equal(suite.T(), flag, true, "Test Failed : CloudAccount Id not found in invalid records")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with non existing cloudAccId")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithNonExistinginstanceType() {
// 	logger.Log.Info("Starting Metering Create API Test with non existing region")
// 	ret_value, err := metering.Create_Usage_Record("nonExistingPropertyFields.region", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with with non existing region")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with non existing region")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithNonExistingRegion() {
// 	logger.Log.Info("Starting Metering Create API Test with non existing instanceType")
// 	ret_value, err := metering.Create_Usage_Record("nonExistingPropertyFields.instanceType", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with with non existing instanceType")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with non existing instanceType")
// }

// func (suite *MeteringAPITestSuite) TestCreateRecordsWithNonExistingServiceType() {
// 	logger.Log.Info("Starting Metering Create API Test with non existing serviceType")
// 	ret_value, err := metering.Create_Usage_Record("nonExistingPropertyFields.serviceType", 400)
// 	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with with non existing serviceType")
// 	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed :Metering Create API Test with non existing serviceType")
// }
