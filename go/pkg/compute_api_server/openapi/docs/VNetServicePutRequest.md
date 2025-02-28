# VNetServicePutRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**VNetServicePutRequestMetadata**](VNetServicePutRequestMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoVNetSpec**](ProtoVNetSpec.md) |  | [optional] 

## Methods

### NewVNetServicePutRequest

`func NewVNetServicePutRequest() *VNetServicePutRequest`

NewVNetServicePutRequest instantiates a new VNetServicePutRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVNetServicePutRequestWithDefaults

`func NewVNetServicePutRequestWithDefaults() *VNetServicePutRequest`

NewVNetServicePutRequestWithDefaults instantiates a new VNetServicePutRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *VNetServicePutRequest) GetMetadata() VNetServicePutRequestMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *VNetServicePutRequest) GetMetadataOk() (*VNetServicePutRequestMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *VNetServicePutRequest) SetMetadata(v VNetServicePutRequestMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *VNetServicePutRequest) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *VNetServicePutRequest) GetSpec() ProtoVNetSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *VNetServicePutRequest) GetSpecOk() (*ProtoVNetSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *VNetServicePutRequest) SetSpec(v ProtoVNetSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *VNetServicePutRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


