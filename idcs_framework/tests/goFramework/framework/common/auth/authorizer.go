package auth

type AuthorizationConfig struct {
	RedirectPort string
	RedirectPath string
	Scope        string
	ClientID     string
	OpenCMD      string
	ClientSecret string
	RedirectUri  string
	Tenant       string
	TenantId     string
	Realm        string
	Username     string
	Password     string
}

var TestConfig AuthorizationConfig
var DefaultConfig = AuthorizationConfig{
	RedirectPort: "3000",
	RedirectPath: "/myapp",
	OpenCMD:      "start",
}

// RedirectURL )
func (c AuthorizationConfig) RedirectURL() string {

	return "https://dev.d3ovv1jda6scbo.amplifyapp.com/"
}
