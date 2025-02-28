# ProtoInstanceStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | Pointer to [**ProtoInstancePhase**](ProtoInstancePhase.md) |  | [optional] [default to PROVISIONING]
**Message** | Pointer to **string** | Additional details about the state or any error conditions. | [optional] 
**Interfaces** | Pointer to [**[]ProtoInstanceInterfaceStatus**](ProtoInstanceInterfaceStatus.md) | A list of network interfaces, along with the private IP address assigned to the interface. | [optional] 
**SshProxy** | Pointer to [**ProtoSshProxyTunnelStatus**](ProtoSshProxyTunnelStatus.md) |  | [optional] 
**UserName** | Pointer to **string** | The user name that should be used to SSH into the instance. | [optional] 

## Methods

### NewProtoInstanceStatus

`func NewProtoInstanceStatus() *ProtoInstanceStatus`

NewProtoInstanceStatus instantiates a new ProtoInstanceStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceStatusWithDefaults

`func NewProtoInstanceStatusWithDefaults() *ProtoInstanceStatus`

NewProtoInstanceStatusWithDefaults instantiates a new ProtoInstanceStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *ProtoInstanceStatus) GetPhase() ProtoInstancePhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *ProtoInstanceStatus) GetPhaseOk() (*ProtoInstancePhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *ProtoInstanceStatus) SetPhase(v ProtoInstancePhase)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *ProtoInstanceStatus) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetMessage

`func (o *ProtoInstanceStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ProtoInstanceStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ProtoInstanceStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ProtoInstanceStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetInterfaces

`func (o *ProtoInstanceStatus) GetInterfaces() []ProtoInstanceInterfaceStatus`

GetInterfaces returns the Interfaces field if non-nil, zero value otherwise.

### GetInterfacesOk

`func (o *ProtoInstanceStatus) GetInterfacesOk() (*[]ProtoInstanceInterfaceStatus, bool)`

GetInterfacesOk returns a tuple with the Interfaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInterfaces

`func (o *ProtoInstanceStatus) SetInterfaces(v []ProtoInstanceInterfaceStatus)`

SetInterfaces sets Interfaces field to given value.

### HasInterfaces

`func (o *ProtoInstanceStatus) HasInterfaces() bool`

HasInterfaces returns a boolean if a field has been set.

### GetSshProxy

`func (o *ProtoInstanceStatus) GetSshProxy() ProtoSshProxyTunnelStatus`

GetSshProxy returns the SshProxy field if non-nil, zero value otherwise.

### GetSshProxyOk

`func (o *ProtoInstanceStatus) GetSshProxyOk() (*ProtoSshProxyTunnelStatus, bool)`

GetSshProxyOk returns a tuple with the SshProxy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSshProxy

`func (o *ProtoInstanceStatus) SetSshProxy(v ProtoSshProxyTunnelStatus)`

SetSshProxy sets SshProxy field to given value.

### HasSshProxy

`func (o *ProtoInstanceStatus) HasSshProxy() bool`

HasSshProxy returns a boolean if a field has been set.

### GetUserName

`func (o *ProtoInstanceStatus) GetUserName() string`

GetUserName returns the UserName field if non-nil, zero value otherwise.

### GetUserNameOk

`func (o *ProtoInstanceStatus) GetUserNameOk() (*string, bool)`

GetUserNameOk returns a tuple with the UserName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserName

`func (o *ProtoInstanceStatus) SetUserName(v string)`

SetUserName sets UserName field to given value.

### HasUserName

`func (o *ProtoInstanceStatus) HasUserName() bool`

HasUserName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


