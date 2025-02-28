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

var (
	AccountTypes = []pb.AccountType{pb.AccountType_ACCOUNT_TYPE_PREMIUM, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING, pb.AccountType_ACCOUNT_TYPE_INTEL, pb.AccountType_ACCOUNT_TYPE_STANDARD}
)

type ProductClient struct {
	ProductCatalogClient pb.ProductCatalogServiceClient
	ProductVendorClient  pb.ProductVendorServiceClient
}

func NewProductClient(ctx context.Context, resolver grpcutil.Resolver) (*ProductClient, error) {
	logger := log.FromContext(ctx).WithName("ProductClient.InitProductClient")
	var productCatalogConn *grpc.ClientConn

	productCatalogAddr := os.Getenv("PRODUCTCATALOG_ADDR")
	if productCatalogAddr == "" {
		productCatalogAddr, err := resolver.Resolve(ctx, "productcatalog")
		if err != nil {
			logger.Error(err, "grpc resolver not able to connect", "addr", productCatalogAddr)
			return nil, err
		}
	}

	productCatalogConn, err := grpcConnect(ctx, productCatalogAddr)
	if err != nil {
		return nil, err
	}
	pc := pb.NewProductCatalogServiceClient(productCatalogConn)
	pvc := pb.NewProductVendorServiceClient(productCatalogConn)
	return &ProductClient{
		ProductCatalogClient: pc,
		ProductVendorClient:  pvc}, nil
}

func (productClient *ProductClient) GetProductCatalogProducts(ctx context.Context) ([]*pb.Product, error) {
	return productClient.GetProductCatalogProductsForAccountTypes(ctx, AccountTypes)
}

func (productClient *ProductClient) GetProductCatalogProductsForAccountTypes(ctx context.Context, accountTypes []pb.AccountType) ([]*pb.Product, error) {
	productMap := make(map[string]*pb.Product)
	products := []*pb.Product{}
	var err error
	for _, acctType := range accountTypes {
		catalogProducts, lerr := productClient.GetProductCatalogProductsWithFilter(ctx, &pb.ProductFilter{AccountType: &acctType})
		err = lerr
		if err != nil {
			break
		}
		for _, catalogProduct := range catalogProducts {
			if product, ok := productMap[catalogProduct.GetId()]; ok {
				product.Rates = append(product.Rates, catalogProduct.Rates...)
			} else {
				productMap[catalogProduct.GetId()] = catalogProduct
			}
		}
	}
	for _, product := range productMap {
		products = append(products, product)
	}
	return products, err
}

func (productClient *ProductClient) GetProductCatalogProductsWithFilter(ctx context.Context, filter *pb.ProductFilter) ([]*pb.Product, error) {
	logger := log.FromContext(ctx).WithName("ProductClient.GetProductCatalogProducts")
	productResponse, err := productClient.ProductCatalogClient.AdminRead(ctx, filter)
	if err != nil {
		logger.Error(err, "error in product catalog client response")
		return nil, err
	}
	logger.Info("product catalog response", "productResponse", productResponse)
	return productResponse.GetProducts(), nil
}

func (productClient *ProductClient) SetProductStatus(ctx context.Context, productStatus *pb.SetProductStatusRequest) error {
	logger := log.FromContext(ctx).WithName("ProductClient.SetProductStatus")
	_, err := productClient.ProductCatalogClient.SetStatus(ctx, productStatus)
	if err != nil {
		logger.Error(err, "error in product catalog client response")
		return err
	}
	return nil
}

func (productClient *ProductClient) GetProductCatalogVendors(ctx context.Context) ([]*pb.Vendor, error) {
	logger := log.FromContext(ctx).WithName("ProductClient.GetProductCatalogVendors")
	vendorResponse, err := productClient.ProductVendorClient.Read(ctx, &pb.VendorFilter{})
	if err != nil {
		logger.Error(err, "error in product vendor client response")
		return nil, err
	}
	logger.Info("product catalog response", "vendorResponse", vendorResponse)
	return vendorResponse.GetVendors(), nil
}
