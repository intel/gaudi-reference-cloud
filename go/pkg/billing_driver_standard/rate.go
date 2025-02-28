// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

import "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

type StandardBillingRateService struct {
	pb.UnimplementedBillingRateServiceServer
}

func (svc *StandardBillingRateService) Read(in *pb.BillingRateFilter, outStream pb.BillingRateService_ReadServer) error {
	// TODO: read rates from the productcatalog service
	// Query the cloudaccount service to identify the
	// cloudaccount type to determine which rate to use.
	// Don't assume the account type is standard, because
	// billing-standard is currently also used for intel
	// accounts.
	return nil
}
