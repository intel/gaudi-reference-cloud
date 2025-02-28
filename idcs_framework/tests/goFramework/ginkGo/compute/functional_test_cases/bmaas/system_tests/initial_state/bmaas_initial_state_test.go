package initial_state_test

import (
	"fmt"
	"strings"

	kube "goFramework/framework/library/bmaas/kube"
	netbox "goFramework/framework/library/bmaas/netbox"

	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testSystem1 = "device-1"
	testSystem2 = "device-2"
	testSystem3 = "device-3"
)

var _ = Describe("Cluster Initial State", Ordered, Label("large"), func() {
	var devicesList *dcim.DcimDevicesListOK

	BeforeAll(func() {
		var err error
		devicesList, err = netbox.GetDevicesList()
		Expect(err).Error().ShouldNot(HaveOccurred(), "error getting device list form netbox")
	})

	It("Netbox devices should not be nil", func() {
		Expect(devicesList).NotTo(BeNil())
	})

	It("Netbox Devices should be three devices", func() {
		Expect(*devicesList.Payload.Count >= (int64(3))).To(BeTrue())
	})

	DescribeTable("Device status should be active and role other",
		func(deviceName string, label string, role string) {
			deviceFound := &models.DeviceWithConfigContext{}
			wasFound := false

			for i := 0; i < len(devicesList.Payload.Results); i++ {
				if strings.Compare(deviceName, *devicesList.Payload.Results[i].Name) == 0 {
					deviceFound = devicesList.Payload.Results[i]
					wasFound = true
					break
				}
			}
			// line below used as inital catch to avoid nil refs, equivalent to 'if wasfalse == false return;' statement
			Expect(wasFound).To(BeTrue(), "Expected device was found")
			Expect(*deviceFound.Name).To(Equal(deviceName), "Device name is not the expected value")
			Expect(*deviceFound.DeviceRole.Name).To(Equal(role), fmt.Sprintf("%s Role is not expected value", deviceName))
			Expect(*deviceFound.Status.Label).To(Equal(label), fmt.Sprintf("%s Label is not expected value", deviceName))
		},
		Entry(fmt.Sprintf("Device %s", testSystem1), testSystem1, "Active", "other"),
		Entry(fmt.Sprintf("Device %s", testSystem2), testSystem2, "Active", "other"),
		Entry(fmt.Sprintf("Device %s", testSystem3), testSystem3, "Active", "other"),
	)

	It("Device should have a BMC interface", func() {
		for i := 0; i < len(devicesList.Payload.Results); i++ {
			bmcInterface, _ := netbox.GetBMCInterfaceForDevice(*devicesList.Payload.Results[i].Name)
			Expect(bmcInterface).NotTo(BeNil())
			Expect(bmcInterface.Label).NotTo(BeEmpty())
		}
	})

	// "X" means muted as this might change in the future
	XIt("Metal3 Namespace does not yet exist", func() {
		_, err := kube.GetNamespace("metal3")
		Expect(err).Error().Should(HaveOccurred())
		//TODO: Refactor Check why commented line below is not working, needed to split like above
		// Expect(kube.GetNamespace("metal3")).Error().Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("namespaces \"metal3\" not found"))
	})

})
