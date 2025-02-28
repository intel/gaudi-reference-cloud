package training

import (
	"fmt"
	"goFramework/framework/frisby_client"
)

func CreateCluster(baseUrl string, token string, cloudAccountId string, payload map[string]interface{}) (int, string) {
	url := fmt.Sprintf("%s/v1/cloudaccounts/%s/clusters", baseUrl, cloudAccountId)
	response := frisby_client.Post(url, token, payload)
	return frisby_client.LogFrisbyInfo(response, "")
}

func GetCluster(baseUrl string, token string, cloudAccountId string, clusterId string) (int, string) {
	url := fmt.Sprintf("%s/v1/cloudaccounts/%s/clusters/%s", baseUrl, cloudAccountId, clusterId)
	response := frisby_client.Get(url, token)
	return frisby_client.LogFrisbyInfo(response, "")
}

func ListClusters(baseUrl string, token string, cloudAccountId string) (int, string) {
	url := fmt.Sprintf("%s/v1/cloudaccounts/%s/clusters", baseUrl, cloudAccountId)
	response := frisby_client.Get(url, token)
	return frisby_client.LogFrisbyInfo(response, "")
}
