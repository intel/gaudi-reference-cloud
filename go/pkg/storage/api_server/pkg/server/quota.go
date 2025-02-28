// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database/query"
)

type QuotaService struct {
	qmsClient             v1.QuotaManagementPrivateServiceClient
	accountQuotaCacheFile map[string]query.UsedQuotaFile
	accountQuotaCacheObj  map[string]query.UsedQuotaObj
	accountConfigQuota    CloudAccountQuota
	session               *sql.DB
	updateTimestamp       time.Time
	mu                    sync.Mutex
}

const (
	quotaCacheExpireSecs = 3600 // Every 1 hour
	storageServiceName   = "storage"
	filesystemQuota      = "filesystems"
	filesystemSizeQuota  = "totalSizeTB"
	bucketQuota          = "buckets"
	bucketSizeQuota      = "bucketSizeInTB"
)

func (qSvc *QuotaService) Init(ctx context.Context, session *sql.DB, configQuota CloudAccountQuota, qmsClient v1.QuotaManagementPrivateServiceClient) error {
	logger := log.FromContext(ctx).WithName("QuotaService.Init")
	logger.Info("initializing quota service ")
	qSvc.qmsClient = qmsClient
	qSvc.accountConfigQuota = configQuota
	qSvc.session = session
	qSvc.accountQuotaCacheFile = map[string]query.UsedQuotaFile{}
	qSvc.accountQuotaCacheObj = map[string]query.UsedQuotaObj{}

	if err := qSvc.refreshCache(ctx); err != nil {
		return err
	}

	return nil
}

func (qSvc *QuotaService) refreshCache(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("QuotaService.refreshCache")
	qSvc.accountQuotaCacheFile = map[string]query.UsedQuotaFile{}
	qSvc.accountQuotaCacheObj = map[string]query.UsedQuotaObj{}
	if err := query.UpdateUsedFileQuotaForAllAccounts(ctx, qSvc.session, &qSvc.accountQuotaCacheFile, timestampInfinityStr); err != nil {
		return fmt.Errorf("error initializing file quota")
	}
	if err := query.UpdateUsedObjQuotaForAllAccounts(ctx, qSvc.session, &qSvc.accountQuotaCacheObj); err != nil {
		return fmt.Errorf("error initializing object quota")
	}
	qSvc.updateTimestamp = time.Now()
	logger.Info("Refreshed cache completed : ", logkeys.AccountCacheAtReturn, qSvc.accountQuotaCacheFile)

	return nil
}

