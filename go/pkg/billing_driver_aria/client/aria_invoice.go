// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

type AriaInvoiceClient struct {
	ariaClient      *AriaClient
	ariaCredentials *AriaCredentials
}

func NewAriaInvoiceClient(ariaClient *AriaClient, ariaCredentials *AriaCredentials) *AriaInvoiceClient {
	return &AriaInvoiceClient{
		ariaClient:      ariaClient,
		ariaCredentials: ariaCredentials,
	}
}

func (ariaInvoiceClient *AriaInvoiceClient) GetInvoiceDetails(ctx context.Context, clientAcctId string, invoiceNo int64, clientMasterPlanInstanceId string) (*response.GetInvoiceDetailsMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaInvoiceClient.GetInvoiceDetails").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getInvoiceDetailsRequest := request.GetInvoiceDetails{
		AriaRequest:                request.AriaRequest{RestCall: "get_invoice_details_m"},
		OutputFormat:               "json",
		ClientNo:                   ariaInvoiceClient.ariaCredentials.clientNo,
		AuthKey:                    ariaInvoiceClient.ariaCredentials.authKey,
		AltCallerId:                AriaClientId,
		ClientAcctId:               clientAcctId,
		InvoiceNo:                  invoiceNo,
		ClientMasterPlanInstanceId: clientMasterPlanInstanceId,
	}
	return CallAria[response.GetInvoiceDetailsMResponse](ctx, ariaInvoiceClient.ariaClient, &getInvoiceDetailsRequest, FailedToGetInvoiceDetailsError)
}

func (ariaInvoiceClient *AriaInvoiceClient) GetStatementForInvoice(ctx context.Context, clientAcctId string, invoiceNo int64) (*response.GetStatementForInvoiceMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaInvoiceClient.GetStatementForInvoice").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getStatementForInvoiceRequest := request.GetStatementForInvoiceMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_statement_for_invoice_m",
		},
		OutputFormat: "json",
		ClientNo:     ariaInvoiceClient.ariaCredentials.clientNo,
		AuthKey:      ariaInvoiceClient.ariaCredentials.authKey,
		AltCallerId:  AriaClientId,
		ClientAcctId: clientAcctId,
		InvoiceNo:    invoiceNo,
	}
	return CallAria[response.GetStatementForInvoiceMResponse](ctx, ariaInvoiceClient.ariaClient, &getStatementForInvoiceRequest, FailedToGetStatementForInvoice)
}

const (
	FILTER_MASTER_PLAN_ALL        = -1
	FILTER_CLIENT_MASTER_PLAN_ALL = ""
	NO_START_DATE                 = ""
	NO_END_DATE                   = ""
)

func (ariaInvoiceClient *AriaInvoiceClient) GetInvoiceHistory(ctx context.Context, clientAccountId string, clientMasterPlanInstanceId string,
	startBillDate string, endBillDate string) (*response.GetInvoiceHistory, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaInvoiceClient.GetInvoiceHistory").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	invoiceHistory := request.GetInvoiceHistory{
		AriaRequest: request.AriaRequest{
			RestCall: "get_invoice_history_m"},
		OutputFormat:               "json",
		ClientNo:                   ariaInvoiceClient.ariaCredentials.clientNo,
		AuthKey:                    ariaInvoiceClient.ariaCredentials.authKey,
		ClientAcctId:               clientAccountId,
		StartBillDate:              startBillDate,
		EndBillDate:                endBillDate,
		ClientMasterPlanInstanceId: clientMasterPlanInstanceId,
	}
	if clientMasterPlanInstanceId == FILTER_CLIENT_MASTER_PLAN_ALL {
		invoiceHistory.MasterPlanInstanceId = FILTER_MASTER_PLAN_ALL
	}
	return CallAria[response.GetInvoiceHistory](ctx, ariaInvoiceClient.ariaClient, &invoiceHistory, FailedToGetInvoiceHistory)
}

// For Testing
func (ariaInvoiceClient *AriaInvoiceClient) GetPendingInvoiceNo(ctx context.Context, clientAccountId string) (*response.GetPendingInvoice, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaInvoiceClient.GetPendingInvoiceNo").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	pendingInvoiceNo := request.Invoice{
		AriaRequest: request.AriaRequest{
			RestCall: "get_pending_invoice_no_m"},
		OutputFormat: "json",
		ClientNo:     ariaInvoiceClient.ariaCredentials.clientNo,
		AuthKey:      ariaInvoiceClient.ariaCredentials.authKey,
		ClientAcctId: clientAccountId,
	}
	return CallAria[response.GetPendingInvoice](ctx, ariaInvoiceClient.ariaClient, &pendingInvoiceNo, FailedToGetPendingInvoiceNo)
}

func (ariaInvoiceClient *AriaInvoiceClient) ManagePendingInvoice(ctx context.Context, clientAccountId string, clientMasterPlanInstanceId string, actionDirective int64) (*response.ManagePendingInvoice, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaInvoiceClient.ManagePendingInvoice").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	pendingInvovice := ariaInvoiceClient.getPendingInvoviceRequest(clientAccountId, actionDirective)
	pendingInvovice.ClientMasterPlanInstanceId = clientMasterPlanInstanceId
	return CallAria[response.ManagePendingInvoice](ctx, ariaInvoiceClient.ariaClient, &pendingInvovice, FailedToManagePendingInvoice)
}

func (ariaInvoiceClient *AriaInvoiceClient) ManagePendingInvoiceWithInoviceNo(ctx context.Context, clientAccountId string, invoiceNo int64, actionDirective int64) (*response.ManagePendingInvoice, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaInvoiceClient.ManagePendingInvoiceWithInoviceNo").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	pendingInvovice := ariaInvoiceClient.getPendingInvoviceRequest(clientAccountId, actionDirective)
	pendingInvovice.InvoiceNo = invoiceNo
	return CallAria[response.ManagePendingInvoice](ctx, ariaInvoiceClient.ariaClient, &pendingInvovice, FailedToManagePendingInvoice)
}

func (ariaInvoiceClient *AriaInvoiceClient) getPendingInvoviceRequest(clientAccountId string, actionDirective int64) request.ManageInvoice {
	pendingInvovice := request.ManageInvoice{
		AriaRequest: request.AriaRequest{
			RestCall: "manage_pending_invoice_m"},
		OutputFormat:    "json",
		ClientNo:        ariaInvoiceClient.ariaCredentials.clientNo,
		AuthKey:         ariaInvoiceClient.ariaCredentials.authKey,
		ClientAcctId:    clientAccountId,
		ActionDirective: actionDirective,
	}
	return pendingInvovice
}
