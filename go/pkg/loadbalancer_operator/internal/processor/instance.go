// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package processor

import (
	"context"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	k8sv1 "k8s.io/api/core/v1"
)

func IsInstanceRunning(instance *privatecloudv1alpha1.Instance) bool {
	cond := FindStatusCondition(instance.Status.Conditions, privatecloudv1alpha1.InstanceConditionRunning)
	return cond != nil && cond.Status == k8sv1.ConditionTrue
}

// Utility to find a condition of the given type.
func FindStatusCondition(conditions []privatecloudv1alpha1.InstanceCondition, conditionType privatecloudv1alpha1.InstanceConditionType) *privatecloudv1alpha1.InstanceCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

// ReconcileListenerPoolMembers is used reconcile the members of a Pool with the desired configuration,
// of the LB Listener. It also removes Finalizers from any Instance which is no longer a part of a LB Pool.
func (p *Processor) ReconcileListenerPoolMembers(ctx context.Context, namespace string, listener loadbalancerv1alpha1.LoadbalancerListener,
	poolMembersReady []privatecloudv1alpha1.Instance, listenerStatus *ListenerStatusResult) (*ListenerStatusResult, error) {

	log := log.FromContext(ctx)

	var err error
	var updatedStatusMessage string

	// Can only reconcile a pool if the pool exists
	if listenerStatus.ListenerStatusConditions.PoolCreated {
		updatedStatusMessage, _, err = p.LBProvider.ObserveCurrentAndReconcile(ctx, namespace, listener,
			listenerStatus.ListenerStatus.PoolID, poolMembersReady)
		if err != nil {
			log.Error(err, "failed to observe and reconcile: "+updatedStatusMessage)
			return listenerStatus, err
		}
	}

	// Update status of the pool members in the Load Balancer object
	var poolStatusMembers []loadbalancerv1alpha1.PoolStatusMember
	for _, instance := range poolMembersReady {

		// Append to the poolStatusMembers list so it can be added to the listener status object later
		poolStatusMembers = append(poolStatusMembers, loadbalancerv1alpha1.PoolStatusMember{
			InstanceResourceId: instance.Name,
			IPAddress:          instance.Status.Interfaces[0].Addresses[0],
		})

		// Add finalizers to all instances that are part of the pool if not already added
		if err := p.PersistInstanceFinalizer(ctx, add, instance); err != nil {
			log.Error(err, "failed to add finalizer", logkeys.InstanceName, instance.Name, logkeys.InstanceNamespace, instance.Namespace)
			return listenerStatus, err
		}
	}

	// Set the PoolMembers of the listener
	listenerStatus.ListenerStatus.PoolMembers = poolStatusMembers

	return listenerStatus, nil
}

func (p *Processor) getReadyPoolMembers(ctx context.Context, namespace string, listener loadbalancerv1alpha1.LoadbalancerListener) ([]privatecloudv1alpha1.Instance, error) {
	var poolMembersReady []privatecloudv1alpha1.Instance

	// Get the members of the load balancer
	poolMembers, err := p.GetLoadbalancerInstances(ctx, namespace, listener)
	if err != nil {
		return poolMembers, err
	}

	for _, p := range poolMembers {
		// An instance is ready if it is "Running: true" && not being deleted
		if IsInstanceRunning(&p) && p.DeletionTimestamp.IsZero() {
			poolMembersReady = append(poolMembersReady, p)
		}
	}

	return poolMembersReady, nil
}
