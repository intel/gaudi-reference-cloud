package financials

import (
	"encoding/json"
	"fmt"
	"time"

	"goFramework/framework/frisby_client"
)

func getRates(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func updateRate(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func updateRateSet(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteRate(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getMetadata(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createMetadata(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func updateMetadata(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteMetadata(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getServiceType(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createServiceType(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func updateServiceType(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteServiceType(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getVendors(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func postProductCatalog(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getRegions(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getFamilies(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getRegion(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getFamily(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func getVendor(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createVendor(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createFamily(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createRegion(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createProduct(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func updateProduct(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteProduct(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "Delete API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func updateFamily(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteFamily(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "Delete API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func updateRegion(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteRegion(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func updateVendor(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteVendor(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func whitelistCloudaccount(url string, admintoken string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, admintoken, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func CreateVendors(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var create_vendor_url = api_base_url + "/v1/vendors"
	response_status, response_body := createVendor(create_vendor_url, token, jsonMap)
	return response_status, response_body
}

func CreateRegion(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var create_region_url = api_base_url + "/v1/regions"
	response_status, response_body := createRegion(create_region_url, token, jsonMap)
	return response_status, response_body
}

func CreateFamily(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var create_region_url = api_base_url + "/v1/families"
	response_status, response_body := createFamily(create_region_url, token, jsonMap)
	return response_status, response_body
}

func CreateRateSet(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var create_region_url = api_base_url + "/v1/rateset"
	response_status, response_body := createFamily(create_region_url, token, jsonMap)
	return response_status, response_body
}

func CreateRate(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var create_region_url = api_base_url + "/v1/rate"
	response_status, response_body := createFamily(create_region_url, token, jsonMap)
	return response_status, response_body
}

func UpdateRateSet(api_base_url string, token string, ProductCatalog_api_payload string, rateSetId string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var create_region_url = api_base_url + "/v1/rateset/" + rateSetId
	response_status, response_body := updateRateSet(create_region_url, token, jsonMap)
	return response_status, response_body
}

func UpdateRate(api_base_url string, token string, ProductCatalog_api_payload string, rateId string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var create_region_url = api_base_url + "/v1/rate/" + rateId
	response_status, response_body := updateRate(create_region_url, token, jsonMap)
	return response_status, response_body
}

func DeleteRateSet(api_base_url string, token string, rateSetId string) (int, string) {
	var create_region_url = api_base_url + "/v1/rateset/" + rateSetId
	response_status, response_body := deleteRate(create_region_url, token)
	return response_status, response_body
}

func DeleteRate(api_base_url string, token string, rateId string) (int, string) {
	var create_region_url = api_base_url + "/v1/rate/" + rateId
	response_status, response_body := deleteRate(create_region_url, token)
	return response_status, response_body
}

func GetProducts(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var get_product_url = api_base_url + "/v1/products"
	response_status, response_body := postProductCatalog(get_product_url, token, jsonMap)
	return response_status, response_body
}

func GetVendors(api_base_url string, token string) (int, string) {
	var get_vendor_url = api_base_url + "/v1/vendors"
	get_response_byid_status, get_response_byid_body := getVendors(get_vendor_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetFamilies(api_base_url string, token string) (int, string) {
	var get_families_url = api_base_url + "/v1/families"
	get_response_byid_status, get_response_byid_body := getFamilies(get_families_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetVendorsV2(api_base_url string, token string) (int, string) {
	var get_vendor_url = api_base_url + "/v1/vendors"
	get_response_byid_status, get_response_byid_body := getVendors(get_vendor_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetRegions(api_base_url string, token string) (int, string) {
	var get_regions_url = api_base_url + "/v1/regions"
	get_response_byid_status, get_response_byid_body := getRegions(get_regions_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetProductsV2(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var get_products_url = api_base_url + "/v1/products"
	get_response_byid_status, get_response_byid_body := postProductCatalog(get_products_url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func GetProductsAPI(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var get_products_url = api_base_url + "/v1/api/products"
	get_response_byid_status, get_response_byid_body := postProductCatalog(get_products_url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func GetProductsV2Clustered(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var get_products_url = api_base_url + "/v1/products"
	get_response_byid_status, get_response_byid_body := postProductCatalog(get_products_url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func GetFamiliesAdmin(api_base_url string, token string) (int, string) {
	var get_families_url = api_base_url + "/v1/families"
	get_response_byid_status, get_response_byid_body := getFamilies(get_families_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetFamiliesAdminByName(api_base_url string, token string, name string) (int, string) {
	var get_families_url = api_base_url + "/v1/families?name=" + name
	get_response_byid_status, get_response_byid_body := getFamilies(get_families_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetVendorsAdmin(api_base_url string, token string) (int, string) {
	var get_vendors_url = api_base_url + "/v1/vendors"
	get_response_byid_status, get_response_byid_body := getVendors(get_vendors_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetVendorsAdminByName(api_base_url string, token string, name string) (int, string) {
	var get_vendors_url = api_base_url + "/v1/vendors?name=" + name
	get_response_byid_status, get_response_byid_body := getVendors(get_vendors_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetRegionsAdmin(api_base_url string, token string) (int, string) {
	var get_regions_url = api_base_url + "/v1/regions"
	get_response_byid_status, get_response_byid_body := getRegions(get_regions_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetRegionsAdminByName(api_base_url string, token string, name string) (int, string) {
	var get_regions_url = api_base_url + "/v1/regions?name=" + name
	get_response_byid_status, get_response_byid_body := getRegions(get_regions_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetProductsAdmin(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var post_products_url = api_base_url + "/v1/products/admin"
	get_response_byid_status, get_response_byid_body := postProductCatalog(post_products_url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func GetFamily(api_base_url string, token string, familyId string) (int, string) {
	var get_families_url = api_base_url + "/v1/families/id/" + familyId
	get_response_byid_status, get_response_byid_body := getFamily(get_families_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetVendor(api_base_url string, token string, vendorId string) (int, string) {
	var get_vendors_url = api_base_url + "/v1/vendors/id/" + vendorId
	get_response_byid_status, get_response_byid_body := getVendor(get_vendors_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetRegion(api_base_url string, token string, regionId string) (int, string) {
	var get_regions_url = api_base_url + "/v1/regions/id/" + regionId
	get_response_byid_status, get_response_byid_body := getRegion(get_regions_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetProductsAdminV2(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var post_products_url = api_base_url + "/v1/products/admin"
	get_response_byid_status, get_response_byid_body := postProductCatalog(post_products_url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func CreateChangeRequest(api_base_url string, token string, payload string) (int, string) {
	fmt.Println("Enable Change Request Payload: ", payload)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	var create_change_request_url = api_base_url + "/v1/products/cr"
	//Create Change Request
	get_response_byid_status, get_response_byid_body := postProductCatalog(create_change_request_url, token, jsonMap)
	fmt.Println("Create Change request response: ", get_response_byid_status)
	fmt.Println("Create Change request body: ", get_response_byid_body)
	return get_response_byid_status, get_response_byid_body
}

func EnableChangeRequest(api_base_url string, token string, product_id string) (int, string) {
	//Enable Change request
	payload := fmt.Sprintf(`{
	  "enable": true
	}`)
	fmt.Println("Enable Change Request Payload: ", payload)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	var create_change_request_url = api_base_url + "/v1/products/cr/enable/" + product_id
	get_response_byid_status, get_response_byid_body := postProductCatalog(create_change_request_url, token, jsonMap)
	fmt.Println("Enable change request response: ", get_response_byid_status)
	fmt.Println("Enable Change request body: ", get_response_byid_body)
	return get_response_byid_status, get_response_byid_body
}

func ApproveChangeRequest(api_base_url string, token string, product_id string) (int, string) {
	//Approve Change request
	payload := fmt.Sprintf(`{
	  "approve": true,
  	  "id": ` + product_id + `
	}`)
	fmt.Println("Approve Change Request Payload: ", payload)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	var create_change_request_url = api_base_url + "/v1/products/cr/review"
	get_response_byid_status, get_response_byid_body := postProductCatalog(create_change_request_url, token, jsonMap)
	fmt.Println("Approve change request response: ", get_response_byid_status)
	fmt.Println("Approve Change request body: ", get_response_byid_body)
	return get_response_byid_status, get_response_byid_body
}

func GetInterests(api_base_url string, token string, cloudaccount_id string, cloudaccount_email string, product_id string, region_id string) (int, string) {
	//Approve Change request
	payload := fmt.Sprintf(`{}`)
	fmt.Println("Approve Change Request Payload: ", payload)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	var create_change_request_url = api_base_url + "/v1/products/interests?cloudaccountId=" + cloudaccount_id + "&userEmail=" + cloudaccount_email + "&productId=" + product_id + "&regionId=" + region_id
	get_response_byid_status, get_response_byid_body := getRegions(create_change_request_url, token)
	fmt.Println("Approve change request response: ", get_response_byid_status)
	fmt.Println("Approve Change request body: ", get_response_byid_body)
	return get_response_byid_status, get_response_byid_body
}

func UpdateFamily(api_base_url string, token string, familyId string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var update_families_url = api_base_url + "/v1/families/" + familyId
	get_response_byid_status, get_response_byid_body := updateFamily(update_families_url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func DeleteFamily(api_base_url string, token string, familyName string) (int, string) {
	var update_families_url = api_base_url + "/v1/families/" + familyName
	get_response_byid_status, get_response_byid_body := deleteFamily(update_families_url, token)
	return get_response_byid_status, get_response_byid_body
}

func UpdateVendor(api_base_url string, token string, vendorId string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var update_vendors_url = api_base_url + "/v1/vendors/" + vendorId
	get_response_byid_status, get_response_byid_body := updateVendor(update_vendors_url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func DeleteVendor(api_base_url string, token string, vendorName string) (int, string) {
	var update_vendors_url = api_base_url + "/v1/vendors/" + vendorName
	get_response_byid_status, get_response_byid_body := deleteVendor(update_vendors_url, token)
	return get_response_byid_status, get_response_byid_body
}

func UpdateRegion(api_base_url string, token string, regionId string, ProductCatalog_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	var update_regions_url = api_base_url + "/v1/regions/" + regionId
	get_response_byid_status, get_response_byid_body := updateRegion(update_regions_url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func DeleteRegion(api_base_url string, token string, regionName string) (int, string) {
	var url = api_base_url + "/v1/regions/" + regionName
	get_response_byid_status, get_response_byid_body := deleteRegion(url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetMetadataSet(api_base_url string, token string) (int, string) {
	var url = api_base_url + "/v1/metadataset"
	get_response_byid_status, get_response_byid_body := getMetadata(url, token)
	return get_response_byid_status, get_response_byid_body
}

func CreateMetadataSet(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var url = api_base_url + "/v1/metadataset"
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	get_response_byid_status, get_response_byid_body := createMetadata(url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func UpdateMetadataSet(api_base_url string, token string, metadataSetId string, ProductCatalog_api_payload string) (int, string) {
	var url = api_base_url + "/v1/metadataset/" + metadataSetId
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	get_response_byid_status, get_response_byid_body := updateMetadata(url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func DeleteMetadataSet(api_base_url string, token string, metadataSetId string) (int, string) {
	var url = api_base_url + "/v1/metadataset/" + metadataSetId
	get_response_byid_status, get_response_byid_body := deleteMetadata(url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetMetadata(api_base_url string, token string) (int, string) {
	var url = api_base_url + "/v1/metadata"
	get_response_byid_status, get_response_byid_body := getMetadata(url, token)
	return get_response_byid_status, get_response_byid_body
}

func CreateMetadata(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var url = api_base_url + "/v1/metadata"
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	get_response_byid_status, get_response_byid_body := createMetadata(url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func UpdateMetadata(api_base_url string, token string, metadataId string, ProductCatalog_api_payload string) (int, string) {
	var url = api_base_url + "/v1/metadata/" + metadataId
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	get_response_byid_status, get_response_byid_body := updateMetadata(url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func DeleteMetadata(api_base_url string, token string, metadataId string) (int, string) {
	var url = api_base_url + "/v1/metadata/" + metadataId
	get_response_byid_status, get_response_byid_body := deleteMetadata(url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetRates(api_base_url string, token string) (int, string) {
	var url = api_base_url + "/v1/rate"
	get_response_byid_status, get_response_byid_body := getRates(url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetRateSets(api_base_url string, token string) (int, string) {
	var url = api_base_url + "/v1/rateset"
	get_response_byid_status, get_response_byid_body := getRates(url, token)
	return get_response_byid_status, get_response_byid_body
}

func CreateServiceType(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var url = api_base_url + "/v1/intelcloudserviceregistration"
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	get_response_byid_status, get_response_byid_body := createServiceType(url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func GetServiceType(api_base_url string, token string) (int, string) {
	var url = api_base_url + "/v1/intelcloudserviceregistration"
	get_response_byid_status, get_response_byid_body := getServiceType(url, token)
	return get_response_byid_status, get_response_byid_body
}

func UpdateServiceType(api_base_url string, token string, name string, location string, ProductCatalog_api_payload string) (int, string) {
	var url = api_base_url + "/v1/intelcloudserviceregistration/" + name + "/" + location
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	get_response_byid_status, get_response_byid_body := updateServiceType(url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func DeleteServiceType(api_base_url string, token string, name string, location string) (int, string) {
	var url = api_base_url + "/v1/intelcloudserviceregistration/" + name + "/" + location
	get_response_byid_status, get_response_byid_body := deleteServiceType(url, token)
	return get_response_byid_status, get_response_byid_body
}

func AddProduct(api_base_url string, token string, ProductCatalog_api_payload string) (int, string) {
	var url = api_base_url + "/v1/products/add"
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	get_response_byid_status, get_response_byid_body := createProduct(url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func UpdateProduct(api_base_url string, token string, ProductCatalog_api_payload string, name string, regionName string) (int, string) {
	var url = api_base_url + "/v1/products/" + name + "/regions/" + regionName
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(ProductCatalog_api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Failed Unmarshalling payload: " + err.Error())
	}
	get_response_byid_status, get_response_byid_body := updateProduct(url, token, jsonMap)
	return get_response_byid_status, get_response_byid_body
}

func DeleteProduct(api_base_url string, token string, name string, regionName string) (int, string) {
	var url = api_base_url + "/v1/products/" + name + "/regions/" + regionName
	get_response_byid_status, get_response_byid_body := deleteProduct(url, token)
	return get_response_byid_status, get_response_byid_body
}

func WhitelistCloudaccount(api_base_url string, token string, cloudaccount_id string, admin_name string, family_id string, createdTime string, product_id string, vendor_id string) (int, string) {
	var jsonMap map[string]interface{}
	whitelist_payload := fmt.Sprintf(`{
		"adminName": "%s",
		"cloudaccountId": "%s",
		"created": "%s",
		"familyId": "%s",
		"productId": "%s",
		"vendorId": "%s"
	}`, admin_name, cloudaccount_id, createdTime, family_id, product_id, vendor_id)
	json.Unmarshal([]byte(whitelist_payload), &jsonMap)
	var whitelist_url = api_base_url + "/v1/products/acl/add"
	response_status, response_body := whitelistCloudaccount(whitelist_url, token, jsonMap)
	return response_status, response_body
}

func Whitelist_Cloud_Account_STaaS(base_url string, admin_token string, cloudaccount_id string, admin_name string, serviceType string) (int, string) {
	familyId := "fabe738e-1edd-4d07-b1b4-5d9eadc9f28d"
	vendorId := "4015bb99-0522-4387-b47e-c821596dc735"
	current_time := time.Now().Add(-1 * time.Hour)
	current_time_d := current_time.Format(time.RFC3339Nano)
	fmt.Println("Creation time to be set: ", current_time_d)

	if serviceType == "ObjectStorageAsAService" {
		productId := "6e9f2eab-76ee-496e-8805-65b5fffca406"
		code, body := WhitelistCloudaccount(base_url, admin_token, cloudaccount_id, admin_name, familyId, current_time_d, productId, vendorId)
		return code, body
	} else {
		productId := "6e9f2eab-76ee-496e-8805-ce9dd17d2d9c"
		code, body := WhitelistCloudaccount(base_url, admin_token, cloudaccount_id, admin_name, familyId, current_time_d, productId, vendorId)
		return code, body
	}
}

func Whitelist_Cloud_Account_MaaS(base_url string, admin_token string, cloudaccount_id string, admin_name string) (int, string) {
	familyId := "4004b07c-14b8-446a-8e63-ed7f1508ee1b"
	vendorId := "4015bb99-0522-4387-b47e-c821596dc735"
	current_time := time.Now().Add(-1 * time.Hour)
	current_time_d := current_time.Format(time.RFC3339Nano)
	fmt.Println("Creation time to be set: ", current_time_d)

	var productIds []string
	productIds = []string{
		"8d728109-0fb2-46c7-a406-1113634d72ab",
		"ba5d2874-dc83-425e-af98-c810f11dad79",
		"269c3034-e6c7-4359-9e77-c3efedfaa778",
		"0a0bffe9-2f62-4ebb-8aee-118ede22b816",
	}

	for _, productId := range productIds {
		code, body := WhitelistCloudaccount(base_url, admin_token, cloudaccount_id, admin_name, familyId, current_time_d, productId, vendorId)
		if code != 200 {
			return code, body
		}
	}

	return 200, "All MaaS products whitelisted successfully"
}

func Whitelist_Cloud_Account_IKS(base_url string, admin_token string, cloudaccount_id string, admin_name string) (int, string) {
	familyId := "61befbee-0607-47c5-b140-c4509dfef835"
	vendorId := "4015bb99-0522-4387-b47e-c821596dc735"
	productId := "3bc52387-da79-4947-a562-95bad32e1db2"
	current_time := time.Now().Add(-1 * time.Hour)
	current_time_d := current_time.Format(time.RFC3339Nano)
	fmt.Println("Creation time to be set: ", current_time_d)
	code, body := WhitelistCloudaccount(base_url, admin_token, cloudaccount_id, admin_name, familyId, current_time_d, productId, vendorId)
	return code, body
}

func Whitelist_Cloud_Account_Gaudi3(base_url string, admin_token string, cloudaccount_id string, admin_name string) (int, string) {
	productId := "f48cbff1-11ab-4c82-b52f-8d07013c98d4"
	vendorId := "4015bb99-0522-4387-b47e-c821596dc735"
	familyId := "61befbee-0607-47c5-b140-c4509dfef835"
	current_time := time.Now().Add(-1 * time.Hour)
	current_time_d := current_time.Format(time.RFC3339Nano)
	fmt.Println("Creation time to be set: ", current_time_d)
	code, body := WhitelistCloudaccount(base_url, admin_token, cloudaccount_id, admin_name, familyId, current_time_d, productId, vendorId)
	return code, body
}

func Whitelist_Cloud_Account_Gaudi3_8node(base_url string, admin_token string, cloudaccount_id string, admin_name string) (int, string) {
	productId := "0a397d4d-930b-41c4-9f04-f8c14b725a12"
	vendorId := "4015bb99-0522-4387-b47e-c821596dc735"
	familyId := "61befbee-0607-47c5-b140-c4509dfef835"
	current_time := time.Now().Add(-1 * time.Hour)
	current_time_d := current_time.Format(time.RFC3339Nano)
	fmt.Println("Creation time to be set: ", current_time_d)
	code, body := WhitelistCloudaccount(base_url, admin_token, cloudaccount_id, admin_name, familyId, current_time_d, productId, vendorId)
	return code, body
}

func Whitelist_Cloud_Account_Gaudi2_32_node(base_url string, admin_token string, cloudaccount_id string, admin_name string) (int, string) {
	productId := "b503fc5b-3ba6-436c-b11d-6ca16af9397e"
	vendorId := "4015bb99-0522-4387-b47e-c821596dc735"
	familyId := "61befbee-0607-47c5-b140-c4509dfef835"
	current_time := time.Now().Add(-1 * time.Hour)
	current_time_d := current_time.Format(time.RFC3339Nano)
	fmt.Println("Creation time to be set: ", current_time_d)
	code, body := WhitelistCloudaccount(base_url, admin_token, cloudaccount_id, admin_name, familyId, current_time_d, productId, vendorId)
	return code, body
}

func Whitelist_Cloud_Account_Gaudi2_single_node(base_url string, admin_token string, cloudaccount_id string, admin_name string) (int, string) {
	productId := "a8482d63-543e-40c5-ab93-804af7bd59fa"
	vendorId := "4015bb99-0522-4387-b47e-c821596dc735"
	familyId := "61befbee-0607-47c5-b140-c4509dfef835"
	current_time := time.Now().Add(-1 * time.Hour)
	current_time_d := current_time.Format(time.RFC3339Nano)
	fmt.Println("Creation time to be set: ", current_time_d)
	code, body := WhitelistCloudaccount(base_url, admin_token, cloudaccount_id, admin_name, familyId, current_time_d, productId, vendorId)
	return code, body
}
