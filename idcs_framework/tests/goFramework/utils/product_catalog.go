package utils

import (
        "fmt"
        "goFramework/testify/test_config"
        "os"
        "path/filepath"

        "runtime"

        "github.com/tidwall/gjson"
)


func Get_PC_config_file_path() (path string) {
        _, filename, _, _ := runtime.Caller(0)
        suite_path := "/" + os.Getenv("Test_Suite") + "/test_config/product_catalog.json"
        filePath := filepath.Clean(filepath.Join(filename, "../../")) + suite_path
        return filePath
}

func Get_PC_config_file_data() string {
        configFile := Get_PC_config_file_path()
        // To DO Handle Error
        configData, _ := test_config.LoadConfig(configFile)
        return configData
}

func Get_PC_Base_Url() string {
        pcjsonData := Get_PC_config_file_data()
        url := gjson.Get(pcjsonData, "urls.base_url").String()
        return url
}

func Get_PC_GRPC_Host() string {
        configData := Get_PC_config_file_data()
        host := gjson.Get(configData, "grpc.status_host").String()
        return host
}

//product catalogue utils

func Get_Product_Get_Payload(product_tag string) string {
        pcjsonData := Get_PC_config_file_data()
        path := "productGetPayload" + "." + product_tag
        json := gjson.Get(pcjsonData, path).String()
        if !gjson.Valid(json) {
                fmt.Println("invalid json", json)
        }        
        return gjson.Get(pcjsonData, path).String()
}

func Get_Product_Get_Payload_Admin(product_tag string) string {
        pcjsonData := Get_PC_config_file_data()
        path := "productGetPayloadAdmin" + "." + product_tag + ".payload"
        json := gjson.Get(pcjsonData, path).String()
        if !gjson.Valid(json) {
                fmt.Println("invalid json", json)
        }        
        return gjson.Get(pcjsonData, path).String()
}

func Get_Vendor_Get_Payload(vendor_tag string) string {
        pcjsonData := Get_PC_config_file_data()
        path := "vendorGetPayload" + "." + vendor_tag
        json := gjson.Get(pcjsonData, path).String()
        if !gjson.Valid(json) {
                fmt.Println("invalid json", json)
        }
        return gjson.Get(pcjsonData, path).String()
}

func Get_Product_Get_Response(product_tag string) string {
        pcjsonData := Get_PC_config_file_data()
        path := "productGetResponse" + "." + product_tag
        json := gjson.Get(pcjsonData, path).String()
        if !gjson.Valid(json) {
                fmt.Println("invalid json", json)
        }
        return gjson.Get(pcjsonData, path).String()
}

func Get_Vendor_Get_Response(vendor_tag string) string {
        pcjsonData := Get_PC_config_file_data()
        path := "vendorGetResponse" + "." + vendor_tag
        json := gjson.Get(pcjsonData, path).String()
        if !gjson.Valid(json) {
                fmt.Println("invalid json", json)
        }
        return gjson.Get(pcjsonData, path).String()
}