func (qSvc *QuotaService) checkAndUpdateFileQuota(ctx context.Context, cloudAccount *v1.CloudAccount, requestedSize int64, isUpdateFileRequest bool) bool {
	logger := log.FromContext(ctx).WithName("QuotaService.checkAndUpdateFileQuota")
	logger.Info("entering cloudaccount quota check ", logkeys.CloudAccountId, cloudAccount.GetId(), logkeys.CloudAccountType, cloudAccount.GetType())
	var storageQuotaByAccount *query.StorageQuotaByAccount
	var filevolumesQuota int64
	var filesizeQuotaInTB int64

	// Get volume quota
	if qSvc.qmsClient != nil {
		req := &v1.ServiceQuotaResourceRequestPrivate{
			ServiceName:    storageServiceName,
			CloudAccountId: cloudAccount.Id,
		}
		// Get cloudaccount storage quotas from QMS
		// Get volume quota
		req.ResourceType = filesystemQuota
		resp, err := qSvc.qmsClient.GetResourceQuotaPrivate(ctx, req)
		if err != nil {
			logger.Error(err, "error when calling QMS for filevolumesQuota")
			return false
		}
		if len(resp.CustomQuota.ServiceResources) != 0 {
			filevolumesQuota = resp.CustomQuota.ServiceResources[0].QuotaConfig.Limits
		} else {
			filevolumesQuota = resp.DefaultQuota.ServiceResources[0].QuotaConfig.Limits
		}
		logger.Info("qms volume count quota", "quota", filevolumesQuota)

		// Get volume size quota
		req.ResourceType = filesystemSizeQuota
		resp, err = qSvc.qmsClient.GetResourceQuotaPrivate(ctx, req)
		if err != nil {
			logger.Error(err, "error when calling QMS for filesizeQuotaInTB")
			return false
		}
		if len(resp.CustomQuota.ServiceResources) != 0 {
			filesizeQuotaInTB = resp.CustomQuota.ServiceResources[0].QuotaConfig.Limits
		} else {
			filesizeQuotaInTB = resp.DefaultQuota.ServiceResources[0].QuotaConfig.Limits
		}
		logger.Info("qms volume size quota", "quota", filesizeQuotaInTB)
	} else {
		// Start a new transaction
		tx, err := qSvc.session.BeginTx(ctx, nil)
		if err != nil {
			logger.Error(err, "error starting transaction")
			// handle error
		}
		// Get the storage quota by account
		storageQuotaByAccount, err = query.GetStorageQuotaByAccount(ctx, tx, cloudAccount.GetId())
		if err != nil {
			logger.Error(err, "error getting storage quota by account")
			// handle error
		}
	}

	qSvc.mu.Lock()
	defer qSvc.mu.Unlock()

	//ensure the cache is not stale, else refresh it
	logger.Info("quota cache stale-ness check", logkeys.TimeInSecondsSinceUpdate, time.Since(qSvc.updateTimestamp).Seconds(), logkeys.QuotaCacheTtl, quotaCacheExpireSecs)
	if time.Since(qSvc.updateTimestamp).Seconds() > quotaCacheExpireSecs {
		if err := qSvc.refreshCache(ctx); err != nil {
			logger.Error(err, "error refreshing cache")
			// soft handle the error,
			// continue using the stale cache
		}
	}

	var qtFilesystems int64
	var qtSize int64

	// If there's a storage quota for the account, use it
	if qSvc.qmsClient != nil {
		qtFilesystems = filevolumesQuota
		qtSize = filesizeQuotaInTB * 1000 // convert the TB to GB
	} else {
		if storageQuotaByAccount != nil {
			qtFilesystems = storageQuotaByAccount.FilevolumesQuota
			qtSize = storageQuotaByAccount.FilesizeQuotaInTB * 1000 // convert the TB to GB

		} else {
			// If there's no storage quota for the account, use the launch quota
			launchQuota, qCfgfound := qSvc.accountConfigQuota.CloudAccounts[getAccountTypeStr(cloudAccount.Type)]
			if !qCfgfound {
				return true
			}
			qtFilesystems = int64(launchQuota.StorageQuota["filesystems"])
			qtSize = int64(launchQuota.StorageQuota["totalSizeGB"])
		}
	}

	logger.Info("storage quota setting ", logkeys.QuotaFilesystem, qtFilesystems, logkeys.QuotaSize, qtSize)
	usedQuota, found := (qSvc.accountQuotaCacheFile)[cloudAccount.GetId()]
	if !found {
		logger.Info("no used quota found for account", logkeys.CloudAccountId, cloudAccount.GetId())
		(qSvc.accountQuotaCacheFile)[cloudAccount.GetId()] = query.UsedQuotaFile{
			TotalFileInstances: 1,
			TotalFileSizeInGB:  requestedSize,
		}
		return true
	}
	logger.Info("accountcache", "used quota", usedQuota)
	if !isUpdateFileRequest && qtFilesystems <= int64(usedQuota.TotalFileInstances) {
		return false
	}
	if qtSize < int64(usedQuota.TotalFileSizeInGB+requestedSize) {
		return false
	}

	if !isUpdateFileRequest {
		usedQuota.TotalFileInstances++
	}

	usedQuota.TotalFileSizeInGB += requestedSize
	(qSvc.accountQuotaCacheFile)[cloudAccount.GetId()] = usedQuota
	logger.Info("quota check passed for account", logkeys.CloudAccountId, cloudAccount.GetId())
	return true
}

