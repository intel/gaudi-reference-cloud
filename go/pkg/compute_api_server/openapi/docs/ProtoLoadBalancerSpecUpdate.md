# ProtoLoadBalancerSpecUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Listeners** | Pointer to [**[]ProtoLoadBalancerListener**](ProtoLoadBalancerListener.md) |  | [optional] 
**Security** | Pointer to [**ProtoLoadBalancerSecurity**](ProtoLoadBalancerSecurity.md) |  | [optional] 

## Methods

### NewProtoLoadBalancerSpecUpdate

`func NewProtoLoadBalancerSpecUpdate() *ProtoLoadBalancerSpecUpdate`

NewProtoLoadBalancerSpecUpdate instantiates a new ProtoLoadBalancerSpecUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerSpecUpdateWithDefaults

`func NewProtoLoadBalancerSpecUpdateWithDefaults() *ProtoLoadBalancerSpecUpdate`

NewProtoLoadBalancerSpecUpdateWithDefaults instantiates a new ProtoLoadBalancerSpecUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetListeners

`func (o *ProtoLoadBalancerSpecUpdate) GetListeners() []ProtoLoadBalancerListener`

GetListeners returns the Listeners field if non-nil, zero value otherwise.

### GetListenersOk

`func (o *ProtoLoadBalancerSpecUpdate) GetListenersOk() (*[]ProtoLoadBalancerListener, bool)`

GetListenersOk returns a tuple with the Listeners field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetListeners

`func (o *ProtoLoadBalancerSpecUpdate) SetListeners(v []ProtoLoadBalancerListener)`

SetListeners sets Listeners field to given value.

### HasListeners

`func (o *ProtoLoadBalancerSpecUpdate) HasListeners() bool`

HasListeners returns a boolean if a field has been set.

### GetSecurity

`func (o *ProtoLoadBalancerSpecUpdate) GetSecurity() ProtoLoadBalancerSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *ProtoLoadBalancerSpecUpdate) GetSecurityOk() (*ProtoLoadBalancerSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *ProtoLoadBalancerSpecUpdate) SetSecurity(v ProtoLoadBalancerSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *ProtoLoadBalancerSpecUpdate) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


