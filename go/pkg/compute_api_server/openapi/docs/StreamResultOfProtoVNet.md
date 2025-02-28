# StreamResultOfProtoVNet

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Result** | Pointer to [**ProtoVNet**](ProtoVNet.md) |  | [optional] 
**Error** | Pointer to [**RpcStatus**](RpcStatus.md) |  | [optional] 

## Methods

### NewStreamResultOfProtoVNet

`func NewStreamResultOfProtoVNet() *StreamResultOfProtoVNet`

NewStreamResultOfProtoVNet instantiates a new StreamResultOfProtoVNet object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStreamResultOfProtoVNetWithDefaults

`func NewStreamResultOfProtoVNetWithDefaults() *StreamResultOfProtoVNet`

NewStreamResultOfProtoVNetWithDefaults instantiates a new StreamResultOfProtoVNet object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResult

`func (o *StreamResultOfProtoVNet) GetResult() ProtoVNet`

GetResult returns the Result field if non-nil, zero value otherwise.

### GetResultOk

`func (o *StreamResultOfProtoVNet) GetResultOk() (*ProtoVNet, bool)`

GetResultOk returns a tuple with the Result field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResult

`func (o *StreamResultOfProtoVNet) SetResult(v ProtoVNet)`

SetResult sets Result field to given value.

### HasResult

`func (o *StreamResultOfProtoVNet) HasResult() bool`

HasResult returns a boolean if a field has been set.

### GetError

`func (o *StreamResultOfProtoVNet) GetError() RpcStatus`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *StreamResultOfProtoVNet) GetErrorOk() (*RpcStatus, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *StreamResultOfProtoVNet) SetError(v RpcStatus)`

SetError sets Error field to given value.

### HasError

`func (o *StreamResultOfProtoVNet) HasError() bool`

HasError returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


