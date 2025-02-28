//go:build Smoke || Products || Positive
// +build Smoke Products Positive

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"

	"github.com/stretchr/testify/assert"
)

var test_ret_val bool

func (suite *PcAPITestSuite) TestSmokeGetproducts() {
	logger.Log.Info("Starting Test Get Products with No filters")
	test_ret_val = financials.Get_Products("noFilters", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestSmokeGetProductsWithNameFilter() {
	logger.Log.Info("Starting Test Get Products with Name Filter")
	test_ret_val = financials.Get_Products("filterbyName", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Name Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetProductsWithIdFilter() {
	logger.Log.Info("Starting Test Get Products with Id Filter")
	test_ret_val = financials.Get_Products("filterbyId", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Id Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetProductsWithNameandIdFilter() {
	logger.Log.Info("Starting Test Get Products By Name and Id Filter ")
	test_ret_val = financials.Get_Products("filterbyIdName", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Name and Id Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetProductsFilterByVendorId() {
	logger.Log.Info("Starting Test Get Products by VendorId Filter ")
	test_ret_val = financials.Get_Products("filterbyVendorId", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Vendor Id Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetProductsFilterByFamilyId() {
	logger.Log.Info("Starting Test Get Products by FamilyId Filter")
	test_ret_val = financials.Get_Products("filterbyfamilyId", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Family Id Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetProductsFilterByDescription() {
	logger.Log.Info("Starting Test Get Products by FamilyId Filter ")
	test_ret_val = financials.Get_Products("filterbydescription", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Description Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetProductsFilterByMetadata() {
	logger.Log.Info("Starting Test Get Products by FamilyId Filter ")
	test_ret_val = financials.Get_Products("filterbymetadata", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Metadata Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetProductsFilterByeccn() {
	logger.Log.Info("Starting Test Get Products by eccn Filter ")
	test_ret_val = financials.Get_Products("filterbyeccn", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with eccn Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetProductsFilterBypcq() {
	logger.Log.Info("Starting Test Get Products by pcq Filter ")
	test_ret_val = financials.Get_Products("filterbypcq", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with pcq Filter")
}
