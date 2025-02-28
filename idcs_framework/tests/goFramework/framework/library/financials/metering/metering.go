package metering

import (
	"bytes"
	"encoding/json"
	"goFramework/framework/common/http_client"
	"goFramework/framework/common/logger"
	"goFramework/utils"
	"strings"

	//"strconv"

	"github.com/google/uuid"
	//"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/instance"
	"github.com/tidwall/gjson"
	//"google.golang.org/protobuf/types/known/timestamppb"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/testify/test_cases/testutils"
	"strconv"
	"time"

	_ "google.golang.org/protobuf/types/known/timestamppb"
)

// Util Functions

func Validate_Search_Response(data []byte) bool {
	var structResponse PostResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func Validate_FindPrevious_Response(data []byte) bool {
	var structResponse PreviousRecordStruct
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func Create_multiple_Records_with_CloudAccountId(RECORDS_COUNT_ALL int) (bool, CreatePostStruct) {
	url := utils.Get_Metering_Base_Url()
	search_url := utils.Get_Metering_Base_Url() + "/search"
	var jsonData CreatePostStruct
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	CloudAccountId, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	props := map[string]string{
		"availabilityZone": "us-west-1a",
		"clusterId":        "harvester1",
		"deleted":          "false",
		"instanceType":     "tiny",
		"region":           "us-west-1",
		"runningSeconds":   "192624.136468704",
	}
	for i := 0; i < RECORDS_COUNT_ALL; i++ {
		jsonData = CreatePostStruct{
			TransactionId:  (uuid.New().String()),
			CloudAccountId: CloudAccountId,
			ResourceId:     (uuid.New().String()),
			Timestamp:      "2022-11-29T13:34:00.000Z",
			Properties:     props,
		}
		jsonPayload, _ := json.Marshal(jsonData)
		req := []byte(jsonPayload)
		reqBody := bytes.NewBuffer(req)
		jsonStr, _ := http_client.Post(url, reqBody, 200)
		if jsonStr == "Failed" {
			logger.Logf.Info("Failed to create Usage Record with request body ", reqBody)
			return false, jsonData
		}

	}
	//Validate number of usage records
	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		CloudAccountId: &CloudAccountId,
	}

	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)

	// Validate the output

	//Validate number of usage records
	result := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.result.cloudAccountId")
	for _, v := range arr.Array() {
		if v.String() == jsonData.CloudAccountId {
			count = count + 1
		}
	}

	if count != RECORDS_COUNT_ALL {
		logger.Logf.Info("Expected number of records not found, Search Result is ", result)
		return false, jsonData
	}
	logger.Logf.Info("Expected number of records found, Search Result is ", result)
	return true, jsonData
}

func Create_multiple_Records_with_transactionId(RECORDS_COUNT_ALL int) (bool, CreatePostStruct) {
	url := utils.Get_Metering_Base_Url()
	search_url := utils.Get_Metering_Base_Url() + "/search"
	var jsonData CreatePostStruct
	TransactionId := (uuid.New().String())
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	CloudAccountId, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	props := map[string]string{
		"availabilityZone": "us-west-1a",
		"clusterId":        "harvester1",
		"deleted":          "false",
		"instanceType":     "tiny",
		"region":           "us-west-1",
		"runningSeconds":   "192624.136468704",
	}
	for i := 0; i < RECORDS_COUNT_ALL; i++ {
		jsonData = CreatePostStruct{
			TransactionId:  TransactionId,
			CloudAccountId: CloudAccountId,
			ResourceId:     (uuid.New().String()),
			Timestamp:      "2022-11-29T13:34:00.000Z",
			Properties:     props,
		}
		jsonPayload, _ := json.Marshal(jsonData)
		req := []byte(jsonPayload)
		reqBody := bytes.NewBuffer(req)
		jsonStr, _ := http_client.Post(url, reqBody, 200)

		if jsonStr == "Failed" {
			logger.Logf.Info("Failed to create Usage Record with request body ", reqBody)
			return false, jsonData
		}

	}
	//Validate number of usage records
	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		TransactionId: &TransactionId,
	}

	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)
	//transId := gjson.Get(sejsonStr, "result.transactionId").String()

	// Validate the output

	//Validate number of usage records
	result := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.result.transactionId")
	for _, v := range arr.Array() {
		if v.String() == jsonData.TransactionId {
			count = count + 1
		}
	}

	if count != RECORDS_COUNT_ALL {
		logger.Logf.Info("Expected number of records not found, Search Result is ", result)
		return false, jsonData
	}
	logger.Logf.Info("Expected number of records found, Search Result is ", result)
	return true, jsonData
}

