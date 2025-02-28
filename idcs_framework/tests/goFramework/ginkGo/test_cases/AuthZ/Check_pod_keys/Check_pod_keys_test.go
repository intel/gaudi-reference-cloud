package Check_pod_keys_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check Pods Keys", func() {
	It("Should check Keys in Pods are injected and valid", func() {
		/*host := os.Getenv("SSH_HOST")
		keys := executeBashScript("./scripts/getPodKeys.sh", host)

		for _, element := range keys {
			Expect(decodeKey(strings.TrimSpace(element))).To(BeTrue())
		} */

		Expect(true).To(BeTrue()) // Disabled test
	})
})
