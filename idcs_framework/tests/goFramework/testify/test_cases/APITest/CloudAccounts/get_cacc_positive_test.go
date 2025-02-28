//go:build All || Functional || GetCloudAccounts || Positive || Regression
// +build All Functional GetCloudAccounts Positive Regression

package CaAPITest

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/financials/cloudAccounts"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"goFramework/utils"
)

// IDC_1.0_CAccSvc_GetById
func (suite *CaAPITestSuite) Test_CAccSvc_GetById() {
	logger.Log.Info("Retrieve the cloud account via GET method using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false, true,
		false, "ACCOUNT_TYPE_INTEL", 200)
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAccSvc_GetByName
func (suite *CaAPITestSuite) Test_CAccSvc_GetByName() {
	logger.Log.Info("Retrieve the cloud account via GET method using name")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false, false,
		false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating the cloud account")
	ret_value, _ := cloudAccounts.GetCAccByName(name, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed to get cloudaccount by name")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_Membership_AddMembers
func (suite *CaAPITestSuite) Test_CAcc_AddMembers_withId() {
	logger.Log.Info("Add Members to Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false, true,
		true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	ret_value, _ := cloudAccounts.AddMemberstoCAcc(1, get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to add member to the cloud account")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAccMemberSvc_GetById
func (suite *CaAPITestSuite) Test_CAccMemberSvc_GetById() {
	logger.Log.Info("Retrieving Members of a Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false, true,
		true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	ret_value, _ := cloudAccounts.GetCAccMembersById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to get members of the cloud account using Id")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_Delete
func (suite *CaAPITestSuite) Test_CAcc_Delete() {
	logger.Log.Info("Creating an premium User Cloud Account")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false, true,
		true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAccSvc_GetAcType_1
func (suite *CaAPITestSuite) Test_CAccSvc_GetAcType_1() {
	logger.Log.Info("Retrieve the cloud account via GET method using account type - STANDARD")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, create_payload := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating the cloud account")
	CAcc_Type := create_payload.Type
	ret_value, _ := cloudAccounts.GetCAccByAccType(CAcc_Type, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed to Get CloudAccounts with account type 1")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAccSvc_GetAcType_2
func (suite *CaAPITestSuite) Test_CAccSvc_GetAcType_2() {
	logger.Log.Info("Retrieve the cloud account via GET method using account type - PREMIUM")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, create_payload := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating the cloud account")
	CAcc_Type := create_payload.Type
	ret_value, _ := cloudAccounts.GetCAccByAccType(CAcc_Type, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed to Get CloudAccounts with account type 2")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAccSvc_GetAcType_3
func (suite *CaAPITestSuite) Test_CAccSvc_GetAcType_3() {
	logger.Log.Info("Retrieve the cloud account via GET method using account type - ENTERPRISE PENDING")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, create_payload := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_ENTERPRISE_PENDING", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating the cloud account")
	CAcc_Type := create_payload.Type
	ret_value, _ := cloudAccounts.GetCAccByAccType(CAcc_Type, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed to Get CloudAccounts with account type 3")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAccSvc_GetAcType_4
func (suite *CaAPITestSuite) Test_CAccSvc_GetAcType_4() {
	logger.Log.Info("Retrieve the cloud account via GET method using account type - ENTERPRISE")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, create_payload := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_ENTERPRISE", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating the cloud account")
	CAcc_Type := create_payload.Type
	ret_value, _ := cloudAccounts.GetCAccByAccType(CAcc_Type, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed to Get CloudAccounts with account type 4")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAccSvc_GetAcType_5
func (suite *CaAPITestSuite) Test_CAccSvc_GetAcType_5() {
	logger.Log.Info("Retrieve the cloud account via GET method using account type - INTEL")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, create_payload := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating the cloud account")
	CAcc_Type := create_payload.Type
	ret_value, _ := cloudAccounts.GetCAccByAccType(CAcc_Type, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed to Get CloudAccounts with account type 5")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAccSvc_GetAcType_PUEnrolled
func (suite *CaAPITestSuite) Test_CAccSvc_GetAcType_PUEnrolled() {
	logger.Log.Info("Retrieve the cloud account via GET method using account type - PREMIUM")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, create_payload := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating the cloud account")
	CAcc_Type := create_payload.Type
	CAcc_enroll_value := create_payload.Enrolled
	ret_value, _ := cloudAccounts.GetCAccPUEnrolled(CAcc_Type, CAcc_enroll_value, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed to Get cloud account with enrolled value")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// Commenting this TC as development is not done yet on retrieving members using username.
// IDC_1.0_CAccMemberSvc_GetByName
// func (suite *CaAPITestSuite) Test_CAccMemberSvc_GetByName() {
// 	logger.Log.Info("Retrieving Members of a Cloud Account using name")
// 	get_CAcc_id, create_payload := cloudAccounts.CreateCloudAccount("ACCOUNT_TYPE_PREMIUM", 200)
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
// 	CAcc_name := create_payload.Name
// 	ret_value, _ := cloudAccounts.GetCAccMembersById(CAcc_name, 200)
// 	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to get cloud account members using name")
// 	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
// }

// IDC_1.0_CAcc_Membership_RemoveMembers
func (suite *CaAPITestSuite) Test_CAcc_RemoveMembers() {
	logger.Log.Info("Removing members of Cloud Account")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account")
	ret_value, create_payload := cloudAccounts.AddMemberstoCAcc(1, get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while adding member to cloud account")
	CAcc_addedMember := create_payload.Members
	remove_CAcc, _ := cloudAccounts.DeleteMembersofCAcc(get_CAcc_id, CAcc_addedMember, 200)
	assert.NotEqual(suite.T(), remove_CAcc, "False", "Test Failed while removing a member from cloud account")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_Getaccount_by_owner
func (suite *CaAPITestSuite) Test_CAcc_FilterByOwner() {
	logger.Log.Info("Retrieving the Cloud Account using owner")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	ret_value, _ := cloudAccounts.GetCAccByOwner(owner, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to get cloud account by owner")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_Filter
func (suite *CaAPITestSuite) Test_CAcc_Filter() {
	logger.Log.Info("Retrieving the Cloud Account using any string")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	ret_value, _ := cloudAccounts.GetCAccByAnyString(parentid, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to get cloud account by any string")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_GetBytid
func (suite *CaAPITestSuite) Test_CAccSvc_GetBytid() {
	logger.Log.Info("Retrieving Cloud Account using tid")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	ret_value, _ := cloudAccounts.GetCAccByTid(tid, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to get cloud account by tid")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_GetByoid
func (suite *CaAPITestSuite) Test_CAccSvc_GetByoid() {
	logger.Log.Info("Retrieving Cloud Account using oid")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	ret_value, _ := cloudAccounts.GetCAccByOid(oid, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to get cloud account by oid")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_Getaccount_by_tid_oid
func (suite *CaAPITestSuite) Test_CAccSvc_GetByTid_Oid() {
	logger.Log.Info("Retrieving Cloud Account using Tid and Oid")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	ret_value, _ := cloudAccounts.GetCAccByTid_Oid(tid, oid, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to get cloud account by tid and oid")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_check_if_billingaccount_created_for_PU
func (suite *CaAPITestSuite) Test_CAccSvc_check_BillingAcc_PU_with_Id() {
	logger.Log.Info("Checking if billing account is created for premium user cloud account")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	check_Billing, _ := cloudAccounts.CAcc_check_BillingAcc(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), check_Billing, "False", "Test Failed while checking billing account for Cloud account")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_check_if_billingaccount_created_for_EU
func (suite *CaAPITestSuite) Test_CAccSvc_check_BillingAcc_EU_with_Id() {
	logger.Log.Info("Checking if billing account is created for enterprise user cloud account")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_ENTERPRISE", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	check_Billing, _ := cloudAccounts.CAcc_check_BillingAcc(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), check_Billing, "False", "Test Failed while checking billing account for Cloud account")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_check_if_billingaccount_created_for_SU
func (suite *CaAPITestSuite) Test_CAccSvc_check_BillingAcc_SU_with_Id() {
	logger.Log.Info("Checking if billing account is created for standard user cloud account")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating cloud account")
	check_Billing, _ := cloudAccounts.CAcc_check_BillingAcc(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), check_Billing, "False", "Test Failed while checking billing account for Cloud account")
	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_Update_name_ById
func (suite *CaAPITestSuite) Test_Update_CAcc_Name_ById() {
	logger.Log.Info("Update Name of Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account")
	ret_value, _ := cloudAccounts.UpdateCAccById("ValidPayload1", get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to update name of the cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_CAcc_Update_owner_ById
func (suite *CaAPITestSuite) Test_Update_CAcc_owner_ById() {
	logger.Log.Info("Update owner of Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_PREMIUM", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a premium user cloud account")
	ret_value, _ := cloudAccounts.UpdateCAccById("ValidPayload2", get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to update owner of the cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_Update_terminateMessageQueued_by_Id
func (suite *CaAPITestSuite) Test_Update_terminateMessageQueued_by_Id() {
	logger.Log.Info("Update terminate messaged queued alert for Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, true, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	ret_value, _ := cloudAccounts.UpdateCAccById("ValidPayload3", get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to update messaged queued alert for cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_Update_terminatePaidServices_by_Id
func (suite *CaAPITestSuite) Test_Update_terminatePaidServices_by_Id() {
	logger.Log.Info("Update terminate paid services alert for Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, true,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	ret_value, _ := cloudAccounts.UpdateCAccById("ValidPayload4", get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to update terminate paid services alert for cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_Update_lowCredits_by_Id
func (suite *CaAPITestSuite) Test_Update_lowCredits_by_Id() {
	logger.Log.Info("Update low credits alert for Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, true, false, true, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	ret_value, _ := cloudAccounts.UpdateCAccById("ValidPayload5", get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to update low credits alert for cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_Update_parentId_by_Id
func (suite *CaAPITestSuite) Test_Update_parentId_by_Id() {
	logger.Log.Info("Update parent id for Cloud Account using id")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	ret_value, _ := cloudAccounts.UpdateCAccById("ValidPayload6", get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value, "False", "Test Failed while trying to update parent id for cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}

// IDC_1.0_Update_nameby_create_CAcc_with_duplicate_oid
func (suite *CaAPITestSuite) Test_UpdateNameBy_create_CAcc_with_duplicate_oid() {
	logger.Log.Info("Update cloud account name by creating cloud account using duplicate oid")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "standard1-eid1"
	userName := utils.Get_UserName("Standard")
	get_CAcc_id, _, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account using Enroll API")
	tid1 := cloudAccounts.Rand_token_payload_gen()
	userName1 := utils.Get_UserName("Standard")
	get_CAcc_id1, _, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid1, userName1, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id1, "False", "Test Failed while creating a cloud account using duplicate oid")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id1, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the first cloud account")
}

// IDC_1.0_CAcc_check_with_Ensure
func (suite *CaAPITestSuite) Test_CAcc_check_with_Ensure() {
	logger.Log.Info("Checking cloud account being created using ensure API")
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, req_payload := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, true, false, true, false,
		false, false, "ACCOUNT_TYPE_STANDARD", 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating a standard user cloud account")
	cacc_ensure, responsePayload := cloudAccounts.CAccEnsure(req_payload, 200)
	assert.NotEqual(suite.T(), cacc_ensure, "False", "Test Failed while checking with ensure API")
	cloudacc_id := gjson.Get(responsePayload, "id").String()
	assert.Equal(suite.T(), get_CAcc_id, cloudacc_id, "Test Failed as ensure API failed to get same cloud accounts details")
	ret_value2, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value2, "False", "Test Failed while deleting the cloud account")
}
