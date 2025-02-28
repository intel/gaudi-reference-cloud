// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	usageControllerLinkChannel = make(chan bool)
	usageControllerTicker      *time.Ticker
)

func StartReportUsageScheduler(ctx context.Context, usageController *UsageController) {
	_, logger, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.StartReportUsageScheduler").Start()
	logger.Info("BEGIN")
	defer logger.Info("END")
	usageControllerTicker = time.NewTicker(time.Duration(config.Cfg.ReportUsageSchedulerInterval) * time.Second)
	go reportUsageLoop(context.Background(), usageController)
}

func StopReportUsageScheduler() {
	if usageControllerLinkChannel != nil {
		close(usageControllerLinkChannel)
		usageControllerLinkChannel = nil
	}
}

func reportUsageLoop(ctx context.Context, usageController *UsageController) {
	ctx, logger, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.reportUsageLoop").Start()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	for {
		err := usageController.ReportAllUsage(ctx)
		if err != nil {
			logger.Error(err, "failed to report all resource usages", "context", "ReportAllUsage")
		}
		select {
		case <-usageControllerLinkChannel:
			return
		case tm := <-usageControllerTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

type UsageController struct {
	cloudAccountClient pb.CloudAccountServiceClient
	usageServiceClient pb.UsageServiceClient
	ariaUsageClient    *client.AriaUsageClient
	ariaAccountClient  *client.AriaAccountClient
}

func NewUsageController(ariaCredentials *client.AriaCredentials, cloudAccountClient pb.CloudAccountServiceClient,
	usageServiceClient pb.UsageServiceClient, ariaUsageClient *client.AriaUsageClient, ariaAccountClient *client.AriaAccountClient) *UsageController {
	return &UsageController{
		cloudAccountClient: cloudAccountClient,
		usageServiceClient: usageServiceClient,
		ariaUsageClient:    ariaUsageClient,
		ariaAccountClient:  ariaAccountClient}
}

func (usageController *UsageController) getCloudAccts(ctx context.Context) []*pb.CloudAccount {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.getCloudAccts").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	acctTypes := []pb.AccountType{pb.AccountType_ACCOUNT_TYPE_PREMIUM, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE}
	cloudAccts := []*pb.CloudAccount{}

	for _, acctType := range acctTypes {
		cloudAccountSearchClient, err :=
			usageController.cloudAccountClient.Search(ctx, &pb.CloudAccountFilter{Type: &acctType})
		if err != nil {
			logger.Error(err, "failed to get cloud account client for searching on", "type", acctType.String())
			span.RecordError(fmt.Errorf("failed to get cloud account client for searching on %w", err))
			continue
		}

		for {
			cloudAccount, err := cloudAccountSearchClient.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				logger.Error(err, "failed to get cloud account")
				continue
			}
			cloudAccts = append(cloudAccts, cloudAccount)
		}
	}
	return cloudAccts
}

func (usageController *UsageController) ReportAllUsage(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.ReportAllUsage").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	cloudAccts := usageController.getCloudAccts(ctx)

	for _, cloudAcct := range cloudAccts {
		resourceUsages, err := usageController.GetResourceUsageOverStream(ctx, cloudAcct)
		if err != nil {
			logger.Error(err, "failed to report usages for cloud account", "cloudAccountId", cloudAcct.Id, "cloudAccountType", &cloudAcct.Type, "context", "SearchResourceUsages")
			span.RecordError(fmt.Errorf("failed to report usages for cloud account %w", err))
			return err
		}

		if resourceUsages == nil {
			logger.Info("no resource usages")
			continue
		}

		billingUsagesToUpload := []*client.BillingUsage{}
		for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
			billingUsagesToUpload = append(billingUsagesToUpload, &client.BillingUsage{
				CloudAccountId: resourceUsageRecord.CloudAccountId,
				ProductId:      resourceUsageRecord.ProductId,
				ResourceId:     resourceUsageRecord.ResourceId,
				ResourceName:   resourceUsageRecord.ResourceName,
				Amount:         resourceUsageRecord.Quantity,
				TransactionId:  resourceUsageRecord.Id,
				UsageId:        resourceUsageRecord.Id,
				RegionName:     resourceUsageRecord.Region,
				UsageDate:      resourceUsageRecord.EndTime.AsTime(),
				UsageUnitType:  resourceUsageRecord.UsageUnitType,
			})
		}
		billingUsagesUploaded, err := usageController.ReportUsageToAria(ctx, billingUsagesToUpload)

		if err != nil {
			span.RecordError(fmt.Errorf("failed to upload usage for cloud account %w", err))
			logger.Error(err, "failed to upload usage for cloud account", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "ReportUsageToAria")
		}
		for _, billingUsageUploaded := range billingUsagesUploaded {
			_, err = usageController.usageServiceClient.MarkResourceUsageAsReported(ctx, &pb.ResourceUsageId{Id: billingUsageUploaded.UsageId})
			if err != nil {
				span.RecordError(fmt.Errorf("failed to mark resource usage as reported for %w", err))
				logger.Error(err, "failed to mark resource usage as reported for ", "usageId", billingUsageUploaded.UsageId)
			}
		}
	}

	return nil
}

