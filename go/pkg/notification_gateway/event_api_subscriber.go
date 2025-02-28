// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"
	"sync"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const waitForClientMins = 1

type ApiSubscriber struct {
	stream               pb.NotificationGatewayService_SubscribeServer
	apiSubscriberChannel chan<- bool
}

type EventApiSubscriber struct {
	apiSubscribers sync.Map
}

func NewEventApiSubscriber() *EventApiSubscriber {
	return &EventApiSubscriber{}
}

func (ep *EventApiSubscriber) addSubscriber(subscriberId string, apiSubscriber ApiSubscriber) {
	ep.apiSubscribers.Store(subscriberId, apiSubscriber)
}

func (ep *EventApiSubscriber) sendNotification(ctx context.Context, notification Notification) {
	logger := log.FromContext(ctx).WithName("EventManager.sendNotification")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	go ep.sendNotificationToSubscribers(ctx, notification)
}

func (ep *EventApiSubscriber) sendNotificationToSubscribers(ctx context.Context, notification Notification) error {
	logger := log.FromContext(ctx).WithName("EventManager.sendNotificationToSubscribers")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	for {
		time.Sleep(waitForClientMins)

		var unsubscribe []string

		// Iterate over all subscribers and send data to each client
		ep.apiSubscribers.Range(func(k, v interface{}) bool {
			clientId, ok := k.(string)
			if !ok {
				return false
			}
			apiSubscrinber, ok := v.(ApiSubscriber)
			if !ok {
				return false
			}
			var notificationsR []*pb.Notification
			notificationR := &pb.Notification{
				Id:               notification.Id,
				CloudAccountId:   &notification.CloudAccountId,
				UserId:           &notification.UserId,
				NotificationType: notification.NotificationType,
				ServiceName:      &notification.ServiceName,
				Creation:         timestamppb.New(notification.Creation),
				Expiration:       timestamppb.New(notification.Expiration),
				Status:           notification.Status,
				Severity:         &notification.Severity,
				Message:          &notification.Message,
				Properties:       notification.Properties,
				ClientRecordId:   notification.ClientRecordId,
			}

			notificationsR = append(notificationsR, notificationR)
			if err := apiSubscrinber.stream.Send(&pb.Events{
				NumberOfNotifications: int32(1),
				NumberOfAlerts:        int32(0),
				Notifications:         notificationsR,
				Alerts:                nil,
			}); err != nil {
				select {
				case apiSubscrinber.apiSubscriberChannel <- true:
					logger.Info("unsubscribing client with", "id", clientId)
				default:
				}
				unsubscribe = append(unsubscribe, clientId)
			}
			return true
		})

		for _, id := range unsubscribe {
			ep.apiSubscribers.Delete(id)
		}
	}
}

func (ep *EventApiSubscriber) sendAlert(ctx context.Context, alert Alert) {
	logger := log.FromContext(ctx).WithName("EventManager.sendAlert")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	go ep.sendAlertToSubscribers(ctx, alert)
}

func (ep *EventApiSubscriber) sendAlertToSubscribers(ctx context.Context, alert Alert) error {
	logger := log.FromContext(ctx).WithName("EventManager.sendAlertToSubscribers")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	for {
		time.Sleep(waitForClientMins)

		var unsubscribe []string

		// Iterate over all subscribers and send data to each client
		ep.apiSubscribers.Range(func(k, v interface{}) bool {
			clientId, ok := k.(string)
			if !ok {
				return false
			}
			apiSubscrinber, ok := v.(ApiSubscriber)
			if !ok {
				return false
			}
			var alertsR []*pb.Alert
			alertR := &pb.Alert{
				Id:             alert.Id,
				CloudAccountId: &alert.CloudAccountId,
				UserId:         &alert.UserId,
				AlertType:      alert.AlertType,
				ServiceName:    &alert.ServiceName,
				Creation:       timestamppb.New(alert.Creation),
				Expiration:     timestamppb.New(alert.Expiration),
				Status:         alert.Status,
				Severity:       &alert.Severity,
				Message:        &alert.Message,
				Properties:     alert.Properties,
				ClientRecordId: alert.ClientRecordId,
			}

			alertsR = append(alertsR, alertR)
			if err := apiSubscrinber.stream.Send(&pb.Events{
				NumberOfNotifications: int32(0),
				NumberOfAlerts:        int32(1),
				Notifications:         nil,
				Alerts:                alertsR,
			}); err != nil {
				select {
				case apiSubscrinber.apiSubscriberChannel <- true:
					logger.Info("unsubscribing client with", "id", clientId)
				default:
				}
				unsubscribe = append(unsubscribe, clientId)
			}
			return true
		})

		for _, id := range unsubscribe {
			ep.apiSubscribers.Delete(id)
		}
	}
}
