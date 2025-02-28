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

// ProtoLoadBalancerListener struct for ProtoLoadBalancerListener
type ProtoLoadBalancerListener struct {
	// The public port of the load balancer.
	Port *int32                 `json:"port,omitempty"`
	Pool *ProtoLoadBalancerPool `json:"pool,omitempty"`
}

// NewProtoLoadBalancerListener instantiates a new ProtoLoadBalancerListener object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoLoadBalancerListener() *ProtoLoadBalancerListener {
	this := ProtoLoadBalancerListener{}
	return &this
}

// NewProtoLoadBalancerListenerWithDefaults instantiates a new ProtoLoadBalancerListener object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoLoadBalancerListenerWithDefaults() *ProtoLoadBalancerListener {
	this := ProtoLoadBalancerListener{}
	return &this
}

// GetPort returns the Port field value if set, zero value otherwise.
func (o *ProtoLoadBalancerListener) GetPort() int32 {
	if o == nil || isNil(o.Port) {
		var ret int32
		return ret
	}
	return *o.Port
}

// GetPortOk returns a tuple with the Port field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerListener) GetPortOk() (*int32, bool) {
	if o == nil || isNil(o.Port) {
		return nil, false
	}
	return o.Port, true
}

// HasPort returns a boolean if a field has been set.
func (o *ProtoLoadBalancerListener) HasPort() bool {
	if o != nil && !isNil(o.Port) {
		return true
	}

	return false
}

// SetPort gets a reference to the given int32 and assigns it to the Port field.
func (o *ProtoLoadBalancerListener) SetPort(v int32) {
	o.Port = &v
}

// GetPool returns the Pool field value if set, zero value otherwise.
func (o *ProtoLoadBalancerListener) GetPool() ProtoLoadBalancerPool {
	if o == nil || isNil(o.Pool) {
		var ret ProtoLoadBalancerPool
		return ret
	}
	return *o.Pool
}

// GetPoolOk returns a tuple with the Pool field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerListener) GetPoolOk() (*ProtoLoadBalancerPool, bool) {
	if o == nil || isNil(o.Pool) {
		return nil, false
	}
	return o.Pool, true
}

// HasPool returns a boolean if a field has been set.
func (o *ProtoLoadBalancerListener) HasPool() bool {
	if o != nil && !isNil(o.Pool) {
		return true
	}

	return false
}

// SetPool gets a reference to the given ProtoLoadBalancerPool and assigns it to the Pool field.
func (o *ProtoLoadBalancerListener) SetPool(v ProtoLoadBalancerPool) {
	o.Pool = &v
}

func (o ProtoLoadBalancerListener) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Port) {
		toSerialize["port"] = o.Port
	}
	if !isNil(o.Pool) {
		toSerialize["pool"] = o.Pool
	}
	return json.Marshal(toSerialize)
}

type NullableProtoLoadBalancerListener struct {
	value *ProtoLoadBalancerListener
	isSet bool
}

func (v NullableProtoLoadBalancerListener) Get() *ProtoLoadBalancerListener {
	return v.value
}

func (v *NullableProtoLoadBalancerListener) Set(val *ProtoLoadBalancerListener) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoLoadBalancerListener) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoLoadBalancerListener) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoLoadBalancerListener(val *ProtoLoadBalancerListener) *NullableProtoLoadBalancerListener {
	return &NullableProtoLoadBalancerListener{value: val, isSet: true}
}

func (v NullableProtoLoadBalancerListener) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoLoadBalancerListener) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