func Create_multiple_Records_with_resourceId(RECORDS_COUNT_ALL int) (bool, CreatePostStruct) {
	url := utils.Get_Metering_Base_Url()
	search_url := utils.Get_Metering_Base_Url() + "/search"
	create_count := 0
	var jsonData CreatePostStruct
	var timestamp string
	ResourceId := (uuid.New().String())
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	CloudAccountId, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	props := map[string]string{
		"availabilityZone": "us-west-1a",
		"clusterId":        "harvester1",
		"deleted":          "false",
		"instanceType":     "tiny",
		"region":           "us-west-1",
		"runningSeconds":   "192624.136468704",
	}
	for i := 0; i < RECORDS_COUNT_ALL; i++ {
		if create_count <= 8 {
			timestamp = "2022-11-29T13:34:00.000Z"
		}
		if create_count > 8 {
			timestamp = "2022-11-30T13:34:45Z"
		}

		jsonData = CreatePostStruct{
			Timestamp:      timestamp,
			TransactionId:  (uuid.New().String()),
			CloudAccountId: CloudAccountId,
			ResourceId:     ResourceId,
			Properties:     props,
		}
		jsonPayload, _ := json.Marshal(jsonData)
		req := []byte(jsonPayload)
		reqBody := bytes.NewBuffer(req)
		jsonStr, _ := http_client.Post(url, reqBody, 200)
		if jsonStr == "Failed" {
			logger.Logf.Info("Failed to create Usage Record with request body ", reqBody)
			return false, jsonData
		}
		create_count = create_count + 1

	}
	//Validate number of usage records
	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		ResourceId: &ResourceId,
	}

	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)

	// Validate the output

	//Validate number of usage records
	result := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.result.resourceId")
	for _, v := range arr.Array() {
		if v.String() == jsonData.ResourceId {
			count = count + 1
		}
	}

	if count != RECORDS_COUNT_ALL {
		logger.Logf.Info("Expected number of records not found, Search Result is ", result)
		return false, jsonData
	}
	logger.Logf.Info("Expected number of records found, Search Result is ", result)
	return true, jsonData
}

func Create_Record_and_Get_Id() (string, CreatePostStruct) {
	url := utils.Get_Metering_Base_Url()
	search_url := utils.Get_Metering_Base_Url() + "/search"
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	CloudAccountId, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	props := map[string]string{

		"availabilityZone": "us-west-1a",
		"clusterId":        "harvester1",
		"deleted":          "false",
		"instanceType":     "tiny",
		"region":           "us-west-1",
		"runningSeconds":   "192624.136468704",
	}

	jsonData := CreatePostStruct{
		Timestamp:      "2023-02-18T20:23:53.136481Z",
		TransactionId:  (uuid.New().String()),
		CloudAccountId: CloudAccountId,
		ResourceId:     (uuid.New().String()),
		Properties:     props,
	}

	jsonPayload, _ := json.Marshal(jsonData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, 200)
	if jsonStr == "Failed" {
		return "", jsonData
	}
	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		CloudAccountId: &CloudAccountId,
	}
	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)
	// Validate the output
	cloudAccId := gjson.Get(sejsonStr, "result.cloudAccountId").String()
	transId := gjson.Get(sejsonStr, "result.transactionId").String()
	resId := gjson.Get(sejsonStr, "result.resourceId").String()

	if cloudAccId != CloudAccountId {
		logger.Logf.Info("Cloud Account Id Mismatch from Search Result, Actual : %s   Expected : %s "+cloudAccId, CloudAccountId)
		return "", jsonData
	}

	if transId != jsonData.TransactionId {
		logger.Logf.Info("Transaction Id Mismatch from Search Result, Actual : %s   Expected : %s "+transId, " ", jsonData.TransactionId)
		return "", jsonData
	}

	if resId != jsonData.ResourceId {
		logger.Logf.Info("Resource Id Mismatch from Search Result, Actual : %s   Expected : %s "+resId, jsonData.ResourceId)
		return "", jsonData
	}

	id := gjson.Get(sejsonStr, "result.id").String()
	logger.Logf.Info("Id created ", id)
	return id, jsonData
}

