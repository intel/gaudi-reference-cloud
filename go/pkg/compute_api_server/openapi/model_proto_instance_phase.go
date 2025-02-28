/*
compute.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
	"fmt"
)

// ProtoInstancePhase  - Provisioning: The system is creating and starting the instance. Default.  - Ready: The instance is running and has completed the running startup process.  - Stopping: The instance is in the process of being stopped.  - Stopped: The instance is stopped.  - Terminating: The instance and its associated resources are in the process of being deleted.  - Failed: The instance crashed, failed, or is otherwise unavailable.  - Starting: The instance is in the process of startup.  - Started: The instance has completed startup and is available to use.
type ProtoInstancePhase string

// List of protoInstancePhase
const (
	PROVISIONING ProtoInstancePhase = "Provisioning"
	READY        ProtoInstancePhase = "Ready"
	STOPPING     ProtoInstancePhase = "Stopping"
	STOPPED      ProtoInstancePhase = "Stopped"
	TERMINATING  ProtoInstancePhase = "Terminating"
	FAILED       ProtoInstancePhase = "Failed"
	STARTING     ProtoInstancePhase = "Starting"
	STARTED      ProtoInstancePhase = "Started"
)

// All allowed values of ProtoInstancePhase enum
var AllowedProtoInstancePhaseEnumValues = []ProtoInstancePhase{
	"Provisioning",
	"Ready",
	"Stopping",
	"Stopped",
	"Terminating",
	"Failed",
	"Starting",
	"Started",
}

func (v *ProtoInstancePhase) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := ProtoInstancePhase(value)
	for _, existing := range AllowedProtoInstancePhaseEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid ProtoInstancePhase", value)
}

// NewProtoInstancePhaseFromValue returns a pointer to a valid ProtoInstancePhase
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewProtoInstancePhaseFromValue(v string) (*ProtoInstancePhase, error) {
	ev := ProtoInstancePhase(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for ProtoInstancePhase: valid values are %v", v, AllowedProtoInstancePhaseEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v ProtoInstancePhase) IsValid() bool {
	for _, existing := range AllowedProtoInstancePhaseEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to protoInstancePhase value
func (v ProtoInstancePhase) Ptr() *ProtoInstancePhase {
	return &v
}

type NullableProtoInstancePhase struct {
	value *ProtoInstancePhase
	isSet bool
}

func (v NullableProtoInstancePhase) Get() *ProtoInstancePhase {
	return v.value
}

func (v *NullableProtoInstancePhase) Set(val *ProtoInstancePhase) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoInstancePhase) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoInstancePhase) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoInstancePhase(val *ProtoInstancePhase) *NullableProtoInstancePhase {
	return &NullableProtoInstancePhase{value: val, isSet: true}
}

func (v NullableProtoInstancePhase) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoInstancePhase) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
