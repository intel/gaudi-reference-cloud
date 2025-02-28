// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAriaApiTestsFlow(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aria Systems API Client Suite")
}

func ShouldRunTest(test string) bool {
	// TODO:condition to determine if the specific test should be run
	switch test {
	case "account":
		return true
	case "invoice":
		return true
	case "enterprise":
		return true
	}
	return false
}

var _ = BeforeSuite(func() {
	usageTypeClient := common.GetAriaUsageTypeClient()
	testUsagesTypeCode := client.GetMinsUsageTypeCode()
	ctx := context.Background()
	_, err := usageTypeClient.GetUsageTypeDetails(ctx, testUsagesTypeCode)
	if err != nil && strings.Contains(err.Error(), "error code:1010") {
		usageUnitTypes, err := usageTypeClient.GetMinuteUsageUnitType(ctx)
		Expect(err).NotTo(HaveOccurred())
		_, err = usageTypeClient.CreateUsageType(ctx, client.USAGE_TYPE_NAME, client.USAGE_TYPE_DESC, usageUnitTypes.UsageUnitTypeNo, testUsagesTypeCode)
		Expect(err).NotTo(HaveOccurred())
	}
})
