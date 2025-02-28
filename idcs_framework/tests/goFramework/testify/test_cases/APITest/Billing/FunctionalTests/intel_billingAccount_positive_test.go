//go:build Functional || BillingAccountRest || Billing || Regression || Positive || Intel
// +build Functional BillingAccountRest Billing Regression Positive Intel

package BillingAPITest

import (
	//"encoding/json"
	//"fmt"
	"github.com/stretchr/testify/assert"
	//"github.com/tidwall/gjson"
	"goFramework/framework/common/logger"
	"goFramework/utils"

	//"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	//"time"
	"os"
)

func (suite *BillingAPITestSuite) TestCreateCloudAccountIntel() {
	os.Setenv("intel_user_test", "True")
	logger.Log.Info("Creating Intel User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	userName := utils.Get_UserName("Intel")
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an intel user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account(Intel User)")
	os.Unsetenv("intel_user_test")
}
