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

// ProtoLoadBalancerListenerStatus struct for ProtoLoadBalancerListenerStatus
type ProtoLoadBalancerListenerStatus struct {
	Name        *string                             `json:"name,omitempty"`
	VipID       *int32                              `json:"vipID,omitempty"`
	Message     *string                             `json:"message,omitempty"`
	PoolMembers []ProtoLoadBalancerPoolStatusMember `json:"poolMembers,omitempty"`
	PoolID      *int32                              `json:"poolID,omitempty"`
	State       *string                             `json:"state,omitempty"`
	Port        *int32                              `json:"port,omitempty"`
}

// NewProtoLoadBalancerListenerStatus instantiates a new ProtoLoadBalancerListenerStatus object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewProtoLoadBalancerListenerStatus() *ProtoLoadBalancerListenerStatus {
	this := ProtoLoadBalancerListenerStatus{}
	return &this
}

// NewProtoLoadBalancerListenerStatusWithDefaults instantiates a new ProtoLoadBalancerListenerStatus object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewProtoLoadBalancerListenerStatusWithDefaults() *ProtoLoadBalancerListenerStatus {
	this := ProtoLoadBalancerListenerStatus{}
	return &this
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *ProtoLoadBalancerListenerStatus) GetName() string {
	if o == nil || isNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerListenerStatus) GetNameOk() (*string, bool) {
	if o == nil || isNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *ProtoLoadBalancerListenerStatus) HasName() bool {
	if o != nil && !isNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *ProtoLoadBalancerListenerStatus) SetName(v string) {
	o.Name = &v
}

// GetVipID returns the VipID field value if set, zero value otherwise.
func (o *ProtoLoadBalancerListenerStatus) GetVipID() int32 {
	if o == nil || isNil(o.VipID) {
		var ret int32
		return ret
	}
	return *o.VipID
}

// GetVipIDOk returns a tuple with the VipID field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerListenerStatus) GetVipIDOk() (*int32, bool) {
	if o == nil || isNil(o.VipID) {
		return nil, false
	}
	return o.VipID, true
}

// HasVipID returns a boolean if a field has been set.
func (o *ProtoLoadBalancerListenerStatus) HasVipID() bool {
	if o != nil && !isNil(o.VipID) {
		return true
	}

	return false
}

// SetVipID gets a reference to the given int32 and assigns it to the VipID field.
func (o *ProtoLoadBalancerListenerStatus) SetVipID(v int32) {
	o.VipID = &v
}

// GetMessage returns the Message field value if set, zero value otherwise.
func (o *ProtoLoadBalancerListenerStatus) GetMessage() string {
	if o == nil || isNil(o.Message) {
		var ret string
		return ret
	}
	return *o.Message
}

// GetMessageOk returns a tuple with the Message field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerListenerStatus) GetMessageOk() (*string, bool) {
	if o == nil || isNil(o.Message) {
		return nil, false
	}
	return o.Message, true
}

// HasMessage returns a boolean if a field has been set.
func (o *ProtoLoadBalancerListenerStatus) HasMessage() bool {
	if o != nil && !isNil(o.Message) {
		return true
	}

	return false
}

// SetMessage gets a reference to the given string and assigns it to the Message field.
func (o *ProtoLoadBalancerListenerStatus) SetMessage(v string) {
	o.Message = &v
}

// GetPoolMembers returns the PoolMembers field value if set, zero value otherwise.
func (o *ProtoLoadBalancerListenerStatus) GetPoolMembers() []ProtoLoadBalancerPoolStatusMember {
	if o == nil || isNil(o.PoolMembers) {
		var ret []ProtoLoadBalancerPoolStatusMember
		return ret
	}
	return o.PoolMembers
}

// GetPoolMembersOk returns a tuple with the PoolMembers field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerListenerStatus) GetPoolMembersOk() ([]ProtoLoadBalancerPoolStatusMember, bool) {
	if o == nil || isNil(o.PoolMembers) {
		return nil, false
	}
	return o.PoolMembers, true
}

// HasPoolMembers returns a boolean if a field has been set.
func (o *ProtoLoadBalancerListenerStatus) HasPoolMembers() bool {
	if o != nil && !isNil(o.PoolMembers) {
		return true
	}

	return false
}

