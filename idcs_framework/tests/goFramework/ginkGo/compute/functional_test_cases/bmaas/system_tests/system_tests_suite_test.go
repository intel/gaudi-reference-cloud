package system_tests_test

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

func TestSystemTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "System Tests Suite")
}
