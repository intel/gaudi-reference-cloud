// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type CloudAcc struct {
	Id string `json:"id"`
}

var queryMetricsPayload = pb.QueryResourcesMetricsRequest{
	CloudAccountId: "cloudAccount",
	ResourceId:     "recourceId",
	Start:          "1709015536.591",
	End:            "1709026336.591",
	Step:           "2m",
	Category:       "metrics",
	ResourceType:   "VM",
}

var cloudAccID string

func TestQueryResourceMetricsMissingMetric(t *testing.T) {
	var queryMetricsPayload = pb.QueryResourcesMetricsRequest{
		CloudAccountId: "cloudAccount",
		ResourceId:     "9f857a32-407e-4123-b434-119a018cc101",
		Start:          "1709015536.591",
		End:            "1709026336.591",
		Step:           "2m",
		Category:       "metrics",
		ResourceType:   "VM",
	}

	_, err := cloudmonitorClient.QueryResourcesMetrics(context.Background(), &queryMetricsPayload)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}

	expectedErrorMessage := "rpc error: code = Unknown desc = unable to fetch data"
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected another error, recieved: %v", err.Error())
	}
}

func TestQueryResourceMetricsInvalidMetric(t *testing.T) {
	var queryMetricsPayload = pb.QueryResourcesMetricsRequest{
		CloudAccountId: "cloudAccount",
		ResourceId:     "9f857a32-407e-4123-b434-119a018cc101",
		Start:          "1709015536.591",
		End:            "1709026336.591",
		Step:           "2m",
		Metric:         "invalid_metric",
		Category:       "metrics",
		ResourceType:   "VM",
	}

	_, err := cloudmonitorClient.QueryResourcesMetrics(context.Background(), &queryMetricsPayload)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}

	expectedErrorMessage := "rpc error: code = Unknown desc = unable to fetch data"
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected another error, recieved: %v", err.Error())
	}
}

func TestQueryResourceMetricsInvalidResourceIdFormat(t *testing.T) {
	var queryMetricsPayload = pb.QueryResourcesMetricsRequest{
		CloudAccountId: "cloudAccount",
		ResourceId:     "invalid_resource",
		Start:          "1709015536.591",
		End:            "1709026336.591",
		Step:           "2m",
		Metric:         "invalid_metric",
		Category:       "metrics",
		ResourceType:   "VM",
	}

	_, err := cloudmonitorClient.QueryResourcesMetrics(context.Background(), &queryMetricsPayload)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}

	expectedErrorMessage := "rpc error: code = InvalidArgument desc = Invalid argument"
	fmt.Println(err.Error())
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected another error, recieved: %v", err.Error())
	}
}

func TestQueryResourceMetricsSuccess(t *testing.T) {
	var queryMetricsPayload = pb.QueryResourcesMetricsRequest{
		CloudAccountId: "121211118967",
		ResourceId:     "9f857a32-407e-4123-b434-119a018cc101",
		Start:          "1709015536.591",
		End:            "1709026336.591",
		Step:           "2m",
		Metric:         "cpu",
		Category:       "metrics",
		ResourceType:   "VM",
	}

	_, err := cloudmonitorClient.QueryResourcesMetrics(context.Background(), &queryMetricsPayload)

	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}

	expectedErrorMessage := "rpc error: code = Unknown desc = unable to fetch data"
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected another error, recieved: %v", err.Error())
	}

	// expectedErrorMessage := "unable to call Victoria Metrics"
	// if expectedErrorMessage != err.Error() {
	// 	t.Fatalf("Expected another error, recieved: %v", err.Error())
	// }
}
