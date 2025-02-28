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

// ProtoResourceMetadataCreate struct for ProtoResourceMetadataCreate
type ProtoResourceMetadataCreate struct {
	CloudAccountId *string `json:"cloudAccountId,omitempty"`
	// If Name is not empty, it must be unique within the cloudAccountId. It will be generated if empty.
	Name *string `json:"name,omitempty"`
	// Not implemented.
	Labels *map[string]string `json:"labels,omitempty"`
}

// NewProtoResourceMetadataCreate instantiates a new ProtoResourceMetadataCreate object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoResourceMetadataCreate() *ProtoResourceMetadataCreate {
	this := ProtoResourceMetadataCreate{}
	return &this
}

// NewProtoResourceMetadataCreateWithDefaults instantiates a new ProtoResourceMetadataCreate object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoResourceMetadataCreateWithDefaults() *ProtoResourceMetadataCreate {
	this := ProtoResourceMetadataCreate{}
	return &this
}

// GetCloudAccountId returns the CloudAccountId field value if set, zero value otherwise.
func (o *ProtoResourceMetadataCreate) GetCloudAccountId() string {
	if o == nil || isNil(o.CloudAccountId) {
		var ret string
		return ret
	}
	return *o.CloudAccountId
}

// GetCloudAccountIdOk returns a tuple with the CloudAccountId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoResourceMetadataCreate) GetCloudAccountIdOk() (*string, bool) {
	if o == nil || isNil(o.CloudAccountId) {
		return nil, false
	}
	return o.CloudAccountId, true
}

// HasCloudAccountId returns a boolean if a field has been set.
func (o *ProtoResourceMetadataCreate) HasCloudAccountId() bool {
	if o != nil && !isNil(o.CloudAccountId) {
		return true
	}

	return false
}

// SetCloudAccountId gets a reference to the given string and assigns it to the CloudAccountId field.
func (o *ProtoResourceMetadataCreate) SetCloudAccountId(v string) {
	o.CloudAccountId = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ProtoResourceMetadataCreate) GetName() string {
	if o == nil || isNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoResourceMetadataCreate) GetNameOk() (*string, bool) {
	if o == nil || isNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ProtoResourceMetadataCreate) HasName() bool {
	if o != nil && !isNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ProtoResourceMetadataCreate) SetName(v string) {
	o.Name = &v
}

// GetLabels returns the Labels field value if set, zero value otherwise.
func (o *ProtoResourceMetadataCreate) GetLabels() map[string]string {
	if o == nil || isNil(o.Labels) {
		var ret map[string]string
		return ret
	}
	return *o.Labels
}

// GetLabelsOk returns a tuple with the Labels field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoResourceMetadataCreate) GetLabelsOk() (*map[string]string, bool) {
	if o == nil || isNil(o.Labels) {
		return nil, false
	}
	return o.Labels, true
}

// HasLabels returns a boolean if a field has been set.
func (o *ProtoResourceMetadataCreate) HasLabels() bool {
	if o != nil && !isNil(o.Labels) {
		return true
	}

	return false
}

// SetLabels gets a reference to the given map[string]string and assigns it to the Labels field.
func (o *ProtoResourceMetadataCreate) SetLabels(v map[string]string) {
	o.Labels = &v
}

func (o ProtoResourceMetadataCreate) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.CloudAccountId) {
		toSerialize["cloudAccountId"] = o.CloudAccountId
	}
	if !isNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !isNil(o.Labels) {
		toSerialize["labels"] = o.Labels
	}
	return json.Marshal(toSerialize)
}

type NullableProtoResourceMetadataCreate struct {
	value *ProtoResourceMetadataCreate
	isSet bool
}

func (v NullableProtoResourceMetadataCreate) Get() *ProtoResourceMetadataCreate {
	return v.value
}

func (v *NullableProtoResourceMetadataCreate) Set(val *ProtoResourceMetadataCreate) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoResourceMetadataCreate) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoResourceMetadataCreate) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoResourceMetadataCreate(val *ProtoResourceMetadataCreate) *NullableProtoResourceMetadataCreate {
	return &NullableProtoResourceMetadataCreate{value: val, isSet: true}
}

func (v NullableProtoResourceMetadataCreate) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoResourceMetadataCreate) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
