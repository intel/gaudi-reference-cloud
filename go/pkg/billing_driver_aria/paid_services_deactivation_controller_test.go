// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"fmt"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"gotest.tools/assert"
)

func TestDetermineMessageTypeAndAccountId(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestDetermineMessageTypeAndAccountId")
	failedEventMsg := "version=3%2E3&sender=A&transaction_id=701817463&action=A&class=T&auth_key=&client_receipt_id=&client_no=5025555&acct_no=33798544&userid=test%2E070954857102&senior_acct_no=&plan_instance_no%5B%5D=&client_plan_instance_id%5B%5D=&resp_cd%5B%5D=&responsibility_level_code%5B%5D=&resp_plan_instance_no%5B%5D=&financial_trans_id=521732802&financial_trans_granular_id=215864208&financial_trans_type_no=3&financial_trans_type_label=Electronic%20Payment&financial_trans_gl_type=C&financial_trans_date=2023-10-09%2005%3A42%3A41&financial_trans_status_label=&financial_trans_status_desc=&financial_trans_client_notes=&financial_trans_amount=204%2E66&financial_trans_applied_amount=204%2E66&charge_trans_id%5B%5D=521732803&payment_trans_id%5B%5D=521732802&applied_amount%5B%5D=204%2E66&applied_trans_id%5B%5D=521732803&applied_trans_type_no%5B%5D=-3&applied_trans_type_label%5B%5D=Voided%20Electronic%20Payment&future_manual_allocation=0&event_id%5B%5D=905&event_label%5B%5D=Electronic%20Payment%20Created&event_id%5B%5D=947&event_label%5B%5D=Electronic%20Payment%20Failed"
	formattedFailedEventMsg := fmt.Sprintf("%s&client_acct_id=%s%s", failedEventMsg, config.Cfg.ClientIdPrefix, "%2E070954857102")
	paidServicesDeactivationController := NewPaidServicesDeactivationController(nil, nil, AriaService.cloudAccountClient)
	messageType, cloudAcctId, dunningState := paidServicesDeactivationController.determineMessageTypeAndAccountId(ctx, formattedFailedEventMsg)
	logger.Info("message type and cloud account id for payment failed", "messageType", messageType, "accountId", cloudAcctId, "dunningState", dunningState)
	assert.Equal(t, messageType, "FailedPayment")
	dunningEventMsg := "version=3%2E2&sender=A&transaction_id=341428564&action=M&class_name=A&auth_key=&client_receipt_id=&client_no=3760759&acct_no=38184591&userid=amardee2%2E883035491865&password=$SH1$5000$8$565AE6860AD5FDBF$32$709D0404504F3B66850C496CBCF5FE9CB2649E6A74E09B879B53F05901E28BB9&pin=&secret_question=&secret_answer=&status_cd=1&senior_acct_no=&notify_method=10&currency=usd&test_acct=N&last_acct_comment=Statement%20sending%20for%20invoice%20no%2E%20257673829%20suppressed%20due%20to%20-%20invoice%20do%20not%20have%20selected%20transactions%20to%20send&acct_contact_first_name=&acct_contact_middle_initial=&acct_contact_last_name=&acct_contact_company_name=&acct_contact_address1=&acct_contact_address2=&acct_contact_address3=&acct_contact_city=&acct_contact_state_prov=&acct_contact_locality=&acct_contact_postal_code=&acct_contact_country=US&acct_contact_phone=&acct_contact_phone_ext=&acct_contact_work_phone=&acct_contact_work_phone_ext=&acct_contact_cell_phone=&acct_contact_email=&acct_contact_fax=&master_plan_instance_no%5B%5D=53144265&client_master_plan_instance_id%5B%5D=amardee2%2E883035491865%2E3404829f-eaae-4e75-93d9-51803097b305&mpi_plan_no%5B%5D=10443632&mpi_client_plan_id%5B%5D=amardee2%2E3404829f-eaae-4e75-93d9-51803097b305&mpi_plan_name%5B%5D=Test%20Product%2FPlan2285abb8-9&mpi_plan_activation_date%5B%5D=2023-10-12&mpi_plan_termination_date%5B%5D=&mpi_status_cd%5B%5D=1&mpi_queued_status_cd%5B%5D=&mpi_resp_level_cd%5B%5D=1&mpi_resp_plan_instance_no%5B%5D=&mpi_plan_units%5B%5D=1&mpi_queued_units%5B%5D=&mpi_units_change_date%5B%5D=&mpi_prov_reason_cd%5B%5D=53144265&mpi_promo_cd%5B%5D=&mpi_transistion_from_plan%5B%5D=&mpi_transistion_to_plan%5B%5D=&mpi_billing_group_no%5B%5D=19847774&mpi_client_billing_group_id%5B%5D=amardee2%2E883035491865%2Ebilling_group&mpi_dunning_group_no%5B%5D=21947300&mpi_client_dunning_group_id%5B%5D=amardee2%2E883035491865%2Edunning_group&mpi_dunning_state%5B%5D=1&mpi_dunning_step%5B%5D=1&mpi_bill_day%5B%5D=12&mpi_created_date%5B%5D=2023-10-12&mpi_next_bill_date%5B%5D=2023-11-12&mpi_last_bill_date%5B%5D=2023-10-12&mpi_recurring_bill_thru_date%5B%5D=2023-11-11&mpi_usage_bill_thru_date%5B%5D=2023-10-11&mpi_plan_date%5B%5D=2023-10-12&mpi_status_date%5B%5D=2023-10-12&mpi_next_dunning_date%5B%5D=2023-10-13&mpi_queued_status_change_date%5B%5D=&mpi_bill_lag_days%5B%5D=7&mpi_pui_no%5B%5D=56965303&mpi_pui_cd_id%5B%5D=56965303&mpi_pui_status%5B%5D=Active&event_id%5B%5D=744&event_label%5B%5D=Account%20Master%20Plan%20Instance%20Dunning%20Degrade%20Date%20Changed"
	formatteddunningEventMsg := fmt.Sprintf("%s&client_acct_id=%s%s", dunningEventMsg, config.Cfg.ClientIdPrefix, "%2E070954857102")
	messageType, cloudAcctId, dunningState = paidServicesDeactivationController.determineMessageTypeAndAccountId(ctx, formatteddunningEventMsg)
	logger.Info("message type and cloud account id for dunning", "messageType", messageType, "accountId", cloudAcctId, "dunningState", dunningState)
	assert.Equal(t, messageType, "Dunning")
}
