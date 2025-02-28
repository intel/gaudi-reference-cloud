// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudmonitor/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Status struct {
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Status    string `json:"status"`
}

type QueryResult struct {
	Status    string `json:"status"`
	IsPartial bool   `json:"isPartial"`
	Data      Data   `json:"data"`
}

type Data struct {
	ResultType string   `json:"resultType"`
	Result     []Result `json:"result"`
}

type Result struct {
	Metric Metric          `json:"metric"`
	Values [][]interface{} `json:"values"`
}

type Metric struct {
	Drive      string `json:"drive"`
	Hostname   string `json:"hostname"`
	Verb       string `json:"verb"`
	Code       string `json:"code"`
	Device     string `json:"device"`
	Mode       string `json:"mode"`
	Mountpoint string `json:"mountpoint"`
}

var fieldOpts []protodb.FieldOptions = []protodb.FieldOptions{}

func NewCloudMonitorService(session *sql.DB, cfg config.Config) (*Server, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}
	return &Server{
		session: session,
		cfg:     cfg,
	}, nil
}

var tracer trace.Tracer
var (
	port = flag.Int("port", 50051, "gRPC server port")
)

type Server struct {
	pb.UnimplementedCloudMonitorServiceServer
	session *sql.DB
	cfg     config.Config
}

