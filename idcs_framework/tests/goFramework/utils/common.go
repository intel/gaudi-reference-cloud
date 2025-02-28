package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	auth1 "goFramework/framework/common/auth"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/testify/test_config"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/tidwall/gjson"
)

var jsonData string

var configData string

func Get_config_file_path() (path string) {
	_, filename, _, _ := runtime.Caller(0)
	suite_path := "/" + os.Getenv("Test_Suite") + "/test_config/config.json"
	filePath := filepath.Clean(filepath.Join(filename, "../../")) + suite_path
	return filePath
}

func Get_config_file_data() string {
	configFile := Get_config_file_path()
	// To DO Handle Error
	configData, _ = test_config.LoadConfig(configFile)
	//fmt.Println("Config Data", configData)
	return configData

}

func Get_cluster_type() string {
	return gjson.Get(configData, "cluster_info.name").String()
}

func Get_Bearer_Token() string {
	authConfig := auth1.DefaultConfig
	clientId, clientSecret, tenant, scope, tenantId, realm, username, password, redirect_uri := Get_auth_data_from_config()
	authConfig.ClientID = clientId
	authConfig.ClientSecret = clientSecret
	authConfig.Tenant = tenant
	authConfig.Scope = scope
	authConfig.TenantId = tenantId
	authConfig.Realm = realm
	authConfig.Username = username
	authConfig.Password = password
	authConfig.RedirectUri = redirect_uri
	authCode, err := auth1.GetTokens(authConfig)
	if err != nil {
		logger.Logf.Infof("Fetching access token failed with error, Tests can not continue : %s", err)
		panic(err)
	}

	logger.Logf.Infof("Access Token Generated: %s", authCode)
	//logger.Logf.Infof("Access Token: %s", authCode.RefreshToken)

	// os.Setenv("AUTH_TOKEN", authCode.AccessToken)
	// os.Setenv("REFRESH_TOKEN", authCode.RefreshToken)
	return authCode.AccessToken
}

func Get_Admin_Role_Token() string {
	if _, ok := os.LookupEnv("cloudAccTest"); ok {
		token := os.Getenv("cloudAccToken")
		return token
	}
	configData = Get_config_file_data()
	token := Get_Admin_Token_Type()
	if token == "azure" {
		logger.Logf.Infof("Generating Admin token from Azure")
		adminToken := auth.Get_Azure_Admin_Bearer_Token()
		return adminToken
	}
	logger.Logf.Infof("Generating Admin token from Zitadel")
	var jsonStr string
	url := gjson.Get(configData, "adminRoleToken.url").String()
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Error("JWT Genration Failed with Error " + err.Error())
		panic("Failed to obtain JWT token, Hence Stopping test execution")
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Error("JWT Genration Failed with Error " + err.Error())
		panic("Failed to obtain JWT token, Hence Stopping test execution")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		logger.Log.Error("JWT Genration Failed with Error " + err.Error())
		panic("Failed to obtain JWT token, Hence Stopping test execution")
	}
	// Log the request body
	jsonStr = string(body)
	logger.Log.Info("Admin Bearer Token" + jsonStr)
	return jsonStr
}

func Get_Admin_Token_Type() string {
	tokenType := gjson.Get(configData, "token.method").String()
	return tokenType
}

func Get_UserName(usertype string) string {
	var userName string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userType := gjson.Get(data, "userType").String()
		if userType == usertype {
			userName = gjson.Get(data, "email").String()
			return false
		}

		return true // keep iterating
	})

	return userName
}

func Get_MemberType(usertype string) string {
	var memberType string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userType := gjson.Get(data, "userType").String()
		if userType == usertype {
			memberType = gjson.Get(data, "memberType").String()
			return false
		}

		return true // keep iterating
	})

	return memberType
}

func GetMailSlurpKey() string {
	mailSlurpKey := gjson.Get(configData, "authConfig.mailSlurpKey").String()
	return mailSlurpKey
}

func GetInboxIdEnterprise() string {
	var inboxIdEnterprise string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userType := gjson.Get(data, "userType").String()
		if userType == "Enterprise" {
			inboxIdEnterprise = gjson.Get(data, "inboxIdEnterprise").String()
			return false
		}

		return true // keep iterating
	})

	return inboxIdEnterprise
}

func GetCountryCodeEnterprise() string {
	var countryCode string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userType := gjson.Get(data, "userType").String()
		if userType == "Enterprise" {
			countryCode = gjson.Get(data, "countryCode").String()
			return false
		}

		return true // keep iterating
	})

	return countryCode
}

func GetInboxIdPremium() string {
	var inboxIdPremium string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userType := gjson.Get(data, "userType").String()
		if userType == "Premium" {
			inboxIdPremium = gjson.Get(data, "inboxIdPremium").String()
			return false
		}

		return true // keep iterating
	})

	return inboxIdPremium
}

