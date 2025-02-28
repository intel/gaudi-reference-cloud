//go:build Functional || FindUsageRecordREST || Metering || Regression
// +build Functional FindUsageRecordREST Metering Regression

package MeteringAPITest

import (
	_ "fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/metering"
	

	"github.com/stretchr/testify/assert"
	
)

func (suite *MeteringAPITestSuite) TestFindPreviousUsageRecords() {
	logger.Logf.Infof("Starting Metering Find Previous API Test")
	ret, data := metering.Create_Record_and_Get_Id()
	filter := metering.UsagePrevious{
		ResourceId: data.ResourceId,
		Id:ret,
	}
	res, _ := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 200, 1)
	assert.Equal(suite.T(), res, false, "Test Failed: Metering Find Previous API Test")
}

func (suite *MeteringAPITestSuite) TestFindPreviousRecordsWithOutId() {
	logger.Logf.Infof("Starting Metering Find Previous API Test without Id")
	_, data := metering.Create_multiple_Records_with_resourceId(10)
	logger.Logf.Info("Result of resourceId is ", data.ResourceId)
	filter := metering.UsagePrevious{
		ResourceId: data.ResourceId,
	}
	res, _ := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 400, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Find Previous API Test Without Id")

}

func (suite *MeteringAPITestSuite) TestFindPreviousRecordsWithOutResourceId() {
	logger.Logf.Infof("Starting Metering Find Previous API Test without ResourceId")
	ret, data := metering.Create_Record_and_Get_Id()
	logger.Logf.Info("Result of ret is ", ret)
	filter := metering.UsagePrevious{
		Id: ret,
	}
	res, _ := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 400, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Find Previous API Test without Reported")
	

}

func (suite *MeteringAPITestSuite) TestFindPreviousRecordsWithInvalidId() {
	logger.Logf.Infof("Starting Metering Find Previous API Test with invalid Id")
	_, data := metering.Create_Record_and_Get_Id()
	filter := metering.UsagePrevious{
		Id:         "abc",
		ResourceId: data.ResourceId,
	}
	res, _ := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 400, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Find Previous API Test with invalid Id")
}
	
func (suite *MeteringAPITestSuite) TestFindPreviousRecordsWithInvalidResourceid() {
	logger.Logf.Infof("Starting Metering Find Previous API Test With Invalid Resource Id")
	ret, data := metering.Create_Record_and_Get_Id()
	logger.Logf.Info("Result of ret is ", ret)
	filter := metering.UsagePrevious{
		Id:         ret,
		ResourceId: "10",
	}
	res, err := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 404, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Find Previous API Test with Invalid Resource Id")
	assert.Contains(suite.T(),err,"no matching records found")
}

func (suite *MeteringAPITestSuite) TestFindPreviousRecordsWithEmptyPayload() {
	logger.Logf.Infof("Starting Metering Find Previous API Test with empty payload")
	ret_value, err := metering.Find_Previous_Usage_Record("missingFields.empty", 400)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Metering Find Previous API Test with empty payload")
	assert.Contains(suite.T(),err,"invalid input arguments, ignoring find previous")
	
}

func (suite *MeteringAPITestSuite) TestFindPreviousRecordsWithExtraField() {
	logger.Logf.Infof("Starting Metering Find Previous API Test with extra field")
	ret_value, err := metering.Find_Previous_Usage_Record("invalid_values.extraField", 400)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Metering Find Previous API Test with extra field")
	assert.Contains(suite.T(),err, "invalid input arguments, ignoring find previous","Test Failed:Metering Find Previous API Test with extra field")
	
}

func (suite *MeteringAPITestSuite) TestFindPreviousRecordsWithMultipleIDs() {
	logger.Logf.Infof("Starting Metering Find Previous API Test with multiple IDs")
	ret_value, err := metering.Find_Previous_Usage_Record("multipleIDs", 400)
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Metering Find Previous API Test with multiple IDs")
	assert.Contains(suite.T(),err, "invalid input arguments, ignoring find previous","Test Failed:Metering Find Previous API Test with multiple IDs")
	

}
