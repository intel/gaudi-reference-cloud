package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/service_apis"

	. "github.com/onsi/gomega"
	"github.com/sethvargo/go-retry"
	"github.com/tidwall/gjson"
)

const sharedFileName = "sharedfile1.txt"
const scale_sleep_time_in_sec = 7
const Volume_timeout_in_min = 5

func FilesystemCreation(filesytem_base_url string, token string, storage_payload string, vol_name string, size string, vast bool) (int, string) {
	var storage_api_payload = storage_payload
	storage_api_payload = strings.Replace(storage_api_payload, "<<storage-name>>", vol_name, 1)
	storage_api_payload = strings.Replace(storage_api_payload, "<<storage-size>>", size, 1)

	storage_class := "GeneralPurposeStd"
	if !vast {
		storage_class = "GeneralPurpose"
	}
	storage_api_payload = strings.Replace(storage_api_payload, "<<storage-class>>", storage_class, 1)

	response_status, response_body := storage.CreateFilesystem(filesytem_base_url, token, storage_api_payload)
	return response_status, response_body
}

func UserCreation(user_base_url string, token string, volume_name string) (int, string) {
	var user_url = user_base_url + volume_name + "/user"
	response_status, response_body := storage.GetUserCreds(user_url, token)
	return response_status, response_body
}

func DeleteVolumes(storage_endpoint string, token string, resource_ids []string) {
	for _, res_id := range resource_ids {
		logger.Logf.Info("Deleting the volume with resource id: " + res_id)
		delete_response_byid_status, _ := storage.DeleteFilesystemById(storage_endpoint, token, res_id)
		Expect(delete_response_byid_status).To(Equal(200), "assertion failed on response code")
		// Validation
		logger.Log.Info("Validation of Volume Deletion using Id")
		DeleteValidationById := CheckVolumeDeletionById(storage_endpoint, token, res_id)
		Eventually(DeleteValidationById, 8*time.Minute, 30*time.Second).Should(BeTrue())
	}
	logger.Log.Info("All volumes have been deleted.")
}

// Delete multiple volumes with retry
func DeleteMultipleVolumesWithRetry(storage_endpoint string, token string, resource_ids []string) {
	for _, res_id := range resource_ids {
		DeleteVolumeWithRetry(storage_endpoint, token, res_id)
	}
}

// Delete a single volume specified by its resource_id and retry delete if initial attempt failed.
// NOTE: If the first attempt to delete the volume fails then we fail the test.
func DeleteVolumeWithRetry(storage_endpoint string, token string, resource_id string) {
	logger.Logf.Info("Deleting volume with resource id: " + resource_id)
	delete_response_byid_status_initial, _ := storage.DeleteFilesystemById(storage_endpoint, token, resource_id)

	// Only validate if response was 200
	if delete_response_byid_status_initial == 200 {
		// Verify that volume was deleted
		logger.Log.Info("Validation of Volume Deletion using Id")
		DeleteValidationById := CheckVolumeDeletionById(storage_endpoint, token, resource_id)

		Eventually(DeleteValidationById, Volume_timeout_in_min*time.Minute, 15*time.Second).Should(BeTrue())
	} else {
		// Volume deletion failed on first try so call the delete API a few more
		// times in an attempt to clean up the account.
		logger.Log.Info("Initial attempt to delete volume failed. Retrying volume deletion to clean up account")

		ctx := context.Background()
		backoffTimer := retry.NewFibonacci(1 * time.Second)
		backoffTimer = retry.WithMaxDuration(60*time.Second, backoffTimer)

		if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
			delete_response_byid_status, _ := storage.DeleteFilesystemById(storage_endpoint, token, resource_id)

			if delete_response_byid_status != 200 {
				return retry.RetryableError(fmt.Errorf("delete volume call failed, retry again"))
			}
			return nil
		}); err != nil {
			logger.Log.Info("Failed to delete volume after maximum retries")
		}
	}

	// Test is considered failed if first delete attempt failed.
	Expect(delete_response_byid_status_initial).To(Equal(200), "failed to delete volume")
}

func BucketCreation(bucket_base_url string, token string, bucket_payload string, bucket_name string) (int, string) {
	var bucket_api_payload = bucket_payload
	bucket_api_payload = strings.Replace(bucket_api_payload, "<<bucket-name>>", bucket_name, 1)
	response_status, response_body := storage.CreateBucket(bucket_base_url, token, bucket_api_payload)
	return response_status, response_body
}

func DeleteBuckets(storage_endpoint string, token string, resource_ids []string) {
	for _, res_id := range resource_ids {
		logger.Logf.Info("Deleting the bucket with resource id: " + res_id)
		delete_response_byid_status, _ := storage.DeleteBucketById(storage_endpoint, token, res_id)
		Expect(delete_response_byid_status).To(Equal(200), "assertion failed on response code")
		// Validation
		logger.Log.Info("Validation of bucket deletion using Id")
		DeleteValidationById := CheckBucketDeletionById(storage_endpoint, token, res_id)
		Eventually(DeleteValidationById, 1*time.Minute, 5*time.Second).Should(BeTrue())
	}
	logger.Log.Info("All buckets have been deleted.")
}

func DeleteMultipleBucketsWithRetry(storage_endpoint string, token string, resource_ids []string) {
	for _, res_id := range resource_ids {
		DeleteBucketWithRetry(storage_endpoint, token, res_id)
	}
	logger.Log.Info("All buckets have been deleted.")
}

