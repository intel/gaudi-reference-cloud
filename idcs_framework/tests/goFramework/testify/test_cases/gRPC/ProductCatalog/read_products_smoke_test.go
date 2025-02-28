//go:build Smoke || Products || gRPC
// +build Smoke Products gRPC

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/grpc/productcatalog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var value bool

func (suite *PcGrpcTestSuite) TestGetProductsWithName() {
	logger.Log.Info("Starting Test Get Products with Name Filter")
	value, _ = productcatalog.Get_Products("filterbyName", "getproductAll")
	assert.Equal(suite.T(), value, true, "Test Failed to Get Vendors with Name Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsWithId() {
	logger.Log.Info("Starting Test Get Products with Id Filter")
	value, _ = productcatalog.Get_Products("filterbyId", "getproductAll")
	assert.Equal(suite.T(), value, true, "Test Failed to Get Vendors with Id Filter")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByFamilyIdFilter() {
	logger.Log.Info("Starting Test Get Products by  FamilyId ")
	value, _ = productcatalog.Get_Products("filterbyFamilyId", "getproductAll")
	assert.Equal(suite.T(), value, true, "Test Failed to Get Vendors with  FamilyId")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByDescriptionFilter() {
	logger.Log.Info("Starting Test Get Products by  Description ")
	value, _ = productcatalog.Get_Products("filterbyDescription", "getproductAll")
	assert.Equal(suite.T(), value, true, "Test Failed to Get Vendors with  Description")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMetadataFilter() {
	logger.Log.Info("Starting Test Get Products by  Metadata ")
	value, _ = productcatalog.Get_Products("filterbyMetadata", "getproduct1")
	assert.Equal(suite.T(), value, true, "Test Failed to Get Vendors with  Metadata")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByeccnFilter() {
	logger.Log.Info("Starting Test Get Products by  eccn ")
	value, _ = productcatalog.Get_Products("filterbyeccn", "getproductAll")
	assert.Equal(suite.T(), value, true, "Test Failed to Get Vendors with  eccn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterBypqnFilter() {
	logger.Log.Info("Starting Test Get Products by  pqn ")
	value, _ = productcatalog.Get_Products("filterbypqn", "getproductAll")
	assert.Equal(suite.T(), value, true, "Test Failed to Get Vendors with  pqn")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByMatchExpFilter() {
	logger.Log.Info("Starting Test Get Products by  MatchExp ")
	value, _ = productcatalog.Get_Products("filterbyMatchExp", "getproductAll")
	assert.Equal(suite.T(), value, true, "Test Failed to Get Vendors with  MatchExp")
}

func (suite *PcGrpcTestSuite) TestGetProductsFilterByVendorIdFilter() {
	logger.Log.Info("Starting Test Get Products by   VendorId ")
	value, _ = productcatalog.Get_Products("filterbyVendorId", "getproductAll")
	assert.Equal(suite.T(), value, true, "Test Failed to Get Vendors with  VendorId")

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestProductsGrpcSmokeTestSuite(t *testing.T) {
	suite.Run(t, new(PcGrpcTestSuite))
}
