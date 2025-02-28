// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package utils

import (
	"context"
	"fmt"
	"reflect"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
)

// This method internally calls `TrimObjectForLogs`
// `TrimObjectForLogs`  deals with structure of pointers but doesn't return any explicit errors in case something goes wrong
// We don't want to abort the workflow in case of unexpected errors and therefore,
// The method recovers from errors returns nil
// On returning, it resumes API workflow execution
func TrimInstanceCloneForLogs(instance *cloudv1alpha1.Instance) *cloudv1alpha1.Instance {
	logger := log.FromContext(context.Background()).WithName("ComputeUtils.TrimInstanceCloneForLogs")
	// handle unexpected errors in TrimInstanceCloneForLogs
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Errorf("error occurred while sanitizing instance object for logs"), logkeys.Error)
		}
	}()

	// preprocess response
	objReflect := reflect.ValueOf(instance)
	if objReflect.IsValid() && !objReflect.IsZero() {
		objClone := instance.DeepCopy()
		switch reflect.ValueOf(objClone).Type().Kind() {
		case reflect.Ptr:
			log.TrimObjectForLogs(reflect.ValueOf(objClone).Elem())
		case reflect.Struct:
			log.TrimObjectForLogs(reflect.ValueOf(&objClone).Elem())
		}
		// return trimmed object copy
		return objClone
	} else {
		// return original object
		return instance
	}
}
