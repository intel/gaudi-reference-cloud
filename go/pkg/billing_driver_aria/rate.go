// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

type AriaBillingRateService struct {
	pb.UnimplementedBillingRateServiceServer
}

func (svc *AriaBillingRateService) Read(in *pb.BillingRateFilter, outStream pb.BillingRateService_ReadServer) error {
	// TODO: read rates from aria for the cloudaccount/product
	return nil
}
