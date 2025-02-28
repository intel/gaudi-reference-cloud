// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	aria "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	intel "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	standard "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_standard"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	meteringTests "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/tests"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/usage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Used for testing
var SyncWait = atomic.Pointer[sync.WaitGroup]{}

var (
	clientConn                     *grpc.ClientConn
	cloudAccountConn               *grpc.ClientConn
	meteringClientConn             *grpc.ClientConn
	usageConn                      *grpc.ClientConn
	Test                           *TestService
	products                       []*pb.Product
	vendors                        []*pb.Vendor
	xeon3SmallInstanceType         = "vm.xeon3.small"
	idcComputeServiceName          = "compute"
	tickerDuration                 = 1
	usageTickerDuration            = 3600
	testSchedulerCloudAccountState *SchedulerCloudAccountState
)

type TestService struct {
	Service
	testDB manageddb.TestDb
}

func (ts *TestService) Init(ctx context.Context, cfg *Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	cfg.CreditsInstallSchedulerInterval = uint16(tickerDuration)
	cfg.ReportUsageSchedulerInterval = uint16(usageTickerDuration)
	cfg.InitTestConfig()
	var err error
	ts.ManagedDb, err = ts.testDB.Start(ctx)
	Test = ts
	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}

	if err := ts.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}

	addr, err := resolver.Resolve(ctx, "billing")
	if err != nil {
		return err
	}
	if clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	addr, err = resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		return err
	}
	if cloudAccountConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	addr, err = resolver.Resolve(ctx, "metering")
	if err != nil {
		return err
	}
	meteringClientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	testSchedulerCloudAccountState = &SchedulerCloudAccountState{
		AccessTimestamp: "",
	}

	return nil
}

func GetProducts() []*pb.Product {
	if len(products) == 0 {
		SetupProductsAndVendors()
	}
	return products
}

func GetVendors() []*pb.Vendor {
	if len(vendors) == 0 {
		SetupProductsAndVendors()
	}
	return vendors
}

func SetupProductsAndVendors() {
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	computeProductId := uuid.NewString()
	matchExpressionCompute := fmt.Sprintf("billUsage && service == \"%s\" && instanceType == \"%s\"", idcComputeServiceName, xeon3SmallInstanceType)

	vendors = make([]*pb.Vendor, 0)
	products = make([]*pb.Product, 0)

	idcComputeProductFamilies := make([]*pb.ProductFamily, 0)
	idcComputeProductFamily := billing.GetIdcComputeProductFamily(idcComputeProductFamilyId)
	idcComputeProductFamilies = append(idcComputeProductFamilies, idcComputeProductFamily)

	vendor := billing.GetIdcVendor(vendorId, idcComputeProductFamilies)
	vendors = append(vendors, vendor)

	// represents a base SKU
	computeProduct := &pb.Product{
		Name:        "computeProductVMSmallXeon3Name",
		Id:          computeProductId,
		VendorId:    vendorId,
		FamilyId:    idcComputeProductFamilyId,
		Description: uuid.NewString(),
		Rates:       billing.GetRates(),
		MatchExpr:   matchExpressionCompute,
		Metadata:    billing.GetMetadata(),
	}
	products = append(products, computeProduct)
}

func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*Config](&TestService{}, &Config{TestProfile: true})
	aria.EmbedService(ctx)
	standard.EmbedService(ctx)
	intel.EmbedService(ctx)
	billing_driver_intel.EmbedService(ctx)
	cloudaccount.EmbedService(ctx)
	meteringTests.EmbedService(ctx)
	usage.EmbedService(ctx)
}

func (ts *TestService) Done() {
	if err := ts.testDB.Stop(context.Background()); err != nil {
		panic(err)
	}
}
