// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"io/fs"
	"net"
	"strings"
	"testing"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/config"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	migrationDir                string
	grpcServer                  *grpc.Server
	testDb                      *manageddb.TestDb
	managedDb                   *manageddb.ManagedDb
	sqlDb                       *sql.DB
	errc                        chan error
	ComputeInstanceTypeService  v1.InstanceTypeServiceClient
	SshKeyService               v1.SshPublicKeyServiceClient
	VnetClient                  v1.VNetServiceClient
	ProductCatalogServiceClient v1.ProductCatalogServiceClient
	InstanceServiceClient       v1.InstanceServiceClient
)

type Test struct {
	clientConn *grpc.ClientConn
	pool       *dockertest.Pool
	resource   *dockertest.Resource
}

func (test *Test) Init() {
	fmt.Println("In Init()")
	test.grpcInit()
}

func (test *Test) Done() {
	fmt.Println("In Done()")
	defer sqlDb.Close()
	test.grpcDone()
	if err := testDb.Stop(context.Background()); err != nil {
		panic(err)
	}
}

func (test *Test) grpcInit() {
	// Context and Log
	ctx := context.Background()
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	log.Info("Initializing GRPC..")
	fmt.Println("Initializing GRPC..")

	// DB
	testDb = &manageddb.TestDb{}
	var err error
	managedDb, err = testDb.Start(ctx)
	if err != nil {
		log.Error(err, "error starting database")
		return
	}

	err = Migrate(ctx, managedDb,
		db.NewTemplateFs(db.MigrationsFs, map[string]interface{}{"DbSeedEnabled": false}),
		db.MigrationsDir,
		db.MigrationsFsUnit,
		db.MigrationsDirUnit)
	if err != nil {
		log.Error(err, "error migrating database")
		return
	}
	sqlDb, err = managedDb.Open(ctx)
	if err != nil {
		log.Error(err, "error opening database ")
		return
	}
	cfg := config.Config{
		EncryptionKeys: "./encryption_keys",
	}

	// computeClient := test.NewMockComputeServiceClient(*testing.T)
	// GRPC SERVER AND IKS SERVICES
	grpcServer = grpc.NewServer()

	iksSrv, err := server.NewIksService(
		sqlDb,
		ComputeInstanceTypeService,
		SshKeyService,
		VnetClient,
		nil,
		InstanceServiceClient,
		cfg)
	if err != nil {
		log.Error(err, "error initializing iks service")
		return
	}
	iksReconcilerSrv, err := server.NewIksPrivateReconcilerService(sqlDb, cfg)
	if err != nil {
		log.Error(err, "error initializing iks service")
		return
	}

	v1.RegisterIksServer(grpcServer, iksSrv)
	v1.RegisterIksPrivateReconcilerServer(grpcServer, iksReconcilerSrv)

	// OPEN LISTENER AND SERVE
	lis, err := net.Listen("tcp", "localhost:")
	if err != nil {
		log.Error(err, "listen failed")
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error(err, "Serve failed")
		}
	}()

	// CLIENT CONNECTION
	test.clientConn, err = grpc.Dial(lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(err, "gprc.Dial error")
	}

	log.Info("Initializing GRPC Completed..")
}

func (test *Test) grpcDone() {
	grpcServer.GracefulStop()
}

func Migrate(ctx context.Context, m *manageddb.ManagedDb, fsys fs.FS, subdir string, fsysUnit fs.FS, subdirUnit string) error {
	log := log.FromContext(ctx).WithName("ManagedDb.Migrate")
	log.Info("Creating and migrating database")
	dir, err := iofs.New(fsys, subdir)
	if err != nil {
		return fmt.Errorf("create iofs for migrate: %w", err)
	}

	mm, err := migrate.NewWithSourceInstance("iofs", dir, m.DatabaseURL.String())
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err = mm.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate database: %w", err)
	}
	test, err := fs.ReadFile(fsysUnit, subdirUnit+"/999999_ed_temp_data.up.sql")
	body := io.NopCloser(strings.NewReader(string(test)))
	migr, err := migrate.NewMigration(body, "mockdata", 99999999, 99999999)
	err = mm.Run(migr)
	if err != nil {
		return fmt.Errorf("Adding fake data to Db: %w", err)
	}
	log.Info("Database migrated successfully")
	return nil
}

