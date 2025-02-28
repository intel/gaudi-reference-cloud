# ProtoInstanceGroupStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ReadyCount** | Pointer to **int32** | The number of instances with a phase of Ready. The instance group is Ready when this equals InstanceGroupSpec.instanceCount. | [optional] 

## Methods

### NewProtoInstanceGroupStatus

`func NewProtoInstanceGroupStatus() *ProtoInstanceGroupStatus`

NewProtoInstanceGroupStatus instantiates a new ProtoInstanceGroupStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceGroupStatusWithDefaults

`func NewProtoInstanceGroupStatusWithDefaults() *ProtoInstanceGroupStatus`

NewProtoInstanceGroupStatusWithDefaults instantiates a new ProtoInstanceGroupStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetReadyCount

`func (o *ProtoInstanceGroupStatus) GetReadyCount() int32`

GetReadyCount returns the ReadyCount field if non-nil, zero value otherwise.

### GetReadyCountOk

`func (o *ProtoInstanceGroupStatus) GetReadyCountOk() (*int32, bool)`

GetReadyCountOk returns a tuple with the ReadyCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadyCount

`func (o *ProtoInstanceGroupStatus) SetReadyCount(v int32)`

SetReadyCount sets ReadyCount field to given value.

### HasReadyCount

`func (o *ProtoInstanceGroupStatus) HasReadyCount() bool`

HasReadyCount returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


