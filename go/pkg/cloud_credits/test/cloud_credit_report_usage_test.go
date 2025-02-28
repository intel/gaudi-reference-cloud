// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"
	// "time"

	"github.com/google/uuid"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	// intelDriverTesting "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	standardDriverTesting "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_standard"
	cc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	meteringQuery "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/db/query"
	meteringServer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/usage"
	// "google.golang.org/protobuf/types/known/timestamppb"
)

func initializeTests(t *testing.T, ctx context.Context) {
	err := meteringQuery.DeleteAllRecords(ctx, meteringServer.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	usageData := usage.NewUsageData(usage.GetUsageDb())
	err = usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all records")
	}
}

func calculateUsages(t *testing.T, ctx context.Context, cloudAcctId string, meteringQuantity float32) {
	resource1Id := uuid.NewString()
	transaction1Id := uuid.NewString()

	meteringServiceClient := pb.NewMeteringServiceClient(meteringClientConn)
	_, err := meteringServiceClient.Create(context.Background(),
		usage.GetComputeUsageRecordCreate(cloudAcctId, resource1Id, transaction1Id,
			usage.DefaultServiceRegion, usage.GetIdcComputeServiceName(), usage.GetXeon3SmallInstanceType(), meteringQuantity))

	if err != nil {
		t.Fatalf("failed to create metering record: %v", err)
	}

	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	usageData := usage.NewUsageData(usage.GetUsageDb())
	usageRecordData := usage.NewUsageRecordData(usage.GetUsageDb())

	uC := usage.NewUsageController(*billingCommon.NewCloudAccountClientForTest(cloudAcctClient), productClient,
		billingCommon.NewMeteringClientForTest(meteringServiceClient), usageData, usageRecordData)
	uC.CalculateUsages(ctx)
}

func TestStandardCloudCreditReport(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestStandardCloudCreditReport")
	logger.Info("BEGIN")
	defer logger.Info("End")

	initializeTests(t, ctx)
	standardUser := "standard_test" + uuid.NewString() + "@example.com"

	cloudAcct := createAndGetCloudAcctWithCredit(t, ctx, standardUser, pb.AccountType_ACCOUNT_TYPE_STANDARD)

	calculateUsages(t, ctx, cloudAcct.Id, 120)

	cloudCreditUsageReport := cc.NewCreditUsageReportScheduler(standardDriverTesting.GetDriverDb(), TestCloudAccountSvcClient,
		TestUsageSvcClient)
	cloudCreditUsageReport.ReportCloudCreditResourceUsages(ctx)
	cloudCreditUsageReport.ReportCloudCreditProductUsages(ctx)

	billingCreditsUpdated := GetBillingCredits(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})

	if len(billingCreditsUpdated) != 1 {
		t.Fatalf("should have got one billing credit after updated")
	}

	verifyUsageReported(t, ctx, billingCreditsUpdated[0], cloudAcct.Id)
}

// func TestIntelCloudCreditReport(t *testing.T) {
// 	ctx := context.Background()
// 	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditReport")
// 	logger.Info("BEGIN")
// 	defer logger.Info("End")

// 	initializeTests(t, ctx)
// 	intelUser := "intel_test" + uuid.NewString() + "@example.com"

// 	cloudAcct := createAndGetCloudAcctWithCredit(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL)

// 	calculateUsages(t, ctx, cloudAcct.Id, 120)
// 	cloudCreditUsageReport := cc.NewCreditUsageReportScheduler(standardDriverTesting.GetDriverDb(), pb.NewCloudAccountServiceClient(cloudAccountConn),
// 		pb.NewUsageServiceClient(usageConn))
// 	cloudCreditUsageReport.ReportCloudCreditResourceUsages(ctx)
// 	cloudCreditUsageReport.ReportCloudCreditProductUsages(ctx)

// 	billingCreditsUpdated := GetBillingCredits(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})

// 	verifyUsageReported(t, ctx, billingCreditsUpdated[0], cloudAcct.Id)
// }

// func TestIntelCloudCreditReportMultipleUsageRecords(t *testing.T) {
// 	ctx := context.Background()
// 	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditReportMultipleUsageRecords")
// 	logger.Info("BEGIN")
// 	defer logger.Info("End")

// 	initializeTests(t, ctx)

// 	intelUser := "intel_test" + uuid.NewString() + "@example.com"

// 	cloudAcct := createAndGetCloudAcctWithCredit(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL)

