# ProtoMachineImage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**ProtoMachineImageMetadata**](ProtoMachineImageMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoMachineImageSpec**](ProtoMachineImageSpec.md) |  | [optional] 

## Methods

### NewProtoMachineImage

`func NewProtoMachineImage() *ProtoMachineImage`

NewProtoMachineImage instantiates a new ProtoMachineImage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoMachineImageWithDefaults

`func NewProtoMachineImageWithDefaults() *ProtoMachineImage`

NewProtoMachineImageWithDefaults instantiates a new ProtoMachineImage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *ProtoMachineImage) GetMetadata() ProtoMachineImageMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ProtoMachineImage) GetMetadataOk() (*ProtoMachineImageMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ProtoMachineImage) SetMetadata(v ProtoMachineImageMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *ProtoMachineImage) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *ProtoMachineImage) GetSpec() ProtoMachineImageSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ProtoMachineImage) GetSpecOk() (*ProtoMachineImageSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ProtoMachineImage) SetSpec(v ProtoMachineImageSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *ProtoMachineImage) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


