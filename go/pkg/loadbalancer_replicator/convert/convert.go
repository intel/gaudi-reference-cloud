// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package convert

import (
	"encoding/json"
	"fmt"
	"strings"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	lbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Convert between Protobuf LoadBalancerPrivate and K8s private.cloud/v1alpha1/Loadbalancer.
type LoadBalancerConverter struct {
	marshaler          *grpcruntime.JSONPb
	regionId           string
	availabilityZoneId string
}

func NewLoadBalancerConverter(regionId, availabilityZoneId string) (*LoadBalancerConverter, error) {

	if regionId == "" {
		return nil, fmt.Errorf("regionId is required")
	}

	if availabilityZoneId == "" {
		return nil, fmt.Errorf("availabilityZoneId is required")
	}

	return &LoadBalancerConverter{
		marshaler: &grpcruntime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				// Force fields with default values, including for enums.
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		},
		regionId:           regionId,
		availabilityZoneId: availabilityZoneId,
	}, nil
}

func (c *LoadBalancerConverter) PbToK8s(loadbalancer *pb.LoadBalancerPrivate) (*lbv1alpha1.Loadbalancer, error) {
	if loadbalancer == nil {
		return nil, fmt.Errorf("LoadBalancerConverter.PbToK8s: loadbalancer is nil")
	}
	if loadbalancer.Metadata == nil {
		return nil, fmt.Errorf("LoadBalancerConverter.PbToK8s: required field Metadata is nil")
	}

	if loadbalancer.Spec == nil {
		return nil, fmt.Errorf("LoadBalancerConverter.PbToK8s: required field Spec is nil")
	}

	spec := &lbv1alpha1.LoadbalancerSpec{}
	if err := c.fromPb(loadbalancer.Spec, spec); err != nil {
		return nil, fmt.Errorf("LoadBalancerConverter.PbToK8s: Spec: %w", err)
	}

	// User-defined labels are not needed in the k8s load balancer metadata.labels field.
	spec.Labels = loadbalancer.Metadata.Labels
	status := &lbv1alpha1.LoadbalancerStatus{}
	if loadbalancer.Status != nil {
		if err := c.fromPb(loadbalancer.Status, status); err != nil {
			return nil, fmt.Errorf("LoadBalancerConverter.PbToK8s: Status: %w", err)
		}
	}

	listeners := []lbv1alpha1.LoadbalancerListener{}

	// Iterate over each listener
	for _, listener := range loadbalancer.Spec.Listeners {

		// Calculate static members
		members := []lbv1alpha1.VMember{}
		for _, resourceId := range listener.Pool.InstanceResourceIds {
			members = append(members, lbv1alpha1.VMember{
				InstanceResourceId: resourceId,
			})
		}

		listeners = append(listeners, lbv1alpha1.LoadbalancerListener{
			VIP: lbv1alpha1.VServer{
				Port:       int(listener.Port),
				IPProtocol: string(lbv1alpha1.IPProtocol_TCP),
				IPType:     string(lbv1alpha1.IPType_PUBLIC),
			},
			Pool: lbv1alpha1.VPool{
				Port:              int(listener.Pool.Port),
				Monitor:           listener.Pool.Monitor.String(),
				InstanceSelectors: listener.Pool.InstanceSelectors,
				Members:           members,
			},
			Owner: "",
		})
	}

	spec.Listeners = listeners

	lb := &lbv1alpha1.Loadbalancer{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: "private.cloud.intel.com/v1alpha1",
			Kind:       "Loadbalancer",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:              loadbalancer.Metadata.ResourceId,
			Namespace:         loadbalancer.Metadata.CloudAccountId,
			ResourceVersion:   loadbalancer.Metadata.ResourceVersion,
			CreationTimestamp: derefTime(fromPbTimestamp(loadbalancer.Metadata.CreationTimestamp)),
			DeletionTimestamp: fromPbTimestamp(loadbalancer.Metadata.DeletionTimestamp),
			// Add labels used by IDC operators.
			Labels: map[string]string{
				"cloud-account-id":  loadbalancer.Metadata.CloudAccountId,
				"regionId":          c.regionId,
				"availabiltyZoneId": c.availabilityZoneId,
			},
		},
		Spec:   *spec,
		Status: *status,
	}

	return lb, nil
}

