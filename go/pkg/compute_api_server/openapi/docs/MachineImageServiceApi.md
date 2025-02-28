# \MachineImageServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**MachineImageServiceDelete**](MachineImageServiceApi.md#MachineImageServiceDelete) | **Post** /proto.MachineImageService/Delete | Delete a machine image.
[**MachineImageServiceGet**](MachineImageServiceApi.md#MachineImageServiceGet) | **Get** /v1/machineimages/{metadata.name} | Get a machine image.
[**MachineImageServicePut**](MachineImageServiceApi.md#MachineImageServicePut) | **Post** /proto.MachineImageService/Put | Create or update a machine image.
[**MachineImageServiceSearch**](MachineImageServiceApi.md#MachineImageServiceSearch) | **Get** /v1/machineimages | List machine images.
[**MachineImageServiceSearchStream**](MachineImageServiceApi.md#MachineImageServiceSearchStream) | **Post** /proto.MachineImageService/SearchStream | List machine images as a stream.



## MachineImageServiceDelete

> map[string]interface{} MachineImageServiceDelete(ctx).Body(body).Execute()

Delete a machine image.

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
    body := *openapiclient.NewProtoMachineImageDeleteRequest() // ProtoMachineImageDeleteRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.MachineImageServiceApi.MachineImageServiceDelete(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MachineImageServiceApi.MachineImageServiceDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `MachineImageServiceDelete`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `MachineImageServiceApi.MachineImageServiceDelete`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiMachineImageServiceDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ProtoMachineImageDeleteRequest**](ProtoMachineImageDeleteRequest.md) |  | 

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


## MachineImageServiceGet

> ProtoMachineImage MachineImageServiceGet(ctx, metadataName).Execute()

Get a machine image.

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
    metadataName := "metadataName_example" // string | Unique name of the machine image.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.MachineImageServiceApi.MachineImageServiceGet(context.Background(), metadataName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MachineImageServiceApi.MachineImageServiceGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `MachineImageServiceGet`: ProtoMachineImage
    fmt.Fprintf(os.Stdout, "Response from `MachineImageServiceApi.MachineImageServiceGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataName** | **string** | Unique name of the machine image. | 

### Other Parameters

Other parameters are passed through a pointer to a apiMachineImageServiceGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProtoMachineImage**](ProtoMachineImage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MachineImageServicePut

> map[string]interface{} MachineImageServicePut(ctx).Body(body).Execute()

Create or update a machine image.

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
    body := *openapiclient.NewProtoMachineImage() // ProtoMachineImage | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.MachineImageServiceApi.MachineImageServicePut(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MachineImageServiceApi.MachineImageServicePut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `MachineImageServicePut`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `MachineImageServiceApi.MachineImageServicePut`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiMachineImageServicePutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ProtoMachineImage**](ProtoMachineImage.md) |  | 

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


## MachineImageServiceSearch

> ProtoMachineImageSearchResponse MachineImageServiceSearch(ctx).MetadataInstanceType(metadataInstanceType).Execute()

List machine images.

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
    metadataInstanceType := "metadataInstanceType_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.MachineImageServiceApi.MachineImageServiceSearch(context.Background()).MetadataInstanceType(metadataInstanceType).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MachineImageServiceApi.MachineImageServiceSearch``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `MachineImageServiceSearch`: ProtoMachineImageSearchResponse
    fmt.Fprintf(os.Stdout, "Response from `MachineImageServiceApi.MachineImageServiceSearch`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiMachineImageServiceSearchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **metadataInstanceType** | **string** |  | 

### Return type

[**ProtoMachineImageSearchResponse**](ProtoMachineImageSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MachineImageServiceSearchStream

> StreamResultOfProtoMachineImage MachineImageServiceSearchStream(ctx).Body(body).Execute()

List machine images as a stream.

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
    body := *openapiclient.NewProtoMachineImageSearchRequest() // ProtoMachineImageSearchRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.MachineImageServiceApi.MachineImageServiceSearchStream(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `MachineImageServiceApi.MachineImageServiceSearchStream``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `MachineImageServiceSearchStream`: StreamResultOfProtoMachineImage
    fmt.Fprintf(os.Stdout, "Response from `MachineImageServiceApi.MachineImageServiceSearchStream`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiMachineImageServiceSearchStreamRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ProtoMachineImageSearchRequest**](ProtoMachineImageSearchRequest.md) |  | 

### Return type

[**StreamResultOfProtoMachineImage**](StreamResultOfProtoMachineImage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