// Delete a single bucket specified by its resource_id and retry delete if initial attempt failed.
// NOTE: If the first attempt to delete the volume fails then we fail the test.
func DeleteBucketWithRetry(storage_endpoint string, token string, resource_id string) {
	logger.Logf.Info("Deleting bucket with resource id: " + resource_id)
	delete_response_byid_status_initial, _ := storage.DeleteBucketById(storage_endpoint, token, resource_id)

	// Only validate if response was 200
	if delete_response_byid_status_initial == 200 {
		// Verify that bucket was deleted
		logger.Log.Info("Validation of Bucket Deletion using Id")
		DeleteValidationById := CheckBucketDeletionById(storage_endpoint, token, resource_id)
		Eventually(DeleteValidationById, 1*time.Minute, 5*time.Second).Should(BeTrue())
	} else {
		// Bucket deletion failed on first try so call the delete API a few more
		// times in an attempt to clean up the account.
		logger.Log.Info("Initial attempt to delete volume failed. Retrying volume deletion to clean up account")

		ctx := context.Background()
		backoffTimer := retry.NewFibonacci(1 * time.Second)
		backoffTimer = retry.WithMaxDuration(60*time.Second, backoffTimer)

		if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
			delete_response_byid_status, _ := storage.DeleteBucketById(storage_endpoint, token, resource_id)

			if delete_response_byid_status != 200 {
				return retry.RetryableError(fmt.Errorf("delete bucket call failed, retry again"))
			}
			return nil
		}); err != nil {
			logger.Log.Info("Failed to delete bucket after maximum retries")
		}
	}

	// Test is considered failed if first delete attempt failed.
	Expect(delete_response_byid_status_initial).To(Equal(200), "failed to delete bucket")
}

func PrincipalCreation(user_base_url string, token string, user_payload string, user_name string, bucket_name string) (int, string) {
	var user_api_payload = user_payload
	user_api_payload = strings.Replace(user_api_payload, "<<user-name>>", user_name, 1)
	user_api_payload = strings.Replace(user_api_payload, "<<bucket-name>>", bucket_name, 1)
	response_status, response_body := storage.CreatePrincipal(user_base_url, token, user_api_payload)
	return response_status, response_body
}

func DeletePrincipals(user_endpoint string, token string, user_ids []string) {
	for _, user_id := range user_ids {
		logger.Logf.Info("Deleting the user with user id: " + user_id)
		delete_response_byid_status, _ := storage.DeletePrincipalById(user_endpoint, token, user_id)
		Expect(delete_response_byid_status).To(Equal(200), "assertion failed on response code")
		// Validation
		logger.Log.Info("Validation of user deletion using Id")
		DeleteValidationById := CheckUserDeletionById(user_endpoint, token, user_id)
		Eventually(DeleteValidationById, 1*time.Minute, 5*time.Second).Should(BeTrue())
	}
	logger.Log.Info("All users have been deleted.")
}

func DeleteMultiplePrincipalsWithRetry(user_endpoint string, token string, user_ids []string) {
	for _, user_id := range user_ids {
		DeletePrincipalWithRetry(user_endpoint, token, user_id)
	}
	logger.Log.Info("All users have been deleted.")
}

func DeletePrincipalWithRetry(user_endpoint string, token string, user_id string) {
	logger.Logf.Info("Deleting the principal with user id: " + user_id)
	delete_response_byid_status_initial, _ := storage.DeletePrincipalById(user_endpoint, token, user_id)

	if delete_response_byid_status_initial == 200 {
		// Validation
		logger.Log.Info("Validation of principal deletion using Id")
		DeleteValidationById := CheckUserDeletionById(user_endpoint, token, user_id)
		Eventually(DeleteValidationById, 1*time.Minute, 5*time.Second).Should(BeTrue())
	} else {
		// Principal deletion failed on first try so call the delete API a few more
		// times in an attempt to clean up the account.
		logger.Log.Info("Initial attempt to delete principal failed. Retrying principal deletion to clean up account")

		ctx := context.Background()
		backoffTimer := retry.NewFibonacci(1 * time.Second)
		backoffTimer = retry.WithMaxDuration(60*time.Second, backoffTimer)

		if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
			delete_response_byid_status, _ := storage.DeletePrincipalById(user_endpoint, token, user_id)

			if delete_response_byid_status != 200 {
				return retry.RetryableError(fmt.Errorf("delete principal call failed, retry again"))
			}
			return nil
		}); err != nil {
			logger.Log.Info("Failed to delete principal after maximum retries")
		}
	}

	Expect(delete_response_byid_status_initial).To(Equal(200), "failed to delete principal")
}

func UpdatePrincipalPolicy(user_base_url string, token string, user_payload string, user_name string, userId string, bucket_name string, user_actions []string,
	user_permissions []string, testType string, positive bool) (int, string) {
	var response_status int
	var response_body string
	var user_api_payload = user_payload
	actionsString := fmt.Sprintf(`%s`, strings.Join(user_actions, `","`))
	permissionsString := fmt.Sprintf(`%s`, strings.Join(user_permissions, `","`))
	user_api_payload = strings.Replace(user_api_payload, "<<updated-user-name>>", user_name, 1)
	user_api_payload = strings.Replace(user_api_payload, "<<updated-bucket-name>>", bucket_name, 1)
	if positive {
		user_api_payload = strings.Replace(user_api_payload, "<<user-actions>>", actionsString, 1)
		user_api_payload = strings.Replace(user_api_payload, "<<user-permissions>>", permissionsString, 1)
	}
	// conditionally choose put function based on testType
	switch testType {
	case "Id":
		response_status, response_body = storage.PutPolicyPrincipalById(user_base_url, token, userId, user_api_payload) // Call your utility function for "id"
	case "Name":
		response_status, response_body = storage.PutPolicyPrincipalByName(user_base_url, token, user_name, user_api_payload) // Call your utility function for "name"
	default:
		// Handle if testType is not defined
		return 0, "Unknown test type. "
	}
	return response_status, response_body
}

