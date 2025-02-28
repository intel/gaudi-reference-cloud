# ProtoLoadBalancerConditionsListenerStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Port** | Pointer to **int32** |  | [optional] 
**PoolCreated** | Pointer to **bool** |  | [optional] 
**VipCreated** | Pointer to **bool** |  | [optional] 
**VipPoolLinked** | Pointer to **bool** |  | [optional] 

## Methods

### NewProtoLoadBalancerConditionsListenerStatus

`func NewProtoLoadBalancerConditionsListenerStatus() *ProtoLoadBalancerConditionsListenerStatus`

NewProtoLoadBalancerConditionsListenerStatus instantiates a new ProtoLoadBalancerConditionsListenerStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerConditionsListenerStatusWithDefaults

`func NewProtoLoadBalancerConditionsListenerStatusWithDefaults() *ProtoLoadBalancerConditionsListenerStatus`

NewProtoLoadBalancerConditionsListenerStatusWithDefaults instantiates a new ProtoLoadBalancerConditionsListenerStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPort

`func (o *ProtoLoadBalancerConditionsListenerStatus) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *ProtoLoadBalancerConditionsListenerStatus) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *ProtoLoadBalancerConditionsListenerStatus) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *ProtoLoadBalancerConditionsListenerStatus) HasPort() bool`

HasPort returns a boolean if a field has been set.

### GetPoolCreated

`func (o *ProtoLoadBalancerConditionsListenerStatus) GetPoolCreated() bool`

GetPoolCreated returns the PoolCreated field if non-nil, zero value otherwise.

### GetPoolCreatedOk

`func (o *ProtoLoadBalancerConditionsListenerStatus) GetPoolCreatedOk() (*bool, bool)`

GetPoolCreatedOk returns a tuple with the PoolCreated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPoolCreated

`func (o *ProtoLoadBalancerConditionsListenerStatus) SetPoolCreated(v bool)`

SetPoolCreated sets PoolCreated field to given value.

### HasPoolCreated

`func (o *ProtoLoadBalancerConditionsListenerStatus) HasPoolCreated() bool`

HasPoolCreated returns a boolean if a field has been set.

### GetVipCreated

`func (o *ProtoLoadBalancerConditionsListenerStatus) GetVipCreated() bool`

GetVipCreated returns the VipCreated field if non-nil, zero value otherwise.

### GetVipCreatedOk

`func (o *ProtoLoadBalancerConditionsListenerStatus) GetVipCreatedOk() (*bool, bool)`

GetVipCreatedOk returns a tuple with the VipCreated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVipCreated

`func (o *ProtoLoadBalancerConditionsListenerStatus) SetVipCreated(v bool)`

SetVipCreated sets VipCreated field to given value.

### HasVipCreated

`func (o *ProtoLoadBalancerConditionsListenerStatus) HasVipCreated() bool`

HasVipCreated returns a boolean if a field has been set.

### GetVipPoolLinked

`func (o *ProtoLoadBalancerConditionsListenerStatus) GetVipPoolLinked() bool`

GetVipPoolLinked returns the VipPoolLinked field if non-nil, zero value otherwise.

### GetVipPoolLinkedOk

`func (o *ProtoLoadBalancerConditionsListenerStatus) GetVipPoolLinkedOk() (*bool, bool)`

GetVipPoolLinkedOk returns a tuple with the VipPoolLinked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVipPoolLinked

`func (o *ProtoLoadBalancerConditionsListenerStatus) SetVipPoolLinked(v bool)`

SetVipPoolLinked sets VipPoolLinked field to given value.

### HasVipPoolLinked

`func (o *ProtoLoadBalancerConditionsListenerStatus) HasVipPoolLinked() bool`

HasVipPoolLinked returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


