package quota_management

import (
	"encoding/json"
	auth "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/auth/oidc_auth"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("IDC - BMaaS Quota Enforcement", Label("compute", "compute_bu", "bmaas_quota"), Ordered, func() {
	var (
		instanceType             string
		machineImage             string
		intelInstanceList        []string
		standardInstanceList     []string
		premiumInstanceList      []string
		intelInstanceDeletion    bool
		standardInstanceDeletion bool
		premiumInstanceDeletion  bool
	)

	BeforeAll(func() {
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		machineImage = utils.GetJsonValue("machineImage")
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("intel user", func() {
		It("intel user", func() {
			// create the instances for each type of user (intel, standard and premium)
			intelUser := utils.GetJsonValue("intelUser")
			var intelUserJson map[string]interface{}
			json.Unmarshal([]byte(intelUser), &intelUserJson)

			var intelUserQuota int = int(intelUserJson["max_quota"].(float64))
			instancePayload, _ := json.Marshal(intelUserJson["instancePayload"])

			//use the username and get the token
			intelToken := auth.GetUserTokenViaResty(oidcUrl, intelUsername)
			// create the instances as per quota limit
			// TODO: machine needs to be removed from the input, fetch the machine image via instance_type (pattern similar to instance creation)
			intelInstanceList = utils.CreateInstanceWithinQuota(intelInstanceEndpoint, intelToken, intelUserQuota, string(instancePayload), instanceType, intelSSHKeyName, intelVnet, machineImage, availabilityZone)

			// create instance after quota exhaustion
			utils.CreateInstanceAfterQuotaLimit(intelInstanceEndpoint, intelToken, string(instancePayload), instanceType, intelSSHKeyName, intelVnet, machineImage, availabilityZone)

			// delete the instances
			utils.DeleteInstanceWithinQuota(intelInstanceEndpoint, intelToken, intelInstanceList, 10*time.Minute)
			intelInstanceDeletion = true
		})
	})

	// Keeping standard and premium quota tests commented due to hardware limitations
	When("standard user", func() {
		It("standard user", func() {
			// create the instances for each type of user (intel, standard and premium)
			standardUser := utils.GetJsonValue("standardUser")
			var standardUserJson map[string]interface{}
			json.Unmarshal([]byte(standardUser), &standardUserJson)

			var standardUserQuota int = int(standardUserJson["max_quota"].(float64))
			instancePayload, _ := json.Marshal(standardUserJson["instancePayload"])

			//use the username and get the token
			var standardToken = auth.GetUserTokenViaResty(oidcUrl, standardUsername)

			// create the instances as per quota limit
			standardInstanceList = utils.CreateInstanceWithinQuota(standardInstanceEndpoint, standardToken, standardUserQuota, string(instancePayload), instanceType, standardSSHKeyName, standardVnet, machineImage, availabilityZone)

			// create instance after quota exhaustion
			utils.CreateInstanceAfterQuotaLimit(standardInstanceEndpoint, standardToken, string(instancePayload), instanceType, standardSSHKeyName, standardVnet, machineImage, availabilityZone)

			// delete the instances
			utils.DeleteInstanceWithinQuota(standardInstanceEndpoint, standardToken, standardInstanceList, 10*time.Minute)
			standardInstanceDeletion = true
		})
	})
	When("premium user", func() {
		It("premium user", func() {
			// create the instances for each type of user (intel, standard and premium)
			premiumUser := utils.GetJsonValue("premiumUser")
			var premiumUserJson map[string]interface{}
			json.Unmarshal([]byte(premiumUser), &premiumUserJson)

			var premiumUserQuota int = int(premiumUserJson["max_quota"].(float64))
			instancePayload, _ := json.Marshal(premiumUserJson["instancePayload"])

			//use the username and get the token
			var premiumToken = auth.GetUserTokenViaResty(oidcUrl, premiumUsername)

			// create the instances as per quota limit
			premiumInstanceList = utils.CreateInstanceWithinQuota(premiumInstanceEndpoint, premiumToken, premiumUserQuota, string(instancePayload), instanceType, premiumSSHKeyName, premiumVnet, machineImage, availabilityZone)

			// create instance after quota exhaustion
			utils.CreateInstanceAfterQuotaLimit(premiumInstanceEndpoint, premiumToken, string(instancePayload), instanceType, premiumSSHKeyName, premiumVnet, machineImage, availabilityZone)

			// delete the instances
			utils.DeleteInstanceWithinQuota(premiumInstanceEndpoint, premiumToken, premiumInstanceList, 10*time.Minute)
			premiumInstanceDeletion = true
		})
	})

	AfterAll(func() {
		if !intelInstanceDeletion {
			for _, name := range intelInstanceList {
				deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceByName(intelInstanceEndpoint, token, name)
				Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

				// Validation
				logInstance.Println("Validation of Instance Deletion")
				instanceValidation := utils.CheckInstanceDeletionById(intelInstanceEndpoint, token, name)
				Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
			}
		}

		if !standardInstanceDeletion {
			for _, name := range standardInstanceList {
				deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceByName(standardInstanceEndpoint, token, name)
				Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

				// Validation
				logInstance.Println("Validation of Instance Deletion")
				instanceValidation := utils.CheckInstanceDeletionById(standardInstanceEndpoint, token, name)
				Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
			}
		}

		if !premiumInstanceDeletion {
			for _, name := range premiumInstanceList {
				deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceByName(premiumInstanceEndpoint, token, name)
				Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

				// Validation
				logInstance.Println("Validation of Instance Deletion")
				instanceValidation := utils.CheckInstanceDeletionById(premiumInstanceEndpoint, token, name)
				Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
			}
		}
	})
})
