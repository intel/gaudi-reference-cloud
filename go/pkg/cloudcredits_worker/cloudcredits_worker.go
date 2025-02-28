// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcreditsworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	cloudCredits "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudcredits_worker/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	ng "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type cloudAccountServiceClientInterface interface {
	GetById(ctx context.Context, in *pb.CloudAccountId, opts ...grpc.CallOption) (*pb.CloudAccount, error)
	Update(ctx context.Context, in *pb.CloudAccountUpdate, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type CloudCreditsWorker struct {
	BillingCloudAccountClient *billingCommon.CloudAccountSvcClient
	NotificationGatewayClient *billingCommon.NotificationGatewayClient
	CloudAccountClient        cloudAccountServiceClientInterface
	StopChan                  *chan os.Signal
	CloudCreditsSvcClient     *cloudCredits.CloudCreditsServiceClient
	CloudCreditUsageScheduler *cloudCredits.CloudCreditUsageEventScheduler
}

func (w CloudCreditsWorker) StartSQSConsumerProcess(ctx context.Context) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditsWorker.StartSQSConsumerProcess").WithValues("queue", config.Cfg.GetAWSSQSQueueName()).Start()
	logger.Info("starting consumer")
	defer span.End()
	dataChan := make(chan *pb.MessageResponse, config.Cfg.GetChannelCapacity())
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	iSleepTimeSec := config.Cfg.GetInitSleepTimeSeconds()
	sleepTimeSec := iSleepTimeSec
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				close(dataChan)
				logger.Info("terminating writer")
				return
			default:
				if sleepTimeSec > config.Cfg.GetMaxSleepTimeSeconds() {
					sleepTimeSec = iSleepTimeSec
				}
				sleepTimeSec, _ = WriteEventsToChannel(ctx, w.NotificationGatewayClient, dataChan, sleepTimeSec)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case readMsg, ok := <-dataChan:
				if !ok {
					logger.Info("datachannel closed")
					return
				}
				eTime := readMsg.GetAttributes()["Timestamp"]
				logger.V(1).Info("received", "messageId", readMsg.MessageId, "messageBody", readMsg.Body, "EventTime", eTime)

				parsedMsg, err := w.ValidateAndParseMessage(ctx, readMsg)
				if err != nil {
					logger.Error(err, "msg validation failed, skip and continue..")
					continue
				}
				if err = w.OnNotifyCloudCredits(ctx, parsedMsg, readMsg.MessageId, readMsg.ReceiptHandle, timestamppb.New(parsedMsg.EventTime)); err != nil {
					logger.Error(err, "unable to process", "messageId", readMsg.MessageId)
					continue
				}
			case <-stopChan:
				logger.Info("terminating receiver")
				return
			}
		}
	}()
	wg.Wait()
}

func NewCloudCreditsWorker(stopChan *chan os.Signal, notificationClient *billingCommon.NotificationGatewayClient, billingCloudAccountClient *billingCommon.CloudAccountSvcClient, cloudCreditsSvcClient *cloudCredits.CloudCreditsServiceClient,
	cloudCreditUsageScheduler *cloudCredits.CloudCreditUsageEventScheduler) *CloudCreditsWorker {
	return &CloudCreditsWorker{
		StopChan:                  stopChan,
		NotificationGatewayClient: notificationClient,
		BillingCloudAccountClient: billingCloudAccountClient,
		CloudAccountClient:        billingCloudAccountClient.CloudAccountClient,
		CloudCreditsSvcClient:     cloudCreditsSvcClient,
		CloudCreditUsageScheduler: cloudCreditUsageScheduler,
	}
}

func WriteEventsToChannel(ctx context.Context, notificationGatewayClient *billingCommon.NotificationGatewayClient, dataChan chan *pb.MessageResponse, sleepTimeSec int64) (int64, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditsWorker.WriteEventsToChannel").Start()
	defer span.End()

	// receive from queue
	resp, err := notificationGatewayClient.ReceiveEvents(ctx, &pb.ReceiveEventRequest{
		QueueName:           config.Cfg.GetAWSSQSQueueName(),
		MaxNumberOfMessages: int64(config.Cfg.GetAWSSQSMaxNumberOfMessages()),
		WaitTimeSeconds:     config.Cfg.GetWaitTimeSeconds(),
		EventName:           "All",
	})
	if err != nil {
		logger.Error(err, "unable to receive messages, retrying..", "sleepTimeSeconds", sleepTimeSec)
		time.Sleep(time.Duration(sleepTimeSec) * time.Second)
		sleepTimeSec = sleepTimeSec * 2
		return sleepTimeSec, err
	}
	if resp == nil || len(resp.MessageResponse) == 0 {
		// wait for next producer cycle
		logger.Info("writer goroutine paused..", "sleepTimeSeconds", sleepTimeSec)
		time.Sleep(time.Duration(sleepTimeSec) * time.Second)
		sleepTimeSec = sleepTimeSec * 2
		return sleepTimeSec, nil
	}
	for _, messageResponse := range resp.MessageResponse {
		logger.Info("writing to data channel", "MessageId", messageResponse.MessageId)
		dataChan <- messageResponse
	}
	return config.Cfg.GetInitSleepTimeSeconds(), nil
}

