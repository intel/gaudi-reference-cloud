package cloud_account_enrroll_comms_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	enroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount_enroll"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/onsi/ginkgo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	clientConn   *grpc.ClientConn
	enrollClient pb.CloudAccountEnrollServiceClient
	acctClient   pb.CloudAccountServiceClient
)

type TestService struct {
	Service
}

type ICPClientMock struct {
}

func (icp ICPClientMock) getPersonId(ctx context.Context, email string, oid string) (string, error) {

	//Return
	if strings.Contains(email, "123456") {
		return "person-id-01", nil
	} else {
		return "person-id-02", nil
	}
}
func (icp ICPClientMock) IsEnterprisePending(ctx context.Context, email string, oid string) (bool, string, error) {

	//Return
	if strings.Contains(email, "Pending") {
		return true, "person-id-01", nil
	} else {
		return false, "person-id-02", nil
	}
}

const URL_TEST = "https://apis-sandbox.intel.com"

func getICPConfig() (*enroll.ICPConfig, error) {
	url, err := url.Parse(URL_TEST)
	if err != nil {
		return nil, err
	}

	return &enroll.ICPConfig{
		URL:     url,
		Timeout: 1 * time.Millisecond,
	}, nil
}

func (ts *TestService) Init(ctx context.Context, cfg *EnrollConfig,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	const ROLE = "cloudaccount-enroll"
	const DOMAIN = "cloudaccount-enroll.idcs-system.svc.cluster.local"
	icpCfg, err := getICPConfig()
	if err != nil {
		return err
	}

	ts.icpConfig = icpCfg
	var icpClient *enroll.ICPClientMock = &enroll.ICPClientMock{}
	ts.Service.setICPClient(icpClient)

	var result any
	body := GenerateRequestToGetCertificate(DOMAIN, ROLE)
	jsonerr := json.Unmarshal([]byte(body), &result)
	if jsonerr != nil {
		ginkgo.Fail("Error during Unmarshal(): " + jsonerr.Error())
	}
	json := GetFieldInfo(body)

	_, clientTLSConf, err := VaultCertSetup(json["issuing_ca"].(string), json["private_key"].(string), json["certificate"].(string))
	if err != nil {
		panic(err.Error())
	}
	test_config := credentials.NewTLS(clientTLSConf)
	addr, err := resolver.Resolve(ctx, "cloudaccount-enroll")
	if err != nil {
		panic(err.Error())
	}
	if clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(test_config)); err != nil {
		panic(err.Error())
	}

	addr2, err := resolver.Resolve(ctx, "populate-inflow-component-git-to-grpc-synchronizer")
	if err != nil {
		panic(err.Error())
	}
	if clientConn, err = grpc.Dial(addr2, grpc.WithTransportCredentials(test_config)); err != nil {
		panic(err.Error())
	}
	fmt.Print(addr2)
	enrollClient = pb.NewCloudAccountEnrollServiceClient(clientConn)
	acctClient = pb.NewCloudAccountServiceClient(clientConn)

	return nil
}

func EmbedService(ctx context.Context) error {
	var err error
	grpcutil.AddTestService[*EnrollConfig](&TestService{}, &EnrollConfig{})
	cloudaccount.EmbedService(ctx)
	billing.EmbedService(ctx)
	return err
}
