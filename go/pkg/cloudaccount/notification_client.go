// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type NotificationClient struct {
	emailNotificationServiceClient pb.EmailNotificationServiceClient
	cfg                            *config.Config
}

func NewNotificationClient(ctx context.Context, resolver grpcutil.Resolver, cfg *config.Config) (*NotificationClient, error) {
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
	return &NotificationClient{
		emailNotificationServiceClient: pb.EmailNotificationServiceClient(emailNotificationServiceClient),
		cfg:                            cfg,
	}, nil
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

func (notificationClient *NotificationClient) SendInvitationExpiredEmail(ctx context.Context, messageType string, cloudAcctEmail string, memberEmail string, templateName string) error {
	logger := log.FromContext(ctx).WithName("NotificationClient.SendInvitationExpiredEmail")

	senderEmail := notificationClient.cfg.GetSenderEmail()

	templateData := map[string]string{
		"memberEmail": memberEmail,
	}

	emailRequest := &pb.EmailRequest{
		MessageType:  messageType,
		ServiceName:  "cloudaccount",
		Recipient:    cloudAcctEmail,
		Sender:       senderEmail,
		TemplateName: templateName,
		TemplateData: templateData,
	}

	if _, err := notificationClient.SendEmailNotification(ctx, emailRequest); err != nil {
		logger.Error(err, "couldn't send email")
		return err
	}

	return nil
}

func (notificationClient *NotificationClient) GetOTPEmailRequest(ctx context.Context, otpReq *pb.Otp) (*pb.EmailRequest, error) {
	logger := log.FromContext(ctx).WithName("NotificationClient.GetOTPEmailRequest")
	logger.Info("get email request data")
	templateData := map[string]string{
		"OTP": otpReq.OtpCode,
	}
	cloudAccount, err := GetCloudAccount(ctx, otpReq.CloudAccountId)
	if err != nil {
		logger.Error(err, "unable to get cloud account id", "cloudAccountId", otpReq.CloudAccountId)
		return nil, err
	}
	emailRequest := &pb.EmailRequest{
		MessageType:  "OTP",
		ServiceName:  "OtpService",
		Recipient:    cloudAccount.GetOwner(),
		Sender:       config.Cfg.GetSenderEmail(),
		TemplateName: config.Cfg.GetOTPTemplate(),
		TemplateData: templateData,
	}
	return emailRequest, nil
}
