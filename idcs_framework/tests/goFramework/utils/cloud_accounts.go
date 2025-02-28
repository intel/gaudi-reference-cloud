package utils

import (
	//"fmt"
	"goFramework/framework/common/logger"
	"goFramework/testify/test_config"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tidwall/gjson"
)

var caConfigData string

func Get_CA_config_file_path() (path string) {
	_, filename, _, _ := runtime.Caller(0)
	suite_path := "/" + os.Getenv("Test_Suite") + "/test_config/cloud_accounts.json"
	filePath := filepath.Clean(filepath.Join(filename, "../../")) + suite_path
	return filePath
}

func Get_CA_config_file_data() string {
	configFile := Get_CA_config_file_path()
	caConfigData, _ = test_config.LoadConfig(configFile)
	return caConfigData
}

func LoadCAConfig(filepath string) {
	configData, _ := test_config.LoadConfig(filepath)
	url = gjson.Get(configData, "baseUrl").String()
}

func Get_CA_Config_Data() string {
	return caConfigData
}

func Get_CA_Base_Url() string {
	cajsonData := Get_CA_Config_Data()
	url := gjson.Get(cajsonData, "urls.base_url").String()
	return url
}

func Get_ComputeUrl() string {
	cajsonData := Get_CA_Config_Data()
	url := gjson.Get(cajsonData, "urls.computeUrl").String()
	return url
}

func Get_CA_OIDC_Url() string {
	cajsonData := Get_CA_Config_Data()
	url := gjson.Get(cajsonData, "urls.oidc_url").String()
	return url
}

func Get_Token_Type() string {
	cajsonData := Get_CA_Config_Data()
	tokenType := gjson.Get(cajsonData, "token.method").String()
	return tokenType
}

func GenerateInt(n int) string {
	var charset = []rune("0987654321")
	rand.Seed(time.Now().UnixNano())
	str := make([]rune, n)
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}
	return string(str)
}

func Contains(CAcc_type string, CAcc_types1 []string) bool {
	for _, v := range CAcc_types1 {
		if v == CAcc_type {
			return true
		}
	}
	return false
}

func Get_cloudAccounts_Create_Payload(search_tag string) string {
	path := "cloudAccounts.create" + "." + search_tag
	logger.Logf.Info("path is", path)
	json := gjson.Get(caConfigData, path).String()
	if !gjson.Valid(json) {
		logger.Logf.Info("invalid json", json)
	}
	logger.Logf.Info("Json File Is", Get_CA_config_file_path())
	return gjson.Get(caConfigData, path).String()
}

func Get_numAccounts() int64 {
	path := "loadDetails" + "." + "numAccounts"
	logger.Logf.Info("path is", path)
	json := gjson.Get(caConfigData, path).String()
	if !gjson.Valid(json) {
		logger.Logf.Info("invalid json", json)
	}
	logger.Logf.Info("Json File Is", Get_CA_config_file_path())
	numAccounts := gjson.Get(caConfigData, path).Int()
	return numAccounts

}
