// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	UpdateCloudAcctCreditsDepletedError string = "failed to update cloud account for credits depleted"
	GetBillingOptionsError              string = "failed to get billing options"
	UpdateCloudAcctDisablePaidError     string = "failed to update cloud account for disable paid"
	BillLagDays                                = 7
)

type SchedulerCloudAccountState struct {
	AccessTimestamp string
	Mutex           sync.Mutex
}
type CloudAccountLocks struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func NewCloudAccountLocks(ctx context.Context) *CloudAccountLocks {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("NewCloudAccountLocks").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	return &CloudAccountLocks{
		locks: make(map[string]*sync.Mutex),
	}
}

func GetSchedulerError(schedulerError string, err error) error {
	return fmt.Errorf("scheduler error:%s,error:%w", schedulerError, err)
}

func GetBillingOptions(ctx context.Context, cloudAcct *pb.CloudAccount) ([]*pb.BillingOption, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BillingUtil.GetBillingOptions").WithValues("cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	driver, err := GetDriver(ctx, cloudAcct.Id)
	if err != nil {
		return nil, err
	}
	var billingOptions []*pb.BillingOption
	billingOption, err := driver.billingOption.Read(ctx, &pb.BillingOptionFilter{CloudAccountId: &cloudAcct.Id})
	if err != nil {
		return nil, err
	}
	billingOptions = append(billingOptions, billingOption)
	return billingOptions, nil
}

func CheckIfOptionsHasCreditCard(ctx context.Context, billingOptions []*pb.BillingOption) bool {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BillingUtil.CheckIfOptionsHasCreditCard").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	hasCreditCard := false
	for _, billingOption := range billingOptions {
		if billingOption.PaymentType == pb.PaymentType_PAYMENT_CREDIT_CARD {
			hasCreditCard = true
		}
	}
	logger.V(9).Info("billing option has credit card", "hasCreditCard", hasCreditCard)
	return hasCreditCard
}

// todo have to sort out testing enterprise accounts with direct debit.
func CheckIfOptionsHasDirectDebit(billingOptions []*pb.BillingOption) bool {
	return false
}

func UpdateCloudAcctCreditsDepleted(ctx context.Context, cloudAcct *pb.CloudAccount) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BillingUtil.UpdateCloudAcctCreditsDepleted").WithValues("cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.V(9).Info("update cloud account for credits depletion if depletion has not happened")
	// mark credits depleted when found
	if cloudAcct.CreditsDepleted != nil {
		if cloudAcct.CreditsDepleted.AsTime().Unix() == 0 {
			logger.Info("updating cloud account for credits depletion because depletion happened")
			_, err := cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, CreditsDepleted: timestamppb.New(time.Now())})
			if err != nil {
				logger.Error(err, "failed to update cloud account depletion")
				return err
			}
		}

	}
	return nil
}

