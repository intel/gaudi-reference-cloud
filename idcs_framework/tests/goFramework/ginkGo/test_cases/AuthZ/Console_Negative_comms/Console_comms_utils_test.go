package cloud_console_negative_test

import (
	"context"
	"encoding/json"
	"fmt"

	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing"
	console "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/console"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type TestService struct {
	console.Service
}

var (
	invoiceClient pb.ConsoleInvoiceServiceClient
)

func (ts *TestService) Init(ctx context.Context, cfg *grpcutil.ListenConfig,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	const ROLE = "console"
	const DOMAIN = "console.idcs-system.svc.cluster.local"
	if err := ts.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}

	var result any
	body := GenerateRequestToGetCertificate(DOMAIN, ROLE)
	jsonerr := json.Unmarshal([]byte(body), &result)
	if jsonerr != nil {
		Fail("Error during Unmarshal(): " + jsonerr.Error())
	}
	json := GetFieldInfo(body)

	_, clientTLSConf, err := VaultCertSetup(json["issuing_ca"].(string), json["private_key"].(string), json["certificate"].(string))
	if err != nil {
		panic(err)
	}
	test_config := credentials.NewTLS(clientTLSConf)

	addr, err := resolver.Resolve(ctx, "console")
	if err != nil {
		panic(err)
	}

	clientConn, err := grpc.Dial(addr, grpc.WithTransportCredentials(test_config))
	if err != nil {
		panic(err)
	}
	fmt.Print(clientConn)

	addr2, err := resolver.Resolve(ctx, "populate-inflow-component-git-to-grpc-synchronizer")
	if err != nil {
		panic(err.Error())
	}
	if clientConn, err = grpc.Dial(addr2, grpc.WithTransportCredentials(test_config)); err != nil {
		panic(err.Error())
	}

	invoiceClient = pb.NewConsoleInvoiceServiceClient(clientConn)
	return nil
}

func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*grpcutil.ListenConfig](&TestService{}, &grpcutil.ListenConfig{})
	billing.EmbedService(ctx)
}
