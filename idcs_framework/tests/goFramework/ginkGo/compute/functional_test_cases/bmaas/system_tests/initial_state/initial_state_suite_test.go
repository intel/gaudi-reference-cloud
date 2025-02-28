package initial_state_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInitialState(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Initial State Suite")
}
