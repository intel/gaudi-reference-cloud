# ProtoInstanceGroupSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InstanceSpec** | Pointer to [**ProtoInstanceSpec**](ProtoInstanceSpec.md) |  | [optional] 
**InstanceCount** | Pointer to **int32** |  | [optional] 

## Methods

### NewProtoInstanceGroupSpec

`func NewProtoInstanceGroupSpec() *ProtoInstanceGroupSpec`

NewProtoInstanceGroupSpec instantiates a new ProtoInstanceGroupSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceGroupSpecWithDefaults

`func NewProtoInstanceGroupSpecWithDefaults() *ProtoInstanceGroupSpec`

NewProtoInstanceGroupSpecWithDefaults instantiates a new ProtoInstanceGroupSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInstanceSpec

`func (o *ProtoInstanceGroupSpec) GetInstanceSpec() ProtoInstanceSpec`

GetInstanceSpec returns the InstanceSpec field if non-nil, zero value otherwise.

### GetInstanceSpecOk

`func (o *ProtoInstanceGroupSpec) GetInstanceSpecOk() (*ProtoInstanceSpec, bool)`

GetInstanceSpecOk returns a tuple with the InstanceSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceSpec

`func (o *ProtoInstanceGroupSpec) SetInstanceSpec(v ProtoInstanceSpec)`

SetInstanceSpec sets InstanceSpec field to given value.

### HasInstanceSpec

`func (o *ProtoInstanceGroupSpec) HasInstanceSpec() bool`

HasInstanceSpec returns a boolean if a field has been set.

### GetInstanceCount

`func (o *ProtoInstanceGroupSpec) GetInstanceCount() int32`

GetInstanceCount returns the InstanceCount field if non-nil, zero value otherwise.

### GetInstanceCountOk

`func (o *ProtoInstanceGroupSpec) GetInstanceCountOk() (*int32, bool)`

GetInstanceCountOk returns a tuple with the InstanceCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceCount

`func (o *ProtoInstanceGroupSpec) SetInstanceCount(v int32)`

SetInstanceCount sets InstanceCount field to given value.

### HasInstanceCount

`func (o *ProtoInstanceGroupSpec) HasInstanceCount() bool`

HasInstanceCount returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


