package utils

import (
	"fmt"
	"time"

	//"goFramework/framework/common/logger"
	"goFramework/framework/common/logger"
	"goFramework/testify/test_config"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"

	"github.com/tidwall/gjson"
)

var billingConfigData string

func Get_Billing_config_file_path() (path string) {
	_, filename, _, _ := runtime.Caller(0)
	suite_path := "/" + os.Getenv("Test_Suite") + "/test_config/billing.json"
	filePath := filepath.Clean(filepath.Join(filename, "../../")) + suite_path
	return filePath
}

func Get_Billing_config_file_data() string {
	configFile := Get_Billing_config_file_path()
	// To DO Handle Error
	billingConfigData, _ = test_config.LoadConfig(configFile)
	//fmt.Println("File Data", billingConfigData)
	return billingConfigData
}

func Get_Billing_Base_Url() string {
	url := gjson.Get(billingConfigData, "urls.billingBaseUrl").String()
	return url
}

func Get_Credits_Base_Url() string {
	url := gjson.Get(billingConfigData, "urls.creditsUrl").String()
	return url
}

func Get_Base_Url1() string {
	url := gjson.Get(billingConfigData, "urls.baseUrl").String()
	return url
}

func Get_Compute_Base_Url() string {
	url := gjson.Get(billingConfigData, "urls.computeUrl").String()
	return url
}

func Get_Console_Base_Url() string {
	url := gjson.Get(billingConfigData, "urls.consoleUrl").String()
	return url
}

func Get_CC_payload(cardType string) string {
	path := "creditCards." + cardType
	data := gjson.Get(billingConfigData, path).String()
	return data
}

func Get_CC_Replace_Url() string {
	url := gjson.Get(billingConfigData, "urls.ccReplaceUrl").String()
	return url
}

func Get_CloudEnroll_Base_Url() string {
	url := gjson.Get(billingConfigData, "urls.cloudAccUrl").String()
	return url
}

func Get_Aria_Base_Url() string {
	url := gjson.Get(billingConfigData, "urls.ariaBaseUrl").String()
	return url
}

func Get_CloudAccountUrl() string {
	url := gjson.Get(billingConfigData, "urls.cloudAccCreateUrl").String()
	return url
}

func Get_Billing_GRPC_Host() string {
	host := gjson.Get(billingConfigData, "grpc.host").String()
	return host
}

func Get_Rates(userType string, instanceType string) string {
	path := userType + ".rates." + instanceType
	rate := gjson.Get(billingConfigData, path).String()
	return rate
}

//Billing utils REST

func Get_Billing_Account_Create_Payload(search_tag string) string {
	path := "billing.billingAccount.create" + "." + search_tag
	fmt.Println("path is", path)
	json := gjson.Get(billingConfigData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Json File Is", Get_Billing_config_file_path())
	fmt.Println("Json output", gjson.Get(billingConfigData, path).String())
	return gjson.Get(billingConfigData, path).String()
}

func Get_PC_Sync_Payload(search_tag string) string {
	path := "billing.pcSync" + "." + search_tag
	fmt.Println("path is", path)
	json := gjson.Get(billingConfigData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Json File Is", Get_Billing_config_file_path())
	fmt.Println("Json output", gjson.Get(billingConfigData, path).String())
	return gjson.Get(billingConfigData, path).String()
}
func Get_Cloud_Credit_Create_Payload(search_tag string) string {
	path := "billing.cloudCredit" + "." + search_tag
	fmt.Println("path is", path)
	json := gjson.Get(billingConfigData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Json File Is", Get_Billing_config_file_path())
	fmt.Println("Json output", gjson.Get(billingConfigData, path).String())
	return gjson.Get(billingConfigData, path).String()
}

func Get_Coupon_CreatePayload(search_tag string) string {
	path := "billing.coupons.Create" + "." + search_tag + ".payload"
	fmt.Println("path is", path)
	json := gjson.Get(billingConfigData, path).String()
	if !gjson.Valid(json) {
		fmt.Println("invalid json", json)
	}
	fmt.Println("Json File Is", Get_Billing_config_file_path())
	fmt.Println("Json output", gjson.Get(billingConfigData, path).String())
	return gjson.Get(billingConfigData, path).String()

}

//Aria utils

func Get_Aria_Config() (string, string) {
	clientNo := gjson.Get(billingConfigData, "ariaConfig.client_no").String()
	authKey := gjson.Get(billingConfigData, "ariaConfig.auth_key").String()
	return clientNo, authKey
}

func GenerateString(n int) string {
	var charset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321")
	rand.Seed(time.Now().UnixNano())
	str := make([]rune, n)
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}
	return string(str)
}

func GenerateSSHKeyName(n int) string {
	var charset = []rune("abcdefghijklmnopqrstuvwxyz")
	rand.Seed(time.Now().UnixNano())
	str := make([]rune, n)
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}
	return string(str)
}

// Get setuup data for billing

func Get_ssh_Key_Payload(username string, keyname string) string {
	var data string
	result := gjson.Get(billingConfigData, "sshKeys")
	result.ForEach(func(key, value gjson.Result) bool {
		data = value.String()
		userName := gjson.Get(data, "userName").String()
		if userName == username {
			keyval1 := gjson.Get(data, "sshKeyName").String()
			if keyval1 == keyname {
				return false
			}

		}

		return true // keep iterating
	})
	return data
}

func GetSSHKey() string {
	return gjson.Get(billingConfigData, "sshKey").String()
}

func GetSchedulerTimeout() int {
	logger.Logf.Infof("Scheduler wait time .... %s", gjson.Get(billingConfigData, "schedulerTimeout").String())
	return int(gjson.Get(billingConfigData, "schedulerTimeout").Int())
}

func Get_Vet_Payload(username string, vnetname string) string {
	var data string
	result := gjson.Get(billingConfigData, "vnet")
	result.ForEach(func(key, value gjson.Result) bool {
		data = value.String()
		userName := gjson.Get(data, "userName").String()
		if userName == username {
			keyval := gjson.Get(data, "payload.metadata.name").String()
			if keyval == vnetname {
				return false
			}

		}

		return true // keep iterating
	})
	return data
}

func Get_Instance_Payload(username string, instanceName string) string {
	var data string
	result := gjson.Get(billingConfigData, "vmInstance")
	result.ForEach(func(key, value gjson.Result) bool {
		data = value.String()
		userName := gjson.Get(data, "userName").String()
		if userName == username {
			keyval := gjson.Get(data, "name").String()
			if keyval == instanceName {
				return false
			}

		}

		return true // keep iterating
	})
	return data
}
