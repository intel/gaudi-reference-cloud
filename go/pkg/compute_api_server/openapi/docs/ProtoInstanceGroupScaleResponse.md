# ProtoInstanceGroupScaleResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metadata** | Pointer to [**ProtoInstanceGroupMetadata**](ProtoInstanceGroupMetadata.md) |  | [optional] 
**Status** | Pointer to [**ProtoInstanceGroupScaleStatus**](ProtoInstanceGroupScaleStatus.md) |  | [optional] 

## Methods

### NewProtoInstanceGroupScaleResponse

`func NewProtoInstanceGroupScaleResponse() *ProtoInstanceGroupScaleResponse`

NewProtoInstanceGroupScaleResponse instantiates a new ProtoInstanceGroupScaleResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceGroupScaleResponseWithDefaults

`func NewProtoInstanceGroupScaleResponseWithDefaults() *ProtoInstanceGroupScaleResponse`

NewProtoInstanceGroupScaleResponseWithDefaults instantiates a new ProtoInstanceGroupScaleResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetadata

`func (o *ProtoInstanceGroupScaleResponse) GetMetadata() ProtoInstanceGroupMetadata`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ProtoInstanceGroupScaleResponse) GetMetadataOk() (*ProtoInstanceGroupMetadata, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ProtoInstanceGroupScaleResponse) SetMetadata(v ProtoInstanceGroupMetadata)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *ProtoInstanceGroupScaleResponse) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.

### GetStatus

`func (o *ProtoInstanceGroupScaleResponse) GetStatus() ProtoInstanceGroupScaleStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProtoInstanceGroupScaleResponse) GetStatusOk() (*ProtoInstanceGroupScaleStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProtoInstanceGroupScaleResponse) SetStatus(v ProtoInstanceGroupScaleStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ProtoInstanceGroupScaleResponse) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


