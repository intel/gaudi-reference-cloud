// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type NotificationGatewayService struct {
	eventManager *EventManager
	//eventPoll    *EventApiSubscriber
	eventHandler *EventHandler

	pb.UnimplementedNotificationGatewayServiceServer
}

func NewNotificationGatewayService(eventManager *EventManager, eventHandler *EventHandler /**, eventPoll *EventApiSubscriber**/) *NotificationGatewayService {
	return &NotificationGatewayService{eventManager: eventManager, eventHandler: eventHandler /**, eventPoll: eventPoll**/}
}

// // we are not doing this as a micro service and hence will not be needed.
// // only scoped for billing for 1.0
// func (svc *NotificationGatewayService) RegisterService(ctx context.Context, in *pb.RegisterServiceEvents) (*emptypb.Empty, error) {
// 	return nil, status.Error(codes.Unimplemented, "need to implement register service")
// }

// this will be called by aria driver and exposed by billing.
func (svc *NotificationGatewayService) Create(ctx context.Context, in *pb.CreateEvent) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("NotificationGatewayService.Create").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	eventStatus, err := svc.MapStatus(in.Status)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	eventType, err := svc.MapType(in.Type)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid type")
	}

	eventSeverity, err := svc.MapSeverity(*in.Severity)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid severity")
	}

	serviceName, err := svc.MapServiceName(*in.ServiceName)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid severity")
	}

	var eventMessage string
	if in.Message == nil {
		eventMessage = ""
	} else {
		eventMessage = *in.Message
	}

	cloudAcctId := in.GetCloudAccountId()
	var userId string
	if in.UserId == nil {
		userId = ""
	} else {
		userId = *in.UserId
	}

	var region string
	if in.Region == nil {
		region = ""
	} else {
		region = *in.Region
	}

	err = svc.eventManager.Create(ctx, CreateEvent{
		Status:         eventStatus,
		Type:           eventType,
		Severity:       eventSeverity,
		ServiceName:    serviceName,
		Message:        eventMessage,
		CloudAccountId: cloudAcctId,
		UserId:         userId,
		// todo: event sub type needs to match a event sub type provided by the service of name
		// at the time of registration - add check.
		EventSubType: in.EventSubType,
		// todo: support some standardized properties that need to be enforced.
		Properties: in.Properties,
		// todo: this id needs to be bound.
		ClientRecordId: in.ClientRecordId,
		Region:         region,
	})

	if err != nil {
		return &emptypb.Empty{}, status.Error(codes.Internal, "failed to handle event")
	}
	return &emptypb.Empty{}, nil
}

func (svc *NotificationGatewayService) sendEmailNotification(ctx context.Context, email Email) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("EventManager.sendEmailNotification")
	const (
		accessKey    = "/vault/secrets/aws_credentials/access_key_id"
		accessSecret = "/vault/secrets/aws_credentials/secret_access_key"
	)

	// Set up SES client
	sess, err := session.NewSession(&aws.Config{
		// TODO: change to pick it up from values
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(accessKey, accessSecret, ""),
	})

	if err != nil {
		logger.Error(err, "Error creating session:")
		return nil, err
	}

	// Create an SES client using the configuration
	awsSession := ses.New(sess)

	// Specify the email parameters
	input := &ses.SendTemplatedEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{aws.String(email.Recipient)},
		},
		Source:       aws.String(email.Sender),
		Template:     aws.String(email.TemplateName),
		TemplateData: aws.String("{\"name\": email.UserName}"), // Dynamic data for the template
	}

	// Send the email
	result, err := awsSession.SendTemplatedEmail(input)
	if err != nil {
		logger.Error(err, "Error sending email:")
		return nil, err
	}

	logger.Info("Email sent to the user", "messageId", result.MessageId)
	if err != nil {
		return &emptypb.Empty{}, status.Error(codes.Internal, "failed to handle event")
	}
	return &emptypb.Empty{}, nil
}

