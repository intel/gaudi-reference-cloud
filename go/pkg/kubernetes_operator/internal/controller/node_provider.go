// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/pkg/errors"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	compute "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/node_provider/compute"
	harvester "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/node_provider/harvester"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	HarvesterProviderName = "Harvester"
	ComputeProviderName   = "Compute"
)

type nodeProvider interface {
	// CreateNode creates a new node based on the settings of the nodegroup it belongs to.
	CreateNode(context.Context, string, string, string, string, string, privatecloudv1alpha1.Nodegroup) (privatecloudv1alpha1.NodeStatus, error)
	// GetNodes gets all nodes that belong to a nodegroup.
	GetNodes(context.Context, string, string) ([]privatecloudv1alpha1.NodeStatus, error)
	// GetNode gets the individual node that belong to a nodegroup.
	GetNode(context.Context, string, string) (privatecloudv1alpha1.NodeStatus, error)
	// DeleteNode deletes a node.
	DeleteNode(context.Context, string, string) error
	// CreateInstanceGroup creates multiple instances that are part of an instance group. This is used by gaudi clusters.
	CreateInstanceGroup(context.Context, string, string, string, int, privatecloudv1alpha1.Nodegroup) ([]privatecloudv1alpha1.NodeStatus, string, error)
	// CreatePrivateInstanceGroup creates multiple private instances that are part of an instance group. This is used by gaudi clusters.
	CreatePrivateInstanceGroup(context.Context, string, string, string, int, privatecloudv1alpha1.Nodegroup) ([]privatecloudv1alpha1.NodeStatus, string, error)
	// DeleteInstanceGroupMember deletes the specific instance group node in an instance group
	DeleteInstanceGroupMember(context.Context, string, string, string) error
	// ScaleUpInstanceGroup adds the missing nodes in an instance group
	ScaleUpInstanceGroup(context.Context, string, string, string, int, privatecloudv1alpha1.Nodegroup, string) ([]privatecloudv1alpha1.NodeStatus, string, error)
	// SearchInstanceGroup validates the instance group existence for a cloud account
	SearchInstanceGroup(context.Context, string, string) (bool, error)
}

func newNodeProvider(provider string, config *Config) (nodeProvider, error) {
	ctx := context.Background()
	if provider == HarvesterProviderName {
		harvesterProvider, err := harvester.NewHarvesterProvider(
			config.NodeProviders.Harvester.URL,
			config.NodeProviders.Harvester.AccessKey,
			config.NodeProviders.Harvester.SecretKey)
		if err != nil {
			return nil, err
		}

		return harvesterProvider, nil
	}

	if provider == ComputeProviderName {
		creds, err := grpcutil.GetClientCredentials(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "Get client credentials")
		}

		clientOptions := []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		clientCon, err := grpcutil.NewClient(ctx, config.NodeProviders.Compute.URL,
			clientOptions...)
		if err != nil {
			return nil, errors.Wrapf(err, "Create grpc client")
		}

		computeClient := pb.NewInstanceServiceClient(clientCon)
		privateComputeClient := pb.NewInstancePrivateServiceClient(clientCon)
		instanceGroupClient := pb.NewInstanceGroupServiceClient(clientCon)
		computeInstanceGroupClient := pb.NewInstanceGroupPrivateServiceClient(clientCon)

		computeProvider, err := compute.NewComputeProvider(ctx, computeClient, privateComputeClient, instanceGroupClient, computeInstanceGroupClient)
		if err != nil {
			return nil, errors.Wrapf(err, "Create compute provider")
		}

		return computeProvider, nil
	}

	return nil, errors.New("Provider not found")
}
