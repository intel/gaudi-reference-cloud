/*
compute.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// ProtoInstanceGroupStatus struct for ProtoInstanceGroupStatus
type ProtoInstanceGroupStatus struct {
	// The number of instances with a phase of Ready. The instance group is Ready when this equals InstanceGroupSpec.instanceCount.
	ReadyCount *int32 `json:"readyCount,omitempty"`
}

// NewProtoInstanceGroupStatus instantiates a new ProtoInstanceGroupStatus object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoInstanceGroupStatus() *ProtoInstanceGroupStatus {
	this := ProtoInstanceGroupStatus{}
	return &this
}

// NewProtoInstanceGroupStatusWithDefaults instantiates a new ProtoInstanceGroupStatus object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoInstanceGroupStatusWithDefaults() *ProtoInstanceGroupStatus {
	this := ProtoInstanceGroupStatus{}
	return &this
}

// GetReadyCount returns the ReadyCount field value if set, zero value otherwise.
func (o *ProtoInstanceGroupStatus) GetReadyCount() int32 {
	if o == nil || isNil(o.ReadyCount) {
		var ret int32
		return ret
	}
	return *o.ReadyCount
}

// GetReadyCountOk returns a tuple with the ReadyCount field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceGroupStatus) GetReadyCountOk() (*int32, bool) {
	if o == nil || isNil(o.ReadyCount) {
		return nil, false
	}
	return o.ReadyCount, true
}

// HasReadyCount returns a boolean if a field has been set.
func (o *ProtoInstanceGroupStatus) HasReadyCount() bool {
	if o != nil && !isNil(o.ReadyCount) {
		return true
	}

	return false
}

// SetReadyCount gets a reference to the given int32 and assigns it to the ReadyCount field.
func (o *ProtoInstanceGroupStatus) SetReadyCount(v int32) {
	o.ReadyCount = &v
}

func (o ProtoInstanceGroupStatus) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.ReadyCount) {
		toSerialize["readyCount"] = o.ReadyCount
	}
	return json.Marshal(toSerialize)
}

type NullableProtoInstanceGroupStatus struct {
	value *ProtoInstanceGroupStatus
	isSet bool
}

func (v NullableProtoInstanceGroupStatus) Get() *ProtoInstanceGroupStatus {
	return v.value
}

func (v *NullableProtoInstanceGroupStatus) Set(val *ProtoInstanceGroupStatus) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoInstanceGroupStatus) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoInstanceGroupStatus) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoInstanceGroupStatus(val *ProtoInstanceGroupStatus) *NullableProtoInstanceGroupStatus {
	return &NullableProtoInstanceGroupStatus{value: val, isSet: true}
}

func (v NullableProtoInstanceGroupStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoInstanceGroupStatus) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
