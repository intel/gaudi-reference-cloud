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
	"strings"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/addon_provider/kubectl"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/pkg/errors"
)

const (
	addonControllerCertCN = "iks:addon-controller"
	addonControllerCertO  = "iks:addon-controller"
)

// AddonReconciler reconciles a Addon object
type AddonReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	*Config
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=addons,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=addons/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=addons/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Addon object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AddonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("AddonReconciler.Reconcile")
	log.V(0).Info("Starting")
	defer log.V(0).Info("Stopping")

	// Get addon custom resource.
	var addon privatecloudv1alpha1.Addon
	if err := r.Get(ctx, req.NamespacedName, &addon); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Add finalizer.
	if addon.DeletionTimestamp.IsZero() && (!controllerutil.ContainsFinalizer(&addon, deleteAddonFinilizer)) {
		controllerutil.AddFinalizer(&addon, deleteAddonFinilizer)
		if err := r.Update(ctx, &addon); err != nil {
			return ctrl.Result{}, err
		}
	}

	// If addon is a template hosted in S3, we prepare the configuration for the S3 client.
	var s3AddonConfig kubectl.S3AddonConfig
	if strings.HasPrefix(addon.Spec.Artifact, "s3://") {
		log.V(0).Info("Creating Information for s3 Addons with URL", logkeys.AddonsURL, r.Config.S3Addons.URL)
		s3AddonConfig.AccessKey = r.Config.S3Addons.AccessKey
		s3AddonConfig.BucketName = r.Config.S3Addons.BucketName
		s3AddonConfig.S3Path = r.Config.S3Addons.S3Path
		s3AddonConfig.SecretKey = r.Config.S3Addons.SecretKey
		s3AddonConfig.URL = r.Config.S3Addons.URL
		s3AddonConfig.UseSSL = r.Config.S3Addons.UseSSL
	}

	// Delete addon.
	if !addon.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&addon, deleteAddonFinilizer) {
			log.V(0).Info("Deleting addon")

			caCertb, caKeyb, err := getKubernetesCACertKey(ctx, r.Client, addon.Spec.ClusterName, req.Namespace)
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
			}

			// If cluster secret is not found, we don't connect to cluster.
			// This can happen when a cluster is deleted and its secret is deleted before
			// addons.
			if err == nil {
				cert, key, err := getKubernetesClientCerts(caCertb, caKeyb, addonControllerCertCN, addonControllerCertO, r.Config.CertExpirations.ControllerCertExpirationPeriod)
				if err != nil {
					return ctrl.Result{}, err
				}

				provider, err := NewAddonProvider(string(addon.Spec.Type),
					utils.GetKubernetesRestConfig(
						fmt.Sprintf("https://%s:%s", addon.Spec.APIServerLB, addon.Spec.APIServerLBPort),
						addonControllerCertCN,
						caCertb,
						cert,
						key), s3AddonConfig)
				if err != nil {
					return ctrl.Result{}, err
				}

				if err := provider.Delete(ctx, &addon); err != nil {
					return ctrl.Result{}, err
				}
			}

			// Remove Finalizer and Updated Add on Resource
			if err := r.removeAddonFinalizerAndUpdateResource(ctx, req.NamespacedName); err != nil {
				return ctrl.Result{}, errors.Wrapf(err, "Remove Finalizer and Update Addon Resource")
			}
		}

		log.V(0).Info("Addon deleted")
		return ctrl.Result{}, nil
	}

	caCertb, caKeyb, err := getKubernetesCACertKey(ctx, r.Client, addon.Spec.ClusterName, req.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	cert, key, err := getKubernetesClientCerts(caCertb, caKeyb, addonControllerCertCN, addonControllerCertO, r.Config.CertExpirations.ControllerCertExpirationPeriod)
	if err != nil {
		return ctrl.Result{}, err
	}

	provider, err := NewAddonProvider(string(addon.Spec.Type),
		utils.GetKubernetesRestConfig(
			fmt.Sprintf("https://%s:%s", addon.Spec.APIServerLB, addon.Spec.APIServerLBPort),
			addonControllerCertCN,
			caCertb,
			cert,
			key), s3AddonConfig)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Currently we do not observe the state of the addon, we just deploy it
	// if artifact changed.
	observedAddonStatus := addon.Status

	// Reconcile States.
	recError := r.reconcileStates(ctx, addon, &observedAddonStatus, provider)

	// Update addon status.
	if err := r.updateState(ctx, req.NamespacedName, &observedAddonStatus); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, recError
}

func (r *AddonReconciler) updateState(ctx context.Context, key k8stypes.NamespacedName, observedAddonStatus *privatecloudv1alpha1.AddonStatus) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var addon privatecloudv1alpha1.Addon
		if err := r.Get(ctx, key, &addon); err != nil {
			return err
		}

		addon.Status = *observedAddonStatus
		if err := r.Status().Update(ctx, &addon); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *AddonReconciler) removeAddonFinalizerAndUpdateResource(ctx context.Context, key k8stypes.NamespacedName) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var addon privatecloudv1alpha1.Addon
		if err := r.Get(ctx, key, &addon); err != nil {
			return err
		}

		// Delete finalizer from addon object
		controllerutil.RemoveFinalizer(&addon, deleteAddonFinilizer)
		if err := r.Update(ctx, &addon); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *AddonReconciler) reconcileStates(ctx context.Context, addon privatecloudv1alpha1.Addon, observedAddonStatus *privatecloudv1alpha1.AddonStatus, provider addonProvider) error {
	log := log.FromContext(ctx).WithName("AddonReconciler.reconcileStates")

	if addon.Spec.Artifact != observedAddonStatus.Artifact || observedAddonStatus.State == privatecloudv1alpha1.ErrorAddonState {
		log.V(0).Info("Deploying addon", logkeys.DesiredArtifact, addon.Spec.Artifact, logkeys.CurrentArtifact, observedAddonStatus.Artifact, logkeys.AddonState, observedAddonStatus.State)
		observedAddonStatus.Name = addon.Name
		observedAddonStatus.LastUpdate = metav1.Time{Time: time.Now()}
		observedAddonStatus.Artifact = addon.Spec.Artifact
		observedAddonStatus.State = privatecloudv1alpha1.UpdatingAddonState
		observedAddonStatus.Reason = ""
		observedAddonStatus.Message = ""

		if err := provider.Put(ctx, &addon); err != nil {
			observedAddonStatus.State = privatecloudv1alpha1.ErrorAddonState
			observedAddonStatus.Reason = "DeployedFailed"
			observedAddonStatus.Message = err.Error()

			return err
		}

		log.V(0).Info("Addon successfully deployed", logkeys.Artifact, addon.Spec.Artifact)
		observedAddonStatus.State = privatecloudv1alpha1.ActiveAddonState
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AddonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&privatecloudv1alpha1.Addon{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.AddonMaxConcurrentReconciles,
		}).
		Complete(r)
}
