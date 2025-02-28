// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
	"github.com/rs/zerolog/log"
)

func (b *Backend) CreateFilesystem(ctx context.Context, opts CreateFilesystemOpts) (*Filesystem, error) {
	if opts.AuthCreds == nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return nil, err
	}

	resp, err := b.client.CreateFileSystemWithResponse(ctx, v4.CreateFileSystemJSONRequestBody{
		Name:          opts.Name,
		GroupName:     b.config.WekaConfig.TenantFsGroupName,
		TotalCapacity: opts.Quota,
		Encrypted:     &opts.Encrypted,
		AuthRequired:  &opts.AuthRequired,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not create filesystem", resp.StatusCode(), resp.Body)
	}

	return intoFilesystem(*resp.JSON200.Data, b.config.WekaConfig.BackendFQDN)
}

func (b *Backend) DeleteFilesystem(ctx context.Context, opts DeleteFilesystemOpts) error {
	if opts.AuthCreds == nil {
		return errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return err
	}

	purgeObs := true
	resp, err := b.client.DeleteFileSystemWithResponse(ctx, opts.FilesystemID, &v4.DeleteFileSystemParams{
		PurgeFromObs: &purgeObs,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return err
	}

	if resp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not delete filesystem", resp.StatusCode(), resp.Body)
	}

	// Very simple check to wait for FS deletion
	deleted := false
	for i := 0; i < b.config.WekaConfig.FileSystemDeleteWait; i++ {
		resp, err := b.client.GetFileSystemWithResponse(ctx, opts.FilesystemID, nil, v4.RequestEditorFn(token.Intercept))
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Error on delete verification")
			return err
		}
		if resp.StatusCode() == 404 {
			deleted = true
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}

	if !deleted {
		return errors.New("could not delete filesystem within timeout")
	}

	return nil
}

func (b *Backend) GetFilesystem(ctx context.Context, opts GetFilesystemOpts) (*Filesystem, error) {
	if opts.AuthCreds == nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return nil, err
	}

	resp, err := b.client.GetFileSystemWithResponse(ctx, opts.FilesystemID, nil, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get filesystem", resp.StatusCode(), resp.Body)
	}

	return intoFilesystem(*resp.JSON200.Data, b.config.WekaConfig.BackendFQDN)
}

func (b *Backend) ListFilesystems(ctx context.Context, opts ListFilesystemsOpts) ([]*Filesystem, error) {
	if opts.AuthCreds == nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return nil, err
	}

	return b.getFilesystems(ctx, opts.Names, token)
}

func (b *Backend) UpdateFilesystem(ctx context.Context, opts UpdateFilesystemOpts) (*Filesystem, error) {
	if opts.AuthCreds == nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := b.loginUser(ctx, *opts.AuthCreds, opts.NamespaceID)
	if err != nil {
		log.Info().Ctx(ctx).Err(err).Msg("User login unsuccessful")
		return nil, err
	}

	resp, err := b.client.UpdateFileSystemWithResponse(ctx, opts.FilesystemID, v4.UpdateFileSystemJSONRequestBody{
		TotalCapacity: opts.Quota,
		AuthRequired:  opts.AuthRequired,
		NewName:       opts.Name,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err

	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not update filesystem", resp.StatusCode(), resp.Body)
	}

	return intoFilesystem(*resp.JSON200.Data, b.config.WekaConfig.BackendFQDN)
}

func (b *Backend) getFilesystems(ctx context.Context, namesToFilter []string, token *securityprovider.SecurityProviderBearerToken) ([]*Filesystem, error) {
	resp, err := b.client.GetFileSystemsWithResponse(ctx, nil, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		return nil, err
	}

	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get filesystems", resp.StatusCode(), resp.Body)
	}

	filesystems := make([]*Filesystem, 0)

	for _, fs := range *resp.JSON200.Data {
		filesystem, err := intoFilesystem(fs, b.config.WekaConfig.BackendFQDN)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Could not parse filesystem")
			continue
		}

		if len(namesToFilter) > 0 && !slices.Contains(namesToFilter, filesystem.Name) {
			continue
		}

		filesystems = append(filesystems, filesystem)
	}

	return filesystems, nil
}
