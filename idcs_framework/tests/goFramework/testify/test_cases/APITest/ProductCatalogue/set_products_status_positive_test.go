//go:build Functional || Products || Positive || KindRegression || Status
// +build Functional Products Positive KindRegression Status

package ProductCatalogue

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials"
	"goFramework/framework/library/grpc/productcatalog"

	"github.com/stretchr/testify/assert"
)

var ret_value_pos bool

func (suite *PcAPITestSuite) TestSetReadyStatus() {
	logger.Log.Info("Starting Test Set Status")
	familyId, productId, vendorId, val := financials.Fetch_Product_Details_For_Status_Payload("filterbyId", 200)
	if val != true {
		ret_value_pos = false
	} else {
		ret_value_pos = productcatalog.Set_Status("PRODUCT_STATUS_READY", "no error", familyId, productId, vendorId, 200)
	}
	assert.Equal(suite.T(), ret_value_pos, true, "Status Changed to Ready Succesfully")
}

func (suite *PcAPITestSuite) TestSetProvisioningStatus() {
	logger.Log.Info("Starting Test Set Status")
	familyId, productId, vendorId, val := financials.Fetch_Product_Details_For_Status_Payload("filterbyId", 200)
	if val != true {
		ret_value_pos = false
	} else {
		ret_value_pos = productcatalog.Set_Status("PRODUCT_STATUS_PROVISIONING", "no error", familyId, productId, vendorId, 200)
	}
	assert.Equal(suite.T(), ret_value_pos, true, "Status Changed to Ready Succesfully")
}

func (suite *PcAPITestSuite) TestSetErrorStatus() {
	logger.Log.Info("Starting Test Set Status")
	familyId, productId, vendorId, val := financials.Fetch_Product_Details_For_Status_Payload("filterbyId", 200)
	if val != true {
		ret_value_pos = false
	} else {
		ret_value_pos = productcatalog.Set_Status("PRODUCT_STATUS_ERROR", "no error", familyId, productId, vendorId, 200)
	}
	assert.Equal(suite.T(), ret_value_pos, true, "Status Changed to Ready Succesfully")
}
