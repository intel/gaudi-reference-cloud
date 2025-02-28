// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
)

var (
	ProductCatalogClient pb.ProductCatalogServiceClient
	ProductVendorClient  pb.ProductVendorServiceClient
	AccountTypes         = []pb.AccountType{pb.AccountType_ACCOUNT_TYPE_PREMIUM, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING, pb.AccountType_ACCOUNT_TYPE_INTEL, pb.AccountType_ACCOUNT_TYPE_STANDARD}
)

const billingEnabledFlagInProdCatalog = "billingEnable"

type ProductClient struct {
	productCatalogClient pb.ProductCatalogServiceClient
	productVendorClient  pb.ProductVendorServiceClient
}

func NewProductClient(ctx context.Context, resolver grpcutil.Resolver) (*ProductClient, error) {
	logger := log.FromContext(ctx).WithName("ProductClient.InitProductClient")
	var productCatalogConn *grpc.ClientConn
	productCatalogAddr, err := resolver.Resolve(ctx, "productcatalog")
	if err != nil {
		logger.Error(err, "grpc resolver not able to connect", "addr", productCatalogAddr)
		return nil, err
	}
	productCatalogConn, err = grpcConnect(ctx, productCatalogAddr)
	if err != nil {
		return nil, err
	}
	pc := pb.NewProductCatalogServiceClient(productCatalogConn)
	pvc := pb.NewProductVendorServiceClient(productCatalogConn)
	ProductCatalogClient = pc
	ProductVendorClient = pvc
	return &ProductClient{productCatalogClient: pc,
		productVendorClient: pvc}, nil
}

func (productClient *ProductClient) GetProductCatalogProducts(ctx context.Context) ([]*pb.Product, error) {
	return productClient.GetProductCatalogProductsForAccountTypes(ctx, AccountTypes)
}

// this code needs to be cleaned up.
func (productClient *ProductClient) GetProductCatalogProductsForAccountTypes(ctx context.Context, accountTypes []pb.AccountType) ([]*pb.Product, error) {
	logger := log.FromContext(ctx).WithName("ProductClient.GetProductCatalogProductsForAccountTypes")
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
			if billingEnabled, ok := catalogProduct.Metadata[billingEnabledFlagInProdCatalog]; ok {
				if billingEnabled == "false" {
					logger.Info("not handling product because billing is disabled for", "name", catalogProduct.Name)
					continue
				}
			}
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
	productResponse, err := ProductCatalogClient.AdminRead(ctx, filter)
	if err != nil {
		logger.Error(err, "error in product catalog client response")
		return nil, err
	}
	logger.Info("product catalog response", "productResponse", productResponse)

	products := []*pb.Product{}

	for _, product := range productResponse.Products {
		if billingEnabled, ok := product.Metadata[billingEnabledFlagInProdCatalog]; ok {
			if billingEnabled == "false" {
				logger.Info("not handling product because billing is disabled for", "name", product.Name)
				continue
			}
		}
		products = append(products, product)
	}

	return products, nil
}

func (productClient *ProductClient) SetProductStatus(ctx context.Context, productStatus *pb.SetProductStatusRequest) error {
	logger := log.FromContext(ctx).WithName("ProductClient.SetProductStatus")
	_, err := ProductCatalogClient.SetStatus(ctx, productStatus)
	if err != nil {
		logger.Error(err, "error in product catalog client response")
		return err
	}
	return nil
}

func (productClient *ProductClient) GetProductCatalogVendors(ctx context.Context) ([]*pb.Vendor, error) {
	logger := log.FromContext(ctx).WithName("ProductClient.GetProductCatalogVendors")
	vendorResponse, err := ProductVendorClient.Read(ctx, &pb.VendorFilter{})
	if err != nil {
		logger.Error(err, "error in product vendor client response")
		return nil, err
	}
	logger.Info("product catalog response", "vendorResponse", vendorResponse)
	return vendorResponse.GetVendors(), nil
}
