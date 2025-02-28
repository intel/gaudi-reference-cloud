// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CloudAccountSvcClient struct {
	CloudAccountClient pb.CloudAccountServiceClient
}

func NewCloudAccountClient(ctx context.Context, resolver grpcutil.Resolver) (*CloudAccountSvcClient, error) {
	logger := log.FromContext(ctx).WithName("CloudAccountClient.NewCloudAccountClient")
	var cloudAccountConn *grpc.ClientConn
	cloudAccountAddr, err := resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		logger.Error(err, "grpc resolver not able to resolve", "addr", cloudAccountAddr)
		return nil, err
	}
	cloudAccountConn, err = grpcConnect(ctx, cloudAccountAddr)
	if err != nil {
		return nil, err
	}
	ca := pb.NewCloudAccountServiceClient(cloudAccountConn)
	return &CloudAccountSvcClient{CloudAccountClient: ca}, nil
}

func NewCloudAccountClientForTest(cloudAccountServiceClient pb.CloudAccountServiceClient) *CloudAccountSvcClient {
	return &CloudAccountSvcClient{CloudAccountClient: cloudAccountServiceClient}
}

func (cloudAccount *CloudAccountSvcClient) GetCloudAccountType(ctx context.Context, accountId *pb.CloudAccountId) (pb.AccountType, error) {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.GetCloudAccountType")
	account, err := cloudAccount.CloudAccountClient.GetById(ctx, accountId)
	if err != nil {
		logger.Error(err, "error in cloud account client response")
		return pb.AccountType_ACCOUNT_TYPE_UNSPECIFIED, err
	}
	logger.Info("cloudaccount response", "account", account)
	return account.GetType(), nil
}

func (cloudAccount *CloudAccountSvcClient) GetAllCloudAccount(ctx context.Context) ([]*pb.CloudAccount, error) {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.GetAllCloudAccount")
	cloudAccounts := []*pb.CloudAccount{}
	accStream, err := cloudAccount.CloudAccountClient.Search(ctx, &pb.CloudAccountFilter{})
	if err != nil {
		logger.Error(err, "error in cloud account client response")
		return nil, err
	}
	done := make(chan bool)
	go func() {
		for {
			currAcc, err := accStream.Recv()
			if err == io.EOF {
				done <- true //close(done)
				return
			}
			if err != nil {
				logger.Error(err, "failed to read from stream")
				break
			}
			cloudAccounts = append(cloudAccounts, currAcc)
		}
	}()

	<-done

	logger.Info("cloudaccount response", "# cloudaccount ", len(cloudAccounts))
	return cloudAccounts, nil
}

func (cloudAccount *CloudAccountSvcClient) UpdateCloudAcctCreditsDepleted(ctx context.Context, cloudAcct *pb.CloudAccount) error {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.UpdateCloudAcctCreditsDepleted")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("update cloud account", "cloudAcct", cloudAcct)
	// mark credits depleted when found
	if cloudAcct.CreditsDepleted != nil {
		if cloudAcct.CreditsDepleted.AsTime().Unix() == 0 {
			_, err := cloudAccount.CloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, CreditsDepleted: timestamppb.New(time.Now())})
			if err != nil {
				logger.Error(err, "error in updating credits depleted")
				return err
			}
		}
	}
	return nil
}

func (cloudAccount *CloudAccountSvcClient) UpdateCloudAcctDisablePaid(ctx context.Context, cloudAcct *pb.CloudAccount) error {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.UpdateCloudAcctDisablePaid")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	if cloudAcct.PaidServicesAllowed || !cloudAcct.LowCredits {
		paidServicesAllowed := false
		lowCredits := true
		_, err := cloudAccount.CloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{
			Id:                  cloudAcct.Id,
			PaidServicesAllowed: &paidServicesAllowed,
			LowCredits:          &lowCredits})
		if err != nil {
			return err
		}
	}
	return nil
}

