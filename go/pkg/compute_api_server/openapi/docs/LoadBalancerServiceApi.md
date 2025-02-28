# \LoadBalancerServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**LoadBalancerServiceCreate**](LoadBalancerServiceApi.md#LoadBalancerServiceCreate) | **Post** /v1/cloudaccounts/{metadata.cloudAccountId}/loadbalancers | Create a new load balancer.
[**LoadBalancerServiceDelete**](LoadBalancerServiceApi.md#LoadBalancerServiceDelete) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/loadbalancers/id/{metadata.resourceId} | Request deletion of a load balancer.
[**LoadBalancerServiceDelete2**](LoadBalancerServiceApi.md#LoadBalancerServiceDelete2) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/loadbalancers/name/{metadata.name} | Request deletion of a load balancer.
[**LoadBalancerServiceGet**](LoadBalancerServiceApi.md#LoadBalancerServiceGet) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/loadbalancers/id/{metadata.resourceId} | Get the status of a load balancer.
[**LoadBalancerServiceGet2**](LoadBalancerServiceApi.md#LoadBalancerServiceGet2) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/loadbalancers/name/{metadata.name} | Get the status of a load balancer.
[**LoadBalancerServicePing**](LoadBalancerServiceApi.md#LoadBalancerServicePing) | **Post** /proto.LoadBalancerService/Ping | Ping always returns a successful response by the service implementation. It can be used for testing connectivity to the service.
[**LoadBalancerServiceSearch**](LoadBalancerServiceApi.md#LoadBalancerServiceSearch) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/loadbalancers | List load balancers.
[**LoadBalancerServiceSearch2**](LoadBalancerServiceApi.md#LoadBalancerServiceSearch2) | **Post** /v1/cloudaccounts/{metadata.cloudAccountId}/loadbalancers/search | List load balancers.
[**LoadBalancerServiceUpdate**](LoadBalancerServiceApi.md#LoadBalancerServiceUpdate) | **Put** /v1/cloudaccounts/{metadata.cloudAccountId}/loadbalancers/id/{metadata.resourceId} | Update the specification of an load balancer.
[**LoadBalancerServiceUpdate2**](LoadBalancerServiceApi.md#LoadBalancerServiceUpdate2) | **Put** /v1/cloudaccounts/{metadata.cloudAccountId}/loadbalancers/name/{metadata.name} | Update the specification of an load balancer.



## LoadBalancerServiceCreate

> ProtoLoadBalancer LoadBalancerServiceCreate(ctx, metadataCloudAccountId).Body(body).Execute()

Create a new load balancer.

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
    body := *openapiclient.NewLoadBalancerServiceCreateRequest() // LoadBalancerServiceCreateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LoadBalancerServiceApi.LoadBalancerServiceCreate(context.Background(), metadataCloudAccountId).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoadBalancerServiceApi.LoadBalancerServiceCreate``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LoadBalancerServiceCreate`: ProtoLoadBalancer
    fmt.Fprintf(os.Stdout, "Response from `LoadBalancerServiceApi.LoadBalancerServiceCreate`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiLoadBalancerServiceCreateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**LoadBalancerServiceCreateRequest**](LoadBalancerServiceCreateRequest.md) |  | 

### Return type

[**ProtoLoadBalancer**](ProtoLoadBalancer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LoadBalancerServiceDelete

> map[string]interface{} LoadBalancerServiceDelete(ctx, metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()

Request deletion of a load balancer.

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
    metadataResourceVersion := "metadataResourceVersion_example" // string | If provided, the existing record must have this resourceVersion for the request to succeed. (optional)
    metadataReserved1 := "metadataReserved1_example" // string | Reserved. Added this field to overcome openAPi-same-struct issue. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LoadBalancerServiceApi.LoadBalancerServiceDelete(context.Background(), metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoadBalancerServiceApi.LoadBalancerServiceDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LoadBalancerServiceDelete`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `LoadBalancerServiceApi.LoadBalancerServiceDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiLoadBalancerServiceDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataName** | **string** |  | 
 **metadataResourceVersion** | **string** | If provided, the existing record must have this resourceVersion for the request to succeed. | 
 **metadataReserved1** | **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | 

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


## LoadBalancerServiceDelete2

> map[string]interface{} LoadBalancerServiceDelete2(ctx, metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()

Request deletion of a load balancer.

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
    metadataResourceVersion := "metadataResourceVersion_example" // string | If provided, the existing record must have this resourceVersion for the request to succeed. (optional)
    metadataReserved1 := "metadataReserved1_example" // string | Reserved. Added this field to overcome openAPi-same-struct issue. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LoadBalancerServiceApi.LoadBalancerServiceDelete2(context.Background(), metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoadBalancerServiceApi.LoadBalancerServiceDelete2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LoadBalancerServiceDelete2`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `LoadBalancerServiceApi.LoadBalancerServiceDelete2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiLoadBalancerServiceDelete2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataResourceId** | **string** |  | 
 **metadataResourceVersion** | **string** | If provided, the existing record must have this resourceVersion for the request to succeed. | 
 **metadataReserved1** | **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | 

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


## LoadBalancerServiceGet

> ProtoLoadBalancer LoadBalancerServiceGet(ctx, metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()

Get the status of a load balancer.

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
    metadataResourceVersion := "metadataResourceVersion_example" // string | If provided, the existing record must have this resourceVersion for the request to succeed. (optional)
    metadataReserved1 := "metadataReserved1_example" // string | Reserved. Added this field to overcome openAPi-same-struct issue. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LoadBalancerServiceApi.LoadBalancerServiceGet(context.Background(), metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoadBalancerServiceApi.LoadBalancerServiceGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LoadBalancerServiceGet`: ProtoLoadBalancer
    fmt.Fprintf(os.Stdout, "Response from `LoadBalancerServiceApi.LoadBalancerServiceGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiLoadBalancerServiceGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataName** | **string** |  | 
 **metadataResourceVersion** | **string** | If provided, the existing record must have this resourceVersion for the request to succeed. | 
 **metadataReserved1** | **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | 

### Return type

[**ProtoLoadBalancer**](ProtoLoadBalancer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LoadBalancerServiceGet2

> ProtoLoadBalancer LoadBalancerServiceGet2(ctx, metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()

Get the status of a load balancer.

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
    metadataResourceVersion := "metadataResourceVersion_example" // string | If provided, the existing record must have this resourceVersion for the request to succeed. (optional)
    metadataReserved1 := "metadataReserved1_example" // string | Reserved. Added this field to overcome openAPi-same-struct issue. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LoadBalancerServiceApi.LoadBalancerServiceGet2(context.Background(), metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoadBalancerServiceApi.LoadBalancerServiceGet2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LoadBalancerServiceGet2`: ProtoLoadBalancer
    fmt.Fprintf(os.Stdout, "Response from `LoadBalancerServiceApi.LoadBalancerServiceGet2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiLoadBalancerServiceGet2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataResourceId** | **string** |  | 
 **metadataResourceVersion** | **string** | If provided, the existing record must have this resourceVersion for the request to succeed. | 
 **metadataReserved1** | **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | 

### Return type

[**ProtoLoadBalancer**](ProtoLoadBalancer.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LoadBalancerServicePing

> map[string]interface{} LoadBalancerServicePing(ctx).Body(body).Execute()

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
    resp, r, err := apiClient.LoadBalancerServiceApi.LoadBalancerServicePing(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoadBalancerServiceApi.LoadBalancerServicePing``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LoadBalancerServicePing`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `LoadBalancerServiceApi.LoadBalancerServicePing`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiLoadBalancerServicePingRequest struct via the builder pattern


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


## LoadBalancerServiceSearch

> ProtoLoadBalancerSearchResponse LoadBalancerServiceSearch(ctx, metadataCloudAccountId).MetadataReserved1(metadataReserved1).Execute()

List load balancers.

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
    metadataReserved1 := "metadataReserved1_example" // string | Reserved. Added this field to overcome openAPi-same-struct issue. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LoadBalancerServiceApi.LoadBalancerServiceSearch(context.Background(), metadataCloudAccountId).MetadataReserved1(metadataReserved1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoadBalancerServiceApi.LoadBalancerServiceSearch``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LoadBalancerServiceSearch`: ProtoLoadBalancerSearchResponse
    fmt.Fprintf(os.Stdout, "Response from `LoadBalancerServiceApi.LoadBalancerServiceSearch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiLoadBalancerServiceSearchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **metadataReserved1** | **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | 

### Return type

[**ProtoLoadBalancerSearchResponse**](ProtoLoadBalancerSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LoadBalancerServiceSearch2

> ProtoLoadBalancerSearchResponse LoadBalancerServiceSearch2(ctx, metadataCloudAccountId).Body(body).Execute()

List load balancers.

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
    body := *openapiclient.NewLoadBalancerServiceSearch2Request() // LoadBalancerServiceSearch2Request | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LoadBalancerServiceApi.LoadBalancerServiceSearch2(context.Background(), metadataCloudAccountId).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoadBalancerServiceApi.LoadBalancerServiceSearch2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LoadBalancerServiceSearch2`: ProtoLoadBalancerSearchResponse
    fmt.Fprintf(os.Stdout, "Response from `LoadBalancerServiceApi.LoadBalancerServiceSearch2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiLoadBalancerServiceSearch2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**LoadBalancerServiceSearch2Request**](LoadBalancerServiceSearch2Request.md) |  | 

### Return type

[**ProtoLoadBalancerSearchResponse**](ProtoLoadBalancerSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LoadBalancerServiceUpdate

> map[string]interface{} LoadBalancerServiceUpdate(ctx, metadataCloudAccountId, metadataResourceId).Body(body).Execute()

Update the specification of an load balancer.

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
    body := *openapiclient.NewLoadBalancerServiceUpdateRequest() // LoadBalancerServiceUpdateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LoadBalancerServiceApi.LoadBalancerServiceUpdate(context.Background(), metadataCloudAccountId, metadataResourceId).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoadBalancerServiceApi.LoadBalancerServiceUpdate``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LoadBalancerServiceUpdate`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `LoadBalancerServiceApi.LoadBalancerServiceUpdate`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiLoadBalancerServiceUpdateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **body** | [**LoadBalancerServiceUpdateRequest**](LoadBalancerServiceUpdateRequest.md) |  | 

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


## LoadBalancerServiceUpdate2

> map[string]interface{} LoadBalancerServiceUpdate2(ctx, metadataCloudAccountId, metadataName).Body(body).Execute()

Update the specification of an load balancer.

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
    body := *openapiclient.NewLoadBalancerServiceUpdate2Request() // LoadBalancerServiceUpdate2Request | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.LoadBalancerServiceApi.LoadBalancerServiceUpdate2(context.Background(), metadataCloudAccountId, metadataName).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `LoadBalancerServiceApi.LoadBalancerServiceUpdate2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LoadBalancerServiceUpdate2`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `LoadBalancerServiceApi.LoadBalancerServiceUpdate2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiLoadBalancerServiceUpdate2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **body** | [**LoadBalancerServiceUpdate2Request**](LoadBalancerServiceUpdate2Request.md) |  | 

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

