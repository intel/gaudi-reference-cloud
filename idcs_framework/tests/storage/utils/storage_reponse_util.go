package utils

import (
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/service_apis"

	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

func CheckVolumeProvisionState(storage_endpoint string, token string, resource_id string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_response_byid_status, get_response_byid_body := storage.GetFilesystemById(storage_endpoint, token, resource_id)
		Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)

		storagePhase := gjson.Get(get_response_byid_body, "status.phase").String()
		Expect(storagePhase).ToNot(Equal("FSFailed"), "Volume is in failed state")

		logger.Logf.Info("storagePhase: ", storagePhase)
		if storagePhase != "FSReady" {
			logger.Logf.Info("Volume is not in ready state")
			return false
		} else {
			logger.Logf.Info("Volume is in ready state")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for volume to get to ready state: ", elapsedTime)
			return true
		}
	}
}

func CheckVolumeDeletionById(storage_endpoint string, token string, resource_id string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_instancebyid_after_delete, _ := storage.GetFilesystemById(storage_endpoint, token, resource_id)
		if get_instancebyid_after_delete != 404 {
			logger.Logf.Info("Volume is not yet deleted.")
			return false
		} else {
			logger.Logf.Info("Volume has been deleted.")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for volume deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckVolumeDeletionByName(storage_endpoint string, token string, name string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_instancebyname_after_delete, _ := storage.GetFilesystemByName(storage_endpoint, token, name)
		if get_instancebyname_after_delete != 404 {
			logger.Logf.Info("Volume is not yet deleted.")
			return false
		} else {
			logger.Logf.Info("Volume has been deleted.")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for volume deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckBucketProvisionState(bucket_endpoint string, token string, resource_id string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_response_byid_status, get_response_byid_body := storage.GetBucketById(bucket_endpoint, token, resource_id)
		Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)

		bucketPhase := gjson.Get(get_response_byid_body, "status.phase").String()
		Expect(bucketPhase).ToNot(Equal("BucketFailed"), "Bucket is in failed state")

		logger.Logf.Info("bucketPhase: ", bucketPhase)
		if bucketPhase != "BucketReady" {
			logger.Logf.Info("Bucket is not in ready state")
			return false
		} else {
			logger.Logf.Info("Bucket is in ready state")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for bucket to get to ready state: ", elapsedTime)
			return true
		}
	}
}

func CheckBucketDeletionById(bucket_endpoint string, token string, resource_id string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_instancebyid_after_delete, _ := storage.GetBucketById(bucket_endpoint, token, resource_id)
		if get_instancebyid_after_delete != 404 {
			logger.Logf.Info("Bucket is not yet deleted.")
			return false
		} else {
			logger.Logf.Info("Bucket has been deleted.")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for bucket deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckBucketDeletionByName(bucket_endpoint string, token string, name string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_instancebyname_after_delete, _ := storage.GetBucketByName(bucket_endpoint, token, name)
		if get_instancebyname_after_delete != 404 {
			logger.Logf.Info("Bucket is not yet deleted.")
			return false
		} else {
			logger.Logf.Info("Bucket has been deleted.")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for bucket deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckUserProvisionState(user_endpoint string, token string, user_id string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_response_byid_status, get_response_byid_body := storage.GetPrincipalById(user_endpoint, token, user_id)
		Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)

		userPhase := gjson.Get(get_response_byid_body, "status.phase").String()
		Expect(userPhase).ToNot(Equal("ObjectUserFailed"), "Object user in failed state")

		logger.Logf.Info("userPhase: ", userPhase)
		if userPhase != "ObjectUserReady" {
			logger.Logf.Info("Object user is not in ready state")
			return false
		} else {
			logger.Logf.Info("Object user is in ready state")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for object user to get to ready state: ", elapsedTime)
			return true
		}
	}
}

func CheckUserDeletionById(user_endpoint string, token string, user_id string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_instancebyid_after_delete, _ := storage.GetPrincipalById(user_endpoint, token, user_id)
		if get_instancebyid_after_delete != 404 {
			logger.Logf.Info("User is not yet deleted.")
			return false
		} else {
			logger.Logf.Info("User has been deleted.")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for user deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckUserDeletionByName(user_endpoint string, token string, name string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_instancebyname_after_delete, _ := storage.GetPrincipalByName(user_endpoint, token, name)
		if get_instancebyname_after_delete != 404 {
			logger.Logf.Info("User is not yet deleted.")
			return false
		} else {
			logger.Logf.Info("User has been deleted.")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for user deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckRuleDeletionById(rule_endpoint string, token string, resource_id string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_rulebyid_after_delete, _ := storage.GetBucketById(rule_endpoint, token, resource_id)
		if get_rulebyid_after_delete != 404 {
			logger.Logf.Info("Rule is not yet deleted.")
			return false
		} else {
			logger.Logf.Info("Rule has been deleted.")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for Rule deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckRuleProvisionState(rule_endpoint string, token string, bucket_id string, rule_id string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_response_byid_status, get_response_byid_body := storage.GetRuleById(rule_endpoint, bucket_id, token, rule_id)
		Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)

		rulePhase := gjson.Get(get_response_byid_body, "status.phase").String()
		Expect(rulePhase).ToNot(Equal("LFRuleFailed"), "Rule is in failed state")

		logger.Logf.Info("rulePhase: ", rulePhase)
		if rulePhase != "LFRuleReady" {
			logger.Logf.Info("Rule is not in ready state")
			return false
		} else {
			logger.Logf.Info("Rule is in ready state")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for rule to get to ready state: ", elapsedTime)
			return true
		}
	}
}

func GetWekaUrl(storage_endpoint string, token string, resource_id string) string {
	get_response_byid_status, get_response_byid_body := storage.GetFilesystemById(storage_endpoint, token, resource_id)
	Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)

	wekaUrl := gjson.Get(get_response_byid_body, "status.mount.clusterAddr").String()
	return wekaUrl
}

func GetVastUrl(storage_endpoint string, token string, resource_id string) string {
	get_response_byid_status, get_response_byid_body := storage.GetFilesystemById(storage_endpoint, token, resource_id)
	Expect(get_response_byid_status).To(Equal(200), get_response_byid_body)

	wekaUrl := gjson.Get(get_response_byid_body, "status.mount.clusterAddr").String()
	return wekaUrl
}
