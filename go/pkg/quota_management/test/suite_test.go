// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quota_management/database"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	quotaManagementServer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quota_management/pkg/server"

	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	testDb                    *manageddb.TestDb
	managedDb                 *manageddb.ManagedDb
	sqlDb                     *sql.DB
	quotaManagementSvc        *quotaManagementServer.QuotaManagementServiceClient
	bootstrappedConfig        []*quotaManagementServer.BootstrappedService
	bootstrappedConfigInvalid []*quotaManagementServer.BootstrappedService
)

func TestQuotaManagementService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Quota Management Service Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()

	grpcutil.UseTLS = false
	log.SetDefaultLogger()

	By("Starting database")
	testDb = &manageddb.TestDb{}
	var err error
	managedDb, err = testDb.Start(ctx)
	// Check Db setup succeeds
	Expect(err).Should(Succeed())
	Expect(managedDb).ShouldNot(BeNil())
	Expect(managedDb.Migrate(ctx, db.Fsys, "migrations")).Should(Succeed())
	sqlDb, err = managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	Expect(sqlDb).ShouldNot(BeNil())

	// Set up mock cloudaccount
	cloudAccountClient := NewMockCloudAccountServiceClient()

	// Set up bootstrap config
	bootstrappedConfig = NewMockBootstrappedServicesRequest()

	// Set up invalid bootstrap config
	bootstrappedConfigInvalid = bootstrappedConfig
	bootstrappedConfigInvalid[0].MaxLimits["compute"]["instances"] = 8

	// Initialize QuotaManagementServiceClient
	selectedRegion := "us-dev-1"
	quotaManagementSvc, err = quotaManagementServer.NewQuotaManagementServiceClient(ctx, sqlDb, cloudAccountClient, selectedRegion)
	Expect(err).Should(Succeed())
	Expect(quotaManagementSvc).ShouldNot(BeNil())
})

func NewMockCloudAccountServiceClient() pb.CloudAccountServiceClient {
	mockController := gomock.NewController(GinkgoT())
	cloudAccountClient := pb.NewMockCloudAccountServiceClient(mockController)

	// create mock cloudaccount with no quota associated
	cloudAccount := &pb.CloudAccount{
		Type: pb.AccountType_ACCOUNT_TYPE_STANDARD,
		Id:   "123456789012",
		Name: "123456789012",
	}

	cloudAccountClient.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(cloudAccount, nil).AnyTimes()
	return cloudAccountClient
}

func NewMockBootstrappedServicesRequest() []*quotaManagementServer.BootstrappedService {
	bootstrapServices := []*quotaManagementServer.BootstrappedService{
		{
			QuotaUnit: "COUNT",
			MaxLimits: map[string]map[string]int{
				"compute": {
					"instances":       10,
					"instance_groups": 5,
				},
				"storage": {
					"filesystems":      2,
					"totalSizeTB":      50,
					"buckets":          100,
					"bucketPrincipals": 50,
				},
			},
			QuotaAccountType: map[string]quotaManagementServer.QuotaAccountTypeDetails{
				"ENTERPRISE": {
					DefaultLimits: map[string]map[string]int{
						"compute": {
							"instances":       10,
							"instance_groups": 4,
						},
						"storage": {
							"filesystems":      2,
							"totalSizeTB":      40,
							"buckets":          80,
							"bucketPrincipals": 30,
						},
					},
				},
				"PREMIUM": {
					DefaultLimits: map[string]map[string]int{
						"compute": {
							"instances":       10,
							"instance_groups": 4,
						},
						"storage": {
							"filesystems":      2,
							"totalSizeTB":      40,
							"buckets":          80,
							"bucketPrincipals": 30,
						},
					},
				},
			},
		},
	}

	return bootstrapServices
}

var _ = AfterSuite(func() {
	// Clean up any resources here
})
