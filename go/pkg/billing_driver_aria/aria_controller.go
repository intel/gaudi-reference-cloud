// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const kFixMeWeHaventImplementedReasonCodeYet int64 = 1
const defaultCreditAssignmentComment = "defaultCredits"
const acctExistsError = 1009
const creditCardAuthorizationAmount = 500

type AriaController struct {
	ariaAccountClient       *client.AriaAccountClient
	ariaPlanClient          *client.AriaPlanClient
	ariaServiceCreditClient *client.ServiceCreditClient
	ariaUsageTypeClient     *client.AriaUsageTypeClient
	ariaPromoClient         *client.PromoClient
	ariaPaymentClient       *client.AriaPaymentClient
	ariaUsageClient         *client.AriaUsageClient
	ariaInvoiceClient       *client.AriaInvoiceClient
	cloudAccountClient      pb.CloudAccountServiceClient
}

func NewAriaController(ariaClient *client.AriaClient, ariaAdminClient *client.AriaAdminClient, ariaCredentials *client.AriaCredentials) *AriaController {
	ariaAccountClient := client.NewAriaAccountClient(ariaClient, ariaCredentials)
	ariaPlanClient := client.NewAriaPlanClient(config.Cfg, ariaAdminClient, ariaClient, ariaCredentials)
	ariaServiceCreditClient := client.NewServiceCreditClient(ariaClient, ariaCredentials)
	ariaUsageTypeClient := client.NewAriaUsageTypeClient(ariaAdminClient, ariaCredentials)
	ariaPromoClient := client.NewPromoClient(ariaAdminClient, ariaCredentials)
	ariaPaymentClient := client.NewAriaPaymentClient(ariaClient, ariaCredentials)
	ariaUsageClient := client.NewAriaUsageClient(ariaClient, ariaCredentials)
	ariaInvoiceClient := client.NewAriaInvoiceClient(ariaClient, ariaCredentials)
	return &AriaController{ariaAccountClient: ariaAccountClient, ariaPlanClient: ariaPlanClient,
		ariaServiceCreditClient: ariaServiceCreditClient, ariaUsageTypeClient: ariaUsageTypeClient,
		ariaPromoClient: ariaPromoClient, ariaPaymentClient: ariaPaymentClient,
		ariaUsageClient: ariaUsageClient, cloudAccountClient: AriaService.cloudAccountClient, ariaInvoiceClient: ariaInvoiceClient}
}

// Get the plan for a client plan ID for a client number
// The client number is consistent across the driver.
func (ariaController *AriaController) GetClientPlan(ctx context.Context, clientPlanId string) ([]data.AllClientPlanDtl, error) {
	resp, err := ariaController.ariaPlanClient.GetAriaPlanDetailsAllForClientPlanId(ctx, clientPlanId)
	if err != nil {
		return nil, err
	}
	return resp.AllClientPlanDtls, nil
}

// Initialize Aria for use with IDC. This creates the master plan and
// the plan set to along with it.
//
// Usage types should also be created here.
func (ariaController *AriaController) InitAria(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.InitAria").Start()
	defer span.End()
	// ensureMinutesUsageType creates the minutes usage type if it does not already exist.
	// checks if the usage unit type of the type minute does not exist and errors out.
	// if does not exist, creates the usage type.
	// if does exist, and values are different than what is expected, updates the existing usage type.
	if err := ariaController.ensureMinutesUsageType(ctx); err != nil {
		return err
	}
	// ensureStorageUsageType creates the storage usage type if it does not already exist.
	// checks if the usage unit type of the type storage does not exist and errors out.
	// if does not exist, creates the usage type.
	// if does exist, and values are different than what is expected, updates the existing usage type.
	if config.Cfg.GetFeaturesSyncStoragePlan() {
		if err := ariaController.ensureStorageUsageType(ctx); err != nil {
			logger.Error(err, "error ensuring storage usage type")
		}
	}
	// ensureInferenceUsageType creates the inference usage type if it does not already exist.
	// checks if the usage unit type of the type inference does not exist and errors out.
	// if does not exist, creates the usage type.
	// if does exist, and values are different than what is expected, updates the existing usage type.
	if config.Cfg.GetFeaturesSyncInferencePlan() {
		if err := ariaController.ensureInferenceUsageType(ctx); err != nil {
			logger.Error(err, "error ensuring inference usage type")
		}
	}
	// ensureTokenUsageType creates the token usage type if it does not already exist.
	// checks if the usage unit type of the type token does not exist and errors out.
	// if does not exist, creates the usage type.
	// if does exist, and values are different than what is expected, updates the existing usage type.
	if config.Cfg.GetFeaturesSyncTokenPlan() {
		if err := ariaController.ensureTokenUsageType(ctx); err != nil {
			logger.Error(err, "error ensuring token usage type")
		}
	}
	// ensureDefaultPlan creates the default plan if it doesn't already exist
	// The default plan is created before the promo plan set because the
	// default plan gets added to the promo plan set.
	if err := ariaController.ensureDefaultPlan(ctx); err != nil {
		return err
	}

	// EnsurePlanSet creates the plan set to be used with the IDC promo code.
	// The plan set includes the default plan.
	if err := ariaController.ariaPromoClient.EnsurePlanSet(ctx); err != nil {
		return err
	}

	// EnsurePromo creates the IDC promo code, and associates the plan set
	// with the promo code. The default plan, the plan set, and the promo code
	// all use well-known identifiers so they can be linked together without
	// having to pass the indentifiers as parameters.
	return ariaController.ariaPromoClient.EnsurePromo(ctx)
}

