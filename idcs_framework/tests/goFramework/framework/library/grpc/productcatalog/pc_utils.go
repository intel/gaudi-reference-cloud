package productcatalog

import (
	"bytes"
	"encoding/json"
	_ "encoding/json"
	"fmt"

	"goFramework/framework/common/grpc_client"
	"goFramework/framework/common/logger"
	"goFramework/testify/test_cases/testutils"
	"goFramework/utils"
	"os/exec"
	_ "strings"

	"github.com/nsf/jsondiff"
	"github.com/tidwall/gjson"
)

func Validate_Get_Response(data []byte, jsonValidateData string, jsonResponseData string) bool {
	opts := jsondiff.DefaultConsoleOptions()
	result, _ := jsondiff.Compare([]byte(jsonResponseData), []byte(jsonValidateData), &opts)
	return result == jsondiff.FullMatch
}

func Get_Vendors(vendor_tag string, response_tag string) (bool, string) {
	// Read Config file
	jsonData := testutils.Get_Vendor_Get_Payload_Grpc(vendor_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	host := testutils.Get_PC_Grpc_Host()
	req := []byte(jsonPayload)
	output, error := Execute_Read_Vendors_Gcurl(jsonPayload, host)
	fmt.Println("Error is ", error)
	if response_tag == "negative_test" {
		if error == "nil" {
			return false, error
		}
		return true, error

	} else {
		jsonResponseExp := testutils.Get_Vendor_Get_Response_Grpc(response_tag)
		jsonResponseExp = gjson.Get(jsonResponseExp, "response").String()
		logger.Logf.Infof("Read response is %s:", jsonResponseExp)
		flag := Validate_Get_Response(req, jsonResponseExp, output)
		fmt.Println("Test result", flag)
		return flag, error

	}

}

func Get_Products(Product_tag string, response_tag string) (bool, string) {
	// Read Config file
	jsonData := testutils.Get_Product_Get_Payload_Grpc(Product_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	host := testutils.Get_PC_Grpc_Host()
	req := []byte(jsonPayload)
	output, error := Execute_Read_Products_Gcurl(jsonPayload, host)
	fmt.Println("Error is ", error)
	if response_tag == "negative_test" {
		if error == "nil" {
			return false, error
		}
		return true, error

	} else {
		jsonResponseExp := testutils.Get_Product_Get_Response_Grpc(response_tag)
		jsonResponseExp = gjson.Get(jsonResponseExp, "response").String()
		logger.Logf.Infof("Read response is %s:", jsonResponseExp)
		flag := Validate_Get_Response(req, jsonResponseExp, output)
		fmt.Println("Test result", flag)
		return flag, error

	}

}

func Execute_Read_Vendors_Gcurl(payload string, grpc_host string) (string, string) {
	cmd := exec.Command("grpcurl", "-d", payload, "-plaintext", grpc_host, "proto.ProductVendorService/Read")
	logger.Logf.Infof("Search Vendors command: %s\n ", cmd)
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

func Execute_Read_Products_Gcurl(payload string, grpc_host string) (string, string) {
	cmd := exec.Command("grpcurl", "-d", payload, "-plaintext", grpc_host, "proto.ProductCatalogService/Read")
	logger.Logf.Infof("Search Vendors command: %s\n ", cmd)
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

func Set_Status(status string, err string, familyId string, productId string, vendorId string, expected_status_code int) bool {
	statusPayload := &SetStatusPayload{
		Status: []StatusPayload{
			{
				Error:     err,
				FamilyId:  familyId,
				ProductId: productId,
				Status:    status,
				VendorId:  vendorId,
			},
		},
	}
	data, _ := json.Marshal(statusPayload)
	jsonPayload := string(data)
	host := utils.Get_PC_GRPC_Host()
	sejsonStr, outStr := grpc_client.ExecuteGrpcCurlRequest(jsonPayload, host, SETSTATUS_ENDPOINT)
	if outStr != "" {
		logger.Logf.Info("Failed to Set Status of a product ", sejsonStr)
		return false
	}
	result := gjson.Parse(sejsonStr)
	logger.Logf.Infof("Get response is %s:", result)
	return true
}
