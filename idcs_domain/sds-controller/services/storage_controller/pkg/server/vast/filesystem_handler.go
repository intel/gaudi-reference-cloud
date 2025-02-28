// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package vast

import (
	"context"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1/vast"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	vast_backend "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FilesystemHandler struct {
	Backends map[string]backend.Interface
}

// CreateFilesystem implements vast.FilesystemServiceServer.
func (h *FilesystemHandler) CreateFilesystem(ctx context.Context, r *vast.CreateFilesystemRequest) (*vast.CreateFilesystemResponse, error) {
	b := h.Backends[r.GetNamespaceId().GetClusterId().GetUuid()]

	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	vb, ok := b.(vast_backend.Interface)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support Vast operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support Vast operations")
	}

	log.Info().Ctx(ctx).Str("name", r.GetName()).Uint64("quota", r.GetTotalBytes()).Msg("Creating filesystem")

	fs, err := vb.CreateView(ctx, vast_backend.CreateViewOpts{
		NamespaceID: r.GetNamespaceId().GetId(),
		Name:        r.GetName(),
		Path:        r.GetPath(),
		Protocols:   intoProtocols(r.GetProtocols()),
		Quota:       r.GetTotalBytes(),
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", fs.ID).Msg("Created filesystem")

	return &vast.CreateFilesystemResponse{
		Filesystem: intoFilesystem(r.GetNamespaceId(), fs),
	}, nil
}

// DeleteFilesystem implements vast.FilesystemServiceServer.
func (h *FilesystemHandler) DeleteFilesystem(ctx context.Context, r *vast.DeleteFilesystemRequest) (*vast.DeleteFilesystemResponse, error) {
	b := h.Backends[r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()]

	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	vb, ok := b.(vast_backend.Interface)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support Vast operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support Vast operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetFilesystemId().GetId()).Msg("Deleting filesystem")

	err := vb.DeleteView(ctx, vast_backend.DeleteViewOpts{
		NamespaceID: r.GetFilesystemId().GetNamespaceId().GetId(),
		ViewID:      r.GetFilesystemId().GetId(),
	})

	if err != nil {
		return nil, err
	}

	return &vast.DeleteFilesystemResponse{}, nil
}

// GetFilesystem implements vast.FilesystemServiceServer.
func (h *FilesystemHandler) GetFilesystem(ctx context.Context, r *vast.GetFilesystemRequest) (*vast.GetFilesystemResponse, error) {
	b := h.Backends[r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()]

	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	vb, ok := b.(vast_backend.Interface)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support Vast operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support Vast operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetFilesystemId().GetId()).Msg("Getting filesystem")

	fs, err := vb.GetView(ctx, vast_backend.GetViewOpts{
		NamespaceID: r.GetFilesystemId().GetNamespaceId().GetId(),
		ViewID:      r.GetFilesystemId().GetId(),
	})

	if err != nil {
		return nil, err
	}
	return &vast.GetFilesystemResponse{
		Filesystem: intoFilesystem(r.GetFilesystemId().GetNamespaceId(), fs),
	}, nil
}

// ListFilesystems implements vast.FilesystemServiceServer.
func (h *FilesystemHandler) ListFilesystems(ctx context.Context, r *vast.ListFilesystemsRequest) (*vast.ListFilesystemsResponse, error) {
	b := h.Backends[r.GetNamespaceId().GetClusterId().GetUuid()]

	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	vb, ok := b.(vast_backend.Interface)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support Vast operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support Vast operations")
	}

	log.Info().Ctx(ctx).Str("namespace_id", r.GetNamespaceId().GetId()).Msg("List filesystem")

	fss, err := vb.ListViews(ctx, vast_backend.ListViewsOpts{
		NamespaceID: r.GetNamespaceId().GetId(),
		Names:       r.GetFilter().GetNames(),
	})

	if err != nil {
		return nil, err
	}

	filesystems := make([]*vast.Filesystem, 0)

	for _, fs := range fss {
		filesystems = append(filesystems, intoFilesystem(r.GetNamespaceId(), fs))
	}

	return &vast.ListFilesystemsResponse{
		Filesystems: filesystems,
	}, nil
}

// UpdateFilesystem implements vast.FilesystemServiceServer.
func (h *FilesystemHandler) UpdateFilesystem(ctx context.Context, r *vast.UpdateFilesystemRequest) (*vast.UpdateFilesystemResponse, error) {
	b := h.Backends[r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()]

	if b == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	vb, ok := b.(vast_backend.Interface)
	if !ok {
		log.Info().Ctx(ctx).Str("uuid", r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid does not support Vast operations")
		return nil, status.Error(codes.FailedPrecondition, "cluster does not support Vast operations")
	}

	log.Info().Ctx(ctx).Str("id", r.GetFilesystemId().GetId()).Msg("Deleting filesystem")

	fs, err := vb.UpdateView(ctx, vast_backend.UpdateViewOpts{
		NamespaceID: r.GetFilesystemId().GetNamespaceId().GetId(),
		ViewID:      r.GetFilesystemId().GetId(),
		Name:        r.NewName,
		Quota:       r.NewTotalBytes,
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", r.GetFilesystemId().GetId()).Msg("Deleted filesystem")

	return &vast.UpdateFilesystemResponse{
		Filesystem: intoFilesystem(r.GetFilesystemId().GetNamespaceId(), fs),
	}, nil
}

func intoFilesystem(namespaceID *v1.NamespaceIdentifier, view *vast_backend.View) *vast.Filesystem {
	if view == nil || namespaceID == nil {
		return nil
	}

	var protocols []vast.Filesystem_Protocol

	for _, protocol := range view.Protocols {
		if protocol == vast_backend.NFSV3 {
			protocols = append(protocols, vast.Filesystem_PROTOCOL_NFS_V3)
		} else if protocol == vast_backend.NFSV4 {
			protocols = append(protocols, vast.Filesystem_PROTOCOL_NFS_V4)
		} else if protocol == vast_backend.SMB {
			protocols = append(protocols, vast.Filesystem_PROTOCOL_SMB)
		} else {
			protocols = append(protocols, vast.Filesystem_PROTOCOL_UNSPECIFIED)
		}
	}

	return &vast.Filesystem{
		Id: &vast.FilesystemIdentifier{
			NamespaceId: namespaceID,
			Id:          view.ID,
		},
		Name:      view.Name,
		Path:      view.Path,
		Protocols: protocols,
		Capacity: &vast.Filesystem_Capacity{
			TotalBytes:     view.TotalBytes,
			AvailableBytes: view.AvailableBytes,
		},
	}
}

func intoProtocols(protocols []vast.Filesystem_Protocol) []vast_backend.Protocol {
	var resp []vast_backend.Protocol

	for _, protocol := range protocols {
		resp = append(resp, vast_backend.Protocol(protocol))
	}

	return resp
}
