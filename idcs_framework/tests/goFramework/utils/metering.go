package utils

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/testify/test_config"
	"os"
	"path/filepath"

	"runtime"

	"time"

	"github.com/tidwall/gjson"
)

var metconfigData string

func Get_Metering_config_file_path() (path string) {
	_, filename, _, _ := runtime.Caller(0)
	suite_path := "/" + os.Getenv("Test_Suite") + "/test_config/metering.json"
	filePath := filepath.Clean(filepath.Join(filename, "../../")) + suite_path
	return filePath
}

func Get_Metering_config_file_data() string {
	configFile := Get_Metering_config_file_path()
	// To DO Handle Error
	metconfigData, _ = test_config.LoadConfig(configFile)
	//fmt.Println("File Data", metconfigData)
	return metconfigData
}

func Get_Metering_Base_Url() string {
	url := gjson.Get(metconfigData, "urls.base_url").String()
	return url
}

func Get_Metering_GRPC_Host() string {
	host := gjson.Get(metconfigData, "grpc.host").String()
	return host
}

//Metering utils REST

func Get_Metering_Create_Payload(search_tag string) string {
	path := "metering.create" + "." + search_tag
	fmt.Println("path is", path)
	json := gjson.Get(metconfigData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Json File Is", Get_Metering_config_file_path())
	fmt.Println("Json output", gjson.Get(metconfigData, path).String())
	return gjson.Get(metconfigData, path).String()
}

func Get_Metering_Search_Payload(search_tag string) string {
	path := "metering.search" + "." + search_tag
	fmt.Println("Path is", path)
	json := gjson.Get(metconfigData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Data is", gjson.Get(metconfigData, path).String())
	return gjson.Get(metconfigData, path).String()
}

func Get_Usage_Record_Get_Payload(search_tag string) string {
	path := "metering.findPrevious" + "." + search_tag
	fmt.Println("Path is", path)
	json := gjson.Get(metconfigData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid  json", json)
	}
	fmt.Println("Data is", gjson.Get(metconfigData, path).String())
	return gjson.Get(metconfigData, path).String()
}

func Get_Usage_Record_Get_Response(search_tag string) string {
	path := "metering.findPrevious" + "." + search_tag
	fmt.Println("Path is", path)
	json := gjson.Get(metconfigData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Data is", gjson.Get(metconfigData, path).String())
	return gjson.Get(metconfigData, path).String()
}

func Get_Metering_Update_Payload(search_tag string) string {
	path := "metering.update" + "." + search_tag
	fmt.Println("Path is", path)
	json := gjson.Get(metconfigData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Data is", gjson.Get(metconfigData, path).String())
	return gjson.Get(metconfigData, path).String()
}
func Get_numofRecords() int64 {

	path := "numberofrecords" + "." + "numRecords"
	logger.Logf.Info("path is", path)
	json := gjson.Get(metconfigData, path).String()
	if !gjson.Valid(json) {
		logger.Logf.Info("invalid json", json)
	}
	logger.Logf.Info("Json File Is", Get_Metering_config_file_path())
	numRecords := gjson.Get(metconfigData, path).Int()
	return numRecords
}

func CalculateAverageTime(timesTaken []string) (time.Duration, error) {
	sum := 0 * time.Millisecond
	for _, t := range timesTaken {
		duration, err := time.ParseDuration(t)
		if err != nil {
			return 0, err
		}
		sum += duration
	}

	average := sum / time.Duration(len(timesTaken))

	return average, nil
}

func Get_PerformanceInput() int64 {

	path := "inputforperformance" + "." + "numRecordsforperf"
	logger.Logf.Info("path is", path)
	json := gjson.Get(metconfigData, path).String()
	if !gjson.Valid(json) {
		logger.Logf.Info("invalid json", json)
	}
	logger.Logf.Info("Json File Is", Get_Metering_config_file_path())
	numRecordsperf := gjson.Get(metconfigData, path).Int()
	return numRecordsperf
}
func Get_CloudaccountInput() int64 {

	path := "inputforcid" + "." + "cidnumber"
	logger.Logf.Info("path is", path)
	json := gjson.Get(metconfigData, path).String()
	if !gjson.Valid(json) {
		logger.Logf.Info("invalid json", json)
	}
	logger.Logf.Info("Json File Is", Get_Metering_config_file_path())
	numRecordscId := gjson.Get(metconfigData, path).Int()
	return numRecordscId
}
