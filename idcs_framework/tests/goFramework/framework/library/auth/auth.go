package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"goFramework/framework/common/azure_auth"
	"goFramework/framework/common/logger"
	"log"
	"os"
	"strconv"

	"github.com/phayes/freeport"
	"github.com/tidwall/gjson"
)

var jsonData string

var tokens map[string]string

var configData string

// Load the configuration from the provided yaml file.
func LoadConfig(filePath string) (string, error) {
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

func Get_Refresh_Token(username string) string {
	if _, ok := tokens[username]; !ok {
		logger.Log.Info("Entry not found in tokens map, Hence token will be created using authorization flow for the user " + username)
		return ""
	}
	logger.Log.Info("Refresh Token found in token map, hence token will be generated using refresh token for the user " + username)
	return tokens[username]
}

func Get_New_Access_Token(authConfig azure_auth.AuthorizationConfig) string {
	authCode := azure_auth.LoginRequest(authConfig)
	return authCode.AccessToken
}

func Get_Azure_Bearer_Token(userEmail string) (string, error) {
	authConfig := azure_auth.AuthorizationConfig{}
	clientId, clientSecret, authEndpoint, scope, username, password, redirect_uri, redirectPort, tokenEndPoint := Get_Azure_auth_data_from_config(userEmail)
	if password == "" {
		return "", fmt.Errorf("password is empty, Entry for user not found in config file ")
	}

	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatalf("Failed to get a free port: %v", err)
	}

	authConfig.ClientID = clientId
	authConfig.ClientSecret = clientSecret
	authConfig.AuthorizationEndPoint = authEndpoint
	authConfig.Scope = scope
	authConfig.Username = username
	authConfig.Password = password
	authConfig.RedirectUri = "http://localhost:" + strconv.Itoa(port) + "/"
	authConfig.RedirectPort = strconv.Itoa(port)
	authConfig.TokenEndPoint = tokenEndPoint

	fmt.Println("Old Port: ", redirectPort)
	fmt.Println("Old Redirect: ", redirect_uri)
	fmt.Println("New Port: ", authConfig.RedirectPort)
	fmt.Println("New Redirect: ", authConfig.RedirectUri)
	// First Check refresh token is available for user if there then use that to fetch access token
	refreshToken := Get_Refresh_Token(userEmail)
	if refreshToken != "" {
		//Fetch token from refresh token and return it
		authCode, err := azure_auth.GetTokenFromRefreshToken(authConfig, refreshToken)
		if err != nil {
			logger.Log.Info("Empty Auth token returned with error " + err.Error())
			return "", err
		}
		logger.Log.Info("Auth token generated using RefreshToken for user : " + username)
		return authCode.AccessToken, nil

	}
	authCode := azure_auth.LoginRequest(authConfig)
	// Store the refresh token to token file to fetch and reuse  it to get access token
	if tokens == nil {
		tokens = make(map[string]string)
	}
	tokens[username] = authCode.RefreshToken
	logger.Logf.Infof("Storing token data in map for user " + username)
	return authCode.AccessToken, nil
}

func Get_Azure_auth_data_from_config(username string) (string, string, string, string, string, string, string, string, string) {
	var password string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userName := gjson.Get(data, "email").String()
		if userName == username {
			password = gjson.Get(data, "password").String()
			return false
		}

		return true // keep iterating
	})
	clientId := gjson.Get(configData, "authConfig.client_id").String()
	clientSecret := gjson.Get(configData, "authConfig.clientSecret").String()
	scope := gjson.Get(configData, "authConfig.scope").String()
	redirect_uri := gjson.Get(configData, "authConfig.redirectUri").String()
	redirectPort := gjson.Get(configData, "authConfig.redirectPort").String()
	authEndPoint := gjson.Get(configData, "authConfig.authEndPoint").String()
	tokenEndPoint := gjson.Get(configData, "authConfig.tokenEndPoint").String()
	return clientId, clientSecret, authEndPoint, scope, username, password, redirect_uri, redirectPort, tokenEndPoint
}

// Get admin token

func Get_Azure_Admin_config() (string, string, string, string, string, string, string, string, string) {
	username := gjson.Get(configData, "adminAuthConfig.useranme").String()
	password := gjson.Get(configData, "adminAuthConfig.password").String()
	clientId := gjson.Get(configData, "adminAuthConfig.client_id").String()
	clientSecret := gjson.Get(configData, "adminAuthConfig.clientSecret").String()
	scope := gjson.Get(configData, "adminAuthConfig.scope").String()
	redirect_uri := gjson.Get(configData, "adminAuthConfig.redirectUri").String()
	redirectPort := gjson.Get(configData, "adminAuthConfig.redirectPort").String()
	authEndPoint := gjson.Get(configData, "adminAuthConfig.authEndPoint").String()
	tokenEndPoint := gjson.Get(configData, "adminAuthConfig.tokenEndPoint").String()
	return clientId, clientSecret, authEndPoint, scope, username, password, redirect_uri, redirectPort, tokenEndPoint
}

func Get_Azure_Admin_Bearer_Token() string {
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatalf("Failed to get a free port: %v", err)
	}
	authConfig := azure_auth.AuthorizationConfig{}
	clientId, clientSecret, authEndpoint, scope, username, password, redirect_uri, redirectPort, tokenEndPoint := Get_Azure_Admin_config()
	authConfig.ClientID = clientId
	authConfig.ClientSecret = clientSecret
	authConfig.AuthorizationEndPoint = authEndpoint
	authConfig.Scope = scope
	authConfig.Username = username
	authConfig.Password = password
	authConfig.RedirectUri = "http://localhost:" + strconv.Itoa(port) + "/"
	authConfig.RedirectPort = strconv.Itoa(port)
	authConfig.TokenEndPoint = tokenEndPoint
	fmt.Println("Admin auth flow", redirectPort)
	fmt.Println("Old Port: ", redirectPort)
	fmt.Println("Old Redirect: ", redirect_uri)
	fmt.Println("New Port: ", authConfig.RedirectPort)
	fmt.Println("New Redirect: ", authConfig.RedirectUri)
	// First Check refresh token is available for user if there then use that to fetch access token
	refreshToken := Get_Refresh_Token(username)
	if refreshToken != "" {
		//Fetch token from refresh token and return it
		authCode, err := azure_auth.GetAdminTokenFromRefreshToken(authConfig, refreshToken)
		if err != nil {
			logger.Log.Info("Empty Auth token returned with error " + err.Error())
			return ""
		}
		//logger.Log.Info("Auth token generated from refresh token : " + authCode.AccessToken)
		return authCode.AccessToken

	}
	authCode := azure_auth.AdminLoginRequest(authConfig)
	// Store the refresh token to token file to fetch and reuse  it to get access token
	if tokens == nil {
		tokens = make(map[string]string)
	}
	tokens[username] = authCode.RefreshToken
	logger.Logf.Infof("Storing token data in map for user " + username)
	return authCode.AccessToken
}

func Get_UserName(usertype string) string {
	var userName string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userType := gjson.Get(data, "userType").String()
		if userType == usertype {
			userName = gjson.Get(data, "email").String()
			return false
		}

		return true // keep iterating
	})

	return userName
}

func generateRandomString() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("Failed to generate random string: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b)
}
