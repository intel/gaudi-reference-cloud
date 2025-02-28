// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package sesutil

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type SESUtil struct {
	sesClient *ses.SES
}

func (sesUtil *SESUtil) Init(ctx context.Context, region string, credentialsFile string) error {
	logger := log.FromContext(ctx).WithName("SESUtil.Init")
	logger.Info("Init BEGIN")
	defer logger.Info("Init END")
	logger.Info("ses init", "region", region, "credentialsFile", credentialsFile)
	sesClient, err := sesUtil.CreateSESClient(ctx, region, credentialsFile)
	if err != nil {
		logger.Error(err, "failed to create ses client with session")
		return err
	}
	sesUtil.sesClient = sesClient
	return nil
}

func (sesUtil *SESUtil) CreateSESClient(ctx context.Context, region string, credentialsFile string) (*ses.SES, error) {
	logger := log.FromContext(ctx).WithName("SESUtil.CreateSESClient")
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("ses init", "region", region, "credentialsFile", credentialsFile)
	var sess *session.Session
	var err error
	if credentialsFile != "" {
		options := session.Options{
			Config: aws.Config{
				Region: aws.String(region),
			},
			SharedConfigState: session.SharedConfigEnable,
		}
		options.SharedConfigFiles = []string{credentialsFile}
		sess, err = session.NewSessionWithOptions(options)
		if err != nil {
			return nil, err
		}
	} else {
		sess, err = session.NewSession(&aws.Config{
			Region: aws.String("us-west-2"),
		})
		if err != nil {
			return nil, err
		}
	}
	sesClient := ses.New(sess)
	return sesClient, nil
}

func (sesUtil *SESUtil) GetTemplate(ctx context.Context, templateName string) (*ses.Template, error) {
	logger := log.FromContext(ctx).WithName("SESUtil.GetTemplate")
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("template", "templateName", templateName)
	emailTemplateinput := &ses.GetTemplateInput{
		TemplateName: aws.String(templateName),
	}

	template, err := sesUtil.sesClient.GetTemplate(emailTemplateinput)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %v", err)
	}
	logger.V(1).Info("retrieved template", "subject", *template.Template.SubjectPart, "HTML Body", *template.Template.HtmlPart, "Text Body", *template.Template.TextPart)
	return template.Template, nil
}

func (sesUtil *SESUtil) SendEmail(ctx context.Context, destinationEmail string, sourceEmail string, templateName string, templateData map[string]string) (string, error) {
	logger := log.FromContext(ctx).WithName("SESUtil.SendEmail")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	emailInput := sesUtil.GetEmailInput(ctx, destinationEmail, sourceEmail, templateName, templateData)
	msgId, err := sesUtil.SendEmailWithTemplate(ctx, emailInput)
	if err != nil {
		logger.Error(err, "error sending email")
		return "", err
	}
	return msgId, nil

}

func (sesUtil *SESUtil) SendEmailWithTemplate(ctx context.Context, emailInput *ses.SendTemplatedEmailInput) (string, error) {
	logger := log.FromContext(ctx).WithName("SESUtil.SendEmailWithTemplate")
	logger.Info("BEGIN")
	defer logger.Info("END")

	// Send the email
	result, err := sesUtil.sesClient.SendTemplatedEmail(emailInput)
	if err != nil {
		logger.Error(err, "error sending email")
		return "", err
	}
	logger.Info("email sent", "messageId", *result.MessageId)
	return *result.MessageId, nil
}

func (sesUtil *SESUtil) GetEmailInput(ctx context.Context, destinationEmail string, sourceEmail string, templateName string, templateData map[string]string) *ses.SendTemplatedEmailInput {
	logger := log.FromContext(ctx).WithName("SESUtil.GetEmailInput")
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("ses send email")
	destination := &ses.Destination{
		ToAddresses: []*string{aws.String(destinationEmail)},
	}
	emailInput := &ses.SendTemplatedEmailInput{
		Destination:  destination,
		Template:     aws.String(templateName),
		Source:       aws.String(sourceEmail),
		TemplateData: aws.String(mapToString(templateData)),
	}
	return emailInput
}

func mapToString(data map[string]string) string {
	var result string
	for key, value := range data {
		result += fmt.Sprintf(`"%s":"%s",`, key, value)
	}
	return fmt.Sprintf("{%s}", result[:len(result)-1])
}

func toStringPtrSlice(addresses []string) []*string {
	addressesPtr := make([]*string, len(addresses))
	for i, s := range addresses {
		str := s
		addressesPtr[i] = aws.String(str)
	}
	return addressesPtr
}

func (sesUtil *SESUtil) GetSendEmailInput(ctx context.Context, toAddresses []string, ccAddresses []string, bccAddresses []string, sourceEmail string, templateName string, templateData map[string]string) *ses.SendTemplatedEmailInput {
	logger := log.FromContext(ctx).WithName("SESUtil.GetSendEmailInput")
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("ses send email")
	destination := &ses.Destination{
		ToAddresses: toStringPtrSlice(toAddresses),
	}
	if len(ccAddresses) > 0 {
		destination.CcAddresses = toStringPtrSlice(ccAddresses)
	}
	if len(ccAddresses) > 0 {
		destination.BccAddresses = toStringPtrSlice(bccAddresses)
	}
	emailInput := &ses.SendTemplatedEmailInput{
		Destination:  destination,
		Template:     aws.String(templateName),
		Source:       aws.String(sourceEmail),
		TemplateData: aws.String(mapToString(templateData)),
	}
	return emailInput
}

func (sesUtil *SESUtil) SendEmailWithOptions(ctx context.Context, toAddresses []string, ccAddresses []string, bccAddresses []string, sourceEmail string, templateName string, templateData map[string]string) (string, error) {
	logger := log.FromContext(ctx).WithName("SESUtil.SendEmail")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	emailInput := sesUtil.GetSendEmailInput(ctx, toAddresses, ccAddresses, bccAddresses, sourceEmail, templateName, templateData)
	msgId, err := sesUtil.SendEmailWithTemplate(ctx, emailInput)
	if err != nil {
		logger.Error(err, "error sending email")
		return "", err
	}
	return msgId, nil

}
