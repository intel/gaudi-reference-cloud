//go:build Functional || Single || Metering || Regression
// +build Functional Single Metering Regression

package MeteringAPITest

import (
	_ "fmt"
	"goFramework/framework/library/financials/metering"

	"goFramework/framework/common/logger"

	"github.com/stretchr/testify/assert"
)

func (suite *MeteringAPITestSuite) TestSingleTransaction() {
	logger.Log.Info("Starting Metering Single Transaction Test")
	ret_value1, data1 := metering.Create_Record_and_Get_Id()
	logger.Logf.Info("Id created is ", ret_value1)
	assert.NotEqual(suite.T(), ret_value1, "", "Test Failed: Starting Metering Single Transaction API Test")
	ret_value2, _ := metering.Create_Duplicate_Record_and_Get_Id(data1.CloudAccountId, data1.ResourceId, data1.TransactionId)
	logger.Logf.Info("Id created is ", ret_value2)
	assert.Equal(suite.T(), ret_value1, ret_value2, "Test Failed:Starting Metering Single Transaction API Test")

}
