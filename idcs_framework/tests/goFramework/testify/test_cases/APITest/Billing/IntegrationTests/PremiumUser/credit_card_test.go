//go:build Functional || CreditCardUI
// +build Functional CreditCardUI

package PremiumBillingAPITest

import (
	"fmt"
	_ "fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) TestPremiumCreateFreeInstanceWithVisaCard() {
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
	err = financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	if err != nil {
		logger.Logf.Infof("Skipping Test because credit card addition failed")
		suite.T().Skip()
	}
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "1111", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "10/2028", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "Visa", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testuservisa@premium.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Visa", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Visa User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Visa User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

	get_CAcc_id := cloudaccId

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
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Intel user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("instance_payload" + instance_payload)

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

	logger.Log.Info("create_response_body" + create_response_body)
	//VMName := gjson.Get(create_response_body, "metadata.name").String()
	assert.Equal(suite.T(), create_response_status, 200, "Failed: Created Free instances after using all credits")
	//assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

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

func (suite *BillingAPITestSuite) TestPremiumCreatePaidInstanceWithVisaCard() {
	token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

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

	// cloud account creation
	creditCardPayload := utils.Get_CC_payload("visaCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	consoleUrl := utils.Get_Console_Base_Url()
	replaceUrl := utils.Get_CC_Replace_Url()
	_, _, _, _, _, password, _, _, _ := auth.Get_Azure_auth_data_from_config(userName)
	err = financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	if err != nil {
		logger.Logf.Infof("Skipping Test because credit card addition failed")
		suite.T().Skip()
	}
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "1111", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "10/2028", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "Visa", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testuservisa@premium.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Visa", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Visa User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Visa User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

	get_CAcc_id := cloudaccId

	// Now launch paid instance and see API throws 403 error

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

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) TestPremiumCreateFreeInstanceWithMasterCard() {
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
	creditCardPayload := utils.Get_CC_payload("masterCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	consoleUrl := utils.Get_Console_Base_Url()
	replaceUrl := utils.Get_CC_Replace_Url()
	_, _, _, _, _, password, _, _, _ := auth.Get_Azure_auth_data_from_config(userName)
	err = financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	if err != nil {
		logger.Logf.Infof("Skipping Test because credit card addition failed")
		suite.T().Skip()
	}
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "5454", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "11/2029", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "MasterCard", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testusermaster@intel.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Master", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Master User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "paymentType").String(), "PAYMENT_CREDIT_CARD", "Failed to validate payment type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

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
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Intel user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("instance_payload" + instance_payload)

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

	logger.Log.Info("create_response_body" + create_response_body)
	//VMName := gjson.Get(create_response_body, "metadata.name").String()
	assert.Equal(suite.T(), create_response_status, 200, "Failed: Created Free instances after using all credits")
	//assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

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

func (suite *BillingAPITestSuite) TestPremiumCreatePaidInstanceWithMasterCard() {
	token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

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

	// cloud account creation
	creditCardPayload := utils.Get_CC_payload("masterCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	consoleUrl := utils.Get_Console_Base_Url()
	replaceUrl := utils.Get_CC_Replace_Url()
	_, _, _, _, _, password, _, _, _ := auth.Get_Azure_auth_data_from_config(userName)
	err = financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	if err != nil {
		logger.Logf.Infof("Skipping Test because credit card addition failed")
		suite.T().Skip()
	}
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "5454", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "11/2029", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "MasterCard", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testusermaster@intel.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Master", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Master User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "paymentType").String(), "PAYMENT_CREDIT_CARD", "Failed to validate payment type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

	get_CAcc_id := cloudaccId

	// Now launch paid instance and see API throws 403 error

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

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) TestPremiumCreateFreeInstanceWithDiscoverCard() {
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
	creditCardPayload := utils.Get_CC_payload("discoverCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	consoleUrl := utils.Get_Console_Base_Url()
	replaceUrl := utils.Get_CC_Replace_Url()
	_, _, _, _, _, password, _, _, _ := auth.Get_Azure_auth_data_from_config(userName)
	err = financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	if err != nil {
		logger.Logf.Infof("Skipping Test because credit card addition failed")
		suite.T().Skip()
	}
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "0000", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "12/2030", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "Discover", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testuserdiscover@intel.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Discover", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Discover User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "paymentType").String(), "PAYMENT_CREDIT_CARD", "Failed to validate payment type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

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
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Intel user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("instance_payload" + instance_payload)

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

	logger.Log.Info("create_response_body" + create_response_body)
	//VMName := gjson.Get(create_response_body, "metadata.name").String()
	assert.Equal(suite.T(), create_response_status, 200, "Failed: Created Free instances after using all credits")
	//assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

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

func (suite *BillingAPITestSuite) TestPremiumCreatePaidInstanceWithDiscoverCard() {
	token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

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

	// cloud account creation
	creditCardPayload := utils.Get_CC_payload("discoverCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	consoleUrl := utils.Get_Console_Base_Url()
	replaceUrl := utils.Get_CC_Replace_Url()
	_, _, _, _, _, password, _, _, _ := auth.Get_Azure_auth_data_from_config(userName)
	err = financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	if err != nil {
		logger.Logf.Infof("Skipping Test because credit card addition failed")
		suite.T().Skip()
	}
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "0000", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "12/2030", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "Discover", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testuserdiscover@intel.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Discover", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Discover User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "paymentType").String(), "PAYMENT_CREDIT_CARD", "Failed to validate payment type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

	get_CAcc_id := cloudaccId

	// Now launch paid instance and see API throws 403 error

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

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) TestPremiumCreateFreeInstanceWithAmexCard() {
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
	creditCardPayload := utils.Get_CC_payload("amexCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	consoleUrl := utils.Get_Console_Base_Url()
	replaceUrl := utils.Get_CC_Replace_Url()
	_, _, _, _, _, password, _, _, _ := auth.Get_Azure_auth_data_from_config(userName)
	err = financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	if err != nil {
		logger.Logf.Infof("Skipping Test because credit card addition failed")
		suite.T().Skip()
	}
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "8431", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "10/2031", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "Amex", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testuseramex@intel.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Amex", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Amex User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "paymentType").String(), "PAYMENT_CREDIT_CARD", "Failed to validate payment type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

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
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Intel user")

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + get_CAcc_id + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-tny", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("instance_payload" + instance_payload)

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

	logger.Log.Info("create_response_body" + create_response_body)
	//VMName := gjson.Get(create_response_body, "metadata.name").String()
	assert.Equal(suite.T(), create_response_status, 200, "Failed: Created Free instances after using all credits")
	//assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")

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

func (suite *BillingAPITestSuite) TestPremiumCreatePaidInstanceWithAmexCard() {
	token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

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

	// cloud account creation
	creditCardPayload := utils.Get_CC_payload("amexCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	consoleUrl := utils.Get_Console_Base_Url()
	replaceUrl := utils.Get_CC_Replace_Url()
	_, _, _, _, _, password, _, _, _ := auth.Get_Azure_auth_data_from_config(userName)
	err = financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	if err != nil {
		logger.Logf.Infof("Skipping Test because credit card addition failed")
		suite.T().Skip()
	}
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "8431", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "10/2031", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "Amex", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testuseramex@intel.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Amex", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Amex User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "paymentType").String(), "PAYMENT_CREDIT_CARD", "Failed to validate payment type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

	get_CAcc_id := cloudaccId

	// Now launch paid instance and see API throws 403 error

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

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) TestPremiumChangeCreditCardLaunchPaidInstance() {
	token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

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

	// cloud account creation
	creditCardPayload := utils.Get_CC_payload("visaCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	consoleUrl := utils.Get_Console_Base_Url()
	replaceUrl := utils.Get_CC_Replace_Url()
	_, _, _, _, _, password, _, _, _ := auth.Get_Azure_auth_data_from_config(userName)
	err = financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	if err != nil {
		logger.Logf.Infof("Skipping Test because credit card addition failed")
		suite.T().Skip()
	}
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	get_CAcc_id := cloudaccId
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "1111", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "10/2028", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "Visa", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testuservisa@premium.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Visa", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Visa User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "paymentType").String(), "PAYMENT_CREDIT_CARD", "Failed to validate payment type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

	// Now Change card to master card and launch instance

	creditCardPayload = utils.Get_CC_payload("masterCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	financials.ChangeCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	bilingOptions, bilingOptions1 = financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "5454", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "11/2029", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "MasterCard", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testusermaster@intel.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Master", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Master User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "paymentType").String(), "PAYMENT_CREDIT_CARD", "Failed to validate payment type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

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

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(get_CAcc_id, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium user)")
}

func (suite *BillingAPITestSuite) Test_Premium_Generate_Invoice_Using_Credit_Card() {
	// Standard user is already enrolled, so start upgrade
	token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken

	ariaclientId, ariaAuth := utils.Get_Aria_Config()
	logger.Logf.Infof("Aria details, ariaclientId", ariaclientId)
	logger.Logf.Infof("Aria details, ariaAuth", ariaAuth)

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

	// cloud account creation
	creditCardPayload := utils.Get_CC_payload("visaCard")
	fmt.Println("Credit card payload : ", creditCardPayload)
	consoleUrl := utils.Get_Console_Base_Url()
	replaceUrl := utils.Get_CC_Replace_Url()
	_, _, _, _, _, password, _, _, _ := auth.Get_Azure_auth_data_from_config(userName)
	err = financials.EnrollPremiumUserWithCreditCard(creditCardPayload, userName, password, consoleUrl, replaceUrl)
	if err != nil {
		logger.Logf.Infof("Skipping Test because credit card addition failed")
		suite.T().Skip()
	}
	time.Sleep(10 * time.Second)
	// Validate Credit card being added
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "1111", "Failed to validate credit card suffix")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "10/2028", "Failed to validate credit card expiration")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "Visa", "Failed to validate credit card type")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testuservisa@premium.com", "Failed to validate user email")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Visa", "Failed to validate First name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Visa User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Visa User", "Failed to validate Last name")
	assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

	// Push some usage and let credit depletion happen

	auto_app_response_status, auto_app_response_body := financials.SetAutoApprovalToFalse(cloudAccId, ariaclientId, ariaAuth)
	logger.Logf.Infof("Response code auto approval : ", auto_app_response_status)
	logger.Logf.Infof("Response body auto approval : ", auto_app_response_body)

	now := time.Now().UTC()
	previousDate := now.AddDate(0, 0, -25).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-med", "medvm", "180000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	//

	response_status, response_body := financials.GetAriaPendingInvoiceNumberForClientId(cloudAccId, ariaclientId, ariaAuth)
	logger.Logf.Infof("Aria details, response_body", response_body)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve pending invoice number")
	json := gjson.Parse(response_body)
	pendingInvoice := json.Get("pending_invoice")
	var directive int64 = 2
	pendingInvoice.ForEach(func(_, value gjson.Result) bool {
		invoiceNo := value.Get("invoice_no").String()
		fmt.Println("Discarding pending Invoice No:", invoiceNo)
		response_status, response_body = financials.ManageAriaPendingInvoiceForClientId(cloudAccId, invoiceNo, ariaclientId, ariaAuth, directive)
		logger.Logf.Infof("Response code get pending invoices : ", response_status)
		logger.Logf.Infof("Response body get pending invoices : ", response_body)
		assert.Equal(suite.T(), response_status, 200, "Failed to discard pending invoice number")
		return true
	})

	response_status, response_body = financials.GenerateAriaInvoiceForClientId(cloudAccId, ariaclientId, ariaAuth)
	assert.Equal(suite.T(), response_status, 200, "Failed to Generate Invoice")
	json = gjson.Parse(response_body)
	pendingInvoice = json.Get("out_invoices")
	logger.Logf.Infof("Pending invoices ", pendingInvoice)
	var directive1 int64 = 1
	var medInvoiceNo string
	pendingInvoice.ForEach(func(_, value gjson.Result) bool {
		medInvoiceNo = value.Get("invoice_no").String()
		logger.Logf.Infof("Approving pending Invoice No:", medInvoiceNo)
		response_status, response_body = financials.ManageAriaPendingInvoiceForClientId(cloudAccId, medInvoiceNo, ariaclientId, ariaAuth, directive1)
		logger.Logf.Infof("Response code generate invoices : ", response_status)
		logger.Logf.Infof("Response body generate invoices : ", response_body)
		assert.Equal(suite.T(), response_status, 200, "Failed to Approving pending Invoice")
		return true
	})

	logger.Logf.Infof("Get billing invoice for clientId")
	url = base_url + "/v1/billing/invoices"
	respCode, invoices := financials.GetInvoice(url, token, cloudAccId)
	assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice for clientId")
	logger.Logf.Infof("invoices in account :", invoices)

	jsonInvoices := gjson.Parse(invoices).Get("invoices")
	flag := false
	jsonInvoices.ForEach(func(_, value gjson.Result) bool {
		invoiceNo := value.Get("id").String()
		logger.Logf.Infof(" Processing invoiceNo : ", invoiceNo)
		if invoiceNo == medInvoiceNo {
			flag = true
		}
		// Bug is open for download link
		// downloadLink := value.Get("downloadLink").String()
		//Expect(downloadLink).NotTo(BeNil(), "Invoice download link unavailable nil.")

		//invoice details
		url := base_url + "/v1/billing/invoices/detail"
		//TOdo invoiceNo
		respCode, detail := financials.GetInvoicewithInvoiceId(url, token, cloudAccId, invoiceNo)
		assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice details for clientId")

		logger.Logf.Infof("Invoice details : ", detail) // Empty Response

		// invoices statement
		url = base_url + "/v1/billing/invoices/statement"
		//TOdo invoiceNo
		respCode, statement := financials.GetInvoicewithInvoiceId(url, token, cloudAccId, invoiceNo)

		assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice statement for clientId")
		logger.Logf.Infof(" Invoice statement", statement)
		return true
	})

	//invoices unbilled
	url = base_url + "/v1/billing/invoices/unbilled"
	respCode, resp := financials.GetInvoice(url, token, cloudAccId)
	logger.Logf.Infof(" Processing unbilled invoices  : ", resp)
	assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice for clientId")
	assert.Equal(suite.T(), flag, true, "Can not get invoice in user account with number ", medInvoiceNo)

	// Check cloud account attributes after upgrade

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}

func (suite *BillingAPITestSuite) Test_Premium_Generate_Invoice_Using_Coupon() {
	// Standard user is already enrolled, so start upgrade
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	token := userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	//computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	ariaclientId, ariaAuth := utils.Get_Aria_Config()
	logger.Logf.Infof("Aria details, ariaclientId", ariaclientId)
	logger.Logf.Infof("Aria details, ariaAuth", ariaAuth)

	coupon_err := billing.Create_Redeem_Coupon("Premium", int64(1000), int64(2), cloudAccId)
	assert.Equal(suite.T(), coupon_err, nil, "Failed to create coupon, failed with error : ", coupon_err)

	// Push some usage and let credit depletion happen

	auto_app_response_status, auto_app_response_body := financials.SetAutoApprovalToFalse(cloudAccId, ariaclientId, ariaAuth)
	logger.Logf.Infof("Response code auto approval : ", auto_app_response_status)
	logger.Logf.Infof("Response body auto approval : ", auto_app_response_body)

	now := time.Now().UTC()
	previousDate := now.AddDate(0, 0, -25).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), uuid.NewString(), cloudAccId, previousDate, "vm-spr-med", "medvm", "180000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := base_url + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	//

	response_status, response_body := financials.GetAriaPendingInvoiceNumberForClientId(cloudAccId, ariaclientId, ariaAuth)
	logger.Logf.Infof("Aria details, response_body", response_body)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve pending invoice number")
	json := gjson.Parse(response_body)
	pendingInvoice := json.Get("pending_invoice")
	var directive int64 = 2
	pendingInvoice.ForEach(func(_, value gjson.Result) bool {
		invoiceNo := value.Get("invoice_no").String()
		fmt.Println("Discarding pending Invoice No:", invoiceNo)
		response_status, response_body = financials.ManageAriaPendingInvoiceForClientId(cloudAccId, invoiceNo, ariaclientId, ariaAuth, directive)
		logger.Logf.Infof("Response code get pending invoices : ", response_status)
		logger.Logf.Infof("Response body get pending invoices : ", response_body)
		assert.Equal(suite.T(), response_status, 200, "Failed to discard pending invoice number")
		return true
	})

	response_status, response_body = financials.GenerateAriaInvoiceForClientId(cloudAccId, ariaclientId, ariaAuth)
	assert.Equal(suite.T(), response_status, 200, "Failed to Generate Invoice")
	json = gjson.Parse(response_body)
	pendingInvoice = json.Get("out_invoices")
	logger.Logf.Infof("Pending invoices ", pendingInvoice)
	var directive1 int64 = 1
	var medInvoiceNo string
	pendingInvoice.ForEach(func(_, value gjson.Result) bool {
		medInvoiceNo = value.Get("invoice_no").String()
		logger.Logf.Infof("Approving pending Invoice No:", medInvoiceNo)
		response_status, response_body = financials.ManageAriaPendingInvoiceForClientId(cloudAccId, medInvoiceNo, ariaclientId, ariaAuth, directive1)
		logger.Logf.Infof("Response code generate invoices : ", response_status)
		logger.Logf.Infof("Response body generate invoices : ", response_body)
		assert.Equal(suite.T(), response_status, 200, "Failed to Approving pending Invoice")
		return true
	})

	logger.Logf.Infof("Get billing invoice for clientId")
	url := base_url + "/v1/billing/invoices"
	respCode, invoices := financials.GetInvoice(url, token, cloudAccId)
	assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice for clientId")
	logger.Logf.Infof("invoices in account :", invoices)

	jsonInvoices := gjson.Parse(invoices).Get("invoices")
	flag := false
	jsonInvoices.ForEach(func(_, value gjson.Result) bool {
		invoiceNo := value.Get("id").String()
		logger.Logf.Infof(" Processing invoiceNo : ", invoiceNo)
		if invoiceNo == medInvoiceNo {
			flag = true
		}
		// Bug is open for download link
		// downloadLink := value.Get("downloadLink").String()
		//Expect(downloadLink).NotTo(BeNil(), "Invoice download link unavailable nil.")

		//invoice details
		url := base_url + "/v1/billing/invoices/detail"
		//TOdo invoiceNo
		respCode, detail := financials.GetInvoicewithInvoiceId(url, token, cloudAccId, invoiceNo)
		assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice details for clientId")

		logger.Logf.Infof("Invoice details : ", detail) // Empty Response

		// invoices statement
		url = base_url + "/v1/billing/invoices/statement"
		//TOdo invoiceNo
		respCode, statement := financials.GetInvoicewithInvoiceId(url, token, cloudAccId, invoiceNo)

		assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice statement for clientId")
		logger.Logf.Infof(" Invoice statement", statement)
		return true
	})

	//invoices unbilled
	url = base_url + "/v1/billing/invoices/unbilled"
	respCode, resp := financials.GetInvoice(url, token, cloudAccId)
	logger.Logf.Infof(" Processing unbilled invoices  : ", resp)
	assert.Equal(suite.T(), respCode, 200, "Failed to get billing invoice for clientId")
	assert.Equal(suite.T(), flag, true, "Can not get invoice in user account with number ", medInvoiceNo)

	// Check cloud account attributes after upgrade

	ret_value1, _ := cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "true", "Test Failed while deleting the cloud account(Premium user)")

}
