package financials

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goFramework/framework/common/http_client"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/utils"
	"os"

	"github.com/tidwall/gjson"
)

func Validate_Get_Products_Response_Struct(data []byte) bool {
	var structResponse GetProductsResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func Validate_Get_Vendors_Response_Struct(data []byte) bool {
	var structResponse GetVendorsResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func Get_Vendors(vendor_tag string, expected_status_code int) bool {
	// Read Config file
	jsonData := utils.Get_Vendor_Get_Payload(vendor_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	username := "sys_devcloudgen2@intel.com"
	token, _ := auth.Get_Azure_Bearer_Token(username)
	os.Setenv("cloudAccTest", "True")
	os.Setenv("cloudAccToken", token)
	if os.Getenv("CloudAccountId") == "" {
		body, code := Get_CloudAccount(username, token)
		if code != 200 {
			fmt.Println("Failed to get cloudaccount.")
		}
		cloud_account_created := gjson.Get(body, "id").String()
		cloudaccount_type := gjson.Get(body, "type").String()
		fmt.Println("Cloud Account Type: ", cloudaccount_type)
		fmt.Println("Cloud Account ID: ", cloud_account_created)
		os.Setenv("CloudAccountId", cloud_account_created)
	}
	fmt.Println("Vendor Payload: ", jsonPayload)
	jsonResponseExp := utils.Get_Vendor_Get_Response("getvendor")
	jsonPayload = gjson.Get(jsonResponseExp, "response").String()
	filter := gjson.Get(jsonPayload, "productFilter").String()
	if filter == "" {
		filter = `{}`
	}
	//jsonResponseExp := utils.Get_Product_Get_Response("getproduct")
	fmt.Println("Filter: ", filter)
	productFilter := fmt.Sprintf(`{
		"cloudaccountId": "%s",
		"productFilter": %s
	}`, os.Getenv("CloudAccountId"), filter)

	logger.Logf.Info("POST FINAL Payload", productFilter)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	url := utils.Get_PC_Base_Url() + "vendors"
	jsonStr, statusCode := http_client.Get_With_Payload(url, reqBody, expected_status_code)
	logger.Logf.Infof("Get response is %s:", jsonStr)
	if jsonStr == "Failed" {
		return false
	}
	req = []byte(jsonStr)
	if expected_status_code != 200 {
		if statusCode != expected_status_code {
			return false
		}
		return true
	}
	flag := Validate_Get_Vendors_Response_Struct([]byte(jsonStr))
	return flag

}

func Get_CloudAccount(username string, token string) (string, int) {
	url := utils.Get_PC_Base_Url() + "cloudaccounts/name/" + username
	adminToken := auth.Get_Azure_Admin_Bearer_Token()
	fmt.Println("AC URL ", url)
	fmt.Println(adminToken)
	bearer := "Bearer " + adminToken
	fmt.Println("Bearer Admin: ", bearer)
	//body, statusCode := http_client.Get(url, 200)
	body, statusCode := http_client.GetOIDC(url, bearer, 200)
	enroll_payload := fmt.Sprintf(`{
		"premium": false,
		"termsStatus": true
	}`)
	req := []byte(enroll_payload)
	reqBody := bytes.NewBuffer(req)
	if statusCode == 404 {
		bearer := "Bearer " + token
		fmt.Println("Bearer Normal User: ", bearer)
		fmt.Println("CloudAccount not found, Creating it...")
		enrollUrl := utils.Get_PC_Base_Url() + "cloudaccounts/enroll"
		body, statusCode = http_client.PostOIDC(enrollUrl, reqBody, 200, bearer)
	}
	return body, statusCode
}

func Get_Products(Product_tag string, expected_status_code int) bool {
	// Read Config file
	jsonData := utils.Get_Product_Get_Payload(Product_tag)
	logger.Logf.Info("POST Payload", jsonData)
	jsonPayload := jsonData
	logger.Logf.Info("POST Payload", jsonPayload)
	username := "sys_devcloudgen2@intel.com"
	token, _ := auth.Get_Azure_Bearer_Token(username)
	os.Setenv("cloudAccTest", "True")
	os.Setenv("cloudAccToken", token)
	if os.Getenv("CloudAccountId") == "" {
		body, code := Get_CloudAccount(username, token)
		if code != 200 {
			fmt.Println("Failed to get cloudaccount.")
		}
		cloud_account_created := gjson.Get(body, "id").String()
		cloudaccount_type := gjson.Get(body, "type").String()
		fmt.Println("Cloud Account Type: ", cloudaccount_type)
		fmt.Println("Cloud Account ID: ", cloud_account_created)
		os.Setenv("CloudAccountId", cloud_account_created)
	}

	filter := gjson.Get(jsonPayload, "productFilter").String()
	if filter == "" {
		filter = `{}`
	}
	//jsonResponseExp := utils.Get_Product_Get_Response("getproduct")
	fmt.Println("Filter: ", filter)
	productFilter := fmt.Sprintf(`{
		"cloudaccountId": "%s",
		"productFilter": %s
	}`, os.Getenv("CloudAccountId"), filter)

	logger.Logf.Info("POST FINAL Payload", productFilter)
	req := []byte(productFilter)
	reqBody := bytes.NewBuffer(req)
	url := utils.Get_PC_Base_Url() + "products"
	fmt.Println("URL: ", url)
	adminToken := auth.Get_Azure_Admin_Bearer_Token()
	adminBearer := "Bearer " + adminToken
	jsonStr, statusCode := http_client.PostOIDC(url, reqBody, expected_status_code, adminBearer)
	logger.Logf.Infof("Get response is %s:", jsonStr)
	logger.Logf.Infof("Get Status code is %s:", statusCode)
	if jsonStr == "Failed" {
		return false
	}
	req = []byte(jsonStr)
	if expected_status_code != 200 {
		if statusCode != expected_status_code && expected_status_code != 401 {
			return false
		}
		return true
	}
	flag := Validate_Get_Products_Response_Struct([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	os.Unsetenv("cloudAccTest")
	return flag
}

func Get_Products_Admin(Product_tag string, expected_status_code int) bool {
	// Read Config file
	jsonData := utils.Get_Product_Get_Payload_Admin(Product_tag)
	logger.Logf.Info("POST Payload", jsonData)
	jsonPayload := jsonData
	logger.Logf.Info("POST Payload", jsonPayload)
	token := auth.Get_Azure_Admin_Bearer_Token()
	os.Setenv("cloudAccTest", "True")
	os.Setenv("cloudAccToken", token)
	//jsonResponseExp := utils.Get_Product_Get_Response("getproduct")
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	url := utils.Get_PC_Base_Url() + "products/admin"
	jsonStr, statusCode := http_client.Post(url, reqBody, expected_status_code)
	logger.Logf.Infof("Get response is %s:", jsonStr)
	logger.Logf.Infof("Get Status code is %s:", statusCode)
	if jsonStr == "Failed" {
		return false
	}
	req = []byte(jsonStr)
	if expected_status_code != 200 {
		if statusCode != expected_status_code {
			return false
		}
		return true
	}
	flag := Validate_Get_Products_Response_Struct([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	os.Unsetenv("cloudAccTest")
	return flag
}

func getVendorResponse(url string, expected_status_code int) (string, int) {
	responseBody, responseCode := http_client.Get(url, expected_status_code)
	if responseBody == "Failed" {
		return "null", responseCode
	} else {
		return responseBody, responseCode
	}
}

func GetVendorsWithParams(url string, params string, expected_status_code int) (string, int) {
	productcatalog_endpoing_url := url + "?" + "name=" + params
	get_response_body, get_response_status := getVendorResponse(productcatalog_endpoing_url, expected_status_code)
	return get_response_body, get_response_status
}

func Fetch_Product_Details_For_Status_Payload(Product_tag string, expected_status_code int) (string, string, string, bool) {
	// Read Config file
	jsonData := utils.Get_Product_Get_Payload(Product_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	url := utils.Get_PC_Base_Url() + "products"
	jsonStr, statusCode := http_client.Post(url, reqBody, expected_status_code)
	logger.Logf.Infof("Get response is %s:", jsonStr)
	logger.Logf.Infof("Get Status code is %s:", statusCode)
	if jsonStr == "Failed" {
		return "", "", "", false
	}
	req = []byte(jsonStr)
	if expected_status_code != 200 {
		if statusCode != expected_status_code {
			return "", "", "", false
		}
	}
	var structResponse GetProductsResponse
	json.Unmarshal([]byte(jsonStr), &structResponse)
	var familyId string
	var productId string
	var vendorId string
	if len(structResponse.Products) != 0 {
		familyId = structResponse.Products[0].FamilyID
		productId = structResponse.Products[0].ID
		vendorId = structResponse.Products[0].VendorID
	}
	return familyId, productId, vendorId, true
}

func Check_Creation_Time(Product_tag string, expected_status_code int) bool {
	jsonData := utils.Get_Product_Get_Payload(Product_tag)
	logger.Logf.Info("POST Payload", jsonData)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	url := utils.Get_PC_Base_Url() + "products"
	jsonStr, statusCode := http_client.Post(url, reqBody, expected_status_code)
	logger.Logf.Infof("Get response is %s:", jsonStr)
	logger.Logf.Infof("Get Status code is %s:", statusCode)
	if jsonStr == "Failed" {
		return false
	}
	req = []byte(jsonStr)
	if expected_status_code != 200 {
		if statusCode != expected_status_code {
			return false
		}
		return true
	}
	flag := Validate_Creation_Time([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	return flag
}

func Validate_Creation_Time(data []byte) bool {
	var structResponse GetProductsResponse
	err := json.Unmarshal(data, &structResponse)
	if err != nil {
		return false
	}
	for _, product := range structResponse.Products {
		if product.Created.String() == "" {
			return false
		}
	}
	return true
}

func Check_Vendor_Creation_Time(vendor_tag string, expected_status_code int) bool {
	jsonData := utils.Get_Vendor_Get_Payload(vendor_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	url := utils.Get_PC_Base_Url() + "vendors"
	jsonStr, statusCode := http_client.Get_With_Payload(url, reqBody, expected_status_code)
	logger.Logf.Infof("Get response is %s:", jsonStr)
	if jsonStr == "Failed" {
		return false
	}
	req = []byte(jsonStr)
	if expected_status_code != 200 {
		if statusCode != expected_status_code {
			return false
		}
		return true
	}
	flag := Validate_Vendors_Creation_Time([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	return flag
}

func Validate_Vendors_Creation_Time(data []byte) bool {
	var structResponse GetVendorsResponse
	err := json.Unmarshal(data, &structResponse)
	if err != nil {
		return false
	}
	for _, vendor := range structResponse.Vendors {
		if vendor.Created.String() == "" {
			return false
		}
		for _, family := range vendor.Families {
			if family.Created.String() == "" {
				return false
			}
		}
	}
	return true
}
