// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sqsutil"
	"go.opentelemetry.io/otel/attribute"
)

const (
	dunning_eventMessageType       = "Dunning"
	failedPayment_eventMessageType = "FailedPayment"
	dunning_eventSubtype           = "DUNNING_STATE_CHANGED"
	failedPayment_eventSubtype     = "FAILED_PAYMENT"
	dunning_alertMessage           = "this is dunning alert"
	failedPayment_alertMessage     = "this is failed payment alert"
	sqsConnTimeout                 = 1
	sqsMaxMessages                 = 10
	sqsMessageAttrs                = ""
	electronicPaymentFailedMsg     = "Electronic Payment Failed"
	accountDunningMsg              = "Account Master Plan Instance Dunning Degrade Date Changed"
	dunningStateCompleted          = "2"
)

var (
	paidServicesDeactivationControllerChannel = make(chan bool)
	paidServicesDeactivationControllerTicker  *time.Ticker
)

func StartPaidServicesDeactivationController(ctx context.Context, paidServicesDeactivationController *PaidServicesDeactivationController) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PaidServicesDeactivationController.StartPaidServicesDeactivationController").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("cfg", "PaidServicesDeactivationControllerInterval", config.Cfg.PaidServicesDeactivationControllerInterval)

	paidServicesDeactivationControllerTicker = time.NewTicker(time.Duration(config.Cfg.PaidServicesDeactivationControllerInterval) * time.Second)
	go paidServicesDeactivationControllerLoop(ctx, paidServicesDeactivationController)
}

func StopPaidServicesDeactivationControllerScheduler() {
	if paidServicesDeactivationControllerChannel != nil {
		close(paidServicesDeactivationControllerChannel)
		paidServicesDeactivationControllerChannel = nil
	}
}

func paidServicesDeactivationControllerLoop(ctx context.Context, paidServicesDeactivationController *PaidServicesDeactivationController) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PaidServicesDeactivationController.paidServicesDeactivationControllerLoop").Start()
	defer span.End()
	for {
		err := paidServicesDeactivationController.GetAriaSqsEventsForDeativation(ctx)
		if err != nil {
			logger.Error(err, "failed to execute getAriaSqsEventsForDeativation")
		}
		select {
		case <-paidServicesDeactivationControllerChannel:
			return
		case tm := <-paidServicesDeactivationControllerTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

type PaidServicesDeactivationController struct {
	ariaSqsUtil        *sqsutil.SQSUtil
	eventManager       *events.EventManager
	cloudAccountClient pb.CloudAccountServiceClient
}

func NewPaidServicesDeactivationController(ariaSqsUtil *sqsutil.SQSUtil, eventManager *events.EventManager, cloudAccountClient pb.CloudAccountServiceClient) *PaidServicesDeactivationController {
	return &PaidServicesDeactivationController{ariaSqsUtil: ariaSqsUtil, eventManager: eventManager, cloudAccountClient: cloudAccountClient}
}

func (paidServicesDeactivationController *PaidServicesDeactivationController) determineMessageTypeAndAccountId(ctx context.Context, messageBody string) (string, string, string) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PaidServicesDeactivationController.determineMessageTypeAndAccountId").Start()
	defer span.End()
	logger.V(1).Info("BEGIN")
	defer logger.V(1).Info("END")
	messageData, clientAcctId, mpiDunningState, err := paidServicesDeactivationController.parseMessageBody(ctx, messageBody)
	messageType := ""
	accountId := ""
	if err != nil {
		logger.Error(err, "failed to parse message body", "messageBody", messageBody)
		return messageType, accountId, mpiDunningState
	}
	if strings.EqualFold(messageData, accountDunningMsg) {
		messageType = dunning_eventMessageType
	} else if strings.EqualFold(messageData, electronicPaymentFailedMsg) {
		messageType = failedPayment_eventMessageType
	}

	if !strings.EqualFold(clientAcctId, "") && strings.Contains(clientAcctId, ".") {
		clientAcctId := strings.Split(clientAcctId, ".")
		accountId = clientAcctId[1]
	}
	span.SetAttributes(attribute.String("cloudAccountId", accountId))

	logger.V(1).Info("message type and cloud account id", "messageType", messageType, "cloudAccountId", accountId)
	return messageType, accountId, mpiDunningState
}

func (paidServicesDeactivationController *PaidServicesDeactivationController) parseMessageBody(ctx context.Context, messageBody string) (string, string, string, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PaidServicesDeactivationController.parseMessageBody").Start()
	defer span.End()
	// Parse the message body
	queryValues, err := url.ParseQuery(messageBody)
	if err != nil {
		logger.Error(err, "error parsing aria event")
		return "", "", "", err
	}
	event := ""
	clientAcctId := queryValues.Get("client_acct_id")
	if !strings.Contains(clientAcctId, config.Cfg.ClientIdPrefix) {
		return event, "", "", nil
	}

	mpiDunningState := queryValues.Get("mpi_dunning_state[]")
	eventLabelValue := queryValues.Get("event_label[]")
	if strings.Contains(eventLabelValue, "Dunning") {
		return eventLabelValue, clientAcctId, mpiDunningState, nil
	}
	logger.Info("event label[]", "event_label", eventLabelValue)
	for key, values := range queryValues {
		if strings.Contains(key, "event_label") {
			if len(values) > 0 {
				for _, val := range values {
					if strings.Contains(val, "Payment Failed") {
						event = val
					}
				}
			}
			break
		}
	}
	return event, clientAcctId, mpiDunningState, nil
}

func (paidServicesDeactivationController *PaidServicesDeactivationController) createPaymentAlerts(ctx context.Context, cloudAcctId string, message string, eventSubtype string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PaidServicesDeactivationController.createPaymentAlerts").WithValues("cloudAccountId", cloudAcctId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	err := paidServicesDeactivationController.eventManager.Create(ctx, events.CreateEvent{
		Status:         events.EventStatus_ACTIVE,
		Type:           events.EventType_ALERT,
		Severity:       events.EventSeverity_HIGH,
		ServiceName:    "billing",
		Message:        message,
		CloudAccountId: cloudAcctId,
		EventSubType:   eventSubtype,
		ClientRecordId: uuid.NewString(),
	})
	if err != nil {
		logger.Error(err, "failed to create alert for failed payment")
		return err
	}

	return nil
}