func (svc *NotificationGatewayService) MapStatus(eventStatus pb.EventStatus) (string, error) {
	switch eventStatus {
	case pb.EventStatus_ACTIVE:
		return EventStatus_ACTIVE, nil
	case pb.EventStatus_INACTIVE:
		return EventStatus_INACTIVE, nil
	default:
		return "", fmt.Errorf("invalid event status %v", eventStatus)
	}
}

func (svc *NotificationGatewayService) MapType(eventType pb.EventType) (string, error) {
	switch eventType {
	case pb.EventType_ALERT:
		return EventType_ALERT, nil
	case pb.EventType_NOTIFICATION:
		return EventType_NOTIFICATION, nil
	case pb.EventType_ERROR:
		return EventType_ERROR, nil
	default:
		return "", fmt.Errorf("invalid event type %v", eventType)
	}
}

func (svc *NotificationGatewayService) MapSeverity(eventSeverity pb.EventSeverity) (string, error) {
	switch eventSeverity {
	case pb.EventSeverity_LOW:
		return EventSeverity_LOW, nil
	case pb.EventSeverity_MEDIUM:
		return EventSeverity_MEDIUM, nil
	default:
		return "", fmt.Errorf("invalid event severity %v", eventSeverity)
	}
}

func (svc *NotificationGatewayService) MapServiceName(serviceName pb.ServiceName) (string, error) {
	switch serviceName {
	case pb.ServiceName_BILLING:
		return "billing", nil
	default:
		return "", fmt.Errorf("invalid service name %v", serviceName)
	}
}

func (svc *NotificationGatewayService) Read(ctx context.Context, in *pb.EventsFilter) (*pb.Events, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("NotificationGatewayService.Read").WithValues("request", in).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	notifications, err := svc.eventManager.getNotificationsForCloudAcct(ctx, in.CloudAccountId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get notifications")
	}

	alerts, err := svc.eventManager.getAlertsForCloudAcct(ctx, in.CloudAccountId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get alerts")
	}

	var notificationsR []*pb.Notification
	var alertsR []*pb.Alert
	notificationCount := 0
	alertCount := 0

	for _, notification := range notifications {
		if notification.Expiration.After(time.Now()) {
			notificationR := svc.MapFromNotification(notification)
			notificationsR = append(notificationsR, notificationR)
			notificationCount++
		}

	}

	for _, alert := range alerts {
		if alert.Expiration.After(time.Now()) {
			alertR := svc.MapFromAlert(alert)
			alertsR = append(alertsR, alertR)
			alertCount++
		}

	}

	return &pb.Events{
		NumberOfNotifications: int32(notificationCount),
		NumberOfAlerts:        int32(alertCount),
		Notifications:         notificationsR,
		Alerts:                alertsR,
	}, nil
}

// to subscribe to long polling
func (svc *NotificationGatewayService) Subscribe(in *pb.EventsSubscribe, outStream pb.NotificationGatewayService_SubscribeServer) error {
	/**ctx := outStream.Context()
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("NotificationGatewayService.Subscribe").WithValues("CloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	subChannel := make(chan bool)
	svc.eventPoll.addSubscriber(in.ClientId, ApiSubscriber{stream: outStream, apiSubscriberChannel: subChannel})

	for {
		select {
		case <-subChannel:
			logger.Info("closing stream for client ID: %d", in.ClientId)
			return nil
		case <-ctx.Done():
			logger.Info("client", "id", in.ClientId, "disconnected")
			return nil
		}
	}**/
	return nil
}

func (svc *NotificationGatewayService) DismissEvent(ctx context.Context, eventFilter *pb.EventsFilter) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("NotificationGatewayService.DismissEvent")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	err := svc.eventManager.dismissEvent(ctx, eventFilter.CloudAccountId, eventFilter.ClientRecordId)
	if err != nil {
		logger.Info("error in dismissing event", "eventFilter", eventFilter)
		return nil, status.Error(codes.Internal, "failed to dismiss event")
	}
	return &emptypb.Empty{}, nil
}

