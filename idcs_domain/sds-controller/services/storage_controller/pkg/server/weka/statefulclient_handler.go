// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package weka

import (
	"context"

	"google.golang.org/grpc/metadata"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	weka "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1/weka"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StatefulClientHandler struct {
	Backends map[string]backend.Interface
}

// CreateStatefulClient implements weka.StatefulClientServiceServer.
func (h *StatefulClientHandler) CreateStatefulClient(ctx context.Context, r *weka.CreateStatefulClientRequest) (*weka.CreateStatefulClientResponse, error) {
	b := h.Backends[r.GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	c, ok := b.(backend.StatefulClientOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support stateful client ops")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support stateful client ops")
	}

	log.Info().Ctx(ctx).Str("name", r.GetName()).Str("ip", r.GetIp()).Msg("Creating StatefulClient")

	createStatefulClientOpts := backend.CreateStatefulClientOpts{
		Name: r.GetName(),
		Ip:   r.GetIp(),
	}

	sc, err := c.CreateStatefulClient(ctx, createStatefulClientOpts)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Str("uuid", r.GetClusterId().GetUuid()).Msg("Could not create StatefulClient")
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", sc.ID).Msg("Created StatefulClient")

	return &weka.CreateStatefulClientResponse{
		StatefulClient: intoStatefulClient(r.GetClusterId(), sc),
	}, nil
}

// DeleteStatefulClient implements weka.StatefulClientServiceServer.
func (h *StatefulClientHandler) DeleteStatefulClient(ctx context.Context, r *weka.DeleteStatefulClientRequest) (*weka.DeleteStatefulClientResponse, error) {
	b := h.Backends[r.GetStatefulClientId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetStatefulClientId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	cl, ok := b.(backend.StatefulClientOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetStatefulClientId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support stateful client ops")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support stateful client ops")
	}

	log.Info().Ctx(ctx).Str("id", r.GetStatefulClientId().GetId()).Msg("Deleting StatefulClient")

	// Create a new background context for the async operation
	md, _ := metadata.FromIncomingContext(ctx)
	asyncCtx := metadata.NewOutgoingContext(context.Background(), md)
	// Execute b.DeleteStatefulClient asynchronously
	go func(c context.Context) {
		err := cl.DeleteStatefulClient(c, backend.DeleteStatefulClientOpts{
			StatefulClientID: r.GetStatefulClientId().GetId(),
		})
		if err != nil {
			log.Warn().Ctx(c).Err(err).Str("id", r.GetStatefulClientId().GetId()).Msg("Could not delete Statefulclient")
			return
		}

		log.Info().Ctx(c).Str("id", r.GetStatefulClientId().GetId()).Msg("Deleted StatefulClient")
	}(asyncCtx)
	// Return immediately without waiting for the result
	return &weka.DeleteStatefulClientResponse{}, nil
}

// GetStatefulClient implements weka.StatefulClientServiceServer.
func (h *StatefulClientHandler) GetStatefulClient(ctx context.Context, r *weka.GetStatefulClientRequest) (*weka.GetStatefulClientResponse, error) {
	b := h.Backends[r.GetStatefulClientId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetStatefulClientId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	c, ok := b.(backend.StatefulClientOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetStatefulClientId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support stateful client ops")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support stateful client ops")
	}

	log.Info().Ctx(ctx).Str("id", r.GetStatefulClientId().GetId()).Msg("Getting StatefulClient")

	sc, err := c.GetStatefulClient(ctx, backend.GetStatefulClientOpts{
		StatefulClientID: r.GetStatefulClientId().GetId(),
	})

	if err != nil {
		return nil, err
	}

	return &weka.GetStatefulClientResponse{
		StatefulClient: intoStatefulClient(r.GetStatefulClientId().GetClusterId(), sc),
	}, nil
}

// ListStatefulClients implements weka.StatefulClientServiceServer.
func (h *StatefulClientHandler) ListStatefulClients(ctx context.Context, r *weka.ListStatefulClientsRequest) (*weka.ListStatefulClientsResponse, error) {
	b := h.Backends[r.GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	c, ok := b.(backend.StatefulClientOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support stateful client ops")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support stateful client ops")
	}

	log.Info().Ctx(ctx).Strs("names", r.GetFilter().GetNames()).Msg("Listing StatefulClients")

	scs, err := c.ListStatefulClients(ctx, backend.ListStatefulClientsOpts{
		Names: r.GetFilter().GetNames(),
	})

	if err != nil {
		return nil, err
	}

	StatefulClients := make([]*weka.StatefulClient, 0)

	for _, sc := range scs {
		StatefulClients = append(StatefulClients, intoStatefulClient(r.GetClusterId(), sc))
	}

	return &weka.ListStatefulClientsResponse{
		StatefulClients: StatefulClients,
	}, nil
}

func intoStatefulClient(clusterID *v1.ClusterIdentifier, sc *backend.StatefulClient) *weka.StatefulClient {

	var status weka.StatefulClient_PredefinedStatus

	switch sc.Status {
	case string(backend.ContainerStatusUP):
		status.PredefinedStatus = weka.StatefulClient_STATUS_UP
	case string(backend.ContainerStatusDegraded):
		status.PredefinedStatus = weka.StatefulClient_STATUS_DEGRADED_UNSPECIFIED
	case string(backend.ContainerStatusDown):
		status.PredefinedStatus = weka.StatefulClient_STATUS_DOWN
	default:
		return &weka.StatefulClient{
			Id: &weka.StatefulClientIdentifier{
				ClusterId: clusterID,
				Id:        sc.ID,
			},
			Name:   sc.Name,
			Status: &weka.StatefulClient_CustomStatus{CustomStatus: sc.Status},
		}
	}

	return &weka.StatefulClient{
		Id: &weka.StatefulClientIdentifier{
			ClusterId: clusterID,
			Id:        sc.ID,
		},
		Name:   sc.Name,
		Status: &status,
	}
}
