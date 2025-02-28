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

// ProtoVNetPutRequestMetadata struct for ProtoVNetPutRequestMetadata
type ProtoVNetPutRequestMetadata struct {
	CloudAccountId *string `json:"cloudAccountId,omitempty"`
	Name           *string `json:"name,omitempty"`
}

// NewProtoVNetPutRequestMetadata instantiates a new ProtoVNetPutRequestMetadata object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoVNetPutRequestMetadata() *ProtoVNetPutRequestMetadata {
	this := ProtoVNetPutRequestMetadata{}
	return &this
}

// NewProtoVNetPutRequestMetadataWithDefaults instantiates a new ProtoVNetPutRequestMetadata object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoVNetPutRequestMetadataWithDefaults() *ProtoVNetPutRequestMetadata {
	this := ProtoVNetPutRequestMetadata{}
	return &this
}

// GetCloudAccountId returns the CloudAccountId field value if set, zero value otherwise.
func (o *ProtoVNetPutRequestMetadata) GetCloudAccountId() string {
	if o == nil || isNil(o.CloudAccountId) {
		var ret string
		return ret
	}
	return *o.CloudAccountId
}

// GetCloudAccountIdOk returns a tuple with the CloudAccountId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoVNetPutRequestMetadata) GetCloudAccountIdOk() (*string, bool) {
	if o == nil || isNil(o.CloudAccountId) {
		return nil, false
	}
	return o.CloudAccountId, true
}

// HasCloudAccountId returns a boolean if a field has been set.
func (o *ProtoVNetPutRequestMetadata) HasCloudAccountId() bool {
	if o != nil && !isNil(o.CloudAccountId) {
		return true
	}

	return false
}

// SetCloudAccountId gets a reference to the given string and assigns it to the CloudAccountId field.
func (o *ProtoVNetPutRequestMetadata) SetCloudAccountId(v string) {
	o.CloudAccountId = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ProtoVNetPutRequestMetadata) GetName() string {
	if o == nil || isNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoVNetPutRequestMetadata) GetNameOk() (*string, bool) {
	if o == nil || isNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ProtoVNetPutRequestMetadata) HasName() bool {
	if o != nil && !isNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ProtoVNetPutRequestMetadata) SetName(v string) {
	o.Name = &v
}

func (o ProtoVNetPutRequestMetadata) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.CloudAccountId) {
		toSerialize["cloudAccountId"] = o.CloudAccountId
	}
	if !isNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	return json.Marshal(toSerialize)
}

type NullableProtoVNetPutRequestMetadata struct {
	value *ProtoVNetPutRequestMetadata
	isSet bool
}

func (v NullableProtoVNetPutRequestMetadata) Get() *ProtoVNetPutRequestMetadata {
	return v.value
}

func (v *NullableProtoVNetPutRequestMetadata) Set(val *ProtoVNetPutRequestMetadata) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoVNetPutRequestMetadata) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoVNetPutRequestMetadata) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoVNetPutRequestMetadata(val *ProtoVNetPutRequestMetadata) *NullableProtoVNetPutRequestMetadata {
	return &NullableProtoVNetPutRequestMetadata{value: val, isSet: true}
}

func (v NullableProtoVNetPutRequestMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoVNetPutRequestMetadata) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