func (qSvc *QuotaService) decFileQuota(ctx context.Context, cloudAccountId string, deletedSize int64, isUpdateFileRequest bool) bool {
	logger := log.FromContext(ctx).WithName("QuotaService.decFileQuota")
	logger.Info("entering cloudaccount quota update for delete ", logkeys.CloudAccountId, cloudAccountId)
	defer logger.Info("debug info", logkeys.AccountCacheAtReturn, qSvc.accountQuotaCacheFile)
	qSvc.mu.Lock()
	defer qSvc.mu.Unlock()

	usedQuota, found := (qSvc.accountQuotaCacheFile)[cloudAccountId]
	if !found {
		logger.Info("no used quota found for account", logkeys.CloudAccountId, cloudAccountId)
		return true
	}
	logger.Info("updating used quota for account", logkeys.CloudAccountId, cloudAccountId)
	// This is a rare case, but a safeguard to avoid decrementing quota below 0
	if !isUpdateFileRequest {
		logger.Info("Not an update request, decrementing file instances")
		logger.Info("updating used quota for account", "before : ", usedQuota)

		usedQuota.TotalFileInstances--
		logger.Info("updating used quota for account", "after : ", usedQuota)

		if usedQuota.TotalFileInstances < 0 {
			logger.Info("Resetting total file instances to 0")
			usedQuota.TotalFileInstances = 0
		}
	}
	usedQuota.TotalFileSizeInGB -= deletedSize
	if usedQuota.TotalFileSizeInGB < 0 {
		logger.Info("Resetting total file instances to 0")
		usedQuota.TotalFileSizeInGB = 0
	}
	(qSvc.accountQuotaCacheFile)[cloudAccountId] = usedQuota

	return true
}

func getAccountTypeStr(accType v1.AccountType) string {
	switch accType {
	case v1.AccountType_ACCOUNT_TYPE_ENTERPRISE:
		return "ENTERPRISE"
	case v1.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING:
		return "ENTERPRISE_PENDING"
	case v1.AccountType_ACCOUNT_TYPE_INTEL:
		return "INTEL"
	case v1.AccountType_ACCOUNT_TYPE_PREMIUM:
		return "PREMIUM"
	case v1.AccountType_ACCOUNT_TYPE_STANDARD:
		return "STANDARD"
	default:
		return "UNKNOWN"
	}
}

