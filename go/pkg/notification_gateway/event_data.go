// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

const (
	insertNotificationRecordQuery = `
		INSERT INTO notifications (id,cloud_account_id,creation,expiration,status,notification_type,severity,service_name,message,properties,client_record_id) 
		VALUES ($1, $2, $3, $4,$5,$6,$7,$8,$9,$10,$11)
	`

	insertAlertRecordQuery = `
		INSERT INTO alerts (id,cloud_account_id,creation,expiration,status,alert_type,severity,service_name,message,properties,client_record_id) 
		VALUES ($1, $2, $3, $4,$5,$6,$7,$8,$9,$10,$11)
	`

	insertErrorRecordQuery = `
		INSERT INTO errors (id,cloud_account_id,creation,expiration,status,error_type,severity,service_name,message,properties,client_record_id, region)
		VALUES ($1, $2, $3, $4,$5,$6,$7,$8,$9,$10,$11,$12)
		`

	searchNotificationsForCloudAcctQuery = `
	SELECT id, cloud_account_id, creation, expiration, status, notification_type, severity, service_name, message, properties, client_record_id
	FROM notifications WHERE cloud_account_id = $1 ORDER BY creation
	`

	searchAlertsForCloudAcctQuery = `
	SELECT id, cloud_account_id, creation, expiration, status, alert_type, severity, service_name, message, properties, client_record_id
	FROM alerts WHERE cloud_account_id = $1 ORDER BY creation
	`

	searchNotifications = `
	SELECT id, cloud_account_id, creation, expiration, status, notification_type, severity, service_name, message, properties, client_record_id
	FROM notifications ORDER BY creation
	`

	searchAlerts = `
	SELECT id, cloud_account_id, creation, expiration, status, alert_type, severity, service_name, message, properties, client_record_id
	FROM alerts ORDER BY creation
	`
)

type EventData struct {
	db *sql.DB
}

func NewEventData(db *sql.DB) *EventData {
	return &EventData{db: db}
}

func (ed *EventData) storeNotification(ctx context.Context, notificationRecord Notification) error {
	logger := log.FromContext(ctx).WithName("EventData.storeNotification")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	_, err := ed.db.ExecContext(ctx,
		insertNotificationRecordQuery,
		notificationRecord.Id,
		notificationRecord.CloudAccountId,
		notificationRecord.Creation,
		notificationRecord.Expiration,
		notificationRecord.Status,
		notificationRecord.NotificationType,
		notificationRecord.Severity,
		notificationRecord.ServiceName,
		notificationRecord.Message,
		notificationRecord.Properties,
		notificationRecord.ClientRecordId,
	)

	if err != nil {
		logger.Error(err, "failed to store notification record")
		return err
	}

	return nil
}

func (ed *EventData) storeError(ctx context.Context, errorRecord Error) error {
	logger := log.FromContext(ctx).WithName("EventData.storeError")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	_, err := ed.db.ExecContext(ctx,
		insertErrorRecordQuery,
		errorRecord.Id,
		errorRecord.CloudAccountId,
		errorRecord.Creation,
		errorRecord.Expiration,
		errorRecord.Status,
		errorRecord.ErrorType,
		errorRecord.Severity,
		errorRecord.ServiceName,
		errorRecord.Message,
		errorRecord.Properties,
		errorRecord.ClientRecordId,
		errorRecord.Region,
	)

	if err != nil {
		logger.Error(err, "failed to store error record")
		return err
	}

	return nil
}

func (ed *EventData) storeCloudCreditsEvent(ctx context.Context, event CloudCreditsEvent) error {
	logger := log.FromContext(ctx).WithName("EventData.storeNotification")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	_, err := ed.db.ExecContext(ctx,
		insertNotificationRecordQuery,
		event.GetEventBase().Id,
		event.GetEventBase().CloudAccountId,
		event.GetEventBase().Creation,
		event.GetEventBase().Expiration,
		event.GetEventBase().Status,
		event.GetEventName(),
		event.GetEventBase().Severity,
		event.ServiceName,
		event.Message,
		event.GetEventBase().Properties,
		event.GetEventBase().ClientRecordId,
	)

	if err != nil {
		logger.Error(err, "failed to store notification record")
		return err
	}

	return nil
}

