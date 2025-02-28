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

// VNetServicePutRequest struct for VNetServicePutRequest
type VNetServicePutRequest struct {
	Metadata *VNetServicePutRequestMetadata `json:"metadata,omitempty"`
	Spec     *ProtoVNetSpec                 `json:"spec,omitempty"`
}

// NewVNetServicePutRequest instantiates a new VNetServicePutRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewVNetServicePutRequest() *VNetServicePutRequest {
	this := VNetServicePutRequest{}
	return &this
}

// NewVNetServicePutRequestWithDefaults instantiates a new VNetServicePutRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewVNetServicePutRequestWithDefaults() *VNetServicePutRequest {
	this := VNetServicePutRequest{}
	return &this
}

// GetMetadata returns the Metadata field value if set, zero value otherwise.
func (o *VNetServicePutRequest) GetMetadata() VNetServicePutRequestMetadata {
	if o == nil || isNil(o.Metadata) {
		var ret VNetServicePutRequestMetadata
		return ret
	}
	return *o.Metadata
}

// GetMetadataOk returns a tuple with the Metadata field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VNetServicePutRequest) GetMetadataOk() (*VNetServicePutRequestMetadata, bool) {
	if o == nil || isNil(o.Metadata) {
		return nil, false
	}
	return o.Metadata, true
}

// HasMetadata returns a boolean if a field has been set.
func (o *VNetServicePutRequest) HasMetadata() bool {
	if o != nil && !isNil(o.Metadata) {
		return true
	}

	return false
}

// SetMetadata gets a reference to the given VNetServicePutRequestMetadata and assigns it to the Metadata field.
func (o *VNetServicePutRequest) SetMetadata(v VNetServicePutRequestMetadata) {
	o.Metadata = &v
}

// GetSpec returns the Spec field value if set, zero value otherwise.
func (o *VNetServicePutRequest) GetSpec() ProtoVNetSpec {
	if o == nil || isNil(o.Spec) {
		var ret ProtoVNetSpec
		return ret
	}
	return *o.Spec
}

// GetSpecOk returns a tuple with the Spec field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *VNetServicePutRequest) GetSpecOk() (*ProtoVNetSpec, bool) {
	if o == nil || isNil(o.Spec) {
		return nil, false
	}
	return o.Spec, true
}

// HasSpec returns a boolean if a field has been set.
func (o *VNetServicePutRequest) HasSpec() bool {
	if o != nil && !isNil(o.Spec) {
		return true
	}

	return false
}

// SetSpec gets a reference to the given ProtoVNetSpec and assigns it to the Spec field.
func (o *VNetServicePutRequest) SetSpec(v ProtoVNetSpec) {
	o.Spec = &v
}

func (o VNetServicePutRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Metadata) {
		toSerialize["metadata"] = o.Metadata
	}
	if !isNil(o.Spec) {
		toSerialize["spec"] = o.Spec
	}
	return json.Marshal(toSerialize)
}

type NullableVNetServicePutRequest struct {
	value *VNetServicePutRequest
	isSet bool
}

func (v NullableVNetServicePutRequest) Get() *VNetServicePutRequest {
	return v.value
}

func (v *NullableVNetServicePutRequest) Set(val *VNetServicePutRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableVNetServicePutRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableVNetServicePutRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableVNetServicePutRequest(val *VNetServicePutRequest) *NullableVNetServicePutRequest {
	return &NullableVNetServicePutRequest{value: val, isSet: true}
}

func (v NullableVNetServicePutRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableVNetServicePutRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