func (svc *NotificationGatewayService) PublishEvent(ctx context.Context, in *pb.PublishEventRequest) (*pb.PublishEventResponse, error) {
	cloudAccountId := in.CreateEvent.GetCloudAccountId()
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("NotificationGatewayService.PublishEvent").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	subject := in.Subject
	logger.Info("publish event api ", "cloudAccountId", cloudAccountId, "topicName", in.TopicName, "subject", subject, "CreateEvent", in.CreateEvent)
	if len(cloudAccountId) == 0 || utils.IsValidCloudAccountId(cloudAccountId) {
		return &pb.PublishEventResponse{}, status.Errorf(codes.InvalidArgument, "invalid input arguments")
	}
	topicName := in.TopicName
	if len(topicName) == 0 {
		return &pb.PublishEventResponse{}, status.Errorf(codes.InvalidArgument, "invalid topic name")
	}
	messageId, err := svc.eventHandler.HandlePublishEvent(ctx, cloudAccountId, in.CreateEvent, topicName)
	if err != nil {
		return &pb.PublishEventResponse{}, status.Error(codes.Internal, "failed to handle event")
	}

	return &pb.PublishEventResponse{MessageId: messageId}, nil
}

func (svc *NotificationGatewayService) SubscribeEvents(ctx context.Context, in *pb.SubscribeEventRequest) (*pb.SubscribeEventResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("NotificationGatewayService.SubscribeEvents").WithValues("QueueName", in.GetQueueName()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	queueName := in.QueueName
	if len(queueName) == 0 {
		return &pb.SubscribeEventResponse{}, status.Errorf(codes.InvalidArgument, "invalid input queue name")
	}
	topicName := in.TopicName
	if len(topicName) == 0 {
		return &pb.SubscribeEventResponse{}, status.Errorf(codes.InvalidArgument, "invalid input topic name")
	}
	subscriptionArn, err := svc.eventHandler.HandleSubscribeEvents(ctx, queueName, topicName)
	if err != nil {
		return &pb.SubscribeEventResponse{}, status.Error(codes.Internal, "failed to subscribe for event")
	}
	return &pb.SubscribeEventResponse{SubscriptionArn: subscriptionArn}, nil
}

func (svc *NotificationGatewayService) ReceiveEvents(ctx context.Context, in *pb.ReceiveEventRequest) (*pb.ReceiveEventResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("NotificationGatewayService.ReceiveEvents").WithValues("QueueName", in.GetQueueName()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	queueName := in.QueueName
	if len(queueName) == 0 {
		return &pb.ReceiveEventResponse{}, status.Errorf(codes.InvalidArgument, "invalid input queue name")
	}
	eventName := in.EventName
	if len(eventName) == 0 {
		return &pb.ReceiveEventResponse{}, status.Errorf(codes.InvalidArgument, "invalid event name")
	}
	messageResponse, err := svc.eventHandler.HandleReceiveEvents(ctx, queueName, int32(in.WaitTimeSeconds), int32(in.MaxNumberOfMessages), eventName)
	if err != nil {
		return &pb.ReceiveEventResponse{}, status.Error(codes.Internal, "failed to receive event")
	}

	return &pb.ReceiveEventResponse{MessageResponse: messageResponse}, nil
}

func (svc *NotificationGatewayService) DeleteEvents(ctx context.Context, in *pb.DeleteEventRequestList) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("NotificationGatewayService.DeleteEvents").WithValues("QueueName", in.GetQueueName()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	queueName := in.GetQueueName()
	if len(queueName) == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.InvalidArgument, "invalid input queue name")
	}
	deleteEventRequests := in.GetDeleteEventRequests()
	if len(deleteEventRequests) == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.InvalidArgument, "invalid delete event request")
	}
	err := svc.eventHandler.HandleDeleteEvents(ctx, queueName, deleteEventRequests)
	if err != nil {
		return &emptypb.Empty{}, status.Error(codes.Internal, "failed to delete events")
	}

	return &emptypb.Empty{}, nil
}
