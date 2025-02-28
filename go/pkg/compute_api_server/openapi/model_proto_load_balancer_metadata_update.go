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

// ProtoLoadBalancerMetadataUpdate struct for ProtoLoadBalancerMetadataUpdate
type ProtoLoadBalancerMetadataUpdate struct {
	CloudAccountId *string `json:"cloudAccountId,omitempty"`
	Name           *string `json:"name,omitempty"`
	ResourceId     *string `json:"resourceId,omitempty"`
	// If provided, the existing record must have this resourceVersion for the request to succeed.
	ResourceVersion *string `json:"resourceVersion,omitempty"`
	// Map of string keys and values that can be used to organize and categorize load balancers.
	Labels *map[string]string `json:"labels,omitempty"`
	// Reserved. Added this field to overcome openAPi-same-struct issue.
	Reserved1 *string `json:"reserved1,omitempty"`
}

// NewProtoLoadBalancerMetadataUpdate instantiates a new ProtoLoadBalancerMetadataUpdate object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoLoadBalancerMetadataUpdate() *ProtoLoadBalancerMetadataUpdate {
	this := ProtoLoadBalancerMetadataUpdate{}
	return &this
}

// NewProtoLoadBalancerMetadataUpdateWithDefaults instantiates a new ProtoLoadBalancerMetadataUpdate object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoLoadBalancerMetadataUpdateWithDefaults() *ProtoLoadBalancerMetadataUpdate {
	this := ProtoLoadBalancerMetadataUpdate{}
	return &this
}

// GetCloudAccountId returns the CloudAccountId field value if set, zero value otherwise.
func (o *ProtoLoadBalancerMetadataUpdate) GetCloudAccountId() string {
	if o == nil || isNil(o.CloudAccountId) {
		var ret string
		return ret
	}
	return *o.CloudAccountId
}

// GetCloudAccountIdOk returns a tuple with the CloudAccountId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerMetadataUpdate) GetCloudAccountIdOk() (*string, bool) {
	if o == nil || isNil(o.CloudAccountId) {
		return nil, false
	}
	return o.CloudAccountId, true
}

// HasCloudAccountId returns a boolean if a field has been set.
func (o *ProtoLoadBalancerMetadataUpdate) HasCloudAccountId() bool {
	if o != nil && !isNil(o.CloudAccountId) {
		return true
	}

	return false
}

// SetCloudAccountId gets a reference to the given string and assigns it to the CloudAccountId field.
func (o *ProtoLoadBalancerMetadataUpdate) SetCloudAccountId(v string) {
	o.CloudAccountId = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ProtoLoadBalancerMetadataUpdate) GetName() string {
	if o == nil || isNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerMetadataUpdate) GetNameOk() (*string, bool) {
	if o == nil || isNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ProtoLoadBalancerMetadataUpdate) HasName() bool {
	if o != nil && !isNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ProtoLoadBalancerMetadataUpdate) SetName(v string) {
	o.Name = &v
}

// GetResourceId returns the ResourceId field value if set, zero value otherwise.
func (o *ProtoLoadBalancerMetadataUpdate) GetResourceId() string {
	if o == nil || isNil(o.ResourceId) {
		var ret string
		return ret
	}
	return *o.ResourceId
}

// GetResourceIdOk returns a tuple with the ResourceId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerMetadataUpdate) GetResourceIdOk() (*string, bool) {
	if o == nil || isNil(o.ResourceId) {
		return nil, false
	}
	return o.ResourceId, true
}

// HasResourceId returns a boolean if a field has been set.
func (o *ProtoLoadBalancerMetadataUpdate) HasResourceId() bool {
	if o != nil && !isNil(o.ResourceId) {
		return true
	}

	return false
}

// SetResourceId gets a reference to the given string and assigns it to the ResourceId field.
func (o *ProtoLoadBalancerMetadataUpdate) SetResourceId(v string) {
	o.ResourceId = &v
}

// GetResourceVersion returns the ResourceVersion field value if set, zero value otherwise.
func (o *ProtoLoadBalancerMetadataUpdate) GetResourceVersion() string {
	if o == nil || isNil(o.ResourceVersion) {
		var ret string
		return ret
	}
	return *o.ResourceVersion
}

