// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package vast

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
	"github.com/rs/zerolog/log"
)

func (b *Backend) CreateView(ctx context.Context, opts CreateViewOpts) (*View, error) {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	authSource := client.RPC

	// Even in case of SMB
	flavor := "NFS"

	// Enforce NFS TLS by default when creating view policy
	nfsEnforceTLS := true

	tenantId, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("namespace ID has invalid value")
		return nil, err
	}

	protocols := protocolsToStrings(opts.Protocols)

	success := false
	// Creating view policy
	vp, err := b.client.ViewpoliciesCreateWithResponse(ctx, client.ViewPolicy{
		TenantId:      &tenantId,
		Name:          opts.Name,
		AuthSource:    &authSource,
		Flavor:        &flavor,
		NfsEnforceTls: &nfsEnforceTLS,
	}, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling VAST API (ViewpoliciesCreate)")
		return nil, err
	}

	if vp == nil || vp.JSON201 == nil || vp.JSON201.Id == nil {
		return nil, backend.ResponseAsErr("could not create view due to policy error", vp.StatusCode(), vp.Body)
	}

	defer func() {
		if !success {
			log.Error().Ctx(ctx).Msg("Cleanup view policy")
			b.client.ViewpoliciesDeleteWithResponse(ctx, *vp.JSON201.Id, client.RequestEditorFn(token.Intercept))
		}
	}()

	// Creating view

	createDir := true

	resp, err := b.client.ViewsCreateWithResponse(ctx, client.CreateView{
		Name:      &opts.Name,
		Path:      opts.Path,
		TenantId:  &tenantId,
		Protocols: &protocols,
		PolicyId:  *vp.JSON201.Id,
		CreateDir: &createDir,
	}, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API (ViewsCreate)")

		return nil, err
	}

	if resp.StatusCode() != 201 || resp.JSON201 == nil {
		return nil, backend.ResponseAsErr("could not create view", resp.StatusCode(), resp.Body)
	}
	view := resp.JSON201

	defer func() {
		if !success {
			log.Error().Ctx(ctx).Msg("Cleanup view")
			b.client.ViewsDeleteWithResponse(ctx, *resp.JSON201.Id, client.RequestEditorFn(token.Intercept))
		}
	}()

	// Creating Quota for the view
	if opts.Quota != 0 {
		var quota *client.Quota
		quota, err = b.createQuotaForView(ctx,
			token,
			opts.Name,
			tenantId,
			&opts.Quota,
			opts.Path,
		)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API (QuotasCreate)")
			return nil, err
		}

		defer func() {
			if !success {
				log.Error().Ctx(ctx).Msg("Cleanup quota")
				b.client.QuotasDeleteWithResponse(ctx, *quota.Id, client.RequestEditorFn(token.Intercept))
			}
		}()
	}

	success = true
	return intoView(*view, opts.Quota)
}

func (b *Backend) createQuotaForView(ctx context.Context, token *securityprovider.SecurityProviderBearerToken, name string, tenantId int, hardLimit *uint64, path string) (*client.Quota, error) {
	quota, err := b.client.QuotasCreateWithResponse(ctx, client.Quota{
		Name:      name,
		HardLimit: hardLimit,
		TenantId:  &tenantId,
		Path:      path,
	}, client.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API (QuotasCreate)")
		return nil, err
	}

	if quota.StatusCode() != 201 || quota.JSON201 == nil {
		return nil, backend.ResponseAsErr("could not create quota for view", quota.StatusCode(), quota.Body)
	} else {
		return quota.JSON201, nil
	}
}

