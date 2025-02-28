package metering

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goFramework/framework/common/grpc_client"
	"goFramework/testify/test_cases/testutils"
	"time"

	"goFramework/framework/common/logger"
	"goFramework/utils"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
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
	var jsonData CreatePostStruct
	CloudAccountId := fmt.Sprint(time.Now().Nanosecond())[:6]
	for i := 0; i < RECORDS_COUNT_ALL; i++ {
		jsonData = CreatePostStruct{
			TransactionId:  (uuid.New().String()),
			CloudAccountId: CloudAccountId,
			ResourceId:     (uuid.New().String()),
			Timestamp:      "2022-11-29T13:34:00.000Z",
		}
		data, err := json.Marshal(jsonData)
		if err != nil {
			logger.Log.Info("Error marshalling json")
			return false, jsonData
		}
		jsonPayload := string(data)
		req := []byte(jsonPayload)
		reqBody := bytes.NewBuffer(req)
		_, outStr := grpc_client.ExecuteGrpcCurlRequest(jsonPayload, utils.Get_Metering_GRPC_Host(), CREATE_ENDPOINT)
		if outStr != "" {
			logger.Logf.Info("Failed to create Usage Record with request body ", reqBody)
			return false, jsonData
		}

	}
	//Validate number of usage records
	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		CloudAccountId: &CloudAccountId,
	}
	data, err := json.Marshal(searchData)
	if err != nil {
		logger.Log.Info("Error marshalling json")
		return false, jsonData
	}
	searPayload := string(data)
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(searPayload, utils.Get_Metering_GRPC_Host(), SEARCH_ENDPOINT)
	if outStr != "" {
		logger.Logf.Info("Failed to Search Usage Record with request body ", searchData)
		return false, jsonData
	}

	// Validate the output

	//Validate number of usage records
	result := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.cloudAccountId")
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
	var jsonData CreatePostStruct
	TransactionId := (uuid.New().String())
	for i := 0; i < RECORDS_COUNT_ALL; i++ {
		jsonData = CreatePostStruct{
			TransactionId:  TransactionId,
			CloudAccountId: fmt.Sprint(time.Now().Nanosecond())[:6],
			ResourceId:     (uuid.New().String()),
			Timestamp:      "2022-11-29T13:34:00.000Z",
		}

		data, err := json.Marshal(jsonData)
		if err != nil {
			logger.Log.Info("Error marshalling json")
			return false, jsonData
		}
		jsonPayload := string(data)
		_, outStr := grpc_client.ExecuteGrpcCurlRequest(jsonPayload, utils.Get_Metering_GRPC_Host(), CREATE_ENDPOINT)
		if outStr != "" {
			logger.Logf.Info("Failed to create Usage Record with request body ", jsonPayload)
			return false, jsonData
		}

	}
	//Validate number of usage records
	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		TransactionId: &TransactionId,
	}
	data, err := json.Marshal(searchData)
	if err != nil {
		logger.Log.Info("Error marshalling json")
		return false, jsonData
	}
	searPayload := string(data)
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(searPayload, utils.Get_Metering_GRPC_Host(), SEARCH_ENDPOINT)
	if outStr != "" {
		logger.Logf.Info("Failed to Search Usage Record with request body ", sejsonStr)
		return false, jsonData
	}

	// Validate the output

	//Validate number of usage records
	result := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.transactionId")
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
	create_count := 0
	var jsonData CreatePostStruct
	var timestamp string
	ResourceId := (uuid.New().String())
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
			CloudAccountId: fmt.Sprint(time.Now().Nanosecond())[:6],
			ResourceId:     ResourceId,
		}

		data, err := json.Marshal(jsonData)
		if err != nil {
			logger.Log.Info("Error marshalling json")
			return false, jsonData
		}
		jsonPayload := string(data)
		_, outStr := grpc_client.ExecuteGrpcCurlRequest(jsonPayload, utils.Get_Metering_GRPC_Host(), CREATE_ENDPOINT)
		if outStr != "" {
			logger.Logf.Info("Failed to create Usage Record with request body ", jsonPayload)
			return false, jsonData
		}
		create_count = create_count + 1

	}
	//Validate number of usage records
	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		ResourceId: &ResourceId,
	}

	data, err := json.Marshal(searchData)
	if err != nil {
		logger.Log.Info("Error marshalling json")
		return false, jsonData
	}
	searPayload := string(data)
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(searPayload, utils.Get_Metering_GRPC_Host(), SEARCH_ENDPOINT)
	if outStr != "" {
		logger.Logf.Info("Failed to Search Usage Record with request body ", searPayload)
		return false, jsonData
	}
	// Validate the output

	//Validate number of usage records
	result := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.resourceId")
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
	CloudAccountId := fmt.Sprint(time.Now().Nanosecond())[:6]
	props := map[string]string{
		"instance": "small",
	}

	jsonData := CreatePostStruct{
		Timestamp:      "2022-11-29T13:34:00.000Z",
		TransactionId:  (uuid.New().String()),
		CloudAccountId: CloudAccountId,
		ResourceId:     (uuid.New().String()),
		Properties:     props,
	}

	data, err := json.Marshal(jsonData)
	if err != nil {
		logger.Log.Info("Error marshalling json")
		return "", jsonData
	}
	jsonPayload := string(data)
	_, outStr := grpc_client.ExecuteGrpcCurlRequest(jsonPayload, utils.Get_Metering_GRPC_Host(), CREATE_ENDPOINT)
	if outStr != "" {
		logger.Logf.Info("Failed to create Usage Record with request body ", jsonPayload)
		return "", jsonData
	}
	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		CloudAccountId: &CloudAccountId,
	}
	data, err = json.Marshal(searchData)
	if err != nil {
		logger.Log.Info("Error marshalling json")
		return "", jsonData
	}
	searPayload := string(data)
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(searPayload, utils.Get_Metering_GRPC_Host(), SEARCH_ENDPOINT)

	// Validate the output
	cloudAccId := gjson.Get(sejsonStr, "cloudAccountId").String()
	// transId := gjson.Get(sejsonStr, "transactionId").String()
	// resId := gjson.Get(sejsonStr, "resourceId").String()

	if cloudAccId != CloudAccountId {
		logger.Logf.Info("Cloud Account Id Mismatch from Search Result, Actual : %s   Expected : %s ", cloudAccId, CloudAccountId)
		return "", jsonData
	}

	// if transId != jsonData.TransactionId {
	// 	logger.Logf.Info("Transaction Id Mismatch from Search Result, Actual : %s   Expected : %s "+transId, " ", jsonData.TransactionId)
	// 	return "", jsonData
	// }

	// if resId != jsonData.ResourceId {
	// 	logger.Logf.Info("Resource Id Mismatch from Search Result, Actual : %s   Expected : %s "+resId, jsonData.ResourceId)
	// 	return "", jsonData
	// }

	id := gjson.Get(sejsonStr, "id").String()
	logger.Logf.Info("Id created ", id)
	return id, jsonData
}