func (qSvc *QuotaService) checkAndUpdateBucketQuota(ctx context.Context, cloudAccount *v1.CloudAccount) bool {
	logger := log.FromContext(ctx).WithName("QuotaService.checkAndUpdateBucketQuota")
	logger.Info("entering cloudaccount quota check ", logkeys.CloudAccountId, cloudAccount.GetId(), logkeys.CloudAccountType, cloudAccount.GetType())
	var storageQuotaByAccount *query.StorageQuotaByAccount
	var bkQuota int64
	// Get bucket quota

	if qSvc.qmsClient != nil {
		req := &v1.ServiceQuotaResourceRequestPrivate{
			ServiceName:    storageServiceName,
			CloudAccountId: cloudAccount.Id,
		}
		// Get cloudaccount storage quotas from QMS
		req.ResourceType = bucketQuota
		resp, err := qSvc.qmsClient.GetResourceQuotaPrivate(ctx, req)
		if err != nil {
			logger.Error(err, "error when calling QMS for filevolumesQuota")
			return false
		}
		if len(resp.CustomQuota.ServiceResources) != 0 {
			bkQuota = resp.CustomQuota.ServiceResources[0].QuotaConfig.Limits
		} else {
			bkQuota = resp.DefaultQuota.ServiceResources[0].QuotaConfig.Limits
		}
		logger.Info("qms bucket count quota", "quota", bkQuota)
	} else {
		tx, err := qSvc.session.BeginTx(ctx, nil)
		if err != nil {
			logger.Error(err, "error starting transaction")
			// handle error
		}
		// Get the storage quota by account
		storageQuotaByAccount, err = query.GetStorageQuotaByAccount(ctx, tx, cloudAccount.GetId())
		if err != nil {
			logger.Error(err, "error getting storage quota by account")
			// handle error
		}
	}

	qSvc.mu.Lock()
	defer qSvc.mu.Unlock()

	//ensure the cache is not stale, else refresh it
	logger.Info("quota cache stale-ness check", logkeys.TimeInSecondsSinceUpdate,
		time.Since(qSvc.updateTimestamp).Seconds(), "quotaCacheTtl", quotaCacheExpireSecs)
	if time.Since(qSvc.updateTimestamp).Seconds() > quotaCacheExpireSecs {
		if err := qSvc.refreshCache(ctx); err != nil {
			logger.Error(err, "error refreshing cache")
			// soft handle the error,
			// continue using the stale cache
		}
	}
	var qtBuckets int64
	if qSvc.qmsClient != nil {
		qtBuckets = bkQuota
	} else {
		// If there's a storage quota for the account, use it
		if storageQuotaByAccount != nil {
			qtBuckets = storageQuotaByAccount.BucketsQuota
		} else {
			// If there's no storage quota for the account, use the launch quota
			launchQuota, qCfgfound := qSvc.accountConfigQuota.CloudAccounts[getAccountTypeStr(cloudAccount.Type)]
			if !qCfgfound {
				return true
			}
			qtBuckets = int64(launchQuota.StorageQuota["buckets"])
		}
	}

	logger.Info("storage bucket quota setting ", logkeys.QuotaBuckets, qtBuckets)

	usedQuota, found := (qSvc.accountQuotaCacheObj)[cloudAccount.GetId()]
	if !found {
		(qSvc.accountQuotaCacheObj)[cloudAccount.GetId()] = query.UsedQuotaObj{
			TotalBuckets: 1,
		}
		return true
	}
	if qtBuckets <= int64(usedQuota.TotalBuckets) {
		return false
	}
	usedQuota.TotalBuckets++
	(qSvc.accountQuotaCacheObj)[cloudAccount.GetId()] = usedQuota

	return true
}

func (qSvc *QuotaService) decBucketQuota(ctx context.Context, cloudAccountId string) bool {
	logger := log.FromContext(ctx).WithName("QuotaService.DecBucketQuota")
	logger.Info("entering cloudaccount quota update for delete bucket", logkeys.CloudAccountId, cloudAccountId)
	defer logger.Info("debug info", logkeys.AccountCacheAtReturnForBuckets, qSvc.accountQuotaCacheObj)
	qSvc.mu.Lock()
	defer qSvc.mu.Unlock()

	usedQuota, found := (qSvc.accountQuotaCacheObj)[cloudAccountId]
	if !found {
		logger.Info("no used quota found for account", logkeys.CloudAccountId, cloudAccountId)
		return true
	}
	logger.Info("updating used quota for account for buckets", logkeys.CloudAccountId, cloudAccountId)
	// This is a rare case, but a safeguard to avoid decrementing quota below 0
	usedQuota.TotalBuckets--
	if usedQuota.TotalBuckets < 0 {
		usedQuota.TotalBuckets = 0
	}
	(qSvc.accountQuotaCacheObj)[cloudAccountId] = usedQuota
	return true
}
func (qSvc *QuotaService) validateQuotaChangeRequest(ctx context.Context, tx *sql.Tx, cloudAccountId, cloudAccountType string, FilesizeQuotaInTB, filevolumesQuota, bucketsQuota int64) error {
	if cloudAccountId == "" {
		return fmt.Errorf("invalid cloud account id")
	}
	accountType, err := AccountTypeFromString(cloudAccountType)
	if err != nil {
		return fmt.Errorf("invalid account type: %w", err)
	}
	launchQuota, qCfgfound := qSvc.accountConfigQuota.CloudAccounts[getAccountTypeStr(accountType)]
	if !qCfgfound {
		return fmt.Errorf("default config not found")
	}

	var originalFilesizeQuotaInTB int64
	if size, ok := launchQuota.StorageQuota["totalSizeGB"]; ok {
		originalFilesizeQuotaInTB = int64(size) / 1000
	}
	originalFilevolumesQuota := int64(launchQuota.StorageQuota["filesystems"])
	originalBucketsQuota := int64(launchQuota.StorageQuota["buckets"])

	if (FilesizeQuotaInTB < originalFilesizeQuotaInTB) || (filevolumesQuota < originalFilevolumesQuota) || (bucketsQuota < originalBucketsQuota) {
		return fmt.Errorf("quota cannot be decreased")
	}

	// Get the current quota
	_, currentQuota, err := qSvc.GetStorageQuotaByAccount(ctx, tx, cloudAccountId, cloudAccountType)
	if err != nil {
		return fmt.Errorf("failed to get current quota. Default quota should exist: %w", err)
	}

	if currentQuota == nil {
		//No current quota found, so validation needed
		return nil
	}
	// Check if the new quota is less than the current quota
	if (FilesizeQuotaInTB < currentQuota.FilesizeQuotaInTB) || (filevolumesQuota < currentQuota.FilevolumesQuota) || (bucketsQuota < currentQuota.BucketsQuota) {
		return fmt.Errorf("new quota cannot be less than the current quota")
	}

	return nil
}

