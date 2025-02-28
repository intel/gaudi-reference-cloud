//go:build Functional || SearchUsageRecordGRPC || Metering
// +build Functional SearchUsageRecordGRPC Metering

package GRPCTest

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/grpc/metering"

	"goFramework/testify/test_cases/testutils"
	_ "log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithNoFilter() {
	logger.Log.Info("Starting Metering Search API Test Without Filter")
	ret, _ := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, "", " Test Failed: Starting Metering Search API Test Without Filter")

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
		TransactionId:  &jsondata.TransactionId,
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with cloudid and transactionId")
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithCloudIdTransactionIdResourceId() {
	logger.Log.Info("Starting Metering Read API Test")
	ret, jsondata := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test with cloudid and transactionId")
	filter := metering.UsageFilter{
		CloudAccountId: &jsondata.CloudAccountId,
		TransactionId:  &jsondata.TransactionId,
		ResourceId:     &jsondata.ResourceId,
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with cloudid and transactionId")
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithTimestamp() {
	logger.Log.Info("Starting Metering Read API Test with Time stamp")
	ret, data := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test with Time stamp")
	filter := metering.UsageFilter{
		StartTime:      testutils.GetStringPointer("2022-11-29T13:34:00Z"),
		CloudAccountId: &data.CloudAccountId,
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with Time stamp")
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithReported() {
	logger.Log.Info("Starting Metering Read API Test with Reported")
	ret, data := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test with Reported")
	filter := metering.UsageFilter{
		StartTime:      testutils.GetStringPointer("2022-11-29T13:34:00Z"),
		Reported:       testutils.GetBoolPointer(false),
		CloudAccountId: &data.CloudAccountId,
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with Reported")

}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithAllFilters() {
	logger.Log.Info("Starting Metering Read API Test with all Filters")
	ret, jsondata := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test  with all Filters")
	filter := metering.UsageFilter{
		StartTime:      testutils.GetStringPointer("2022-11-29T13:34:00Z"),
		Reported:       testutils.GetBoolPointer(false),
		CloudAccountId: &jsondata.CloudAccountId,
		ResourceId:     &jsondata.ResourceId,
		TransactionId:  &jsondata.TransactionId,
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test  with all Filters")
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithId() {
	logger.Log.Info("Starting Metering Read API Test with id")
	ret, _ := metering.Create_Record_and_Get_Id()
	n, _ := strconv.ParseInt(ret, 10, 64)
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test  with all Filters")
	filter := metering.UsageFilter{
		Id: testutils.GetInt64Pointer(n),
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test  with all Filters")
}

func (suite *GRPCTestSuite) TestReadApiSearchRecordsWithOneCorrectFilter() {
	logger.Log.Info("Starting Metering Read API Test with Non Existing Values in Filters")
	ret, jsondata := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test  with all Filters")
	filter := metering.UsageFilter{
		CloudAccountId: &jsondata.CloudAccountId,
		TransactionId:  testutils.GetStringPointer("qqqqqq"),
		ResourceId:     testutils.GetStringPointer("zzzz"),
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test  with all Filters")

}

func (suite *GRPCTestSuite) TestReadApiSearchNonexistingCloudAccountId() {
	logger.Log.Info("Starting Metering Read API Test with Non existing Cloud Account Id")
	ret, _ := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test  with all Filters")
	filter := metering.UsageFilter{
		Reported:       testutils.GetBoolPointer(false),
		CloudAccountId: testutils.GetStringPointer(fmt.Sprint(time.Now().Nanosecond())[:6]),
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test  with all Filters")
}

func (suite *GRPCTestSuite) TestReadApiSearchNonexistingResourceId() {
	logger.Log.Info("Starting Metering Read API Test Non existing ResourceId")
	ret, _ := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test  with all Filters")
	filter := metering.UsageFilter{
		ResourceId: testutils.GetStringPointer("abcd"),
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test  with all Filters")
}

func (suite *GRPCTestSuite) TestReadApiSearchNonexistingTransactionId() {
	logger.Log.Info("Starting Metering Read API Test with Non existing TransactionId")
	ret, _ := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test  with all Filters")
	filter := metering.UsageFilter{
		TransactionId: testutils.GetStringPointer("abc123d"),
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with Non existing TransactionId")

}

func (suite *GRPCTestSuite) TestReadApiSearchNonexistingId() {
	logger.Log.Info("Starting Metering Read API Test with Non existing Id")
	ret, _ := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret, true, "Test Failed: Metering Read API Test with Non existing Id")
	filter := metering.UsageFilter{
		Id: testutils.GetInt64Pointer(13500),
	}
	res := metering.Search_Usage_Record_with_dynamic_filter(filter, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with Non existing Id")

}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidcloudAccountId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid CloudAccountId")
	ret_value, err := metering.Search_Usage_Record("invalid_values.resourceId")
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid CloudAccountId")
	assert.Contains(suite.T(), err, "error getting request data: bad input: expecting string", "Test Failed : Metering Search API Test with Invalid CloudAccountId")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Id")
	ret_value, err := metering.Search_Usage_Record("invalid_values.Id")
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid Id")
	assert.Contains(suite.T(), err, "error getting request data: bad input: expecting string", "Test Failed : Metering Search API Test with Invalid Id")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidResourceId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Resource Id")
	ret_value, err := metering.Search_Usage_Record("invalid_values.resourceId")
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid Resource Id")
	assert.Contains(suite.T(), err, "error getting request data: bad input: expecting string", "Test Failed : Metering Search API Test with Invalid Resource Id")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidTransactionId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid TransactionID")
	ret_value, err := metering.Search_Usage_Record("invalid_values.transactionId")
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid TransactionID")
	assert.Contains(suite.T(), err, "error getting request data: bad input: expecting string", "Test Failed : Metering Search API Test with Invalid TransactionID")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidstarttime() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Start Time")
	ret_value, err := metering.Search_Usage_Record("invalid_values.starttimestamp")
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid Start Time")
	assert.Contains(suite.T(), err, "error getting request data: json: cannot unmarshal number into Go value of type string", "Test Failed : Metering Search API Test with Invalid Start Time")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidendtime() {
	logger.Log.Info("Starting Metering Search API Test with invalid End Time")
	ret_value, err := metering.Search_Usage_Record("invalid_values.endtimestamp")
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid end Time")
	assert.Contains(suite.T(), err, "error getting request data: json: cannot unmarshal number into Go value of type string", "Test Failed : Metering Search API Test with Invalid end Time")
}

func (suite *GRPCTestSuite) TestSearchRecordsWithInvalidreported() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Reported")
	ret_value, err := metering.Search_Usage_Record("invalid_values.reported")
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid Reported")
	assert.Contains(suite.T(), err, "error getting request data: bad input: expecting boolean", "Test Failed : Metering Search API Test with Invalid Reported")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGrpcSearchMeteringSuite(t *testing.T) {
	suite.Run(t, new(GRPCTestSuite))

}
