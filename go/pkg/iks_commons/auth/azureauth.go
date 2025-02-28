// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks_commons/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"

	"github.com/golang-jwt/jwt"
	cv "github.com/nirasan/go-oauth-pkce-code-verifier"
	"github.com/playwright-community/playwright-go"

	"github.com/tidwall/gjson"
)

var oauthRandom string = fmt.Sprintf("%08d", rand.Intn(100000000))
var timeout float64 = 10

var configData string

func assertErrorToNilf(message string, err error) {
	if err != nil {
		log.Fatalf(message, logkeys.Error, err)
	}
}

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
		fmt.Println("Config file path", filePath)
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

func Get_Azure_Admin_config(testenv string) (string, string, string, string, string, string, string, string, string) {
	username := gjson.Get(configData, getPath(testenv, "adminAuthConfig.username")).String()
	base64Password := gjson.Get(configData, getPath(testenv, "adminAuthConfig.password")).String()
	// decode base64 password
	passwordByte, err := base64.StdEncoding.DecodeString(base64Password)
	if err != nil {
		log.Fatal(logkeys.Error, err)
	}
	password := string(passwordByte)
	clientId := gjson.Get(configData, getPath(testenv, "adminAuthConfig.client_id")).String()
	clientSecret := gjson.Get(configData, getPath(testenv, "adminAuthConfig.clientSecret")).String()
	scope := gjson.Get(configData, getPath(testenv, "adminAuthConfig.scope")).String()
	redirect_uri := gjson.Get(configData, getPath(testenv, "adminAuthConfig.redirectUri")).String()
	redirectPort := gjson.Get(configData, getPath(testenv, "adminAuthConfig.redirectPort")).String()
	authEndPoint := gjson.Get(configData, getPath(testenv, "adminAuthConfig.authEndPoint")).String()
	tokenEndPoint := gjson.Get(configData, getPath(testenv, "adminAuthConfig.tokenEndPoint")).String()
	return clientId, clientSecret, authEndPoint, scope, username, password, redirect_uri, redirectPort, tokenEndPoint
}

func Get_Azure_Admin_Bearer_Token(testenv string) (string, int64) {
	authConfig := AuthorizationConfig{}
	clientId, clientSecret, authEndpoint, scope, username, password, redirect_uri, redirectPort, tokenEndPoint := Get_Azure_Admin_config(testenv)
	authConfig.ClientID = clientId
	authConfig.ClientSecret = clientSecret
	authConfig.AuthorizationEndPoint = authEndpoint
	authConfig.Scope = scope
	authConfig.Username = username
	authConfig.Password = password
	authConfig.RedirectUri = redirect_uri
	authConfig.RedirectPort = redirectPort
	authConfig.TokenEndPoint = tokenEndPoint

	token, expiry := AdminLoginRequest(authConfig)

	token = "Bearer " + token
	return token, expiry
}

func Get_Azure_Bearer_Token(userEmail string) (string, int64, error) {
	authConfig := AuthorizationConfig{}
	clientId, clientSecret, authEndpoint, scope, username, password, redirect_uri, redirectPort, tokenEndPoint := Get_Azure_auth_data_from_config(userEmail)
	if password == "" {
		return "", 0, fmt.Errorf("password is empty, Entry for user not found in config file ")
	}
	authConfig.ClientID = clientId
	authConfig.ClientSecret = clientSecret
	authConfig.AuthorizationEndPoint = authEndpoint
	authConfig.Scope = scope
	authConfig.Username = username
	authConfig.Password = password
	authConfig.RedirectUri = redirect_uri
	authConfig.RedirectPort = redirectPort
	authConfig.TokenEndPoint = tokenEndPoint

	token, expiry := LoginRequest(authConfig)

	token = "Bearer " + token

	return token, expiry, nil
}