func Create_Usage_Record(create_tag string, expected_status_code int) (bool, string) {
	// Read Config file
	url := utils.Get_Metering_Base_Url()
	jsonData := utils.Get_Metering_Create_Payload(create_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != 0 {
		if jsonStr == "nil" {
			return true, jsonStr
		}
	}
	return true, jsonStr
}

func Get_Invalid_Metering_Record(jsonData SearchInvalidMeteringRecord, expected_status_code int, RECORDS_COUNT_ALL int, reason string) (bool, string) {
	// Read Config file
	url := utils.Get_Metering_Base_Url() + "/invalid/search"
	jsonPayload, _ := json.Marshal(jsonData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, ""
	}
	id := gjson.Get(jsonStr, "id").String()

	//Validate number of usage records

	result := gjson.Parse(jsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		r := v.Get("meteringRecordInvalidityReason").String()
		count = count + 1
		if r != reason {
			logger.Logf.Infof("Reason for invalid metering record did not match, Expected : %s, Actual :%s", reason, r)
			return false, id
		}
	}

	logger.Logf.Info("Actual Number of records ", count)
	logger.Logf.Info("Expected Count value is ", RECORDS_COUNT_ALL)
	if count == RECORDS_COUNT_ALL {
		return true, id
	}

	return false, id
}

func Create_Usage_RecordDynamically(expected_status_code int) (string, CreatePostStruct) {
	url := utils.Get_Metering_Base_Url()
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	CloudAccountId, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	createUsageRecord := CreatePostStruct{
		TransactionId:  utils.GenerateString(10),
		ResourceId:     utils.GenerateString(10),
		CloudAccountId: CloudAccountId,
		Timestamp:      "2022-11-29T13:34:00.000Z",
		Properties: map[string]string{
			"instance": "small",
		},
	}
	jsonPayload, _ := json.Marshal(createUsageRecord)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return "False", createUsageRecord
	}
	logger.Logf.Infof("")
	if expected_status_code != 0 {
		if jsonStr == "nil" {
			return "true", createUsageRecord
		}
	}
	id := gjson.Get(jsonStr, "id").String()
	return id, createUsageRecord

}

func Search_Usage_Record(create_tag string, expected_status_code int) (bool, string) {
	// Read Config file
	url := utils.Get_Metering_Base_Url() + "/search"
	jsonData := utils.Get_Metering_Search_Payload(create_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, expected_status_code)
	res := []byte(jsonStr)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	if expected_status_code != 0 {
		if jsonStr == "nil" {
			return true, jsonStr
		} else {
			return false, jsonStr
		}
	}
	ret := Validate_Search_Response(res)
	if !ret {
		return false, jsonStr
	}

	return ret, jsonStr
}

func Search_Usage_Record_with_dynamic_filter(jsonData UsageFilter, expected_status_code int, RECORDS_COUNT_ALL int) (bool, string) {
	// Read Config file
	url := utils.Get_Metering_Base_Url() + "/search"
	jsonPayload, _ := json.Marshal(jsonData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, ""
	}
	id := gjson.Get(jsonStr, "id").String()

	//Validate number of usage records

	result := gjson.Parse(jsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1
	}

	logger.Logf.Info("Actual Number of records ", count)
	logger.Logf.Info("Expected Count value is ", RECORDS_COUNT_ALL)
	if count == RECORDS_COUNT_ALL {
		return true, id
	}

	return false, id
}

