# \InstanceTypeServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**InstanceTypeServiceDelete**](InstanceTypeServiceApi.md#InstanceTypeServiceDelete) | **Post** /proto.InstanceTypeService/Delete | Delete an instance typ.
[**InstanceTypeServiceGet**](InstanceTypeServiceApi.md#InstanceTypeServiceGet) | **Get** /v1/instancetypes/{metadata.name} | Get an instance type.
[**InstanceTypeServicePut**](InstanceTypeServiceApi.md#InstanceTypeServicePut) | **Post** /proto.InstanceTypeService/Put | Create or update an instance type.
[**InstanceTypeServiceSearch**](InstanceTypeServiceApi.md#InstanceTypeServiceSearch) | **Get** /v1/instancetypes | List instance types.
[**InstanceTypeServiceSearchStream**](InstanceTypeServiceApi.md#InstanceTypeServiceSearchStream) | **Post** /proto.InstanceTypeService/SearchStream | List instance types as a stream.



## InstanceTypeServiceDelete

> map[string]interface{} InstanceTypeServiceDelete(ctx).Body(body).Execute()

Delete an instance typ.

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
    body := *openapiclient.NewProtoInstanceTypeDeleteRequest() // ProtoInstanceTypeDeleteRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceTypeServiceApi.InstanceTypeServiceDelete(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceTypeServiceApi.InstanceTypeServiceDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceTypeServiceDelete`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceTypeServiceApi.InstanceTypeServiceDelete`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInstanceTypeServiceDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ProtoInstanceTypeDeleteRequest**](ProtoInstanceTypeDeleteRequest.md) |  | 

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


## InstanceTypeServiceGet

> ProtoInstanceType InstanceTypeServiceGet(ctx, metadataName).Execute()

Get an instance type.

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
    metadataName := "metadataName_example" // string | Unique name of the instance type.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceTypeServiceApi.InstanceTypeServiceGet(context.Background(), metadataName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceTypeServiceApi.InstanceTypeServiceGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceTypeServiceGet`: ProtoInstanceType
    fmt.Fprintf(os.Stdout, "Response from `InstanceTypeServiceApi.InstanceTypeServiceGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataName** | **string** | Unique name of the instance type. | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceTypeServiceGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProtoInstanceType**](ProtoInstanceType.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstanceTypeServicePut

> map[string]interface{} InstanceTypeServicePut(ctx).Body(body).Execute()

Create or update an instance type.

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
    body := *openapiclient.NewProtoInstanceType() // ProtoInstanceType | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceTypeServiceApi.InstanceTypeServicePut(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceTypeServiceApi.InstanceTypeServicePut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceTypeServicePut`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceTypeServiceApi.InstanceTypeServicePut`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInstanceTypeServicePutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ProtoInstanceType**](ProtoInstanceType.md) |  | 

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


## InstanceTypeServiceSearch

> ProtoInstanceTypeSearchResponse InstanceTypeServiceSearch(ctx).Execute()

List instance types.

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
    resp, r, err := apiClient.InstanceTypeServiceApi.InstanceTypeServiceSearch(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceTypeServiceApi.InstanceTypeServiceSearch``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceTypeServiceSearch`: ProtoInstanceTypeSearchResponse
    fmt.Fprintf(os.Stdout, "Response from `InstanceTypeServiceApi.InstanceTypeServiceSearch`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceTypeServiceSearchRequest struct via the builder pattern


### Return type

[**ProtoInstanceTypeSearchResponse**](ProtoInstanceTypeSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstanceTypeServiceSearchStream

> StreamResultOfProtoInstanceType InstanceTypeServiceSearchStream(ctx).Body(body).Execute()

List instance types as a stream.

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
    resp, r, err := apiClient.InstanceTypeServiceApi.InstanceTypeServiceSearchStream(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceTypeServiceApi.InstanceTypeServiceSearchStream``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceTypeServiceSearchStream`: StreamResultOfProtoInstanceType
    fmt.Fprintf(os.Stdout, "Response from `InstanceTypeServiceApi.InstanceTypeServiceSearchStream`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInstanceTypeServiceSearchStreamRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | **map[string]interface{}** |  | 

### Return type

[**StreamResultOfProtoInstanceType**](StreamResultOfProtoInstanceType.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