func (c *LoadBalancerConverter) K8sToPb(loadbalancer *lbv1alpha1.Loadbalancer) (*pb.LoadBalancerPrivate, error) {
	if loadbalancer == nil {
		return nil, fmt.Errorf("LoadBalancerConverter.K8sToPb: instance is nil")
	}

	listeners := []*pb.LoadBalancerListener{}
	for _, listener := range loadbalancer.Spec.Listeners {

		// Convert Monitor Type
		monitorType := pb.LoadBalancerMonitorType_http
		switch strings.ToLower(listener.Pool.Monitor) {
		case string(lbv1alpha1.MonitorType_HTTP):
			monitorType = pb.LoadBalancerMonitorType_http
		case string(lbv1alpha1.MonitorType_HTTPS):
			monitorType = pb.LoadBalancerMonitorType_https
		case string(lbv1alpha1.MonitorType_TCP):
			monitorType = pb.LoadBalancerMonitorType_tcp
		default:
			// return nil, fmt.Errorf("invalid monitor type")
			monitorType = pb.LoadBalancerMonitorType_tcp
		}

		instanceResourceIds := []string{}
		for _, i := range listener.Pool.Members {
			instanceResourceIds = append(instanceResourceIds, i.InstanceResourceId)
		}

		listeners = append(listeners, &pb.LoadBalancerListener{
			Port: int32(listener.VIP.Port),
			Pool: &pb.LoadBalancerPool{
				Port:                int32(listener.Pool.Port),
				InstanceSelectors:   listener.Pool.InstanceSelectors,
				InstanceResourceIds: instanceResourceIds,
				Monitor:             monitorType,
			},
		})
	}

	spec := &pb.LoadBalancerSpecPrivate{
		Listeners: listeners,
		Security: &pb.LoadBalancerSecurity{
			Sourceips: loadbalancer.Spec.Security.Sourceips,
		},
	}
	status := &pb.LoadBalancerStatusPrivate{}
	if err := c.toPb(loadbalancer.Status, status); err != nil {
		return nil, fmt.Errorf("LoadBalancerConverter.K8sToPb: Status: %w", err)
	}
	return &pb.LoadBalancerPrivate{
		Metadata: &pb.LoadBalancerMetadataPrivate{
			CloudAccountId:    loadbalancer.ObjectMeta.Namespace,
			ResourceId:        loadbalancer.ObjectMeta.Name,
			ResourceVersion:   loadbalancer.ObjectMeta.ResourceVersion,
			Labels:            loadbalancer.Spec.Labels,
			CreationTimestamp: fromK8sTimestamp(&loadbalancer.ObjectMeta.CreationTimestamp),
			DeletionTimestamp: fromK8sTimestamp(loadbalancer.ObjectMeta.DeletionTimestamp),
		},
		Spec:   spec,
		Status: status,
	}, nil
}

// Convert Protobuf to JSON then to target.
func (c *LoadBalancerConverter) fromPb(source any, target any) error {
	jsonBytes, err := c.marshaler.Marshal(source)
	if err != nil {
		return fmt.Errorf("LoadBalancerConverter.fromPb: unable to serialize to json: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("LoadBalancerConverter.fromPb: unable to deserialize from json: %w", err)
	}
	return nil
}

// Convert generic struct to JSON then to Protobuf.
func (c *LoadBalancerConverter) toPb(source any, target any) error {
	jsonBytes, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("LoadBalancerConverter.toPb: unable to serialize to json: %w", err)
	}
	if err := c.marshaler.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("LoadBalancerConverter.toPb: unable to deserialize from json: %w", err)
	}
	return nil
}

func derefTime(t *k8smetav1.Time) k8smetav1.Time {
	if t == nil {
		return k8smetav1.Time{}
	}
	return *t
}

func fromPbTimestamp(t *timestamppb.Timestamp) *k8smetav1.Time {
	if t == nil {
		return nil
	}
	mt := k8smetav1.NewTime(t.AsTime())
	return &mt
}

func fromK8sTimestamp(t *k8smetav1.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t.Time)
}
