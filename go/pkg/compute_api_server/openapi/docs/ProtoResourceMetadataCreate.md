# ProtoResourceMetadataCreate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CloudAccountId** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** | If Name is not empty, it must be unique within the cloudAccountId. It will be generated if empty. | [optional] 
**Labels** | Pointer to **map[string]string** | Not implemented. | [optional] 

## Methods

### NewProtoResourceMetadataCreate

`func NewProtoResourceMetadataCreate() *ProtoResourceMetadataCreate`

NewProtoResourceMetadataCreate instantiates a new ProtoResourceMetadataCreate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoResourceMetadataCreateWithDefaults

`func NewProtoResourceMetadataCreateWithDefaults() *ProtoResourceMetadataCreate`

NewProtoResourceMetadataCreateWithDefaults instantiates a new ProtoResourceMetadataCreate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCloudAccountId

`func (o *ProtoResourceMetadataCreate) GetCloudAccountId() string`

GetCloudAccountId returns the CloudAccountId field if non-nil, zero value otherwise.

### GetCloudAccountIdOk

`func (o *ProtoResourceMetadataCreate) GetCloudAccountIdOk() (*string, bool)`

GetCloudAccountIdOk returns a tuple with the CloudAccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloudAccountId

`func (o *ProtoResourceMetadataCreate) SetCloudAccountId(v string)`

SetCloudAccountId sets CloudAccountId field to given value.

### HasCloudAccountId

`func (o *ProtoResourceMetadataCreate) HasCloudAccountId() bool`

HasCloudAccountId returns a boolean if a field has been set.

### GetName

`func (o *ProtoResourceMetadataCreate) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoResourceMetadataCreate) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoResourceMetadataCreate) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoResourceMetadataCreate) HasName() bool`

HasName returns a boolean if a field has been set.

### GetLabels

`func (o *ProtoResourceMetadataCreate) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ProtoResourceMetadataCreate) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ProtoResourceMetadataCreate) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ProtoResourceMetadataCreate) HasLabels() bool`

HasLabels returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


