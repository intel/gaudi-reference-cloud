# ProtoInstanceTypeSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**DisplayName** | Pointer to **string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**InstanceCategory** | Pointer to [**ProtoInstanceCategory**](ProtoInstanceCategory.md) |  | [optional] [default to VIRTUAL_MACHINE]
**Cpu** | Pointer to [**ProtoCpuSpec**](ProtoCpuSpec.md) |  | [optional] 
**Memory** | Pointer to [**ProtoMemorySpec**](ProtoMemorySpec.md) |  | [optional] 
**Disks** | Pointer to [**[]ProtoDiskSpec**](ProtoDiskSpec.md) |  | [optional] 
**Gpu** | Pointer to [**ProtoGpuSpec**](ProtoGpuSpec.md) |  | [optional] 
**HbmMode** | Pointer to **string** |  | [optional] 

## Methods

### NewProtoInstanceTypeSpec

`func NewProtoInstanceTypeSpec() *ProtoInstanceTypeSpec`

NewProtoInstanceTypeSpec instantiates a new ProtoInstanceTypeSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceTypeSpecWithDefaults

`func NewProtoInstanceTypeSpecWithDefaults() *ProtoInstanceTypeSpec`

NewProtoInstanceTypeSpecWithDefaults instantiates a new ProtoInstanceTypeSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ProtoInstanceTypeSpec) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoInstanceTypeSpec) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoInstanceTypeSpec) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoInstanceTypeSpec) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDisplayName

`func (o *ProtoInstanceTypeSpec) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *ProtoInstanceTypeSpec) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *ProtoInstanceTypeSpec) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *ProtoInstanceTypeSpec) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetDescription

`func (o *ProtoInstanceTypeSpec) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ProtoInstanceTypeSpec) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ProtoInstanceTypeSpec) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ProtoInstanceTypeSpec) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetInstanceCategory

`func (o *ProtoInstanceTypeSpec) GetInstanceCategory() ProtoInstanceCategory`

GetInstanceCategory returns the InstanceCategory field if non-nil, zero value otherwise.

### GetInstanceCategoryOk

`func (o *ProtoInstanceTypeSpec) GetInstanceCategoryOk() (*ProtoInstanceCategory, bool)`

GetInstanceCategoryOk returns a tuple with the InstanceCategory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceCategory

`func (o *ProtoInstanceTypeSpec) SetInstanceCategory(v ProtoInstanceCategory)`

SetInstanceCategory sets InstanceCategory field to given value.

### HasInstanceCategory

`func (o *ProtoInstanceTypeSpec) HasInstanceCategory() bool`

HasInstanceCategory returns a boolean if a field has been set.

### GetCpu

`func (o *ProtoInstanceTypeSpec) GetCpu() ProtoCpuSpec`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *ProtoInstanceTypeSpec) GetCpuOk() (*ProtoCpuSpec, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *ProtoInstanceTypeSpec) SetCpu(v ProtoCpuSpec)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *ProtoInstanceTypeSpec) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *ProtoInstanceTypeSpec) GetMemory() ProtoMemorySpec`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *ProtoInstanceTypeSpec) GetMemoryOk() (*ProtoMemorySpec, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *ProtoInstanceTypeSpec) SetMemory(v ProtoMemorySpec)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *ProtoInstanceTypeSpec) HasMemory() bool`

HasMemory returns a boolean if a field has been set.

### GetDisks

`func (o *ProtoInstanceTypeSpec) GetDisks() []ProtoDiskSpec`

GetDisks returns the Disks field if non-nil, zero value otherwise.

### GetDisksOk

`func (o *ProtoInstanceTypeSpec) GetDisksOk() (*[]ProtoDiskSpec, bool)`

GetDisksOk returns a tuple with the Disks field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisks

`func (o *ProtoInstanceTypeSpec) SetDisks(v []ProtoDiskSpec)`

SetDisks sets Disks field to given value.

### HasDisks

`func (o *ProtoInstanceTypeSpec) HasDisks() bool`

HasDisks returns a boolean if a field has been set.

### GetGpu

`func (o *ProtoInstanceTypeSpec) GetGpu() ProtoGpuSpec`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *ProtoInstanceTypeSpec) GetGpuOk() (*ProtoGpuSpec, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *ProtoInstanceTypeSpec) SetGpu(v ProtoGpuSpec)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *ProtoInstanceTypeSpec) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### GetHbmMode

`func (o *ProtoInstanceTypeSpec) GetHbmMode() string`

GetHbmMode returns the HbmMode field if non-nil, zero value otherwise.

### GetHbmModeOk

`func (o *ProtoInstanceTypeSpec) GetHbmModeOk() (*string, bool)`

GetHbmModeOk returns a tuple with the HbmMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHbmMode

`func (o *ProtoInstanceTypeSpec) SetHbmMode(v string)`

SetHbmMode sets HbmMode field to given value.

### HasHbmMode

`func (o *ProtoInstanceTypeSpec) HasHbmMode() bool`

HasHbmMode returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


