// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudmonitor/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudmonitor/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	grpcServer *grpc.Server
	testDb     *manageddb.TestDb
	mDB        *manageddb.ManagedDb
	sqlDb      *sql.DB
	cfg        config.Config
)

//go:embed migrations/*.sql
var fs embed.FS
var dir string = "migrations"

type Test struct {
	clientConn *grpc.ClientConn
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
	var err error
	testDb = &manageddb.TestDb{}
	mDB, err = testDb.Start(ctx)
	if err != nil {
		log.Error(err, "error starting database")
		return
	}

	err = mDB.Migrate(ctx, fs, dir)
	if err != nil {
		log.Error(err, "error migrating database ")
		return
	}

	sqlDb, err = mDB.Open(ctx)
	if err != nil {
		log.Error(err, "error opening database ")
		return
	}

	cfg := config.Config{
		VictoriaMetricsAddr: "https://internal-placeholder.com/select/38/prometheus/api/v1/query_range",
	}

	// GRPC SERVER AND cloudmonitor SERVICES
	grpcServer = grpc.NewServer()
	cloudmonitorSrv, err := server.NewCloudMonitorService(sqlDb, cfg)
	if err != nil {
		log.Error(err, "error initializing cloudmonitor service")
		return
	}

	v1.RegisterCloudMonitorServiceServer(grpcServer, cloudmonitorSrv)
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
