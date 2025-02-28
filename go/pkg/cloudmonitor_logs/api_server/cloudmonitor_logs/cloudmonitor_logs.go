// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudmonitor_logs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudmonitor_logs/api_server/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"net/http"
	"slices"
)

type CloudMonitorLogsService struct {
	pb.UnimplementedCloudMonitorLogsServiceServer
	cfg config.Config
}

func NewCloudMonitorLogsService(config config.Config) (*CloudMonitorLogsService, error) {
	return &CloudMonitorLogsService{
		cfg: config,
	}, nil
}

func (s *CloudMonitorLogsService) SearchAllLogs(ctx context.Context, req *pb.SearchAllLogsRequest) (*pb.SearchAllLogsResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudMonitorLogsService.SearchAllLogs").WithValues("cloudAccountId", req.CloudAccountId, "resourceId", req.ResourceId).Start()
	defer span.End()

	logger.Info("CloudMonitorLogsService: SearchAllLogs invoked")

	if req.ResourceType == "IKS" {
		fmt.Println("Resource type: ", req.ResourceType)
	} else {
		fmt.Println("Invalid resource type!!")
		// logger.Error(err, "Incorrect format of resourceId in input")
		// span.SetAttributes(
		// 	attribute.String("resourceId", req.ResourceId),
		// )
		// span.RecordError(fmt.Errorf("Incorrect format of resourceId in input"))
		return &pb.SearchAllLogsResponse{}, fmt.Errorf("Invalid resource type!!")
	}

	openSearchURL := s.cfg.OpenSearchEndpoint
	indexName := "cm_logs_" + req.CloudAccountId + "*"
	clusterRegion := s.cfg.ClusterRegion
	useProxy := s.cfg.UseProxy
	insecureSkipVerify := s.cfg.InsecureSkipVerify
	maxPageSize := int32(200)
	cloudAccountId := req.CloudAccountId
	clusterId := req.ResourceId

	// validate Timestamp and convert to UTC string with miliseconds
	startTime, endTime := validateTimestamp(req.StartTime, req.EndTime)
	// startTime := ConvertEpochToUTC(startTimeEpoch)
	// endTime := ConvertEpochToUTC(endTimeEpoch)

	// pagination
	size := req.Size
	if size > maxPageSize {
		fmt.Println("size requested is out of range, reseting to max size!!")
		size = maxPageSize
	}
	if size < int32(1) {
		fmt.Println("size requested is out of range, reseting to normal size!!")
		size = 10
	}
	pageNumber := req.PageNumber
	if pageNumber < int32(1) {
		fmt.Println("negative page number requested, reseting to one!!")
		pageNumber = 1
	}
	from := (pageNumber - 1) * size

	query := fmt.Sprintf(
		`{
		"sort": [
			{
				"@timestamp": {
					"order": "desc"
				}
			}
		],
		"size": %d,
		"from": %d,
		"query": {
			"bool": {
				"filter": [
					{
						"term": {
							"cluster_region": "%s"
						}
					},
					{
						"term": {
							"cluster_id": "%s"
						}
					},
					{
						"term": {
							"cloudaccount_id": "%s"
						}
					},
					{
						"range": {
							"@timestamp": {
								"gte": %d,
								"lte": %d,
								"format": "epoch_millis"
							}
						}
					}
				]
			}
		}
	}`, size, from, clusterRegion, clusterId, cloudAccountId, startTime, endTime)

	requestURL := fmt.Sprintf("%s/%s/_search/", openSearchURL, indexName)
	requestMethod := "GET"

	body, _, err := sendRequestToOpenSearch(useProxy, insecureSkipVerify, requestURL, requestMethod, query)
	if err != nil {
		fmt.Printf("unable to fetch data : %v\n", err)
		return &pb.SearchAllLogsResponse{}, fmt.Errorf("Unable to fetch data!!")
	}

	// Parse the JSON response
	var apiResponse pb.QueryResult
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Printf("unable to parse response from OpenSearch: %v\n", err)
		return &pb.SearchAllLogsResponse{}, fmt.Errorf("Unable to parse response data.")
	}

	returnResponse := &pb.SearchAllLogsResponse{
		Data: []*pb.Hits{},
	}
	returnResponse.Data = apiResponse.Hits.Hits

	return returnResponse, nil
}

