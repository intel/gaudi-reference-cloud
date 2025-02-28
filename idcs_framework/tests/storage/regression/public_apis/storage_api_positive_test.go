package public_apis

import (
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/service_apis"
	storage_utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Storage API positive flow", Label("storage", "file_store", "storage_positive"), Ordered, ContinueOnFailure, func() {
	var volume_name string
	var storage_payload string
	var storage_id_created string
	var volume_creation_status_positive int
	var volume_creation_body_positive string
	var isFileSystemCreated = false

	BeforeAll(func() {
		// retrieve the required information from test config
		storage_payload = storage_utils.GetStoragePayload()

		// Volume to be created
		logger.Log.Info("Starting the creation of a volume using filesystem API")
		volume_name = "automation-storage-" + storage_utils.GetRandomString()
		volume_creation_status_positive, volume_creation_body_positive = storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, "1TB", vastEnabled)
		Expect(volume_creation_status_positive).To(Equal(200), volume_creation_body_positive)
		Expect(strings.Contains(volume_creation_body_positive, `"name":"`+volume_name+`"`)).To(BeTrue(), "assertion failed on response body")
		storage_id_created = gjson.Get(volume_creation_body_positive, "metadata.resourceId").String()
		// Validation
		logger.Log.Info("Checking whether volume is in ready state")
		ValidationSt := storage_utils.CheckVolumeProvisionState(storage_endpoint, token, storage_id_created)
		Eventually(ValidationSt, storage_utils.Volume_timeout_in_min*time.Minute, 15*time.Second).Should(BeTrue())
		isFileSystemCreated = true
	})

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("File system creation and its validation - prerequisite", func() {
		It("Validate the file system creation in before all", func() {
			defer GinkgoRecover()
			logger.Log.Info("is File system created? " + strconv.FormatBool(isFileSystemCreated))
			Expect(isFileSystemCreated).Should(BeTrue(), "File system creation failed with following error "+volume_creation_body_positive)
		})
	})

	When("List all file systems", func() {
		It("List all file systems", func() {
			defer GinkgoRecover()
			logger.Log.Info("List all file systems")
			get_response_byid_status, get_response_byid_body := storage.GetAllFilesystems(storage_endpoint, token)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Retrieve the volume via GET method using id", func() {
		It("Retrieve the volume via GET method using id", func() {
			defer GinkgoRecover()
			logger.Log.Info("Retrieve the volume via GET method using id")
			res_id := gjson.Get(volume_creation_body_positive, "metadata.resourceId").String()
			get_response_byid_status, get_response_byid_body := storage.GetFilesystemById(storage_endpoint, token, res_id)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Retrieve the volume via GET method using name", func() {
		It("Retrieve the volume via GET method using name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Retrieve the volume via GET method using name")
			get_response_byid_status, get_response_byid_body := storage.GetFilesystemByName(storage_endpoint, token, volume_name)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Delete the volume using resource name", func() {
		It("Delete the volume using resource name", func() {
			defer GinkgoRecover()
			logger.Log.Info("Delete the volume using resource name")
			// Creation of the volume
			volume_name_tc := "automation-storage-" + storage_utils.GetRandomString()
			volume_creation_status, volume_creation_body := storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name_tc, "1TB", vastEnabled)
			Expect(volume_creation_status).To(Equal(200), volume_creation_body)
			Expect(strings.Contains(volume_creation_body, `"name":"`+volume_name_tc+`"`)).To(BeTrue(), "assertion failed on response body")
			res_id := gjson.Get(volume_creation_body, "metadata.resourceId").String()
			// Validation
			logger.Log.Info("Checking whether volume is in ready state")
			ValidationSt := storage_utils.CheckVolumeProvisionState(storage_endpoint, token, res_id)
			Eventually(ValidationSt, storage_utils.Volume_timeout_in_min*time.Minute, 15*time.Second).Should(BeTrue())

			// Deletion of volume using name
			storage_utils.DeleteFilesystemByNameWithRetry(storage_endpoint, token, volume_name_tc)
		})
	})

	AfterAll(func() {
		// Delete the volume
		logger.Log.Info("Starting the deletion of the volume using filesystem API")
		storage_id_created = gjson.Get(volume_creation_body_positive, "metadata.resourceId").String()
		storage_ids := []string{storage_id_created}
		storage_utils.DeleteMultipleVolumesWithRetry(storage_endpoint, token, storage_ids)
	})
})

// TO-DO : PUT and search test cases will be added later as they is not implemented yet by development team.
