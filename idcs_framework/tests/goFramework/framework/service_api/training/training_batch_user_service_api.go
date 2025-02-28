package training

import (
	"fmt"
	"goFramework/framework/frisby_client"
)

func Register(baseUrl string, token string, cloudAccountId string, payload map[string]interface{}) (int, string) {
	url := fmt.Sprintf("%s/v1/cloudaccounts/%s/trainings", baseUrl, cloudAccountId)
	response := frisby_client.Post(url, token, payload)
	return frisby_client.LogFrisbyInfo(response, "")
}

func GetExpiryTimeById(baseUrl string, token string, cloudAccountId string) (int, string) {
	url := fmt.Sprintf("%s/v1/cloudaccounts/%s/trainings/expiry", baseUrl, cloudAccountId)
	response := frisby_client.Get(url, token)
	return frisby_client.LogFrisbyInfo(response, "")
}
