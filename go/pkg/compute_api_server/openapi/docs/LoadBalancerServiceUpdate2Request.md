# LoadBalancerServiceUpdate2Request

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**LoadBalancerServiceUpdate2RequestMetadata**](LoadBalancerServiceUpdate2RequestMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoLoadBalancerSpecUpdate**](ProtoLoadBalancerSpecUpdate.md) |  | [optional] 

## Methods

### NewLoadBalancerServiceUpdate2Request

`func NewLoadBalancerServiceUpdate2Request() *LoadBalancerServiceUpdate2Request`

NewLoadBalancerServiceUpdate2Request instantiates a new LoadBalancerServiceUpdate2Request object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLoadBalancerServiceUpdate2RequestWithDefaults

`func NewLoadBalancerServiceUpdate2RequestWithDefaults() *LoadBalancerServiceUpdate2Request`

NewLoadBalancerServiceUpdate2RequestWithDefaults instantiates a new LoadBalancerServiceUpdate2Request object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *LoadBalancerServiceUpdate2Request) GetMetadata() LoadBalancerServiceUpdate2RequestMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *LoadBalancerServiceUpdate2Request) GetMetadataOk() (*LoadBalancerServiceUpdate2RequestMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *LoadBalancerServiceUpdate2Request) SetMetadata(v LoadBalancerServiceUpdate2RequestMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *LoadBalancerServiceUpdate2Request) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *LoadBalancerServiceUpdate2Request) GetSpec() ProtoLoadBalancerSpecUpdate`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *LoadBalancerServiceUpdate2Request) GetSpecOk() (*ProtoLoadBalancerSpecUpdate, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *LoadBalancerServiceUpdate2Request) SetSpec(v ProtoLoadBalancerSpecUpdate)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *LoadBalancerServiceUpdate2Request) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


