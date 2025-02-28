// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"fmt"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudcredits_worker/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type NotificationGatewayClient struct {
	NotificationGatewayServiceClient pb.NotificationGatewayServiceClient
	emailNotificationServiceClient   pb.EmailNotificationServiceClient
}

func NewNotificationGatewayClient(ctx context.Context, resolver grpcutil.Resolver) (*NotificationGatewayClient, error) {
	logger := log.FromContext(ctx).WithName("NotificationClient.NewNotificationGatewayClient")
	addr, err := resolver.Resolve(ctx, "notification-gateway")
	if err != nil {
		logger.Error(err, "grpc resolver not able to connect", "addr", addr)
		return nil, err
	}
	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		return nil, err
	}
	notificationGatewayServiceClient := pb.NewNotificationGatewayServiceClient(conn)
	emailNotificationServiceClient := pb.NewEmailNotificationServiceClient(conn)
	return &NotificationGatewayClient{NotificationGatewayServiceClient: notificationGatewayServiceClient, emailNotificationServiceClient: emailNotificationServiceClient}, nil
}

func (notificationGatewayClient *NotificationGatewayClient) PublishEvent(ctx context.Context, req *pb.PublishEventRequest) (*pb.PublishEventResponse, error) {
	logger := log.FromContext(ctx).WithName("NotificationGatewayClient.PublishEvent")
	logger.Info("publish event")

	return notificationGatewayClient.NotificationGatewayServiceClient.PublishEvent(ctx, req)
}

func (notificationGatewayClient *NotificationGatewayClient) SubscribeEvents(ctx context.Context, req *pb.SubscribeEventRequest) (*pb.SubscribeEventResponse, error) {
	logger := log.FromContext(ctx).WithName("NotificationGatewayClient.Subscribe")
	logger.Info("Subscribe events")

	return notificationGatewayClient.NotificationGatewayServiceClient.SubscribeEvents(ctx, req)
}

func (notificationGatewayClient *NotificationGatewayClient) ReceiveEvents(ctx context.Context, req *pb.ReceiveEventRequest) (*pb.ReceiveEventResponse, error) {
	logger := log.FromContext(ctx).WithName("NotificationGatewayClient.ReceiveEvents")
	logger.Info("Receive events")

	return notificationGatewayClient.NotificationGatewayServiceClient.ReceiveEvents(ctx, req)
}

func (notificationGatewayClient *NotificationGatewayClient) SendEmailNotification(ctx context.Context, req *pb.EmailRequest) (*pb.EmailResponse, error) {
	logger := log.FromContext(ctx).WithName("NotificationClient.SendEmailNotification")
	logger.Info("send email notification")
	resp, err := notificationGatewayClient.emailNotificationServiceClient.SendUserEmail(ctx, req)
	if err != nil {
		logger.Error(err, "unable to send email", "Service", req.ServiceName, "TemplateName", req.TemplateName)
		return nil, err
	}
	return resp, nil
}

func (notificationGatewayClient *NotificationGatewayClient) SendCloudCreditUsageEmail(ctx context.Context, messageType string, cloudAcctEmail string, templateName string, serviceName string) error {
	logger := log.FromContext(ctx).WithName("notificationGatewayClient.SendCloudCreditUsageEmail")
	logger.Info("Send email", "cloudaccount", cloudAcctEmail)

	senderEmail := config.Cfg.GetSenderEmail()
	templateData := map[string]string{
		"consoleUrl": config.Cfg.GetConsoleUrl(),
		"paymentUrl": config.Cfg.GetPaymentUrl(),
	}
	emailRequest := &pb.EmailRequest{
		MessageType:  messageType,
		ServiceName:  serviceName,
		Recipient:    cloudAcctEmail,
		Sender:       senderEmail,
		TemplateName: templateName,
		TemplateData: templateData,
	}
	if _, err := notificationGatewayClient.SendEmailNotification(ctx, emailRequest); err != nil {
		logger.Error(err, "couldn't send email")
		return err
	}
	return nil
}

func (notificationGatewayClient *NotificationGatewayClient) SendEmailWithOptions(ctx context.Context, req *pb.SendEmailRequest) (*pb.EmailResponse, error) {
	logger := log.FromContext(ctx).WithName("notificationGatewayClient.SendEmailWithOptions")
	logger.Info("send email notification with options")
	resp, err := notificationGatewayClient.emailNotificationServiceClient.SendEmail(ctx, req)
	if err != nil {
		logger.Error(err, "unable to send email with options", "Service", req.ServiceName, "TemplateName", req.TemplateName)
		return nil, err
	}
	return resp, nil
}

func (notificationGatewayClient *NotificationGatewayClient) SendCloudCreditUsageEmailWithOptions(ctx context.Context, messageType, cloudAcctEmail, templateName, serviceName string) error {
	logger := log.FromContext(ctx).WithName("notificationGatewayClient.SendCloudCreditUsageEmailWithOptions")

	if strings.TrimSpace(config.Cfg.GetSenderEmail()) == "" || strings.TrimSpace(cloudAcctEmail) == "" || strings.TrimSpace(templateName) == "" {
		return fmt.Errorf("invalid email config params")
	}
	senderEmail := config.Cfg.GetSenderEmail()
	templateData := map[string]string{
		"consoleUrl": config.Cfg.GetConsoleUrl(),
		"paymentUrl": config.Cfg.GetPaymentUrl(),
	}
	emailRequest := &pb.SendEmailRequest{
		MessageType:  messageType,
		ServiceName:  serviceName,
		Recipient:    []string{cloudAcctEmail},
		CcRecipients: []string{config.Cfg.GetOpsPDL()},
		Sender:       senderEmail,
		TemplateName: templateName,
		TemplateData: templateData,
	}
	if _, err := notificationGatewayClient.SendEmailWithOptions(ctx, emailRequest); err != nil {
		logger.Error(err, "couldn't send email")
		return err
	}
	return nil
}
