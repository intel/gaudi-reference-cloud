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

// LoadBalancerServiceUpdate2RequestMetadata struct for LoadBalancerServiceUpdate2RequestMetadata
type LoadBalancerServiceUpdate2RequestMetadata struct {
	ResourceId *string `json:"resourceId,omitempty"`
	// If provided, the existing record must have this resourceVersion for the request to succeed.
	ResourceVersion *string `json:"resourceVersion,omitempty"`
	// Map of string keys and values that can be used to organize and categorize load balancers.
	Labels *map[string]string `json:"labels,omitempty"`
	// Reserved. Added this field to overcome openAPi-same-struct issue.
	Reserved1 *string `json:"reserved1,omitempty"`
}

// NewLoadBalancerServiceUpdate2RequestMetadata instantiates a new LoadBalancerServiceUpdate2RequestMetadata object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLoadBalancerServiceUpdate2RequestMetadata() *LoadBalancerServiceUpdate2RequestMetadata {
	this := LoadBalancerServiceUpdate2RequestMetadata{}
	return &this
}

// NewLoadBalancerServiceUpdate2RequestMetadataWithDefaults instantiates a new LoadBalancerServiceUpdate2RequestMetadata object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLoadBalancerServiceUpdate2RequestMetadataWithDefaults() *LoadBalancerServiceUpdate2RequestMetadata {
	this := LoadBalancerServiceUpdate2RequestMetadata{}
	return &this
}

// GetResourceId returns the ResourceId field value if set, zero value otherwise.
func (o *LoadBalancerServiceUpdate2RequestMetadata) GetResourceId() string {
	if o == nil || isNil(o.ResourceId) {
		var ret string
		return ret
	}
	return *o.ResourceId
}

// GetResourceIdOk returns a tuple with the ResourceId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LoadBalancerServiceUpdate2RequestMetadata) GetResourceIdOk() (*string, bool) {
	if o == nil || isNil(o.ResourceId) {
		return nil, false
	}
	return o.ResourceId, true
}

// HasResourceId returns a boolean if a field has been set.
func (o *LoadBalancerServiceUpdate2RequestMetadata) HasResourceId() bool {
	if o != nil && !isNil(o.ResourceId) {
		return true
	}

	return false
}

// SetResourceId gets a reference to the given string and assigns it to the ResourceId field.
func (o *LoadBalancerServiceUpdate2RequestMetadata) SetResourceId(v string) {
	o.ResourceId = &v
}

// GetResourceVersion returns the ResourceVersion field value if set, zero value otherwise.
func (o *LoadBalancerServiceUpdate2RequestMetadata) GetResourceVersion() string {
	if o == nil || isNil(o.ResourceVersion) {
		var ret string
		return ret
	}
	return *o.ResourceVersion
}

// GetResourceVersionOk returns a tuple with the ResourceVersion field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LoadBalancerServiceUpdate2RequestMetadata) GetResourceVersionOk() (*string, bool) {
	if o == nil || isNil(o.ResourceVersion) {
		return nil, false
	}
	return o.ResourceVersion, true
}

// HasResourceVersion returns a boolean if a field has been set.
func (o *LoadBalancerServiceUpdate2RequestMetadata) HasResourceVersion() bool {
	if o != nil && !isNil(o.ResourceVersion) {
		return true
	}

	return false
}

// SetResourceVersion gets a reference to the given string and assigns it to the ResourceVersion field.
func (o *LoadBalancerServiceUpdate2RequestMetadata) SetResourceVersion(v string) {
	o.ResourceVersion = &v
}

// GetLabels returns the Labels field value if set, zero value otherwise.
func (o *LoadBalancerServiceUpdate2RequestMetadata) GetLabels() map[string]string {
	if o == nil || isNil(o.Labels) {
		var ret map[string]string
		return ret
	}
	return *o.Labels
}

// GetLabelsOk returns a tuple with the Labels field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LoadBalancerServiceUpdate2RequestMetadata) GetLabelsOk() (*map[string]string, bool) {
	if o == nil || isNil(o.Labels) {
		return nil, false
	}
	return o.Labels, true
}

// HasLabels returns a boolean if a field has been set.
func (o *LoadBalancerServiceUpdate2RequestMetadata) HasLabels() bool {
	if o != nil && !isNil(o.Labels) {
		return true
	}

	return false
}

// SetLabels gets a reference to the given map[string]string and assigns it to the Labels field.
func (o *LoadBalancerServiceUpdate2RequestMetadata) SetLabels(v map[string]string) {
	o.Labels = &v
}

// GetReserved1 returns the Reserved1 field value if set, zero value otherwise.
func (o *LoadBalancerServiceUpdate2RequestMetadata) GetReserved1() string {
	if o == nil || isNil(o.Reserved1) {
		var ret string
		return ret
	}
	return *o.Reserved1
}

// GetReserved1Ok returns a tuple with the Reserved1 field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *LoadBalancerServiceUpdate2RequestMetadata) GetReserved1Ok() (*string, bool) {
	if o == nil || isNil(o.Reserved1) {
		return nil, false
	}
	return o.Reserved1, true
}

// HasReserved1 returns a boolean if a field has been set.
func (o *LoadBalancerServiceUpdate2RequestMetadata) HasReserved1() bool {
	if o != nil && !isNil(o.Reserved1) {
		return true
	}

	return false
}

// SetReserved1 gets a reference to the given string and assigns it to the Reserved1 field.
func (o *LoadBalancerServiceUpdate2RequestMetadata) SetReserved1(v string) {
	o.Reserved1 = &v
}

func (o LoadBalancerServiceUpdate2RequestMetadata) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
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

type NullableLoadBalancerServiceUpdate2RequestMetadata struct {
	value *LoadBalancerServiceUpdate2RequestMetadata
	isSet bool
}

func (v NullableLoadBalancerServiceUpdate2RequestMetadata) Get() *LoadBalancerServiceUpdate2RequestMetadata {
	return v.value
}

func (v *NullableLoadBalancerServiceUpdate2RequestMetadata) Set(val *LoadBalancerServiceUpdate2RequestMetadata) {
	v.value = val
	v.isSet = true
}

func (v NullableLoadBalancerServiceUpdate2RequestMetadata) IsSet() bool {
	return v.isSet
}

func (v *NullableLoadBalancerServiceUpdate2RequestMetadata) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLoadBalancerServiceUpdate2RequestMetadata(val *LoadBalancerServiceUpdate2RequestMetadata) *NullableLoadBalancerServiceUpdate2RequestMetadata {
	return &NullableLoadBalancerServiceUpdate2RequestMetadata{value: val, isSet: true}
}

func (v NullableLoadBalancerServiceUpdate2RequestMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLoadBalancerServiceUpdate2RequestMetadata) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
