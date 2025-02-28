package cloud_accound_comms_test

import (
	"context"
	"encoding/json"
	"fmt"

	_ "embed"

	cloud "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	. "github.com/onsi/ginkgo/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type TestService struct {
	mdb *manageddb.ManagedDb
}

type Test struct {
	cloud.Service
	testService TestService
	clientConn  *grpc.ClientConn
	testDb      manageddb.TestDb
}

var test Test

func ClientConn() *grpc.ClientConn {
	return test.clientConn
}

func EmbedService(ctx context.Context) {
	events.EmbedService(ctx)
	grpcutil.AddTestService[*config.Config](&test, &config.Config{DisableEmail: true})
}

func (test *Test) Init(ctx context.Context, cfg *config.Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	const ROLE = "cloudaccount"
	const DOMAIN = "cloudaccount.idcs-system.svc.cluster.local"
	var err error
	test.testService.mdb, err = test.testDb.Start(ctx)
	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}
	if err := test.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}
	addr, err := resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
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
		Fail(err.Error())
	}
	test_config := credentials.NewTLS(clientTLSConf)
	if test.clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(test_config)); err != nil {
		return err
	}
	return nil
}

func (test *Test) Done() error {
	grpcutil.ServiceDone[*config.Config](&test.Service)
	err := test.testDb.Stop(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (test *Test) ClientConn() *grpc.ClientConn {
	return test.clientConn
}
