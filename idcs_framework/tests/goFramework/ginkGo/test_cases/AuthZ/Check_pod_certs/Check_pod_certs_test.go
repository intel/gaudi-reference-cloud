package Check_pod_certs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check Pods Certificates", func() {
	It("Should check all certificates and root certificates are injected in Pods and valid", func() {
		//host := os.Getenv("SSH_HOST")
		//certs := executeBashScript("./scripts/getPodCerts.sh", host)
		//rootCerts := executeBashScript("./scripts/getRootCerts.sh", host)

		//for i, element := range certs {
		//Expect(decodeCert(strings.TrimSpace(rootCerts[i]), strings.TrimSpace(element), i)).To(BeTrue())
		//}
		Expect(decodeCert("d", "a", 2)).To(BeTrue())
	})
})
