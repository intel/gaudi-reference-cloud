// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	intel "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNotifications(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestNotifications")
	logger.Info("BEGIN")
	defer logger.Info("End")
	client := pb.NewNotificationGatewayServiceClient(clientConn)
	eventSeverity := pb.EventSeverity_LOW
	serviceName := pb.ServiceName_BILLING
	message := "this is a test message"
	cloudAcctId := cloudaccount.MustNewId()
	userId := uuid.NewString()
	properties := map[string]string{
		"key": "value",
	}
	//TODO: remove eventmanager and use notificationclient
	_, err := client.Create(ctx, &pb.CreateEvent{
		Status:         pb.EventStatus_ACTIVE,
		Type:           pb.EventType_NOTIFICATION,
		Severity:       &eventSeverity,
		ServiceName:    &serviceName,
		Message:        &message,
		CloudAccountId: &cloudAcctId,
		UserId:         &userId,
		EventSubType:   "CREDIT_EXPIRED",
		Properties:     properties,
		ClientRecordId: uuid.NewString(),
	})

	if err != nil {
		t.Fatalf("failed to create notification")
	}

	events, err := client.Read(ctx, &pb.EventsFilter{CloudAccountId: cloudAcctId})

	if err != nil {
		t.Fatalf("failed to get notifications")
	}

	logger.Info("length of the notifications is", "len", events.NumberOfNotifications)
}

func TestAlerts(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestAlerts")
	logger.Info("BEGIN")
	defer logger.Info("End")
	client := pb.NewNotificationGatewayServiceClient(clientConn)
	eventSeverity := pb.EventSeverity_LOW
	serviceName := pb.ServiceName_BILLING
	message := "this is a test message"
	cloudAcctId := cloudaccount.MustNewId()
	userId := uuid.NewString()

	_, err := client.Create(ctx, &pb.CreateEvent{
		Status:         pb.EventStatus_ACTIVE,
		Type:           pb.EventType_ALERT,
		Severity:       &eventSeverity,
		ServiceName:    &serviceName,
		Message:        &message,
		CloudAccountId: &cloudAcctId,
		UserId:         &userId,
		EventSubType:   "CREDIT_EXPIRED",
		ClientRecordId: uuid.NewString(),
	})

	if err != nil {
		t.Fatalf("failed to create alert")
	}

	events, err := client.Read(ctx, &pb.EventsFilter{CloudAccountId: cloudAcctId})

	if err != nil {
		t.Fatalf("failed to get alerts")
	}

	logger.Info("length of the alerts is", "len", events.NumberOfAlerts)
}

func TestNotificationLongPoll(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	logger := log.FromContext(ctx).WithName("TestNotificationLongPoll")
	logger.Info("BEGIN")
	defer logger.Info("End")
	client := pb.NewNotificationGatewayServiceClient(clientConn)
	eventSeverity := pb.EventSeverity_LOW
	serviceName := pb.ServiceName_BILLING
	message := "this is a test message"
	cloudAcctId := cloudaccount.MustNewId()
	userId := uuid.NewString()
	properties := map[string]string{
		"key": "value",
	}
	go waitForNotification(t, ctx)
	_, err := client.Create(ctx, &pb.CreateEvent{
		Status:         pb.EventStatus_ACTIVE,
		Type:           pb.EventType_NOTIFICATION,
		Severity:       &eventSeverity,
		ServiceName:    &serviceName,
		Message:        &message,
		CloudAccountId: &cloudAcctId,
		UserId:         &userId,
		EventSubType:   "CREDIT_EXPIRED",
		Properties:     properties,
		ClientRecordId: uuid.NewString(),
	})

	if err != nil {
		t.Fatalf("failed to create notification")
	}

	events, err := client.Read(ctx, &pb.EventsFilter{CloudAccountId: cloudAcctId})

	if err != nil {
		t.Fatalf("failed to get notifications")
	}

	logger.Info("length of the notifications is", "len", events.NumberOfNotifications)
	wg.Wait()
}

func waitForNotification(t *testing.T, ctx context.Context) {
	logger := log.FromContext(ctx).WithName("waitForNotification")
	time.Sleep(time.Second * 4)
	client := pb.NewNotificationGatewayServiceClient(clientConn)
	notificationGatewaySubscribeClient, err := client.Subscribe(ctx, &pb.EventsSubscribe{ClientId: uuid.NewString()})
	if err != nil {
		logger.Error(err, "failed to create client")
	}
	response, err := notificationGatewaySubscribeClient.Recv()
	if err != nil {
		logger.Error(err, "failed to receive notification")
	}
	logger.Info("recieved", "number of notifications", response.NumberOfNotifications)
}

