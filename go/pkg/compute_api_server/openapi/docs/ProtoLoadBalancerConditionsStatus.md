# ProtoLoadBalancerConditionsStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Listeners** | Pointer to [**[]ProtoLoadBalancerConditionsListenerStatus**](ProtoLoadBalancerConditionsListenerStatus.md) |  | [optional] 
**FirewallRuleCreated** | Pointer to **bool** |  | [optional] 

## Methods

### NewProtoLoadBalancerConditionsStatus

`func NewProtoLoadBalancerConditionsStatus() *ProtoLoadBalancerConditionsStatus`

NewProtoLoadBalancerConditionsStatus instantiates a new ProtoLoadBalancerConditionsStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerConditionsStatusWithDefaults

`func NewProtoLoadBalancerConditionsStatusWithDefaults() *ProtoLoadBalancerConditionsStatus`

NewProtoLoadBalancerConditionsStatusWithDefaults instantiates a new ProtoLoadBalancerConditionsStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetListeners

`func (o *ProtoLoadBalancerConditionsStatus) GetListeners() []ProtoLoadBalancerConditionsListenerStatus`

GetListeners returns the Listeners field if non-nil, zero value otherwise.

### GetListenersOk

`func (o *ProtoLoadBalancerConditionsStatus) GetListenersOk() (*[]ProtoLoadBalancerConditionsListenerStatus, bool)`

GetListenersOk returns a tuple with the Listeners field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetListeners

`func (o *ProtoLoadBalancerConditionsStatus) SetListeners(v []ProtoLoadBalancerConditionsListenerStatus)`

SetListeners sets Listeners field to given value.

### HasListeners

`func (o *ProtoLoadBalancerConditionsStatus) HasListeners() bool`

HasListeners returns a boolean if a field has been set.

### GetFirewallRuleCreated

`func (o *ProtoLoadBalancerConditionsStatus) GetFirewallRuleCreated() bool`

GetFirewallRuleCreated returns the FirewallRuleCreated field if non-nil, zero value otherwise.

### GetFirewallRuleCreatedOk

`func (o *ProtoLoadBalancerConditionsStatus) GetFirewallRuleCreatedOk() (*bool, bool)`

GetFirewallRuleCreatedOk returns a tuple with the FirewallRuleCreated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFirewallRuleCreated

`func (o *ProtoLoadBalancerConditionsStatus) SetFirewallRuleCreated(v bool)`

SetFirewallRuleCreated sets FirewallRuleCreated field to given value.

### HasFirewallRuleCreated

`func (o *ProtoLoadBalancerConditionsStatus) HasFirewallRuleCreated() bool`

HasFirewallRuleCreated returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


