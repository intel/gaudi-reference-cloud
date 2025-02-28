// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing_driver_intel

import "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

type IntelBillingRateService struct {
	pb.UnimplementedBillingRateServiceServer
}

func (svc *IntelBillingRateService) Read(in *pb.BillingRateFilter, outStream pb.BillingRateService_ReadServer) error {
	// TODO: read rates from the productcatalog service
	// Query the cloudaccount service to identify the
	// cloudaccount type to determine which rate to use.
	//
	// This code should be shared with standard driver
	return nil
}
