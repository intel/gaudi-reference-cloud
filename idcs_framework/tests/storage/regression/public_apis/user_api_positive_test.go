package public_apis

import (
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage_utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("User API positive flow", Label("storage", "file_store", "user_positive"), Ordered, ContinueOnFailure, func() {
	var volume_name string
	var storage_payload string
	var storage_id_created string
	var volume_creation_status_positive int
	var volume_creation_body_positive string
	var user_creation_status_positive int
	var user_creation_body_positive string
	var isFileSystemCreated = false
	var username string
	var password string

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

	When("Create the filesystem user", func() {
		It("Create the filesystem user (WEKA)", func() {
			if vastEnabled {
				Skip("Skipping test because VAST is enabled")
			}
			defer GinkgoRecover()
			logger.Log.Info("Create the filesystem user")
			user_creation_status_positive, user_creation_body_positive = storage_utils.UserCreation(user_endpoint, token, volume_name)
			Expect(user_creation_status_positive).To(Equal(200), user_creation_body_positive)
			username = gjson.Get(user_creation_body_positive, "metadata.user").String()
			password = gjson.Get(user_creation_body_positive, "metadata.password").String()
		})
	})

	When("Updating new user credentials", func() {
		It("Updating new user credentials (WEKA)", func() {
			if vastEnabled {
				Skip("Skipping test because VAST is enabled")
			}

			defer GinkgoRecover()
			logger.Log.Info("Updating new user credentials")
			user_creation_status_positive, user_creation_body_positive = storage_utils.UserCreation(user_endpoint, token, volume_name)
			Expect(user_creation_status_positive).To(Equal(200), user_creation_body_positive)
			username = gjson.Get(user_creation_body_positive, "metadata.user").String()
			password = gjson.Get(user_creation_body_positive, "metadata.password").String()
		})
	})

	AfterAll(func() {
		// Delete the volume
		logger.Log.Info("Starting the deletion of the volume using filesystem API")
		storage_id_created = gjson.Get(volume_creation_body_positive, "metadata.resourceId").String()
		storage_ids := []string{storage_id_created}

		storage_utils.DeleteMultipleVolumesWithRetry(storage_endpoint, token, storage_ids)

		Expect(password).NotTo(BeNil())
		Expect(username).NotTo(BeNil())
	})
})

// TO-DO : PUT and search test cases will be added later as they is not implemented yet by development team.