func (w CloudCreditsWorker) ValidateAndParseMessage(ctx context.Context, msg *pb.MessageResponse) (*events.CreateEvent, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditsWorker.ValidateAndParseMessage").WithValues("messageId", msg.MessageId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("messageid and attributes", "MessageId", msg.MessageId, "attributes", msg.Attributes)
	if msg.MessageId == "" || msg.Attributes == nil {
		missingAttrErr := errors.New("missing mgsId or attributes")
		logger.Error(missingAttrErr, msg.Body)
		return nil, missingAttrErr
	}

	parsedMsg := &events.CreateEvent{}
	if err := json.Unmarshal([]byte(msg.Body), parsedMsg); err != nil {
		logger.Error(err, "failed to unmarshal message body")
		return nil, err
	}
	if parsedMsg.CloudAccountId == "" {
		missingCAErr := errors.New("missing cloudAccountId")
		logger.Error(missingCAErr, msg.Body)
		return nil, missingCAErr
	}

	// validate Cloudaccount
	cloudAccount, err := w.CloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: parsedMsg.CloudAccountId})
	if err != nil {
		logger.Error(err, "error getting cloudaccount details from cloudaccount service")
		return nil, err
	}
	span.SetAttributes(attribute.String("cloudAccountId", cloudAccount.Id))

	eventTime, err := parseTime(msg.GetAttributes()["Timestamp"])
	if err != nil {
		logger.Error(err, "error in parseTime")
		return nil, err
	}
	parsedMsg.EventTime = *eventTime

	handleOutdated := true
	csresp, err := CloudCreditsClient.CloudCreditsSvcClient.ReadCreditStateLog(ctx, &pb.CreditsStateFilter{CloudAccountId: cloudAccount.Id})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.NotFound {
			logger.Error(err, "unable to ReadCreditStateLog")
			return nil, err
		} else if ok && st.Code() == codes.NotFound {
			logger.Info("creditStateLog not found", "cloudAccountId", cloudAccount.Id)
			handleOutdated = false
		}
	}

	if eventTime.Before(csresp.GetEventAt().AsTime()) && handleOutdated {
		logger.Info("skip handling outdated event", "EventSubType", parsedMsg.EventSubType, "EventTime", eventTime.String(), "lastEventTime", csresp.GetEventAt().AsTime().String(), "cloudAccountId", cloudAccount.Id)
		err = fmt.Errorf("got outdated event, loggging and deleting")

		// Logging the outdated event
		if err := w.CreateStateLog(ctx, parsedMsg.CloudAccountId, billingCommon.CloudCreditOutDatedEvent, timestamppb.New(*eventTime)); err != nil {
			logger.Error(err, "unable to createStateLogEntry")
			return nil, err
		}

		// delete event from queue
		if err := w.deleteMessageFromQueue(ctx, msg.MessageId, msg.ReceiptHandle); err != nil {
			logger.Error(err, "Unable to delete message from queue")
		}

		return nil, err
	}
	return parsedMsg, nil
}

