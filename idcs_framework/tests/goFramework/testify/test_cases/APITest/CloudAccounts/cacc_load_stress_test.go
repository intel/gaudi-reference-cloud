//go:build NonFunctional || Load
// +build NonFunctional Load

package CaAPITest

import (
	"sync"
	//"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/utils"
	"time"
	"github.com/stretchr/testify/assert"
)


// IDC_1.0_CAcc_Create_Del_SU_CloudAcc_60_Concurrent
func (suite *CaAPITestSuite) Test_CAcc_Create_Del_SU_CloudAcc_60_Concurrent() {
	Accounts := utils.Get_numAccounts()
	numAccounts := int(Accounts)
	const interval = 3 * time.Minute // Interval time
	var group sync.WaitGroup
	for i := 1; i <= numAccounts; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			// Creating a Standard User Cloud Account
			name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
			get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false, 
				false, false, "ACCOUNT_TYPE_STANDARD", 200)
			assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
			time.Sleep(interval)
			ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
			assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account")
		}(i)
	}
	group.Wait()
}

// IDC_1.0_CAcc_Create_Del_IU_CloudAcc_60_Concurrent
func (suite *CaAPITestSuite) Test_CAcc_Create_Del_IU_CloudAcc_60_Concurrent() {
	Accounts := utils.Get_numAccounts()
	numAccounts := int(Accounts)
	const interval = 3 * time.Minute // Interval time
	var group sync.WaitGroup
	for i:=1; i<=numAccounts; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			// Creating a Intel User Cloud Account
			name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
			get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false, 
				true, false, "ACCOUNT_TYPE_INTEL", 200)
			assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a intel user cloud account")
			time.Sleep(interval)
			ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
			assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account")
		}(i)
	}
	group.Wait()
}

// IDC_1.0_CAcc_Create_Del_PU_CloudAcc_60_Concurrent
func (suite *CaAPITestSuite) Test_CAcc_Create_Del_PU_CloudAcc_60_Concurrent() {
	Accounts := utils.Get_numAccounts()
	numAccounts := int(Accounts)
	const interval = 3 * time.Minute // Interval time
	var group sync.WaitGroup
	for i:=1; i<=numAccounts; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			// Creating a Premium User Cloud Account
			name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
			get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false, 
				true, true, "ACCOUNT_TYPE_PREMIUM", 200)
			assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account")
			time.Sleep(interval)
			ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
			assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account")
		}(i)
	}
	group.Wait()
}

// IDC_1.0_CAcc_Create_Del_IU_CloudAcc_60_Concurrent
func (suite *CaAPITestSuite) Test_CAcc_Create_Del_EU_CloudAcc_60_Concurrent() {
	Accounts := utils.Get_numAccounts()
	numAccounts := int(Accounts)
	const interval = 3 * time.Minute // Interval time
	var group sync.WaitGroup
	for i:=1; i<=numAccounts; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			// Creating a Enterprise User Cloud Account
			name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
			get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false, 
				true, true, "ACCOUNT_TYPE_ENTERPRISE", 200)
			assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account")
			time.Sleep(interval)
			ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
			assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account")
		}(i)
	}
	group.Wait()
}
