// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sesutil"
)

type EmailNotificationService struct {
	pb.UnimplementedEmailNotificationServiceServer
	sesUtil *sesutil.SESUtil
}

func NewEmailNotificationService(sesUtil *sesutil.SESUtil) *EmailNotificationService {
	return &EmailNotificationService{sesUtil: sesUtil}
}

func (ens *EmailNotificationService) SendUserEmail(ctx context.Context, emailRequest *pb.EmailRequest) (*pb.EmailResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EmailNotificationService.SendUserEmail").Start()
	defer span.End()
	logger.Info("send email api invoked", "templateName", emailRequest.TemplateName, "templateData", emailRequest.TemplateData)
	messageId, err := ens.sesUtil.SendEmail(ctx, emailRequest.Recipient, emailRequest.Sender, emailRequest.TemplateName, emailRequest.TemplateData)
	if err != nil {
		logger.Error(err, "failed to send email")
		return &pb.EmailResponse{MessageId: messageId, Success: false}, err
	}
	return &pb.EmailResponse{MessageId: messageId, Success: true}, nil
}

func (ens *EmailNotificationService) SendEmail(ctx context.Context, emailRequest *pb.SendEmailRequest) (*pb.EmailResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).Start()
	defer span.End()
	logger.Info("send email api invoked", "templateName", emailRequest.TemplateName, "templateData", emailRequest.TemplateData)
	messageId, err := ens.sesUtil.SendEmailWithOptions(ctx, emailRequest.Recipient, emailRequest.CcRecipients, emailRequest.BccRecipients, emailRequest.Sender, emailRequest.TemplateName, emailRequest.TemplateData)
	if err != nil {
		logger.Error(err, "failed to send email")
		return &pb.EmailResponse{MessageId: messageId, Success: false}, err
	}
	return &pb.EmailResponse{MessageId: messageId, Success: true}, nil
}
