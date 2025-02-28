//go:build Functional || CreateUsageRecordGRPC || Metering
// +build Functional CreateUsageRecordGRPC Metering

package GRPCTest

import (
	_ "fmt"
	"goFramework/framework/library/grpc/metering"

	"testing"

	"goFramework/framework/common/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func (suite *GRPCTestSuite) TestCreateMeteringRecords() {
	logger.Log.Info("Starting Metering Create API Test")
	ret_value, _ := metering.Create_Usage_Record("validPayload")
	assert.Equal(suite.T(), ret_value, true, "Starting Metering Create API Test")
}

func (suite *GRPCTestSuite) TestCreateMeteringRecordsWithoutProperties() {
	logger.Log.Info("Starting Metering Create API Test without properties")
	ret_value, _ := metering.Create_Usage_Record("createWithoutProperties")
	assert.Equal(suite.T(), ret_value, true, "Starting Metering Create API Test without properties")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithOutTransActionId() {
	logger.Log.Info("Starting Metering Create API Test without Transaction Id")
	ret_value, err := metering.Create_Usage_Record("missingFields.transactionId")
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test without Transaction Id")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test without CloudAccountId")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithOutCloudAccountId() {
	logger.Log.Info("Starting Metering Create API Test without CloudAccountId")
	ret_value, err := metering.Create_Usage_Record("missingFields.cloudAccountId")
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test without CloudAccountId")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test without CloudAccountId")
}

func (suite *GRPCTestSuite) TestCreateRecordsandValidateCreation() {
	logger.Log.Info("Starting Metering Create API Test")
	ret_value, _ := metering.Create_Record_and_Get_Id()
	assert.NotEqual(suite.T(), ret_value, "", "Test Failed: Starting Metering Create API Test")

}

func (suite *GRPCTestSuite) TestCreateRecordsWithInvalidcloudAccountId() {
	logger.Log.Info("Starting Metering Create API Test Invalid Cloud Account Id")
	ret_value, err := metering.Create_Usage_Record("invalid_values.cloudAccountId")
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test wit invalid  CloudAccountId")
	assert.Contains(suite.T(), err, "error getting request data: bad input: expecting string", "Test Failed : Starting Metering Create API Test without CloudAccountId")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithInvalidResourceId() {
	logger.Log.Info("Starting Metering Create API Test with Invalid Resource Id")
	ret_value, err := metering.Create_Usage_Record("invalid_values.resourceId")
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test wit invalid  resourceId")
	assert.Contains(suite.T(), err, "error getting request data: bad input: expecting string", "Test Failed : Starting Metering Create API Test without resourceId")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithInvalidTransactionId() {
	logger.Log.Info("Starting Metering Create API Test with Invalid transactionId")
	ret_value, err := metering.Create_Usage_Record("invalid_values.transactionId")
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test wit invalid  transactionId")
	assert.Contains(suite.T(), err, "error getting request data: bad input: expecting string", "Test Failed : Starting Metering Create API Test without transactionId")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithInvalidtimestamp() {
	logger.Log.Info("Starting Metering Create API Test with Invalid timestamp")
	ret_value, err := metering.Create_Usage_Record("invalid_values.timestamp")
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test wit invalid  timestamp")
	assert.Contains(suite.T(), err, "error getting request data: json: cannot unmarshal", "Test Failed : Starting Metering Create API Test without timestamp")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithMissingResourceId() {
	logger.Log.Info("Starting Metering Create API Test with missing resourceId")
	ret_value, err := metering.Create_Usage_Record("missingFields.resourceId")
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test without resourceId")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test without resourceId")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithMissingTransactionId() {
	logger.Log.Info("Starting Metering Create API Test with missing transactionId")
	ret_value, err := metering.Create_Usage_Record("missingFields.transactionId")
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test without Transaction Id")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test without transactionId")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithMissingtimestamp() {
	logger.Log.Info("Starting Metering Create API Test with missing timestamp")
	ret_value, err := metering.Create_Usage_Record("missingFields.timestamp")
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test without timestamp")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test without timestamp")
}

func (suite *GRPCTestSuite) TestCreateRecordsWithEmptyPayload() {
	logger.Log.Info("Starting Metering Create API Test with empty payload")
	ret_value, err := metering.Create_Usage_Record("missingFields.empty")
	assert.Equal(suite.T(), ret_value, true, "Test Failed : Starting Metering Create API Test with empty  payload")
	assert.Contains(suite.T(), err, "invalid input arguments, ignoring record creation", "Test Failed : Starting Metering Create API Test with empty payload")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCreateGRPCTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCTestSuite))
}