func Search_Metering_Records(jsonData UsageFilter, expected_status_code int) (string, int) {
	// Read Config file
	url := utils.Get_Metering_Base_Url() + "/search"
	jsonPayload, _ := json.Marshal(jsonData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, code := http_client.Post(url, reqBody, expected_status_code)
	return jsonStr, code
}

func Update_Usage_Record_with_dynamic_filter(jsonData UsageUpdate, id string, data CreatePostStruct, expected_status_code int, RECORDS_COUNT_ALL int) bool {
	// Read Config file
	url := utils.Get_Metering_Base_Url()
	search_url := utils.Get_Metering_Base_Url() + "/search"
	jsonPayload, _ := json.Marshal(jsonData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, code := http_client.Patch(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false
	}
	logger.Logf.Info("Response from API : ", code)
	logger.Logf.Info("Expected Response from API : ", expected_status_code)
	if code != expected_status_code {
		return false
	}

	//Validate number of usage records

	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		Reported:       &jsonData.Reported,
		CloudAccountId: &data.CloudAccountId,
		ResourceId:     &data.ResourceId,
		TransactionId:  &data.TransactionId,
	}
	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)

	// Validate the output

	result := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		rec_id := v.Get("id").String()
		if rec_id == id {
			count = count + 1
		}

	}

	logger.Logf.Info("Actual Number of records ", count)
	logger.Logf.Info("Expected Number of Records ", RECORDS_COUNT_ALL)
	if count == RECORDS_COUNT_ALL {
		return true
	}

	return false

}

func Find_Usage_Record_with_dynamic_filter(jsonData UsagePrevious, times string, expected_status_code int, RECORDS_COUNT_ALL int) (bool, string) {
	// Read Config file
	resource_id := jsonData.ResourceId
	id := jsonData.Id
	url := utils.Get_Metering_Base_Url() + "/previous" + "?"
	if id != "" && resource_id != "" {
		url = url + "id=" + id + "&resourceId=" + resource_id
	}

	if id != "" && resource_id == "" {
		url = url + "id=" + id
	}

	if id == "" && resource_id != "" {
		url = url + "resourceId=" + resource_id
	}
	logger.Logf.Info("Find Usage Record URL  : ", url)
	search_url := utils.Get_Metering_Base_Url() + "/search"
	jsonStr, statusCode := http_client.Get(url, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}

	if expected_status_code != 200 {
		if statusCode == expected_status_code {
			return true, jsonStr
		}
	}
	logger.Logf.Info("Find Previous result", jsonStr)
	searchData := UsageFilter{
		ResourceId: &resource_id,
	}

	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)

	//Validate number of usage records
	result := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(jsonStr, "timestamp").String()
	logger.Logf.Info("Expected time stamp ", times)
	logger.Logf.Info("Actual time stamp ", arr)

	if arr == times {
		count = count + 1
	}
	if count != 1 {
		logger.Logf.Info("Expected number of records not found, Search Result is ", jsonStr)
		return false, jsonStr
	}
	logger.Logf.Info("Expected number of records found, Search Result is ", result)
	return true, jsonStr
}

func Find_Previous_Usage_Record(search_tag string, expected_status_code int) (bool, string) {
	jsonData := utils.Get_Usage_Record_Get_Payload(search_tag)
	resource_id := gjson.Get(jsonData, "resourceId").String()
	id := gjson.Get(jsonData, "id").String()
	url := utils.Get_Metering_Base_Url() + "/previous" + "?"
	if id != "" && resource_id != "" {
		url = url + "id=" + id + "&resourceId=" + resource_id
	}

	if id != "" && resource_id == "" {
		url = url + "id=" + id
	}

	if id == "" && resource_id != "" {
		url = url + "resourceId=" + resource_id
	}
	logger.Logf.Info("Find Usage Record URL  : ", url)
	jsonStr, _ := http_client.Get(url, expected_status_code)
	logger.Logf.Infof("Get response is %s:", jsonStr)
	if jsonStr == "Failed" {
		return false, jsonStr
	}

	if expected_status_code != 0 {
		return true, jsonStr
	}
	flag := Validate_FindPrevious_Response([]byte(jsonStr))
	return flag, jsonStr

}