// 	calculateUsages(t, ctx, cloudAcct.Id, 120)
// 	calculateUsages(t, ctx, cloudAcct.Id, 120)

// 	cloudCreditUsageReport := cc.NewCreditUsageReportScheduler(standardDriverTesting.GetDriverDb(), pb.NewCloudAccountServiceClient(cloudAccountConn),
// 		pb.NewUsageServiceClient(usageConn))
// 	cloudCreditUsageReport.ReportCloudCreditResourceUsages(ctx)
// 	cloudCreditUsageReport.ReportCloudCreditProductUsages(ctx)

// 	billingCreditsUpdated := GetBillingCredits(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})

// 	if len(billingCreditsUpdated) != 1 {
// 		t.Fatalf("should have got one billing credit after updated")
// 	}

// 	verifyUsageReported(t, ctx, billingCreditsUpdated[0], cloudAcct.Id)
// }

// func TestIntelCloudCreditReportMultipleCredits(t *testing.T) {
// 	ctx := context.Background()
// 	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditReportMultipleCredits")
// 	logger.Info("BEGIN")
// 	defer logger.Info("End")

// 	initializeTests(t, ctx)

// 	intelUser := "intel_test" + uuid.NewString() + "@example.com"

// 	cloudAcct := createAndGetCloudAcctWithCredit(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL)

// 	billingCreditClient := pb.NewBillingCreditServiceClient(clientConn)
// 	currTime := time.Now()
// 	newDate := currTime.AddDate(0, 0, DefaultCreditExpirationDays+50)
// 	expirationDate := timestamppb.New(newDate)
// 	billingCredit := &pb.BillingCredit{
// 		CloudAccountId:  cloudAcct.Id,
// 		Created:         timestamppb.New(time.Now()),
// 		OriginalAmount:  DefaultCloudCreditAmount,
// 		RemainingAmount: DefaultCloudCreditAmount,
// 		Reason:          DefaultCreditReason,
// 		CouponCode:      "LaterCredit",
// 		Expiration:      expirationDate}

// 	_, err := billingCreditClient.Create(context.Background(), billingCredit)

// 	if err != nil {
// 		t.Fatalf("failed to create driver credits: %v", err)
// 	}

// 	calculateUsages(t, ctx, cloudAcct.Id, 120)
// 	calculateUsages(t, ctx, cloudAcct.Id, 120)

// 	cloudCreditUsageReport := cc.NewCreditUsageReportScheduler(standardDriverTesting.GetDriverDb(), pb.NewCloudAccountServiceClient(cloudAccountConn),
// 		pb.NewUsageServiceClient(usageConn))
// 	cloudCreditUsageReport.ReportCloudCreditResourceUsages(ctx)
// 	cloudCreditUsageReport.ReportCloudCreditProductUsages(ctx)

// 	billingCreditsUpdated := GetBillingCredits(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})

// 	if len(billingCreditsUpdated) != 2 {
// 		t.Fatalf("should have got one billing credit after updated")
// 	}

// 	if billingCreditsUpdated[0].CouponCode == "LaterCredit" {
// 		verifyUsageReported(t, ctx, billingCreditsUpdated[1], cloudAcct.Id)
// 		if billingCreditsUpdated[0].AmountUsed != 0 {
// 			t.Fatalf("should not have amount used updated")
// 		}
// 	} else {
// 		verifyUsageReported(t, ctx, billingCreditsUpdated[0], cloudAcct.Id)
// 		if billingCreditsUpdated[1].AmountUsed != 0 {
// 			t.Fatalf("should not have amount used updated")
// 		}
// 	}
// }

// func TestIntelCloudCreditReportHigherUsage(t *testing.T) {
// 	ctx := context.Background()
// 	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditReportHigherUsage")
// 	logger.Info("BEGIN")
// 	defer logger.Info("End")

// 	initializeTests(t, ctx)

// 	intelUser := "intel_test" + uuid.NewString() + "@example.com"

// 	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
// 		Name:  intelUser,
// 		Owner: intelUser,
// 		Tid:   uuid.NewString(),
// 		Oid:   uuid.NewString(),
// 		Type:  pb.AccountType_ACCOUNT_TYPE_INTEL,
// 	})

// 	calculateUsages(t, ctx, cloudAcct.Id, 120)

// 	usageData := usage.NewUsageData(usage.GetUsageDb())
// 	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &cloudAcct.Id})

