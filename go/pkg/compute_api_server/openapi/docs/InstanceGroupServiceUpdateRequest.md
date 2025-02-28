# InstanceGroupServiceUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**InstanceGroupServiceUpdateRequestMetadata**](InstanceGroupServiceUpdateRequestMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoInstanceGroupSpec**](ProtoInstanceGroupSpec.md) |  | [optional] 

## Methods

### NewInstanceGroupServiceUpdateRequest

`func NewInstanceGroupServiceUpdateRequest() *InstanceGroupServiceUpdateRequest`

NewInstanceGroupServiceUpdateRequest instantiates a new InstanceGroupServiceUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstanceGroupServiceUpdateRequestWithDefaults

`func NewInstanceGroupServiceUpdateRequestWithDefaults() *InstanceGroupServiceUpdateRequest`

NewInstanceGroupServiceUpdateRequestWithDefaults instantiates a new InstanceGroupServiceUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *InstanceGroupServiceUpdateRequest) GetMetadata() InstanceGroupServiceUpdateRequestMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *InstanceGroupServiceUpdateRequest) GetMetadataOk() (*InstanceGroupServiceUpdateRequestMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *InstanceGroupServiceUpdateRequest) SetMetadata(v InstanceGroupServiceUpdateRequestMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *InstanceGroupServiceUpdateRequest) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *InstanceGroupServiceUpdateRequest) GetSpec() ProtoInstanceGroupSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InstanceGroupServiceUpdateRequest) GetSpecOk() (*ProtoInstanceGroupSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InstanceGroupServiceUpdateRequest) SetSpec(v ProtoInstanceGroupSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InstanceGroupServiceUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