func Update_Usage_Record(search_tag string, expected_status_code int) (bool, string) {
	// Read Config file
	url := utils.Get_Metering_Base_Url()
	jsonData := utils.Get_Metering_Update_Payload(search_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Patch(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	if expected_status_code != 0 {
		if jsonStr == "nil" {
			return true, jsonStr
		} else {
			return false, jsonStr
		}
	}

	return true, jsonStr
}

func getMeteringRecords(url string, expected_status_code int) (string, int) {
	responseBody, responseCode := http_client.Get(url, expected_status_code)
	if responseBody == "Failed" {
		return "null", responseCode
	} else {
		return responseBody, responseCode
	}
}

func GetMeteringRecordsWithParams(url string, params string, expected_status_code int) (string, int) {
	metering_endpoing_url := url + "?" + "resourceId=" + params
	get_response_body, get_response_status := getMeteringRecords(metering_endpoing_url, expected_status_code)
	return get_response_body, get_response_status
}

func Create_Duplicate_Record_and_Get_Id(cloudaccountid string, resourceid string, transactionid string) (string, CreatePostStruct) {
	url := utils.Get_Metering_Base_Url()
	search_url := utils.Get_Metering_Base_Url() + "/search"
	CloudAccountId := cloudaccountid
	props := map[string]string{
		"availabilityZone": "us-west-1a",
		"clusterId":        "harvester1",
		"deleted":          "false",
		"instanceType":     "tiny",
		"region":           "us-west-1",
		"runningSeconds":   "192624.136468704",
	}

	jsonData := CreatePostStruct{
		Timestamp:      "2022-11-29T13:34:00.000Z",
		TransactionId:  transactionid,
		CloudAccountId: CloudAccountId,
		ResourceId:     resourceid,
		Properties:     props,
	}

	jsonPayload, _ := json.Marshal(jsonData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, 200)
	if jsonStr == "Failed" {
		return "", jsonData
	}
	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		CloudAccountId: &CloudAccountId,
	}
	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)
	// Validate the output
	cloudAccId := gjson.Get(sejsonStr, "result.cloudAccountId").String()
	transId := gjson.Get(sejsonStr, "result.transactionId").String()
	resId := gjson.Get(sejsonStr, "result.resourceId").String()

	if cloudAccId != CloudAccountId {
		logger.Logf.Info("Cloud Account Id Mismatch from Search Result, Actual : %s   Expected : %s "+cloudAccId, CloudAccountId)
		return "", jsonData
	}

	if transId != jsonData.TransactionId {
		logger.Logf.Info("Transaction Id Mismatch from Search Result, Actual : %s   Expected : %s "+transId, " ", jsonData.TransactionId)
		return "", jsonData
	}

	if resId != jsonData.ResourceId {
		logger.Logf.Info("Resource Id Mismatch from Search Result, Actual : %s   Expected : %s "+resId, jsonData.ResourceId)
		return "", jsonData
	}

	id := gjson.Get(sejsonStr, "result.id").String()
	logger.Logf.Info("Id created ", id)
	return id, jsonData
}

