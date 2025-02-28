// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"
	"errors"
	"os"
	"testing"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProductClient tests
func TestNewProductClient(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution with env var", func(t *testing.T) {
		mockResolver := NewMockResolver(func(ctx context.Context, s string) (string, error) {
			// not actually connecting to anything, but needs to be "localhost" to not use TLS
			return "localhost:port", nil
		})

		// set environment variable
		os.Setenv("PRODUCTCATALOG_ADDR", "localhost:port")

		// run the function under test
		productClient, err := NewProductClient(ctx, mockResolver)
		assert.NoError(t, err)
		assert.NotNil(t, productClient)
	})

	t.Run("resolver error", func(t *testing.T) {
		mockResolver := NewMockResolver(func(ctx context.Context, s string) (string, error) {
			return "", errors.New("resolver error")
		})

		// set environment variable
		os.Setenv("PRODUCTCATALOG_ADDR", "")

		// run the function under test
		productClient, err := NewProductClient(ctx, mockResolver)
		assert.Error(t, err)
		assert.Nil(t, productClient)
	})
}

func TestGetProductCatalogProducts(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution", func(t *testing.T) {
		mockProduct := v1.Product{
			Name:        "test-product",
			Id:          "test-product-id",
			Created:     timestamppb.Now(),
			VendorId:    "test-vendor-id",
			FamilyId:    "test-family-id",
			Description: "test description",
			Metadata: map[string]string{
				"test-key": "test-value",
			},
			Eccn:      "test-eccn",
			Pcq:       "test-pcq",
			MatchExpr: "test-match-expr",
			Rates: []*v1.Rate{{
				AccountType: v1.AccountType_ACCOUNT_TYPE_STANDARD,
				Unit:        v1.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE,
				Rate:        "test-rate",
				UsageExpr:   "test-usage-expr",
			}},
			Status: "test-status",
		}
		mockAdminRead := func(ctx context.Context, in *v1.ProductFilter, opts ...grpc.CallOption) (*v1.ProductResponse, error) {
			return &v1.ProductResponse{
				Products: []*v1.Product{&mockProduct},
			}, nil
		}
		mockProductCatalogServiceClient := NewMockProductCatalogServiceClient(mockAdminRead, nil, nil, nil, nil)
		productClient := &ProductClient{
			ProductCatalogClient: mockProductCatalogServiceClient,
			ProductVendorClient:  nil,
		}

		// run the function under test
		products, err := productClient.GetProductCatalogProducts(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, products)
	})
}

func TestGetProductCatalogProductsForAccountTypes(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution", func(t *testing.T) {
		mockProduct := v1.Product{
			Name:        "test-product",
			Id:          "test-product-id",
			Created:     timestamppb.Now(),
			VendorId:    "test-vendor-id",
			FamilyId:    "test-family-id",
			Description: "test description",
			Metadata: map[string]string{
				"test-key": "test-value",
			},
			Eccn:      "test-eccn",
			Pcq:       "test-pcq",
			MatchExpr: "test-match-expr",
			Rates: []*v1.Rate{{
				AccountType: v1.AccountType_ACCOUNT_TYPE_STANDARD,
				Unit:        v1.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE,
				Rate:        "test-rate",
				UsageExpr:   "test-usage-expr",
			}},
			Status: "test-status",
		}
		mockAdminRead := func(ctx context.Context, in *v1.ProductFilter, opts ...grpc.CallOption) (*v1.ProductResponse, error) {
			return &v1.ProductResponse{
				Products: []*v1.Product{&mockProduct},
			}, nil
		}
		mockProductCatalogServiceClient := NewMockProductCatalogServiceClient(mockAdminRead, nil, nil, nil, nil)
		productClient := &ProductClient{
			ProductCatalogClient: mockProductCatalogServiceClient,
			ProductVendorClient:  nil,
		}

		// run the function under test
		products, err := productClient.GetProductCatalogProductsForAccountTypes(ctx, []v1.AccountType{v1.AccountType_ACCOUNT_TYPE_STANDARD})
		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.Equal(t, 1, len(products))
		assert.EqualValues(t, &mockProduct, products[0])
	})

	t.Run("product catalog client read error", func(t *testing.T) {
		mockAdminRead := func(ctx context.Context, in *v1.ProductFilter, opts ...grpc.CallOption) (*v1.ProductResponse, error) {
			return nil, errors.New("product catalog client read error")
		}
		mockProductCatalogServiceClient := NewMockProductCatalogServiceClient(mockAdminRead, nil, nil, nil, nil)
		productClient := &ProductClient{
			ProductCatalogClient: mockProductCatalogServiceClient,
			ProductVendorClient:  nil,
		}

		// run the function under test
		products, err := productClient.GetProductCatalogProductsForAccountTypes(ctx, []v1.AccountType{v1.AccountType_ACCOUNT_TYPE_STANDARD})
		assert.Error(t, err)
		assert.NotNil(t, products)
		assert.Equal(t, 0, len(products))
	})
}

