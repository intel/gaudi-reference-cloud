# StreamResultOfProtoSshPublicKey

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Result** | Pointer to [**ProtoSshPublicKey**](ProtoSshPublicKey.md) |  | [optional] 
**Error** | Pointer to [**RpcStatus**](RpcStatus.md) |  | [optional] 

## Methods

### NewStreamResultOfProtoSshPublicKey

`func NewStreamResultOfProtoSshPublicKey() *StreamResultOfProtoSshPublicKey`

NewStreamResultOfProtoSshPublicKey instantiates a new StreamResultOfProtoSshPublicKey object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStreamResultOfProtoSshPublicKeyWithDefaults

`func NewStreamResultOfProtoSshPublicKeyWithDefaults() *StreamResultOfProtoSshPublicKey`

NewStreamResultOfProtoSshPublicKeyWithDefaults instantiates a new StreamResultOfProtoSshPublicKey object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResult

`func (o *StreamResultOfProtoSshPublicKey) GetResult() ProtoSshPublicKey`

GetResult returns the Result field if non-nil, zero value otherwise.

### GetResultOk

`func (o *StreamResultOfProtoSshPublicKey) GetResultOk() (*ProtoSshPublicKey, bool)`

GetResultOk returns a tuple with the Result field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResult

`func (o *StreamResultOfProtoSshPublicKey) SetResult(v ProtoSshPublicKey)`

SetResult sets Result field to given value.

### HasResult

`func (o *StreamResultOfProtoSshPublicKey) HasResult() bool`

HasResult returns a boolean if a field has been set.

### GetError

`func (o *StreamResultOfProtoSshPublicKey) GetError() RpcStatus`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *StreamResultOfProtoSshPublicKey) GetErrorOk() (*RpcStatus, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *StreamResultOfProtoSshPublicKey) SetError(v RpcStatus)`

SetError sets Error field to given value.

### HasError

`func (o *StreamResultOfProtoSshPublicKey) HasError() bool`

HasError returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


