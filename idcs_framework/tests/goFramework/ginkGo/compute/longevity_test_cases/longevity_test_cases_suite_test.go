package longevity_test_cases_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var instanceType string
var sshPublicKey string

func init() {
	flag.StringVar(&instanceType, "instanceType", "bm-virtual", "")
	flag.StringVar(&sshPublicKey, "sshPublicKey", "", "")
}

func TestLongevityTestCases(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Longevity Tests Cases Suite")
}
