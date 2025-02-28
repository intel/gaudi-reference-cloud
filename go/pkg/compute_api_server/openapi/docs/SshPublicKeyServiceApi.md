# \SshPublicKeyServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SshPublicKeyServiceCreate**](SshPublicKeyServiceApi.md#SshPublicKeyServiceCreate) | **Post** /v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys | Store an SSH public key.
[**SshPublicKeyServiceDelete**](SshPublicKeyServiceApi.md#SshPublicKeyServiceDelete) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys/id/{metadata.resourceId} | Delete an SSH public key.
[**SshPublicKeyServiceDelete2**](SshPublicKeyServiceApi.md#SshPublicKeyServiceDelete2) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys/name/{metadata.name} | Delete an SSH public key.
[**SshPublicKeyServiceGet**](SshPublicKeyServiceApi.md#SshPublicKeyServiceGet) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys/id/{metadata.resourceId} | Retrieve a stored SSH public key.
[**SshPublicKeyServiceGet2**](SshPublicKeyServiceApi.md#SshPublicKeyServiceGet2) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys/name/{metadata.name} | Retrieve a stored SSH public key.
[**SshPublicKeyServicePing**](SshPublicKeyServiceApi.md#SshPublicKeyServicePing) | **Get** /v1/ping | Ping always returns a successful response by the service implementation. It can be used for testing connectivity to the service.
[**SshPublicKeyServiceSearch**](SshPublicKeyServiceApi.md#SshPublicKeyServiceSearch) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys | Get a list of stored SSH public keys.
[**SshPublicKeyServiceSearchStream**](SshPublicKeyServiceApi.md#SshPublicKeyServiceSearchStream) | **Post** /proto.SshPublicKeyService/SearchStream | List stored SSH public keys as a stream. Warning: This does not work with OpenAPI client. Internal-use only.



## SshPublicKeyServiceCreate

> ProtoSshPublicKey SshPublicKeyServiceCreate(ctx, metadataCloudAccountId).Body(body).Execute()

Store an SSH public key.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    metadataCloudAccountId := "metadataCloudAccountId_example" // string | 
    body := *openapiclient.NewSshPublicKeyServiceCreateRequest() // SshPublicKeyServiceCreateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshPublicKeyServiceApi.SshPublicKeyServiceCreate(context.Background(), metadataCloudAccountId).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshPublicKeyServiceApi.SshPublicKeyServiceCreate``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SshPublicKeyServiceCreate`: ProtoSshPublicKey
    fmt.Fprintf(os.Stdout, "Response from `SshPublicKeyServiceApi.SshPublicKeyServiceCreate`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiSshPublicKeyServiceCreateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**SshPublicKeyServiceCreateRequest**](SshPublicKeyServiceCreateRequest.md) |  | 

### Return type

[**ProtoSshPublicKey**](ProtoSshPublicKey.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SshPublicKeyServiceDelete

> map[string]interface{} SshPublicKeyServiceDelete(ctx, metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).Execute()

Delete an SSH public key.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    metadataCloudAccountId := "metadataCloudAccountId_example" // string | 
    metadataResourceId := "metadataResourceId_example" // string | 
    metadataName := "metadataName_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshPublicKeyServiceApi.SshPublicKeyServiceDelete(context.Background(), metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshPublicKeyServiceApi.SshPublicKeyServiceDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SshPublicKeyServiceDelete`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `SshPublicKeyServiceApi.SshPublicKeyServiceDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiSshPublicKeyServiceDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataName** | **string** |  | 

### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SshPublicKeyServiceDelete2

> map[string]interface{} SshPublicKeyServiceDelete2(ctx, metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).Execute()

Delete an SSH public key.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    metadataCloudAccountId := "metadataCloudAccountId_example" // string | 
    metadataName := "metadataName_example" // string | 
    metadataResourceId := "metadataResourceId_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshPublicKeyServiceApi.SshPublicKeyServiceDelete2(context.Background(), metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshPublicKeyServiceApi.SshPublicKeyServiceDelete2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SshPublicKeyServiceDelete2`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `SshPublicKeyServiceApi.SshPublicKeyServiceDelete2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiSshPublicKeyServiceDelete2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataResourceId** | **string** |  | 

### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SshPublicKeyServiceGet

> ProtoSshPublicKey SshPublicKeyServiceGet(ctx, metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).Execute()

Retrieve a stored SSH public key.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    metadataCloudAccountId := "metadataCloudAccountId_example" // string | 
    metadataResourceId := "metadataResourceId_example" // string | 
    metadataName := "metadataName_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshPublicKeyServiceApi.SshPublicKeyServiceGet(context.Background(), metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshPublicKeyServiceApi.SshPublicKeyServiceGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SshPublicKeyServiceGet`: ProtoSshPublicKey
    fmt.Fprintf(os.Stdout, "Response from `SshPublicKeyServiceApi.SshPublicKeyServiceGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiSshPublicKeyServiceGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataName** | **string** |  | 

### Return type

[**ProtoSshPublicKey**](ProtoSshPublicKey.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SshPublicKeyServiceGet2

> ProtoSshPublicKey SshPublicKeyServiceGet2(ctx, metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).Execute()

Retrieve a stored SSH public key.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    metadataCloudAccountId := "metadataCloudAccountId_example" // string | 
    metadataName := "metadataName_example" // string | 
    metadataResourceId := "metadataResourceId_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshPublicKeyServiceApi.SshPublicKeyServiceGet2(context.Background(), metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshPublicKeyServiceApi.SshPublicKeyServiceGet2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SshPublicKeyServiceGet2`: ProtoSshPublicKey
    fmt.Fprintf(os.Stdout, "Response from `SshPublicKeyServiceApi.SshPublicKeyServiceGet2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiSshPublicKeyServiceGet2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataResourceId** | **string** |  | 

### Return type

[**ProtoSshPublicKey**](ProtoSshPublicKey.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SshPublicKeyServicePing

> map[string]interface{} SshPublicKeyServicePing(ctx).Execute()

Ping always returns a successful response by the service implementation. It can be used for testing connectivity to the service.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshPublicKeyServiceApi.SshPublicKeyServicePing(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshPublicKeyServiceApi.SshPublicKeyServicePing``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SshPublicKeyServicePing`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `SshPublicKeyServiceApi.SshPublicKeyServicePing`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiSshPublicKeyServicePingRequest struct via the builder pattern


### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SshPublicKeyServiceSearch

> ProtoSshPublicKeySearchResponse SshPublicKeyServiceSearch(ctx, metadataCloudAccountId).Execute()

Get a list of stored SSH public keys.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    metadataCloudAccountId := "metadataCloudAccountId_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshPublicKeyServiceApi.SshPublicKeyServiceSearch(context.Background(), metadataCloudAccountId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshPublicKeyServiceApi.SshPublicKeyServiceSearch``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SshPublicKeyServiceSearch`: ProtoSshPublicKeySearchResponse
    fmt.Fprintf(os.Stdout, "Response from `SshPublicKeyServiceApi.SshPublicKeyServiceSearch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiSshPublicKeyServiceSearchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProtoSshPublicKeySearchResponse**](ProtoSshPublicKeySearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SshPublicKeyServiceSearchStream

> StreamResultOfProtoSshPublicKey SshPublicKeyServiceSearchStream(ctx).Body(body).Execute()

List stored SSH public keys as a stream. Warning: This does not work with OpenAPI client. Internal-use only.

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    body := *openapiclient.NewProtoSshPublicKeySearchRequest() // ProtoSshPublicKeySearchRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SshPublicKeyServiceApi.SshPublicKeyServiceSearchStream(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SshPublicKeyServiceApi.SshPublicKeyServiceSearchStream``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SshPublicKeyServiceSearchStream`: StreamResultOfProtoSshPublicKey
    fmt.Fprintf(os.Stdout, "Response from `SshPublicKeyServiceApi.SshPublicKeyServiceSearchStream`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSshPublicKeyServiceSearchStreamRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ProtoSshPublicKeySearchRequest**](ProtoSshPublicKeySearchRequest.md) |  | 

### Return type

[**StreamResultOfProtoSshPublicKey**](StreamResultOfProtoSshPublicKey.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