func UpdatePrincipalMultiPolicy(user_base_url string, token string, user_payload string, user_name string, userId string, bucket_name1 string, bucket_name2 string, user_actions []string,
	user_permissions []string, testType string) (int, string) {
	var response_status int
	var response_body string
	var user_api_payload = user_payload
	actionsString := fmt.Sprintf(`%s`, strings.Join(user_actions, `","`))
	permissionsString := fmt.Sprintf(`%s`, strings.Join(user_permissions, `","`))
	user_api_payload = strings.Replace(user_api_payload, "<<updated-user-name>>", user_name, 1)
	user_api_payload = strings.Replace(user_api_payload, "<<updated-bucket-name-1>>", bucket_name1, 1)
	user_api_payload = strings.Replace(user_api_payload, "<<updated-bucket-name-2>>", bucket_name2, 1)
	user_api_payload = strings.Replace(user_api_payload, "<<user-actions>>", actionsString, 1)
	user_api_payload = strings.Replace(user_api_payload, "<<user-actions2>>", actionsString, 1)
	user_api_payload = strings.Replace(user_api_payload, "<<user-permissions>>", permissionsString, 1)
	user_api_payload = strings.Replace(user_api_payload, "<<user-permissions2>>", permissionsString, 1)

	// conditionally choose put function based on testType
	switch testType {
	case "Id":
		response_status, response_body = storage.PutPolicyPrincipalById(user_base_url, token, userId, user_api_payload) // Call your utility function for "id"
	case "Name":
		response_status, response_body = storage.PutPolicyPrincipalByName(user_base_url, token, user_name, user_api_payload) // Call your utility function for "name"
	default:
		// Handle if testType is not defined
		return 0, "Unknown test type. "
	}
	return response_status, response_body
}

func RuleCreation(rule_base_url string, token string, rule_payload string, rule_name string, bucket_id string) (int, string) {
	var rule_api_payload = rule_payload
	request_url := rule_base_url + bucket_id + "/lifecyclerule"
	rule_api_payload = strings.Replace(rule_api_payload, "<<rule-name>>", rule_name, 1)
	response_status, response_body := storage.CreateRule(request_url, token, rule_api_payload)
	return response_status, response_body
}

func DeleteRules(rule_endpoint string, token string, bucket_id string, resource_ids []string) {
	for _, res_id := range resource_ids {
		logger.Logf.Info("Deleting the lifecycle rule with bucket id:" + bucket_id + " rule id: " + res_id)
		delete_response_byid_status, _ := storage.DeleteRuleById(rule_endpoint+bucket_id+"/lifecyclerule/id/", token, res_id)
		Expect(delete_response_byid_status).To(Equal(200), "assertion failed on response code")

		// Validation
		logger.Log.Info("Validation of lifecycle rule deletion using Id")
		DeleteValidationById := CheckRuleDeletionById(rule_endpoint, token, res_id)
		Eventually(DeleteValidationById, 1*time.Minute, 5*time.Second).Should(BeTrue())
	}
	logger.Log.Info("All rules have been deleted.")
}

func DeleteMultipleRulesWithRetry(rule_endpoint string, token string, bucket_id string, resource_ids []string) {
	for _, res_id := range resource_ids {
		DeleteRuleWithRetry(rule_endpoint, token, bucket_id, res_id)
	}
	logger.Log.Info("All rules have been deleted.")
}

func DeleteRuleWithRetry(rule_endpoint string, token string, bucket_id string, resource_id string) {
	logger.Logf.Info("Deleting the lifecycle rule with bucket id:" + bucket_id + " rule id: " + resource_id)
	delete_response_byid_status_initial, _ := storage.DeleteRuleById(rule_endpoint+bucket_id+"/lifecyclerule/id/", token, resource_id)

	// Only validate if response was 200
	if delete_response_byid_status_initial == 200 {
		// Verify that the rule was deleted
		logger.Log.Info("Validation of lifecycle rule deletion using Id")
		DeleteValidationById := CheckRuleDeletionById(rule_endpoint, token, resource_id)
		Eventually(DeleteValidationById, 1*time.Minute, 5*time.Second).Should(BeTrue())
	} else {
		// Rule deletion failed on first try so call the delete API a few more
		// times in an attempt to clean up the account.
		logger.Log.Info("Initial attempt to delete rule failed. Retrying rule deletion to clean up account")

		ctx := context.Background()
		backoffTimer := retry.NewFibonacci(1 * time.Second)
		backoffTimer = retry.WithMaxDuration(60*time.Second, backoffTimer)

		if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
			delete_response_byid_status, _ := storage.DeleteRuleById(rule_endpoint+bucket_id+"/lifecyclerule/id/", token, resource_id)

			if delete_response_byid_status != 200 {
				return retry.RetryableError(fmt.Errorf("delete rule call failed, retry again"))
			}
			return nil
		}); err != nil {
			logger.Log.Info("Failed to delete rule after maximum retries")
		}
	}

	// Test is considered failed if first delete attempt failed.
	Expect(delete_response_byid_status_initial).To(Equal(200), "failed to delete rule")
}

func CreateVM(instance_type_to_be_created string, machineImage string, instance_endpoint string, token string,
	vm_payload string, sshkeyName string, vnet string, availabilityZone string, maxWaitTime int) (string, error) {

	vm_name := "automation-vm-" + GetRandomString()
	logger.Logf.Infof("Create VM with name: %s, Type: %s, Image: %s", vm_name, instance_type_to_be_created, machineImage)

	instance_creation_status := 0
	instance_creation_status, instance_creation_body := storage.CreateInstance(instance_endpoint, token,
		vm_payload, vm_name, instance_type_to_be_created, sshkeyName, vnet, machineImage, availabilityZone)

	if instance_creation_status != 200 {
		return "", fmt.Errorf("assertion failed on response code, expecting 200, actual: %d", instance_creation_status)
	}

	if !strings.Contains(instance_creation_body, `"name":"`+vm_name+`"`) {
		return "", errors.New("assertion failed on response body")
	}

	instance_id_created := gjson.Get(instance_creation_body, "metadata.resourceId").String()

	logger.Log.Info("Checking whether instance is in ready state")
	instanceReady := storage.CheckInstanceState(instance_endpoint, token, instance_id_created, "Ready")

	Eventually(instanceReady, time.Duration(maxWaitTime)*time.Minute, 30*time.Second).Should(BeTrue())
	get_response_byid_status, get_response := storage.GetInstanceById(instance_endpoint, token, instance_id_created)
	instance_creation_body = get_response

	var err error = nil
	if get_response_byid_status != 200 {
		err = fmt.Errorf("CheckInstanceState status code: %d, Expected 200", get_response_byid_status)
	}
	return instance_creation_body, err
}

