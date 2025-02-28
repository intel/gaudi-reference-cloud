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

var _ = Describe("Bucket API negative flow", Label("storage", "object_store", "buckets", "bucket_negative"), Ordered, ContinueOnFailure, func() {
	var bucket_payload string

	BeforeAll(func() {
		// retrieve the required information from test config
		bucket_payload = storage_utils.GetBucketPayload()
	})

	// Function to create the bucket and returning resource_id
	executeBucketCreation := func(bucket_name string) string {
		bucket_creation_status_positive, bucket_creation_body_positive := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)
		Expect(bucket_creation_status_positive).To(Equal(200), bucket_creation_body_positive)
		Expect(strings.Contains(bucket_creation_body_positive, `"name":"`+storage_cloud_account+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
		bucket_id_created := gjson.Get(bucket_creation_body_positive, "metadata.resourceId").String()
		// Validation
		logger.Log.Info("Checking whether bucket is in ready state")
		Validation := storage_utils.CheckBucketProvisionState(bucket_endpoint, token, bucket_id_created)
		Eventually(Validation, 1*time.Minute, 5*time.Second).Should(BeTrue())
		return bucket_id_created
	}

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("Create bucket with too many characters", func() {
		It("Create bucket with too many characters", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create bucket with too many characters")
			bucket_name := "bucketsss-name-to-validate-the-character-length-for-testing-purpose-attempt" + storage_utils.GetRandomString()
			bucket_creation_status_negative, bucket_creation_body_negative := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)
			Expect(bucket_creation_status_negative).To(Equal(400), bucket_creation_body_negative)
			Expect(strings.Contains(bucket_creation_body_negative, `"message":"invalid resource name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create bucket with already used name", func() {
		It("Create bucket with already used name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create bucket with already used name")
			bucket_name := "automation-bucket-" + storage_utils.GetRandomString()
			// First bucket creation
			res_id := executeBucketCreation(bucket_name)
			// Second bucket creation with same name
			bucket_creation_status_negative1, bucket_creation_body_negative1 := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)
			Expect(bucket_creation_status_negative1).To(Equal(409), bucket_creation_body_negative1)
			Expect(strings.Contains(bucket_creation_body_negative1, `"message":"bucket name `+storage_cloud_account+"-"+bucket_name+` already exists"`)).To(BeTrue(), "assertion failed on response body")

			// Deleting the created buckets
			resource_ids := []string{res_id}
			storage_utils.DeleteMultipleBucketsWithRetry(bucket_endpoint, token, resource_ids)
		})
	})

	When("Create bucket with invalid name(includes special chars)", func() {
		It("Create bucket with invalid name(includes special chars)", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create bucket with invalid name(includes special chars)")
			bucket_name := "automation-test" + storage_utils.GetRandomString() + "!@$%"
			bucket_creation_status_negative, bucket_creation_body_negative := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)
			Expect(bucket_creation_status_negative).To(Equal(400), bucket_creation_body_negative)
			Expect(strings.Contains(bucket_creation_body_negative, `"message":"invalid resource name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create bucket with empty name", func() {
		It("Create bucket with empty name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create bucket with empty name")
			bucket_name := ""
			bucket_creation_status_negative, bucket_creation_body_negative := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)
			Expect(bucket_creation_status_negative).To(Equal(400), bucket_creation_body_negative)
			Expect(strings.Contains(bucket_creation_body_negative, `"message":"missing name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Get a bucket with invalid resource Id", func() {
		It("Get a bucket with invalid resource Id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Get a bucket with invalid resource Id")
			res_id := "invalid-resid"
			get_response_byid_status, get_response_byid_body := storage.GetBucketById(bucket_endpoint, token, res_id)
			Expect(get_response_byid_status).To(Equal(400), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"invalid resourceId"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Get a bucket with invalid name", func() {
		It("Get a bucket with invalid name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Get a bucket with invalid name")
			name := "invalid-name"
			get_response_byname_status, get_response_byname_body := storage.GetBucketByName(bucket_endpoint, token, name)
			Expect(get_response_byname_status).To(Equal(404), get_response_byname_body)
			Expect(strings.Contains(get_response_byname_body, `"message":"no matching records found"`)).To(BeTrue(), "assertion failed on response body")

		})
	})

})

// TO-DO : PUT test cases will be added later as they is not implemented yet by development team.
