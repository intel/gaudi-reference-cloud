//go:build Functional || Billing || Regression
// +build Functional Billing Regression

package BillingAPITest

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"os"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// var met_ret bool

func (suite *BillingAPITestSuite) TestIntelCreateFreeInstanceWithoutCredits() {
	suite.T().Skip()
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	// token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	computeUrl := utils.Get_Compute_Base_Url()
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	//paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	//assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Now launch paid instance and see API throws 403 error

	//token := utils.Get_intel_Token()
	logger.Log.Info("Compute Url : " + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)
	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Intel user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	VMName := gjson.Get(create_response_body, "metadata.name").String()
	if create_response_status == 429 || create_response_status == 403 {
		if create_response_status == 403 {
			message := gjson.Get(create_response_body, "message").String()
			if message == "product not found" {
				logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
				suite.T().Skip()
			}
		} else {
			logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
			suite.T().Skip()
		}

	}

	assert.Equal(suite.T(), create_response_status, 200, "Failed: Failed to create VM instance")
	assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) TestaIntelCreatePaidInstanceWithoutCredits() {
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	computeUrl := utils.Get_Compute_Base_Url()
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"

	base_url := utils.Get_Base_Url1()
	url := base_url + "/v1/cloudaccounts"

	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	// paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	// lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	// assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	// assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Now launch paid instance and see API throws 403 error

	//token := utils.Get_intel_Token()
	logger.Log.Info("Compute Url : " + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)

	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Intel user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-med", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)
	fmt.Println("instance_endpoint", instance_endpoint)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	//VMName := gjson.Get(create_response_body, "metadata.name").String()
	if create_response_status == 429 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
		suite.T().Skip()
	}
	assert.Equal(suite.T(), create_response_status, 403, "Failed: Failed to create VM instance")
	//assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) TestIntelCreateFreeInstanceWithCouponCredits() {
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	suite.T().Skip()
	os.Setenv("intel_user_test", "True")
	//Create Cloud Account
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	computeUrl := utils.Get_Compute_Base_Url()
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	// paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	// lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	// assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	// assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Add coupon to the user
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.CreateCouponStruct{
		Amount:  300,
		Creator: "idc_billing@intel.com",
		Expires: expirationtime,
		Start:   creation_time,
		NumUses: 2,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode, 200)

	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get on Coupon Failed")

	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Intel cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	// Now launch paid instance and see API throws 403 error

	//token := utils.Get_intel_Token()
	logger.Log.Info("Compute Url : " + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)
	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Intel user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	if create_response_status == 429 || create_response_status == 403 {
		if create_response_status == 403 {
			message := gjson.Get(create_response_body, "message").String()
			if message == "product not found" {
				logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
				suite.T().Skip()
			}
		} else {
			logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
			suite.T().Skip()
		}

	}

	// hit the api

	VMName := gjson.Get(create_response_body, "metadata.name").String()

	assert.Equal(suite.T(), create_response_status, 200, "Failed: Failed to create VM instance")
	assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

func (suite *BillingAPITestSuite) TestIntelCreatePaidInstanceWithCouponCredits() {
	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	//Create Cloud Account
	computeUrl := utils.Get_Compute_Base_Url()
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	// paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	// lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	// assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	// assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Add coupon to the user
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.CreateCouponStruct{
		Amount:  300,
		Creator: "idc_billing@intel.com",
		Expires: expirationtime,
		Start:   creation_time,
		NumUses: 2,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode, 200)

	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get on Coupon Failed")

	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Intel cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	// Now launch paid instance and see API throws 403 error

	//token := utils.Get_intel_Token()
	logger.Log.Info("Compute Url : " + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)

	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Intel user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-med", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
		suite.T().Skip()

	}
	VMName := gjson.Get(create_response_body, "metadata.name").String()
	assert.Equal(suite.T(), create_response_status, 200, "Failed: Failed to create VM instance")
	assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Intel user)")

}

// Premium User