func NewMockComputeServiceClient(ctx context.Context, t *testing.T) *v1.MockInstanceTypeServiceClient {
	mockController := gomock.NewController(t)
	computeClient := v1.NewMockInstanceTypeServiceClient(mockController)
	instances := &v1.InstanceTypeSearchResponse{
		Items: []*v1.InstanceType{
			{
				Metadata: &v1.InstanceType_Metadata{
					Name: "vm-spr-tny",
				},
				Spec: &v1.InstanceTypeSpec{
					Name:             "vm-spr-tny",
					DisplayName:      "Tiny VM - Intel® Xeon 4th Gen ® Scalable processor",
					Description:      "4th Generation Intel® Xeon® Scalable processor",
					InstanceCategory: v1.InstanceCategory_VirtualMachine,
					Cpu: &v1.CpuSpec{
						Cores:     4,
						Id:        "0x806F8",
						ModelName: "4th Generation Intel® Xeon® Scalable processor",
						Sockets:   1,
						Threads:   1,
					},
					Memory:  &v1.MemorySpec{},
					Disks:   []*v1.DiskSpec{},
					Gpu:     &v1.GpuSpec{},
					HbmMode: "",
				},
			},
			{
				Metadata: &v1.InstanceType_Metadata{
					Name: "vm-spr-lrg",
				},
				Spec: &v1.InstanceTypeSpec{
					Name:             "vm-spr-lrg",
					DisplayName:      "Large VM - Intel® Xeon 4th Gen ® Scalable processor",
					Description:      "4th Generation Intel® Xeon® Scalable processor",
					InstanceCategory: v1.InstanceCategory_VirtualMachine,
					Cpu: &v1.CpuSpec{
						Cores:     32,
						Id:        "0x806F8",
						ModelName: "4th Generation Intel® Xeon® Scalable processor",
						Sockets:   1,
						Threads:   1,
					},
					Memory:  &v1.MemorySpec{},
					Disks:   []*v1.DiskSpec{},
					Gpu:     &v1.GpuSpec{},
					HbmMode: "",
				},
			},
			{
				Metadata: &v1.InstanceType_Metadata{
					Name: "bm-icp-gaudi2",
				},
				Spec: &v1.InstanceTypeSpec{
					Name:             "bm-icp-gaudi2",
					DisplayName:      "8 Gaudi2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk",
					Description:      "8 Gaudi2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk",
					InstanceCategory: v1.InstanceCategory_BareMetalHost,
					Cpu: &v1.CpuSpec{
						Cores:     8,
						Id:        "0x806F8",
						ModelName: "4th Generation Intel® Xeon® Scalable processor",
						Sockets:   1,
						Threads:   1,
					},
					Memory:  &v1.MemorySpec{},
					Disks:   []*v1.DiskSpec{},
					Gpu:     &v1.GpuSpec{},
					HbmMode: "",
				},
			},
		},
	}
	computeClient.EXPECT().Search(ctx, &v1.InstanceTypeSearchRequest{}).Return(instances, nil)
	return computeClient
}

