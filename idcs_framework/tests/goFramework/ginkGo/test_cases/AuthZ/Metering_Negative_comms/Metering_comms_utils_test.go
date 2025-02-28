package metering_service_negative_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/server"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	test       Test
	clientConn *grpc.ClientConn
)

func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*server.Config](&test, &server.Config{
		TestMode: true,
	})
}

type Test struct {
	server.Service
	testDb     manageddb.TestDb
	clientConn *grpc.ClientConn
}

func (test *Test) Init(ctx context.Context, cfg *server.Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	const ROLE = "metering"
	const DOMAIN = "metering.idcs-system.svc.cluster.local"
	var err error
	var result any
	body := GenerateRequestToGetCertificate(DOMAIN, ROLE)
	jsonerr := json.Unmarshal([]byte(body), &result)
	if jsonerr != nil {
		Fail("Error during Unmarshal(): " + jsonerr.Error())
	}
	json := GetFieldInfo(body)

	_, clientTLSConf, err := VaultCertSetup(json["issuing_ca"].(string), json["private_key"].(string), json["certificate"].(string))
	if err != nil {
		Fail(err.Error())
	}
	test_config := credentials.NewTLS(clientTLSConf)
	test.Mdb, err = test.testDb.Start(ctx)
	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}
	if err = test.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		panic(err.Error())
	}
	addr, err := resolver.Resolve(ctx, "metering")
	if err != nil {
		panic(err.Error())
	}
	if test.clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(test_config)); err != nil {
		panic(err.Error())
	}
	addr2, err := resolver.Resolve(ctx, "populate-inflow-component-git-to-grpc-synchronizer")
	if err != nil {
		panic(err.Error())
	}
	if clientConn, err = grpc.Dial(addr2, grpc.WithTransportCredentials(test_config)); err != nil {
		panic(err.Error())
	}
	return nil
}

func (test *Test) Done() error {
	grpcutil.ServiceDone[*server.Config](&test.Service)
	err := test.testDb.Stop(context.Background())
	if err != nil {
		return err
	}
	return nil
}