func (c *Server) QueryResourcesMetrics(ctx context.Context, req *pb.QueryResourcesMetricsRequest) (*pb.QueryResourcesMetricsResponse, error) {
	//first(ctx, req)
	preview := false
	apiEndpointVM := ""
	tokenVM := ""
	query := ""
	unit := ""
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudMonitorService.QueryResourcesMetrics").WithValues("cloudAccountId", req.CloudAccountId,
		"resourceId", req.ResourceId).Start()

	defer span.End()

	// Validate resourceId.
	if strings.HasPrefix(req.ResourceId, "request") || strings.HasPrefix(req.ResourceId, "reservation") {
		preview = true
	} else {
		preview = false
	}

	if !preview && req.ResourceType != "IKS" {
		if _, err := uuid.Parse(req.ResourceId); err != nil {
			log.Error(err, "Incorrect format of resourceId in input")
			span.SetAttributes(
				attribute.String("resourceId", req.ResourceId),
			)
			span.RecordError(fmt.Errorf("Incorrect format of resourceId in input"))
			return &pb.QueryResourcesMetricsResponse{}, status.Error(grpccodes.InvalidArgument, "Invalid argument")

		}
	} else {
		fmt.Println("resourceId:", req.ResourceId)
	}

	// log.Info("Request", "req", req)
	returnValue := &pb.QueryResourcesMetricsResponse{
		Response: []*pb.QueryResourcesMetricsResponseItem{},
	}
	itemized := &pb.QueryResourcesMetricsResponseItem{

		Queryvalue: []*pb.Queryvalue{},
		Unit:       "",
		Item:       "",
	}
	if req.ResourceType == "VM" {
		apiEndpoint := c.cfg.VictoriaMetricsAddr
		filePath := "/vault/secrets/vmtoken"
		token, err := os.ReadFile(filePath)
		if err != nil {
			// fmt.Println("Error creating request:", err)
			log.Error(err, "Unable to read ca for  Victoria Metrics ")
			return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")
		}
		apiEndpointVM = apiEndpoint
		tokenVM = string(token)
	} else if req.ResourceType == "IKS" {
		apiEndpoint := c.cfg.RemoteWriteIKSAddr
		filePath := "/vault/secrets/vmtokeniks"
		token, err := os.ReadFile(filePath)
		if err != nil {
			// fmt.Println("Error creating request:", err)
			log.Error(err, "Unable to read ca for  Victoria Metrics ")
			return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")
		}
		apiEndpointVM = apiEndpoint
		tokenVM = string(token)
	} else if req.ResourceType == "BM" {
		if !c.cfg.EnableMetricsBM {
			return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("Metrics is not enabled for BM resource")
		}

		// cluster := c.cfg.VMClusterName  //"ilens-dev"
		// role := c.cfg.IamRole           //"arn:aws:iam::390677890188:role/cloudmonitor-describe-cluster-role"
		// server := c.cfg.ClusterEndpoint //"https://E9403701F90A90B408CFF9ABCC4144C5.gr7.us-west-2.eks.amazonaws.com"
		// region := c.cfg.AwsVMClusterRegion

		//fmt.Println("Request", "req", req)

		// filePathAccessId := "/vault/secrets/accessid"
		// filePathAccessSecret := "/vault/secrets/accesssecret"
		//caFile := "/vault/secrets/eksserverca"
		apiEndpoint := c.cfg.RemoteReadAddrBM

		filePath := "/vault/secrets/vmtokenbm"
		token, err := os.ReadFile(filePath)
		if err != nil {
			// fmt.Println("Error creating request:", err)
			log.Error(err, "Unable to read token for  Victoria Metrics ")
			return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")
		}
		vmRecord, err := GetVMRecord(req.CloudAccountId, c, ctx)
		if err != nil {
			// fmt.Println("Error creating request:", err)
			log.Error(err, "Unable to get VM ID")
			return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to generate query")
		}
		apiEndpoint = strings.Replace(apiEndpoint, "%id%", vmRecord, -1)

		// token, err := GetVMUserAWSAuth(caFile, cluster, req.CloudAccountId, server, region, role)
		// if err != nil {
		// 	// fmt.Println("Error creating request:", err)
		// 	log.Error(err, "Unable to create a query for Victoria Metrics with Query Generator")
		// 	return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")
		// }
		apiEndpointVM = apiEndpoint
		tokenVM = string(token)
	} else {
		span.RecordError(fmt.Errorf("Incorrect resource type"))
		return &pb.QueryResourcesMetricsResponse{}, status.Error(grpccodes.InvalidArgument, "Invalid argument")
	}
	//"https://internal-placeholder.com/select/11/prometheus/api/v1/query_range"
	if req.ResourceType == "BM" {

		queryBM, unitBM, err := QueryGeneratorBM(req.Metric, req.CloudAccountId, req.ResourceId)
		if err != nil {
			// fmt.Println("Error creating request:", err)
			log.Error(err, "Unable to create a query for Victoria Metrics with Query Generator")
			return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")
		}
		query = queryBM
		unit = unitBM
	} else if req.ResourceType == "IKS" {

		queryBM, unitBM, err := QueryGeneratorIKS(req.Metric, req.CloudAccountId, req.ResourceId)
		if err != nil {
			// fmt.Println("Error creating request:", err)
			log.Error(err, "Unable to create a query for Victoria Metrics with Query Generator")
			return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")
		}
		query = queryBM
		unit = unitBM
	} else {
		queryVM, unitVM, err := QueryGenerator(req.Metric, req.CloudAccountId, req.ResourceId)
		if err != nil {
			// fmt.Println("Error creating request:", err)
			log.Error(err, "Unable to create a query for Victoria Metrics with Query Generator")
			return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")
		}
		query = queryVM
		unit = unitVM
	}

	client := &http.Client{}
	params := url.Values{}
	params.Add("start", req.Start)
	params.Add("end", req.End)
	params.Add("step", req.Step)
	params.Add("query", query)
	apiEndpointVM += "?" + params.Encode() //"query=" + query + "&start=" + req.Start + "&end=" + req.End + "&step=" + req.Step
	//apiEndpoint = `https://internal-placeholder.com/select/11/prometheus/api/v1/query_range?query=sum(avg(rate(kubevirt_vmi_vcpu_seconds%7Bnamespace%3D%22239002718779%22%2C%20name%3D%22eb70f9ed-c996-47ca-a8dd-10a5768f3da6%22%7D%5B5m%5D))%20by%20(domain%2C%20name))&start=1709015351.383&end=1709026151.383&step=30s` //"?" + params.Encode()
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if c.cfg.InsecureSkipVerify != true {

		caCert, err := os.ReadFile("/vault/secrets/rootca")
		if err != nil {
			log.Error(err, "Unable to read CA for Victoria Metrics")
			return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs: caCertPool}
	} else {

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	fmt.Println(apiEndpointVM)

	request, err := http.NewRequest("GET", apiEndpointVM, nil)
	if err != nil {
		// fmt.Println("Error creating request:", err)
		log.Error(err, "Error creating request to call Victoria Metrics endpoint")

		return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")
	}
	request.Header.Add("Authorization", "Bearer "+(string(tokenVM)))
	response, err := client.Do(request)
	if err != nil {
		log.Error(err, "Error calling Victoria Metrics endpoint")
		fmt.Println("Error making HTTP request:", err)
		fmt.Println("Error making HTTP request:", response)
		span.SetAttributes(
			attribute.String("resourceId", req.ResourceId),
		)
		span.SetStatus(codes.Error, "unable to fetch data from VictoriaMetrics")
		span.RecordError(fmt.Errorf("unable to fetch data"))
		return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")

	}
	//Add attributes to a span

	defer response.Body.Close()
	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {

		log.Error(err, "Error reading response body")
		span.SetAttributes(
			attribute.String("resourceId", req.ResourceId),
		)
		span.SetStatus(codes.Error, "unable to read response data from VictoriaMetrics")
		span.RecordError(fmt.Errorf("unable to read response data from VictoriaMetrics"))
		return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to fetch data")

	}

	// Parse the JSON response
	var apiResponse QueryResult
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		log.Error(err, "Error parsing response")
		span.SetAttributes(
			attribute.String("resourceId", req.ResourceId),
		)
		span.SetStatus(codes.Error, "unable to parse data from VictoriaMetrics")
		span.RecordError(fmt.Errorf("unable to parse data"))
		return &pb.QueryResourcesMetricsResponse{}, fmt.Errorf("unable to parse data")

	}

	// Check if Data, Result, and Values exist and have the correct structure

	if len(apiResponse.Data.Result) > 0 {
		for i, _ := range apiResponse.Data.Result {
			itemized = &pb.QueryResourcesMetricsResponseItem{

				Queryvalue: []*pb.Queryvalue{},
				Unit:       "",
				Item:       "",
			}
			// Access the parsed data
			for _, value := range apiResponse.Data.Result[i].Values {

				timeStamp := value[0].(float64)
				valueStr := value[1].(string)
				//fmt.Printf("Timestamp: %.0f, Value: %s\n", timeStamp, valueStr)
				targetObj := &pb.Queryvalue{}
				targetObj.Value = valueStr
				targetObj.Epochtime = strconv.FormatFloat(timeStamp, 'f', -1, 64)
				itemized.Queryvalue = append(itemized.Queryvalue, targetObj)

			}
			itemized.Unit = unit
			if req.ResourceType == "IKS" {
				if req.Metric == "apiserver_requestsbycode" {
					itemized.Item = apiResponse.Data.Result[i].Metric.Code
				} else if req.Metric == "apiserver_requestsbyverb" || req.Metric == "apiserver_errorsbyverb" || req.Metric == "apiserver_latencybyverb" {
					itemized.Item = apiResponse.Data.Result[i].Metric.Verb
				} else if req.Metric == "apiserver_latencybyhostname" {
					itemized.Item = apiResponse.Data.Result[i].Metric.Hostname + " " + apiResponse.Data.Result[i].Metric.Verb
				} else {
					itemized.Item = apiResponse.Data.Result[i].Metric.Hostname
				}
			} else if req.ResourceType == "BM" {
				if req.Metric == "network_transmit_bytes" || req.Metric == "io_traffic_read" || req.Metric == "io_traffic_write" || req.Metric == "network_receive_bytes" {
					itemized.Item = apiResponse.Data.Result[i].Metric.Device
				} else if req.Metric == "cpu" {
					itemized.Item = apiResponse.Data.Result[i].Metric.Mode
				} else if req.Metric == "disk" {
					itemized.Item = apiResponse.Data.Result[i].Metric.Mountpoint
				} else {
					itemized.Item = apiResponse.Data.Result[i].Metric.Hostname
				}
			} else {
				itemized.Item = apiResponse.Data.Result[i].Metric.Drive
			}

			returnValue.Response = append(returnValue.Response, itemized)
		}

		log.Info("Response sent")
		return returnValue, nil
	} else {
		span.SetAttributes(
			attribute.String("resourceId", req.ResourceId),
		)
		span.SetStatus(codes.Error, "data unavailable at the moment for this instance")
		span.RecordError(fmt.Errorf("data unavailable at the moment for this instance"))
		return &pb.QueryResourcesMetricsResponse{}, status.Error(grpccodes.Unavailable, "Data is unavailable at the moment. Please retry after sometime.")
	}
}

