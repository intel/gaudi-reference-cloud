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

var _ = Describe("Lifecycle rule API negative flow", Label("storage", "object_store", "lifecycle_rules", "lifecycle_rule_negative"), Ordered, ContinueOnFailure, func() {
	var (
		bucket_name       string
		bucket_payload    string
		bucket_id_created string

		rule_payload          string
		rule_invalid_payload  string
		rule_invalid_payload2 string
		rule_id_created       string
	)
	BeforeAll(func() {
		// retrieve the required information from test config
		bucket_payload = storage_utils.GetBucketPayload()
		rule_payload = storage_utils.GetRulePayload()
		rule_invalid_payload = storage_utils.GetInvalidRulePayload()
		rule_invalid_payload2 = storage_utils.GetInvalidRulePayload2()

		// Bucket to be created
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

	executeRuleCreation := func(rule_name string, bucket_id string) string {
		logger.Log.Info("Starting the creation of a rule")
		rule_creation_status_positive, rule_creation_body_positive := storage_utils.RuleCreation(rule_endpoint, token, rule_payload, rule_name, bucket_id_created)
		Expect(rule_creation_status_positive).To(Equal(200), rule_creation_body_positive)
		Expect(strings.Contains(rule_creation_body_positive, `"ruleName":"`+rule_name+`"`)).To(BeTrue(), "assertion failed on response body")
		rule_id_created = gjson.Get(rule_creation_body_positive, "metadata.resourceId").String()
		// Validation
		logger.Log.Info("Checking whether rule is in ready state")
		ValidationSt := storage_utils.CheckRuleProvisionState(rule_endpoint, token, bucket_id, rule_id_created)
		Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
		return rule_id_created
	}

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("Search rule using invalid bucket id", func() {
		It("Search rule using invalid bucket id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Search rule using invalid bucket id")
			get_response_byid_status, get_response_byid_body := storage.GetAllRules(rule_endpoint, token, "invalid-bucket-id8723h")
			Expect(get_response_byid_status).To(Equal(400), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"invalid bucketId"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create rule with too many characters", func() {
		It("Create rule with too many characters", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create rule with too many characters")
			rule_name := "rules-invalid-name-to-validate-the-character-length-for-testing-purpose-attempt" + storage_utils.GetRandomString()
			rule_creation_status_negative, rule_creation_body_negative := storage_utils.RuleCreation(rule_endpoint, token, rule_payload, rule_name, bucket_id_created)
			Expect(rule_creation_status_negative).To(Equal(400), rule_creation_body_negative)
			Expect(strings.Contains(rule_creation_body_negative, `"message":"invalid resource name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create rule with already used name", func() {
		It("Create rule with already used name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create rule with already used name")
			rule_name := "test-rule"
			// First rule creation
			res_id := executeRuleCreation(rule_name, bucket_id_created)
			// Second rule creation with same name
			rule_creation_status_negative, rule_creation_body_negative := storage_utils.RuleCreation(rule_endpoint, token, rule_payload, rule_name, bucket_id_created)
			Expect(rule_creation_status_negative).To(Equal(409), rule_creation_body_negative)
			Expect(strings.Contains(rule_creation_body_negative, `"message":"bucket lifecycle rule name `+rule_name+` already exists"`)).To(BeTrue(), "assertion failed on response body")

			// Deleting the created rules
			resource_ids := []string{res_id}
			storage_utils.DeleteMultipleRulesWithRetry(rule_endpoint, token, bucket_id_created, resource_ids)
		})
	})

	When("Create rule with invalid name(includes special chars)", func() {
		It("Create rule with invalid name(includes special chars)", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create rule with invalid name(includes special chars)")
			rule_name := "test-rule!@$%"
			rule_creation_status_negative, rule_creation_body_negative := storage_utils.RuleCreation(rule_endpoint, token, rule_payload, rule_name, bucket_id_created)
			Expect(rule_creation_status_negative).To(Equal(400), rule_creation_body_negative)
			Expect(strings.Contains(rule_creation_body_negative, `"message":"invalid resource name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create rule with invalid bucketId", func() {
		It("Create rule with invalid bucketId", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create rule with invalid bucketId")
			rule_name := "test-rule"
			rule_creation_status_negative, rule_creation_body_negative := storage_utils.RuleCreation(rule_endpoint, token, rule_payload, rule_name, "invalid-bucket-id1j2b332b")
			Expect(rule_creation_status_negative).To(Equal(400), rule_creation_body_negative)
			Expect(strings.Contains(rule_creation_body_negative, `"message":"quering bucket lifecycle rule failed"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create rule with empty name", func() {
		It("Create rule with empty name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create rule with missing name")
			rule_name := ""
			rule_creation_status_negative, rule_creation_body_negative := storage_utils.RuleCreation(rule_endpoint, token, rule_payload, rule_name, bucket_id_created)
			Expect(rule_creation_status_negative).To(Equal(400), rule_creation_body_negative)
			Expect(strings.Contains(rule_creation_body_negative, `"message":"missing ruleName in metadata"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create rule with both expire days and delete marker", func() {
		It("Create rule with both expire days and delete marker", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create rule with both expire days and delete marker")
			rule_name := "test-rule"
			rule_creation_status_negative, rule_creation_body_negative := storage_utils.RuleCreation(rule_endpoint, token, rule_invalid_payload, rule_name, bucket_id_created)
			Expect(rule_creation_status_negative).To(Equal(400), rule_creation_body_negative)
			Expect(strings.Contains(rule_creation_body_negative, `"message":"delete marker and expiry days cannot both be set"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Get a rule with invalid rule Id", func() {
		It("Get a rule with invalid rule Id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Get a rule with invalid resource Id")
			res_id := "Tuvb5026-d516-48C8-bfd3-5998547265U2"
			get_response_byid_status, get_response_byid_body := storage.GetRuleById(rule_endpoint, bucket_id_created, token, res_id)
			Expect(get_response_byid_status).To(Equal(400), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"invalid ruleId"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Get a rule with invalid bucket Id", func() {
		It("Get a rule with invalid bucket Id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Get a rule with invalid bucket Id")
			bucket_id := "invalid-bucket_id"
			get_response_byid_status, get_response_byid_body := storage.GetRuleById(rule_endpoint, bucket_id, token, rule_id_created)
			Expect(get_response_byid_status).To(Equal(400), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"invalid bucketId"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Update a rule with invalid rule Id", func() {
		It("Update a rule with invalid rule Id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update a rule with invalid resource Id")
			res_id := "invalid-resid"
			get_response_byid_status, get_response_byid_body := storage.PutRuleById(rule_endpoint, bucket_id_created, token, res_id, rule_payload)
			Expect(get_response_byid_status).To(Equal(400), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"invalid ruleId"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Update a rule with invalid bucket Id", func() {
		It("Update a rule with invalid bucket Id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update a rule with invalid bucket Id")
			bucket_id := "invalid-bucket-id"
			get_response_byid_status, get_response_byid_body := storage.PutRuleById(rule_endpoint, bucket_id, token, rule_id_created, rule_payload)
			Expect(get_response_byid_status).To(Equal(400), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"invalid bucketId"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Update a rule with invalid payload", func() {
		It("Update a rule with invalid payload", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create rule")
			rule_name := "test-rule"
			// First rule creation
			res_id := executeRuleCreation(rule_name, bucket_id_created)

			logger.Log.Info("Update a rule with invalid payload")
			get_response_byid_status, get_response_byid_body := storage.PutRuleById(rule_endpoint, bucket_id_created, token, res_id, rule_invalid_payload2)
			Expect(get_response_byid_status).To(Equal(400), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"delete marker and expiry days cannot both be set"`)).To(BeTrue(), "assertion failed on response body")
			// Deleting the created rules
			resource_ids := []string{res_id}
			storage_utils.DeleteMultipleRulesWithRetry(rule_endpoint, token, bucket_id_created, resource_ids)
		})
	})

	When("Delete a rule with invalid ruleId", func() {
		It("Delete a rule with invalid ruleId", func() {
			defer GinkgoRecover()
			logger.Log.Info("Delete a rule with invalid ruleId")
			res_id := "invalid-rule-id"
			delete_rule_endpoint := rule_endpoint + bucket_id_created + "/lifecyclerule"
			get_response_byid_status, get_response_byid_body := storage.DeleteRuleById(delete_rule_endpoint, token, res_id)
			Expect(get_response_byid_status).To(Equal(404), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"Not Found"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Delete a rule with invalid bucketId", func() {
		It("Delete a rule with invalid bucketId", func() {
			defer GinkgoRecover()
			logger.Log.Info("Delete a rule with invalid bucketId")
			bucket_id := "invalid-bucket-id"
			delete_rule_endpoint := rule_endpoint + bucket_id + "/lifecyclerule"
			get_response_byid_status, get_response_byid_body := storage.DeleteRuleById(delete_rule_endpoint, token, rule_id_created)
			Expect(get_response_byid_status).To(Equal(404), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"Not Found"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	AfterAll(func() {
		// Delete all buckets
		logger.Log.Info("Starting the deletion of bucket/s using API")
		resource_ids := []string{bucket_id_created}
		storage_utils.DeleteMultipleBucketsWithRetry(bucket_endpoint, token, resource_ids)

	})

})
