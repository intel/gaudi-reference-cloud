# ProtoLoadBalancerMetadataReference

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CloudAccountId** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**ResourceId** | Pointer to **string** |  | [optional] 
**ResourceVersion** | Pointer to **string** | If provided, the existing record must have this resourceVersion for the request to succeed. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 

## Methods

### NewProtoLoadBalancerMetadataReference

`func NewProtoLoadBalancerMetadataReference() *ProtoLoadBalancerMetadataReference`

NewProtoLoadBalancerMetadataReference instantiates a new ProtoLoadBalancerMetadataReference object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerMetadataReferenceWithDefaults

`func NewProtoLoadBalancerMetadataReferenceWithDefaults() *ProtoLoadBalancerMetadataReference`

NewProtoLoadBalancerMetadataReferenceWithDefaults instantiates a new ProtoLoadBalancerMetadataReference object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCloudAccountId

`func (o *ProtoLoadBalancerMetadataReference) GetCloudAccountId() string`

GetCloudAccountId returns the CloudAccountId field if non-nil, zero value otherwise.

### GetCloudAccountIdOk

`func (o *ProtoLoadBalancerMetadataReference) GetCloudAccountIdOk() (*string, bool)`

GetCloudAccountIdOk returns a tuple with the CloudAccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloudAccountId

`func (o *ProtoLoadBalancerMetadataReference) SetCloudAccountId(v string)`

SetCloudAccountId sets CloudAccountId field to given value.

### HasCloudAccountId

`func (o *ProtoLoadBalancerMetadataReference) HasCloudAccountId() bool`

HasCloudAccountId returns a boolean if a field has been set.

### GetName

`func (o *ProtoLoadBalancerMetadataReference) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoLoadBalancerMetadataReference) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoLoadBalancerMetadataReference) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoLoadBalancerMetadataReference) HasName() bool`

HasName returns a boolean if a field has been set.

### GetResourceId

`func (o *ProtoLoadBalancerMetadataReference) GetResourceId() string`

GetResourceId returns the ResourceId field if non-nil, zero value otherwise.

### GetResourceIdOk

`func (o *ProtoLoadBalancerMetadataReference) GetResourceIdOk() (*string, bool)`

GetResourceIdOk returns a tuple with the ResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceId

`func (o *ProtoLoadBalancerMetadataReference) SetResourceId(v string)`

SetResourceId sets ResourceId field to given value.

### HasResourceId

`func (o *ProtoLoadBalancerMetadataReference) HasResourceId() bool`

HasResourceId returns a boolean if a field has been set.

### GetResourceVersion

`func (o *ProtoLoadBalancerMetadataReference) GetResourceVersion() string`

GetResourceVersion returns the ResourceVersion field if non-nil, zero value otherwise.

### GetResourceVersionOk

`func (o *ProtoLoadBalancerMetadataReference) GetResourceVersionOk() (*string, bool)`

GetResourceVersionOk returns a tuple with the ResourceVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceVersion

`func (o *ProtoLoadBalancerMetadataReference) SetResourceVersion(v string)`

SetResourceVersion sets ResourceVersion field to given value.

### HasResourceVersion

`func (o *ProtoLoadBalancerMetadataReference) HasResourceVersion() bool`

HasResourceVersion returns a boolean if a field has been set.

### GetReserved1

`func (o *ProtoLoadBalancerMetadataReference) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *ProtoLoadBalancerMetadataReference) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *ProtoLoadBalancerMetadataReference) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *ProtoLoadBalancerMetadataReference) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


