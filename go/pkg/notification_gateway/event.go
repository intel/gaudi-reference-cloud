// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import "time"

const (
	EventType_NOTIFICATION = "notification"
	EventType_ALERT        = "alert"
	EventType_EMAIL        = "email"
	EventType_RESOURCE     = "resource"
	EventType_OPERATION    = "operation"
	EventType_AUDIT        = "audit"
	EventType_ERROR        = "error"
)

const (
	EventSeverity_LOW    = "low"
	EventSeverity_MEDIUM = "medium"
	EventSeverity_HIGH   = "high"
)

const (
	EventStatus_ACTIVE   = "active"
	EventStatus_INACTIVE = "inactive"
)

const (
	EventReceiver_PORTAL  = "portal"
	EventReceiver_SMTP    = "smtp"
	EventReceiver_SQS     = "sqs"
	EventReceiver_SES     = "ses"
	EventReceiver_DEFAULT = "default"
)

const (
	EventName_CloudCreditsAvailableEvent        = "CloudCreditsAvailable"
	EventName_CloudCreditsUsedEvent             = "CloudCreditsUsed"
	EventName_CloudCreditsExpiredEvent          = "CloudCreditsExpired"
	EventName_CloudCreditsThresholdReachedEvent = "CloudCreditsThresholdReached"
)

type Event interface {
	GetEventName() string
	GetEventBase() EventBase
	GetUsageIds() []string
	GetCloudCredits() []CloudCredits
}

type EventBase struct {
	Id             string
	Version        string
	CloudAccountId string
	UserId         string
	ServiceName    string
	Type           string
	SubType        string
	Creation       time.Time
	Expiration     time.Time
	Status         string
	Severity       string
	Properties     map[string]string
	ClientRecordId string
}

type CreateEvent struct {
	Status         string
	Type           string
	Severity       string
	ServiceName    string
	Message        string
	CloudAccountId string
	UserId         string
	EventSubType   string
	Properties     map[string]string
	ClientRecordId string
	Region         string
	EventTime      time.Time
}

type Notification struct {
	EventBase
	NotificationType string
	ServiceName      string
	Message          string
}

type Alert struct {
	EventBase
	AlertType   string
	ServiceName string
	Message     string
}

type Email struct {
	EventBase
	UserName     string
	AlertType    string
	ServiceName  string
	Recipient    string
	Sender       string
	TemplateName string
}

type Error struct {
	EventBase
	ErrorType   string
	ServiceName string
	Message     string
	Region      string
}

type CloudCredits struct {
	CouponCode     string
	OriginalAmount float64
	UsedAmount     float64
	Expiry         time.Time
}

type CalculateCloudCreditUsage struct {
	EventBase
	UsageIds []string
}

type CalculateCloudCreditExpiry struct {
	EventBase
	Credits []CloudCredits
}

type RegisterServiceEvents struct {
	ServiceName           string
	EventSubTypeRecievers []*RegisterEventSubTypeReceiver
}

type RegisterEventSubTypeReceiver struct {
	EventSubType  string
	EventReceiver string
}

type EventsSubscribe struct {
	ClientId int32
}
type CloudCreditsEvent struct {
	Name        string
	CreditEvent EventBase
	UsageIds    []string
	Credits     []CloudCredits
	ServiceName string
	Message     string
}

type CloudCreditsUsageEvent struct {
	Name        string
	CreditEvent EventBase
	UsageIds    []string
	Credits     []CloudCredits
}

type CloudCreditsUsedEvent struct {
	Name        string
	CreditEvent EventBase
	UsageIds    []string
	Credits     []CloudCredits
}

type CloudCreditsExpiryEvent struct {
	Name        string
	CreditEvent EventBase
	UsageIds    []string
	Credits     []CloudCredits
}

func (e CloudCreditsUsageEvent) GetEventName() string            { return e.Name }
func (e CloudCreditsUsageEvent) GetEventBase() EventBase         { return e.CreditEvent }
func (e CloudCreditsUsageEvent) GetUsageIds() []string           { return e.UsageIds }
func (e CloudCreditsUsageEvent) GetCloudCredits() []CloudCredits { return e.Credits }

func (e CloudCreditsUsedEvent) GetEventName() string            { return e.Name }
func (e CloudCreditsUsedEvent) GetEventBase() EventBase         { return e.CreditEvent }
func (e CloudCreditsUsedEvent) GetUsageIds() []string           { return e.UsageIds }
func (e CloudCreditsUsedEvent) GetCloudCredits() []CloudCredits { return e.Credits }

func (e CloudCreditsExpiryEvent) GetEventName() string            { return e.Name }
func (e CloudCreditsExpiryEvent) GetEventBase() EventBase         { return e.CreditEvent }
func (e CloudCreditsExpiryEvent) GetUsageIds() []string           { return e.UsageIds }
func (e CloudCreditsExpiryEvent) GetCloudCredits() []CloudCredits { return e.Credits }

func (e CloudCreditsEvent) GetEventName() string    { return e.Name }
func (e CloudCreditsEvent) GetEventBase() EventBase { return e.CreditEvent }
func (e CloudCreditsEvent) GetUsageIds() []string   { return e.UsageIds }

func (e CloudCreditsEvent) GetCloudCredits() []CloudCredits { return e.Credits }