func MountWEKAVolume(instance_creation_body, accountNum, volumePw, mountPath, volName, wekaUrl, network string) error {
	logger.Logf.Infof("Mount volume using network: %s", network)

	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	// Example mountwekavolume.sh script parameters:
	// mountwekavolume.sh <account num> <volume pw> <mount path> <volume name> <WEKA URL>

	cmdstr := fmt.Sprintf("sudo /home/ubuntu/mountwekavolume.sh %s '%s' '%s' '%s' '%s' '%s'", accountNum, volumePw, mountPath, volName, wekaUrl, network)
	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, cmdstr}

	out, commandErr :=
		RunCommand(sshCommand)
	logger.Logf.Infof("Output RunCommand mountwekavolume.sh script: [%s]", out)
	logger.Logf.Infof("commandErr returned by mountwekavolume.sh script: [%v]", commandErr)
	return commandErr
}

func MountVASTVolume(instance_creation_body, mountPath, volName, vastUrl string) error {
	logger.Logf.Infof("Mount VAST volume")

	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	// Example mountvastvolume.sh script parameters:
	// mountvastvolume.sh <mount path> <volume name> <VAST URL>
	cmdstr := fmt.Sprintf("sudo /home/ubuntu/mountvastvolume.sh '%s' '%s' '%s'", mountPath, volName, vastUrl)
	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, cmdstr}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output RunCommand mountvastvolume.sh script: [%s]", out)
	logger.Logf.Infof("commandErr returned by mountvastvolume.sh script: [%v]", commandErr)
	return commandErr
}

// Wait for cloud-init to complete
func WaitForCloudInit(instance_creation_body string, maxWaitTimeInSec int) error {
	logger.Logf.Infof("Waiting for cloud-init to complete...")

	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	// Example: waitforcloudinit.sh <max time to wait in sec>
	cmdstr := fmt.Sprintf("sudo /home/ubuntu/waitforcloudinit.sh %d", maxWaitTimeInSec)
	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, cmdstr}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output RunCommand waitforcloudinit.sh script: [%s]", out)
	logger.Logf.Infof("commandErr returned by waitforcloudinit.sh script: [%v]", commandErr)
	return commandErr
}

func WaitForMount(instance_creation_body, mountPath string) error {
	logger.Log.Info("Wait for mount to become available")
	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	cmdstr := fmt.Sprintf("sudo /home/ubuntu/waitformount.sh '%s'", mountPath)
	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, cmdstr}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output of waitformount.sh script: [%s]", out)
	logger.Logf.Infof("commandErr returned by waitformount.sh script: [%v]", commandErr)

	return commandErr
}

func RunTestsOnVolume(instance_creation_body, mountPath string) error {
	logger.Log.Info("Run FIO benchmark tests")
	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	cmdstr := fmt.Sprintf("sudo /home/ubuntu/runtests.sh '%s'", mountPath)
	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, cmdstr}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output RunCommand runtests.sh script: [%s]", out)
	return commandErr
}

func CreateTestFileOnVolume(instance_creation_body, mountPath string) error {
	logger.Log.Info("Run FIO benchmark tests")
	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	cmdstr := fmt.Sprintf("sudo touch %s/%s", mountPath, sharedFileName)
	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, cmdstr}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output of RunCommand: [%s]", out)
	return commandErr
}

func CheckTestFileOnVolume(instance_creation_body, mountPath string) error {
	logger.Log.Info("Run FIO benchmark tests")
	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	cmdstr := fmt.Sprintf("sudo ls %s/%s", mountPath, sharedFileName)
	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, cmdstr}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output of RunCommand: [%s]", out)
	return commandErr
}

func VerifyMountPathIsInvalid(instance_creation_body, mountPath string) error {
	logger.Log.Info("Verify mount path is invalid")
	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	cmdstr := fmt.Sprintf("sudo /home/ubuntu/waitforunmount.sh %s", mountPath)
	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, cmdstr}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output RunCommand runtests.sh script: [%s]", out)
	logger.Logf.Infof("commandErr returned by waitforunmount.sh script: [%v]", commandErr)
	return commandErr
}

func UnmountVolume(instance_creation_body, mountPath string) error {
	logger.Log.Info("Unmount volume")
	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	cmdstr := fmt.Sprintf("sudo /home/ubuntu/unmountvolume.sh '%s'", mountPath)
	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, cmdstr}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output RunCommand unmountvolume.sh script: [%s]", out)
	return commandErr
}

func DeleteVM(instance_creation_body string, instance_endpoint string, token string, maxWaitTime int) error {
	logger.Log.Info("Delete VM")

	// Delete VM
	logger.Log.Info("Remove the instance via DELETE api using resource id")
	instance_id_created := gjson.Get(instance_creation_body, "metadata.resourceId").String()
	delete_response_byid_status, _ := storage.DeleteInstanceById(instance_endpoint, token, instance_id_created)

	if delete_response_byid_status != 200 {
		return fmt.Errorf("assertion failed on response code, expecting 200, actual: %d", delete_response_byid_status)
	}

	logger.Log.Info("Validation of Instance Deletion")
	instanceValidation := storage.CheckInstanceDeletionById(instance_endpoint, token, instance_id_created)
	Eventually(instanceValidation, time.Duration(maxWaitTime)*time.Minute, 30*time.Second).Should(BeTrue())

	return nil
}

