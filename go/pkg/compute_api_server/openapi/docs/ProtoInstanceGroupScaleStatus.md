# ProtoInstanceGroupScaleStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CurrentCount** | Pointer to **int32** |  | [optional] 
**DesiredCount** | Pointer to **int32** |  | [optional] 
**ReadyCount** | Pointer to **int32** | The number of instances with a phase of Ready. | [optional] 
**CurrentMembers** | Pointer to **[]string** | The names of existing and non-deleting instances in the instanceGroup. | [optional] 
**NewMembers** | Pointer to **[]string** | The names of newly created instances in the instanceGroup. | [optional] 
**ReadyMembers** | Pointer to **[]string** | The names of instances with a phase of Ready. | [optional] 

## Methods

### NewProtoInstanceGroupScaleStatus

`func NewProtoInstanceGroupScaleStatus() *ProtoInstanceGroupScaleStatus`

NewProtoInstanceGroupScaleStatus instantiates a new ProtoInstanceGroupScaleStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceGroupScaleStatusWithDefaults

`func NewProtoInstanceGroupScaleStatusWithDefaults() *ProtoInstanceGroupScaleStatus`

NewProtoInstanceGroupScaleStatusWithDefaults instantiates a new ProtoInstanceGroupScaleStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCurrentCount

`func (o *ProtoInstanceGroupScaleStatus) GetCurrentCount() int32`

GetCurrentCount returns the CurrentCount field if non-nil, zero value otherwise.

### GetCurrentCountOk

`func (o *ProtoInstanceGroupScaleStatus) GetCurrentCountOk() (*int32, bool)`

GetCurrentCountOk returns a tuple with the CurrentCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentCount

`func (o *ProtoInstanceGroupScaleStatus) SetCurrentCount(v int32)`

SetCurrentCount sets CurrentCount field to given value.

### HasCurrentCount

`func (o *ProtoInstanceGroupScaleStatus) HasCurrentCount() bool`

HasCurrentCount returns a boolean if a field has been set.

### GetDesiredCount

`func (o *ProtoInstanceGroupScaleStatus) GetDesiredCount() int32`

GetDesiredCount returns the DesiredCount field if non-nil, zero value otherwise.

### GetDesiredCountOk

`func (o *ProtoInstanceGroupScaleStatus) GetDesiredCountOk() (*int32, bool)`

GetDesiredCountOk returns a tuple with the DesiredCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredCount

`func (o *ProtoInstanceGroupScaleStatus) SetDesiredCount(v int32)`

SetDesiredCount sets DesiredCount field to given value.

### HasDesiredCount

`func (o *ProtoInstanceGroupScaleStatus) HasDesiredCount() bool`

HasDesiredCount returns a boolean if a field has been set.

### GetReadyCount

`func (o *ProtoInstanceGroupScaleStatus) GetReadyCount() int32`

GetReadyCount returns the ReadyCount field if non-nil, zero value otherwise.

### GetReadyCountOk

`func (o *ProtoInstanceGroupScaleStatus) GetReadyCountOk() (*int32, bool)`

GetReadyCountOk returns a tuple with the ReadyCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadyCount

`func (o *ProtoInstanceGroupScaleStatus) SetReadyCount(v int32)`

SetReadyCount sets ReadyCount field to given value.

### HasReadyCount

`func (o *ProtoInstanceGroupScaleStatus) HasReadyCount() bool`

HasReadyCount returns a boolean if a field has been set.

### GetCurrentMembers

`func (o *ProtoInstanceGroupScaleStatus) GetCurrentMembers() []string`

GetCurrentMembers returns the CurrentMembers field if non-nil, zero value otherwise.

### GetCurrentMembersOk

`func (o *ProtoInstanceGroupScaleStatus) GetCurrentMembersOk() (*[]string, bool)`

GetCurrentMembersOk returns a tuple with the CurrentMembers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentMembers

`func (o *ProtoInstanceGroupScaleStatus) SetCurrentMembers(v []string)`

SetCurrentMembers sets CurrentMembers field to given value.

### HasCurrentMembers

`func (o *ProtoInstanceGroupScaleStatus) HasCurrentMembers() bool`

HasCurrentMembers returns a boolean if a field has been set.

### GetNewMembers

`func (o *ProtoInstanceGroupScaleStatus) GetNewMembers() []string`

GetNewMembers returns the NewMembers field if non-nil, zero value otherwise.

### GetNewMembersOk

`func (o *ProtoInstanceGroupScaleStatus) GetNewMembersOk() (*[]string, bool)`

GetNewMembersOk returns a tuple with the NewMembers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewMembers

`func (o *ProtoInstanceGroupScaleStatus) SetNewMembers(v []string)`

SetNewMembers sets NewMembers field to given value.

### HasNewMembers

`func (o *ProtoInstanceGroupScaleStatus) HasNewMembers() bool`

HasNewMembers returns a boolean if a field has been set.

### GetReadyMembers

`func (o *ProtoInstanceGroupScaleStatus) GetReadyMembers() []string`

GetReadyMembers returns the ReadyMembers field if non-nil, zero value otherwise.

### GetReadyMembersOk

`func (o *ProtoInstanceGroupScaleStatus) GetReadyMembersOk() (*[]string, bool)`

GetReadyMembersOk returns a tuple with the ReadyMembers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadyMembers

`func (o *ProtoInstanceGroupScaleStatus) SetReadyMembers(v []string)`

SetReadyMembers sets ReadyMembers field to given value.

### HasReadyMembers

`func (o *ProtoInstanceGroupScaleStatus) HasReadyMembers() bool`

HasReadyMembers returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


