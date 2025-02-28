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

var _ = Describe("Principal API positive flow", Label("storage", "object_store", "principal", "principal_positive"), Ordered, ContinueOnFailure, func() {
	var (
		bucket_name                           string
		bucket_name2                          string
		principal_name                        string
		bucket_payload                        string
		principal_payload                     string
		principal_policy_update_payload       string
		principal_policy_multi_update_payload string
		bucket_id_created                     string
		bucket_id_created2                    string
		principal_id_created                  string
		principal_creation_body_positive      string
		principal_creation_status_positive    int
	)

	BeforeAll(func() {
		// retrieve the required information from test config
		bucket_payload = storage_utils.GetBucketPayload()
		principal_payload = storage_utils.GetPrincipalPayload()
		principal_policy_update_payload = storage_utils.GetPrincipalPutPolicyPayload()
		principal_policy_multi_update_payload = storage_utils.GetPrincipalMultiPutPolicyPayload()

		// Bucket to be created
		logger.Log.Info("Starting the creation of first bucket using API")
		bucket_name = "automation-bucket-1-" + storage_utils.GetRandomString()
		bucket_creation_status_positive, bucket_creation_body_positive := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)
		Expect(bucket_creation_status_positive).To(Equal(200), bucket_creation_body_positive)
		Expect(strings.Contains(bucket_creation_body_positive, `"name":"`+storage_cloud_account+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
		bucket_id_created = gjson.Get(bucket_creation_body_positive, "metadata.resourceId").String()
		// Validation
		logger.Log.Info("Checking whether bucket is in ready state")
		ValidationSt := storage_utils.CheckBucketProvisionState(bucket_endpoint, token, bucket_id_created)
		Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())

		// Bucket to be created
		logger.Log.Info("Starting the creation of second bucket using API")
		bucket_name2 = "automation-bucket-2-" + storage_utils.GetRandomString()
		bucket_creation_status_positive2, bucket_creation_body_positive2 := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name2)
		Expect(bucket_creation_status_positive2).To(Equal(200), bucket_creation_body_positive2)
		Expect(strings.Contains(bucket_creation_body_positive2, `"name":"`+storage_cloud_account+"-"+bucket_name2+`"`)).To(BeTrue(), "assertion failed on response body")
		bucket_id_created2 = gjson.Get(bucket_creation_body_positive2, "metadata.resourceId").String()
		// Validation
		logger.Log.Info("Checking whether bucket is in ready state")
		ValidationSt2 := storage_utils.CheckBucketProvisionState(bucket_endpoint, token, bucket_id_created2)
		Eventually(ValidationSt2, 1*time.Minute, 5*time.Second).Should(BeTrue())
	})

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("Starting the creation of a principal with single bucket access", func() {
		It("Starting the creation of a principal with single bucket access", func() {
			defer GinkgoRecover()
			logger.Log.Info("Starting the creation of a principal with single bucket access")
			principal_name = "automation-principal-" + storage_utils.GetRandomString()
			principal_creation_status_positive, principal_creation_body_positive = storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload, principal_name, storage_cloud_account+"-"+bucket_name)
			Expect(principal_creation_status_positive).To(Equal(200), principal_creation_body_positive)
			Expect(strings.Contains(principal_creation_body_positive, `"name":"`+principal_name+`"`)).To(BeTrue(), "assertion failed on response body")
			Expect(strings.Contains(principal_creation_body_positive, `"bucketId":"`+storage_cloud_account+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
			principal_id_created = gjson.Get(principal_creation_body_positive, "metadata.userId").String()
			// Validation
			logger.Log.Info("Checking whether principal is in ready state")
			ValidationSt := storage_utils.CheckUserProvisionState(principal_endpoint, token, principal_id_created)
			Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("List all principals", func() {
		It("List All principals", func() {
			defer GinkgoRecover()
			logger.Log.Info("List All principals")
			get_response_byid_status, get_response_byid_body := storage.GetAllPrincipals(principal_endpoint, token)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Retrieve a principal via GET method using id", func() {
		It("Retrieve a principal via GET method using id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Retrieve a principal via GET method using id")
			//principal_id :=  gjson.Get(principal_creation_body_positive, "metadata.principalId").String()
			get_response_byid_status, get_response_byid_body := storage.GetPrincipalById(principal_endpoint, token, principal_id_created)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Retrieve a principal via GET method using name", func() {
		It("Retrieve a principal via GET method using name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Retrieve a principal via GET method using name")
			get_response_byid_status, get_response_byid_body := storage.GetPrincipalByName(principal_endpoint, token, principal_name)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Ensure the principal has access to the bucket by using the GET method on the bucket", func() {
		It("Ensure the principal has access to the bucket by using the GET method on the bucket", func() {
			defer GinkgoRecover()
			logger.Log.Info("Ensure the principal has access to the bucket by using the GET method on the bucket")
			get_response_byid_status, get_response_byid_body := storage.GetBucketById(bucket_endpoint, token, bucket_id_created)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
			// Validate principal name is same
			allPolicies := gjson.Get(get_response_byid_body, "status.policy.userAccessPolicies").Array()
			principalFound := false
			for _, policy := range allPolicies {
				name := policy.Get("metadata.name").String()
				if name == principal_name {
					principalFound = true
					break
				}
			}
			Expect(principalFound).To(BeTrue(), "User not found in response")
		})
	})

	When("Update credentials for object service principal using Id", func() {
		It("Update credentials for object service principal using Id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update credentials for object service principal using Id")
			put_response_byid_status, put_response_byid_body := storage.PutCredentialPrincipalById(principal_endpoint, token, principal_id_created, "")
			secretKey := gjson.Get(put_response_byid_body, "status.principal.credentials.secretKey").String()
			Expect(put_response_byid_status).To(Equal(200), put_response_byid_body)
			Expect(secretKey).ToNot(BeEmpty(), "secretKey is empty. updating credentials failed.")
		})
	})

	When("Update credentials for object service principal using principal name", func() {
		It("Update credentials for object service principal using principal name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update credentials for object service principal using principal name")
			put_response_byname_status, put_response_byname_body := storage.PutCredentialPrincipalByName(principal_endpoint, token, principal_name, "")
			secretKey := gjson.Get(put_response_byname_body, "status.principal.credentials.secretKey").String()
			Expect(put_response_byname_status).To(Equal(200), put_response_byname_body)
			Expect(secretKey).ToNot(BeEmpty(), "secretKey is empty. updating credentials failed.")
		})
	})

	When("Update policy for object service principal using principalId", func() {
		It("Update policy for object service principal using principalId", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update policy for object service principal using principalId")
			principal_actions := []string{"GetBucketLocation", "GetBucketPolicy", "ListBucket", "ListBucketMultipartUploads", "ListMultipartUploadParts", "GetBucketTagging"}
			principal_permissions := []string{"ReadBucket", "WriteBucket", "DeleteBucket"}
			put_response_byid_status, put_response_byid_body := storage_utils.UpdatePrincipalPolicy(principal_endpoint, token, principal_policy_update_payload,
				principal_name, principal_id_created, storage_cloud_account+"-"+bucket_name, principal_actions, principal_permissions, "Id", true)
			Expect(put_response_byid_status).To(Equal(200), put_response_byid_body)
		})
	})

	When("Update policy for object service principal using name", func() {
		It("Update policy for object service principal using name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update policy for object service principal using name")
			principal_actions := []string{"GetBucketLocation", "ListBucketMultipartUploads", "ListMultipartUploadParts", "GetBucketTagging"}
			principal_permissions := []string{"ReadBucket", "DeleteBucket"}
			put_response_byname_status, put_response_byname_body := storage_utils.UpdatePrincipalPolicy(principal_endpoint, token, principal_policy_update_payload,
				principal_name, principal_id_created, storage_cloud_account+"-"+bucket_name, principal_actions, principal_permissions, "Name", true)
			Expect(put_response_byname_status).To(Equal(200), put_response_byname_body)
		})
	})

	When("Update policy for object service principal, add bucket policy", func() {
		It("Update policy for object service principal, add bucket access", func() {
			defer GinkgoRecover()
			logger.Log.Info("Update policy for object service principal, add bucket access")

			principal_actions := []string{"GetBucketLocation", "ListBucketMultipartUploads", "ListMultipartUploadParts", "GetBucketTagging"}
			principal_permissions := []string{"ReadBucket", "DeleteBucket"}
			put_response_byname_status, put_response_byname_body := storage_utils.UpdatePrincipalMultiPolicy(principal_endpoint, token, principal_policy_multi_update_payload,
				principal_name, principal_id_created, storage_cloud_account+"-"+bucket_name, storage_cloud_account+"-"+bucket_name2, principal_actions, principal_permissions, "Name")
			Expect(put_response_byname_status).To(Equal(200), put_response_byname_body)
		})
	})

	When("Delete the principal using name", func() {
		It("Delete the principal using name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Delete the principal using name")
			// Creation of the principal
			principal_name_tc := "automation-principal-" + storage_utils.GetRandomString()
			principal_creation_status, principal_creation_body := storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload, principal_name_tc, storage_cloud_account+"-"+bucket_name)
			Expect(principal_creation_status).To(Equal(200), principal_creation_body)
			Expect(strings.Contains(principal_creation_body, `"name":"`+principal_name_tc+`"`)).To(BeTrue(), "assertion failed on response body")
			res_id := gjson.Get(principal_creation_body, "metadata.userId").String()
			// Validation
			logger.Log.Info("Checking whether principal is in ready state")
			ValidationSt := storage_utils.CheckUserProvisionState(principal_endpoint, token, res_id)
			Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
			// Deletion of principal using name
			delete_response_byname_status, delete_response_byname_body := storage.DeletePrincipalByName(principal_endpoint, token, principal_name_tc)
			Expect(delete_response_byname_status).To(Equal(200), delete_response_byname_body)

			logger.Log.Info("Validation of principal deletion using name")
			DeleteValidationByName := storage_utils.CheckUserDeletionByName(principal_endpoint, token, principal_name_tc)
			Eventually(DeleteValidationByName, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	AfterAll(func() {
		// Delete all principals
		logger.Log.Info("Starting the deletion of principal/s using API")
		principal_ids := []string{principal_id_created}
		storage_utils.DeleteMultiplePrincipalsWithRetry(principal_endpoint, token, principal_ids)

		// Delete all buckets
		logger.Log.Info("Starting the deletion of bucket/s using API")
		resource_ids := []string{bucket_id_created, bucket_id_created2}
		storage_utils.DeleteMultipleBucketsWithRetry(bucket_endpoint, token, resource_ids)
	})
})
