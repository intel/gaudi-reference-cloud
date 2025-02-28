// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/instance"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/instance_group"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/instance_type"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/ip_resource_manager"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/loadbalancer"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/machine_image"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/ssh_public_key"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/vnet"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcService struct {
	ManagedDb                          *manageddb.ManagedDb
	VmInstanceSchedulingService        pb.InstanceSchedulingServiceClient
	BillingDeactivateInstancesService  pb.BillingDeactivateInstancesServiceClient
	InstanceService                    *instance.InstanceService
	LoadBalancerService                *loadbalancer.Service
	listener                           net.Listener
	grpcServer                         *grpc.Server
	db                                 *sql.DB
	errc                               chan error
	cfg                                config.Config
	cloudAccountServiceClient          pb.CloudAccountServiceClient
	cloudAccountAppClientServiceClient pb.CloudAccountAppClientServiceClient
	// objectStorageServicePrivateClient may be nil
	objectStorageServicePrivateClient pb.ObjectStorageServicePrivateClient
	// fleetAdminServiceClient may be nil
	fleetAdminServiceClient pb.FleetAdminServiceClient
	qmsClient               pb.QuotaManagementPrivateServiceClient
}

func New(
	ctx context.Context,
	cfg *config.Config,
	managedDb *manageddb.ManagedDb,
	vmInstanceSchedulingService pb.InstanceSchedulingServiceClient,
	billingDeactivateInstancesService pb.BillingDeactivateInstancesServiceClient,
	cloudAccountServiceClient pb.CloudAccountServiceClient,
	cloudAccountAppClientServiceClient pb.CloudAccountAppClientServiceClient,
	objectStorageServicePrivateClient pb.ObjectStorageServicePrivateClient,
	fleetAdminServiceClient pb.FleetAdminServiceClient,
	qmsClient pb.QuotaManagementPrivateServiceClient,
	listener net.Listener,
) (*GrpcService, error) {
	if cfg.PurgeInstanceInterval == 0*time.Second || cfg.PurgeInstanceAge == 0*time.Second || cfg.GetDeactivateInstancesInterval == 0*time.Second {
		return nil, fmt.Errorf("values of purgeInstanceInterval, purgeInstanceAge and getDeactivateInstancesInterval should be set")
	}

	return &GrpcService{
		ManagedDb:                          managedDb,
		VmInstanceSchedulingService:        vmInstanceSchedulingService,
		BillingDeactivateInstancesService:  billingDeactivateInstancesService,
		listener:                           listener,
		cfg:                                *cfg,
		cloudAccountServiceClient:          cloudAccountServiceClient,
		cloudAccountAppClientServiceClient: cloudAccountAppClientServiceClient,
		objectStorageServicePrivateClient:  objectStorageServicePrivateClient,
		fleetAdminServiceClient:            fleetAdminServiceClient,
		qmsClient:                          qmsClient,
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
	log := log.FromContext(ctx).WithName("GrpcService.DeleteInstanceRecords")
	// Calculate the time to delete records that are older than 'purgeInstanceAge' from 'currentTime'
	deleteBeforeTime := currentTime.Add(-purgeInstanceAge)

	// Limit records purged to known working number. Any records not purged this time will be purged at the next
	// PurgeInstanceInterval. At the current interval of 1 hour, the limit of 3600 represents an average delete
	// of 1 instance per second. The "Instance Delete many records" unit test may be used to test increased values.
	query := `select resource_id,cloud_account_id from instance where deleted_timestamp < $1 limit 3600`
	args := []any{deleteBeforeTime}
	log.Info("Executing query", "query", query, "args", args)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var resourceId, cloudAccountId string
	var resourceIds []string
	for rows.Next() {
		if err := rows.Scan(&resourceId, &cloudAccountId); err != nil {
			return 0, err
		}
		log.Info("Deleting instance record", logkeys.ResourceId, resourceId, logkeys.CloudAccountId, cloudAccountId)
		resourceIds = append(resourceIds, resourceId)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	query = `delete from instance where resource_id = any($1)`
	args = []any{resourceIds}
	log.Info("Executing query", "query", query, "args", args)
	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rowsAffected, err
	}
	return rowsAffected, nil
}

func (s *GrpcService) deleteDeactivatedInstancesThread(ctx context.Context) {
	log := log.FromContext(ctx).WithName("GrpcService.deleteDeactivatedInstancesThread")

	// Loop through periodically for every 'GetDeactivateInstancesInterval' of time to obtain the instance types that needs to be deleted for each CloudAccountId
	ticker := time.NewTicker(s.cfg.GetDeactivateInstancesInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			numOfDeactivatedInstancesDeleted, err := s.DeletedDeactivatedInstances(ctx)
			if err != nil {
				log.Error(err, "Error occured when deleting deactivated instances")
				continue
			}
			log.Info("Deactivated instances that are deleted from database", "numOfDeactivatedInstancesDeleted", numOfDeactivatedInstancesDeleted)
		}
	}
}
func (s *GrpcService) DeletedDeactivatedInstances(ctx context.Context) (int64, error) {
	var rowsAffected int64
	log := log.FromContext(ctx).WithName("GrpcService.DeletedDeactivatedInstances")
	deactivatedInstanceStream, err := s.BillingDeactivateInstancesService.GetDeactivateInstancesStream(ctx, &emptypb.Empty{})

	if err != nil {
		return 0, err
	}
	log.Info("DeletedDeactivatedInstances stream started")
	query := `
		select resource_id, name, resource_version from instance
		where cloud_account_id = $1
		and  value->'spec'->>'instanceType' = $2
		and  deleted_timestamp = $3
	`
	for {
		listOfDeactivatedInstances, err := deactivatedInstanceStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error(err, "failed to recv deactivated instance data from billing")
			return 0, err
		}
		for _, deactivatedInstance := range listOfDeactivatedInstances.DeactivationList {
			cloudAccountId := deactivatedInstance.CloudAccountId
			log.Info("Receiving Stream", "CloudaccountId", cloudAccountId)
			for _, quota := range deactivatedInstance.Quotas {
				instanceType := quota.InstanceType
				// run a select query to find the instance records that needs to be deleted
				rows, err := s.db.QueryContext(ctx, query, cloudAccountId, instanceType, "infinity")
				if err != nil {
					return 0, err
				}
				defer rows.Close()

				var rowCount int64
				var resourceId string
				var name string
				var resourceVersion string
				for rows.Next() {
					if err := rows.Scan(&resourceId, &name, &resourceVersion); err != nil {
						return 0, err
					}
					// delete the instance
					log.Info("Deleting Deactivated Instance", "instanceName", name, logkeys.CloudAccountId, cloudAccountId, logkeys.ResourceId, resourceId)
					err := s.InstanceService.DeleteInstanceInternal(ctx, cloudAccountId, resourceId, name, resourceVersion)
					if err != nil {
						// continue with next row or next cloudAccount
						log.Error(err, "Error occured when deleting deactivated instance", "name", name, logkeys.CloudAccountId, cloudAccountId, logkeys.ResourceId, resourceId)
						continue
					}
					rowCount++
				}
				if err := rows.Err(); err != nil {
					return 0, err
				}
				rowsAffected += rowCount
			}
		}
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
	// Update the maximum idle connection count.
	// The default MaxIdle connections is set to 20. The reason for this change are twofold
	// a. The performance impact of establishing a new connection for bursty requests is high.
	// b. The pprof heap dump indicated that a cum and flat values referenced by pgx.connect was high.
	// 	  Increasing the idle connection count reduces this value in the heap.
	log.Info("Setting database MaxIdleConnections configuration", "dbMaxIdleConnectionCount", s.cfg.DbMaxIdleConnectionCount)
	db.SetMaxIdleConns(int(s.cfg.DbMaxIdleConnectionCount))
	if err != nil {
		return err
	}
	s.db = db

	go s.instancePurgeThread(ctx)
	go s.deleteDeactivatedInstancesThread(ctx)

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    120 * time.Second, // The time a connection is kept alive without any activity.
			Timeout: 20 * time.Second,  // Maximum time the server waits for activity before closing the connection.
		}),
	}

	s.grpcServer, err = grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return err
	}

	sshPublicKeyService, err := ssh_public_key.NewSshPublicKeyService(s.cloudAccountAppClientServiceClient, db)
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
	vNetService, err := vnet.NewVNetService(db, ipResourceManagerService, s.objectStorageServicePrivateClient, s.cfg.VNetService)
	if err != nil {
		return err
	}
	instanceService, err := instance.NewInstanceService(db, sshPublicKeyService, instanceTypeService, machineImageService,
		vNetService, s.VmInstanceSchedulingService, s.cfg, s.cloudAccountServiceClient, s.fleetAdminServiceClient, s.qmsClient)
	if err != nil {
		return err
	}
	s.InstanceService = instanceService

	instanceGroupService, err := instance_group.NewInstanceGroupService(instanceService)
	if err != nil {
		return err
	}
	loadBalancerService, err := loadbalancer.NewLoadBalancerService(db, s.cfg, instanceService)
	if err != nil {
		return err
	}
	s.LoadBalancerService = loadBalancerService

	pb.RegisterSshPublicKeyServiceServer(s.grpcServer, sshPublicKeyService)
	pb.RegisterInstanceServiceServer(s.grpcServer, instanceService)
	pb.RegisterInstancePrivateServiceServer(s.grpcServer, instanceService)
	pb.RegisterInstanceTypeServiceServer(s.grpcServer, instanceTypeService)
	pb.RegisterMachineImageServiceServer(s.grpcServer, machineImageService)
	pb.RegisterIpResourceManagerServiceServer(s.grpcServer, ipResourceManagerService)
	pb.RegisterVNetServiceServer(s.grpcServer, vNetService)
	pb.RegisterVNetPrivateServiceServer(s.grpcServer, vNetService)
	pb.RegisterInstanceGroupServiceServer(s.grpcServer, instanceGroupService)
	pb.RegisterInstanceGroupPrivateServiceServer(s.grpcServer, instanceGroupService)
	pb.RegisterLoadBalancerServiceServer(s.grpcServer, loadBalancerService)
	pb.RegisterLoadBalancerPrivateServiceServer(s.grpcServer, loadBalancerService)
	reflection.Register(s.grpcServer)
	log.Info("Service running")
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
