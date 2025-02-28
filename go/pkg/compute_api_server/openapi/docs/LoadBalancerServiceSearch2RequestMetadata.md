# LoadBalancerServiceSearch2RequestMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Labels** | Pointer to **map[string]string** | If not empty, only return load balancers that have these key/value pairs. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 

## Methods

### NewLoadBalancerServiceSearch2RequestMetadata

`func NewLoadBalancerServiceSearch2RequestMetadata() *LoadBalancerServiceSearch2RequestMetadata`

NewLoadBalancerServiceSearch2RequestMetadata instantiates a new LoadBalancerServiceSearch2RequestMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLoadBalancerServiceSearch2RequestMetadataWithDefaults

`func NewLoadBalancerServiceSearch2RequestMetadataWithDefaults() *LoadBalancerServiceSearch2RequestMetadata`

NewLoadBalancerServiceSearch2RequestMetadataWithDefaults instantiates a new LoadBalancerServiceSearch2RequestMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLabels

`func (o *LoadBalancerServiceSearch2RequestMetadata) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *LoadBalancerServiceSearch2RequestMetadata) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *LoadBalancerServiceSearch2RequestMetadata) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *LoadBalancerServiceSearch2RequestMetadata) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetReserved1

`func (o *LoadBalancerServiceSearch2RequestMetadata) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *LoadBalancerServiceSearch2RequestMetadata) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *LoadBalancerServiceSearch2RequestMetadata) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *LoadBalancerServiceSearch2RequestMetadata) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


