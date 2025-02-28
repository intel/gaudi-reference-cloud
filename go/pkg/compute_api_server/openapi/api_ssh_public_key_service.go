/*
compute.proto

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: version not set
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// SshPublicKeyServiceApiService SshPublicKeyServiceApi service
type SshPublicKeyServiceApiService service

type ApiSshPublicKeyServiceCreateRequest struct {
	ctx                    context.Context
	ApiService             *SshPublicKeyServiceApiService
	metadataCloudAccountId string
	body                   *SshPublicKeyServiceCreateRequest
}

func (r ApiSshPublicKeyServiceCreateRequest) Body(body SshPublicKeyServiceCreateRequest) ApiSshPublicKeyServiceCreateRequest {
	r.body = &body
	return r
}

func (r ApiSshPublicKeyServiceCreateRequest) Execute() (*ProtoSshPublicKey, *http.Response, error) {
	return r.ApiService.SshPublicKeyServiceCreateExecute(r)
}

/*
SshPublicKeyServiceCreate Store an SSH public key.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param metadataCloudAccountId
	@return ApiSshPublicKeyServiceCreateRequest
*/
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceCreate(ctx context.Context, metadataCloudAccountId string) ApiSshPublicKeyServiceCreateRequest {
	return ApiSshPublicKeyServiceCreateRequest{
		ApiService:             a,
		ctx:                    ctx,
		metadataCloudAccountId: metadataCloudAccountId,
	}
}