func (ariaController *AriaController) CreateAriaAccount(ctx context.Context, cloudAccountId string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.CreateAriaAccount").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	cloudAccount, err := ariaController.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil {
		logger.Error(err, "failed to get cloud account for cloud account id", "id", cloudAccountId)
	}

	clientPlanId := client.GetDefaultPlanClientId()
	planExists, err := ariaController.ClientPlanDetailsExists(ctx, clientPlanId)
	if err != nil {
		logger.Error(err, "failed to create account , cannot check if default plan exists ", "clientPlanId", clientPlanId)
		return err
	}
	if !planExists {
		err := errors.New("client plan does not exist when expected")
		logger.Error(err, "default client plan for the creation of account needs to exist prior to account creation")
		return err
	}
	logger.Info("client plan exists ", "clientPlanId", clientPlanId, "exists", planExists)
	acctDetails, err := ariaController.ariaAccountClient.GetAriaAccountDetailsAllForClientId(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		if acctDetails != nil && acctDetails.ErrorCode != acctExistsError {
			logger.Error(err, "failed to check if account exists")
			return err
		}

		var acctType string
		if cloudAccount.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
			acctType = client.ACCOUNT_TYPE_PREMIUM
		} else if cloudAccount.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING {
			acctType = client.ACCOUNT_TYPE_ENTERPRISE_PENDING
		} else if cloudAccount.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE {
			acctType = client.ACCOUNT_TYPE_ENTERPRISE
		}

		_, err := ariaController.ariaAccountClient.CreateAriaAccount(ctx, client.GetAccountClientId(cloudAccountId), clientPlanId, acctType)
		if err != nil {
			logger.Error(err, "failed to create Aria account", "cloudAccountId", cloudAccountId)
			return err
		}
	}
	logger.Info("create account", "cloudAccountId", cloudAccountId, "clientPlanId", clientPlanId)
	//TWC4727-397: Remove default credit for Premium user
	// if cloudAccount.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
	// 	logger.V(1).Info("assing default credit to account", "cloudAccountId", cloudAccountId)
	// 	err := ariaController.assignDefaultCloudCreditsToPremium(ctx, cloudAccountId)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	logger.V(1).Info("set account notify template group id", "cloudAccountId", cloudAccountId)
	err = ariaController.SetAriaAccountNotifyTemplateGroup(ctx, cloudAccount)
	if err != nil {
		logger.Error(err, "failed to SetAriaAccountNotifyTemplateGroup", "cloudAccountId", cloudAccountId)
		return err
	}
	err = ariaController.UpdateAriaAccountDunningGroup(ctx, cloudAccount)
	if err != nil {
		logger.Error(err, "failed to UpdateAriaAccountDunningGroup", "cloudAccountId", cloudAccountId)
		return err
	}
	return ariaController.UpdateAriaAccountContact(ctx, cloudAccount)
}

// Aria controller function to call Service_credit's ApplyCreditService method to apply credits to aria account
func (ariaController *AriaController) AccountCreditMigrate(ctx context.Context, cloudAccountId string, remainingAmount float64) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.AccountCreditMigrate").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	_, err := ariaController.ariaServiceCreditClient.ApplyCreditService(ctx, client.GetAccountClientId(cloudAccountId), remainingAmount)
	if err != nil {
		logger.Error(err, "failed to migrate credit", "cloudAccountId", cloudAccountId, "remainingAmount", remainingAmount, "context", "ApplyCreditService")
		return err
	}
	return nil
}

func (ariaController *AriaController) assignDefaultCloudCreditsToPremium(ctx context.Context, cloudAccountId string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.assignDefaultCloudCreditsToPremium").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	acctCredits, err := ariaController.GetAccountCredits(ctx, cloudAccountId)
	if err != nil {
		logger.Error(err, "failed to get account credits", "cloudAccountId", cloudAccountId)
		return err
	}

	for _, acctCredit := range acctCredits {
		if acctCredit.ReasonCode == kFixMeWeHaventImplementedReasonCodeYet {
			logger.Info("default cloud credit has been assigned to premium account", "cloudAccountId", cloudAccountId)
			return nil
		}
	}

	currentDate := time.Now()
	newDate := currentDate.AddDate(0, 0, config.Cfg.PremiumDefaultCreditExpirationDays)
	expirationDate := fmt.Sprintf("%04d-%02d-%02d", newDate.Year(), newDate.Month(), newDate.Day())
	logger.V(9).Info("invoked CreateServiceCredit ", "cloudAccountId", cloudAccountId)
	return ariaController.CreateServiceCredit(ctx, cloudAccountId, config.Cfg.PremiumDefaultCreditAmount, expirationDate, defaultCreditAssignmentComment)
}

func (ariaController *AriaController) CreateServiceCredit(ctx context.Context, cloudAccountId string, amount float64, expirationDate string, comments string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.CreateServiceCredit").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	resp, err := ariaController.ariaServiceCreditClient.CreateServiceCredits(ctx, client.GetAccountClientId(cloudAccountId), amount, kFixMeWeHaventImplementedReasonCodeYet, expirationDate, comments)
	if err != nil {
		logger.Error(err, "failed to create service credit ", "resp", resp)
		return err
	}
	logger.Info("create service credit", "cloudAccountId", cloudAccountId)
	return nil
}

