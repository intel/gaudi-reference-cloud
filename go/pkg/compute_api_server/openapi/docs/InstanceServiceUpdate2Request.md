# InstanceServiceUpdate2Request

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**InstanceServiceUpdate2RequestMetadata**](InstanceServiceUpdate2RequestMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoInstanceSpec**](ProtoInstanceSpec.md) |  | [optional] 

## Methods

### NewInstanceServiceUpdate2Request

`func NewInstanceServiceUpdate2Request() *InstanceServiceUpdate2Request`

NewInstanceServiceUpdate2Request instantiates a new InstanceServiceUpdate2Request object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInstanceServiceUpdate2RequestWithDefaults

`func NewInstanceServiceUpdate2RequestWithDefaults() *InstanceServiceUpdate2Request`

NewInstanceServiceUpdate2RequestWithDefaults instantiates a new InstanceServiceUpdate2Request object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *InstanceServiceUpdate2Request) GetMetadata() InstanceServiceUpdate2RequestMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *InstanceServiceUpdate2Request) GetMetadataOk() (*InstanceServiceUpdate2RequestMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *InstanceServiceUpdate2Request) SetMetadata(v InstanceServiceUpdate2RequestMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *InstanceServiceUpdate2Request) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *InstanceServiceUpdate2Request) GetSpec() ProtoInstanceSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InstanceServiceUpdate2Request) GetSpecOk() (*ProtoInstanceSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InstanceServiceUpdate2Request) SetSpec(v ProtoInstanceSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InstanceServiceUpdate2Request) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


