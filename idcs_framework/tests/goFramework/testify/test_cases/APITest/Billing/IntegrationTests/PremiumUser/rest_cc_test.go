//go:build Functional || TestTT
// +build Functional TestTT

package PremiumBillingAPITest

import (
	"fmt"
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
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func (suite *BillingAPITestSuite) Test_Premium_Usage_Multiple_VM1() {

	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(20), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.Equal(suite.T(), "UPGRADE_COMPLETE", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	var sml_vm bool = true
	var med_vm bool = true
	var lrg_vm bool = true
	var bm_vm bool = true

	vm_creation_error, create_response_body, skip_val := billing.Create_Vm_Instance(cloudAccId, "vm-spr-med", userToken, 200)
	if skip_val {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error)
		med_vm = false
	}
	assert.Equal(suite.T(), vm_creation_error, nil, "Failed to create vm with error : %s", vm_creation_error)

	vm_creation_error1, create_response_body1, skip_val1 := billing.Create_Vm_Instance(cloudAccId, "vm-spr-sml", userToken, 200)
	if skip_val1 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error1)
		sml_vm = false
	}

	vm_creation_error2, create_response_body2, skip_val2 := billing.Create_Vm_Instance(cloudAccId, "vm-spr-lrg", userToken, 200)
	if skip_val2 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error2)
		lrg_vm = false
	}
	assert.Equal(suite.T(), vm_creation_error2, nil, "Failed to create vm with error : %s", vm_creation_error2)

	vm_creation_error3, create_response_body3, skip_val3 := billing.Create_Vm_Instance(cloudAccId, "bm-spr", userToken, 200)
	if skip_val3 {
		logger.Logf.Infof("Skipping Test because create instance returned error %s : ", vm_creation_error3)
		bm_vm = false
	}
	assert.Equal(suite.T(), vm_creation_error3, nil, "Failed to create vm with error : %s", vm_creation_error3)

	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	if vm_creation_error != nil {
		if med_vm {
			time.Sleep(5 * time.Minute)
			instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
			instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
			// delete the instance created
			delete_status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
			assert.Equal(suite.T(), 200, delete_status, "Failed : Instance not deleted after all credits expired")
			time.Sleep(10 * time.Second)
		}
	}

	if vm_creation_error1 != nil {
		if sml_vm {
			time.Sleep(5 * time.Minute)
			instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
			instance_id_created1 := gjson.Get(create_response_body1, "metadata.resourceId").String()
			// delete the instance created
			delete_status1, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created1)
			assert.Equal(suite.T(), 200, delete_status1, "Failed : Instance not deleted after all credits expired")
			time.Sleep(10 * time.Second)
		}
	}

	if vm_creation_error2 != nil {
		if lrg_vm {
			time.Sleep(5 * time.Minute)
			instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
			instance_id_created1 := gjson.Get(create_response_body2, "metadata.resourceId").String()
			// delete the instance created
			delete_status1, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created1)
			assert.Equal(suite.T(), 200, delete_status1, "Failed : Instance not deleted after all credits expired")
			time.Sleep(10 * time.Second)
		}
	}

	if vm_creation_error3 != nil {
		if bm_vm {
			time.Sleep(5 * time.Minute)
			instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
			instance_id_created1 := gjson.Get(create_response_body3, "metadata.resourceId").String()
			// delete the instance created
			delete_status1, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created1)
			assert.Equal(suite.T(), 200, delete_status1, "Failed : Instance not deleted after all credits expired")
			time.Sleep(10 * time.Second)
		}
	}

	if sml_vm {
		usage_err := billing.ValidateUsageNotZero(cloudAccId, float64(0.0075), "vm-spr-sml", authToken)
		assert.Equal(suite.T(), usage_err, nil, "Failed to validate usages, validation failed with error : ", usage_err)
	}
	if med_vm {
		usage_err1 := billing.ValidateUsageNotZero(cloudAccId, float64(0.015), "vm-spr-med", authToken)
		assert.Equal(suite.T(), usage_err1, nil, "Failed to validate usages, validation failed with error : ", usage_err1)
	}

	if lrg_vm {
		usage_err1 := billing.ValidateUsageNotZero(cloudAccId, float64(0.03), "vm-spr-lrg", authToken)
		assert.Equal(suite.T(), usage_err1, nil, "Failed to validate usages, validation failed with error : ", usage_err1)
	}

	time.Sleep(5 * time.Minute)
	credits_err := billing.ValidateCreditsNonZeroDepletion(cloudAccId, 20, authToken)
	assert.Equal(suite.T(), credits_err, nil, "Failed to validate credits, validation failed with error : ", credits_err)

	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium User)")

}

