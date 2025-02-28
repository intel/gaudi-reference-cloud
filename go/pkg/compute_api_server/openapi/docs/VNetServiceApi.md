# \VNetServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**VNetServiceDelete**](VNetServiceApi.md#VNetServiceDelete) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/vnets/id/{metadata.resourceId} | Delete an VNet from the DB. Returns FailedPrecondition if VNet has running instances or other consumed IP addresses.
[**VNetServiceDelete2**](VNetServiceApi.md#VNetServiceDelete2) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/vnets/name/{metadata.name} | Delete an VNet from the DB. Returns FailedPrecondition if VNet has running instances or other consumed IP addresses.
[**VNetServiceGet**](VNetServiceApi.md#VNetServiceGet) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/vnets/id/{metadata.resourceId} | Retrieve a VNet record from DB
[**VNetServiceGet2**](VNetServiceApi.md#VNetServiceGet2) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/vnets/name/{metadata.name} | Retrieve a VNet record from DB
[**VNetServicePing**](VNetServiceApi.md#VNetServicePing) | **Post** /proto.VNetService/Ping | Ping always returns a successful response by the service implementation. It can be used for testing connectivity to the service.
[**VNetServicePut**](VNetServiceApi.md#VNetServicePut) | **Post** /v1/cloudaccounts/{metadata.cloudAccountId}/vnets | Create or update a VNet.
[**VNetServiceSearch**](VNetServiceApi.md#VNetServiceSearch) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/vnets | Get a list of stored VNets.
[**VNetServiceSearchStream**](VNetServiceApi.md#VNetServiceSearchStream) | **Post** /proto.VNetService/SearchStream | List stored VNets as a stream.



## VNetServiceDelete

> map[string]interface{} VNetServiceDelete(ctx, metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).Execute()

Delete an VNet from the DB. Returns FailedPrecondition if VNet has running instances or other consumed IP addresses.

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
    resp, r, err := apiClient.VNetServiceApi.VNetServiceDelete(context.Background(), metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `VNetServiceApi.VNetServiceDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `VNetServiceDelete`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `VNetServiceApi.VNetServiceDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiVNetServiceDeleteRequest struct via the builder pattern


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


## VNetServiceDelete2

> map[string]interface{} VNetServiceDelete2(ctx, metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).Execute()

Delete an VNet from the DB. Returns FailedPrecondition if VNet has running instances or other consumed IP addresses.

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
    resp, r, err := apiClient.VNetServiceApi.VNetServiceDelete2(context.Background(), metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `VNetServiceApi.VNetServiceDelete2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `VNetServiceDelete2`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `VNetServiceApi.VNetServiceDelete2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiVNetServiceDelete2Request struct via the builder pattern


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


## VNetServiceGet

> ProtoVNet VNetServiceGet(ctx, metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).Execute()

Retrieve a VNet record from DB

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
    resp, r, err := apiClient.VNetServiceApi.VNetServiceGet(context.Background(), metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `VNetServiceApi.VNetServiceGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `VNetServiceGet`: ProtoVNet
    fmt.Fprintf(os.Stdout, "Response from `VNetServiceApi.VNetServiceGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiVNetServiceGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataName** | **string** |  | 

### Return type

[**ProtoVNet**](ProtoVNet.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## VNetServiceGet2

> ProtoVNet VNetServiceGet2(ctx, metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).Execute()

Retrieve a VNet record from DB

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
    resp, r, err := apiClient.VNetServiceApi.VNetServiceGet2(context.Background(), metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `VNetServiceApi.VNetServiceGet2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `VNetServiceGet2`: ProtoVNet
    fmt.Fprintf(os.Stdout, "Response from `VNetServiceApi.VNetServiceGet2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiVNetServiceGet2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataResourceId** | **string** |  | 

### Return type

[**ProtoVNet**](ProtoVNet.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## VNetServicePing

> map[string]interface{} VNetServicePing(ctx).Body(body).Execute()

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
    body := map[string]interface{}{ ... } // map[string]interface{} | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.VNetServiceApi.VNetServicePing(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `VNetServiceApi.VNetServicePing``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `VNetServicePing`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `VNetServiceApi.VNetServicePing`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiVNetServicePingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | **map[string]interface{}** |  | 

### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## VNetServicePut

> ProtoVNet VNetServicePut(ctx, metadataCloudAccountId).Body(body).Execute()

Create or update a VNet.

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
    body := *openapiclient.NewVNetServicePutRequest() // VNetServicePutRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.VNetServiceApi.VNetServicePut(context.Background(), metadataCloudAccountId).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `VNetServiceApi.VNetServicePut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `VNetServicePut`: ProtoVNet
    fmt.Fprintf(os.Stdout, "Response from `VNetServiceApi.VNetServicePut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiVNetServicePutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**VNetServicePutRequest**](VNetServicePutRequest.md) |  | 

### Return type

[**ProtoVNet**](ProtoVNet.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## VNetServiceSearch

> ProtoVNetSearchResponse VNetServiceSearch(ctx, metadataCloudAccountId).Execute()

Get a list of stored VNets.

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
    resp, r, err := apiClient.VNetServiceApi.VNetServiceSearch(context.Background(), metadataCloudAccountId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `VNetServiceApi.VNetServiceSearch``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `VNetServiceSearch`: ProtoVNetSearchResponse
    fmt.Fprintf(os.Stdout, "Response from `VNetServiceApi.VNetServiceSearch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiVNetServiceSearchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProtoVNetSearchResponse**](ProtoVNetSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## VNetServiceSearchStream

> StreamResultOfProtoVNet VNetServiceSearchStream(ctx).Body(body).Execute()

List stored VNets as a stream.

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
    body := *openapiclient.NewProtoVNetSearchRequest() // ProtoVNetSearchRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.VNetServiceApi.VNetServiceSearchStream(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `VNetServiceApi.VNetServiceSearchStream``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `VNetServiceSearchStream`: StreamResultOfProtoVNet
    fmt.Fprintf(os.Stdout, "Response from `VNetServiceApi.VNetServiceSearchStream`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiVNetServiceSearchStreamRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ProtoVNetSearchRequest**](ProtoVNetSearchRequest.md) |  | 

### Return type

[**StreamResultOfProtoVNet**](StreamResultOfProtoVNet.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

