# ProtoSshProxyTunnelStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProxyUser** | Pointer to **string** | The username required to connect to the SSH proxy. | [optional] 
**ProxyAddress** | Pointer to **string** | The IP address or FQDN of the SSH proxy. | [optional] 
**ProxyPort** | Pointer to **int32** | The TCP port for the SSH proxy. | [optional] 

## Methods

### NewProtoSshProxyTunnelStatus

`func NewProtoSshProxyTunnelStatus() *ProtoSshProxyTunnelStatus`

NewProtoSshProxyTunnelStatus instantiates a new ProtoSshProxyTunnelStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoSshProxyTunnelStatusWithDefaults

`func NewProtoSshProxyTunnelStatusWithDefaults() *ProtoSshProxyTunnelStatus`

NewProtoSshProxyTunnelStatusWithDefaults instantiates a new ProtoSshProxyTunnelStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProxyUser

`func (o *ProtoSshProxyTunnelStatus) GetProxyUser() string`

GetProxyUser returns the ProxyUser field if non-nil, zero value otherwise.

### GetProxyUserOk

`func (o *ProtoSshProxyTunnelStatus) GetProxyUserOk() (*string, bool)`

GetProxyUserOk returns a tuple with the ProxyUser field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProxyUser

`func (o *ProtoSshProxyTunnelStatus) SetProxyUser(v string)`

SetProxyUser sets ProxyUser field to given value.

### HasProxyUser

`func (o *ProtoSshProxyTunnelStatus) HasProxyUser() bool`

HasProxyUser returns a boolean if a field has been set.

### GetProxyAddress

`func (o *ProtoSshProxyTunnelStatus) GetProxyAddress() string`

GetProxyAddress returns the ProxyAddress field if non-nil, zero value otherwise.

### GetProxyAddressOk

`func (o *ProtoSshProxyTunnelStatus) GetProxyAddressOk() (*string, bool)`

GetProxyAddressOk returns a tuple with the ProxyAddress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProxyAddress

`func (o *ProtoSshProxyTunnelStatus) SetProxyAddress(v string)`

SetProxyAddress sets ProxyAddress field to given value.

### HasProxyAddress

`func (o *ProtoSshProxyTunnelStatus) HasProxyAddress() bool`

HasProxyAddress returns a boolean if a field has been set.

### GetProxyPort

`func (o *ProtoSshProxyTunnelStatus) GetProxyPort() int32`

GetProxyPort returns the ProxyPort field if non-nil, zero value otherwise.

### GetProxyPortOk

`func (o *ProtoSshProxyTunnelStatus) GetProxyPortOk() (*int32, bool)`

GetProxyPortOk returns a tuple with the ProxyPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProxyPort

`func (o *ProtoSshProxyTunnelStatus) SetProxyPort(v int32)`

SetProxyPort sets ProxyPort field to given value.

### HasProxyPort

`func (o *ProtoSshProxyTunnelStatus) HasProxyPort() bool`

HasProxyPort returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