func (suite *BillingAPITestSuite) Test_Premium_Sanity_Validate_all_Usages() {
	userName := utils.Get_UserName("Premium")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	base_url := utils.Get_Base_Url1()
	computeUrl := utils.Get_Compute_Base_Url()
	cloudAccId, _ := testsetup.GetCloudAccountId(userName, base_url, authToken)

	// Check cloud account attributes before upgrade

	ret_value1, responsePayload := cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_STANDARD", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for Account type")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	// Upgrade standard account using coupon
	upgrade_err := billing.Standard_to_premium_upgrade_with_coupon("Premium", int64(100), 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")
	assert.Equal(suite.T(), upgrade_err, nil, "upgrade from standard to premium with coupon failed, with error : ", upgrade_err)

	migrate_err := billing.Credit_Migrate(cloudAccId, authToken)
	assert.Equal(suite.T(), migrate_err, nil, "upgrade from standard to premium with coupon failed, with error : ", migrate_err)

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = cloudAccounts.GetCAccById(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")
	assert.Equal(suite.T(), "ACCOUNT_TYPE_PREMIUM", gjson.Get(responsePayload, "type").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "true", gjson.Get(responsePayload, "paidServicesAllowed").String(), "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "lowCredits").String(), "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.Equal(suite.T(), "false", gjson.Get(responsePayload, "terminatePaidServices").String(), "Failed: Validation Failed , Get on cloud account for terminatePaidServices")
	assert.Equal(suite.T(), "UPGRADE_COMPLETE", gjson.Get(responsePayload, "upgradedToPremium").String(), "Failed: Validation Failed , Get on cloud account for upgradedToPremium")
	assert.Equal(suite.T(), "UPGRADE_NOT_INITIATED", gjson.Get(responsePayload, "upgradedToEnterprise").String(), "Failed: Validation Failed , Get on cloud account for upgradedToEnterprise")

	//token := utils.Get_Standard_Token()
	logger.Log.Info("Compute Url : " + computeUrl)
	baseUrl := utils.Get_Base_Url1()
	// Create an ssh key  for the user
	//authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/" + "sshpublickeys"
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

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	assert.Equal(suite.T(), vnet_creation_status, 200, "Failed: Vnet  creation failed for Premium User")

	skip_small := false
	skip_med := false
	skip_large := false
	skip_bm := false
	skip_gaudi := false

	var med_count int
	var small_count int
	var bm_count int
	var large_count int
	var gaudi_count int

	var VMName string
	var instance_id_created1 string
	var instance_id_created2 string
	var instance_id_created3 string
	var instance_id_created4 string
	var instance_id_created5 string

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Failed to create instance of type small due to high demand %s : ", create_response_body)
		skip_small = true

	} else {
		instance_id_created1 = gjson.Get(create_response_body, "metadata.resourceId").String()
		VMName = gjson.Get(create_response_body, "metadata.name").String()
		assert.Equal(suite.T(), create_response_status, 200, "Failed: Vm Instance creation failed")
		assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	}

	vm_name = "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload = compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-med", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Failed to create instance of type medium due to high demand %s : ", create_response_body)
		skip_med = true

	} else {
		instance_id_created2 = gjson.Get(create_response_body, "metadata.resourceId").String()
		VMName = gjson.Get(create_response_body, "metadata.name").String()
		assert.Equal(suite.T(), create_response_status, 200, "Failed: Vm Instance creation failed")
		assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	}

	vm_name = "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload = compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "vm-spr-lrg", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Failed to create instance of type large due to high demand %s : ", create_response_body)
		skip_large = true
	} else {
		instance_id_created3 = gjson.Get(create_response_body, "metadata.resourceId").String()
		VMName = gjson.Get(create_response_body, "metadata.name").String()
		assert.Equal(suite.T(), create_response_status, 200, "Failed: Vm Instance creation failed")
		assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	}

	vm_name = "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload = compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "bm-spr", "ubuntu-22.04-spr-metal-cloudimg-amd64-v20240115", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Failed to create instance of type BM due to high demand %s : ", create_response_body)
		skip_bm = true
	} else {
		instance_id_created4 = gjson.Get(create_response_body, "metadata.resourceId").String()
		VMName = gjson.Get(create_response_body, "metadata.name").String()
		assert.Equal(suite.T(), create_response_status, 200, "Failed: Vm Instance creation failed")
		assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	}

	vm_name = "autovm-" + utils.GenerateSSHKeyName(4)
	instance_payload = compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, "bm-icp-gaudi2", "ubuntu-20.04-gaudi-metal-cloudimg-amd64-v20231013", ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("Instance Creation Response Body : " + create_response_body)
	if create_response_status == 429 {
		logger.Logf.Infof("Failed to create instance of type BM due to high demand %s : ", create_response_body)
		skip_gaudi = true
	} else {
		instance_id_created5 = gjson.Get(create_response_body, "metadata.resourceId").String()
		VMName = gjson.Get(create_response_body, "metadata.name").String()
		assert.Equal(suite.T(), create_response_status, 200, "Failed: Vm Instance creation failed")
		assert.Equal(suite.T(), vm_name, VMName, "Failed to create VM instance, resposne validation failed")
	}

	// Push Metering Data to use all Credits

	now := time.Now().UTC()
	previousDate := now.Add(25 * time.Minute).Format("2006-01-02T15:04:05.999999Z")
	fmt.Println("Metering Date", previousDate)
	create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
		uuid.NewString(), utils.GenerateString(10), cloudAccId, previousDate, "vm-spr-sml", "smallvm", "180000")
	fmt.Println("create_payload", create_payload)
	metering_api_base_url := baseUrl + "/v1/meteringrecords"
	response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, authToken, create_payload)
	assert.Equal(suite.T(), response_status, 200, "Failed: Failed to create metering data")

	// Wait for the coupon to expire
	logger.Logf.Infof("Waiting for %d Minutes to get usages... ", utils.GetSchedulerTimeout())
	time.Sleep(time.Duration(utils.GetSchedulerTimeout()) * time.Minute)

	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, authToken)
	assert.Equal(suite.T(), usage_response_status, 200, "Failed: Failed to validate usage_response_status")
	logger.Logf.Info("usage_response_body: ", usage_response_body)
	tamt := 20
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")
	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if skip_small != true {
			if gjson.Get(data, "productType").String() == "vm-spr-sml" {
				small_count = 1
				Amount := gjson.Get(data, "amount").String()
				actualAMount, _ := strconv.ParseFloat(Amount, 64)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")

				minsUsed := gjson.Get(data, "minsUsed").String()
				minsFactor, _ := strconv.ParseFloat(minsUsed, 64)
				assert.Greater(suite.T(), minsFactor, float64(0), "Failed: Failed to validate minsUsed")

				rate := gjson.Get(data, "rate").String()
				rateFactor, _ := strconv.ParseFloat(rate, 64)
				assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

				logger.Logf.Infof("Actual Usage :    ", actualAMount)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")
				assert.Equal(suite.T(), small_count, 1, "Failed: Small Instance Usage not displayed")
			}

		}

		if skip_med != true {
			if gjson.Get(data, "productType").String() == "vm-spr-med" {
				med_count = 1
				Amount := gjson.Get(data, "amount").String()
				actualAMount, _ := strconv.ParseFloat(Amount, 64)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")

				minsUsed := gjson.Get(data, "minsUsed").String()
				minsFactor, _ := strconv.ParseFloat(minsUsed, 64)
				assert.Greater(suite.T(), minsFactor, float64(0), "Failed: Failed to validate minsUsed")

				rate := gjson.Get(data, "rate").String()
				rateFactor, _ := strconv.ParseFloat(rate, 64)
				assert.Equal(suite.T(), 0.015, rateFactor, "Failed: Failed to validate rate amount")
				assert.Equal(suite.T(), med_count, 1, "Failed: Medium Usage not displayed")

			}

		}

		if skip_large != true {
			if gjson.Get(data, "productType").String() == "vm-spr-lrg" {
				large_count = 1
				Amount := gjson.Get(data, "amount").String()
				actualAMount, _ := strconv.ParseFloat(Amount, 64)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")

				minsUsed := gjson.Get(data, "minsUsed").String()
				minsFactor, _ := strconv.ParseFloat(minsUsed, 64)
				assert.Greater(suite.T(), minsFactor, float64(0), "Failed: Failed to validate minsUsed")

				rate := gjson.Get(data, "rate").String()
				rateFactor, _ := strconv.ParseFloat(rate, 64)
				assert.Equal(suite.T(), 0.03, rateFactor, "Failed: Failed to validate rate amount")
				assert.Equal(suite.T(), large_count, 1, "Failed: Large Usage not displayed")

			}

		}

		if skip_bm != true {
			if gjson.Get(data, "productType").String() == "bm-spr" {
				bm_count = 1
				Amount := gjson.Get(data, "amount").String()
				actualAMount, _ := strconv.ParseFloat(Amount, 64)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")

				rate := gjson.Get(data, "rate").String()
				rateFactor, _ := strconv.ParseFloat(rate, 64)
				assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

				minsUsed := gjson.Get(data, "minsUsed").String()
				minsFactor, _ := strconv.ParseFloat(minsUsed, 64)
				assert.Greater(suite.T(), minsFactor, float64(0), "Failed: Failed to validate minsUsed")

				logger.Logf.Infof("Actual Usage :    ", actualAMount)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")
				assert.Equal(suite.T(), bm_count, 1, "Failed: Bm Usage not displayed")
			}

		}

		if skip_gaudi != true {
			if gjson.Get(data, "productType").String() == "bm-icp-gaudi2" {
				gaudi_count = 1
				Amount := gjson.Get(data, "amount").String()
				actualAMount, _ := strconv.ParseFloat(Amount, 64)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")

				rate := gjson.Get(data, "rate").String()
				rateFactor, _ := strconv.ParseFloat(rate, 64)
				assert.Equal(suite.T(), 0.0075, rateFactor, "Failed: Failed to validate rate amount")

				minsUsed := gjson.Get(data, "minsUsed").String()
				minsFactor, _ := strconv.ParseFloat(minsUsed, 64)
				assert.Greater(suite.T(), minsFactor, float64(0), "Failed: Failed to validate minsUsed")

				logger.Logf.Infof("Actual Usage :    ", actualAMount)
				assert.Greater(suite.T(), actualAMount, float64(0), "Failed: Failed to validate usage amount")
				assert.Equal(suite.T(), gaudi_count, 1, "Failed: Gaudi Usage not displayed")
			}

		}

		return true // keep iterating
	})

	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	assert.GreaterOrEqual(suite.T(), total_amount_from_response, float64(tamt), "Failed: Failed to validate usage amount")

	//Validate credits

	time.Sleep(2 * time.Minute)
	baseUrl = utils.Get_Credits_Base_Url() + "/credit"
	response_status, responseBody := financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Billing Account Cloud Credits")
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").String()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	amt := 20
	assert.GreaterOrEqual(suite.T(), usedAmount, float64(amt), "Failed to validate used credits")
	assert.Greater(suite.T(), remainingAmount, "50", "Failed to validate remaining credits")
	res := billing.GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	assert.Greater(suite.T(), unappliedCredits, "0", "Failed : Unapplied cloud credit did not become zero")

	if create_response_status == 200 {
		time.Sleep(5 * time.Minute)
		// delete the instance created
		instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
		if skip_small != true {
			get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created1)
			logger.Logf.Info("get_response_status: ", get_response_status)
			logger.Logf.Info("get_response_body: ", get_response_body)
			assert.Equal(suite.T(), 200, get_response_status, "Failed : Small Instance not deleted after all credits used")
		}
		if skip_med != true {
			get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created2)
			logger.Logf.Info("get_response_status: ", get_response_status)
			logger.Logf.Info("get_response_body: ", get_response_body)
			assert.Equal(suite.T(), 200, get_response_status, "Failed : Medium Instance not deleted after all credits used")
		}

		if skip_large != true {
			get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created3)
			logger.Logf.Info("get_response_status: ", get_response_status)
			logger.Logf.Info("get_response_body: ", get_response_body)
			assert.Equal(suite.T(), 200, get_response_status, "Failed : Large Instance not deleted after all credits used")
		}

		if skip_bm != true {
			get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created4)
			logger.Logf.Info("get_response_status: ", get_response_status)
			logger.Logf.Info("get_response_body: ", get_response_body)
			assert.Equal(suite.T(), 200, get_response_status, "Failed : BM Instance not deleted after all credits used")
		}

		if skip_gaudi != true {
			get_response_status, get_response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created5)
			logger.Logf.Info("get_response_status: ", get_response_status)
			logger.Logf.Info("get_response_body: ", get_response_body)
			assert.Equal(suite.T(), 200, get_response_status, "Failed : Gaudi Instance not deleted after all credits used")
		}
		// del_response_status, del_response_body := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		// logger.Logf.Info("del_response_status: ", del_response_status)
		// logger.Logf.Info("del_response_body: ", del_response_body)
		time.Sleep(10 * time.Second)
	}

	// Validate cloud credit data
	zeroamt := 0
	response_status, responseBody = financials.GetCredits(baseUrl, userToken, cloudAccId)
	assert.Equal(suite.T(), response_status, 200, "Failed to retrieve Credit details")
	totalRemainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	assert.Greater(suite.T(), totalRemainingAmount, float64(zeroamt), "Failed: Failed to validate expired credits")
	//totalRemainingAmount = testsetup.RoundFloat(totalRemainingAmount, 2)
	fmt.Println("totalRemainingAmount", totalRemainingAmount)
	assert.Less(suite.T(), totalRemainingAmount, float64(100), "Test Failed while deleting the cloud account(Premium User)")
	ret_value1, _ = cloudAccounts.DeleteCloudAccount(cloudAccId, 200)
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed while deleting the cloud account(Premium User)")

}
