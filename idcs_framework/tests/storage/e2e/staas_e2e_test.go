package e2e

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

var _ = Describe("Storage STaaS E2E Bucket Test Execution", Ordered, ContinueOnFailure, Label("staas-e2e", "storage-staas-e2e", "staas-e2e-os", "object_store"), func() {
	var instance_creation_status int
	var instance_creation_body string
	var vm_name string
	var vm_payload string
	var bucket_name string
	var bucket_payload string
	var principal_name string
	var principal_payload string
	var principal_id_created string
	var instance_id_created string
	var bucket_id_created string
	var access_key string
	var secret_key string
	var endpoint string
	var bucketClusterName string
	var principalClusterName string
	var instance_group_payload_os string
	var instance_group_name string

	const default_bucketops_sleep_time_in_sec = 90

	BeforeAll(func() {
		instance_group_payload_os = storage_utils.GetStInstanceGroupPayloadOS()
		vm_name = "automation-vm-" + storage_utils.GetRandomString()
		vm_payload = storage_utils.GetStInstancePayload()
		bucket_payload = storage_utils.GetBucketPayload()
		principal_payload = storage_utils.GetPrincipalPayload()
	})

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("Create Bucket", func() {
		It("Create Bucket", func() {
			// Bucket to be created
			logger.Log.Info("Starting the creation of bucket using API")
			bucket_name = "automation-bucket-1-" + storage_utils.GetRandomString()
			bucket_creation_status_positive, bucket_creation_body_positive := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)
			Expect(bucket_creation_status_positive).To(Equal(200), bucket_creation_body_positive)
			Expect(strings.Contains(bucket_creation_body_positive, `"name":"`+cloudAccount+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
			bucket_id_created = gjson.Get(bucket_creation_body_positive, "metadata.resourceId").String()

			// Validation
			logger.Log.Info("Checking whether bucket is in ready state")
			ValidationSt := storage_utils.CheckBucketProvisionState(bucket_endpoint, token, bucket_id_created)
			Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Get Bucket", func() {
		It("Get Bucket", func() {
			logger.Log.Info("Starting the creation of bucket using API")
			get_response_byid_status, get_response_byid_body := storage.GetBucketById(bucket_endpoint, token, bucket_id_created)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)

			// populate endpoint url
			endpoint = gjson.Get(get_response_byid_body, "status.cluster.accessEndpoint").String()
			logger.Log.Info("endpoint")
			logger.Log.Info(endpoint)

			// Verify that the cluster which the bucket is created on is a general purpose cluster
			bucketClusterName = gjson.Get(get_response_byid_body, "status.cluster.clusterName").String()
			logger.Logf.Infof("Bucket was created on cluster: %s", bucketClusterName)
			Expect(storage_utils.IsGPCluster(region, bucketClusterName)).To(Equal(true))
		})
	})

	When("Create Principal", func() {
		It("Create Principal", func() {
			logger.Log.Info("Starting the creation of a principal with single bucket access")
			principal_name = "automation-principal-" + storage_utils.GetRandomString()
			principal_creation_status_positive, principal_creation_body_positive := storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload,
				principal_name, cloudAccount+"-"+bucket_name)
			Expect(principal_creation_status_positive).To(Equal(200), principal_creation_body_positive)
			Expect(strings.Contains(principal_creation_body_positive, `"name":"`+principal_name+`"`)).To(BeTrue(), "assertion failed on response body")
			Expect(strings.Contains(principal_creation_body_positive, `"bucketId":"`+cloudAccount+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
			principal_id_created = gjson.Get(principal_creation_body_positive, "metadata.userId").String()

			// Validation
			logger.Log.Info("Checking whether principal is in ready state")
			ValidationSt := storage_utils.CheckUserProvisionState(principal_endpoint, token, principal_id_created)
			Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())

			access_key = gjson.Get(principal_creation_body_positive, "status.principal.credentials.accessKey").String()
			secret_key = gjson.Get(principal_creation_body_positive, "status.principal.credentials.secretKey").String()

			// Verify that the principal is created on a general purpose cluster
			principalClusterName = gjson.Get(principal_creation_body_positive, "status.principal.cluster.clusterName").String()
			logger.Logf.Infof("Principal was created on cluster: %s", principalClusterName)
			Expect(storage_utils.IsGPCluster(region, principalClusterName)).To(Equal(true))

			// Verify that the principal and bucket are created on the same general purpose cluster
			Expect(principalClusterName).To(Equal(bucketClusterName))
		})
	})

	When("Create Instance, execute bucket operation scripts via cloud-init", func() {
		It("Create Instance, execute bucket operation scripts via cloud-init", func() {
			if superComputeRun {
				Skip("Skipping test because this is a Supercompute test run...")
			}

			logger.Log.Info("Starting Instance Creation flow via Instance creation API...")
			// set access credentials
			vm_payload = strings.Replace(vm_payload, "${BUCKET_ID}", cloudAccount+"-"+bucket_name, -1)
			vm_payload = strings.Replace(vm_payload, "${ACCESS_KEY}", access_key, -1)
			vm_payload = strings.Replace(vm_payload, "${SECRET_KEY}", secret_key, -1)
			vm_payload = strings.Replace(vm_payload, "${ENDPOINT_URL}", endpoint, -1)

			instance_creation_status, instance_creation_body = storage.CreateInstance(instance_endpoint, token, vm_payload, vm_name,
				instanceType, sshkeyName, vnet, machineImage, availabilityZone)

			Expect(instance_creation_status).To(Equal(200), "assertion failed on response code")
			Expect(strings.Contains(instance_creation_body, `"name":"`+vm_name+`"`)).To(BeTrue(), "assertion failed on response body")
			instance_id_created = gjson.Get(instance_creation_body, "metadata.resourceId").String()

			logger.Log.Info("Checking whether instance is in ready state")

			instanceValidation := storage.CheckInstanceState(instance_endpoint, token, instance_id_created, "Ready")
			Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())

			maxWaitTimeInMin := max_vm_creation_timeout_in_min
			if strings.Contains(strings.ToLower(instanceType), "bm") || strings.Contains(strings.ToLower(machineImage), "metal") || strings.Contains(strings.ToLower(machineImage), "gaudi") {
				maxWaitTimeInMin = 45
			}
			Eventually(instanceValidation, time.Duration(maxWaitTimeInMin)*time.Minute, 60*time.Second).Should(BeTrue())
			_, get_response := storage.GetInstanceById(instance_endpoint, token, instance_id_created)

			instance_creation_body = get_response
		})
	})

	When("Create Gaudi instance group", func() {
		It("Create Gaudi instance group", func() {
			if !superComputeRun {
				Skip("Skipping test because this is not a Supercompute test run...")
			}

			logger.Logf.Infof("instance_group_endpoint: %s", instance_group_endpoint)
			logger.Logf.Infof("instance_group_payload: %s", instance_group_payload_os)

			instance_group_payload_os = strings.Replace(instance_group_payload_os, "${BUCKET_ID}", cloudAccount+"-"+bucket_name, -1)
			instance_group_payload_os = strings.Replace(instance_group_payload_os, "${ACCESS_KEY}", access_key, -1)
			instance_group_payload_os = strings.Replace(instance_group_payload_os, "${SECRET_KEY}", secret_key, -1)
			instance_group_payload_os = strings.Replace(instance_group_payload_os, "${ENDPOINT_URL}", endpoint, -1)

			var err error
			instance_group_name, instance_creation_body, err = storage_utils.CreateGaudiInstanceGroup(instanceType, machineImage,
				instance_group_endpoint, token, instance_group_payload_os, sshkeyName, vnet, availabilityZone)
			Expect(err).NotTo(HaveOccurred())

			// Get the resource IDs of first machine in the instance group
			get_response_byname_status, get_response_byname_body := storage.GetInstancesWithinGroup(instance_endpoint, token, instance_group_name)
			Expect(get_response_byname_status).To(Equal(200))

			// Get resource IDs for first two machines in instance group
			resource_id1 := gjson.Get(get_response_byname_body, "items.0.metadata.resourceId").String()
			logger.Logf.Infof("Resource ID of Gaudi machine 1: [%s] ", resource_id1)

			// Get instance information using resource IDs
			var status int
			status, instance_creation_body = storage.GetInstanceById(instance_endpoint, token, resource_id1)
			Expect(status).To(Equal(200))
		})
	})

	When("Verify cloud init is successful", func() {
		It("Verify cloud init is successful", func() {
			sleepTimeInSec := 150
			if strings.Contains(strings.ToLower(instanceType), "bm") || strings.Contains(strings.ToLower(machineImage), "metal") || strings.Contains(strings.ToLower(machineImage), "gaudi") {
				sleepTimeInSec = 300
			}
			logger.Logf.Infof("Waiting for cloud init scripts, sleep for %d seconds", sleepTimeInSec)
			time.Sleep(time.Duration(sleepTimeInSec) * time.Second)

			// Refresh token before creating the instance group
			_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)

			logger.Log.Info("Checking cloud init logs")
			res, err := storage_utils.CheckCloudInit(instance_creation_body)
			Expect(err).To(BeNil())
			Expect(res).To(ContainSubstring("STAAS script completed"))
		})
	})

	When("Upload Objects to Bucket", func() {
		It("Upload Objects to Bucket", func() {
			logger.Log.Info("Upload Objects to Bucket")
			res, err := storage_utils.UploadBucket(instance_creation_body)

			Expect(err).To(BeNil())
			Expect(res).NotTo(ContainSubstring("failed"))
		})
	})

	When("Delete Bucket that is not empty", func() {
		It("Delete Bucket that is not empty", func() {
			logger.Log.Info("Delete Bucket that is not empty")

			sleep_time_in_sec := default_bucketops_sleep_time_in_sec
			if testEnv == "production" {
				// For production environments we have to wait at least 10 minutes for the bucket state to update
				sleep_time_in_sec = 600
			}
			logger.Logf.Infof("Sleeping for %d seconds to allow bucket state to be updated after adding objects...", sleep_time_in_sec)
			time.Sleep(time.Duration(sleep_time_in_sec) * time.Second)

			delete_response_byname_status, delete_response_byname_body := storage.DeleteBucketByName(bucket_endpoint, token, cloudAccount+"-"+bucket_name)
			Expect(delete_response_byname_status).NotTo(Equal(200), delete_response_byname_body)
		})
	})

	When("Delete all objects versions from Bucket", func() {
		It("Delete all objects from Bucket", func() {
			logger.Log.Info("Delete all objects from Bucket")
			res, err := storage_utils.EmptyBucket(instance_creation_body)
			Expect(err).To(BeNil())
			Expect(res).NotTo(ContainSubstring("failed"))
		})
	})

	When("Delete Bucket that is empty", func() {
		It("Delete Bucket that is empty", func() {
			logger.Log.Info("Delete Bucket that is empty")

			sleep_time_in_sec := default_bucketops_sleep_time_in_sec
			if testEnv == "production" {
				// For production environments we have to wait at least 10 minutes for the bucket state to update after all objects are deleted
				sleep_time_in_sec = 600
			}
			logger.Logf.Infof("Sleeping for %d seconds to allow bucket state to be updated after all objects are deleted...", sleep_time_in_sec)
			time.Sleep(time.Duration(sleep_time_in_sec) * time.Second)

			delete_response_byname_status, delete_response_byname_body := storage.DeleteBucketByName(bucket_endpoint, token, cloudAccount+"-"+bucket_name)
			Expect(delete_response_byname_status).To(Equal(200), delete_response_byname_body)

			logger.Log.Info("Validation of bucket deletion using name")
			DeleteValidationByName := storage_utils.CheckBucketDeletionByName(bucket_endpoint, token, cloudAccount+"-"+bucket_name)
			Eventually(DeleteValidationByName, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Delete Principal", func() {
		It("Delete Principal", func() {
			// Deletion of principal using name
			delete_response_byname_status, delete_response_byname_body := storage.DeletePrincipalByName(principal_endpoint, token, principal_name)
			Expect(delete_response_byname_status).To(Equal(200), delete_response_byname_body)

			logger.Log.Info("Validation of principal deletion using name")
			DeleteValidationByName := storage_utils.CheckUserDeletionByName(principal_endpoint, token, principal_name)
			Eventually(DeleteValidationByName, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Delete the created instance", func() {
		It("Delete the created instance", func() {
			if superComputeRun {
				Skip("Skipping test because this is a Supercompute test run...")
			}

			logger.Log.Info("Remove the instance via DELETE api using resource id")
			instance_id_created := gjson.Get(instance_creation_body, "metadata.resourceId").String()
			delete_response_byid_status, _ := storage.DeleteInstanceById(instance_endpoint, token, instance_id_created)
			Expect(delete_response_byid_status).To(Equal(200), "assertion failed on response code")

			logger.Log.Info("Validation of Instance Deletion")
			instanceValidation := storage.CheckInstanceDeletionById(instance_endpoint, token, instance_id_created)
			Eventually(instanceValidation, time.Duration(max_vm_deletion_timeout_in_min)*time.Minute, 30*time.Second).Should(BeTrue())
		})
	})

	// Delete Instance Group
	When("Delete Instance Group", func() {
		It("Delete Instance Group", func() {
			if !superComputeRun {
				Skip("Skipping test because this is not a Supercompute test run...")
			}

			err := storage_utils.DeleteInstanceGroup(instance_group_name, instance_group_endpoint, token)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	// The following set of tests do the following:
	// - Create bucket and principal 1
	// - Verify that both were created on the same GP cluster
	// - Delete principal 1
	// - Create principal 2
	// - Verify that new principal was created on the same GP cluster as the bucket.
	// - Delete bucket
	// - Delete principal 2
	When("Verify General Purpose Cluster - Create bucket and principal", func() {
		It("Verify General Purpose Cluster - Create bucket and principal", func() {
			// Create bucket
			logger.Log.Info("Starting the creation of bucket using API")
			bucket_name = "automation-bucket-2-" + storage_utils.GetRandomString()
			bucket_creation_status_positive, bucket_creation_body_positive := storage_utils.BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)
			Expect(bucket_creation_status_positive).To(Equal(200), bucket_creation_body_positive)
			Expect(strings.Contains(bucket_creation_body_positive, `"name":"`+cloudAccount+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
			bucket_id_created = gjson.Get(bucket_creation_body_positive, "metadata.resourceId").String()

			// Wait for bucket to get to ready state
			logger.Log.Info("Checking whether bucket is in ready state")
			ValidationSt := storage_utils.CheckBucketProvisionState(bucket_endpoint, token, bucket_id_created)
			Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())

			// Get Bucket info
			get_response_byid_status, get_response_byid_body := storage.GetBucketById(bucket_endpoint, token, bucket_id_created)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)

			// Verify that the bucket was created on a general purpose cluster
			bucketClusterName = gjson.Get(get_response_byid_body, "status.cluster.clusterName").String()
			logger.Logf.Infof("Bucket was created on cluster: %s", bucketClusterName)
			Expect(storage_utils.IsGPCluster(region, bucketClusterName)).To(Equal(true))

			// Create principal 1
			logger.Log.Info("Starting the creation of a principal with single bucket access")
			principal_name = "automation-principal-" + storage_utils.GetRandomString()
			principal_creation_status_positive, principal_creation_body_positive := storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload,
				principal_name, cloudAccount+"-"+bucket_name)
			Expect(principal_creation_status_positive).To(Equal(200), principal_creation_body_positive)
			Expect(strings.Contains(principal_creation_body_positive, `"name":"`+principal_name+`"`)).To(BeTrue(), "assertion failed on response body")
			Expect(strings.Contains(principal_creation_body_positive, `"bucketId":"`+cloudAccount+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
			principal_id_created = gjson.Get(principal_creation_body_positive, "metadata.userId").String()

			// Verify that the principal was created on a general purpose cluster
			principalClusterName = gjson.Get(principal_creation_body_positive, "status.principal.cluster.clusterName").String()
			logger.Logf.Infof("Principal was created on cluster: %s", principalClusterName)
			Expect(principalClusterName).To(Equal(bucketClusterName))

			// Verify that the principal and the bucket were created on the same GP cluster
			Expect(principalClusterName).To(Equal(bucketClusterName))

			// Check that principal 1 gets to ready state
			logger.Log.Info("Checking whether principal is in ready state")
			ValidationSt = storage_utils.CheckUserProvisionState(principal_endpoint, token, principal_id_created)
			Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Verify General Purpose Cluster - Delete Principal 1", func() {
		It("Verify General Purpose Cluster Part 2", func() {
			// Delete principal 1
			delete_response_byname_status, delete_response_byname_body := storage.DeletePrincipalByName(principal_endpoint, token, principal_name)
			Expect(delete_response_byname_status).To(Equal(200), delete_response_byname_body)

			logger.Log.Info("Validation of principal deletion using name")
			DeleteValidationByName := storage_utils.CheckUserDeletionByName(principal_endpoint, token, principal_name)
			Eventually(DeleteValidationByName, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Verify General Purpose Cluster - Create Principal 2", func() {
		It("Verify General Purpose Cluster - Create Principal 2", func() {
			// Create principal 2
			logger.Log.Info("Starting the creation of a principal with single bucket access")
			principal_name = "automation-principal-" + storage_utils.GetRandomString()
			principal_creation_status_positive, principal_creation_body_positive := storage_utils.PrincipalCreation(principal_endpoint, token, principal_payload,
				principal_name, cloudAccount+"-"+bucket_name)
			Expect(principal_creation_status_positive).To(Equal(200), principal_creation_body_positive)
			Expect(strings.Contains(principal_creation_body_positive, `"name":"`+principal_name+`"`)).To(BeTrue(), "assertion failed on response body")
			Expect(strings.Contains(principal_creation_body_positive, `"bucketId":"`+cloudAccount+"-"+bucket_name+`"`)).To(BeTrue(), "assertion failed on response body")
			principal_id_created = gjson.Get(principal_creation_body_positive, "metadata.userId").String()

			// Verify that the cluster the bucket is created on is a general purpose cluster
			principalClusterName = gjson.Get(principal_creation_body_positive, "status.principal.cluster.clusterName").String()
			logger.Logf.Infof("Principal 2 was created on cluster: %s", principalClusterName)
			Expect(storage_utils.IsGPCluster(region, principalClusterName)).To(Equal(true))

			// Verify that the principal 2 was created on the same GP cluster as the bucket
			Expect(principalClusterName).To(Equal(bucketClusterName))

			// Check that principal 2 gets to ready state
			logger.Log.Info("Checking whether principal is in ready state")
			ValidationSt := storage_utils.CheckUserProvisionState(principal_endpoint, token, principal_id_created)
			Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Verify General Purpose Cluster - Delete Bucket", func() {
		It("Verify General Purpose Cluster - Delete Bucket", func() {
			// Delete bucket
			logger.Log.Info("Delete Bucket that is empty")
			delete_response_byname_status, delete_response_byname_body := storage.DeleteBucketByName(bucket_endpoint, token, cloudAccount+"-"+bucket_name)
			Expect(delete_response_byname_status).To(Equal(200), delete_response_byname_body)

			logger.Log.Info("Validation of bucket deletion using name")
			DeleteValidationByName := storage_utils.CheckBucketDeletionByName(bucket_endpoint, token, cloudAccount+"-"+bucket_name)
			Eventually(DeleteValidationByName, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Verify General Purpose Cluster - Delete Principal 2", func() {
		It("Verify General Purpose Cluster  - Delete Principal 2", func() {
			// Delete principal 2
			delete_response_byname_status, delete_response_byname_body := storage.DeletePrincipalByName(principal_endpoint, token, principal_name)
			Expect(delete_response_byname_status).To(Equal(200), delete_response_byname_body)

			logger.Log.Info("Validation of principal deletion using name")
			DeleteValidationByName := storage_utils.CheckUserDeletionByName(principal_endpoint, token, principal_name)
			Eventually(DeleteValidationByName, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})
})