func GetCountryCodePremium() string {
	var countryCode string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userType := gjson.Get(data, "userType").String()
		if userType == "Premium" {
			countryCode = gjson.Get(data, "countryCode").String()
			return false
		}

		return true // keep iterating
	})

	return countryCode
}

func GetCountryCode(userType1 string) string {
	var countryCode string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userType := gjson.Get(data, "userType").String()
		if userType == userType1 {
			countryCode = gjson.Get(data, "countryCode").String()
			return false
		}

		return true // keep iterating
	})

	return countryCode
}

func GetInboxIdStandard() string {
	var inboxIdStandard string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userType := gjson.Get(data, "userType").String()
		if userType == "Standard" {
			inboxIdStandard = gjson.Get(data, "inboxIdStandard").String()
			return false
		}

		return true // keep iterating
	})

	return inboxIdStandard
}

func GetCountryCodeStandard() string {
	var countryCode string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userType := gjson.Get(data, "userType").String()
		if userType == "Standard" {
			countryCode = gjson.Get(data, "countryCode").String()
			return false
		}

		return true // keep iterating
	})

	return countryCode
}

func Get_Base_Url() string {
	url := gjson.Get(configData, "urls.base_url").String()
	return url
}

func Get_GRPC_Host() string {
	host := gjson.Get(configData, "grpc.host").String()
	return host
}

func Get_Config_Data() string {
	return configData
}

func Get_Vm_Post_Payload(vm_tag string, action string) string {
	path := action + "." + vm_tag
	return gjson.Get(configData, path).String()
}

func Get_Vm_Post_Response(vm_tag string, action string) string {
	path := action + "." + vm_tag + "." + "response_tiny_vm"
	return gjson.Get(configData, path).String()
}

func Get_Vm_Get_Response(vm_tag string) string {
	path := "vmGetResponse" + "." + vm_tag
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	return gjson.Get(configData, path).String()
}

func Get_Vm_Delete_Response(vm_tag string) string {
	path := "vmDeleteResponse" + "." + vm_tag
	return gjson.Get(configData, path).String()
}

//storage utils

func Get_Storage_Post_Payload(storage_tag string, action string) string {

	path := action + "." + storage_tag
	return gjson.Get(configData, path).String()
}

func Get_Storage_Post_Response(storage_tag string, action string) string {

	path := action + "." + storage_tag + "." + "response"
	return gjson.Get(configData, path).String()
}

func Get_Storage_Get_Response(storage_tag string) string {

	path := "storageGetResponse" + "." + storage_tag
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	return gjson.Get(configData, path).String()
}

func Get_Storage_Delete_Response(storage_tag string) string {

	path := "storageDeleteResponse" + "." + storage_tag
	return gjson.Get(configData, path).String()
}

// ssh utils

func Get_Ssh_Post_Payload(tag string, action string) string {

	path := action + "." + tag
	return gjson.Get(configData, path).String()
}

func Get_Ssh_Delete_Response(tag string) string {

	path := "sshDeleteResponse" + "." + tag
	return gjson.Get(configData, path).String()
}

func Get_Ssh_Get_Response(tag string) string {

	path := "sshGetResponse" + "." + tag
	json := gjson.Get(configData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	return gjson.Get(configData, path).String()
}

func CompareJSONToStruct(bytes []byte, empty interface{}) bool {
	var mapped map[string]interface{}

	if err := json.Unmarshal(bytes, &mapped); err != nil {
		return false
	}

	emptyValue := reflect.ValueOf(empty).Type()

	// check if number of fields is the same
	if len(mapped) != emptyValue.NumField() {
		return false
	}

	// check if field names are the same
	for key := range mapped {
		// @todo go deeper into nested struct fields
		if field, found := emptyValue.FieldByName(key); found {
			if !strings.EqualFold(key, strings.Split(field.Tag.Get("json"), ",")[0]) {
				return false
			}
		}
	}

	// @todo check for field type mismatching

	return true
}

func Get_auth_data_from_config() (string, string, string, string, string, string, string, string, string) {
	configData = Get_config_file_data()
	clientId := gjson.Get(configData, "authConfig.client_id").String()
	clientSecret := gjson.Get(configData, "authConfig.client_secret").String()
	tenant := gjson.Get(configData, "authConfig.tenant").String()
	scope := gjson.Get(configData, "authConfig.scope").String()
	tenantId := gjson.Get(configData, "authConfig.tenant_id").String()
	realm := gjson.Get(configData, "authConfig.realm").String()
	username := gjson.Get(configData, "authConfig.username").String()
	password := gjson.Get(configData, "authConfig.password").String()
	redirect_uri := gjson.Get(configData, "authConfig.redirect_uri").String()
	return clientId, clientSecret, tenant, scope, tenantId, realm, username, password, redirect_uri
}

func Get_Expired_Bearer_Token() string {
	configData = Get_config_file_data()
	oldBearerToken := gjson.Get(configData, "expiredBearerToken").String()
	return oldBearerToken
}
