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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// define virtual server configuration
type vServer struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Port        int    `json:"port"`
	IPType      string `json:"iptype"`
	Persist     string `json:"persist"`
	IPProtocol  string `json:"ipprotocol"`
	Environment int    `json:"environment"`
	UserGroup   int    `json:"usergroup"`
}

// Define pool configuration
type vPool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Port        int    `json:"port,omitempty"`
	// +kubebuilder:validation:MinItems=0
	// +optional
	Members           []VMember `json:"members"`
	Environment       int       `json:"environment"`
	UserGroup         int       `json:"usergroup"`
	LoadBalancingMode string    `json:"loadbalancingmode"`
	MinActiveMembers  int       `json:"minactivemembers"`
	Monitor           string    `json:"monitor"`
}

// receiver method for vPool - DeepCopyInto
func (in *vPool) DeepCopyInto(out *vPool) {
	out.Members = make([]VMember, len(in.Members))
	for i, member := range in.Members {
		out.Members[i] = member
	}
}

// Define member configuration -- part of vPool
type VMember struct {
	Name            string `json:"name,omitempty"`
	IP              string `json:"ip"`
	ConnectionLimit int    `json:"connectionLimit"`
	PriorityGroup   int    `json:"priorityGroup"`
	Ratio           int    `json:"ratio"`
	AdminState      string `json:"adminState"`
}

type State string

const (
	PENDING    State = "Pending"
	READY      State = "Active"
	TERMINATED State = "Deleting"
	ERROR      State = "Error"
)

type ilbConditions struct {
	PoolCreated   bool `json:"poolCreated"`
	VIPCreated    bool `json:"vipCreated"`
	VIPPoolLinked bool `json:"vipPoolLinked"`
}

// IlbSpec defines the desired state of Ilb
type IlbSpec struct {
	// +optional
	Pool  vPool   `json:"pool"`
	VIP   vServer `json:"vip"`
	Owner string  `json:"owner"`
}

// IlbStatus defines the observed state of Ilb
type IlbStatus struct {
	//define observed state of cluster
	Name       string        `json:"name"`
	State      State         `json:"state"`
	Message    string        `json:"message"`
	Vip        string        `json:"vip"`
	PoolID     int           `json:"poolID"`
	VipID      int           `json:"vipID"`
	Conditions ilbConditions `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ilb is the Schema for the ilbs API
type Ilb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IlbSpec   `json:"spec,omitempty"`
	Status IlbStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IlbList contains a list of Ilb
type IlbList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ilb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ilb{}, &IlbList{})
}
