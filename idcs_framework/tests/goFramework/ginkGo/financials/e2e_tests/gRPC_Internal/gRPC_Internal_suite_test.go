package gRPC_Internal_test

import (
	"goFramework/ginkGo/financials/financials_utils"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var cognitoUrl string
var grpcGlobalUrl string
var cognitoClientId string
var cognitoUserPool string
var clientSecret string

var _ = BeforeSuite(func() {
	financials_utils.LoadE2EConfig("../../data", "billing.json")
	cognitoUrl = financials_utils.GetCognitoUrl()
	grpcGlobalUrl = financials_utils.GetgRPCGlobalUrl()
	cognitoClientId = financials_utils.GetCognitoClientId()
	cognitoUserPool = financials_utils.GetcognitoUserPool()
	clientSecret = financials_utils.GetcognitoClientSecret()
})

func TestGRPCInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GRPCInternal Suite")
}
