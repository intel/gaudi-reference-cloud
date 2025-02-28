# \PortsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DevcloudV2EnvironmentConfigurePortAccessPut**](PortsApi.md#DevcloudV2EnvironmentConfigurePortAccessPut) | **Put** /devcloud/v2/{environment}/configure/port/access | 
[**DevcloudV2EnvironmentConfigurePortTrunkPut**](PortsApi.md#DevcloudV2EnvironmentConfigurePortTrunkPut) | **Put** /devcloud/v2/{environment}/configure/port/trunk | 
[**DevcloudV2EnvironmentListPortsGet**](PortsApi.md#DevcloudV2EnvironmentListPortsGet) | **Get** /devcloud/v2/{environment}/list/ports | 
[**DevcloudV2EnvironmentPortDetailsGet**](PortsApi.md#DevcloudV2EnvironmentPortDetailsGet) | **Get** /devcloud/v2/{environment}/port/details | 



## DevcloudV2EnvironmentConfigurePortAccessPut

> DevcloudV2EnvironmentConfigurePortAccessPut200Response DevcloudV2EnvironmentConfigurePortAccessPut(ctx, environment).DevcloudV2EnvironmentConfigurePortAccessPutRequest(devcloudV2EnvironmentConfigurePortAccessPutRequest).Execute()



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
    environment := "environment_example" // string | 
    devcloudV2EnvironmentConfigurePortAccessPutRequest := *openapiclient.NewDevcloudV2EnvironmentConfigurePortAccessPutRequest() // DevcloudV2EnvironmentConfigurePortAccessPutRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PortsApi.DevcloudV2EnvironmentConfigurePortAccessPut(context.Background(), environment).DevcloudV2EnvironmentConfigurePortAccessPutRequest(devcloudV2EnvironmentConfigurePortAccessPutRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PortsApi.DevcloudV2EnvironmentConfigurePortAccessPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DevcloudV2EnvironmentConfigurePortAccessPut`: DevcloudV2EnvironmentConfigurePortAccessPut200Response
    fmt.Fprintf(os.Stdout, "Response from `PortsApi.DevcloudV2EnvironmentConfigurePortAccessPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**environment** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDevcloudV2EnvironmentConfigurePortAccessPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **devcloudV2EnvironmentConfigurePortAccessPutRequest** | [**DevcloudV2EnvironmentConfigurePortAccessPutRequest**](DevcloudV2EnvironmentConfigurePortAccessPutRequest.md) |  | 

### Return type

[**DevcloudV2EnvironmentConfigurePortAccessPut200Response**](DevcloudV2EnvironmentConfigurePortAccessPut200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DevcloudV2EnvironmentConfigurePortTrunkPut

> DevcloudV2EnvironmentConfigurePortAccessPut200Response DevcloudV2EnvironmentConfigurePortTrunkPut(ctx, environment).DevcloudV2EnvironmentConfigurePortTrunkPutRequest(devcloudV2EnvironmentConfigurePortTrunkPutRequest).Execute()



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
    environment := "environment_example" // string | 
    devcloudV2EnvironmentConfigurePortTrunkPutRequest := *openapiclient.NewDevcloudV2EnvironmentConfigurePortTrunkPutRequest() // DevcloudV2EnvironmentConfigurePortTrunkPutRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PortsApi.DevcloudV2EnvironmentConfigurePortTrunkPut(context.Background(), environment).DevcloudV2EnvironmentConfigurePortTrunkPutRequest(devcloudV2EnvironmentConfigurePortTrunkPutRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PortsApi.DevcloudV2EnvironmentConfigurePortTrunkPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DevcloudV2EnvironmentConfigurePortTrunkPut`: DevcloudV2EnvironmentConfigurePortAccessPut200Response
    fmt.Fprintf(os.Stdout, "Response from `PortsApi.DevcloudV2EnvironmentConfigurePortTrunkPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**environment** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDevcloudV2EnvironmentConfigurePortTrunkPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **devcloudV2EnvironmentConfigurePortTrunkPutRequest** | [**DevcloudV2EnvironmentConfigurePortTrunkPutRequest**](DevcloudV2EnvironmentConfigurePortTrunkPutRequest.md) |  | 

### Return type

[**DevcloudV2EnvironmentConfigurePortAccessPut200Response**](DevcloudV2EnvironmentConfigurePortAccessPut200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DevcloudV2EnvironmentListPortsGet

> DevcloudV2EnvironmentListPortsGet200Response DevcloudV2EnvironmentListPortsGet(ctx, environment).SwitchFqdn(switchFqdn).Execute()



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
    switchFqdn := "switchFqdn_example" // string | 
    environment := "environment_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PortsApi.DevcloudV2EnvironmentListPortsGet(context.Background(), environment).SwitchFqdn(switchFqdn).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PortsApi.DevcloudV2EnvironmentListPortsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DevcloudV2EnvironmentListPortsGet`: DevcloudV2EnvironmentListPortsGet200Response
    fmt.Fprintf(os.Stdout, "Response from `PortsApi.DevcloudV2EnvironmentListPortsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**environment** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDevcloudV2EnvironmentListPortsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **switchFqdn** | **string** |  | 


### Return type

[**DevcloudV2EnvironmentListPortsGet200Response**](DevcloudV2EnvironmentListPortsGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DevcloudV2EnvironmentPortDetailsGet

> DevcloudV2EnvironmentPortDetailsGet200Response DevcloudV2EnvironmentPortDetailsGet(ctx, environment).SwitchFqdn(switchFqdn).SwitchPort(switchPort).Execute()



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
    switchFqdn := "switchFqdn_example" // string | 
    switchPort := "switchPort_example" // string | 
    environment := "environment_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PortsApi.DevcloudV2EnvironmentPortDetailsGet(context.Background(), environment).SwitchFqdn(switchFqdn).SwitchPort(switchPort).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PortsApi.DevcloudV2EnvironmentPortDetailsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DevcloudV2EnvironmentPortDetailsGet`: DevcloudV2EnvironmentPortDetailsGet200Response
    fmt.Fprintf(os.Stdout, "Response from `PortsApi.DevcloudV2EnvironmentPortDetailsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**environment** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDevcloudV2EnvironmentPortDetailsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **switchFqdn** | **string** |  | 
 **switchPort** | **string** |  | 


### Return type

[**DevcloudV2EnvironmentPortDetailsGet200Response**](DevcloudV2EnvironmentPortDetailsGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

