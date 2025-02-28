// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// This is a very basic test implementation of the product catalog client to behave the way a testing scenario needs it to behave.
// This is not emulation or simulation. This is merely to verify acting on certain expected responses.
// Expecations of the response are important to assert choices or decisions or operations etc to be worked on based on expectations.
// This can get as fancy as needed :-)

type ProductTestClient struct {
	productCatalogVendors  []*pb.Vendor
	productCatalogProducts []*pb.Product
}

func NewProductTestClient(productCatalogVendors []*pb.Vendor, productCatalogProducts []*pb.Product) *ProductTestClient {
	return &ProductTestClient{productCatalogVendors: productCatalogVendors, productCatalogProducts: productCatalogProducts}
}

func (productTestClient *ProductTestClient) GetProductCatalogProducts(ctx context.Context) ([]*pb.Product, error) {
	logger := log.FromContext(ctx).WithName("ProductTestClient.GetProductCatalogProducts")
	logger.Info("returning test product catalog products")
	return productTestClient.productCatalogProducts, nil
}

func (productTestClient *ProductTestClient) GetProductCatalogProductsWithFilter(ctx context.Context, filter *pb.ProductFilter) ([]*pb.Product, error) {
	logger := log.FromContext(ctx).WithName("ProductTestClient.GetProductCatalogProducts")
	logger.Info("returning test product catalog products")
	return productTestClient.productCatalogProducts, nil
}

func (productTestClient *ProductTestClient) GetProductCatalogProductsForAccountTypes(ctx context.Context, accountTypes []pb.AccountType) ([]*pb.Product, error) {
	logger := log.FromContext(ctx).WithName("ProductTestClient.GetProductCatalogProducts")
	logger.Info("returning test product catalog products")
	return productTestClient.productCatalogProducts, nil
}

func (productTestClient *ProductTestClient) SetProductStatus(ctx context.Context, productStatus *pb.SetProductStatusRequest) error {
	logger := log.FromContext(ctx).WithName("ProductTestClient.SetProductStatus")
	logger.Info("test product client does not implement set product status")
	return nil
}

func (productTestClient *ProductTestClient) GetProductCatalogVendors(ctx context.Context) ([]*pb.Vendor, error) {
	logger := log.FromContext(ctx).WithName("productTestClient.GetProductCatalogVendors")
	logger.Info("returning test product catalog vendors")
	return productTestClient.productCatalogVendors, nil
}
