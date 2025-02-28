// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubernetes 1.24 kube-scheduler (https://github.com/kubernetes/kubernetes/tree/73da4d3652771d6c6dfe904fe8fae594a1a72e2b/pkg/scheduler).
// To see changes made, run diff-kube-scheduler.sh.

/*
Copyright 2021 The Kubernetes Authors.

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

package names

const (
	InterPodAffinity                = "InterPodAffinity"
	NodeAffinity                    = "NodeAffinity"
	NodeResourcesBalancedAllocation = "NodeResourcesBalancedAllocation"
	NodeResourcesFit                = "NodeResourcesFit"
	NodeUnschedulable               = "NodeUnschedulable"
	PodTopologySpread               = "PodTopologySpread"
	TaintToleration                 = "TaintToleration"
)
