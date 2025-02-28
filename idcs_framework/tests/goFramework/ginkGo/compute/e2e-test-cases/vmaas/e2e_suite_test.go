package vmaas

import (
	"flag"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var instanceType string
var sshPublicKey string
var proxyIp string
var sshPrivateKeyPath string

func init() {
	os.Setenv("http_proxy", "")
	os.Setenv("https_proxy", "")
	os.Setenv("no_proxy", "")
	flag.StringVar(&instanceType, "instanceType", "vm-spr-tny", "")
	flag.StringVar(&sshPublicKey, "sshPublicKey", "", "")
	flag.StringVar(&proxyIp, "proxyIp", "10.165.62.252", "")
	flag.StringVar(&sshPrivateKeyPath, "sshPrivateKeyPath", "../../ansible/id-rsa.pub", "")
}

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VM Instance Suite")
}

type GetProductsResponse struct {
	Products []struct {
		Name        string    `json:"name"`
		ID          string    `json:"id"`
		Created     time.Time `json:"created"`
		VendorID    string    `json:"vendorId"`
		FamilyID    string    `json:"familyId"`
		Description string    `json:"description"`
		Metadata    struct {
			Category     string `json:"category"`
			Disks        string `json:"disks.size"`
			DisplayName  string `json:"displayName"`
			Desc         string `json:"family.displayDescription"`
			DispName     string `json:"family.displayName"`
			Highlight    string `json:"highlight"`
			Information  string `json:"information"`
			InstanceType string `json:"instanceType"`
			Memory       string `json:"memory.size"`
			Processor    string `json:"processor"`
			Region       string `json:"region"`
			Service      string `json:"service"`
		} `json:"metadata"`
		Eccn      string `json:"eccn"`
		Pcq       string `json:"pcq"`
		MatchExpr string `json:"matchExpr"`
		Rates     []struct {
			AccountType string `json:"accountType"`
			Rate        string `json:"rate"`
			Unit        string `json:"unit"`
			UsageExpr   string `json:"usageExpr"`
		} `json:"rates"`
	} `json:"products"`
}
