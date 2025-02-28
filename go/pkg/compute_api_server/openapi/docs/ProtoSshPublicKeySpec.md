# ProtoSshPublicKeySpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SshPublicKey** | Pointer to **string** | SSH public key in authorized_keys format \&quot;ssh-rsa ... comment\&quot;. | [optional] 
**OwnerEmail** | Pointer to **string** |  | [optional] 

## Methods

### NewProtoSshPublicKeySpec

`func NewProtoSshPublicKeySpec() *ProtoSshPublicKeySpec`

NewProtoSshPublicKeySpec instantiates a new ProtoSshPublicKeySpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProtoSshPublicKeySpecWithDefaults

`func NewProtoSshPublicKeySpecWithDefaults() *ProtoSshPublicKeySpec`

NewProtoSshPublicKeySpecWithDefaults instantiates a new ProtoSshPublicKeySpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSshPublicKey

`func (o *ProtoSshPublicKeySpec) GetSshPublicKey() string`

GetSshPublicKey returns the SshPublicKey field if non-nil, zero value otherwise.

### GetSshPublicKeyOk

`func (o *ProtoSshPublicKeySpec) GetSshPublicKeyOk() (*string, bool)`

GetSshPublicKeyOk returns a tuple with the SshPublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSshPublicKey

`func (o *ProtoSshPublicKeySpec) SetSshPublicKey(v string)`

SetSshPublicKey sets SshPublicKey field to given value.

### HasSshPublicKey

`func (o *ProtoSshPublicKeySpec) HasSshPublicKey() bool`

HasSshPublicKey returns a boolean if a field has been set.

### GetOwnerEmail

`func (o *ProtoSshPublicKeySpec) GetOwnerEmail() string`

GetOwnerEmail returns the OwnerEmail field if non-nil, zero value otherwise.

### GetOwnerEmailOk

`func (o *ProtoSshPublicKeySpec) GetOwnerEmailOk() (*string, bool)`

GetOwnerEmailOk returns a tuple with the OwnerEmail field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwnerEmail

`func (o *ProtoSshPublicKeySpec) SetOwnerEmail(v string)`

SetOwnerEmail sets OwnerEmail field to given value.

### HasOwnerEmail

`func (o *ProtoSshPublicKeySpec) HasOwnerEmail() bool`

HasOwnerEmail returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


