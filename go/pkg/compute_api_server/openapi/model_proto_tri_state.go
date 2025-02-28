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

// ProtoTriState the model 'ProtoTriState'
type ProtoTriState string

// List of protoTriState
const (
	UNDEFINED ProtoTriState = "Undefined"
	TRUE      ProtoTriState = "True"
	FALSE     ProtoTriState = "False"
)

// All allowed values of ProtoTriState enum
var AllowedProtoTriStateEnumValues = []ProtoTriState{
	"Undefined",
	"True",
	"False",
}

func (v *ProtoTriState) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := ProtoTriState(value)
	for _, existing := range AllowedProtoTriStateEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid ProtoTriState", value)
}

// NewProtoTriStateFromValue returns a pointer to a valid ProtoTriState
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewProtoTriStateFromValue(v string) (*ProtoTriState, error) {
	ev := ProtoTriState(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for ProtoTriState: valid values are %v", v, AllowedProtoTriStateEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v ProtoTriState) IsValid() bool {
	for _, existing := range AllowedProtoTriStateEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to protoTriState value
func (v ProtoTriState) Ptr() *ProtoTriState {
	return &v
}

type NullableProtoTriState struct {
	value *ProtoTriState
	isSet bool
}

func (v NullableProtoTriState) Get() *ProtoTriState {
	return v.value
}

func (v *NullableProtoTriState) Set(val *ProtoTriState) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoTriState) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoTriState) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoTriState(val *ProtoTriState) *NullableProtoTriState {
	return &NullableProtoTriState{value: val, isSet: true}
}

func (v NullableProtoTriState) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoTriState) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
