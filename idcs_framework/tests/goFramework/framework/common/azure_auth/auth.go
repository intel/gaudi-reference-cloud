package azure_auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"

	"goFramework/framework/common/logger"
	"strings"

	"github.com/dgrijalva/jwt-go"
	cv "github.com/nirasan/go-oauth-pkce-code-verifier"
	"github.com/pkg/errors"
)

var oauthRandom string = fmt.Sprintf("%08d", rand.Intn(100000000))

// AuthorizationCode is a value provided after initial successful
// authentication/authorization, it is used to get access/refresh tokens
type AuthorizationCode struct {
	Value string
}

// Tokens holds access and refresh tokens
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
}

func verifyIfTokenIsValid(token string) (string, error) {

	if token == "" {
		return "", fmt.Errorf("empty token in cache, need to fetch the latest token")
	}
	// No need to validate the signature since it is an internally generated token.
	token_p, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		logger.Logf.Errorf("Error trying to parse the cached jwt token", err)
		return "", err
	}
	if err := token_p.Claims.Valid(); err != nil {
		// error is thrown if iat, exp, nbf claims fail validation.
		logger.Logf.Errorf("Claims are not valid", err)
		return "", err
	}
	return token, nil
}

func GetAdminTokenFromRefreshToken(c AuthorizationConfig, refreshToken string) (t Tokens, err error) {
	// set the url and form-encoded data for the POST to the access token endpoint
	url := c.TokenEndPoint
	data := fmt.Sprintf(
		"grant_type=refresh_token&client_id=%s"+
			"&refresh_token=%s"+
			"&scope=%s"+
			"&redirect_uri=%s",
		c.ClientID, refreshToken, c.Scope, c.RedirectUri)
	payload := strings.NewReader(data)

	// create the request and execute it
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "http://localhost:3000")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("snap: HTTP error: %s", err)
		return t, err
	}

	// process the response
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	// unmarshal the json into a string map
	err = json.Unmarshal(body, &t)
	if err != nil {
		fmt.Printf("snap: JSON error: %s", err)
		return t, err
	}
	return t, nil
}

func GetTokenFromRefreshToken(c AuthorizationConfig, refreshToken string) (t Tokens, err error) {
	TokenURL := c.TokenEndPoint
	formVals := url.Values{}
	formVals.Set("grant_type", "refresh_token")
	formVals.Set("refresh_token", refreshToken)
	formVals.Set("scope", c.Scope)
	formVals.Set("redirect_uri", c.RedirectUri)
	formVals.Set("client_id", c.ClientID)
	response, err := http.PostForm(TokenURL, formVals)
	if err != nil {
		return t, errors.Wrap(err, "error while trying to get tokens")
	}
	body, err := ioutil.ReadAll(response.Body)
	//logger.Logf.Infof("Token Response Body", string(body))

	if err != nil {
		return t, errors.Wrap(err, "error while trying to read token json body")
	}

	err = json.Unmarshal(body, &t)

	if err != nil {
		return t, errors.Wrap(err, "error while trying to parse token json body")
	}

	return
}

func GetTokens(c AuthorizationConfig, codeVerifier string, code string) (t Tokens, err error) {
	fmt.Println("Retrieving tokens....", c.TokenEndPoint)
	TokenURL := c.TokenEndPoint
	formVals := url.Values{}
	formVals.Set("grant_type", "authorization_code")
	formVals.Set("code", code)
	formVals.Set("scope", c.Scope)
	formVals.Set("redirect_uri", c.RedirectUri)
	formVals.Set("code_verifier", codeVerifier)
	formVals.Set("client_id", c.ClientID)
	response, err := http.PostForm(TokenURL, formVals)
	if err != nil {
		return t, errors.Wrap(err, "error while trying to get tokens")
	}
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return t, errors.Wrap(err, "error while trying to read token json body")
	}

	err = json.Unmarshal(body, &t)
	if err != nil {
		return t, errors.Wrap(err, "error while trying to parse token json body")
	}

	return
}

func GetAdminTokens(c AuthorizationConfig, codeVerifier string, code string) (t Tokens, err error) {
	// set the url and form-encoded data for the POST to the access token endpoint
	url := c.TokenEndPoint
	data := fmt.Sprintf(
		"grant_type=authorization_code&client_id=%s"+
			"&code_verifier=%s"+
			"&code=%s"+
			"&redirect_uri=%s"+
			"&scope=%s",
		c.ClientID, codeVerifier, code, c.RedirectUri, c.Scope)
	payload := strings.NewReader(data)

	// create the request and execute it
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "http://localhost:3000")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("snap: HTTP error: %s", err)
		return t, err
	}

	// process the response
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	// unmarshal the json into a string map
	err = json.Unmarshal(body, &t)
	if err != nil {
		fmt.Printf("snap: JSON error: %s", err)
		return t, err
	}

	return t, nil
}

func AdminLoginRequest(c AuthorizationConfig) (token Tokens) {
	URL, err := url.Parse(c.AuthorizationEndPoint)
	if err != nil {
		logger.Logf.Infof("Error while parsing the auth url", err)
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
	_, token = AuthenticateAdminUser(url, c, codeVerifier)
	if _, err := verifyIfTokenIsValid(token.AccessToken); err != nil {
		logger.Logf.Infof("Failed to Validate the generated Token")
		return
	}

	return
}

// LoginRequest asks the os to open the login URL and starts a listening on the
// configured port for the authorizaton code. This is used on initial login to
// get the initial token pairs
// uses Interactive UI flow
func LoginRequest(c AuthorizationConfig) (token Tokens) {
	URL, err := url.Parse(c.AuthorizationEndPoint)
	if err != nil {
		logger.Logf.Infof("Error while parsing the auth url", err)
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
	parameters.Set("response_mode", "query")
	parameters.Set("x-client-VER", "2.37.0")

	URL.RawQuery = parameters.Encode()

	// Open URL and send username and password
	url := URL.String()
	_, token = AuthenticateUser(url, c, codeVerifier)

	if _, err := verifyIfTokenIsValid(token.AccessToken); err != nil {
		logger.Logf.Infof("Failed to Validate the generated Token")
		return
	}

	return
}
