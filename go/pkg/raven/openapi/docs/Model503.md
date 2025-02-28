# Model503

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Error** | Pointer to **string** |  | [optional] [default to "Service Temporarily Unavailable"]
**Message** | Pointer to **[]string** |  | [optional] 
**Status** | Pointer to **bool** |  | [optional] [default to false]

## Methods

### NewModel503

`func NewModel503() *Model503`

NewModel503 instantiates a new Model503 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModel503WithDefaults

`func NewModel503WithDefaults() *Model503`

NewModel503WithDefaults instantiates a new Model503 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetError

`func (o *Model503) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *Model503) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *Model503) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *Model503) HasError() bool`

HasError returns a boolean if a field has been set.

### GetMessage

`func (o *Model503) GetMessage() []string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *Model503) GetMessageOk() (*[]string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *Model503) SetMessage(v []string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *Model503) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetStatus

`func (o *Model503) GetStatus() bool`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Model503) GetStatusOk() (*bool, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Model503) SetStatus(v bool)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Model503) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


