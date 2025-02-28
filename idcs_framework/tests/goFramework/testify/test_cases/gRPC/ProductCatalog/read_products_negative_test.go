//go:build Functional || Products || Negative || gRPC
// +build Functional Products Negative gRPC

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/grpc/productcatalog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var ret_value_neg bool
var pc_neg_error string

func (suite *PcGrpcTestSuite) TestGetProductsWithInvalidNameFilter() {
	logger.Log.Info("Starting Test Get Products with Name Filter")
	ret_value_neg, pc_neg_error = productcatalog.Get_Products("filterbyInvalidName", "negative_test")
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Name Filter")
	assert.Contains(suite.T(), pc_neg_error, "pc_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with invalid Name Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsWithInvalidIdFilter() {
	logger.Log.Info("Starting Test Get Products with Id Filter")
	ret_value_neg, pc_neg_error = productcatalog.Get_Products("filterbyInvalidId", "negative_test")
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Id Filter")
	assert.Contains(suite.T(), pc_neg_error, "pc_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with invalid Id Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByInvalidVendorId() {
	logger.Log.Info("Starting Test Get Products by Invalid VendorId Filter ")
	ret_value_neg, pc_neg_error = productcatalog.Get_Products("filterbyInvalidVendorId", "negative_test")
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Vendor Id Filter")
	assert.Contains(suite.T(), pc_neg_error, "pc_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with invalid Vendor Id Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByInvalidFamilyId() {
	logger.Log.Info("Starting Test Get Products by Invalid FamilyId Filter ")
	ret_value_neg, pc_neg_error = productcatalog.Get_Products("filterbyInvalidfamilyId", "negative_test")
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Family Id Filter")
	assert.Contains(suite.T(), pc_neg_error, "pc_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with invalid Family Id Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByInvalidDescription() {
	logger.Log.Info("Starting Test Get Products by Invalid FamilyId Filter ")
	ret_value_neg, pc_neg_error = productcatalog.Get_Products("filterbyInvaliddescription", "negative_test")
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Description Filter")
	assert.Contains(suite.T(), pc_neg_error, "pc_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with invalid Description Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByInvalidMetadata() {
	logger.Log.Info("Starting Test Get Products by Invalid FamilyId Filter ")
	ret_value_neg, pc_neg_error = productcatalog.Get_Products("filterbyInvalidmetadata", "negative_test")
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Metadata Filter")
	assert.Contains(suite.T(), pc_neg_error, "pc_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with invalid Metadata Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByInvalideccn() {
	logger.Log.Info("Starting Test Get Products by Invalid eccn Filter ")
	ret_value_neg, pc_neg_error = productcatalog.Get_Products("filterbyInvalideccn", "negative_test")
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid eccn Filter")
	assert.Contains(suite.T(), pc_neg_error, "pc_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with invalid eccn Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByInvalidpcq() {
	logger.Log.Info("Starting Test Get Products by Invalid pcq Filter ")
	ret_value_neg, pc_neg_error = productcatalog.Get_Products("filterbyInvalidpcq", "negative_test")
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid pcq Filter")
	assert.Contains(suite.T(), pc_neg_error, "pc_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with invalid pcq Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByInvalidMatchExp() {
	logger.Log.Info("Starting Test Get Products by Invalid Match Expression Filter ")
	ret_value_neg, pc_neg_error = productcatalog.Get_Products("filterbyInvalidmatchExpr", "negative_test")
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Match Expression  Filter")
	assert.Contains(suite.T(), pc_neg_error, "pc_neg_error getting request data: bad input: expecting string", "Test Failed to Get Vendors with invalid Match Expression  Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByInvalidFilter() {
	logger.Log.Info("Starting Test Get Products by Invalid Filter ")
	ret_value_neg, pc_neg_error = productcatalog.Get_Products("filterbyInvalidFilter", "negative_test")
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid  Filter")
	assert.Contains(suite.T(), pc_neg_error, "pc_neg_error getting request data: message type proto.ProductFilter has no known field named testFilter", "Test Failed to Get Vendors with invalid  Filter")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestProductsNegGrpcTestSuite(t *testing.T) {
	suite.Run(t, new(PcGrpcTestSuite))
}
