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

// Headers struct for Headers
type Headers struct {
	Authorization *string `json:"Authorization,omitempty"`
}

// NewHeaders instantiates a new Headers object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewHeaders() *Headers {
	this := Headers{}
	return &this
}

// NewHeadersWithDefaults instantiates a new Headers object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewHeadersWithDefaults() *Headers {
	this := Headers{}
	return &this
}

// GetAuthorization returns the Authorization field value if set, zero value otherwise.
func (o *Headers) GetAuthorization() string {
	if o == nil || isNil(o.Authorization) {
		var ret string
		return ret
	}
	return *o.Authorization
}

// GetAuthorizationOk returns a tuple with the Authorization field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Headers) GetAuthorizationOk() (*string, bool) {
	if o == nil || isNil(o.Authorization) {
		return nil, false
	}
	return o.Authorization, true
}

// HasAuthorization returns a boolean if a field has been set.
func (o *Headers) HasAuthorization() bool {
	if o != nil && !isNil(o.Authorization) {
		return true
	}

	return false
}

// SetAuthorization gets a reference to the given string and assigns it to the Authorization field.
func (o *Headers) SetAuthorization(v string) {
	o.Authorization = &v
}

func (o Headers) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Authorization) {
		toSerialize["Authorization"] = o.Authorization
	}
	return json.Marshal(toSerialize)
}

type NullableHeaders struct {
	value *Headers
	isSet bool
}

func (v NullableHeaders) Get() *Headers {
	return v.value
}

func (v *NullableHeaders) Set(val *Headers) {
	v.value = val
	v.isSet = true
}

func (v NullableHeaders) IsSet() bool {
	return v.isSet
}

func (v *NullableHeaders) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableHeaders(val *Headers) *NullableHeaders {
	return &NullableHeaders{value: val, isSet: true}
}

func (v NullableHeaders) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableHeaders) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
