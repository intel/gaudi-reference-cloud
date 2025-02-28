// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package vast

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
	"github.com/rs/zerolog/log"
)

// This is used to apply quota to the whole tenant namespace
var RootPath = "/"

func (b *Backend) CreateNamespace(ctx context.Context, opts backend.CreateNamespaceOpts) (*backend.Namespace, error) {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	var ipRanges *[][]string = nil
	if opts.IPRanges != nil {
		ipRanges = &opts.IPRanges
	}

	resp, err := b.client.TenantsCreateWithResponse(ctx, client.Tenant{
		Name:           opts.Name,
		ClientIpRanges: ipRanges,
	}, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API")
		return nil, err
	}

	if resp.StatusCode() != 201 || resp.JSON201 == nil || resp.JSON201.Id == nil {
		return nil, backend.ResponseAsErr("could not create namespace", resp.StatusCode(), resp.Body)
	}

	quota, err := b.client.QuotasCreateWithResponse(ctx, client.Quota{
		Name:      opts.Name,
		HardLimit: &opts.Quota,
		TenantId:  resp.JSON201.Id,
		Path:      RootPath,
	}, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API")
		return nil, err
	}

	if quota.StatusCode() != 201 || quota.JSON201 == nil {
		// Rollback NS
		err = b.DeleteNamespace(ctx, backend.DeleteNamespaceOpts{NamespaceID: strconv.Itoa(*resp.JSON201.Id)})
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Could not rollback ns due to quota not set")
		}
		return nil, backend.ResponseAsErr("could not create quota", resp.StatusCode(), resp.Body)
	}

	return intoNamespace(*resp.JSON201, opts.Quota)
}

func (b *Backend) DeleteNamespace(ctx context.Context, opts backend.DeleteNamespaceOpts) error {
	if slices.Contains(b.config.VastConfig.ProtectedOrgIds, opts.NamespaceID) {
		log.Error().Ctx(ctx).Str("namespace_id", opts.NamespaceID).Msg("Attempt to delete protected org")
		return errors.New("attempt to delete protected org")
	}

	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	id, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("ID has invalid value")
		return err
	}

	ns, err := b.GetNamespace(ctx, backend.GetNamespaceOpts{NamespaceID: opts.NamespaceID})
	if err != nil || ns == nil {
		log.Error().Ctx(ctx).Err(err).Msg("could not get namespace")
		return err
	}

	quotas, err := b.listQuotas(ctx, ListQuotasOpts{
		NamespaceID: opts.NamespaceID,
	}, token)
	if err == nil && quotas != nil {
		for _, quota := range quotas {
			var resp *client.QuotasDeleteResponse
			resp, err = b.client.QuotasDeleteWithResponse(ctx, quota.ID, client.RequestEditorFn(token.Intercept))
			// We will not be able to delete this ns with quota intact
			if err != nil {
				return err
			} else if resp.StatusCode() != 204 {
				return backend.ResponseAsErr("could not delete namespace", resp.StatusCode(), resp.Body)
			}
		}
	}

	resp, err := b.client.TenantsDeleteWithResponse(ctx, id, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API")
		return err
	}

	if resp.StatusCode() != 204 {
		return backend.ResponseAsErr("could not delete namespace", resp.StatusCode(), resp.Body)
	}

	return nil
}

func (b *Backend) GetNamespace(ctx context.Context, opts backend.GetNamespaceOpts) (*backend.Namespace, error) {
	if slices.Contains(b.config.VastConfig.ProtectedOrgIds, opts.NamespaceID) {
		log.Error().Ctx(ctx).Str("namespace_id", opts.NamespaceID).Msg("Attempt to get protected org")
		return nil, errors.New("attempt to get protected org")
	}

	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	id, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("ID has invalid value")
		return nil, err
	}

	resp, err := b.client.TenantsReadWithResponse(ctx, id, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, backend.ResponseAsErr("could not get cluster", resp.StatusCode(), resp.Body)
	}

	quotas, err := b.listQuotas(ctx, ListQuotasOpts{
		NamespaceID: opts.NamespaceID,
		Paths:       []string{RootPath},
	}, token)
	if err != nil {
		log.Warn().Ctx(ctx).Str("namespace_id", opts.NamespaceID).Err(err).Msg("could not get quotas")
	}

	var q uint64
	if len(quotas) > 0 {
		q = quotas[0].Quota
	}

	return intoNamespace(*resp.JSON200, q)
}

