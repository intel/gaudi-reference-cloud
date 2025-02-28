// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package convert

import (
	"encoding/json"
	"fmt"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Convert between Protobuf InstancePrivate and K8s private.cloud/v1alpha1/Instance.
type InstanceConverter struct {
	marshaler *grpcruntime.JSONPb
}

func NewInstanceConverter() *InstanceConverter {
	return &InstanceConverter{
		marshaler: &grpcruntime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				// Force fields with default values, including for enums.
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
	}
}

func (c *InstanceConverter) PbToK8s(instance *pb.InstancePrivate) (*privatecloudv1alpha1.Instance, error) {
	if instance == nil {
		return nil, fmt.Errorf("InstanceConverter.PbToK8s: instance is nil")
	}
	if instance.Metadata == nil {
		return nil, fmt.Errorf("InstanceConverter.PbToK8s: required field Metadata is nil")
	}
	spec := &privatecloudv1alpha1.InstanceSpec{}
	if instance.Spec != nil {
		if err := c.fromPb(instance.Spec, spec); err != nil {
			return nil, fmt.Errorf("InstanceConverter.PbToK8s: Spec: %w", err)
		}
	}
	spec.InstanceName = instance.Metadata.Name
	// User-defined labels are not needed in the k8s Instance metadata.labels field.
	// Store labels in the spec so they can be copied to the Kubevirt VirtualMachine metadata.labels field.
	spec.Labels = instance.Metadata.Labels
	status := &privatecloudv1alpha1.InstanceStatus{}
	if instance.Status != nil {
		if err := c.fromPb(instance.Status, status); err != nil {
			return nil, fmt.Errorf("InstanceConverter.PbToK8s: Status: %w", err)
		}
	}
	return &privatecloudv1alpha1.Instance{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: "private.cloud.intel.com/v1alpha1",
			Kind:       "Instance",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:              instance.Metadata.ResourceId,
			Namespace:         instance.Metadata.CloudAccountId,
			ResourceVersion:   instance.Metadata.ResourceVersion,
			CreationTimestamp: derefTime(fromPbTimestamp(instance.Metadata.CreationTimestamp)),
			DeletionTimestamp: fromPbTimestamp(instance.Metadata.DeletionTimestamp),
			// Add labels used by IDC operators.
			Labels: map[string]string{
				"cloud-account-id":      instance.Metadata.CloudAccountId,
				"instance-category":     string(spec.InstanceTypeSpec.InstanceCategory),
				"supercompute-group-id": string(spec.SuperComputeGroupId),
				"cluster-group-id":      string(spec.ClusterGroupId),
				"cluster-id":            string(spec.ClusterId),
				"region":                string(spec.Region),
				"node-id":               string(spec.NodeId),
				"instance-group":        string(spec.InstanceGroup),
			},
		},
		Spec:   *spec,
		Status: *status,
	}, nil
}

func (c *InstanceConverter) K8sToPb(instance *privatecloudv1alpha1.Instance) (*pb.InstancePrivate, error) {
	if instance == nil {
		return nil, fmt.Errorf("InstanceConverter.K8sToPb: instance is nil")
	}
	spec := &pb.InstanceSpecPrivate{}
	if err := c.toPb(instance.Spec, spec); err != nil {
		return nil, fmt.Errorf("InstanceConverter.K8sToPb: Spec: %w", err)
	}
	status := &pb.InstanceStatusPrivate{}
	if err := c.toPb(instance.Status, status); err != nil {
		return nil, fmt.Errorf("InstanceConverter.K8sToPb: Status: %w", err)
	}
	return &pb.InstancePrivate{
		Metadata: &pb.InstanceMetadataPrivate{
			CloudAccountId:    instance.ObjectMeta.Namespace,
			ResourceId:        instance.ObjectMeta.Name,
			ResourceVersion:   instance.ObjectMeta.ResourceVersion,
			Labels:            instance.Spec.Labels,
			CreationTimestamp: fromK8sTimestamp(&instance.ObjectMeta.CreationTimestamp),
			DeletionTimestamp: fromK8sTimestamp(instance.ObjectMeta.DeletionTimestamp),
		},
		Spec:   spec,
		Status: status,
	}, nil
}

// Convert Protobuf to JSON then to target.
func (c *InstanceConverter) fromPb(source any, target any) error {
	jsonBytes, err := c.marshaler.Marshal(source)
	if err != nil {
		return fmt.Errorf("InstanceConverter.fromPb: unable to serialize to json: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("InstanceConverter.fromPb: unable to deserialize from json: %w", err)
	}
	return nil
}

// Convert generic struct to JSON then to Protobuf.
func (c *InstanceConverter) toPb(source any, target any) error {
	jsonBytes, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("InstanceConverter.toPb: unable to serialize to json: %w", err)
	}
	if err := c.marshaler.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("InstanceConverter.toPb: unable to deserialize from json: %w", err)
	}
	return nil
}

func fromPbTimestamp(t *timestamppb.Timestamp) *k8smetav1.Time {
	if t == nil {
		return nil
	}
	mt := k8smetav1.NewTime(t.AsTime())
	return &mt
}

func derefTime(t *k8smetav1.Time) k8smetav1.Time {
	if t == nil {
		return k8smetav1.Time{}
	}
	return *t
}

func fromK8sTimestamp(t *k8smetav1.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t.Time)
}