// Execute executes the request
//
//	@return ProtoSshPublicKey
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceCreateExecute(r ApiSshPublicKeyServiceCreateRequest) (*ProtoSshPublicKey, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodPost
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *ProtoSshPublicKey
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SshPublicKeyServiceApiService.SshPublicKeyServiceCreate")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys"
	localVarPath = strings.Replace(localVarPath, "{"+"metadata.cloudAccountId"+"}", url.PathEscape(parameterToString(r.metadataCloudAccountId, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.body == nil {
		return localVarReturnValue, nil, reportError("body is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.body
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	closeErr := localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}
	if closeErr != nil {
		return localVarReturnValue, localVarHTTPResponse, closeErr
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RpcStatus
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiSshPublicKeyServiceDeleteRequest struct {
	ctx                    context.Context
	ApiService             *SshPublicKeyServiceApiService
	metadataCloudAccountId string
	metadataResourceId     string
	metadataName           *string
}

func (r ApiSshPublicKeyServiceDeleteRequest) MetadataName(metadataName string) ApiSshPublicKeyServiceDeleteRequest {
	r.metadataName = &metadataName
	return r
}

func (r ApiSshPublicKeyServiceDeleteRequest) Execute() (map[string]interface{}, *http.Response, error) {
	return r.ApiService.SshPublicKeyServiceDeleteExecute(r)
}

/*
SshPublicKeyServiceDelete Delete an SSH public key.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param metadataCloudAccountId
	@param metadataResourceId
	@return ApiSshPublicKeyServiceDeleteRequest
*/
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceDelete(ctx context.Context, metadataCloudAccountId string, metadataResourceId string) ApiSshPublicKeyServiceDeleteRequest {
	return ApiSshPublicKeyServiceDeleteRequest{
		ApiService:             a,
		ctx:                    ctx,
		metadataCloudAccountId: metadataCloudAccountId,
		metadataResourceId:     metadataResourceId,
	}
}

// Execute executes the request
//
//	@return map[string]interface{}
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceDeleteExecute(r ApiSshPublicKeyServiceDeleteRequest) (map[string]interface{}, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodDelete
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue map[string]interface{}
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SshPublicKeyServiceApiService.SshPublicKeyServiceDelete")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys/id/{metadata.resourceId}"
	localVarPath = strings.Replace(localVarPath, "{"+"metadata.cloudAccountId"+"}", url.PathEscape(parameterToString(r.metadataCloudAccountId, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"metadata.resourceId"+"}", url.PathEscape(parameterToString(r.metadataResourceId, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.metadataName != nil {
		localVarQueryParams.Add("metadata.name", parameterToString(*r.metadataName, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	closeErr := localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}
	if closeErr != nil {
		return localVarReturnValue, localVarHTTPResponse, closeErr
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RpcStatus
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiSshPublicKeyServiceDelete2Request struct {
	ctx                    context.Context
	ApiService             *SshPublicKeyServiceApiService
	metadataCloudAccountId string
	metadataName           string
	metadataResourceId     *string
}

func (r ApiSshPublicKeyServiceDelete2Request) MetadataResourceId(metadataResourceId string) ApiSshPublicKeyServiceDelete2Request {
	r.metadataResourceId = &metadataResourceId
	return r
}

func (r ApiSshPublicKeyServiceDelete2Request) Execute() (map[string]interface{}, *http.Response, error) {
	return r.ApiService.SshPublicKeyServiceDelete2Execute(r)
}

/*
SshPublicKeyServiceDelete2 Delete an SSH public key.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param metadataCloudAccountId
	@param metadataName
	@return ApiSshPublicKeyServiceDelete2Request
*/
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceDelete2(ctx context.Context, metadataCloudAccountId string, metadataName string) ApiSshPublicKeyServiceDelete2Request {
	return ApiSshPublicKeyServiceDelete2Request{
		ApiService:             a,
		ctx:                    ctx,
		metadataCloudAccountId: metadataCloudAccountId,
		metadataName:           metadataName,
	}
}

// Execute executes the request
//
//	@return map[string]interface{}
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceDelete2Execute(r ApiSshPublicKeyServiceDelete2Request) (map[string]interface{}, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodDelete
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue map[string]interface{}
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SshPublicKeyServiceApiService.SshPublicKeyServiceDelete2")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys/name/{metadata.name}"
	localVarPath = strings.Replace(localVarPath, "{"+"metadata.cloudAccountId"+"}", url.PathEscape(parameterToString(r.metadataCloudAccountId, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"metadata.name"+"}", url.PathEscape(parameterToString(r.metadataName, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.metadataResourceId != nil {
		localVarQueryParams.Add("metadata.resourceId", parameterToString(*r.metadataResourceId, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	closeErr := localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}
	if closeErr != nil {
		return localVarReturnValue, localVarHTTPResponse, closeErr
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RpcStatus
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiSshPublicKeyServiceGetRequest struct {
	ctx                    context.Context
	ApiService             *SshPublicKeyServiceApiService
	metadataCloudAccountId string
	metadataResourceId     string
	metadataName           *string
}

func (r ApiSshPublicKeyServiceGetRequest) MetadataName(metadataName string) ApiSshPublicKeyServiceGetRequest {
	r.metadataName = &metadataName
	return r
}

func (r ApiSshPublicKeyServiceGetRequest) Execute() (*ProtoSshPublicKey, *http.Response, error) {
	return r.ApiService.SshPublicKeyServiceGetExecute(r)
}

/*
SshPublicKeyServiceGet Retrieve a stored SSH public key.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param metadataCloudAccountId
	@param metadataResourceId
	@return ApiSshPublicKeyServiceGetRequest
*/
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceGet(ctx context.Context, metadataCloudAccountId string, metadataResourceId string) ApiSshPublicKeyServiceGetRequest {
	return ApiSshPublicKeyServiceGetRequest{
		ApiService:             a,
		ctx:                    ctx,
		metadataCloudAccountId: metadataCloudAccountId,
		metadataResourceId:     metadataResourceId,
	}
}

// Execute executes the request
//
//	@return ProtoSshPublicKey
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceGetExecute(r ApiSshPublicKeyServiceGetRequest) (*ProtoSshPublicKey, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *ProtoSshPublicKey
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SshPublicKeyServiceApiService.SshPublicKeyServiceGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys/id/{metadata.resourceId}"
	localVarPath = strings.Replace(localVarPath, "{"+"metadata.cloudAccountId"+"}", url.PathEscape(parameterToString(r.metadataCloudAccountId, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"metadata.resourceId"+"}", url.PathEscape(parameterToString(r.metadataResourceId, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.metadataName != nil {
		localVarQueryParams.Add("metadata.name", parameterToString(*r.metadataName, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	closeErr := localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}
	if closeErr != nil {
		return localVarReturnValue, localVarHTTPResponse, closeErr
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RpcStatus
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiSshPublicKeyServiceGet2Request struct {
	ctx                    context.Context
	ApiService             *SshPublicKeyServiceApiService
	metadataCloudAccountId string
	metadataName           string
	metadataResourceId     *string
}

func (r ApiSshPublicKeyServiceGet2Request) MetadataResourceId(metadataResourceId string) ApiSshPublicKeyServiceGet2Request {
	r.metadataResourceId = &metadataResourceId
	return r
}

func (r ApiSshPublicKeyServiceGet2Request) Execute() (*ProtoSshPublicKey, *http.Response, error) {
	return r.ApiService.SshPublicKeyServiceGet2Execute(r)
}

/*
SshPublicKeyServiceGet2 Retrieve a stored SSH public key.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param metadataCloudAccountId
	@param metadataName
	@return ApiSshPublicKeyServiceGet2Request
*/
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceGet2(ctx context.Context, metadataCloudAccountId string, metadataName string) ApiSshPublicKeyServiceGet2Request {
	return ApiSshPublicKeyServiceGet2Request{
		ApiService:             a,
		ctx:                    ctx,
		metadataCloudAccountId: metadataCloudAccountId,
		metadataName:           metadataName,
	}
}

// Execute executes the request
//
//	@return ProtoSshPublicKey
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceGet2Execute(r ApiSshPublicKeyServiceGet2Request) (*ProtoSshPublicKey, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *ProtoSshPublicKey
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SshPublicKeyServiceApiService.SshPublicKeyServiceGet2")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys/name/{metadata.name}"
	localVarPath = strings.Replace(localVarPath, "{"+"metadata.cloudAccountId"+"}", url.PathEscape(parameterToString(r.metadataCloudAccountId, "")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"metadata.name"+"}", url.PathEscape(parameterToString(r.metadataName, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.metadataResourceId != nil {
		localVarQueryParams.Add("metadata.resourceId", parameterToString(*r.metadataResourceId, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	closeErr := localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}
	if closeErr != nil {
		return localVarReturnValue, localVarHTTPResponse, closeErr
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RpcStatus
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiSshPublicKeyServicePingRequest struct {
	ctx        context.Context
	ApiService *SshPublicKeyServiceApiService
}

func (r ApiSshPublicKeyServicePingRequest) Execute() (map[string]interface{}, *http.Response, error) {
	return r.ApiService.SshPublicKeyServicePingExecute(r)
}

/*
SshPublicKeyServicePing Ping always returns a successful response by the service implementation. It can be used for testing connectivity to the service.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@return ApiSshPublicKeyServicePingRequest
*/
func (a *SshPublicKeyServiceApiService) SshPublicKeyServicePing(ctx context.Context) ApiSshPublicKeyServicePingRequest {
	return ApiSshPublicKeyServicePingRequest{
		ApiService: a,
		ctx:        ctx,
	}
}

// Execute executes the request
//
//	@return map[string]interface{}
func (a *SshPublicKeyServiceApiService) SshPublicKeyServicePingExecute(r ApiSshPublicKeyServicePingRequest) (map[string]interface{}, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue map[string]interface{}
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SshPublicKeyServiceApiService.SshPublicKeyServicePing")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/ping"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	closeErr := localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}
	if closeErr != nil {
		return localVarReturnValue, localVarHTTPResponse, closeErr
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RpcStatus
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiSshPublicKeyServiceSearchRequest struct {
	ctx                    context.Context
	ApiService             *SshPublicKeyServiceApiService
	metadataCloudAccountId string
}

func (r ApiSshPublicKeyServiceSearchRequest) Execute() (*ProtoSshPublicKeySearchResponse, *http.Response, error) {
	return r.ApiService.SshPublicKeyServiceSearchExecute(r)
}

/*
SshPublicKeyServiceSearch Get a list of stored SSH public keys.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@param metadataCloudAccountId
	@return ApiSshPublicKeyServiceSearchRequest
*/
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceSearch(ctx context.Context, metadataCloudAccountId string) ApiSshPublicKeyServiceSearchRequest {
	return ApiSshPublicKeyServiceSearchRequest{
		ApiService:             a,
		ctx:                    ctx,
		metadataCloudAccountId: metadataCloudAccountId,
	}
}

// Execute executes the request
//
//	@return ProtoSshPublicKeySearchResponse
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceSearchExecute(r ApiSshPublicKeyServiceSearchRequest) (*ProtoSshPublicKeySearchResponse, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodGet
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *ProtoSshPublicKeySearchResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SshPublicKeyServiceApiService.SshPublicKeyServiceSearch")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/v1/cloudaccounts/{metadata.cloudAccountId}/sshpublickeys"
	localVarPath = strings.Replace(localVarPath, "{"+"metadata.cloudAccountId"+"}", url.PathEscape(parameterToString(r.metadataCloudAccountId, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	closeErr := localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}
	if closeErr != nil {
		return localVarReturnValue, localVarHTTPResponse, closeErr
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RpcStatus
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiSshPublicKeyServiceSearchStreamRequest struct {
	ctx        context.Context
	ApiService *SshPublicKeyServiceApiService
	body       *ProtoSshPublicKeySearchRequest
}

func (r ApiSshPublicKeyServiceSearchStreamRequest) Body(body ProtoSshPublicKeySearchRequest) ApiSshPublicKeyServiceSearchStreamRequest {
	r.body = &body
	return r
}

func (r ApiSshPublicKeyServiceSearchStreamRequest) Execute() (*StreamResultOfProtoSshPublicKey, *http.Response, error) {
	return r.ApiService.SshPublicKeyServiceSearchStreamExecute(r)
}

/*
SshPublicKeyServiceSearchStream List stored SSH public keys as a stream. Warning: This does not work with OpenAPI client. Internal-use only.

	@param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
	@return ApiSshPublicKeyServiceSearchStreamRequest
*/
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceSearchStream(ctx context.Context) ApiSshPublicKeyServiceSearchStreamRequest {
	return ApiSshPublicKeyServiceSearchStreamRequest{
		ApiService: a,
		ctx:        ctx,
	}
}

// Execute executes the request
//
//	@return StreamResultOfProtoSshPublicKey
func (a *SshPublicKeyServiceApiService) SshPublicKeyServiceSearchStreamExecute(r ApiSshPublicKeyServiceSearchStreamRequest) (*StreamResultOfProtoSshPublicKey, *http.Response, error) {
	var (
		localVarHTTPMethod  = http.MethodPost
		localVarPostBody    interface{}
		formFiles           []formFile
		localVarReturnValue *StreamResultOfProtoSshPublicKey
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SshPublicKeyServiceApiService.SshPublicKeyServiceSearchStream")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/proto.SshPublicKeyService/SearchStream"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.body == nil {
		return localVarReturnValue, nil, reportError("body is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.body
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	closeErr := localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}
	if closeErr != nil {
		return localVarReturnValue, localVarHTTPResponse, closeErr
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		var v RpcStatus
		err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
		if err != nil {
			newErr.error = err.Error()
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		newErr.error = formatErrorMessage(localVarHTTPResponse.Status, &v)
		newErr.model = v
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
