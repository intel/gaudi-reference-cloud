package financials

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/frisby_client"
)

// createFileSystem is a helper function that sends a POST request to create a file system.
// It takes in the URL, token, and payload as parameters.
// It returns the response code and response body.
func createFileSystem(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

// getFileSystems is a helper function that sends a GET request to retrieve file systems.
// It takes in the URL and token as parameters.
// It returns the response code and response body.
func getFileSystems(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	return responseCode, responseBody
}

// getFileSystemStatusByResourceId is a helper function that sends a GET request to retrieve the status of a file system by its resource ID.
// It takes in the URL and token as parameters.
// It returns the response code and response body.
func getFileSystemStatusByResourceId(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	return responseCode, responseBody
}

// deleteFileSystembyByResourceId is a helper function that sends a DELETE request to delete a file system by its resource ID.
// It takes in the URL and token as parameters.
// It returns the response code and response body.
func deleteFileSystembyByResourceId(url string, token string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte("{}"), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	frisby_response := frisby_client.Delete_With_Json(url, token, jsonMap)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	return responseCode, responseBody
}

// updateFileSystemByResourceId is a helper function that sends a PUT request to update a file system by its resource ID.
// It takes in the URL, token, and payload as parameters.
// It returns the response code and response body.
func updateFileSystemByResourceId(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	return responseCode, responseBody
}

// getFileSystemCredentialsByResourceId is a helper function that sends a GET request to retrieve the credentials of a file system by its resource ID.
// It takes in the URL and token as parameters.
// It returns the response code and response body.
func getFileSystemCredentialsByResourceId(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	return responseCode, responseBody
}

// getFileSystemStatusByName is a helper function that sends a GET request to retrieve the status of a file system by its name.
// It takes in the URL and token as parameters.
// It returns the response code and response body.
func getFileSystemStatusByName(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	return responseCode, responseBody
}

// deleteFileSystembyByName is a helper function that sends a DELETE request to delete a file system by its name.
// It takes in the URL and token as parameters.
// It returns the response code and response body.
func deleteFileSystembyByName(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	return responseCode, responseBody
}

// updateFileSystemByName is a helper function that sends a PUT request to update a file system by its name.
// It takes in the URL, token, and payload as parameters.
// It returns the response code and response body.
func updateFileSystemByName(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	return responseCode, responseBody
}

// getFileSystemCredentialsByName is a helper function that sends a GET request to retrieve the credentials of a file system by its name.
// It takes in the URL and token as parameters.
// It returns the response code and response body.
func getFileSystemCredentialsByName(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	return responseCode, responseBody
}

// CreateFileSystem is a function that creates a volume using the provided URL, token, payload, and cloud account ID.
// It makes a POST request to the specified URL with the given token and payload.
// It returns the response code and response body.
func CreateFileSystem(base_url string, token string, api_payload string, cloudAccountId string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/filesystems", base_url, cloudAccountId)
	fmt.Println("STaaS", base_url)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := createFileSystem(base_url, token, jsonMap)
	return response_status, response_body
}

func CreateObjectStorage(base_url string, token string, api_payload string, cloudAccountId string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/objects/buckets", base_url, cloudAccountId)
	fmt.Println("base_url", base_url)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := createFileSystem(base_url, token, jsonMap)
	return response_status, response_body
}

// GetFileSystems retrieves volumes information for a given cloud account ID.
// It makes a GET request to the specified base URL and returns the response status and body.
func GetFileSystems(base_url string, token string, cloudAccountId string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/filesystems", base_url, cloudAccountId)
	response_status, response_body := getFileSystems(base_url, token)
	return response_status, response_body
}

// GetFileSystemStatusByResourceId retrieves the status of a file system by its resource ID.
// It takes in the base URL, token, cloud account ID, and resource ID as parameters.
// It returns the response code and response body.
func GetFileSystemStatusByResourceId(base_url string, token string, cloudAccountId string, resourceId string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/filesystems/id/%s", base_url, cloudAccountId, resourceId)
	fmt.Println("base_url", base_url)
	response_status, response_body := getFileSystemStatusByResourceId(base_url, token)
	return response_status, response_body
}

func GetObjectStorageStatusByResourceId(base_url string, token string, cloudAccountId string, resourceId string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/objects/buckets/id/%s", base_url, cloudAccountId, resourceId)
	fmt.Println("base_url", base_url)
	response_status, response_body := getFileSystemStatusByResourceId(base_url, token)
	return response_status, response_body
}

// DeleteFileSystemByResourceId deletes a file system by its resource ID.
// It takes in the base URL, token, cloud account ID, and resource ID as parameters.
// It returns the response code and response body.
func DeleteFileSystemByResourceId(base_url string, token string, cloudAccountId string, resourceId string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/filesystems/id/%s", base_url, cloudAccountId, resourceId)
	response_status, response_body := deleteFileSystembyByResourceId(base_url, token)
	return response_status, response_body
}

// DeleteObjectStorageByResourceId deletes a object storage by its resource ID.
// It takes in the base URL, token, cloud account ID, and resource ID as parameters.
// It returns the response code and response body.
func DeleteObjectStorageByResourceId(base_url string, token string, cloudAccountId string, resourceId string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/objects/buckets/id/%s", base_url, cloudAccountId, resourceId)
	response_status, response_body := deleteFileSystembyByResourceId(base_url, token)
	return response_status, response_body
}

// UpdateFileSystemByResourceId updates a file system by its resource ID.
// It takes in the base URL, token, payload, cloud account ID, and resource ID as parameters.
// It returns the response code and response body.
func UpdateFileSystemByResourceId(base_url string, token string, api_payload string, cloudAccountId string, resourceId string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/filesystems/id/%s", base_url, cloudAccountId, resourceId)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := updateFileSystemByResourceId(base_url, token, jsonMap)
	return response_status, response_body
}

// GetFileSystemCredentialsByResourceId retrieves the credentials of a file system by its resource ID.
// It takes in the base URL, token, cloud account ID, and resource ID as parameters.
// It returns the response code and response body.
func GetFileSystemCredentialsByResourceId(base_url string, token string, cloudAccountId string, resourceId string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/filesystems/id/%s/user", base_url, cloudAccountId, resourceId)
	response_status, response_body := getFileSystemCredentialsByResourceId(base_url, token)
	return response_status, response_body
}

func GetFileSystemCredentialsByName(base_url string, token string, cloudAccountId string, name string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/filesystems/id/%s/user", base_url, cloudAccountId, name)
	response_status, response_body := getFileSystemCredentialsByName(base_url, token)
	return response_status, response_body
}

func GetFileSystemStatusByName(base_url string, token string, cloudAccountId string, name string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/filesystems/id/%s", base_url, cloudAccountId, name)
	response_status, response_body := getFileSystemStatusByName(base_url, token)
	return response_status, response_body
}

func UpdateFileSystemByName(base_url string, token string, api_payload string, cloudAccountId string, name string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/filesystems/id/%s", base_url, cloudAccountId, name)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	response_status, response_body := updateFileSystemByName(base_url, token, jsonMap)
	return response_status, response_body
}

func DeleteFileSystemByName(base_url string, token string, cloudAccountId string, name string) (int, string) {
	base_url = fmt.Sprintf("%s/v1/cloudaccounts/%s/filesystems/id/%s", base_url, cloudAccountId, name)
	response_status, response_body := deleteFileSystembyByName(base_url, token)
	return response_status, response_body
}
