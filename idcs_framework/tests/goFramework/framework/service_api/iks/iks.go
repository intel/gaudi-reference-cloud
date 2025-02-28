package iks

import (
	"encoding/json"
	"goFramework/framework/frisby_client"
)

func GetVersions(url string, token string) (int, string) {
	response := frisby_client.Get(url, token)
	return frisby_client.LogFrisbyInfo(response, "GET IKS Versions API")
}

func GetInstanceTypes(url string, token string) (int, string) {
	response := frisby_client.Get(url, token)
	return frisby_client.LogFrisbyInfo(response, "GET IKS Instances API")
}

func CreateCluster(url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response := frisby_client.Post(url, token, jsonMap)
	return frisby_client.LogFrisbyInfo(response, "Create IKS Cluster API")
}

func CreateSCCluster(url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response := frisby_client.Post(url, token, jsonMap)
	return frisby_client.LogFrisbyInfo(response, "Create SC Cluster API")
}

func CreateWorkerNode(url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response := frisby_client.Post(url, token, jsonMap)
	return frisby_client.LogFrisbyInfo(response, "Create IKS Worker Node API")
}
