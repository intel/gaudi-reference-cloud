package storage

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	service_apis "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/service_apis"
	storage_utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Storage API positive flow", Label("storage", "storage_bu"), Ordered, ContinueOnFailure, func() {
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
		volume_creation_status_positive, volume_creation_body_positive = storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, "5GB", vastEnabled)
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
			logger.Log.Info("is instance created? " + strconv.FormatBool(isFileSystemCreated))
			Expect(isFileSystemCreated).Should(BeTrue(), "Instance creation failed with following error "+volume_creation_body_positive)
		})
	})

	When("Increase the file system size", func() {
		It("size increment should be successful", func() {
			logger.Log.Info("Attempt to increase the file system size")
			storage_put_payload := storage_utils.GetStoragePutPayload()
			storage_put_payload = strings.Replace(storage_put_payload, "<<updated-size>>", "7GB", 1)
			fmt.Println(storage_put_payload)
			volume_update_status_positive, volume_update_body_positive := service_apis.PutFilesystemById(storage_endpoint, token, storage_id_created, storage_put_payload)
			Expect(volume_update_status_positive).To(Equal(200), volume_update_body_positive)
			Expect(strings.Contains(volume_update_body_positive, `"storage":"7GB"`)).To(BeTrue(), "assertion failed on response body")
		})
	})

	When("Decrease the file system size", func() {
		It("size increment shouldn't be allowed", func() {
			logger.Log.Info("Attempt to decrease the file system size")
			storage_put_payload := storage_utils.GetStoragePutPayload()
			storage_put_payload = strings.Replace(storage_put_payload, "<<updated-size>>", "3GB", 1)
			volume_update_status_positive, volume_update_body_positive := service_apis.PutFilesystemById(storage_endpoint, token, storage_id_created, storage_put_payload)
			Expect(volume_update_status_positive).To(Equal(400), volume_update_body_positive)
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
