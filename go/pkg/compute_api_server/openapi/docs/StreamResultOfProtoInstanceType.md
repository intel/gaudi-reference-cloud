# StreamResultOfProtoInstanceType

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Result** | Pointer to [**ProtoInstanceType**](ProtoInstanceType.md) |  | [optional] 
**Error** | Pointer to [**RpcStatus**](RpcStatus.md) |  | [optional] 

## Methods

### NewStreamResultOfProtoInstanceType

`func NewStreamResultOfProtoInstanceType() *StreamResultOfProtoInstanceType`

NewStreamResultOfProtoInstanceType instantiates a new StreamResultOfProtoInstanceType object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStreamResultOfProtoInstanceTypeWithDefaults

`func NewStreamResultOfProtoInstanceTypeWithDefaults() *StreamResultOfProtoInstanceType`

NewStreamResultOfProtoInstanceTypeWithDefaults instantiates a new StreamResultOfProtoInstanceType object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResult

`func (o *StreamResultOfProtoInstanceType) GetResult() ProtoInstanceType`

GetResult returns the Result field if non-nil, zero value otherwise.

### GetResultOk

`func (o *StreamResultOfProtoInstanceType) GetResultOk() (*ProtoInstanceType, bool)`

GetResultOk returns a tuple with the Result field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResult

`func (o *StreamResultOfProtoInstanceType) SetResult(v ProtoInstanceType)`

SetResult sets Result field to given value.

### HasResult

`func (o *StreamResultOfProtoInstanceType) HasResult() bool`

HasResult returns a boolean if a field has been set.

### GetError

`func (o *StreamResultOfProtoInstanceType) GetError() RpcStatus`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *StreamResultOfProtoInstanceType) GetErrorOk() (*RpcStatus, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *StreamResultOfProtoInstanceType) SetError(v RpcStatus)`

SetError sets Error field to given value.

### HasError

`func (o *StreamResultOfProtoInstanceType) HasError() bool`

HasError returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


