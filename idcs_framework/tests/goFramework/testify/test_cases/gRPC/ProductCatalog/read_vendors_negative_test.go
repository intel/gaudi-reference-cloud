//go:build Functional || Vendors || Negative || gRPC
// +build Functional Vendors Negative gRPC

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/grpc/productcatalog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var ret_val_neg bool
var vendor_neg_error string

func (suite *PcGrpcTestSuite) TestGetVendorsWithInvalidNameFilter() {
	logger.Log.Info("Starting Test Get Vendors with invalid Name value filter ")
	ret_val_neg, vendor_neg_error = productcatalog.Get_Vendors("filterbyInvalidName", "negative_test")
	assert.Equal(suite.T(), ret_val_neg, true, "Test Failed to Get Vendors with Invlaid Name Filter")
	assert.Contains(suite.T(), vendor_neg_error, "vendor_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with Invlaid Name Filter")
}

func (suite *PcGrpcTestSuite) TestGetVendorsWithInvalidIdFilter() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_val_neg, vendor_neg_error = productcatalog.Get_Vendors("filterbyInvalidId", "negative_test")
	assert.Equal(suite.T(), ret_val_neg, true, "Test Failed to Get Vendors with Invalid Id Filter")
	assert.Contains(suite.T(), vendor_neg_error, "vendor_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with Invlaid Name Filter")
}

func (suite *PcGrpcTestSuite) TestGetVendorsWithInvalidFilter() {
	logger.Log.Info("Starting Test Get Vendors with Invalid Filter ")
	ret_val_neg, vendor_neg_error = productcatalog.Get_Vendors("filterbyInvalidFilter", "negative_test")
	assert.Equal(suite.T(), ret_val_neg, true, "Test Failed to Get Vendors with Invalid Filter")
	assert.Contains(suite.T(), vendor_neg_error, "message type proto.VendorFilter has no known field named testFilter", "Test Failed to Get Vendors with Invlaid Name Filter")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestVendorsNegGrpcTestSuite(t *testing.T) {
	suite.Run(t, new(PcGrpcTestSuite))
}
