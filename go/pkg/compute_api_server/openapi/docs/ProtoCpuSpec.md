# ProtoCpuSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Cores** | Pointer to **int32** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**ModelName** | Pointer to **string** |  | [optional] 
**Sockets** | Pointer to **int32** |  | [optional] 
**Threads** | Pointer to **int32** |  | [optional] 

## Methods

### NewProtoCpuSpec

`func NewProtoCpuSpec() *ProtoCpuSpec`

NewProtoCpuSpec instantiates a new ProtoCpuSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoCpuSpecWithDefaults

`func NewProtoCpuSpecWithDefaults() *ProtoCpuSpec`

NewProtoCpuSpecWithDefaults instantiates a new ProtoCpuSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCores

`func (o *ProtoCpuSpec) GetCores() int32`

GetCores returns the Cores field if non-nil, zero value otherwise.

### GetCoresOk

`func (o *ProtoCpuSpec) GetCoresOk() (*int32, bool)`

GetCoresOk returns a tuple with the Cores field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCores

`func (o *ProtoCpuSpec) SetCores(v int32)`

SetCores sets Cores field to given value.

### HasCores

`func (o *ProtoCpuSpec) HasCores() bool`

HasCores returns a boolean if a field has been set.

### GetId

`func (o *ProtoCpuSpec) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProtoCpuSpec) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProtoCpuSpec) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ProtoCpuSpec) HasId() bool`

HasId returns a boolean if a field has been set.

### GetModelName

`func (o *ProtoCpuSpec) GetModelName() string`

GetModelName returns the ModelName field if non-nil, zero value otherwise.

### GetModelNameOk

`func (o *ProtoCpuSpec) GetModelNameOk() (*string, bool)`

GetModelNameOk returns a tuple with the ModelName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModelName

`func (o *ProtoCpuSpec) SetModelName(v string)`

SetModelName sets ModelName field to given value.

### HasModelName

`func (o *ProtoCpuSpec) HasModelName() bool`

HasModelName returns a boolean if a field has been set.

### GetSockets

`func (o *ProtoCpuSpec) GetSockets() int32`

GetSockets returns the Sockets field if non-nil, zero value otherwise.

### GetSocketsOk

`func (o *ProtoCpuSpec) GetSocketsOk() (*int32, bool)`

GetSocketsOk returns a tuple with the Sockets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSockets

`func (o *ProtoCpuSpec) SetSockets(v int32)`

SetSockets sets Sockets field to given value.

### HasSockets

`func (o *ProtoCpuSpec) HasSockets() bool`

HasSockets returns a boolean if a field has been set.

### GetThreads

`func (o *ProtoCpuSpec) GetThreads() int32`

GetThreads returns the Threads field if non-nil, zero value otherwise.

### GetThreadsOk

`func (o *ProtoCpuSpec) GetThreadsOk() (*int32, bool)`

GetThreadsOk returns a tuple with the Threads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreads

`func (o *ProtoCpuSpec) SetThreads(v int32)`

SetThreads sets Threads field to given value.

### HasThreads

`func (o *ProtoCpuSpec) HasThreads() bool`

HasThreads returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


