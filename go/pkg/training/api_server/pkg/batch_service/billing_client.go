// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
)

type BillingClient struct {
	BillingCreditServiceClient pb.BillingCreditServiceClient
	BillingCouponServiceClient pb.BillingCouponServiceClient
}

func NewBillingClient(ctx context.Context, resolver grpcutil.Resolver) (*BillingClient, error) {
	logger := log.FromContext(ctx).WithName("BillingClient.NewBillingClient")
	var billingConn *grpc.ClientConn

	billingAddr := os.Getenv("PRODUCTCATALOG_ADDR")
	if billingAddr == "" {
		billingAddr, err := resolver.Resolve(ctx, "productcatalog")
		if err != nil {
			logger.Error(err, "grpc resolver not able to resolve", "addr", billingAddr)
			return nil, err
		}
	}

	billingConn, err := grpcConnect(ctx, billingAddr)
	if err != nil {
		return nil, err
	}
	creditSvc := pb.NewBillingCreditServiceClient(billingConn)
	couponSvc := pb.NewBillingCouponServiceClient(billingConn)

	return &BillingClient{BillingCreditServiceClient: creditSvc, BillingCouponServiceClient: couponSvc}, nil
}

func (billingClient *BillingClient) GetCloudAccountCredits(ctx context.Context, accountId string, history bool) ([]*pb.BillingCredit, error) {
	logger := log.FromContext(ctx).WithName("BillingClient.GetCloudAccountCredits")

	in := &pb.BillingCreditFilter{CloudAccountId: accountId, History: &history}
	account, err := billingClient.BillingCreditServiceClient.Read(ctx, in)
	if err != nil {
		logger.Error(err, "error in billing credit client response")
		return nil, err
	}
	logger.Info("billingCreditClient response", "account", account)
	return account.GetCredits(), nil
}

func (billingClient *BillingClient) GetCouponExpiry(ctx context.Context, couponCode string) ([]*pb.BillingCoupon, error) {
	logger := log.FromContext(ctx).WithName("GetCloudAccountCredits.GetCouponDetails")
	in := &pb.BillingCouponFilter{
		Code: &couponCode,
	}
	coupons, err := billingClient.BillingCouponServiceClient.Read(ctx, in)
	if err != nil {
		logger.Error(err, "error in billing credit coupon response")
		return nil, err
	}
	logger.Info("billingCouponClient response", "coupons", coupons)
	return coupons.GetCoupons(), nil
}
