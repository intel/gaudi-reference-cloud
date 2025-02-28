// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright 2023.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// define virtual server configuration
type VServer struct {
	Port       int    `json:"port"`
	IPType     string `json:"iptype"`
	Persist    string `json:"persist"`
	IPProtocol string `json:"ipprotocol"`

	// SSL is an optional field which allows for SSL configuration at the vServer
	SSLConfig *SSL `json:"ssl,omitempty"`
}

type SSL struct {
	Profile Profile `json:"profile,omitempty"`
}

// Profile is the SSL profile configuration for the specific listener
type Profile struct {
	Id int `json:"id,omitempty"`
}

// Define pool configuration
type VPool struct {
	Port int `json:"port,omitempty"`
	// +kubebuilder:validation:MinItems=0
	// +optional
	Members           []VMember         `json:"members"`
	LoadBalancingMode string            `json:"loadbalancingmode"`
	MinActiveMembers  int               `json:"minactivemembers"`
	Monitor           string            `json:"monitor"`
	InstanceSelectors map[string]string `json:"instanceSelectors,omitempty"`
}

// receiver method for vPool - DeepCopyInto
func (in *VPool) DeepCopyInto(out *VPool) {
	out.Members = make([]VMember, len(in.Members))
	for i, member := range in.Members {
		out.Members[i] = member
	}
}

// Define member configuration -- part of vPool
type VMember struct {
	Name               string `json:"name,omitempty"`
	ConnectionLimit    int    `json:"connectionLimit"`
	PriorityGroup      int    `json:"priorityGroup"`
	Ratio              int    `json:"ratio"`
	AdminState         string `json:"adminState"`
	MonitorStatus      string `json:"monitorStatus"`
	InstanceResourceId string `json:"instanceRef"`
}

type State string

const (
	PENDING  State = "Pending"
	DEGRADED State = "Degraded"
	READY    State = "Active"
	DELETING State = "Deleting"
	DELETED  State = "Deleted"
)

type IPType string

const (
	IPType_PUBLIC   IPType = "public"
	IPType_PRIVATE  IPType = "private"
	IPType_EXISTING IPType = "existing"
)

type IPProtocol string

const (
	IPProtocol_TCP IPType = "tcp"
	IPProtocol_UDP IPType = "udp"
)

type MonitorType string

const (
	MonitorType_HTTP  MonitorType = "http"
	MonitorType_HTTPS MonitorType = "https"
	MonitorType_TCP   MonitorType = "tcp"
)

// LoadbalancerSpec defines the desired state of Loadbalancer
type LoadbalancerSpec struct {
	Listeners []LoadbalancerListener `json:"listeners"`

	Security LoadbalancerSecurity `json:"security"`

	// User defined labels when a new load balancer is created
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

type LoadbalancerSecurity struct {
	// +kubebuilder:validation:MinItems=1
	Sourceips []string `json:"sourceips"`
}

type LoadbalancerListener struct {
	// +optional
	Pool  VPool   `json:"pool"`
	VIP   VServer `json:"vip"`
	Owner string  `json:"owner"`
}

type PoolStatusMember struct {
	InstanceResourceId string `json:"instanceRef"`
	IPAddress          string `json:"ip"`
}

// LoadbalancerStatus defines the observed state of Loadbalancer
type LoadbalancerStatus struct {
	//define observed state of cluster
	State      State            `json:"state"`
	Message    string           `json:"message"`
	Vip        string           `json:"vip"`
	Conditions ConditionsStatus `json:"conditions"`
	Listeners  []ListenerStatus `json:"listeners"`
}

type ConditionsStatus struct {
	Listeners           []ConditionsListenerStatus `json:"listeners"`
	FirewallRuleCreated bool                       `json:"firewallRuleCreated"`
}

type ConditionsListenerStatus struct {
	Port          int  `json:"port"`
	PoolCreated   bool `json:"poolCreated"`
	VIPPoolLinked bool `json:"vipPoolLinked"`
	VIPCreated    bool `json:"vipCreated"`
}

type ListenerStatus struct {
	Port        int                `json:"port"`
	Name        string             `json:"name"`
	State       State              `json:"state"`
	Message     string             `json:"message"`
	PoolMembers []PoolStatusMember `json:"poolMembers,omitempty"`
	PoolID      int                `json:"poolID"`
	VipID       int                `json:"vipID"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=lb;
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="State status",priority=1
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.message",description="Message"
// +kubebuilder:printcolumn:name="VIP",type="string",JSONPath=".status.vip",description="VIP"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of Loadbalancer"

// Loadbalancer is the Schema for the loadbalancers API
type Loadbalancer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadbalancerSpec   `json:"spec,omitempty"`
	Status LoadbalancerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LoadbalancerList contains a list of Loadbalancer
type LoadbalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Loadbalancer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Loadbalancer{}, &LoadbalancerList{})
}
