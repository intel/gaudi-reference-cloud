// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true

// BucketMeteMeteringMonitorConfig is the Schema for the bucketmeteringmonitorconfigs API.
// It stores the configuration for the Bucket Metering Monitor.
type BucketMeteringMonitorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ControllerManagerConfigurationSpec returns the configurations for controllers
	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	// Controller addr returns the adrress for weka controllers
	StorageControllerServerAddr string `json:"storageControllerServerAddr"`

	StorageControllerServerUseMtls bool `json:"storageControllerServerUseMtls"`

	// The address of the global metering service in format "host:port".
	MeteringServerAddr string `json:"meteringServerAddr"`

	// Usage records for running instances will be sent periodically with this interval.
	// The actual interval will be slightly less by a random value to avoid bursts.
	MaxUsageRecordSendIntervalMinutes int64 `json:"maxUsageRecordSendIntervalMinutes"`

	// product catalog match expression value for objectStorage service
	ServiceType string `json:"serviceType"`

	// Region
	Region string `json:"region"`
}

func init() {
	SchemeBuilder.Register(&BucketMeteringMonitorConfig{})
}
