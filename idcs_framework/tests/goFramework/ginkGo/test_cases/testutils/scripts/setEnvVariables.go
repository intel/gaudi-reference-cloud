package main

import (
	"fmt"
	"goFramework/ginkGo/test_cases/testutils"
	"os"
)

func main() {
	testutils.SetEnvironmentVariables()

	fmt.Println("export GLOBAL_URL=" + os.Getenv("GLOBAL_URL"))
	fmt.Println("export REGIONAL_URL=" + os.Getenv("REGIONAL_URL"))
	fmt.Println("export VAULT_ADDR=" + os.Getenv("VAULT_ADDR"))
	fmt.Println("export VAULT_ADDR_CA=" + os.Getenv("VAULT_ADDR_CA"))
	fmt.Println("export VAULT_ADDR_1A=" + os.Getenv("VAULT_ADDR_1A"))
	fmt.Println("export USE_PROXY=" + os.Getenv("USE_PROXY"))
	//fmt.Println("export VAULT_TOKEN=" + os.Getenv("VAULT_TOKEN"))
}
