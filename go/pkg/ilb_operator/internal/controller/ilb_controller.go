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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ilbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ilb_operator/api/v1alpha1"
	lb_provider "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ilb_operator/internal/LB_Provider"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

const (
	deleteIlbFinalizer = "private.cloud.intel.com/deleteIlb"
)

// IlbReconciler reconciles a Ilb object
type IlbReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	*LoadBalancerProviderConfig
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=ilbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=ilbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=ilbs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ilb object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *IlbReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("IlbReconciler.Reconcile").Start()
	defer span.End()

	// Get the provider that we will be working with and the ilb_instance
	var ilb_instance ilbv1alpha1.Ilb
	if err := r.Get(ctx, req.NamespacedName, &ilb_instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	//get the load balancer provider we will be using for this instance.
	p, err := lb_provider.NewLoadBalancerProvider(r.LoadBalancerProviderConfig.ProviderType, &lb_provider.Config{
		BaseURL:         r.LoadBalancerProviderConfig.BaseURL,
		Domain:          r.LoadBalancerProviderConfig.Domain,
		UserName:        r.LoadBalancerProviderConfig.UserName,
		Secret:          r.LoadBalancerProviderConfig.Secret,
		HighwireTimeout: r.LoadBalancerProviderConfig.HighwireTimeout,
	})
	if err != nil {
		log.Error(err, "Could not initialized the provider", logkeys.ProviderType, r.LoadBalancerProviderConfig.ProviderType)
		return ctrl.Result{}, err
	}
	log.Info("The Provider is initialized", logkeys.ProviderType, r.LoadBalancerProviderConfig.ProviderType, logkeys.ProviderURL, r.LoadBalancerProviderConfig.BaseURL)

	// Remove the associated artifacts if ILB Instance is deleted
	if !ilb_instance.DeletionTimestamp.IsZero() {
		log.Info("The deletetion Time stamp Exist", logkeys.DeletionTimestamp, ilb_instance.DeletionTimestamp.GoString())

		if err = p.ProcessFinalizers(&ilb_instance); err != nil {
			return ctrl.Result{}, err
		}

		//remove the finalizer if successful
		controllerutil.RemoveFinalizer(&ilb_instance, deleteIlbFinalizer)
		//Update the VIP STATE -- this may not be needed when our finalizer removal works as expected but no harm in changing the state explicitly
		ilb_instance.Status.State = ilbv1alpha1.TERMINATED

		if err := r.Update(ctx, &ilb_instance); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Successfully removed the finalizers")
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&ilb_instance, deleteIlbFinalizer) {
		controllerutil.AddFinalizer(&ilb_instance, deleteIlbFinalizer)
		if err := r.Update(ctx, &ilb_instance); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	if err = p.GetStatus(&ilb_instance); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.Status().Update(ctx, &ilb_instance); err != nil {
		return ctrl.Result{}, err
	}

	// Create Virtual Server
	if !ilb_instance.Status.Conditions.VIPCreated {
		if err := p.CreateVirtualServer(&ilb_instance); err != nil {
			if err := r.Status().Update(ctx, &ilb_instance); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	// Create pool
	if !ilb_instance.Status.Conditions.PoolCreated && (len(ilb_instance.Spec.Pool.Members) >= 1) {
		if err := p.CreatePool(&ilb_instance); err != nil {
			if err := r.Status().Update(ctx, &ilb_instance); err != nil {
				return ctrl.Result{}, err
			}
			//here we need to understand the error code and requeue
			//If the error is due data issue (400), we may not requeue , otherwise we shall requeue (what if the end point is not reachable and available later)
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	// Create the linkage between Virtual server and Pool (if it doesnt exist already)
	if !ilb_instance.Status.Conditions.VIPPoolLinked && (ilb_instance.Status.Conditions.VIPCreated && ilb_instance.Status.Conditions.PoolCreated) {
		//call the linkVSToPool  .. pass the pool id, VIP ID in a PUT Request
		if err = p.LinkVSToPool(&ilb_instance); err != nil {
			if err := r.Status().Update(ctx, &ilb_instance); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	// Observe if any changes in the pool
	// This is only pool member reconciliation as of now. Do this only if Pool is created
	if ilb_instance.Status.Conditions.PoolCreated { //we may need to add another condition here .. should we observe only after READY ?
		if err = p.ObserveCurrentAndReconcile(&ilb_instance); err != nil {
			log.Error(err, "failed to observe and reconcile")
		}
	}

	// Requeue if pool has not been associated to virtual server.
	if ilb_instance.Status.Conditions.PoolCreated && ilb_instance.Status.Conditions.VIPCreated && !ilb_instance.Status.Conditions.VIPPoolLinked {
		log.V(0).Info("requeuing to associate pool to virtual server")
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IlbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ilbv1alpha1.Ilb{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.IlbMaxConcurrentReconciles,
		}).
		Complete(r)
}