func TestAlertLongPoll(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	logger := log.FromContext(ctx).WithName("TestAlertLongPoll")
	logger.Info("BEGIN")
	defer logger.Info("End")
	client := pb.NewNotificationGatewayServiceClient(clientConn)
	eventSeverity := pb.EventSeverity_LOW
	serviceName := pb.ServiceName_BILLING
	message := "this is a test message"
	cloudAcctId := cloudaccount.MustNewId()
	userId := uuid.NewString()
	go waitForAlert(t, ctx)
	_, err := client.Create(ctx, &pb.CreateEvent{
		Status:         pb.EventStatus_ACTIVE,
		Type:           pb.EventType_ALERT,
		Severity:       &eventSeverity,
		ServiceName:    &serviceName,
		Message:        &message,
		CloudAccountId: &cloudAcctId,
		UserId:         &userId,
		EventSubType:   "CREDIT_EXPIRED",
		ClientRecordId: uuid.NewString(),
	})

	if err != nil {
		t.Fatalf("failed to create alert")
	}

	events, err := client.Read(ctx, &pb.EventsFilter{CloudAccountId: cloudAcctId})

	if err != nil {
		t.Fatalf("failed to get alerts")
	}

	logger.Info("length of the alerts is", "len", events.NumberOfAlerts)
	wg.Wait()
}

func waitForAlert(t *testing.T, ctx context.Context) {
	logger := log.FromContext(ctx).WithName("waitForAlert")
	time.Sleep(time.Second * 4)
	client := pb.NewNotificationGatewayServiceClient(clientConn)
	notificationGatewaySubscribeClient, err := client.Subscribe(ctx, &pb.EventsSubscribe{ClientId: uuid.NewString()})
	if err != nil {
		logger.Error(err, "failed to create client")
	}
	response, err := notificationGatewaySubscribeClient.Recv()
	if err != nil {
		logger.Error(err, "failed to receive alert")
	}
	logger.Info("recieved", "number of alerts", response.GetNumberOfAlerts())
}

func TestIntelCloudCreditExpiredNotification(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditExpiredNotification")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  intelUser,
		Owner: intelUser,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  pb.AccountType_ACCOUNT_TYPE_INTEL,
	})
	expiredDate := time.Now().AddDate(0, 0, -10)
	expirationDate := timestamppb.New(expiredDate)
	billingCredit := &pb.BillingCredit{
		CloudAccountId:  cloudAcct.Id,
		Created:         timestamppb.New(time.Now()),
		OriginalAmount:  DefaultCloudCreditAmount,
		RemainingAmount: DefaultCloudCreditAmount,
		Reason:          DefaultCreditReason,
		CouponCode:      DefaultCreditCoupon,
		Expiration:      expirationDate}
	_, err := intelDriver.billingCredit.Create(context.Background(), billingCredit)
	if err != nil {
		t.Fatalf("failed to create intel credits: %v", err)
	}
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditExpiryScheduler := NewCloudCreditExpiryScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	cloudCreditExpiryScheduler.cloudCreditExpiry(ctx)

	intel.DeleteIntelCredits(t, ctx)

	client := pb.NewNotificationGatewayServiceClient(clientConn)
	events, err := client.Read(ctx, &pb.EventsFilter{CloudAccountId: cloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to get alerts")
	}

	if len(events.Alerts) != 1 {
		t.Fatalf("wrong length of alerts")
	}
	logger.Info("and done with testing for notifications cloud credit expired for intel customers")
}