func (ariaController *AriaController) GetUnbilledUsage(ctx context.Context, cloudAccountId string) (float64, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.GetUnbilledUsage").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	acctPlans, err := ariaController.ariaAccountClient.GetAcctPlans(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		return 0, err
	}
	var unbilledUsage float64
	for _, acctPlan := range acctPlans.AcctPlansM {
		ubilledUsageSummary, err := ariaController.ariaUsageClient.GetUnbilledUsageSummary(ctx,
			client.GetAccountClientId(cloudAccountId), acctPlan.ClientMasterPlanInstanceId)
		if err != nil {
			return 0, err
		}
		unbilledUsage += ubilledUsageSummary.MtdBalanceAmount
	}
	return unbilledUsage, nil
}

func (ariaController *AriaController) GetUnAppliedServiceCredits(ctx context.Context, cloudAccountId string) (float32, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.GetUnAppliedServiceCredits").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	unAppliedServiceCreditsDetails, err := ariaController.ariaServiceCreditClient.GetUnappliedServiceCredits(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		logger.Error(err, "failed to get unapplied service credits")
		return 0, err
	}
	// first treat Aria as the source of truth.
	// if the credits have been applied to a invoice, then the unapplied amount is reflected in Aria.
	var unappliedAmount float32
	for _, unappliedCredits := range unAppliedServiceCreditsDetails.UnappliedServiceCreditsDetails {
		unappliedAmount += unappliedCredits.AmountLeftToApply
	}
	logger.Info("unapplied amount left to apply ", "unappliedAmount", unappliedAmount)
	acctPlans, err := ariaController.ariaAccountClient.GetAcctPlans(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		logger.Error(err, "failed to get account plans")
		return 0, err
	}
	for _, acctPlan := range acctPlans.AcctPlansM {
		ubilledUsageSummary, err := ariaController.ariaUsageClient.GetUnbilledUsageSummary(ctx,
			client.GetAccountClientId(cloudAccountId), acctPlan.ClientMasterPlanInstanceId)
		if err != nil {
			logger.Error(err, "failed to get unbilled usage history")
			return 0, err
		}
		unappliedAmount = unappliedAmount - float32(ubilledUsageSummary.MtdBalanceAmount)
	}
	logger.Info("unapplied amount after usages", "unappliedAmount", unappliedAmount)
	if unappliedAmount < 0 {
		unappliedAmount = 0
	}
	return unappliedAmount, nil
}

func (ariaController *AriaController) GetClientPlanDetails(ctx context.Context, clientPlanId string) (*response.GetPlanDetailResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.GetClientPlanDetails").WithValues("clientPlanId", clientPlanId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	resp, err := ariaController.ariaPlanClient.GetAriaPlanDetails(ctx, clientPlanId)
	logger.Info("plan details response", "resp", resp)
	if err != nil {
		logger.Error(err, "plan details response error")
		return resp, err
	}
	return resp, nil
}

// Todo: Why do we have two ways to check for the same thing?
// We are doing a get all and find one for creation of account.
// And we are doing a find one for the ensuring of the default plan. WHY>> :-()

func (ariaController *AriaController) ClientPlanDetailsExists(ctx context.Context, clientPlanId string) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.ClientPlanDetailsExists").WithValues("clientPlanId", clientPlanId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	clientPlanDtls, err := ariaController.GetClientPlanDetails(ctx, clientPlanId)
	if err != nil {
		return false, err
	}
	if clientPlanDtls.ErrorCode != 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (ariaController *AriaController) ensureDefaultPlan(ctx context.Context) error {
	planId := client.GetDefaultPlanClientId()
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.ensureDefaultPlan").WithValues("clientPlanId", planId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	exists, err := ariaController.ClientPlanDetailsExists(ctx, planId)
	if err != nil {
		if !strings.Contains(err.Error(), "error code:1010") {
			logger.Error(err, "client plan id plan get error", "planId", planId)
			return err
		}
	}
	if exists {
		return nil
	}
	logger.Info("default plan not found, creating")
	return ariaController.ariaPlanClient.CreateDefaultPlan(ctx)
}

func (ariaController *AriaController) getMinutesUsageTypeDetails(ctx context.Context) (*response.GetUsageTypeDetails, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.getMinutesUsageTypeDetails").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	usageTypeDetail, err := ariaController.ariaUsageTypeClient.GetUsageTypeDetails(ctx, client.GetMinsUsageTypeCode())
	if err != nil {
		const INVALID_USAGE_TYPE_CODE = 1010
		if usageTypeDetail != nil && usageTypeDetail.ErrorCode == INVALID_USAGE_TYPE_CODE {
			return nil, nil
		} else {
			logger.Error(err, "failed to get usage type details for the usage type code", "code", client.GetMinsUsageTypeCode())
			return nil, err
		}
	}
	return usageTypeDetail, nil
}

