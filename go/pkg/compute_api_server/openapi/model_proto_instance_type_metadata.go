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

// ProtoInstanceTypeMetadata struct for ProtoInstanceTypeMetadata
type ProtoInstanceTypeMetadata struct {
	// Unique name of the instance type.
	Name *string `json:"name,omitempty"`
}

// NewProtoInstanceTypeMetadata instantiates a new ProtoInstanceTypeMetadata object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoInstanceTypeMetadata() *ProtoInstanceTypeMetadata {
	this := ProtoInstanceTypeMetadata{}
	return &this
}

// NewProtoInstanceTypeMetadataWithDefaults instantiates a new ProtoInstanceTypeMetadata object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoInstanceTypeMetadataWithDefaults() *ProtoInstanceTypeMetadata {
	this := ProtoInstanceTypeMetadata{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ProtoInstanceTypeMetadata) GetName() string {
	if o == nil || isNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceTypeMetadata) GetNameOk() (*string, bool) {
	if o == nil || isNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ProtoInstanceTypeMetadata) HasName() bool {
	if o != nil && !isNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ProtoInstanceTypeMetadata) SetName(v string) {
	o.Name = &v
}

func (o ProtoInstanceTypeMetadata) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	return json.Marshal(toSerialize)
}

type NullableProtoInstanceTypeMetadata struct {
	value *ProtoInstanceTypeMetadata
	isSet bool
}

func (v NullableProtoInstanceTypeMetadata) Get() *ProtoInstanceTypeMetadata {
	return v.value
}

func (v *NullableProtoInstanceTypeMetadata) Set(val *ProtoInstanceTypeMetadata) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoInstanceTypeMetadata) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoInstanceTypeMetadata) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoInstanceTypeMetadata(val *ProtoInstanceTypeMetadata) *NullableProtoInstanceTypeMetadata {
	return &NullableProtoInstanceTypeMetadata{value: val, isSet: true}
}

func (v NullableProtoInstanceTypeMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoInstanceTypeMetadata) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