func (qSvc *QuotaService) InsertStorageQuotaByAccount(ctx context.Context, tx *sql.Tx, cloudAccountId, cloudAccountType, reason string, FilesizeQuotaInTB, filevolumesQuota, bucketsQuota int64) (*query.StorageQuotaByAccount, error) {
	logger := log.FromContext(ctx).WithName("QuotaService.InsertStorageQuotaByAccount")
	logger.Info("entering insert quota by account", logkeys.CloudAccountId, cloudAccountId)
	defer logger.Info("inserted quota by account for", logkeys.CloudAccountId, cloudAccountId)
	if err := qSvc.refreshCache(ctx); err != nil {
		logger.Error(err, "error refreshing cache")
		// soft handle the error,
		// continue using the stale cache
	}
	if err := qSvc.validateQuotaChangeRequest(ctx, tx, cloudAccountId, cloudAccountType, FilesizeQuotaInTB, filevolumesQuota, bucketsQuota); err != nil {
		return nil, err
	}
	return query.InsertStorageQuotaByAccount(ctx, tx, cloudAccountId, cloudAccountType, reason, FilesizeQuotaInTB, filevolumesQuota, bucketsQuota)
}

func (qSvc *QuotaService) UpdateStorageQuotaByAccount(ctx context.Context, tx *sql.Tx, cloudAccountId, cloudAccountType, reason string, FilesizeQuotaInTB, filevolumesQuota, bucketsQuota int64) (*query.StorageQuotaByAccount, error) {
	logger := log.FromContext(ctx).WithName("QuotaService.UpdateStorageQuotaByAccount")
	logger.Info("entering update quota by account", logkeys.CloudAccountId, cloudAccountId)
	defer logger.Info("updated quota by account for", logkeys.CloudAccountId, cloudAccountId)
	if err := qSvc.refreshCache(ctx); err != nil {
		logger.Error(err, "error refreshing cache")
		// soft handle the error,
		// continue using the stale cache
	}
	if err := qSvc.validateQuotaChangeRequest(ctx, tx, cloudAccountId, cloudAccountType, FilesizeQuotaInTB, filevolumesQuota, bucketsQuota); err != nil {
		return nil, err
	}
	return query.UpdateStorageQuotaByAccount(ctx, tx, cloudAccountId, cloudAccountType, reason, FilesizeQuotaInTB, filevolumesQuota, bucketsQuota)
}

