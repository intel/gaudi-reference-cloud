package bmaas

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"os"
	"flag"
	"testing"
)

var instanceType string
var sshPublicKey string

func init() {
	os.Setenv("IDC_GLOBAL_URL_PREFIX", "https://dev.api.cloud.intel.com.kind.local")
    os.Setenv("IDC_REGIONAL_URL_PREFIX", "https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local")
    os.Setenv("TOKEN_URL_PREFIX","http://dev.oidc.cloud.intel.com.kind.local:80")
	os.Setenv("https_proxy", "")
	flag.StringVar(&instanceType, "instanceType", "bm-virtual", "")
	flag.StringVar(&sshPublicKey, "sshPublicKey", "", "")
}

func TestBmaas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BM API Suite")
}
