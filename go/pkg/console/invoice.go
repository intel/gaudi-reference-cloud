// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package console

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConsoleInvoiceServer struct {
	pb.UnimplementedConsoleInvoiceServiceServer
}

func (cis *ConsoleInvoiceServer) ReadUnbilled(*pb.ConsoleInvoiceUnbilledRequest, pb.
	ConsoleInvoiceService_ReadUnbilledServer) error {
	// use billingInvoiceClient to make gRPC calls to the billing service
	return status.Errorf(codes.Unimplemented, "not implemented")
}
