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
	"encoding/json"

	"time"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	iks "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_reconciler/iks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	IKSKubernetesProviderName = "iks"
	HarvesterProviderName     = "Harvester"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	IKSAPIClient *iks.Client
	Config       *Config
}

// +kubebuilder:rbac:groups=private.cloud.intel.com,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=private.cloud.intel.com,resources=clusters/status,verbs=get
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("ClusterReconciler.Reconcile")

	// Get cluster object
	var cluster privatecloudv1alpha1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		// If cluster not found, it has been deleted and we need to
		// update API to inform about the deletion and stop reconciler
		if client.IgnoreNotFound(err) == nil {
			log.V(0).Info("cluster deleted", logkeys.ClusterName, req.Name)
			if err := r.IKSAPIClient.DeleteCluster(ctx, req.Name); err != nil {
				return ctrl.Result{}, errors.Wrapf(err, "delete cluster %s", req.Name)
			}

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Update cluster status
	if string(cluster.Status.State) != "" {
		log.V(0).Info("pushing cluster status", logkeys.ClusterName, cluster.Name)
		if err := r.IKSAPIClient.UpdateClusterStatus(ctx, cluster.Name, cluster.Status); err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update status for cluster %s", cluster.Name)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// This will poll clusters with changes.
	go r.pollPendingChanges()

	return ctrl.NewControllerManagedBy(mgr).
		For(&privatecloudv1alpha1.Cluster{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Config.ClusterMaxConcurrentReconciles,
		}).
		Complete(r)
}

func (r *ClusterReconciler) pollPendingChanges() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("ClusterReconciler.pollPendingChanges")

	for {
		time.Sleep(r.Config.PollPeriodicity)
		log.V(0).Info("polling clusters")

		// Create or update clusters
		apiClusters, err := r.IKSAPIClient.Get(ctx, iks.PendingRevisionState)
		if err != nil {
			log.Error(err, "get clusters in Pending state")
			continue
		}

		log.V(0).Info("total Pending clusters", logkeys.Total, len(apiClusters))

		for _, apiCluster := range apiClusters {
			var desiredCluster privatecloudv1alpha1.Cluster
			if err := json.Unmarshal([]byte(apiCluster.DesiredJson), &desiredCluster); err != nil {
				log.Error(err, "unmarshal desiredJson into cluster", logkeys.ClusterRevId, apiCluster.ClusterRevId)
			}

			// We're hardcoding namespace to default
			// TODO: Define correct namespace to be used by kubernetes operator.
			desiredCluster.Namespace = defaultNamespace

			clusterFound := true
			var cluster privatecloudv1alpha1.Cluster
			if err = r.Get(ctx, k8stypes.NamespacedName{Name: desiredCluster.Name, Namespace: desiredCluster.Namespace}, &cluster); err != nil {
				if !k8serrors.IsNotFound(err) {
					log.Error(err, "get cluster", logkeys.ClusterName, cluster.Name)
					continue
				}

				clusterFound = false
			}

			if !clusterFound {
				log.V(0).Info("create cluster", logkeys.ClusterName, desiredCluster.Name, logkeys.ClusterNamespace, desiredCluster.Namespace)

				if err := r.Create(ctx, &desiredCluster, &client.CreateOptions{}); err != nil {
					log.Error(err, "create cluster error", logkeys.ClusterName, desiredCluster.Name, logkeys.ClusterNamespace, desiredCluster.Namespace)
					continue
				}

				if err := r.IKSAPIClient.AppliedRevision(ctx, int(apiCluster.ClusterRevId)); err != nil {
					log.Error(err, "set rev change as a applied")
					continue
				}

				if err := r.IKSAPIClient.UpdateClusterState(ctx, desiredCluster.Name, string(privatecloudv1alpha1.UpdatingNodegroupState)); err != nil {
					log.Error(err, "set cluster state to creating")
					continue
				}
			} else {
				log.V(0).Info("update cluster", logkeys.ClusterName, cluster.Name, logkeys.ClusterNamespace, cluster.Namespace)
				cluster.Spec = desiredCluster.Spec
				if err := r.Update(ctx, &cluster, &client.UpdateOptions{}); err != nil {
					log.Error(err, "update cluster", logkeys.ClusterName, cluster.Name, logkeys.ClusterNamespace, cluster.Namespace)
					continue
				}

				if err := r.IKSAPIClient.AppliedRevision(ctx, int(apiCluster.ClusterRevId)); err != nil {
					log.Error(err, "set rev change as a applied")
					continue
				}

				if err := r.IKSAPIClient.UpdateClusterState(ctx, cluster.Name, string(privatecloudv1alpha1.UpdatingNodegroupState)); err != nil {
					log.Error(err, "set cluster state to updating")
					continue
				}
			}
		}

		// Delete clusters
		apiClusters, err = r.IKSAPIClient.Get(ctx, iks.DeletePendingRevisionState)
		if err != nil {
			log.Error(err, "get clusters in Pending state")
			continue
		}

		log.V(0).Info("total DeletePending clusters", logkeys.Total, len(apiClusters))

		for _, apiCluster := range apiClusters {
			var cluster privatecloudv1alpha1.Cluster
			if err := json.Unmarshal([]byte(apiCluster.DesiredJson), &cluster); err != nil {
				log.Error(err, "unmarshal desiredJson into cluster", logkeys.ClusterRevId, apiCluster.ClusterRevId)
			}

			// We're hardcoding namespace to default
			cluster.Namespace = defaultNamespace

			clusterFound := true
			if err = r.Get(ctx, k8stypes.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, &cluster); err != nil {
				if !k8serrors.IsNotFound(err) {
					log.Error(err, "get cluster", logkeys.ClusterName, cluster.Name)
					continue
				}
				log.V(0).Info("delete cluster, cluster not found", logkeys.ClusterName, cluster.Name, logkeys.ClusterRevId, apiCluster.ClusterRevId)
				clusterFound = false
			}

			if clusterFound {
				log.V(0).Info("deleting cluster", logkeys.ClusterName, cluster.Name, logkeys.ClusterNamespace, cluster.Namespace)
				if err := r.Delete(ctx, &cluster, &client.DeleteOptions{}); err != nil {
					log.Error(err, "deleting cluster error", logkeys.ClusterName, cluster.Name, logkeys.ClusterNamespace, cluster.Namespace)
					continue
				}

				if err := r.IKSAPIClient.UpdateClusterState(ctx, cluster.Name, string(privatecloudv1alpha1.DeletingNodegroupState)); err != nil {
					log.Error(err, "set cluster state to deleting")
					continue
				}
			}
		}
	}
}
