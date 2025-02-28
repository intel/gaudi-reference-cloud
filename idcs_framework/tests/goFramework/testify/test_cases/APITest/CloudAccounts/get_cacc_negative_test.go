//go:build All || NonFunctional || GetCloudAccounts || Negative || Regression
// +build All NonFunctional GetCloudAccounts Negative Regression

package CaAPITest

import (
	"github.com/stretchr/testify/assert"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/cloudAccounts"
)

// IDC_1.0_CAcc_SU
func (suite *CaAPITestSuite) Test_CAcc_SU() {
	logger.Log.Info("Ensure no related Aria account is created for standard user")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	check_Billing, _ := cloudAccounts.CAcc_check_BillingAcc(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), check_Billing, "False", "Test Failed while checking billing account for Cloud account")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_IU
func (suite *CaAPITestSuite) Test_CAcc_IU() {
	logger.Log.Info("Ensure no related Aria account is created for intel user")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	check_Billing, _ := cloudAccounts.CAcc_check_BillingAcc(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), check_Billing, "False", "Test Failed while checking billing account for Cloud account")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_GetById_Invalid
func (suite *CaAPITestSuite) Test_CAcc_GetById_Invalid() {
	logger.Log.Info("Retrieve Cloud Account by looking up Invalid Cloud Account ID")
	ret_value1, jsonStr := cloudAccounts.GetCAccById("12334", 400)
	logger.Logf.Info("jsonStr", jsonStr)
	assert.Equal(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Contains(suite.T(), jsonStr, "invalid CloudAccountId")
}

// IDC_1.0_CAcc_GetByName_Invalid
func (suite *CaAPITestSuite) Test_CAcc_GetByName_Invalid() {
	logger.Log.Info("Retrieve Cloud Account by looking up Invalid Cloud Account Name")
	ret_value, jsonStr := cloudAccounts.GetCAccByName("12345", 404)
	logger.Logf.Info("jsonStr", jsonStr)
	assert.Equal(suite.T(), ret_value, "False", "Test Failed to get cloudaccount by name")
}

// IDC_1.0_get_CAcc_with_Invalidendpoint
// func (suite *CaAPITestSuite) Test_get_CAcc_with_Invalidendpoint() {
// 	logger.Log.Info("Retrieving the Cloud Account using invalid endpoint")
// 	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
// 	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
// 		false, false, "ACCOUNT_TYPE_STANDARD", 200)
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
// 	ret_value, _ := cloudAccounts.GetCAccByInvalidEndpoint(parentid, 200)
// 	assert.Equal(suite.T(), ret_value, "False", "Test Failed while trying to get cloud account with invalid endpoint")
// 	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
// }

// IDC_1.0_Create_CAcc_duplicate_email_in_JWTtoken
// func (suite *CaAPITestSuite) Test_Create_CAcc_duplicate_email_in_JWTtoken() {
// 	logger.Log.Info("Creating a Enroll User Cloud Account using Enroll API with duplicate email")
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	get_CAcc_id, _, _ := cloudAccounts.CreateCAccwithEnroll("intel", enterpriseId, "zxcvbn@intel.com", tid, false, 200)
// 	logger.Log.Info("Created first account")
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a intel user cloud account using Enroll API")
// 	enterpriseId1, _, tid1 := cloudAccounts.Rand_token_payload_gen()
// 	get_CAcc_id1, _, _ := cloudAccounts.CreateCAccwithEnroll("intel", enterpriseId1, "zxcvbn@intel.com", tid1, false, 404)
// 	logger.Log.Info("Tried creating second account with duplicate email")
// 	assert.Equal(suite.T(), get_CAcc_id1, "False", "Failed while trying to create with duplicate email.")
// 	ret_value, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while deleting the cloud account")
// }

// IDC_1.0_patch_oid_with_duplicate_oid
func (suite *CaAPITestSuite) Test_Update_oid_with_duplicate_oid() {
	logger.Log.Info("Update oid with duplicate oid for a Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	name1, oid1, owner1, parentid1, tid1 := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id2, _ := cloudAccounts.CreateCloudAccount(name1, oid1, owner1, parentid1, tid1, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id2, "False", "Test Failed while creating a premium user cloud account")
	ret_value, _ := cloudAccounts.UpdateCAccDuplicateOidById(oid, get_CAcc_id2, 500)
	assert.Equal(suite.T(), ret_value, "False", "Test Failed while trying to update oid with duplicate oid for cloud account")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the first cloud account")
	ret_value3, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id2, 200)
	assert.NotEqual(suite.T(), ret_value3, "False", "Test Failed while deleting the second cloud account")
}

// IDC_1.0_patch_name_with_duplicate_name
func (suite *CaAPITestSuite) Test_Update_CAccName_with_duplicateName() {
	logger.Log.Info("Update name with duplicate name of Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	name1, oid1, owner1, parentid1, tid1 := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id2, _ := cloudAccounts.CreateCloudAccount(name1, oid1, owner1, parentid1, tid1, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id2, "False", "Test Failed while creating a premium user cloud account")
	ret_value, _ := cloudAccounts.UpdateCAccById(name, get_CAcc_id2, 500)
	assert.NotEqual(suite.T(), ret_value, "True", "Test Failed while trying to update name with duplicate name of the cloud account")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the first cloud account")
	ret_value3, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id2, 200)
	assert.NotEqual(suite.T(), ret_value3, "False", "Test Failed while deleting the second cloud account")
}

// IDC_1.0_create_CAcc_with_empty_email_id
// func (suite *CaAPITestSuite) Test_create_CAcc_with_empty_email_id() {
// 	logger.Log.Info("Creating a cloud account without email in jwt token")
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	enterpriseId := "testid"
// 	get_CAcc_id, _, _ := cloudAccounts.CreateCAccwithEnroll("standard", enterpriseId, "", tid, false, 200)
// 	assert.Equal(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account using Enroll API")
// }

// IDC_1.0_Patch_oid_by_Id
func (suite *CaAPITestSuite) Test_Update_oid_by_Id() {
	logger.Log.Info("check update oid for Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	ret_value, _ := cloudAccounts.UpdateCAccById("ValidPayload7", get_CAcc_id, 500)
	assert.Equal(suite.T(), ret_value, "False", "Test Failed while trying to update oid for cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_patch_cloudaccountId
func (suite *CaAPITestSuite) Test_Update_CAccId_by_Id() {
	logger.Log.Info("check update id for Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	ret_value, _ := cloudAccounts.UpdateCAccById("ValidPayload8", get_CAcc_id, 500)
	assert.Equal(suite.T(), ret_value, "False", "Test Failed while trying to update id for cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_MultipleMembers_1CAcc
func (suite *CaAPITestSuite) Test_CAcc_MultipleMembers_1CAcc() {
	logger.Log.Info("Add two Members to Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	//assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	_, _ = cloudAccounts.AddMemberstoCAcc(2, get_CAcc_id, 200)
	_, _ = cloudAccounts.GetCAccMembersById(get_CAcc_id, 200)
	//assert.Equal(suite.T(), ret_value, "False", "Test Failed while trying to add member to the cloud account")
	//assert.NotEqual(suite.T(), get_call, "False", "Test Failed while trying to get members of cloud account")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}
