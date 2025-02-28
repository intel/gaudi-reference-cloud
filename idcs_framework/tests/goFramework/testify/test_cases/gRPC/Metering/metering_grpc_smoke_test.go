//go:build Functional || MeteringGRPCSmokeSuite || Metering
// +build Functional MeteringGRPCSmokeSuite Metering

package GRPCTest

import (
	_ "fmt"
	"goFramework/framework/library/grpc/metering"
	"strconv"
	"testing"

	"goFramework/framework/common/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// var met_ret bool

func (suite *GRPCTestSuite) TestCreateMeteringRecords() {
	logger.Log.Info("Starting Metering Create API Test")
	ret_value, _ := metering.Create_Usage_Record("validPayload")
	assert.Equal(suite.T(), ret_value, true, "Starting Metering Create API Test")
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithNoFilter() {
	logger.Log.Info("Starting Metering Search API Test Without Filter")
	ret, _ := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, "", "Test Failed: Starting Metering Search API Test Without Filter")

}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithTransactionIdFilter() {
	logger.Log.Info("Starting Metering Read API Test with TransactionId filter")
	ret, _ := metering.Create_multiple_Records_with_transactionId(10)
	assert.Equal(suite.T(), ret, true, "Test Failed: Starting Metering Read API Test with TransactionId filter")
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithCloudIdFilter() {
	logger.Log.Info("Starting Metering Read API Test with cloudId filter")
	ret, _ := metering.Create_multiple_Records_with_CloudAccountId(10)
	assert.Equal(suite.T(), ret, true, "Test Failed: Starting Metering Read API Test with CloudId filter")
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithResourceIdFilter() {
	logger.Log.Info("Starting Metering Read API Test with ResourceId Filter")
	ret, _ := metering.Create_multiple_Records_with_resourceId(10)
	assert.Equal(suite.T(), ret, true, "Test Failed: Starting Metering Read API Test with resourceId filter")
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithCloudIdResourceId() {
	logger.Log.Info("Starting Metering Read API Test with cloudid and resourceId")
	ret, jsondata := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test with cloudid and resourceId")
	filter := metering.UsageFilter{
		CloudAccountId: &jsondata.CloudAccountId,
		ResourceId:     &jsondata.ResourceId,
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with cloudid and resourceId")
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithCloudIdTransactionId() {
	logger.Log.Info("Starting Metering Read API Test")
	ret, jsondata := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test with cloudid and transactionId")
	filter := metering.UsageFilter{
		CloudAccountId: &jsondata.CloudAccountId,
		ResourceId:     &jsondata.TransactionId,
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with cloudid and transactionId")
}

func (suite *GRPCTestSuite) TestUpdateRecords() {
	logger.Logf.Infof("Starting Metering Update API Test")
	i, data := metering.Create_Record_and_Get_Id()
	ret, _ := strconv.ParseInt(i, 10, 64)
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Update API Test")
	filter := metering.UsageUpdate{
		Id:       []int64{ret},
		Reported: true,
	}
	res := metering.Update_Usage_Record_with_dynamic_filter(filter, ret, data, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Update API Test")
}

func (suite *GRPCTestSuite) TestFindPreviousRecords() {
	logger.Logf.Infof("Starting Metering Find Previous API Test")
	ret, data := metering.Create_multiple_Records_with_resourceId(10)
	assert.Equal(suite.T(), ret, true, "Test Failed: Metering Find Previous API Test")
	filter := metering.UsagePrevious{
		ResourceId: data.ResourceId,
	}
	res, _ := metering.Find_Usage_Record_with_dynamic_filter(filter, data.Timestamp, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Find Previous API Test")
}

func TestGRPCSmokeTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCTestSuite))
}