func UpdateCloudAcctDisablePaid(ctx context.Context, cloudAcct *pb.CloudAccount) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BillingUtil.UpdateCloudAcctDisablePaid").WithValues("cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	if cloudAcct.PaidServicesAllowed || !cloudAcct.LowCredits {
		paidServicesAllowed := false
		lowCredits := true
		_, err := cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{
			Id:                  cloudAcct.Id,
			PaidServicesAllowed: &paidServicesAllowed,
			LowCredits:          &lowCredits})
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateCloudAcctNoCredits(ctx context.Context, cloudAcct *pb.CloudAccount) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BillingUtil.UpdateCloudAcctNoCredits").WithValues("cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.V(9).Info("updating cloud acct with no credits", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
	logger.V(9).Info("updating cloud acct depletion", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
	err := UpdateCloudAcctCreditsDepleted(ctx, cloudAcct)
	if err != nil {
		logger.Error(err, "updating cloud acct with no credits failed", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
		return GetSchedulerError(UpdateCloudAcctCreditsDepletedError, err)
	}
	if (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_STANDARD) || (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_INTEL) {
		logger.V(9).Info("updating cloud acct disable paid", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
		err = UpdateCloudAcctDisablePaid(ctx, cloudAcct)
		if err != nil {
			return GetSchedulerError(UpdateCloudAcctDisablePaidError, err)
		}
	} else if (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM) ||
		(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE) {
		billingOptions, err := GetBillingOptions(ctx, cloudAcct)
		if err != nil {
			return GetSchedulerError(GetBillingOptionsError, err)
		}
		if (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM &&
			!CheckIfOptionsHasCreditCard(ctx, billingOptions)) ||
			(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE &&
				!CheckIfOptionsHasDirectDebit(billingOptions)) {
			logger.V(9).Info("updating cloud acct disable paid", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
			err = UpdateCloudAcctDisablePaid(ctx, cloudAcct)
			if err != nil {
				return GetSchedulerError(UpdateCloudAcctDisablePaidError, err)
			}
		}
	}
	return nil
}

func UpdateCloudAcctHasCredits(ctx context.Context, cloudAcct *pb.CloudAccount, hasCredit bool) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BillingUtil.UpdateCloudAcctHasCredits").WithValues("cloudAccountId", cloudAcct.Id, "paidServicesAllowed", cloudAcct.GetPaidServicesAllowed(), "terminatePaidServices", cloudAcct.GetTerminatePaidServices()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	if hasCredit && cloudAcct.CreditsDepleted.AsTime().Unix() != 0 {
		_, err := cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, CreditsDepleted: timestamppb.New(time.Unix(0, 0))})
		if err != nil {
			return err
		}
	}

	if (cloudAcct.GetTerminatePaidServices() || cloudAcct.GetTerminateMessageQueued() || !cloudAcct.GetPaidServicesAllowed()) &&
		(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_STANDARD || cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_INTEL || (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM && cloudAcct.Enrolled)) {
		logger.V(9).Info("updating terminate paid false, terminate mq false, paid services allowed true for cloud acct",
			"cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type)
		terminatePaidServices := cloudAcct.GetTerminatePaidServices()
		// todo: this should clear out the message from the Q
		terminateMessageQueued := false
		paidServicesAllowed := cloudAcct.GetPaidServicesAllowed()
		if !cloudAcct.GetPaidServicesAllowed() && hasCredit {
			paidServicesAllowed = true
			terminatePaidServices = false
		}
		logger.V(9).Info("cloud account update", "id", cloudAcct.Id, "paidServicesAllowed", paidServicesAllowed, "terminatePaidServices", terminatePaidServices, "terminateMessageQueued", terminateMessageQueued)
		_, err := cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{
			Id:                     cloudAcct.Id,
			TerminatePaidServices:  &terminatePaidServices,
			TerminateMessageQueued: &terminateMessageQueued,
			PaidServicesAllowed:    &paidServicesAllowed,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func IsCloudAccountWithCreditCard(ctx context.Context, cloudAcct *pb.CloudAccount) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BillingUtil.IsCloudAccountWithCreditCard").WithValues("cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	if cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		billingOptions, err := GetBillingOptions(ctx, cloudAcct)
		if err != nil {
			err = GetSchedulerError(GetBillingOptionsError, err)
			logger.Error(err, "error in getting billing options")
			return true, nil
		}
		logger.V(9).Info("premium account", "billingOptions", billingOptions)
		if CheckIfOptionsHasCreditCard(ctx, billingOptions) {
			return true, nil
		}
	}
	return false, nil
}

func GetAccountTypes(ctx context.Context, configuredAccountTypes []string) []pb.AccountType {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BillingUtil.GetAccountTypes").Start()
	defer span.End()
	accountTypes := []pb.AccountType{}
	for _, configuredAccountType := range configuredAccountTypes {
		switch configuredAccountType {
		case pb.AccountType_ACCOUNT_TYPE_PREMIUM.String():
			accountTypes = append(accountTypes, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
		case pb.AccountType_ACCOUNT_TYPE_STANDARD.String():
			accountTypes = append(accountTypes, pb.AccountType_ACCOUNT_TYPE_STANDARD)
		case pb.AccountType_ACCOUNT_TYPE_INTEL.String():
			accountTypes = append(accountTypes, pb.AccountType_ACCOUNT_TYPE_INTEL)
		default:
			logger.Error(fmt.Errorf("invalid account type %v", configuredAccountType), "invalid account type")
		}
	}
	return accountTypes
}
func (cloudAccountLocks *CloudAccountLocks) Lock(ctx context.Context, cloudAccountID string) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("Lock").WithValues("cloudAccountID", cloudAccountID).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	cloudAccountLocks.mu.Lock()

	if _, exists := cloudAccountLocks.locks[cloudAccountID]; !exists {
		cloudAccountLocks.locks[cloudAccountID] = &sync.Mutex{}
	}
	cloudAccountLocks.mu.Unlock()
	cloudAccountLocks.locks[cloudAccountID].Lock()
}

func (cloudAccountLocks *CloudAccountLocks) Unlock(ctx context.Context, cloudAccountID string) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("Unlock").WithValues("cloudAccountID", cloudAccountID).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	cloudAccountLocks.mu.Lock()
	defer cloudAccountLocks.mu.Unlock()

	if lock, exists := cloudAccountLocks.locks[cloudAccountID]; exists {
		lock.Unlock()
	}
}
