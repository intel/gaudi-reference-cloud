package main

import (
	"fmt"
	"os"
	"strings"

	old_auth_token "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/auth"
)

func main() {
	if len(os.Args) != 4 {
		//fmt.Println("Usage: program <authConfigPath> <userEmail>")
		os.Exit(1)
	}
	testEnv := os.Args[1]
	authConfigPath := os.Args[2]
	userEmail := os.Args[3]
	if testEnv == "staging" || testEnv == "qa1" || testEnv == "dev3" {
		old_auth_token.Get_config_file_data(authConfigPath)
		userToken, err := old_auth_token.Get_Azure_Bearer_Token(userEmail)
		if err == nil {
			userToken = strings.Split(userToken, " ")[1]
			fmt.Println(userToken)
		} else {
			fmt.Println("")
		}
	}
}