func Create_CloudAccount_and_Get_Id(cloudaccountid string, resourceid string) (bool, string) {
	url := utils.Get_Metering_Base_Url()
	//search_url := utils.Get_Metering_Base_Url() + "/search"
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	CloudAccountId, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	ResourceId := resourceid
	Timestamp := time.Now().Format(time.RFC3339)

	props := map[string]string{

		"availabilityZone":    "us-west-1a",
		"clusterId":           "harvester1",
		"deleted":             "false",
		"instanceType":        "tiny",
		"region":              "us-west-1",
		"runningSeconds":      "192624.136468704",
		"serviceType":         "ComputeAsAService",
		"firstReadyTimestamp": "2023-03-15T22:23:47Z",
	}

	jsonData := CreatePostStruct{
		Timestamp:      Timestamp,
		TransactionId:  (uuid.New().String()),
		CloudAccountId: CloudAccountId,
		ResourceId:     ResourceId,
		Properties:     props,
		Reported:       false,
	}

	jsonPayload, _ := json.Marshal(jsonData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, 200)
	if jsonStr == "Failed" {
		return false, jsonStr
	}

	return true, jsonStr
}

func extractValue(line string) string {
	// Split the line by whitespace
	parts := strings.Fields(line)

	// Assuming the value is the second part (index 1)
	if len(parts) >= 2 {
		return parts[1]
	}

	return ""
}
func Search_Usage_Record_with_Cloud_AccountId(jsonData UsageFilter, expected_status_code int) (bool, int) {
	// Read Config file
	url := utils.Get_Metering_Base_Url() + "/search"
	search_url := utils.Get_Metering_Base_Url() + "/search"
	jsonPayload, _ := json.Marshal(jsonData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	CloudAccountId := jsonData.CloudAccountId
	Id := jsonData.Id
	jsonStr, _ := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, 0

	}
	//id := gjson.Get(jsonStr, "id").String()

	//Validate number of usage records

	result := gjson.Parse(jsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1
	}
	searchData := UsageFilter{
		CloudAccountId: CloudAccountId,
		Id:             Id,
	}
	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)
	cloudAccId := gjson.Get(sejsonStr, "result.cloudAccountId").String()

	logger.Logf.Info("Actual Number of records ", count)
	if cloudAccId == *jsonData.CloudAccountId {
		return true, count
	}

	return false, count
}

func Create_Record_and_Get_Id_Performance_Testing(cloudaccountid string, resourceid string) (string, string) {
	url := utils.Get_Metering_Base_Url()
	search_url := utils.Get_Metering_Base_Url() + "/search"
	name, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	CloudAccountId, _ := cloudAccounts.CreateCloudAccount(name, oid, owner, parentid, tid, false, false, false, false, false,
		true, false, "ACCOUNT_TYPE_INTEL", 200)
	ResourceId := resourceid
	Timestamp := time.Now().Format(time.RFC3339)
	props := map[string]string{

		"availabilityZone":    "us-west-1a",
		"clusterId":           "harvester1",
		"deleted":             "false",
		"instanceType":        "tiny",
		"region":              "us-west-1",
		"runningSeconds":      "192624.136468704",
		"serviceType":         "ComputeAsAService",
		"firstReadyTimestamp": "2023-03-15T22:23:47Z",
	}

	jsonData := CreatePostStruct{
		Timestamp:      Timestamp,
		TransactionId:  (uuid.New().String()),
		CloudAccountId: CloudAccountId,
		ResourceId:     ResourceId,
		Properties:     props,
	}

	jsonPayload, _ := json.Marshal(jsonData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, 200)
	if jsonStr == "Failed" {
		return "", ""
	}
	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		CloudAccountId: &CloudAccountId,
		ResourceId:     &ResourceId,
		StartTime:      &Timestamp,
		TransactionId:  &jsonData.TransactionId,
	}
	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)
	// Validate the output
	cloudAccId := gjson.Get(sejsonStr, "result.cloudAccountId").String()
	transId := gjson.Get(sejsonStr, "result.transactionId").String()
	resId := gjson.Get(sejsonStr, "result.resourceId").String()

	if cloudAccId != CloudAccountId {
		logger.Logf.Info("Cloud Account Id Mismatch from Search Result, Actual : %s   Expected : %s "+cloudAccId, CloudAccountId)
		return "", sejsonStr
	}

	if transId != jsonData.TransactionId {
		logger.Logf.Info("Transaction Id Mismatch from Search Result, Actual : %s   Expected : %s "+transId, " ", jsonData.TransactionId)
		return "", sejsonStr
	}

	if resId != jsonData.ResourceId {
		logger.Logf.Info("Resource Id Mismatch from Search Result, Actual : %s   Expected : %s "+resId, jsonData.ResourceId)
		return "", sejsonStr
	}

	id := gjson.Get(sejsonStr, "result.id").String()
	logger.Logf.Info("Id created ", id)
	return id, sejsonStr
}

