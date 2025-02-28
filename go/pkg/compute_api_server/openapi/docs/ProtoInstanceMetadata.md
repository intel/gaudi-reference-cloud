# ProtoInstanceMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CloudAccountId** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**ResourceId** | Pointer to **string** |  | [optional] 
**ResourceVersion** | Pointer to **string** | resourceVersion can be provided with Update and Delete for concurrency control. | [optional] 
**Labels** | Pointer to **map[string]string** | Map of string keys and values that can be used to organize and categorize instances. This is also used by TopologySpreadConstraints. | [optional] 
**CreationTimestamp** | Pointer to **time.Time** | Not implemented. | [optional] 
**DeletionTimestamp** | Pointer to **time.Time** | Timestamp when resource was requested to be deleted. | [optional] 

## Methods

### NewProtoInstanceMetadata

`func NewProtoInstanceMetadata() *ProtoInstanceMetadata`

NewProtoInstanceMetadata instantiates a new ProtoInstanceMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceMetadataWithDefaults

`func NewProtoInstanceMetadataWithDefaults() *ProtoInstanceMetadata`

NewProtoInstanceMetadataWithDefaults instantiates a new ProtoInstanceMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCloudAccountId

`func (o *ProtoInstanceMetadata) GetCloudAccountId() string`

GetCloudAccountId returns the CloudAccountId field if non-nil, zero value otherwise.

### GetCloudAccountIdOk

`func (o *ProtoInstanceMetadata) GetCloudAccountIdOk() (*string, bool)`

GetCloudAccountIdOk returns a tuple with the CloudAccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloudAccountId

`func (o *ProtoInstanceMetadata) SetCloudAccountId(v string)`

SetCloudAccountId sets CloudAccountId field to given value.

### HasCloudAccountId

`func (o *ProtoInstanceMetadata) HasCloudAccountId() bool`

HasCloudAccountId returns a boolean if a field has been set.

### GetName

`func (o *ProtoInstanceMetadata) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoInstanceMetadata) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoInstanceMetadata) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoInstanceMetadata) HasName() bool`

HasName returns a boolean if a field has been set.

### GetResourceId

`func (o *ProtoInstanceMetadata) GetResourceId() string`

GetResourceId returns the ResourceId field if non-nil, zero value otherwise.

### GetResourceIdOk

`func (o *ProtoInstanceMetadata) GetResourceIdOk() (*string, bool)`

GetResourceIdOk returns a tuple with the ResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceId

`func (o *ProtoInstanceMetadata) SetResourceId(v string)`

SetResourceId sets ResourceId field to given value.

### HasResourceId

`func (o *ProtoInstanceMetadata) HasResourceId() bool`

HasResourceId returns a boolean if a field has been set.

### GetResourceVersion

`func (o *ProtoInstanceMetadata) GetResourceVersion() string`

GetResourceVersion returns the ResourceVersion field if non-nil, zero value otherwise.

### GetResourceVersionOk

`func (o *ProtoInstanceMetadata) GetResourceVersionOk() (*string, bool)`

GetResourceVersionOk returns a tuple with the ResourceVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceVersion

`func (o *ProtoInstanceMetadata) SetResourceVersion(v string)`

SetResourceVersion sets ResourceVersion field to given value.

### HasResourceVersion

`func (o *ProtoInstanceMetadata) HasResourceVersion() bool`

HasResourceVersion returns a boolean if a field has been set.

### GetLabels

`func (o *ProtoInstanceMetadata) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ProtoInstanceMetadata) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ProtoInstanceMetadata) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ProtoInstanceMetadata) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetCreationTimestamp

`func (o *ProtoInstanceMetadata) GetCreationTimestamp() time.Time`

GetCreationTimestamp returns the CreationTimestamp field if non-nil, zero value otherwise.

### GetCreationTimestampOk

`func (o *ProtoInstanceMetadata) GetCreationTimestampOk() (*time.Time, bool)`

GetCreationTimestampOk returns a tuple with the CreationTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreationTimestamp

`func (o *ProtoInstanceMetadata) SetCreationTimestamp(v time.Time)`

SetCreationTimestamp sets CreationTimestamp field to given value.

### HasCreationTimestamp

`func (o *ProtoInstanceMetadata) HasCreationTimestamp() bool`

HasCreationTimestamp returns a boolean if a field has been set.

### GetDeletionTimestamp

`func (o *ProtoInstanceMetadata) GetDeletionTimestamp() time.Time`

GetDeletionTimestamp returns the DeletionTimestamp field if non-nil, zero value otherwise.

### GetDeletionTimestampOk

`func (o *ProtoInstanceMetadata) GetDeletionTimestampOk() (*time.Time, bool)`

GetDeletionTimestampOk returns a tuple with the DeletionTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletionTimestamp

`func (o *ProtoInstanceMetadata) SetDeletionTimestamp(v time.Time)`

SetDeletionTimestamp sets DeletionTimestamp field to given value.

### HasDeletionTimestamp

`func (o *ProtoInstanceMetadata) HasDeletionTimestamp() bool`

HasDeletionTimestamp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


