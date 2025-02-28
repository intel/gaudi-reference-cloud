// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"
	"database/sql"
	"embed"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sesutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/snsutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sqsutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Service struct {
	mdb *manageddb.ManagedDb
	Sql *sql.DB
}

func (svc *Service) Init(ctx context.Context, cfg *config.Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	logger := log.FromContext(ctx).WithName("Service.Init")
	logger.Info("service initialization")
	config.Cfg = cfg
	var err error
	if svc.mdb == nil {
		svc.mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			logger.Error(err, "manageddb.New failed")
			return err
		}
	}
	svc.Sql, err = openDb(ctx, svc.mdb)
	if err != nil {
		logger.Error(err, "openDb failed")
		return err
	}
	awsSESRegion := config.Cfg.GetAWSSESRegion()
	awsCredentialsFile := config.Cfg.GetAWSCredentialsFile()
	awsTopicARN := config.Cfg.GetAWSSNSDefaultTopicArn()
	logger.Info("AWS SES", "awsSESRegion", awsSESRegion, "awsCredentialsFile", awsCredentialsFile)
	snsUtil := snsutil.SNSUtil{}
	if err := snsUtil.Init(ctx, awsSESRegion, awsTopicARN, awsCredentialsFile); err != nil {
		logger.Error(err, "couldn't init ses client")
	}
	sqsUtil := sqsutil.SQSUtil{}
	awsQueueUrl := config.Cfg.GetAWSSQSDefaultQueueUrl()
	if err := sqsUtil.Init(ctx, awsSESRegion, awsQueueUrl, awsCredentialsFile); err != nil {
		logger.Error(err, "couldn't init ses client")
	}
	eventDispatcher := NewEventDispatcher()
	eventData := NewEventData(svc.Sql)
	eventManager := NewEventManager(eventData, eventDispatcher)
	eventPublisher := NewEventPublisher(&snsUtil)
	eventReceiver := NewEventReceiver(&snsUtil, &sqsUtil)
	eventHandler := NewEventHandler(eventData, eventPublisher, eventReceiver)
	notificationGwService := NewNotificationGatewayService(eventManager, eventHandler)
	pb.RegisterNotificationGatewayServiceServer(grpcServer, notificationGwService)
	sesUtil := sesutil.SESUtil{}
	//TODO: IAM changes
	if err := sesUtil.Init(ctx, awsSESRegion, awsCredentialsFile); err != nil {
		logger.Error(err, "couldn't init ses client")
	}
	emailService := NewEmailNotificationService(&sesUtil)
	pb.RegisterEmailNotificationServiceServer(grpcServer, emailService)
	reflection.Register(grpcServer)
	return nil
}

func (*Service) Name() string {
	return "notification-gateway"
}

//var db *sql.DB

//go:embed sql/*.sql
var fs embed.FS

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
