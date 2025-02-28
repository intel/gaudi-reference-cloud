# ProtoInstanceSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AvailabilityZone** | Pointer to **string** | Not implemented. | [optional] 
**InstanceType** | Pointer to **string** | The name of an InstanceType. | [optional] 
**MachineImage** | Pointer to **string** | The name of a MachineImage. Not implemented. | [optional] 
**RunStrategy** | Pointer to [**ProtoRunStrategy**](ProtoRunStrategy.md) |  | [optional] [default to RERUN_ON_FAILURE]
**SshPublicKeyNames** | Pointer to **[]string** | The name of a previously stored SSH public key. Users can use the corresponding SSH private key to SSH to this instance. | [optional] 
**Interfaces** | Pointer to [**[]ProtoNetworkInterface**](ProtoNetworkInterface.md) | Not implemented. | [optional] 
**TopologySpreadConstraints** | Pointer to [**[]ProtoTopologySpreadConstraints**](ProtoTopologySpreadConstraints.md) | This controls how instances are spread across the failure domains within the availability zone. This can help to achieve high availability. If this contains at least one key/value pair in matchLabels, then instances that have all of these key/value pairs will be placed evenly across failure domains. | [optional] 
**UserData** | Pointer to **string** |  | [optional] 
**InstanceGroup** | Pointer to **string** | If not empty, this instance is part of the named instance group. | [optional] 
**QuickConnectEnabled** | Pointer to [**ProtoTriState**](ProtoTriState.md) |  | [optional] [default to UNDEFINED]

## Methods

### NewProtoInstanceSpec

`func NewProtoInstanceSpec() *ProtoInstanceSpec`

NewProtoInstanceSpec instantiates a new ProtoInstanceSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceSpecWithDefaults

`func NewProtoInstanceSpecWithDefaults() *ProtoInstanceSpec`

NewProtoInstanceSpecWithDefaults instantiates a new ProtoInstanceSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAvailabilityZone

`func (o *ProtoInstanceSpec) GetAvailabilityZone() string`

GetAvailabilityZone returns the AvailabilityZone field if non-nil, zero value otherwise.

### GetAvailabilityZoneOk

`func (o *ProtoInstanceSpec) GetAvailabilityZoneOk() (*string, bool)`

GetAvailabilityZoneOk returns a tuple with the AvailabilityZone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvailabilityZone

`func (o *ProtoInstanceSpec) SetAvailabilityZone(v string)`

SetAvailabilityZone sets AvailabilityZone field to given value.

### HasAvailabilityZone

`func (o *ProtoInstanceSpec) HasAvailabilityZone() bool`

HasAvailabilityZone returns a boolean if a field has been set.

### GetInstanceType

`func (o *ProtoInstanceSpec) GetInstanceType() string`

GetInstanceType returns the InstanceType field if non-nil, zero value otherwise.

### GetInstanceTypeOk

`func (o *ProtoInstanceSpec) GetInstanceTypeOk() (*string, bool)`

GetInstanceTypeOk returns a tuple with the InstanceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceType

`func (o *ProtoInstanceSpec) SetInstanceType(v string)`

SetInstanceType sets InstanceType field to given value.

### HasInstanceType

`func (o *ProtoInstanceSpec) HasInstanceType() bool`

HasInstanceType returns a boolean if a field has been set.

### GetMachineImage

`func (o *ProtoInstanceSpec) GetMachineImage() string`

GetMachineImage returns the MachineImage field if non-nil, zero value otherwise.

### GetMachineImageOk

`func (o *ProtoInstanceSpec) GetMachineImageOk() (*string, bool)`

GetMachineImageOk returns a tuple with the MachineImage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMachineImage

`func (o *ProtoInstanceSpec) SetMachineImage(v string)`

SetMachineImage sets MachineImage field to given value.

### HasMachineImage

`func (o *ProtoInstanceSpec) HasMachineImage() bool`

HasMachineImage returns a boolean if a field has been set.

### GetRunStrategy

`func (o *ProtoInstanceSpec) GetRunStrategy() ProtoRunStrategy`

GetRunStrategy returns the RunStrategy field if non-nil, zero value otherwise.

### GetRunStrategyOk

`func (o *ProtoInstanceSpec) GetRunStrategyOk() (*ProtoRunStrategy, bool)`

GetRunStrategyOk returns a tuple with the RunStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunStrategy

`func (o *ProtoInstanceSpec) SetRunStrategy(v ProtoRunStrategy)`

SetRunStrategy sets RunStrategy field to given value.

### HasRunStrategy

`func (o *ProtoInstanceSpec) HasRunStrategy() bool`

HasRunStrategy returns a boolean if a field has been set.

