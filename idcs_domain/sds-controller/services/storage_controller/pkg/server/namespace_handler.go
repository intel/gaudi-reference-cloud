// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NamespaceHandler struct {
	Backends map[string]backend.Interface
}

// CreateNamespace implements v1.NamespaceServiceServer.
func (h *NamespaceHandler) CreateNamespace(ctx context.Context, r *v1.CreateNamespaceRequest) (*v1.CreateNamespaceResponse, error) {
	b := h.Backends[r.GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	n, ok := b.(backend.NamespaceOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support namespace operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support namespace operations")
	}

	log.Info().Ctx(ctx).Str("name", r.GetName()).Uint64("quota", r.GetQuota().GetTotalBytes()).Msg("Creating namespace")

	var ranges [][]string

	for _, rng := range r.GetIpFilters() {
		if ranges == nil {
			ranges = [][]string{}
		}
		ranges = append(ranges, []string{rng.GetStart(), rng.GetEnd()})
	}

	ns, err := n.CreateNamespace(ctx, backend.CreateNamespaceOpts{
		Name:          r.GetName(),
		Quota:         r.GetQuota().GetTotalBytes(),
		IPRanges:      ranges,
		AdminName:     r.GetAdminUser().GetName(),
		AdminPassword: r.GetAdminUser().GetPassword(),
	})
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Str("uuid", r.GetClusterId().GetUuid()).Msg("Could not create namespace")
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", ns.ID).Msg("Created namespace")

	return &v1.CreateNamespaceResponse{
		Namespace: intoNamespace(r.GetClusterId(), ns),
	}, nil
}

// DeleteNamespace implements v1.NamespaceServiceServer.
func (h *NamespaceHandler) DeleteNamespace(ctx context.Context, r *v1.DeleteNamespaceRequest) (*v1.DeleteNamespaceResponse, error) {
	b := h.Backends[r.GetNamespaceId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	n, ok := b.(backend.NamespaceOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support namespace operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support namespace operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetNamespaceId().GetId()).Msg("Deleting namespace")

	err := n.DeleteNamespace(ctx, backend.DeleteNamespaceOpts{
		NamespaceID: r.GetNamespaceId().GetId(),
	})

	if err != nil {
		return nil, err
	}

	return &v1.DeleteNamespaceResponse{}, nil
}

// GetNamespace implements v1.NamespaceServiceServer.
func (h *NamespaceHandler) GetNamespace(ctx context.Context, r *v1.GetNamespaceRequest) (*v1.GetNamespaceResponse, error) {
	b := h.Backends[r.GetNamespaceId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	n, ok := b.(backend.NamespaceOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support namespace operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support namespace operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetNamespaceId().GetId()).Msg("Getting namespace")

	ns, err := n.GetNamespace(ctx, backend.GetNamespaceOpts{
		NamespaceID: r.GetNamespaceId().GetId(),
	})

	if err != nil {
		return nil, err
	}

	return &v1.GetNamespaceResponse{
		Namespace: intoNamespace(r.GetNamespaceId().GetClusterId(), ns),
	}, nil
}

// ListNamespaces implements v1.NamespaceServiceServer.
func (h *NamespaceHandler) ListNamespaces(ctx context.Context, r *v1.ListNamespacesRequest) (*v1.ListNamespacesResponse, error) {
	b := h.Backends[r.GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exist")
	}
	n, ok := b.(backend.NamespaceOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support namespace operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support namespace operations")
	}

	log.Info().Ctx(ctx).Strs("names", r.GetFilter().GetNames()).Msg("Listing namespaces")

	nss, err := n.ListNamespaces(ctx, backend.ListNamespacesOpts{
		Names: r.GetFilter().GetNames(),
	})

	if err != nil {
		return nil, err
	}

	namespaces := make([]*v1.Namespace, 0)

	for _, ns := range nss {
		namespaces = append(namespaces, intoNamespace(r.GetClusterId(), ns))
	}

	return &v1.ListNamespacesResponse{
		Namespaces: namespaces,
	}, nil
}

// UpdateNamespace implements v1.NamespaceServiceServer.
func (h *NamespaceHandler) UpdateNamespace(ctx context.Context, r *v1.UpdateNamespaceRequest) (*v1.UpdateNamespaceResponse, error) {
	b := h.Backends[r.GetNamespaceId().GetClusterId().GetUuid()]
	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}
	n, ok := b.(backend.NamespaceOps)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support namespace operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support namespace operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetNamespaceId().GetId()).Uint64("quota", r.GetQuota().GetTotalBytes()).Msg("Updating namespace")

	var ranges [][]string

	for _, rng := range r.GetIpFilters() {
		if ranges == nil {
			ranges = [][]string{}
		}
		ranges = append(ranges, []string{rng.GetStart(), rng.GetEnd()})
	}

	ns, err := n.UpdateNamespace(ctx, backend.UpdateNamespaceOpts{
		NamespaceID: r.GetNamespaceId().GetId(),
		Quota:       r.GetQuota().GetTotalBytes(),
		IPRanges:    ranges,
	})

	if err != nil {
		return nil, err
	}

	return &v1.UpdateNamespaceResponse{
		Namespace: intoNamespace(r.GetNamespaceId().GetClusterId(), ns),
	}, nil
}

func intoNamespace(clusterID *v1.ClusterIdentifier, ns *backend.Namespace) *v1.Namespace {
	if ns == nil || clusterID == nil {
		return nil
	}
	var filters []*v1.Namespace_IpFilter

	for _, filter := range ns.IPRanges {
		if len(filter) != 2 {
			log.Error().Any("filter", filter).Msg("Wrong IP Range filter pair")
			continue
		}
		filters = append(filters, &v1.Namespace_IpFilter{
			Start: filter[0],
			End:   filter[1],
		})
	}
	return &v1.Namespace{
		Id: &v1.NamespaceIdentifier{
			ClusterId: clusterID,
			Id:        ns.ID,
		},
		Name: ns.Name,
		Quota: &v1.Namespace_Quota{
			TotalBytes: ns.QuotaTotal,
		},
		IpFilters: filters,
	}
}
