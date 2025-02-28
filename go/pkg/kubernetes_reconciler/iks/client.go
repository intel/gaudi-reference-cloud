// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iks

import (
	"context"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	PendingRevisionState       RevisionState = "Pending"
	DeletePendingRevisionState RevisionState = "DeletePending"
)

type Client struct {
	Addr       string
	Ctx        context.Context
	GRPCClient pb.IksPrivateReconcilerClient
}

type RevisionState string

func NewClient(addr string) (*Client, error) {
	ctx := context.Background()

	// By default this loads the certs from /vault/secrets/, or
	// some env vars can be set with right path where certs are
	// stored.
	creds, err := grpcutil.GetClientCredentials(ctx)
	if err != nil {
		return nil, err
	}

	clientOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}
	clientConn, err := grpcutil.NewClient(ctx,
		addr,
		clientOptions...,
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		Addr:       addr,
		Ctx:        ctx,
		GRPCClient: pb.NewIksPrivateReconcilerClient(clientConn),
	}, nil
}

func (c *Client) Get(ctx context.Context, state RevisionState) ([]*pb.ClusterReconcilerResponseCluster, error) {
	resp, err := c.GRPCClient.GetClustersReconciler(ctx, &pb.ClusterReconcilerRequest{
		State: string(state),
	})
	if err != nil {
		return nil, err
	}

	return resp.Clusters, nil
}

