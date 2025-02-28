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

var _ = Describe("Lifecycle rule API positive flow", Label("storage", "object_store", "lifecycle_rules", "lifecycle_rule_positive"), Ordered, ContinueOnFailure, func() {
	var (
		bucket_name       string
		bucket_payload    string
		bucket_id_created string

		rule_name                     string
		rule_payload                  string
		rule_id_created               string
		rule_creation_status_positive int
		rule_creation_body_positive   string
	)
	BeforeAll(func() {
		// retrieve the required information from test config
		bucket_payload = storage_utils.GetBucketPayload()
		rule_payload = storage_utils.GetRulePayload()

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

		// Rule to be created
		logger.Log.Info("Starting the creation of a lifecycle rule using API")
		rule_name = "test-rule"
		rule_creation_status_positive, rule_creation_body_positive = storage_utils.RuleCreation(rule_endpoint, token, rule_payload, rule_name, bucket_id_created)
		Expect(rule_creation_status_positive).To(Equal(200), rule_creation_body_positive)
		Expect(strings.Contains(rule_creation_body_positive, `"ruleName":"`+rule_name+`"`)).To(BeTrue(), "assertion failed on response body")
		rule_id_created = gjson.Get(rule_creation_body_positive, "metadata.resourceId").String()

	})

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("Search all lifecycle rules", func() {
		It("Search all rules", func() {
			defer GinkgoRecover()
			logger.Log.Info("Search all rules")
			get_response_byid_status, get_response_byid_body := storage.GetAllRules(rule_endpoint, token, bucket_id_created)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Scale Test - Search all lifecycle rules", func() {
		It("Scale Test - Search all lifecycle rules", func() {
			if !scaleTestRun {
				Skip("Skipping test because this is not a scale test run")
			}

			logger.Log.Info("Search all rules for scale test")
			Expect(scale_first_bucket_id).ToNot(BeEmpty())
			get_response_byid_status, get_response_byid_body := storage.GetAllRules(rule_endpoint, token, scale_first_bucket_id)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Retrieve the lifecycle rule via GET method using id", func() {
		It("Retrieve the rule via GET method using id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Retrieve the rule via GET method using id")
			res_id := gjson.Get(rule_creation_body_positive, "metadata.resourceId").String()
			get_response_byid_status, get_response_byid_body := storage.GetRuleById(rule_endpoint, bucket_id_created, token, res_id)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Updating the lifecycle rule via PUT method using id", func() {
		It("Updating the rule via PUT method using id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update the rule via PUT method using id")
			res_id := gjson.Get(rule_creation_body_positive, "metadata.resourceId").String()
			get_response_byid_status, get_response_byid_body := storage.PutRuleById(rule_endpoint, bucket_id_created, token, res_id, rule_payload)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Delete the lifecycle rule using id", func() {
		It("Delete the rule using id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Delete the bucket using id")
			// Creation of the rule
			logger.Log.Info("Starting the creation of a lifecycle rule using API")
			rule_name = "delete-rule"
			rule_creation_status, rule_creation_body := storage_utils.RuleCreation(rule_endpoint, token, rule_payload, rule_name, bucket_id_created)
			Expect(rule_creation_status).To(Equal(200), rule_creation_body)
			Expect(strings.Contains(rule_creation_body, `"ruleName":"`+rule_name+`"`)).To(BeTrue(), "assertion failed on response body")
			rule_id := gjson.Get(rule_creation_body, "metadata.resourceId").String()

			// Deletion of rule using id
			logger.Log.Info("Starting the deletion of rule/s using API")
			rule_ids := []string{rule_id}
			storage_utils.DeleteMultipleRulesWithRetry(rule_endpoint, token, bucket_id_created, rule_ids)
		})
	})

	AfterAll(func() {
		// Delete all rules
		logger.Log.Info("Starting the deletion of rule(s) using API")
		rule_id_created = gjson.Get(rule_creation_body_positive, "metadata.resourceId").String()
		rule_ids := []string{rule_id_created}
		storage_utils.DeleteMultipleRulesWithRetry(rule_endpoint, token, bucket_id_created, rule_ids)

		// Delete all buckets
		logger.Log.Info("Starting the deletion of bucket(s) using API")
		resource_ids := []string{bucket_id_created}
		storage_utils.DeleteMultipleBucketsWithRetry(bucket_endpoint, token, resource_ids)

	})
})
