// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type ProductClientInterface interface {
	GetProductCatalogProducts(ctx context.Context) ([]*pb.Product, error)
	GetProductCatalogProductsWithFilter(ctx context.Context, filter *pb.ProductFilter) ([]*pb.Product, error)
	GetProductCatalogProductsForAccountTypes(ctx context.Context, accountTypes []pb.AccountType) ([]*pb.Product, error)
	SetProductStatus(ctx context.Context, productStatus *pb.SetProductStatusRequest) error
	GetProductCatalogVendors(ctx context.Context) ([]*pb.Vendor, error)
}
