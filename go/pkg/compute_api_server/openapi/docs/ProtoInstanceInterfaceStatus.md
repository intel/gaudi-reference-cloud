# ProtoInstanceInterfaceStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Not implemented. | [optional] 
**VNet** | Pointer to **string** | Not implemented. | [optional] 
**DnsName** | Pointer to **string** | Fully qualified domain name (FQDN) of interface. | [optional] 
**PrefixLength** | Pointer to **int32** | Subnet prefix length. | [optional] 
**Addresses** | Pointer to **[]string** | List of IP addresses. | [optional] 
**Subnet** | Pointer to **string** | Subnet IP address in format \&quot;1.2.3.4\&quot;. | [optional] 
**Gateway** | Pointer to **string** | Gateway IP address. | [optional] 

## Methods

### NewProtoInstanceInterfaceStatus

`func NewProtoInstanceInterfaceStatus() *ProtoInstanceInterfaceStatus`

NewProtoInstanceInterfaceStatus instantiates a new ProtoInstanceInterfaceStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceInterfaceStatusWithDefaults

`func NewProtoInstanceInterfaceStatusWithDefaults() *ProtoInstanceInterfaceStatus`

NewProtoInstanceInterfaceStatusWithDefaults instantiates a new ProtoInstanceInterfaceStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ProtoInstanceInterfaceStatus) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoInstanceInterfaceStatus) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoInstanceInterfaceStatus) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoInstanceInterfaceStatus) HasName() bool`

HasName returns a boolean if a field has been set.

### GetVNet

`func (o *ProtoInstanceInterfaceStatus) GetVNet() string`

GetVNet returns the VNet field if non-nil, zero value otherwise.

### GetVNetOk

`func (o *ProtoInstanceInterfaceStatus) GetVNetOk() (*string, bool)`

GetVNetOk returns a tuple with the VNet field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVNet

`func (o *ProtoInstanceInterfaceStatus) SetVNet(v string)`

SetVNet sets VNet field to given value.

### HasVNet

`func (o *ProtoInstanceInterfaceStatus) HasVNet() bool`

HasVNet returns a boolean if a field has been set.

### GetDnsName

`func (o *ProtoInstanceInterfaceStatus) GetDnsName() string`

GetDnsName returns the DnsName field if non-nil, zero value otherwise.

### GetDnsNameOk

`func (o *ProtoInstanceInterfaceStatus) GetDnsNameOk() (*string, bool)`

GetDnsNameOk returns a tuple with the DnsName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDnsName

`func (o *ProtoInstanceInterfaceStatus) SetDnsName(v string)`

SetDnsName sets DnsName field to given value.

### HasDnsName

`func (o *ProtoInstanceInterfaceStatus) HasDnsName() bool`

HasDnsName returns a boolean if a field has been set.

### GetPrefixLength

`func (o *ProtoInstanceInterfaceStatus) GetPrefixLength() int32`

GetPrefixLength returns the PrefixLength field if non-nil, zero value otherwise.

### GetPrefixLengthOk

`func (o *ProtoInstanceInterfaceStatus) GetPrefixLengthOk() (*int32, bool)`

GetPrefixLengthOk returns a tuple with the PrefixLength field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrefixLength

`func (o *ProtoInstanceInterfaceStatus) SetPrefixLength(v int32)`

SetPrefixLength sets PrefixLength field to given value.

### HasPrefixLength

`func (o *ProtoInstanceInterfaceStatus) HasPrefixLength() bool`

HasPrefixLength returns a boolean if a field has been set.

### GetAddresses

`func (o *ProtoInstanceInterfaceStatus) GetAddresses() []string`

GetAddresses returns the Addresses field if non-nil, zero value otherwise.

### GetAddressesOk

`func (o *ProtoInstanceInterfaceStatus) GetAddressesOk() (*[]string, bool)`

GetAddressesOk returns a tuple with the Addresses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAddresses

`func (o *ProtoInstanceInterfaceStatus) SetAddresses(v []string)`

SetAddresses sets Addresses field to given value.

### HasAddresses

`func (o *ProtoInstanceInterfaceStatus) HasAddresses() bool`

HasAddresses returns a boolean if a field has been set.

### GetSubnet

`func (o *ProtoInstanceInterfaceStatus) GetSubnet() string`

GetSubnet returns the Subnet field if non-nil, zero value otherwise.

### GetSubnetOk

`func (o *ProtoInstanceInterfaceStatus) GetSubnetOk() (*string, bool)`

GetSubnetOk returns a tuple with the Subnet field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubnet

`func (o *ProtoInstanceInterfaceStatus) SetSubnet(v string)`

SetSubnet sets Subnet field to given value.

### HasSubnet

`func (o *ProtoInstanceInterfaceStatus) HasSubnet() bool`

HasSubnet returns a boolean if a field has been set.

### GetGateway

`func (o *ProtoInstanceInterfaceStatus) GetGateway() string`

GetGateway returns the Gateway field if non-nil, zero value otherwise.

### GetGatewayOk

`func (o *ProtoInstanceInterfaceStatus) GetGatewayOk() (*string, bool)`

GetGatewayOk returns a tuple with the Gateway field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGateway

`func (o *ProtoInstanceInterfaceStatus) SetGateway(v string)`

SetGateway sets Gateway field to given value.

### HasGateway

`func (o *ProtoInstanceInterfaceStatus) HasGateway() bool`

HasGateway returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


