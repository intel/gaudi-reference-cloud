package billing_standard_comms_test

import (
	"context"

	standard "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_standard"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type TestService struct {
	standard.Service
}

var test TestService

func (ts *TestService) Hosts() []string {
	return []string{ts.Name(), "billing-intel"}
}

func (*TestService) Init(ctx context.Context, cfg *grpcutil.ListenConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {

	logger := log.FromContext(ctx).WithName("BillingStandardDriver Service.Init")

	logger.Info("initializing billing standard service")

	pb.RegisterBillingAccountServiceServer(grpcServer, &standard.StandardBillingAccountService{})
	pb.RegisterBillingOptionServiceServer(grpcServer, &standard.StandardBillingOptionService{})
	pb.RegisterBillingRateServiceServer(grpcServer, &standard.StandardBillingRateService{})
	pb.RegisterBillingCreditServiceServer(grpcServer, &standard.StandardBillingCreditService{})
	reflection.Register(grpcServer)
	return nil
}

func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*grpcutil.ListenConfig](&test, &grpcutil.ListenConfig{})
}
