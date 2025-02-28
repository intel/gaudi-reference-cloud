// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package weka

import (
	"context"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
	"github.com/rs/zerolog/log"
)

func (b *Backend) GetStatus(ctx context.Context) (*backend.ClusterStatus, error) {
	token, err := b.login(ctx, b.adminCredentials, "Root")
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	resp, err := b.client.GetClusterStatusWithResponse(ctx, v4.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Weka API")
		return nil, err
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil || resp.JSON200.Data == nil {
		return nil, backend.ResponseAsErr("could not get cluster, status", resp.StatusCode(), resp.Body)
	}
	var status backend.HealthStatus

	if resp.JSON200.Data.Status != nil && *resp.JSON200.Data.Status == "OK" {
		status = backend.Healthy
	} else {
		status = backend.Unhealthy
	}

	var availableBytes uint64
	var totalBytes uint64

	if resp.JSON200.Data.Capacity != nil && resp.JSON200.Data.Capacity.UnprovisionedBytes != nil {
		availableBytes = *resp.JSON200.Data.Capacity.UnprovisionedBytes
	}

	if resp.JSON200.Data.Capacity != nil && resp.JSON200.Data.Capacity.TotalBytes != nil {
		totalBytes = *resp.JSON200.Data.Capacity.TotalBytes
	}

	namespaces, err := b.getNamespaces(ctx, make([]string, 0), token)

	if err != nil {
		log.Error().Ctx(ctx).Msg("failed to list namespaces")
	}

	labels := make(map[string]string)

	if resp.JSON200.Data.Guid != nil {
		labels["wekaGuid"] = *resp.JSON200.Data.Guid
	}

	if resp.JSON200.Data.Name != nil {
		labels["wekaName"] = *resp.JSON200.Data.Name
	}

	return &backend.ClusterStatus{
		AvailableBytes:      availableBytes,
		TotalBytes:          totalBytes,
		NamespacesLimit:     256,
		NamespacesAvailable: 256 - int32(len(namespaces)) - int32(len(b.config.WekaConfig.ProtectedOrgIds)),
		HealthStatus:        status,
		Labels:              labels,
	}, nil
}
