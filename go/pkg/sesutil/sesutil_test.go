// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package sesutil

import (
	"context"

	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func TestSendEmailForOTP(t *testing.T) {
	sesUtil := SESUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("ses.TestSendEmail")
	logger.Info("TestSendEmailForOTP BEGIN")
	defer logger.Info("TestSendEmailForOTP END")

	if err := sesUtil.Init(ctx, "us-west-2", ""); err != nil {
		logger.Error(err, "couldn't init ses")
		return
	}
	destinationEmail := "agpremium211223@proton.me"
	sourceEmail := "test@dev3.api.idcservice.net"
	templateName := "Multi-User-OTP-Template"
	templateData := map[string]string{
		"OTP": "5915",
	}
	logger.Info("calling send email")
	if _, err := sesUtil.SendEmail(ctx, destinationEmail, sourceEmail, templateName, templateData); err != nil {
		logger.Error(err, "couldn't send mail")
	}

}

func TestSendEmailForInvitation(t *testing.T) {
	sesUtil := SESUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("ses.TestSendEmailForInvitation")
	logger.Info("TestSendEmailForInvitation BEGIN")
	defer logger.Info("TestSendEmailForInvitation END")

	if err := sesUtil.Init(ctx, "us-west-2", ""); err != nil {
		logger.Error(err, "couldn't init ses")
		return
	}
	destinationEmail := "agpremium211223@proton.me"
	sourceEmail := "test@dev3.api.idcservice.net"
	templateName := "Multi-User-Invitation-Template"
	templateData := map[string]string{
		"invitationLink": "https://consumerint.intel.com/intelcorpintb2c.onmicrosoft.com/B2C_1A_UnifiedLogin_SISU_CML_SAML/generic/login?entityId=wcm-qa.intel.com&ui_locales=en",
		"invitationCode": "12345678",
	}
	logger.Info("calling send email")
	if _, err := sesUtil.SendEmail(ctx, destinationEmail, sourceEmail, templateName, templateData); err != nil {
		logger.Error(err, "couldn't send mail")
	}

}

func TestSendEmailWithOptions(t *testing.T) {
	sesUtil := SESUtil{}
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("ses.TestSendEmailWithOptions")
	logger.Info("TestSendEmailWithOptions BEGIN")
	defer logger.Info("TestSendEmailWithOptions END")

	if err := sesUtil.Init(ctx, "us-west-2", ""); err != nil {
		logger.Error(err, "couldn't init ses")
		return
	}
	toAddresses := []string{"agpremium211223@proton.me"}
	ccAddresses := []string{"amardeepcg1@proton.me"}
	sourceEmail := "test@dev3.api.idcservice.net"
	templateName := "Terminate-Instance-100-Used-Template-Dev"
	templateData := map[string]string{
		"consoleUrl": "https://dev3.console.idcservice.net/",
		"paymentUrl": "https://dev3.console.idcservice.net/upgradeaccount",
	}
	logger.Info("calling send email with options")
	if _, err := sesUtil.SendEmailWithOptions(ctx, toAddresses, ccAddresses, []string{}, sourceEmail, templateName, templateData); err != nil {
		logger.Error(err, "couldn't send mail with options", "toAddresses", toAddresses, "ccAddresses", ccAddresses)
	}

}
