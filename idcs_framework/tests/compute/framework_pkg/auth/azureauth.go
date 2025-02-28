package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	cv "github.com/nirasan/go-oauth-pkce-code-verifier"
	"github.com/playwright-community/playwright-go"

	"github.com/tidwall/gjson"
)

var oauthRandom string = fmt.Sprintf("%08d", rand.Intn(100000000))
var timeout float64 = 30

var configData string
var logInstance *logger.CustomLogger

func SetLogger(logger *logger.CustomLogger) {
	logInstance = logger
}

func assertErrorToNilf(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func Get_Azure_Admin_config(testenv string) (string, string, string, string, string, string, string, string, string) {
	username := gjson.Get(configData, getPath(testenv, "adminAuthConfig.username")).String()
	password := gjson.Get(configData, getPath(testenv, "adminAuthConfig.password")).String()
	clientId := gjson.Get(configData, getPath(testenv, "adminAuthConfig.client_id")).String()
	clientSecret := gjson.Get(configData, getPath(testenv, "adminAuthConfig.clientSecret")).String()
	scope := gjson.Get(configData, getPath(testenv, "adminAuthConfig.scope")).String()
	redirect_uri := gjson.Get(configData, getPath(testenv, "adminAuthConfig.redirectUri")).String()
	redirectPort := gjson.Get(configData, getPath(testenv, "adminAuthConfig.redirectPort")).String()
	authEndPoint := gjson.Get(configData, getPath(testenv, "adminAuthConfig.authEndPoint")).String()
	tokenEndPoint := gjson.Get(configData, getPath(testenv, "adminAuthConfig.tokenEndPoint")).String()
	return clientId, clientSecret, authEndPoint, scope, username, password, redirect_uri, redirectPort, tokenEndPoint
}

func Get_Azure_Admin_Bearer_Token(testenv string) string {
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

	token := AdminLoginRequest(authConfig)

	token = "Bearer " + token
	return token
}

func Get_Azure_Bearer_Token(userEmail string) (string, error) {
	authConfig := AuthorizationConfig{}
	clientId, clientSecret, authEndpoint, scope, username, password, redirect_uri, redirectPort, tokenEndPoint := Get_Azure_auth_data_from_config(userEmail)
	if password == "" {
		return "", fmt.Errorf("password is empty, Entry for user not found in config file ")
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

	token := LoginRequest(authConfig)

	token = "Bearer " + token

	return token, nil
}

func AuthenticateUser(url string, c AuthorizationConfig, codeVerifier string) (token string) {
	err := playwright.Install()
	if err != nil {
		logInstance.Println("Playwrite browser installation failed")
	}
	pw, err := playwright.Run()
	assertErrorToNilf("could not launch playwright: %w", err)
	browser, err := pw.Chromium.Launch()
	assertErrorToNilf("could not launch Chromium: %w", err)
	context, err := browser.NewContext()
	assertErrorToNilf("could not create context: %w", err)
	page, err := context.NewPage()
	page.SetDefaultNavigationTimeout(timeout)
	assertErrorToNilf("could not create page: %w", err)
	for i := 0; i < 5; i++ {
		_, err = page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float(0)})
		if err == nil {
			break
		}
		continue
	}
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

	token = GetTokenFromLocalStorage(results)

	return token
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

func verifyIfTokenIsValid(token string) (string, error) {

	if token == "" {
		return "", fmt.Errorf("empty token in cache, need to fetch the latest token")
	}
	// No need to validate the signature since it is an internally generated token.
	token_p, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		logInstance.Println("Error trying to parse the cached jwt token" + err.Error())
		return "", err
	}
	if err := token_p.Claims.Valid(); err != nil {
		// error is thrown if iat, exp, nbf claims fail validation.
		logInstance.Println("Claims are not valid" + err.Error())
		return "", err
	}
	return token, nil
}

func LogCheckingK6Environment(message string, print bool) {

	//Adding this to avoid problems when running auth script in k6 environment
	if os.Getenv("K6") != "" && print == true {
		logInstance.Println(message)
	}
	if os.Getenv("K6") == "" {
		logInstance.Println(message)
	}
}

func AdminLoginRequest(c AuthorizationConfig) (token string) {
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
	token = AuthenticateAdminUser(url, c, codeVerifier)
	if _, err := verifyIfTokenIsValid(token); err != nil {
		LogCheckingK6Environment("Failed to Validate the generated Token", true)
		return
	}

	return token
}

// Update on Sept 25, 2024 for modified UI flow
func AuthenticateAdminUser(url string, c AuthorizationConfig, codeVerifier string) (token string) {
	err := playwright.Install()
	if err != nil {
		logInstance.Println("Playwrite browser installation failed")
	}
	pw, err := playwright.Run()

	assertErrorToNilf("could not launch playwright: %w", err)
	browser, err := pw.Chromium.Launch()
	assertErrorToNilf("could not launch Chromium: %w", err)
	context, err := browser.NewContext()
	assertErrorToNilf("could not create context: %w", err)
	page, err := context.NewPage()
	page.SetDefaultNavigationTimeout(timeout)
	assertErrorToNilf("could not create page: %w", err)
	for i := 0; i < 5; i++ {
		_, err = page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float(0)})
		if err == nil {
			break
		}
		continue
	}
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

	token = GetTokenFromLocalStorage(results)

	return token
}

func LoginRequest(c AuthorizationConfig) (token string) {
	URL, err := url.Parse(c.AuthorizationEndPoint)
	if err != nil {
		logInstance.Println("Error while parsing the auth url" + err.Error())
		return
	}
	var CodeVerifier, _ = cv.CreateCodeVerifier()
	codeChallenge := CodeVerifier.CodeChallengeS256()
	codeVerifier := CodeVerifier.String()
	//LogCheckingK6Environment("code  Verifier"+codeVerifier, false)
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
	//LogCheckingK6Environment("Authorization Url: "+URL.String(), false)

	// Open URL and send username and password
	url := URL.String()
	token = AuthenticateUser(url, c, codeVerifier)

	/*if _, err := verifyIfTokenIsValid(token); err != nil {
		LogCheckingK6Environment("Failed to Validate the generated Token", true)
		return
	} */

	return token
}

func GetTokenFromLocalStorage(results any) (token string) {
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
			}
		}
	}

	return token
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

func Get_Azure_Admin_Bearer_Token_Load_Test() string {
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

	token := AdminLoginRequest(authConfig)

	token = "Bearer " + token
	return token
}
