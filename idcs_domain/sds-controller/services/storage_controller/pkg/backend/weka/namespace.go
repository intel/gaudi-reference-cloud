// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"
	"errors"
	"slices"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
	"github.com/rs/zerolog/log"
)

func (b *Backend) CreateNamespace(ctx context.Context, opts backend.CreateNamespaceOpts) (*backend.Namespace, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	ssdQuota := uint64(0)

	resp, err := b.client.CreateOrganizationWithResponse(ctx, v4.CreateOrganizationJSONRequestBody{
		Name:       opts.Name,
		Username:   opts.AdminName,
		Password:   opts.AdminPassword,
		SsdQuota:   &ssdQuota,
		TotalQuota: &opts.Quota,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not create namespace", resp.StatusCode(), resp.Body)
	}

	return intoNamespace(*resp.JSON200.Data)
}

func (b *Backend) DeleteNamespace(ctx context.Context, opts backend.DeleteNamespaceOpts) error {
	if slices.Contains(b.config.WekaConfig.ProtectedOrgIds, opts.NamespaceID) {
		log.Error().Ctx(ctx).Str("namespace_id", opts.NamespaceID).Msg("Attempt to delete protected org")
		return errors.New("attempt to delete protected org")
	}

	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	resp, err := b.client.DeleteOrganizationWithResponse(ctx, opts.NamespaceID, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return err
	}

	if resp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not delete namespace", resp.StatusCode(), resp.Body)
	}

	return nil
}

func (b *Backend) GetNamespace(ctx context.Context, opts backend.GetNamespaceOpts) (*backend.Namespace, error) {
	if slices.Contains(b.config.WekaConfig.ProtectedOrgIds, opts.NamespaceID) {
		log.Error().Ctx(ctx).Str("namespace_id", opts.NamespaceID).Msg("Attempt to get protected org")
		return nil, errors.New("attempt to get protected org")
	}

	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	resp, err := b.client.GetOrganizationWithResponse(ctx, opts.NamespaceID, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get cluster", resp.StatusCode(), resp.Body)
	}

	return intoNamespace(*resp.JSON200.Data)
}

func (b *Backend) ListNamespaces(ctx context.Context, opts backend.ListNamespacesOpts) ([]*backend.Namespace, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	return b.getNamespaces(ctx, opts.Names, token)
}

func (b *Backend) UpdateNamespace(ctx context.Context, opts backend.UpdateNamespaceOpts) (*backend.Namespace, error) {
	if slices.Contains(b.config.WekaConfig.ProtectedOrgIds, opts.NamespaceID) {
		log.Error().Ctx(ctx).Str("namespace_id", opts.NamespaceID).Msg("Attempt to update protected org")
		return nil, errors.New("attempt to update protected org")
	}

	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	if opts.Quota == 0 {
		return nil, errors.New("new quota is not set")
	}

	resp, err := b.client.SetOrganizationLimitWithResponse(ctx, opts.NamespaceID, v4.SetOrganizationLimitJSONRequestBody{
		TotalQuota: &opts.Quota,
	}, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not update namespace", resp.StatusCode(), resp.Body)
	}

	return intoNamespace(*resp.JSON200.Data)
}

func (b *Backend) getNamespaces(ctx context.Context, namesToFilter []string, token *securityprovider.SecurityProviderBearerToken) ([]*backend.Namespace, error) {
	resp, err := b.client.GetOrganizationsWithResponse(ctx, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		return nil, err
	}

	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get namespaces, status", resp.StatusCode(), resp.Body)
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	b.orgNames = make(map[string]string)

	namespaces := make([]*backend.Namespace, 0)

	for _, org := range *resp.JSON200.Data {
		ns, err := intoNamespace(org)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Could not parse namespace from org")
			continue
		}

		b.orgNames[ns.ID] = ns.Name

		if slices.Contains(b.config.WekaConfig.ProtectedOrgIds, ns.ID) ||
			(len(namesToFilter) > 0 && !slices.Contains(namesToFilter, ns.Name)) {
			continue
		}

		namespaces = append(namespaces, ns)
	}

	return namespaces, nil
}