func (qSvc *QuotaService) DeleteStorageQuotaByAccount(ctx context.Context, tx *sql.Tx, cloudAccountId string) error {
	logger := log.FromContext(ctx).WithName("QuotaService.DeleteStorageQuotaByAccount")
	logger.Info("entering delete quota by account", logkeys.CloudAccountId, cloudAccountId)
	defer logger.Info("deleted quota by account for", logkeys.CloudAccountId, cloudAccountId)
	return query.DeleteStorageQuotaByAccount(ctx, tx, cloudAccountId)
}

func (qSvc *QuotaService) GetStorageQuotaByAccount(ctx context.Context, tx *sql.Tx, cloudAccountId string, cloudAccountType string) (*query.StorageQuotaByAccount, *query.StorageQuotaByAccount, error) {
	logger := log.FromContext(ctx).WithName("QuotaService.GetStorageQuotaByAccount")
	logger.Info("entering get quota by account", logkeys.CloudAccountId, cloudAccountId)
	defer logger.Info("get quota by account for", logkeys.CloudAccountId, cloudAccountId)
	updatedQuota, err := query.GetStorageQuotaByAccount(ctx, tx, cloudAccountId)
	if err != nil {
		return nil, nil, err
	}

	accountType, err := AccountTypeFromString(cloudAccountType)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid account type: %w", err)
	}

	launchQuota, qCfgfound := qSvc.accountConfigQuota.CloudAccounts[getAccountTypeStr(accountType)]
	if !qCfgfound {
		return &query.StorageQuotaByAccount{}, updatedQuota, nil
	}
	var filesizeQuotaInTB int64
	if size, ok := launchQuota.StorageQuota["totalSizeGB"]; ok {
		filesizeQuotaInTB = int64(size) / 1000
	}
	defaultQuota := &query.StorageQuotaByAccount{
		CloudAccountId:    cloudAccountId,
		CloudAccountType:  getAccountTypeStr(accountType),
		FilesizeQuotaInTB: filesizeQuotaInTB,
		FilevolumesQuota:  int64(launchQuota.StorageQuota["filesystems"]),
		BucketsQuota:      int64(launchQuota.StorageQuota["buckets"]),
	}

	return defaultQuota, updatedQuota, nil
}

func (qSvc *QuotaService) GetAllStorageQuota(ctx context.Context, tx *sql.Tx) ([]query.StorageQuotaByAccount, error) {
	return query.GetAllStorageQuota(ctx, tx)
}

func AccountTypeFromString(s string) (v1.AccountType, error) {
	switch s {
	case "ENTERPRISE":
		return v1.AccountType_ACCOUNT_TYPE_ENTERPRISE, nil
	case "ENTERPRISE_PENDING":
		return v1.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING, nil
	case "INTEL":
		return v1.AccountType_ACCOUNT_TYPE_INTEL, nil
	case "PREMIUM":
		return v1.AccountType_ACCOUNT_TYPE_PREMIUM, nil
	case "STANDARD":
		return v1.AccountType_ACCOUNT_TYPE_STANDARD, nil
	default:
		return 0, fmt.Errorf("invalid account type: %s", s)
	}
}

func (qSvc *QuotaService) GetStorageQuotaByType(ctx context.Context, tx *sql.Tx, cloudAccountType v1.AccountType) (*query.StorageQuotaByType, error) {

	launchQuota, qCfgfound := qSvc.accountConfigQuota.CloudAccounts[getAccountTypeStr(cloudAccountType)]
	if !qCfgfound {
		return &query.StorageQuotaByType{}, nil
	}
	var filesizeQuotaInTB int64
	if size, ok := launchQuota.StorageQuota["totalSizeGB"]; ok {
		filesizeQuotaInTB = int64(size) / 1000
	}
	defaultQuota := &query.StorageQuotaByType{
		CloudAccountType:  getAccountTypeStr(cloudAccountType),
		FilesizeQuotaInTB: filesizeQuotaInTB,
		FilevolumesQuota:  int64(launchQuota.StorageQuota["filesystems"]),
		BucketsQuota:      int64(launchQuota.StorageQuota["buckets"]),
	}

	return defaultQuota, nil
}
