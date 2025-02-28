// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"

	v1 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1"
	weka "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/intel/storagecontroller/v1/weka"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	weka_backend "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/server/helpers"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FilesystemHandler struct {
	Backends map[string]backend.Interface
}

// CreateFilesystem implements weka.FilesystemServiceServer.
func (h *FilesystemHandler) CreateFilesystem(ctx context.Context, r *weka.CreateFilesystemRequest) (*weka.CreateFilesystemResponse, error) {
	b := h.Backends[r.GetNamespaceId().GetClusterId().GetUuid()]
	wb, ok := b.(weka_backend.Interface)
	if !ok || wb == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	log.Info().Ctx(ctx).Str("name", r.GetName()).Uint64("quota", r.GetTotalBytes()).Msg("Creating filesystem")

	fs, err := wb.CreateFilesystem(ctx, weka_backend.CreateFilesystemOpts{
		NamespaceID:  r.GetNamespaceId().GetId(),
		Name:         r.GetName(),
		Quota:        r.GetTotalBytes(),
		AuthRequired: r.GetAuthRequired(),
		Encrypted:    r.GetEncrypted(),
		AuthCreds:    helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	log.Info().Ctx(ctx).Str("id", fs.ID).Msg("Created filesystem")

	return &weka.CreateFilesystemResponse{
		Filesystem: intoFilesystem(r.GetNamespaceId(), fs),
	}, nil
}

// DeleteFilesystem implements weka.FilesystemServiceServer.
func (h *FilesystemHandler) DeleteFilesystem(ctx context.Context, r *weka.DeleteFilesystemRequest) (*weka.DeleteFilesystemResponse, error) {
	b := h.Backends[r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()]
	wb, ok := b.(weka_backend.Interface)
	if !ok || wb == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	log.Info().Ctx(ctx).Str("id", r.GetFilesystemId().GetId()).Msg("Deleting filesystem")

	err := wb.DeleteFilesystem(ctx, weka_backend.DeleteFilesystemOpts{
		NamespaceID:  r.GetFilesystemId().GetNamespaceId().GetId(),
		FilesystemID: r.GetFilesystemId().GetId(),
		AuthCreds:    helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	return &weka.DeleteFilesystemResponse{}, nil
}

// GetFilesystem implements weka.FilesystemServiceServer.
func (h *FilesystemHandler) GetFilesystem(ctx context.Context, r *weka.GetFilesystemRequest) (*weka.GetFilesystemResponse, error) {
	b := h.Backends[r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()]
	wb, ok := b.(weka_backend.Interface)
	if !ok || wb == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	log.Info().Ctx(ctx).Str("id", r.GetFilesystemId().GetId()).Msg("Getting filesystem")

	fs, err := wb.GetFilesystem(ctx, weka_backend.GetFilesystemOpts{
		NamespaceID:  r.GetFilesystemId().GetNamespaceId().GetId(),
		FilesystemID: r.GetFilesystemId().GetId(),
		AuthCreds:    helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	return &weka.GetFilesystemResponse{
		Filesystem: intoFilesystem(r.GetFilesystemId().GetNamespaceId(), fs),
	}, nil
}

// ListFilesystems implements weka.FilesystemServiceServer.
func (h *FilesystemHandler) ListFilesystems(ctx context.Context, r *weka.ListFilesystemsRequest) (*weka.ListFilesystemsResponse, error) {
	b := h.Backends[r.GetNamespaceId().GetClusterId().GetUuid()]
	wb, ok := b.(weka_backend.Interface)
	if !ok || wb == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	log.Info().Ctx(ctx).Strs("names", r.GetFilter().GetNames()).Any("namespace_id", r.GetNamespaceId()).Msg("Listing filesystems")

	fss, err := wb.ListFilesystems(ctx, weka_backend.ListFilesystemsOpts{
		NamespaceID: r.GetNamespaceId().GetId(),
		Names:       r.GetFilter().GetNames(),
		AuthCreds:   helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	filesystems := make([]*weka.Filesystem, 0)

	for _, fs := range fss {
		filesystems = append(filesystems, intoFilesystem(r.GetNamespaceId(), fs))
	}

	return &weka.ListFilesystemsResponse{
		Filesystems: filesystems,
	}, nil
}

// UpdateFilesystem implements weka.FilesystemServiceServer.
func (h *FilesystemHandler) UpdateFilesystem(ctx context.Context, r *weka.UpdateFilesystemRequest) (*weka.UpdateFilesystemResponse, error) {
	b := h.Backends[r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()]
	wb, ok := b.(weka_backend.Interface)
	if !ok || wb == nil {
		log.Info().Ctx(ctx).Str("uuid", r.GetFilesystemId().GetNamespaceId().GetClusterId().GetUuid()).Msg("Cluster with provided uuid was not found")
		return nil, status.Error(codes.NotFound, "cluster does not exists")
	}

	log.Info().Ctx(ctx).Str("id", r.GetFilesystemId().GetId()).Msg("Updating filesystem")

	fs, err := wb.UpdateFilesystem(ctx, weka_backend.UpdateFilesystemOpts{
		NamespaceID:  r.GetFilesystemId().GetNamespaceId().GetId(),
		FilesystemID: r.GetFilesystemId().GetId(),
		Quota:        r.NewTotalBytes,
		Name:         r.NewName,
		AuthRequired: r.NewAuthRequired,
		AuthCreds:    helpers.IntoAuthCreds(r.GetAuthCtx()),
	})

	if err != nil {
		return nil, err
	}

	return &weka.UpdateFilesystemResponse{
		Filesystem: intoFilesystem(r.GetFilesystemId().GetNamespaceId(), fs),
	}, nil
}

func intoFilesystem(namespaceID *v1.NamespaceIdentifier, fs *weka_backend.Filesystem) *weka.Filesystem {
	if fs == nil || namespaceID == nil {
		return nil
	}

	return &weka.Filesystem{
		Id: &weka.FilesystemIdentifier{
			NamespaceId: namespaceID,
			Id:          fs.ID,
		},
		Name:         fs.Name,
		Status:       weka.Filesystem_Status(fs.FilesystemStatus), // Such conversation can be dangerous if enums description change
		IsEncrypted:  fs.Encrypted,
		AuthRequired: fs.AuthRequired,
		Capacity: &weka.Filesystem_Capacity{
			TotalBytes:     fs.TotalBytes,
			AvailableBytes: fs.AvailableBytes,
		},
		Backend: fs.BackendFQDN,
	}
}
