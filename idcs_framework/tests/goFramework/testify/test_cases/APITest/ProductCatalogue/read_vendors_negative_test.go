//go:build Functional || Vendors || Regression || Negative
// +build Functional Vendors Regression Negative

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"

	"github.com/stretchr/testify/assert"
)

var ret_val_neg bool

// TODO: TWC4726-207:  bug needs to be fixed in 1.0.1, enable this test once fixed
// func (suite *PcAPITestSuite) TestGetVendorsWithInvalidNameFilter() {
// 	logger.Log.Info("Starting Test Get Vendors with invalid Name value filter ")
// 	ret_val_neg = financials.Get_Vendors("filterbyInvalidName", 200)
// 	assert.NotEqual(suite.T(), ret_val_neg, true, "Test Failed to Get Vendors with Invlaid Name Filter")
// }

// TODO: TWC4726-207:  bug needs to be fixed in 1.0.1, enable this test once fixed
// func (suite *PcAPITestSuite) TestGetVendorsWithInvalidIdFilter() {
// 	logger.Log.Info("Starting Test Get Vendors ")
// 	ret_val_neg = financials.Get_Vendors("filterbyInvalidId", 200)
// 	assert.NotEqual(suite.T(), ret_val_neg, true, "Test Failed to Get Vendors with Invalid Id Filter")
// }

// TODO: TWC4726-207:  bug needs to be fixed in 1.0.1, enable this test once fixed
// func (suite *PcAPITestSuite) TestGetVendorsWithInvalidFilter() {
// 	logger.Log.Info("Starting Test Get Vendors with Invalid Filter ")
// 	ret_val_neg = financials.Get_Vendors("filterbyInvalidFilter", 200)
// 	assert.NotEqual(suite.T(), ret_val_neg, true, "Test Failed to Get Vendors with Invalid Filter")
// }

func (suite *PcAPITestSuite) TestGetVendorsWithWrongToken() {
	logger.Log.Info("Starting Test Get Vendors with Wrong Token")
	ret_val_neg = financials.Get_Vendors("noFilters", 401)
	assert.Equal(suite.T(), ret_val_neg, true, "Test Failed to Get Vendors with Wrong Token")
}

func (suite *PcAPITestSuite) TestGetVendorsWithExpiredToken() {
	logger.Log.Info("Starting Test Get Vendors with Wrong Token")
	ret_val_neg = financials.Get_Vendors("noFilters", 401)
	assert.Equal(suite.T(), ret_val_neg, true, "Test Failed to Get Vendors with Expired Token")
}
