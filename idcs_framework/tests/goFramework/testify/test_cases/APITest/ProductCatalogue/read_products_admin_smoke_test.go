//go:build Smoke || Products || Positive || Admin
// +build Smoke Products Positive Admin

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"

	"github.com/stretchr/testify/assert"
)

var test_ret_val bool

func (suite *PcAPITestSuite) TestSmokeGetproductsAdmin() {
	logger.Log.Info("Starting Test Get Products with No filters")
	test_ret_val = financials.Get_Products_Admin("noFilters", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestSmokeGetproductsAdminWithNameFilter() {
	logger.Log.Info("Starting Test Get Products with Name Filter")
	test_ret_val = financials.Get_Products_Admin("filterbyName", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Name Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetproductsAdminWithIdFilter() {
	logger.Log.Info("Starting Test Get Products with Id Filter")
	test_ret_val = financials.Get_Products_Admin("filterbyId", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Id Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetproductsAdminWithNameandIdFilter() {
	logger.Log.Info("Starting Test Get Products By Name and Id Filter ")
	test_ret_val = financials.Get_Products_Admin("filterbyIdName", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Name and Id Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetproductsAdminFilterByVendorId() {
	logger.Log.Info("Starting Test Get Products by VendorId Filter ")
	test_ret_val = financials.Get_Products_Admin("filterbyVendorId", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Vendor Id Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetproductsAdminFilterByFamilyId() {
	logger.Log.Info("Starting Test Get Products by FamilyId Filter")
	test_ret_val = financials.Get_Products_Admin("filterbyfamilyId", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Family Id Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetproductsAdminFilterByDescription() {
	logger.Log.Info("Starting Test Get Products by FamilyId Filter ")
	test_ret_val = financials.Get_Products_Admin("filterbydescription", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Description Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetproductsAdminFilterByMetadata() {
	logger.Log.Info("Starting Test Get Products by FamilyId Filter ")
	test_ret_val = financials.Get_Products_Admin("filterbymetadata", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with Metadata Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetproductsAdminFilterByeccn() {
	logger.Log.Info("Starting Test Get Products by eccn Filter ")
	test_ret_val = financials.Get_Products_Admin("filterbyeccn", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with eccn Filter")
}

func (suite *PcAPITestSuite) TestSmokeGetproductsAdminFilterBypcq() {
	logger.Log.Info("Starting Test Get Products by pcq Filter ")
	test_ret_val = financials.Get_Products_Admin("filterbypcq", 200)
	assert.Equal(suite.T(), test_ret_val, true, "Test Failed to Get Vendors with pcq Filter")
}
