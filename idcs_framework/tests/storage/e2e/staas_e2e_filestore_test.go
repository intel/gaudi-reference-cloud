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

var _ = Describe("Storage STaaS File Storage E2E Test Execution", Ordered, ContinueOnFailure, Label("staas-e2e", "storage-staas-e2e", "staas-e2e-fs", "file_store"), func() {
	var vm_payload string
	var volume_name1 string
	var volume_name2 string
	var filesystem_storage_payload string
	var filesystem_storage_payload_fast string
	var storage_id_created1 string
	var storage_id_created2 string
	var volume_creation_status_positive int
	var volume_creation_body_positive1 string
	var volume_creation_body_positive2 string
	var user_creation_status_positive int
	var user_creation_body_positive string
	var username1 string
	var password1 string
	var instance_creation_body1 string
	var instance_creation_body2 string
	var prevStepOk bool
	var weka_url1 string
	var weka_url2 string
	var mountPath1 string
	var mountPath2 string
	var instance_group_payload_fs string
	var vast_url string

	BeforeAll(func() {
		logger.Log.Info("Starting file storage E2E tests")

		instance_group_payload_fs = storage_utils.GetStInstanceGroupPayloadFS()
		vm_payload = storage_utils.GetStInstancePayloadFS()
		filesystem_storage_payload = storage_utils.GetStFilesystemStoragePayload()
		filesystem_storage_payload_fast = storage_utils.GetStFilesystemStoragePayloadFast()
		mountPath1 = "/mnt/test1"
		mountPath2 = "/mnt/test2"
	})

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("Create Volumes", func() {
		It("Create Volumes", func() {
			prevStepOk = false

			// Volume to be created
			logger.Log.Info("Starting the creation of a volume using filesystem API")

			volume_name1 = "automation-storage-" + storage_utils.GetRandomString()
			logger.Logf.Infof("volume_name1: %s", volume_name1)

			payload := filesystem_storage_payload
			if isGaudiInstance {
				// Create a volume on a fast cluster for Gaudi machines
				logger.Log.Info("Create a fast volume for use with Gaudi instances")
				payload = filesystem_storage_payload_fast
			}

			volume_creation_status_positive, volume_creation_body_positive1 = storage_utils.FilesystemCreation(storage_endpoint, token, payload, volume_name1, "1TB", vastEnabled)
			Expect(volume_creation_status_positive).To(Equal(200), volume_creation_body_positive1)

			Expect(strings.Contains(volume_creation_body_positive1, `"name":"`+volume_name1+`"`)).To(BeTrue(), "assertion failed on response body")
			storage_id_created1 = gjson.Get(volume_creation_body_positive1, "metadata.resourceId").String()

			// Validation
			logger.Log.Info("Checking whether volume is in ready state")
			ValidationSt := storage_utils.CheckVolumeProvisionState(storage_endpoint, token, storage_id_created1)
			Eventually(ValidationSt, storage_utils.Volume_timeout_in_min*time.Minute, 30*time.Second).Should(BeTrue())

			if vastEnabled {
				vast_url = storage_utils.GetVastUrl(storage_endpoint, token, storage_id_created1)
				logger.Logf.Infof("VAST URL for volume 1: %s", vast_url)
			} else {
				weka_url1 = storage_utils.GetWekaUrl(storage_endpoint, token, storage_id_created1)
				logger.Logf.Infof("WEKA URL for volume 1: [%s]", weka_url1)
			}

			// Create second volume
			volume_name2 = "automation-storage-" + storage_utils.GetRandomString()
			logger.Logf.Infof("Volume 2: %s", volume_name2)

			volume_creation_status_positive, volume_creation_body_positive2 = storage_utils.FilesystemCreation(storage_endpoint, token, payload, volume_name2, "2TB", vastEnabled)
			Expect(volume_creation_status_positive).To(Equal(200), volume_creation_body_positive2)
			Expect(strings.Contains(volume_creation_body_positive2, `"name":"`+volume_name2+`"`)).To(BeTrue(), "assertion failed on response body")
			storage_id_created2 = gjson.Get(volume_creation_body_positive2, "metadata.resourceId").String()

			// Validation
			logger.Log.Info("Checking whether volume is in ready state")
			ValidationSt = storage_utils.CheckVolumeProvisionState(storage_endpoint, token, storage_id_created2)
			Eventually(ValidationSt, storage_utils.Volume_timeout_in_min*time.Minute, 30*time.Second).Should(BeTrue())

			if !vastEnabled {
				weka_url2 = storage_utils.GetWekaUrl(storage_endpoint, token, storage_id_created2)
				logger.Logf.Infof("WEKA URL for volume 2: [%s]", weka_url2)
			}

			prevStepOk = true
		})
	})

	When("Create File System User", func() {
		It("Create File System User", func() {
			if !prevStepOk || vastEnabled {
				Skip("Skipping test because previous step failed or VAST volumes are enabled")
			}

			// Create File System User
			logger.Log.Info("Create the filesystem user")
			user_creation_status_positive, user_creation_body_positive = storage_utils.UserCreation(user_endpoint, token, volume_name1)

			if user_creation_status_positive != 200 {
				prevStepOk = false
			}
			Expect(user_creation_status_positive).To(Equal(200), user_creation_body_positive)

			username1 = gjson.Get(user_creation_body_positive, "user").String()
			password1 = gjson.Get(user_creation_body_positive, "password").String()

			Expect(password1).NotTo(BeNil())
			Expect(username1).NotTo(BeNil())
		})
	})

	When("Create Instance1", func() {
		It("Create Instance1", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			if superComputeRun {
				Skip("Skipping test because this is a Supercompute test run")
			}

			var err error
			instance_creation_body1, err = storage_utils.CreateVM(instanceType, machineImage, instance_endpoint,
				token, vm_payload, sshkeyName, vnet, availabilityZone, max_vm_creation_timeout_in_min)

			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("Create Gaudi instance group", func() {
		It("Create Gaudi instance group", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			if !superComputeRun {
				Skip("Skipping test because this is not a Supercompute test run...")
			}

			logger.Logf.Infof("instance_group_endpoint: %s", instance_group_endpoint)
			logger.Logf.Infof("instance_group_payload: %s", instance_group_payload_fs)

			// Refresh token before creating the instance group
			_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)

			var err error
			var ig_name string
			ig_name, instance_creation_body1, err = storage_utils.CreateGaudiInstanceGroup(instanceType, machineImage,
				instance_group_endpoint, token, instance_group_payload_fs, sshkeyName, vnet, availabilityZone)

			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())

			// Refresh token
			_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)

			// Get the resource IDs of machines in the instance group with name ig_name
			get_response_byname_status, get_response_byname_body := storage.GetInstancesWithinGroup(instance_endpoint, token, ig_name)
			prevStepOk = (get_response_byname_status == 200)
			Expect(err).NotTo(HaveOccurred())

			// Get resource IDs for first two machines in instance group
			resource_id1 := gjson.Get(get_response_byname_body, "items.0.metadata.resourceId").String()
			logger.Logf.Infof("Resource ID of Gaudi machine 1: [%s] ", resource_id1)

			resource_id2 := gjson.Get(get_response_byname_body, "items.1.metadata.resourceId").String()
			logger.Logf.Infof("Resource ID of Gaudi machine 2: [%s] ", resource_id2)

			// Get instance information using resource IDs
			var status int
			status, instance_creation_body1 = storage.GetInstanceById(instance_endpoint, token, resource_id1)
			prevStepOk = (status == 200)
			Expect(err).NotTo(HaveOccurred())

			status, instance_creation_body2 = storage.GetInstanceById(instance_endpoint, token, resource_id2)
			prevStepOk = (status == 200)
			Expect(err).NotTo(HaveOccurred())

		})
	})

	When("Wait for cloud-init to complete on Instance1", func() {
		It("Wait for cloud-init to complete on Instance1", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			err := storage_utils.WaitForCloudInit(instance_creation_body1, 600)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("Mount Volume1 on Instance1", func() {
		It("Mount Volume1 on Instance1", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			network := "udp"
			if isGaudiInstance {
				// Use fast network for Gaudi machines
				logger.Log.Info("Using fast network for mounting on Gaudi instances")
				network = "storage0-tenant"
			}

			// Mount first volume to VM
			logger.Log.Info("Mounting Volume1 on Instance1...")

			var err error
			if vastEnabled {
				err = storage_utils.MountVASTVolume(instance_creation_body1, mountPath1, volume_name1, vast_url)
			} else {
				err = storage_utils.MountWEKAVolume(instance_creation_body1, cloudAccount, password1, mountPath1,
					volume_name1, weka_url1, network)
			}
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())

			logger.Log.Info("Waiting for Volume1 mount to complete on Instance1...")
			err = storage_utils.WaitForMount(instance_creation_body1, mountPath1)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
			logger.Log.Info("Volume1 was mounted successfully on Instance1")
		})
	})

	When("Create test file on Volume1 from Instance1", func() {
		It("Create test file on Volume1 from Instance1", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			logger.Log.Info("Creating test file on Volume1 from Instance1...")
			err := storage_utils.CreateTestFileOnVolume(instance_creation_body1, mountPath1)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("Run file storage tests on Volume1 from Instance1", func() {
		It("Run file storage tests on Volume1 from Instance1", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			logger.Log.Info("Running tests on first volume...")
			err := storage_utils.RunTestsOnVolume(instance_creation_body1, mountPath1)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("Mount Volume2 on Instance1", func() {
		It("Mount Volume2 on Instance1", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			network := "udp"
			if isGaudiInstance {
				// Use fast network for Gaudi machines
				logger.Log.Info("Using fast network for mounting on Gaudi instances")
				network = "storage0-tenant"
			}

			logger.Log.Info("Mounting Volume2 on Instance1...")

			var err error
			if vastEnabled {
				err = storage_utils.MountVASTVolume(instance_creation_body1, mountPath2, volume_name2, vast_url)
			} else {
				err = storage_utils.MountWEKAVolume(instance_creation_body1, cloudAccount, password1, mountPath2,
					volume_name2, weka_url2, network)
			}
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())

			logger.Log.Info("Waiting for Volume2 mount to complete on Instance1...")
			err = storage_utils.WaitForMount(instance_creation_body1, mountPath2)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
			logger.Log.Info("Volume2 was mounted successfully on Instance1")
		})
	})

	When("Run file storage tests on Volume2 from Instance1", func() {
		It("Run file storage tests on Volume2 from Instance1", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			logger.Log.Info("Running tests on second volume...")
			err := storage_utils.RunTestsOnVolume(instance_creation_body1, mountPath2)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("Unmount Volume2 from Instance1", func() {
		It("Unmount second volume", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			err := storage_utils.UnmountVolume(instance_creation_body1, mountPath1)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())

			err = storage_utils.UnmountVolume(instance_creation_body1, mountPath2)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("Create Instance2", func() {
		It("Create Instance2", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			if superComputeRun {
				Skip("Skipping test because this is a Supercompute test run")
			}

			var err error
			instance_creation_body2, err = storage_utils.CreateVM("vm-spr-sml", "ubuntu-2204-jammy-v20240308", instance_endpoint,
				token, vm_payload, sshkeyName, vnet, availabilityZone, max_vm_creation_timeout_in_min)

			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("Wait for cloud-init to complete on Instance2", func() {
		It("Wait for cloud-init to complete on Instance2", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			err := storage_utils.WaitForCloudInit(instance_creation_body2, 600)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("Mount Volume1 on Instance2", func() {
		It("Mount Volume1 on Instance2", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			network := "udp"
			if isGaudiInstance {
				// Use fast network for Gaudi machines
				logger.Log.Info("Using fast network for mounting on Gaudi instances")
				network = "storage0-tenant"
			}

			// Mount Volume1 on Instance2
			logger.Log.Info("Mounting volume1 on Instance2...")

			var err error
			if vastEnabled {
				err = storage_utils.MountVASTVolume(instance_creation_body2, mountPath1, volume_name1, vast_url)
			} else {
				err = storage_utils.MountWEKAVolume(instance_creation_body2, cloudAccount, password1, mountPath1,
					volume_name1, weka_url1, network)
			}
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())

			logger.Log.Info("Waiting for volume1 mount to complete on Instance2...")
			err = storage_utils.WaitForMount(instance_creation_body2, mountPath1)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
			logger.Log.Info("Volume1 was mounted successfully on Instance2")
		})
	})

	When("Check that shared file is visible to Instance2", func() {
		It("Check that shared file is visible to Instance2", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			logger.Log.Info("Creating test file on Volume1 from Instance1...")
			err := storage_utils.CheckTestFileOnVolume(instance_creation_body2, mountPath1)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("Run file storage tests on Volume1 from Instance1 and Instance2", func() {
		It("Run file storage tests on Volume1 from Instance1 and Instance2", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			logger.Log.Info("Running tests on Volume1 from Instance2...")
			err := storage_utils.RunTestsOnVolume(instance_creation_body1, mountPath1)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())

			logger.Log.Info("Running tests on Volume1 from Instance2...")
			err = storage_utils.RunTestsOnVolume(instance_creation_body2, mountPath1)
			prevStepOk = (err == nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	// Delete Instance2 with Volume2 still mounted.
	When("Delete Instance2", func() {
		It("Delete Instance2", func() {
			if superComputeRun {
				Skip("Skipping test because this is a Supercompute test run...")
			}

			err := storage_utils.DeleteVM(instance_creation_body2, instance_endpoint, token, max_vm_deletion_timeout_in_min)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	// Delete Volume2 after Instance2 was deleted without unmounting Volume2.  Should succeed.
	When("Delete Volume2", func() {
		It("Delete Volume2", func() {
			// Delete the volume
			logger.Log.Info("Starting the deletion of the Volume2 using filesystem API")

			storage_ids := []string{storage_id_created2}
			storage_utils.DeleteVolumes(storage_endpoint, token, storage_ids)
		})
	})

	// Delete Volume1 even though it is still mounted to Instance1 which is still running.
	When("Delete Volume1", func() {
		It("Delete Volume1", func() {
			// Delete the volume
			logger.Log.Info("Starting the deletion of the Volume1 using filesystem API")

			storage_ids := []string{storage_id_created1}
			storage_utils.DeleteVolumes(storage_endpoint, token, storage_ids)
		})
	})

	// Verify that mount path 1 on Instance1 cannot be accessed since the Volume1 was deleted
	When("Check Volume1 mount is inaccessible on Instance1", func() {
		It("Check Volume1 mount is inaccessible on Instance1", func() {
			if !prevStepOk {
				Skip("Skipping test because previous step failed")
			}

			logger.Log.Info("Checking that mount on Instance1 is inaccessible after Volume1 was deleted")
			err := storage_utils.VerifyMountPathIsInvalid(instance_creation_body1, mountPath1)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	// Delete Instance1
	When("Delete Instance1", func() {
		It("Delete Instance1", func() {
			if superComputeRun {
				Skip("Skipping test because this is a Supercompute test run...")
			}

			err := storage_utils.DeleteVM(instance_creation_body1, instance_endpoint, token, max_vm_deletion_timeout_in_min)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	// Delete Instance Group
	When("Delete Instance Group", func() {
		It("Delete Instance Group", func() {
			if !superComputeRun {
				Skip("Skipping test because this is not a Supercompute test run...")
			}

			err := storage_utils.DeleteInstanceGroup("gaudi-test", instance_group_endpoint, token)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
