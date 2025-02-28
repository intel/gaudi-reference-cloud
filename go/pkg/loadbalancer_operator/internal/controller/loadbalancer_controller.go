// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/processor"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/provider"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// LoadbalancerReconciler reconciles a Loadbalancer object
type LoadbalancerReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	LBProvider provider.Provider
	*processor.Processor
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=loadbalancers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=loadbalancers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=loadbalancers/finalizers,verbs=update
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=firewallrules/status,verbs=get;update;patch

func (r *LoadbalancerReconciler) MapInstanceToLoadbalancer(ctx context.Context, object client.Object) []reconcile.Request {

	log := log.FromContext(ctx).WithName("LoadbalancerReconciler.MapInstanceToLoadBalancer")

	instance := object.(*privatecloudv1alpha1.Instance)

	// Look up all lbs that this instance is attached to.
	lbs, err := r.Processor.GetLoadbalancers(ctx, instance)
	if err != nil {
		return nil
	}

	requests := []reconcile.Request{}

	// Iterate over each load balancer and enqueue a reconcile event for all the load balancers
	// which are reference in this Instance.
	for _, lb := range lbs {
		log.Info("instances enqueue", logkeys.LoadBalancerName, lb.Name, logkeys.LoadBalancerNamespace, lb.Namespace)
		requests = append(requests, reconcile.Request{
			NamespacedName: client.ObjectKeyFromObject(lb),
		})
	}

	if len(requests) > 0 {
		log.Info("instances enqueue", logkeys.NumOfLoadBalancers, len(requests))
	}

	return requests
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *LoadbalancerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("LoadbalancerReconciler.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	log.Info("BEGIN")

	result, reconcileErr := func() (ctrl.Result, error) {
		// Get the provider that we will be working with and the loadbalancer_instance
		loadbalancer, err := r.Processor.GetLoadbalancer(ctx, req.NamespacedName)
		if err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		if loadbalancer == nil {
			log.Info("Ignoring reconcile request because source loadbalancer was not found in cache")
			return ctrl.Result{}, nil
		}

		log.Info("reconcile load balancer", logkeys.LoadBalancerName, loadbalancer.Name)

		result, processErr := func() (ctrl.Result, error) {

			var requeue bool
			var err error

			if !loadbalancer.DeletionTimestamp.IsZero() {
				_, err = r.Processor.ProcessDeleteLoadbalancer(ctx, loadbalancer)
				if err != nil {
					return ctrl.Result{}, err
				}
			} else {
				requeue, err = r.ProcessLoadbalancer(ctx, loadbalancer)
				if err != nil {
					return ctrl.Result{}, err
				}
			}

			if err = r.Processor.PersistLoadbalancerStatusUpdate(ctx, loadbalancer, req.NamespacedName); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{Requeue: requeue}, nil
		}()

		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "InstanceReconciler.Reconcile: error reconciling Loadbalancer")
	}
	log.Info("END", "result", result, "err", reconcileErr)
	return result, reconcileErr
}
