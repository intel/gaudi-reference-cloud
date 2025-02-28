//go:build Functional || SearchUsageRecordREST || Metering || Regression
// +build Functional SearchUsageRecordREST Metering Regression

package MeteringAPITest

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/metering"

	"goFramework/testify/test_cases/testutils"
	_ "log"
	"strconv"

	_ "time"

	"github.com/stretchr/testify/assert"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var met_ret bool
var test MeteringAPITestSuite
var searchId int64
var starttime timestamppb.Timestamp
var endtimestamp timestamppb.Timestamp

func (suite *MeteringAPITestSuite) TestReadApiSearchUsageRecordsWithNoFilter() {
	logger.Log.Info("Starting Metering Search API Test Without Filter")
	i, _ := metering.Create_Record_and_Get_Id()
	ret, _ := strconv.ParseInt(i, 10, 64)
	logger.Logf.Info("Result of ret is ", ret)
	filter := metering.UsageFilter{}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 1)

	//validateresid, _ := strconv.ParseInt(resid, 10, 64)
	assert.Equal(suite.T(), res, false, "Test Failed: Starting Metering Search API Test Without Filter")
	//assert.Contains(suite.T(),res,false,"Search results were more than expected number of records")

}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithTransactionIdFilter() {
	logger.Log.Info("Starting Metering Read API Test with TransactionId filter")
	_, jsonData := metering.Create_multiple_Records_with_transactionId(10)
	logger.Logf.Info("Result of transactionId created is ", jsonData.TransactionId)
	filter := metering.UsageFilter{
		TransactionId: &jsonData.TransactionId,
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 10)

	assert.Equal(suite.T(), res, true, "Test Failed: Starting Metering Search API Test with TransactionId filter")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithCloudIdFilter() {
	logger.Log.Info("Starting Metering Read API Test with cloudId filter")
	_, jsonData := metering.Create_multiple_Records_with_CloudAccountId(10)
	logger.Logf.Info("Result of cloudAccount created is ", jsonData.CloudAccountId)
	filter := metering.UsageFilter{
		CloudAccountId: &jsonData.CloudAccountId,
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 10)
	assert.Equal(suite.T(), res, true, "Test Failed: Starting Metering Read API Test with CloudId filter")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithResourceIdFilter() {
	logger.Log.Info("Starting Metering Read API Test with ResourceId Filter")
	_, jsonData := metering.Create_multiple_Records_with_resourceId(10)
	logger.Logf.Info("Result of resourceId created is ", jsonData.ResourceId)
	filter := metering.UsageFilter{
		ResourceId: &jsonData.ResourceId,
	}
	ret, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 10)
	assert.Equal(suite.T(), ret, true, "Test Failed: Starting Metering Read API Test with resourceId filter")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithCloudIdResourceId() {
	logger.Log.Info("Starting Metering Read API Test with cloudid and resourceId")
	_, jsondata := metering.Create_Record_and_Get_Id()
	filter := metering.UsageFilter{
		CloudAccountId: &jsondata.CloudAccountId,
		ResourceId:     &jsondata.ResourceId,
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with cloudid and resourceId")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithCloudIdTransactionId() {
	logger.Log.Info("Starting Metering Read API Test")
	_, jsondata := metering.Create_Record_and_Get_Id()
	filter := metering.UsageFilter{
		CloudAccountId: &jsondata.CloudAccountId,
		TransactionId:  &jsondata.TransactionId,
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with cloudid and transactionId")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithTimestamp() {
	logger.Log.Info("Starting Metering Read API Test with Time stamp")
	_, jsondata := metering.Create_Record_and_Get_Id()
	filter := metering.UsageFilter{
		StartTime: &jsondata.Timestamp,
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 1)
	assert.Equal(suite.T(), res, false, "Test Failed: Metering Read API Test with Time stamp")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithReported() {
	logger.Log.Info("Starting Metering Read API Test with Reported")
	_, jsondata := metering.Create_Record_and_Get_Id()
	filter := metering.UsageFilter{
		StartTime: &jsondata.Timestamp,
		Reported:  testutils.GetBoolPointer(true),
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 1)
	assert.Equal(suite.T(), res, false, "Test Failed: Metering Read API Test with Reported")

}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithAllFilters() {
	logger.Log.Info("Starting Metering Read API Test with all Filters")
	_, jsondata := metering.Create_Record_and_Get_Id()
	filter := metering.UsageFilter{
		StartTime:      &jsondata.Timestamp,
		Reported:       testutils.GetBoolPointer(true),
		CloudAccountId: &jsondata.CloudAccountId,
		ResourceId:     &jsondata.ResourceId,
		TransactionId:  &jsondata.TransactionId,
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test  with all Filters")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithId() {
	logger.Log.Info("Starting Metering Read API Test with id")
	ret, _ := metering.Create_Record_and_Get_Id()
	n, _ := strconv.ParseInt(ret, 10, 64)
	filter := metering.UsageFilter{
		Id: testutils.GetInt64Pointer(n),
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test  with id")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithOneCorrectFilter() {
	logger.Log.Info("Starting Metering Read API Test with one correct filter")
	_, jsondata := metering.Create_Record_and_Get_Id()
	filter := metering.UsageFilter{
		CloudAccountId: &jsondata.CloudAccountId,
		TransactionId:  testutils.GetStringPointer("qqqqqq"),
		ResourceId:     testutils.GetStringPointer("zzzz"),
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test  with one correct filter")

}

func (suite *MeteringAPITestSuite) TestReadApiSearchNonexistingCloudAccountId() {
	logger.Log.Info("Starting Metering Read API Test with Non existing Cloud Account Id")
	ret, _ := metering.Create_Record_and_Get_Id()
	logger.Logf.Info("result of ret is", ret)
	filter := metering.UsageFilter{
		Reported:       testutils.GetBoolPointer(false),
		CloudAccountId: testutils.GetStringPointer("888888"),
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test  with non existing cloud account id")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchNonexistingResourceId() {
	logger.Log.Info("Starting Metering Read API Test Non existing ResourceId")
	ret, _ := metering.Create_Record_and_Get_Id()
	logger.Logf.Info("result of ret is", ret)
	filter := metering.UsageFilter{
		ResourceId: testutils.GetStringPointer("abcd"),
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test  with non existing resource id")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchNonexistingTransactionId() {
	logger.Log.Info("Starting Metering Read API Test with Non existing TransactionId")
	ret, _ := metering.Create_Record_and_Get_Id()
	logger.Logf.Info("result of ret is", ret)
	filter := metering.UsageFilter{
		TransactionId: testutils.GetStringPointer("abc123d"),
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with Non existing TransactionId")

}

func (suite *MeteringAPITestSuite) TestReadApiSearchNonexistingId() {
	logger.Log.Info("Starting Metering Read API Test with Non existing Id")
	ret, _ := metering.Create_Record_and_Get_Id()
	logger.Logf.Info("result of ret is", ret)
	filter := metering.UsageFilter{

		Id: testutils.GetInt64Pointer(100000000000000),
	}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Read API Test with Non existing Id")

}

func (suite *MeteringAPITestSuite) TestSearchRecordsWithInvalidcloudAccountId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid CloudAccountId")
	ret_value, err := metering.Search_Usage_Record("invalid_values.resourceId", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid CloudAccountId")
	assert.Contains(suite.T(), err, "", "Test Failed : Metering Search API Test with Invalid CloudAccountId")
}

func (suite *MeteringAPITestSuite) TestSearchRecordsWithInvalidId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Id")
	ret_value, err := metering.Search_Usage_Record("invalid_values.id", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid Id")
	assert.Contains(suite.T(), err, "", "Test Failed : Metering Search API Test with Invalid Id")
}

func (suite *MeteringAPITestSuite) TestSearchRecordsWithInvalidResourceId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Resource Id")
	ret_value, err := metering.Search_Usage_Record("invalid_values.resourceId", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid Resource Id")
	assert.Contains(suite.T(), err, "", "Test Failed : Metering Search API Test with Invalid Resource Id")
}

func (suite *MeteringAPITestSuite) TestSearchRecordsWithInvalidTransactionId() {
	logger.Log.Info("Starting Metering Search API Test with Invalid TransactionID")
	ret_value, err := metering.Search_Usage_Record("invalid_values.transactionId", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid TransactionID")
	assert.Contains(suite.T(), err, "", "Test Failed : Metering Search API Test with Invalid TransactionID")
}

func (suite *MeteringAPITestSuite) TestSearchRecordsWithInvalidstarttime() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Start Time")
	ret_value, err := metering.Search_Usage_Record("invalid_values.starttimestamp", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid Start Time")
	assert.Contains(suite.T(), err, "", "Test Failed : Metering Search API Test with Invalid Start Time")
}

func (suite *MeteringAPITestSuite) TestSearchRecordsWithInvalidendtime() {
	logger.Log.Info("Starting Metering Search API Test with invalid End Time")
	ret_value, err := metering.Search_Usage_Record("invalid_values.endtimestamp", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid end Time")
	assert.Contains(suite.T(), err, "", "Test Failed : Metering Search API Test with Invalid end Time")
}

func (suite *MeteringAPITestSuite) TestSearchRecordsWithInvalidreported() {
	logger.Log.Info("Starting Metering Search API Test with Invalid Reported")
	ret_value, err := metering.Search_Usage_Record("invalid_values.reported", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Search API Test with Invalid Reported")
	assert.Contains(suite.T(), err, "", "Test Failed : Metering Search API Test with Invalid Reported")
}
