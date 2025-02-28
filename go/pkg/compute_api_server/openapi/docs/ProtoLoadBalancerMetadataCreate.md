# ProtoLoadBalancerMetadataCreate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CloudAccountId** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** | Name will be generated if empty. | [optional] 
**Labels** | Pointer to **map[string]string** | Map of string keys and values that can be used to organize and categorize load balancers. | [optional] 
**Reserved1** | Pointer to **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | [optional] 

## Methods

### NewProtoLoadBalancerMetadataCreate

`func NewProtoLoadBalancerMetadataCreate() *ProtoLoadBalancerMetadataCreate`

NewProtoLoadBalancerMetadataCreate instantiates a new ProtoLoadBalancerMetadataCreate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoLoadBalancerMetadataCreateWithDefaults

`func NewProtoLoadBalancerMetadataCreateWithDefaults() *ProtoLoadBalancerMetadataCreate`

NewProtoLoadBalancerMetadataCreateWithDefaults instantiates a new ProtoLoadBalancerMetadataCreate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCloudAccountId

`func (o *ProtoLoadBalancerMetadataCreate) GetCloudAccountId() string`

GetCloudAccountId returns the CloudAccountId field if non-nil, zero value otherwise.

### GetCloudAccountIdOk

`func (o *ProtoLoadBalancerMetadataCreate) GetCloudAccountIdOk() (*string, bool)`

GetCloudAccountIdOk returns a tuple with the CloudAccountId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloudAccountId

`func (o *ProtoLoadBalancerMetadataCreate) SetCloudAccountId(v string)`

SetCloudAccountId sets CloudAccountId field to given value.

### HasCloudAccountId

`func (o *ProtoLoadBalancerMetadataCreate) HasCloudAccountId() bool`

HasCloudAccountId returns a boolean if a field has been set.

### GetName

`func (o *ProtoLoadBalancerMetadataCreate) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProtoLoadBalancerMetadataCreate) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProtoLoadBalancerMetadataCreate) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProtoLoadBalancerMetadataCreate) HasName() bool`

HasName returns a boolean if a field has been set.

### GetLabels

`func (o *ProtoLoadBalancerMetadataCreate) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ProtoLoadBalancerMetadataCreate) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ProtoLoadBalancerMetadataCreate) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ProtoLoadBalancerMetadataCreate) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetReserved1

`func (o *ProtoLoadBalancerMetadataCreate) GetReserved1() string`

GetReserved1 returns the Reserved1 field if non-nil, zero value otherwise.

### GetReserved1Ok

`func (o *ProtoLoadBalancerMetadataCreate) GetReserved1Ok() (*string, bool)`

GetReserved1Ok returns a tuple with the Reserved1 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReserved1

`func (o *ProtoLoadBalancerMetadataCreate) SetReserved1(v string)`

SetReserved1 sets Reserved1 field to given value.

### HasReserved1

`func (o *ProtoLoadBalancerMetadataCreate) HasReserved1() bool`

HasReserved1 returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