func (paidServicesDeactivationController *PaidServicesDeactivationController) GetAriaSqsEventsForDeativation(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PaidServicesDeactivationController.GetAriaSQSEventsForDeativation").Start()
	defer span.End()
	logger.Info("getting Aria SQS Events for failed payments and dunning activities.")

	messages, err := paidServicesDeactivationController.ariaSqsUtil.ReceiveMessage(ctx, sqsConnTimeout, sqsMaxMessages, sqsMessageAttrs)
	if err != nil {
		logger.Error(err, "could not retrieve messages from the aria event queue")
		return err
	}

	if len(messages) > 0 {
		for _, message := range messages {
			messageType, accountId, mpiDunningState := paidServicesDeactivationController.determineMessageTypeAndAccountId(ctx, *message.Body)
			if messageType == "" || accountId == "" {
				continue
			}
			if (messageType == failedPayment_eventMessageType || messageType == dunning_eventMessageType) && accountId == "" {
				logger.Error(fmt.Errorf("empty cloudAccountId"), "no valid account id found, ignoring this message")
				continue
			}
			var cloudAcct *pb.CloudAccount
			if strings.EqualFold(messageType, failedPayment_eventMessageType) || strings.EqualFold(messageType, dunning_eventMessageType) {
				cloudAcct, err = paidServicesDeactivationController.cloudAccountClient.GetById(ctx, &pb.CloudAccountId{Id: accountId})
				logger.V(1).Info("cloud account event", "cloudAcct", cloudAcct, "messageType", messageType)
				if err != nil {
					logger.Error(err, "error getting acct details from cloud acct service")
					continue
				}
				if cloudAcct == nil {
					logger.V(1).Info("cloud account not found", "cloudAcct", cloudAcct)
					continue
				}
			}

			switch messageType {

			// Failed Payments Events
			case failedPayment_eventMessageType:
				logger.Info("electronic payment failed", "cloudAccountId", accountId)
				if (cloudAcct != nil) && cloudAcct.PaidServicesAllowed {
					paidServicesAllowed := false
					_, err := paidServicesDeactivationController.cloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, PaidServicesAllowed: &paidServicesAllowed})
					if err != nil {
						logger.Error(err, "failed to update cloud account paid services allowed", "cloudAccountId", accountId)
						continue
					}
				}

				// generate alert notification for failed payment
				alert_err := paidServicesDeactivationController.createPaymentAlerts(ctx, accountId, dunning_alertMessage, dunning_eventSubtype)
				if alert_err != nil {
					logger.Error(alert_err, "failed to create payment alert", alert_err)
				}
				// delete message from the queue.
				err := paidServicesDeactivationController.ariaSqsUtil.DeleteMessage(ctx, message)
				if err != nil {
					logger.Error(err, "failed to delete message", err)
					continue
				}

			// Dunning Events
			case dunning_eventMessageType:
				logger.Info("account master plan instance dunning state changed", "cloudAccountId", accountId)

				if (cloudAcct != nil) && !cloudAcct.TerminatePaidServices && strings.EqualFold(mpiDunningState, dunningStateCompleted) {
					terminatePaidServices := true
					_, err := paidServicesDeactivationController.cloudAccountClient.Update(ctx, &pb.CloudAccountUpdate{Id: cloudAcct.Id, TerminatePaidServices: &terminatePaidServices})
					if err != nil {
						logger.Error(err, "failed to update cloud account terminate paid services", "cloudAccountId", accountId)
						continue
					}
				}

				// generate alert notification for dunning state changed
				if err := paidServicesDeactivationController.createPaymentAlerts(ctx, accountId, failedPayment_alertMessage, failedPayment_eventSubtype); err != nil {
					logger.Error(err, "failed to create payment alert", err)
					return err
				}

				// delete message from the queue.
				err := paidServicesDeactivationController.ariaSqsUtil.DeleteMessage(ctx, message)
				if err != nil {
					logger.Error(err, "failed to delete message")
					return err
				}

			default:
				logger.Info("invalid messageType, ignoring this message.")
			}
		}

	} else {
		logger.Info("no events found for failed payment or dunning in aria sqs")
	}

	return nil
}
