# DevcloudPort

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** |  | [optional] 
**InterfaceMode** | Pointer to **string** |  | [optional] 
**LinkStatus** | Pointer to **string** |  | [optional] 
**Mode** | Pointer to **string** |  | [optional] 
**NativeVlan** | Pointer to [**DevcloudPortNativeVlan**](DevcloudPortNativeVlan.md) |  | [optional] 
**Port** | Pointer to **string** |  | [optional] 
**TrunkGroups** | Pointer to **string** |  | [optional] 
**UntaggedVlan** | Pointer to [**DevcloudPortNativeVlan**](DevcloudPortNativeVlan.md) |  | [optional] 

## Methods

### NewDevcloudPort

`func NewDevcloudPort() *DevcloudPort`

NewDevcloudPort instantiates a new DevcloudPort object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDevcloudPortWithDefaults

`func NewDevcloudPortWithDefaults() *DevcloudPort`

NewDevcloudPortWithDefaults instantiates a new DevcloudPort object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *DevcloudPort) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DevcloudPort) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DevcloudPort) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DevcloudPort) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetInterfaceMode

`func (o *DevcloudPort) GetInterfaceMode() string`

GetInterfaceMode returns the InterfaceMode field if non-nil, zero value otherwise.

### GetInterfaceModeOk

`func (o *DevcloudPort) GetInterfaceModeOk() (*string, bool)`

GetInterfaceModeOk returns a tuple with the InterfaceMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInterfaceMode

`func (o *DevcloudPort) SetInterfaceMode(v string)`

SetInterfaceMode sets InterfaceMode field to given value.

### HasInterfaceMode

`func (o *DevcloudPort) HasInterfaceMode() bool`

HasInterfaceMode returns a boolean if a field has been set.

### GetLinkStatus

`func (o *DevcloudPort) GetLinkStatus() string`

GetLinkStatus returns the LinkStatus field if non-nil, zero value otherwise.

### GetLinkStatusOk

`func (o *DevcloudPort) GetLinkStatusOk() (*string, bool)`

GetLinkStatusOk returns a tuple with the LinkStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLinkStatus

`func (o *DevcloudPort) SetLinkStatus(v string)`

SetLinkStatus sets LinkStatus field to given value.

### HasLinkStatus

`func (o *DevcloudPort) HasLinkStatus() bool`

HasLinkStatus returns a boolean if a field has been set.

### GetMode

`func (o *DevcloudPort) GetMode() string`

GetMode returns the Mode field if non-nil, zero value otherwise.

### GetModeOk

`func (o *DevcloudPort) GetModeOk() (*string, bool)`

GetModeOk returns a tuple with the Mode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMode

`func (o *DevcloudPort) SetMode(v string)`

SetMode sets Mode field to given value.

### HasMode

`func (o *DevcloudPort) HasMode() bool`

HasMode returns a boolean if a field has been set.

### GetNativeVlan

`func (o *DevcloudPort) GetNativeVlan() DevcloudPortNativeVlan`

GetNativeVlan returns the NativeVlan field if non-nil, zero value otherwise.

### GetNativeVlanOk

`func (o *DevcloudPort) GetNativeVlanOk() (*DevcloudPortNativeVlan, bool)`

GetNativeVlanOk returns a tuple with the NativeVlan field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNativeVlan

`func (o *DevcloudPort) SetNativeVlan(v DevcloudPortNativeVlan)`

SetNativeVlan sets NativeVlan field to given value.

### HasNativeVlan

`func (o *DevcloudPort) HasNativeVlan() bool`

HasNativeVlan returns a boolean if a field has been set.

### GetPort

`func (o *DevcloudPort) GetPort() string`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *DevcloudPort) GetPortOk() (*string, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *DevcloudPort) SetPort(v string)`

SetPort sets Port field to given value.

### HasPort

`func (o *DevcloudPort) HasPort() bool`

HasPort returns a boolean if a field has been set.

### GetTrunkGroups

`func (o *DevcloudPort) GetTrunkGroups() string`

GetTrunkGroups returns the TrunkGroups field if non-nil, zero value otherwise.

### GetTrunkGroupsOk

`func (o *DevcloudPort) GetTrunkGroupsOk() (*string, bool)`

GetTrunkGroupsOk returns a tuple with the TrunkGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrunkGroups

`func (o *DevcloudPort) SetTrunkGroups(v string)`

SetTrunkGroups sets TrunkGroups field to given value.

### HasTrunkGroups

`func (o *DevcloudPort) HasTrunkGroups() bool`

HasTrunkGroups returns a boolean if a field has been set.

### GetUntaggedVlan

`func (o *DevcloudPort) GetUntaggedVlan() DevcloudPortNativeVlan`

GetUntaggedVlan returns the UntaggedVlan field if non-nil, zero value otherwise.

### GetUntaggedVlanOk

`func (o *DevcloudPort) GetUntaggedVlanOk() (*DevcloudPortNativeVlan, bool)`

GetUntaggedVlanOk returns a tuple with the UntaggedVlan field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUntaggedVlan

`func (o *DevcloudPort) SetUntaggedVlan(v DevcloudPortNativeVlan)`

SetUntaggedVlan sets UntaggedVlan field to given value.

### HasUntaggedVlan

`func (o *DevcloudPort) HasUntaggedVlan() bool`

HasUntaggedVlan returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


