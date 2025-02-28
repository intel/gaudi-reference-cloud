// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"strings"
	"testing"
	"time"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests"
	clientTestCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"gotest.tools/assert"
)

var Invoiceid int64

func BeforeInvoiceControllerTest(ctx context.Context, test string, t *testing.T) {
	BeforeUsageControllerTest(ctx, test, t)
	Invoiceid = tests.ManageInvoice(ctx, t, cloudAcctId)

}

func TestGetStatement(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	BeforeInvoiceControllerTest(ctx, "TestGetStatement", t)
	invoiceController := NewInvoiceController(clientTestCommon.GetAriaClient(), clientTestCommon.GetAriaCredentials(), AriaService.cloudAccountClient)
	invoiceId := &pb.InvoiceId{
		CloudAccountId: cloudAcctId,
		InvoiceId:      Invoiceid,
	}
	statement, err := invoiceController.GetStatement(ctx, invoiceId.CloudAccountId, invoiceId.InvoiceId)
	if err != nil && !strings.Contains(err.Error(), "3006") {
		t.Fatalf("failed to get statement: %v", err)
	}
	if err == nil || !strings.Contains(err.Error(), "3006") {
		assert.Equal(t, statement.MimeType, "text/html")
	}
}

func TestGetInvoice(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	BeforeInvoiceControllerTest(ctx, "TestGetInvoice", t)
	invoiceController := NewInvoiceController(clientTestCommon.GetAriaClient(), clientTestCommon.GetAriaCredentials(), AriaService.cloudAccountClient)
	invoiceFilter := &pb.BillingInvoiceFilter{
		CloudAccountId: cloudAcctId,
	}
	billingInvoiceResp, err := invoiceController.GetInvoice(ctx, invoiceFilter.CloudAccountId, nil, nil)
	if err != nil {
		t.Fatalf("failed to get invoice: %v", err)
	}
	assert.Equal(t, len(billingInvoiceResp.GetInvoices()), 1)
	currentTime := time.Now()
	start := billingCommon.ToTimestamp(currentTime)
	end := billingCommon.ToTimestamp(currentTime.Add(-30))
	invoiceDateFilter := &pb.BillingInvoiceFilter{
		CloudAccountId: cloudAcctId,
		SearchStart:    start,
		SearchEnd:      end,
	}
	billingInvoiceResp, err = invoiceController.GetInvoice(ctx, invoiceDateFilter.CloudAccountId, invoiceDateFilter.SearchStart, invoiceDateFilter.SearchEnd)
	if err != nil {
		t.Fatalf("failed to get invoice: %v", err)
	}
	assert.Equal(t, len(billingInvoiceResp.GetInvoices()), 1)
}

/*func TestGetInvoiceWithAssignedPlan(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	BeforeInvoiceControllerTest(ctx, "TestGetInvoice", t)
	invoiceController := NewInvoiceController(clientTestCommon.GetAriaClient(), clientTestCommon.GetAriaCredentials(), AriaService.cloudAccountClient)
	invoiceFilter := &pb.BillingInvoiceFilter{
		CloudAccountId: cloudAcctId,
	}
	ReportTestUsage(ctx, t, idcProducts)
	tests.ManageInvoice(ctx, t, cloudAcctId)
	billingInvoiceResp, err := invoiceController.GetInvoice(ctx, invoiceFilter.CloudAccountId, nil, nil)
	if err != nil {
		t.Fatalf("failed to get invoice: %v", err)
	}
	assert.Equal(t, len(billingInvoiceResp.GetInvoices()), 3)
	currentTime := time.Now()
	start := billingCommon.ToTimestamp(currentTime)
	end := billingCommon.ToTimestamp(currentTime.Add(-30))
	invoiceDateFilter := &pb.BillingInvoiceFilter{
		CloudAccountId: cloudAcctId,
		SearchStart:    start,
		SearchEnd:      end,
	}
	billingInvoiceResp, err = invoiceController.GetInvoice(ctx, invoiceDateFilter.CloudAccountId, invoiceDateFilter.SearchStart, invoiceDateFilter.SearchEnd)
	if err != nil {
		t.Fatalf("failed to get invoice: %v", err)
	}
	assert.Equal(t, len(billingInvoiceResp.GetInvoices()), 3)
}*/
