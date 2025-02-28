// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	mock "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/aria_mock_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type error_resp struct {
	ErrorCode int64  `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

// Type checkers
func CheckJsonType(w http.ResponseWriter, content_type string) {
	if content_type != "application/json" {
		err := error_resp{
			ErrorCode: 1004,
			ErrorMsg:  "File type not supported",
		}
		err_str := fmt.Sprintf("%#v", err)
		http.Error(w, err_str, http.StatusBadGateway)
	}
}
func CheckFormEncodedType(w http.ResponseWriter, content_type string) {
	if content_type != "application/x-www-form-urlencoded" {
		err := error_resp{
			ErrorCode: 1004,
			ErrorMsg:  "File type not supported",
		}
		err_str := fmt.Sprintf("%#v", err)
		http.Error(w, err_str, http.StatusBadGateway)
	}
}

// --- Main Handler:- Admin ---
func RestCallAssignerWrapperAdmin(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(context.Background()).WithName("RestCallAssignerWrapperAdmin")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error reading request body:", err)
		return
	}
	req, err := ParseQuery(string(body))
	if err != nil {
		logger.Error(err, "failed to parse the url body of type x-www-form-urlencoded")
	}
	rest_call := fmt.Sprint(req["rest_call"])
	if err != nil {
		logger.Error(err, "failed to marshal the request to json")
	}
	CheckFormEncodedType(w, r.Header.Get("Content-Type"))
	if !mock.AuthChecker(req, w) {
		return
	}
	switch rest_call {
	case "get_usage_types_m":
		mock.GetUsageTypeHandler(w, req)
	case "get_usage_unit_types_m":
		mock.GetUsageUnitsHandler(w, req)
	case "get_usage_type_details_m":
		mock.GetUsageTypeDetailsHandler(w, req)
	case "delete_plans_m":
		mock.DeletePlanClientHandler(w, req)
	case "edit_plan_m":
		mock.DeactivatePlanHandler(w, req)
	case "create_new_plan_m":
		mock.CreatePlanClientHandler(w, req)
	case "create_usage_type_m":
		mock.CreateUsageTypehandler(w, req)
	case "update_usage_type_m":
		mock.UpdateUsageTypeHandler(w, req)
	case "get_service_details_m":
		mock.GetAriaServiceDetailsHandler(w, req)
	case "create_service_m":
		mock.CreateAriaServiceHandler(w, req)
	case "update_service_m":
		mock.UpdateAriaServicehandler(w, req)
	case "get_plan_details_m":
		mock.GetPlanDetailsHandler(w, req)
	}

}

// ---- Main Handler for Client ----
func RestCallAssignerWrapper(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(context.Background()).WithName("RestCallAssignerWrapper")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error reading request body:", err)
		return
	}
	var req map[string]interface{}
	err = json.Unmarshal(body, &req)
	// logger.Info(string(body))
	if err != nil {
		logger.Error(err, "failed to Unmarshal request")

	}
	rest_call := fmt.Sprintf("%v", req["rest_call"])
	CheckJsonType(w, r.Header.Get("Content-Type"))
	if !mock.AuthChecker(req, w) {
		return
	}
	switch rest_call {
	case "get_client_plans_all_m":
		mock.GetPlansHandler(w, body)
	case "create_acct_complete_m":
		mock.CreateAriaAccountHandler(w, body)
	case "get_acct_details_all_m":
		mock.GetAriaAccountHandler(w, body)
	case "get_acct_payment_methods_and_terms_m":
		mock.GetPaymentMethodHandler(w, body)
	case "assign_collections_acct_group_m":
		mock.AssignCollectionsAccountGroupHandler(w, body)
	case "update_acct_billing_group_m":
		mock.AddAccountPaymentMethodHandler(w, body)
	case "create_advanced_service_credit_m":
		mock.CreateServiceCreditsHandler(w, body)
	case "get_unapplied_service_credits_m":
		mock.GetUnappliedServiceCreditsHandler(w, body)
	case "set_session_m":
		mock.SetSessionHandler(w, body)
	case "create_acct_billing_group_m":
		mock.CreateBillingGroupHandler(w, body)
	case "create_acct_dunning_group_m":
		mock.CreateDunningGroupHandler(w, body)
	case "get_acct_billing_group_details_m":
		mock.GetBillingGroupHandler(w, body)
	case "get_acct_dunning_group_details_m":
		mock.GetDunningGroupHandler(w, body)
	case "get_acct_plans_all_m":
		mock.GetAcctPlansHandler(w, body)
	case "get_acct_no_from_user_id_m":
		mock.GetAccountNoFromUserIdHandler(w, body)
	case "assign_acct_plan_m":
		mock.AssignPlanToAccountHandler(w, body)
	case "get_acct_credits_m":
		mock.GetAccountCreditHandler(w, body)
	case "get_credit_details_m":
		mock.GetAccountCreditDetailsHandler(w, body)
	case "set_acct_notify_tmplt_grp_m":
		mock.SetAccountNotifyTemplateGroupHandler(w, body)
	case "get_client_plan_service_rates_m":
		mock.GetClientPlanServiceRatesHandler(w, body)
	case "bulk_record_usage_m":
		mock.CreateBulkUsageRecord(w, body)
	}
}