// GetResourceVersionOk returns a tuple with the ResourceVersion field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerMetadataUpdate) GetResourceVersionOk() (*string, bool) {
	if o == nil || isNil(o.ResourceVersion) {
		return nil, false
	}
	return o.ResourceVersion, true
}

// HasResourceVersion returns a boolean if a field has been set.
func (o *ProtoLoadBalancerMetadataUpdate) HasResourceVersion() bool {
	if o != nil && !isNil(o.ResourceVersion) {
		return true
	}

	return false
}

// SetResourceVersion gets a reference to the given string and assigns it to the ResourceVersion field.
func (o *ProtoLoadBalancerMetadataUpdate) SetResourceVersion(v string) {
	o.ResourceVersion = &v
}

// GetLabels returns the Labels field value if set, zero value otherwise.
func (o *ProtoLoadBalancerMetadataUpdate) GetLabels() map[string]string {
	if o == nil || isNil(o.Labels) {
		var ret map[string]string
		return ret
	}
	return *o.Labels
}

// GetLabelsOk returns a tuple with the Labels field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerMetadataUpdate) GetLabelsOk() (*map[string]string, bool) {
	if o == nil || isNil(o.Labels) {
		return nil, false
	}
	return o.Labels, true
}

// HasLabels returns a boolean if a field has been set.
func (o *ProtoLoadBalancerMetadataUpdate) HasLabels() bool {
	if o != nil && !isNil(o.Labels) {
		return true
	}

	return false
}

// SetLabels gets a reference to the given map[string]string and assigns it to the Labels field.
func (o *ProtoLoadBalancerMetadataUpdate) SetLabels(v map[string]string) {
	o.Labels = &v
}

// GetReserved1 returns the Reserved1 field value if set, zero value otherwise.
func (o *ProtoLoadBalancerMetadataUpdate) GetReserved1() string {
	if o == nil || isNil(o.Reserved1) {
		var ret string
		return ret
	}
	return *o.Reserved1
}

// GetReserved1Ok returns a tuple with the Reserved1 field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerMetadataUpdate) GetReserved1Ok() (*string, bool) {
	if o == nil || isNil(o.Reserved1) {
		return nil, false
	}
	return o.Reserved1, true
}

// HasReserved1 returns a boolean if a field has been set.
func (o *ProtoLoadBalancerMetadataUpdate) HasReserved1() bool {
	if o != nil && !isNil(o.Reserved1) {
		return true
	}

	return false
}

// SetReserved1 gets a reference to the given string and assigns it to the Reserved1 field.
func (o *ProtoLoadBalancerMetadataUpdate) SetReserved1(v string) {
	o.Reserved1 = &v
}

func (o ProtoLoadBalancerMetadataUpdate) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.CloudAccountId) {
		toSerialize["cloudAccountId"] = o.CloudAccountId
	}
	if !isNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !isNil(o.ResourceId) {
		toSerialize["resourceId"] = o.ResourceId
	}
	if !isNil(o.ResourceVersion) {
		toSerialize["resourceVersion"] = o.ResourceVersion
	}
	if !isNil(o.Labels) {
		toSerialize["labels"] = o.Labels
	}
	if !isNil(o.Reserved1) {
		toSerialize["reserved1"] = o.Reserved1
	}
	return json.Marshal(toSerialize)
}

type NullableProtoLoadBalancerMetadataUpdate struct {
	value *ProtoLoadBalancerMetadataUpdate
	isSet bool
}

func (v NullableProtoLoadBalancerMetadataUpdate) Get() *ProtoLoadBalancerMetadataUpdate {
	return v.value
}

func (v *NullableProtoLoadBalancerMetadataUpdate) Set(val *ProtoLoadBalancerMetadataUpdate) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoLoadBalancerMetadataUpdate) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoLoadBalancerMetadataUpdate) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoLoadBalancerMetadataUpdate(val *ProtoLoadBalancerMetadataUpdate) *NullableProtoLoadBalancerMetadataUpdate {
	return &NullableProtoLoadBalancerMetadataUpdate{value: val, isSet: true}
}

func (v NullableProtoLoadBalancerMetadataUpdate) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoLoadBalancerMetadataUpdate) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