// ensure the minutes usage type.
// will check if the usage type of the right code exists.
// if it does exist - map to the expected set of values for the attributes.
// if the mapping of is not correct, then update.
// if does not exist - create with the expected set of values for the attributes.
func (ariaController *AriaController) ensureMinutesUsageType(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.ensureMinutesUsageType").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	// first check for the usage unit type of the type minutes
	usageUnitType, err := ariaController.ariaUsageTypeClient.GetMinuteUsageUnitType(ctx)
	if err != nil {
		logger.Error(err, "failed to get minute usage unit type")
		return err
	}
	if usageUnitType == nil {
		return errors.New("usage unit type of type minute does not exist")
	}
	usageTypeDetails, err := ariaController.getMinutesUsageTypeDetails(ctx)
	if err != nil {
		logger.Error(err, "failed to get minutes usage type details")
		return err
	}
	if usageTypeDetails == nil {
		logger.Info("minutes usage type does not exist")
		_, err := ariaController.ariaUsageTypeClient.CreateUsageType(ctx, client.USAGE_TYPE_NAME, client.GetMinsUsageTypeDesc(), usageUnitType.UsageUnitTypeNo, client.GetMinsUsageTypeCode())
		if err != nil {
			logger.Error(err, "failed to create usage type of type minues")
			return err
		}
	} else {
		if usageTypeDetails.UsageTypeName != client.USAGE_TYPE_NAME ||
			usageTypeDetails.UsageTypeDesc != client.GetMinsUsageTypeDesc() ||
			usageTypeDetails.UsageUnitType != usageUnitType.UsageUnitTypeDesc {
			_, err := ariaController.ariaUsageTypeClient.UpdateUsageType(ctx, client.USAGE_TYPE_NAME, client.GetMinsUsageTypeDesc(), usageUnitType.UsageUnitTypeNo, client.GetMinsUsageTypeCode())
			if err != nil {
				logger.Error(err, "failed to update usage type of the type minutes")
				return err
			}
		}
	}
	return nil
}

func (ariaController *AriaController) AddPaymentPreProcessing(ctx context.Context, cloudAccountId string) (string, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.AddPaymentPreProcessing").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	clientAccountId := client.GetAccountClientId(cloudAccountId)
	clientAcctGroupId := client.GetClientAcctGroupId()

	resp, err := ariaController.ariaPaymentClient.AssignCollectionsAccountGroup(context.Background(), clientAccountId, clientAcctGroupId)
	if err != nil {
		if resp != nil && resp.ErrorCode == 12004 {
			logger.Info("account already assigned to this group")
		} else {
			logger.Error(err, "failed to assign collection account group", "resp", resp)
			return "", err
		}
	}

	setSessionResp, err := ariaController.ariaPaymentClient.SetSession(ctx, clientAccountId)
	if err != nil {
		logger.Error(err, "failed to get session id ", "setSessionResp", setSessionResp)
		return "", err
	}
	logger.Info("get session id", "sessionId", setSessionResp.SessionId)
	logger.Info("exit: AddPaymentPreProcessing")

	return setSessionResp.SessionId, nil
}

func (ariaController *AriaController) AddPaymentPostProcessing(ctx context.Context, cloudAccountId string, primaryPaymentMethodNo int64) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.AddPaymentPostProcessing").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	clientBillingGroupResponse, err := ariaController.ariaAccountClient.GetBillingGroup(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		logger.Error(err, "failed to get account billing group")
		return err
	}

	//TODO: support for multiple billing groups
	clientBillingGroupId := clientBillingGroupResponse.BillingGroupDetails[0].ClientBillingGroupId
	logger.Info("clientBillingGroupId", "clientBillingGroupId", clientBillingGroupId)

	resp, err := ariaController.ariaPaymentClient.UpdateAccountBillingGroup(context.Background(), client.GetAccountClientId(cloudAccountId), clientBillingGroupId, primaryPaymentMethodNo)
	if err != nil {
		logger.Error(err, "failed to update account billing group", "resp", resp)
		return err
	}

	getPaymentMethodsResp, err := ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), client.GetAccountClientId(cloudAccountId))
	if err != nil {
		logger.Error(err, "failed to fetch account payment methods", "resp", getPaymentMethodsResp)
		return err
	}

	for index, _ := range getPaymentMethodsResp.AccountPaymentMethods {
		if getPaymentMethodsResp.AccountPaymentMethods[index].PaymentMethodNo != primaryPaymentMethodNo {
			_, err = ariaController.ariaPaymentClient.RemovePaymentMethod(context.Background(), client.GetAccountClientId(cloudAccountId), getPaymentMethodsResp.AccountPaymentMethods[index].PaymentMethodNo)
			if err != nil {
				logger.Error(err, "failed to remove payment method")
				return err
			}
		}
	}

	if config.Cfg.AuthorizationEnabled {
		logger.Info("authorization invoked")
		clientBillingGroupNo := clientBillingGroupResponse.BillingGroupDetails[0].BillingGroupNo
		logger.Info("authorize cc clientBillingGroupNo", "clientBillingGroupNo", clientBillingGroupNo)

		AuthorizeElectronicPaymentResponse, err := ariaController.ariaPaymentClient.AuthorizeElectronicPayment(context.Background(), client.GetAccountClientId(cloudAccountId), clientBillingGroupNo, creditCardAuthorizationAmount)
		if err != nil {
			logger.Error(err, "failed to authorize electronic payment", "resp", AuthorizeElectronicPaymentResponse)
			return err
		}
	}

	account, err := ariaController.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil {
		logger.Error(err, "error in cloud account client response")
		return err
	}
	upgradeStatus := account.UpgradedToPremium
	if upgradeStatus == pb.UpgradeStatus_UPGRADE_PENDING_CC {
		upgradeStatus = pb.UpgradeStatus_UPGRADE_PENDING_CC_VERIFIED
	}
	paidServicesAllowed := true
	lowCredits := false
	terminatePaidServices := false
	terminateMessageQueued := false

	_, err = ariaController.cloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{
		Id:                     cloudAccountId,
		PaidServicesAllowed:    &paidServicesAllowed,
		LowCredits:             &lowCredits,
		TerminatePaidServices:  &terminatePaidServices,
		TerminateMessageQueued: &terminateMessageQueued,
		UpgradedToPremium:      &upgradeStatus})
	if err != nil {
		logger.Error(err, "failed to update paid services allowed flag in cloud account")
		return err
	}
	logger.Info("credit card added successfully", "cloudAccountId", cloudAccountId)
	logger.Info("exit: AddPaymentPostProcessing")
	return nil
}

