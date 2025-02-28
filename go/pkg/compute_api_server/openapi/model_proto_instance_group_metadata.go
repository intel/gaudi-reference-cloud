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

// ProtoInstanceGroupMetadata struct for ProtoInstanceGroupMetadata
type ProtoInstanceGroupMetadata struct {
	CloudAccountId *string `json:"cloudAccountId,omitempty"`
	Name           *string `json:"name,omitempty"`
	// Reserved. Added this field to overcome openAPi-same-struct issue.
	Reserved2 *string `json:"reserved2,omitempty"`
}

// NewProtoInstanceGroupMetadata instantiates a new ProtoInstanceGroupMetadata object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoInstanceGroupMetadata() *ProtoInstanceGroupMetadata {
	this := ProtoInstanceGroupMetadata{}
	return &this
}

// NewProtoInstanceGroupMetadataWithDefaults instantiates a new ProtoInstanceGroupMetadata object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoInstanceGroupMetadataWithDefaults() *ProtoInstanceGroupMetadata {
	this := ProtoInstanceGroupMetadata{}
	return &this
}

// GetCloudAccountId returns the CloudAccountId field value if set, zero value otherwise.
func (o *ProtoInstanceGroupMetadata) GetCloudAccountId() string {
	if o == nil || isNil(o.CloudAccountId) {
		var ret string
		return ret
	}
	return *o.CloudAccountId
}

// GetCloudAccountIdOk returns a tuple with the CloudAccountId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceGroupMetadata) GetCloudAccountIdOk() (*string, bool) {
	if o == nil || isNil(o.CloudAccountId) {
		return nil, false
	}
	return o.CloudAccountId, true
}

// HasCloudAccountId returns a boolean if a field has been set.
func (o *ProtoInstanceGroupMetadata) HasCloudAccountId() bool {
	if o != nil && !isNil(o.CloudAccountId) {
		return true
	}

	return false
}

// SetCloudAccountId gets a reference to the given string and assigns it to the CloudAccountId field.
func (o *ProtoInstanceGroupMetadata) SetCloudAccountId(v string) {
	o.CloudAccountId = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ProtoInstanceGroupMetadata) GetName() string {
	if o == nil || isNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceGroupMetadata) GetNameOk() (*string, bool) {
	if o == nil || isNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ProtoInstanceGroupMetadata) HasName() bool {
	if o != nil && !isNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ProtoInstanceGroupMetadata) SetName(v string) {
	o.Name = &v
}

// GetReserved2 returns the Reserved2 field value if set, zero value otherwise.
func (o *ProtoInstanceGroupMetadata) GetReserved2() string {
	if o == nil || isNil(o.Reserved2) {
		var ret string
		return ret
	}
	return *o.Reserved2
}

// GetReserved2Ok returns a tuple with the Reserved2 field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceGroupMetadata) GetReserved2Ok() (*string, bool) {
	if o == nil || isNil(o.Reserved2) {
		return nil, false
	}
	return o.Reserved2, true
}

// HasReserved2 returns a boolean if a field has been set.
func (o *ProtoInstanceGroupMetadata) HasReserved2() bool {
	if o != nil && !isNil(o.Reserved2) {
		return true
	}

	return false
}

// SetReserved2 gets a reference to the given string and assigns it to the Reserved2 field.
func (o *ProtoInstanceGroupMetadata) SetReserved2(v string) {
	o.Reserved2 = &v
}

func (o ProtoInstanceGroupMetadata) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.CloudAccountId) {
		toSerialize["cloudAccountId"] = o.CloudAccountId
	}
	if !isNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !isNil(o.Reserved2) {
		toSerialize["reserved2"] = o.Reserved2
	}
	return json.Marshal(toSerialize)
}

type NullableProtoInstanceGroupMetadata struct {
	value *ProtoInstanceGroupMetadata
	isSet bool
}

func (v NullableProtoInstanceGroupMetadata) Get() *ProtoInstanceGroupMetadata {
	return v.value
}

func (v *NullableProtoInstanceGroupMetadata) Set(val *ProtoInstanceGroupMetadata) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoInstanceGroupMetadata) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoInstanceGroupMetadata) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoInstanceGroupMetadata(val *ProtoInstanceGroupMetadata) *NullableProtoInstanceGroupMetadata {
	return &NullableProtoInstanceGroupMetadata{value: val, isSet: true}
}

func (v NullableProtoInstanceGroupMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoInstanceGroupMetadata) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
