// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package vast

import (
	"context"
	"errors"

	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend"
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
	"github.com/rs/zerolog/log"
)

func (b *Backend) GetStatus(ctx context.Context) (*backend.ClusterStatus, error) {
	token, err := b.login(ctx, b.adminCredentials)
	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Admin login unsuccessful")
		return nil, err
	}

	clusters, err := b.client.ClustersListWithResponse(ctx, &client.ClustersListParams{}, client.RequestEditorFn(token.Intercept))

	if err != nil {
		log.Error().Ctx(ctx).Err(err).Msg("Error calling Vast API")
		return nil, err
	}

	if clusters.JSON200 == nil || len(*clusters.JSON200) != 1 {
		log.Error().Ctx(ctx).Any("response", clusters).Msg("Invalid clsuters response")
		return nil, errors.New("cluster response is not valid")
	}

	cluster := (*clusters.JSON200)[0]
	log.Info().Any("cluster", cluster).Msg("Vast cluster found")

	var status backend.HealthStatus
	if cluster.State != nil && *cluster.State == "ONLINE" {
		status = backend.Healthy
	} else {
		status = backend.Unhealthy
	}

	var usedSpace uint64
	var totalBytes uint64

	if cluster.LogicalSpaceInUse != nil {
		usedSpace = *cluster.LogicalSpaceInUse
	}

	if cluster.LogicalSpace != nil {
		totalBytes = *cluster.LogicalSpace
	}

	return &backend.ClusterStatus{
		HealthStatus:        status,
		TotalBytes:          totalBytes,
		AvailableBytes:      totalBytes - usedSpace,
		NamespacesLimit:     512,
		NamespacesAvailable: 512,
	}, nil
}
