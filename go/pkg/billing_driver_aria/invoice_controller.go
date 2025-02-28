// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	FILTER_MASTER_PLAN_INSTANCE = -1
	INVOICE_MAX_DEFAULT_HISTORY = 12
)

type InvoiceController struct {
	ariaInvoiceClient  *client.AriaInvoiceClient
	cloudAccountClient pb.CloudAccountServiceClient
}

type TimeRange struct {
	StartTime *timestamppb.Timestamp
	EndTime   *timestamppb.Timestamp
}

func NewInvoiceController(ariaClient *client.AriaClient, ariaCredentials *client.AriaCredentials, cloudAccountClient pb.CloudAccountServiceClient) *InvoiceController {
	ariaInvoiceClient := client.NewAriaInvoiceClient(ariaClient, ariaCredentials)
	return &InvoiceController{ariaInvoiceClient: ariaInvoiceClient, cloudAccountClient: cloudAccountClient}
}

func (invoiceController InvoiceController) GetStatement(ctx context.Context, cloudAccountId string, invoiceId int64) (*pb.Statement, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvoiceController.GetStatement").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	clientAccountId := client.GetAccountClientId(cloudAccountId)
	statementResp, err := invoiceController.ariaInvoiceClient.GetStatementForInvoice(ctx, clientAccountId, invoiceId)
	if err != nil {
		logger.Error(err, client.GetDriverError(FailedToGetStatementForInvoice, err).Error())
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToGetStatementForInvoice, err).Error())
	}
	statementBytes := []byte(statementResp.OutStatement)
	statement := &pb.Statement{
		Statement: statementBytes,
		MimeType:  statementResp.MimeType,
	}
	return statement, nil
}

func (invoiceController InvoiceController) GetInvoice(ctx context.Context, cloudAccountId string, startTime *timestamppb.Timestamp, endTime *timestamppb.Timestamp) (*pb.BillingInvoiceResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvoiceController.GetInvoice").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	clientAcctId := client.GetAccountClientId(cloudAccountId)

	// get cloud account information - creation date
	cloudAccount, err := invoiceController.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
	if err != nil {
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToGetAccountCreationTime, err).Error())
	}

	if startTime == nil && endTime == nil {
		// if both start and end are nil, set appropriate date range for default case.
		dateRange, err := GetDefaultInvoiceDates(ctx, startTime, endTime, cloudAccount.Created)
		if err != nil {
			return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToSetDefaultInvoicePeriod, err).Error())
		}
		startTime = dateRange.StartTime
		endTime = dateRange.EndTime
	}

	// validate invoice date range
	if !validateInvoiceDateRange(ctx, startTime, endTime) {
		return nil, status.Errorf(codes.Internal, client.GetDriverError(InvalidDateRangeForInvoice, err).Error())
	}

	// get invoices from aria
	invoicesResponse, err := invoiceController.ariaInvoiceClient.GetInvoiceHistory(ctx, clientAcctId, client.FILTER_CLIENT_MASTER_PLAN_ALL, client.TimestampToAriaDateFormat(startTime), client.TimestampToAriaDateFormat(endTime))
	if err != nil || invoicesResponse.ErrorCode != 0 {
		logger.Error(err, client.GetDriverError(FailedToGetInvoiceHistory, err).Error())
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToGetInvoiceHistory, err).Error())
	}
	invoices, err := mapAndGetInvoices(ctx, invoicesResponse.InvoiceHist, cloudAccountId)
	if err != nil {
		logger.Error(err, client.GetDriverError(FailedToMapDataForInvoice, err).Error())
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToMapDataForInvoice, err).Error())
	}
	invoiceResponse := &pb.BillingInvoiceResponse{
		LastUpdated: timestamppb.Now(),
		Invoices:    invoices,
	}
	return invoiceResponse, nil
}

