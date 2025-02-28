// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package netbox_validation

import (
	"context"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	testbmhelper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/compute/helper/bm_helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

var (
	bmaasRoleID  = int64(1)
	testSystem1  = "device-1"
	testSystem2  = "device-2"
	testSystem3  = "device-3"
	testSystem4  = "device-4"
	testSystem5  = "device-5"
	testSystem6  = "device-6"
	testSystem7  = "device-7"
	testSystem8  = "device-8"
	testSystem9  = "device-9"
	testSystem10 = "device-10"
	testSystem11 = "device-11"
	testSystem12 = "device-12"
)

func TestNetboxEnrollment(ctx context.Context) {

	testSystems := []string{testSystem1, testSystem2, testSystem3, testSystem4, testSystem5, testSystem6, testSystem7, testSystem8,
		testSystem9, testSystem10, testSystem11, testSystem12}

	for _, testSystem := range testSystems {
		time.Sleep(10 * time.Second)
		err := testbmhelper.UpdateDeviceEnrollmentStatus(testSystem, dcim.BMEnroll)
		Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting status")

		err = testbmhelper.UpdateDeviceRole(testSystem, bmaasRoleID)
		Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting Role")
	}

	for _, testSystem := range testSystems {
		By("Wait until testSystem1 is in available state before validation")
		succeded, err := testbmhelper.CheckBMHState(testSystem, "available", 7200)
		Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
			"unable to reach state within expected time")
		Expect(succeded).To(Equal(true))
		By("Wait until testSystem1 has been verified")
		Eventually(func(g Gomega) {
			bmh, err := testbmhelper.GetBmhByName(testSystem)
			Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting Role")
			g.Expect(bmh.Labels).Should(HaveKeyWithValue("cloud.intel.com/verified", "true"))
		}, 25*time.Minute, 1*time.Second).Should(Succeed())
		By("Wait until testSystem1 is in available state")
		succeded, err = testbmhelper.CheckBMHState(testSystem, "available", 7200)
		Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
			"unable to reach state within expected time")
		Expect(succeded).To(Equal(true))
	}

	return
}
