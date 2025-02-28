// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"sync"
	"testing"

	//"testing"

	"github.com/golang/mock/gomock"
	//"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria"
	//"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_standard"
	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	//aria "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria"
	//"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	//standard "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_standard"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	//"google.golang.org/protobuf/types/known/emptypb"
)

//Todo: This is neither a unit test nor a integration test.
// For it to go across drives, it needs to either be a integration test that uses mocked interactions with Aria or
// use real integration with Aria.
// It cannot be a unit test because it calls drivers.

func InitMockProductServiceClient(mockController *gomock.Controller) {
	productCatalogClient := pb.NewMockProductCatalogServiceClient(mockController)
	productVendorClient := pb.NewMockProductVendorServiceClient(mockController)

	// Mock Create
	productCatalogReader := &sync.Map{}

	productMockRead := func(ctx context.Context, req *pb.ProductFilter, opts ...grpc.CallOption) (*pb.ProductResponse, error) {
		productCatalogReader.Store(0, req)
		return GetMockProductResponse(), nil
	}

	vendorMockRead := func(ctx context.Context, req *pb.VendorFilter, opts ...grpc.CallOption) (*pb.VendorResponse, error) {
		productCatalogReader.Store(1, req)
		return GetMockProductVendorResponse(), nil
	}
	productMockSetStatus := func(ctx context.Context, req *pb.SetProductStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	}

	productCatalogClient.EXPECT().AdminRead(gomock.Any(), gomock.Any()).DoAndReturn(productMockRead).AnyTimes()
	productCatalogClient.EXPECT().SetStatus(gomock.Any(), gomock.Any()).DoAndReturn(productMockSetStatus).AnyTimes()
	productVendorClient.EXPECT().Read(gomock.Any(), gomock.Any()).DoAndReturn(vendorMockRead).AnyTimes()
	billing.ProductVendorClient = productVendorClient
	billing.ProductCatalogClient = productCatalogClient
}

func GetMockProductResponse() *pb.ProductResponse {
	products := GetProducts()
	response := &pb.ProductResponse{
		Products: products,
	}
	return response
}

func GetMockProductVendorResponse() *pb.VendorResponse {
	vendors := GetVendors()
	response := &pb.VendorResponse{
		Vendors: vendors,
	}
	return response
}

func TestProductCatalogSync(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestProductCatalogSync")
	logger.Info("BEGIN")
	defer logger.Info("End")
	mockController := gomock.NewController(t)
	defer mockController.Finish()
	InitMockProductServiceClient(mockController)
	/**client := pb.NewBillingProductCatalogSyncServiceClient(clientConn)

	// Configure the drivers with a WaitGroup so we can synchronize
	// with them.
	wg := sync.WaitGroup{}
	wg.Add(len(driverSpecs))
	standard.SyncWait.Store(&wg)
	defer standard.SyncWait.Store(nil)

	//aria.SyncWait.Store(&wg)
	//defer aria.SyncWait.Store(nil)
	billing_driver_intel.SyncWait.Store(&wg)
	defer billing_driver_intel.SyncWait.Store(nil)
	_, err := client.Sync(ctx, &emptypb.Empty{})
	if err != nil {
		t.Error(err)
	}
	wg.Wait()**/
}