func GetDefaultInvoiceDates(ctx context.Context, startTime *timestamppb.Timestamp, endTime *timestamppb.Timestamp, cloudAccountCreationTime *timestamppb.Timestamp) (*TimeRange, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvoiceController.GetDefaultInvoiceDates").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	invoiceDates := &TimeRange{}
	currentTime := time.Now()

	if endTime == nil {
		currentTime := time.Now()
		firstOfMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, time.Local)
		lastOfMonth := firstOfMonth.AddDate(0, 1, 0).Add(time.Nanosecond * -1)

		endTime = &timestamppb.Timestamp{
			Seconds: lastOfMonth.Unix(),
			Nanos:   int32(lastOfMonth.Nanosecond()),
		}

	}

	if startTime == nil {
		startTimeTmp := currentTime.AddDate(0, -INVOICE_MAX_DEFAULT_HISTORY, 0)

		// compare startTime and cloudAccountCreationTime and set startTime with whichever is after
		if cloudAccountCreationTime.AsTime().After(startTimeTmp) {
			logger.Info("Invoices", "cloudAccountCreationTime", cloudAccountCreationTime.AsTime())
			logger.Info("Invoices", "cloudAccountCreationTimePB", cloudAccountCreationTime)
			startTimeTmp = cloudAccountCreationTime.AsTime()
		}

		startTimeTmp = time.Date(startTimeTmp.Year(), startTimeTmp.Month(), 1, 0, 0, 0, 0, currentTime.Location())
		startTime = timestamppb.New(startTimeTmp)
	}

	invoiceDates.StartTime = startTime
	invoiceDates.EndTime = endTime

	return invoiceDates, nil
}

func validateInvoiceDateRange(ctx context.Context, startTime *timestamppb.Timestamp, endTime *timestamppb.Timestamp) bool {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvoiceController.validateInvoiceDateRange").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	// validate input start and end time
	if startTime != nil && endTime != nil && endTime.AsTime().Before(startTime.AsTime()) {
		logger.Info("endTime is before the starttime", "startTime", startTime.AsTime().Format("January 2, 2006"), "endTime", endTime.AsTime().Format("January 2, 2006"))
		return false
	}

	// Invoices do not display the information for the current month or future months.
	if isInFutureMonth(startTime) || isInFutureMonth(endTime) {
		logger.Info("start or end time cannot be within the current month or future month.", "startTime", startTime.AsTime().Format("January 2, 2006"), "endTime", endTime.AsTime().Format("January 2, 2006"))
		return false
	}

	// if either start or end is nil
	if (startTime == nil && endTime != nil) || (startTime != nil && endTime == nil) {
		logger.Info("invalid start time")
		return false
	}

	return true

}

func isInFutureMonth(inputTime *timestamppb.Timestamp) bool {
	if inputTime != nil {
		currentTime := time.Now()
		inputDate := time.Unix(inputTime.Seconds, int64(inputTime.Nanos))
		if inputDate.Year() > currentTime.Year() || (inputDate.Year() == currentTime.Year() && inputDate.Month() > currentTime.Month()) {
			return true
		}
	}
	return false
}

func mapAndGetInvoices(ctx context.Context, invoiceHist []data.InvoiceHist, cloudAccountId string) ([]*pb.Invoice, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("InvoiceController.mapAndGetInvoices").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	invoices := []*pb.Invoice{}
	for _, invoice := range invoiceHist {
		startDate, err := client.ParseAriaDate(invoice.UsageBillFrom)
		if err != nil {
			logger.Error(err, "error in usage bill from date")
			return nil, fmt.Errorf("error in usage bill from date %v", err)
		}
		endDate, err := client.ParseAriaDate(invoice.UsageBillThru)
		if err != nil {
			logger.Error(err, "error in usage bill-through date of the invoice")
			return nil, fmt.Errorf("error in usage bill-through date of the invoice %v", err)
		}

		billDate, err := client.ParseAriaDate(invoice.BillDate)
		if err != nil {
			logger.Error(err, "error in date the invoice was billed")
			return nil, fmt.Errorf("error in date the invoice was billed %v", err)
		}
		statusCd := "Due"
		var paidDate *timestamppb.Timestamp
		if len(invoice.PaidDate) > 0 {
			paidDate, err = client.ParseAriaDate(invoice.PaidDate)
			statusCd = "Paid"
			if err != nil {
				logger.Error(err, "error in date the invoice was paid")
				return nil, fmt.Errorf("error date the invoice was paid %v", err)
			}
		}
		dueAmount := invoice.Amount - invoice.Credit
		if dueAmount == 0 {
			statusCd = "Paid"
		}
		invoice := &pb.Invoice{
			CloudAccountId: cloudAccountId,
			Id:             uint64(invoice.InvoiceNo),
			Total:          float64(invoice.Amount),
			Paid:           float64(invoice.Credit),
			Due:            float64(dueAmount),
			Start:          startDate,
			End:            endDate,
			InvoiceDate:    billDate,
			PaidDate:       paidDate,
			BillingPeriod:  client.TimeToMonthYearFormat(billDate.AsTime()),
			Status:         statusCd,
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}
