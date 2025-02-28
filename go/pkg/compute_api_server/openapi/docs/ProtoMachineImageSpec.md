# ProtoMachineImageSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DisplayName** | Pointer to **string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**UserName** | Pointer to **string** |  | [optional] 
**Icon** | Pointer to **string** |  | [optional] 
**InstanceCategories** | Pointer to [**[]ProtoInstanceCategory**](ProtoInstanceCategory.md) | If not empty, this machine image is only compatible with the specified instance categories. | [optional] 
**InstanceTypes** | Pointer to **[]string** | If not empty, this machine image is only compatible with the specified instance types. | [optional] 
**Md5sum** | Pointer to **string** |  | [optional] 
**Sha256sum** | Pointer to **string** |  | [optional] 
**Sha512sum** | Pointer to **string** |  | [optional] 
**Labels** | Pointer to **map[string]string** |  | [optional] 
**ImageCategories** | Pointer to **[]string** |  | [optional] 
**Components** | Pointer to [**[]ProtoMachineImageComponent**](ProtoMachineImageComponent.md) |  | [optional] 
**Hidden** | Pointer to **bool** | If true, this machine image will not be returned by the MachineImageService.Search method but it can still be used to launch instances. | [optional] 
**VirtualSizeBytes** | Pointer to **string** |  | [optional] 

## Methods

### NewProtoMachineImageSpec

`func NewProtoMachineImageSpec() *ProtoMachineImageSpec`

NewProtoMachineImageSpec instantiates a new ProtoMachineImageSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoMachineImageSpecWithDefaults

`func NewProtoMachineImageSpecWithDefaults() *ProtoMachineImageSpec`

NewProtoMachineImageSpecWithDefaults instantiates a new ProtoMachineImageSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDisplayName

`func (o *ProtoMachineImageSpec) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *ProtoMachineImageSpec) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *ProtoMachineImageSpec) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *ProtoMachineImageSpec) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetDescription

`func (o *ProtoMachineImageSpec) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ProtoMachineImageSpec) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ProtoMachineImageSpec) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ProtoMachineImageSpec) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetUserName

`func (o *ProtoMachineImageSpec) GetUserName() string`

GetUserName returns the UserName field if non-nil, zero value otherwise.

### GetUserNameOk

`func (o *ProtoMachineImageSpec) GetUserNameOk() (*string, bool)`

GetUserNameOk returns a tuple with the UserName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserName

`func (o *ProtoMachineImageSpec) SetUserName(v string)`

SetUserName sets UserName field to given value.

### HasUserName

`func (o *ProtoMachineImageSpec) HasUserName() bool`

HasUserName returns a boolean if a field has been set.

### GetIcon

`func (o *ProtoMachineImageSpec) GetIcon() string`

GetIcon returns the Icon field if non-nil, zero value otherwise.

### GetIconOk

`func (o *ProtoMachineImageSpec) GetIconOk() (*string, bool)`

GetIconOk returns a tuple with the Icon field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIcon

`func (o *ProtoMachineImageSpec) SetIcon(v string)`

SetIcon sets Icon field to given value.

### HasIcon

`func (o *ProtoMachineImageSpec) HasIcon() bool`

HasIcon returns a boolean if a field has been set.

### GetInstanceCategories

`func (o *ProtoMachineImageSpec) GetInstanceCategories() []ProtoInstanceCategory`

GetInstanceCategories returns the InstanceCategories field if non-nil, zero value otherwise.

### GetInstanceCategoriesOk

`func (o *ProtoMachineImageSpec) GetInstanceCategoriesOk() (*[]ProtoInstanceCategory, bool)`

GetInstanceCategoriesOk returns a tuple with the InstanceCategories field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceCategories

`func (o *ProtoMachineImageSpec) SetInstanceCategories(v []ProtoInstanceCategory)`

SetInstanceCategories sets InstanceCategories field to given value.

### HasInstanceCategories

`func (o *ProtoMachineImageSpec) HasInstanceCategories() bool`

HasInstanceCategories returns a boolean if a field has been set.

### GetInstanceTypes

`func (o *ProtoMachineImageSpec) GetInstanceTypes() []string`

GetInstanceTypes returns the InstanceTypes field if non-nil, zero value otherwise.

### GetInstanceTypesOk

`func (o *ProtoMachineImageSpec) GetInstanceTypesOk() (*[]string, bool)`

GetInstanceTypesOk returns a tuple with the InstanceTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceTypes

`func (o *ProtoMachineImageSpec) SetInstanceTypes(v []string)`

SetInstanceTypes sets InstanceTypes field to given value.

### HasInstanceTypes

`func (o *ProtoMachineImageSpec) HasInstanceTypes() bool`

HasInstanceTypes returns a boolean if a field has been set.

### GetMd5sum