func (ariaController *AriaController) GetBillingOptions(ctx context.Context, cloudAccountId string) ([]data.AccountPaymentMethods, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.GetBillingOptions").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	resp, err := ariaController.ariaPaymentClient.GetPaymentMethods(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		logger.Error(err, "failed to get billing options ", "resp", resp)
		return nil, err
	}

	return resp.AccountPaymentMethods, nil
}

func (ariaController *AriaController) GetAcctNoFromCloudAcctId(ctx context.Context, cloudAccountId string) (int64, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.GetAcctNoFromCloudAcctId").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	resp, err := ariaController.ariaAccountClient.GetAccountNoFromUserId(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		logger.Error(err, "failed to get account number from cloud account id", "resp", resp)
		return 0, err
	}

	logger.Info("get account number from cloud account id ", "account no", resp.AcctNo)

	return resp.AcctNo, nil
}

func (ariaController *AriaController) GetAccountCredits(ctx context.Context, cloudAccountId string) ([]data.AllCredit, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.GetAccountCredits").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	accountNo, err := ariaController.GetAcctNoFromCloudAcctId(ctx, cloudAccountId)
	if err != nil || accountNo == 0 {
		logger.Error(err, "failed to get account no", "accountNo", accountNo)
		return nil, err
	}

	resp, err := ariaController.ariaAccountClient.GetAccountCredits(ctx, accountNo)
	if err != nil {
		logger.Error(err, "failed to get account credits", "resp", resp)
		return nil, err
	}

	logger.Info("get account credits - all", "acct_credits", resp.AllCredits)

	return resp.AllCredits, nil
}

func (ariaController *AriaController) GetAccountServiceCredits(ctx context.Context, cloudAccountId string) ([]data.AllCredit, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.GetAccountServiceCredits").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	accountNo, err := ariaController.GetAcctNoFromCloudAcctId(ctx, cloudAccountId)
	if err != nil || accountNo == 0 {
		logger.Error(err, "failed to get account no", "accountNo", accountNo)
		return nil, err
	}

	resp, err := ariaController.ariaAccountClient.GetAccountCredits(ctx, accountNo)
	if err != nil {
		logger.Error(err, "failed to get account credits", "resp", resp)
		return nil, err
	}
	logger.V(9).Info("get account credits - all", "acct_credits", resp.AllCredits)
	serviceCredits := make([]data.AllCredit, 0, len(resp.AllCredits))
	for _, credit := range resp.AllCredits {
		if credit.CreditType == SERVICE_CREDITS {
			serviceCredits = append(serviceCredits, credit)
		}
	}

	logger.V(9).Info("account service credits", "serviceCredits", serviceCredits)

	return serviceCredits, nil
}

func (ariaController *AriaController) GetCreditDetails(ctx context.Context, clientAcctId string, creditNo int64) (*response.GetCreditDetails, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.GetCreditDetails").WithValues("clientAcctId", clientAcctId).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	creditDetails, err := ariaController.ariaAccountClient.GetAccountCreditDetails(ctx, clientAcctId, creditNo)
	if err != nil {
		logger.Error(err, "failed to get credit details")
		return nil, err
	}
	logger.Info("got credit details", "details", creditDetails)
	return creditDetails, nil
}

func (ariaController *AriaController) SetAriaAccountNotifyTemplateGroup(ctx context.Context, cloudAccount *pb.CloudAccount) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.SetAriaAccountNotifyTemplateGroup").WithValues("cloudAccountId", cloudAccount.GetId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("set account notify template groupId", "cloudAccount", cloudAccount)
	if cloudAccount.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM || cloudAccount.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE {
		country := "US"
		if len(cloudAccount.CountryCode) != 0 {
			country = cloudAccount.CountryCode
		}
		clientAccountId := client.GetAccountClientId(cloudAccount.GetId())
		logger.V(1).Info("account notify template groupId", "accountType", cloudAccount.GetType(), "countryCode", country)
		clientNotificationTemplateGroupId := client.GetClientNotificationTemplateGroupId(cloudAccount.GetType(), country)
		respGet, err := ariaController.ariaAccountClient.GetAccountNotificationDetails(ctx, clientAccountId)
		if err != nil {
			if respGet != nil && respGet.ErrorCode != acctExistsError {
				logger.Error(err, "failed to get account notification details", "respGet", respGet)
				return err
			}
		}
		if respGet != nil && len(respGet.AccountNotificationDetails) > 0 {
			if client.ContainsNotificationTemplateGroupId(respGet.AccountNotificationDetails, clientNotificationTemplateGroupId) {
				logger.V(1).Info("client notification template group id already set", "clientNotificationTemplateGroupId", clientNotificationTemplateGroupId)
				return nil
			}
		}
		respSet, err := ariaController.ariaAccountClient.SetAccountNotifyTemplateGroup(ctx, clientAccountId, clientNotificationTemplateGroupId)
		if err != nil {
			logger.Error(err, "failed to set account notify template group id", "respSet", respSet)
			return err
		}
	}
	return nil
}

