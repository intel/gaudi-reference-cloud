package testsetup

import (
	//"bytes"

	"fmt"
	"os"
	//"os/exec"
	"goFramework/framework/common/logger"
	"github.com/tidwall/gjson"
)

var ConfigData string
var TestSetupData string

// Load the configuration from the provided yaml file.
func LoadConfig(filePath string) (string, error) {
	fmt.Println("Config file path", filePath)
	fileData, err := os.ReadFile(filePath) // if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(fileData), nil
}

func Get_config_file_data(configFile string) {
	// To DO Handle Error
	var err error
	ConfigData, err = LoadConfig(configFile)
	if err != nil {
		fmt.Println(err)
	}

}

func Load_Test_Setup_Data(configFile string) {
	// To DO Handle Error
	var err error
	TestSetupData, err = LoadConfig(configFile)
	if err != nil {
		fmt.Println(err)
		TestSetupData = ""
	}

}

// Cloud Accounts

func Get_User_Token(userName string, accType string, eid string) string {
	var idp string
	oidcUrl := gjson.Get(ConfigData, "oidcUrl").String()	
	logger.Logf.Infof("Generating Auth token for user ", userName)	
	tid := Rand_token_payload_gen()
	if accType == "ACCOUNT_TYPE_ENTERPRISE" {
		idp = "https://login.microsoftonline.com/24d2eec2-c04e-4c44-830d-f374a7b9559e/v2.0"
	} else if accType == "ACCOUNT_TYPE_STANDARD" || accType == "ACCOUNT_TYPE_PREMIUM" || accType == "ACCOUNT_TYPE_INTEL" {
		idp = "intelcorpintb2c.onmicrosoft.com"
	}
	logger.Logf.Infof("Get Token for User ", userName)	
	authToken := "Bearer " + Get_OIDC_Enroll_Token(oidcUrl, tid, userName, eid, idp)
	return authToken
}

func EnrollUsers() error {
	result := gjson.Get(ConfigData, "enrollUsers")
	baseUrl := gjson.Get(ConfigData, "baseUrl").String()
	oidcUrl := gjson.Get(ConfigData, "oidcUrl").String()
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userName := gjson.Get(data, "userName").String()
	    accType := gjson.Get(data, "accType").String()
	    eid := gjson.Get(data, "eid").String()
	    tid := Rand_token_payload_gen()
	    var idp string		
	    logger.Logf.Infof("Starting Enrollment for User ", userName)	    
	    if accType == "ACCOUNT_TYPE_ENTERPRISE" {
		    idp = "https://login.microsoftonline.com/24d2eec2-c04e-4c44-830d-f374a7b9559e/v2.0"
	    } else if accType == "ACCOUNT_TYPE_STANDARD" || accType == "ACCOUNT_TYPE_PREMIUM" || accType == "ACCOUNT_TYPE_INTEL" {
		    idp = "intelcorpintb2c.onmicrosoft.com"
	    }
	    logger.Logf.Infof("Starting Enrollment for User ", userName)	   
		authToken := "Bearer " + Get_OIDC_Enroll_Token(oidcUrl, tid, userName, eid, idp)	
		err := EnrollUsersOIDC(data, userName, accType, baseUrl, authToken)
		if err != nil {
			fmt.Println("Enroll Failed")
			//return false
		}
		return true // keep iterating
	})
	return nil
}

func DeleteCloudAccounts() error {
	result := gjson.Get(ConfigData, "enrollUsers")	
	result.ForEach(func(key, value gjson.Result) bool {
		baseUrl := gjson.Get(ConfigData, "baseUrl").String()
		authToken := Get_OIDC_Admin_Token()
		data := value.String()
		userName := gjson.Get(data, "userName").String()
		err := DeleteCloudAccount(userName, baseUrl, authToken)
		if err != nil {
			fmt.Println("Enroll Failed")
			return false
		}
		return true // keep iterating
	})
	return nil
}

// func Enroll_Azure() error {
// 	result := gjson.Get(ConfigData, "enrollUsers")
// 	var premium bool
// 	result.ForEach(func(key, value gjson.Result) bool {
// 		data := value.String()
// 		err := EnrollUsersAzure(data)
// 		if err != nil {
// 			fmt.Println("Hello")
// 		}
// 		return true // keep iterating
// 	})
// 	return nil
// }

