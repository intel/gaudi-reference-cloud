// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"context"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProductCatalogServer struct {
	pb.UnimplementedProductCatalogServiceServer
	products []*pb.Product
}

func NewProductCatalogServer() *ProductCatalogServer {
	defaultProducts := []*pb.Product{
		{
			Name: "test123",
			Id:   "123",
			Metadata: map[string]string{
				"Model": "model_test_123",
			},
		},
	}
	return &ProductCatalogServer{
		products: defaultProducts,
	}
}

func NewProductCatalogServerWithProducts(products []*pb.Product) *ProductCatalogServer {
	return &ProductCatalogServer{
		products: products,
	}
}

func (p *ProductCatalogServer) AdminRead(_ context.Context, _ *pb.ProductFilter) (*pb.ProductResponse, error) {
	return &pb.ProductResponse{
		Products: p.products,
	}, nil
}

func (p *ProductCatalogServer) UserRead(_ context.Context, _ *pb.ProductUserFilter) (*pb.ProductResponse, error) {
	return &pb.ProductResponse{
		Products: p.products,
	}, nil
}

func (p *ProductCatalogServer) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
