// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"flag"
	"fmt"
	"os"

	auth_admin "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks_commons/auth"

	"gopkg.in/yaml.v2"
)

func main() {

	var admin, user, adminToken, userToken string
	var adminTokenExpiry, userTokenExpiry int64
	var err error

	authConfigPath := "../iks_integration_test/config/auth_config.json"
	tokenOutputPath := "../iks_integration_test/config/config.yaml"

	flag.StringVar(&admin, "a", "", "Invoke Admin Token Generator")
	flag.StringVar(&user, "u", "", "Invoke User Token Generator")
	flag.Parse()

	// parse auth config file
	auth_admin.Get_config_file_data(authConfigPath)

	if admin != "" {
		fmt.Println("Generating Admin Token.")
		adminToken, adminTokenExpiry = auth_admin.Get_Azure_Admin_Bearer_Token(admin)
	} else if user != "" {
		//obtain user token
		fmt.Println("Generating User Token.")
		userToken, userTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(user)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println("Select proper flag \"a\" or \"u\"")
		return
	}

	// Read the YAML file
	yamlFile, err := os.ReadFile(tokenOutputPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Unmarshal the YAML data into a map
	var data map[interface{}]interface{}
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		fmt.Println(err)
		return
	}

	// populate bearer_token and admin_token fields
	if admin != "" {
		data["admin_token"] = string(adminToken)
		data["admin_token_expiry"] = adminTokenExpiry
	} else if user != "" {
		data["bearer_token"] = string(userToken)
		data["bearer_token_expiry"] = userTokenExpiry
	}
	// Marshal the map back into YAML
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Write the YAML data to a file
	err = os.WriteFile(tokenOutputPath, yamlData, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Successfully generated token.")

}
