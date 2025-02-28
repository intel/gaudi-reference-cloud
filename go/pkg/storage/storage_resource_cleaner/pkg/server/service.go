// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"io"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type StorageResourceCleanerService struct {
	billingClient    pb.BillingDeactivateInstancesServiceClient
	storageFSClient  pb.FilesystemPrivateServiceClient
	storageBKClient  pb.ObjectStorageServicePrivateClient
	emailClient      pb.EmailNotificationServiceClient
	templateName     string
	senderEmail      string
	consoleUrl       string
	paymentUrl       string
	cleanUpThreshold int
	delayInterval    int
	enabled          bool
	dryRun           bool
}

func NewStorageResourceCleaner(storageFSClient pb.FilesystemPrivateServiceClient, storageBKClient pb.ObjectStorageServicePrivateClient, billingClient pb.BillingDeactivateInstancesServiceClient, emailClient pb.EmailNotificationServiceClient, cfg *Config) (*StorageResourceCleanerService, error) {
	return &StorageResourceCleanerService{
		billingClient:    billingClient,
		storageFSClient:  storageFSClient,
		storageBKClient:  storageBKClient,
		emailClient:      emailClient,
		cleanUpThreshold: int(cfg.Threshold),
		delayInterval:    int(cfg.Interval),
		senderEmail:      cfg.SenderEmail,
		templateName:     cfg.TemplateName,
		consoleUrl:       cfg.ConsoleUrl,
		paymentUrl:       cfg.PaymentUrl,
		enabled:          cfg.ServiceEnabled,
		dryRun:           cfg.DryRun,
	}, nil
}

func (svc *StorageResourceCleanerService) scanAccounts(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("scanAccounts")
	logger.Info("Begin retrieve list of account with storage")

	if !svc.enabled {
		logger.Info("service in disabled mode")
		return
	}
	// Fetch list of all cloudaccounts with storage resources
	fsMap, bkMap := svc.fetchAccounts(ctx)
	logger.Info("accounts with volumes", "cloudaccounts", fsMap)
	logger.Info("accounts with buckets", "cloudaccounts", bkMap)

	// Clean resources for applicable accounts
	logger.Info("Begin filtering accounts that depleted credits")
	err := svc.cleanResources(ctx, fsMap, bkMap)
	if err != nil {
		logger.Error(err, "error encountered during resource cleanup")
	}

}

// Retrieve all cloudaccounts that have active storage resources (volumes/buckets)
func (svc *StorageResourceCleanerService) fetchAccounts(ctx context.Context) (map[string][]string, map[string][]string) {
	logger := log.FromContext(ctx).WithName("scanAccounts")
	logger.Info("Begin Scan")
	fsMap := make(map[string][]string)
	bkMap := make(map[string][]string)
	// Fetch list of all filesystems
	searchFSReq := pb.FilesystemSearchStreamPrivateRequest{ResourceVersion: "0", AvailabilityZone: "az1"}
	fsStream, err := svc.storageFSClient.SearchFilesystemRequests(ctx, &searchFSReq)
	if err != nil {
		logger.Error(err, "error reading filesystem requests")
	}
	var fsResp *pb.FilesystemRequestResponse
	var bkResp *pb.ObjectBucketSearchPrivateResponse
	for {
		fsResp, err = fsStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error(err, "error reading from stream")
			break
		}
		if fsResp == nil {
			logger.Info("received empty response")
			break
		}
		cloudAcc := fsResp.Filesystem.Metadata.CloudAccountId
		//skip adding to map if iks volume
		if fsResp.Filesystem.Spec.Prefix != "" {
			continue
		}
		if _, exists := fsMap[cloudAcc]; !exists {
			fsMap[cloudAcc] = append(fsMap[cloudAcc], fsResp.Filesystem.Metadata.Name)
		} else {
			fsMap[cloudAcc] = []string{fsResp.Filesystem.Metadata.Name}
		}

	}
	searchBKReq := pb.ObjectBucketSearchPrivateRequest{ResourceVersion: "0", AvailabilityZone: "az1"}
	bkStream, err := svc.storageBKClient.SearchBucketPrivate(ctx, &searchBKReq)
	if err != nil {
		logger.Error(err, "error reading bucket requests")
	}
	for {
		bkResp, err = bkStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error(err, "error reading from stream")
			break
		}
		if bkResp == nil {
			logger.Info("received empty response")
			break
		}
		cloudAcc := bkResp.Bucket.Metadata.CloudAccountId
		if _, exists := bkMap[cloudAcc]; !exists {
			bkMap[cloudAcc] = append(bkMap[cloudAcc], bkResp.Bucket.Metadata.Name)
		} else {
			bkMap[cloudAcc] = []string{bkResp.Bucket.Metadata.Name}
		}
	}
	return fsMap, bkMap
}