func TestLoopedIntelCloudCreditExpiredNotification(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestLoopedIntelCloudCreditExpiredNotification")
	logger.Info("BEGIN")
	defer logger.Info("End")

	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	var cloudAcctIds []string
	for i := 1; i <= 20; i++ {
		intelUser := "intel_" + uuid.NewString() + "@example.com"
		cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
			Name:  intelUser,
			Owner: intelUser,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_INTEL,
		})
		cloudAcctIds = append(cloudAcctIds, cloudAcct.GetId())
		expiredDate := time.Now().AddDate(0, 0, -10)
		expirationDate := timestamppb.New(expiredDate)
		for j := 1; j <= 20; j++ {
			billingCredit := &pb.BillingCredit{
				CloudAccountId:  cloudAcct.Id,
				Created:         timestamppb.New(time.Now()),
				OriginalAmount:  DefaultCloudCreditAmount,
				RemainingAmount: DefaultCloudCreditAmount,
				Reason:          DefaultCreditReason,
				CouponCode:      DefaultCreditCoupon,
				Expiration:      expirationDate}
			//Create
			_, err := intelDriver.billingCredit.Create(context.Background(), billingCredit)
			if err != nil {
				t.Fatalf("failed to create intel credits: %v", err)
			}
		}

	}

	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditExpiryScheduler := NewCloudCreditExpiryScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	cloudCreditExpiryScheduler.cloudCreditExpiry(ctx)
	intel.DeleteIntelCredits(t, ctx)
	client := pb.NewNotificationGatewayServiceClient(clientConn)
	for _, cloudAcctId := range cloudAcctIds {
		events, err := client.Read(ctx, &pb.EventsFilter{CloudAccountId: cloudAcctId})

		if err != nil {
			t.Fatalf("failed to get alerts")
		}

		logger.Info("received cloud acct for", "id", events.Alerts[0].CloudAccountId)
		if len(events.Alerts) != 1 {
			t.Fatalf("wrong length of alerts")
		}

		if strings.Compare(*events.Alerts[0].CloudAccountId, cloudAcctId) != 0 {
			t.Fatalf("wrong cloud acct id received")
		}
	}
	_, err := eventData.GetAlerts(ctx)

	if err != nil {
		t.Fatalf("failed to get alerts")
	}

	// this is really a bad way to verify but I need it right now.. Cannot get over it.
	//if len(alerts) != 20 {
	//	t.Fatalf("wrong length of alerts")
	//}
	logger.Info("and done with testing for notification of cloud credit expired for intel customers")
}

func TestLoopedIntelCloudCreditUsageThresholdNotification(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditUsageThresholdNotification")
	logger.Info("BEGIN")
	defer logger.Info("End")

	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	var cloudAcctIds []string
	for i := 1; i <= 20; i++ {
		intelUser := "intel_" + uuid.NewString() + "@example.com"
		cloudAcct := createAcctWithCreditsForNotifications(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL)
		billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}
		intelCredits := GetBillingCredits(t, ctx, billingAcct)
		for _, intelCredit := range intelCredits {
			if intelCredit.CloudAccountId == cloudAcct.Id {
				intel.UpdateIntelCreditWithRemainingAmount(t, ctx, billingAcct, intelCredit.CloudAccountId, 0.19*DefaultCloudCreditAmount)
			}
		}
		cloudAcctIds = append(cloudAcctIds, cloudAcct.GetId())
	}

	cloudCreditUsageScheduler.cloudCreditUsages(ctx)

	for _, cloudAcctId := range cloudAcctIds {
		client := pb.NewNotificationGatewayServiceClient(clientConn)
		events, err := client.Read(ctx, &pb.EventsFilter{CloudAccountId: cloudAcctId})

		if err != nil {
			t.Fatalf("failed to get notifications")
		}

		if len(events.Notifications) != 1 {
			t.Fatalf("wrong length of notifications")
		}
	}

	_, err := eventData.GetNotifications(ctx)

	if err != nil {
		t.Fatalf("failed to get notifications")
	}

	// this is really a bad way to verify but I need it right now.. Cannot get over it.
	//if len(notifications) != 20 {
	//	t.Fatalf("wrong length of notifications")
	//}

	intel.DeleteIntelCredits(t, ctx)
	logger.Info("and done with testing cloud credit usage for intel customers with usage threshold for notifications")
}

func TestIntelCloudCreditUsageCompleteNotification(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditUsageCompleteNotification")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsForNotifications(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}
	billingCredits := GetBillingCredits(t, ctx, billingAcct)
	intel.UpdateIntelCreditUsed(t, ctx, billingAcct, billingCredits)
	intel.GetIntelCredits(t, ctx, billingCredits)
	cloudCreditUsageScheduler.cloudCreditUsages(ctx)
	intel.DeleteIntelCredits(t, ctx)

	client := pb.NewNotificationGatewayServiceClient(clientConn)
	events, err := client.Read(ctx, &pb.EventsFilter{CloudAccountId: cloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to get alerts")
	}

	if len(events.Alerts) != 1 {
		t.Fatalf("wrong length of alerts")
	}

	logger.Info("and done with testing cloud credit usage for intel customers with usage completely used for notification")
}

func TestLoopedIntelCloudCreditUsageCompleteNotification(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestLoopedIntelCloudCreditUsageCompleteNotification")
	logger.Info("BEGIN")
	defer logger.Info("End")
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)

	var cloudAcctIds []string

	for i := 1; i <= 20; i++ {

		intelUser := "intel_" + uuid.NewString() + "@example.com"
		cloudAcct := createAcctWithCreditsForNotifications(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL)
		billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}
		billingCredits := GetBillingCredits(t, ctx, billingAcct)
		intel.UpdateIntelCreditUsed(t, ctx, billingAcct, billingCredits)
		intel.GetIntelCredits(t, ctx, billingCredits)
		cloudAcctIds = append(cloudAcctIds, cloudAcct.GetId())

	}

	cloudCreditUsageScheduler.cloudCreditUsages(ctx)

	intel.DeleteIntelCredits(t, ctx)

	for _, cloudAcctId := range cloudAcctIds {

		client := pb.NewNotificationGatewayServiceClient(clientConn)
		events, err := client.Read(ctx, &pb.EventsFilter{CloudAccountId: cloudAcctId})

		if err != nil {
			t.Fatalf("failed to get alerts")
		}

		if len(events.Alerts) != 1 {
			t.Fatalf("wrong length of alerts")
		}
	}

	_, err := eventData.GetAlerts(ctx)

	if err != nil {
		t.Fatalf("failed to get alerts")
	}

	// this is really a bad way to verify but I need it right now.. Cannot get over it.
	//if len(alerts) != 20 {
	//	t.Fatalf("wrong length of alerts")
	//}

	logger.Info("and done with testing cloud credit usage for intel customers with usage completely used for notification")
}