func (ariaController *AriaController) UpdateAriaAccountDunningGroup(ctx context.Context, cloudAccount *pb.CloudAccount) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.UpdateAriaAccountDunningGroup").WithValues("cloudAccountId", cloudAccount.GetId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	if cloudAccount.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE {
		logger.Info("update account dunning group", "cloudAccount", cloudAccount)
		clientAccountId := client.GetAccountClientId(cloudAccount.GetId())
		resp, err := ariaController.ariaAccountClient.UpdateAccountDunningGroup(ctx, clientAccountId, client.GetDunningGroupId(clientAccountId), client.ENTERPRISE_ClIENT_DUNNING_PROCESS_ID)
		if err != nil {
			logger.Error(err, "failed to update account dunning group", "resp", resp)
			return err
		}
	}
	return nil
}

func (ariaController *AriaController) UpdateAriaAccountContact(ctx context.Context, cloudAccount *pb.CloudAccount) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.UpdateAriaAccountContact").WithValues("cloudAccountId", cloudAccount.GetId()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("update account contact", "cloudAccount", cloudAccount)
	clientAccountId := client.GetAccountClientId(cloudAccount.GetId())
	resp, err := ariaController.ariaAccountClient.UpdateAccountContact(ctx, clientAccountId, cloudAccount.GetOwner())
	if err != nil {
		logger.Error(err, "failed to update account contact", "resp", resp)
		return err
	}
	return nil
}

func (ariaController *AriaController) DowngradePremiumtoStandard(ctx context.Context, cloudAccountId string, force bool) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.DowngradePremiumtoStandard").WithValues("cloudAccountId", cloudAccountId, "force", force).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// 1 verify account exist in aria
	acctDetails, err := ariaController.ariaAccountClient.GetAriaAccountDetailsAllForClientId(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		if acctDetails != nil && acctDetails.ErrorCode != acctExistsError {
			logger.Error(err, "failed to check if account exists")
			return status.Errorf(codes.NotFound, "failed to check if account exists")
		}
	}

	// 2  validation for unbilled usages and pending invoices
	if !force { // preconditions will be ignored if force

		unbilledUsage, err := ariaController.GetUnbilledUsage(ctx, cloudAccountId)
		if err != nil {
			logger.Error(err, "failed to get unbilled usage")
			return status.Errorf(codes.Internal, client.GetDriverError(FailedToGetUnbilledUsage, err).Error())
		}
		if unbilledUsage > 0 {
			logger.Error(err, "provided account have pending unbilled usages")
			return status.Errorf(codes.FailedPrecondition, "provided account have pending unbilled usages")
		}

		clientAcctId := client.GetAccountClientId(cloudAccountId)
		invoiceHistoryResponse, err := ariaController.ariaInvoiceClient.GetInvoiceHistory(ctx, clientAcctId, client.FILTER_CLIENT_MASTER_PLAN_ALL, "", "")
		if err != nil {
			logger.Error(err, "failed to get invoice history")
			return status.Errorf(codes.Internal, "failed to get invoice history")
		}

		for _, invoice := range invoiceHistoryResponse.InvoiceHist {
			if len(invoice.PaidDate) == 0 {
				logger.Error(err, "provided account has pending invoices")
				return status.Errorf(codes.Internal, "provided account has pending invoices")
			}
		}
	}

	/* we are skipping this step for now
	// 4 apply credits to account
	unAppliedCredits, err := ariaController.GetUnAppliedServiceCredits(ctx, cloudAccountId)
	if err != nil {
		logger.Error(err, "fail to get unAppliedCredits")
		return status.Errorf(codes.FailedPrecondition, "fail to get unAppliedCredits")
	}

	if unAppliedCredits > 0 {
		_, err = ariaController.ariaServiceCreditClient.ApplyCreditService(ctx, client.GetAccountClientId(cloudAccountId), float64(unAppliedCredits))
		if err != nil {
			logger.Error(err, "fail to apply credits")
			return status.Errorf(codes.FailedPrecondition, "fail to apply credits")
		}
	}
	*/

	// 5 delete the credit card
	getPaymentMethodsResp, err := ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), client.GetAccountClientId(cloudAccountId))
	if err != nil {
		logger.Error(err, "failed to fetch account payment methods", "resp", getPaymentMethodsResp)
		return err
	}

	for _, paymentMethod := range getPaymentMethodsResp.AccountPaymentMethods {
		paymentMethodNo := paymentMethod.PaymentMethodNo
		_, err = ariaController.ariaPaymentClient.RemovePaymentMethod(ctx, client.GetAccountClientId(cloudAccountId), paymentMethodNo)
		if err != nil {
			logger.Error(err, "failed to remove payment method to aria account", "paymentMethodNo", paymentMethodNo)
			return status.Errorf(codes.Internal, "failed to remove payment method %v to aria account", paymentMethodNo)
		}
	}

	logger.V(9).Info("cloud account payment methods were removed succesfully")

	// 6 deactivate the Aria account
	_, err = ariaController.ariaAccountClient.UpdateAriaAccountStatus(ctx, acctDetails.AcctNo2, 0)
	if err != nil {
		logger.Error(err, "failed to deactivate Aria Account", "id", cloudAccountId, "accountNo", acctDetails.AcctNo2)
		return status.Errorf(codes.Internal, "failed to deactivate Aria Account")
	}

	// 7 verify payments methods were removed on the aria account
	getPaymentMethodsResp, err = ariaController.ariaPaymentClient.GetPaymentMethods(context.Background(), client.GetAccountClientId(cloudAccountId))
	if err != nil {
		logger.Error(err, "failed to fetch account payment methods", "resp", getPaymentMethodsResp)
		return err
	}
	if len(getPaymentMethodsResp.AccountPaymentMethods) != 0 {
		logger.Error(errors.New("payment methods not removed in aria"), "aria account payment methods were no removed in aria on downgrading account premium to standard")
		return status.Errorf(codes.Internal, "aria account payment methods were no removed in aria on downgrading account premium to standard")
	}

	// 8 verify account was disabled in aria
	acctDetails, err = ariaController.ariaAccountClient.GetAriaAccountDetailsAllForClientId(ctx, client.GetAccountClientId(cloudAccountId))
	if err != nil {
		if acctDetails != nil && acctDetails.ErrorCode != acctExistsError {
			logger.Error(err, "failed to check if account exists")
			return status.Errorf(codes.NotFound, "failed to check if account exists")
		}
	}
	if acctDetails.StatusCd != 0 {
		logger.Error(errors.New("statusCd was not updated in aria"), "aria account status was not changed on downgrading account premium to standard")
		return status.Errorf(codes.Internal, "aria account status was not changed on downgrading account premium to standard")
	}

	logger.V(9).Info("cloud account deactivated succesfully on aria ", "acctDetails", acctDetails)

	return nil
}