func (s *CloudMonitorLogsService) SearchLogsByTerm(ctx context.Context, req *pb.SearchLogsByTermRequest) (*pb.SearchLogsByTermResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudMonitorLogsService.SearchLogsByTerm").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()

	logger.Info("CloudMonitorLogsService: SearchLogsByTerm invoked")

	if req.ResourceType == "IKS" {
		fmt.Println("Resource type: ", req.ResourceType)
	} else {
		fmt.Println("Invalid resource type!!")
		return &pb.SearchLogsByTermResponse{}, fmt.Errorf("Invalid resource type!!")
	}

	openSearchURL := s.cfg.OpenSearchEndpoint
	indexName := "cm_logs_" + req.CloudAccountId + "*"
	clusterRegion := s.cfg.ClusterRegion
	useProxy := s.cfg.UseProxy
	insecureSkipVerify := s.cfg.InsecureSkipVerify
	maxPageSize := int32(200)
	cloudAccountId := req.CloudAccountId
	clusterId := req.ResourceId
	searchTerm := req.SearchTerm

	// validate Timestamp and convert to UTC string with miliseconds
	startTime, endTime := validateTimestamp(req.StartTime, req.EndTime)
	// startTime := ConvertEpochToUTC(startTimeEpoch)
	// endTime := ConvertEpochToUTC(endTimeEpoch)

	// pagination
	size := req.Size
	if size > maxPageSize {
		fmt.Println("size requested is out of range, reseting to max size!!")
		size = maxPageSize
	}
	if size < int32(1) {
		fmt.Println("size requested is out of range, reseting to normal size!!")
		size = 10
	}
	pageNumber := req.PageNumber
	if pageNumber < int32(1) {
		fmt.Println("invalid page number requested, reseting to zero!!")
		pageNumber = 1
	}
	from := (pageNumber - 1) * size

	query := fmt.Sprintf(
		`{
		"sort": [
			{
				"@timestamp": {
					"order": "desc"
				}
			}
		],
		"size": %d,
		"from": %d,
		"query": {
			"bool": {
				"filter": [
					{
						"multi_match": {
							"type": "best_fields",
							"query": "%s",
							"lenient": true
						}
					},
					{
						"term": {
							"cluster_region": "%s"
						}
					},
					{
						"term": {
							"cluster_id": "%s"
						}
					},
					{
						"term": {
							"cloudaccount_id": "%s"
						}
					},
					{
						"range": {
							"@timestamp": {
								"gte": %d,
								"lte": %d,
								"format": "epoch_millis"
							}
						}
					}
				]
			}
		}
	}`, size, from, searchTerm, clusterRegion, clusterId, cloudAccountId, startTime, endTime)

	requestURL := fmt.Sprintf("%s/%s/_search/", openSearchURL, indexName)
	requestMethod := "GET"

	body, _, err := sendRequestToOpenSearch(useProxy, insecureSkipVerify, requestURL, requestMethod, query)
	if err != nil {
		fmt.Printf("unable to fetch data : %v\n", err)
		return &pb.SearchLogsByTermResponse{}, fmt.Errorf("Unable to fetch data!!")
	}

	// Parse the JSON response
	var apiResponse pb.QueryResult
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Printf("unable to parse response from OpenSearch: %v\n", err)
		return &pb.SearchLogsByTermResponse{}, fmt.Errorf("Unable to parse response data.")
	}

	returnResponse := &pb.SearchLogsByTermResponse{
		Data: []*pb.Hits{},
	}
	returnResponse.Data = apiResponse.Hits.Hits

	return returnResponse, nil
}

