package Instance_replicator_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInstanceReplicator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InstanceReplicator Suite")
}
