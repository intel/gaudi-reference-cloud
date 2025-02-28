// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

var maxId int64 = 1_000_000_000_000

const (
	defaultExpirationOfEventYears  = 0
	defaultExpirationOfEventMonths = 1
	defaultExpirationOfEventDays   = 0
	notificationsTable             = "notifications"
	alertsTable                    = "alerts"
)

func NewId() (string, error) {
	intId, err := rand.Int(rand.Reader, big.NewInt(maxId))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%012d", intId), nil
}

func MustNewId() string {
	id, err := NewId()
	if err != nil {
		panic(err)
	}
	return id
}

type EventManager struct {
	eventData       *EventData
	eventDispatcher *EventDispatcher
}

func NewEventManager(eventData *EventData, eventDispatcher *EventDispatcher) *EventManager {
	return &EventManager{eventData: eventData, eventDispatcher: eventDispatcher}
}

func (em *EventManager) RegisterService(ctx context.Context, registerServiceEvents RegisterServiceEvents) error {
	return nil
}

func (em *EventManager) Create(ctx context.Context, createEvent CreateEvent) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("EventManager.create").WithValues("cloudAccountId", createEvent.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	eventId, err := NewId()

	if err != nil {
		logger.Error(err, "failed to generate id")
	}

	eventBase := EventBase{
		Id:             eventId,
		CloudAccountId: createEvent.CloudAccountId,
		UserId:         createEvent.UserId,
		Creation:       time.Now(),
		Expiration:     time.Now().AddDate(defaultExpirationOfEventYears, 1, defaultExpirationOfEventDays),
		Status:         createEvent.Status,
		Severity:       createEvent.Severity,
		Properties:     createEvent.Properties,
		ClientRecordId: createEvent.ClientRecordId,
	}
	switch createEvent.Type {

	case EventType_NOTIFICATION:
		notification := Notification{
			EventBase:        eventBase,
			NotificationType: createEvent.EventSubType,
			ServiceName:      createEvent.ServiceName,
			Message:          createEvent.Message,
		}
		if err = em.eventData.storeNotification(ctx, notification); err != nil {
			return err
		}
		return em.eventDispatcher.dispatchNotification(ctx, notification)

	case EventType_ALERT:
		alert := Alert{
			EventBase:   eventBase,
			AlertType:   createEvent.EventSubType,
			ServiceName: createEvent.ServiceName,
			Message:     createEvent.Message,
		}
		if err = em.eventData.storeAlert(ctx, alert); err != nil {
			return err
		}
		return em.eventDispatcher.dispatchAlert(ctx, alert)

	case EventType_EMAIL:
		email := Email{
			EventBase:   eventBase,
			AlertType:   createEvent.EventSubType,
			ServiceName: createEvent.ServiceName,
		}
		return em.eventDispatcher.dispatchEmail(ctx, email)
	case EventType_ERROR:
		errorRecord := Error{
			EventBase:   eventBase,
			ServiceName: createEvent.ServiceName,
			Message:     createEvent.Message,
			Region:      createEvent.Region,
		}
		if err = em.eventData.storeError(ctx, errorRecord); err != nil {
			return err
		}
		return em.eventDispatcher.dispatchError(ctx, errorRecord)
	default:
		return fmt.Errorf("invalid event type %v", createEvent.Type)
	}
}

func (em *EventManager) getNotificationsForCloudAcct(ctx context.Context, cloudAcctId string) ([]*Notification, error) {
	logger := log.FromContext(ctx).WithName("EventManager.getNotificationsForCloudAcct")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	return em.eventData.getNotificationsForCloudAcct(ctx, cloudAcctId)
}

func (em *EventManager) getAlertsForCloudAcct(ctx context.Context, cloudAcctId string) ([]*Alert, error) {
	logger := log.FromContext(ctx).WithName("EventManager.getAlertsForCloudAcct")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	return em.eventData.getAlertsForCloudAcct(ctx, cloudAcctId)
}

func (em *EventManager) dismissEvent(ctx context.Context, cloudAcctId string, clientRecordId string) error {
	logger := log.FromContext(ctx).WithName("EventManager.dismissEvent")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	eventBase, err := em.eventData.getEvent(ctx, notificationsTable, cloudAcctId, clientRecordId)
	if err != nil {
		logger.Error(err, "failed to get notficication", "cloudAccountId", cloudAcctId, "clientRecordId", clientRecordId)
		return err
	}
	if eventBase != nil {
		logger.Info("delete notification", "eventBase", eventBase)
		err = em.eventData.deleteNotification(ctx, eventBase.Id)
		if err != nil {
			logger.Error(err, "failed to delete notficication", "cloudAccountId", cloudAcctId, "clientRecordId", clientRecordId)
			return err
		}
	}
	eventBase, err = em.eventData.getEvent(ctx, alertsTable, cloudAcctId, clientRecordId)
	if err != nil {
		logger.Error(err, "failed to get alert", "cloudAccountId", cloudAcctId, "clientRecordId", clientRecordId)
		return err
	}
	if eventBase != nil {
		logger.Info("delete alert", "eventBase", eventBase)
		err = em.eventData.deleteAlert(ctx, eventBase.Id)
		if err != nil {
			logger.Error(err, "failed to delete alert", "cloudAccountId", cloudAcctId, "clientRecordId", clientRecordId)
			return err
		}
	}
	return nil
}

// todo: this will be implemented for ops.
func (em *EventManager) getNotifications(ctx context.Context) ([]*Notification, error) {
	return nil, nil
}

// todo: this will be implemented for ops.
func (em *EventManager) getAlerts(ctx context.Context) ([]*Alert, error) {
	return nil, nil
}

// to subscribe to long polling
func (em *EventManager) subscribe(ctx context.Context, eventSubscribe *EventsSubscribe) error {
	return nil
}