func (s *CloudMonitorLogsService) SearchPanelData(ctx context.Context, req *pb.SearchPanelDataRequest) (*pb.SearchPanelDataResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudMonitorLogsService.GetSystemComponents").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()

	logger.Info("CloudMonitorLogsService: GetSystemComponents invoked")

	allowedAggregationFieldNames := []string{}
	if req.ResourceType == "IKS" {
		allowedAggregationFieldNames = s.cfg.IKSAggregationFieldNames
	} else {
		fmt.Println("Invalid resource type!!")
		return &pb.SearchPanelDataResponse{}, fmt.Errorf("Invalid resource type!!")
	}

	fieldName := ""
	// validate field name for selected resource type
	if slices.Contains(allowedAggregationFieldNames, req.FieldName) {
		fieldName = req.FieldName
	} else {
		fmt.Println("Invalid fieldName aggregation for selected resource type!!")
		return &pb.SearchPanelDataResponse{}, fmt.Errorf("Invalid fieldName provided for aggregation!!")
	}

	openSearchURL := s.cfg.OpenSearchEndpoint
	indexName := "cm_logs_" + req.CloudAccountId + "*"
	clusterRegion := s.cfg.ClusterRegion
	useProxy := s.cfg.UseProxy
	insecureSkipVerify := s.cfg.InsecureSkipVerify
	cloudAccountId := req.CloudAccountId
	clusterId := req.ResourceId

	// validate Timestamp and convert to UTC string with miliseconds
	startTime, endTime := validateTimestamp(req.StartTime, req.EndTime)
	// startTime := ConvertEpochToUTC(startTimeEpoch)
	// endTime := ConvertEpochToUTC(endTimeEpoch)

	size := req.Size
	// maximum allowed unique component
	maxSize := int32(200)
	if size > 200 || size < int32(1) {
		fmt.Println("size requested is out of range, reseting to max size!!")
		size = maxSize
	}

	query := fmt.Sprintf(
		`{
			"size": 0,
			"aggs": {
				"unique_field_values": {
					"terms": {
						"field": "%s",
						"size": %d
					}
				}
			},
			"query": {
				"bool": {
					"filter": [
						{
							"term": {
								"cluster_region": "%s"
							}
						},
						{
							"term": {
								"cluster_id": "%s"
							}
						},
						{
							"term": {
								"cloudaccount_id": "%s"
							}
						},
						{
							"range": {
								"@timestamp": {
									"gte": %d,
									"lte": %d,
									"format": "epoch_millis"
								}
							}
						}
					]
				}
			}
		}`, fieldName, size, clusterRegion, clusterId, cloudAccountId, startTime, endTime)

	requestURL := fmt.Sprintf("%s/%s/_search/", openSearchURL, indexName)
	requestMethod := "GET"

	body, _, err := sendRequestToOpenSearch(useProxy, insecureSkipVerify, requestURL, requestMethod, query)
	if err != nil {
		fmt.Printf("unable to fetch data : %v\n", err)
		return &pb.SearchPanelDataResponse{}, fmt.Errorf("Unable to fetch data!!")
	}

	// Parse the JSON response
	var apiResponse pb.QueryResult
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Printf("unable to parse response from OpenSearch: %v\n", err)
		return &pb.SearchPanelDataResponse{}, fmt.Errorf("Unable to parse response data.")
	}

	returnResponse := &pb.SearchPanelDataResponse{
		Data: []*pb.Buckets{},
	}
	returnResponse.Data = apiResponse.Aggregations.UniqueFieldValues.Buckets

	return returnResponse, nil
}

