# \InstanceGroupServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**InstanceGroupServiceCreate**](InstanceGroupServiceApi.md#InstanceGroupServiceCreate) | **Post** /v1/cloudaccounts/{metadata.cloudAccountId}/instancegroups | Launch a new group of instances.
[**InstanceGroupServiceDelete**](InstanceGroupServiceApi.md#InstanceGroupServiceDelete) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/instancegroups/name/{metadata.name} | Request deletion (termination) of an instance group.
[**InstanceGroupServiceDeleteMember**](InstanceGroupServiceApi.md#InstanceGroupServiceDeleteMember) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/instancegroups/name/{metadata.name}/instance/id/{instanceResourceId} | Request deletion (termination) of an instance in a group. It always retains at least one instance in the group to use a template. To delete the entire group, use Delete API instead.
[**InstanceGroupServiceDeleteMember2**](InstanceGroupServiceApi.md#InstanceGroupServiceDeleteMember2) | **Delete** /v1/cloudaccounts/{metadata.cloudAccountId}/instancegroups/name/{metadata.name}/instance/name/{instanceName} | Request deletion (termination) of an instance in a group. It always retains at least one instance in the group to use a template. To delete the entire group, use Delete API instead.
[**InstanceGroupServicePing**](InstanceGroupServiceApi.md#InstanceGroupServicePing) | **Post** /proto.InstanceGroupService/Ping | Ping always returns a successful response by the service implementation. It can be used for testing connectivity to the service.
[**InstanceGroupServiceScaleUp**](InstanceGroupServiceApi.md#InstanceGroupServiceScaleUp) | **Patch** /v1/cloudaccounts/{metadata.cloudAccountId}/instancegroups/name/{metadata.name}/scale-up | Create new instances for the group to reach to the desired count. This returns an error if the desired count is less than the current count.
[**InstanceGroupServiceSearch**](InstanceGroupServiceApi.md#InstanceGroupServiceSearch) | **Get** /v1/cloudaccounts/{metadata.cloudAccountId}/instancegroups | List instance groups.
[**InstanceGroupServiceUpdate**](InstanceGroupServiceApi.md#InstanceGroupServiceUpdate) | **Put** /v1/cloudaccounts/{metadata.cloudAccountId}/instancegroups/name/{metadata.name} | Update the specification of an instanceGroup



## InstanceGroupServiceCreate

> ProtoInstanceGroup InstanceGroupServiceCreate(ctx, metadataCloudAccountId).Body(body).Execute()

Launch a new group of instances.

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
    body := *openapiclient.NewInstanceGroupServiceCreateRequest() // InstanceGroupServiceCreateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceGroupServiceApi.InstanceGroupServiceCreate(context.Background(), metadataCloudAccountId).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceGroupServiceApi.InstanceGroupServiceCreate``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceGroupServiceCreate`: ProtoInstanceGroup
    fmt.Fprintf(os.Stdout, "Response from `InstanceGroupServiceApi.InstanceGroupServiceCreate`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceGroupServiceCreateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**InstanceGroupServiceCreateRequest**](InstanceGroupServiceCreateRequest.md) |  | 

### Return type

[**ProtoInstanceGroup**](ProtoInstanceGroup.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstanceGroupServiceDelete

> map[string]interface{} InstanceGroupServiceDelete(ctx, metadataCloudAccountId, metadataName).Execute()

Request deletion (termination) of an instance group.

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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceGroupServiceApi.InstanceGroupServiceDelete(context.Background(), metadataCloudAccountId, metadataName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceGroupServiceApi.InstanceGroupServiceDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceGroupServiceDelete`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceGroupServiceApi.InstanceGroupServiceDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceGroupServiceDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



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


## InstanceGroupServiceDeleteMember

> map[string]interface{} InstanceGroupServiceDeleteMember(ctx, metadataCloudAccountId, metadataName, instanceResourceId).MetadataReserved2(metadataReserved2).InstanceName(instanceName).Execute()

Request deletion (termination) of an instance in a group. It always retains at least one instance in the group to use a template. To delete the entire group, use Delete API instead.

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
    instanceResourceId := "instanceResourceId_example" // string | 
    metadataReserved2 := "metadataReserved2_example" // string | Reserved. Added this field to overcome openAPi-same-struct issue. (optional)
    instanceName := "instanceName_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceGroupServiceApi.InstanceGroupServiceDeleteMember(context.Background(), metadataCloudAccountId, metadataName, instanceResourceId).MetadataReserved2(metadataReserved2).InstanceName(instanceName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceGroupServiceApi.InstanceGroupServiceDeleteMember``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceGroupServiceDeleteMember`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceGroupServiceApi.InstanceGroupServiceDeleteMember`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 
**instanceResourceId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceGroupServiceDeleteMemberRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **metadataReserved2** | **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | 
 **instanceName** | **string** |  | 

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


## InstanceGroupServiceDeleteMember2

> map[string]interface{} InstanceGroupServiceDeleteMember2(ctx, metadataCloudAccountId, metadataName, instanceName).MetadataReserved2(metadataReserved2).InstanceResourceId(instanceResourceId).Execute()

Request deletion (termination) of an instance in a group. It always retains at least one instance in the group to use a template. To delete the entire group, use Delete API instead.

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
    instanceName := "instanceName_example" // string | 
    metadataReserved2 := "metadataReserved2_example" // string | Reserved. Added this field to overcome openAPi-same-struct issue. (optional)
    instanceResourceId := "instanceResourceId_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceGroupServiceApi.InstanceGroupServiceDeleteMember2(context.Background(), metadataCloudAccountId, metadataName, instanceName).MetadataReserved2(metadataReserved2).InstanceResourceId(instanceResourceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceGroupServiceApi.InstanceGroupServiceDeleteMember2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceGroupServiceDeleteMember2`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceGroupServiceApi.InstanceGroupServiceDeleteMember2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 
**instanceName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceGroupServiceDeleteMember2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **metadataReserved2** | **string** | Reserved. Added this field to overcome openAPi-same-struct issue. | 
 **instanceResourceId** | **string** |  | 

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


## InstanceGroupServicePing

> map[string]interface{} InstanceGroupServicePing(ctx).Body(body).Execute()

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
    resp, r, err := apiClient.InstanceGroupServiceApi.InstanceGroupServicePing(context.Background()).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceGroupServiceApi.InstanceGroupServicePing``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceGroupServicePing`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceGroupServiceApi.InstanceGroupServicePing`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiInstanceGroupServicePingRequest struct via the builder pattern


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


## InstanceGroupServiceScaleUp

> ProtoInstanceGroupScaleResponse InstanceGroupServiceScaleUp(ctx, metadataCloudAccountId, metadataName).Body(body).Execute()

Create new instances for the group to reach to the desired count. This returns an error if the desired count is less than the current count.

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
    body := *openapiclient.NewInstanceGroupServiceScaleUpRequest() // InstanceGroupServiceScaleUpRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceGroupServiceApi.InstanceGroupServiceScaleUp(context.Background(), metadataCloudAccountId, metadataName).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceGroupServiceApi.InstanceGroupServiceScaleUp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceGroupServiceScaleUp`: ProtoInstanceGroupScaleResponse
    fmt.Fprintf(os.Stdout, "Response from `InstanceGroupServiceApi.InstanceGroupServiceScaleUp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceGroupServiceScaleUpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **body** | [**InstanceGroupServiceScaleUpRequest**](InstanceGroupServiceScaleUpRequest.md) |  | 

### Return type

[**ProtoInstanceGroupScaleResponse**](ProtoInstanceGroupScaleResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstanceGroupServiceSearch

> ProtoInstanceGroupSearchResponse InstanceGroupServiceSearch(ctx, metadataCloudAccountId).Execute()

List instance groups.

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
    resp, r, err := apiClient.InstanceGroupServiceApi.InstanceGroupServiceSearch(context.Background(), metadataCloudAccountId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceGroupServiceApi.InstanceGroupServiceSearch``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceGroupServiceSearch`: ProtoInstanceGroupSearchResponse
    fmt.Fprintf(os.Stdout, "Response from `InstanceGroupServiceApi.InstanceGroupServiceSearch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceGroupServiceSearchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProtoInstanceGroupSearchResponse**](ProtoInstanceGroupSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InstanceGroupServiceUpdate

> map[string]interface{} InstanceGroupServiceUpdate(ctx, metadataCloudAccountId, metadataName).Body(body).Execute()

Update the specification of an instanceGroup

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
    body := *openapiclient.NewInstanceGroupServiceUpdateRequest() // InstanceGroupServiceUpdateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InstanceGroupServiceApi.InstanceGroupServiceUpdate(context.Background(), metadataCloudAccountId, metadataName).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InstanceGroupServiceApi.InstanceGroupServiceUpdate``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `InstanceGroupServiceUpdate`: map[string]interface{}
    fmt.Fprintf(os.Stdout, "Response from `InstanceGroupServiceApi.InstanceGroupServiceUpdate`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**metadataCloudAccountId** | **string** |  | 
**metadataName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiInstanceGroupServiceUpdateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **body** | [**InstanceGroupServiceUpdateRequest**](InstanceGroupServiceUpdateRequest.md) |  | 

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

