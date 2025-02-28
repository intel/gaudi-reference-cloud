// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import "time"

// Application configuration
type Config struct {
	InstanceSchedulerAddr              string        `koanf:"instanceSchedulerAddr"`
	FleetAdminServerAddr               string        `koanf:"fleetAdminServerAddr"`
	SchedulerStatisticsPollingInterval time.Duration `koanf:"schedulerStatisticsPollingInterval"`
}
