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

// ProtoVNetSearchRequestMetadata struct for ProtoVNetSearchRequestMetadata
type ProtoVNetSearchRequestMetadata struct {
	CloudAccountId *string `json:"cloudAccountId,omitempty"`
}

// NewProtoVNetSearchRequestMetadata instantiates a new ProtoVNetSearchRequestMetadata object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoVNetSearchRequestMetadata() *ProtoVNetSearchRequestMetadata {
	this := ProtoVNetSearchRequestMetadata{}
	return &this
}

// NewProtoVNetSearchRequestMetadataWithDefaults instantiates a new ProtoVNetSearchRequestMetadata object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoVNetSearchRequestMetadataWithDefaults() *ProtoVNetSearchRequestMetadata {
	this := ProtoVNetSearchRequestMetadata{}
	return &this
}

// GetCloudAccountId returns the CloudAccountId field value if set, zero value otherwise.
func (o *ProtoVNetSearchRequestMetadata) GetCloudAccountId() string {
	if o == nil || isNil(o.CloudAccountId) {
		var ret string
		return ret
	}
	return *o.CloudAccountId
}

// GetCloudAccountIdOk returns a tuple with the CloudAccountId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoVNetSearchRequestMetadata) GetCloudAccountIdOk() (*string, bool) {
	if o == nil || isNil(o.CloudAccountId) {
		return nil, false
	}
	return o.CloudAccountId, true
}

// HasCloudAccountId returns a boolean if a field has been set.
func (o *ProtoVNetSearchRequestMetadata) HasCloudAccountId() bool {
	if o != nil && !isNil(o.CloudAccountId) {
		return true
	}

	return false
}

// SetCloudAccountId gets a reference to the given string and assigns it to the CloudAccountId field.
func (o *ProtoVNetSearchRequestMetadata) SetCloudAccountId(v string) {
	o.CloudAccountId = &v
}

func (o ProtoVNetSearchRequestMetadata) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.CloudAccountId) {
		toSerialize["cloudAccountId"] = o.CloudAccountId
	}
	return json.Marshal(toSerialize)
}

type NullableProtoVNetSearchRequestMetadata struct {
	value *ProtoVNetSearchRequestMetadata
	isSet bool
}

func (v NullableProtoVNetSearchRequestMetadata) Get() *ProtoVNetSearchRequestMetadata {
	return v.value
}

func (v *NullableProtoVNetSearchRequestMetadata) Set(val *ProtoVNetSearchRequestMetadata) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoVNetSearchRequestMetadata) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoVNetSearchRequestMetadata) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoVNetSearchRequestMetadata(val *ProtoVNetSearchRequestMetadata) *NullableProtoVNetSearchRequestMetadata {
	return &NullableProtoVNetSearchRequestMetadata{value: val, isSet: true}
}

func (v NullableProtoVNetSearchRequestMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoVNetSearchRequestMetadata) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