// 	if err != nil {
// 		t.Fatalf("failed to search resource usages: %v", err)
// 	}

// 	var amountToBeReported float64 = 0

// 	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
// 		amountToBeReported += resourceUsageRecord.Rate * resourceUsageRecord.Quantity
// 	}

// 	billingCreditClient := pb.NewBillingCreditServiceClient(clientConn)
// 	currTime := time.Now()
// 	newDate := currTime.AddDate(0, 0, DefaultCreditExpirationDays+50)
// 	expirationDate := timestamppb.New(newDate)
// 	creditAmount := amountToBeReported - 1
// 	billingCredit := &pb.BillingCredit{
// 		CloudAccountId:  cloudAcct.Id,
// 		Created:         timestamppb.New(time.Now()),
// 		OriginalAmount:  creditAmount,
// 		RemainingAmount: creditAmount,
// 		Reason:          DefaultCreditReason,
// 		CouponCode:      DefaultCreditCoupon,
// 		Expiration:      expirationDate}

// 	_, err = billingCreditClient.Create(context.Background(), billingCredit)

// 	if err != nil {
// 		t.Fatalf("failed to create driver credits: %v", err)
// 	}

// 	cloudCreditUsageReport := cc.NewCreditUsageReportScheduler(standardDriverTesting.GetDriverDb(), pb.NewCloudAccountServiceClient(cloudAccountConn),
// 		pb.NewUsageServiceClient(usageConn))
// 	cloudCreditUsageReport.ReportCloudCreditResourceUsages(ctx)
// 	cloudCreditUsageReport.ReportCloudCreditProductUsages(ctx)

// 	billingCreditsUpdated := GetBillingCredits(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})

// 	if len(billingCreditsUpdated) != 1 {
// 		t.Fatalf("should have got one billing credit after updated")
// 	}

// 	if billingCreditsUpdated[0].AmountUsed != creditAmount {
// 		t.Fatalf("incorrect amount used")
// 	}

// 	if billingCreditsUpdated[0].RemainingAmount != 0 {
// 		t.Fatalf("incorrect amount used")
// 	}

// 	resourceUsagesAfterUpdate, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &cloudAcct.Id})

// 	if err != nil {
// 		t.Fatalf("failed to search resource usages: %v", err)
// 	}

// 	var allReported bool = true

// 	for _, resourceUsage := range resourceUsagesAfterUpdate.ResourceUsages {
// 		allReported = allReported && resourceUsage.Reported
// 	}

// 	if allReported {
// 		t.Fatalf("at least one should not have been reported")
// 	}
// }

// func TestIntelCloudCreditReportMultiple(t *testing.T) {
// 	ctx := context.Background()
// 	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditReportMultiple")
// 	logger.Info("BEGIN")
// 	defer logger.Info("End")

// 	initializeTests(t, ctx)

// 	intelUser := "intel_test" + uuid.NewString() + "@example.com"

// 	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
// 		Name:  intelUser,
// 		Owner: intelUser,
// 		Tid:   uuid.NewString(),
// 		Oid:   uuid.NewString(),
// 		Type:  pb.AccountType_ACCOUNT_TYPE_INTEL,
// 	})

// 	calculateUsages(t, ctx, cloudAcct.Id, 120)
// 	calculateUsages(t, ctx, cloudAcct.Id, 120)
// 	usageData := usage.NewUsageData(usage.GetUsageDb())

// 	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &cloudAcct.Id})

// 	if err != nil {
// 		t.Fatalf("failed to search resource usages: %v", err)
// 	}

// 	var amountToBeReported float64 = 0

// 	for _, resourceUsage := range resourceUsages.ResourceUsages {
// 		amountToBeReported += resourceUsage.Rate * resourceUsage.Quantity
// 	}

// 	billingCreditClient := pb.NewBillingCreditServiceClient(clientConn)
// 	currTime := time.Now()
// 	newDate := currTime.AddDate(0, 0, DefaultCreditExpirationDays+50)
// 	expirationDate := timestamppb.New(newDate)
// 	billingCredit := &pb.BillingCredit{
// 		CloudAccountId:  cloudAcct.Id,
// 		Created:         timestamppb.New(time.Now()),
// 		OriginalAmount:  amountToBeReported / 2,
// 		RemainingAmount: amountToBeReported / 2,
// 		Reason:          DefaultCreditReason,
// 		CouponCode:      DefaultCreditCoupon,
// 		Expiration:      expirationDate}