func EmptyBucket(instance_creation_body string) (string, error) {
	logger.Log.Info("emptying bucket")
	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, "python3 /home/ubuntu/delete_object.py"}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output RunCommand python3 /home/ubuntu/delete_object.py: [%s]", out)
	output := out.String()
	return output, commandErr
}

func UploadBucket(instance_creation_body string) (string, error) {
	logger.Log.Info("upload to bucket")
	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, "python3 /home/ubuntu/upload_object.py"}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output RunCommand python3 /home/ubuntu/upload_object.py: [%s]", out)
	output := out.String()
	return output, commandErr
}

func CheckCloudInit(instance_creation_body string) (string, error) {
	logger.Log.Info("checking for errors")
	machine_ip := ExtractMachineIPFromResponse(instance_creation_body)
	proxy_ip := ExtractProxyIPFromResponse(instance_creation_body)
	proxy_user := ExtractProxyUserFromResponse(instance_creation_body)
	machine_user := ExtractMachineUserFromResponse(instance_creation_body)

	sshCommand := []string{"ssh", "-J", proxy_user + "@" + proxy_ip, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		machine_user + "@" + machine_ip, "sudo cat /var/log/cloud-init-output.log"}

	out, commandErr := RunCommand(sshCommand)
	logger.Logf.Infof("Output RunCommand sudo cat /var/log/cloud-init-output.log: [%s]", out)
	output := out.String()
	return output, commandErr
}

// Delete a single file system (i.e. volume) specified by its name and retry delete if initial attempt failed.
// NOTE: If the first attempt to delete the volume fails then we fail the test.
func DeleteFilesystemByNameWithRetry(storage_endpoint string, token string, volume_name string) {
	logger.Logf.Infof("Deleting volume with name: %s", volume_name)
	delete_response_byname_status_initial, _ := storage.DeleteFilesystemByName(storage_endpoint, token, volume_name)

	// Only validate if response was 200
	if delete_response_byname_status_initial == 200 {
		// Verify that volume was deleted
		logger.Log.Info("Validation of Volume Deletion using name")
		DeleteValidationByName := CheckVolumeDeletionByName(storage_endpoint, token, volume_name)
		Eventually(DeleteValidationByName, Volume_timeout_in_min*time.Minute, 15*time.Second).Should(BeTrue())
	} else {
		// Volume deletion failed on first try so call the delete API a few more
		// times in an attempt to clean up the account.
		logger.Log.Info("Initial attempt to delete volume failed. Retrying volume deletion to clean up account")

		ctx := context.Background()
		backoffTimer := retry.NewFibonacci(1 * time.Second)
		backoffTimer = retry.WithMaxDuration(60*time.Second, backoffTimer)

		if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
			delete_response_byname_status, _ := storage.DeleteFilesystemByName(storage_endpoint, token, volume_name)

			if delete_response_byname_status != 200 {
				return retry.RetryableError(fmt.Errorf("delete volume call failed, retry again"))
			}
			return nil
		}); err != nil {
			logger.Log.Info("Failed to delete volume after maximum retries")
		}
	}

	// Test is considered failed if first delete attempt failed.
	Expect(delete_response_byname_status_initial).To(Equal(200), "failed to delete volume")
}

func CreateGaudiInstanceGroup(instance_type_to_be_created string, machineImage string, instance_group_endpoint string, token string,
	instance_group_payload string, sshkeyName string, vnet string, availabilityZone string) (string, string, error) {

	instance_group_name := "gaudi-test"

	logger.Logf.Infof("Creating an instance group with name: %s, instance type: %s, machine image: %s", instance_group_name, instance_type_to_be_created, machineImage)

	instance_creation_status := 0
	instance_creation_status, instance_creation_body := storage.InstanceGroupCreation(instance_group_endpoint,
		token, instance_group_payload, instance_group_name, "8", instance_type_to_be_created, sshkeyName, vnet, machineImage, availabilityZone)

	logger.Logf.Infof("instance_creation_body: [%s]", instance_creation_body)

	if instance_creation_status != 200 {
		return instance_group_name, "", fmt.Errorf("assertion failed on response code, expecting 200, actual: %d", instance_creation_status)
	}

	get_instances_group_response_status, get_instances_group_response_body := storage.GetAllInstanceGroups(instance_group_endpoint, token)

	if get_instances_group_response_status != 200 {
		return instance_group_name, "", fmt.Errorf("assertion failed on response code, expecting 200, actual: %d", get_instances_group_response_status)
	}

	totalMachines, _ := strconv.Atoi(gjson.Get(get_instances_group_response_body, "items.0.spec.instanceCount").String())
	machineReadyCount, _ := strconv.Atoi(gjson.Get(get_instances_group_response_body, "items.0.status.readyCount").String())
	logger.Logf.Infof("total machines: [%d], ready count: [%d]", totalMachines, machineReadyCount)

	// Loop until readyCount = totalMachines with a timeout of 30 minutes
	logger.Log.Info("Waiting for instance group machines to get to ready state")
	targetReached := false
	i := 0
	maxWaitTimeInMinutes := 60
	for i = 0; i < maxWaitTimeInMinutes; i++ {
		logger.Logf.Infof("Iteration: %d, Total machines: %d, Ready count: %d", i, totalMachines, machineReadyCount)
		if machineReadyCount == totalMachines {
			targetReached = true
			break
		}
		_, get_instances_group_response_body = storage.GetAllInstanceGroups(instance_group_endpoint, token)
		machineReadyCount, _ = strconv.Atoi(gjson.Get(get_instances_group_response_body, "items.0.status.readyCount").String())
		time.Sleep(1 * time.Minute)
	}

	if targetReached {
		logger.Log.Info("Success: All machines are in ready state")
		return instance_group_name, instance_creation_body, nil
	} else {
		logger.Log.Info("Error: All machines did not reach ready state")
		return instance_group_name, "", fmt.Errorf("all instance group machines did not reach ready state")
	}

}

