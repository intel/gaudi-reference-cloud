// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

import (
	"context"

	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StandardBillingInvoiceService struct {
	pb.UnimplementedBillingInvoiceServiceServer
	meteringServiceClient *billing.MeteringClient
	productServiceClient  *billing.ProductClient
	cloudAccountClient    *billing.CloudAccountSvcClient
	config                *Config
}

func (svc *StandardBillingInvoiceService) Read(ctx context.Context, in *pb.BillingInvoiceFilter) (*pb.BillingInvoiceResponse, error) {
	// Unimplemented: not required by invoice
	return nil, nil
}

func (svc *StandardBillingInvoiceService) ReadDetail(in *pb.InvoiceId, outStream pb.BillingInvoiceService_ReadDetailServer) error {
	// Unimplemented: not required by invoice
	return nil
}

func (svc *StandardBillingInvoiceService) ReadUnbilled(in *pb.BillingAccount, outStream pb.BillingInvoiceService_ReadUnbilledServer) error {
	// Unimplemented: not required by invoice
	return nil

}

func (svc *StandardBillingInvoiceService) ReadStatement(context.Context, *pb.InvoiceId) (*pb.Statement, error) {
	// TODO: implement statement by querying metering data
	// Code should be shared with standard driver
	return nil, status.Error(codes.Unimplemented, "statements not implemented yet")
}
