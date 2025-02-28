# InstanceServiceSearch2RequestMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Labels** | Pointer to **map[string]string** | If not empty, only return instances that have these key/value pairs. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 
**InstanceGroup** | Pointer to **string** |  | [optional] 
**InstanceGroupFilter** | Pointer to [**ProtoSearchFilterCriteria**](ProtoSearchFilterCriteria.md) |  | [optional] [default to DEFAULT]

## Methods

### NewInstanceServiceSearch2RequestMetadata

`func NewInstanceServiceSearch2RequestMetadata() *InstanceServiceSearch2RequestMetadata`

NewInstanceServiceSearch2RequestMetadata instantiates a new InstanceServiceSearch2RequestMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstanceServiceSearch2RequestMetadataWithDefaults

`func NewInstanceServiceSearch2RequestMetadataWithDefaults() *InstanceServiceSearch2RequestMetadata`

NewInstanceServiceSearch2RequestMetadataWithDefaults instantiates a new InstanceServiceSearch2RequestMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLabels

`func (o *InstanceServiceSearch2RequestMetadata) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *InstanceServiceSearch2RequestMetadata) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *InstanceServiceSearch2RequestMetadata) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *InstanceServiceSearch2RequestMetadata) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetReserved1

`func (o *InstanceServiceSearch2RequestMetadata) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *InstanceServiceSearch2RequestMetadata) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *InstanceServiceSearch2RequestMetadata) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *InstanceServiceSearch2RequestMetadata) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.

### GetInstanceGroup

`func (o *InstanceServiceSearch2RequestMetadata) GetInstanceGroup() string`

GetInstanceGroup returns the InstanceGroup field if non-nil, zero value otherwise.

### GetInstanceGroupOk

`func (o *InstanceServiceSearch2RequestMetadata) GetInstanceGroupOk() (*string, bool)`

GetInstanceGroupOk returns a tuple with the InstanceGroup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceGroup

`func (o *InstanceServiceSearch2RequestMetadata) SetInstanceGroup(v string)`

SetInstanceGroup sets InstanceGroup field to given value.

### HasInstanceGroup

`func (o *InstanceServiceSearch2RequestMetadata) HasInstanceGroup() bool`

HasInstanceGroup returns a boolean if a field has been set.

### GetInstanceGroupFilter

`func (o *InstanceServiceSearch2RequestMetadata) GetInstanceGroupFilter() ProtoSearchFilterCriteria`

GetInstanceGroupFilter returns the InstanceGroupFilter field if non-nil, zero value otherwise.

### GetInstanceGroupFilterOk

`func (o *InstanceServiceSearch2RequestMetadata) GetInstanceGroupFilterOk() (*ProtoSearchFilterCriteria, bool)`

GetInstanceGroupFilterOk returns a tuple with the InstanceGroupFilter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceGroupFilter

`func (o *InstanceServiceSearch2RequestMetadata) SetInstanceGroupFilter(v ProtoSearchFilterCriteria)`

SetInstanceGroupFilter sets InstanceGroupFilter field to given value.

### HasInstanceGroupFilter

`func (o *InstanceServiceSearch2RequestMetadata) HasInstanceGroupFilter() bool`

HasInstanceGroupFilter returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