// VMaaS utilities

// Instance

func CreateVmInstances() {
	result := gjson.Get(ConfigData, "vmInstance")
	result.ForEach(func(key, value gjson.Result) bool {
		baseUrl := gjson.Get(ConfigData, "baseUrl").String()
		regionalUrl := gjson.Get(ConfigData, "regionalUrl").String()
		authToken := Get_OIDC_Admin_Token()
		data := value.String()
		err := CreateVmInstance(data, regionalUrl, baseUrl, authToken)
		if err != nil {
			fmt.Printf("error creating vm instance %s ", err.Error())
			return false
		}
		return true // keep iterating
	})
}

// Vnet

func CreateVents() {
	result := gjson.Get(ConfigData, "vnet")
	result.ForEach(func(key, value gjson.Result) bool {
		baseUrl := gjson.Get(ConfigData, "baseUrl").String()
		regionalUrl := gjson.Get(ConfigData, "regionalUrl").String()
		authToken := Get_OIDC_Admin_Token()
		data := value.String()
		err := CreateVnet(data, regionalUrl, baseUrl, authToken)
		if err != nil {
			fmt.Printf("error creating vnet %s", err.Error())
			return false
		}
		return true // keep iterating
	})
}

func CreateSSHKeys() {
	result := gjson.Get(ConfigData, "sshKeys")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()		
		baseUrl := gjson.Get(ConfigData, "baseUrl").String()
		regionalUrl := gjson.Get(ConfigData, "regionalUrl").String()
		authToken := Get_OIDC_Admin_Token()
		err := CreateSSHKey(data, regionalUrl, baseUrl, authToken)
		if err != nil {
			fmt.Printf("error creating vnet %s ", err.Error())
			return false
		}
		return true // keep iterating
	})
}

func CreateandRedeemCoupons() {
	result := gjson.Get(ConfigData, "coupons")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		baseUrl := gjson.Get(ConfigData, "baseUrl").String()
		authToken := Get_OIDC_Admin_Token()
		err := CreateandRedeemCoupon(data, baseUrl, authToken)
		if err != nil {
			fmt.Printf("error in create and redeem coupons %s", err.Error())
		}
		return true // keep iterating
	})
}

func DeleteInstances() error {
	result := gjson.Get(ConfigData, "vmInstance")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userName := gjson.Get(data, "userName").String()
		instanceName := gjson.Get(data, "name").String()
		baseUrl := gjson.Get(ConfigData, "baseUrl").String()
		regionalUrl := gjson.Get(ConfigData, "regionalUrl").String()
	    authToken := Get_OIDC_Admin_Token()
		err := DeleteInstance(instanceName, userName, regionalUrl, baseUrl, authToken)
		if err != nil {
			fmt.Println("Enroll Failed")
		}
		return true // keep iterating
	})
	return nil
}

func DeleteVNets() error {
	result := gjson.Get(ConfigData, "vnet")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		baseUrl := gjson.Get(ConfigData, "baseUrl").String()
		regionalUrl := gjson.Get(ConfigData, "regionalUrl").String()
	    authToken := Get_OIDC_Admin_Token()
		userName := gjson.Get(data, "userName").String()
		vNetName := gjson.Get(data, "payload.metadata.name").String()
		err := DeleteVnet(vNetName, userName, regionalUrl, baseUrl, authToken)
		if err != nil {
			fmt.Println("Delete VNet Failed")
			return false
		}
		return true // keep iterating
	})
	return nil
}

func DeleteSSHKeys() error {
	result := gjson.Get(ConfigData, "sshKeys")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		baseUrl := gjson.Get(ConfigData, "baseUrl").String()
		regionalUrl := gjson.Get(ConfigData, "regionalUrl").String()
	    authToken := Get_OIDC_Admin_Token()
		userName := gjson.Get(data, "userName").String()
		sshKeyName := gjson.Get(data, "sshKeyName").String()
		err := DeleteSSHKey(sshKeyName, userName, regionalUrl, baseUrl, authToken)
		if err != nil {
			fmt.Println("Delete SSH KEY Failed")
			return false
		}
		return true // keep iterating
	})
	return nil
}