func (suite *BillingAPITestSuite) TestPremiumCreateFreeInstanceWithoutCredits() {
	logger.Log.Info("Starting Test : Launch Paid instance without coupon or credit card")
	// Enroll customer
	suite.T().Skip()
	logger.Log.Info("Creating a Premium User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "premimum1-eid1"
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)

	computeUrl := utils.Get_Compute_Base_Url()
	logger.Log.Info("Compute Url" + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	url := base_url + "/v1/cloudaccounts"
	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}
	time.Sleep(1 * time.Minute)

	userToken = "Bearer " + userToken

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Now launch paid instance and see API throws 403 error

	upgrade_err := billing.Upgrade_to_Premium_with_coupon(get_CAcc_id, authToken, userToken, int64(1))
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	//token := utils.Get_intel_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)
	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Premium user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)

	if create_response_status == 429 || create_response_status == 403 {
		if create_response_status == 403 {
			message := gjson.Get(create_response_body, "message").String()
			if message == "product not found" {
				logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
				suite.T().Skip()
			}
		} else {
			logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
			suite.T().Skip()
		}

	}

	VMName := gjson.Get(create_response_body, "metadata.name").String()
	assert.Equal(suite.T(), create_response_status, 200, "Failed: Failed to create VM instance")
	assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

// func (suite *BillingAPITestSuite) TestPremiumCreatePaidInstanceWithoutCredits() {
// 	logger.Log.Info("Starting Test : Launch Paid instance without coupon or credit card")
// 	// Enroll customer
// 	logger.Log.Info("Creating a Premium User Cloud Account using Enroll API")
// 	tid := cloudAccounts.Rand_token_payload_gen()
// 	enterpriseId := "premimum1-eid1"
// 	userName := utils.Get_UserName("Premium")
// 	userToken, _ := auth.Get_Azure_Bearer_Token(userName)

// 	computeUrl := utils.Get_Compute_Base_Url()
// 	logger.Log.Info("Compute Url" + computeUrl)
// 	//baseUrl := utils.Get_Base_Url1()
// 	// Create an ssh key  for the user
// 	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
// 	base_url := utils.Get_Base_Url1()
// 	url := base_url + "/v1/cloudaccounts"
// 	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
// 	if err == nil {
// 		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
// 	}
// 	time.Sleep(1 * time.Minute)

// 	userToken = "Bearer " + userToken

// 	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
// 	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
// 	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
// 	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

// 	// Now launch paid instance and see API throws 403 error

// 	upgrade_err := billing.Upgrade_to_Premium_with_coupon(get_CAcc_id, authToken, userToken, int64(1))
// 	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

// 	// Now launch paid instance and see API throws 403 error

// 	fmt.Println("Starting the SSH-Public-Key Creation via API...")
// 	// form the endpoint and payload
// 	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
// 	sshPublicKey := utils.GetSSHKey()
// 	fmt.Println("SSH key is" + sshPublicKey)
// 	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
// 	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
// 	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
// 	// hit the api
// 	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)

// 	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
// 	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
// 	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

// 	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
// 	vnet_name := compute_utils.GetVnetName()
// 	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
// 	// hit the api
// 	fmt.Println("Vnet end point ", vnet_endpoint)
// 	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
// 	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
// 	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Premium user")

// 	fmt.Println("Starting the Instance Creation via API...")
// 	// form the endpoint and payload
// 	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
// 	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
// 	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-med", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
// 	fmt.Println("instance_payload", instance_payload)

// 	// hit the api
// 	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
// 	logger.Log.Info("create_response_body" + create_response_body)

// 	if create_response_status == 429 {
// 		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
// 		suite.T().Skip()

// 	}
// 	//VMName := gjson.Get(create_response_body, "metadata.name").String()
// 	assert.Equal(suite.T(), create_response_status, 403, "Failed: Failed to create VM instance")
// 	//assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

// 	if create_response_status == 200 {
// 		time.Sleep(10 * time.Second)
// 		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
// 		// delete the instance created
// 		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
// 		time.Sleep(10 * time.Second)
// 	}

// 	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
// 	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

// }