// Update updates the cluster status into the api.
func (c *Client) UpdateClusterState(ctx context.Context, clusterUUID string, clusterState string) error {
	if _, err := c.GRPCClient.PutClusterStateReconciler(ctx, &pb.UpdateClusterStateRequest{
		Uuid:  clusterUUID,
		State: clusterState,
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateClusterStatus(ctx context.Context, clusterUUID string, clusterStatus privatecloudv1alpha1.ClusterStatus) error {
	var clusterStatusRequest pb.ClusterStatusRequest
	clusterStatusRequest.State = string(clusterStatus.State)
	clusterStatusRequest.LastUpdate = clusterStatus.LastUpdate.Format(time.RFC3339)
	clusterStatusRequest.Reason = clusterStatus.Reason
	clusterStatusRequest.Message = clusterStatus.Message
	clusterStatusRequest.Nodegroups = make([]*pb.NodeGroupStatusRequest, 0, len(clusterStatus.Nodegroups))
	clusterStatusRequest.Ilbs = make([]*pb.IlbStatusRequest, 0, len(clusterStatus.ILBS))
	clusterStatusRequest.Addons = make([]*pb.AddonStatusRequest, 0, len(clusterStatus.Addons))
	clusterStatusRequest.Firewall = make([]*pb.FirewallStatusRequest, 0, len(clusterStatus.Firewall))

	for _, nodegroup := range clusterStatus.Nodegroups {
		var nodeGroupStatusRequest pb.NodeGroupStatusRequest
		nodeGroupStatusRequest.Count = int32(nodegroup.Count)
		nodeGroupStatusRequest.Name = nodegroup.Name
		nodeGroupStatusRequest.State = string(nodegroup.State)
		nodeGroupStatusRequest.Type = string(nodegroup.Type)
		nodeGroupStatusRequest.Reason = nodegroup.Reason
		nodeGroupStatusRequest.Message = nodegroup.Message

		nodeGroupStatusRequest.Nodes = make([]*pb.NodeStatusRequest, 0, len(nodegroup.Nodes))
		for _, node := range nodegroup.Nodes {
			var nodeStatusRequest pb.NodeStatusRequest
			nodeStatusRequest.CreationTime = node.CreationTime.Format(time.RFC3339)
			nodeStatusRequest.InstanceIMI = node.InstanceIMI
			nodeStatusRequest.IpAddress = node.IpAddress
			nodeStatusRequest.KubeProxyVersion = node.KubeProxyVersion
			nodeStatusRequest.KubeletVersion = node.KubeletVersion
			nodeStatusRequest.LastUpdate = node.LastUpdate.Format(time.RFC3339)
			nodeStatusRequest.Message = node.Message
			nodeStatusRequest.Name = node.Name
			nodeStatusRequest.Reason = node.Reason
			nodeStatusRequest.State = string(node.State)
			nodeStatusRequest.DnsName = node.DNSName
			nodeStatusRequest.WekaStorage = &pb.WekaStorageStatus{
				ClientId:     node.WekaStorageStatus.ClientId,
				Status:       node.WekaStorageStatus.Status,
				CustomStatus: node.WekaStorageStatus.CustomStatus,
				Message:      node.WekaStorageStatus.Message,
			}

			nodeGroupStatusRequest.Nodes = append(nodeGroupStatusRequest.Nodes, &nodeStatusRequest)
		}

		clusterStatusRequest.Nodegroups = append(clusterStatusRequest.Nodegroups, &nodeGroupStatusRequest)
	}

	for _, ilb := range clusterStatus.ILBS {
		var ilbStatusRequest pb.IlbStatusRequest
		ilbStatusRequest.VipID = int32(ilb.VipID)
		ilbStatusRequest.Vip = ilb.Vip
		ilbStatusRequest.State = string(ilb.State)
		ilbStatusRequest.PoolID = int32(ilb.PoolID)
		ilbStatusRequest.Name = ilb.Name
		ilbStatusRequest.Message = ilb.Message
		ilbStatusRequest.Conditions = &pb.IlbConditionsStatusRequest{
			PoolCreated:   ilb.Conditions.PoolCreated,
			VipCreated:    ilb.Conditions.VIPCreated,
			VipPoolLinked: ilb.Conditions.VIPPoolLinked,
		}

		clusterStatusRequest.Ilbs = append(clusterStatusRequest.Ilbs, &ilbStatusRequest)
	}

	for _, addon := range clusterStatus.Addons {
		var addonStatusRequest pb.AddonStatusRequest
		addonStatusRequest.Artifact = addon.Artifact
		addonStatusRequest.LastUpdate = addon.LastUpdate.Format(time.RFC3339)
		addonStatusRequest.Message = addon.Message
		addonStatusRequest.Name = addon.Name
		addonStatusRequest.Reason = addon.Reason
		addonStatusRequest.State = string(addon.State)

		clusterStatusRequest.Addons = append(clusterStatusRequest.Addons, &addonStatusRequest)
	}

	for _, storage := range clusterStatus.Storage {
		var storageStatusRequest pb.StorageStatusRequest
		storageStatusRequest.NamespaceName = storage.NamespaceName
		storageStatusRequest.NamespaceCreated = storage.NamespaceCreated
		storageStatusRequest.NamespaceState = string(storage.NamespaceState)
		storageStatusRequest.StorageState = string(storage.State)
		storageStatusRequest.Reason = storage.Reason
		storageStatusRequest.Message = storage.Message
		storageStatusRequest.Provider = string(storage.Provider)
		storageStatusRequest.Size = storage.Size
		storageStatusRequest.SecretCreated = storage.SecretCreated

		clusterStatusRequest.Storages = append(clusterStatusRequest.Storages, &storageStatusRequest)
	}

	for _, firewall := range clusterStatus.Firewall {
		var firewallStatusRequest pb.FirewallStatusRequest
		firewallStatusRequest.Firewallstate = string(firewall.Firewallrulestatus.State)
		firewallStatusRequest.Destinationip = firewall.DestinationIp
		firewallStatusRequest.Port = int32(firewall.Port)
		firewallStatusRequest.Protocol = firewall.Protocol
		firewallStatusRequest.Sourceips = firewall.SourceIps

		clusterStatusRequest.Firewall = append(clusterStatusRequest.Firewall, &firewallStatusRequest)
	}
	if _, err := c.GRPCClient.PutClusterStatusReconciler(ctx, &pb.UpdateClusterStatusRequest{
		Uuid:          clusterUUID,
		ClusterStatus: &clusterStatusRequest,
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) AppliedRevision(ctx context.Context, revID int) error {
	if _, err := c.GRPCClient.PutClusterChangeAppliedReconciler(ctx, &pb.UpdateClusterChangeAppliedRequest{
		ClusterrevId:  int32(revID),
		ChangeApplied: true,
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteCluster(ctx context.Context, clusterUUID string) error {
	if _, err := c.GRPCClient.DeleteClusterReconciler(ctx, &pb.ClusterDeletionRequest{
		Uuid: clusterUUID,
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateClusterSecret(ctx context.Context, clusterUUID, caCert, caKey, etcdCaCert, etcdCaKey, saPrivate, saPublic, controlplaneRegistrationCmd, workerRegistrationCmd, etcdEncryptionConfigs string) error {
	if _, err := c.GRPCClient.PutClusterCertsReconciler(ctx, &pb.UpdateClusterCertsRequest{
		Uuid:               clusterUUID,
		CaCert:             caCert,
		CaKey:              caKey,
		EtcdCaCert:         etcdCaCert,
		EtcdCaKey:          etcdCaKey,
		SaKey:              saPrivate,
		SaPub:              saPublic,
		CpRegistrationCmd:  controlplaneRegistrationCmd,
		WrkRegistrationCmd: workerRegistrationCmd,
		EtcdCaRotationKey:  etcdEncryptionConfigs,
	}); err != nil {
		return err
	}

	return nil
}
