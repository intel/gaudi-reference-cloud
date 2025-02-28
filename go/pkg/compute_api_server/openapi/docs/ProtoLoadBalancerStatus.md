# ProtoLoadBalancerStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Conditions** | Pointer to [**ProtoLoadBalancerConditionsStatus**](ProtoLoadBalancerConditionsStatus.md) |  | [optional] 
**Listeners** | Pointer to [**[]ProtoLoadBalancerListenerStatus**](ProtoLoadBalancerListenerStatus.md) |  | [optional] 
**State** | Pointer to **string** |  | [optional] 
**Vip** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 

## Methods

### NewProtoLoadBalancerStatus

`func NewProtoLoadBalancerStatus() *ProtoLoadBalancerStatus`

NewProtoLoadBalancerStatus instantiates a new ProtoLoadBalancerStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerStatusWithDefaults

`func NewProtoLoadBalancerStatusWithDefaults() *ProtoLoadBalancerStatus`

NewProtoLoadBalancerStatusWithDefaults instantiates a new ProtoLoadBalancerStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConditions

`func (o *ProtoLoadBalancerStatus) GetConditions() ProtoLoadBalancerConditionsStatus`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ProtoLoadBalancerStatus) GetConditionsOk() (*ProtoLoadBalancerConditionsStatus, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ProtoLoadBalancerStatus) SetConditions(v ProtoLoadBalancerConditionsStatus)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ProtoLoadBalancerStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetListeners

`func (o *ProtoLoadBalancerStatus) GetListeners() []ProtoLoadBalancerListenerStatus`

GetListeners returns the Listeners field if non-nil, zero value otherwise.

### GetListenersOk

`func (o *ProtoLoadBalancerStatus) GetListenersOk() (*[]ProtoLoadBalancerListenerStatus, bool)`

GetListenersOk returns a tuple with the Listeners field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetListeners

`func (o *ProtoLoadBalancerStatus) SetListeners(v []ProtoLoadBalancerListenerStatus)`

SetListeners sets Listeners field to given value.

### HasListeners

`func (o *ProtoLoadBalancerStatus) HasListeners() bool`

HasListeners returns a boolean if a field has been set.

### GetState

`func (o *ProtoLoadBalancerStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ProtoLoadBalancerStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ProtoLoadBalancerStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *ProtoLoadBalancerStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetVip

`func (o *ProtoLoadBalancerStatus) GetVip() string`

GetVip returns the Vip field if non-nil, zero value otherwise.

### GetVipOk

`func (o *ProtoLoadBalancerStatus) GetVipOk() (*string, bool)`

GetVipOk returns a tuple with the Vip field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVip

`func (o *ProtoLoadBalancerStatus) SetVip(v string)`

SetVip sets Vip field to given value.

### HasVip

`func (o *ProtoLoadBalancerStatus) HasVip() bool`

HasVip returns a boolean if a field has been set.

### GetMessage

`func (o *ProtoLoadBalancerStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ProtoLoadBalancerStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ProtoLoadBalancerStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ProtoLoadBalancerStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