func DeleteInstanceGroup(instance_group_name string, instance_group_endpoint string, token string) error {
	logger.Log.Info("Delete Instance Group")
	status, _ := storage.DeleteInstanceGroupByName(instance_group_endpoint, token, instance_group_name)

	if status != 200 {
		return fmt.Errorf("assertion failed on response code, expecting 200, actual: %d", status)
	}

	return nil
}

func CreateMultipleFilesystems(filesytem_base_url, token, storage_payload_template, size string, volumeCount int) {
	storage_payload_template = strings.Replace(storage_payload_template, "<<storage-size>>", size, 1)
	var temp_storage_payload = storage_payload_template

	for i := 0; i < volumeCount; i++ {
		volName := fmt.Sprintf("scale-volume-%04d", i)
		logger.Logf.Infof("Creating volume: %s", volName)
		temp_storage_payload = strings.Replace(storage_payload_template, "<<storage-name>>", volName, 1)
		response_status, response_body := storage.CreateFilesystem(filesytem_base_url, token, temp_storage_payload)
		Expect(response_status).To(Equal(200), response_body)

		storage_id_created := gjson.Get(response_body, "metadata.resourceId").String()

		// Validation
		logger.Log.Info("Checking whether volume is in ready state")
		ValidationSt := CheckVolumeProvisionState(filesytem_base_url, token, storage_id_created)
		Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())

		time.Sleep(scale_sleep_time_in_sec * time.Second)
	}
}

func CreateMultipleBuckets(bucket_endpoint, token, bucket_payload, cloud_account, bucket_name_prefix string, bucketCount int) {
	logger.Logf.Infof("Starting creation of %d buckets", bucketCount)

	for i := 0; i < bucketCount; i++ {
		bucketName := fmt.Sprintf("%s%04d", bucket_name_prefix, i)
		logger.Logf.Infof("Creating bucket: %s", bucketName)
		bucket_creation_status_positive, bucket_creation_body_positive := BucketCreation(bucket_endpoint, token, bucket_payload, bucketName)
		Expect(bucket_creation_status_positive).To(Equal(200), bucket_creation_body_positive)
		bucket_id_created := gjson.Get(bucket_creation_body_positive, "metadata.resourceId").String()

		newBucketName := fmt.Sprintf("%s-%s", cloud_account, bucketName)
		logger.Logf.Infof("Waiting for bucket [%s] to get to ready state", newBucketName)
		ValidationSt := CheckBucketProvisionState(bucket_endpoint, token, bucket_id_created)
		Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())

		time.Sleep(scale_sleep_time_in_sec * time.Second)
	}
}

func CreateMultiplePrincipals(principal_endpoint string, token string, principal_payload_template string, bucket_name string, principalCount int) {
	logger.Logf.Infof("Starting creation of %d principals", principalCount)
	principal_payload_template = strings.Replace(principal_payload_template, "<<bucket-name>>", bucket_name, 1)

	for i := 0; i < principalCount; i++ {
		principal_name := fmt.Sprintf("sc-pr-%04d", i)
		principal_api_payload := strings.Replace(principal_payload_template, "<<user-name>>", principal_name, 1)

		logger.Logf.Infof("Creating principal %s for bucket %s", principal_name, bucket_name)
		response_status, response_body := storage.CreatePrincipal(principal_endpoint, token, principal_api_payload)
		Expect(response_status).To(Equal(200), response_body)

		principal_id_created := gjson.Get(response_body, "metadata.userId").String()

		// Validation
		logger.Log.Info("Checking whether principal is in ready state")
		ValidationSt := CheckUserProvisionState(principal_endpoint, token, principal_id_created)
		Eventually(ValidationSt, 1*time.Minute, 5*time.Second).Should(BeTrue())

		time.Sleep(scale_sleep_time_in_sec * time.Second)
	}
}

func CreateMultipleRules(rule_base_url string, token string, rule_payload_template string, bucket_id string, ruleCount int) ([]string, error) {
	logger.Logf.Infof("Starting creation of %d lifecycle rules", ruleCount)
	ruleIds := make([]string, 0)

	for i := 0; i < ruleCount; i++ {
		rule_name := fmt.Sprintf("scale-rule-%04d", i)
		logger.Logf.Infof("Creating lifecycle rule: %s", rule_name)

		rule_payload := strings.Replace(rule_payload_template, "<<rule-name>>", rule_name, 1)
		rule_creation_status, rule_creation_body := RuleCreation(rule_base_url, token, rule_payload, rule_name, bucket_id)
		if rule_creation_status != 200 {
			return ruleIds, fmt.Errorf("lifecycle rule creation failed: %s", rule_creation_body)
		}

		rule_id_created := gjson.Get(rule_creation_body, "metadata.resourceId").String()
		ruleIds = append(ruleIds, rule_id_created)

		time.Sleep(scale_sleep_time_in_sec * time.Second)
	}

	return ruleIds, nil
}

func DeleteMultipleBuckets(bucket_endpoint, token, cloud_account, bucket_name_prefix string, bucketCount int) {
	logger.Logf.Infof("Starting deletion of %d buckets", bucketCount)
	for i := 0; i < bucketCount; i++ {
		bucketName := fmt.Sprintf("%s-%s%04d", cloud_account, bucket_name_prefix, i)
		logger.Logf.Infof("Deleting bucket: %s", bucketName)
		delete_response_byname_status, delete_response_byname_body := storage.DeleteBucketByName(bucket_endpoint, token, bucketName)
		if delete_response_byname_status != 200 {
			logger.Logf.Infof("ERROR: Failed to delete bucket: %s", bucketName)
			logger.Logf.Infof("Delete request body: %s", delete_response_byname_body)
		}

		time.Sleep(scale_sleep_time_in_sec * time.Second)
	}
}

