//go:build Functional || ProductCatalogSyncRESTA || Regression || Billing || Intel
// +build Functional ProductCatalogSyncRESTA Regression Billing Intel

package BillingAPITest

import (
	_ "fmt"
	"goFramework/framework/library/financials/billing"

	"goFramework/framework/common/logger"

	"github.com/stretchr/testify/assert"
	
)

// var met_ret bool

func (suite *BillingAPITestSuite) TestSyncProductCatalog() {
	logger.Log.Info("Starting Product Catalog Sync API Test")
	ret_value, _ := billing.SyncProductCatalog("Test01 Master Plan", "validPayload", 200)
	assert.Equal(suite.T(), ret_value, false, "Failed: Test: Product Catalog Sync API Test")
}