func NewMockSshServiceClient(ctx context.Context, t *testing.T) *v1.MockSshPublicKeyServiceClient {
	mockController := gomock.NewController(t)
	sshClient := v1.NewMockSshPublicKeyServiceClient(mockController)

	sshkeysearchresponse := &v1.SshPublicKeySearchResponse{
		Items: []*v1.SshPublicKey{
			{
				Metadata: &v1.ResourceMetadata{
					CloudAccountId: "iks_user",
					Name:           "6e35780c-04_iks_user_ssh_key",
					ResourceId:     "c608f460-276d-40a2-bdff-76d7e813115c",
					CreationTimestamp: &timestamppb.Timestamp{
						Seconds: 1699140227,
						Nanos:   580620000,
					},
				},
				Spec: &v1.SshPublicKeySpec{
					SshPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDOX0rNUQLB8/tkyAg8exMnA4ledt6/DosUnp5Phxu/A1Mypbyvqo9OgrfjqAW3pAMVXrH1FuAHdP5IGdsPyvP7hd5S43BA4JR0NZMVwKUx11xtE+EGEhWdz8kfTvVVxhvcR1LbiArRdHYfEDKbdd4s2pivqoELKDXZfdn4JDNLuEq/XpcYn53m9z+Q6dz9MLjuUmDqOavc2Ih4xrYmbM6VcN2zHnOBzQjEpCGZM1/o2c3KQzloxOVRp2iFqK4PVau2mL/UaIjHADifeEveWQS18D9KQHP/4ATk+Nj/5JgPirmmZ39vJRzFJNpUsrBrSodvDJrRfbx5sCKQhgs9wZOviCnE4iNqzTefsPs3E1Cf+rgowLZ9kJIS2RzMjwBE3HsIOpvbbdbaGoEfIAvPQG/3c65PwPUGzFVM0QXFvQUdKu3fxQ566t3cCTHiJIiF3/1PKR1pN30CXsJ4xJvARKUddVA9qgAXHLJMQEPx+BlO8q4auI+kVoYETap0obPh+Ouz28FgT4SbOprg070JiIjk/6jYsE7keN4ettg2n1/H0tH9vmVkYTOBWD4SPB8u/4yOswC6qhLVTfFWPC/YlT3yb2xo/r65zCtUh/Psm4THzmfVSjvo2ODERb5MxW9Zvqm3v6pbC5vThH3LuBFwAhyTI1iDCK/Mp0pSUGeTU096JQ==\n",
				},
			},
		},
	}
	skey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDOX0rNUQLB8/tkyAg8exMnA4ledt6/DosUnp5Phxu/A1Mypbyvqo9OgrfjqAW3pAMVXrH1FuAHdP5IGdsPyvP7hd5S43BA4JR0NZMVwKUx11xtE+EGEhWdz8kfTvVVxhvcR1LbiArRdHYfEDKbdd4s2pivqoELKDXZfdn4JDNLuEq/XpcYn53m9z+Q6dz9MLjuUmDqOavc2Ih4xrYmbM6VcN2zHnOBzQjEpCGZM1/o2c3KQzloxOVRp2iFqK4PVau2mL/UaIjHADifeEveWQS18D9KQHP/4ATk+Nj/5JgPirmmZ39vJRzFJNpUsrBrSodvDJrRfbx5sCKQhgs9wZOviCnE4iNqzTefsPs3E1Cf+rgowLZ9kJIS2RzMjwBE3HsIOpvbbdbaGoEfIAvPQG/3c65PwPUGzFVM0QXFvQUdKu3fxQ566t3cCTHiJIiF3/1PKR1pN30CXsJ4xJvARKUddVA9qgAXHLJMQEPx+BlO8q4auI+kVoYETap0obPh+Ouz28FgT4SbOprg070JiIjk/6jYsE7keN4ettg2n1/H0tH9vmVkYTOBWD4SPB8u/4yOswC6qhLVTfFWPC/YlT3yb2xo/r65zCtUh/Psm4THzmfVSjvo2ODERb5MxW9Zvqm3v6pbC5vThH3LuBFwAhyTI1iDCK/Mp0pSUGeTU096JQ==\n"
	sshkey := &v1.SshPublicKey{
		Metadata: &v1.ResourceMetadata{
			CloudAccountId: "iks_user",
			Name:           "584db5cd-87--h15g",
		},
		Spec: &v1.SshPublicKeySpec{
			SshPublicKey: skey,
		},
	}

	sshClient.EXPECT().Search(gomock.Any(), &v1.SshPublicKeySearchRequest{Metadata: &v1.ResourceMetadataSearch{}}).Return(sshkeysearchresponse, nil).AnyTimes()
	sshClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(sshkey, nil).AnyTimes()
	return sshClient
}

func NewMockVnetServiceClient(ctx context.Context, t *testing.T) *v1.MockVNetServiceClient {
	mockController := gomock.NewController(t)
	vnetClient := v1.NewMockVNetServiceClient(mockController)

	vnetGetResponse := &v1.VNet{
		Metadata: &v1.VNet_Metadata{
			CloudAccountId: "iks_user",
			Name:           "us-dev-1a-default",
			ResourceId:     "5e10818b-6430-464e-a763-e311164f0a6b",
		},
		Spec: &v1.VNetSpec{
			Region:           "us-dev-1",
			AvailabilityZone: "us-dev-1a",
			PrefixLength:     24,
		},
	}

	vnetClient.EXPECT().Get(ctx, gomock.Any()).Return(vnetGetResponse, nil)
	return vnetClient
}

func NewMockInstanceServiceClient(
	t *testing.T,
	mockDelMethod bool,
	c codes.Code,
	msg string) *v1.MockInstanceServiceClient {
	instanceServiceClient := v1.NewMockInstanceServiceClient(gomock.NewController(t))
	if mockDelMethod {
		instanceServiceClient.EXPECT().
			Delete(
				gomock.Any(),
				gomock.Any(),
			).
			Return(
				&emptypb.Empty{},
				status.Error(c, msg),
			)
	}

	return instanceServiceClient
}

func NewMockProductCatalogServiceClient(ctx context.Context, t *testing.T) *v1.MockProductCatalogServiceClient {
	mockController := gomock.NewController(t)
	productCatalogClient := v1.NewMockProductCatalogServiceClient(mockController)
	fileProductName := "storage-file"

	m := make(map[string]string)
	// Below sizes are in TB
	m["volume.size.min"] = "1"
	m["volume.size.max"] = "100"

	// define product
	p1 := &v1.Product{
		Name:     fileProductName,
		Metadata: m,
	}

	// Set mocking behavior
	productCatalogClient.EXPECT().AdminRead(context.Background(), gomock.Any()).Return(&v1.ProductResponse{
		Products: []*v1.Product{
			p1,
		},
	}, nil).AnyTimes()

	return productCatalogClient
}
