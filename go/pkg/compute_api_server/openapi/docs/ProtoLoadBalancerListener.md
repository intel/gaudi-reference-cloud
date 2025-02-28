# ProtoLoadBalancerListener

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Port** | Pointer to **int32** | The public port of the load balancer. | [optional] 
**Pool** | Pointer to [**ProtoLoadBalancerPool**](ProtoLoadBalancerPool.md) |  | [optional] 

## Methods

### NewProtoLoadBalancerListener

`func NewProtoLoadBalancerListener() *ProtoLoadBalancerListener`

NewProtoLoadBalancerListener instantiates a new ProtoLoadBalancerListener object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerListenerWithDefaults

`func NewProtoLoadBalancerListenerWithDefaults() *ProtoLoadBalancerListener`

NewProtoLoadBalancerListenerWithDefaults instantiates a new ProtoLoadBalancerListener object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPort

`func (o *ProtoLoadBalancerListener) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *ProtoLoadBalancerListener) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *ProtoLoadBalancerListener) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *ProtoLoadBalancerListener) HasPort() bool`

HasPort returns a boolean if a field has been set.

### GetPool

`func (o *ProtoLoadBalancerListener) GetPool() ProtoLoadBalancerPool`

GetPool returns the Pool field if non-nil, zero value otherwise.

### GetPoolOk

`func (o *ProtoLoadBalancerListener) GetPoolOk() (*ProtoLoadBalancerPool, bool)`

GetPoolOk returns a tuple with the Pool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPool

`func (o *ProtoLoadBalancerListener) SetPool(v ProtoLoadBalancerPool)`

SetPool sets Pool field to given value.

### HasPool

`func (o *ProtoLoadBalancerListener) HasPool() bool`

HasPool returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