func Get_Azure_auth_data_from_config(username string) (string, string, string, string, string, string, string, string, string) {
	var base64Password, password string
	result := gjson.Get(configData, "authConfig.userAccounts")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		userName := gjson.Get(data, "email").String()
		if userName == username {
			base64Password = gjson.Get(data, "password").String()
			// decode base64 password
			passwordByte, err := base64.StdEncoding.DecodeString(base64Password)
			if err != nil {
				log.Fatal(logkeys.Error, err)
			}
			password = string(passwordByte)
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

func verifyIfTokenIsValid(token string) (string, error) {

	if token == "" {
		return "", fmt.Errorf("empty token in cache, need to fetch the latest token")
	}
	// No need to validate the signature since it is an internally generated token.
	token_p, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		logger.Log.Error("Error trying to parse the cached jwt token" + err.Error())
		return "", err
	}
	if err := token_p.Claims.Valid(); err != nil {
		// error is thrown if iat, exp, nbf claims fail validation.
		logger.Log.Error("Claims are not valid" + err.Error())
		return "", err
	}
	return token, nil
}

func LogCheckingK6Environment(message string, print bool) {

	//Adding this to avoid problems when running auth script in k6 environment
	if os.Getenv("K6") != "" && print == true {
		fmt.Println(message)
	}
	if os.Getenv("K6") == "" {
		logger.Log.Info(message)
	}
}

func AdminLoginRequest(c AuthorizationConfig) (token string, expiry int64) {
	URL, err := url.Parse(c.AuthorizationEndPoint)

	if err != nil {
		LogCheckingK6Environment("Error while parsing the auth url"+err.Error(), true)
		return
	}
	var CodeVerifier, _ = cv.CreateCodeVerifier()
	codeChallenge := CodeVerifier.CodeChallengeS256()
	codeVerifier := CodeVerifier.String()
	parameters := url.Values{}
	parameters.Set("code_challenge", codeChallenge)
	parameters.Set("code_challenge_method", "S256")
	parameters.Set("client_id", c.ClientID)
	parameters.Set("scope", c.Scope)
	parameters.Set("redirect_uri", c.RedirectUri)
	parameters.Set("response_type", "code")
	parameters.Set("state", oauthRandom)
	parameters.Set("nonce", oauthRandom)
	parameters.Set("sso_reload", "true")
	parameters.Set("response_mode", "query")
	parameters.Set("x-client-VER", "2.30.0")
	URL.RawQuery = parameters.Encode()

	// Open URL and send username and password
	url := URL.String()
	token, expiry = AuthenticateAdminUser(url, c, codeVerifier)
	if _, err := verifyIfTokenIsValid(token); err != nil {
		LogCheckingK6Environment("Failed to Validate the generated Token", true)
		return
	}

	return token, expiry
}

func AuthenticateAdminUser(url string, c AuthorizationConfig, codeVerifier string) (token string, expiry int64) {
	// err := playwright.Install()
	// if err != nil {
	// 	fmt.Println("Playwrite browser installation failed")
	// }
	pw, err := playwright.Run()

	assertErrorToNilf("could not launch playwright: %w", err)
	browser, err := pw.Chromium.Launch()
	assertErrorToNilf("could not launch Chromium: %w", err)
	context, err := browser.NewContext()
	assertErrorToNilf("could not create context: %w", err)
	page, err := context.NewPage()
	page.SetDefaultNavigationTimeout(timeout)
	assertErrorToNilf("could not create page: %w", err)
	_, err = page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float(0), WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	assertErrorToNilf("could not goto: %w", err)
	// To Do : Test whether page time out is working, comment  sleep accorrdingly
	time.Sleep(10 * time.Second)

	assertErrorToNilf("could not click: %v", page.Click("#i0116"))
	assertErrorToNilf("could not type: %v", page.Type("#i0116", c.Username))
	assertErrorToNilf("could not press: %v", page.Click("#idSIButton9"))
	time.Sleep(10 * time.Second)
	assertErrorToNilf("could not click: %v", page.Click("#i0118"))
	assertErrorToNilf("could not type: %v", page.Type("#i0118", c.Password))
	assertErrorToNilf("could not press: %v", page.Click("#idSIButton9"))

	time.Sleep(10 * time.Second)

	results, _ := page.Evaluate("() => localStorage")

	token, expiry = GetTokenFromLocalStorage(results)
	return token, expiry
}

func AuthenticateUser(url string, c AuthorizationConfig, codeVerifier string) (token string, expiry int64) {
	// err := playwright.Install()
	// if err != nil {
	// 	fmt.Println("Playwrite browser installation failed")
	// }
	pw, err := playwright.Run()
	assertErrorToNilf("could not launch playwright: %w", err)
	browser, err := pw.Chromium.Launch()
	assertErrorToNilf("could not launch Chromium: %w", err)
	context, err := browser.NewContext()
	assertErrorToNilf("could not create context: %w", err)
	page, err := context.NewPage()
	page.SetDefaultNavigationTimeout(timeout)
	assertErrorToNilf("could not create page: %w", err)
	_, err = page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float(0)})
	assertErrorToNilf("could not goto: %w", err)
	// To Do : Test whether page time out is working, comment  sleep accorrdingly
	time.Sleep(10 * time.Second)
	assertErrorToNilf("could not click: %v", page.Click("#signInName"))
	assertErrorToNilf("could not type: %v", page.Type("#signInName", c.Username))
	assertErrorToNilf("could not press: %v", page.Click("#continue"))
	time.Sleep(10 * time.Second)

	fl := strings.Contains(c.Username, "intel.com")
	if fl {
		time.Sleep(20 * time.Second)
		b2b_message := page.Locator("#b2b_message_continue")
		count, _ := b2b_message.Count()
		//assertErrorToNilf("could not determine b2b mesasge: %w", err)
		if count != 0 {
			assertErrorToNilf("could not press: %v", page.Click("#b2b_message_continue"))
			time.Sleep(20 * time.Second)
		}
	}

	if fl {
		time.Sleep(1 * time.Minute)
		assertErrorToNilf("could not click: %v", page.Click("#i0118"))
		assertErrorToNilf("could not type: %v", page.Type("#i0118", c.Password))
		assertErrorToNilf("could not press: %v", page.Click("#idSIButton9"))
		time.Sleep(40 * time.Second)
		if err = page.Click("#KmsiCheckboxField"); err == nil {
			time.Sleep(40 * time.Second)
			page.Click("#idSIButton9")
		}
	} else {
		assertErrorToNilf("could not click: %v", page.Click("#password"))
		assertErrorToNilf("could not type: %v", page.Type("#password", c.Password))
		assertErrorToNilf("could not press: %v", page.Click("#continue"))
	}

	time.Sleep(20 * time.Second)

	results, _ := page.Evaluate("() => localStorage")
	token, expiry = GetTokenFromLocalStorage(results)

	return token, expiry
}