func (s *CloudMonitorLogsService) SearchLogsByFilter(ctx context.Context, req *pb.SearchLogsByFilterRequest) (*pb.SearchLogsByFilterResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudMonitorLogsService.SearchLogsByFilter").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()

	logger.Info("CloudMonitorLogsService: SearchLogsByFilter invoked")

	// select valid filters based on resource type
	allowedFilterFieldNames := []string{}
	if req.ResourceType == "IKS" {
		allowedFilterFieldNames = s.cfg.IKSFilterFieldNames
	} else {
		fmt.Println("Invalid resource type!!")
		return &pb.SearchLogsByFilterResponse{}, fmt.Errorf("Invalid resource type!!")
	}

	// length of FieldNames and FieldValues should be same
	if len(req.FieldNames) != len(req.FieldValues) {
		fmt.Println("fieldNames and fieldValues should be of the same length.")
		return &pb.SearchLogsByFilterResponse{}, fmt.Errorf("number of selected filters and their values should be of the same length.")
	}

	// validate filter field names
	fieldNames, fieldValues := validateFiledNames(allowedFilterFieldNames, req.FieldNames, req.FieldValues)

	if len(fieldNames) < 1 {
		fmt.Println("selected filters are not allowed for resource type, unable to fetch data.")
		return &pb.SearchLogsByFilterResponse{}, fmt.Errorf("unable to fetch data for selected filters.")
	}

	openSearchURL := s.cfg.OpenSearchEndpoint
	indexName := "cm_logs_" + req.CloudAccountId + "*"
	clusterRegion := s.cfg.ClusterRegion
	useProxy := s.cfg.UseProxy
	insecureSkipVerify := s.cfg.InsecureSkipVerify
	maxPageSize := int32(200)
	cloudAccountId := req.CloudAccountId
	clusterId := req.ResourceId

	// validate Timestamp and convert to UTC string with miliseconds
	startTime, endTime := validateTimestamp(req.StartTime, req.EndTime)
	// startTime := ConvertEpochToUTC(startTimeEpoch)
	// endTime := ConvertEpochToUTC(endTimeEpoch)

	// pagination
	size := req.Size
	if size > maxPageSize {
		fmt.Println("size requested is out of range, reseting to max size!!")
		size = maxPageSize
	}
	if size < int32(1) {
		fmt.Println("size requested is out of range, reseting to normal size!!")
		size = 10
	}
	pageNumber := req.PageNumber
	if pageNumber < int32(1) {
		fmt.Println("invalid page number requested, reseting to zero!!")
		pageNumber = 1
	}
	from := (pageNumber - 1) * size

	// build dynamic query
	query, err := BuildQuery(fieldNames, fieldValues, cloudAccountId, clusterRegion, clusterId, size, from, startTime, endTime)
	if err != nil {
		fmt.Printf("Error building query: %v\n", err)
		return &pb.SearchLogsByFilterResponse{}, fmt.Errorf("Unable to fetch data!!")
	}

	requestURL := fmt.Sprintf("%s/%s/_search/", openSearchURL, indexName)
	requestMethod := "GET"

	body, _, err := sendRequestToOpenSearch(useProxy, insecureSkipVerify, requestURL, requestMethod, query)
	if err != nil {
		fmt.Printf("unable to fetch data : %v\n", err)
		return &pb.SearchLogsByFilterResponse{}, fmt.Errorf("Unable to fetch data!!")
	}

	// Parse the JSON response
	var apiResponse pb.QueryResult
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Printf("unable to parse response from OpenSearch: %v\n", err)
		return &pb.SearchLogsByFilterResponse{}, fmt.Errorf("Unable to parse response data.")
	}

	returnResponse := &pb.SearchLogsByFilterResponse{
		Data: []*pb.Hits{},
	}
	returnResponse.Data = apiResponse.Hits.Hits

	return returnResponse, nil

}

