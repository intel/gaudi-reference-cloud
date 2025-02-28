# InstanceGroupServiceCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**InstanceGroupServiceCreateRequestMetadata**](InstanceGroupServiceCreateRequestMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoInstanceGroupSpec**](ProtoInstanceGroupSpec.md) |  | [optional] 

## Methods

### NewInstanceGroupServiceCreateRequest

`func NewInstanceGroupServiceCreateRequest() *InstanceGroupServiceCreateRequest`

NewInstanceGroupServiceCreateRequest instantiates a new InstanceGroupServiceCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstanceGroupServiceCreateRequestWithDefaults

`func NewInstanceGroupServiceCreateRequestWithDefaults() *InstanceGroupServiceCreateRequest`

NewInstanceGroupServiceCreateRequestWithDefaults instantiates a new InstanceGroupServiceCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *InstanceGroupServiceCreateRequest) GetMetadata() InstanceGroupServiceCreateRequestMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *InstanceGroupServiceCreateRequest) GetMetadataOk() (*InstanceGroupServiceCreateRequestMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *InstanceGroupServiceCreateRequest) SetMetadata(v InstanceGroupServiceCreateRequestMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *InstanceGroupServiceCreateRequest) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *InstanceGroupServiceCreateRequest) GetSpec() ProtoInstanceGroupSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InstanceGroupServiceCreateRequest) GetSpecOk() (*ProtoInstanceGroupSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InstanceGroupServiceCreateRequest) SetSpec(v ProtoInstanceGroupSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InstanceGroupServiceCreateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