// 	_, err = billingCreditClient.Create(context.Background(), billingCredit)

// 	if err != nil {
// 		t.Fatalf("failed to create driver credits: %v", err)
// 	}

// 	_, err = billingCreditClient.Create(context.Background(), billingCredit)

// 	if err != nil {
// 		t.Fatalf("failed to create driver credits: %v", err)
// 	}

// 	cloudCreditUsageReport := cc.NewCreditUsageReportScheduler(standardDriverTesting.GetDriverDb(), pb.NewCloudAccountServiceClient(cloudAccountConn),
// 		pb.NewUsageServiceClient(usageConn))
// 	cloudCreditUsageReport.ReportCloudCreditResourceUsages(ctx)
// 	cloudCreditUsageReport.ReportCloudCreditProductUsages(ctx)

// 	billingCreditsUpdated := GetBillingCredits(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})

// 	if len(billingCreditsUpdated) != 2 {
// 		t.Fatalf("should have got two billing credit after updated")
// 	}

// 	for _, billingCreditUpdated := range billingCreditsUpdated {
// 		if billingCreditUpdated.AmountUsed != (amountToBeReported / 2) {
// 			t.Fatalf("invalid amount to be reported")
// 		}
// 		if billingCreditUpdated.RemainingAmount != 0 {
// 			t.Fatalf("invalid remaining amount")
// 		}
// 	}

// 	resourceUsagesUpdated, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &cloudAcct.Id})

// 	if err != nil {
// 		t.Fatalf("failed to search resource usages: %v", err)
// 	}

// 	for _, resourceUsage := range resourceUsagesUpdated.ResourceUsages {
// 		if resourceUsage.Reported != true {
// 			t.Fatalf("resource usage should be updated as reported")
// 		}
// 	}

// 	cloudCreditUsageReport.ReportCloudCreditResourceUsages(ctx)
// 	cloudCreditUsageReport.ReportCloudCreditProductUsages(ctx)

// 	billingCreditNotUpdated := &pb.BillingCredit{
// 		CloudAccountId:  cloudAcct.Id,
// 		Created:         timestamppb.New(time.Now()),
// 		OriginalAmount:  amountToBeReported / 2,
// 		RemainingAmount: amountToBeReported / 2,
// 		Reason:          DefaultCreditReason,
// 		CouponCode:      "NotUpdated",
// 		Expiration:      expirationDate}

// 	_, _ = billingCreditClient.Create(context.Background(), billingCreditNotUpdated)

// 	billingCreditsAgainUpdated := GetBillingCredits(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})

// 	if len(billingCreditsAgainUpdated) != 3 {
// 		t.Fatalf("should have got two billing credit after updated")
// 	}

// 	for _, billingCreditUpdated := range billingCreditsAgainUpdated {
// 		if billingCreditUpdated.CouponCode == "NotUpdated" {
// 			if billingCreditUpdated.AmountUsed != 0 {
// 				t.Fatalf("invalid amount to be reported")
// 			}
// 		} else {
// 			if billingCreditUpdated.AmountUsed != (amountToBeReported / 2) {
// 				t.Fatalf("invalid amount to be reported")
// 			}
// 			if billingCreditUpdated.RemainingAmount != 0 {
// 				t.Fatalf("invalid remaining amount")
// 			}
// 		}
// 	}
// }

func createAndGetCloudAcctWithCredit(t *testing.T, ctx context.Context, user string, acctType pb.AccountType) *pb.CloudAccount {

	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  acctType,
	})

	CreateBillingCredit(t, ctx, cloudAcct, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
	billingCredits := GetBillingCredits(t, ctx, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})

	if len(billingCredits) != 1 {
		t.Fatalf("should have got one billing credit")
	}

	return cloudAcct
}

func verifyUsageReported(t *testing.T, ctx context.Context, billingCredit *pb.BillingCredit, cloudAcctId string) {

	usageData := usage.NewUsageData(usage.GetUsageDb())
	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &cloudAcctId})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	var amountToBeReported float64 = 0

	for _, resourceUsage := range resourceUsages.ResourceUsages {
		amountToBeReported += resourceUsage.Rate * resourceUsage.Quantity
	}

	if billingCredit.AmountUsed != amountToBeReported {
		t.Fatalf("incorrect amount reported")
	}

	for _, resourceUsage := range resourceUsages.ResourceUsages {
		if resourceUsage.Reported != true {
			t.Fatalf("resource usage should be updated as reported")
		}
	}
}