func (cloudAccount *CloudAccountSvcClient) UpdateCloudAcctNoCredits(ctx context.Context, cloudAcct *pb.CloudAccount, driver *BillingDriverClients) error {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.UpdateCloudAcctNoCredits")
	logger.Info("BEGIN")
	defer logger.Info("End")
	logger.V(1).Info("updating cloud acct with no credits", "id", cloudAcct.Id, "type", cloudAcct.Type)

	err := cloudAccount.UpdateCloudAcctCreditsDepleted(ctx, cloudAcct)
	if err != nil {
		logger.Error(err, "error updating cloud account credits depleted", "id", cloudAcct.Id, "type", cloudAcct.Type)
		return GetSchedulerError("failed to update cloud account for credits depleted", err)
	}
	if (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_STANDARD) || (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_INTEL) {
		err = cloudAccount.UpdateCloudAcctDisablePaid(ctx, cloudAcct)
		if err != nil {
			logger.Error(err, "error updating cloud account PaidServicesAllowed and LowCredits", "cloudAcct", cloudAcct)
			return GetSchedulerError("failed to update cloud account for disablinf paid services", err)
		}
	} else if (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM) ||
		(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE) {
		billingOptions, err := cloudAccount.GetBillingOptions(ctx, cloudAcct, driver)
		if err != nil {
			logger.Error(err, "error in get billing options", "cloudAcct", cloudAcct)
			return GetSchedulerError("error in getting billing options", err)
		}
		if (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM &&
			!CheckIfOptionsHasCreditCard(ctx, billingOptions)) ||
			(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE &&
				!CheckIfOptionsHasDirectDebit(billingOptions)) {
			err = cloudAccount.UpdateCloudAcctDisablePaid(ctx, cloudAcct)
			if err != nil {
				logger.Error(err, "error updating cloud account PaidServicesAllowed and LowCredits", "cloudAcct", cloudAcct)
				return GetSchedulerError("failed to update cloud account for disablinf paid services", err)
			}
		}
	}
	return nil
}

