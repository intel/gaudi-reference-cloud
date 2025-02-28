//go:build NonFunctional || Longevity
// +build NonFunctional Longevity

package CaAPITest

import (
	// "goFramework/framework/common/logger"
	"goFramework/framework/library/financials/cloudAccounts"
	"github.com/stretchr/testify/assert"
	"time"
)

// IDC_1.0_CAcc_Create_Del_CloudAcc_Random_with_Members
func (suite *CaAPITestSuite) Test_CAcc_Create_Del_CloudAcc_Random_with_Members() {
	const duration = 48 * time.Hour // time defined for 2 days
	const interval = 3 * time.Minute // Interval time

	endTime := time.Now().Add(duration)
	for time.Now().Before(endTime) {
		name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
		get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false, 
			false, false, "ACCOUNT_TYPE_STANDARD", 200)
		assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
		ret_value, create_payload := cloudAccounts.AddMemberstoCAcc(1, get_CAcc_id, 200)
		assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to add member to the cloud account")
		time.Sleep(interval)
		CAcc_addedMember := create_payload.Members
		remove_CAcc, _ := cloudAccounts.DeleteMembersofCAcc(get_CAcc_id, CAcc_addedMember, 200)
		assert.NotEqual(suite.T(), remove_CAcc, "False", "Test Failed while removing a member from cloud account")
		ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
		assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
	}
}

