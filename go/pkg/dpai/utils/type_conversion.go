// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertTimeToTimestamppb(timeValue time.Time) *timestamppb.Timestamp {
	if !timeValue.IsZero() {
		timestampValue := timestamppb.New(timeValue)
		return timestampValue
	}
	return nil
}

func GetPostgressNamespaceName(name string) string {
	return fmt.Sprintf("postgres-%s", name)
}

func ConvertDpaiServiceTypeToPbEnum(val string) (*pb.DpaiServiceType, error) {
	var enumValue pb.DpaiServiceType
	switch val {
	case "DPAI_WORKSPACE":
		enumValue = pb.DpaiServiceType_DPAI_WORKSPACE
	case "DPAI_HMS":
		enumValue = pb.DpaiServiceType_DPAI_HMS
	case "DPAI_SPARK":
		enumValue = pb.DpaiServiceType_DPAI_SPARK
	case "DPAI_AIRFLOW":
		enumValue = pb.DpaiServiceType_DPAI_AIRFLOW
	case "DPAI_POSTGRES":
		enumValue = pb.DpaiServiceType_DPAI_POSTGRES
	default:
		return nil, fmt.Errorf("invalid ServiceType value: %s", val)
	}
	return &enumValue, nil
}

func ConvertDpaiDeploymentChangeIndicatorToPbEnum(val string) (*pb.DpaiDeploymentChangeIndicator, error) {
	var enumValue pb.DpaiDeploymentChangeIndicator
	switch val {
	case "DPAI_CREATE":
		enumValue = pb.DpaiDeploymentChangeIndicator_DPAI_CREATE
	case "DPAI_UPDATE":
		enumValue = pb.DpaiDeploymentChangeIndicator_DPAI_UPDATE
	case "DPAI_DELETE":
		enumValue = pb.DpaiDeploymentChangeIndicator_DPAI_DELETE
	default:
		return nil, fmt.Errorf("invalid changeIndicator value: %s", val)
	}
	return &enumValue, nil
}

func ConvertDpaiDeploymentStateToPbEnum(val string) (*pb.DpaiDeploymentState, error) {
	var enumValue pb.DpaiDeploymentState
	switch val {
	case "DPAI_ACCEPTED":
		enumValue = pb.DpaiDeploymentState_DPAI_ACCEPTED
	case "DPAI_CANCELLED":
		enumValue = pb.DpaiDeploymentState_DPAI_CANCELLED
	case "DPAI_FAILED":
		enumValue = pb.DpaiDeploymentState_DPAI_FAILED
	case "DPAI_PENDING":
		enumValue = pb.DpaiDeploymentState_DPAI_PENDING
	case "DPAI_QUEUED":
		enumValue = pb.DpaiDeploymentState_DPAI_QUEUED
	case "DPAI_RUNNING":
		enumValue = pb.DpaiDeploymentState_DPAI_RUNNING
	case "DPAI_SUCCESS":
		enumValue = pb.DpaiDeploymentState_DPAI_SUCCESS
	case "DPAI_UPSTREAM_CANCELLED":
		enumValue = pb.DpaiDeploymentState_DPAI_UPSTREAM_CANCELLED
	case "DPAI_UPSTREAM_FAILED":
		enumValue = pb.DpaiDeploymentState_DPAI_UPSTREAM_FAILED
	case "DPAI_WAITING_FOR_UPSTREAM":
		enumValue = pb.DpaiDeploymentState_DPAI_WAITING_FOR_UPSTREAM
	default:
		return nil, fmt.Errorf("invalid State value: %s", val)
	}
	return &enumValue, nil
}

func ConvertBytesToTags(b []byte) (map[string]string, error) {
	var tags map[string]string
	if err := json.Unmarshal(b, &tags); err != nil {
		return nil, fmt.Errorf("unable to convert bytes %+v to map[string]string. Error: %+v", b, err)
	}
	return tags, nil
}

func ConvertTagsToBytes(tags map[string]string) ([]byte, error) {
	b, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("not able to parse the tags: %+v \n Failed with error: %+v", tags, err)
	}
	return b, nil
}

// ChartReference
func ConvertBytesToChartReference(b []byte) (*pb.DpaiChartReference, error) {
	var chart pb.DpaiChartReference
	if err := json.Unmarshal(b, &chart); err != nil {
		return nil, fmt.Errorf("unable to convert bytes to DpaiChartReference. Error: %+v", err)
	}
	return &chart, nil
}