func (usageController *UsageController) ReportUsageToAria(ctx context.Context, billingUsagesToUpload []*client.BillingUsage) ([]*client.BillingUsage, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.ReportUsageToAria").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// do post bulk usage once.
	billingUsagesUploaded, billingUsagesFailedToUpload, err := usageController.postBulkUsageRecords(ctx, billingUsagesToUpload)
	if err != nil {
		span.RecordError(fmt.Errorf("failed to post usage records %w", err))
		logger.Error(err, "failed to post usage records", "billingUsagesToUpload", billingUsagesToUpload, "context", "postBulkUsageRecords")
		return nil, err
	}

	type accountProd struct {
		cloudAccountId string
		productId      string
	}

	// build a unique mapping between cloud accounts and products.
	failedCloudAccountProductMapping := map[accountProd][]*client.BillingUsage{}
	for _, billingUsageFailedToUpload := range billingUsagesFailedToUpload {
		ap := accountProd{
			cloudAccountId: billingUsageFailedToUpload.CloudAccountId,
			productId:      billingUsageFailedToUpload.ProductId,
		}
		failedCloudAccountProductMapping[ap] = append(failedCloudAccountProductMapping[ap], billingUsageFailedToUpload)
	}

	// Install plans in accounts and pick out usages to retry
	billingUsageRetryToUpload := []*client.BillingUsage{}
	for ap, usage := range failedCloudAccountProductMapping {
		err = usageController.addPlanToAccount(ctx, ap.cloudAccountId, ap.productId)
		if err != nil {
			span.RecordError(fmt.Errorf("failed to add correlation for cloud account id %v  error %w", ap.cloudAccountId, err))
			logger.Error(err, "failed to add correlation for", "cloudAccountId", ap.cloudAccountId, "productId", ap.productId)
			continue
		}
		billingUsageRetryToUpload = append(billingUsageRetryToUpload, usage...)
	}

	// retry failed usages after installing plans
	if len(billingUsageRetryToUpload) > 0 {
		billingUsagesRetryUploaded, _, err := usageController.postBulkUsageRecords(ctx, billingUsageRetryToUpload)
		if err != nil {
			span.RecordError(fmt.Errorf("failed to post usage records after associations  %w", err))
			logger.Error(err, "failed to post usage records after associations for", "billingUsageRetryToUpload", billingUsageRetryToUpload, "context", "postBulkUsageRecords")
			return nil, err
		}
		billingUsagesUploaded = append(billingUsagesUploaded, billingUsagesRetryUploaded...)
	}

	return billingUsagesUploaded, nil
}

