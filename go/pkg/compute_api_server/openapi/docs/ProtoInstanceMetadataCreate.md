# ProtoInstanceMetadataCreate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CloudAccountId** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** | Name will be generated if empty. | [optional] 
**Labels** | Pointer to **map[string]string** | Map of string keys and values that can be used to organize and categorize instances. This is also used by TopologySpreadConstraints. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 
**ProductId** | Pointer to **string** |  | [optional] 

## Methods

### NewProtoInstanceMetadataCreate

`func NewProtoInstanceMetadataCreate() *ProtoInstanceMetadataCreate`

NewProtoInstanceMetadataCreate instantiates a new ProtoInstanceMetadataCreate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoInstanceMetadataCreateWithDefaults

`func NewProtoInstanceMetadataCreateWithDefaults() *ProtoInstanceMetadataCreate`

NewProtoInstanceMetadataCreateWithDefaults instantiates a new ProtoInstanceMetadataCreate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCloudAccountId

`func (o *ProtoInstanceMetadataCreate) GetCloudAccountId() string`

GetCloudAccountId returns the CloudAccountId field if non-nil, zero value otherwise.

### GetCloudAccountIdOk

`func (o *ProtoInstanceMetadataCreate) GetCloudAccountIdOk() (*string, bool)`

GetCloudAccountIdOk returns a tuple with the CloudAccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloudAccountId

`func (o *ProtoInstanceMetadataCreate) SetCloudAccountId(v string)`

SetCloudAccountId sets CloudAccountId field to given value.

### HasCloudAccountId

`func (o *ProtoInstanceMetadataCreate) HasCloudAccountId() bool`

HasCloudAccountId returns a boolean if a field has been set.

### GetName

`func (o *ProtoInstanceMetadataCreate) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoInstanceMetadataCreate) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoInstanceMetadataCreate) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoInstanceMetadataCreate) HasName() bool`

HasName returns a boolean if a field has been set.

### GetLabels

`func (o *ProtoInstanceMetadataCreate) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ProtoInstanceMetadataCreate) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ProtoInstanceMetadataCreate) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ProtoInstanceMetadataCreate) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetReserved1

`func (o *ProtoInstanceMetadataCreate) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *ProtoInstanceMetadataCreate) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *ProtoInstanceMetadataCreate) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *ProtoInstanceMetadataCreate) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.

### GetProductId

`func (o *ProtoInstanceMetadataCreate) GetProductId() string`

GetProductId returns the ProductId field if non-nil, zero value otherwise.

### GetProductIdOk

`func (o *ProtoInstanceMetadataCreate) GetProductIdOk() (*string, bool)`

GetProductIdOk returns a tuple with the ProductId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProductId

`func (o *ProtoInstanceMetadataCreate) SetProductId(v string)`

SetProductId sets ProductId field to given value.

### HasProductId

`func (o *ProtoInstanceMetadataCreate) HasProductId() bool`

HasProductId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


