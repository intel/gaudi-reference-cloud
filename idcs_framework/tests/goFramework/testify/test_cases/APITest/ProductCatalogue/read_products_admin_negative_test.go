//go:build Functional || Products || Regression || Negative || Admin
// +build Functional Products Regression Negative Admin

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"

	"github.com/stretchr/testify/assert"
)

var ret_value_neg_admin bool

func (suite *PcAPITestSuite) TestGetproductsAdminWithInvalidNameFilter() {
	logger.Log.Info("Starting Test Get Products with Name Filter")
	ret_value_neg_admin = financials.Get_Products_Admin("filterbyInvalidName", 400)
	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Vendors with invalid Name Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminWithInvalidIdFilter() {
	logger.Log.Info("Starting Test Get Products with Id Filter")
	ret_value_neg_admin = financials.Get_Products_Admin("filterbyInvalidId", 400)
	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Vendors with invalid Id Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByInvalidVendorId() {
	logger.Log.Info("Starting Test Get Products by Invalid VendorId Filter ")
	ret_value_neg_admin = financials.Get_Products_Admin("filterbyInvalidVendorId", 400)
	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Vendors with invalid Vendor Id Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByInvalidFamilyId() {
	logger.Log.Info("Starting Test Get Products by Invalid FamilyId Filter ")
	ret_value_neg_admin = financials.Get_Products_Admin("filterbyInvalidfamilyId", 400)
	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Vendors with invalid Family Id Filter")
}

// TODO: TWC4726-207:  bug needs to be fixed in 1.0.1, enable this test once fixed
// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByInvalidDescription() {
// 	logger.Log.Info("Starting Test Get Products by Invalid FamilyId Filter ")
// 	ret_value_neg_admin = financials.Get_Products_Admin("filterbyInvaliddescription", 400)
// 	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Vendors with invalid Description Filter")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByInvalidMetadata() {
	logger.Log.Info("Starting Test Get Products by Invalid FamilyId Filter ")
	ret_value_neg_admin = financials.Get_Products_Admin("filterbyInvalidmetadata", 400)
	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Vendors with invalid Metadata Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByInvalideccn() {
	logger.Log.Info("Starting Test Get Products by Invalid eccn Filter ")
	ret_value_neg_admin = financials.Get_Products_Admin("filterbyInvalideccn", 400)
	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Vendors with invalid eccn Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByInvalidpcq() {
	logger.Log.Info("Starting Test Get Products by Invalid pcq Filter ")
	ret_value_neg_admin = financials.Get_Products_Admin("filterbyInvalidpcq", 400)
	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Vendors with invalid pcq Filter")
}

func (suite *PcAPITestSuite) TestGetproductsAdminFilterByInvalidMatchExp() {
	logger.Log.Info("Starting Test Get Products by Invalid Match Expression Filter ")
	ret_value_neg_admin = financials.Get_Products_Admin("filterbyInvalidmatchExpr", 400)
	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Vendors with invalid Match Expression  Filter")
}



// TODO: TWC4726-207:  bug needs to be fixed in 1.0.1, enable this test once fixed
// func (suite *PcAPITestSuite) TestGetproductsAdminFilterByInvalidFilter() {
// 	logger.Log.Info("Starting Test Get Products by Invalid Filter ")
// 	ret_value_neg_admin = financials.Get_Products_Admin("filterbyInvalidFilter", 400)
// 	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Vendors with invalid  Filter")
// }

func (suite *PcAPITestSuite) TestGetproductsAdminWithWrongToken() {
	logger.Log.Info("Starting Test Get Products With Wrong Token ")
	ret_value_neg_admin = financials.Get_Products_Admin("noFilter", 401)
	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Products with wrong token")
}

func (suite *PcAPITestSuite) TestGetproductsAdminWithExpiredAuthToken() {
	logger.Log.Info("Starting Test Get Products With Expired Token ")
	ret_value_neg_admin = financials.Get_Products_Admin("noFilter", 401)
	assert.Equal(suite.T(), ret_value_neg_admin, true, "Test Failed to Get Products with expired token")
}