// Clean up resources for accounts meeting criteria
func (svc *StorageResourceCleanerService) cleanResources(ctx context.Context, fsMap map[string][]string, bkMap map[string][]string) error {
	logger := log.FromContext(ctx).WithName("cleanResources")
	logger.Info("Begin Scan for accounts in bad financial standing")
	storageFamily := "Storage as a Service"
	elapsedTime := "0" // filter for how much time has elapsed since credit depletion
	bctx := context.Background()
	deactivation, err := svc.billingClient.GetDeactivatedServiceAccounts(bctx, &pb.GetDeactivatedAccountsRequest{
		ProductFamily:    &storageFamily,
		CleanupThreshold: &elapsedTime,
	})
	if err != nil {
		logger.Error(err, "error retrieving cloudaccounts for deactivation")
	}
	accountCount := 0
	for {
		account, err := deactivation.Recv()
		if err == io.EOF {
			logger.Info("receiving accounts eof", "err", err, "accountCount", accountCount)
			break
		}
		if err != nil {
			logger.Info("error receiving accounts", "account", account, "err", err, "accountCount", accountCount)
			logger.Error(err, "error reading from stream")
			break
		}
		if account == nil {
			logger.Info("error receiving account null", "err", err, "accountCount", accountCount)
			logger.Error(err, "received empty response")
			break
		}
		accountCount = accountCount + 1
		cloudAcc := account.CloudAccountId
		//Check if any resources for this account
		var match bool
		if _, exists := fsMap[cloudAcc]; exists {
			match = true
		}
		if _, exists := bkMap[cloudAcc]; exists {
			match = true
		}

		if match {
			logger.Info("account with resources found", logkeys.CloudAccount, cloudAcc)
			depeletionTime := account.CreditsDepleted.AsTime()
			logger.Info("cloud credit depletion time", "time", depeletionTime)
			currTime := time.Now()
			duration := currTime.Sub(depeletionTime)
			logger.Info("time elapsed since credit depletion", logkeys.Duration, duration)
			// check if send email
			if duration <= 1*24*time.Hour { //send email within first day of credit depletion
				email := account.Email
				err := svc.sendNotification(ctx, email)
				if err != nil {
					logger.Error(err, "failed to send notification")
				}
			}
			// Delete resources for account if 30 or more days have elapsed
			if duration >= time.Duration(svc.cleanUpThreshold)*24*time.Hour {
				err := svc.deleteBuckets(ctx, cloudAcc, bkMap[cloudAcc])
				if err != nil {
					logger.Info("failed to delete buckets for", logkeys.CloudAccount, cloudAcc)
					logger.Error(err, "error cleaning bucket")
				}
				// err = svc.deleteFilesystems(ctx, cloudAcc, fsMap[cloudAcc])
				// if err != nil {
				// 	logger.Info("failed to delete volumes for", logkeys.CloudAccount, cloudAcc)
				// 	logger.Error(err, "error cleaning volume")
				// }
			}
		} // end match found
	} // end for loop
	logger.Info("finish clean up")
	return nil
}

