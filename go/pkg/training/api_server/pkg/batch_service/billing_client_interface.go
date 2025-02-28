// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type BillingClientInterface interface {
	GetCloudAccountCredits(ctx context.Context, accountId string, history bool) ([]*pb.BillingCredit, error)
	GetCouponExpiry(ctx context.Context, couponCode string) ([]*pb.BillingCoupon, error)
}