func Create_Usage_Record(create_tag string) (bool, string) {
	// Read Config file
	jsonData := utils.Get_Metering_Create_Payload(create_tag)
	fmt.Println("Json output in lib file", jsonData)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	jsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(jsonPayload, utils.Get_Metering_GRPC_Host(), CREATE_ENDPOINT)
	if outStr != "" && jsonStr != "" {
		return false, outStr
	}
	return true, outStr
}

func Search_Usage_Record(create_tag string) (bool, string) {
	// Read Config file
	jsonData := utils.Get_Metering_Search_Payload(create_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	jsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(jsonPayload, utils.Get_Metering_GRPC_Host(), SEARCH_ENDPOINT)
	if outStr != "" {
		logger.Logf.Info("Failed to create Usage Record with request body ", jsonPayload)
		return false, outStr
	}
	ret := Validate_Search_Response([]byte(jsonStr))
	if !ret {
		return false, outStr
	}

	return ret, outStr
}

func Search_Usage_Record_with_dynamic_filter(jsonData UsageFilter, RECORDS_COUNT_ALL int) bool {
	// Read Config file
	data, _ := json.Marshal(jsonData)
	jsonPayload := string(data)
	jsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(jsonPayload, utils.Get_Metering_GRPC_Host(), SEARCH_ENDPOINT)
	if outStr != "" {
		return false
	}

	//Validate number of usage records

	result := gjson.Parse(jsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.id")
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		count = count + 1
	}

	logger.Logf.Info("Count value is ", count)
	logger.Logf.Info("Expected Count value is ", RECORDS_COUNT_ALL)
	if count == RECORDS_COUNT_ALL {
		return true
	}

	return false
}

func Update_Usage_Record_with_dynamic_filter(jsonData UsageUpdate, id int64, data CreatePostStruct, RECORDS_COUNT_ALL int) bool {
	// Read Config file
	data1, _ := json.Marshal(jsonData)
	jsonPayload := string(data1)
	_, outStr := grpc_client.ExecuteGrpcCurlRequest(jsonPayload, utils.Get_Metering_GRPC_Host(), UPDATE_ENDPOINT)
	if outStr != "" {
		return false
	}
	//Validate number of usage records

	// Get Id by searching with cloudaccountID

	searchData := UsageFilter{
		Reported: &jsonData.Reported,
		Id:       testutils.GetInt64Pointer(id),
	}
	data1, _ = json.Marshal(searchData)
	searPayload := string(data1)
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(searPayload, utils.Get_Metering_GRPC_Host(), SEARCH_ENDPOINT)

	// Validate the output

	result := gjson.Parse(sejsonStr)
	logger.Logf.Info("Result of search", result)
	count := 0
	arr := gjson.Get(result.String(), "..#.id")
	for _, v := range arr.Array() {
		logger.Logf.Info("Result of search with filter is ", v)
		rec_id := v.Int()
		if rec_id == id {
			count = count + 1
		}

	}

	logger.Logf.Info("Count value is ", count)
	logger.Logf.Info("Expected Count value is ", RECORDS_COUNT_ALL)
	if count == RECORDS_COUNT_ALL {
		return true
	}

	return false

}

func Find_Usage_Record_with_dynamic_filter(jsonData UsagePrevious, times string, RECORDS_COUNT_ALL int) (bool, string) {
	// Read Config file
	data, _ := json.Marshal(jsonData)
	searPayload := string(data)
	resource_id := jsonData.ResourceId
	fpjsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(searPayload, utils.Get_Metering_GRPC_Host(), FINDPREVIOUS_ENDPOINT)
	if outStr != "" {
		logger.Logf.Info("Finding Previous usage record failed with error", outStr)
		return false, outStr
	}

	logger.Logf.Info("Find Previous result", fpjsonStr)
	actual_resid := gjson.Get(fpjsonStr, "resourceId").String()
	actual_timestamp := gjson.Get(fpjsonStr, "timestamp").String()

	if actual_resid != resource_id && actual_timestamp != times {
		return false, fpjsonStr
	}

	return true, fpjsonStr
}

func Find_Previous_Usage_Record(search_tag string) (bool, string) {
	jsonData := utils.Get_Usage_Record_Get_Payload(search_tag)
	searPayload := gjson.Get(jsonData, "payload").String()
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(searPayload, utils.Get_Metering_GRPC_Host(), FINDPREVIOUS_ENDPOINT)
	if outStr != "" {
		logger.Logf.Info("Finding Previous usage record failed with error", outStr)
		return false, outStr
	}

	return true, sejsonStr
}

func Update_Usage_Record(search_tag string) (bool, string) {
	// Read Config file
	jsonData := utils.Get_Metering_Update_Payload(search_tag)
	searPayload := gjson.Get(jsonData, "payload").String()
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(searPayload, utils.Get_Metering_GRPC_Host(), UPDATE_ENDPOINT)
	if outStr != "" {
		logger.Logf.Info("Finding Previous usage record failed with error", outStr)
		return false, outStr
	}

	return true, sejsonStr
}