func (svc *StorageResourceCleanerService) sendNotification(ctx context.Context, email string) error {
	logger := log.FromContext(ctx).WithName("sendNotification")
	logger.Info("Begin notify")
	toAddresses := []string{email}
	ccAddresses := []string{"idc-staas-notification@intel.com"}
	sourceEmail := svc.senderEmail
	templateName := svc.templateName
	templateData := map[string]string{
		"consoleUrl": svc.consoleUrl,
		"paymentUrl": svc.paymentUrl,
	}
	request := &pb.SendEmailRequest{
		ServiceName:  "StorageAsAService",
		Recipient:    toAddresses,
		CcRecipients: ccAddresses,
		Sender:       sourceEmail,
		TemplateName: templateName,
		TemplateData: templateData,
	}
	if svc.emailClient != nil {
		if _, err := svc.emailClient.SendEmail(ctx, request); err != nil {
			logger.Error(err, "couldn't send mail with options", "toAddresses", toAddresses, "ccAddresses", ccAddresses)
			return err
		}
	}

	return nil
}

func (svc *StorageResourceCleanerService) deleteFilesystems(ctx context.Context, cloudaccount string, fsList []string) error {
	logger := log.FromContext(ctx).WithName("deleteFilesystems")
	logger.Info("Begin Volume cleanup for", logkeys.CloudAccount, cloudaccount)

	for _, fsName := range fsList {
		// if flag true, write to logs and skip deletion
		if svc.dryRun {
			logger.Info("dry run delete volume for", logkeys.CloudAccount, cloudaccount, logkeys.FilesystemName, fsName)
			continue
		}
		_, err := svc.storageFSClient.DeletePrivate(ctx, &pb.FilesystemDeleteRequestPrivate{
			Metadata: &pb.FilesystemMetadataReference{
				CloudAccountId: cloudaccount,
				NameOrId: &pb.FilesystemMetadataReference_Name{
					Name: fsName,
				},
			},
		})
		if err != nil {
			logger.Error(err, "failed to delete volume for ", logkeys.CloudAccount, cloudaccount, logkeys.FilesystemName, fsName)
		} else {
			logger.Info("successfully deleted volume for", logkeys.CloudAccount, cloudaccount, logkeys.FilesystemName, fsName)
		}

	}
	return nil
}

func (svc *StorageResourceCleanerService) deleteBuckets(ctx context.Context, cloudaccount string, bkList []string) error {
	logger := log.FromContext(ctx).WithName("deleteBuckets")
	logger.Info("Begin bucket cleanup for", logkeys.CloudAccount, cloudaccount)
	for _, bkName := range bkList {
		// if flag true, write to logs and skip deletion
		if svc.dryRun {
			logger.Info("dry run delete bucket for", logkeys.CloudAccount, cloudaccount, logkeys.BucketName, bkName)
			continue
		}
		_, err := svc.storageBKClient.DeleteBucketPrivate(ctx, &pb.ObjectBucketDeletePrivateRequest{
			Metadata: &pb.ObjectBucketMetadataRef{
				CloudAccountId: cloudaccount,
				NameOrId: &pb.ObjectBucketMetadataRef_BucketName{
					BucketName: bkName,
				},
			},
		})
		if err != nil {
			logger.Error(err, "failed to delete bucket for ", logkeys.CloudAccount, cloudaccount, logkeys.BucketName, bkName)
		} else {
			logger.Info("successfully deleted bucket for", logkeys.CloudAccount, cloudaccount, logkeys.BucketName, bkName)
		}
	}

	return nil
}

func (svc *StorageResourceCleanerService) StartAccountWatcher(ctx context.Context, interval time.Duration) {
	logger := log.FromContext(ctx).WithName("StartAccountWatcher")
	logger.Info("Start Service")
	// Periodically update the metric
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			svc.scanAccounts(ctx)
		}
	}
}
