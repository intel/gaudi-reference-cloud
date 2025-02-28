//go:build Functional || MeteringRestSmokeSuite || Metering || Regression
// +build Functional MeteringRestSmokeSuite Metering Regression

package MeteringAPITest

import (
	_ "fmt"
	"goFramework/framework/library/financials/metering"
	"strconv"

	"goFramework/framework/common/logger"

	"github.com/stretchr/testify/assert"
)

// var met_ret bool

func (suite *MeteringAPITestSuite) TestSmokeCreateMeteringRecords() {
	logger.Log.Info("Starting Metering Create API Test")
	ret_value, err := metering.Create_Usage_Record("validPayload", 200)
	assert.Equal(suite.T(), ret_value, true, "Starting Metering Create API Test")
	assert.Contains(suite.T(), err, "{}", "Records created")
}

func (suite *MeteringAPITestSuite) TestReadApiSearchRecordsWithNoFilter() {
	logger.Log.Info("Starting Metering Search API Test Without Filter")
	i, _ := metering.Create_Record_and_Get_Id()
	ret, _ := strconv.ParseInt(i, 10, 64)
	//assert.Equal(suite.T(), i, "", "Test Failed: Starting Metering Search API Test Without Filter")
	filter := metering.UsageFilter{}
	res, _ := metering.Search_Usage_Record_with_dynamic_filter(filter, 200, 1)
	logger.Logf.Info("Result of ret is ", ret)
	//validateresid, _ := strconv.ParseInt(resid, 10, 64)
	assert.Equal(suite.T(), res, false, "Test Failed: Starting Metering Search API Test Without Filter")
	//assert.Contains(suite.T(),res,false,"Search results were more than expected number of records")

}

func (suite *MeteringAPITestSuite) TestUpdateRecords() {
	logger.Logf.Infof("Starting Metering Update API Test")
	i, data := metering.Create_Record_and_Get_Id()
	ret, _ := strconv.ParseInt(i, 10, 64)
	logger.Logf.Info("Result of ret is ", ret)
	//assert.Equal(suite.T(), ret, ret, "Test Failed: Metering Update API Test")
	filter := metering.UsageUpdate{
		Id:       []int64{ret},
		Reported: false,
	}
	res := metering.Update_Usage_Record_with_dynamic_filter(filter, i, data, 200, 1)
	assert.Equal(suite.T(), res, true, "Test Failed: Metering Update API Test")
}

func (suite *MeteringAPITestSuite) TestFindPreviousRecords() {
	logger.Logf.Infof("Starting Metering Find Previous API Test")
	i, data := metering.Create_Record_and_Get_Id()
	ret, _ := strconv.ParseInt(i, 10, 64)
	logger.Logf.Info("Result of ret is ", ret)
	//assert.Equal(suite.T(), ret, ret, "Test Failed: Metering Find Previous API Test")
	filter := metering.UsagePrevious{
		ResourceId: data.ResourceId,
		Id:         i,
	}
	res, _ := metering.Find_Usage_Record_with_dynamic_filter(filter, data.ResourceId, 200, 1)
	assert.Equal(suite.T(), res, false, "Test Failed: Metering Find Previous API Test")
}