func (ariaController *AriaController) getStorageUsageTypeDetails(ctx context.Context) (*response.GetUsageTypeDetails, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.getStorageUsageTypeDetails").Start()
	defer span.End()
	usageTypeDetail, err := ariaController.ariaUsageTypeClient.GetUsageTypeDetails(ctx, client.GetStorageUsageUnitTypeCode())
	if err != nil {
		const INVALID_USAGE_TYPE_CODE = 1010
		if usageTypeDetail != nil && usageTypeDetail.ErrorCode == INVALID_USAGE_TYPE_CODE {
			return nil, nil
		} else {
			logger.Error(err, "failed to get usage type details for the usage type code", "code", client.GetStorageUsageUnitTypeCode())
			return nil, err
		}
	}
	return usageTypeDetail, nil
}

func (ariaController *AriaController) getInferenceUsageTypeDetails(ctx context.Context) (*response.GetUsageTypeDetails, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.getInferenceUsageTypeDetails").Start()
	defer span.End()
	usageTypeDetail, err := ariaController.ariaUsageTypeClient.GetUsageTypeDetails(ctx, client.GetInferenceUsageUnitTypeCode())
	if err != nil {
		const INVALID_USAGE_TYPE_CODE = 1010
		if usageTypeDetail != nil && usageTypeDetail.ErrorCode == INVALID_USAGE_TYPE_CODE {
			return nil, nil
		} else {
			logger.Error(err, "failed to get usage type details for the usage type code", "code", client.GetInferenceUsageUnitTypeCode())
			return nil, err
		}
	}
	return usageTypeDetail, nil
}

func (ariaController *AriaController) getTokenUsageTypeDetails(ctx context.Context) (*response.GetUsageTypeDetails, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.getTokenUsageTypeDetails").Start()
	defer span.End()
	usageTypeDetail, err := ariaController.ariaUsageTypeClient.GetUsageTypeDetails(ctx, client.GetTokenUsageUnitTypeCode())
	if err != nil {
		const INVALID_USAGE_TYPE_CODE = 1010
		if usageTypeDetail != nil && usageTypeDetail.ErrorCode == INVALID_USAGE_TYPE_CODE {
			return nil, nil
		} else {
			logger.Error(err, "failed to get usage type details for the usage type code", "code", client.GetTokenUsageUnitTypeCode())
			return nil, err
		}
	}
	return usageTypeDetail, nil
}

// ensure the GB Per Hour usage type.
// will check if the usage type of the right code exists.
// if it does exist - map to the expected set of values for the attributes.
// if the mapping of is not correct, then update.
// if does not exist - create with the expected set of values for the attributes.
func (ariaController *AriaController) ensureStorageUsageType(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.ensureStorageUsageType").Start()
	defer span.End()
	// first check for the usage unit type of the type storage
	usageUnitType, err := ariaController.ariaUsageTypeClient.GetStorageUsageUnitType(ctx)
	if err != nil {
		logger.Error(err, "failed to get storage usage unit type", "usageUnitType", client.GetAriaSystemStorageUsageUnitTypeName())
		return err
	}
	if usageUnitType == nil {
		return errors.New("usage unit type of type storage does not exist")
	}
	usageTypeDetails, err := ariaController.getStorageUsageTypeDetails(ctx)
	if err != nil {
		logger.Error(err, "failed to get storage usage type details")
		return err
	}
	if usageTypeDetails == nil {
		logger.Info("storage usage type does not exist")
		_, err := ariaController.ariaUsageTypeClient.CreateUsageType(ctx, client.GetProductStorageUsageUnitTypeName(), client.GetStorageUsageUnitTypeDesc(), usageUnitType.UsageUnitTypeNo, client.GetStorageUsageUnitTypeCode())
		if err != nil {
			logger.Error(err, "failed to create usage type of type storage")
			return err
		}
	} else {
		if usageTypeDetails.UsageTypeName != client.GetProductStorageUsageUnitTypeName() ||
			usageTypeDetails.UsageTypeDesc != client.GetStorageUsageUnitTypeDesc() ||
			usageTypeDetails.UsageUnitType != usageUnitType.UsageUnitTypeDesc {
			_, err := ariaController.ariaUsageTypeClient.UpdateUsageType(ctx, client.GetProductStorageUsageUnitTypeName(), client.GetStorageUsageUnitTypeDesc(), usageUnitType.UsageUnitTypeNo, client.GetStorageUsageUnitTypeCode())
			if err != nil {
				logger.Error(err, "failed to update usage type of the type storage")
				return err
			}
		}
	}
	return nil
}

