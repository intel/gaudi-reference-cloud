// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubernetes 1.24 kube-scheduler (https://github.com/kubernetes/kubernetes/tree/73da4d3652771d6c6dfe904fe8fae594a1a72e2b/pkg/scheduler).
// To see changes made, run diff-kube-scheduler.sh.

/*
Copyright 2019 The Kubernetes Authors.

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

package podtopologyspread

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type topologyPair struct {
	key   string
	value string
}

// topologySpreadConstraint is an internal version for v1.TopologySpreadConstraint
// and where the selector is parsed.
// Fields are exported for comparison during testing.
type topologySpreadConstraint struct {
	MaxSkew     int32
	TopologyKey string
	Selector    labels.Selector
	MinDomains  int32
}

// buildDefaultConstraints builds the constraints for a pod using
// .DefaultConstraints and the selectors from the services, replication
// controllers, replica sets and stateful sets that match the pod.
func (pl *PodTopologySpread) buildDefaultConstraints(p *v1.Pod, action v1.UnsatisfiableConstraintAction) ([]topologySpreadConstraint, error) {
	return nil, nil
}

// nodeLabelsMatchSpreadConstraints checks if ALL topology keys in spread Constraints are present in node labels.
func nodeLabelsMatchSpreadConstraints(nodeLabels map[string]string, constraints []topologySpreadConstraint) bool {
	for _, c := range constraints {
		if _, ok := nodeLabels[c.TopologyKey]; !ok {
			return false
		}
	}
	return true
}

func filterTopologySpreadConstraints(constraints []v1.TopologySpreadConstraint, action v1.UnsatisfiableConstraintAction, enableMinDomainsInPodTopologySpread bool) ([]topologySpreadConstraint, error) {
	var result []topologySpreadConstraint
	for _, c := range constraints {
		if c.WhenUnsatisfiable == action {
			selector, err := metav1.LabelSelectorAsSelector(c.LabelSelector)
			if err != nil {
				return nil, err
			}
			tsc := topologySpreadConstraint{
				MaxSkew:     c.MaxSkew,
				TopologyKey: c.TopologyKey,
				Selector:    selector,
				MinDomains:  1, // if MinDomains is nil, we treat MinDomains as 1.
			}
			if enableMinDomainsInPodTopologySpread && c.MinDomains != nil {
				tsc.MinDomains = *c.MinDomains
			}
			result = append(result, tsc)
		}
	}
	return result, nil
}

func countPodsMatchSelector(podInfos []*framework.PodInfo, selector labels.Selector, ns string) int {
	count := 0
	for _, p := range podInfos {
		// Bypass terminating Pod (see #87621).
		if p.Pod.DeletionTimestamp != nil || p.Pod.Namespace != ns {
			continue
		}
		if selector.Matches(labels.Set(p.Pod.Labels)) {
			count++
		}
	}
	return count
}
