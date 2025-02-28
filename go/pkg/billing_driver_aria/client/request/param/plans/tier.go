// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package plans

type Tier struct {
	Schedule []TierSchedule `json:"schedule,omitempty" url:"schedule,omitempty"`
}