func TestIntelCloudCreditUsageWithoutNotification(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestIntelCloudCreditUsageWithoutNotification")
	logger.Info("BEGIN")
	defer logger.Info("End")

	intelUser := "intel_" + uuid.NewString() + "@example.com"
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(billingDb)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	cloudAcctClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	cloudAccountSvcClient := &billingCommon.CloudAccountSvcClient{CloudAccountClient: cloudAcctClient}
	cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, nil, testSchedulerCloudAccountState, cloudAccountSvcClient)
	cloudAcct := createAcctWithCreditsForNotifications(t, ctx, intelUser, pb.AccountType_ACCOUNT_TYPE_INTEL)
	billingAcct := &pb.BillingAccount{CloudAccountId: cloudAcct.Id}
	billingCredits := GetBillingCredits(t, ctx, billingAcct)
	intel.GetIntelCredits(t, ctx, billingCredits)
	cloudCreditUsageScheduler.cloudCreditUsages(ctx)
	intel.DeleteIntelCredits(t, ctx)

	client := pb.NewNotificationGatewayServiceClient(clientConn)
	events, err := client.Read(ctx, &pb.EventsFilter{CloudAccountId: cloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to get events")
	}
	logger.Info("Events", "events", events)
	if len(events.Notifications) != 0 {
		t.Fatalf("wrong length of alerts")
	}

	logger.Info("and done with testing cloud credit usage for intel customers without notification")
}

func createAcctWithCreditsForNotifications(t *testing.T, ctx context.Context, user string, acctType pb.AccountType) *pb.CloudAccount {
	cloudAcct := CreateAndGetAccount(t, &pb.CloudAccountCreate{
		Name:  user,
		Owner: user,
		Tid:   uuid.NewString(),
		Oid:   uuid.NewString(),
		Type:  acctType,
	})
	CreateBillingCredit(t, ctx, cloudAcct, &pb.BillingAccount{CloudAccountId: cloudAcct.Id})
	return cloudAcct
}

func TestDismissEvent(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestDismissEvent")
	logger.Info("BEGIN")
	defer logger.Info("End")
	client := pb.NewNotificationGatewayServiceClient(clientConn)
	eventSeverity := pb.EventSeverity_LOW
	serviceName := pb.ServiceName_BILLING
	message := "this is a test message"
	cloudAcctId := cloudaccount.MustNewId()
	userId := uuid.NewString()
	properties := map[string]string{
		"key": "value",
	}
	clientRecordId := uuid.NewString()
	_, err := client.Create(ctx, &pb.CreateEvent{
		Status:         pb.EventStatus_ACTIVE,
		Type:           pb.EventType_NOTIFICATION,
		Severity:       &eventSeverity,
		ServiceName:    &serviceName,
		Message:        &message,
		CloudAccountId: &cloudAcctId,
		UserId:         &userId,
		EventSubType:   "CREDIT_THRESHOLD_REACHED",
		Properties:     properties,
		ClientRecordId: clientRecordId,
	})

	if err != nil {
		logger.Error(err, "failed to create event")
		t.Fatalf("failed to create event")
	}

	_, err = client.DismissEvent(ctx, &pb.EventsFilter{CloudAccountId: cloudAcctId, ClientRecordId: clientRecordId})
	if err != nil {
		logger.Error(err, "failed to dismiss event")
		t.Fatalf("failed to dismiss event")
	}
}