func DeleteMultipleRules(rule_base_url string, token string, bucket_id string, rule_ids []string) {
	logger.Logf.Infof("Starting deletion of %d lifecycle rules", len(rule_ids))

	var rule_name string
	for i := 0; i < len(rule_ids); i++ {
		rule_id := rule_ids[i]

		// Get the rule name
		status, body := storage.GetRuleById(rule_base_url, bucket_id, token, rule_id)
		rule_name = ""
		if status == 200 {
			rule_name := gjson.Get(body, "metadata.resourceId").String()
			logger.Logf.Infof("Deleting lifecycle rule: %s", rule_name)
		}

		status, body = storage.DeleteRuleById(rule_base_url+bucket_id+"/lifecyclerule/id/", token, rule_id)
		if status != 200 {
			logger.Logf.Infof("ERROR: Failed to delete lifecycle rule: %s, ID: %s", rule_name, rule_id)
			logger.Logf.Infof("Rule delete request body: %s", body)
		}

		time.Sleep(scale_sleep_time_in_sec * time.Second)
	}
}

func DeleteMultiplePrincipals(principal_endpoint string, token string, principalCount int) {
	logger.Logf.Infof("Starting deletion of %d principals", principalCount)
	for i := 0; i < principalCount; i++ {
		principal_name := fmt.Sprintf("sc-pr-%04d", i)

		logger.Logf.Infof("Deleting principal: %s", principal_name)

		delete_response_byname_status, delete_response_byname_body := storage.DeletePrincipalByName(principal_endpoint, token, principal_name)
		if delete_response_byname_status != 200 {
			logger.Logf.Infof("ERROR: Failed to delete principal: %s", principal_name)
			logger.Logf.Infof("Principal delete request body: %s", delete_response_byname_body)
		}

		time.Sleep(scale_sleep_time_in_sec * time.Second)
	}
}

func DeleteMultipleFilesystems(filesytem_base_url, token string, volCount int) {
	for i := 0; i < volCount; i++ {
		volName := fmt.Sprintf("scale-volume-%04d", i)
		logger.Logf.Infof("Deleting volume: %s", volName)

		status, body := storage.GetFilesystemByName(filesytem_base_url, token, volName)

		if status == 200 {
			res_id := gjson.Get(body, "metadata.resourceId").String()
			storage.DeleteFilesystemById(filesytem_base_url, token, res_id)
		} else {
			logger.Logf.Infof("Failed to delete volume: %s", volName)
			logger.Logf.Infof("Volume delete request body: %s", body)
		}

		time.Sleep(scale_sleep_time_in_sec * time.Second)
	}
}

func ConcurrencyCreateVolume(storage_endpoint, token, storage_payload, volume_name_prefix, size string, vast bool, id int, wg *sync.WaitGroup) {
	defer wg.Done() // Notify the WaitGroup that this worker is done

	volume_name := fmt.Sprintf("%s%d", volume_name_prefix, id)
	volume_creation_status, volume_creation_body := FilesystemCreation(storage_endpoint, token, storage_payload, volume_name, size, vast)

	if volume_creation_status == 200 {
		resource_id := gjson.Get(volume_creation_body, "metadata.resourceId").String()

		// Wait for volume to get to a ready state
		logger.Log.Info("Checking whether volume is in ready state")

		for i := 0; i <= 15; i++ {
			get_response_byid_status, get_response_byid_body := storage.GetFilesystemById(storage_endpoint, token, resource_id)
			if get_response_byid_status == 200 {
				storagePhase := gjson.Get(get_response_byid_body, "status.phase").String()
				if storagePhase == "FSReady" {
					break
				}
			}
			time.Sleep(5 * time.Second)
		}
	} else {
		logger.Logf.Warnf("ERROR: Volume creation API failed. Volume name: %s, Status: %d, Body: %s", volume_name, volume_creation_status, volume_creation_body)
	}
}

func ConcurrencyDeleteVolume(storage_endpoint, token, volume_name_prefix string, id int, wg *sync.WaitGroup) {
	defer wg.Done() // Notify the WaitGroup that this worker is done

	volume_name := fmt.Sprintf("%s%d", volume_name_prefix, id)
	del_status, del_body := storage.DeleteFilesystemByName(storage_endpoint, token, volume_name)

	if del_status != 200 {
		logger.Logf.Warnf("ERROR: Failed to delete volume: %s, Status: %d, Body: %s", volume_name, del_status, del_body)

		// 404 error means the volume was not created, so no need to retry
		if del_status != 404 {
			// Initial attempt to delete volume failed, so try again
			for i := 0; i <= 5; i++ {
				del_status, _ = storage.DeleteFilesystemByName(storage_endpoint, token, volume_name)
				if del_status == 200 {
					break
				}

				time.Sleep(5 * time.Second)
			}

		}
	}
}

// This goroutine will create a bucket and wait for it to get to a ready state
func ConcurrencyCreateBucket(bucket_endpoint, token, storage_payload, bucket_name_prefix, cloud_account string, id int, wg *sync.WaitGroup) {
	defer wg.Done() // Notify the WaitGroup that this worker is done
	bucket_name := fmt.Sprintf("%s%d", bucket_name_prefix, id)
	bucket_creation_status, bucket_creation_body := BucketCreation(bucket_endpoint, token, bucket_payload, bucket_name)

	if bucket_creation_status == 200 {
		bucket_id_created := gjson.Get(bucket_creation_body, "metadata.resourceId").String()

		// Wait for bucket to get to a ready state
		logger.Log.Info("Checking whether bucket is in ready state")

		for i := 0; i <= 15; i++ {
			get_response_byid_status, get_response_byid_body := storage.GetBucketById(bucket_endpoint, token, bucket_id_created)
			if get_response_byid_status == 200 {
				bucketPhase := gjson.Get(get_response_byid_body, "status.phase").String()
				if bucketPhase == "BucketReady" {
					break
				}
				time.Sleep(5 * time.Second)
			}
		}
	} else {
		logger.Logf.Warnf("ERROR: Bucket creation API failed. Bucket name: %s, Status: %d, Body: %s", bucket_name, bucket_creation_status, bucket_creation_body)
	}
}

