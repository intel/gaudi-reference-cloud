package auth

import (
	"fmt"
	"os"
)

type AuthorizationConfig struct {
	RedirectPort             string
	RedirectPath             string
	Scope                    string
	ClientID                 string
	OpenCMD                  string
	ClientSecret             string
	RedirectUri              string
	Username                 string
	Password                 string
	AuthorizationEndPoint    string
	TokenEndPoint            string
	GenerateFromRefreshToken bool
}

func LoadConfig(filePath string) (string, error) {
	//Adding this to avoid problems when running auth script in k6 environment
	if os.Getenv("K6") == "" {
		//fmt.Println("Config file path", filePath)
	}
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
	configData, err = LoadConfig(configFile)
	if err != nil {
		fmt.Println(err)
	}
}

func getPath(testenv, subPath string) string {
	return "env." + testenv + "." + subPath
}
