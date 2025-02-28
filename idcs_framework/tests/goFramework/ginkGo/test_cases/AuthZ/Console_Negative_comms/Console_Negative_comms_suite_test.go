package cloud_console_negative_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConsoleNegativeComms(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConsoleNegativeComms Suite")
}
