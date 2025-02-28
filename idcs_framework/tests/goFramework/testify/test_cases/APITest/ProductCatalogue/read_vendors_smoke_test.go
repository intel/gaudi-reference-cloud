//go:build Smoke || ProductsSmoke || Positive
// +build Smoke ProductsSmoke Positive

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"

	"github.com/stretchr/testify/assert"
)

var ret_val bool

func (suite *PcAPITestSuite) TestSmokeGetVendors() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_val = financials.Get_Vendors("noFilters", 200)
	assert.Equal(suite.T(), ret_val, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestSmokeGetVendorsWithNameFilter() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_val = financials.Get_Vendors("filterbyName", 200)
	assert.Equal(suite.T(), ret_val, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestSmokeGetVendorsWithIdFilter() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_val = financials.Get_Vendors("filterbyId", 200)
	assert.Equal(suite.T(), ret_val, true, "Test Failed to Get Vendors with No filters")
}

func (suite *PcAPITestSuite) TestSmokeGetVendorsWithNameandIdFilter() {
	logger.Log.Info("Starting Test Get Vendors ")
	ret_val = financials.Get_Vendors("filterbyIdName", 200)
	assert.Equal(suite.T(), ret_val, true, "Test Failed to Get Vendors with No filters")
}
