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

// LoadBalancerServiceCreateRequest struct for LoadBalancerServiceCreateRequest
type LoadBalancerServiceCreateRequest struct {
	Metadata *LoadBalancerServiceCreateRequestMetadata `json:"metadata,omitempty"`
	Spec     *ProtoLoadBalancerSpec                    `json:"spec,omitempty"`
}

// NewLoadBalancerServiceCreateRequest instantiates a new LoadBalancerServiceCreateRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLoadBalancerServiceCreateRequest() *LoadBalancerServiceCreateRequest {
	this := LoadBalancerServiceCreateRequest{}
	return &this
}

// NewLoadBalancerServiceCreateRequestWithDefaults instantiates a new LoadBalancerServiceCreateRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLoadBalancerServiceCreateRequestWithDefaults() *LoadBalancerServiceCreateRequest {
	this := LoadBalancerServiceCreateRequest{}
	return &this
}

// GetMetadata returns the Metadata field value if set, zero value otherwise.
func (o *LoadBalancerServiceCreateRequest) GetMetadata() LoadBalancerServiceCreateRequestMetadata {
	if o == nil || isNil(o.Metadata) {
		var ret LoadBalancerServiceCreateRequestMetadata
		return ret
	}
	return *o.Metadata
}

// GetMetadataOk returns a tuple with the Metadata field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LoadBalancerServiceCreateRequest) GetMetadataOk() (*LoadBalancerServiceCreateRequestMetadata, bool) {
	if o == nil || isNil(o.Metadata) {
		return nil, false
	}
	return o.Metadata, true
}

// HasMetadata returns a boolean if a field has been set.
func (o *LoadBalancerServiceCreateRequest) HasMetadata() bool {
	if o != nil && !isNil(o.Metadata) {
		return true
	}

	return false
}

// SetMetadata gets a reference to the given LoadBalancerServiceCreateRequestMetadata and assigns it to the Metadata field.
func (o *LoadBalancerServiceCreateRequest) SetMetadata(v LoadBalancerServiceCreateRequestMetadata) {
	o.Metadata = &v
}

// GetSpec returns the Spec field value if set, zero value otherwise.
func (o *LoadBalancerServiceCreateRequest) GetSpec() ProtoLoadBalancerSpec {
	if o == nil || isNil(o.Spec) {
		var ret ProtoLoadBalancerSpec
		return ret
	}
	return *o.Spec
}

// GetSpecOk returns a tuple with the Spec field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LoadBalancerServiceCreateRequest) GetSpecOk() (*ProtoLoadBalancerSpec, bool) {
	if o == nil || isNil(o.Spec) {
		return nil, false
	}
	return o.Spec, true
}

// HasSpec returns a boolean if a field has been set.
func (o *LoadBalancerServiceCreateRequest) HasSpec() bool {
	if o != nil && !isNil(o.Spec) {
		return true
	}

	return false
}

// SetSpec gets a reference to the given ProtoLoadBalancerSpec and assigns it to the Spec field.
func (o *LoadBalancerServiceCreateRequest) SetSpec(v ProtoLoadBalancerSpec) {
	o.Spec = &v
}

func (o LoadBalancerServiceCreateRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Metadata) {
		toSerialize["metadata"] = o.Metadata
	}
	if !isNil(o.Spec) {
		toSerialize["spec"] = o.Spec
	}
	return json.Marshal(toSerialize)
}

type NullableLoadBalancerServiceCreateRequest struct {
	value *LoadBalancerServiceCreateRequest
	isSet bool
}

func (v NullableLoadBalancerServiceCreateRequest) Get() *LoadBalancerServiceCreateRequest {
	return v.value
}

func (v *NullableLoadBalancerServiceCreateRequest) Set(val *LoadBalancerServiceCreateRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableLoadBalancerServiceCreateRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableLoadBalancerServiceCreateRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLoadBalancerServiceCreateRequest(val *LoadBalancerServiceCreateRequest) *NullableLoadBalancerServiceCreateRequest {
	return &NullableLoadBalancerServiceCreateRequest{value: val, isSet: true}
}

func (v NullableLoadBalancerServiceCreateRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLoadBalancerServiceCreateRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
