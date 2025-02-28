// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

var (
	eventExpiryTicker *time.Ticker
)

func StartEventExpiryScheduler(ctx context.Context, eventExpiryScheduler *EventExpiryScheduler, eventExpirySchedulerInterval uint16) {
	eventExpiryTicker = time.NewTicker(time.Duration(eventExpirySchedulerInterval) * time.Minute)
	go eventExpiryLoop(context.Background(), eventExpiryScheduler)
}

func eventExpiryLoop(ctx context.Context, eventExpiryScheduler *EventExpiryScheduler) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventExpiryScheduler.eventExpiryLoop").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	for {
		eventExpiry(ctx, &logger, eventExpiryScheduler)
		tm := <-eventExpiryTicker.C
		if tm.IsZero() {
			return
		}
	}
}

func eventExpiry(ctx context.Context, logger *logr.Logger, eventExpiryScheduler *EventExpiryScheduler) {
	err := eventExpiryScheduler.handleEventExpiry(ctx)
	if err != nil {
		logger.Error(err, "failed to handle event expiry")
	}
}

type EventExpiryScheduler struct {
	eventData *EventData
}

func NewEventExpiryScheduler(eventData *EventData) *EventExpiryScheduler {
	return &EventExpiryScheduler{eventData: eventData}
}

func (eventExpiryScheduler *EventExpiryScheduler) handleEventExpiry(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("EventExpiryScheduler.handleEventExpiry").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	notifications, err := eventExpiryScheduler.eventData.GetNotifications(ctx)
	if err != nil {
		logger.Error(err, "failed to get notifications")
	}
	for _, notification := range notifications {
		if time.Now().Before(notification.Expiration) {
			err = eventExpiryScheduler.eventData.deleteNotification(ctx, notification.Id)
			if err != nil {
				logger.Error(err, "failed to delete notification upon expiry", "notificationId", notification.Id)
			}
		}
	}

	alerts, err := eventExpiryScheduler.eventData.GetAlerts(ctx)
	if err != nil {
		logger.Error(err, "failed to get alerts")
	}
	for _, alert := range alerts {
		if time.Now().Before(alert.Expiration) {
			err = eventExpiryScheduler.eventData.deleteAlert(ctx, alert.Id)
			if err != nil {
				logger.Error(err, "failed to delete alert upon expiry", "alertId", alert.Id)
			}
		}
	}
	return nil
}
