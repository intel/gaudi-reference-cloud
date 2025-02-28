# ProtoLoadBalancerListenerStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**VipID** | Pointer to **int32** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**PoolMembers** | Pointer to [**[]ProtoLoadBalancerPoolStatusMember**](ProtoLoadBalancerPoolStatusMember.md) |  | [optional] 
**PoolID** | Pointer to **int32** |  | [optional] 
**State** | Pointer to **string** |  | [optional] 
**Port** | Pointer to **int32** |  | [optional] 

## Methods

### NewProtoLoadBalancerListenerStatus

`func NewProtoLoadBalancerListenerStatus() *ProtoLoadBalancerListenerStatus`

NewProtoLoadBalancerListenerStatus instantiates a new ProtoLoadBalancerListenerStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerListenerStatusWithDefaults

`func NewProtoLoadBalancerListenerStatusWithDefaults() *ProtoLoadBalancerListenerStatus`

NewProtoLoadBalancerListenerStatusWithDefaults instantiates a new ProtoLoadBalancerListenerStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ProtoLoadBalancerListenerStatus) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoLoadBalancerListenerStatus) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoLoadBalancerListenerStatus) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoLoadBalancerListenerStatus) HasName() bool`

HasName returns a boolean if a field has been set.

### GetVipID

`func (o *ProtoLoadBalancerListenerStatus) GetVipID() int32`

GetVipID returns the VipID field if non-nil, zero value otherwise.

### GetVipIDOk

`func (o *ProtoLoadBalancerListenerStatus) GetVipIDOk() (*int32, bool)`

GetVipIDOk returns a tuple with the VipID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVipID

`func (o *ProtoLoadBalancerListenerStatus) SetVipID(v int32)`

SetVipID sets VipID field to given value.

### HasVipID

`func (o *ProtoLoadBalancerListenerStatus) HasVipID() bool`

HasVipID returns a boolean if a field has been set.

### GetMessage

`func (o *ProtoLoadBalancerListenerStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ProtoLoadBalancerListenerStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ProtoLoadBalancerListenerStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ProtoLoadBalancerListenerStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetPoolMembers

`func (o *ProtoLoadBalancerListenerStatus) GetPoolMembers() []ProtoLoadBalancerPoolStatusMember`

GetPoolMembers returns the PoolMembers field if non-nil, zero value otherwise.

### GetPoolMembersOk

`func (o *ProtoLoadBalancerListenerStatus) GetPoolMembersOk() (*[]ProtoLoadBalancerPoolStatusMember, bool)`

GetPoolMembersOk returns a tuple with the PoolMembers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPoolMembers

`func (o *ProtoLoadBalancerListenerStatus) SetPoolMembers(v []ProtoLoadBalancerPoolStatusMember)`

SetPoolMembers sets PoolMembers field to given value.

### HasPoolMembers

`func (o *ProtoLoadBalancerListenerStatus) HasPoolMembers() bool`

HasPoolMembers returns a boolean if a field has been set.

### GetPoolID

`func (o *ProtoLoadBalancerListenerStatus) GetPoolID() int32`

GetPoolID returns the PoolID field if non-nil, zero value otherwise.

### GetPoolIDOk

`func (o *ProtoLoadBalancerListenerStatus) GetPoolIDOk() (*int32, bool)`

GetPoolIDOk returns a tuple with the PoolID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPoolID

`func (o *ProtoLoadBalancerListenerStatus) SetPoolID(v int32)`

SetPoolID sets PoolID field to given value.

### HasPoolID

`func (o *ProtoLoadBalancerListenerStatus) HasPoolID() bool`

HasPoolID returns a boolean if a field has been set.

### GetState

`func (o *ProtoLoadBalancerListenerStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ProtoLoadBalancerListenerStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ProtoLoadBalancerListenerStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *ProtoLoadBalancerListenerStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetPort

`func (o *ProtoLoadBalancerListenerStatus) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *ProtoLoadBalancerListenerStatus) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *ProtoLoadBalancerListenerStatus) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *ProtoLoadBalancerListenerStatus) HasPort() bool`

HasPort returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


