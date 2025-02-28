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

var _ = Describe("Bucket API positive flow", Label("storage", "object_store", "buckets", "bucket_positive"), Ordered, ContinueOnFailure, func() {
	var (
		bucket_name                     string
		bucket_payload                  string
		bucket_id_created               string
		bucket_creation_status_positive int
		bucket_creation_body_positive   string
	)

	BeforeAll(func() {
		// retrieve the required information from test config
		bucket_payload = storage_utils.GetBucketPayload()

		// Bucket to be created
		logger.Log.Info("Starting the creation of a bucket using API")
		bucket_name = "automation-bucket-" + storage_utils.GetRandomString()
		bucket_creation_status_positive, bucket_creation_body_positive = storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)
		Expect(bucket_creation_status_positive).To(Equal(200), bucket_creation_body_positive)
		Expect(strings.Contains(bucket_creation_body_positive, `"name":"`+storage_cloud_account+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
		bucket_id_created = gjson.Get(bucket_creation_body_positive, "metadata.resourceId").String()
		// Validation
		logger.Log.Info("Checking whether bucket is in ready state")
		ValidationSt := storage_utils.CheckBucketProvisionState(bucket_endpoint, token, bucket_id_created)
		Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
	})

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("List all buckets", func() {
		It("List all buckets", func() {
			defer GinkgoRecover()
			logger.Log.Info("List all buckets")
			get_response_byid_status, get_response_byid_body := storage.GetAllBuckets(bucket_endpoint, token)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Retrieve the bucket via GET method using id", func() {
		It("Retrieve the bucket via GET method using id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Retrieve the bucket via GET method using id")
			res_id := gjson.Get(bucket_creation_body_positive, "metadata.resourceId").String()
			get_response_byid_status, get_response_byid_body := storage.GetBucketById(bucket_endpoint, token, res_id)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Retrieve the bucket via GET method using name", func() {
		It("Retrieve the bucket via GET method using name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Retrieve the bucket via GET method using name")
			get_response_byid_status, get_response_byid_body := storage.GetBucketByName(bucket_endpoint, token, storage_cloud_account+"-"+bucket_name)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Delete the bucket using name", func() {
		It("Delete the bucket using name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Delete the bucket using name")
			// Creation of the bucket
			bucket_name_tc := "automation-storage-" + storage_utils.GetRandomString()
			bucket_creation_status, bucket_creation_body := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name_tc)
			Expect(bucket_creation_status).To(Equal(200), bucket_creation_body)
			Expect(strings.Contains(bucket_creation_body, `"name":"`+storage_cloud_account+"-"+bucket_name_tc+`"`)).To(BeTrue(), "assertion failed on response body")
			res_id := gjson.Get(bucket_creation_body, "metadata.resourceId").String()
			// Validation
			logger.Log.Info("Checking whether bucket is in ready state")
			ValidationSt := storage_utils.CheckBucketProvisionState(bucket_endpoint, token, res_id)
			Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
			// Deletion of bucket using name
			delete_response_byname_status, delete_response_byname_body := storage.DeleteBucketByName(bucket_endpoint, token, storage_cloud_account+"-"+bucket_name_tc)
			Expect(delete_response_byname_status).To(Equal(200), delete_response_byname_body)

			logger.Log.Info("Validation of bucket deletion using name")
			DeleteValidationByName := storage_utils.CheckBucketDeletionByName(bucket_endpoint, token, storage_cloud_account+"-"+bucket_name_tc)
			Eventually(DeleteValidationByName, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	AfterAll(func() {
		// Delete all buckets
		logger.Log.Info("Starting the deletion of bucket(s) using API")
		bucket_id_created = gjson.Get(bucket_creation_body_positive, "metadata.resourceId").String()

		resource_ids := []string{bucket_id_created}
		storage_utils.DeleteMultipleBucketsWithRetry(bucket_endpoint, token, resource_ids)
	})
})
