package main

import (
	"fmt"
	"goFramework/framework/authentication"
	"goFramework/framework/library/auth"
	"goFramework/framework/common/logger"
	"os"
)

func main() {
	var admin string
	logger.InitializeZapCustomLogger()
	authentication.Get_config_file_data("./../test_config/prod_catalog_load_auth_config.json")
	auth.Get_config_file_data("./../test_config/prod_catalog_load_auth_config.json")
	username := auth.Get_UserName("Premium")
	if os.Getenv("ADMIN") != "" {
		admin_token := authentication.Get_Azure_Admin_Bearer_Token_Load_Test()
		//admin = fmt.Sprintf("export AZURE_ADMIN_TOKEN='%s'", admin_token)
		admin = fmt.Sprintf(admin_token)
		fmt.Println(admin)
		//test
	} else {
		user_token, _ := authentication.Get_Azure_Bearer_Token(username)
		//user := fmt.Sprintf("export AZURE_TOKEN='%s'", user_token)
		user := fmt.Sprintf(user_token)
		fmt.Println(user)
	}
}