func (usageController *UsageController) addPlanToAccount(ctx context.Context, cloudAccountId string, productId string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.addPlanToAccount").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	cloudAccount, err := usageController.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil {
		span.RecordError(fmt.Errorf("failed to get cloud account id %v and type %v error %w ", cloudAccount.Id, cloudAccount.Type, err))
		logger.Error(err, "failed to get cloud account", "cloudAccountId", cloudAccount.Id, "type", cloudAccount.Type, "context", "GetById")
		return err
	}

	var cloudAccountType string
	if cloudAccount.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE {
		cloudAccountType = "enterprise"
		clientAccountId := client.GetAccountClientId(cloudAccountId)

		masterPlansInfo, err := usageController.ariaAccountClient.GetEnterpriseParentAccountPlan(ctx, clientAccountId)
		if err != nil {
			logger.Error(err, "failed to get enterprise account plan", "cloudAccountId", cloudAccount.Id, "cloudAccountType", cloudAccount.Type, "context", "GetById")
			return err
		}
		for _, masterPlanInfo := range masterPlansInfo {
			if masterPlanInfo.ClientMasterPlanId == client.GetDefaultPlanClientId() {
				_, err := usageController.ariaAccountClient.AssignPlanToEnterpriseChildAccount(ctx, client.GetAccountClientId(cloudAccountId),
					client.GetPlanClientId(productId),
					client.GetClientMasterPlanInstanceId(cloudAccountId, productId),
					client.GetRateScheduleClientId(productId, cloudAccountType),
					client.BILL_LAG_DAYS, client.ALT_BILL_DAY, masterPlanInfo.LastBillThruDate, client.PRORATE_FIRST_INVOICE, client.PARENT_PAY_FOR_CHILD_ACCOUNT, masterPlanInfo.ClientMasterPlanInstanceId)
				if err != nil {
					span.RecordError(fmt.Errorf("failed to assign plan to enterprise child account Id %v type %v clientAccountId  %v  productId %v billLagDays %v altBillDay %v lastBillThruDate %v prorateFirstInvoice %v payForChildAccount %v clientMasterPlanInstanceId %v error %w", cloudAccount.Id, cloudAccount.Type, clientAccountId, productId, client.BILL_LAG_DAYS, client.ALT_BILL_DAY, masterPlanInfo.LastBillThruDate, client.PRORATE_FIRST_INVOICE, client.PARENT_PAY_FOR_CHILD_ACCOUNT, masterPlanInfo.ClientMasterPlanInstanceId, err))
					logger.Error(err, "failed to assign plan to enterprise child account", "cloudAccountId", cloudAccount.Id, "type", cloudAccount.Type,
						"clientAccountId", clientAccountId, "productId", productId, "billLagDays", client.BILL_LAG_DAYS, "altBillDay", client.ALT_BILL_DAY, "lastBillThruDate", masterPlanInfo.LastBillThruDate,
						"prorateFirstInvoice", client.PRORATE_FIRST_INVOICE, "payForChildAccount", client.PARENT_PAY_FOR_CHILD_ACCOUNT, "clientMasterPlanInstanceId", masterPlanInfo.ClientMasterPlanInstanceId, "context", "AssignPlanToEnterpriseChildAccount")
					return err
				}
			}
		}
	} else if cloudAccount.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		acctPlans, err := usageController.ariaAccountClient.GetAcctPlans(ctx, client.GetAccountClientId(cloudAccountId))
		if err != nil {
			logger.Error(err, "failed to get account plans for assigning plans to acct", "cloudAccountId", cloudAccountId, "cloudAccountType", cloudAccount.Type, "context", "GetAcctPlans")
			return err
		}

		cloudAccountType = "premium"

		for _, acctPlan := range acctPlans.AcctPlansM {
			if acctPlan.ClientPlanId == client.GetDefaultPlanClientId() {
				_, err := usageController.ariaAccountClient.AssignPlanToPremiumAcct(ctx, client.GetAccountClientId(cloudAccountId),
					client.GetPlanClientId(productId),
					client.GetClientMasterPlanInstanceId(cloudAccountId, productId),
					client.GetRateScheduleClientId(productId, cloudAccountType),
					client.BILL_LAG_DAYS, client.ALT_BILL_DAY, acctPlan.ClientMasterPlanInstanceId)

				if err != nil {
					span.RecordError(fmt.Errorf("failed to assign plan to premium account , cloudAccountId %v type %v  productId %v billLagDays %v  altBillDay %v  clientMasterPlanInstanceId %v error %w", cloudAccount.Id, cloudAccount.Type, productId, client.BILL_LAG_DAYS, client.ALT_BILL_DAY, acctPlan.ClientMasterPlanInstanceId, err))
					logger.Error(err, "failed to assign plan to premium account", "cloudAccountId", cloudAccount.Id, "type", cloudAccount.Type,
						"productId", productId, "billLagDays", client.BILL_LAG_DAYS, "altBillDay", client.ALT_BILL_DAY, "clientMasterPlanInstanceId", acctPlan.ClientMasterPlanInstanceId, "context", "AssignPlanToPremiumAcct")
					return err
				}
				return nil

			}
		}

		_, err = usageController.ariaAccountClient.AssignPlanToAccountWithBillingAndDunningGroup(ctx, client.GetAccountClientId(cloudAccountId),
			client.GetPlanClientId(productId),
			client.GetClientMasterPlanInstanceId(cloudAccountId, productId),
			client.GetRateScheduleClientId(productId, cloudAccountType),
			client.BILL_LAG_DAYS, client.ALT_BILL_DAY)

		if err != nil {
			span.RecordError(fmt.Errorf("failed to assign plan to account with billing and dunning group cloudAccountId %v type %v  productId %v billLagDays %v  altBillDay %v error %w", cloudAccount.Id, cloudAccount.Type, productId, client.BILL_LAG_DAYS, client.ALT_BILL_DAY, err))
			logger.Error(err, "failed to assign plan to account with billing and dunning group", "cloudAccountId", cloudAccount.Id, "type", cloudAccount.Type,
				"productId", productId, "billLagDays", client.BILL_LAG_DAYS, "altBillDay", client.ALT_BILL_DAY, "context", "AssignPlanToAccountWithBillingAndDunningGroup")
			return err
		}
		return nil

	} else {
		return errors.New("not supported cloud account type")
	}
	return nil
}

