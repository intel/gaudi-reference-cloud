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

// StreamResultOfProtoVNet struct for StreamResultOfProtoVNet
type StreamResultOfProtoVNet struct {
	Result *ProtoVNet `json:"result,omitempty"`
	Error  *RpcStatus `json:"error,omitempty"`
}

// NewStreamResultOfProtoVNet instantiates a new StreamResultOfProtoVNet object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewStreamResultOfProtoVNet() *StreamResultOfProtoVNet {
	this := StreamResultOfProtoVNet{}
	return &this
}

// NewStreamResultOfProtoVNetWithDefaults instantiates a new StreamResultOfProtoVNet object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewStreamResultOfProtoVNetWithDefaults() *StreamResultOfProtoVNet {
	this := StreamResultOfProtoVNet{}
	return &this
}

// GetResult returns the Result field value if set, zero value otherwise.
func (o *StreamResultOfProtoVNet) GetResult() ProtoVNet {
	if o == nil || isNil(o.Result) {
		var ret ProtoVNet
		return ret
	}
	return *o.Result
}

// GetResultOk returns a tuple with the Result field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *StreamResultOfProtoVNet) GetResultOk() (*ProtoVNet, bool) {
	if o == nil || isNil(o.Result) {
		return nil, false
	}
	return o.Result, true
}

// HasResult returns a boolean if a field has been set.
func (o *StreamResultOfProtoVNet) HasResult() bool {
	if o != nil && !isNil(o.Result) {
		return true
	}

	return false
}

// SetResult gets a reference to the given ProtoVNet and assigns it to the Result field.
func (o *StreamResultOfProtoVNet) SetResult(v ProtoVNet) {
	o.Result = &v
}

// GetError returns the Error field value if set, zero value otherwise.
func (o *StreamResultOfProtoVNet) GetError() RpcStatus {
	if o == nil || isNil(o.Error) {
		var ret RpcStatus
		return ret
	}
	return *o.Error
}

// GetErrorOk returns a tuple with the Error field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *StreamResultOfProtoVNet) GetErrorOk() (*RpcStatus, bool) {
	if o == nil || isNil(o.Error) {
		return nil, false
	}
	return o.Error, true
}

// HasError returns a boolean if a field has been set.
func (o *StreamResultOfProtoVNet) HasError() bool {
	if o != nil && !isNil(o.Error) {
		return true
	}

	return false
}

// SetError gets a reference to the given RpcStatus and assigns it to the Error field.
func (o *StreamResultOfProtoVNet) SetError(v RpcStatus) {
	o.Error = &v
}

func (o StreamResultOfProtoVNet) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Result) {
		toSerialize["result"] = o.Result
	}
	if !isNil(o.Error) {
		toSerialize["error"] = o.Error
	}
	return json.Marshal(toSerialize)
}

type NullableStreamResultOfProtoVNet struct {
	value *StreamResultOfProtoVNet
	isSet bool
}

func (v NullableStreamResultOfProtoVNet) Get() *StreamResultOfProtoVNet {
	return v.value
}

func (v *NullableStreamResultOfProtoVNet) Set(val *StreamResultOfProtoVNet) {
	v.value = val
	v.isSet = true
}

func (v NullableStreamResultOfProtoVNet) IsSet() bool {
	return v.isSet
}

func (v *NullableStreamResultOfProtoVNet) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableStreamResultOfProtoVNet(val *StreamResultOfProtoVNet) *NullableStreamResultOfProtoVNet {
	return &NullableStreamResultOfProtoVNet{value: val, isSet: true}
}

func (v NullableStreamResultOfProtoVNet) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableStreamResultOfProtoVNet) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
