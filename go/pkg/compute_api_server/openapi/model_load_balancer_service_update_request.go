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

// LoadBalancerServiceUpdateRequest struct for LoadBalancerServiceUpdateRequest
type LoadBalancerServiceUpdateRequest struct {
	Metadata *LoadBalancerServiceUpdateRequestMetadata `json:"metadata,omitempty"`
	Spec     *ProtoLoadBalancerSpecUpdate              `json:"spec,omitempty"`
}

// NewLoadBalancerServiceUpdateRequest instantiates a new LoadBalancerServiceUpdateRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLoadBalancerServiceUpdateRequest() *LoadBalancerServiceUpdateRequest {
	this := LoadBalancerServiceUpdateRequest{}
	return &this
}

// NewLoadBalancerServiceUpdateRequestWithDefaults instantiates a new LoadBalancerServiceUpdateRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLoadBalancerServiceUpdateRequestWithDefaults() *LoadBalancerServiceUpdateRequest {
	this := LoadBalancerServiceUpdateRequest{}
	return &this
}

// GetMetadata returns the Metadata field value if set, zero value otherwise.
func (o *LoadBalancerServiceUpdateRequest) GetMetadata() LoadBalancerServiceUpdateRequestMetadata {
	if o == nil || isNil(o.Metadata) {
		var ret LoadBalancerServiceUpdateRequestMetadata
		return ret
	}
	return *o.Metadata
}

// GetMetadataOk returns a tuple with the Metadata field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LoadBalancerServiceUpdateRequest) GetMetadataOk() (*LoadBalancerServiceUpdateRequestMetadata, bool) {
	if o == nil || isNil(o.Metadata) {
		return nil, false
	}
	return o.Metadata, true
}

// HasMetadata returns a boolean if a field has been set.
func (o *LoadBalancerServiceUpdateRequest) HasMetadata() bool {
	if o != nil && !isNil(o.Metadata) {
		return true
	}

	return false
}

// SetMetadata gets a reference to the given LoadBalancerServiceUpdateRequestMetadata and assigns it to the Metadata field.
func (o *LoadBalancerServiceUpdateRequest) SetMetadata(v LoadBalancerServiceUpdateRequestMetadata) {
	o.Metadata = &v
}

// GetSpec returns the Spec field value if set, zero value otherwise.
func (o *LoadBalancerServiceUpdateRequest) GetSpec() ProtoLoadBalancerSpecUpdate {
	if o == nil || isNil(o.Spec) {
		var ret ProtoLoadBalancerSpecUpdate
		return ret
	}
	return *o.Spec
}

// GetSpecOk returns a tuple with the Spec field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LoadBalancerServiceUpdateRequest) GetSpecOk() (*ProtoLoadBalancerSpecUpdate, bool) {
	if o == nil || isNil(o.Spec) {
		return nil, false
	}
	return o.Spec, true
}

// HasSpec returns a boolean if a field has been set.
func (o *LoadBalancerServiceUpdateRequest) HasSpec() bool {
	if o != nil && !isNil(o.Spec) {
		return true
	}

	return false
}

// SetSpec gets a reference to the given ProtoLoadBalancerSpecUpdate and assigns it to the Spec field.
func (o *LoadBalancerServiceUpdateRequest) SetSpec(v ProtoLoadBalancerSpecUpdate) {
	o.Spec = &v
}

func (o LoadBalancerServiceUpdateRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Metadata) {
		toSerialize["metadata"] = o.Metadata
	}
	if !isNil(o.Spec) {
		toSerialize["spec"] = o.Spec
	}
	return json.Marshal(toSerialize)
}

type NullableLoadBalancerServiceUpdateRequest struct {
	value *LoadBalancerServiceUpdateRequest
	isSet bool
}

func (v NullableLoadBalancerServiceUpdateRequest) Get() *LoadBalancerServiceUpdateRequest {
	return v.value
}

func (v *NullableLoadBalancerServiceUpdateRequest) Set(val *LoadBalancerServiceUpdateRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableLoadBalancerServiceUpdateRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableLoadBalancerServiceUpdateRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLoadBalancerServiceUpdateRequest(val *LoadBalancerServiceUpdateRequest) *NullableLoadBalancerServiceUpdateRequest {
	return &NullableLoadBalancerServiceUpdateRequest{value: val, isSet: true}
}

func (v NullableLoadBalancerServiceUpdateRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLoadBalancerServiceUpdateRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
