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

// Package profile holds the definition of a scheduling Profile.
package profile

import (
	"errors"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	frameworkruntime "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/runtime"
)

// newProfile builds a Profile for the given configuration.
func newProfile(cfg config.KubeSchedulerProfile, r frameworkruntime.Registry,
	opts ...frameworkruntime.Option) (framework.Framework, error) {
	return frameworkruntime.NewFramework(r, &cfg, opts...)
}

// Map holds frameworks indexed by scheduler name.
type Map map[string]framework.Framework

// NewMap builds the frameworks given by the configuration, indexed by name.
func NewMap(cfgs []config.KubeSchedulerProfile, r frameworkruntime.Registry,
	opts ...frameworkruntime.Option) (Map, error) {
	m := make(Map)
	v := cfgValidator{m: m}

	for _, cfg := range cfgs {
		p, err := newProfile(cfg, r, opts...)
		if err != nil {
			return nil, fmt.Errorf("creating profile for scheduler name %s: %v", cfg.SchedulerName, err)
		}
		if err := v.validate(cfg, p); err != nil {
			return nil, err
		}
		m[cfg.SchedulerName] = p
	}
	return m, nil
}

// HandlesSchedulerName returns whether a profile handles the given scheduler name.
func (m Map) HandlesSchedulerName(name string) bool {
	_, ok := m[name]
	return ok
}

type cfgValidator struct {
	m Map
}

func (v *cfgValidator) validate(cfg config.KubeSchedulerProfile, f framework.Framework) error {
	if len(f.ProfileName()) == 0 {
		return errors.New("scheduler name is needed")
	}
	if cfg.Plugins == nil {
		return fmt.Errorf("plugins required for profile with scheduler name %q", f.ProfileName())
	}
	if v.m[f.ProfileName()] != nil {
		return fmt.Errorf("duplicate profile with scheduler name %q", f.ProfileName())
	}
	return nil
}
