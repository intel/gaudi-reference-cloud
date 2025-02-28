# SshPublicKeyServiceCreateRequestMetadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | If Name is not empty, it must be unique within the cloudAccountId. It will be generated if empty. | [optional] 
**Labels** | Pointer to **map[string]string** | Not implemented. | [optional] 

## Methods

### NewSshPublicKeyServiceCreateRequestMetadata

`func NewSshPublicKeyServiceCreateRequestMetadata() *SshPublicKeyServiceCreateRequestMetadata`

NewSshPublicKeyServiceCreateRequestMetadata instantiates a new SshPublicKeyServiceCreateRequestMetadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSshPublicKeyServiceCreateRequestMetadataWithDefaults

`func NewSshPublicKeyServiceCreateRequestMetadataWithDefaults() *SshPublicKeyServiceCreateRequestMetadata`

NewSshPublicKeyServiceCreateRequestMetadataWithDefaults instantiates a new SshPublicKeyServiceCreateRequestMetadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *SshPublicKeyServiceCreateRequestMetadata) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SshPublicKeyServiceCreateRequestMetadata) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SshPublicKeyServiceCreateRequestMetadata) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *SshPublicKeyServiceCreateRequestMetadata) HasName() bool`

HasName returns a boolean if a field has been set.

### GetLabels

`func (o *SshPublicKeyServiceCreateRequestMetadata) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *SshPublicKeyServiceCreateRequestMetadata) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *SshPublicKeyServiceCreateRequestMetadata) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *SshPublicKeyServiceCreateRequestMetadata) HasLabels() bool`

HasLabels returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


