package testutils

import (
	"bytes"
	_ "encoding/json"
	"goFramework/framework/common/logger"
	"goFramework/utils"
	"os/exec"
	_ "strings"

	"github.com/tidwall/gjson"
)

const (
	RECORDS_COUNT_ALL       int    = 1000
	RECORDS_COUNT_ONE       int    = 1
	RECORDS_READAFTER_ID    uint64 = 10
	RECORDS_READAFTER_COUNT int    = 990
)

func Execute_Create_Usage_Record_Gcurl(payload string) (string, string) {
	cmd := exec.Command("grpcurl", "-d", payload, "-plaintext", "localhost:50051", "idc_metering.MeteringService/Create")
	logger.Logf.Infof("Create Usage Record command: %s\n ", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger.Logf.Infof("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	logger.Logf.Infof("out:\n%s\nerr:\n%s\n", outStr, errStr)
	return outStr, errStr
}

func Execute_Read_Usage_Record_Gcurl(payload string) (string, string) {
	cmd := exec.Command("grpcurl", "-d", payload, "-plaintext", "localhost:50051", "idc_metering.MeteringService/Search")
	logger.Logf.Infof("Search Usage Record command: %s\n ", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger.Logf.Infof("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	logger.Logf.Infof("out:\n%s\nerr:\n%s\n", outStr, errStr)
	return outStr, errStr
}

func Execute_Update_Usage_Record_Gcurl(payload string) (string, string) {
	cmd := exec.Command("grpcurl", "-d", payload, "-plaintext", "localhost:50051", "idc_metering.MeteringService/Update")
	logger.Logf.Infof("Update Usage Record command: %s\n ", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger.Logf.Infof("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	logger.Logf.Infof("out:\n%s\nerr:\n%s\n", outStr, errStr)
	return outStr, errStr
}

func Execute_FindPrevious_Usage_Record_Gcurl(payload string) (string, string) {
	cmd := exec.Command("grpcurl", "-d", payload, "-plaintext", "localhost:50051", "idc_metering.MeteringService/FindPrevious")
	logger.Logf.Infof("Find Previous Usage Record command: %s\n ", cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger.Logf.Infof("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdout.String(), stderr.String()
	logger.Logf.Infof("out:\n%s\nerr:\n%s\n", outStr, errStr)
	return outStr, errStr
}

func Get_Invalid_Value(action string, field string) string {
	jsonData := utils.Get_Config_Data()
	path := "metering" + "." + action + "." + "invalid_values" + "." + field
	return gjson.Get(jsonData, path).String()
}

func Get_Missing_Value(action string, field string) string {
	jsonData := utils.Get_Config_Data()
	path := "metering" + "." + action + "." + "missingFields" + "." + field
	return gjson.Get(jsonData, path).String()
}

func Get_Payload(action string, field string) string {
	jsonData := utils.Get_Config_Data()
	path := "metering" + "." + action + "." + field
	return gjson.Get(jsonData, path).String()
}