func LoginRequest(c AuthorizationConfig) (token string, expiry int64) {
	URL, err := url.Parse(c.AuthorizationEndPoint)
	if err != nil {
		logger.Log.Info("Error while parsing the auth url" + err.Error())
		return
	}
	var CodeVerifier, _ = cv.CreateCodeVerifier()
	codeChallenge := CodeVerifier.CodeChallengeS256()
	codeVerifier := CodeVerifier.String()
	// LogCheckingK6Environment("code  Verifier"+codeVerifier, false)
	parameters := url.Values{}
	parameters.Set("code_challenge", codeChallenge)
	parameters.Set("code_challenge_method", "S256")
	parameters.Set("client_id", c.ClientID)
	parameters.Set("scope", c.Scope)
	parameters.Set("redirect_uri", c.RedirectUri)
	parameters.Set("response_type", "code")
	parameters.Set("state", oauthRandom)
	parameters.Set("nonce", oauthRandom)
	parameters.Set("response_mode", "query")
	parameters.Set("x-client-VER", "2.37.0")

	URL.RawQuery = parameters.Encode()
	// LogCheckingK6Environment("Authorization Url: "+URL.String(), false)

	// Open URL and send username and password
	url := URL.String()
	token, expiry = AuthenticateUser(url, c, codeVerifier)

	/*if _, err := verifyIfTokenIsValid(token); err != nil {
		LogCheckingK6Environment("Failed to Validate the generated Token", true)
		return
	} */

	return token, expiry
}

func GetTokenFromLocalStorage(results any) (token string, expiry int64) {
	var err error
	v := reflect.ValueOf(results)

	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			if strings.Contains(key.String(), "accesstoken") {
				strct := v.MapIndex(key)

				accessTokenValue := fmt.Sprintf("%v", strct)

				accessTokenByte := []byte(accessTokenValue)
				var accesstokenMap map[string]interface{}
				json.Unmarshal(accessTokenByte, &accesstokenMap)

				token = accesstokenMap["secret"].(string)
				expiry, err = strconv.ParseInt(accesstokenMap["expiresOn"].(string), 10, 64)
				if err != nil {
					logger.Log.Info("Error while convertion string to int64" + err.Error())
					return
				}
			}
		}
	}

	return token, expiry
}

func Get_Azure_Admin_config_Load() (string, string, string, string, string, string, string, string, string) {
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

func Get_Azure_Admin_Bearer_Token_Load_Test() (string, int64) {
	authConfig := AuthorizationConfig{}
	clientId, clientSecret, authEndpoint, scope, username, password, redirect_uri, redirectPort, tokenEndPoint := Get_Azure_Admin_config_Load()
	authConfig.ClientID = clientId
	authConfig.ClientSecret = clientSecret
	authConfig.AuthorizationEndPoint = authEndpoint
	authConfig.Scope = scope
	authConfig.Username = username
	authConfig.Password = password
	authConfig.RedirectUri = redirect_uri
	authConfig.RedirectPort = redirectPort
	authConfig.TokenEndPoint = tokenEndPoint

	token, expiry := AdminLoginRequest(authConfig)

	token = "Bearer " + token
	return token, expiry
}