func (suite *BillingAPITestSuite) TestPremiumCreateFreeInstanceWithCouponCredits() {
	logger.Log.Info("Starting Test : Launch Paid instance without coupon or credit card")
	// Enroll customer
	suite.T().Skip()
	logger.Log.Info("Creating a Premium User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "premimum1-eid1"
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)

	computeUrl := utils.Get_Compute_Base_Url()
	logger.Log.Info("Compute Url" + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	url := base_url + "/v1/cloudaccounts"
	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}
	time.Sleep(1 * time.Minute)

	userToken = "Bearer " + userToken

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Now launch paid instance and see API throws 403 error

	upgrade_err := billing.Upgrade_to_Premium_with_coupon(get_CAcc_id, authToken, userToken, int64(300))
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	// Add coupon to the user

	// Now launch paid instance and see API throws 403 error

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)
	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Premium user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)

	if create_response_status == 429 || create_response_status == 403 {
		if create_response_status == 403 {
			message := gjson.Get(create_response_body, "message").String()
			if message == "product not found" {
				logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
				suite.T().Skip()
			}
		} else {
			logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
			suite.T().Skip()
		}

	}

	VMName := gjson.Get(create_response_body, "metadata.name").String()
	assert.Equal(suite.T(), create_response_status, 200, "Failed: Failed to create VM instance")
	assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) TestPremiumCreatePaidInstanceWithCouponCredits() {
	logger.Log.Info("Starting Test : Launch Paid instance without coupon or credit card")
	// Enroll customer
	logger.Log.Info("Creating a Premium User Cloud Account using Enroll API")
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "premimum1-eid1"
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)

	computeUrl := utils.Get_Compute_Base_Url()
	logger.Log.Info("Compute Url" + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	url := base_url + "/v1/cloudaccounts"
	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}
	time.Sleep(1 * time.Minute)

	userToken = "Bearer " + userToken

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, _ := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Now launch paid instance and see API throws 403 error

	upgrade_err := billing.Upgrade_to_Premium_with_coupon(get_CAcc_id, authToken, userToken, int64(300))
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	// Add coupon to the user

	// Now launch paid instance and see API throws 403 error

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)

	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Premium user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-med", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("create_response_body" + create_response_body)

	if create_response_status == 429 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
		suite.T().Skip()
	}

	VMName := gjson.Get(create_response_body, "metadata.name").String()
	assert.Equal(suite.T(), create_response_status, 200, "Failed: Failed to create VM instance")
	assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) TestPremiumCreateFreeInstanceWithCreditCard() {
	logger.Logf.Infof("Skipping Test because credit card is not enabled")
	suite.T().Skip()
	token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	computeUrl := utils.Get_Compute_Base_Url()
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	logger.Log.Info("Compute Url" + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	url := base_url + "/v1/cloudaccounts"
	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}
	time.Sleep(1 * time.Minute)

	// cloud account creation
	creditCardPayload := utils.Get_CC_payload("visaCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	consoleUrl := utils.Get_Console_Base_Url()
	replaceUrl := utils.Get_CC_Replace_Url()
	_, _, _, _, _, password, _, _, _ := auth.Get_Azure_auth_data_from_config(userName)
	financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	fmt.Println("bilingOptions1", bilingOptions1)
	fmt.Println("bilingOptions ", bilingOptions)
	//Expect(bilingOptions).To(Equal(200), "Failed to create VNet")
	// Expect(gjson.Get(bilingOptions1, "creditCard.suffix").String()).To(Equal("1111"), "Failed to validate credit card details")
	// Expect(gjson.Get(bilingOptions1, "creditCard.expiration").String()).To(Equal("10/2026"), "Failed to validate credit card details")
	// Expect(gjson.Get(bilingOptions1, "creditCard.type").String()).To(Equal("Visa"), "Failed to validate credit card details")
	// Expect(gjson.Get(bilingOptions1, "paymentType").String()).To(Equal("PAYMENT_CREDIT_CARD"), "Failed to validate credit card details")
	// Expect(gjson.Get(bilingOptions1, "email").String()).To(Equal("testuser@premium.com"), "Failed to validate credit card details")
	// Expect(gjson.Get(bilingOptions1, "firstName").String()).To(Equal("Test"), "Failed to validate credit card details")
	// Expect(gjson.Get(bilingOptions1, "lastName").String()).To(Equal("User"), "Failed to validate credit card details")
	// Expect(gjson.Get(bilingOptions1, "address1").String()).To(Equal("Intel Technologies"), "Failed to validate credit card details")
	// Expect(gjson.Get(bilingOptions1, "city").String()).To(Equal("Bangalore"), "Failed to validate credit card details")
	// Expect(gjson.Get(bilingOptions1, "cloudAccountId").String()).To(Equal(cloudaccId), "Failed to validate credit card details")

	get_CAcc_id := cloudaccId

	// Now launch paid instance and see API throws 403 error

	//token := utils.Get_intel_Token()
	//computeUrl := utils.Get_Compute_Base_Url()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)
	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Premium user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)

	if create_response_status == 429 || create_response_status == 403 {
		if create_response_status == 403 {
			message := gjson.Get(create_response_body, "message").String()
			if message == "product not found" {
				logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
				suite.T().Skip()
			}
		} else {
			logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
			suite.T().Skip()
		}

	}
	VMName := gjson.Get(create_response_body, "metadata.name").String()

	assert.Equal(suite.T(), create_response_status, 200, "Failed: Failed to create VM instance")
	assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")

}