func ConvertChartReferenceToBytes(chart *pb.DpaiChartReference) ([]byte, error) {
	output, err := json.Marshal(chart)
	if err != nil {
		return nil, fmt.Errorf("not able to parse the tags: %+v \n Failed with error: %+v", chart, err)
	}
	return output, nil
}

func MergeChartReference(t1, t2 *pb.DpaiChartReference) *pb.DpaiChartReference {

	// Create a new instance of test
	merged := &pb.DpaiChartReference{}

	// Use reflection to iterate over the fields of the struct
	t1Value := reflect.ValueOf(t1).Elem()
	t2Value := reflect.ValueOf(t2).Elem()
	mergedValue := reflect.ValueOf(merged).Elem()

	for i := 0; i < t1Value.NumField(); i++ {
		// Copy the value of each field from t1 to merged
		field := mergedValue.Field(i)
		if field.CanSet() {
			field.Set(t1Value.Field(i))
		}
	}

	// Use reflection to iterate over the fields of the struct
	for i := 0; i < t2Value.NumField(); i++ {
		field := t2Value.Field(i)
		if field.Type().Kind() == reflect.String && field.Len() > 0 {
			// If it's a non-empty string, copy it to merged
			mergedValue.Field(i).Set(field)
		}
	}

	return merged
}

// ImageReference
func ConvertBytesToImageReference(b []byte) (*pb.DpaiImageReference, error) {
	var chart pb.DpaiImageReference
	if err := json.Unmarshal(b, &chart); err != nil {
		return nil, fmt.Errorf("unable to convert bytes to DpaiImageReference. Error: %+v", err)
	}
	return &chart, nil
}

func ConvertImageReferenceToBytes(chart *pb.DpaiImageReference) ([]byte, error) {
	output, err := json.Marshal(chart)
	if err != nil {
		return nil, fmt.Errorf("not able to parse the tags: %+v \n Failed with error: %+v", chart, err)
	}
	return output, nil
}

func MergeImageReference(t1, t2 *pb.DpaiImageReference) *pb.DpaiImageReference {

	// Create a new instance of test
	merged := &pb.DpaiImageReference{}

	// Use reflection to iterate over the fields of the struct
	t1Value := reflect.ValueOf(t1).Elem()
	t2Value := reflect.ValueOf(t2).Elem()
	mergedValue := reflect.ValueOf(merged).Elem()

	for i := 0; i < t1Value.NumField(); i++ {
		// Copy the value of each field from t1 to merged
		field := mergedValue.Field(i)
		if field.CanSet() {
			field.Set(t1Value.Field(i))
		}
	}

	// Use reflection to iterate over the fields of the struct
	for i := 0; i < t2Value.NumField(); i++ {
		field := t2Value.Field(i)
		if field.Type().Kind() == reflect.String && field.Len() > 0 {
			// If it's a non-empty string, copy it to merged
			mergedValue.Field(i).Set(field)
		}
	}

	return merged
}

// SecretReference
func ConvertBytesToSecretReference(b []byte) (*pb.DpaiSecretReference, error) {
	var data pb.DpaiSecretReference
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("unable to convert bytes to DpaiSecretReference. Error: %+v", err)
	}
	return &data, nil
}

func ConvertSecretReferenceToBytes(data *pb.DpaiSecretReference) ([]byte, error) {
	output, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("not able to parse the tags: %+v \n Failed with error: %+v", data, err)
	}
	return output, nil
}

func MergeStructs(t1, t2 interface{}) interface{} {
	// Get the reflect.Value of each struct
	v1 := reflect.ValueOf(t1)
	v2 := reflect.ValueOf(t2)

	// Make sure both values are structs
	if v1.Kind() != reflect.Struct || v2.Kind() != reflect.Struct {
		return nil
	}

	// Create a new instance of the same type as t1
	result := reflect.New(v1.Type()).Elem()

	// Iterate through fields of t1 and copy values to the result
	for i := 0; i < v1.NumField(); i++ {
		result.Field(i).Set(v1.Field(i))
	}

	// Iterate through fields of t2 and copy values to the result
	for i := 0; i < v2.NumField(); i++ {
		result.Field(i).Set(v2.Field(i))
	}

	return result.Interface()
}
