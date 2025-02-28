// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package productcatalog

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog/handlers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog/sync"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog_operator/apis/private.cloud/v1alpha1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type GrpcService struct {
	// Format should be ":30002"
	ListenAddr       string
	Config           *config.Config
	CloudAccountDb   *manageddb.ManagedDb
	ProductCatalogDb *manageddb.ManagedDb
	grpcServer       *grpc.Server
	kubeConfig       *rest.Config
}

// TODO: update to the default service startup procedure
func (s *GrpcService) Start(ctx context.Context, defaultRestConfig *rest.Config) error {
	log := log.FromContext(ctx).WithName("GrpcService.Start")
	log.Info("BEGIN", "ListenAddr", s.ListenAddr)
	defer log.Info("END")

	if s.CloudAccountDb == nil {
		return fmt.Errorf("cloudAccountDb not provided")
	}
	cloudAccountDb, err := s.CloudAccountDb.Open(ctx)
	if err != nil {
		return err
	}

	s.kubeConfig = defaultRestConfig

	// Create k8s client to get CRDs
	restClient, err := initKubeRestClient(s.kubeConfig)
	if err != nil {
		return err
	}

	// Create cloudaccount client
	cloudAcctClient, err := initCloudAccountClient(ctx)
	if err != nil {
		return err
	}

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
		log.Error(err, "can't start secure server: ", "err", err)
		return err
	}

	productAccessService, err := handlers.NewProductAccessService(cloudAccountDb)
	if err != nil {
		return err
	}
	pb.RegisterProductAccessServiceServer(s.grpcServer, productAccessService)

	productVendorService, err := handlers.NewProductVendorService(restClient)
	if err != nil {
		return err
	}
	pb.RegisterProductVendorServiceServer(s.grpcServer, productVendorService)

	// register the product service
	productCatalogService, err := handlers.NewProductCatalogService(restClient, cloudAcctClient, cloudAccountDb)
	if err != nil {
		log.Error(err, "unable to create product catalog service")
		return err
	}
	pb.RegisterProductCatalogServiceServer(s.grpcServer, productCatalogService)

	// Product Catalog v2 Service Initialization starts
	var productCatalogDb *sql.DB

	if s.ProductCatalogDb != nil {
		productCatalogDb, err = openDb(ctx, s.ProductCatalogDb)
		if err != nil {
			log.Error(err, "unable to resolve productCatalogDb address")
			return err
		}

		// Region Repository
		regionRepository, err := handlers.NewRegionRepository(productCatalogDb)
		if err != nil {
			log.Error(err, "error initializing region repository")
			return err
		}

		// register the region service
		regionService, err := handlers.NewRegionService(regionRepository, cloudAcctClient)
		if err != nil {
			log.Error(err, "unable to resolve region service address")
			return err
		}
		pb.RegisterRegionServiceServer(s.grpcServer, regionService)

		productSyncService, err := sync.NewProductSyncService(restClient, productCatalogDb, s.Config.DefaultRegions)
		if err != nil {
			log.Error(err, "unable to create product sync service")
			return err
		}
		pb.RegisterProductSyncServiceServer(s.grpcServer, productSyncService)

		// Permission Repository
		regionAccessRepository, err := handlers.NewRegionAccessRepository(productCatalogDb)
		if err != nil {
			log.Error(err, "error initializing region access repository")
			return err
		}

		regionAccessService, err := handlers.NewRegionAccessService(regionAccessRepository, cloudAcctClient)
		if err != nil {
			log.Error(err, "unable to create cloud account region access service")
			return err
		}
		pb.RegisterRegionAccessServiceServer(s.grpcServer, regionAccessService)

		log.Info("product catalog v2 services enabled")
	}

	// Product Catalog v2 Service Initialization ends

	// Register reflection service on gRPC server.
	// required to test with grpc clients
	reflection.Register(s.grpcServer)

	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}
	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			log.Error(err, "serve failed")
		}
	}()

	return nil
}

func initKubeRestClient(kubeConfig *rest.Config) (*rest.RESTClient, error) {
	err := v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("error adding to scheme: %v", err)
	}

	kubeConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupName, Version: "v1alpha1"}
	kubeConfig.APIPath = "/apis"
	kubeConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	kubeConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	restClient, err := rest.UnversionedRESTClientFor(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating kube client: %v", err)
	}
	return restClient, nil
}

func initCloudAccountClient(ctx context.Context) (pb.CloudAccountServiceClient, error) {
	resolver := &grpcutil.DnsResolver{}
	addr, err := resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		return nil, fmt.Errorf("error resolving cloudaccount client: %v", err)
	}

	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("error creating grpc client: %v", err)
	}
	return pb.NewCloudAccountServiceClient(conn), nil
}

func (s *GrpcService) Stop(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("GrpcService.Stop")
	log.Info("BEGIN")
	defer log.Info("END")
	s.grpcServer.GracefulStop()
	return nil
}

//go:embed sql/*.sql
var fs embed.FS

// Open service database and run migrations.
func openDb(ctx context.Context, mdb *manageddb.ManagedDb) (*sql.DB, error) {
	log := log.FromContext(ctx)
	if err := mdb.Migrate(ctx, fs, "sql"); err != nil {
		log.Error(err, "migrate:")
		return nil, err
	}

	var err error
	db, err := mdb.Open(ctx)
	if err != nil {
		log.Error(err, "mdb.Open failed")
		return nil, err
	}
	return db, nil
}
