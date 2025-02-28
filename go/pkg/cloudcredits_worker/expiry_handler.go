// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcreditsworker

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

const CloudCreditExpired = "Cloud Credits Expired"

func (w CloudCreditsWorker) HandleNearExpiryCredits(ctx context.Context, accountId string) error {
	logger := log.FromContext(ctx).WithName("CloudCreditsWorker.HandleNearExpiryCredits")
	logger.Info("cloud credits near expiry event", "accountId", accountId)
	defer logger.Info("No action required to be taken.")
	// Note: No requirement to send email notification

	// cloudAcct, err := w.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: accountId})
	// if err != nil {
	// 	logger.Error(err, "error getting acct details from cloud acct service")
	// 	return nil
	// }
	// if cloudAcct == nil {
	// 	logger.Info("cloud account not found", "cloudAcct", cloudAcct)
	// 	return nil
	// }

	// logger.Info("sending email for near cloud credit expiry for cloud account", "id", cloudAcct.Id)
	// if err = w.billingNotificationClient.SendCloudCreditUsageEmail(ctx, CloudCreditExpired, cloudAcct.Owner, billing.Cfg.GetCloudCreditExpiryTemplate()); err != nil {
	// 	logger.Error(err, "couldn't send email for low credits", "Owner", cloudAcct.Owner)
	// }

	return nil
}
