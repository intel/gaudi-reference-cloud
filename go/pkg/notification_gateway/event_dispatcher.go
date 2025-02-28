// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"
	"sync"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

const fixNumberOfRetrials = 1

type ServiceEventDispatcher struct {
	serviceName string
	// when we support plugins - this will be a map of event type to grpc connections.
	// until then, it is in the same code base and hence we only store the kind of receiver.
	serviceEventTypeDispatcher map[string]string
}

var serviceEventTypeMap = sync.Map{}

type EventDispatcher struct {
	//eventApiSubscriber *EventApiSubscriber
}

func NewEventDispatcher( /**eventPoll *EventApiSubscriber**/ ) *EventDispatcher {
	return &EventDispatcher{}
}

func (ed *EventDispatcher) dispatchNotification(ctx context.Context, notification Notification) error {
	logger := log.FromContext(ctx).WithName("EventDispatcher.dispatchNotification")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	//ed.eventApiSubscriber.sendNotification(ctx, notification)
	return nil
}

func (ed *EventDispatcher) retryDispatchNotification(ctx context.Context, notification Notification) error {
	logger := log.FromContext(ctx).WithName("EventDispatcher.retryDispatchNotification")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	for i := 1; i <= fixNumberOfRetrials; i++ {
		err := ed.dispatchNotification(ctx, notification)
		if err == nil {
			break
		}
	}
	return nil
}

func (ed *EventDispatcher) dispatchAlert(ctx context.Context, alert Alert) error {
	logger := log.FromContext(ctx).WithName("EventDispatcher.dispatchAlert")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	//ed.eventApiSubscriber.sendAlert(ctx, alert)
	return nil
}

func (ed *EventDispatcher) dispatchError(ctx context.Context, error Error) error {
	logger := log.FromContext(ctx).WithName("EventDispatcher.dispatchError")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	//ed.eventApiSubscriber.sendAlert(ctx, error)
	return nil
}

func (ed *EventDispatcher) dispatchEmail(ctx context.Context, email Email) error {
	logger := log.FromContext(ctx).WithName("EventDispatcher.dispatchEmail")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	// ed.eventApiSubscriber.sendAlert(ctx, alert)
	return nil
}

func (ed *EventDispatcher) retryDispatchAlert(ctx context.Context, alert Alert) error {
	logger := log.FromContext(ctx).WithName("EventDispatcher.retryDispatchAlert")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	for i := 1; i <= fixNumberOfRetrials; i++ {
		err := ed.dispatchAlert(ctx, alert)
		if err == nil {
			break
		}
	}
	return nil
}

func (ed *EventDispatcher) addDispatcher(ctx context.Context, serviceName string, eventSubTypeReceiverRegistration []*RegisterEventSubTypeReceiver) {

}
