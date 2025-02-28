# ProtoLoadBalancerMetadataSearch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CloudAccountId** | Pointer to **string** |  | [optional] 
**Labels** | Pointer to **map[string]string** | If not empty, only return load balancers that have these key/value pairs. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 

## Methods

### NewProtoLoadBalancerMetadataSearch

`func NewProtoLoadBalancerMetadataSearch() *ProtoLoadBalancerMetadataSearch`

NewProtoLoadBalancerMetadataSearch instantiates a new ProtoLoadBalancerMetadataSearch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerMetadataSearchWithDefaults

`func NewProtoLoadBalancerMetadataSearchWithDefaults() *ProtoLoadBalancerMetadataSearch`

NewProtoLoadBalancerMetadataSearchWithDefaults instantiates a new ProtoLoadBalancerMetadataSearch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCloudAccountId

`func (o *ProtoLoadBalancerMetadataSearch) GetCloudAccountId() string`

GetCloudAccountId returns the CloudAccountId field if non-nil, zero value otherwise.

### GetCloudAccountIdOk

`func (o *ProtoLoadBalancerMetadataSearch) GetCloudAccountIdOk() (*string, bool)`

GetCloudAccountIdOk returns a tuple with the CloudAccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloudAccountId

`func (o *ProtoLoadBalancerMetadataSearch) SetCloudAccountId(v string)`

SetCloudAccountId sets CloudAccountId field to given value.

### HasCloudAccountId

`func (o *ProtoLoadBalancerMetadataSearch) HasCloudAccountId() bool`

HasCloudAccountId returns a boolean if a field has been set.

### GetLabels

`func (o *ProtoLoadBalancerMetadataSearch) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ProtoLoadBalancerMetadataSearch) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ProtoLoadBalancerMetadataSearch) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ProtoLoadBalancerMetadataSearch) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetReserved1

`func (o *ProtoLoadBalancerMetadataSearch) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *ProtoLoadBalancerMetadataSearch) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *ProtoLoadBalancerMetadataSearch) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *ProtoLoadBalancerMetadataSearch) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


