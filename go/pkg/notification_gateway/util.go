// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (svc *NotificationGatewayService) MapFromNotification(notification *Notification) *pb.Notification {
	return &pb.Notification{
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
}

func (svc *NotificationGatewayService) MapFromAlert(alert *Alert) *pb.Alert {
	return &pb.Alert{
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
}
