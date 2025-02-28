// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubernetes 1.24 kube-scheduler (https://github.com/kubernetes/kubernetes/tree/73da4d3652771d6c6dfe904fe8fae594a1a72e2b/pkg/scheduler).
// To see changes made, run diff-kube-scheduler.sh.

/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package profile

import (
	"fmt"
	"strings"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	frameworkruntime "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/runtime"
)

var (
	registry frameworkruntime.Registry
	opts     []frameworkruntime.Option
)

func TestNewMap(t *testing.T) {
	cases := []struct {
		name    string
		cfgs    []config.KubeSchedulerProfile
		wantErr string
	}{
		{
			name: "valid",
			cfgs: []config.KubeSchedulerProfile{
				{
					SchedulerName: "profile-1",
					Plugins:       &config.Plugins{},
				},
				{
					SchedulerName: "profile-2",
					Plugins:       &config.Plugins{},
				},
			},
		},
		{
			name: "duplicate scheduler name",
			cfgs: []config.KubeSchedulerProfile{
				{
					SchedulerName: "profile-1",
					Plugins:       &config.Plugins{},
				},
				{
					SchedulerName: "profile-1",
					Plugins:       &config.Plugins{},
				},
			},
			wantErr: "duplicate profile",
		},
		{
			name: "scheduler name is needed",
			cfgs: []config.KubeSchedulerProfile{
				{
					Plugins: &config.Plugins{},
				},
			},
			wantErr: "scheduler name is needed",
		},
		{
			name: "plugins required for profile",
			cfgs: []config.KubeSchedulerProfile{
				{
					SchedulerName: "profile-1",
				},
			},
			wantErr: "plugins required for profile",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := NewMap(tc.cfgs, registry, opts...)
			if err := checkErr(err, tc.wantErr); err != nil {
				t.Fatal(err)
			}
			if len(tc.wantErr) != 0 {
				return
			}
			if len(m) != len(tc.cfgs) {
				t.Errorf("got %d profiles, want %d", len(m), len(tc.cfgs))
			}
		})
	}
}

func checkErr(err error, wantErr string) error {
	if len(wantErr) == 0 {
		return err
	}
	if err == nil || !strings.Contains(err.Error(), wantErr) {
		return fmt.Errorf("got error %q, want %q", err, wantErr)
	}
	return nil
}
