// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type BillingDriverClients struct {
	BillingDriver
	billingAcct   pb.BillingAccountServiceClient
	billingCredit pb.BillingCreditServiceClient
}

func NewBillingDriverClients(name string, conn *grpc.ClientConn) *BillingDriverClients {
	billingDriver := BillingDriver{
		name:             name,
		conn:             conn,
		billingOption:    pb.NewBillingOptionServiceClient(conn),
		billingRate:      pb.NewBillingRateServiceClient(conn),
		billingInvoice:   pb.NewBillingInvoiceServiceClient(conn),
		payment:          pb.NewPaymentServiceClient(conn),
		billingInstances: pb.NewBillingInstancesServiceClient(conn),
	}
	billingDriverClients := BillingDriverClients{
		BillingDriver: billingDriver,
		billingAcct:   pb.NewBillingAccountServiceClient(conn),
		billingCredit: pb.NewBillingCreditServiceClient(conn),
	}
	return &billingDriverClients
}

var (
	standardDriver  *BillingDriverClients
	intelDriver     *BillingDriverClients
	ariaDriver      *BillingDriverClients
	cloudacctClient pb.CloudAccountServiceClient
	usageClient     *billingCommon.MeteringClient
	productClient   billingCommon.ProductClientInterface
)

type tDriverSpec struct {
	name    string
	service string
	driver  **BillingDriverClients
}

var driverSpecs = []tDriverSpec{
	{name: "standard", service: "billing-standard", driver: &standardDriver},
	{name: "intel", service: "billing-intel", driver: &intelDriver},
	{name: "aria", service: "billing-aria", driver: &ariaDriver},
}

var driverConnections = make(map[string]*grpc.ClientConn)

func grpcConnect(ctx context.Context, addr string) (*grpc.ClientConn, error) {

	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}
	conn, err := grpcutil.NewClient(ctx, addr, dialOptions...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func InitDrivers(ctx context.Context, resolver grpcutil.Resolver) error {
	addr, err := resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		return err
	}
	conn, err := grpcConnect(ctx, addr)
	if err != nil {
		return err
	}
	cloudacctClient = pb.NewCloudAccountServiceClient(conn)

	// metering usage client
	usageClient, err = billingCommon.NewMeteringClient(ctx, resolver)
	if err != nil {
		return err
	}

	// product catalog client
	productClient, err = billingCommon.NewProductClient(ctx, resolver)

	if err != nil {
		return err
	}

	for _, spec := range driverSpecs {
		addr, err := resolver.Resolve(ctx, spec.service)
		if err != nil {
			return err
		}
		conn, err := grpcConnect(ctx, addr)
		if err != nil {
			return err
		}
		*spec.driver = NewBillingDriverClients(spec.name, conn)
		driverConnections[spec.name] = conn
	}
	return nil
}

// TODO: Examine the constraints of utilizing the proxy pattern to support selective calls for enterprise pending.
func GetDriverByType(typ pb.AccountType) (*BillingDriverClients, error) {
	switch typ {
	case pb.AccountType_ACCOUNT_TYPE_STANDARD:
		return standardDriver, nil
	case pb.AccountType_ACCOUNT_TYPE_PREMIUM:
		fallthrough
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE:
		fallthrough
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING:
		return ariaDriver, nil
	case pb.AccountType_ACCOUNT_TYPE_INTEL:
		return intelDriver, nil
	default:
		return nil, fmt.Errorf("invalid account type %v", typ)
	}
}

func GetDriverAllByType(typ pb.AccountType) (*BillingDriverClients, error) {
	switch typ {
	case pb.AccountType_ACCOUNT_TYPE_STANDARD:
		return standardDriver, nil
	case pb.AccountType_ACCOUNT_TYPE_PREMIUM:
		fallthrough
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING:
		fallthrough
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE:
		return ariaDriver, nil
	case pb.AccountType_ACCOUNT_TYPE_INTEL:
		return intelDriver, nil
	default:
		return nil, fmt.Errorf("invalid account type %v", typ)
	}
}

func GetDriver(ctx context.Context, cloudAcctId string) (*BillingDriverClients, error) {
	cloudAcct, err := cloudacctClient.GetById(ctx,
		&pb.CloudAccountId{Id: cloudAcctId})
	// cloud acct needs to return a more valid error.
	if err != nil {
		return nil, GetBillingInternalError(FailedToReadCloudAccount, err)
	}
	return GetDriverByType(cloudAcct.GetType())
}

func GetDriverAll(ctx context.Context, cloudAcctId string) (*BillingDriverClients, error) {
	cloudAcct, err := cloudacctClient.GetById(ctx,
		&pb.CloudAccountId{Id: cloudAcctId})
	// cloud acct needs to return a more valid error.
	if err != nil {
		return nil, GetBillingInternalError(FailedToReadCloudAccount, err)
	}
	return GetDriverAllByType(cloudAcct.GetType())
}
