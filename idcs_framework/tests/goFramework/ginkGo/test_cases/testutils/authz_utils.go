package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/tidwall/gjson"
)

type Domain struct {
	Common_Name string `json:"common_name"`
}

func VaultRequest(token string, url_passed string, method string) []byte {
	use_proxy, err := strconv.ParseBool(os.Getenv("USE_PROXY"))
	if err != nil {
		ginkgo.Fail("Error getting use proxy variable" + err.Error())
	}
	if use_proxy {
		proxyURL, _ := url.Parse(os.Getenv("https_proxy"))
		proxy := http.ProxyURL(proxyURL)
		transport := &http.Transport{Proxy: proxy}
		client := &http.Client{Transport: transport}
		req, err := http.NewRequest(method, url_passed, nil)
		if err != nil {
			ginkgo.Fail("url " + url_passed + "failed: " + err.Error())
		}
		req.Header.Set("X-Vault-Token", token)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Print(err)
			ginkgo.Fail(err.Error())
		}
		defer resp.Body.Close()
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			ginkgo.Fail(err.Error())
		}

		return bodyText
	} else {
		client := &http.Client{}
		req, err := http.NewRequest(method, url_passed, nil)
		if err != nil {
			ginkgo.Fail("url " + url_passed + "failed: " + err.Error())
		}
		req.Header.Set("X-Vault-Token", token)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Print(err)
			ginkgo.Fail(err.Error())
		}
		defer resp.Body.Close()
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			ginkgo.Fail(err.Error())
		}

		return bodyText
	}
}

func GenerateCertificate(v_token string, role string, common_name string, url_passed string, method string) bool {

	f := Domain{
		Common_Name: common_name,
	}

	data, err := json.Marshal(f)

	if err != nil {
		ginkgo.Fail(err.Error())
	}
	reader := bytes.NewReader(data)

	use_proxy, err := strconv.ParseBool(os.Getenv("USE_PROXY"))
	if err != nil {
		ginkgo.Fail("Error getting use proxy variable" + err.Error())
	}
	if use_proxy {
		proxyURL, _ := url.Parse(os.Getenv("https_proxy"))
		proxy := http.ProxyURL(proxyURL)
		transport := &http.Transport{Proxy: proxy}
		client := &http.Client{Transport: transport}
		req, err := http.NewRequest(method, url_passed, reader)
		if err != nil {
			ginkgo.Fail(err.Error())
		}
		req.Header.Set("X-Vault-Token", v_token)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			ginkgo.Fail(err.Error())
		}
		defer resp.Body.Close()

		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			ginkgo.Fail(err.Error())
		}

		var result map[string]any
		errors := json.Unmarshal([]byte(bodyText), &result)
		if errors != nil {
			ginkgo.Fail(errors.Error())
		}
	} else {
		if use_proxy {
			client := &http.Client{}
			req, err := http.NewRequest(method, url_passed, reader)
			if err != nil {
				ginkgo.Fail(err.Error())
			}
			req.Header.Set("X-Vault-Token", v_token)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp, err := client.Do(req)
			if err != nil {
				ginkgo.Fail(err.Error())
			}
			defer resp.Body.Close()

			bodyText, err := io.ReadAll(resp.Body)
			if err != nil {
				ginkgo.Fail(err.Error())
			}

			var result map[string]any
			errors := json.Unmarshal([]byte(bodyText), &result)
			if errors != nil {
				ginkgo.Fail(errors.Error())
			}
		}
	}

	return true
}

func GenerateAndGetCertificate(v_token string, role string, common_name string, url_passed string, method string) []byte {
	use_proxy, err := strconv.ParseBool(os.Getenv("USE_PROXY"))
	if err != nil {
		ginkgo.Fail("Error getting use proxy variable" + err.Error())
	}
	if use_proxy {
		var reqData = fmt.Sprintf(`{"common_name": "%s"}`, common_name)

		proxyURL, _ := url.Parse(os.Getenv("https_proxy"))
		proxy := http.ProxyURL(proxyURL)
		transport := &http.Transport{Proxy: proxy}
		client := &http.Client{Transport: transport}
		var data = strings.NewReader(reqData)
		req, err := http.NewRequest(method, url_passed+role, data)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("X-Vault-Token", v_token)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		return bodyText
	} else {
		var reqData = fmt.Sprintf(`{"common_name": "%s"}`, common_name)
		client := &http.Client{}
		var data = strings.NewReader(reqData)
		req, err := http.NewRequest(method, url_passed+role, data)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("X-Vault-Token", v_token)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		return bodyText

	}
}

