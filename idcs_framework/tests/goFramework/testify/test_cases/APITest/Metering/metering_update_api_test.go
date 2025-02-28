//go:build Functional || UpdateUsageRecordREST || Metering || Regression
// +build Functional UpdateUsageRecordREST Metering Regression

package MeteringAPITest

import (
	_ "fmt"
	"goFramework/framework/common/logger"

	"goFramework/framework/library/financials/metering"
	"strconv"

	"github.com/stretchr/testify/assert"
)

func (suite *MeteringAPITestSuite) TestUpdateUsageRecords() {
	logger.Logf.Info("Starting Metering Update API Test")
	i, data := metering.Create_Record_and_Get_Id()
	ret, _ := strconv.ParseInt(i, 10, 64)
	logger.Logf.Info("Result of ret is", ret)
	filter := metering.UsageUpdate{
		Id:       []int64{ret},
		Reported: true,
	}
	res := metering.Update_Usage_Record_with_dynamic_filter(filter, i, data, 200, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Update API Test")
}

func (suite *MeteringAPITestSuite) TestUpdateRecordsWithOutId() {
	logger.Logf.Info("Starting Metering Update API Test without Id")
	ret, data := metering.Create_Record_and_Get_Id()
	logger.Logf.Info("Result of ret is", ret)

	filter := metering.UsageUpdate{
		Reported: true,
	}
	res := metering.Update_Usage_Record_with_dynamic_filter(filter, ret, data, 400, 0)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Update API Test without Id")

}

func (suite *MeteringAPITestSuite) TestUpdateRecordsWithOutReported() {
	logger.Logf.Infof("Starting Metering Update API Test without Reported")
	i, data := metering.Create_Record_and_Get_Id()
	ret, _ := strconv.ParseInt(i, 10, 64)
	logger.Logf.Info("Result of ret is", ret)

	filter := metering.UsageUpdate{
		Id: []int64{ret},
	}
	res := metering.Update_Usage_Record_with_dynamic_filter(filter, i, data, 200, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Update API Test without Reported")

}

func (suite *MeteringAPITestSuite) TestUpdateRecordsWithInvalidId() {
	logger.Logf.Infof("Starting Metering Update API Test with invalid Id")
	ret_value, _ := metering.Update_Usage_Record("invalid_values.id", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Update API Test with invalid Id")
	//assert.Contains(suite.T(), err, "{\"code\":3,\"message\":\"invalid input arguments, ignoring record update.\",\"details\":[]}", "Test Failed : Metering Update API Test with invalid Id")
}

func (suite *MeteringAPITestSuite) TestUpdateRecordsWithInvalidreported() {
	logger.Logf.Infof("Starting Metering Update API Test With Invalid Reported")
	ret_value, _ := metering.Update_Usage_Record("invalid_values.reported", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Update API Test With Invalid Reported")
	//assert.Contains(suite.T(), err, "{\"code\":3,\"message\":\"invalid input arguments, ignoring record update.\",\"details\":[]}", "Test Failed : Update API Test With Invalid Reported")
}

func (suite *MeteringAPITestSuite) TestUpdateRecordsWithEmptyPayload() {
	logger.Logf.Infof("Starting Metering Update API Test with empty payload")
	ret_value, err := metering.Update_Usage_Record("missingFields.empty", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Update API Test with empty payload")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record update", "Test Failed : Metering Update API Test with empty payload")
}

func (suite *MeteringAPITestSuite) TestUpdateRecordsWithExtraField() {
	logger.Logf.Infof("Starting Metering Update API Test with extra field")
	ret_value, _ := metering.Update_Usage_Record("invalid_values.extraField", 400)
	assert.Equal(suite.T(), ret_value, false, "Test Failed : Metering Update API Test with extra field")
	//assert.Contains(suite.T(), err, "{\"code\":3,\"message\":\"invalid input arguments, ignoring record update.\",\"details\":[]}", "Test Failed : Metering Update API Test with extra field")
}
