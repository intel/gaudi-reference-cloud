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
	"fmt"
	"reflect"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/internal/provider"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	FirewallFinalizer = "private.cloud.intel.com/firewallOperator"
)

// FirewallRuleReconciler reconciles a FirewallRule object
type FirewallRuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	*FirewallRuleProviderConfig
	Provider provider.FirewallProvider
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=firewallrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=firewallrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=firewallrules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *FirewallRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Get the provider that we will be working with and the FirewallRule
	var firewallRule v1alpha1.FirewallRule
	if err := r.Get(ctx, req.NamespacedName, &firewallRule); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Remove the associated artifacts if FW Rule is deleted
	if !firewallRule.DeletionTimestamp.IsZero() {
		log.Info("The deletion time stamp exists", "Deletion time stamp", firewallRule.DeletionTimestamp.GoString())

		conditions := &firewallRule.Status.Conditions
		meta.SetStatusCondition(conditions, metav1.Condition{
			Type:               "Reconciling",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: firewallRule.Generation,
			LastTransitionTime: metav1.Now(),
			Message:            "Firewall Rule is reconciling",
			Reason:             "Reconciling",
		})
		meta.SetStatusCondition(conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			ObservedGeneration: firewallRule.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             "Ready",
			Message:            "Firewall rules are applied",
		})

		firewallRule.Status.State = v1alpha1.DELETING

		if err := r.PersistStatusUpdate(ctx, &firewallRule, req.NamespacedName); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("BEGIN rules remove")

		resp, err := r.Provider.RemoveAccess(ctx, firewallRule)
		if err != nil {
			log.Error(err, "error removing ports for firewall rule", "firewallRule", req.NamespacedName)
			return ctrl.Result{}, err
		}
		log.Info("END rules remove", "firewallRule", req.NamespacedName, "Response", resp)

		// Set the status to be terminated
		meta.SetStatusCondition(conditions, metav1.Condition{
			Type:               "Terminated",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: firewallRule.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             "Terminated",
			Message:            "Firewall rules are terminated",
		})

		firewallRule.Status.State = v1alpha1.DELETED

		if err := r.PersistStatusUpdate(ctx, &firewallRule, req.NamespacedName); err != nil {
			return ctrl.Result{}, err
		}

		// remove the finalizer if successful
		if err := r.PersistFinalizer(ctx, remove, firewallRule); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Successfully removed the finalizers")
		return ctrl.Result{}, nil
	}

	conditions := &firewallRule.Status.Conditions
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               "Reconciling",
		Status:             metav1.ConditionTrue,
		ObservedGeneration: firewallRule.Generation,
		LastTransitionTime: metav1.Now(),
		Message:            "Firewall Rule is reconciling",
		Reason:             "Reconciling",
	})
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: firewallRule.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "Ready",
		Message:            "Firewall rules are applied",
	})
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               "Terminated",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: firewallRule.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "Terminated",
		Message:            "Firewall rule is terminated",
	})

	firewallRule.Status.State = v1alpha1.RECONCILING
	firewallRule.Status.Message = "Rule is reconciling"

	if err := r.PersistStatusUpdate(ctx, &firewallRule, req.NamespacedName); err != nil {
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(&firewallRule, FirewallFinalizer) {
		log.Info("adding finalizer")
		if err := r.PersistFinalizer(ctx, add, firewallRule); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	if err := r.SyncRules(ctx, firewallRule); err != nil {

		firewallRule.Status.State = v1alpha1.RECONCILING
		firewallRule.Status.Message = err.Error()

		if err := r.PersistStatusUpdate(ctx, &firewallRule, req.NamespacedName); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	}

	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               "Reconciling",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: firewallRule.Generation,
		LastTransitionTime: metav1.Now(),
		Message:            "Firewall Rule is reconciling",
		Reason:             "Reconciling",
	})
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		ObservedGeneration: firewallRule.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "Ready",
		Message:            "Firewall rule is applied",
	})
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               "Terminated",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: firewallRule.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             "Terminated",
		Message:            "Firewall rule is terminated",
	})

	firewallRule.Status.State = v1alpha1.READY
	firewallRule.Status.Message = "Rule is ready"

	if err := r.PersistStatusUpdate(ctx, &firewallRule, req.NamespacedName); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FirewallRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.FirewallRule{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		Complete(r)
}

