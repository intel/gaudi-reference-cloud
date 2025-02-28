/*
Devcloud API

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 1.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// DevcloudV2EnvironmentConfigurePortAccessPut200Response struct for DevcloudV2EnvironmentConfigurePortAccessPut200Response
type DevcloudV2EnvironmentConfigurePortAccessPut200Response struct {
	// If task failed, contains a human-readable messages
	Message []string `json:"message,omitempty"`
	// Indicates if task succeeded
	Success *bool `json:"success,omitempty"`
}

// NewDevcloudV2EnvironmentConfigurePortAccessPut200Response instantiates a new DevcloudV2EnvironmentConfigurePortAccessPut200Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewDevcloudV2EnvironmentConfigurePortAccessPut200Response() *DevcloudV2EnvironmentConfigurePortAccessPut200Response {
	this := DevcloudV2EnvironmentConfigurePortAccessPut200Response{}
	return &this
}

// NewDevcloudV2EnvironmentConfigurePortAccessPut200ResponseWithDefaults instantiates a new DevcloudV2EnvironmentConfigurePortAccessPut200Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewDevcloudV2EnvironmentConfigurePortAccessPut200ResponseWithDefaults() *DevcloudV2EnvironmentConfigurePortAccessPut200Response {
	this := DevcloudV2EnvironmentConfigurePortAccessPut200Response{}
	return &this
}

// GetMessage returns the Message field value if set, zero value otherwise.
func (o *DevcloudV2EnvironmentConfigurePortAccessPut200Response) GetMessage() []string {
	if o == nil || isNil(o.Message) {
		var ret []string
		return ret
	}
	return o.Message
}

// GetMessageOk returns a tuple with the Message field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *DevcloudV2EnvironmentConfigurePortAccessPut200Response) GetMessageOk() ([]string, bool) {
	if o == nil || isNil(o.Message) {
		return nil, false
	}
	return o.Message, true
}

// HasMessage returns a boolean if a field has been set.
func (o *DevcloudV2EnvironmentConfigurePortAccessPut200Response) HasMessage() bool {
	if o != nil && !isNil(o.Message) {
		return true
	}

	return false
}

// SetMessage gets a reference to the given []string and assigns it to the Message field.
func (o *DevcloudV2EnvironmentConfigurePortAccessPut200Response) SetMessage(v []string) {
	o.Message = v
}

// GetSuccess returns the Success field value if set, zero value otherwise.
func (o *DevcloudV2EnvironmentConfigurePortAccessPut200Response) GetSuccess() bool {
	if o == nil || isNil(o.Success) {
		var ret bool
		return ret
	}
	return *o.Success
}

// GetSuccessOk returns a tuple with the Success field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *DevcloudV2EnvironmentConfigurePortAccessPut200Response) GetSuccessOk() (*bool, bool) {
	if o == nil || isNil(o.Success) {
		return nil, false
	}
	return o.Success, true
}

// HasSuccess returns a boolean if a field has been set.
func (o *DevcloudV2EnvironmentConfigurePortAccessPut200Response) HasSuccess() bool {
	if o != nil && !isNil(o.Success) {
		return true
	}

	return false
}

// SetSuccess gets a reference to the given bool and assigns it to the Success field.
func (o *DevcloudV2EnvironmentConfigurePortAccessPut200Response) SetSuccess(v bool) {
	o.Success = &v
}

func (o DevcloudV2EnvironmentConfigurePortAccessPut200Response) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Message) {
		toSerialize["message"] = o.Message
	}
	if !isNil(o.Success) {
		toSerialize["success"] = o.Success
	}
	return json.Marshal(toSerialize)
}

type NullableDevcloudV2EnvironmentConfigurePortAccessPut200Response struct {
	value *DevcloudV2EnvironmentConfigurePortAccessPut200Response
	isSet bool
}

func (v NullableDevcloudV2EnvironmentConfigurePortAccessPut200Response) Get() *DevcloudV2EnvironmentConfigurePortAccessPut200Response {
	return v.value
}

func (v *NullableDevcloudV2EnvironmentConfigurePortAccessPut200Response) Set(val *DevcloudV2EnvironmentConfigurePortAccessPut200Response) {
	v.value = val
	v.isSet = true
}

func (v NullableDevcloudV2EnvironmentConfigurePortAccessPut200Response) IsSet() bool {
	return v.isSet
}

func (v *NullableDevcloudV2EnvironmentConfigurePortAccessPut200Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableDevcloudV2EnvironmentConfigurePortAccessPut200Response(val *DevcloudV2EnvironmentConfigurePortAccessPut200Response) *NullableDevcloudV2EnvironmentConfigurePortAccessPut200Response {
	return &NullableDevcloudV2EnvironmentConfigurePortAccessPut200Response{value: val, isSet: true}
}

func (v NullableDevcloudV2EnvironmentConfigurePortAccessPut200Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableDevcloudV2EnvironmentConfigurePortAccessPut200Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
