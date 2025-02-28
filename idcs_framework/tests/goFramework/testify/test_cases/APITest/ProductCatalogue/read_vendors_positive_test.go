//go:build Functional || Vendors || Regression || Positive
// +build Functional Vendors Regression Positive

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"

	"github.com/stretchr/testify/assert"
)

var ret_ven_val_pos bool

func (suite *PcAPITestSuite) TestGetVendors() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_ven_val_pos = financials.Get_Vendors("noFilters", 200)
	assert.Equal(suite.T(), ret_ven_val_pos, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestGetVendorsWithNameFilter() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_ven_val_pos = financials.Get_Vendors("filterbyName", 200)
	assert.Equal(suite.T(), ret_ven_val_pos, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestGetVendorsWithIdFilter() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_ven_val_pos = financials.Get_Vendors("filterbyId", 200)
	assert.Equal(suite.T(), ret_ven_val_pos, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestGetVendorsWithNameandIdFilter() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_ven_val_pos = financials.Get_Vendors("filterbyIdName", 200)
	assert.Equal(suite.T(), ret_ven_val_pos, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestGetVendorsCreationTime() {
	logger.Log.Info("Starting Test Get Vendors and check creation time is not null")
	ret_ven_val_pos = financials.Check_Vendor_Creation_Time("noFilters", 200)
	assert.Equal(suite.T(), ret_ven_val_pos, true, "Test Failed to check creation time in vendors")
}
