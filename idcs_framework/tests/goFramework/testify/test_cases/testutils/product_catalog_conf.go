package testutils

import (
	"fmt"
	"goFramework/testify/test_config"
	"os"
	"path/filepath"

	"runtime"

	"github.com/tidwall/gjson"
)

var configData string

func Get_PC_config_file_path() (path string) {
	_, filename, _, _ := runtime.Caller(0)
	suite_path := "/" + os.Getenv("Test_Suite") + "/test_config/product_catalog.json"
	fmt.Println("Suite Path", suite_path)
	fmt.Println("File Path", filename)
	filePath := filepath.Clean(filepath.Join(filename, "../../../../")) + suite_path
	return filePath
}

func Get_PC_config_file_data() {
	configFile := Get_PC_config_file_path()
	// To DO Handle Error
	configData, _ = test_config.LoadConfig(configFile)
}

func Get_PC_Base_Url() string {
	url := gjson.Get(configData, "urls.base_url").String()
	return url
}

//product catalogue utils REST

func Get_Product_Get_Payload(product_tag string) string {
	path := "productGetPayload" + "." + product_tag
	fmt.Println("path is", path)
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Json output", gjson.Get(configData, path).String())
	return gjson.Get(configData, path).String()
}

func Get_Vendor_Get_Payload(vendor_tag string) string {
	path := "vendorGetPayload" + "." + vendor_tag
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	return gjson.Get(configData, path).String()
}

func Get_Product_Get_Response(product_tag string) string {
	path := "productGetResponse" + "." + product_tag
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	return gjson.Get(configData, path).String()
}

func Get_Vendor_Get_Response(vendor_tag string) string {
	path := "vendorGetResponse" + "." + vendor_tag
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	return gjson.Get(configData, path).String()
}

// grpc utils

func Get_Product_Get_Payload_Grpc(product_tag string) string {
	path := "grpc.productGetPayload" + "." + product_tag
	fmt.Println("path is", path)
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Json output", gjson.Get(configData, path).String())
	return gjson.Get(configData, path).String()
}

func Get_Vendor_Get_Payload_Grpc(vendor_tag string) string {
	path := "grpc.vendorGetPayload" + "." + vendor_tag
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	return gjson.Get(configData, path).String()
}

func Get_Product_Get_Response_Grpc(product_tag string) string {
	path := "grpc.productGetResponse" + "." + product_tag
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Data is", gjson.Get(configData, path).String())
	return gjson.Get(configData, path).String()
}

func Get_Vendor_Get_Response_Grpc(vendor_tag string) string {
	path := "grpc.vendorGetResponse" + "." + vendor_tag
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	return gjson.Get(configData, path).String()
}

func Get_PC_Grpc_Host() string {
	url := gjson.Get(configData, "grpc.host").String()
	return url
}
