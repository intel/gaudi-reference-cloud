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

// ProtoInstanceGroupMetadataUpdate struct for ProtoInstanceGroupMetadataUpdate
type ProtoInstanceGroupMetadataUpdate struct {
	CloudAccountId *string `json:"cloudAccountId,omitempty"`
	Name           *string `json:"name,omitempty"`
	// Reserved. Added this field to overcome openAPi-same-struct issue.
	Reserved3 *string `json:"reserved3,omitempty"`
}

// NewProtoInstanceGroupMetadataUpdate instantiates a new ProtoInstanceGroupMetadataUpdate object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoInstanceGroupMetadataUpdate() *ProtoInstanceGroupMetadataUpdate {
	this := ProtoInstanceGroupMetadataUpdate{}
	return &this
}

// NewProtoInstanceGroupMetadataUpdateWithDefaults instantiates a new ProtoInstanceGroupMetadataUpdate object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoInstanceGroupMetadataUpdateWithDefaults() *ProtoInstanceGroupMetadataUpdate {
	this := ProtoInstanceGroupMetadataUpdate{}
	return &this
}

// GetCloudAccountId returns the CloudAccountId field value if set, zero value otherwise.
func (o *ProtoInstanceGroupMetadataUpdate) GetCloudAccountId() string {
	if o == nil || isNil(o.CloudAccountId) {
		var ret string
		return ret
	}
	return *o.CloudAccountId
}

// GetCloudAccountIdOk returns a tuple with the CloudAccountId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceGroupMetadataUpdate) GetCloudAccountIdOk() (*string, bool) {
	if o == nil || isNil(o.CloudAccountId) {
		return nil, false
	}
	return o.CloudAccountId, true
}

// HasCloudAccountId returns a boolean if a field has been set.
func (o *ProtoInstanceGroupMetadataUpdate) HasCloudAccountId() bool {
	if o != nil && !isNil(o.CloudAccountId) {
		return true
	}

	return false
}

// SetCloudAccountId gets a reference to the given string and assigns it to the CloudAccountId field.
func (o *ProtoInstanceGroupMetadataUpdate) SetCloudAccountId(v string) {
	o.CloudAccountId = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ProtoInstanceGroupMetadataUpdate) GetName() string {
	if o == nil || isNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceGroupMetadataUpdate) GetNameOk() (*string, bool) {
	if o == nil || isNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ProtoInstanceGroupMetadataUpdate) HasName() bool {
	if o != nil && !isNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ProtoInstanceGroupMetadataUpdate) SetName(v string) {
	o.Name = &v
}

// GetReserved3 returns the Reserved3 field value if set, zero value otherwise.
func (o *ProtoInstanceGroupMetadataUpdate) GetReserved3() string {
	if o == nil || isNil(o.Reserved3) {
		var ret string
		return ret
	}
	return *o.Reserved3
}

// GetReserved3Ok returns a tuple with the Reserved3 field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceGroupMetadataUpdate) GetReserved3Ok() (*string, bool) {
	if o == nil || isNil(o.Reserved3) {
		return nil, false
	}
	return o.Reserved3, true
}

// HasReserved3 returns a boolean if a field has been set.
func (o *ProtoInstanceGroupMetadataUpdate) HasReserved3() bool {
	if o != nil && !isNil(o.Reserved3) {
		return true
	}

	return false
}

// SetReserved3 gets a reference to the given string and assigns it to the Reserved3 field.
func (o *ProtoInstanceGroupMetadataUpdate) SetReserved3(v string) {
	o.Reserved3 = &v
}

func (o ProtoInstanceGroupMetadataUpdate) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.CloudAccountId) {
		toSerialize["cloudAccountId"] = o.CloudAccountId
	}
	if !isNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !isNil(o.Reserved3) {
		toSerialize["reserved3"] = o.Reserved3
	}
	return json.Marshal(toSerialize)
}

type NullableProtoInstanceGroupMetadataUpdate struct {
	value *ProtoInstanceGroupMetadataUpdate
	isSet bool
}

func (v NullableProtoInstanceGroupMetadataUpdate) Get() *ProtoInstanceGroupMetadataUpdate {
	return v.value
}

func (v *NullableProtoInstanceGroupMetadataUpdate) Set(val *ProtoInstanceGroupMetadataUpdate) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoInstanceGroupMetadataUpdate) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoInstanceGroupMetadataUpdate) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoInstanceGroupMetadataUpdate(val *ProtoInstanceGroupMetadataUpdate) *NullableProtoInstanceGroupMetadataUpdate {
	return &NullableProtoInstanceGroupMetadataUpdate{value: val, isSet: true}
}

func (v NullableProtoInstanceGroupMetadataUpdate) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoInstanceGroupMetadataUpdate) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
