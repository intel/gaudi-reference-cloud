//go:build Functional || FindUsageRecordGRPC || Metering
// +build Functional FindUsageRecordGRPC Metering

package GRPCTest

import (
	_ "fmt"
	"goFramework/framework/common/logger"
	_ "goFramework/framework/library"
	"goFramework/framework/library/grpc/metering"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func (suite *GRPCTestSuite) TestFindPreviousRecords() {
	logger.Logf.Infof("Starting Metering Find Previous API Test")
	ret, data := metering.Create_multiple_Records_with_resourceId(10)
	assert.Equal(suite.T(), ret, true, "Test Failed: Metering Find Previous API Test")
	filter := metering.UsagePrevious{
		ResourceId: data.ResourceId,
	}
	res, _ := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 1)
	assert.Equal(suite.T(), res, true, " Test Failed: Metering Find Previous API Test")
}

func (suite *GRPCTestSuite) TestFindPreviousRecordsWithOutId() {
	logger.Logf.Infof("Starting Metering Find Previous API Test without Id")
	ret, data := metering.Create_multiple_Records_with_resourceId(10)
	assert.Equal(suite.T(), ret, true, "Test Failed: Metering Find Previous API Test")
	filter := metering.UsagePrevious{
		ResourceId: data.ResourceId,
	}
	res, _ := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Find Previous API Test Without Id")

}

func (suite *GRPCTestSuite) TestFindPreviousRecordsWithOutResourceId() {
	logger.Logf.Infof("Starting Metering Find Previous API Test without ResourceId")
	ret, data := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Find Previous API Test without Reported")
	filter := metering.UsagePrevious{
		Id: ret,
	}
	res, err := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Find Previous API Test without Reported")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring find request", "Test Failed : Metering Find Previous API Test without ResourceId")

}

func (suite *GRPCTestSuite) TestFindPreviousRecordsWithInvalidId() {
	logger.Logf.Infof("Starting Metering Find Previous API Test with invalid Id")
	ret, data := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Find Previous API Test without Reported")
	filter := metering.UsagePrevious{
		Id:         "abc",
		ResourceId: data.ResourceId,
	}
	res, err := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Find Previous API Test with invalid Id")
	assert.Contains(suite.T(), err, "error getting request data", "Test Failed : Metering Find Previous API Test with invalid Id")
}

func (suite *GRPCTestSuite) TestFindPreviousRecordsWithInvalidResourceid() {
	logger.Logf.Infof("Starting Metering Find Previous API Test With Invalid Resource Id")
	ret, data := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Find Previous API Test without Reported")
	filter := metering.UsagePrevious{
		Id:         "10",
		ResourceId: "10",
	}
	res, err := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Find Previous API Test with invalid Id")
	assert.Contains(suite.T(), err, "error getting request data", "Test Failed : Metering Find Previous API Test with invalid Id")
}

func (suite *GRPCTestSuite) TestFindPreviousRecordsWithEmptyPayload() {
	logger.Logf.Infof("Starting Metering Find Previous API Test with empty payload")
	ret_value, err := metering.Find_Previous_Usage_Record("missingFields.empty")
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Find Previous API Test with empty payload")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring find request", "Test Failed : Metering Find Previous API Test with empty payload")
}

func (suite *GRPCTestSuite) TestFindPreviousRecordsWithExtraField() {
	logger.Logf.Infof("Starting Metering Find Previous API Test with extra field")
	ret_value, err := metering.Find_Previous_Usage_Record("invalid_values.extraField")
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Find Previous API Test with extra field")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring find request", "Test Failed : Metering Find Previous API Test with extra field")
}

// In order for 'go test' to run this suite, we need to create
//
//	a normal test function and pass our suite to suite.Run
func TestGrpcFindPreviousUsageSuite(t *testing.T) {

	//Single client used for testing the APIs
	suite.Run(t, new(GRPCTestSuite))

}
