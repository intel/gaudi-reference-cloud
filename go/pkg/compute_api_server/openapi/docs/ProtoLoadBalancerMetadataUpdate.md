# ProtoLoadBalancerMetadataUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CloudAccountId** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**ResourceId** | Pointer to **string** |  | [optional] 
**ResourceVersion** | Pointer to **string** | If provided, the existing record must have this resourceVersion for the request to succeed. | [optional] 
**Labels** | Pointer to **map[string]string** | Map of string keys and values that can be used to organize and categorize load balancers. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 

## Methods

### NewProtoLoadBalancerMetadataUpdate

`func NewProtoLoadBalancerMetadataUpdate() *ProtoLoadBalancerMetadataUpdate`

NewProtoLoadBalancerMetadataUpdate instantiates a new ProtoLoadBalancerMetadataUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerMetadataUpdateWithDefaults

`func NewProtoLoadBalancerMetadataUpdateWithDefaults() *ProtoLoadBalancerMetadataUpdate`

NewProtoLoadBalancerMetadataUpdateWithDefaults instantiates a new ProtoLoadBalancerMetadataUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCloudAccountId

`func (o *ProtoLoadBalancerMetadataUpdate) GetCloudAccountId() string`

GetCloudAccountId returns the CloudAccountId field if non-nil, zero value otherwise.

### GetCloudAccountIdOk

`func (o *ProtoLoadBalancerMetadataUpdate) GetCloudAccountIdOk() (*string, bool)`

GetCloudAccountIdOk returns a tuple with the CloudAccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloudAccountId

`func (o *ProtoLoadBalancerMetadataUpdate) SetCloudAccountId(v string)`

SetCloudAccountId sets CloudAccountId field to given value.

### HasCloudAccountId

`func (o *ProtoLoadBalancerMetadataUpdate) HasCloudAccountId() bool`

HasCloudAccountId returns a boolean if a field has been set.

### GetName

`func (o *ProtoLoadBalancerMetadataUpdate) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoLoadBalancerMetadataUpdate) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoLoadBalancerMetadataUpdate) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoLoadBalancerMetadataUpdate) HasName() bool`

HasName returns a boolean if a field has been set.

### GetResourceId

`func (o *ProtoLoadBalancerMetadataUpdate) GetResourceId() string`

GetResourceId returns the ResourceId field if non-nil, zero value otherwise.

### GetResourceIdOk

`func (o *ProtoLoadBalancerMetadataUpdate) GetResourceIdOk() (*string, bool)`

GetResourceIdOk returns a tuple with the ResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceId

`func (o *ProtoLoadBalancerMetadataUpdate) SetResourceId(v string)`

SetResourceId sets ResourceId field to given value.

### HasResourceId

`func (o *ProtoLoadBalancerMetadataUpdate) HasResourceId() bool`

HasResourceId returns a boolean if a field has been set.

### GetResourceVersion

`func (o *ProtoLoadBalancerMetadataUpdate) GetResourceVersion() string`

GetResourceVersion returns the ResourceVersion field if non-nil, zero value otherwise.

### GetResourceVersionOk

`func (o *ProtoLoadBalancerMetadataUpdate) GetResourceVersionOk() (*string, bool)`

GetResourceVersionOk returns a tuple with the ResourceVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceVersion

`func (o *ProtoLoadBalancerMetadataUpdate) SetResourceVersion(v string)`

SetResourceVersion sets ResourceVersion field to given value.

### HasResourceVersion

`func (o *ProtoLoadBalancerMetadataUpdate) HasResourceVersion() bool`

HasResourceVersion returns a boolean if a field has been set.

### GetLabels

`func (o *ProtoLoadBalancerMetadataUpdate) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ProtoLoadBalancerMetadataUpdate) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ProtoLoadBalancerMetadataUpdate) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ProtoLoadBalancerMetadataUpdate) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetReserved1

`func (o *ProtoLoadBalancerMetadataUpdate) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *ProtoLoadBalancerMetadataUpdate) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *ProtoLoadBalancerMetadataUpdate) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *ProtoLoadBalancerMetadataUpdate) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