func (b *Backend) DeleteView(ctx context.Context, opts DeleteViewOpts) error {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	id, err := strconv.Atoi(opts.ViewID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("ID has invalid value")
		return err
	}

	view, err := b.GetView(ctx, GetViewOpts{
		NamespaceID: opts.NamespaceID,
		ViewID:      opts.ViewID,
	})

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Could not get view by ID")
		return err
	}

	tenantId, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("namespace ID has invalid value")
		return err
	}

	// ToDo should we remove all the quotas by path prefix? Dangling quotas prevent
	// new quotas with the same path to be created.
	quotas, err := b.listQuotas(ctx, ListQuotasOpts{
		NamespaceID: opts.NamespaceID,
		Paths:       []string{view.Path},
	}, token)

	if err == nil && len(quotas) != 0 {
		for _, quota := range quotas {
			var resp *client.QuotasDeleteResponse
			resp, err = b.client.QuotasDeleteWithResponse(ctx, quota.ID, client.RequestEditorFn(token.Intercept))
			if err != nil {
				return err
			} else if resp.StatusCode() != 204 {
				return backend.ResponseAsErr("could not delete quota", resp.StatusCode(), resp.Body)
			}
		}
	} else {
		log.Info().Ctx(ctx).Err(err).Msgf("no quotas found for tenant: %s and path: %s", opts.NamespaceID, view.Path)
	}

	resp, err := b.client.ViewsDeleteWithResponse(ctx, id, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API")
		return err
	}

	if resp.StatusCode() != 204 {
		return backend.ResponseAsErr("could not delete view", resp.StatusCode(), resp.Body)
	}

	// Try to delete the folder.
	folder, err := b.client.FoldersDeleteFolderWithResponse(ctx, client.FolderDelete{
		TenantId: &tenantId,
		Path:     view.Path,
	}, client.RequestEditorFn(token.Intercept))

	if folder.StatusCode() != 200 {
		log.Info().Ctx(ctx).
			Int("status", folder.StatusCode()).
			Str("body", string(folder.Body)).
			Msgf("can't delete folder %s for namespace %s", view.Path, opts.NamespaceID)
		// ignore the error and continue: it could be that there was no folder
	}

	// Policy cleanup
	vpDelResp, err := b.client.ViewpoliciesDeleteWithResponse(ctx, view.PolicyID, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API")
		return err
	}

	if vpDelResp.StatusCode() != 204 {
		log.Error().Ctx(ctx).Str("body", string(resp.Body)).Int("statusCode", resp.StatusCode()).Msg("could not delete view policy")
		return fmt.Errorf("could not delete view policy for view: %v", view)
	}

	return nil
}

func (b *Backend) GetView(ctx context.Context, opts GetViewOpts) (*View, error) {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	id, err := strconv.Atoi(opts.ViewID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("ID has invalid value")
		return nil, err
	}

	resp, err := b.client.ViewsReadWithResponse(ctx, id, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling vast API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.TenantId == nil {
		return nil, backend.ResponseAsErr("could not get view", resp.StatusCode(), resp.Body)
	}

	tenantId, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("namespace ID has invalid value")
		return nil, err
	}

	if *resp.JSON200.TenantId != tenantId {
		return nil, fmt.Errorf("could not get view as it does not belongs to namespace %s!=%s", strconv.Itoa(*resp.JSON200.TenantId), opts.NamespaceID)
	}

	quotas, err := b.listQuotas(ctx, ListQuotasOpts{
		NamespaceID: opts.NamespaceID,
		Paths:       []string{resp.JSON200.Path},
	}, token)
	if err != nil {
		return nil, err
	}

	var totalBytes uint64 = 0
	if len(quotas) > 0 {
		totalBytes = quotas[0].Quota
	}

	return intoView(*resp.JSON200, totalBytes)
}

func (b *Backend) ListViews(ctx context.Context, opts ListViewsOpts) ([]*View, error) {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	tenantId, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("namespace ID has invalid value")
		return nil, err
	}

	return b.getViews(ctx, tenantId, opts.Names, token)
}

