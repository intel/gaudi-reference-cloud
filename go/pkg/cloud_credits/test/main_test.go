// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var creditsClient pb.CloudCreditsCreditServiceClient
var couponClient pb.CloudCreditsCouponServiceClient

func TestMain(m *testing.M) {
	log.SetDefaultLogger()
	ctx := context.Background()
	cloudaccount.EmbedService(ctx)
	EmbedService(ctx)
	grpcutil.StartTestServices(ctx)
	defer grpcutil.StopTestServices()

	// Single client for testing cloudcredit's Coupon and Credit APIs
	creditsClient = pb.NewCloudCreditsCreditServiceClient(testService.clientConn)
	couponClient = pb.NewCloudCreditsCouponServiceClient(testService.clientConn)
	//m.Run()
}
