// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package weka

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
	"github.com/rs/zerolog/log"
)

func (b *Backend) CreateStatefulClient(ctx context.Context, opts backend.CreateStatefulClientOpts) (*backend.StatefulClient, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	noWait := true
	addContainerResp, err := b.client.AddContainerWithResponse(ctx, v4.AddContainerJSONRequestBody{
		ContainerName: opts.Name,
		Ip:            &opts.Ip,
		NoWait:        &noWait,
	}, v4.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka Add container API")
		return nil, err
	}

	if addContainerResp.StatusCode() != 200 || addContainerResp.JSON200 == nil || addContainerResp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not add the StatefulClient", addContainerResp.StatusCode(), addContainerResp.Body)
	}

	sc, err := intoStatefulClient(*addContainerResp.JSON200.Data)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error in getting the stateful client configuration")
		return nil, err
	}
	return sc, nil
}

func (b *Backend) DeleteStatefulClient(ctx context.Context, opts backend.DeleteStatefulClientOpts) error {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return err
	}

	deactivateContainerResp, err := b.client.DeactivateContainerWithResponse(ctx, opts.StatefulClientID, v4.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling deactivate container Weka API")
		return err
	}

	if deactivateContainerResp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not deactivate StatefulClient", deactivateContainerResp.StatusCode(), deactivateContainerResp.Body)
	}

	removeContainerResp, err := b.client.RemoveContainerWithResponse(ctx, opts.StatefulClientID, v4.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling remove container Weka API")
		return err
	}

	if removeContainerResp.StatusCode() != 200 {
		return backend.ResponseAsErr("could not remove StatefulClient", removeContainerResp.StatusCode(), removeContainerResp.Body)
	}

	return nil
}

func (b *Backend) GetStatefulClient(ctx context.Context, opts backend.GetStatefulClientOpts) (*backend.StatefulClient, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	resp, err := b.client.GetSingleContainerWithResponse(ctx, opts.StatefulClientID, v4.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get stateful client", resp.StatusCode(), resp.Body)
	}

	if resp.JSON200.Data.Mode != nil && *resp.JSON200.Data.Mode != string(backend.ContainerModeClient) {
		errorMessage := "client container not found"
		return nil, backend.ResponseAsErr("could not get stateful clients", http.StatusNotFound, []byte(errorMessage))
	}

	sc, err := intoStatefulClient(*resp.JSON200.Data)
	if err != nil {
		return nil, fmt.Errorf("could not get stateful clients parameter: %s", err)
	}

	if sc.Status == string(backend.ContainerStatusUP) {
		// The management processes for the client are operational.
		// However, one or more frontend processes are not operational.
		// We should wait until both the management and frontend processes to be in the operational.
		return b.getStatefulClientProcessesStatus(ctx, *sc)
	}

	return sc, nil
}

func (b *Backend) ListStatefulClients(ctx context.Context, opts backend.ListStatefulClientsOpts) ([]*backend.StatefulClient, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	return b.getStatefulClients(ctx, opts.Names, token)
}

func (b *Backend) getStatefulClients(ctx context.Context, namesToFilter []string, token *securityprovider.SecurityProviderBearerToken) ([]*backend.StatefulClient, error) {
	resp, err := b.client.GetContainersWithResponse(ctx, v4.RequestEditorFn(token.Intercept))
	if err != nil {
		return nil, err
	}

	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get stateful clients", resp.StatusCode(), resp.Body)
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	statefulClients := make([]*backend.StatefulClient, 0)

	for _, scs := range *resp.JSON200.Data {
		sc, err := intoStatefulClient(scs)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Could not parse statefulclient from containers info")
			continue
		}

		if len(namesToFilter) > 0 && !slices.Contains(namesToFilter, sc.Name) {
			continue
		}

		if sc.Mode == string(backend.ContainerModeClient) {
			statefulClients = append(statefulClients, sc)
		}
	}

	return statefulClients, nil
}

func (b *Backend) getStatefulClientProcessesStatus(ctx context.Context, fsc backend.StatefulClient) (*backend.StatefulClient, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	resp, err := b.client.GetProcessesWithResponse(ctx, v4.RequestEditorFn(token.Intercept))
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get stateful client processes", resp.StatusCode(), resp.Body)
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	statefulClientProcesses := make([]*backend.Process, 0)

	for _, scp := range *resp.JSON200.Data {
		p, err := intoStatefulClientProcess(scp)
		if err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("Could not parse statefulclient processes info")
			continue
		}

		if fsc.Name != p.Hostname {
			continue
		}

		if p.Mode == string(backend.ContainerModeClient) {
			statefulClientProcesses = append(statefulClientProcesses, p)
		}
	}

	if len(statefulClientProcesses) != (fsc.Cores + 1) {
		fsc.Status = string(backend.ContainerStatusProcessesNotUP)
		return &fsc, nil
	}

	for _, ps := range statefulClientProcesses {
		if ps.Role != string(backend.ContainerRoleFrontend) {
			continue
		}
		if ps.Status != string(backend.ContainerStatusUP) {
			fsc.Status = string(backend.ContainerStatusProcessesNotUP)
			return &fsc, nil
		}
	}

	return &fsc, nil

}