// Standard User

func (suite *BillingAPITestSuite) TestStandardCreateFreeInstanceWithoutCredits() {
	suite.T().Skip()
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	computeUrl := utils.Get_Compute_Base_Url()
	//base_url := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	//paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	//assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "IN", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Now launch paid instance and see API throws 403 error

	//token := utils.Get_Standard_Token()
	logger.Log.Info("Compute Url" + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)
	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Standard user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	VMName := gjson.Get(create_response_body, "metadata.name").String()

	if create_response_status == 429 || create_response_status == 403 {
		if create_response_status == 403 {
			message := gjson.Get(create_response_body, "message").String()
			if message == "product not found" {
				logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
				suite.T().Skip()
			}
		} else {
			logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
			suite.T().Skip()
		}

	}

	assert.Equal(suite.T(), create_response_status, 200, "Failed: Failed to create VM instance")
	assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) TestAStandardCreatePaidInstanceWithoutCredits() {
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	computeUrl := utils.Get_Compute_Base_Url()
	base_url := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()

	url := base_url + "/v1/cloudaccounts"
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"
	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}
	time.Sleep(1 * time.Minute)
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "IN", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Now launch paid instance and see API throws 403 error

	time.Sleep(1 * time.Minute)
	//token := utils.Get_Standard_Token()
	logger.Log.Info("Compute Url" + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)

	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Standard user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("create_response_body" + create_response_body)

	if create_response_status == 429 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
		suite.T().Skip()
	}

	//VMName := gjson.Get(create_response_body, "metadata.name").String()
	assert.Equal(suite.T(), create_response_status, 403, "Failed: Failed to create VM instance")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) TestStandardCreateFreeInstanceWithCouponCredits() {
	suite.T().Skip()
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	computeUrl := utils.Get_Compute_Base_Url()
	//base_url := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "IN", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Add coupon to the user
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    2,
		IsStandard: true,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode, 200)

	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get on Coupon Failed")

	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Standard cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	// Now launch paid instance and see API throws 403 error

	//token := utils.Get_Standard_Token()
	logger.Log.Info("Compute Url" + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)
	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Standard user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	VMName := gjson.Get(create_response_body, "metadata.name").String()

	if create_response_status == 429 || create_response_status == 403 {
		if create_response_status == 403 {
			message := gjson.Get(create_response_body, "message").String()
			if message == "product not found" {
				logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
				suite.T().Skip()
			}
		} else {
			logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
			suite.T().Skip()
		}

	}

	assert.Equal(suite.T(), create_response_status, 200, "Failed: Failed to create VM instance")
	assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}

