//go:enroll_device || enroll
//go:build enroll_device || enroll
// +build enroll_device enroll

package system_tests_test

import (
	"goFramework/framework/library/bmaas/kube"
	nb "goFramework/framework/library/bmaas/netbox"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	otherRoleID = int64(2)
	bmaasRoleID = int64(1)
	testSystem1 = "device-1"
	testSystem3 = "device-3"
)

var _ = Describe("Cluster Enroll Device", Ordered, Label("large"), func() {

	It("Update device enrollment status to 'enroll'", func() {
		err := nb.UpdateDeviceEnrollmentStatus(testSystem1, dcim.BMEnroll)
		Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting status")
	})

	It("Update Role", func() {
		err := nb.UpdateDeviceRole(testSystem1, bmaasRoleID)
		Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting Role")
	})

	It("Update device enrollment status to 'enroll'", func() {
		err := nb.UpdateDeviceEnrollmentStatus(testSystem3, dcim.BMEnroll)
		Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting status")
	})

	It("Update Role", func() {
		err := nb.UpdateDeviceRole(testSystem3, bmaasRoleID)
		Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting Role")

	})

	It("Check system state (kube)", func() {
		By("Wait until testSystem1 is in available state before validation")
		succeded, err := kube.CheckBMHState(testSystem1, "available", 1800)
		Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
			"unable to reach state within expected time")
		Expect(succeded).To(Equal(true))
		By("Wait until testSystem1 has been verified")
		Eventually(func(g Gomega) {
			bmh, err := kube.GetBmhByName(testSystem1)
			Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting Role")
			g.Expect(bmh.Labels).Should(HaveKeyWithValue("cloud.intel.com/verified", "true"))
		}, 25*time.Minute, 1*time.Second).Should(Succeed())
		By("Wait until testSystem1 is in available state")
		succeded, err = kube.CheckBMHState(testSystem1, "available", 1800)
		Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
			"unable to reach state within expected time")
		Expect(succeded).To(Equal(true))

		By("Wait until testSystem3 has been verified")
		Eventually(func(g Gomega) {
			bmh, err := kube.GetBmhByName(testSystem3)
			Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting Role")
			g.Expect(bmh.Labels).Should(HaveKeyWithValue("cloud.intel.com/verified", "true"))
		}, 25*time.Minute, 1*time.Second).Should(Succeed())
		By("Wait until testSystem3 is in available state")
		succeded, err = kube.CheckBMHState(testSystem3, "available", 1800)
		Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
			"unable to reach state within expected time")
		Expect(succeded).To(Equal(true))
	})

	It("Check device enrollment status is 'enrolled'", func() {
		succeded, err := nb.CheckDeviceEnrollmentStatus(testSystem1, dcim.BMEnrolled, 30*time.Minute)
		Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
			"unable to reach state within expected time")
		Expect(succeded).To(Equal(true))
	})

})
