# ProtoNetworkInterface

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Name of the network interface as known by the operating system. Not implemented. | [optional] 
**VNet** | Pointer to **string** | Name of the VNet that the network interface connects to. | [optional] 

## Methods

### NewProtoNetworkInterface

`func NewProtoNetworkInterface() *ProtoNetworkInterface`

NewProtoNetworkInterface instantiates a new ProtoNetworkInterface object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoNetworkInterfaceWithDefaults

`func NewProtoNetworkInterfaceWithDefaults() *ProtoNetworkInterface`

NewProtoNetworkInterfaceWithDefaults instantiates a new ProtoNetworkInterface object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ProtoNetworkInterface) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoNetworkInterface) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoNetworkInterface) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoNetworkInterface) HasName() bool`

HasName returns a boolean if a field has been set.

### GetVNet

`func (o *ProtoNetworkInterface) GetVNet() string`

GetVNet returns the VNet field if non-nil, zero value otherwise.

### GetVNetOk

`func (o *ProtoNetworkInterface) GetVNetOk() (*string, bool)`

GetVNetOk returns a tuple with the VNet field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVNet

`func (o *ProtoNetworkInterface) SetVNet(v string)`

SetVNet sets VNet field to given value.

### HasVNet

`func (o *ProtoNetworkInterface) HasVNet() bool`

HasVNet returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


