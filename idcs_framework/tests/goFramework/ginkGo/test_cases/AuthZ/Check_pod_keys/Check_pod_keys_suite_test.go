package Check_pod_keys_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCheckPodKeys(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CheckPodKeys Suite")
}