func ResourceRequest(token string, url_passed string, method string) []byte {
	use_proxy, err := strconv.ParseBool(os.Getenv("USE_PROXY"))
	if err != nil {
		ginkgo.Fail("Error getting use proxy variable" + err.Error())
	}
	if use_proxy {
		proxyURL, _ := url.Parse(os.Getenv("https_proxy"))
		proxy := http.ProxyURL(proxyURL)
		transport := &http.Transport{Proxy: proxy}
		client := &http.Client{Transport: transport}
		req, err := http.NewRequest(method, url_passed, nil)
		if err != nil {
			ginkgo.Fail("url " + url_passed + "failed: " + err.Error())
		}

		req.Header.Add("Authorization", token)
		resp, err := client.Do(req)
		if err != nil {
			ginkgo.Fail(err.Error())
		}
		defer resp.Body.Close()

		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			ginkgo.Fail(err.Error())
		}

		if resp.StatusCode == 500 || resp.StatusCode == 501 || resp.StatusCode == 503 {
			ginkgo.Fail("Error reaching endpoint: " + url_passed + " " + strconv.Itoa(resp.StatusCode) + " Body" + string(bodyText))
		}

		return bodyText
	} else {
		client := &http.Client{}
		req, err := http.NewRequest(method, url_passed, nil)
		if err != nil {
			ginkgo.Fail("url " + url_passed + "failed: " + err.Error())
		}

		req.Header.Add("Authorization", token)
		resp, err := client.Do(req)
		if err != nil {
			ginkgo.Fail(err.Error())
		}
		defer resp.Body.Close()

		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			ginkgo.Fail(err.Error())
		}

		if resp.StatusCode == 500 || resp.StatusCode == 501 || resp.StatusCode == 503 {
			ginkgo.Fail("Error reaching endpoint: " + url_passed + " " + strconv.Itoa(resp.StatusCode) + " Body" + string(bodyText))
		}

		return bodyText
	}
}

func retrieveValuesFromJson(path string) (string, string, string, string, string, string, string) {
	jsonFile, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Print(err.Error())
	}

	globalUrl := gjson.Get(string(jsonFile), "global_url").String()
	regional_url := gjson.Get(string(jsonFile), "regional_url").String()
	vault_addr := gjson.Get(string(jsonFile), "vault_addr").String()
	vault_addr_ca := gjson.Get(string(jsonFile), "vault_addr_ca").String()
	vault_addr_1a := gjson.Get(string(jsonFile), "vault_addr_1a").String()
	use_proxy := gjson.Get(string(jsonFile), "use_proxy").String()
	vault_token := gjson.Get(string(jsonFile), "vault_token").String()

	return globalUrl, regional_url, vault_addr, vault_addr_ca, vault_addr_1a, use_proxy, vault_token
}

func SetEnvironmentVariables() {
	const CONFIG_FILE = "../../../test_config/authz_resources/authz_config.json"

	// Get variables from json
	globalUrl, regional_url, vault_addr, vault_addr_ca, vault_addr_1a, use_proxy, _ := retrieveValuesFromJson(CONFIG_FILE)

	// Set env variables
	os.Setenv("GLOBAL_URL", globalUrl)
	os.Setenv("REGIONAL_URL", regional_url)
	os.Setenv("VAULT_ADDR", vault_addr)
	os.Setenv("VAULT_ADDR_CA", vault_addr_ca)
	os.Setenv("VAULT_ADDR_1A", vault_addr_1a)
	os.Setenv("USE_PROXY", use_proxy)
	//os.Setenv("VAULT_TOKEN", vault_token)
}

func CheckEnvironmentAndGetLocalHost(port uint16) string {
	containerized, err := strconv.ParseBool(os.Getenv("CONTAINERIZED"))
	var address string
	if err != nil {
		ginkgo.Fail("Error getting use proxy variable" + err.Error())
	}
	if containerized {
		address = fmt.Sprintf("host.docker.internal:%d", port)
		return address
	} else {
		address = fmt.Sprintf("localhost:%d", port)
		return address
	}
}
