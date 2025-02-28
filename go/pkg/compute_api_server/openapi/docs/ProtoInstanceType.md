# ProtoInstanceType

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**ProtoInstanceTypeMetadata**](ProtoInstanceTypeMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoInstanceTypeSpec**](ProtoInstanceTypeSpec.md) |  | [optional] 

## Methods

### NewProtoInstanceType

`func NewProtoInstanceType() *ProtoInstanceType`

NewProtoInstanceType instantiates a new ProtoInstanceType object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceTypeWithDefaults

`func NewProtoInstanceTypeWithDefaults() *ProtoInstanceType`

NewProtoInstanceTypeWithDefaults instantiates a new ProtoInstanceType object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *ProtoInstanceType) GetMetadata() ProtoInstanceTypeMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ProtoInstanceType) GetMetadataOk() (*ProtoInstanceTypeMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ProtoInstanceType) SetMetadata(v ProtoInstanceTypeMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *ProtoInstanceType) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *ProtoInstanceType) GetSpec() ProtoInstanceTypeSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ProtoInstanceType) GetSpecOk() (*ProtoInstanceTypeSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ProtoInstanceType) SetSpec(v ProtoInstanceTypeSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *ProtoInstanceType) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