// SetPoolMembers gets a reference to the given []ProtoLoadBalancerPoolStatusMember and assigns it to the PoolMembers field.
func (o *ProtoLoadBalancerListenerStatus) SetPoolMembers(v []ProtoLoadBalancerPoolStatusMember) {
	o.PoolMembers = v
}

// GetPoolID returns the PoolID field value if set, zero value otherwise.
func (o *ProtoLoadBalancerListenerStatus) GetPoolID() int32 {
	if o == nil || isNil(o.PoolID) {
		var ret int32
		return ret
	}
	return *o.PoolID
}

// GetPoolIDOk returns a tuple with the PoolID field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerListenerStatus) GetPoolIDOk() (*int32, bool) {
	if o == nil || isNil(o.PoolID) {
		return nil, false
	}
	return o.PoolID, true
}

// HasPoolID returns a boolean if a field has been set.
func (o *ProtoLoadBalancerListenerStatus) HasPoolID() bool {
	if o != nil && !isNil(o.PoolID) {
		return true
	}

	return false
}

// SetPoolID gets a reference to the given int32 and assigns it to the PoolID field.
func (o *ProtoLoadBalancerListenerStatus) SetPoolID(v int32) {
	o.PoolID = &v
}

// GetState returns the State field value if set, zero value otherwise.
func (o *ProtoLoadBalancerListenerStatus) GetState() string {
	if o == nil || isNil(o.State) {
		var ret string
		return ret
	}
	return *o.State
}

// GetStateOk returns a tuple with the State field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerListenerStatus) GetStateOk() (*string, bool) {
	if o == nil || isNil(o.State) {
		return nil, false
	}
	return o.State, true
}

// HasState returns a boolean if a field has been set.
func (o *ProtoLoadBalancerListenerStatus) HasState() bool {
	if o != nil && !isNil(o.State) {
		return true
	}

	return false
}

// SetState gets a reference to the given string and assigns it to the State field.
func (o *ProtoLoadBalancerListenerStatus) SetState(v string) {
	o.State = &v
}

// GetPort returns the Port field value if set, zero value otherwise.
func (o *ProtoLoadBalancerListenerStatus) GetPort() int32 {
	if o == nil || isNil(o.Port) {
		var ret int32
		return ret
	}
	return *o.Port
}

// GetPortOk returns a tuple with the Port field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ProtoLoadBalancerListenerStatus) GetPortOk() (*int32, bool) {
	if o == nil || isNil(o.Port) {
		return nil, false
	}
	return o.Port, true
}

// HasPort returns a boolean if a field has been set.
func (o *ProtoLoadBalancerListenerStatus) HasPort() bool {
	if o != nil && !isNil(o.Port) {
		return true
	}

	return false
}

// SetPort gets a reference to the given int32 and assigns it to the Port field.
func (o *ProtoLoadBalancerListenerStatus) SetPort(v int32) {
	o.Port = &v
}

func (o ProtoLoadBalancerListenerStatus) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	if !isNil(o.VipID) {
		toSerialize["vipID"] = o.VipID
	}
	if !isNil(o.Message) {
		toSerialize["message"] = o.Message
	}
	if !isNil(o.PoolMembers) {
		toSerialize["poolMembers"] = o.PoolMembers
	}
	if !isNil(o.PoolID) {
		toSerialize["poolID"] = o.PoolID
	}
	if !isNil(o.State) {
		toSerialize["state"] = o.State
	}
	if !isNil(o.Port) {
		toSerialize["port"] = o.Port
	}
	return json.Marshal(toSerialize)
}

type NullableProtoLoadBalancerListenerStatus struct {
	value *ProtoLoadBalancerListenerStatus
	isSet bool
}

func (v NullableProtoLoadBalancerListenerStatus) Get() *ProtoLoadBalancerListenerStatus {
	return v.value
}

func (v *NullableProtoLoadBalancerListenerStatus) Set(val *ProtoLoadBalancerListenerStatus) {
	v.value = val
	v.isSet = true
}

func (v NullableProtoLoadBalancerListenerStatus) IsSet() bool {
	return v.isSet
}

func (v *NullableProtoLoadBalancerListenerStatus) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableProtoLoadBalancerListenerStatus(val *ProtoLoadBalancerListenerStatus) *NullableProtoLoadBalancerListenerStatus {
	return &NullableProtoLoadBalancerListenerStatus{value: val, isSet: true}
}

func (v NullableProtoLoadBalancerListenerStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableProtoLoadBalancerListenerStatus) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
