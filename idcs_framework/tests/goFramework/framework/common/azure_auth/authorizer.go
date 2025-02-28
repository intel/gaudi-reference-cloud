package azure_auth

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
