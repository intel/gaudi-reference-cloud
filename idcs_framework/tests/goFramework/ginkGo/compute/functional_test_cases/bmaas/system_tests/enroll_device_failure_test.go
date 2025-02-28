//go:enroll_device_failure || enroll
//go:build enroll_device_failure || enroll
// +build enroll_device_failure enroll

package system_tests_test

import (
	"goFramework/framework/library/bmaas/kube"
	netbox "goFramework/framework/library/bmaas/netbox"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	bmaasRoleId = int64(1)
	testSystem2 = "device-2"
)

var _ = Describe("Cluster Enroll Device Failure", Ordered, Label("large"), func() {

	It("Interface update", func() {
		bmcInterface, _ := netbox.GetBMCInterfaceForDevice(testSystem2)
		Expect(bmcInterface).NotTo(BeNil())
		err := netbox.UpdateInterfaceLabel(bmcInterface, "1234")
		Expect(err).Error().ShouldNot(HaveOccurred(), "error updating interface label")
	})

	It("Update Role", func() {
		err := netbox.UpdateDeviceRole(testSystem2, bmaasRoleId)
		Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting Role")

	})

	It("Update device enrollment status to 'enroll'", func() {
		err := netbox.UpdateDeviceEnrollmentStatus(testSystem2, dcim.BMEnroll)
		Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting status")
	})

	It("Check device state", func() {
		succeded, err := kube.CheckBMHState(testSystem2, "available", 60)
		Expect(err).Error().Should(HaveOccurred(), "Timeout reached; "+
			"unable to reach state within expected time")
		Expect(succeded).To(Equal(false))
	})

})
