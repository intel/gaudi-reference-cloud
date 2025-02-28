package compute_util

import (
	"encoding/json"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	. "github.com/onsi/gomega"
)

func ComputeCreateVMInstance(computeInstanceUrl string, token string, payload string) *resty.Response {
	var jsonPayload map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonPayload)
	response := client.Post(computeInstanceUrl, token, jsonPayload)

	instanceId := gjson.Get(string(response.Body()), "metadata.resourceId").String()
	instanceValidation := utils.CheckInstancePhase(computeInstanceUrl, token, instanceId)
	Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
	return response
}

func ComputeCreateBMInstance(computeInstanceUrl string, token string, payload string) *resty.Response {
	var jsonPayload map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonPayload)
	response := client.Post(computeInstanceUrl, token, jsonPayload)

	instanceId := gjson.Get(string(response.Body()), "metadata.resourceId").String()
	instanceValidation := utils.CheckInstancePhase(computeInstanceUrl, token, instanceId)
	Eventually(instanceValidation, 30*time.Minute, 10*time.Second).Should(BeTrue())
	return response
}

func ComputeGetAllInstance(computeInstanceUrl string, token string) *resty.Response {
	response := client.Get(computeInstanceUrl, token)
	return response
}

func ComputeGetInstanceById(computeInstanceUrl string, token string, instanceId string) *resty.Response {
	var getInstanceByIdUrl = computeInstanceUrl + "/id/" + instanceId
	response := client.Get(getInstanceByIdUrl, token)
	return response
}

func ComputeDeleteVMInstanceById(computeInstanceUrl string, token string, instanceId string) *resty.Response {
	var deleteInstanceByIdUrl = computeInstanceUrl + "/id/" + instanceId
	response := client.Delete(deleteInstanceByIdUrl, token)

	instanceValidation := utils.CheckInstanceDeletionById(computeInstanceUrl, token, instanceId)
	Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
	return response
}

func ComputeDeleteBMInstanceById(computeInstanceUrl string, token string, instanceId string) *resty.Response {
	var deleteInstanceByIdUrl = computeInstanceUrl + "/id/" + instanceId
	response := client.Delete(deleteInstanceByIdUrl, token)

	instanceValidation := utils.CheckInstanceDeletionById(computeInstanceUrl, token, instanceId)
	Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
	return response
}
