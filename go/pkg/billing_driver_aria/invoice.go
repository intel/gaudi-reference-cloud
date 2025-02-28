// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type AriaBillingInvoiceService struct {
	invoiceController *InvoiceController
	pb.UnimplementedBillingInvoiceServiceServer
}

func (svc *AriaBillingInvoiceService) Read(ctx context.Context, in *pb.BillingInvoiceFilter) (*pb.BillingInvoiceResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaBillingInvoiceService.Read").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("get invoice from aria")
	invoiceResponse, err := svc.invoiceController.GetInvoice(ctx, in.CloudAccountId, in.SearchStart, in.SearchEnd)
	if err != nil {
		logger.Error(err, "failed to get invoice")
		return nil, err
	}
	return invoiceResponse, nil
}

func (svc *AriaBillingInvoiceService) ReadDetail(in *pb.InvoiceId, outStream pb.BillingInvoiceService_ReadDetailServer) error {
	// TODO: implement invoices by querying aria
	return nil
}

func (svc *AriaBillingInvoiceService) ReadUnbilled(in *pb.BillingAccount, outStream pb.BillingInvoiceService_ReadUnbilledServer) error {
	// TODO: implement invoices by querying aria
	return nil
}

func (svc *AriaBillingInvoiceService) ReadStatement(ctx context.Context, in *pb.InvoiceId) (*pb.Statement, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaBillingInvoiceService.ReadStatement").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("get statement from aria")
	statement, err := svc.invoiceController.GetStatement(ctx, in.CloudAccountId, in.InvoiceId)
	if err != nil {
		logger.Error(err, "failed to get statement")
		return nil, err
	}
	return statement, nil
}
