# ProtoInstanceGroup

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**ProtoInstanceGroupMetadata**](ProtoInstanceGroupMetadata.md) |  | [optional] 
**Spec** | Pointer to [**ProtoInstanceGroupSpec**](ProtoInstanceGroupSpec.md) |  | [optional] 
**Status** | Pointer to [**ProtoInstanceGroupStatus**](ProtoInstanceGroupStatus.md) |  | [optional] 

## Methods

### NewProtoInstanceGroup

`func NewProtoInstanceGroup() *ProtoInstanceGroup`

NewProtoInstanceGroup instantiates a new ProtoInstanceGroup object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceGroupWithDefaults

`func NewProtoInstanceGroupWithDefaults() *ProtoInstanceGroup`

NewProtoInstanceGroupWithDefaults instantiates a new ProtoInstanceGroup object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *ProtoInstanceGroup) GetMetadata() ProtoInstanceGroupMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ProtoInstanceGroup) GetMetadataOk() (*ProtoInstanceGroupMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ProtoInstanceGroup) SetMetadata(v ProtoInstanceGroupMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *ProtoInstanceGroup) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetSpec

`func (o *ProtoInstanceGroup) GetSpec() ProtoInstanceGroupSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ProtoInstanceGroup) GetSpecOk() (*ProtoInstanceGroupSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ProtoInstanceGroup) SetSpec(v ProtoInstanceGroupSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *ProtoInstanceGroup) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetStatus

`func (o *ProtoInstanceGroup) GetStatus() ProtoInstanceGroupStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProtoInstanceGroup) GetStatusOk() (*ProtoInstanceGroupStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProtoInstanceGroup) SetStatus(v ProtoInstanceGroupStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ProtoInstanceGroup) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