`func (o *ProtoMachineImageSpec) GetMd5sum() string`

GetMd5sum returns the Md5sum field if non-nil, zero value otherwise.

### GetMd5sumOk

`func (o *ProtoMachineImageSpec) GetMd5sumOk() (*string, bool)`

GetMd5sumOk returns a tuple with the Md5sum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMd5sum

`func (o *ProtoMachineImageSpec) SetMd5sum(v string)`

SetMd5sum sets Md5sum field to given value.

### HasMd5sum

`func (o *ProtoMachineImageSpec) HasMd5sum() bool`

HasMd5sum returns a boolean if a field has been set.

### GetSha256sum

`func (o *ProtoMachineImageSpec) GetSha256sum() string`

GetSha256sum returns the Sha256sum field if non-nil, zero value otherwise.

### GetSha256sumOk

`func (o *ProtoMachineImageSpec) GetSha256sumOk() (*string, bool)`

GetSha256sumOk returns a tuple with the Sha256sum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSha256sum

`func (o *ProtoMachineImageSpec) SetSha256sum(v string)`

SetSha256sum sets Sha256sum field to given value.

### HasSha256sum

`func (o *ProtoMachineImageSpec) HasSha256sum() bool`

HasSha256sum returns a boolean if a field has been set.

### GetSha512sum

`func (o *ProtoMachineImageSpec) GetSha512sum() string`

GetSha512sum returns the Sha512sum field if non-nil, zero value otherwise.

### GetSha512sumOk

`func (o *ProtoMachineImageSpec) GetSha512sumOk() (*string, bool)`

GetSha512sumOk returns a tuple with the Sha512sum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSha512sum

`func (o *ProtoMachineImageSpec) SetSha512sum(v string)`

SetSha512sum sets Sha512sum field to given value.

### HasSha512sum

`func (o *ProtoMachineImageSpec) HasSha512sum() bool`

HasSha512sum returns a boolean if a field has been set.

### GetLabels

`func (o *ProtoMachineImageSpec) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ProtoMachineImageSpec) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ProtoMachineImageSpec) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ProtoMachineImageSpec) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetImageCategories

`func (o *ProtoMachineImageSpec) GetImageCategories() []string`

GetImageCategories returns the ImageCategories field if non-nil, zero value otherwise.

### GetImageCategoriesOk

`func (o *ProtoMachineImageSpec) GetImageCategoriesOk() (*[]string, bool)`

GetImageCategoriesOk returns a tuple with the ImageCategories field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageCategories

`func (o *ProtoMachineImageSpec) SetImageCategories(v []string)`

SetImageCategories sets ImageCategories field to given value.

### HasImageCategories

`func (o *ProtoMachineImageSpec) HasImageCategories() bool`

HasImageCategories returns a boolean if a field has been set.

### GetComponents

`func (o *ProtoMachineImageSpec) GetComponents() []ProtoMachineImageComponent`

GetComponents returns the Components field if non-nil, zero value otherwise.

### GetComponentsOk

`func (o *ProtoMachineImageSpec) GetComponentsOk() (*[]ProtoMachineImageComponent, bool)`

GetComponentsOk returns a tuple with the Components field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComponents

`func (o *ProtoMachineImageSpec) SetComponents(v []ProtoMachineImageComponent)`

SetComponents sets Components field to given value.

### HasComponents

`func (o *ProtoMachineImageSpec) HasComponents() bool`

HasComponents returns a boolean if a field has been set.

### GetHidden

`func (o *ProtoMachineImageSpec) GetHidden() bool`

GetHidden returns the Hidden field if non-nil, zero value otherwise.

### GetHiddenOk

`func (o *ProtoMachineImageSpec) GetHiddenOk() (*bool, bool)`

GetHiddenOk returns a tuple with the Hidden field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHidden

`func (o *ProtoMachineImageSpec) SetHidden(v bool)`

SetHidden sets Hidden field to given value.

### HasHidden

`func (o *ProtoMachineImageSpec) HasHidden() bool`

HasHidden returns a boolean if a field has been set.

### GetVirtualSizeBytes

`func (o *ProtoMachineImageSpec) GetVirtualSizeBytes() string`

GetVirtualSizeBytes returns the VirtualSizeBytes field if non-nil, zero value otherwise.

### GetVirtualSizeBytesOk

`func (o *ProtoMachineImageSpec) GetVirtualSizeBytesOk() (*string, bool)`

GetVirtualSizeBytesOk returns a tuple with the VirtualSizeBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVirtualSizeBytes

`func (o *ProtoMachineImageSpec) SetVirtualSizeBytes(v string)`

SetVirtualSizeBytes sets VirtualSizeBytes field to given value.

### HasVirtualSizeBytes

`func (o *ProtoMachineImageSpec) HasVirtualSizeBytes() bool`

HasVirtualSizeBytes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


