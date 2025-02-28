# ProtoInstance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**ProtoInstanceMetadata**](ProtoInstanceMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoInstanceSpec**](ProtoInstanceSpec.md) |  | [optional] 
**Status** | Pointer to [**ProtoInstanceStatus**](ProtoInstanceStatus.md) |  | [optional] 

## Methods

### NewProtoInstance

`func NewProtoInstance() *ProtoInstance`

NewProtoInstance instantiates a new ProtoInstance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceWithDefaults

`func NewProtoInstanceWithDefaults() *ProtoInstance`

NewProtoInstanceWithDefaults instantiates a new ProtoInstance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *ProtoInstance) GetMetadata() ProtoInstanceMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ProtoInstance) GetMetadataOk() (*ProtoInstanceMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ProtoInstance) SetMetadata(v ProtoInstanceMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *ProtoInstance) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *ProtoInstance) GetSpec() ProtoInstanceSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ProtoInstance) GetSpecOk() (*ProtoInstanceSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ProtoInstance) SetSpec(v ProtoInstanceSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *ProtoInstance) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetStatus

`func (o *ProtoInstance) GetStatus() ProtoInstanceStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProtoInstance) GetStatusOk() (*ProtoInstanceStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProtoInstance) SetStatus(v ProtoInstanceStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ProtoInstance) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


