//go:build Functional || UpdateUsageRecordGRPC || Metering
// +build Functional UpdateUsageRecordGRPC Metering

package GRPCTest

import (
	_ "fmt"
	"goFramework/framework/common/logger"
	_ "goFramework/framework/library"

	//"io"
	"goFramework/framework/library/grpc/metering"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func (suite *GRPCTestSuite) TestUpdateRecords() {
	logger.Logf.Infof("Starting Metering Update API Test")
	i, data := metering.Create_Record_and_Get_Id()
	ret, _ := strconv.ParseInt(i, 10, 64)
	assert.NotEqual(suite.T(), ret, true, " Test Failed: Metering Update API Test")
	filter := metering.UsageUpdate{
		Id:       []int64{ret},
		Reported: true,
	}
	res := metering.Update_Usage_Record_with_dynamic_filter(filter, ret, data, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Update API Test")
}

func (suite *GRPCTestSuite) TestUpdateRecordsWithOutId() {
	logger.Logf.Infof("Starting Metering Update API Test without Id")
	i, data := metering.Create_Record_and_Get_Id()
	ret, _ := strconv.ParseInt(i, 10, 64)
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Update API Test without Id")
	filter := metering.UsageUpdate{
		Reported: true,
	}
	res := metering.Update_Usage_Record_with_dynamic_filter(filter, ret, data, 0)
	assert.NotEqual(suite.T(), res, true, "Test Failed: Metering Update API Test without Id")

}

func (suite *GRPCTestSuite) TestUpdateRecordsWithOutReported() {
	logger.Logf.Infof("Starting Metering Update API Test without Reported")
	i, data := metering.Create_Record_and_Get_Id()
	ret, _ := strconv.ParseInt(i, 10, 64)
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Update API Test without Reported")
	filter := metering.UsageUpdate{
		Id: []int64{ret},
	}
	res := metering.Update_Usage_Record_with_dynamic_filter(filter, ret, data, 0)
	assert.NotEqual(suite.T(), res, true, "Test Failed: Metering Update API Test without Reported")

}

func (suite *GRPCTestSuite) TestUpdateRecordsWithInvalidId() {
	logger.Logf.Infof("Starting Metering Update API Test with invalid Id")
	ret_value, err := metering.Update_Usage_Record("invalid_values.id")
	assert.NotEqual(suite.T(), ret_value, true, "Test Failed : Metering Update API Test with invalid Id")
	assert.Contains(suite.T(), err, "error getting request data: strconv.ParseInt", "Test Failed : Metering Update API Test with invalid Id")
}

func (suite *GRPCTestSuite) TestUpdateRecordsWithInvalidreported() {
	logger.Logf.Infof("Starting Metering Update API Test With Invalid Reported")
	ret_value, err := metering.Update_Usage_Record("invalid_values.reported")
	assert.NotEqual(suite.T(), ret_value, true, "Test Failed : Update API Test With Invalid Reported")
	assert.Contains(suite.T(), err, "error getting request data: bad input:", "Test Failed : Update API Test With Invalid Reported")
}

func (suite *GRPCTestSuite) TestUpdateRecordsWithEmptyPayload() {
	logger.Logf.Infof("Starting Metering Update API Test with empty payload")
	ret_value, err := metering.Update_Usage_Record("missingFields.empty")
	assert.NotEqual(suite.T(), ret_value, true, "Test Failed : Metering Update API Test with empty payload")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record update", "Test Failed : Metering Update API Test with empty payload")
}

func (suite *GRPCTestSuite) TestUpdateRecordsWithExtraField() {
	logger.Logf.Infof("Starting Metering Update API Test with extra field")
	ret_value, err := metering.Update_Usage_Record("invalid_values.extraField")
	assert.NotEqual(suite.T(), ret_value, true, "Test Failed : Metering Update API Test with extra field")
	assert.Contains(suite.T(), err, "error getting request data: message type proto.UsageUpdate has no known field named extraField", "Test Failed : Metering Update API Test with extra field")
}

// In order for 'go test' to run this suite, we need to create
//
//	a normal test function and pass our suite to suite.Run
func TestGrpcUpdateUsageSuite(t *testing.T) {
	//Single client used for testing the APIs
	suite.Run(t, new(GRPCTestSuite))

}
