package public_apis

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/service_apis"
	storage_utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/utils"

	"fmt"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Storage Concurrency tests", Label("storage", "concurrency"), Ordered, ContinueOnFailure, func() {
	const (
		volume_name_prefix      = "concurr-vol-"
		bucket_name_prefix      = "concurr-bk-"
		principal_name_prefix   = "concurr-pr-"
		rule_name_prefix        = "cc-rule-"
		volume_count            = 10
		bucket_count            = 10
		principal_count         = 5
		rule_count              = 10
		sleep_time_in_sec       = 10
		defaultVolumeQuotaLimit = 2
	)

	var storage_payload string
	var principal_payload string
	var rule_payload string
	var first_bucket_name string
	var rule_ids = make([]string, rule_count)
	var quotaTestPrevStepOk bool

	BeforeAll(func() {
		logger.Logf.Infof("Starting concurrency test suite")

		// retrieve the required information from test config
		storage_payload = storage_utils.GetStoragePayload()
		bucket_payload = storage_utils.GetBucketPayload()
		principal_payload = storage_utils.GetPrincipalPayload()
		rule_payload = storage_utils.GetRulePayload()
	})

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	// Note: The volume quota test assumes that the target account has the default volume quota of 2 volumes.
	// Steps:
	// Create two volumes sequentally
	// Wait for volumes to get to ready state
	When("Volume quota concurrency test - Create two volumes", func() {
		It("Volume quota concurrency test - Create two volumes", func() {
			if !concurrTestRun || cloudAccount == concurr_account && region == concurr_region {
				Skip("This volume quota concurrency test can only run in accounts with a low volume quota")
			}

			quotaTestPrevStepOk = false

			get_response_byid_status, get_response_byid_body := storage.GetAllFilesystems(storage_endpoint, token)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
			currVolCount := strings.Count(get_response_byid_body, `"spec":`)
			Expect(currVolCount).To(Equal(0), "Test requires no volumes to exist in the account")

			size := "1TB"

			// Create volumes sequentally up to the quota limit and wait for each volume to get to Ready state
			for i := 1; i <= defaultVolumeQuotaLimit; i++ {
				volume_name := fmt.Sprintf("%s%d", volume_name_prefix, i)
				volume_creation_status, volume_creation_body := storage_utils.FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, size, vastEnabled)
				Expect(volume_creation_status).To(Equal(200), volume_creation_body)

				status, body := storage.GetFilesystemByName(storage_endpoint, token, volume_name)
				Expect(status).To(Equal(200), body)
				storage_id_created := gjson.Get(body, "metadata.resourceId").String()

				logger.Logf.Infof("Checking whether volume %s is in ready state", volume_name)
				ValidationSt := storage_utils.CheckVolumeProvisionState(storage_endpoint, token, storage_id_created)
				Eventually(ValidationSt, storage_utils.Volume_timeout_in_min*time.Minute, 15*time.Second).Should(BeTrue())
			}

			quotaTestPrevStepOk = true
		})
	})

	// In parallel delete the two volumes created in the above step and create 10 volumes.
	When("Volume quota concurrency test - Create and delete volumes in parallel", func() {
		It("Volume quota concurrency test - Create and delete volumes in parallel", func() {
			if !concurrTestRun || cloudAccount == concurr_account && region == concurr_region {
				Skip("This volume quota concurrency test can only run in accounts with a low volume quota")
			}

			if !quotaTestPrevStepOk {
				Skip("Skipping test because previous step failed")
			}

			var wg1 sync.WaitGroup // Used to wait for all create volume goroutines to finish

			// Launch goroutines for creating volumes
			for i := (defaultVolumeQuotaLimit + 1); i <= (volume_count + defaultVolumeQuotaLimit); i++ {
				wg1.Add(1) // Increment the WaitGroup counter

				logger.Logf.Infof("Creating volume %d in parallel", i)
				go storage_utils.ConcurrencyCreateVolume(storage_endpoint, token, storage_payload, volume_name_prefix, "1TB", vastEnabled, i, &wg1) // Start the goroutine
			}

			var wg2 sync.WaitGroup // Used to wait for all delete volume goroutines to finish

			// Launch goroutines for deleting volumes
			for i := 1; i <= defaultVolumeQuotaLimit; i++ {
				wg2.Add(1) // Increment the WaitGroup counter

				logger.Logf.Infof("Deleting volume %d", i)
				go storage_utils.ConcurrencyCreateVolume(storage_endpoint, token, storage_payload, volume_name_prefix, "1TB", vastEnabled, i, &wg2) // Start the goroutine
			}

			wg1.Wait() // Wait for all create volume goroutines to finish
			logger.Log.Info("All create volume workers completed")

			wg2.Wait() // Wait for all delete volume goroutines to finish
			logger.Log.Info("All delete volume workers completed")
		})
	})

	// Verify volume count
	When("Volume quota concurrency test - Verify volume count", func() {
		It("Volume quota concurrency test - Verify volume count", func() {
			if !concurrTestRun || cloudAccount == concurr_account && region == concurr_region {
				Skip("This concurrency test can only run in accounts with a low volume quota")
			}

			if !quotaTestPrevStepOk {
				Skip("Skipping test because previous step failed")
			}

			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			logger.Log.Info("Verifying number of volumes created")
			get_response_byid_status, get_response_byid_body := storage.GetAllFilesystems(storage_endpoint, token)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)

			volCount := strings.Count(get_response_byid_body, `"spec":`)
			logger.Logf.Infof("number of volumes: %d", volCount)
			Expect(volCount).To(Equal(defaultVolumeQuotaLimit))
		})
	})

	// Clean up the volumes create by the quota concurrency test sequentally
	When("Volume quota concurrency test - Cleanup volumes", func() {
		It("Volume quota concurrency test - Cleanup volumes", func() {
			if !concurrTestRun || cloudAccount == concurr_account && region == concurr_region {
				Skip("This volume quota concurrency test can only run in accounts with a low volume quota")
			}

			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			logger.Logf.Infof("Deleting volumes sequentally...")

			// Delete all volumes created by the quota concurrency tests
			// NOTE: Some of the volumes won't exist because of the quota limit.  In this case we just ignore the 404 errors
			for i := 1; i <= (volume_count + defaultVolumeQuotaLimit); i++ {
				volume_name := fmt.Sprintf("%s%d", volume_name_prefix, i)
				logger.Logf.Infof("Deleting volume: %s", volume_name)
				status, body := storage.DeleteFilesystemByName(storage_endpoint, token, volume_name)
				if status != 200 && status != 404 {
					logger.Logf.Infof("ERROR: Attempt to delete volume %s resulted in error %d.  Body: %s", status, body)
				}
			}

		})
	})

	When("Create buckets concurrently", func() {
		It("Create buckets concurrently", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Creating buckets concurrently")
			var wg sync.WaitGroup // Used to wait for all goroutines to finish

			// Launch goroutines
			for i := 1; i <= bucket_count; i++ {
				wg.Add(1) // Increment the WaitGroup counter

				logger.Logf.Infof("Creating bucket %d", i)
				go storage_utils.ConcurrencyCreateBucket(bucket_endpoint, token, bucket_payload, bucket_name_prefix, cloudAccount, i, &wg) // Start the goroutine
			}

			wg.Wait() // Wait for all goroutines to finish
			logger.Log.Info("All workers completed")

			first_bucket_name = fmt.Sprintf("%s-%s1", cloudAccount, bucket_name_prefix)
		})
	})

	When("List all buckets", func() {
		It("List all buckets", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("List all buckets")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			get_response_byid_status, get_response_byid_body := storage.GetAllBuckets(bucket_endpoint, token)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Check buckets", func() {
		It("Check buckets ", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Checking all buckets")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			for i := 1; i <= bucket_count; i++ {
				bucket_name := fmt.Sprintf("%s-%s%d", cloudAccount, bucket_name_prefix, i)
				logger.Logf.Infof("Checking bucket %s", bucket_name)

				status, body := storage.GetBucketByName(bucket_endpoint, token, bucket_name)
				Expect(status).To(Equal(200), body)

				bucketPhase := gjson.Get(body, "status.phase").String()
				Expect(bucketPhase).To(Equal("BucketReady"), body)
			}
		})
	})

	// Create lifecycle rules concurrently
	// Prereq: Buckets must be create before this step is run
	When("Create rules concurrently", func() {
		It("Create rules concurrently", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Creating rules concurrently")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			var wg sync.WaitGroup // Used to wait for all goroutines to finish

			Expect(first_bucket_name).NotTo(BeEmpty())
			status, body := storage.GetBucketByName(bucket_endpoint, token, first_bucket_name)
			Expect(status).To(Equal(200), body)
			first_bucket_id := gjson.Get(body, "metadata.resourceId").String()

			// Launch goroutines
			for i := 1; i <= rule_count; i++ {
				wg.Add(1) // Increment the WaitGroup counter

				logger.Logf.Infof("Creating rule %d", i)
				rule_name := fmt.Sprintf("%s%d", rule_name_prefix, i)
				go storage_utils.ConcurrencyCreateRule(rule_endpoint, token, rule_payload, rule_name, first_bucket_id, i, rule_ids, &wg)
			}

			wg.Wait() // Wait for all goroutines to finish
			logger.Log.Info("All workers completed")

			logger.Logf.Infof("rule ids: %s", rule_ids)
		})
	})

	When("Check rules", func() {
		It("Check rules ", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Checking all rules")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			Expect(first_bucket_name).NotTo(BeEmpty())
			status, body := storage.GetBucketByName(bucket_endpoint, token, first_bucket_name)
			Expect(status).To(Equal(200), body)
			first_bucket_id := gjson.Get(body, "metadata.resourceId").String()

			for i := 1; i <= rule_count; i++ {
				rule_name := fmt.Sprintf("%s%d", rule_name_prefix, i)
				logger.Logf.Infof("Checking rule %s", rule_name)

				status, body = storage.GetRuleById(rule_endpoint, first_bucket_id, token, rule_ids[i-1])
				Expect(status).To(Equal(200), body)

				// Get rule name from body
				rule_name_actual := gjson.Get(body, "metadata.ruleName").String()
				Expect(rule_name_actual).To(Equal(rule_name))
			}
		})
	})

	When("Delete rules concurrently", func() {
		It("Delete rules concurrently", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Deleting rules concurrently")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			var wg sync.WaitGroup

			Expect(first_bucket_name).NotTo(BeEmpty())
			status, body := storage.GetBucketByName(bucket_endpoint, token, first_bucket_name)
			Expect(status).To(Equal(200), body)
			first_bucket_id := gjson.Get(body, "metadata.resourceId").String()

			// Launch goroutines
			for i := 1; i <= rule_count; i++ {
				wg.Add(1) // Increment the WaitGroup counter

				logger.Logf.Infof("Deleting rule with ID: %s", rule_ids[i-1])
				go storage_utils.ConcurrencyDeleteRule(rule_endpoint, token, first_bucket_id, rule_ids[i-1], &wg) // Start the goroutine
			}

			wg.Wait() // Wait for all goroutines to finish
			logger.Log.Info("All workers completed")
		})
	})

	// Create principals concurrently
	// Prereq: Buckets must be created before this step is run
	When("Create principals concurrently", func() {
		It("Create principals concurrently", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Creating principals concurrently")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			var wg sync.WaitGroup // Used to wait for all goroutines to finish

			// Launch goroutines
			for i := 1; i <= principal_count; i++ {
				wg.Add(1) // Increment the WaitGroup counter

				principal_name := fmt.Sprintf("%s%d", principal_name_prefix, i)
				logger.Logf.Infof("Creating principal %s", principal_name)
				go storage_utils.ConcurrencyCreatePrincipal(principal_endpoint, token, principal_payload, principal_name, first_bucket_name, i, &wg) // Start the goroutine
			}

			wg.Wait() // Wait for all goroutines to finish
			logger.Log.Info("All workers completed")
		})
	})

	When("Check state of each principal", func() {
		It("Check state of each principal", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Checking each principal")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			for i := 1; i <= principal_count; i++ {
				principal_name := fmt.Sprintf("%s%d", principal_name_prefix, i)
				logger.Logf.Infof("Checking principal %s", principal_name)

				status, body := storage.GetPrincipalByName(principal_endpoint, token, principal_name)
				Expect(status).To(Equal(200), body)

				principal_id := gjson.Get(body, "metadata.userId").String()

				logger.Log.Info("Checking whether principal is in ready state")
				ValidationSt := storage_utils.CheckUserProvisionState(principal_endpoint, token, principal_id)
				Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())
			}
		})
	})

	When("Delete principals concurrently", func() {
		It("Delete principals concurrently", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Deleting principals concurrently")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			var wg sync.WaitGroup // Used to wait for all goroutines to finish

			// Launch goroutines
			for i := 1; i <= principal_count; i++ {
				wg.Add(1) // Increment the WaitGroup counter

				principal_name := fmt.Sprintf("%s%d", principal_name_prefix, i)
				logger.Logf.Infof("Deleting principal %s", principal_name)

				go storage_utils.ConcurrencyDeletePrincipal(principal_endpoint, token, principal_payload, principal_name, &wg)
			}

			wg.Wait() // Wait for all goroutines to finish
			logger.Log.Info("All workers completed")
		})
	})

	When("Delete buckets concurrently", func() {
		It("Delete buckets concurrently", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Deleting buckets concurrently")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			var wg sync.WaitGroup // Used to wait for all goroutines to finish

			// Launch goroutines
			for i := 1; i <= bucket_count; i++ {
				wg.Add(1) // Increment the WaitGroup counter

				// Start the goroutine
				go storage_utils.ConcurrencyDeleteBucket(bucket_endpoint, token, bucket_name_prefix, cloudAccount, i, &wg)
			}

			wg.Wait() // Wait for all goroutines to finish
			logger.Log.Info("All workers completed")
		})
	})

	When("Create volumes concurrently", func() {
		It("Create volumes concurrently", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Creating volumes concurrently")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			var wg sync.WaitGroup // Used to wait for all goroutines to finish

			// Launch goroutines
			for i := 1; i <= volume_count; i++ {
				wg.Add(1) // Increment the WaitGroup counter

				logger.Logf.Infof("Creating volume %d", i)
				go storage_utils.ConcurrencyCreateVolume(storage_endpoint, token, storage_payload, volume_name_prefix, "1TB", vastEnabled, i, &wg) // Start the goroutine
			}

			wg.Wait() // Wait for all goroutines to finish
			logger.Log.Info("All workers completed")
		})
	})

	When("Check state of each volume", func() {
		It("Check state of each volume", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Checking each volume")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			for i := 1; i <= volume_count; i++ {
				volume_name := fmt.Sprintf("%s%d", volume_name_prefix, i)
				logger.Logf.Infof("Checking state of volume %s", volume_name)

				status, body := storage.GetFilesystemByName(storage_endpoint, token, volume_name)
				Expect(status).To(Equal(200), body)
				storage_id_created := gjson.Get(body, "metadata.resourceId").String()

				// Validation
				logger.Log.Info("Checking whether volume is in ready state")
				ValidationSt := storage_utils.CheckVolumeProvisionState(storage_endpoint, token, storage_id_created)
				Eventually(ValidationSt, storage_utils.Volume_timeout_in_min*time.Minute, 15*time.Second).Should(BeTrue())
			}
		})
	})

	When("List all file systems", func() {
		It("List all file systems", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("List all file systems")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			get_response_byid_status, get_response_byid_body := storage.GetAllFilesystems(storage_endpoint, token)
			Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)
		})
	})

	When("Delete volumes concurrently", func() {
		It("Delete volumes concurrently", func() {
			if shouldSkipSuite() {
				Skip(fmt.Sprintf("Concurrency suite can only run in account: %s and region: %s", concurr_account, concurr_region))
			}

			logger.Log.Info("Deleting volumes concurrently")
			logger.Logf.Infof("Sleeping for %d seconds to avoid rate limit errors", sleep_time_in_sec)
			time.Sleep(sleep_time_in_sec * time.Second)

			var wg2 sync.WaitGroup

			// Launch goroutines
			for i := 1; i <= volume_count; i++ {
				wg2.Add(1) // Increment the WaitGroup counter

				// Start the goroutine
				go storage_utils.ConcurrencyDeleteVolume(storage_endpoint, token, volume_name_prefix, i, &wg2)
			}

			wg2.Wait() // Wait for all goroutines to finish
			logger.Log.Info("All workers completed")
		})
	})

	AfterAll(func() {
	})
})

func shouldSkipSuite() bool {
	rv := false
	if !concurrTestRun || cloudAccount != concurr_account || region != concurr_region {
		rv = true
	}
	return rv
}