func (suite *BillingAPITestSuite) TestStandardCreatePaidInstanceWithCouponCredits() {
	os.Setenv("standard_user_test", "True")
	logger.Log.Info("Creating a enterprise User Cloud Account using Enroll API")
	userName := utils.Get_UserName("Standard")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	computeUrl := utils.Get_Compute_Base_Url()
	base_url := utils.Get_Base_Url1()
	tid := cloudAccounts.Rand_token_payload_gen()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	enterpriseId := "testeid-01"

	url := base_url + "/v1/cloudaccounts"

	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}
	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("standard", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an Standard user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_STANDARD", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	//lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "IN", CountryCode, "Failed: Validation Failed in Coupon Redemption, Get on cloud account for paidServices")
	//assert.Equal(suite.T(), "false", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

	// Add coupon to the user
	creation_time, expirationtime := billing.GetCreationExpirationTime()
	createCoupon := billing.StandardCreateCouponStruct{
		Amount:     300,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    2,
		IsStandard: true,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := billing.CreateCoupon(req, 200)
	couponCode := gjson.Get(data, "code").String()
	assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(data, "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
	assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(data, "creator").String(), "Failed: Create Coupon Failed to validate Creator")
	assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(data, "expires").String(), "Failed: Create Coupon Failed to validate Expires")
	assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(data, "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), createCoupon.Start, gjson.Get(data, "start").String(), "Failed: Create Coupon Failed to validate numUses")
	assert.Equal(suite.T(), ret_value, true, "Failed: Create Coupon Failed")
	// Get coupon and validate
	getret_value, getdata := billing.GetCoupons(couponCode, 200)

	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), createCoupon.Amount, gjson.Get(val.String(), "amount").Int(), "Failed: Create Coupon Failed to validate Amount")
		assert.Equal(suite.T(), createCoupon.Creator, gjson.Get(val.String(), "creator").String(), "Failed: Create Coupon Failed to validate Creator")
		assert.Equal(suite.T(), createCoupon.Expires, gjson.Get(val.String(), "expires").String(), "Failed: Create Coupon Failed to validate Expires")
		assert.Equal(suite.T(), createCoupon.NumUses, gjson.Get(val.String(), "numUses").Int(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), createCoupon.Start, gjson.Get(val.String(), "start").String(), "Failed: Create Coupon Failed to validate numUses")
		assert.Equal(suite.T(), "0", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}
	assert.Equal(suite.T(), getret_value, true, "Failed: Get on Coupon Failed")

	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Standard cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := billing.RedeemCouponStruct{
		CloudAccountID: get_CAcc_id,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ret_value, data = billing.RedeemCoupon(req, 200)

	// Get coupon and validate
	getret_value, getdata = billing.GetCoupons(couponCode, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		assert.Equal(suite.T(), get_CAcc_id, gjson.Get(val.String(), "cloudAccountId").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), couponCode, gjson.Get(val.String(), "code").String(), "Failed: Validation Failed in Coupon Redemption")
		assert.Equal(suite.T(), "true", gjson.Get(val.String(), "installed").String(), "Failed: Validation Failed in Coupon Redemption")
	}
	for _, val := range couponData.Array() {
		assert.Equal(suite.T(), "1", gjson.Get(val.String(), "numRedeemed").String(), "Failed: Create Coupon Failed to validate numUses")
	}

	// Now launch paid instance and see API throws 403 error

	//token := utils.Get_Standard_Token()
	logger.Log.Info("Compute Url" + computeUrl)
	//baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)

	assert.Equal(suite.T(), sshkey_creation_status, 200, "Failed: Failed to create SSH Public key")
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), sshkey_name, ssh_publickey_name_created, "Failed: Failed to create SSH Public key, response validation failed")

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Standard user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("create_response_body" + create_response_body)

	if create_response_status == 429 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", create_response_body)
		suite.T().Skip()
	}

	VMName := gjson.Get(create_response_body, "metadata.name").String()
	assert.Equal(suite.T(), create_response_status, 200, "Failed: Failed to create VM instance")
	assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

	if create_response_status == 200 {
		time.Sleep(10 * time.Second)
		instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
		// delete the instance created
		_, _ = frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		time.Sleep(10 * time.Second)
	}

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Standard user)")

}