func (cloudAccount *CloudAccountSvcClient) UpdateCloudAcctHasCredits(ctx context.Context, cloudAcct *pb.CloudAccount) error {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.UpdateCloudAcctHasCredits")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("updating cloud acct has credits", "id", cloudAcct.Id)

	lowCredits := false
	_, err := cloudAccount.CloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, LowCredits: &lowCredits, CreditsDepleted: &timestamppb.Timestamp{Seconds: 0, Nanos: 0}})
	if err != nil {
		logger.Error(err, "error in cloud account client response")
		return err
	}

	if (cloudAcct.GetTerminatePaidServices() || cloudAcct.GetTerminateMessageQueued() || !cloudAcct.GetPaidServicesAllowed()) &&
		(cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_STANDARD || cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_INTEL || (cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM && cloudAcct.Enrolled)) {
		logger.V(1).Info("updating terminate paid false, terminate mq false, paid services allowed true for cloud acct",
			"id", cloudAcct.Id, "type", cloudAcct.Type)
		terminatePaidServices := cloudAcct.GetTerminatePaidServices()
		// todo: this should clear out the message from the queue
		terminateMessageQueued := false
		paidServicesAllowed := cloudAcct.GetPaidServicesAllowed()
		if !cloudAcct.GetPaidServicesAllowed() {
			paidServicesAllowed = true
			terminatePaidServices = false
		}
		logger.V(1).Info("cloud account update", "id", cloudAcct.Id, "paidServicesAllowed", paidServicesAllowed, "terminatePaidServices", terminatePaidServices, "terminateMessageQueued", terminateMessageQueued)
		_, err := cloudAccount.CloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{
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

func (cloudAccount *CloudAccountSvcClient) SearchCloudAccounts(ctx context.Context, filter *pb.CloudAccountFilter) ([]*pb.CloudAccount, error) {
	cloudAccts := []*pb.CloudAccount{}
	res, err := cloudAccount.CloudAccountClient.Search(ctx, filter)
	if err != nil {
		return nil, err
	}
	for {
		acct, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		cloudAccts = append(cloudAccts, acct)
	}
	return cloudAccts, nil
}

func (cloudAccount *CloudAccountSvcClient) GetCloudAcctsOfTypes(ctx context.Context, accountTypes []pb.AccountType) ([]*pb.CloudAccount, error) {
	cloudAccts := []*pb.CloudAccount{}
	for _, accountType := range accountTypes {
		filter := &pb.CloudAccountFilter{Type: &accountType}
		res, err := cloudAccount.SearchCloudAccounts(ctx, filter)
		if err != nil {
			return nil, err
		}
		cloudAccts = append(cloudAccts, res...)
	}
	return cloudAccts, nil
}

func (cloudAccount *CloudAccountSvcClient) GetCloudAccountsWithCredits(ctx context.Context) ([]*pb.CloudAccount, error) {
	accountTypes := []pb.AccountType{pb.AccountType_ACCOUNT_TYPE_STANDARD, pb.AccountType_ACCOUNT_TYPE_PREMIUM, pb.AccountType_ACCOUNT_TYPE_INTEL}
	return cloudAccount.GetCloudAcctsOfTypes(ctx, accountTypes)
}

func (cloudAccount *CloudAccountSvcClient) GetBillingOptions(ctx context.Context, cloudAcct *pb.CloudAccount, driver *BillingDriverClients) ([]*pb.BillingOption, error) {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.GetBillingOptions")
	logger.Info("BEGIN")
	defer logger.Info("End")
	var billingOptions []*pb.BillingOption
	billingOption, err := driver.BillingOption.Read(ctx, &pb.BillingOptionFilter{CloudAccountId: &cloudAcct.Id})
	if err != nil {
		logger.Error(err, "Error in reading Billing Option")
		return nil, err
	}
	billingOptions = append(billingOptions, billingOption)
	return billingOptions, nil
}

func (cloudAccount *CloudAccountSvcClient) IsCloudAccountWithCreditCard(ctx context.Context, cloudAcct *pb.CloudAccount, driver *BillingDriverClients) (bool, error) {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.HasCreditCard")
	logger.Info("BEGIN")
	defer logger.Info("END")
	if cloudAcct.Type == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		billingOptions, err := cloudAccount.GetBillingOptions(ctx, cloudAcct, driver)
		if err != nil {
			err = GetSchedulerError("error in getting billing options", err)
			logger.Error(err, "error in getting billing options")
			return true, nil
		}
		logger.V(1).Info("premium account", "billingOptions", billingOptions)
		if CheckIfOptionsHasCreditCard(ctx, billingOptions) {
			return true, nil
		}
	}
	return false, nil
}

func CheckIfOptionsHasCreditCard(ctx context.Context, billingOptions []*pb.BillingOption) bool {
	logger := log.FromContext(ctx).WithName("CloudAccountSvcClient.CheckIfOptionsHasCreditCard")
	logger.Info("BEGIN")
	defer logger.Info("End")
	hasCreditCard := false
	for _, billingOption := range billingOptions {
		if billingOption.PaymentType == pb.PaymentType_PAYMENT_CREDIT_CARD {
			hasCreditCard = true
		}
	}
	logger.V(1).Info("billing option has credit card", "hasCreditCard", hasCreditCard)
	return hasCreditCard
}

// todo have to sort out testing enterprise accounts with direct debit.
func CheckIfOptionsHasDirectDebit(billingOptions []*pb.BillingOption) bool {
	return false
}
func (cloudAccount *CloudAccountSvcClient) GetCloudAcctsOfTypesWithFilter(ctx context.Context, accountTypes []pb.AccountType, filter *pb.CloudAccountFilter) ([]*pb.CloudAccount, error) {
	cloudAccts := []*pb.CloudAccount{}
	for _, accountType := range accountTypes {
		filter.Type = &accountType
		res, err := cloudAccount.SearchCloudAccounts(ctx, filter)
		if err != nil {
			return nil, err
		}
		cloudAccts = append(cloudAccts, res...)
	}
	return cloudAccts, nil
}

func (cloudAccount *CloudAccountSvcClient) GetStandardAndIntelAccounts(ctx context.Context) []*pb.CloudAccount {
	logger := log.FromContext(ctx).WithName("cloudAccount.GetStandardAndIntelAccounts")
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	acctTypes := []pb.AccountType{pb.AccountType_ACCOUNT_TYPE_STANDARD, pb.AccountType_ACCOUNT_TYPE_INTEL}
	cloudAccts := []*pb.CloudAccount{}

	for _, acctType := range acctTypes {
		cloudAccountSearchClient, err :=
			cloudAccount.CloudAccountClient.Search(ctx, &pb.CloudAccountFilter{Type: &acctType})
		if err != nil {
			logger.Error(err, "failed to get cloud account client for searching on", "accountType", acctType.String())
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
