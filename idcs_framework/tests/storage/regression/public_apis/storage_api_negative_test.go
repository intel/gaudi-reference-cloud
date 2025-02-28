package public_apis

import (
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/service_apis"
	storage_utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Storage API Negative flow", Label("storage", "file_store", "storage_negative"), Ordered, ContinueOnFailure, func() {
	var storage_payload string
	var resource_ids []string

	BeforeAll(func() {
		// retrieve the required information from test config
		storage_payload = storage_utils.GetStoragePayload()

	})

	// Function to create the volume and returning resource_id
	executeVolumeCreation := func(volume_name string) string {
		volume_creation_status_positive, volume_creation_body_positive := storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, "1TB", vastEnabled)
		Expect(volume_creation_status_positive).To(Equal(200), volume_creation_body_positive)
		Expect(strings.Contains(volume_creation_body_positive, `"name":"`+volume_name+`"`)).To(BeTrue(), "assertion failed on response body")
		storage_id_created := gjson.Get(volume_creation_body_positive, "metadata.resourceId").String()
		// Validation
		logger.Log.Info("Checking whether volume is in ready state")
		Validation := storage_utils.CheckVolumeProvisionState(storage_endpoint, token, storage_id_created)
		Eventually(Validation, storage_utils.Volume_timeout_in_min*time.Minute, 15*time.Second).Should(BeTrue())
		resource_ids = append(resource_ids, storage_id_created)
		return storage_id_created
	}

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("Create volume with already used name", func() {
		It("Create volume with already used name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create volume with already used name")
			volume_name := "automation-storage-" + storage_utils.GetRandomString()
			// First volume creation
			executeVolumeCreation(volume_name)

			// Temporary fix to avoid backend corruption issues
			logger.Log.Info("Sleeping for 2 minutes...")
			time.Sleep(2 * time.Minute)

			// Second volume creation with same name
			volume_creation_status_negative1, volume_creation_body_negative1 := storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, "1TB", vastEnabled)
			Expect(volume_creation_status_negative1).To(Equal(409), volume_creation_body_negative1)
			Expect(strings.Contains(volume_creation_body_negative1, `"message":"insert: filesystem name `+volume_name+` already exists"`)).To(BeTrue(), "assertion failed on response body")

			// Deleting the created volumes
			storage_utils.DeleteMultipleVolumesWithRetry(storage_endpoint, token, resource_ids)
			// Empty resource_ids array
			resource_ids = []string{}
		})
	})

	When("Create volume with too many characters", func() {
		It("Create volume with too many characters", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create volume with too many characters")
			volume_name := "Filesystem-name-to-validate-the-character-length-for-testing-purpose-attempt" + utils.GetRandomString()
			volume_creation_status_negative, volume_creation_body_negative := storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, "1TB", vastEnabled)
			Expect(volume_creation_status_negative).To(Equal(400), volume_creation_body_negative)
			Expect(strings.Contains(volume_creation_body_negative, `"message":"invalid resource name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create volume with invalid name(includes special chars)", func() {
		It("Create volume with invalid name(includes special chars)", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create volume with invalid name(includes special chars)")
			volume_name := "automation-test" + storage_utils.GetRandomString() + "!@$%"
			volume_creation_status_negative, volume_creation_body_negative := storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, "1TB", vastEnabled)
			Expect(volume_creation_status_negative).To(Equal(400), volume_creation_body_negative)
			Expect(strings.Contains(volume_creation_body_negative, `"message":"invalid resource name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create volume with empty name", func() {
		It("Create volume with empty name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create volume with empty name")
			volume_name := ""
			volume_creation_status_negative, volume_creation_body_negative := storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, "1TB", vastEnabled)
			Expect(volume_creation_status_negative).To(Equal(400), volume_creation_body_negative)
			Expect(strings.Contains(volume_creation_body_negative, `"message":"missing name"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create volume exceeding max size permitted (100TB)", func() {
		It("Create volume exceeding max size permitted", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create volume exceeding max size permitted (100TB)")
			volume_name := "automation-test" + storage_utils.GetRandomString()
			volume_creation_status_negative, volume_creation_body_negative := storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, "101TB", vastEnabled)
			Expect(volume_creation_status_negative).To(Equal(400), volume_creation_body_negative)
			Expect(strings.Contains(volume_creation_body_negative, `"message":"invalid storage size is outside allowed range"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Create volume smaller than min size permitted (5GB)", func() {
		It("Create volume smaller than min size permitted (5GB)", func() {
			defer GinkgoRecover()
			logger.Log.Info("Create volume smaller than min size permitted (5GB)")
			volume_name := "automation-test" + storage_utils.GetRandomString()
			volume_creation_status_negative, volume_creation_body_negative := storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, "0GB", vastEnabled)
			Expect(volume_creation_status_negative).To(Equal(400), volume_creation_body_negative)
			Expect(strings.Contains(volume_creation_body_negative, `"message":"invalid storage size is outside allowed range"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Get a volume with invalid resource Id", func() {
		It("Get a volume with invalid resource Id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Get a volume with invalid resource Id")
			res_id := "invalid-resid"
			get_response_byid_status, get_response_byid_body := storage.GetFilesystemById(storage_endpoint, token, res_id)
			Expect(get_response_byid_status).To(Equal(400), get_response_byid_body)
			Expect(strings.Contains(get_response_byid_body, `"message":"invalid resourceId"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Get a volume with invalid name", func() {
		It("Get a volume with invalid name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Get a volume with invalid name")
			name := "invalid-name"
			get_response_byname_status, get_response_byname_body := storage.GetFilesystemByName(storage_endpoint, token, name)
			Expect(get_response_byname_status).To(Equal(404), get_response_byname_body)
			Expect(strings.Contains(get_response_byname_body, `"message":"no matching records found"`)).To(BeTrue(), "assertion failed on response body")

		})
	})

	When("Create 3 volumes and validate quota check", func() {
		It("Create 3 volumes and validate quota check", func() {
			if (cloudAccount == "440440958336" || cloudAccount == "810990250449") && region == "us-staging-1" {
				Skip("Skipping test due to increased quota in test account")
			}

			defer GinkgoRecover()
			logger.Log.Info("Create 3 volumes and validate quota check")
			resource_ids = []string{}
			// Create first volume
			volume_name1 := "automation-storage-1-" + storage_utils.GetRandomString()
			executeVolumeCreation(volume_name1)
			// Create second volume
			volume_name2 := "automation-storage-2-" + storage_utils.GetRandomString()
			executeVolumeCreation(volume_name2)
			// Create third volume
			volume_name3 := "automation-storage-3-" + storage_utils.GetRandomString()
			volume_creation_status_negative, volume_creation_body_negative := storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name3, "1TB", vastEnabled)
			Expect(volume_creation_status_negative).To(Equal(400), volume_creation_body_negative)
			Expect(strings.Contains(volume_creation_body_negative, `"message":"quota check failed"`)).To(BeTrue(), "assertion failed on response body")

			logger.Log.Info("Starting the deletion of the volume using filesystem API")
			storage_utils.DeleteMultipleVolumesWithRetry(storage_endpoint, token, resource_ids)
			// Empty resource_ids array
			resource_ids = []string{}
		})
	})

	AfterAll(func() {
		// Delete the volume
		logger.Log.Info("Starting the deletion of the volume using filesystem API")
		// Delete volumes if any
		if len(resource_ids) > 0 {
			storage_utils.DeleteMultipleVolumesWithRetry(storage_endpoint, token, resource_ids)
		}

	})

})
