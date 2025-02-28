//go:build Functional || Products || Negative || KindRegression || Status
// +build Functional Products Negative KindRegression Status

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"
	"goFramework/framework/library/grpc/productcatalog"

	"github.com/stretchr/testify/assert"
)

var ret_value_neg1 bool

func (suite *PcAPITestSuite) TestInvalidProductId() {
	logger.Log.Info("Starting Test Set Status")
	familyId, productId, vendorId, val := financials.Fetch_Product_Details_For_Status_Payload("filterbyInvalidId", 400)
	if val != true {
		ret_value_neg1 = false
	} else {
		ret_value_neg1 = productcatalog.Set_Status("PRODUCT_STATUS_READY", "no error", familyId, productId, vendorId, 200)
	}
	assert.NotEqual(suite.T(), ret_value_neg1, true, "No matching product would be found for empty product id")
}

func (suite *PcAPITestSuite) TestInvalidVendorId() {
	logger.Log.Info("Starting Test Set Status")
	familyId, productId, vendorId, val := financials.Fetch_Product_Details_For_Status_Payload("filterbyInvalidVendorId", 400)
	if val != true {
		ret_value_neg1 = false
	} else {
		ret_value_neg1 = productcatalog.Set_Status("PRODUCT_STATUS_PROVISIONING", "no error", familyId, productId, vendorId, 200)
	}
	assert.NotEqual(suite.T(), ret_value_neg1, true, "No matching product would be found for empty vendor id")
}

func (suite *PcAPITestSuite) TestInvalidFamilyId() {
	logger.Log.Info("Starting Test Set Status")
	familyId, productId, vendorId, val := financials.Fetch_Product_Details_For_Status_Payload("filterbyInvalidfamilyId", 400)
	if val != true {
		ret_value_neg1 = false
	} else {
		ret_value_neg1 = productcatalog.Set_Status("PRODUCT_STATUS_ERROR", "no error", familyId, productId, vendorId, 200)
	}
	assert.NotEqual(suite.T(), ret_value_neg1, true, "No matching product would be found for empty family id")
}
