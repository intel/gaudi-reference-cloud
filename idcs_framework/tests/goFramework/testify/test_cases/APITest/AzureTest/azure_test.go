//go:build All || Functional || AzureTest || Positive || Regression
// +build All Functional AzureTest Positive Regression

package AzureTest

import (
	"bytes"
	"encoding/json"
	"goFramework/framework/common/http_client"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	//"goFramework/utils"
)

type CreateCloudAccountEnrollStruct struct {
	Premium bool `json:"premium"`
}

// IDC_1.0_CAccSvc_GetById
func (suite *CaAPITestSuite) Test_Auth_Flow() {
	logger.Log.Info("Retrieve the cloud account Azure token")
	auth.Get_config_file_data("../../../test_config/config.json")
	authToken := "Bearer " + auth.Get_Azure_Bearer_Token("testify.stand.2023@proton.me")
	url := "https://internal-placeholder.com/v1/cloudaccounts/enroll"
	data := CreateCloudAccountEnrollStruct{
		Premium: false,
	}
	jsonPayload, _ := json.Marshal(data)
	req := []byte(jsonPayload)
	reqBodyCreate := bytes.NewBuffer(req)
	jsonStr, Code := http_client.Post_Azure(url, reqBodyCreate, authToken, 200)
	logger.Log.Info("Response :" + jsonStr)
	assert.Equal(suite.T(), Code, 200, "Azure Test Failed")
	assert.Equal(suite.T(), gjson.Get(jsonStr, "cloudAccountType").String(), "ACCOUNT_TYPE_STANDARD", "Azure test  failed in validating Clooud Acc Type")
}

// IDC_1.0_CAccSvc_GetById
func (suite *CaAPITestSuite) Test_Refresh_Token() {
	logger.Log.Info("Retrieve the cloud account Azure token")
	url := "https://internal-placeholder.com/v1/cloudaccounts/enroll"
	authToken := "Bearer " + auth.Get_Azure_Bearer_Token("testify.stand.2023@proton.me")
	data := CreateCloudAccountEnrollStruct{
		Premium: false,
	}
	jsonPayload, _ := json.Marshal(data)
	req := []byte(jsonPayload)
	reqBodyCreate := bytes.NewBuffer(req)
	jsonStr, Code := http_client.Post_Azure(url, reqBodyCreate, authToken, 200)
	logger.Log.Info("Response :" + jsonStr)
	assert.Equal(suite.T(), Code, 200, "Azure Test Failed")
	assert.Equal(suite.T(), gjson.Get(jsonStr, "cloudAccountType").String(), "ACCOUNT_TYPE_STANDARD", "Azure test  failed in validating Clooud Acc Type")
}

// IDC_1.0_CAccSvc_GetById
func (suite *CaAPITestSuite) Test_Refresh_Token2() {
	logger.Log.Info("Retrieve the cloud account Azure token")
	url := "https://internal-placeholder.com/v1/cloudaccounts/enroll"
	authToken := "Bearer " + auth.Get_Azure_Bearer_Token("testify.stand.2023@proton.me")
	data := CreateCloudAccountEnrollStruct{
		Premium: false,
	}
	jsonPayload, _ := json.Marshal(data)
	req := []byte(jsonPayload)
	reqBodyCreate := bytes.NewBuffer(req)
	jsonStr, Code := http_client.Post_Azure(url, reqBodyCreate, authToken, 200)
	logger.Log.Info("Response :" + jsonStr)
	assert.Equal(suite.T(), Code, 200, "Azure Test Failed")
	assert.Equal(suite.T(), gjson.Get(jsonStr, "cloudAccountType").String(), "ACCOUNT_TYPE_STANDARD", "Azure test  failed in validating Clooud Acc Type")
}
