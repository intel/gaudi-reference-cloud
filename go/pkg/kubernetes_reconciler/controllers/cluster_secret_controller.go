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

package controllers

import (
	"context"
	"fmt"

	iks "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_reconciler/iks"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"encoding/base64"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	clusterSecretControllerLabelKey   = "controller"
	clusterSecretControllerLabelValue = "cluster"
	clusterSecretServiceLabelKey      = "service"
	clusterSecretServiceLabelValue    = "iks"
)

// ClusterReconciler reconciles a Cluster object
type ClusterSecretReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	IKSAPIClient *iks.Client
	Config       *Config
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
func (r *ClusterSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("ClusterSecretReconciler.Reconcile")
	log.V(0).Info("Pushing cluster secret data")

	// Get cluster secret object
	var clusterSecret v1.Secret
	if err := r.Get(ctx, req.NamespacedName, &clusterSecret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	caCert, ok := clusterSecret.Data["ca.crt"]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("ca.crt not found")
	}

	caKey, ok := clusterSecret.Data["ca.key"]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("ca.key not found")
	}

	etcdCaCert, ok := clusterSecret.Data["etcd-ca.crt"]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("etcd-ca.crt not found")
	}

	etcdCaKey, ok := clusterSecret.Data["etcd-ca.key"]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("etcd-ca.key not found")
	}

	saKey, ok := clusterSecret.Data["sa.key"]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("sa.key not found")
	}

	saPub, ok := clusterSecret.Data["sa.pub"]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("sa.pub not found")
	}

	controlplaneRegistrationCmd, ok := clusterSecret.Data["controlplane-registration-cmd"]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("controlplane-registration-cmd not found")
	}

	workerRegistrationCmd, ok := clusterSecret.Data["worker-registration-cmd"]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("worker-registration-cmd not found")
	}

	etcdEncryptionConfigs, ok := clusterSecret.Data["etcd-encryption-configs"]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("etcd-encryption-configs not found")
	}

	if err := r.IKSAPIClient.UpdateClusterSecret(ctx,
		clusterSecret.Name,
		base64.StdEncoding.EncodeToString(caCert),
		base64.StdEncoding.EncodeToString(caKey),
		base64.StdEncoding.EncodeToString(etcdCaCert),
		base64.StdEncoding.EncodeToString(etcdCaKey),
		base64.StdEncoding.EncodeToString(saKey),
		base64.StdEncoding.EncodeToString(saPub),
		base64.StdEncoding.EncodeToString(controlplaneRegistrationCmd),
		base64.StdEncoding.EncodeToString(workerRegistrationCmd),
		base64.StdEncoding.EncodeToString(etcdEncryptionConfigs)); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "update cluster secret %s", clusterSecret.Name)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	selectorPredicate, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchLabels: map[string]string{
			clusterSecretControllerLabelKey: clusterSecretControllerLabelValue,
			clusterSecretServiceLabelKey:    clusterSecretServiceLabelValue,
		},
	})
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Secret{}, builder.WithPredicates(selectorPredicate)).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Config.ClusterSecretMaxConcurrentReconciles,
		}).
		Complete(r)
}
