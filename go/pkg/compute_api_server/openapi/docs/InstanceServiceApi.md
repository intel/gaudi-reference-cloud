# \InstanceServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**InstanceServiceCreate**](InstanceServiceApi.md#InstanceServiceCreate) | **Post** /v1/cloudaccounts/{metadata.cloudAccountId}/instances | Launch a new baremetal or virtual machine instance.
[**InstanceServiceDelete**](InstanceServiceApi.md#InstanceServiceDelete) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/instances/id/{metadata.resourceId} | Request deletion (termination) of an instance.
[**InstanceServiceDelete2**](InstanceServiceApi.md#InstanceServiceDelete2) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/instances/name/{metadata.name} | Request deletion (termination) of an instance.
[**InstanceServiceGet**](InstanceServiceApi.md#InstanceServiceGet) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/instances/id/{metadata.resourceId} | Get the status of an instance.
[**InstanceServiceGet2**](InstanceServiceApi.md#InstanceServiceGet2) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/instances/name/{metadata.name} | Get the status of an instance.
[**InstanceServicePing**](InstanceServiceApi.md#InstanceServicePing) | **Post** /proto.InstanceService/Ping | Ping always returns a successful response by the service implementation. It can be used for testing connectivity to the service.
[**InstanceServiceSearch**](InstanceServiceApi.md#InstanceServiceSearch) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/instances | List instances.
[**InstanceServiceSearch2**](InstanceServiceApi.md#InstanceServiceSearch2) | **Post** /v1/cloudaccounts/{metadata.cloudAccountId}/instances/search | List instances.
[**InstanceServiceUpdate**](InstanceServiceApi.md#InstanceServiceUpdate) | **Put** /v1/cloudaccounts/{metadata.cloudAccountId}/instances/id/{metadata.resourceId} | Update the specification of an instance.
[**InstanceServiceUpdate2**](InstanceServiceApi.md#InstanceServiceUpdate2) | **Put** /v1/cloudaccounts/{metadata.cloudAccountId}/instances/name/{metadata.name} | Update the specification of an instance.



## InstanceServiceCreate

> ProtoInstance InstanceServiceCreate(ctx, metadataCloudAccountId).Body(body).Execute()

Launch a new baremetal or virtual machine instance.

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
    body := *openapiclient.NewInstanceServiceCreateRequest() // InstanceServiceCreateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceServiceApi.InstanceServiceCreate(context.Background(), metadataCloudAccountId).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceServiceApi.InstanceServiceCreate``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceServiceCreate`: ProtoInstance
    fmt.Fprintf(os.Stdout, "Response from `InstanceServiceApi.InstanceServiceCreate`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceServiceCreateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**InstanceServiceCreateRequest**](InstanceServiceCreateRequest.md) |  | 

### Return type

[**ProtoInstance**](ProtoInstance.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstanceServiceDelete

> map[string]interface{} InstanceServiceDelete(ctx, metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()

Request deletion (termination) of an instance.

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
    resp, r, err := apiClient.InstanceServiceApi.InstanceServiceDelete(context.Background(), metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceServiceApi.InstanceServiceDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceServiceDelete`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceServiceApi.InstanceServiceDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceServiceDeleteRequest struct via the builder pattern


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


## InstanceServiceDelete2

> map[string]interface{} InstanceServiceDelete2(ctx, metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()

Request deletion (termination) of an instance.

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
    resp, r, err := apiClient.InstanceServiceApi.InstanceServiceDelete2(context.Background(), metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceServiceApi.InstanceServiceDelete2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceServiceDelete2`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceServiceApi.InstanceServiceDelete2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceServiceDelete2Request struct via the builder pattern


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


## InstanceServiceGet

> ProtoInstance InstanceServiceGet(ctx, metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()

Get the status of an instance.

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
    resp, r, err := apiClient.InstanceServiceApi.InstanceServiceGet(context.Background(), metadataCloudAccountId, metadataResourceId).MetadataName(metadataName).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceServiceApi.InstanceServiceGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceServiceGet`: ProtoInstance
    fmt.Fprintf(os.Stdout, "Response from `InstanceServiceApi.InstanceServiceGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceServiceGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataName** | **string** |  | 
 **metadataResourceVersion** | **string** | If provided, the existing record must have this resourceVersion for the request to succeed. | 
 **metadataReserved1** | **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | 

### Return type

[**ProtoInstance**](ProtoInstance.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstanceServiceGet2

> ProtoInstance InstanceServiceGet2(ctx, metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()

Get the status of an instance.

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
    resp, r, err := apiClient.InstanceServiceApi.InstanceServiceGet2(context.Background(), metadataCloudAccountId, metadataName).MetadataResourceId(metadataResourceId).MetadataResourceVersion(metadataResourceVersion).MetadataReserved1(metadataReserved1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceServiceApi.InstanceServiceGet2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceServiceGet2`: ProtoInstance
    fmt.Fprintf(os.Stdout, "Response from `InstanceServiceApi.InstanceServiceGet2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceServiceGet2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **metadataResourceId** | **string** |  | 
 **metadataResourceVersion** | **string** | If provided, the existing record must have this resourceVersion for the request to succeed. | 
 **metadataReserved1** | **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | 

### Return type

[**ProtoInstance**](ProtoInstance.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstanceServicePing

> map[string]interface{} InstanceServicePing(ctx).Body(body).Execute()

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
    resp, r, err := apiClient.InstanceServiceApi.InstanceServicePing(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceServiceApi.InstanceServicePing``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceServicePing`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceServiceApi.InstanceServicePing`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInstanceServicePingRequest struct via the builder pattern


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


## InstanceServiceSearch

> ProtoInstanceSearchResponse InstanceServiceSearch(ctx, metadataCloudAccountId).MetadataReserved1(metadataReserved1).MetadataInstanceGroup(metadataInstanceGroup).MetadataInstanceGroupFilter(metadataInstanceGroupFilter).Execute()

List instances.

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
    metadataInstanceGroup := "metadataInstanceGroup_example" // string | If instanceGroupFilter is ExactValue, return instances in this instance group. Otherwise, this field is ignored (optional)
    metadataInstanceGroupFilter := "metadataInstanceGroupFilter_example" // string | Filter instances by instance group. If Default, this behaves like Empty and returns instances that are not in any instance group.   - Default: Use the default behavior, which is described in the specific SearchFilterCriteria field.  - Any: Return records with any value in this field (including empty).  - Empty: Return records with an empty value in this field  - NonEmpty: Return records with a non-empty value in this field  - ExactValue: Return records with an exact value in this field (optional) (default to "Default")

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceServiceApi.InstanceServiceSearch(context.Background(), metadataCloudAccountId).MetadataReserved1(metadataReserved1).MetadataInstanceGroup(metadataInstanceGroup).MetadataInstanceGroupFilter(metadataInstanceGroupFilter).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceServiceApi.InstanceServiceSearch``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceServiceSearch`: ProtoInstanceSearchResponse
    fmt.Fprintf(os.Stdout, "Response from `InstanceServiceApi.InstanceServiceSearch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceServiceSearchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **metadataReserved1** | **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | 
 **metadataInstanceGroup** | **string** | If instanceGroupFilter is ExactValue, return instances in this instance group. Otherwise, this field is ignored | 
 **metadataInstanceGroupFilter** | **string** | Filter instances by instance group. If Default, this behaves like Empty and returns instances that are not in any instance group.   - Default: Use the default behavior, which is described in the specific SearchFilterCriteria field.  - Any: Return records with any value in this field (including empty).  - Empty: Return records with an empty value in this field  - NonEmpty: Return records with a non-empty value in this field  - ExactValue: Return records with an exact value in this field | [default to &quot;Default&quot;]

### Return type

[**ProtoInstanceSearchResponse**](ProtoInstanceSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstanceServiceSearch2

> ProtoInstanceSearchResponse InstanceServiceSearch2(ctx, metadataCloudAccountId).Body(body).Execute()

List instances.

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
    body := *openapiclient.NewInstanceServiceSearch2Request() // InstanceServiceSearch2Request | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceServiceApi.InstanceServiceSearch2(context.Background(), metadataCloudAccountId).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceServiceApi.InstanceServiceSearch2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceServiceSearch2`: ProtoInstanceSearchResponse
    fmt.Fprintf(os.Stdout, "Response from `InstanceServiceApi.InstanceServiceSearch2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceServiceSearch2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**InstanceServiceSearch2Request**](InstanceServiceSearch2Request.md) |  | 

### Return type

[**ProtoInstanceSearchResponse**](ProtoInstanceSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstanceServiceUpdate

> map[string]interface{} InstanceServiceUpdate(ctx, metadataCloudAccountId, metadataResourceId).Body(body).Execute()

Update the specification of an instance.

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
    body := *openapiclient.NewInstanceServiceUpdateRequest() // InstanceServiceUpdateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceServiceApi.InstanceServiceUpdate(context.Background(), metadataCloudAccountId, metadataResourceId).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceServiceApi.InstanceServiceUpdate``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceServiceUpdate`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceServiceApi.InstanceServiceUpdate`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceServiceUpdateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **body** | [**InstanceServiceUpdateRequest**](InstanceServiceUpdateRequest.md) |  | 

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


## InstanceServiceUpdate2

> map[string]interface{} InstanceServiceUpdate2(ctx, metadataCloudAccountId, metadataName).Body(body).Execute()

Update the specification of an instance.

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
    body := *openapiclient.NewInstanceServiceUpdate2Request() // InstanceServiceUpdate2Request | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceServiceApi.InstanceServiceUpdate2(context.Background(), metadataCloudAccountId, metadataName).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceServiceApi.InstanceServiceUpdate2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceServiceUpdate2`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceServiceApi.InstanceServiceUpdate2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceServiceUpdate2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **body** | [**InstanceServiceUpdate2Request**](InstanceServiceUpdate2Request.md) |  | 

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