func (s *CloudMonitorLogsService) SearchLogsByPhrase(ctx context.Context, req *pb.SearchLogsByPhraseRequest) (*pb.SearchLogsByPhraseResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudMonitorLogsService.SearchLogsByPhrase").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()

	logger.Info("CloudMonitorLogsService: SearchLogsByPhrase invoked")

	if req.ResourceType == "IKS" {
		fmt.Println("Resource type: ", req.ResourceType)
	} else {
		fmt.Println("Invalid resource type!!")
		return &pb.SearchLogsByPhraseResponse{}, fmt.Errorf("Invalid resource type!!")
	}

	openSearchURL := s.cfg.OpenSearchEndpoint
	indexName := "cm_logs_" + req.CloudAccountId + "*"
	clusterRegion := s.cfg.ClusterRegion
	useProxy := s.cfg.UseProxy
	insecureSkipVerify := s.cfg.InsecureSkipVerify
	maxPageSize := int32(200)
	cloudAccountId := req.CloudAccountId
	clusterId := req.ResourceId
	searchPhrase := req.SearchPhrase

	// validate Timestamp and convert to UTC string with miliseconds
	startTime, endTime := validateTimestamp(req.StartTime, req.EndTime)
	// startTime := ConvertEpochToUTC(startTimeEpoch)
	// endTime := ConvertEpochToUTC(endTimeEpoch)

	// pagination
	size := req.Size
	if size > maxPageSize {
		fmt.Println("size requested is out of range, reseting to max size!!")
		size = maxPageSize
	}
	if size < int32(1) {
		fmt.Println("size requested is out of range, reseting to normal size!!")
		size = 10
	}
	pageNumber := req.PageNumber
	if pageNumber < int32(1) {
		fmt.Println("invalid page number requested, reseting to zero!!")
		pageNumber = 1
	}
	from := (pageNumber - 1) * size

	query := fmt.Sprintf(
		`{
		"sort": [
			{
				"@timestamp": {
					"order": "desc"
				}
			}
		],
		"size": %d,
		"from": %d,
		"query": {
			"bool": {
				"filter": [
					{
						"multi_match": {
							"type": "phrase",
							"query": "%s",
							"lenient": true
						}
					},
					{
						"term": {
							"cluster_region": "%s"
						}
					},
					{
						"term": {
							"cluster_id": "%s"
						}
					},
					{
						"term": {
							"cloudaccount_id": "%s"
						}
					},
					{
						"range": {
							"@timestamp": {
								"gte": %d,
								"lte": %d,
								"format": "epoch_millis"
							}
						}
					}
				]
			}
		}
	}`, size, from, searchPhrase, clusterRegion, clusterId, cloudAccountId, startTime, endTime)

	requestURL := fmt.Sprintf("%s/%s/_search/", openSearchURL, indexName)
	requestMethod := "GET"

	body, _, err := sendRequestToOpenSearch(useProxy, insecureSkipVerify, requestURL, requestMethod, query)
	if err != nil {
		fmt.Printf("unable to fetch data : %v\n", err)
		return &pb.SearchLogsByPhraseResponse{}, fmt.Errorf("Unable to fetch data!!")
	}

	// Parse the JSON response
	var apiResponse pb.QueryResult
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Printf("unable to parse response from OpenSearch: %v\n", err)
		return &pb.SearchLogsByPhraseResponse{}, fmt.Errorf("Unable to parse response data.")
	}

	returnResponse := &pb.SearchLogsByPhraseResponse{
		Data: []*pb.Hits{},
	}
	returnResponse.Data = apiResponse.Hits.Hits

	return returnResponse, nil

}

