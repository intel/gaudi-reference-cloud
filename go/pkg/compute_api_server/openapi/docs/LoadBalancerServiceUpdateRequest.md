# LoadBalancerServiceUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**LoadBalancerServiceUpdateRequestMetadata**](LoadBalancerServiceUpdateRequestMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoLoadBalancerSpecUpdate**](ProtoLoadBalancerSpecUpdate.md) |  | [optional] 

## Methods

### NewLoadBalancerServiceUpdateRequest

`func NewLoadBalancerServiceUpdateRequest() *LoadBalancerServiceUpdateRequest`

NewLoadBalancerServiceUpdateRequest instantiates a new LoadBalancerServiceUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLoadBalancerServiceUpdateRequestWithDefaults

`func NewLoadBalancerServiceUpdateRequestWithDefaults() *LoadBalancerServiceUpdateRequest`

NewLoadBalancerServiceUpdateRequestWithDefaults instantiates a new LoadBalancerServiceUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *LoadBalancerServiceUpdateRequest) GetMetadata() LoadBalancerServiceUpdateRequestMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *LoadBalancerServiceUpdateRequest) GetMetadataOk() (*LoadBalancerServiceUpdateRequestMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *LoadBalancerServiceUpdateRequest) SetMetadata(v LoadBalancerServiceUpdateRequestMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *LoadBalancerServiceUpdateRequest) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *LoadBalancerServiceUpdateRequest) GetSpec() ProtoLoadBalancerSpecUpdate`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *LoadBalancerServiceUpdateRequest) GetSpecOk() (*ProtoLoadBalancerSpecUpdate, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *LoadBalancerServiceUpdateRequest) SetSpec(v ProtoLoadBalancerSpecUpdate)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *LoadBalancerServiceUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