func QueryGenerator(metric string, cloudaccountid string, resourseid string) (string, string, error) {

	query := ""
	if metric == "cpu" {
		query = "avg(rate(kubevirt_vmi_vcpu_seconds{namespace=\"%namespace%\", name=\"%resourceid%\"}[2m]))"
		query = strings.Replace(query, "%namespace%", cloudaccountid, -1)
		query = strings.Replace(query, "%resourceid%", resourseid, -1)
		return query, "percentage", nil
	} else if metric == "memory" {
		query = `sum((kubevirt_vmi_memory_available_bytes{namespace="%namespace%", name="%resourceid%"} -kubevirt_vmi_memory_unused_bytes{namespace="%namespace%", name="%resourceid%"})/kubevirt_vmi_memory_available_bytes{namespace="%namespace%", name="%resourceid%"})`
		query = strings.Replace(query, "%namespace%", cloudaccountid, -1)
		query = strings.Replace(query, "%resourceid%", resourseid, -1)
		return query, "percentage", nil
	} else if metric == "network_receive_bytes" {
		query = `irate(kubevirt_vmi_network_receive_bytes_total{namespace="%namespace%", name="%resourceid%"}[2m])*8`
		query = strings.Replace(query, "%namespace%", cloudaccountid, -1)
		query = strings.Replace(query, "%resourceid%", resourseid, -1)
		return query, "b/s", nil

	} else if metric == "network_transmit_bytes" {
		query = `irate(kubevirt_vmi_network_transmit_bytes_total{namespace="%namespace%", name="%resourceid%"}[2m])*8`
		query = strings.Replace(query, "%namespace%", cloudaccountid, -1)
		query = strings.Replace(query, "%resourceid%", resourseid, -1)
		return query, "b/s", nil
	} else if metric == "storage_read_traffic_bytes" {
		query = `irate(kubevirt_vmi_storage_read_traffic_bytes_total{namespace="%namespace%", name="%resourceid%"}[2m])`
		query = strings.Replace(query, "%namespace%", cloudaccountid, -1)
		query = strings.Replace(query, "%resourceid%", resourseid, -1)
		return query, "B/s", nil

	} else if metric == "storage_write_traffic_bytes" {
		query = `irate(kubevirt_vmi_storage_write_traffic_bytes_total{namespace="%namespace%", name="%resourceid%"}[2m])`
		query = strings.Replace(query, "%namespace%", cloudaccountid, -1)
		query = strings.Replace(query, "%resourceid%", resourseid, -1)
		return query, "B/s", nil

	} else if metric == "storage_iops_read_total" {
		query = `irate(kubevirt_vmi_storage_iops_read_total{namespace="%namespace%", name="%resourceid%"}[2m])`
		query = strings.Replace(query, "%namespace%", cloudaccountid, -1)
		query = strings.Replace(query, "%resourceid%", resourseid, -1)
		return query, "io/s", nil

	} else if metric == "storage_iops_write_total" {
		query = `irate(kubevirt_vmi_storage_iops_write_total{namespace="%namespace%", name="%resourceid%"}[2m])`
		query = strings.Replace(query, "%namespace%", cloudaccountid, -1)
		query = strings.Replace(query, "%resourceid%", resourseid, -1)
		return query, "io/s", nil

	} else if metric == "storage_read_times_ms_total" {
		query = `irate(kubevirt_vmi_storage_read_times_ms_total{namespace="%namespace%", name="%resourceid%"}[2m])`
		query = strings.Replace(query, "%namespace%", cloudaccountid, -1)
		query = strings.Replace(query, "%resourceid%", resourseid, -1)
		return query, "ms", nil

	} else if metric == "storage_write_times_ms_total" {
		query = `irate(kubevirt_vmi_storage_write_times_ms_total{namespace="%namespace%", name="%resourceid%"}[2m])`
		query = strings.Replace(query, "%namespace%", cloudaccountid, -1)
		query = strings.Replace(query, "%resourceid%", resourseid, -1)
		return query, "ms", nil

	} else {
		// log.Info("Unsupported Metric")
		return "", "", fmt.Errorf("unsupported metric")
	}

}

var maxId int64 = 1_000_000_000_000

func NewId() (string, error) {
	intId, err := rand.Int(rand.Reader, big.NewInt(maxId))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%012d", intId), nil
}