func ConcurrencyDeleteBucket(bucket_endpoint, token, bucket_name_prefix, cloud_account string, id int, wg *sync.WaitGroup) {
	defer wg.Done() // Notify the WaitGroup that this worker is done

	bucket_name := fmt.Sprintf("%s-%s%d", cloud_account, bucket_name_prefix, id)
	del_status, del_body := storage.DeleteBucketByName(bucket_endpoint, token, bucket_name)

	if del_status != 200 {
		logger.Logf.Warnf("ERROR: Failed to delete bucket %s, Status: %d, Body: %s", del_status, del_body)

		// Initial attempt to delete bucket failed, so try again
		for i := 0; i <= 5; i++ {
			del_status, _ = storage.DeleteBucketByName(bucket_endpoint, token, bucket_name)
			if del_status == 200 {
				break
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func ConcurrencyCreatePrincipal(principal_endpoint, token, principal_payload, principal_name, bucket_name string, id int, wg *sync.WaitGroup) {
	defer wg.Done() // Notify the WaitGroup that this worker is done

	principal_creation_status, principal_creation_body := PrincipalCreation(principal_endpoint, token, principal_payload, principal_name, bucket_name)

	if principal_creation_status == 200 {
		principal_id_created := gjson.Get(principal_creation_body, "metadata.userId").String()

		// Wait for principal to get to a ready state
		logger.Log.Info("Checking whether principal is in ready state")

		for i := 0; i <= 15; i++ {
			get_response_byid_status, get_response_byid_body := storage.GetPrincipalById(principal_endpoint, token, principal_id_created)
			if get_response_byid_status == 200 {
				userPhase := gjson.Get(get_response_byid_body, "status.phase").String()
				if userPhase == "ObjectUserReady" {
					break
				}
			} else {
				logger.Logf.Infof("ERROR: Call to GetPrincipalById failed with status %d. Body: %s", get_response_byid_status, get_response_byid_body)
			}

			time.Sleep(5 * time.Second)
		}
	} else {
		logger.Logf.Warnf("ERROR: Failed to create principal: %s. Status: %d, Body: %s", principal_name, principal_creation_status, principal_creation_body)
	}
}

func ConcurrencyDeletePrincipal(principal_endpoint, token, principal_payload, principal_name string, wg *sync.WaitGroup) {
	defer wg.Done() // Notify the WaitGroup that this worker is done

	status, body := storage.DeletePrincipalByName(principal_endpoint, token, principal_name)
	if status != 200 {
		logger.Logf.Warnf("Cannot delete principal %s.  Status: %d, Body: %s", principal_name, status, body)

		// Initial attempt to delete principal failed, try again
		for i := 0; i <= 5; i++ {
			status, _ := storage.DeletePrincipalByName(principal_endpoint, token, principal_name)
			if status == 200 {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func ConcurrencyCreateRule(rule_endpoint, token, rule_payload, rule_name, bucket_id string, i int, rule_ids []string, wg *sync.WaitGroup) {
	defer wg.Done() // Notify the WaitGroup that this worker is done

	rule_payload = strings.Replace(rule_payload, "<<rule-name>>", rule_name, 1)
	rule_creation_status, rule_creation_body := RuleCreation(rule_endpoint, token, rule_payload, rule_name, bucket_id)

	if rule_creation_status == 200 {
		rule_id := gjson.Get(rule_creation_body, "metadata.resourceId").String()
		rule_ids[i-1] = rule_id
	} else {
		logger.Logf.Warnf("Failed to create rule: %s, Status: %d, Body: %s", rule_name, rule_creation_status, rule_creation_body)
	}
}

func ConcurrencyDeleteRule(rule_endpoint, token, bucket_id, rule_id string, wg *sync.WaitGroup) {
	defer wg.Done() // Notify the WaitGroup that this worker is done

	status, body := storage.DeleteRuleById(rule_endpoint+bucket_id+"/lifecyclerule/id/", token, rule_id)
	if status != 200 {
		logger.Logf.Warnf("Failed to delete rule with ID: %s, Status: %d, Body: %s", rule_id, status, body)

		// Initial rule deletion attempt failed, so try again
		for i := 0; i <= 5; i++ {
			status, _ := storage.DeleteRuleById(rule_endpoint+bucket_id+"/lifecyclerule/id/", token, rule_id)
			if status == 200 {
				break
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func GetRandomString() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 6+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : 6+2]
}

func ExtractMachineIPFromResponse(response string) string {
	return gjson.Get(response, "status.interfaces.0.addresses.0").String()
}

func ExtractProxyIPFromResponse(response string) string {
	return gjson.Get(response, "status.sshProxy.proxyAddress").String()
}

func ExtractProxyUserFromResponse(response string) string {
	return gjson.Get(response, "status.sshProxy.proxyUser").String()
}

func ExtractMachineUserFromResponse(response string) string {
	return gjson.Get(response, "status.userName").String()
}

func RunCommand(sshCommand []string) (*bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(sshCommand[0], sshCommand[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Command failed with error: %v\n", err)
		if stderr.Len() > 0 {
			fmt.Printf("Error: %s\n", stderr.String())
		}
		return nil, err
	}
	return &stdout, nil
}

func ConvertFileToString(filePath string, filename string) (string, error) {
	//fmt.Println("Config file path", filePath)
	wd, _ := os.Getwd()
	wd = filepath.Clean(filepath.Join(wd, filePath))
	configData, err := os.ReadFile(wd + "/" + filename)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(configData), nil
}

func GetAvailabilityZone(region string, enabled bool) string {
	if region[:7] == "us-dev-" {
		if enabled {
			return region + "b"
		} else {
			return region + "a"
		}
	} else {
		return region + "a"
	}
}