// ensure the token usage type.
// will check if the usage type of the right code exists.
// if it does exist - map to the expected set of values for the attributes.
// if the mapping of is not correct, then update.
// if does not exist - create with the expected set of values for the attributes.
func (ariaController *AriaController) ensureTokenUsageType(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.ensureTokenUsageType").Start()
	defer span.End()
	// first check for the usage unit type of the type token
	usageUnitType, err := ariaController.ariaUsageTypeClient.GetTokenUsageUnitType(ctx)
	if err != nil {
		logger.Error(err, "failed to get token usage unit type", "usageUnitType", client.GetAriaSystemTokenUsageUnitTypeName())
		return err
	}
	if usageUnitType == nil {
		return errors.New("usage unit type of type token does not exist")
	}
	usageTypeDetails, err := ariaController.getTokenUsageTypeDetails(ctx)
	if err != nil {
		logger.Error(err, "failed to get token usage type details")
		return err
	}
	if usageTypeDetails == nil {
		logger.Info("token usage type does not exist")
		_, err := ariaController.ariaUsageTypeClient.CreateUsageType(ctx, client.GetProductTokenUsageUnitTypeName(), client.GetTokenUsageUnitTypeDesc(), usageUnitType.UsageUnitTypeNo, client.GetTokenUsageUnitTypeCode())
		if err != nil {
			logger.Error(err, "failed to create usage type of type token")
			return err
		}
	} else {
		if usageTypeDetails.UsageTypeName != client.GetProductTokenUsageUnitTypeName() ||
			usageTypeDetails.UsageTypeDesc != client.GetTokenUsageUnitTypeDesc() ||
			usageTypeDetails.UsageUnitType != usageUnitType.UsageUnitTypeDesc {
			_, err := ariaController.ariaUsageTypeClient.UpdateUsageType(ctx, client.GetProductTokenUsageUnitTypeName(), client.GetTokenUsageUnitTypeDesc(), usageUnitType.UsageUnitTypeNo, client.GetTokenUsageUnitTypeCode())
			if err != nil {
				logger.Error(err, "failed to update usage type of the type token")
				return err
			}
		}
	}
	return nil
}

// ensure the dollar per inference usage type.
// will check if the usage type of the right code exists.
// if it does exist - map to the expected set of values for the attributes.
// if the mapping of is not correct, then update.
// if does not exist - create with the expected set of values for the attributes.
func (ariaController *AriaController) ensureInferenceUsageType(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaController.ensureInferenceUsageType").Start()
	defer span.End()
	// first check for the usage unit type of the type inference
	usageUnitType, err := ariaController.ariaUsageTypeClient.GetInferenceUsageUnitType(ctx)
	if err != nil {
		logger.Error(err, "failed to get inference usage unit type", "usageUnitType", client.GetAriaSystemInferenceUsageUnitTypeName())
		return err
	}
	if usageUnitType == nil {
		return errors.New("usage unit type of type inference does not exist")
	}
	usageTypeDetails, err := ariaController.getInferenceUsageTypeDetails(ctx)
	if err != nil {
		logger.Error(err, "failed to get inference usage type details")
		return err
	}
	if usageTypeDetails == nil {
		logger.Info("inference usage type does not exist")
		_, err := ariaController.ariaUsageTypeClient.CreateUsageType(ctx, client.GetProductInferenceUsageUnitTypeName(), client.GetInferenceUsageUnitTypeDesc(), usageUnitType.UsageUnitTypeNo, client.GetInferenceUsageUnitTypeCode())
		if err != nil {
			logger.Error(err, "failed to create usage type of type inference")
			return err
		}
	} else {
		if usageTypeDetails.UsageTypeName != client.GetProductInferenceUsageUnitTypeName() ||
			usageTypeDetails.UsageTypeDesc != client.GetInferenceUsageUnitTypeDesc() ||
			usageTypeDetails.UsageUnitType != usageUnitType.UsageUnitTypeDesc {
			_, err := ariaController.ariaUsageTypeClient.UpdateUsageType(ctx, client.GetProductInferenceUsageUnitTypeName(), client.GetInferenceUsageUnitTypeDesc(), usageUnitType.UsageUnitTypeNo, client.GetInferenceUsageUnitTypeCode())
			if err != nil {
				logger.Error(err, "failed to update usage type of the type inference")
				return err
			}
		}
	}
	return nil
}
