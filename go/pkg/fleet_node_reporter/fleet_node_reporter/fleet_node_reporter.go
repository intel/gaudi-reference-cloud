// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package fleet_node_reporter

import (
	"context"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_node_reporter/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

// FleetNodeReporter is a struct that holds the instance scheduling service client and configuration.
type FleetNodeReporter struct {
	instanceSchedulingServiceClient pb.InstanceSchedulingServiceClient
	fleetAdminServiceClient         pb.FleetAdminServiceClient
	cfg                             config.Config
}

// NewFleetNodeReporter creates a new instance of FleetNodeReporter.
func NewFleetNodeReporter(ctx context.Context, instanceSchedulingServiceClient pb.InstanceSchedulingServiceClient,
	fleetAdminServiceClient pb.FleetAdminServiceClient, cfg config.Config) (*FleetNodeReporter, error) {
	return &FleetNodeReporter{
		instanceSchedulingServiceClient: instanceSchedulingServiceClient,
		fleetAdminServiceClient:         fleetAdminServiceClient,
		cfg:                             cfg,
	}, nil
}

// Start begins the process of periodically fetching statistics.
func (s *FleetNodeReporter) Start(ctx context.Context) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("FleetNodeReporter.Start").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	// Start a new goroutine to periodically fetch statistics.
	go func() {
		ticker := time.NewTicker(s.cfg.SchedulerStatisticsPollingInterval)

		// Ensure the ticker is stopped when the function exits.
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info("Context done, stopping ticker")

				// Exit the goroutine if the context is done.
				return
			case tickTime := <-ticker.C:
				log.Info("Ticker ticked", "tickTime", tickTime)

				// Fetch statistics on each tick.
				if err := s.getStatistics(ctx); err != nil {
					log.Error(err, "Error fetching statistics")
				}
			}
		}
	}()
}

// getStatistics fetches statistical data from the instance scheduling service.
func (s *FleetNodeReporter) getStatistics(ctx context.Context) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("FleetNodeReporter.getStatistics").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	// Call the GetStatistics method on the instance scheduling service client.
	log.Info("Fetching statistical data from instanceSchedulingService")
	resp, err := s.instanceSchedulingServiceClient.GetStatistics(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("schedulingServiceClient.GetStatistics returned empty response")
	}
	// Log the response received.
	log.Info("Statistics Response", "resp", resp)

	log.Info("Calling ReportNodeStatistics")
	// Call the ReportNodeStatistics method on fleetAdminService to sync NodeStatistics
	_, fleetAdminErr := s.fleetAdminServiceClient.ReportNodeStatistics(ctx, &pb.ReportNodeStatisticsRequest{
		SchedulerNodeStatistics: resp.SchedulerNodeStatistics,
	})
	if fleetAdminErr != nil {
		return fleetAdminErr
	}
	log.Info("ReportNodeStatistics call succeed")

	return nil
}
