package bmaas

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	"os"
	"goFramework/framework/common/logger"
	"flag"
	"time"
	"testing"
)

var instanceType string
var sshPublicKey string

func init() {
	flag.StringVar(&instanceType, "instanceType", "bm-spr", "")
	flag.StringVar(&sshPublicKey, "sshPublicKey", "", "")
}

func TestBmaas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BM Instance Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
	logger.Log.Info("Starting test suite")
})

type GetProductsResponse struct {
	Products []struct {
		Name  string `json:"name"`
		ID          string    `json:"id"`
		Created     time.Time `json:"created"`
		VendorID string    `json:"vendorId"`
		FamilyID    string    `json:"familyId"`
		Description string `json:"description"`
		Metadata    struct {
			Category string `json:"category"`
			Disks string `json:"disks.size"`
			DisplayName string `json:"displayName"`
			Desc string `json:"family.displayDescription"`
			DispName string `json:"family.displayName"`
			Highlight string `json:"highlight"`
			Information string `json:"information"`
			InstanceType string `json:"instanceType"`
			Memory string `json:"memory.size"`
			Processor string `json:"processor"`
			Region string `json:"region"`
			Service string `json:"service"`
		} `json:"metadata"`
		Eccn        string    `json:"eccn"`
		Pcq   string `json:"pcq"`
		MatchExpr   string    `json:"matchExpr"`
		Rates []struct {
			AccountType string `json:"accountType"`
			Rate        string `json:"rate"`
			Unit        string `json:"unit"`
			UsageExpr   string `json:"usageExpr"`
		} `json:"rates"`
	} `json:"products"`
}