func (w CloudCreditsWorker) OnNotifyCloudCredits(ctx context.Context, msg *events.CreateEvent, msgId string, rhandle string, eventTime *timestamppb.Timestamp) error {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudCreditsWorker.OnNotifyCloudCredit").WithValues("messageId", msgId, "cloudAccountId", msg.CloudAccountId).Start()
	defer span.End()

	if strings.EqualFold(msg.Type, ng.EventType_OPERATION) {
		// accountId := msg.CloudAccountId
		switch msg.EventSubType {

		case billingCommon.CloudCreditsExpiredEventSubType:
			if err := w.HandleCreditsUsed(ctx, msg, true); err != nil {
				return err
			}
		case billingCommon.CloudCreditsAboutToExpireEventSubType:
			logger.Info("No action required to be taken for CloudCreditsAboutToExpire event.")
			if err := w.deleteMessageFromQueue(ctx, msgId, rhandle); err != nil {
				return err
			}
			return nil
		case billingCommon.CloudCreditsAvailableEventSubType:
			if err := w.HandleCreditsAvailable(ctx, msg); err != nil {
				return err
			}
		case billingCommon.CloudCreditsThresholdReachedEventSubType:
			if err := w.HandleThresholdReached(ctx, msg); err != nil {
				return err
			}
		case billingCommon.CloudCreditsUsedEventSubType:
			if err := w.HandleCreditsUsed(ctx, msg, false); err != nil {
				return err
			}
		default:
			logger.V(1).Info("invalid message subtype, ignoring this message..")
			return nil
		}

		if err := w.CreateStateLog(ctx, msg.CloudAccountId, msg.EventSubType, eventTime); err != nil {
			logger.Error(err, "unable to createStateLogEntry")
			return err
		}

		if err := w.deleteMessageFromQueue(ctx, msgId, rhandle); err != nil {
			return err
		}

	} else {
		logger.V(1).Info("invalid message type, ignoring this message..")
	}
	return nil
}

func (w CloudCreditsWorker) deleteMessageFromQueue(ctx context.Context, messageId string, receiptHandle string) error {
	logger := log.FromContext(ctx).WithName("CloudCreditsWorkerService.deleteMessageFromQueue")

	// delete message from the queue.
	logger.Info("deleting message from queue", "messageId", messageId)

	obj := pb.DeleteEventRequest{MessageId: messageId, ReceiptHandle: receiptHandle}
	request := pb.DeleteEventRequestList{QueueName: config.Cfg.GetAWSSQSQueueName(), DeleteEventRequests: []*pb.DeleteEventRequest{&obj}}
	_, err := w.NotificationGatewayClient.NotificationGatewayServiceClient.DeleteEvents(ctx, &request)
	if err != nil {
		logger.Error(err, "failed to delete message from Queue")
		return err
	}
	return nil
}

func (*CloudCreditsWorker) Name() string {
	return "cloudcredits-worker"
}

func parseTime(s string) (*time.Time, error) {
	layout := "2006-01-02T15:04:05Z"
	s = strings.TrimSuffix(s, " 00:00:00")
	t, err := time.Parse(layout, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (w CloudCreditsWorker) CreateStateLog(ctx context.Context, cloudaccountId string, eventState string, eventTime *timestamppb.Timestamp) error {
	logger := log.FromContext(ctx).WithName("getCreditStateFromEvent")
	logger.Info("add credits_state_log", "cloudaccountId", cloudaccountId, "state", eventState)
	state := getCreditStateFromEvent(ctx, eventState)
	in := pb.CreditsState{
		CloudAccountId: cloudaccountId,
		State:          state,
		EventAt:        eventTime,
	}
	_, err := CloudCreditsClient.CloudCreditsSvcClient.CreateCreditStateLog(ctx, &in)
	if err != nil {
		return err
	}
	return nil
}

func getCreditStateFromEvent(ctx context.Context, eventSubType string) pb.CloudCreditsState {
	logger := log.FromContext(ctx).WithName("getCreditStateFromEvent")
	switch eventSubType {
	case billingCommon.CloudCreditsExpiredEventSubType:
		return *pb.CloudCreditsState_CLOUD_CREDITS_STATE_EXPIRED.Enum()
	case billingCommon.CloudCreditsAvailableEventSubType:
		return *pb.CloudCreditsState_CLOUD_CREDITS_STATE_AVAILABLE.Enum()
	case billingCommon.CloudCreditsThresholdReachedEventSubType:
		return *pb.CloudCreditsState_CLOUD_CREDITS_STATE_THRESHOLD_REACHED.Enum()
	case billingCommon.CloudCreditsUsedEventSubType:
		return *pb.CloudCreditsState_CLOUD_CREDITS_STATE_USED.Enum()
	case billingCommon.CloudCreditOutDatedEvent:
		return *pb.CloudCreditsState_CLOUD_CREDITS_STATE_OUTDATED.Enum()
	default:
		logger.Info("invalid message subtype, can't get credit state")
		return *pb.CloudCreditsState_CLOUD_CREDITS_STATE_UNSPECIFIED.Enum()
	}
}
