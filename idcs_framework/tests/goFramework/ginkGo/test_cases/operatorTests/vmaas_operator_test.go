package operatorTests

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstanceOperator Test, Create Instance", func() {
	//ctx := context.Background()
	Context("Instance Operator tests, Create Instance - ", func() {
		It("Should create Instance successfully", func() {
			fmt.Fprintf(GinkgoWriter, "Starting Test Create Tiny VM: %v\n", true)
			logger.Log.Info("Starting Test Create Tiny VM")
			ret := library.Create_vm_instance("tiny_vm", 0)
			Expect(ret).Should(Equal(true))
			logger.Log.Info("Starting Test Delete Tiny VM")
			ret = library.Terminate_vm_instance("tiny_vm", 0)
			Expect(ret).Should(Equal(true))
		})

	})

	Context("Instance Operator tests, Create Instance with out ssh Key - ", func() {
		It("Should fail creating instance without ssh key", func() {
			logger.Log.Info("Starting Test Create Tiny VM")
			ret := library.Create_vm_instance("tiny_vm_without_sshPublicKeyNames", 422)
			Expect(ret).Should(Equal(true))
		})

	})

})
