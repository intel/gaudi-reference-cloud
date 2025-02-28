// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type NotificationClient struct {
	emailNotificationServiceClient pb.EmailNotificationServiceClient
}

func NewNotificationClient(ctx context.Context, resolver grpcutil.Resolver) (*NotificationClient, error) {
	logger := log.FromContext(ctx).WithName("NotificationClient.InitNotificationClient")
	addr, err := resolver.Resolve(ctx, "notification-gateway")
	if err != nil {
		logger.Error(err, "grpc resolver not able to connect", "addr", addr)
		return nil, err
	}
	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		return nil, err
	}
	emailNotificationServiceClient := pb.NewEmailNotificationServiceClient(conn)
	return &NotificationClient{emailNotificationServiceClient: pb.EmailNotificationServiceClient(emailNotificationServiceClient)}, nil
}

func (notificationClient *NotificationClient) SendEmailNotification(ctx context.Context, req *pb.EmailRequest) (*pb.EmailResponse, error) {
	logger := log.FromContext(ctx).WithName("NotificationClient.SendEmailNotification")
	logger.Info("send email notification")
	resp, err := notificationClient.emailNotificationServiceClient.SendUserEmail(ctx, req)
	if err != nil {
		logger.Error(err, "unable to send email", "Service", req.ServiceName, "TemplateName", req.TemplateName)
		return nil, err
	}
	return resp, nil
}

func (notificationClient *NotificationClient) SendCloudCreditUsageEmail(ctx context.Context, messageType string, cloudAcctEmail string, templateName string) error {
	logger := log.FromContext(ctx).WithName("SendCloudCreditUsageEmail")
	logger.Info("send email for cloud credit usage")
	//TODO: Make all these variables configurables
	senderEmail := Cfg.GetSenderEmail()
	templateData := map[string]string{
		"consoleUrl": Cfg.GetConsoleUrl(),
		"paymentUrl": Cfg.GetPaymentUrl(),
	}
	emailRequest := &pb.EmailRequest{
		MessageType:  messageType,
		ServiceName:  "billing",
		Recipient:    cloudAcctEmail,
		Sender:       senderEmail,
		TemplateName: templateName,
		TemplateData: templateData,
	}
	logger.Info("sending email")
	if _, err := notificationClient.SendEmailNotification(ctx, emailRequest); err != nil {
		logger.Error(err, "couldn't send email")
		return err
	}
	return nil
}

func (notificationClient *NotificationClient) SendEmailWithOptions(ctx context.Context, req *pb.SendEmailRequest) (*pb.EmailResponse, error) {
	logger := log.FromContext(ctx).WithName("NotificationClient.SendEmailWithOptions")
	logger.Info("send email notification with options")
	resp, err := notificationClient.emailNotificationServiceClient.SendEmail(ctx, req)
	if err != nil {
		logger.Error(err, "unable to send email with options", "Service", req.ServiceName, "TemplateName", req.TemplateName)
		return nil, err
	}
	return resp, nil
}

func (notificationClient *NotificationClient) SendCloudCreditUsageEmailWithOptions(ctx context.Context, messageType string, cloudAcctEmail string, templateName string) error {
	logger := log.FromContext(ctx).WithName("SendCloudCreditUsageEmail")
	logger.Info("send email for cloud credit usage")
	//TODO: Make all these variables configurables
	senderEmail := Cfg.GetSenderEmail()
	templateData := map[string]string{
		"consoleUrl": Cfg.GetConsoleUrl(),
		"paymentUrl": Cfg.GetPaymentUrl(),
	}
	emailRequest := &pb.SendEmailRequest{
		MessageType:  messageType,
		ServiceName:  "billing",
		Recipient:    []string{cloudAcctEmail},
		CcRecipients: []string{Cfg.GetPDLEmail()},
		Sender:       senderEmail,
		TemplateName: templateName,
		TemplateData: templateData,
	}
	logger.Info("sending email")
	if _, err := notificationClient.SendEmailWithOptions(ctx, emailRequest); err != nil {
		logger.Error(err, "couldn't send email")
		return err
	}
	return nil
}
