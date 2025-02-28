// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import "fmt"

const (
	FailedToReadCloudAccount                         = "FAILED_TO_READ_CLOUD_ACCT"
	InvalidCloudAccountId                            = "INVALID_CLOUD_ACCT_ID"
	InvalidCouponCode                                = "INVALID_COUPON_CODE"
	InvalidBillingCreditAmount                       = "INVALID_BILLING_CREDIT_AMOUNT"
	InvalidBillingCreditExpiration                   = "INVALID_BILLING_CREDIT_EXPIRATION"
	InvalidBillingCreditReason                       = "INVALID_BILLING_CREDIT_REASON"
	FailedToLookUpCloudAccount                       = "FAILED_TO_LOOK_UP_CLOUD_ACCT"
	InvalidCloudAcct                                 = "INVALID_CLOUD_ACCOUNT"
	FailedToCreateBillingAccount                     = "FAILED_TO_CREATE_BILLING_ACCT"
	FailedToDowngradeBillingAccountPremiumToStandard = "FAILED_TO_DOWNGRADE_BILLING_ACCT_PREMIUM_TO_STANDARD"
	FailedToCreateBillingCredit                      = "FAILED_TO_CREATE_BILLING_CREDIT"
	FailedToMigrateCredits                           = "FAILED_TO_MIGRATE_CREDITS"
	FailedToReadBillingCredit                        = "FAILED_TO_READ_BILLING_CREDIT"
	FailedToReadUnappliedCreditBalance               = "FAILED_TO_READ_UNAPPLIED_CREDIT_BALANCE"
	InvalidSchedulerOpsAction                        = "INVALID_SCHEDULER_OPS_ACTION"
	FailedToDeleteMigratedCredits                    = "FAILED_TO_DELETE_MIGRATED_CREDITS"
	FailedToGetUsages                                = "FAILED_TO_GET_USAGES"
	FailedToParseFilter                              = "FAILED_TO_PARSE_FILTER"
	FailedToGetUsagesProduct                         = "FAILED_TO_GET_USAGE_PRODUCT"
)

func GetBillingError(billingApiError string, err error) error {
	return fmt.Errorf("billing api error:%s,service error:%v", billingApiError, err)
}

func GetBillingInternalError(billingInternalError string, err error) error {
	return fmt.Errorf("billing internal error:%s,service error:%v", billingInternalError, err)
}
