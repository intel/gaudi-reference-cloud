# ProtoLoadBalancerSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Listeners** | Pointer to [**[]ProtoLoadBalancerListener**](ProtoLoadBalancerListener.md) |  | [optional] 
**Security** | Pointer to [**ProtoLoadBalancerSecurity**](ProtoLoadBalancerSecurity.md) |  | [optional] 

## Methods

### NewProtoLoadBalancerSpec

`func NewProtoLoadBalancerSpec() *ProtoLoadBalancerSpec`

NewProtoLoadBalancerSpec instantiates a new ProtoLoadBalancerSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerSpecWithDefaults

`func NewProtoLoadBalancerSpecWithDefaults() *ProtoLoadBalancerSpec`

NewProtoLoadBalancerSpecWithDefaults instantiates a new ProtoLoadBalancerSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetListeners

`func (o *ProtoLoadBalancerSpec) GetListeners() []ProtoLoadBalancerListener`

GetListeners returns the Listeners field if non-nil, zero value otherwise.

### GetListenersOk

`func (o *ProtoLoadBalancerSpec) GetListenersOk() (*[]ProtoLoadBalancerListener, bool)`

GetListenersOk returns a tuple with the Listeners field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetListeners

`func (o *ProtoLoadBalancerSpec) SetListeners(v []ProtoLoadBalancerListener)`

SetListeners sets Listeners field to given value.

### HasListeners

`func (o *ProtoLoadBalancerSpec) HasListeners() bool`

HasListeners returns a boolean if a field has been set.

### GetSecurity

`func (o *ProtoLoadBalancerSpec) GetSecurity() ProtoLoadBalancerSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *ProtoLoadBalancerSpec) GetSecurityOk() (*ProtoLoadBalancerSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *ProtoLoadBalancerSpec) SetSecurity(v ProtoLoadBalancerSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *ProtoLoadBalancerSpec) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


