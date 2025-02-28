// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package metrics

const (
	// Context and authentication errors
	ErrorContextRequestID      = "error_request_id"
	ErrorContextUserEmail      = "error_user_email"
	ErrorContextAuthentication = "error_authentication"

	// Request processing errors
	ErrorProcessingRequestCopy  = "error_request_copy"
	ErrorProcessingResponseCopy = "error_response_copy"

	// Communication errors
	ErrorDispatcherCall = "error_dispatcher_call"
	ErrorStreamReceive  = "error_stream_receive"
	ErrorStreamSend     = "error_stream_send"

	// Usage and billing errors
	ErrorUsageRecord = "error_usage_record"

	// Generic errors
	ErrorInternal = "error_internal"
	ErrorTimeout  = "error_timeout"
)