func (s *CloudMonitorLogsService) UserRegistration(ctx context.Context, req *pb.UserRegistrationRequest) (*pb.UserRegistrationResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudMonitorLogsService.UserRegistration").WithValues("cloudAccountId", req.CloudAccountId).Start()
	defer span.End()

	logger.Info("CloudMonitorLogsService: UserRegistration invoked")

	if req.ResourceType == "IKS" {
		fmt.Println("Resource type: ", req.ResourceType)
	} else {
		fmt.Println("Invalid resource type!!")
		return &pb.UserRegistrationResponse{}, fmt.Errorf("Invalid resource type!!")
	}

	responseItem := &pb.ResponseData{
		IsTemplateCreated: false,
		IsIndexCreated:    false,
		IsAliasCreated:    false,
	}

	returnResponse := &pb.UserRegistrationResponse{
		Response: responseItem,
	}

	useProxy := s.cfg.UseProxy
	insecureSkipVerify := s.cfg.InsecureSkipVerify
	openSearchURL := s.cfg.OpenSearchEndpoint
	cloudAccountId := req.CloudAccountId
	templateName := "cm_logs_" + cloudAccountId + "_tmplt"
	indexName := "cm_logs_" + cloudAccountId + "*"
	aliasName := "cm_logs_" + cloudAccountId + "_wrt_als"

	// check if template exist for given cloudAccountID
	requestURL := fmt.Sprintf("%s/_index_template/%s", openSearchURL, templateName)
	requestMethod := "HEAD"
	_, respStatusCode, err := sendRequestToOpenSearch(useProxy, insecureSkipVerify, requestURL, requestMethod, "")
	if err != nil {
		fmt.Printf("unable to fetch data from OpenSearch : %v\n", err)
		return &pb.UserRegistrationResponse{}, fmt.Errorf("Unable to fetch data from OpenSearch")
	}

	templateStatus := false
	if respStatusCode == http.StatusOK {
		fmt.Printf("Template %s exists\n", templateName)
		templateStatus = true
		responseItem.IsTemplateCreated = true
	} else if respStatusCode == http.StatusNotFound {
		fmt.Printf("Template %s does not exist\n", templateName)
		templateStatus = false
	} else {
		fmt.Printf("Unexpected status code for template: %d\n", respStatusCode)
		templateStatus = false
	}

	// if template does not exists
	if !templateStatus {
		// create a template
		url, body, err := CreateTemplate(cloudAccountId, openSearchURL)
		if err != nil {
			fmt.Printf("Error creating a template: %v\n", err)
			return &pb.UserRegistrationResponse{}, fmt.Errorf("Error creating a template for Index")
		}

		// call to sendRequestToOpenSearch() func to create a template
		_, respStatusCode, err := sendRequestToOpenSearch(useProxy, insecureSkipVerify, url, "PUT", body)
		if err != nil {
			fmt.Printf("unable to fetch data from OpenSearch : %v\n", err)
			return &pb.UserRegistrationResponse{}, fmt.Errorf("Unable to fetch data from OpenSearch")
		}

		if respStatusCode == http.StatusOK {
			fmt.Println("Template created successfully")
			responseItem.IsTemplateCreated = true
		} else {
			fmt.Println("Error while creating an index")
			return &pb.UserRegistrationResponse{}, fmt.Errorf("Error creating a template for Index")
		}
	}

	// check if index exist for given cloudAccountID
	requestURL = fmt.Sprintf("%s/%s", openSearchURL, aliasName)
	_, respStatusCode, err = sendRequestToOpenSearch(useProxy, insecureSkipVerify, requestURL, "HEAD", "")
	if err != nil {
		fmt.Printf("unable to fetch data from OpenSearch : %v\n", err)
		return &pb.UserRegistrationResponse{}, fmt.Errorf("Unable to fetch data from OpenSearch")
	}

	indexStatus := false
	if respStatusCode == http.StatusOK {
		fmt.Printf("Index %s exists with alias name %s\n", indexName, aliasName)
		indexStatus = true
		responseItem.IsIndexCreated = true
		responseItem.IsAliasCreated = true
	} else if respStatusCode == http.StatusNotFound {
		fmt.Printf("Index %s does not exist\n", indexName)
		indexStatus = false
	} else {
		fmt.Printf("Unexpected status code for template: %d\n", respStatusCode)
		indexStatus = false
	}

	// if index does not exists
	if !indexStatus {
		// create an index
		url, body, err := CreateIndex(cloudAccountId, openSearchURL)
		if err != nil {
			fmt.Printf("Error creating index: %v\n", err)
			return &pb.UserRegistrationResponse{}, fmt.Errorf("Error creating an Index")
		}

		// call to OpenSearch to create an Index
		_, respStatusCode, err := sendRequestToOpenSearch(useProxy, insecureSkipVerify, url, "PUT", body)
		if err != nil {
			fmt.Printf("unable to fetch data from OpenSearch : %v\n", err)
			return &pb.UserRegistrationResponse{}, fmt.Errorf("Unable to fetch data from OpenSearch")
		}

		if respStatusCode == http.StatusOK {
			fmt.Println("Index and Alias created successfully")
			responseItem.IsIndexCreated = true
			responseItem.IsAliasCreated = true
		} else if respStatusCode == http.StatusBadRequest {
			fmt.Println("Index already exists with same name")
		} else {
			fmt.Println("Error while creating an index")
			return &pb.UserRegistrationResponse{}, fmt.Errorf("Error creating an Index")
		}
	}

	return returnResponse, nil
}