func (b *Backend) UpdateView(ctx context.Context, opts UpdateViewOpts) (*View, error) {
	// Parameter validation
	tenantId, err := strconv.Atoi(opts.NamespaceID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("namespace ID has invalid value")
		return nil, err
	}

	id, err := strconv.Atoi(opts.ViewID)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("ID has invalid value")
		return nil, err
	}

	// Request view and quota
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	viewResp, err := b.client.ViewsReadWithResponse(ctx, id, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling vast API")
		return nil, err
	}

	if viewResp.StatusCode() != 200 || viewResp.JSON200 == nil || viewResp.JSON200.TenantId == nil {
		return nil, backend.ResponseAsErr("could not get view", viewResp.StatusCode(), viewResp.Body)
	}

	view := viewResp.JSON200

	quotas, err := b.listQuotas(ctx, ListQuotasOpts{
		NamespaceID: opts.NamespaceID,
		Paths:       []string{view.Path},
	}, token)

	if err != nil {
		return nil, err
	}

	var totalBytes uint64 = 0
	if len(quotas) > 0 {
		totalBytes = quotas[0].Quota
	}

	if opts.Name != nil {
		// update name
		var resp *client.ViewsUpdateResponse
		resp, err = b.client.ViewsUpdateWithResponse(ctx, id, client.View{
			Name:     opts.Name,
			TenantId: &tenantId,
		}, client.RequestEditorFn(token.Intercept))
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API (ViewsUpdate)")
			return nil, err
		}

		if resp.JSON200 == nil {
			return nil, backend.ResponseAsErr("could not update name of the view", resp.StatusCode(), resp.Body)
		}
		view.Name = opts.Name
	}

	if opts.Quota != nil {
		if len(quotas) > 0 {
			// update existing quota
			quotaId := quotas[0].ID
			var resp *client.QuotasPartialUpdateResponse
			resp, err = b.client.QuotasPartialUpdateWithResponse(ctx, quotaId, client.QuotaUpdate{
				HardLimit: opts.Quota,
			}, client.RequestEditorFn(token.Intercept))
			if err != nil {
				log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API (QuotaUpdate)")
				return nil, err
			}

			if resp.StatusCode() != 200 {
				return nil, backend.ResponseAsErr("could not update quota for the view", resp.StatusCode(), resp.Body)
			}
		} else {
			// create a new quota
			_, err = b.createQuotaForView(ctx, token, *view.Name, tenantId, opts.Quota, view.Path)
			if err != nil {
				return nil, err
			}
		}
		totalBytes = *opts.Quota
	}

	return intoView(*view, totalBytes)
}

func (b *Backend) getViews(ctx context.Context, nsId int, namesToFilter []string, token *securityprovider.SecurityProviderBearerToken) ([]*View, error) {
	resp, err := b.client.ViewsListWithResponse(ctx, &client.ViewsListParams{}, client.RequestEditorFn(token.Intercept))

	if err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, backend.ResponseAsErr("could not get views", resp.StatusCode(), resp.Body)
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	views := make([]*View, 0)
	// prepare quotas request
	paths := make([]string, 0)
	quotas := make(map[string]uint64)
	for _, view := range *resp.JSON200 {
		paths = append(paths, view.Path)
	}
	quotasResp, err := b.listQuotas(ctx, ListQuotasOpts{
		NamespaceID: strconv.Itoa(nsId),
		Paths:       paths,
	}, token)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("could not issue ListQuotas request")
	}
	// index quota results
	for _, quota := range quotasResp {
		quotas[quota.Path] = quota.Quota
	}

	for _, view := range *resp.JSON200 {
		var totalBytes uint64 = 0
		if quota, ok := quotas[view.Path]; ok {
			totalBytes = quota
		}
		v, err := intoView(view, totalBytes)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Could not parse namespace from org")
			continue
		}

		if view.TenantId == nil || *view.TenantId != nsId ||
			(len(namesToFilter) > 0 && !slices.Contains(namesToFilter, v.Name)) {
			continue
		}

		views = append(views, v)
	}

	return views, nil
}

func protocolsToStrings(protocols []Protocol) []client.CreateViewProtocols {
	var resp []client.CreateViewProtocols

	for _, protocol := range protocols {
		switch protocol {
		case NFSV3:
			resp = append(resp, client.CreateViewProtocolsNFS)
		case NFSV4:
			resp = append(resp, client.CreateViewProtocolsNFS4)
		case SMB:
			resp = append(resp, client.CreateViewProtocolsSMB)
		default:
			log.Info().Int("protocol", int(protocol)).Msg("Unknown protocol from gRPC")
		}
	}

	return resp
}

func stringsToProtocols(protocols *[]client.ViewProtocols) []Protocol {
	var resp []Protocol
	if protocols == nil {
		return nil
	}

	for _, protocol := range *protocols {
		switch protocol {
		case client.NFS:
			resp = append(resp, 1)
		case client.NFS4:
			resp = append(resp, 2)
		case client.SMB:
			resp = append(resp, 3)
		case client.DATABASE, client.ENDPOINT, client.S3:
			// just ignore them
		default:
			log.Info().Str("protocol", string(protocol)).Msg("Unknown protocol from Vast")
		}
	}

	return resp
}
