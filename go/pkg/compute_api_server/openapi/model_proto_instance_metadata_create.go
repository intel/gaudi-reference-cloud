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

// ProtoInstanceMetadataCreate struct for ProtoInstanceMetadataCreate
type ProtoInstanceMetadataCreate struct {
	CloudAccountId *string `json:"cloudAccountId,omitempty"`
	// Name will be generated if empty.
	Name *string `json:"name,omitempty"`
	// Map of string keys and values that can be used to organize and categorize instances. This is also used by TopologySpreadConstraints.
	Labels *map[string]string `json:"labels,omitempty"`
	// Reserved. Added this field to overcome openAPi-same-struct issue.
	Reserved1 *string `json:"reserved1,omitempty"`
	ProductId *string `json:"productId,omitempty"`
}

// NewProtoInstanceMetadataCreate instantiates a new ProtoInstanceMetadataCreate object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoInstanceMetadataCreate() *ProtoInstanceMetadataCreate {
	this := ProtoInstanceMetadataCreate{}
	return &this
}

// NewProtoInstanceMetadataCreateWithDefaults instantiates a new ProtoInstanceMetadataCreate object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoInstanceMetadataCreateWithDefaults() *ProtoInstanceMetadataCreate {
	this := ProtoInstanceMetadataCreate{}
	return &this
}

// GetCloudAccountId returns the CloudAccountId field value if set, zero value otherwise.
func (o *ProtoInstanceMetadataCreate) GetCloudAccountId() string {
	if o == nil || isNil(o.CloudAccountId) {
		var ret string
		return ret
	}
	return *o.CloudAccountId
}

// GetCloudAccountIdOk returns a tuple with the CloudAccountId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceMetadataCreate) GetCloudAccountIdOk() (*string, bool) {
	if o == nil || isNil(o.CloudAccountId) {
		return nil, false
	}
	return o.CloudAccountId, true
}

// HasCloudAccountId returns a boolean if a field has been set.
func (o *ProtoInstanceMetadataCreate) HasCloudAccountId() bool {
	if o != nil && !isNil(o.CloudAccountId) {
		return true
	}

	return false
}

// SetCloudAccountId gets a reference to the given string and assigns it to the CloudAccountId field.
func (o *ProtoInstanceMetadataCreate) SetCloudAccountId(v string) {
	o.CloudAccountId = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ProtoInstanceMetadataCreate) GetName() string {
	if o == nil || isNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceMetadataCreate) GetNameOk() (*string, bool) {
	if o == nil || isNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ProtoInstanceMetadataCreate) HasName() bool {
	if o != nil && !isNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ProtoInstanceMetadataCreate) SetName(v string) {
	o.Name = &v
}

// GetLabels returns the Labels field value if set, zero value otherwise.
func (o *ProtoInstanceMetadataCreate) GetLabels() map[string]string {
	if o == nil || isNil(o.Labels) {
		var ret map[string]string
		return ret
	}
	return *o.Labels
}

// GetLabelsOk returns a tuple with the Labels field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceMetadataCreate) GetLabelsOk() (*map[string]string, bool) {
	if o == nil || isNil(o.Labels) {
		return nil, false
	}
	return o.Labels, true
}

// HasLabels returns a boolean if a field has been set.
func (o *ProtoInstanceMetadataCreate) HasLabels() bool {
	if o != nil && !isNil(o.Labels) {
		return true
	}

	return false
}

// SetLabels gets a reference to the given map[string]string and assigns it to the Labels field.
func (o *ProtoInstanceMetadataCreate) SetLabels(v map[string]string) {
	o.Labels = &v
}

// GetReserved1 returns the Reserved1 field value if set, zero value otherwise.
func (o *ProtoInstanceMetadataCreate) GetReserved1() string {
	if o == nil || isNil(o.Reserved1) {
		var ret string
		return ret
	}
	return *o.Reserved1
}

// GetReserved1Ok returns a tuple with the Reserved1 field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceMetadataCreate) GetReserved1Ok() (*string, bool) {
	if o == nil || isNil(o.Reserved1) {
		return nil, false
	}
	return o.Reserved1, true
}

// HasReserved1 returns a boolean if a field has been set.
func (o *ProtoInstanceMetadataCreate) HasReserved1() bool {
	if o != nil && !isNil(o.Reserved1) {
		return true
	}

	return false
}

// SetReserved1 gets a reference to the given string and assigns it to the Reserved1 field.
func (o *ProtoInstanceMetadataCreate) SetReserved1(v string) {
	o.Reserved1 = &v
}

// GetProductId returns the ProductId field value if set, zero value otherwise.
func (o *ProtoInstanceMetadataCreate) GetProductId() string {
	if o == nil || isNil(o.ProductId) {
		var ret string
		return ret
	}
	return *o.ProductId
}

// GetProductIdOk returns a tuple with the ProductId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoInstanceMetadataCreate) GetProductIdOk() (*string, bool) {
	if o == nil || isNil(o.ProductId) {
		return nil, false
	}
	return o.ProductId, true
}

// HasProductId returns a boolean if a field has been set.
func (o *ProtoInstanceMetadataCreate) HasProductId() bool {
	if o != nil && !isNil(o.ProductId) {
		return true
	}

	return false
}

// SetProductId gets a reference to the given string and assigns it to the ProductId field.
func (o *ProtoInstanceMetadataCreate) SetProductId(v string) {
	o.ProductId = &v
}

func (o ProtoInstanceMetadataCreate) MarshalJSON() ([]byte, error) {
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
	if !isNil(o.Reserved1) {
		toSerialize["reserved1"] = o.Reserved1
	}
	if !isNil(o.ProductId) {
		toSerialize["productId"] = o.ProductId
	}
	return json.Marshal(toSerialize)
}

type NullableProtoInstanceMetadataCreate struct {
	value *ProtoInstanceMetadataCreate
	isSet bool
}

func (v NullableProtoInstanceMetadataCreate) Get() *ProtoInstanceMetadataCreate {
	return v.value
}

func (v *NullableProtoInstanceMetadataCreate) Set(val *ProtoInstanceMetadataCreate) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoInstanceMetadataCreate) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoInstanceMetadataCreate) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoInstanceMetadataCreate(val *ProtoInstanceMetadataCreate) *NullableProtoInstanceMetadataCreate {
	return &NullableProtoInstanceMetadataCreate{value: val, isSet: true}
}

func (v NullableProtoInstanceMetadataCreate) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoInstanceMetadataCreate) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