func (ed *EventData) storeAlert(ctx context.Context, alertRecord Alert) error {
	logger := log.FromContext(ctx).WithName("EventData.storeAlert")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	_, err := ed.db.ExecContext(ctx,
		insertAlertRecordQuery,
		alertRecord.Id,
		alertRecord.CloudAccountId,
		alertRecord.Creation,
		alertRecord.Expiration,
		alertRecord.Status,
		alertRecord.AlertType,
		alertRecord.Severity,
		alertRecord.ServiceName,
		alertRecord.Message,
		alertRecord.Properties,
		alertRecord.ClientRecordId,
	)

	if err != nil {
		logger.Error(err, "failed to store alert record")
		return err
	}

	return nil
}

func (ed *EventData) getNotificationsForCloudAcct(ctx context.Context, cloudAcctId string) ([]*Notification, error) {
	logger := log.FromContext(ctx).WithName("EventData.getNotificationsForCloudAcct")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	rows, err := ed.db.QueryContext(ctx, searchNotificationsForCloudAcctQuery, cloudAcctId)
	if err != nil {
		logger.Error(err, "failed to find notifications for cloud acct", "id", cloudAcctId)
		return nil, err
	}

	var notifications []*Notification
	defer rows.Close()
	for rows.Next() {
		notification := Notification{}
		props := []byte{}

		if err := rows.Scan(&notification.Id,
			&notification.CloudAccountId, &notification.Creation,
			&notification.Expiration, &notification.Status,
			&notification.NotificationType, &notification.Severity,
			&notification.ServiceName, &notification.Message, &props, &notification.ClientRecordId); err != nil {

			logger.Error(err, "failed to read notifications for cloud acct", "id", cloudAcctId)
			return nil, err
		}

		if props != nil {
			if err := json.Unmarshal([]byte(props), &notification.Properties); err != nil {
				logger.Error(err, "invalid properties for notification for cloud acct", "id", cloudAcctId)
				return nil, err
			}
		}

		notifications = append(notifications, &notification)

	}

	return notifications, nil
}

func (ed *EventData) getAlertsForCloudAcct(ctx context.Context, cloudAcctId string) ([]*Alert, error) {
	logger := log.FromContext(ctx).WithName("EventData.getAlertsForCloudAcct")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	rows, err := ed.db.QueryContext(ctx, searchAlertsForCloudAcctQuery, cloudAcctId)
	if err != nil {
		logger.Error(err, "failed to find alerts for cloud acct", "id", cloudAcctId)
		return nil, err
	}

	var alerts []*Alert
	defer rows.Close()
	for rows.Next() {
		alert := Alert{}
		props := []byte{}

		if err := rows.Scan(&alert.Id,
			&alert.CloudAccountId, &alert.Creation,
			&alert.Expiration, &alert.Status,
			&alert.AlertType, &alert.Severity,
			&alert.ServiceName, &alert.Message, &props, &alert.ClientRecordId); err != nil {

			logger.Error(err, "failed to read alerts for cloud acct", "id", cloudAcctId)
			return nil, err
		}

		if props != nil {
			if err := json.Unmarshal([]byte(props), &alert.Properties); err != nil {
				logger.Error(err, "invalid properties for alerts for cloud acct", "id", cloudAcctId)
				return nil, err
			}
		}

		alerts = append(alerts, &alert)

	}

	return alerts, nil
}

func (ed *EventData) GetNotifications(ctx context.Context) ([]*Notification, error) {
	logger := log.FromContext(ctx).WithName("EventData.getNotifications")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	rows, err := ed.db.QueryContext(ctx, searchNotifications)
	if err != nil {
		logger.Error(err, "failed to find notifications")
		return nil, err
	}

	var notifications []*Notification
	defer rows.Close()
	for rows.Next() {
		notification := Notification{}
		props := []byte{}

		if err := rows.Scan(&notification.Id,
			&notification.CloudAccountId, &notification.Creation,
			&notification.Expiration, &notification.Status,
			&notification.NotificationType, &notification.Severity,
			&notification.ServiceName, &notification.Message, &props, &notification.ClientRecordId); err != nil {

			logger.Error(err, "failed to read notifications")
			return nil, err
		}

		if props != nil {
			if err := json.Unmarshal([]byte(props), &notification.Properties); err != nil {
				logger.Error(err, "invalid properties for notification")
				return nil, err
			}
		}

		notifications = append(notifications, &notification)

	}

	return notifications, nil
}

