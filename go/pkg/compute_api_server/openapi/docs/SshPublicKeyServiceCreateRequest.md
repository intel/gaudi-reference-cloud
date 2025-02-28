# SshPublicKeyServiceCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**SshPublicKeyServiceCreateRequestMetadata**](SshPublicKeyServiceCreateRequestMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoSshPublicKeySpec**](ProtoSshPublicKeySpec.md) |  | [optional] 

## Methods

### NewSshPublicKeyServiceCreateRequest

`func NewSshPublicKeyServiceCreateRequest() *SshPublicKeyServiceCreateRequest`

NewSshPublicKeyServiceCreateRequest instantiates a new SshPublicKeyServiceCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSshPublicKeyServiceCreateRequestWithDefaults

`func NewSshPublicKeyServiceCreateRequestWithDefaults() *SshPublicKeyServiceCreateRequest`

NewSshPublicKeyServiceCreateRequestWithDefaults instantiates a new SshPublicKeyServiceCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *SshPublicKeyServiceCreateRequest) GetMetadata() SshPublicKeyServiceCreateRequestMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *SshPublicKeyServiceCreateRequest) GetMetadataOk() (*SshPublicKeyServiceCreateRequestMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *SshPublicKeyServiceCreateRequest) SetMetadata(v SshPublicKeyServiceCreateRequestMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *SshPublicKeyServiceCreateRequest) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *SshPublicKeyServiceCreateRequest) GetSpec() ProtoSshPublicKeySpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *SshPublicKeyServiceCreateRequest) GetSpecOk() (*ProtoSshPublicKeySpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *SshPublicKeyServiceCreateRequest) SetSpec(v ProtoSshPublicKeySpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *SshPublicKeyServiceCreateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


