package utils

import (
	"github.com/tidwall/gjson"
)

func ExtractMachineIPFromResponse(response string) string {
	return gjson.Get(response, "status.interfaces.0.addresses.0").String()
}

func ExtractProxyIPFromResponse(response string) string {
	return gjson.Get(response, "status.sshProxy.proxyAddress").String()
}

func ExtractProxyUserFromResponse(response string) string {
	return gjson.Get(response, "status.sshProxy.proxyUser").String()
}

func ExtractMachineUserFromResponse(response string) string {
	return gjson.Get(response, "status.userName").String()
}

func ExtractInterfaceDetailsFromResponse(response string) (string, string, string, string) {
	var machineAddress = gjson.Get(response, "status.interfaces.0.addresses.0").String()
	var proxyAddress = gjson.Get(response, "status.sshProxy.proxyAddress").String()
	var proxyUser = gjson.Get(response, "status.sshProxy.proxyUser").String()
	var userName = gjson.Get(response, "status.userName").String()
	return machineAddress, proxyAddress, proxyUser, userName
}
