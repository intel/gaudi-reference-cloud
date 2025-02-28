# InstanceGroupServiceScaleUpRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**InstanceGroupServiceScaleUpRequestMetadata**](InstanceGroupServiceScaleUpRequestMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoInstanceGroupSpec**](ProtoInstanceGroupSpec.md) |  | [optional] 

## Methods

### NewInstanceGroupServiceScaleUpRequest

`func NewInstanceGroupServiceScaleUpRequest() *InstanceGroupServiceScaleUpRequest`

NewInstanceGroupServiceScaleUpRequest instantiates a new InstanceGroupServiceScaleUpRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstanceGroupServiceScaleUpRequestWithDefaults

`func NewInstanceGroupServiceScaleUpRequestWithDefaults() *InstanceGroupServiceScaleUpRequest`

NewInstanceGroupServiceScaleUpRequestWithDefaults instantiates a new InstanceGroupServiceScaleUpRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *InstanceGroupServiceScaleUpRequest) GetMetadata() InstanceGroupServiceScaleUpRequestMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *InstanceGroupServiceScaleUpRequest) GetMetadataOk() (*InstanceGroupServiceScaleUpRequestMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *InstanceGroupServiceScaleUpRequest) SetMetadata(v InstanceGroupServiceScaleUpRequestMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *InstanceGroupServiceScaleUpRequest) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *InstanceGroupServiceScaleUpRequest) GetSpec() ProtoInstanceGroupSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InstanceGroupServiceScaleUpRequest) GetSpecOk() (*ProtoInstanceGroupSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InstanceGroupServiceScaleUpRequest) SetSpec(v ProtoInstanceGroupSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InstanceGroupServiceScaleUpRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


