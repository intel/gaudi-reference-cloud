package public_apis

import (
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/service_apis"
	storage_utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Principal API negative flow", Label("storage", "object_store", "principals", "principal_negative"), Ordered, ContinueOnFailure, func() {
	var (
		bucket_name                              string
		bucket_payload                           string
		principal_payload                        string
		bucket_id_created                        string
		principal_policy_update_negative_payload string
	)

	BeforeAll(func() {
		// retrieve the required information from test config
		bucket_payload = storage_utils.GetBucketPayload()
		principal_payload = storage_utils.GetPrincipalPayload()
		principal_policy_update_negative_payload = storage_utils.GetPrincipalPutPolicyNegativePayload()

		// Function to create the bucket and returning resource_id
		logger.Log.Info("Starting the creation of a bucket using API")
		bucket_name = "automation-bucket-" + storage_utils.GetRandomString()
		bucket_creation_status_positive, bucket_creation_body_positive := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)
		Expect(bucket_creation_status_positive).To(Equal(200), bucket_creation_body_positive)
		Expect(strings.Contains(bucket_creation_body_positive, `"name":"`+storage_cloud_account+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
		bucket_id_created = gjson.Get(bucket_creation_body_positive, "metadata.resourceId").String()
		// Validation
		logger.Log.Info("Checking whether bucket is in ready state")
		ValidationSt := storage_utils.CheckBucketProvisionState(bucket_endpoint, token, bucket_id_created)
		Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
	})

	executePrincipalCreation := func(principal_name string) string {
		logger.Log.Info("Starting the creation of a principal with single principal access")
		principal_creation_status_positive, principal_creation_body_positive := storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload, principal_name, storage_cloud_account+"-"+bucket_name)
		Expect(principal_creation_status_positive).To(Equal(200), principal_creation_body_positive)
		Expect(strings.Contains(principal_creation_body_positive, `"name":"`+principal_name+`"`)).To(BeTrue(), "assertion failed on response body")
		Expect(strings.Contains(principal_creation_body_positive, `"bucketId":"`+storage_cloud_account+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
		principal_id_created := gjson.Get(principal_creation_body_positive, "metadata.userId").String()
		// Validation
		logger.Log.Info("Checking whether principal is in ready state")
		ValidationSt := storage_utils.CheckUserProvisionState(principal_endpoint, token, principal_id_created)
		Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
		return principal_id_created
	}

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("Create principal with too many characters", func() {
		It("Create principal with too many characters", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create principal with too many characters")
			principal_name := "principals-invalid-name-to-validate-the-character-length-for-testing-purpose-attempt" + storage_utils.GetRandomString()
			principal_creation_status_negative, principal_creation_body_negative := storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload, principal_name, storage_cloud_account+"-"+bucket_name)
			Expect(principal_creation_status_negative).To(Equal(400), principal_creation_body_negative)
			Expect(strings.Contains(principal_creation_body_negative, `"message":"invalid resource name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create principal with already used name", func() {
		It("Create principal with already used name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create principal with already used name")
			principal_name := "automation-principal-" + storage_utils.GetRandomString()
			// First principal creation
			res_id := executePrincipalCreation(principal_name)
			// Second principal creation with same name
			principal_creation_status_negative, principal_creation_body_negative := storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload, principal_name, storage_cloud_account+"-"+bucket_name)
			Expect(principal_creation_status_negative).To(Equal(409), principal_creation_body_negative)
			Expect(strings.Contains(principal_creation_body_negative, `"message":"bucket user name `+principal_name+` already exists"`)).To(BeTrue(), "assertion failed on response body")

			// Deleting the created principals
			resource_ids := []string{res_id}
			storage_utils.DeleteMultiplePrincipalsWithRetry(principal_endpoint, token, resource_ids)
		})
	})

	When("Create principal with invalid name(includes special chars)", func() {
		It("Create principal with invalid name(includes special chars)", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create principal with invalid name(includes special chars)")
			principal_name := "automation-test" + storage_utils.GetRandomString() + "!@$%"
			principal_creation_status_negative, principal_creation_body_negative := storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload, principal_name, storage_cloud_account+"-"+bucket_name)
			Expect(principal_creation_status_negative).To(Equal(400), principal_creation_body_negative)
			Expect(strings.Contains(principal_creation_body_negative, `"message":"invalid resource name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create principal with invalid bucketId", func() {
		It("Create principal with invalid bucketId", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create principal with invalid bucketId")
			principal_name := "automation-test" + storage_utils.GetRandomString()
			principal_creation_status_negative, principal_creation_body_negative := storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload, principal_name, "invalid")
			Expect(principal_creation_status_negative).To(Equal(404), principal_creation_body_negative)
			Expect(strings.Contains(principal_creation_body_negative, `"message":"no matching records found"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create principal with empty name", func() {
		It("Create principal with empty name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create principal with empty name")
			principal_name := ""
			principal_creation_status_negative, principal_creation_body_negative := storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload, principal_name, storage_cloud_account+"-"+bucket_name)
			Expect(principal_creation_status_negative).To(Equal(400), principal_creation_body_negative)
			Expect(strings.Contains(principal_creation_body_negative, `"message":"missing name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Get a principal with invalid Id", func() {
		It("Get a principal with invalid Id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Get a principal with invalid resource Id")
			res_id := "invalid-resid"
			get_response_byid_status, get_response_byid_body := storage.GetPrincipalById(principal_endpoint, token, res_id)
			Expect(get_response_byid_status).To(Equal(400), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"invalid userId"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Get a principal with invalid name", func() {
		It("Get a principal with invalid name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Get a principal with invalid name")
			name := "invalid-name"
			get_response_byname_status, get_response_byname_body := storage.GetPrincipalByName(principal_endpoint, token, name)
			Expect(get_response_byname_status).To(Equal(404), get_response_byname_body)
			Expect(strings.Contains(get_response_byname_body, `"message":"no matching records found"`)).To(BeTrue(), "assertion failed on response body")

		})
	})

	When("Update policy for object service principal with invalid actions using Id", func() {
		It("Update policy for object service principal with invalid actions using Id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update policy for object service principal with invalid actions using Id")
			// Create principal
			principal_name := "automation-principal-" + storage_utils.GetRandomString()
			principal_id_created := executePrincipalCreation(principal_name)
			principal_actions := []string{"invalid"}
			principal_permissions := []string{"ReadBucket", "WriteBucket", "DeleteBucket"}
			put_response_byid_status, put_response_byid_body := storage_utils.UpdatePrincipalPolicy(principal_endpoint, token, principal_policy_update_negative_payload,
				principal_name, principal_id_created, storage_cloud_account+"-"+bucket_name, principal_actions, principal_permissions, "Id", false)
			Expect(put_response_byid_status).To(Equal(400), put_response_byid_body)
			// Delete principal
			principal_ids := []string{principal_id_created}
			storage_utils.DeleteMultiplePrincipalsWithRetry(principal_endpoint, token, principal_ids)
		})
	})

	When("Update policy for object service principal with invalid actions using name", func() {
		It("Update policy for object service principal with invalid actions using name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update policy for object service principal with invalid actions using name")
			// Create principal
			principal_name := "automation-principal-" + storage_utils.GetRandomString()
			principal_id_created := executePrincipalCreation(principal_name)
			principal_actions := []string{"invalid"}
			principal_permissions := []string{"ReadBucket", "WriteBucket", "DeleteBucket"}
			put_response_byid_status, put_response_byid_body := storage_utils.UpdatePrincipalPolicy(principal_endpoint, token, principal_policy_update_negative_payload,
				principal_name, principal_id_created, storage_cloud_account+"-"+bucket_name, principal_actions, principal_permissions, "Name", false)
			Expect(put_response_byid_status).To(Equal(400), put_response_byid_body)
			// Delete principal
			principal_ids := []string{principal_id_created}
			storage_utils.DeleteMultiplePrincipalsWithRetry(principal_endpoint, token, principal_ids)
		})
	})

	When("Update policy for object service principal with invalid permissions using Id", func() {
		It("Update policy for object service principal with invalid permissions using Id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update policy for object service principal with invalid permissions using Id")
			// Create principal
			principal_name := "automation-principal-" + storage_utils.GetRandomString()
			principal_id_created := executePrincipalCreation(principal_name)
			principal_actions := []string{"GetBucketLocation"}
			principal_permissions := []string{"Invalid"}
			put_response_byid_status, put_response_byid_body := storage_utils.UpdatePrincipalPolicy(principal_endpoint, token, principal_policy_update_negative_payload,
				principal_name, principal_id_created, storage_cloud_account+"-"+bucket_name, principal_actions, principal_permissions, "Id", false)
			Expect(put_response_byid_status).To(Equal(400), put_response_byid_body)
			// Delete principal
			principal_ids := []string{principal_id_created}
			storage_utils.DeleteMultiplePrincipalsWithRetry(principal_endpoint, token, principal_ids)
		})
	})

	AfterAll(func() {
		// Delete all buckets
		logger.Log.Info("Starting the deletion of bucket(s) using API")
		resource_ids := []string{bucket_id_created}
		storage_utils.DeleteMultipleBucketsWithRetry(bucket_endpoint, token, resource_ids)
	})

})