func (ed *EventData) GetAlerts(ctx context.Context) ([]*Alert, error) {
	logger := log.FromContext(ctx).WithName("EventData.GetAlerts")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	rows, err := ed.db.QueryContext(ctx, searchAlerts)
	if err != nil {
		logger.Error(err, "failed to find alerts")
		return nil, err
	}

	var alerts []*Alert
	defer rows.Close()
	for rows.Next() {
		alert := Alert{}
		props := []byte{}

		if err := rows.Scan(&alert.Id,
			&alert.CloudAccountId, &alert.Creation,
			&alert.Expiration, &alert.Status,
			&alert.AlertType, &alert.Severity,
			&alert.ServiceName, &alert.Message, &props, &alert.ClientRecordId); err != nil {

			logger.Error(err, "failed to read alerts")
			return nil, err
		}

		if props != nil {
			if err := json.Unmarshal([]byte(props), &alert.Properties); err != nil {
				logger.Error(err, "invalid properties for alerts")
				return nil, err
			}
		}

		alerts = append(alerts, &alert)

	}

	return alerts, nil
}

func (ed *EventData) deleteNotification(ctx context.Context, notificationId string) error {
	logger := log.FromContext(ctx).WithName("EventData.deleteNotification")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	deleteNotification := "DELETE FROM notifications WHERE id = $1"
	_, err := ed.db.ExecContext(ctx, deleteNotification, notificationId)
	if err != nil {
		logger.Error(err, "failed to delete notification", "id", notificationId)
		return err
	}
	return nil
}

func (ed *EventData) deleteAlert(ctx context.Context, alertId string) error {
	logger := log.FromContext(ctx).WithName("EventData.deleteAlert")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	deleteAlert := "DELETE FROM alerts WHERE id = $1"
	_, err := ed.db.ExecContext(ctx, deleteAlert, alertId)
	if err != nil {
		logger.Error(err, "failed to delete alert", "id", alertId)
		return err
	}
	return nil
}

func (ed *EventData) getEvent(ctx context.Context, eventTable string, cloudAccountId string, clientRecordId string) (*EventBase, error) {
	logger := log.FromContext(ctx).WithName("EventData.getEvent")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	selectEvent := fmt.Sprintf("SELECT  id, cloud_account_id, creation, expiration, status, severity, properties, client_record_id FROM %s WHERE cloud_account_id = $1 AND client_record_id = $2", eventTable)
	rows, err := ed.db.QueryContext(ctx, selectEvent, cloudAccountId, clientRecordId)
	if err != nil {
		logger.Error(err, "failed to get event", "eventTable", eventTable, "cloudAccountId", cloudAccountId, "clientRecordId", clientRecordId)
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	eventBase, err := ed.scanEventBase(ctx, rows)
	if err != nil {
		logger.Error(err, "failed to scan event", "eventTable", eventTable, "cloudAccountId", cloudAccountId, "clientRecordId", clientRecordId)
		return nil, err
	}
	logger.Info("Event detail", "eventBase", eventBase)
	return eventBase, nil
}

func (*EventData) scanEventBase(ctx context.Context, rows *sql.Rows) (*EventBase, error) {
	logger := log.FromContext(ctx).WithName("EventData.getNotification")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	eventBase := EventBase{}
	props := []byte{}
	if err := rows.Scan(&eventBase.Id,
		&eventBase.CloudAccountId, &eventBase.Creation,
		&eventBase.Expiration, &eventBase.Status,
		&eventBase.Severity, &props, &eventBase.ClientRecordId); err != nil {

		logger.Error(err, "failed to read event")
		return nil, err
	}
	if props != nil {
		if err := json.Unmarshal([]byte(props), &eventBase.Properties); err != nil {
			logger.Error(err, "invalid properties for event")
			return nil, err
		}
	}
	return &eventBase, nil
}

func (ed *EventData) DeleteAllNotifications(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("EventData.DeleteAllNotifications")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	deleteNotification := "DELETE FROM notifications"
	_, err := ed.db.ExecContext(ctx, deleteNotification)
	if err != nil {
		logger.Error(err, "failed to delete all notifications")
		return err
	}
	return nil
}

func (ed *EventData) DeleteAllAlerts(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("EventData.DeleteAllAlerts")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	deleteAlert := "DELETE FROM alerts"
	_, err := ed.db.ExecContext(ctx, deleteAlert)
	if err != nil {
		logger.Error(err, "failed to delete all alerts")
		return err
	}
	return nil
}
