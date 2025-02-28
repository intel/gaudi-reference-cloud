# InstanceServiceUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**InstanceServiceUpdateRequestMetadata**](InstanceServiceUpdateRequestMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoInstanceSpec**](ProtoInstanceSpec.md) |  | [optional] 

## Methods

### NewInstanceServiceUpdateRequest

`func NewInstanceServiceUpdateRequest() *InstanceServiceUpdateRequest`

NewInstanceServiceUpdateRequest instantiates a new InstanceServiceUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstanceServiceUpdateRequestWithDefaults

`func NewInstanceServiceUpdateRequestWithDefaults() *InstanceServiceUpdateRequest`

NewInstanceServiceUpdateRequestWithDefaults instantiates a new InstanceServiceUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *InstanceServiceUpdateRequest) GetMetadata() InstanceServiceUpdateRequestMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *InstanceServiceUpdateRequest) GetMetadataOk() (*InstanceServiceUpdateRequestMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *InstanceServiceUpdateRequest) SetMetadata(v InstanceServiceUpdateRequestMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *InstanceServiceUpdateRequest) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *InstanceServiceUpdateRequest) GetSpec() ProtoInstanceSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InstanceServiceUpdateRequest) GetSpecOk() (*ProtoInstanceSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InstanceServiceUpdateRequest) SetSpec(v ProtoInstanceSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InstanceServiceUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


