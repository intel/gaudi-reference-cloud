# ProtoVNetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Region** | Pointer to **string** |  | [optional] 
**AvailabilityZone** | Pointer to **string** |  | [optional] 
**PrefixLength** | Pointer to **int32** | The reserved subnet will have a prefix length with this value or less. | [optional] 

## Methods

### NewProtoVNetSpec

`func NewProtoVNetSpec() *ProtoVNetSpec`

NewProtoVNetSpec instantiates a new ProtoVNetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoVNetSpecWithDefaults

`func NewProtoVNetSpecWithDefaults() *ProtoVNetSpec`

NewProtoVNetSpecWithDefaults instantiates a new ProtoVNetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRegion

`func (o *ProtoVNetSpec) GetRegion() string`

GetRegion returns the Region field if non-nil, zero value otherwise.

### GetRegionOk

`func (o *ProtoVNetSpec) GetRegionOk() (*string, bool)`

GetRegionOk returns a tuple with the Region field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegion

`func (o *ProtoVNetSpec) SetRegion(v string)`

SetRegion sets Region field to given value.

### HasRegion

`func (o *ProtoVNetSpec) HasRegion() bool`

HasRegion returns a boolean if a field has been set.

### GetAvailabilityZone

`func (o *ProtoVNetSpec) GetAvailabilityZone() string`

GetAvailabilityZone returns the AvailabilityZone field if non-nil, zero value otherwise.

### GetAvailabilityZoneOk

`func (o *ProtoVNetSpec) GetAvailabilityZoneOk() (*string, bool)`

GetAvailabilityZoneOk returns a tuple with the AvailabilityZone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvailabilityZone

`func (o *ProtoVNetSpec) SetAvailabilityZone(v string)`

SetAvailabilityZone sets AvailabilityZone field to given value.

### HasAvailabilityZone

`func (o *ProtoVNetSpec) HasAvailabilityZone() bool`

HasAvailabilityZone returns a boolean if a field has been set.

### GetPrefixLength

`func (o *ProtoVNetSpec) GetPrefixLength() int32`

GetPrefixLength returns the PrefixLength field if non-nil, zero value otherwise.

### GetPrefixLengthOk

`func (o *ProtoVNetSpec) GetPrefixLengthOk() (*int32, bool)`

GetPrefixLengthOk returns a tuple with the PrefixLength field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrefixLength

`func (o *ProtoVNetSpec) SetPrefixLength(v int32)`

SetPrefixLength sets PrefixLength field to given value.

### HasPrefixLength

`func (o *ProtoVNetSpec) HasPrefixLength() bool`

HasPrefixLength returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


