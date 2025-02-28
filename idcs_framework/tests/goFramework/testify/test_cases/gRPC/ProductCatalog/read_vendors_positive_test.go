//go:build Functional || Vendors || Positive || gRPC
// +build Functional Vendors Positive gRPC

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/grpc/productcatalog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var ret_val bool

func (suite *PcGrpcTestSuite) TestGetVendors() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_val, _ = productcatalog.Get_Vendors("noFilters", "getvendorGrpc")
	assert.Equal(suite.T(), ret_val, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcGrpcTestSuite) TestGetVendorsWithNameFilter() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_val, _ = productcatalog.Get_Vendors("filterbyName", "getvendorGrpc")
	assert.Equal(suite.T(), ret_val, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcGrpcTestSuite) TestGetVendorsWithIdFilter() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_val, _ = productcatalog.Get_Vendors("filterbyId", "getvendorGrpc")
	assert.Equal(suite.T(), ret_val, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcGrpcTestSuite) TestGetVendorsWithNameandIdFilter() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_val, _ = productcatalog.Get_Vendors("filterbyIdName", "getvendorGrpc")
	assert.Equal(suite.T(), ret_val, true, "Test Failed to Get Vendors with No filters")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestVendorsGrpcTestSuite(t *testing.T) {
	suite.Run(t, new(PcGrpcTestSuite))
}
