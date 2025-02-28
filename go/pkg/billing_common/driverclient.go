// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
)

type BillingDriverClients struct {
	BillingAcct   pb.BillingAccountServiceClient
	BillingCredit pb.BillingCreditServiceClient
	BillingOption pb.BillingOptionServiceClient
}

func NewBillingDriverClients(name string, conn *grpc.ClientConn) *BillingDriverClients {
	billingDriverClients := BillingDriverClients{
		BillingCredit: pb.NewBillingCreditServiceClient(conn),
		BillingAcct:   pb.NewBillingAccountServiceClient(conn),
		BillingOption: pb.NewBillingOptionServiceClient(conn),
	}
	return &billingDriverClients
}

var (
	standardDriver  *BillingDriverClients
	intelDriver     *BillingDriverClients
	ariaDriver      *BillingDriverClients
	cloudacctClient pb.CloudAccountServiceClient
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

func InitDrivers(ctx context.Context, cloudaccntClient pb.CloudAccountServiceClient, resolver grpcutil.Resolver) error {
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
	cloudacctClient = cloudaccntClient
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
		return nil, GetBillingInternalError("failed to read cloud account", err)
	}
	return GetDriverByType(cloudAcct.GetType())
}

func GetDriverAll(ctx context.Context, cloudAcctId string) (*BillingDriverClients, error) {
	cloudAcct, err := cloudacctClient.GetById(ctx,
		&pb.CloudAccountId{Id: cloudAcctId})
	// cloud acct needs to return a more valid error.
	if err != nil {
		return nil, GetBillingInternalError("failed to read cloud account", err)
	}
	return GetDriverAllByType(cloudAcct.GetType())
}