func TestGetProductCatalogProductsWithFilter(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution", func(t *testing.T) {
		mockProduct := v1.Product{
			Name:        "test-product",
			Id:          "test-product-id",
			Created:     timestamppb.Now(),
			VendorId:    "test-vendor-id",
			FamilyId:    "test-family-id",
			Description: "test description",
			Metadata: map[string]string{
				"test-key": "test-value",
			},
			Eccn:      "test-eccn",
			Pcq:       "test-pcq",
			MatchExpr: "test-match-expr",
			Rates: []*v1.Rate{{
				AccountType: v1.AccountType_ACCOUNT_TYPE_STANDARD,
				Unit:        v1.RateUnit_RATE_UNIT_DOLLARS_PER_MINUTE,
				Rate:        "test-rate",
				UsageExpr:   "test-usage-expr",
			}},
			Status: "test-status",
		}
		mockAdminRead := func(ctx context.Context, in *v1.ProductFilter, opts ...grpc.CallOption) (*v1.ProductResponse, error) {
			return &v1.ProductResponse{
				Products: []*v1.Product{&mockProduct},
			}, nil
		}
		mockProductCatalogServiceClient := NewMockProductCatalogServiceClient(mockAdminRead, nil, nil, nil, nil)
		productClient := &ProductClient{
			ProductCatalogClient: mockProductCatalogServiceClient,
			ProductVendorClient:  nil,
		}

		// run the function under test
		products, err := productClient.GetProductCatalogProductsWithFilter(ctx, &v1.ProductFilter{})
		assert.NoError(t, err)
		assert.NotNil(t, products)
		assert.Equal(t, 1, len(products))
		assert.EqualValues(t, &mockProduct, products[0])
	})

	t.Run("product catalog client read error", func(t *testing.T) {
		mockAdminRead := func(ctx context.Context, in *v1.ProductFilter, opts ...grpc.CallOption) (*v1.ProductResponse, error) {
			return nil, errors.New("product catalog client read error")
		}
		mockProductCatalogServiceClient := NewMockProductCatalogServiceClient(mockAdminRead, nil, nil, nil, nil)
		productClient := &ProductClient{
			ProductCatalogClient: mockProductCatalogServiceClient,
			ProductVendorClient:  nil,
		}

		// run the function under test
		products, err := productClient.GetProductCatalogProductsWithFilter(ctx, &v1.ProductFilter{})
		assert.Error(t, err)
		assert.Nil(t, products)
	})
}

func TestSetProductStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution", func(t *testing.T) {
		mockSetStatus := func(ctx context.Context, in *v1.SetProductStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			return &emptypb.Empty{}, nil
		}
		mockProductCatalogServiceClient := NewMockProductCatalogServiceClient(nil, nil, nil, mockSetStatus, nil)
		productClient := &ProductClient{
			ProductCatalogClient: mockProductCatalogServiceClient,
			ProductVendorClient:  nil,
		}

		// run the function under test
		err := productClient.SetProductStatus(ctx, &v1.SetProductStatusRequest{})
		assert.NoError(t, err)
	})

	t.Run("product catalog client set status error", func(t *testing.T) {
		mockSetStatus := func(ctx context.Context, in *v1.SetProductStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			return nil, errors.New("product catalog client set status error")
		}
		mockProductCatalogServiceClient := NewMockProductCatalogServiceClient(nil, nil, nil, mockSetStatus, nil)
		productClient := &ProductClient{
			ProductCatalogClient: mockProductCatalogServiceClient,
			ProductVendorClient:  nil,
		}

		// run the function under test
		err := productClient.SetProductStatus(ctx, &v1.SetProductStatusRequest{})
		assert.Error(t, err)
	})
}

func TestGetProductCatalogVendors(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution", func(t *testing.T) {
		mockVendor := v1.Vendor{
			Name:        "test-vendor",
			Id:          "test-vendor-id",
			Created:     timestamppb.Now(),
			Description: "test description",
			Families:    []*v1.ProductFamily{},
		}
		mockRead := func(ctx context.Context, in *v1.VendorFilter, opts ...grpc.CallOption) (*v1.VendorResponse, error) {
			return &v1.VendorResponse{
				Vendors: []*v1.Vendor{&mockVendor},
			}, nil
		}
		mockProductVendorServiceClient := NewMockProductVendorServiceClient(mockRead)
		productClient := &ProductClient{
			ProductCatalogClient: nil,
			ProductVendorClient:  mockProductVendorServiceClient,
		}

		// run the function under test
		vendors, err := productClient.GetProductCatalogVendors(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, vendors)
		assert.Equal(t, 1, len(vendors))
		assert.EqualValues(t, &mockVendor, vendors[0])
	})

	t.Run("product vendor client read error", func(t *testing.T) {
		mockRead := func(ctx context.Context, in *v1.VendorFilter, opts ...grpc.CallOption) (*v1.VendorResponse, error) {
			return nil, errors.New("product vendor client read error")
		}
		mockProductVendorServiceClient := NewMockProductVendorServiceClient(mockRead)
		productClient := &ProductClient{
			ProductCatalogClient: nil,
			ProductVendorClient:  mockProductVendorServiceClient,
		}

		// run the function under test
		vendors, err := productClient.GetProductCatalogVendors(ctx)
		assert.Error(t, err)
		assert.Nil(t, vendors)
	})
}