func (b *Backend) ListNamespaces(ctx context.Context, opts backend.ListNamespacesOpts) ([]*backend.Namespace, error) {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	return b.getNamespaces(ctx, opts.Names, token)
}

func (b *Backend) UpdateNamespace(ctx context.Context, opts backend.UpdateNamespaceOpts) (*backend.Namespace, error) {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	ns, err := b.GetNamespace(ctx, backend.GetNamespaceOpts{
		NamespaceID: opts.NamespaceID,
	})
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Could not find ns to update")
		return nil, err
	}

	quotas, err := b.listQuotas(ctx, ListQuotasOpts{
		NamespaceID: opts.NamespaceID,
		Paths:       []string{RootPath},
	}, token)
	if err == nil && len(quotas) > 0 {
		resp, err := b.client.QuotasPartialUpdateWithResponse(ctx, quotas[0].ID, client.QuotaUpdate{
			HardLimit: &opts.Quota,
		}, client.RequestEditorFn(token.Intercept))
		if err != nil {
			return nil, fmt.Errorf("could not update quota for tenant %s(%s): %w", ns.Name, opts.NamespaceID, err)
		}
		if resp.StatusCode() != 200 {
			return nil, backend.ResponseAsErr("could not update quota", resp.StatusCode(), resp.Body)
		}
		log.Info().Ctx(ctx).Str("namespace_id", opts.NamespaceID).Msg("Quota updated for tenant")
		ns.QuotaTotal = opts.Quota
	}

	if opts.IPRanges != nil {
		id, err := strconv.Atoi(opts.NamespaceID)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("ID has invalid value")
			return nil, err
		}

		resp, err := b.client.TenantsPartialUpdateWithResponse(ctx, id, client.TenantUpdate{
			ClientIpRanges: &opts.IPRanges,
		}, client.RequestEditorFn(token.Intercept))

		if resp != nil && resp.StatusCode() != 200 {
			return nil, backend.ResponseAsErr("could not update ip ranges", resp.StatusCode(), resp.Body)
		} else {
			ns.IPRanges = opts.IPRanges
		}
	}

	return ns, nil
}

func (b *Backend) getNamespaces(ctx context.Context, namesToFilter []string, token *securityprovider.SecurityProviderBearerToken) ([]*backend.Namespace, error) {
	params := &client.TenantsListParams{}
	if len(namesToFilter) > 0 {
		if len(namesToFilter) == 1 {
			log.Info().Ctx(ctx).Str("name", namesToFilter[0]).Msg("Issuing TenantsList API with name param")
			params.Name = &namesToFilter[0]
		} else {
			allNames := strings.Join(namesToFilter, ",")
			log.Info().Ctx(ctx).Str("name__in", allNames).Msg("Issuing TenantsList API with name__in param")
			params.NameIn = &allNames
		}
	}

	resp, err := b.client.TenantsListWithResponse(ctx, params, client.RequestEditorFn(token.Intercept))
	if err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, backend.ResponseAsErr("could not get namespaces", resp.StatusCode(), resp.Body)
	}

	namespaces := make([]*backend.Namespace, 0)

	for _, tenant := range *resp.JSON200 {
		namespaceId := strconv.Itoa(*tenant.Id)

		quotas, err := b.listQuotas(ctx, ListQuotasOpts{
			NamespaceID: namespaceId,
			Paths:       []string{RootPath},
		}, token)
		if err != nil {
			log.Warn().Ctx(ctx).Str("namespace_id", namespaceId).Err(err).Msg("could not get quota")
		}

		var q uint64
		if quotas != nil && len(quotas) > 0 {
			q = quotas[0].Quota
		}

		ns, err := intoNamespace(tenant, q)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Could not parse namespace from org")
			continue
		}

		if slices.Contains(b.config.VastConfig.ProtectedOrgIds, ns.ID) ||
			(len(namesToFilter) > 0 && !slices.Contains(namesToFilter, ns.Name)) {
			continue
		}

		namespaces = append(namespaces, ns)
		log.Info().Ctx(ctx).Str("Tenant name", ns.Name).Msg("Tenant namespace name")
	}

	return namespaces, nil
}
