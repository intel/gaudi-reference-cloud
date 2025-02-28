# InstanceServiceUpdate2RequestMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ResourceId** | Pointer to **string** |  | [optional] 
**ResourceVersion** | Pointer to **string** | If provided, the existing record must have this resourceVersion for the request to succeed. | [optional] 
**Labels** | Pointer to **map[string]string** | The entire set of labels will be replaced with these labels. Not implemented. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 

## Methods

### NewInstanceServiceUpdate2RequestMetadata

`func NewInstanceServiceUpdate2RequestMetadata() *InstanceServiceUpdate2RequestMetadata`

NewInstanceServiceUpdate2RequestMetadata instantiates a new InstanceServiceUpdate2RequestMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstanceServiceUpdate2RequestMetadataWithDefaults

`func NewInstanceServiceUpdate2RequestMetadataWithDefaults() *InstanceServiceUpdate2RequestMetadata`

NewInstanceServiceUpdate2RequestMetadataWithDefaults instantiates a new InstanceServiceUpdate2RequestMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResourceId

`func (o *InstanceServiceUpdate2RequestMetadata) GetResourceId() string`

GetResourceId returns the ResourceId field if non-nil, zero value otherwise.

### GetResourceIdOk

`func (o *InstanceServiceUpdate2RequestMetadata) GetResourceIdOk() (*string, bool)`

GetResourceIdOk returns a tuple with the ResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceId

`func (o *InstanceServiceUpdate2RequestMetadata) SetResourceId(v string)`

SetResourceId sets ResourceId field to given value.

### HasResourceId

`func (o *InstanceServiceUpdate2RequestMetadata) HasResourceId() bool`

HasResourceId returns a boolean if a field has been set.

### GetResourceVersion

`func (o *InstanceServiceUpdate2RequestMetadata) GetResourceVersion() string`

GetResourceVersion returns the ResourceVersion field if non-nil, zero value otherwise.

### GetResourceVersionOk

`func (o *InstanceServiceUpdate2RequestMetadata) GetResourceVersionOk() (*string, bool)`

GetResourceVersionOk returns a tuple with the ResourceVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceVersion

`func (o *InstanceServiceUpdate2RequestMetadata) SetResourceVersion(v string)`

SetResourceVersion sets ResourceVersion field to given value.

### HasResourceVersion

`func (o *InstanceServiceUpdate2RequestMetadata) HasResourceVersion() bool`

HasResourceVersion returns a boolean if a field has been set.

### GetLabels

`func (o *InstanceServiceUpdate2RequestMetadata) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *InstanceServiceUpdate2RequestMetadata) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *InstanceServiceUpdate2RequestMetadata) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *InstanceServiceUpdate2RequestMetadata) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetReserved1

`func (o *InstanceServiceUpdate2RequestMetadata) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *InstanceServiceUpdate2RequestMetadata) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *InstanceServiceUpdate2RequestMetadata) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *InstanceServiceUpdate2RequestMetadata) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