func (usageController *UsageController) postBulkUsageRecords(ctx context.Context, billingUsagesToUpload []*client.BillingUsage) ([]*client.BillingUsage, []*client.BillingUsage, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.postBulkUsageRecords").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	bulkUsageRecordsResponse, err := usageController.ariaUsageClient.CreateBulkUsageRecords(ctx, billingUsagesToUpload)
	if err != nil {
		span.RecordError(fmt.Errorf("failed to post usage records billingUsagesToUpload %v error %w", billingUsagesToUpload, err))
		logger.Error(err, "failed to post usage records", "billingUsagesToUpload", billingUsagesToUpload, "context", "CreateBulkUsageRecords")
		return nil, nil, err
	}

	var billingUsageFailedToUpload []*client.BillingUsage
	var billingUsageUploaded []*client.BillingUsage
	// get the error records for the ones that failed.
	errorRecords := map[string]bool{}
	for _, errorRecord := range bulkUsageRecordsResponse.ErrorRecords {
		errorRecords[errorRecord.OutClientRecordId] = true
	}
	for _, billingUsageToUpload := range billingUsagesToUpload {
		if errorRecords[billingUsageToUpload.TransactionId] {
			billingUsageFailedToUpload = append(billingUsageFailedToUpload, billingUsageToUpload)
		} else {
			billingUsageUploaded = append(billingUsageUploaded, billingUsageToUpload)
		}
	}
	return billingUsageUploaded, billingUsageFailedToUpload, nil
}

type SortedUsageRecordHistory []data.UsageHistoryRec

func (usage SortedUsageRecordHistory) Len() int      { return len(usage) }
func (usage SortedUsageRecordHistory) Swap(i, j int) { usage[i], usage[j] = usage[j], usage[i] }
func (usage SortedUsageRecordHistory) Less(i, j int) bool {
	return client.ParseDateAndTime(usage[i].UsageDate, usage[i].UsageTime).Before(client.ParseDateAndTime(usage[j].UsageDate, usage[j].UsageTime))
}

// To be depericated
func (usageController *UsageController) GetResourceUsage(ctx context.Context, cloudAcct *pb.CloudAccount) (*pb.ResourceUsages, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.GetResourceUsage").WithValues("cloudAccountId", cloudAcct.GetId()).Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")
	reported := false
	resourceUsages, err := usageController.usageServiceClient.SearchResourceUsages(ctx,
		&pb.ResourceUsagesFilter{
			CloudAccountId: &cloudAcct.Id,
			Reported:       &reported,
		})
	if err != nil {
		logger.Error(err, "failed to report cloud credit usages for cloud account", "cloudAccountId", cloudAcct.GetId())
		return nil, err
	}
	return resourceUsages, nil
}

func (usageController *UsageController) GetResourceUsageOverStream(ctx context.Context, cloudAcct *pb.CloudAccount) (*pb.ResourceUsages, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.GetResourceUsageOverStream").WithValues("cloudAccountId", cloudAcct.GetId()).Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	reported := false
	resourceUsagesFilter := &pb.ResourceUsagesFilter{
		CloudAccountId: &cloudAcct.Id,
		Reported:       &reported,
	}
	resourceUsagesStream, err := usageController.usageServiceClient.StreamSearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		logger.Error(err, "failed to get resource usages for cloud account")
		return nil, err
	}
	resourceUsages := &pb.ResourceUsages{}
	for {

		resourceUsage, err := resourceUsagesStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			logger.Error(err, "failed to recv resource usages for cloud account")
			return nil, err
		}

		resourceUsages.ResourceUsages = append(resourceUsages.ResourceUsages, resourceUsage)
	}
	return resourceUsages, nil
}
