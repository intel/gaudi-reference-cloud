// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package replicator

import (
	"testing"

	lbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
)

func Test_areEqual(t *testing.T) {
	tests := map[string]struct {
		oldStatus lbv1alpha1.LoadbalancerStatus
		newStatus lbv1alpha1.LoadbalancerStatus
		want      bool
	}{
		"empty": {
			oldStatus: lbv1alpha1.LoadbalancerStatus{},
			newStatus: lbv1alpha1.LoadbalancerStatus{},
			want:      true,
		},
		"simple": {
			oldStatus: lbv1alpha1.LoadbalancerStatus{
				State:   lbv1alpha1.PENDING,
				Vip:     "1.2.3.4",
				Message: "foo",
				Conditions: lbv1alpha1.ConditionsStatus{
					Listeners: []lbv1alpha1.ConditionsListenerStatus{{
						Port:          9090,
						PoolCreated:   true,
						VIPPoolLinked: true,
						VIPCreated:    true,
					}, {
						Port:          8081,
						PoolCreated:   false,
						VIPPoolLinked: false,
						VIPCreated:    true,
					}},
					FirewallRuleCreated: true,
				},
				Listeners: []lbv1alpha1.ListenerStatus{{
					Port:    9090,
					Name:    "mylb",
					Message: "current status is foo",
					PoolMembers: []lbv1alpha1.PoolStatusMember{{
						InstanceResourceId: "c-d-e-f",
						IPAddress:          "2.2.2.2",
					}, {
						InstanceResourceId: "a-b-c-d",
						IPAddress:          "1.1.1.1",
					}},
					PoolID: 11,
					VipID:  12,
				}},
			},
			newStatus: lbv1alpha1.LoadbalancerStatus{
				State:   lbv1alpha1.PENDING,
				Vip:     "1.2.3.4",
				Message: "foo",
				Conditions: lbv1alpha1.ConditionsStatus{
					Listeners: []lbv1alpha1.ConditionsListenerStatus{{
						Port:          9090,
						PoolCreated:   true,
						VIPPoolLinked: true,
						VIPCreated:    true,
					}, {
						Port:          8081,
						PoolCreated:   false,
						VIPPoolLinked: false,
						VIPCreated:    true,
					}},
					FirewallRuleCreated: true,
				},
				Listeners: []lbv1alpha1.ListenerStatus{{
					Port:    9090,
					Name:    "mylb",
					Message: "current status is foo",
					PoolMembers: []lbv1alpha1.PoolStatusMember{{
						InstanceResourceId: "c-d-e-f",
						IPAddress:          "2.2.2.2",
					}, {
						InstanceResourceId: "a-b-c-d",
						IPAddress:          "1.1.1.1",
					}},
					PoolID: 11,
					VipID:  12,
				}},
			},
			want: true,
		},
		"different states": {
			oldStatus: lbv1alpha1.LoadbalancerStatus{
				State:   lbv1alpha1.PENDING,
				Vip:     "1.2.3.4",
				Message: "foo",
				Conditions: lbv1alpha1.ConditionsStatus{
					Listeners: []lbv1alpha1.ConditionsListenerStatus{{
						Port:          9090,
						PoolCreated:   true,
						VIPPoolLinked: true,
						VIPCreated:    true,
					}, {
						Port:          8081,
						PoolCreated:   false,
						VIPPoolLinked: false,
						VIPCreated:    true,
					}},
					FirewallRuleCreated: true,
				},
				Listeners: []lbv1alpha1.ListenerStatus{{
					Port:    9090,
					Name:    "mylb",
					Message: "current status is foo",
					PoolMembers: []lbv1alpha1.PoolStatusMember{{
						InstanceResourceId: "c-d-e-f",
						IPAddress:          "2.2.2.2",
					}, {
						InstanceResourceId: "a-b-c-d",
						IPAddress:          "1.1.1.1",
					}},
					PoolID: 11,
					VipID:  12,
				}},
			},
			newStatus: lbv1alpha1.LoadbalancerStatus{
				State:   lbv1alpha1.DELETING,
				Vip:     "1.2.3.4",
				Message: "foo",
				Conditions: lbv1alpha1.ConditionsStatus{
					Listeners: []lbv1alpha1.ConditionsListenerStatus{{
						Port:          9090,
						PoolCreated:   true,
						VIPPoolLinked: true,
						VIPCreated:    true,
					}, {
						Port:          8081,
						PoolCreated:   false,
						VIPPoolLinked: false,
						VIPCreated:    true,
					}},
					FirewallRuleCreated: true,
				},
				Listeners: []lbv1alpha1.ListenerStatus{{
					Port:    9090,
					Name:    "mylb",
					Message: "current status is foo",
					PoolMembers: []lbv1alpha1.PoolStatusMember{{
						InstanceResourceId: "c-d-e-f",
						IPAddress:          "2.2.2.2",
					}, {
						InstanceResourceId: "a-b-c-d",
						IPAddress:          "1.1.1.1",
					}},
					PoolID: 11,
					VipID:  12,
				}},
			},
			want: false,
		},
		"different conditions, listeners": {
			oldStatus: lbv1alpha1.LoadbalancerStatus{
				State: lbv1alpha1.PENDING,
				Vip:   "1.2.3.4",
				Conditions: lbv1alpha1.ConditionsStatus{
					Listeners: []lbv1alpha1.ConditionsListenerStatus{{
						Port:          9090,
						PoolCreated:   true,
						VIPPoolLinked: true,
						VIPCreated:    true,
					}, {
						Port:          8081,
						PoolCreated:   false,
						VIPPoolLinked: false,
						VIPCreated:    true,
					}},
					FirewallRuleCreated: true,
				},
				Listeners: []lbv1alpha1.ListenerStatus{{
					Port:    9090,
					Name:    "mylb",
					Message: "current status is foo",
					PoolMembers: []lbv1alpha1.PoolStatusMember{{
						InstanceResourceId: "c-d-e-f",
						IPAddress:          "2.2.2.2",
					}, {
						InstanceResourceId: "a-b-c-d",
						IPAddress:          "1.1.1.1",
					}},
					PoolID: 11,
					VipID:  12,
				}},
			},
			newStatus: lbv1alpha1.LoadbalancerStatus{
				State: lbv1alpha1.PENDING,
				Vip:   "1.2.3.4",
				Conditions: lbv1alpha1.ConditionsStatus{
					Listeners: []lbv1alpha1.ConditionsListenerStatus{{
						Port:          8080, // <--- changed
						PoolCreated:   true,
						VIPPoolLinked: true,
						VIPCreated:    true,
					}},
					FirewallRuleCreated: true,
				},
				Listeners: []lbv1alpha1.ListenerStatus{{
					Port:    9090,
					Name:    "mylb",
					Message: "current status is foo",
					PoolMembers: []lbv1alpha1.PoolStatusMember{{
						InstanceResourceId: "a-b-c-d",
						IPAddress:          "1.1.1.1",
					}, {
						InstanceResourceId: "c-d-e-f",
						IPAddress:          "2.2.2.2",
					}},
					PoolID: 11,
					VipID:  12,
				}},
			},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := areEqual(tt.oldStatus, tt.newStatus); got != tt.want {
				t.Errorf("areEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
