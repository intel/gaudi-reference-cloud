# InstanceServiceUpdateRequestMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**ResourceVersion** | Pointer to **string** | If provided, the existing record must have this resourceVersion for the request to succeed. | [optional] 
**Labels** | Pointer to **map[string]string** | The entire set of labels will be replaced with these labels. Not implemented. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 

## Methods

### NewInstanceServiceUpdateRequestMetadata

`func NewInstanceServiceUpdateRequestMetadata() *InstanceServiceUpdateRequestMetadata`

NewInstanceServiceUpdateRequestMetadata instantiates a new InstanceServiceUpdateRequestMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstanceServiceUpdateRequestMetadataWithDefaults

`func NewInstanceServiceUpdateRequestMetadataWithDefaults() *InstanceServiceUpdateRequestMetadata`

NewInstanceServiceUpdateRequestMetadataWithDefaults instantiates a new InstanceServiceUpdateRequestMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *InstanceServiceUpdateRequestMetadata) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *InstanceServiceUpdateRequestMetadata) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *InstanceServiceUpdateRequestMetadata) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *InstanceServiceUpdateRequestMetadata) HasName() bool`

HasName returns a boolean if a field has been set.

### GetResourceVersion

`func (o *InstanceServiceUpdateRequestMetadata) GetResourceVersion() string`

GetResourceVersion returns the ResourceVersion field if non-nil, zero value otherwise.

### GetResourceVersionOk

`func (o *InstanceServiceUpdateRequestMetadata) GetResourceVersionOk() (*string, bool)`

GetResourceVersionOk returns a tuple with the ResourceVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceVersion

`func (o *InstanceServiceUpdateRequestMetadata) SetResourceVersion(v string)`

SetResourceVersion sets ResourceVersion field to given value.

### HasResourceVersion

`func (o *InstanceServiceUpdateRequestMetadata) HasResourceVersion() bool`

HasResourceVersion returns a boolean if a field has been set.

### GetLabels

`func (o *InstanceServiceUpdateRequestMetadata) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *InstanceServiceUpdateRequestMetadata) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *InstanceServiceUpdateRequestMetadata) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *InstanceServiceUpdateRequestMetadata) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetReserved1

`func (o *InstanceServiceUpdateRequestMetadata) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *InstanceServiceUpdateRequestMetadata) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *InstanceServiceUpdateRequestMetadata) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *InstanceServiceUpdateRequestMetadata) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


