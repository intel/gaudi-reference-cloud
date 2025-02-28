# ProtoLoadBalancer

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**ProtoLoadBalancerMetadata**](ProtoLoadBalancerMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoLoadBalancerSpec**](ProtoLoadBalancerSpec.md) |  | [optional] 
**Status** | Pointer to [**ProtoLoadBalancerStatus**](ProtoLoadBalancerStatus.md) |  | [optional] 

## Methods

### NewProtoLoadBalancer

`func NewProtoLoadBalancer() *ProtoLoadBalancer`

NewProtoLoadBalancer instantiates a new ProtoLoadBalancer object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerWithDefaults

`func NewProtoLoadBalancerWithDefaults() *ProtoLoadBalancer`

NewProtoLoadBalancerWithDefaults instantiates a new ProtoLoadBalancer object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *ProtoLoadBalancer) GetMetadata() ProtoLoadBalancerMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ProtoLoadBalancer) GetMetadataOk() (*ProtoLoadBalancerMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ProtoLoadBalancer) SetMetadata(v ProtoLoadBalancerMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *ProtoLoadBalancer) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *ProtoLoadBalancer) GetSpec() ProtoLoadBalancerSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ProtoLoadBalancer) GetSpecOk() (*ProtoLoadBalancerSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ProtoLoadBalancer) SetSpec(v ProtoLoadBalancerSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *ProtoLoadBalancer) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetStatus

`func (o *ProtoLoadBalancer) GetStatus() ProtoLoadBalancerStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProtoLoadBalancer) GetStatusOk() (*ProtoLoadBalancerStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProtoLoadBalancer) SetStatus(v ProtoLoadBalancerStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ProtoLoadBalancer) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


