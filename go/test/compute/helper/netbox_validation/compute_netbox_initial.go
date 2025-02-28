// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package netbox_validation

import (
	"context"
	"fmt"
	testbmhelper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/compute/helper/bm_helper"
	"strings"

	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/models"
	. "github.com/onsi/gomega"
)

func TestNetboxInitial(ctx context.Context) {
	var devicesList *dcim.DcimDevicesListOK

	testSystems := []string{testSystem1, testSystem2, testSystem3, testSystem4, testSystem5, testSystem6, testSystem7, testSystem8, testSystem9,
		testSystem10, testSystem11, testSystem12}

	var err error
	devicesList, err = testbmhelper.GetDevicesList()
	Expect(err).Error().ShouldNot(HaveOccurred(), "error getting device list form netbox")

	Expect(devicesList).NotTo(BeNil())

	Expect(*devicesList.Payload.Count >= (int64(3))).To(BeTrue())

	for _, testSystem := range testSystems {
		checkDevice(devicesList, testSystem, "Active", "other")
	}

	for i := 0; i < len(devicesList.Payload.Results); i++ {
		bmcInterface, _ := testbmhelper.GetBMCInterfaceForDevice(*devicesList.Payload.Results[i].Name)
		Expect(bmcInterface).NotTo(BeNil())
		Expect(bmcInterface.Label).NotTo(BeEmpty())
	}

}

func checkDevice(list *dcim.DcimDevicesListOK, deviceName string, label string, role string) {
	deviceFound := &models.DeviceWithConfigContext{}
	wasFound := false

	for i := 0; i < len(list.Payload.Results); i++ {
		if strings.Compare(deviceName, *list.Payload.Results[i].Name) == 0 {
			deviceFound = list.Payload.Results[i]
			wasFound = true
			break
		}
	}
	// line below used as inital catch to avoid nil refs, equivalent to 'if wasfalse == false return;' statement
	Expect(wasFound).To(BeTrue(), "Expected device was found")
	Expect(*deviceFound.Name).To(Equal(deviceName), "Device name is not the expected value")
	Expect(*deviceFound.DeviceRole.Name).To(Equal(role), fmt.Sprintf("%s Role is not expected value", deviceName))
	Expect(*deviceFound.Status.Label).To(Equal(label), fmt.Sprintf("%s Label is not expected value", deviceName))
}
