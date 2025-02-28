// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: fleet_common.proto

package pb

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on GroupVersionResource with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GroupVersionResource) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GroupVersionResource with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GroupVersionResourceMultiError, or nil if none found.
func (m *GroupVersionResource) ValidateAll() error {
	return m.validate(true)
}

func (m *GroupVersionResource) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Group

	// no validation rules for Version

	// no validation rules for Resource

	if len(errors) > 0 {
		return GroupVersionResourceMultiError(errors)
	}

	return nil
}

// GroupVersionResourceMultiError is an error wrapping multiple validation
// errors returned by GroupVersionResource.ValidateAll() if the designated
// constraints aren't met.
type GroupVersionResourceMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GroupVersionResourceMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GroupVersionResourceMultiError) AllErrors() []error { return m }

// GroupVersionResourceValidationError is the validation error returned by
// GroupVersionResource.Validate if the designated constraints aren't met.
type GroupVersionResourceValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GroupVersionResourceValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GroupVersionResourceValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GroupVersionResourceValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GroupVersionResourceValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GroupVersionResourceValidationError) ErrorName() string {
	return "GroupVersionResourceValidationError"
}

// Error satisfies the builtin error interface
func (e GroupVersionResourceValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGroupVersionResource.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GroupVersionResourceValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GroupVersionResourceValidationError{}

// Validate checks the field values on SchedulerStatistics with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *SchedulerStatistics) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on SchedulerStatistics with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// SchedulerStatisticsMultiError, or nil if none found.
func (m *SchedulerStatistics) ValidateAll() error {
	return m.validate(true)
}

func (m *SchedulerStatistics) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetSchedulerNodeStatistics() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, SchedulerStatisticsValidationError{
						field:  fmt.Sprintf("SchedulerNodeStatistics[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, SchedulerStatisticsValidationError{
						field:  fmt.Sprintf("SchedulerNodeStatistics[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return SchedulerStatisticsValidationError{
					field:  fmt.Sprintf("SchedulerNodeStatistics[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return SchedulerStatisticsMultiError(errors)
	}

	return nil
}

// SchedulerStatisticsMultiError is an error wrapping multiple validation
// errors returned by SchedulerStatistics.ValidateAll() if the designated
// constraints aren't met.
type SchedulerStatisticsMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m SchedulerStatisticsMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m SchedulerStatisticsMultiError) AllErrors() []error { return m }

// SchedulerStatisticsValidationError is the validation error returned by
// SchedulerStatistics.Validate if the designated constraints aren't met.
type SchedulerStatisticsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SchedulerStatisticsValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SchedulerStatisticsValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SchedulerStatisticsValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SchedulerStatisticsValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SchedulerStatisticsValidationError) ErrorName() string {
	return "SchedulerStatisticsValidationError"
}

// Error satisfies the builtin error interface
func (e SchedulerStatisticsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSchedulerStatistics.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SchedulerStatisticsValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SchedulerStatisticsValidationError{}

// Validate checks the field values on SchedulerNodeStatistics with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *SchedulerNodeStatistics) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on SchedulerNodeStatistics with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// SchedulerNodeStatisticsMultiError, or nil if none found.
func (m *SchedulerNodeStatistics) ValidateAll() error {
	return m.validate(true)
}

func (m *SchedulerNodeStatistics) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetSchedulerNode()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, SchedulerNodeStatisticsValidationError{
					field:  "SchedulerNode",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, SchedulerNodeStatisticsValidationError{
					field:  "SchedulerNode",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetSchedulerNode()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return SchedulerNodeStatisticsValidationError{
				field:  "SchedulerNode",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	for idx, item := range m.GetInstanceTypeStatistics() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, SchedulerNodeStatisticsValidationError{
						field:  fmt.Sprintf("InstanceTypeStatistics[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, SchedulerNodeStatisticsValidationError{
						field:  fmt.Sprintf("InstanceTypeStatistics[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return SchedulerNodeStatisticsValidationError{
					field:  fmt.Sprintf("InstanceTypeStatistics[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if all {
		switch v := interface{}(m.GetNodeResources()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, SchedulerNodeStatisticsValidationError{
					field:  "NodeResources",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, SchedulerNodeStatisticsValidationError{
					field:  "NodeResources",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetNodeResources()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return SchedulerNodeStatisticsValidationError{
				field:  "NodeResources",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return SchedulerNodeStatisticsMultiError(errors)
	}

	return nil
}

// SchedulerNodeStatisticsMultiError is an error wrapping multiple validation
// errors returned by SchedulerNodeStatistics.ValidateAll() if the designated
// constraints aren't met.
type SchedulerNodeStatisticsMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m SchedulerNodeStatisticsMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m SchedulerNodeStatisticsMultiError) AllErrors() []error { return m }

// SchedulerNodeStatisticsValidationError is the validation error returned by
// SchedulerNodeStatistics.Validate if the designated constraints aren't met.
type SchedulerNodeStatisticsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SchedulerNodeStatisticsValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SchedulerNodeStatisticsValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SchedulerNodeStatisticsValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SchedulerNodeStatisticsValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SchedulerNodeStatisticsValidationError) ErrorName() string {
	return "SchedulerNodeStatisticsValidationError"
}

// Error satisfies the builtin error interface
func (e SchedulerNodeStatisticsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSchedulerNodeStatistics.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SchedulerNodeStatisticsValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SchedulerNodeStatisticsValidationError{}

// Validate checks the field values on SchedulerNode with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *SchedulerNode) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on SchedulerNode with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in SchedulerNodeMultiError, or
// nil if none found.
func (m *SchedulerNode) ValidateAll() error {
	return m.validate(true)
}

func (m *SchedulerNode) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Region

	// no validation rules for AvailabilityZone

	// no validation rules for NodeName

	// no validation rules for ClusterId

	// no validation rules for Namespace

	// no validation rules for Partition

	// no validation rules for ClusterGroup

	// no validation rules for NetworkMode

	if all {
		switch v := interface{}(m.GetSourceGvr()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, SchedulerNodeValidationError{
					field:  "SourceGvr",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, SchedulerNodeValidationError{
					field:  "SourceGvr",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetSourceGvr()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return SchedulerNodeValidationError{
				field:  "SourceGvr",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return SchedulerNodeMultiError(errors)
	}

	return nil
}

// SchedulerNodeMultiError is an error wrapping multiple validation errors
// returned by SchedulerNode.ValidateAll() if the designated constraints
// aren't met.
type SchedulerNodeMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m SchedulerNodeMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m SchedulerNodeMultiError) AllErrors() []error { return m }

// SchedulerNodeValidationError is the validation error returned by
// SchedulerNode.Validate if the designated constraints aren't met.
type SchedulerNodeValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SchedulerNodeValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SchedulerNodeValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SchedulerNodeValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SchedulerNodeValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SchedulerNodeValidationError) ErrorName() string { return "SchedulerNodeValidationError" }

// Error satisfies the builtin error interface
func (e SchedulerNodeValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSchedulerNode.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SchedulerNodeValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SchedulerNodeValidationError{}

// Validate checks the field values on InstanceTypeStatistics with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *InstanceTypeStatistics) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on InstanceTypeStatistics with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// InstanceTypeStatisticsMultiError, or nil if none found.
func (m *InstanceTypeStatistics) ValidateAll() error {
	return m.validate(true)
}

func (m *InstanceTypeStatistics) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for InstanceType

	// no validation rules for RunningInstances

	// no validation rules for MaxNewInstances

	// no validation rules for InstanceCategory

	if len(errors) > 0 {
		return InstanceTypeStatisticsMultiError(errors)
	}

	return nil
}

// InstanceTypeStatisticsMultiError is an error wrapping multiple validation
// errors returned by InstanceTypeStatistics.ValidateAll() if the designated
// constraints aren't met.
type InstanceTypeStatisticsMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m InstanceTypeStatisticsMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m InstanceTypeStatisticsMultiError) AllErrors() []error { return m }

// InstanceTypeStatisticsValidationError is the validation error returned by
// InstanceTypeStatistics.Validate if the designated constraints aren't met.
type InstanceTypeStatisticsValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e InstanceTypeStatisticsValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e InstanceTypeStatisticsValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e InstanceTypeStatisticsValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e InstanceTypeStatisticsValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e InstanceTypeStatisticsValidationError) ErrorName() string {
	return "InstanceTypeStatisticsValidationError"
}

// Error satisfies the builtin error interface
func (e InstanceTypeStatisticsValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sInstanceTypeStatistics.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = InstanceTypeStatisticsValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = InstanceTypeStatisticsValidationError{}

// Validate checks the field values on NodeResources with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *NodeResources) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on NodeResources with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in NodeResourcesMultiError, or
// nil if none found.
func (m *NodeResources) ValidateAll() error {
	return m.validate(true)
}

func (m *NodeResources) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for FreeMilliCPU

	// no validation rules for UsedMilliCPU

	// no validation rules for FreeMemoryBytes

	// no validation rules for UsedMemoryBytes

	// no validation rules for FreeGPU

	// no validation rules for UsedGPU

	if len(errors) > 0 {
		return NodeResourcesMultiError(errors)
	}

	return nil
}

// NodeResourcesMultiError is an error wrapping multiple validation errors
// returned by NodeResources.ValidateAll() if the designated constraints
// aren't met.
type NodeResourcesMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m NodeResourcesMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m NodeResourcesMultiError) AllErrors() []error { return m }

// NodeResourcesValidationError is the validation error returned by
// NodeResources.Validate if the designated constraints aren't met.
type NodeResourcesValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e NodeResourcesValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e NodeResourcesValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e NodeResourcesValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e NodeResourcesValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e NodeResourcesValidationError) ErrorName() string { return "NodeResourcesValidationError" }

// Error satisfies the builtin error interface
func (e NodeResourcesValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sNodeResources.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = NodeResourcesValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = NodeResourcesValidationError{}
