package Check_pod_certs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCheckPodCerts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CheckPodCerts Suite")
}
