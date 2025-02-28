//go:build Functional || Products || Regression || Negative
// +build Functional Products Regression Negative

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"

	"github.com/stretchr/testify/assert"
)

var ret_value_neg bool

func (suite *PcAPITestSuite) TestGetProductsWithInvalidNameFilter() {
	logger.Log.Info("Starting Test Get Products with Name Filter")
	ret_value_neg = financials.Get_Products("filterbyInvalidName", 400)
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Name Filter")
}

func (suite *PcAPITestSuite) TestGetProductsWithInvalidIdFilter() {
	logger.Log.Info("Starting Test Get Products with Id Filter")
	ret_value_neg = financials.Get_Products("filterbyInvalidId", 400)
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Id Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByInvalidVendorId() {
	logger.Log.Info("Starting Test Get Products by Invalid VendorId Filter ")
	ret_value_neg = financials.Get_Products("filterbyInvalidVendorId", 400)
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Vendor Id Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByInvalidFamilyId() {
	logger.Log.Info("Starting Test Get Products by Invalid FamilyId Filter ")
	ret_value_neg = financials.Get_Products("filterbyInvalidfamilyId", 400)
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Family Id Filter")
}

// TODO: TWC4726-207:  bug needs to be fixed in 1.0.1, enable this test once fixed
// func (suite *PcAPITestSuite) TestGetProductsFilterByInvalidDescription() {
// 	logger.Log.Info("Starting Test Get Products by Invalid FamilyId Filter ")
// 	ret_value_neg = financials.Get_Products("filterbyInvaliddescription", 400)
// 	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Description Filter")
// }

func (suite *PcAPITestSuite) TestGetProductsFilterByInvalidMetadata() {
	logger.Log.Info("Starting Test Get Products by Invalid FamilyId Filter ")
	ret_value_neg = financials.Get_Products("filterbyInvalidmetadata", 400)
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Metadata Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByInvalideccn() {
	logger.Log.Info("Starting Test Get Products by Invalid eccn Filter ")
	ret_value_neg = financials.Get_Products("filterbyInvalideccn", 400)
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid eccn Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByInvalidpcq() {
	logger.Log.Info("Starting Test Get Products by Invalid pcq Filter ")
	ret_value_neg = financials.Get_Products("filterbyInvalidpcq", 400)
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid pcq Filter")
}

func (suite *PcAPITestSuite) TestGetProductsFilterByInvalidMatchExp() {
	logger.Log.Info("Starting Test Get Products by Invalid Match Expression Filter ")
	ret_value_neg = financials.Get_Products("filterbyInvalidmatchExpr", 400)
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid Match Expression  Filter")
}



// TODO: TWC4726-207:  bug needs to be fixed in 1.0.1, enable this test once fixed
// func (suite *PcAPITestSuite) TestGetProductsFilterByInvalidFilter() {
// 	logger.Log.Info("Starting Test Get Products by Invalid Filter ")
// 	ret_value_neg = financials.Get_Products("filterbyInvalidFilter", 400)
// 	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Vendors with invalid  Filter")
// }

func (suite *PcAPITestSuite) TestGetProductsWithWrongToken() {
	logger.Log.Info("Starting Test Get Products With Wrong Token ")
	ret_value_neg = financials.Get_Products("noFilter", 401)
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Products with wrong token")
}

func (suite *PcAPITestSuite) TestGetProductsWithExpiredAuthToken() {
	logger.Log.Info("Starting Test Get Products With Expired Token ")
	ret_value_neg = financials.Get_Products("noFilter", 401)
	assert.Equal(suite.T(), ret_value_neg, true, "Test Failed to Get Products with expired token")
}
