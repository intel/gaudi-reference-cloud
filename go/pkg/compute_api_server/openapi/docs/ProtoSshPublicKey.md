# ProtoSshPublicKey

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**ProtoResourceMetadata**](ProtoResourceMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoSshPublicKeySpec**](ProtoSshPublicKeySpec.md) |  | [optional] 

## Methods

### NewProtoSshPublicKey

`func NewProtoSshPublicKey() *ProtoSshPublicKey`

NewProtoSshPublicKey instantiates a new ProtoSshPublicKey object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoSshPublicKeyWithDefaults

`func NewProtoSshPublicKeyWithDefaults() *ProtoSshPublicKey`

NewProtoSshPublicKeyWithDefaults instantiates a new ProtoSshPublicKey object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *ProtoSshPublicKey) GetMetadata() ProtoResourceMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ProtoSshPublicKey) GetMetadataOk() (*ProtoResourceMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ProtoSshPublicKey) SetMetadata(v ProtoResourceMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *ProtoSshPublicKey) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *ProtoSshPublicKey) GetSpec() ProtoSshPublicKeySpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ProtoSshPublicKey) GetSpecOk() (*ProtoSshPublicKeySpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ProtoSshPublicKey) SetSpec(v ProtoSshPublicKeySpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *ProtoSshPublicKey) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


