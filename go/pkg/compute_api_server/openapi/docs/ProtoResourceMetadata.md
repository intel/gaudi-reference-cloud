# ProtoResourceMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CloudAccountId** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**ResourceId** | Pointer to **string** | A globally unique UUID. | [optional] 
**Labels** | Pointer to **map[string]string** |  | [optional] 
**CreationTimestamp** | Pointer to **time.Time** |  | [optional] 
**AllowDelete** | Pointer to **bool** |  | [optional] 

## Methods

### NewProtoResourceMetadata

`func NewProtoResourceMetadata() *ProtoResourceMetadata`

NewProtoResourceMetadata instantiates a new ProtoResourceMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoResourceMetadataWithDefaults

`func NewProtoResourceMetadataWithDefaults() *ProtoResourceMetadata`

NewProtoResourceMetadataWithDefaults instantiates a new ProtoResourceMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCloudAccountId

`func (o *ProtoResourceMetadata) GetCloudAccountId() string`

GetCloudAccountId returns the CloudAccountId field if non-nil, zero value otherwise.

### GetCloudAccountIdOk

`func (o *ProtoResourceMetadata) GetCloudAccountIdOk() (*string, bool)`

GetCloudAccountIdOk returns a tuple with the CloudAccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloudAccountId

`func (o *ProtoResourceMetadata) SetCloudAccountId(v string)`

SetCloudAccountId sets CloudAccountId field to given value.

### HasCloudAccountId

`func (o *ProtoResourceMetadata) HasCloudAccountId() bool`

HasCloudAccountId returns a boolean if a field has been set.

### GetName

`func (o *ProtoResourceMetadata) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoResourceMetadata) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoResourceMetadata) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoResourceMetadata) HasName() bool`

HasName returns a boolean if a field has been set.

### GetResourceId

`func (o *ProtoResourceMetadata) GetResourceId() string`

GetResourceId returns the ResourceId field if non-nil, zero value otherwise.

### GetResourceIdOk

`func (o *ProtoResourceMetadata) GetResourceIdOk() (*string, bool)`

GetResourceIdOk returns a tuple with the ResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceId

`func (o *ProtoResourceMetadata) SetResourceId(v string)`

SetResourceId sets ResourceId field to given value.

### HasResourceId

`func (o *ProtoResourceMetadata) HasResourceId() bool`

HasResourceId returns a boolean if a field has been set.

### GetLabels

`func (o *ProtoResourceMetadata) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ProtoResourceMetadata) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ProtoResourceMetadata) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ProtoResourceMetadata) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetCreationTimestamp

`func (o *ProtoResourceMetadata) GetCreationTimestamp() time.Time`

GetCreationTimestamp returns the CreationTimestamp field if non-nil, zero value otherwise.

### GetCreationTimestampOk

`func (o *ProtoResourceMetadata) GetCreationTimestampOk() (*time.Time, bool)`

GetCreationTimestampOk returns a tuple with the CreationTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreationTimestamp

`func (o *ProtoResourceMetadata) SetCreationTimestamp(v time.Time)`

SetCreationTimestamp sets CreationTimestamp field to given value.

### HasCreationTimestamp

`func (o *ProtoResourceMetadata) HasCreationTimestamp() bool`

HasCreationTimestamp returns a boolean if a field has been set.

### GetAllowDelete

`func (o *ProtoResourceMetadata) GetAllowDelete() bool`

GetAllowDelete returns the AllowDelete field if non-nil, zero value otherwise.

### GetAllowDeleteOk

`func (o *ProtoResourceMetadata) GetAllowDeleteOk() (*bool, bool)`

GetAllowDeleteOk returns a tuple with the AllowDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowDelete

`func (o *ProtoResourceMetadata) SetAllowDelete(v bool)`

SetAllowDelete sets AllowDelete field to given value.

### HasAllowDelete

`func (o *ProtoResourceMetadata) HasAllowDelete() bool`

HasAllowDelete returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