// SyncRules looks at all the rules defined for a tenant and syncronizes them to what exists
// in the firewall API.
func (r *FirewallRuleReconciler) SyncRules(ctx context.Context, fwRule v1alpha1.FirewallRule) error {

	log := log.FromContext(ctx)

	log.Info("BEGIN sync rules")
	defer log.Info("END sync rules")

	vip := fwRule.Spec.DestinationIP

	cloudAccountId, err := provider.GetCloudAccountId(fwRule)
	if err != nil {
		return err
	}

	// Get all rules for the tenant
	var firewallRules v1alpha1.FirewallRuleList
	if err := r.List(ctx, &firewallRules, client.InNamespace(fwRule.Namespace)); err != nil {
		log.Error(err, "could not list firewall rules for tenant", logkeys.CloudAccountId, cloudAccountId)
		return err
	}

	var rules []v1alpha1.FirewallRule
	// filter out any rules that are deleted, or those that don't match the same Destination IP,
	// they shouldn't be added or calculated in the set.
	for _, rule := range firewallRules.Items {
		if !rule.DeletionTimestamp.IsZero() || rule.Spec.DestinationIP != fwRule.Spec.DestinationIP {
			continue
		}
		rules = append(rules, rule)
	}

	existingRules, err := r.Provider.GetExistingCustomerAccess(ctx, cloudAccountId, vip)
	if err != nil {
		log.Error(err, "error getting existing access", logkeys.CloudAccountId, cloudAccountId)
		return err
	}

	err = r.Provider.SyncFirewallRules(ctx, rules, existingRules, vip, cloudAccountId)
	if err != nil {
		log.Error(err, "error syncing ports for firewall rule", logkeys.CloudAccountId, fwRule.Namespace)
		return err
	}

	return nil
}

func (r *FirewallRuleReconciler) PersistStatusUpdate(ctx context.Context, firewallRule *v1alpha1.FirewallRule, namespacedName types.NamespacedName) error {
	log := log.FromContext(ctx).WithName("FirewallRuleReconciler.PersistStatusUpdate")
	log.Info("BEGIN")

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestFWRule, err := r.GetFirewallRule(ctx, namespacedName)
		// firewallRule is deleted
		if latestFWRule == nil {
			log.Info("firewallRule not found", "firewallRule", firewallRule)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get the firewallRule: %+v. error:%w", firewallRule, err)
		}
		if !equality.Semantic.DeepEqual(firewallRule.Status, latestFWRule.Status) {
			log.Info("firewallRule status mismatch", "firewallRule status", firewallRule.Status, "latest firewallRule status", latestFWRule.Status)
			// update latest fwrule status
			firewallRule.Status.DeepCopyInto(&latestFWRule.Status)
			if err := r.Client.Status().Update(ctx, latestFWRule); err != nil {
				return fmt.Errorf("PersistStatusUpdate: %w", err)
			}
		} else {
			log.Info("firewallRule status does not need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update firewallRule status: %w", err)
	}
	log.Info("END")
	return nil
}

// Get firewallrule from K8s.
// Returns (nil, nil) if not found.
func (r *FirewallRuleReconciler) GetFirewallRule(ctx context.Context, namespacedName types.NamespacedName) (*v1alpha1.FirewallRule, error) {
	firewallRule := &v1alpha1.FirewallRule{}
	err := r.Client.Get(ctx, namespacedName, firewallRule)
	if errors.IsNotFound(err) || reflect.ValueOf(firewallRule).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetFirewallRule: %w", err)
	}
	return firewallRule, nil
}

type finalizerOperation string

const (
	add    finalizerOperation = "add"
	remove finalizerOperation = "remove"
)

func (r *FirewallRuleReconciler) PersistFinalizer(ctx context.Context, operation finalizerOperation, firewallRule v1alpha1.FirewallRule) error {
	log := log.FromContext(ctx).WithName("FirewallRuleReconciler.PersistFinalizer")
	log.Info("BEGIN")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {

		latestFWRule, err := r.GetFirewallRule(ctx, types.NamespacedName{
			Name: firewallRule.Name, Namespace: firewallRule.Namespace,
		})
		// firewallRule is deleted
		if latestFWRule == nil {
			log.Info("firewallRule not found", "firewallRule", firewallRule)
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get the firewallRule: %+v. error:%w", firewallRule, err)
		}

		switch operation {
		case add:
			if controllerutil.ContainsFinalizer(latestFWRule, FirewallFinalizer) {
				return nil
			}
			controllerutil.AddFinalizer(latestFWRule, FirewallFinalizer)
		case remove:
			if !controllerutil.ContainsFinalizer(latestFWRule, FirewallFinalizer) {
				return nil
			}
			controllerutil.RemoveFinalizer(latestFWRule, FirewallFinalizer)
		}

		// Use the Client to update since this resource.
		if err := r.Client.Update(ctx, latestFWRule); err != nil {
			return fmt.Errorf("PersistFinalizer: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update fwRule: %w", err)
	}
	log.Info("END")
	return nil
}

// HasControllerReference returns true if the object
// has an owner ref with controller equal to true
func HasControllerReference(object metav1.Object) bool {
	owners := object.GetOwnerReferences()
	for _, owner := range owners {
		isTrue := owner.Controller
		if owner.Controller != nil &&
			*isTrue &&
			owner.Kind == "Loadbalancer" {
			return true
		}
	}
	return false
}
