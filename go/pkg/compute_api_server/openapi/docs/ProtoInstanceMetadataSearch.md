# ProtoInstanceMetadataSearch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CloudAccountId** | Pointer to **string** |  | [optional] 
**Labels** | Pointer to **map[string]string** | If not empty, only return instances that have these key/value pairs. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 
**InstanceGroup** | Pointer to **string** |  | [optional] 
**InstanceGroupFilter** | Pointer to [**ProtoSearchFilterCriteria**](ProtoSearchFilterCriteria.md) |  | [optional] [default to DEFAULT]

## Methods

### NewProtoInstanceMetadataSearch

`func NewProtoInstanceMetadataSearch() *ProtoInstanceMetadataSearch`

NewProtoInstanceMetadataSearch instantiates a new ProtoInstanceMetadataSearch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceMetadataSearchWithDefaults

`func NewProtoInstanceMetadataSearchWithDefaults() *ProtoInstanceMetadataSearch`

NewProtoInstanceMetadataSearchWithDefaults instantiates a new ProtoInstanceMetadataSearch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCloudAccountId

`func (o *ProtoInstanceMetadataSearch) GetCloudAccountId() string`

GetCloudAccountId returns the CloudAccountId field if non-nil, zero value otherwise.

### GetCloudAccountIdOk

`func (o *ProtoInstanceMetadataSearch) GetCloudAccountIdOk() (*string, bool)`

GetCloudAccountIdOk returns a tuple with the CloudAccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloudAccountId

`func (o *ProtoInstanceMetadataSearch) SetCloudAccountId(v string)`

SetCloudAccountId sets CloudAccountId field to given value.

### HasCloudAccountId

`func (o *ProtoInstanceMetadataSearch) HasCloudAccountId() bool`

HasCloudAccountId returns a boolean if a field has been set.

### GetLabels

`func (o *ProtoInstanceMetadataSearch) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ProtoInstanceMetadataSearch) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ProtoInstanceMetadataSearch) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ProtoInstanceMetadataSearch) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetReserved1

`func (o *ProtoInstanceMetadataSearch) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *ProtoInstanceMetadataSearch) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *ProtoInstanceMetadataSearch) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *ProtoInstanceMetadataSearch) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.

### GetInstanceGroup

`func (o *ProtoInstanceMetadataSearch) GetInstanceGroup() string`

GetInstanceGroup returns the InstanceGroup field if non-nil, zero value otherwise.

### GetInstanceGroupOk

`func (o *ProtoInstanceMetadataSearch) GetInstanceGroupOk() (*string, bool)`

GetInstanceGroupOk returns a tuple with the InstanceGroup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceGroup

`func (o *ProtoInstanceMetadataSearch) SetInstanceGroup(v string)`

SetInstanceGroup sets InstanceGroup field to given value.

### HasInstanceGroup

`func (o *ProtoInstanceMetadataSearch) HasInstanceGroup() bool`

HasInstanceGroup returns a boolean if a field has been set.

### GetInstanceGroupFilter

`func (o *ProtoInstanceMetadataSearch) GetInstanceGroupFilter() ProtoSearchFilterCriteria`

GetInstanceGroupFilter returns the InstanceGroupFilter field if non-nil, zero value otherwise.

### GetInstanceGroupFilterOk

`func (o *ProtoInstanceMetadataSearch) GetInstanceGroupFilterOk() (*ProtoSearchFilterCriteria, bool)`

GetInstanceGroupFilterOk returns a tuple with the InstanceGroupFilter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceGroupFilter

`func (o *ProtoInstanceMetadataSearch) SetInstanceGroupFilter(v ProtoSearchFilterCriteria)`

SetInstanceGroupFilter sets InstanceGroupFilter field to given value.

### HasInstanceGroupFilter

`func (o *ProtoInstanceMetadataSearch) HasInstanceGroupFilter() bool`

HasInstanceGroupFilter returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