func Find_Usage_Record_with_id(jsonData UsagePrevious, expected_status_code int, total_records int) (bool, int) {
	// Read Config file
	resource_id := jsonData.ResourceId
	id := jsonData.Id
	url := utils.Get_Metering_Base_Url() + "/previous" + "?"
	if id != "" && resource_id != "" {
		url = url + "id=" + id + "&resourceId=" + resource_id
	}

	if id != "" && resource_id == "" {
		url = url + "id=" + id
	}

	if id == "" && resource_id != "" {
		url = url + "resourceId=" + resource_id
	}
	logger.Logf.Info("Find Usage Record URL  : ", url)
	search_url := utils.Get_Metering_Base_Url() + "/search"
	jsonStr, statusCode := http_client.Get(url, expected_status_code)
	if jsonStr == "Failed" {
		return false, 0
	}

	if expected_status_code != 200 {
		if statusCode != expected_status_code {
			return false, 0
		}
	}
	logger.Logf.Info("Find Previous result", jsonStr)
	result := gjson.Parse(jsonStr)
	logger.Logf.Info("Result of search", result)
	idf := result.Get("id").String()
	logger.Logf.Info("value of idf ", idf)
	err_msg := result.Get("code").String()
	logger.Logf.Info("value of error code", err_msg)

	res, _ := strconv.ParseInt(idf, 10, 64)
	searchData := UsageFilter{
		Id: testutils.GetInt64Pointer(res),
	}

	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)

	//Validate number of usage records
	result1 := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result1)
	searchid := result1.Get("result.id").String()
	logger.Logf.Info("value of searchid is ", searchid)
	count := 0

	if searchid == idf && searchid != "" && err_msg != "5" {
		count = count + 1
		return true, count
	}
	if err_msg == "5" {
		//count = count + 1
		return true, 0
	}
	/*if count != 1  {
		logger.Logf.Info("Expected number of records not found, Search Result is ", sejsonStr)
		return false, 0
	}*/

	return false, 0

}

func Update_Usage_Record_with_rid(jsonData UsageUpdate, rid string, id *int64, expected_status_code int, RECORDS_COUNT_ALL int) bool {
	// Read Config file
	url := utils.Get_Metering_Base_Url()
	search_url := utils.Get_Metering_Base_Url() + "/search"
	jsonPayload, _ := json.Marshal(jsonData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, code := http_client.Patch(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false
	}
	logger.Logf.Info("Response from API : ", code)
	logger.Logf.Info("Expected Response from API : ", expected_status_code)
	if code != expected_status_code {
		return false
	}

	//Validate number of usage records

	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{

		ResourceId: &rid,
		Id:         id,
	}
	searPayload, _ := json.Marshal(searchData)
	req1 := []byte(searPayload)
	sereqBody := bytes.NewBuffer(req1)
	sejsonStr, _ := http_client.Post(search_url, sereqBody, 200)

	// Validate the output

	result := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.result")
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		rec_id := v.Get("id").String()

		//res, _ := strconv.ParseInt(rec_id, 10, 64)

		//res1 := testutils.GetInt64Pointer(res)
		logger.Logf.Info("value of rec_id is ", rec_id)
		id1 := strconv.FormatInt(*id, 10)
		logger.Logf.Info("value of id ", id1)
		if rec_id == id1 {

			count = count + 1
		}

	}

	logger.Logf.Info("Actual Number of records ", count)
	logger.Logf.Info("Expected Number of Records ", RECORDS_COUNT_ALL)
	if count == RECORDS_COUNT_ALL {
		return true
	}

	return false

}
