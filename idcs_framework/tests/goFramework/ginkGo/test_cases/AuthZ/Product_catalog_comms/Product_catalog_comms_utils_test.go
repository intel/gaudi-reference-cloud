package product_catalog_service_comms_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	billing_test "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing"
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
	. "github.com/onsi/ginkgo/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	clientConn               *grpc.ClientConn
	productCatalogClientConn *grpc.ClientConn
	products                 []*pb.Product
	vendors                  []*pb.Vendor
	xeon3SmallInstanceType   = "vm.xeon3.small"
	idcComputeServiceName    = "compute"
	tickerDuration           = 1
)

type TestService struct {
	billing_test.Service
	testDB manageddb.TestDb
}

func (ts *TestService) Init(ctx context.Context, cfg *billing_test.Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	cfg.CreditsInstallSchedulerInterval = uint16(tickerDuration)
	cfg.InitTestConfig()
	var err error
	const ROLE = "productcatalog"
	const DOMAIN = "productcatalog.idcs-system.svc.cluster.local"
	cfg.CreditsInstallSchedulerInterval = 1
	var result any

	body := GenerateRequestToGetCertificate(DOMAIN, ROLE)
	jsonerr := json.Unmarshal([]byte(body), &result)
	if jsonerr != nil {
		Fail("Error during Unmarshal(): " + err.Error())
	}

	json := GetFieldInfo(body)

	_, clientTLSConf, err := VaultCertSetup(json["issuing_ca"].(string), json["private_key"].(string), json["certificate"].(string))
	if err != nil {
		fmt.Print(err.Error())
	}

	test_config := credentials.NewTLS(clientTLSConf)

	ts.ManagedDb, err = ts.testDB.Start(ctx)
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
	if clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(test_config)); err != nil {
		return err
	}

	fmt.Print("Connected to service...")
	addr, err = resolver.Resolve(ctx, "productcatalog")
	if err != nil {
		return err
	}
	productCatalogClientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(test_config))
	if err != nil {
		return err
	}

	fmt.Print("Connected to Product Catalog service...")

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
	grpcutil.AddTestService[*billing_test.Config](&TestService{}, &billing_test.Config{})
	aria.EmbedService(ctx)
	standard.EmbedService(ctx)
	intel.EmbedService(ctx)
	billing_driver_intel.EmbedService(ctx)
	cloudaccount.EmbedService(ctx)
	meteringTests.EmbedService(ctx)
}
