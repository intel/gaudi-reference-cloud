# ProtoLoadBalancerMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CloudAccountId** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**ResourceId** | Pointer to **string** |  | [optional] 
**ResourceVersion** | Pointer to **string** | resourceVersion can be provided with Update and Delete for concurrency control. | [optional] 
**Labels** | Pointer to **map[string]string** | Map of string keys and values that can be used to organize and categorize load balancers. | [optional] 
**CreationTimestamp** | Pointer to **time.Time** | Not implemented. | [optional] 
**DeletionTimestamp** | Pointer to **time.Time** | Timestamp when resource was requested to be deleted. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 

## Methods

### NewProtoLoadBalancerMetadata

`func NewProtoLoadBalancerMetadata() *ProtoLoadBalancerMetadata`

NewProtoLoadBalancerMetadata instantiates a new ProtoLoadBalancerMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerMetadataWithDefaults

`func NewProtoLoadBalancerMetadataWithDefaults() *ProtoLoadBalancerMetadata`

NewProtoLoadBalancerMetadataWithDefaults instantiates a new ProtoLoadBalancerMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCloudAccountId

`func (o *ProtoLoadBalancerMetadata) GetCloudAccountId() string`

GetCloudAccountId returns the CloudAccountId field if non-nil, zero value otherwise.

### GetCloudAccountIdOk

`func (o *ProtoLoadBalancerMetadata) GetCloudAccountIdOk() (*string, bool)`

GetCloudAccountIdOk returns a tuple with the CloudAccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloudAccountId

`func (o *ProtoLoadBalancerMetadata) SetCloudAccountId(v string)`

SetCloudAccountId sets CloudAccountId field to given value.

### HasCloudAccountId

`func (o *ProtoLoadBalancerMetadata) HasCloudAccountId() bool`

HasCloudAccountId returns a boolean if a field has been set.

### GetName

`func (o *ProtoLoadBalancerMetadata) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoLoadBalancerMetadata) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoLoadBalancerMetadata) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoLoadBalancerMetadata) HasName() bool`

HasName returns a boolean if a field has been set.

### GetResourceId

`func (o *ProtoLoadBalancerMetadata) GetResourceId() string`

GetResourceId returns the ResourceId field if non-nil, zero value otherwise.

### GetResourceIdOk

`func (o *ProtoLoadBalancerMetadata) GetResourceIdOk() (*string, bool)`

GetResourceIdOk returns a tuple with the ResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceId

`func (o *ProtoLoadBalancerMetadata) SetResourceId(v string)`

SetResourceId sets ResourceId field to given value.

### HasResourceId

`func (o *ProtoLoadBalancerMetadata) HasResourceId() bool`

HasResourceId returns a boolean if a field has been set.

### GetResourceVersion

`func (o *ProtoLoadBalancerMetadata) GetResourceVersion() string`

GetResourceVersion returns the ResourceVersion field if non-nil, zero value otherwise.

### GetResourceVersionOk

`func (o *ProtoLoadBalancerMetadata) GetResourceVersionOk() (*string, bool)`

GetResourceVersionOk returns a tuple with the ResourceVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceVersion

`func (o *ProtoLoadBalancerMetadata) SetResourceVersion(v string)`

SetResourceVersion sets ResourceVersion field to given value.

### HasResourceVersion

`func (o *ProtoLoadBalancerMetadata) HasResourceVersion() bool`

HasResourceVersion returns a boolean if a field has been set.

### GetLabels

`func (o *ProtoLoadBalancerMetadata) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ProtoLoadBalancerMetadata) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ProtoLoadBalancerMetadata) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ProtoLoadBalancerMetadata) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetCreationTimestamp

`func (o *ProtoLoadBalancerMetadata) GetCreationTimestamp() time.Time`

GetCreationTimestamp returns the CreationTimestamp field if non-nil, zero value otherwise.

### GetCreationTimestampOk

`func (o *ProtoLoadBalancerMetadata) GetCreationTimestampOk() (*time.Time, bool)`

GetCreationTimestampOk returns a tuple with the CreationTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreationTimestamp

`func (o *ProtoLoadBalancerMetadata) SetCreationTimestamp(v time.Time)`

SetCreationTimestamp sets CreationTimestamp field to given value.

### HasCreationTimestamp

`func (o *ProtoLoadBalancerMetadata) HasCreationTimestamp() bool`

HasCreationTimestamp returns a boolean if a field has been set.

### GetDeletionTimestamp

`func (o *ProtoLoadBalancerMetadata) GetDeletionTimestamp() time.Time`

GetDeletionTimestamp returns the DeletionTimestamp field if non-nil, zero value otherwise.

### GetDeletionTimestampOk

`func (o *ProtoLoadBalancerMetadata) GetDeletionTimestampOk() (*time.Time, bool)`

GetDeletionTimestampOk returns a tuple with the DeletionTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletionTimestamp

`func (o *ProtoLoadBalancerMetadata) SetDeletionTimestamp(v time.Time)`

SetDeletionTimestamp sets DeletionTimestamp field to given value.

### HasDeletionTimestamp

`func (o *ProtoLoadBalancerMetadata) HasDeletionTimestamp() bool`

HasDeletionTimestamp returns a boolean if a field has been set.

### GetReserved1

`func (o *ProtoLoadBalancerMetadata) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *ProtoLoadBalancerMetadata) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *ProtoLoadBalancerMetadata) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *ProtoLoadBalancerMetadata) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


