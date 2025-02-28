// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package processor

import (
	"context"

	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

// Result struct of info which is used to drive the next step after this
// method is complete.
type ListenerStatusResult struct {
	Requeue                  bool
	ListenerStatusConditions *loadbalancerv1alpha1.ConditionsListenerStatus
	ListenerStatus           *loadbalancerv1alpha1.ListenerStatus
}

// Processes a single Listener for a Loadbalancer.
func (p *Processor) processListener(ctx context.Context, loadbalancer *loadbalancerv1alpha1.Loadbalancer,
	listener loadbalancerv1alpha1.LoadbalancerListener, primaryVIPExisting bool,
	listenerStatusConditions *loadbalancerv1alpha1.ConditionsListenerStatus, listenerStatus *loadbalancerv1alpha1.ListenerStatus) (*ListenerStatusResult, error) {

	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Processor.processListener").WithValues("name", loadbalancer.Name).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	result := &ListenerStatusResult{
		Requeue:                  false,
		ListenerStatusConditions: listenerStatusConditions,
		ListenerStatus:           listenerStatus,
	}

	// Create Virtual Server attach ssl while creating if provided in resource
	if !result.ListenerStatusConditions.VIPCreated {

		lbType := loadbalancerv1alpha1.IPType_PUBLIC
		if primaryVIPExisting {
			lbType = loadbalancerv1alpha1.IPType_EXISTING
		}

		updatedStatusMessage, err := p.LBProvider.CreateVirtualServer(ctx, loadbalancer.Name, loadbalancer.Namespace,
			listener.VIP, result.ListenerStatus.PoolID, string(lbType), loadbalancer.Status.Vip)
		if err != nil {
			result.ListenerStatus.Message = updatedStatusMessage
			return result, err
		}

		log.Info("create virtual server: requeuing lb")
		result.Requeue = true
		return result, nil
	}

	// Get the members of the load balancer
	poolMembersReady, err := p.getReadyPoolMembers(ctx, loadbalancer.Namespace, listener)
	if err != nil {
		return result, err
	}

	// Create pool
	if !result.ListenerStatusConditions.PoolCreated && len(poolMembersReady) >= 1 {

		// Add finalizers to all instances that are part of the pool first before configuring the pool.
		// This ensures that the instance is tracked via finalizer before being a part of the lb.
		for _, instance := range poolMembersReady {
			if err := p.PersistInstanceFinalizer(ctx, add, instance); err != nil {
				return result, err
			}
		}

		updatedStatusMessage, err := p.LBProvider.CreatePool(ctx, loadbalancer.Name, loadbalancer.Namespace, listener.Pool, poolMembersReady, listener.VIP.Port)
		if err != nil {
			result.ListenerStatus.Message = updatedStatusMessage
			// Need to understand the error code and requeue.
			// If the error is due data issue (400), we may not requeue, otherwise we shall
			// requeue (what if the end point is not reachable and available later)
			log.Error(err, "error creating pool", logkeys.ListenPort, listener.VIP.Port, logkeys.LoadBalancerName, loadbalancer.Name)
			return result, err
		}

		log.Info("create pool: requeuing lb")
		result.Requeue = true
		return result, nil
	}

	// Create the linkage between Virtual server and Pool (if it does not exist already)
	if !result.ListenerStatusConditions.VIPPoolLinked &&
		(result.ListenerStatusConditions.VIPCreated &&
			result.ListenerStatusConditions.PoolCreated) {
		// Call the linkVSToPool  .. pass the pool id, VIP ID in a PUT Request
		updatedStatusMessage, err := p.LBProvider.LinkVSToPool(ctx, result.ListenerStatus.VipID, result.ListenerStatus.PoolID)
		if err != nil {
			result.ListenerStatus.Message = updatedStatusMessage
			return result, err
		}

		log.Info("link virtual server & pool: requeuing lb")
		result.Requeue = true
		return result, nil
	}

	// Observe if any changes in the pool
	result, err = p.ReconcileListenerPoolMembers(ctx, loadbalancer.Namespace, listener, poolMembersReady, result)
	if err != nil {
		return result, err
	}

	// Requeue if pool has not been associated to virtual server.
	if result.ListenerStatusConditions.PoolCreated && result.ListenerStatusConditions.VIPCreated &&
		!result.ListenerStatusConditions.VIPPoolLinked {
		log.Info("re-queuing to associate pool to virtual server")
		result.Requeue = true
		return result, nil
	}

	return result, nil
}
