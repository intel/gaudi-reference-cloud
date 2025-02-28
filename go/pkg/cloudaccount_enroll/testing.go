// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount_enroll

import (
	"context"
	"net/url"
	"strings"
	"time"

	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	icp "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount_enroll/icpintel"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (icp ICPClientMock) GetPersonId(ctx context.Context, email string, oid string) (string, error) {

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

func getICPConfig() (*icp.ICPConfig, error) {
	url, err := url.Parse(URL_TEST)
	if err != nil {
		return nil, err
	}

	return &icp.ICPConfig{
		URL:     url,
		Timeout: 1 * time.Millisecond,
	}, nil
}

func (ts *TestService) Init(ctx context.Context, cfg *EnrollConfig,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	icpCfg, err := getICPConfig()
	if err != nil {
		return err
	}
	ts.icpConfig = icpCfg
	var icpClient *ICPClientMock = &ICPClientMock{}
	ts.Service.setICPClient(icpClient)
	if err := ts.Service.initEnrollService(ctx, icpClient, resolver, grpcServer, cfg); err != nil {
		return err
	}
	addr, err := resolver.Resolve(ctx, "cloudaccount-enroll")
	if err != nil {
		return err
	}
	if clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	enrollClient = pb.NewCloudAccountEnrollServiceClient(clientConn)
	acctClient = pb.NewCloudAccountServiceClient(clientConn)
	return nil
}

func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*EnrollConfig](&TestService{}, &EnrollConfig{})
	cloudaccount.EmbedService(ctx)
	billing.EmbedService(ctx)
}
