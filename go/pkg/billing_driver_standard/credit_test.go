// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

// func TestCredit(t *testing.T) {
// 	ctx := context.Background()

// 	logger := log.FromContext(ctx).WithName("TestCredit")
// 	logger.Info("BEGIN")
// 	defer logger.Info("End")

// 	standardDriverCreditClient := pb.NewBillingCreditServiceClient(standardDriverConn)

// 	cloudAcct := CreateStanardCloudAcct(t, ctx)

// 	couponCode := "SomeCode"

// 	currentDate := time.Now()
// 	newDate := currentDate.AddDate(0, 0, 40)

// 	billingCredit := &pb.BillingCredit{
// 		Created:         timestamppb.New(currentDate),
// 		Expiration:      timestamppb.New(newDate),
// 		CloudAccountId:  cloudAcct.Id,
// 		Reason:          pb.BillingCreditReason_CREDIT_INITIAL,
// 		OriginalAmount:  100,
// 		RemainingAmount: 100,
// 		CouponCode:      couponCode,
// 	}

// 	_, err := standardDriverCreditClient.Create(ctx, billingCredit)
// 	if err != nil {
// 		t.Fatalf("failed to create cloud credit: %v", err)
// 	}

// 	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}

// 	standardCreditReadClient, err := standardDriverCreditClient.ReadInternal(ctx, billingAcct)
// 	if err != nil {
// 		t.Fatalf("failed to get client for reading billing credit: %v", err)
// 	}

// 	for {
// 		billingCreditR, err := standardCreditReadClient.Recv()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			t.Fatalf("failed to read billing credit: %v", err)
// 		}
// 		if billingCreditR.OriginalAmount != billingCredit.OriginalAmount {
// 			t.Fatalf("original amount does not match")
// 		}
// 	}

// }
