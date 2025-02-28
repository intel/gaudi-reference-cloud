package git_to_grpc_comms_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/instance"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/instance_type"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/ip_resource_manager"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/machine_image"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/ssh_public_key"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/vnet"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type GrpcService struct {
	ManagedDb                   *manageddb.ManagedDb
	VmInstanceSchedulingService pb.InstanceSchedulingServiceClient
	listener                    net.Listener
	grpcServer                  *grpc.Server
	db                          *sql.DB
	errc                        chan error
	cfg                         config.Config
	cloudAccountServiceClient   pb.CloudAccountServiceClient
}

func New(ctx context.Context, cfg *config.Config, managedDb *manageddb.ManagedDb, vmInstanceSchedulingService pb.InstanceSchedulingServiceClient, listener net.Listener) (*GrpcService, error) {
	if cfg.PurgeInstanceInterval == 0*time.Second || cfg.PurgeInstanceAge == 0*time.Second {
		return nil, fmt.Errorf("values of purgeInstanceInterval and purgeInstanceAge should be set")
	}

	return &GrpcService{
		ManagedDb:                   managedDb,
		VmInstanceSchedulingService: vmInstanceSchedulingService,
		listener:                    listener,
		cfg:                         *cfg,
	}, nil
}

// Run service, blocking until an error occurs.
func (s *GrpcService) Run(ctx context.Context) error {
	if err := s.Start(ctx); err != nil {
		return err
	}
	// Wait for ListenAndServe to return, return error.
	return <-s.errc
}

func (s *GrpcService) instancePurgeThread(ctx context.Context) {
	log := log.FromContext(ctx).WithName("GrpcService.instancePurgeThread")
	purgeInstanceAge := s.cfg.PurgeInstanceAge

	// Loop through periodically for every 'PurgeInstanceInterval' of time to purge old instance records from database
	ticker := time.NewTicker(s.cfg.PurgeInstanceInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			numOfInstancesDeleted, err := s.DeleteInstanceRecords(ctx, now, purgeInstanceAge)
			if err != nil {
				log.Error(err, "Error Occured when deleting instance records")
			}

			log.Info("Instances that are purged from database", "numOfInstancesDeleted", numOfInstancesDeleted)
		}
	}
}

func (s *GrpcService) DeleteInstanceRecords(ctx context.Context, currentTime time.Time, purgeInstanceAge time.Duration) (int64, error) {
	// Calculate the time to delete records that are older than 'purgeInstanceAge' from 'currentTime'
	deleteBeforeTime := currentTime.Add(-purgeInstanceAge)
	query := `delete from instance where deleted_timestamp < $1`
	result, err := s.db.ExecContext(ctx, query, deleteBeforeTime)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rowsAffected, err
	}
	return rowsAffected, nil
}

func (s *GrpcService) Start(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("GrpcService.Start")
	log.Info("BEGIN", "ListenAddr", s.listener.Addr().(*net.TCPAddr).Port)
	defer log.Info("END")
	if s.ManagedDb == nil {
		return fmt.Errorf("ManagedDb not provided")
	}
	if err := s.ManagedDb.Migrate(ctx, db.MigrationsFs, db.MigrationsDir); err != nil {
		return err
	}
	db, err := s.ManagedDb.Open(ctx)
	if err != nil {
		return err
	}
	s.db = db

	go s.instancePurgeThread(ctx)

	const ROLE = "populate-inflow-component-git-to-grpc-synchronizer"
	const DOMAIN = "populate-inflow-component-git-to-grpc-synchronizer.idcs-system.svc.cluster.local"
	var result any

	body := GenerateRequestToGetCertificate(DOMAIN, ROLE)
	jsonerr := json.Unmarshal([]byte(body), &result)
	if jsonerr != nil {
		Fail("Error during Unmarshal(): " + err.Error())
	}

	json := GetFieldInfo(body)

	fmt.Print(json)

	serverTLSConf, _, err := VaultCertSetup(json["issuing_ca"].(string), json["private_key"].(string), json["certificate"].(string))

	if err != nil {
		fmt.Print(err.Error())
	}

	test_config := credentials.NewTLS(serverTLSConf)

	fmt.Print(test_config)

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	}

	s.grpcServer, err = grpc.NewServer(serverOptions...), nil
	if err != nil {
		return err
	}

	sshPublicKeyService, err := ssh_public_key.NewSshPublicKeyService(db)
	if err != nil {
		return err
	}
	instanceTypeService, err := instance_type.NewInstanceTypeService(db)
	if err != nil {
		return err
	}
	machineImageService, err := machine_image.NewMachineImageService(db)
	if err != nil {
		return err
	}
	ipResourceManagerService, err := ip_resource_manager.NewIpResourceManagerService(db)
	if err != nil {
		return err
	}

	objectStorageServicePrivateClient := pb.NewMockObjectStorageServicePrivateClient(gomock.NewController(GinkgoT()))
	Expect(objectStorageServicePrivateClient).ShouldNot(BeNil())
	vNetService, err := vnet.NewVNetService(db, ipResourceManagerService, objectStorageServicePrivateClient, s.cfg.VNetService)
	if err != nil {
		return err
	}
	instanceService, err := instance.NewInstanceService(db, sshPublicKeyService, instanceTypeService, machineImageService,
		vNetService, s.VmInstanceSchedulingService, s.cfg, s.cloudAccountServiceClient, nil)
	if err != nil {
		return err
	}

	pb.RegisterSshPublicKeyServiceServer(s.grpcServer, sshPublicKeyService)
	pb.RegisterInstanceServiceServer(s.grpcServer, instanceService)
	pb.RegisterInstancePrivateServiceServer(s.grpcServer, instanceService)
	pb.RegisterInstanceTypeServiceServer(s.grpcServer, instanceTypeService)
	pb.RegisterMachineImageServiceServer(s.grpcServer, machineImageService)
	pb.RegisterIpResourceManagerServiceServer(s.grpcServer, ipResourceManagerService)
	pb.RegisterVNetServiceServer(s.grpcServer, vNetService)
	pb.RegisterVNetPrivateServiceServer(s.grpcServer, vNetService)
	reflection.Register(s.grpcServer)
	fmt.Print("LISTENER...", s.listener.Addr())
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}
	s.errc = make(chan error, 1)
	go func() {
		s.errc <- s.grpcServer.Serve(s.listener)
		close(s.errc)
	}()

	return nil
}

func (s *GrpcService) Stop(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("GrpcService.Stop")
	log.Info("BEGIN")
	defer log.Info("END")
	// Stop immediately. Do not wait for existing RPCs because streaming RPCs may never end.
	if s != nil && s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	return nil
}

func NewServer(ctx context.Context, serverOptions ...grpc.ServerOption) (*grpc.Server, error) {
	logger := log.FromContext(ctx)
	var creds credentials.TransportCredentials
	UseTLS := false
	logger.Info("Server mTLS status", "UseTLS", UseTLS)
	if UseTLS {
		var err error
		if err != nil {
			logger.Error(err, "can't load TLS credentials for server")
			return nil, err
		}
	} else {
		creds = insecure.NewCredentials()
	}

	serverOptions = append(serverOptions, grpc.Creds(creds))
	return grpc.NewServer(serverOptions...), nil
}