### GetSshPublicKeyNames

`func (o *ProtoInstanceSpec) GetSshPublicKeyNames() []string`

GetSshPublicKeyNames returns the SshPublicKeyNames field if non-nil, zero value otherwise.

### GetSshPublicKeyNamesOk

`func (o *ProtoInstanceSpec) GetSshPublicKeyNamesOk() (*[]string, bool)`

GetSshPublicKeyNamesOk returns a tuple with the SshPublicKeyNames field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSshPublicKeyNames

`func (o *ProtoInstanceSpec) SetSshPublicKeyNames(v []string)`

SetSshPublicKeyNames sets SshPublicKeyNames field to given value.

### HasSshPublicKeyNames

`func (o *ProtoInstanceSpec) HasSshPublicKeyNames() bool`

HasSshPublicKeyNames returns a boolean if a field has been set.

### GetInterfaces

`func (o *ProtoInstanceSpec) GetInterfaces() []ProtoNetworkInterface`

GetInterfaces returns the Interfaces field if non-nil, zero value otherwise.

### GetInterfacesOk

`func (o *ProtoInstanceSpec) GetInterfacesOk() (*[]ProtoNetworkInterface, bool)`

GetInterfacesOk returns a tuple with the Interfaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInterfaces

`func (o *ProtoInstanceSpec) SetInterfaces(v []ProtoNetworkInterface)`

SetInterfaces sets Interfaces field to given value.

### HasInterfaces

`func (o *ProtoInstanceSpec) HasInterfaces() bool`

HasInterfaces returns a boolean if a field has been set.

### GetTopologySpreadConstraints

`func (o *ProtoInstanceSpec) GetTopologySpreadConstraints() []ProtoTopologySpreadConstraints`

GetTopologySpreadConstraints returns the TopologySpreadConstraints field if non-nil, zero value otherwise.

### GetTopologySpreadConstraintsOk

`func (o *ProtoInstanceSpec) GetTopologySpreadConstraintsOk() (*[]ProtoTopologySpreadConstraints, bool)`

GetTopologySpreadConstraintsOk returns a tuple with the TopologySpreadConstraints field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopologySpreadConstraints

`func (o *ProtoInstanceSpec) SetTopologySpreadConstraints(v []ProtoTopologySpreadConstraints)`

SetTopologySpreadConstraints sets TopologySpreadConstraints field to given value.

### HasTopologySpreadConstraints

`func (o *ProtoInstanceSpec) HasTopologySpreadConstraints() bool`

HasTopologySpreadConstraints returns a boolean if a field has been set.

### GetUserData

`func (o *ProtoInstanceSpec) GetUserData() string`

GetUserData returns the UserData field if non-nil, zero value otherwise.

### GetUserDataOk

`func (o *ProtoInstanceSpec) GetUserDataOk() (*string, bool)`

GetUserDataOk returns a tuple with the UserData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserData

`func (o *ProtoInstanceSpec) SetUserData(v string)`

SetUserData sets UserData field to given value.

### HasUserData

`func (o *ProtoInstanceSpec) HasUserData() bool`

HasUserData returns a boolean if a field has been set.

### GetInstanceGroup

`func (o *ProtoInstanceSpec) GetInstanceGroup() string`

GetInstanceGroup returns the InstanceGroup field if non-nil, zero value otherwise.

### GetInstanceGroupOk

`func (o *ProtoInstanceSpec) GetInstanceGroupOk() (*string, bool)`

GetInstanceGroupOk returns a tuple with the InstanceGroup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceGroup

`func (o *ProtoInstanceSpec) SetInstanceGroup(v string)`

SetInstanceGroup sets InstanceGroup field to given value.

### HasInstanceGroup

`func (o *ProtoInstanceSpec) HasInstanceGroup() bool`

HasInstanceGroup returns a boolean if a field has been set.

### GetQuickConnectEnabled

`func (o *ProtoInstanceSpec) GetQuickConnectEnabled() ProtoTriState`

GetQuickConnectEnabled returns the QuickConnectEnabled field if non-nil, zero value otherwise.

### GetQuickConnectEnabledOk

`func (o *ProtoInstanceSpec) GetQuickConnectEnabledOk() (*ProtoTriState, bool)`

GetQuickConnectEnabledOk returns a tuple with the QuickConnectEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuickConnectEnabled

`func (o *ProtoInstanceSpec) SetQuickConnectEnabled(v ProtoTriState)`

SetQuickConnectEnabled sets QuickConnectEnabled field to given value.

### HasQuickConnectEnabled

`func (o *ProtoInstanceSpec) HasQuickConnectEnabled() bool`

HasQuickConnectEnabled returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


