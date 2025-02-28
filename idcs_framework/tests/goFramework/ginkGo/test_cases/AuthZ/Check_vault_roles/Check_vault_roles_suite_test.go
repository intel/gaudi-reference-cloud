package Check_vault_roles_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCheckVaultRoles(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CheckVaultRoles Suite")
}
