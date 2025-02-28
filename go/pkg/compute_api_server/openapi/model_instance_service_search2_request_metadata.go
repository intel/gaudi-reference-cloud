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

// InstanceServiceSearch2RequestMetadata struct for InstanceServiceSearch2RequestMetadata
type InstanceServiceSearch2RequestMetadata struct {
	// If not empty, only return instances that have these key/value pairs.
	Labels *map[string]string `json:"labels,omitempty"`
	// Reserved. Added this field to overcome openAPi-same-struct issue.
	Reserved1           *string                    `json:"reserved1,omitempty"`
	InstanceGroup       *string                    `json:"instanceGroup,omitempty"`
	InstanceGroupFilter *ProtoSearchFilterCriteria `json:"instanceGroupFilter,omitempty"`
}

// NewInstanceServiceSearch2RequestMetadata instantiates a new InstanceServiceSearch2RequestMetadata object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewInstanceServiceSearch2RequestMetadata() *InstanceServiceSearch2RequestMetadata {
	this := InstanceServiceSearch2RequestMetadata{}
	var instanceGroupFilter ProtoSearchFilterCriteria = DEFAULT
	this.InstanceGroupFilter = &instanceGroupFilter
	return &this
}

// NewInstanceServiceSearch2RequestMetadataWithDefaults instantiates a new InstanceServiceSearch2RequestMetadata object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewInstanceServiceSearch2RequestMetadataWithDefaults() *InstanceServiceSearch2RequestMetadata {
	this := InstanceServiceSearch2RequestMetadata{}
	var instanceGroupFilter ProtoSearchFilterCriteria = DEFAULT
	this.InstanceGroupFilter = &instanceGroupFilter
	return &this
}

// GetLabels returns the Labels field value if set, zero value otherwise.
func (o *InstanceServiceSearch2RequestMetadata) GetLabels() map[string]string {
	if o == nil || isNil(o.Labels) {
		var ret map[string]string
		return ret
	}
	return *o.Labels
}

// GetLabelsOk returns a tuple with the Labels field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceServiceSearch2RequestMetadata) GetLabelsOk() (*map[string]string, bool) {
	if o == nil || isNil(o.Labels) {
		return nil, false
	}
	return o.Labels, true
}

// HasLabels returns a boolean if a field has been set.
func (o *InstanceServiceSearch2RequestMetadata) HasLabels() bool {
	if o != nil && !isNil(o.Labels) {
		return true
	}

	return false
}

// SetLabels gets a reference to the given map[string]string and assigns it to the Labels field.
func (o *InstanceServiceSearch2RequestMetadata) SetLabels(v map[string]string) {
	o.Labels = &v
}

// GetReserved1 returns the Reserved1 field value if set, zero value otherwise.
func (o *InstanceServiceSearch2RequestMetadata) GetReserved1() string {
	if o == nil || isNil(o.Reserved1) {
		var ret string
		return ret
	}
	return *o.Reserved1
}

// GetReserved1Ok returns a tuple with the Reserved1 field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceServiceSearch2RequestMetadata) GetReserved1Ok() (*string, bool) {
	if o == nil || isNil(o.Reserved1) {
		return nil, false
	}
	return o.Reserved1, true
}

// HasReserved1 returns a boolean if a field has been set.
func (o *InstanceServiceSearch2RequestMetadata) HasReserved1() bool {
	if o != nil && !isNil(o.Reserved1) {
		return true
	}

	return false
}

// SetReserved1 gets a reference to the given string and assigns it to the Reserved1 field.
func (o *InstanceServiceSearch2RequestMetadata) SetReserved1(v string) {
	o.Reserved1 = &v
}

// GetInstanceGroup returns the InstanceGroup field value if set, zero value otherwise.
func (o *InstanceServiceSearch2RequestMetadata) GetInstanceGroup() string {
	if o == nil || isNil(o.InstanceGroup) {
		var ret string
		return ret
	}
	return *o.InstanceGroup
}

// GetInstanceGroupOk returns a tuple with the InstanceGroup field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceServiceSearch2RequestMetadata) GetInstanceGroupOk() (*string, bool) {
	if o == nil || isNil(o.InstanceGroup) {
		return nil, false
	}
	return o.InstanceGroup, true
}

// HasInstanceGroup returns a boolean if a field has been set.
func (o *InstanceServiceSearch2RequestMetadata) HasInstanceGroup() bool {
	if o != nil && !isNil(o.InstanceGroup) {
		return true
	}

	return false
}

// SetInstanceGroup gets a reference to the given string and assigns it to the InstanceGroup field.
func (o *InstanceServiceSearch2RequestMetadata) SetInstanceGroup(v string) {
	o.InstanceGroup = &v
}

// GetInstanceGroupFilter returns the InstanceGroupFilter field value if set, zero value otherwise.
func (o *InstanceServiceSearch2RequestMetadata) GetInstanceGroupFilter() ProtoSearchFilterCriteria {
	if o == nil || isNil(o.InstanceGroupFilter) {
		var ret ProtoSearchFilterCriteria
		return ret
	}
	return *o.InstanceGroupFilter
}

// GetInstanceGroupFilterOk returns a tuple with the InstanceGroupFilter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceServiceSearch2RequestMetadata) GetInstanceGroupFilterOk() (*ProtoSearchFilterCriteria, bool) {
	if o == nil || isNil(o.InstanceGroupFilter) {
		return nil, false
	}
	return o.InstanceGroupFilter, true
}

// HasInstanceGroupFilter returns a boolean if a field has been set.
func (o *InstanceServiceSearch2RequestMetadata) HasInstanceGroupFilter() bool {
	if o != nil && !isNil(o.InstanceGroupFilter) {
		return true
	}

	return false
}

// SetInstanceGroupFilter gets a reference to the given ProtoSearchFilterCriteria and assigns it to the InstanceGroupFilter field.
func (o *InstanceServiceSearch2RequestMetadata) SetInstanceGroupFilter(v ProtoSearchFilterCriteria) {
	o.InstanceGroupFilter = &v
}

func (o InstanceServiceSearch2RequestMetadata) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Labels) {
		toSerialize["labels"] = o.Labels
	}
	if !isNil(o.Reserved1) {
		toSerialize["reserved1"] = o.Reserved1
	}
	if !isNil(o.InstanceGroup) {
		toSerialize["instanceGroup"] = o.InstanceGroup
	}
	if !isNil(o.InstanceGroupFilter) {
		toSerialize["instanceGroupFilter"] = o.InstanceGroupFilter
	}
	return json.Marshal(toSerialize)
}

type NullableInstanceServiceSearch2RequestMetadata struct {
	value *InstanceServiceSearch2RequestMetadata
	isSet bool
}

func (v NullableInstanceServiceSearch2RequestMetadata) Get() *InstanceServiceSearch2RequestMetadata {
	return v.value
}

func (v *NullableInstanceServiceSearch2RequestMetadata) Set(val *InstanceServiceSearch2RequestMetadata) {
	v.value = val
	v.isSet = true
}

func (v NullableInstanceServiceSearch2RequestMetadata) IsSet() bool {
	return v.isSet
}

func (v *NullableInstanceServiceSearch2RequestMetadata) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableInstanceServiceSearch2RequestMetadata(val *InstanceServiceSearch2RequestMetadata) *NullableInstanceServiceSearch2RequestMetadata {
	return &NullableInstanceServiceSearch2RequestMetadata{value: val, isSet: true}
}

func (v NullableInstanceServiceSearch2RequestMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableInstanceServiceSearch2RequestMetadata) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
