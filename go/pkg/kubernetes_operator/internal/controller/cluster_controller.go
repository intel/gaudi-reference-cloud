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
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"encoding/base64"
	"encoding/json"

	"sigs.k8s.io/yaml"

	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	helmaddonprovider "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/addon_provider/helm"
	kubectladdonprovider "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/addon_provider/kubectl"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	apiserverv1 "k8s.io/apiserver/pkg/apis/config/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	fwv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	ilbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ilb_operator/api/v1alpha1"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/etcd"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	controlplaneDesiredCount = 3

	apiserverILBName       = "apiserver"
	publicApiserverILBName = "public-apiserver"
	etcdILBName            = "etcd"
	konnectivityILBName    = "konnectivity"

	customerILBOwner = "customer"
	systemILBOwner   = "system"

	clusterSecretControllerLabelKey   = "controller"
	clusterSecretControllerLabelValue = "cluster"
	clusterSecretServiceLabelKey      = "service"
	clusterSecretServiceLabelValue    = "iks"

	wekaStorageProvider     = "weka"
	wekaStorageSecretName   = "csi-wekafs"
	storageNamespace        = "storage-system"
	vastStorageProvider     = "vast"
	vastStorageSecretName   = "vast-mgmt"
	superComputeClusterType = "supercompute"
)

var (
	ownerKey = ".metadata.controller"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	*Config
	FsOrgPrivateClient      pb.FilesystemOrgPrivateServiceClient
	FsPrivateClient         pb.FilesystemPrivateServiceClient
	StorageKMSPrivateClient pb.StorageKMSPrivateServiceClient
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=clusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=nodegroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=nodegroups/status,verbs=get
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=nodegroups/finalizers,verbs=update
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=addons,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=addons/status,verbs=get
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=addons/finalizers,verbs=update
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=ilbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=ilbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=ilbs/finalizers,verbs=update
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=firewallrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=firewallrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=firewallrules/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Addon object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("ClusterReconciler.Reconcile")
	log.V(0).Info("Starting")
	defer log.V(0).Info("Stopping")

	// Get cluster custom resource that triggered the reconciliation.
	var cluster privatecloudv1alpha1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Add finalizer to ensure everything is cleaned up before deletion of cluster CR.
	if cluster.DeletionTimestamp.IsZero() && (!controllerutil.ContainsFinalizer(&cluster, deleteClusterFinalizer)) {
		controllerutil.AddFinalizer(&cluster, deleteClusterFinalizer)
		if err := r.Update(ctx, &cluster); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Delete the cluster.
	if !cluster.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&cluster, deleteClusterFinalizer) {
			log.V(0).Info("Deleting cluster")
			for _, s := range cluster.Spec.Storage {
				storageProvider := r.NewStorageProvider(ctx, s.Provider)
				if storageProvider == nil {
					return ctrl.Result{}, fmt.Errorf("create storage provider %s", s.Provider)
				}

				log.V(0).Info("Deleting storage", logkeys.Provider, s.Provider)
				if err := storageProvider.DeleteStorage(
					ctx,
					cluster.Name,
					cluster.Spec.CustomerCloudAccountId,
					getWekaPrefix(cluster.Name)); err != nil {
					if grpcstatus.Code(err) != grpccodes.NotFound {
						return ctrl.Result{}, errors.Wrapf(err, "delete storage")
					}
				}
			}

			kubernetesProvider, err := r.getKubernetesProvider(
				ctx,
				false,
				cluster.Name,
				cluster.Namespace,
				"",
				"",
				cluster.Spec.KubernetesProvider)
			if err != nil {
				return ctrl.Result{}, err
			}

			if err = kubernetesProvider.CleanUpCluster(ctx, cluster.Name); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(&cluster, deleteClusterFinalizer)
			if err := r.Update(ctx, &cluster); err != nil {
				return ctrl.Result{}, err
			}
		}

		log.V(0).Info("Cluster deleted")

		return ctrl.Result{}, nil
	}

	// Get status of custom resources created by this cluster.
	var nodegroupList privatecloudv1alpha1.NodegroupList
	if err := r.List(ctx, &nodegroupList, client.InNamespace(req.Namespace), client.MatchingFields{ownerKey: req.Name}); err != nil {
		return ctrl.Result{}, err
	}

	var addonList privatecloudv1alpha1.AddonList
	if err := r.List(ctx, &addonList, client.InNamespace(req.Namespace), &client.MatchingFields{ownerKey: req.Name}); err != nil {
		return ctrl.Result{}, err
	}

	var ilbList ilbv1alpha1.IlbList
	if err := r.List(ctx, &ilbList, client.InNamespace(req.Namespace), client.MatchingFields{ownerKey: req.Name}); err != nil {
		return ctrl.Result{}, err
	}

	var fwList fwv1alpha1.FirewallRuleList
	if err := r.List(ctx, &fwList, client.InNamespace(req.Namespace), client.MatchingFields{ownerKey: req.Name}); err != nil {
		return ctrl.Result{}, err
	}

	observedClusterStatus, err := r.observeCurrentState(ctx, cluster, nodegroupList, addonList, ilbList, fwList, r.FsOrgPrivateClient)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "Observe current state")
	}

	// TODO: All of this should happen in reconcileStates function.
	cluster.Status = observedClusterStatus

	controlplaneNodegroup := getControlplaneNodegroup(nodegroupList)

	// Let's update the cluster status until controlplane is created.
	if controlplaneNodegroup != nil {
		// Cluster state should be equal to controlplane nodegroup, since
		// the controlplane is the actual cluster.
		cluster.Status.State = privatecloudv1alpha1.ClusterState(controlplaneNodegroup.Status.State)

		// If controlplane nodegroup count is not equal to the desired count,
		// cluster should be in updating state.
		// This happens when the controlplane nodegroup instanceIMI is being updated since
		// the cluster controller increases its count from 3 to 4.
		if controlplaneNodegroup.Spec.Count != controlplaneDesiredCount || controlplaneNodegroup.Status.Count != controlplaneDesiredCount {
			cluster.Status.State = privatecloudv1alpha1.UpdatingClusterState
		}

		if err := r.updateClusterStatus(ctx, req.NamespacedName, cluster.Status); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		cluster.Status.State = privatecloudv1alpha1.UpdatingClusterState
		cluster.Status.Message = "Creating controlplane"
		if err := r.updateClusterStatus(ctx, req.NamespacedName, cluster.Status); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Create controlplane ilbs.
	var etcdVIPReady, apiserverVIPReady, publicApiserverVIPReady, konnectivityVIPReady bool
	var etcdVIP, apiserverVIP, publicApiserverVIP, konnectivityVIP string
	var etcdPort, apiserverPort, publicApiserverPort, konnectivityPort int
	for _, ilbSpec := range cluster.Spec.ILBS {
		if ilbSpec.Owner != systemILBOwner {
			continue
		}

		log.V(0).Info("Find or create ilb", logkeys.IlbName, ilbSpec.Name)
		var currentILB ilbv1alpha1.Ilb
		VIPReady, VIP, port, currentILB, err := r.findOrCreateILB(ctx, ilbSpec, cluster, ilbList)
		if err != nil {
			return ctrl.Result{}, errors.WithMessagef(err, "find or create ilb %s", etcdILBName)
		}

		if controlplaneNodegroup != nil && VIPReady {
			log.V(0).Info("Add controlplane nodes into ilb", logkeys.IlbName, ilbSpec.Name)
			if err := r.addNodesAndUpdateILB(ctx, controlplaneNodegroup.Status.Nodes, ilbSpec, currentILB); err != nil {
				return ctrl.Result{}, errors.WithMessagef(err, "add controlplane node to ilb %s", currentILB.Name)
			}
		}

		if ilbSpec.Name == etcdILBName {
			etcdVIPReady = VIPReady
			etcdVIP = VIP
			etcdPort = port
		} else if ilbSpec.Name == apiserverILBName {
			apiserverVIPReady = VIPReady
			apiserverVIP = VIP
			apiserverPort = port
		} else if ilbSpec.Name == konnectivityILBName {
			konnectivityVIPReady = VIPReady
			konnectivityVIP = VIP
			konnectivityPort = port
		} else if ilbSpec.Name == publicApiserverILBName {
			publicApiserverVIPReady = VIPReady
			publicApiserverVIP = VIP
			publicApiserverPort = port
		}
	}

	// VIP ready means that the ILB has an IP that can be used to create controlplane nodes.
	if !etcdVIPReady || !apiserverVIPReady || !konnectivityVIPReady || !publicApiserverVIPReady {
		log.V(0).Info("Ilb VIP not ready", logkeys.EtcdVIP, etcdVIPReady, logkeys.ApiServerVIP, apiserverVIPReady, logkeys.PublicApiServerVIP, publicApiserverVIPReady, logkeys.KonnectivityVIP, konnectivityVIPReady)
		r.updateClusterStatusMessage(ctx, req.NamespacedName, cluster.Status, "Provisioning network", int32(0))

		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Cluster is really the controlplane nodegroup, without controlplane nodegroup there is
	// not a cluster. Cluster state is equal to controlplane nodegroup state.

	// Create controlplane nodegroup.
	if controlplaneNodegroup == nil {
		log.V(0).Info("Creating controlplane")
		r.updateClusterStatusMessage(ctx, req.NamespacedName, cluster.Status, "Provisioning compute", int32(0))

		kubernetesProvider, err := r.getKubernetesProvider(
			ctx,
			false,
			cluster.Name,
			cluster.Namespace,
			"",
			"",
			cluster.Spec.KubernetesProvider)
		if err != nil {
			return ctrl.Result{}, err
		}

		if err := r.initCluster(ctx, &cluster, kubernetesProvider, etcdVIP, apiserverVIP, publicApiserverVIP, konnectivityVIP, etcdPort, apiserverPort, publicApiserverPort); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.createControlplaneNodegroup(ctx,
			strings.Replace(cluster.Name, "cl-", "cp-", 1),
			1,
			&cluster,
			etcdVIP,
			apiserverVIP,
			strconv.Itoa(etcdPort),
			strconv.Itoa(apiserverPort)); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if controlplaneNodegroup.Spec.Count != controlplaneNodegroup.Status.Count {
		log.V(0).Info("Waiting for controlplane desired nodes", logkeys.CurrentNodeCount, controlplaneNodegroup.Status.Count, logkeys.DesiredNodeCount, controlplaneNodegroup.Spec.Count)
		return ctrl.Result{}, nil
	}

	if controlplaneNodegroup.Status.State != privatecloudv1alpha1.ActiveNodegroupState {
		log.V(0).Info("Waiting for controlplane nodegroup to be Active", logkeys.ControlplaneNodegroupState, controlplaneNodegroup.Status.State)
		return ctrl.Result{}, nil
	}

	if controlplaneNodegroup.Spec.Count < controlplaneDesiredCount {
		log.V(0).Info("Adding controlplane node", logkeys.CurrentNodeCount, controlplaneNodegroup.Spec.Count, logkeys.DesiredNodeCount, controlplaneDesiredCount)
		r.updateClusterStatusMessage(ctx, req.NamespacedName, cluster.Status, "Provisioning compute", int32(0))

		controlplaneNodegroup.Spec.Count++
		if err := r.Update(ctx, controlplaneNodegroup); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Upgrade controlplane nodes.
	if controlplaneNodegroup.Spec.InstanceIMI != cluster.Spec.InstanceIMI {
		log.V(0).Info("Updating InstanceIMI", logkeys.CurrentInstanceIMI, controlplaneNodegroup.Spec.InstanceIMI, logkeys.DesiredInstanceIMI, cluster.Spec.InstanceIMI)
		r.updateClusterStatusMessage(ctx, req.NamespacedName, cluster.Status, "Upgrading cluster", int32(0))

		// Get cluster secret and etcd encryption config in use.
		etcdEncryptionConfig, clusterSecret, err := GetEtcdEncryptionConfigFromClusterSecret(
			ctx,
			r.Client,
			req.NamespacedName,
			controlplaneNodegroup.Spec.InstanceIMI)
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "get etcd encryption config and secret")
		}

		// Rotate etcd encryption keys.
		secret, err := GenerateRandomString(32)
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "create new etcd encryption secret")
		}

		rotatedKeys, err := rotateEtcdEncryptionKey(etcdEncryptionConfig, time.Now().Format(time.RFC3339), base64.StdEncoding.EncodeToString(secret))
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "rotate etcd encryption keys")
		}
		rotatedEtcdEncryptionConfig := apiserverv1.EncryptionConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "EncryptionConfiguration",
				APIVersion: "apiserver.config.k8s.io/v1",
			},
			Resources: []apiserverv1.ResourceConfiguration{
				{
					Resources: []string{
						"secrets",
					},
					Providers: []apiserverv1.ProviderConfiguration{
						{
							AESCBC: &apiserverv1.AESConfiguration{
								Keys: rotatedKeys,
							},
						},
						{
							Identity: &apiserverv1.IdentityConfiguration{},
						},
					},
				},
			},
		}

		// Add new etcd encryption config into cluster secret, let's also use the new instanceIMI.
		etcdEncryptionConfigs, err := addEtcdEncryptionConfigInClusterSecret(cluster.Spec.InstanceIMI, rotatedEtcdEncryptionConfig, clusterSecret)
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update etcd encryption configs")
		}

		etcdEncryptionConfigsBytes, err := json.Marshal(etcdEncryptionConfigs)
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "can not marshal etcd encryption configs")
		}

		clusterSecret.Data["etcd-encryption-configs"] = etcdEncryptionConfigsBytes

		if err := r.Update(ctx, clusterSecret); err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "update cluster secret with rotated etcd encryption config")
		}

		// Update controlplane instanceIMI and count. We add a fourth node as a way to
		// safely update controlplane nodes.
		controlplaneNodegroup.Spec.InstanceIMI = cluster.Spec.InstanceIMI
		controlplaneNodegroup.Spec.Count = controlplaneDesiredCount + 1

		return ctrl.Result{}, r.Update(ctx, controlplaneNodegroup)
	}

	// If upgrade is in progress.
	if controlplaneNodegroup.Spec.Count == (controlplaneDesiredCount + 1) {
		// Ensure we have all four controlplane nodes before doing the instanceIMI check.
		if controlplaneNodegroup.Status.Count != (controlplaneDesiredCount + 1) {
			return ctrl.Result{}, nil
		}

		// Ensure controlplane nodes have the correct instanceIMI.
		for _, node := range controlplaneNodegroup.Status.Nodes {
			if node.InstanceIMI != cluster.Spec.InstanceIMI {
				return ctrl.Result{}, nil
			}
		}

		// If upgrade is done, let's go back to three nodes.
		controlplaneNodegroup.Spec.Count = controlplaneDesiredCount
		return ctrl.Result{}, r.Update(ctx, controlplaneNodegroup)
	}

	// Wait until controlplane nodegroup has the desired default controlplane count.
	if controlplaneNodegroup.Status.Count != controlplaneDesiredCount {
		return ctrl.Result{}, nil
	}

	// Wait until controlplane nodegroup instanceIMI is set in all nodes.
	for _, node := range controlplaneNodegroup.Status.Nodes {
		if node.InstanceIMI != controlplaneNodegroup.Spec.InstanceIMI {
			return ctrl.Result{}, nil
		}
	}

	// compare the storage request between whats exisiting and whats in the new CRD
	// if the size of the storage is different for the cluster ID, call the updateFileSystemOrg command

	// Enable storage.
	for _, s := range cluster.Spec.Storage {
		storageProvider := r.NewStorageProvider(ctx, s.Provider)
		if storageProvider == nil {
			return ctrl.Result{}, fmt.Errorf("create storage provider %s", s.Provider)
		}

		storageStatus := getStorageStatus(s.Provider, &cluster)
		if storageStatus == nil {
			log.V(0).Info("Storage not found in status", logkeys.Provider, s.Provider)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}

		// Create storage what really is doing is initializing the storage org / namespace
		// and reserving some storage space.

		// Convert both sizes to TB
		currentStorageSize := convertToTB(storageStatus.Size)
		if currentStorageSize == -1 {
			return ctrl.Result{}, errors.New("Invalid current storage size")
		}
		newStorageSize := convertToTB(s.Size)
		if newStorageSize == -1 {
			return ctrl.Result{}, errors.New("Invalid new storage size")
		}

		if !storageStatus.NamespaceCreated {
			log.V(0).Info("Creating storage", logkeys.StorageProvider, s.Provider, logkeys.StorageSize, s.Size)

			if len(cluster.Spec.VNETS) < 1 {
				return ctrl.Result{}, errors.New("VNETS not found")
			}

			if len(cluster.Spec.CustomerCloudAccountId) == 0 {
				return ctrl.Result{}, errors.New("CustomerCloudAccountId not found")
			}

			if err := storageProvider.CreateStorage(
				ctx,
				cluster.Spec.VNETS[0].AvailabilityZone,
				cluster.Name,
				cluster.Spec.CustomerCloudAccountId,
				s.Size,
				getWekaPrefix(cluster.Name),
				cluster.Spec.ClusterType); err != nil {
				return ctrl.Result{}, errors.Wrapf(err, "enable storage")
			}

			return ctrl.Result{RequeueAfter: time.Second}, nil

		} else if storageStatus.State == privatecloudv1alpha1.ActiveStorageState && currentStorageSize != newStorageSize {
			// call updateStorage api
			log.V(0).Info("Updating Storage", logkeys.CloudAccountId, cluster.Spec.CustomerCloudAccountId, logkeys.Provider, s.Provider, logkeys.StorageSize, s.Size)
			if err := storageProvider.UpdateStorage(
				ctx,
				cluster.Spec.VNETS[0].AvailabilityZone,
				cluster.Name,
				cluster.Spec.CustomerCloudAccountId,
				s.Size,
				getWekaPrefix(cluster.Name),
				cluster.Spec.ClusterType); err != nil {
				return ctrl.Result{}, errors.Wrapf(err, "update storage")
			}

			return ctrl.Result{RequeueAfter: time.Second}, nil
		} else {
			// Recreating storage if it did not get in Active state at least once.
			timeSinceCreation := time.Since(storageStatus.CreatedAt.Time)
			if storageStatus.NamespaceState != privatecloudv1alpha1.ActiveStorageState &&
				timeSinceCreation > r.Config.Weka.RecreateGracePeriod &&
				storageStatus.ActiveAt.IsZero() {
				log.V(0).Info("Recreating storage", logkeys.StorageProvider, s.Provider, logkeys.StorageNamespaceState, storageStatus.NamespaceState, logkeys.TimeInSecondsSinceCreation, timeSinceCreation.Seconds())
				if err := storageProvider.DeleteStorage(ctx, cluster.Name, cluster.Spec.CustomerCloudAccountId, getWekaPrefix(cluster.Name)); err != nil {
					return ctrl.Result{}, errors.Wrapf(err, "delete storage to recreate it")
				}

				return ctrl.Result{RequeueAfter: time.Second}, nil
			}
		}

		if storageStatus.NamespaceState == privatecloudv1alpha1.ActiveStorageState && !storageStatus.SecretCreated {
			// Create weka secret
			if s.Provider == wekaStorageProvider {
				log.V(0).Info("Creating weka secret", logkeys.StorageProvider, s.Provider)
				filesystemOrg, err := r.FsOrgPrivateClient.GetFilesystemOrgPrivate(ctx, &pb.FilesystemOrgGetRequestPrivate{
					Metadata: &pb.FilesystemMetadataReference{
						NameOrId: &pb.FilesystemMetadataReference_Name{
							Name: cluster.Name,
						},
						CloudAccountId: cluster.Spec.CustomerCloudAccountId,
					},
				})
				if err != nil {
					return ctrl.Result{}, errors.Wrapf(err, "get filesystem org to read credentials path")
				}

				filesystemOrgCredentials, err := r.StorageKMSPrivateClient.Get(ctx, &pb.GetSecretRequest{
					KeyPath: filesystemOrg.Spec.Scheduler.Namespace.CredentialsPath,
				})
				if err != nil {
					return ctrl.Result{}, errors.Wrapf(err, "Get storage credentials")
				}

				var namespace v1.Namespace
				namespace.Name = storageNamespace

				kubernetesProvider, err := r.getKubernetesProvider(
					ctx,
					true,
					cluster.Name,
					cluster.Namespace,
					apiserverVIP,
					strconv.Itoa(apiserverPort),
					cluster.Spec.KubernetesProvider)
				if err != nil {
					return ctrl.Result{}, err
				}

				if err := kubernetesProvider.CreateNamespace(ctx, &namespace); err != nil {
					if !k8serrors.IsAlreadyExists(err) {
						return ctrl.Result{}, errors.Wrapf(err, "Create storage namespace")
					}
				}

				clusterEndpoint := filesystemOrg.Spec.Scheduler.Cluster.ClusterAddr
				clusterEndpointSplit := strings.Split(clusterEndpoint, ":")
				if len(clusterEndpointSplit) == 1 {
					clusterEndpoint = clusterEndpoint + ":" + r.Config.Weka.ClusterPort
				}
				log.V(0).Info("Creating storage Secret Request", logkeys.ClusterEndpoint, clusterEndpoint)

				var storageSecret v1.Secret
				storageSecret.Name = wekaStorageSecretName
				storageSecret.Namespace = storageNamespace
				storageSecret.Type = v1.SecretTypeOpaque
				storageSecret.StringData = map[string]string{
					"username":           filesystemOrgCredentials.Secrets["username"],
					"password":           filesystemOrgCredentials.Secrets["password"],
					"organization":       storageStatus.NamespaceName,
					"endpoints":          clusterEndpoint,
					"scheme":             r.Config.Weka.Scheme,
					"localContainerName": "",
				}

				if err := kubernetesProvider.CreateSecret(ctx, storageNamespace, &storageSecret); err != nil {
					return ctrl.Result{}, errors.Wrapf(err, "Create storage secret")
				}

			} else if s.Provider == vastStorageProvider {
				// Check if the cluster name starts with "cl-"
				var fsName string
				if strings.HasPrefix(cluster.Name, "cl-") {
					// Remove the prefix
					fsName = strings.TrimPrefix(cluster.Name, "cl-")
				} else {
					fsName = cluster.Name
				}

				// Create vast secret
				log.V(0).Info("Creating vast secret", logkeys.StorageProvider, s.Provider)
				filesystemCred, err := r.FsPrivateClient.CreateorGetUserPrivate(ctx, &pb.FilesystemGetUserRequestPrivate{
					Metadata: &pb.FilesystemMetadataReference{
						NameOrId: &pb.FilesystemMetadataReference_Name{
							Name: fsName,
						},
						CloudAccountId: cluster.Spec.CustomerCloudAccountId,
					},
				})
				if err != nil {
					return ctrl.Result{}, errors.Wrapf(err, "CreateorGetUserPrivate to read credentials error")
				}

				// Fetch the Cluster Endpoint
				filesystemPriv, err := r.FsPrivateClient.GetPrivate(ctx, &pb.FilesystemGetRequestPrivate{
					Metadata: &pb.FilesystemMetadataReference{
						NameOrId: &pb.FilesystemMetadataReference_Name{
							Name: fsName,
						},
						CloudAccountId: cluster.Spec.CustomerCloudAccountId,
					},
				})
				if err != nil {
					return ctrl.Result{}, errors.Wrapf(err, "Get Private error")
				}

				clusterEndpoint := filesystemPriv.Status.ClusterInfo["VastVMSEndpoint"]
				fmt.Println("Getting the endpoing for vast secret", clusterEndpoint)

				var namespace v1.Namespace
				namespace.Name = storageNamespace

				kubernetesProvider, err := r.getKubernetesProvider(
					ctx,
					true,
					cluster.Name,
					cluster.Namespace,
					apiserverVIP,
					strconv.Itoa(apiserverPort),
					cluster.Spec.KubernetesProvider)
				if err != nil {
					return ctrl.Result{}, err
				}

				if err := kubernetesProvider.CreateNamespace(ctx, &namespace); err != nil {
					if !k8serrors.IsAlreadyExists(err) {
						return ctrl.Result{}, errors.Wrapf(err, "Create storage namespace")
					}
				}

				log.V(0).Info("Creating vast storage Secret Request")
				var storageSecret v1.Secret
				storageSecret.Name = vastStorageSecretName
				storageSecret.Namespace = storageNamespace
				storageSecret.Type = v1.SecretTypeOpaque
				storageSecret.StringData = map[string]string{
					"username": filesystemCred.User,
					"password": filesystemCred.Password,
					"endpoint": clusterEndpoint,
				}

				if err := kubernetesProvider.CreateSecret(ctx, storageNamespace, &storageSecret); err != nil {
					return ctrl.Result{}, errors.Wrapf(err, "Create storage secret")
				}
			}
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}

		if storageStatus.State != privatecloudv1alpha1.ActiveStorageState {
			log.V(0).Info("Waiting for storage provider to be active", logkeys.StorageProvider, s.Provider, logkeys.StorageState, storageStatus.State)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
	}

	// Create, delete and update addons.
	// Get desired and current addons as a map for easy look up.
	r.updateClusterStatusMessage(ctx, req.NamespacedName, cluster.Status, "Reconciling addons", int32(0))
	desiredAddons := make(map[string]privatecloudv1alpha1.AddonTemplateSpec, 0)
	for _, addon := range cluster.Spec.Addons {
		desiredAddons[cluster.Name+"-"+addon.Name] = addon
	}

	currentAddons := make(map[string]privatecloudv1alpha1.Addon, 0)
	for _, addon := range addonList.Items {
		currentAddons[addon.Name] = addon
	}

	if err := r.createAddons(ctx, desiredAddons, currentAddons, &cluster, apiserverVIP, apiserverPort, konnectivityVIP, konnectivityPort); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.updateAddons(ctx, desiredAddons, currentAddons); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.deleteAddons(ctx, desiredAddons, currentAddons); err != nil {
		return ctrl.Result{}, err
	}

	// Create, delete and update nodegroups.
	// Get desired and current nodegroups as a map for easy look up.
	r.updateClusterStatusMessage(ctx, req.NamespacedName, cluster.Status, "Reconciling worker nodegroups", int32(0))
	desiredNodegroups := make(map[string]privatecloudv1alpha1.NodegroupTemplateSpec, 0)
	for _, nodegroup := range cluster.Spec.Nodegroups {
		desiredNodegroups[nodegroup.Name] = nodegroup
	}

	currentNodegroups := make(map[string]privatecloudv1alpha1.Nodegroup, 0)
	for _, nodegroup := range nodegroupList.Items {
		currentNodegroups[nodegroup.Name] = nodegroup
	}

	if err := r.createNodegroups(ctx, desiredNodegroups, currentNodegroups, &cluster, etcdVIP, apiserverVIP, strconv.Itoa(etcdPort), strconv.Itoa(apiserverPort)); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.updateNodegroups(ctx, desiredNodegroups, currentNodegroups, &cluster); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.deleteNodegroups(ctx, desiredNodegroups, currentNodegroups); err != nil {
		return ctrl.Result{}, err
	}

	// Create, delete and update worker ilbs.
	// Get desired and current ilbs as a map for easy look up.
	r.updateClusterStatusMessage(ctx, req.NamespacedName, cluster.Status, "Reconciling customer ILBs", int32(0))
	desiredILBS := make(map[string]privatecloudv1alpha1.ILBTemplateSpec, 0)
	for _, ilb := range cluster.Spec.ILBS {
		if ilb.Owner == customerILBOwner {
			desiredILBS[cluster.Name+"-"+ilb.Name] = ilb
		}
	}

	currentILBS := make(map[string]ilbv1alpha1.Ilb, 0)
	for _, ilb := range ilbList.Items {
		if ilb.Spec.Owner == customerILBOwner {
			currentILBS[ilb.Name] = ilb
		}
	}

	if err := r.createILBS(ctx, desiredILBS, currentILBS, &cluster); err != nil {
		return ctrl.Result{}, err
	}

	var nodes []privatecloudv1alpha1.NodeStatus
	for _, nodegroup := range nodegroupList.Items {
		if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.WorkerNodegroupType {
			nodes = append(nodes, nodegroup.Status.Nodes...)
		}
	}
	if err := r.updateILBS(ctx, desiredILBS, currentILBS, nodes); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.deleteILBS(ctx, desiredILBS, currentILBS, &cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Create, delete and update nodegroups.
	// Get desired and current nodegroups as a map for easy look up.
	r.updateClusterStatusMessage(ctx, req.NamespacedName, cluster.Status, "Reconciling Firewall Rules", int32(0))
	desiredFws := make(map[string]privatecloudv1alpha1.FirewallSpec, 0)
	for _, fw := range cluster.Spec.Firewall {
		desiredFws[fw.DestinationIp] = fw
	}

	currentFws := make(map[string]fwv1alpha1.FirewallRule, 0)
	for _, fw := range fwList.Items {
		currentFws[fw.Spec.DestinationIP] = fw
	}

	if err := r.createFwRule(ctx, desiredFws, currentFws, &cluster); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.updateFwRule(ctx, desiredFws, currentFws); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.deleteFwRule(ctx, desiredFws, currentFws, &cluster); err != nil {
		return ctrl.Result{}, err
	}

	r.updateClusterStatusMessage(ctx, req.NamespacedName, cluster.Status, "Cluster ready", int32(0))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.S3Snapshots.Enabled {
		go etcdSnapshots(context.Background(), mgr.GetClient(), r.Config)
	}

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&privatecloudv1alpha1.Nodegroup{},
		ownerKey,
		func(rawObj client.Object) []string {
			nodegroup := rawObj.(*privatecloudv1alpha1.Nodegroup)
			owner := metav1.GetControllerOf(nodegroup)

			if owner == nil {
				return nil
			}

			return []string{owner.Name}
		}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&privatecloudv1alpha1.Addon{},
		ownerKey,
		func(rawObj client.Object) []string {
			addon := rawObj.(*privatecloudv1alpha1.Addon)
			owner := metav1.GetControllerOf(addon)

			if owner == nil {
				return nil
			}

			return []string{owner.Name}
		}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&ilbv1alpha1.Ilb{},
		ownerKey,
		func(rawObj client.Object) []string {
			ilb := rawObj.(*ilbv1alpha1.Ilb)
			owner := metav1.GetControllerOf(ilb)

			if owner == nil {
				return nil
			}

			return []string{owner.Name}
		}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&fwv1alpha1.FirewallRule{},
		ownerKey,
		func(rawObj client.Object) []string {
			fw := rawObj.(*fwv1alpha1.FirewallRule)
			owner := metav1.GetControllerOf(fw)

			if owner == nil {
				return nil
			}

			return []string{owner.Name}
		}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1.Secret{},
		ownerKey,
		func(rawObj client.Object) []string {
			secret := rawObj.(*v1.Secret)
			owner := metav1.GetControllerOf(secret)

			if owner == nil {
				return nil
			}

			return []string{owner.Name}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&privatecloudv1alpha1.Cluster{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&privatecloudv1alpha1.Nodegroup{}).
		Owns(&privatecloudv1alpha1.Addon{}).
		Owns(&ilbv1alpha1.Ilb{}).
		Owns(&fwv1alpha1.FirewallRule{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.ClusterMaxConcurrentReconciles,
		}).
		Complete(r)
}

func etcdSnapshots(ctx context.Context, c client.Client, config *Config) {
	log := log.FromContext(ctx).WithName("etcdSnapshots")

	for {
		// This is here to wait for cache
		time.Sleep(config.S3Snapshots.SnapshotsPeriodicity)
		log.V(0).Info("etcd-snapshots: starting etcd snapshots for all clusters")

		var clusterList privatecloudv1alpha1.ClusterList
		if err := c.List(ctx, &clusterList, &client.ListOptions{}); err != nil {
			log.Error(err, "etcd-snapshots: get clusters")
			continue
		}

		for _, cluster := range clusterList.Items {
			if cluster.Status.State != privatecloudv1alpha1.ActiveClusterState {
				continue
			}

			log.V(0).Info("etcd-snapshots: creating etcd snapshot", logkeys.ClusterName, cluster.Name)
			// Retrieve the cluster etcd LB IP and port
			ilbSpecList := cluster.Spec.ILBS
			ilbStatusList := cluster.Status.ILBS
			var etcdIlbPort string
			var etcdIlbIp string

			for _, ilbSpec := range ilbSpecList {
				if ilbSpec.Name == "etcd" {
					etcdIlbPort = strconv.Itoa(ilbSpec.Port)
				}
			}
			for _, ilbStatus := range ilbStatusList {
				if strings.HasSuffix(ilbStatus.Name, "etcd") {
					etcdIlbIp = ilbStatus.Vip
				}
			}
			log.V(0).Info("etcd-snapshots: cluster etcd ILB ip and port", logkeys.EtcdIlbIp, etcdIlbIp, logkeys.EtcdIlbPort, etcdIlbPort, logkeys.ClusterName, cluster.Name)

			go func(clusterName, clusterNamespace string) {
				// Get etcd client
				expirationPeriod := time.Now().Add(config.CertExpirations.ControllerCertExpirationPeriod)
				etcdClientCert, etcdClientKey, etcdCACert, err := getEtcdCertificates(ctx, c, clusterName, clusterNamespace, &expirationPeriod)
				if err != nil {
					log.Error(err, "etcd-snapshots: get etcd client certificate", logkeys.ClusterName, clusterName)
					return
				}

				etcdClient, err := etcd.NewClient(etcdIlbIp, etcdIlbPort, etcdClientCert, etcdClientKey, etcdCACert)
				if err != nil {
					log.Error(err, "etcd-snapshots: create etcd client", logkeys.ClusterName, clusterName)
					return
				}
				defer etcdClient.Close()

				// Create etcd snapshot locally
				etcdLogger := etcdClient.GetLogger()
				ts := time.Now().UTC().Format(time.RFC3339)
				etcdSnapshotsPath := config.S3Snapshots.SnapshotsFolder + clusterName + fmt.Sprintf("-%v.db", strings.Replace(ts, ":", "", -1))
				log.V(0).Info("etcd-snapshots: saving etcd snapshot", logkeys.EtcdSnapshotsPath, etcdSnapshotsPath, logkeys.ClusterName, clusterName)
				if err := etcdClient.EtcdSnapshot(ctx, etcdLogger, etcdSnapshotsPath); err != nil {
					log.Error(err, "etcd-snapshots: saving etcd snapshot", logkeys.ClusterName, clusterName)
					return
				}

				// Compress etcd snapshot
				files := []string{etcdSnapshotsPath}
				etcdSnapshotCompressedFileName := clusterName + fmt.Sprintf("-%v.tar.gz", strings.Replace(ts, ":", "", -1))
				etcdSnapshotCompressedPath := config.S3Snapshots.SnapshotsFolder + clusterName + fmt.Sprintf("-%v.tar.gz", strings.Replace(ts, ":", "", -1))
				out, err := os.Create(etcdSnapshotCompressedPath)
				if err != nil {
					log.Error(err, "etcd-snapshots: create archive path", logkeys.ClusterName, clusterName)
					return
				}
				defer out.Close()

				if err = createArchive(files, out); err != nil {
					log.Error(err, "etcd-snapshots: create archive", logkeys.ClusterName, clusterName)
					return
				}
				log.V(3).Info("etcd-snapshots: archived and compressed snaphost successfully", logkeys.EtcdSnapshotsPath, etcdSnapshotCompressedPath, logkeys.ClusterName, clusterName)

				log.V(3).Info("etcd-snapshots: deleting uncompressed snaphost file", logkeys.EtcdSnapshotsPath, etcdSnapshotsPath, logkeys.ClusterName, clusterName)
				if err = os.Remove(etcdSnapshotsPath); err != nil {
					log.Error(err, "etcd-snapshots: delete uncompressed snaphost file", logkeys.ClusterName, clusterName)
					return
				}

				// Upload the compressed etcd snapshot to S3
				log.V(3).Info("etcd-snapshots: uploading the compressed snapshot to S3 endpoint", logkeys.ClusterName, clusterName)
				// Initialize minio client object.
				minioClient, err := minio.New(config.S3Snapshots.URL, &minio.Options{
					Creds:  credentials.NewStaticV4(config.S3Snapshots.AccessKey, config.S3Snapshots.SecretKey, ""),
					Secure: config.S3Snapshots.UseSSL,
				})
				if err != nil {
					log.Error(err, "etcd-snapshots: setup the S3 client", logkeys.ClusterName, clusterName)
					return
				}

				// Upload the file with FPutObject
				info, err := minioClient.FPutObject(ctx, config.S3Snapshots.BucketName, config.S3Snapshots.S3Path+clusterName+"/"+etcdSnapshotCompressedFileName, etcdSnapshotCompressedPath, minio.PutObjectOptions{ContentType: config.S3Snapshots.ContentType})
				if err != nil {
					log.Error(err, "etcd-snapshots: upload the file to S3", logkeys.ClusterName, clusterName)
					return
				}
				log.V(3).Info("etcd-snapshots: successfully uploaded the file to S3", logkeys.UploadInfoSize, info.Size, logkeys.ClusterName, clusterName)

				log.V(3).Info("etcd-snapshots: deleting the compressed snapshot file", logkeys.ClusterName, clusterName)
				err = os.Remove(etcdSnapshotCompressedPath)
				if err != nil {
					log.Error(err, "etcd-snapshots: delete compressed snaphost file", logkeys.ClusterName, clusterName)
					return
				}
			}(cluster.Name, cluster.Namespace)
		}
	}
}

func (r *ClusterReconciler) updateClusterStatus(ctx context.Context, key k8stypes.NamespacedName, clusterStatus privatecloudv1alpha1.ClusterStatus) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var cluster privatecloudv1alpha1.Cluster
		if err := r.Get(ctx, key, &cluster); err != nil {
			return err
		}

		cluster.Status = clusterStatus
		cluster.Status.LastUpdate = metav1.Time{Time: time.Now()}
		if err := r.Status().Update(ctx, &cluster); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterReconciler) updateClusterStatusMessage(ctx context.Context, key k8stypes.NamespacedName, clusterStatus privatecloudv1alpha1.ClusterStatus, message string, errorcode int32) {
	log := log.FromContext(ctx).WithName("ClusterReconciler.updateClusterStatusMessage")

	operatorMessage := OperatorMessage{
		Message:   message,
		ErrorCode: errorcode,
	}
	operatorMessageString, err := json.Marshal(operatorMessage)
	if err != nil {
		log.Error(err, "could not update nodegroup status while Marshalling", logkeys.Message, message)
	}

	clusterStatus.Message = string(operatorMessageString)
	if err := r.updateClusterStatus(ctx, key, clusterStatus); err != nil {
		log.Error(err, "could not update cluster status message", logkeys.Message, message)
	}
}

func (r *ClusterReconciler) createControlplaneNodegroup(ctx context.Context, name string, count int, cluster *privatecloudv1alpha1.Cluster, etcdLB string, apiserverLB string, etcdLBPort string, apiserverLBPort string) error {
	var nodegroup privatecloudv1alpha1.Nodegroup

	nodegroup.Name = name
	nodegroup.ObjectMeta.Namespace = cluster.Namespace
	nodegroup.ObjectMeta.Annotations = cluster.Annotations

	nodegroup.Spec.InstanceIMI = cluster.Spec.InstanceIMI
	nodegroup.Spec.ClusterName = cluster.Name
	nodegroup.Spec.NodegroupType = privatecloudv1alpha1.ControlplaneNodegroupType
	nodegroup.Spec.Count = count
	nodegroup.Spec.KubernetesProvider = cluster.Spec.KubernetesProvider
	nodegroup.Spec.KubernetesVersion = cluster.Spec.KubernetesVersion
	nodegroup.Spec.NodeProvider = cluster.Spec.NodeProvider
	nodegroup.Spec.SSHKey = cluster.Spec.SSHKey
	nodegroup.Spec.CloudAccountId = cluster.Spec.CloudAccountId
	nodegroup.Spec.InstanceType = cluster.Spec.InstanceType
	nodegroup.Spec.ClusterType = cluster.Spec.ClusterType
	nodegroup.Spec.Region = cluster.Spec.Network.Region
	nodegroup.Spec.VNETS = cluster.Spec.VNETS
	nodegroup.Spec.EtcdLB = etcdLB
	nodegroup.Spec.APIServerLB = apiserverLB
	nodegroup.Spec.EtcdLBPort = etcdLBPort
	nodegroup.Spec.APIServerLBPort = apiserverLBPort
	nodegroup.Spec.ContainerRuntime = ""
	nodegroup.Spec.ContainerRuntimeArgs = make(map[string]string)

	controllerutil.AddFinalizer(&nodegroup, deleteNodesFinalizer)

	if err := ctrl.SetControllerReference(cluster, &nodegroup, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, &nodegroup); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func (r *ClusterReconciler) initCluster(ctx context.Context, cluster *privatecloudv1alpha1.Cluster, kubernetesProvider kubernetesProvider, etcdLB string, apiserverLB string, publicApiserverLB string, konnectivityLB string, etcdPort int, apiserverPort int, publicApiserverLBPort int) error {
	var clusterSecret v1.Secret
	clusterSecret.Name = cluster.Name
	clusterSecret.Namespace = cluster.Namespace
	clusterSecret.Type = v1.SecretTypeOpaque
	clusterSecret.Labels = map[string]string{
		clusterSecretControllerLabelKey: clusterSecretControllerLabelValue,
		clusterSecretServiceLabelKey:    clusterSecretServiceLabelValue,
	}
	clusterSecret.Data = make(map[string][]byte, 0)

	if err := kubernetesProvider.InitCluster(ctx, &clusterSecret, cluster, etcdLB, apiserverLB, publicApiserverLB, konnectivityLB, etcdPort, apiserverPort, publicApiserverLBPort); err != nil {
		return err
	}

	// Create etcd encryption configuration.
	// We create two keys the first time a cluster is being created, so that we can
	// use the second one as encryption key when controlplane nodes are being rotated and since
	// existing controlplane nodes already have second key they can decrypt as well. This is
	// because we can not ssh into existing nodes and rotate the key for encryption.
	secret1, err := GenerateRandomString(32)
	if err != nil {
		return errors.WithMessagef(err, "can not create first etcd encryption secret")
	}

	secret2, err := GenerateRandomString(32)
	if err != nil {
		return errors.WithMessagef(err, "can not create second etcd encryption secret")
	}

	etcdEncryptionConfigBytes, err := yaml.Marshal(apiserverv1.EncryptionConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EncryptionConfiguration",
			APIVersion: "apiserver.config.k8s.io/v1",
		},
		Resources: []apiserverv1.ResourceConfiguration{
			{
				Resources: []string{
					"secrets",
				},
				Providers: []apiserverv1.ProviderConfiguration{
					{
						AESCBC: &apiserverv1.AESConfiguration{
							Keys: []apiserverv1.Key{
								{
									Name:   time.Now().Format(time.RFC3339),
									Secret: base64.StdEncoding.EncodeToString(secret1),
								},
								{
									// We add 5m to the name of the second key just to differentiate them.
									Name:   time.Now().Add(time.Minute * 5).Format(time.RFC3339),
									Secret: base64.StdEncoding.EncodeToString(secret2),
								},
							},
						},
					},
					{
						Identity: &apiserverv1.IdentityConfiguration{},
					},
				},
			},
		},
	})
	if err != nil {
		return errors.Wrapf(err, "marshal etcd encryption config")
	}

	etcdEncryptionConfigs, err := json.Marshal(map[string]string{
		cluster.Spec.InstanceIMI: string(etcdEncryptionConfigBytes),
	})
	if err != nil {
		return errors.Wrapf(err, "marshal etcd encryption configs")
	}

	clusterSecret.Data["etcd-encryption-configs"] = etcdEncryptionConfigs

	if err := ctrl.SetControllerReference(cluster, &clusterSecret, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, &clusterSecret); err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func (r *ClusterReconciler) createILBS(ctx context.Context, desiredILBS map[string]privatecloudv1alpha1.ILBTemplateSpec, currentILBS map[string]ilbv1alpha1.Ilb, cluster *privatecloudv1alpha1.Cluster) error {
	for ilbName, ilbTemplate := range desiredILBS {
		if _, found := currentILBS[ilbName]; !found {
			var ilb ilbv1alpha1.Ilb
			ilb.Name = ilbName
			ilb.Namespace = cluster.Namespace

			ilb.Spec.Owner = ilbTemplate.Owner

			ilb.Spec.VIP.Name = ilbName + "-" + strconv.Itoa(ilbTemplate.Port)
			ilb.Spec.VIP.Description = ilbTemplate.Description
			ilb.Spec.VIP.Port = ilbTemplate.Port
			ilb.Spec.VIP.IPType = ilbTemplate.IPType
			ilb.Spec.VIP.Persist = ilbTemplate.Persist
			ilb.Spec.VIP.IPProtocol = ilbTemplate.IPProtocol
			ilb.Spec.VIP.Environment = ilbTemplate.Environment
			ilb.Spec.VIP.UserGroup = ilbTemplate.Usergroup

			ilb.Spec.Pool.Name = cluster.Name + "-" + ilbTemplate.Pool.Name + "-" + strconv.Itoa(ilbTemplate.Pool.Port)
			ilb.Spec.Pool.Description = ilbTemplate.Pool.Description
			ilb.Spec.Pool.Port = ilbTemplate.Pool.Port
			ilb.Spec.Pool.Environment = ilbTemplate.Environment
			ilb.Spec.Pool.UserGroup = ilbTemplate.Usergroup
			ilb.Spec.Pool.LoadBalancingMode = ilbTemplate.Pool.LoadBalancingMode
			ilb.Spec.Pool.MinActiveMembers = ilbTemplate.Pool.MinActiveMembers
			ilb.Spec.Pool.Monitor = ilbTemplate.Pool.Monitor

			if err := ctrl.SetControllerReference(cluster, &ilb, r.Scheme); err != nil {
				return err
			}

			if err := r.Create(ctx, &ilb); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) deleteILBS(ctx context.Context, desiredILBS map[string]privatecloudv1alpha1.ILBTemplateSpec, currentILBS map[string]ilbv1alpha1.Ilb, cluster *privatecloudv1alpha1.Cluster) error {
	for name, ilb := range currentILBS {
		if _, found := desiredILBS[name]; !found {
			if err := r.Delete(ctx, &ilb, &client.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) updateILBS(ctx context.Context, desiredILBS map[string]privatecloudv1alpha1.ILBTemplateSpec, currentILBS map[string]ilbv1alpha1.Ilb, nodes []privatecloudv1alpha1.NodeStatus) error {
	for name, ilb := range currentILBS {
		if desiredILB, found := desiredILBS[name]; found {
			ilb.Spec.Pool.Members = make([]ilbv1alpha1.VMember, 0)
			for _, node := range nodes {
				ilb.Spec.Pool.Members = append(ilb.Spec.Pool.Members, ilbv1alpha1.VMember{
					IP:              node.IpAddress,
					ConnectionLimit: desiredILB.Pool.MemberConnectionLimit,
					PriorityGroup:   desiredILB.Pool.MemberPriorityGroup,
					Ratio:           desiredILB.Pool.MemberRatio,
					AdminState:      desiredILB.Pool.MemberAdminState,
				})
			}

			if err := r.Update(ctx, &ilb); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) createAddons(ctx context.Context, desiredAddons map[string]privatecloudv1alpha1.AddonTemplateSpec, currentAddons map[string]privatecloudv1alpha1.Addon, cluster *privatecloudv1alpha1.Cluster, apiserverLB string, apiserverLBPort int, konnectivityLB string, konnectivityLBPort int) error {
	for addonName, addonTemplate := range desiredAddons {
		if _, found := currentAddons[addonName]; !found {
			var addon privatecloudv1alpha1.Addon
			addon.ObjectMeta.Name = addonName
			addon.ObjectMeta.Namespace = cluster.Namespace
			addon.Spec.ClusterName = cluster.Name
			addon.Spec.Type = addonTemplate.Type
			addon.Spec.Artifact = addonTemplate.Artifact
			addon.Spec.Args = make(map[string]string, 0)
			addon.Spec.APIServerLB = apiserverLB
			addon.Spec.APIServerLBPort = strconv.Itoa(apiserverLBPort)

			// create vast addon
			if addonName == cluster.Name+"-csi-driver" {
				// Check if the cluster name starts with "cl-"
				var fsName string
				if strings.HasPrefix(cluster.Name, "cl-") {
					// Remove the prefix
					fsName = strings.TrimPrefix(cluster.Name, "cl-")
				} else {
					fsName = cluster.Name
				}
				// Fetch the Cluster Endpoint
				filesystemPriv, err := r.FsPrivateClient.GetPrivate(ctx, &pb.FilesystemGetRequestPrivate{
					Metadata: &pb.FilesystemMetadataReference{
						NameOrId: &pb.FilesystemMetadataReference_Name{
							Name: fsName,
						},
						CloudAccountId: cluster.Spec.CustomerCloudAccountId,
					},
				})
				if err != nil {
					return errors.Wrapf(err, "Get Private error")
				}

				clusterEndpoint := filesystemPriv.Status.ClusterInfo["VastVMSEndpoint"]
				fmt.Println("Updating the endpoing in csi-vast-controller deamonset", clusterEndpoint)

				vastStorageVolumePath := "/" + fsName
				viewPolicy := cluster.Spec.CustomerCloudAccountId + "-" + fsName
				addon.Spec.Args = map[string]string{
					"namespace":   storageNamespace,
					"name":        "csi-driver",
					"repoUrl":     "https://vast-data.github.io/vast-csi",
					"endpoint":    clusterEndpoint,
					"storagePath": vastStorageVolumePath,
					"viewPolicy":  viewPolicy,
				}
				fmt.Println("Vast Addon Spec: ", addon.Spec.Args)
			}

			if addonName == cluster.Name+"-"+helmaddonprovider.WekaFsPluginHelmChartName {
				addon.Spec.Args = map[string]string{
					helmaddonprovider.WekaFsPluginNamespaceKey: storageNamespace,
					helmaddonprovider.WekaFsPluginNameKey:      r.Config.Weka.HelmchartName,
					helmaddonprovider.WekaFsPluginRepoUrlKey:   r.Config.Weka.HelmchartRepoUrl,
					helmaddonprovider.WekaFsPluginPrefix:       getWekaPrefix(cluster.Name),
				}
			}

			if addonName == cluster.Name+"-"+kubectladdonprovider.WekaStorageclassTemplateConfigName {
				addon.Spec.Args = map[string]string{
					kubectladdonprovider.WekaStorageclassSecretNameKey:              wekaStorageSecretName,
					kubectladdonprovider.WekaStorageclassSecretNamespaceKey:         storageNamespace,
					kubectladdonprovider.WekaStorageclassFilesystemGroupNameKey:     r.Config.Weka.FilesystemGroupName,
					kubectladdonprovider.WekaStorageclassInitialFilesystemSizeGBKey: r.Config.Weka.InitialFilesystemSizeGB,
					kubectladdonprovider.WekaStorageclassReclaimPolicyKey:           r.Config.Weka.ReclaimPolicy,
				}
			}

			if addonName == cluster.Name+"-"+kubectladdonprovider.KubeProxyTemplateConfigName {
				addon.Spec.Args = map[string]string{
					kubectladdonprovider.KubeProxyClusterCIDRKey: cluster.Spec.Network.PodCIDR,
					kubectladdonprovider.KubeProxyClusterVIPKey:  "https://" + apiserverLB + ":" + strconv.Itoa(apiserverLBPort),
				}
			}

			if addonName == cluster.Name+"-"+kubectladdonprovider.CorednsTemplateConfigName {
				addon.Spec.Args = map[string]string{
					kubectladdonprovider.CorednsClusterIPKey: cluster.Spec.Network.ClusterDNS,
				}
			}

			if addonName == cluster.Name+"-"+kubectladdonprovider.CalicoTemplateConfigName {
				addon.Spec.Args = map[string]string{
					kubectladdonprovider.CalicoClusterCIDRKey: cluster.Spec.Network.PodCIDR,
				}
			}

			if addonName == cluster.Name+"-"+kubectladdonprovider.KonnectivityAgentTemplateConfigName {
				addon.Spec.Args = map[string]string{
					kubectladdonprovider.KonnectivityProxyServerHostKey: konnectivityLB,
					kubectladdonprovider.KonnectivityProxyServerPortKey: strconv.Itoa(konnectivityLBPort),
				}
			}

			controllerutil.AddFinalizer(&addon, deleteAddonFinilizer)

			if err := ctrl.SetControllerReference(cluster, &addon, r.Scheme); err != nil {
				return err
			}

			if err := r.Create(ctx, &addon); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) deleteAddons(ctx context.Context, desiredAddons map[string]privatecloudv1alpha1.AddonTemplateSpec, currentAddons map[string]privatecloudv1alpha1.Addon) error {
	for name, addon := range currentAddons {
		if _, found := desiredAddons[name]; !found {
			if err := r.Delete(ctx, &addon, &client.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) updateAddons(ctx context.Context, desiredAddons map[string]privatecloudv1alpha1.AddonTemplateSpec, currentAddons map[string]privatecloudv1alpha1.Addon) error {
	for name, addon := range currentAddons {
		if desiredaddon, found := desiredAddons[name]; found {
			if desiredaddon.Artifact != addon.Spec.Artifact {
				addon.Spec.Artifact = desiredaddon.Artifact
				if err := r.Update(ctx, &addon); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) createNodegroups(ctx context.Context, desiredNodegroups map[string]privatecloudv1alpha1.NodegroupTemplateSpec, currentNodegroups map[string]privatecloudv1alpha1.Nodegroup, cluster *privatecloudv1alpha1.Cluster, etcdLB string, apiserverLB string, etcdLBPort string, apiserverLBPort string) error {
	log := log.FromContext(ctx).WithName("createNodegroups")
	for nodegroupName, nodegroupTemplate := range desiredNodegroups {
		if _, found := currentNodegroups[nodegroupName]; !found {
			var nodegroup privatecloudv1alpha1.Nodegroup

			nodegroup.ObjectMeta.Name = nodegroupName
			nodegroup.ObjectMeta.Namespace = cluster.Namespace
			nodegroup.ObjectMeta.Annotations = nodegroupTemplate.Annotations
			nodegroup.ObjectMeta.Labels = nodegroupTemplate.Labels

			nodegroup.Spec.ClusterName = cluster.Name
			nodegroup.Spec.NodegroupType = privatecloudv1alpha1.WorkerNodegroupType
			nodegroup.Spec.KubernetesProvider = cluster.Spec.KubernetesProvider
			nodegroup.Spec.NodeProvider = cluster.Spec.NodeProvider
			nodegroup.Spec.Count = nodegroupTemplate.Count
			nodegroup.Spec.InstanceIMI = nodegroupTemplate.InstanceIMI
			nodegroup.Spec.InstanceType = nodegroupTemplate.InstanceType
			nodegroup.Spec.ClusterType = nodegroupTemplate.ClusterType
			nodegroup.Spec.SSHKey = nodegroupTemplate.SSHKey
			nodegroup.Spec.CloudAccountId = nodegroupTemplate.CloudAccountId
			nodegroup.Spec.Region = cluster.Spec.Network.Region
			nodegroup.Spec.VNETS = nodegroupTemplate.VNETS
			nodegroup.Spec.EtcdLB = etcdLB
			nodegroup.Spec.APIServerLB = apiserverLB
			nodegroup.Spec.EtcdLBPort = etcdLBPort
			nodegroup.Spec.APIServerLBPort = apiserverLBPort
			nodegroup.Spec.ContainerRuntime = nodegroupTemplate.ContainerRuntime
			nodegroup.Spec.ContainerRuntimeArgs = nodegroupTemplate.ContainerRuntimeArgs
			nodegroup.Spec.UserDataURL = nodegroupTemplate.UserDataURL

			if storage := getStorageSpec(wekaStorageProvider, cluster); storage != nil {
				if storageStatus := getStorageStatus(wekaStorageProvider, cluster); storageStatus != nil {
					// Check instanceType and enable weka storage for only non vm instancetype
					if storageStatus.State == privatecloudv1alpha1.ActiveStorageState && !strings.Contains(nodegroup.Spec.InstanceType, "vm") {
						nodegroup.Spec.WekaStorage = privatecloudv1alpha1.WekaStorage{
							Enable:    true,
							ClusterId: storageStatus.ClusterId,
							NumCores:  storage.NumCores,
							Mode:      storage.Mode,
						}
					}
				}
			}

			// Check VAST Storage Status and set it accordingly
			if storage := getStorageSpec(vastStorageProvider, cluster); storage != nil {
				if storageStatus := getStorageStatus(vastStorageProvider, cluster); storageStatus != nil {
					// Check instanceType and enable weka storage for only non vm instancetype
					if storageStatus.State == privatecloudv1alpha1.ActiveStorageState {
						nodegroup.Spec.WekaStorage = privatecloudv1alpha1.WekaStorage{
							Enable:    false,
							ClusterId: storageStatus.ClusterId,
							Mode:      "vast",
						}
						log.Info("Setting VAST Storage in Storage Spec", "nodegroupName", nodegroupName, "WekaStorage", nodegroup.Spec.WekaStorage)
					}
				}
			}

			controllerutil.AddFinalizer(&nodegroup, deleteNodesFinalizer)

			if err := ctrl.SetControllerReference(cluster, &nodegroup, r.Scheme); err != nil {
				return err
			}

			if err := r.Create(ctx, &nodegroup); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) deleteNodegroups(ctx context.Context, desiredNodegroups map[string]privatecloudv1alpha1.NodegroupTemplateSpec, currentNodegroups map[string]privatecloudv1alpha1.Nodegroup) error {
	for _, nodegroup := range currentNodegroups {
		if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
			continue
		}

		if _, found := desiredNodegroups[nodegroup.Name]; !found {
			if err := r.Delete(ctx, &nodegroup, &client.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) updateNodegroups(ctx context.Context, desiredNodegroups map[string]privatecloudv1alpha1.NodegroupTemplateSpec, currentNodegroups map[string]privatecloudv1alpha1.Nodegroup, cluster *privatecloudv1alpha1.Cluster) error {
	for _, nodegroup := range currentNodegroups {
		if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
			continue
		}

		if desirednodegroup, found := desiredNodegroups[nodegroup.Name]; found {
			nodegroup.ObjectMeta.Annotations = desirednodegroup.Annotations
			nodegroup.ObjectMeta.Labels = desirednodegroup.Labels

			nodegroup.Spec.Count = desirednodegroup.Count
			nodegroup.Spec.InstanceIMI = desirednodegroup.InstanceIMI
			nodegroup.Spec.InstanceType = desirednodegroup.InstanceType
			nodegroup.Spec.ClusterType = desirednodegroup.ClusterType
			nodegroup.Spec.SSHKey = desirednodegroup.SSHKey

			if storage := getStorageSpec(wekaStorageProvider, cluster); storage != nil {
				if storageStatus := getStorageStatus(wekaStorageProvider, cluster); storageStatus != nil {
					if storageStatus.State == privatecloudv1alpha1.ActiveStorageState && !strings.Contains(nodegroup.Spec.InstanceType, "vm") {
						nodegroup.Spec.WekaStorage = privatecloudv1alpha1.WekaStorage{
							Enable:    true,
							ClusterId: storageStatus.ClusterId,
							NumCores:  storage.NumCores,
							Mode:      storage.Mode,
						}
					}
				}
			} /*else {
				// Initialize weka storage with defaults isnt required as it will reset the storage mode "vast"
				nodegroup.Spec.WekaStorage = privatecloudv1alpha1.WekaStorage{}
			}*/
			if err := r.Update(ctx, &nodegroup); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) findOrCreateILB(ctx context.Context, ilbSpec privatecloudv1alpha1.ILBTemplateSpec, cluster privatecloudv1alpha1.Cluster, ilbList ilbv1alpha1.IlbList) (bool, string, int, ilbv1alpha1.Ilb, error) {
	for _, currentILB := range ilbList.Items {
		if currentILB.Name == cluster.Name+"-"+ilbSpec.Name {
			// Check if VIP IP has been populated.
			// This means VIP is ready.
			if len(currentILB.Status.Vip) > 0 {
				return true, currentILB.Status.Vip, ilbSpec.Port, currentILB, nil
			}
			return false, "", ilbSpec.Port, ilbv1alpha1.Ilb{}, nil
		}
	}

	var ilb ilbv1alpha1.Ilb
	ilb.Name = cluster.Name + "-" + ilbSpec.Name
	ilb.Namespace = cluster.Namespace

	ilb.Spec.VIP.Name = cluster.Name + "-" + ilbSpec.Name + "-" + strconv.Itoa(ilbSpec.Port)
	ilb.Spec.VIP.Description = ""
	ilb.Spec.VIP.Port = ilbSpec.Port
	ilb.Spec.VIP.IPType = ilbSpec.IPType
	ilb.Spec.VIP.Persist = ilbSpec.Persist
	ilb.Spec.VIP.IPProtocol = ilbSpec.IPProtocol
	ilb.Spec.VIP.Environment = ilbSpec.Environment
	ilb.Spec.VIP.UserGroup = ilbSpec.Usergroup
	ilb.Spec.Owner = ilbSpec.Owner

	ilb.Spec.Pool.Name = cluster.Name + "-" + ilbSpec.Pool.Name + "-" + strconv.Itoa(ilbSpec.Pool.Port)
	ilb.Spec.Pool.Description = ilbSpec.Pool.Description
	ilb.Spec.Pool.Port = ilbSpec.Pool.Port
	ilb.Spec.Pool.Environment = ilbSpec.Environment
	ilb.Spec.Pool.UserGroup = ilbSpec.Usergroup
	ilb.Spec.Pool.LoadBalancingMode = ilbSpec.Pool.LoadBalancingMode
	ilb.Spec.Pool.MinActiveMembers = ilbSpec.Pool.MinActiveMembers
	ilb.Spec.Pool.Monitor = ilbSpec.Pool.Monitor

	if err := ctrl.SetControllerReference(&cluster, &ilb, r.Scheme); err != nil {
		return false, "", ilbSpec.Port, ilbv1alpha1.Ilb{}, err
	}

	if err := r.Create(ctx, &ilb); err != nil {
		return false, "", ilbSpec.Port, ilbv1alpha1.Ilb{}, err
	}

	return false, "", ilbSpec.Port, ilbv1alpha1.Ilb{}, nil
}

func (r *ClusterReconciler) addNodesAndUpdateILB(ctx context.Context, nodes []privatecloudv1alpha1.NodeStatus, ilbSpec privatecloudv1alpha1.ILBTemplateSpec, ilb ilbv1alpha1.Ilb) error {
	ilb.Spec.Pool.Members = make([]ilbv1alpha1.VMember, 0)
	for _, node := range nodes {
		// A node could not have an ip address due to errors in the node provider,
		// so if this happens, we shouldn't add the node to the pool.
		if len(node.IpAddress) != 0 {
			ilb.Spec.Pool.Members = append(ilb.Spec.Pool.Members, ilbv1alpha1.VMember{
				IP:              node.IpAddress,
				ConnectionLimit: ilbSpec.Pool.MemberConnectionLimit,
				PriorityGroup:   ilbSpec.Pool.MemberPriorityGroup,
				Ratio:           ilbSpec.Pool.MemberRatio,
				AdminState:      ilbSpec.Pool.MemberAdminState,
			})
		}
	}

	if err := r.Update(ctx, &ilb); err != nil {
		return err
	}

	return nil
}

func (r *ClusterReconciler) createFwRule(ctx context.Context, desiredFWS map[string]privatecloudv1alpha1.FirewallSpec, currentFWS map[string]fwv1alpha1.FirewallRule, cluster *privatecloudv1alpha1.Cluster) error {

	id := uuid.New()
	const cloudAccountIdLabel = "cloud-account-id"
	const componentLabel = "app.kubernetes.io/component"
	const createdLabel = "app.kubernetes.io/created-by"
	const managedbyLabel = "app.kubernetes.io/managed-by"
	const nameLabel = "app.kubernetes.io/name"
	const partLabel = "app.kubernetes.io/part-of"
	const fwcomponentLabel = "app.kubernetes.io/part-of"
	const cp = "control-plane"
	const azid = "availabiltyZoneId"
	const regionid = "regionId"

	log := log.FromContext(ctx).WithName("createFwRule")
	log.Info("creating firewall rules")

	for fwName, fwTemplate := range desiredFWS {
		if _, found := currentFWS[fwName]; !found {
			var fw fwv1alpha1.FirewallRule
			fw.Name = "iks" + cluster.Name + id.String()
			fw.Namespace = cluster.Namespace
			fw.Labels = map[string]string{
				cloudAccountIdLabel: cluster.Spec.CustomerCloudAccountId,
				componentLabel:      "firewall",
				createdLabel:        "loadbalancer-operator",
				managedbyLabel:      "iks-operator",
				nameLabel:           fw.Name,
				partLabel:           "firewall-operator",
				cp:                  "controller-manager",
				azid:                cluster.Spec.VNETS[0].AvailabilityZone,
				regionid:            cluster.Spec.Network.Region,
			}
			fw.Spec.SourceIPs = fwTemplate.SourceIps
			fw.Spec.DestinationIP = fwTemplate.DestinationIp

			if fwTemplate.Protocol == "TCP" {
				fw.Spec.Protocol = fwv1alpha1.Protocol_TCP
			}
			if fwTemplate.Protocol == "UDP" {
				fw.Spec.Protocol = fwv1alpha1.Protocol_UDP
			}
			fw.Spec.Port = strconv.Itoa(fwTemplate.Port)

			if err := ctrl.SetControllerReference(cluster, &fw, r.Scheme); err != nil {
				return err
			}

			if err := r.Create(ctx, &fw); err != nil {
				return err
			}
			log.Info("created firewall rule", "ruleDestinationIp", fwName)
		}
	}

	return nil
}

func (r *ClusterReconciler) deleteFwRule(ctx context.Context, desiredFWS map[string]privatecloudv1alpha1.FirewallSpec, currentFWS map[string]fwv1alpha1.FirewallRule, cluster *privatecloudv1alpha1.Cluster) error {
	log := log.FromContext(ctx).WithName("deleteFwRule")
	log.Info("deleting firewall rules")
	for name, fw := range currentFWS {
		if _, found := desiredFWS[name]; !found {
			log.V(0).Info("trying to delete", "ruleDestinationIp", name)
			if err := r.Delete(ctx, &fw, &client.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
				return err
			}
			log.V(0).Info("rule deleted", "ruleDestinationIp", name)
		}
	}

	return nil
}

func (r *ClusterReconciler) updateFwRule(ctx context.Context, desiredFWS map[string]privatecloudv1alpha1.FirewallSpec, currentFWS map[string]fwv1alpha1.FirewallRule) error {
	log := log.FromContext(ctx).WithName("updateFwRule")
	log.Info("updating firewall rules")
	for name, fw := range currentFWS {
		if desiredFW, found := desiredFWS[name]; found {
			if r.fwRuleChanged(fw, desiredFWS[name]) {

				fw.Spec.SourceIPs = desiredFW.SourceIps
				if desiredFW.Protocol == "TCP" {
					fw.Spec.Protocol = fwv1alpha1.Protocol_TCP
				}
				if desiredFW.Protocol == "UDP" {
					fw.Spec.Protocol = fwv1alpha1.Protocol_UDP
				}
				if err := r.Update(ctx, &fw); err != nil {
					return err
				}
				// Set the status of the Firewall Rule to "Reconciling" since the spec changed. This is to
				// overcome a timing issue that the Firewall Operator can only process a single rule at at time.
				// This can lead the user to think the rule is "Active" when it hasn't yet been processed.
				fw.Status.State = fwv1alpha1.RECONCILING
				if err := r.persistFirewallRuleStatusUpdate(ctx, &fw); err != nil {
					return err
				}
				log.Info("updated firewall rule", "ruleDestinationIp", name)
			} else {
				log.Info("firewall rule didn't changed, will skip the update",
					"ruleDestinationIp", name)
			}
		}
	}

	return nil
}

func (r *ClusterReconciler) fwRuleChanged(currentFwRule fwv1alpha1.FirewallRule,
	desiredFwRule privatecloudv1alpha1.FirewallSpec) bool {
	// check if protocol changed
	if currentFwRule.Spec.Protocol != fwv1alpha1.Protocol(desiredFwRule.Protocol) {
		return true
	}
	// check if port has been changed
	currentPort, err := strconv.Atoi(currentFwRule.Spec.Port)
	if err != nil {
		return true
	}
	if currentPort != desiredFwRule.Port {
		return true
	}
	// check if source ips has been changed
	if len(currentFwRule.Spec.SourceIPs) != len(desiredFwRule.SourceIps) {
		return true
	}
	for _, desiredSrcIp := range desiredFwRule.SourceIps {
		found := false
		for _, currentSrcIp := range currentFwRule.Spec.SourceIPs {
			if desiredSrcIp == currentSrcIp {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	// no changes, no need to update the fw rule
	return false
}

// Update FirewallRule Status.
func (r *ClusterReconciler) persistFirewallRuleStatusUpdate(ctx context.Context, firewallRule *fwv1alpha1.FirewallRule) error {
	log := log.FromContext(ctx).WithName("ClusterReconciler.persistFirewallRuleStatusUpdate")
	log.Info("BEGIN")

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestFirewallRule, err := r.getFirewallRule(ctx, k8stypes.NamespacedName{Name: firewallRule.Name, Namespace: firewallRule.Namespace})
		if err != nil {
			return fmt.Errorf("failed to get the firewallrule: %+v. error:%w", firewallRule, err)
		}
		if latestFirewallRule == nil {
			log.Info("firewallrule not found", logkeys.FirewallRuleName, firewallRule.GetName())
			return nil
		}

		if !equality.Semantic.DeepEqual(firewallRule.Status, latestFirewallRule.Status) {
			// update latest firewallrule status
			firewallRule.Status.DeepCopyInto(&latestFirewallRule.Status)
			if err := r.Status().Update(ctx, latestFirewallRule); err != nil {
				return fmt.Errorf("persistFirewallRuleStatusUpdate: %w", err)
			}
		} else {
			log.Info("firewallrule status does not need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update firewallrule status: %w", err)
	}
	log.Info("END")
	return nil
}

// Get firewallrule from K8s.
// Returns (nil, nil) if not found.
func (r *ClusterReconciler) getFirewallRule(ctx context.Context, namespacedName k8stypes.NamespacedName) (*fwv1alpha1.FirewallRule, error) {
	firewallRule := &fwv1alpha1.FirewallRule{}
	err := r.Get(ctx, namespacedName, firewallRule)
	if k8serrors.IsNotFound(err) || reflect.ValueOf(firewallRule).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getFirewallRule: %w", err)
	}
	return firewallRule, nil
}

func (r *ClusterReconciler) observeCurrentState(ctx context.Context, cluster privatecloudv1alpha1.Cluster, nodegroupList privatecloudv1alpha1.NodegroupList, addonList privatecloudv1alpha1.AddonList, ilbList ilbv1alpha1.IlbList, fwList fwv1alpha1.FirewallRuleList, fsOrgPrivateClient pb.FilesystemOrgPrivateServiceClient) (privatecloudv1alpha1.ClusterStatus, error) {
	log := log.FromContext(ctx).WithName("ClusterReconciler.observeCurrentState")
	var clusterStatus = cluster.Status

	clusterStatus.Nodegroups = make([]privatecloudv1alpha1.NodegroupStatus, 0, len(nodegroupList.Items))
	for _, nodegroup := range nodegroupList.Items {
		clusterStatus.Nodegroups = append(clusterStatus.Nodegroups, nodegroup.Status)
	}

	clusterStatus.Addons = make([]privatecloudv1alpha1.AddonStatus, 0, len(addonList.Items))
	for _, addon := range addonList.Items {
		clusterStatus.Addons = append(clusterStatus.Addons, addon.Status)
	}

	var apiserverLBState ilbv1alpha1.State
	var apiserverVIP string
	var apiserverPort int
	clusterStatus.ILBS = make([]ilbv1alpha1.IlbStatus, 0, len(ilbList.Items))
	for _, ilb := range ilbList.Items {
		clusterStatus.ILBS = append(clusterStatus.ILBS, ilb.Status)

		if ilb.Name == cluster.Name+"-"+apiserverILBName {
			apiserverLBState = ilb.Status.State
			apiserverVIP = ilb.Status.Vip
			apiserverPort = ilb.Spec.VIP.Port
		}
	}

	clusterStatus.Storage = make([]privatecloudv1alpha1.StorageStatus, 0, len(cluster.Spec.Storage))
	for _, s := range cluster.Spec.Storage {

		storageProvider := r.NewStorageProvider(ctx, s.Provider)
		if storageProvider == nil {
			return clusterStatus, errors.New("Invalid storage provider " + s.Provider)
		}

		var oldStorageStatus privatecloudv1alpha1.StorageStatus
		oldStorageStatusPtr := getStorageStatus(s.Provider, &cluster)
		if oldStorageStatusPtr != nil {
			oldStorageStatus = *oldStorageStatusPtr
		}

		storageStatus, err := storageProvider.GetStorage(
			ctx,
			cluster.Name,
			cluster.Spec.CustomerCloudAccountId,
			s,
			oldStorageStatus,
		)
		if err != nil {
			return clusterStatus, errors.Wrapf(err, "Get storage status error")
		}

		// Check if the storage secret exists.
		if apiserverLBState == ilbv1alpha1.READY {
			kubernetesProvider, err := r.getKubernetesProvider(
				ctx,
				true,
				cluster.Name,
				cluster.Namespace,
				apiserverVIP,
				strconv.Itoa(apiserverPort),
				cluster.Spec.KubernetesProvider)
			if err != nil {
				return clusterStatus, err
			}

			if s.Provider == wekaStorageProvider {
				_, err = kubernetesProvider.GetSecret(ctx, wekaStorageSecretName, storageNamespace)
			} else if s.Provider == vastStorageProvider {
				_, err = kubernetesProvider.GetSecret(ctx, vastStorageSecretName, storageNamespace)
			}
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					return clusterStatus, errors.Wrapf(err, "Get storage secret")
				}

				log.V(0).Info("storage secret not found. Controller will create it")
			} else {
				storageStatus.SecretCreated = true

				if storageStatus.NamespaceState == privatecloudv1alpha1.ActiveStorageState {
					storageStatus.State = privatecloudv1alpha1.ActiveStorageState
					/* Set the vast storage status in spec*/
					for i, nodegroup := range nodegroupList.Items {
						if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.WorkerNodegroupType && s.Provider == vastStorageProvider {
							nodegroupList.Items[i].Spec.WekaStorage = privatecloudv1alpha1.WekaStorage{
								Enable:    false,
								ClusterId: storageStatus.ClusterId,
								Mode:      "vast",
							}
							log.Info("Setting VAST Storage in Storage Spec", "nodegroupName", nodegroup.Status.Name, "WekaStorage", nodegroup.Spec.WekaStorage)
						}
					}
				}
			}
		}

		clusterStatus.Storage = append(clusterStatus.Storage, storageStatus)
	}

	// for every firewall rule , FirewallStatus
	var fwstatus privatecloudv1alpha1.FirewallStatus
	clusterStatus.Firewall = make([]privatecloudv1alpha1.FirewallStatus, 0, len(fwList.Items))
	for _, fw := range fwList.Items {
		fwstatus.Firewallrulestatus = fw.Status
		fwstatus.SourceIps = fw.Spec.SourceIPs
		fwstatus.DestinationIp = fw.Spec.DestinationIP
		if fw.Spec.Protocol == fwv1alpha1.Protocol_TCP {
			fwstatus.Protocol = "TCP"
		}
		if fw.Spec.Protocol == fwv1alpha1.Protocol_UDP {
			fwstatus.Protocol = "UDP"
		}
		port, err := strconv.Atoi(fw.Spec.Port)
		if err != nil {
			return clusterStatus, errors.New("Invalid port ")
		}
		fwstatus.Port = port

		clusterStatus.Firewall = append(clusterStatus.Firewall, fwstatus)
	}

	return clusterStatus, nil
}

func getControlplaneNodegroup(nodegroupList privatecloudv1alpha1.NodegroupList) *privatecloudv1alpha1.Nodegroup {
	for _, nodegroup := range nodegroupList.Items {
		if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
			return nodegroup.DeepCopy()
		}
	}

	return nil
}

func rotateEtcdEncryptionKey(encrypConfigString string, keyName string, keySecret string) ([]apiserverv1.Key, error) {
	var encrypConfig apiserverv1.EncryptionConfiguration
	if err := yaml.Unmarshal([]byte(encrypConfigString), &encrypConfig); err != nil {
		return nil, errors.Wrapf(err, "unmarshal etcd encryption configuration")
	}

	if len(encrypConfig.Resources) < 1 {
		return nil, errors.New("there are no resources in encryption config")
	}

	if len(encrypConfig.Resources[0].Providers) < 1 {
		return nil, errors.New("there are no providers in encryption config")
	}

	// Rotation of keys.
	// First key => move as third key.
	// Second key => move as First key.
	// New key => as second key
	keys := encrypConfig.Resources[0].Providers[0].AESCBC.Keys
	rotatedKeys := make([]apiserverv1.Key, 3)
	rotatedKeys[2] = keys[0]
	rotatedKeys[0] = keys[1]
	rotatedKeys[1].Name = keyName
	rotatedKeys[1].Secret = keySecret

	return rotatedKeys, nil
}

func addEtcdEncryptionConfigInClusterSecret(instanceIMI string, rotatedEtcdEncryptionConfig apiserverv1.EncryptionConfiguration, clusterSecret *v1.Secret) (map[string]string, error) {
	etcdEncryptionConfigsBytes, found := clusterSecret.Data[etcdEncryptionConfigsClusterSecretKey]
	if !found {
		return nil, errors.Errorf("can not find %s in cluster secret %s", etcdEncryptionConfigsClusterSecretKey, clusterSecret.Name)
	}

	var etcdEncryptionConfigs map[string]string
	if err := json.Unmarshal(etcdEncryptionConfigsBytes, &etcdEncryptionConfigs); err != nil {
		return nil, errors.Wrapf(err, "unmarshal etcd encription configs")
	}

	rotatedEtcdEncryptionConfigBytes, err := yaml.Marshal(rotatedEtcdEncryptionConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal rotated etcd encryption config")
	}

	etcdEncryptionConfigs[instanceIMI] = string(rotatedEtcdEncryptionConfigBytes)
	// We only store latest three etcd encryption configurations.
	if len(etcdEncryptionConfigs) <= 3 {
		return etcdEncryptionConfigs, nil
	}

	// If we have more than 3 configs, we get the oldest one
	// and we remove it.
	var oldestIMI string
	for IMI := range etcdEncryptionConfigs {
		if len(oldestIMI) == 0 {
			oldestIMI = IMI
			continue
		}

		if IMI < oldestIMI {
			oldestIMI = IMI
		}
	}
	delete(etcdEncryptionConfigs, oldestIMI)

	return etcdEncryptionConfigs, nil
}

func (r *ClusterReconciler) getKubernetesProvider(ctx context.Context, createKubernetesClient bool, clusterName, clusterSecretNamespace, apiServerLB, apiServerLBPort, kubernetesProvider string) (kubernetesProvider, error) {
	var kubernetesClient *kubernetes.Clientset
	if createKubernetesClient {
		caCertb, caKeyb, err := getKubernetesCACertKey(ctx, r.Client, clusterName, clusterSecretNamespace)
		if err != nil {
			return nil, errors.Wrapf(err, "Get CA cert and key from cluster secret")
		}

		// TODO: there should be a cluster role just for the cluster controller.
		cert, key, err := getKubernetesClientCerts(
			caCertb,
			caKeyb,
			nodegroupControllerCertCN,
			nodegroupControllerCertO,
			r.Config.CertExpirations.ControllerCertExpirationPeriod)
		if err != nil {
			return nil, errors.Wrapf(err, "Get kubernetes client certs")
		}

		kubernetesClient, err = utils.GetKubernetesClientFromConfig(
			utils.GetKubernetesRestConfig(
				fmt.Sprintf("https://%s:%s", apiServerLB, apiServerLBPort),
				nodegroupControllerCertCN,
				caCertb,
				cert,
				key))
		if err != nil {
			return nil, errors.Wrapf(err, "Get kubernetes client")
		}
	}

	return newKubernetesProvider(kubernetesProvider, r.Config, false, kubernetesClient)
}

func (r *ClusterReconciler) NewStorageProvider(ctx context.Context, provider string) StorageProvider {
	log := log.FromContext(ctx).WithName("ClusterReconciler.NewStorageProvider")
	switch provider {
	case wekaStorageProvider:
		log.Info("Weka storage provider isn't supported anymore")
		return &WekaStorageProvider{
			FsOrgPrivateClient: r.FsOrgPrivateClient,
		}
	case vastStorageProvider:
		return &VastStorageProvider{
			FsPrivateClient: r.FsPrivateClient,
		}
	default:
		log.Info("Invalid Storage Provider")
	}
	return nil
}

// Storage provider implementation.
type StorageProvider interface {
	CreateStorage(ctx context.Context, availabilityZone string, fsOrgName string, cloudAccountId string, size string, prefix string, clusterType string) error
	DeleteStorage(ctx context.Context, fsOrgName, cloudAccountId, prefix string) error
	UpdateStorage(ctx context.Context, availabilityZone string, fsOrgName string, cloudAccountId string, size string, prefix string, clusterType string) error
	GetStorage(ctx context.Context, fsOrgName string, cloudAccountId string, storage privatecloudv1alpha1.Storage, oldStorageStatus privatecloudv1alpha1.StorageStatus) (privatecloudv1alpha1.StorageStatus, error)
}

type VastStorageProvider struct {
	FsPrivateClient pb.FilesystemPrivateServiceClient
}

func (p *VastStorageProvider) CreateStorage(
	ctx context.Context,
	availabilityZone string,
	fsOrgName string,
	cloudAccountId string,
	size string,
	prefix string,
	clusterType string) error {

	log := log.FromContext(ctx).WithName("VastStorageProvider.CreateStorage")
	log.V(0).Info("Creating vast storage", logkeys.FsOrgName, fsOrgName, logkeys.CloudAccountId, cloudAccountId)

	var fsName string

	// Check if the cluster name starts with "cl-"
	if strings.HasPrefix(fsOrgName, "cl-") {
		// Remove the prefix
		fsName = strings.TrimPrefix(fsOrgName, "cl-")
	} else {
		fsName = fsOrgName
	}
	vastStorageVolumePath := "/" + fsName

	// Create Filesystem
	_, err := p.FsPrivateClient.CreatePrivate(ctx, &pb.FilesystemCreateRequestPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			Name:           fsName,
			CloudAccountId: cloudAccountId,
			SkipQuotaCheck: true,
		},
		Spec: &pb.FilesystemSpecPrivate{
			AvailabilityZone: availabilityZone,
			Request: &pb.FilesystemCapacity{
				Storage: size,
			},
			FilesystemType: pb.FilesystemType_ComputeKubernetes,
			MountProtocol:  pb.FilesystemMountProtocols_NFS,
			Encrypted:      true,
			VolumePath:     vastStorageVolumePath,
			Prefix:         prefix,
			StorageClass:   pb.FilesystemStorageClass_GeneralPurposeStd,
		},
	})
	if err != nil {
		return errors.Wrapf(err, "create Private error")
	}

	return nil
}

func (p *VastStorageProvider) DeleteStorage(ctx context.Context, fsOrgName, cloudAccountId, prefix string) error {
	log := log.FromContext(ctx).WithName("VastStorageProvider.DeleteStorage")
	log.V(0).Info("Deleting vast storage", logkeys.FsOrgName, fsOrgName, logkeys.CloudAccountId, cloudAccountId)

	var fsName string
	// Check if the cluster name starts with "cl-"
	if strings.HasPrefix(fsOrgName, "cl-") {
		// Remove the prefix
		fsName = strings.TrimPrefix(fsOrgName, "cl-")
	} else {
		fsName = fsOrgName
	}

	_, err := p.FsPrivateClient.DeletePrivate(ctx, &pb.FilesystemDeleteRequestPrivate{
		Metadata: &pb.FilesystemMetadataReference{
			NameOrId:       &pb.FilesystemMetadataReference_Name{Name: fsName},
			CloudAccountId: cloudAccountId,
		},
	})

	if err != nil {
		return err
	}
	return nil
}

func (p *VastStorageProvider) UpdateStorage(
	ctx context.Context,
	availabilityZone string,
	fsOrgName string,
	cloudAccountId string,
	size string,
	prefix string,
	clusterType string) error {

	log := log.FromContext(ctx).WithName("VastStorageProvider.UpdateStorage")

	var fsName string
	// Check if the cluster name starts with "cl-"
	if strings.HasPrefix(fsOrgName, "cl-") {
		// Remove the prefix
		fsName = strings.TrimPrefix(fsOrgName, "cl-")
	} else {
		fsName = fsOrgName
	}

	_, err := p.FsPrivateClient.UpdatePrivate(ctx, &pb.FilesystemUpdateRequestPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			Name:           fsName,
			CloudAccountId: cloudAccountId,
			SkipQuotaCheck: true,
		},
		Spec: &pb.FilesystemSpecPrivate{
			AvailabilityZone: availabilityZone,
			Request: &pb.FilesystemCapacity{
				Storage: size,
			},
			FilesystemType: pb.FilesystemType_ComputeKubernetes,
			MountProtocol:  pb.FilesystemMountProtocols_NFS,
			Encrypted:      true,
			Prefix:         prefix,
			StorageClass:   pb.FilesystemStorageClass_GeneralPurpose,
		},
	})
	if err != nil {
		log.Info("Error Message from Storage: ", "message", err.Error())
		if strings.Contains(err.Error(), "only size extension is allowed for file storage") {
			return nil
		}
		return errors.Wrapf(err, "Update filesystem org")
	}
	return nil
}

func (p *VastStorageProvider) GetStorage(
	ctx context.Context,
	fsOrgName string,
	cloudAccountId string,
	storage privatecloudv1alpha1.Storage,
	oldStorageStatus privatecloudv1alpha1.StorageStatus) (privatecloudv1alpha1.StorageStatus, error) {
	log := log.FromContext(ctx).WithName("VastStorageProvider.GetStorage")

	var storageStatus privatecloudv1alpha1.StorageStatus
	storageStatus.Provider = storage.Provider
	storageStatus.State = privatecloudv1alpha1.UpdatingStorageState
	storageStatus.NamespaceState = privatecloudv1alpha1.UpdatingStorageState
	storageStatus.LastUpdate = metav1.Now()
	var fsName string

	// Check if the cluster name starts with "cl-"
	if strings.HasPrefix(fsOrgName, "cl-") {
		// Remove the prefix
		fsName = strings.TrimPrefix(fsOrgName, "cl-")
	} else {
		fsName = fsOrgName
	}

	resp, err := p.FsPrivateClient.GetPrivate(ctx, &pb.FilesystemGetRequestPrivate{
		Metadata: &pb.FilesystemMetadataReference{
			NameOrId: &pb.FilesystemMetadataReference_Name{
				Name: fsName,
			},
			CloudAccountId: cloudAccountId,
		},
	})
	if err != nil {
		if grpcstatus.Code(err) == grpccodes.NotFound {
			return storageStatus, nil
		}
		return storageStatus, errors.Wrapf(err, "Get Private error")
	}
	log.V(0).Info("Vast Storage status", logkeys.StatusPhase, resp.Status.Phase, logkeys.StatusMessage, resp.Status.Message)

	// If storage has been created, we get the status of it.
	storageStatus.NamespaceCreated = true
	storageStatus.CreatedAt = metav1.NewTime(resp.Metadata.CreationTimestamp.AsTime())
	storageStatus.Message = resp.Status.Message
	storageStatus.Size = resp.Spec.Request.Storage

	log.Info("Vast Current storage Size: ", "Current Storage Size", resp.Spec.Request.Storage)
	log.Info("Vast New storage Size: ", "New Storage Size", storage.Size)

	switch resp.Status.Phase {
	case pb.FilesystemPhase_FSProvisioning:
		storageStatus.NamespaceState = privatecloudv1alpha1.UpdatingStorageState
	case pb.FilesystemPhase_FSReady:
		storageStatus.NamespaceState = privatecloudv1alpha1.ActiveStorageState
		storageStatus.NamespaceName = resp.Spec.Scheduler.Namespace.Name
		storageStatus.ClusterId = resp.Spec.Scheduler.Cluster.ClusterUUID

		// Set the time only if not set already.
		if oldStorageStatus.ActiveAt.IsZero() {
			storageStatus.ActiveAt = metav1.Now()
		} else {
			storageStatus.ActiveAt = oldStorageStatus.ActiveAt
		}
	case pb.FilesystemPhase_FSDeleting:
		storageStatus.NamespaceState = privatecloudv1alpha1.DeletingStorageState
	case pb.FilesystemPhase_FSFailed:
		storageStatus.NamespaceState = privatecloudv1alpha1.ErrorStorageState
	default:
		storageStatus.NamespaceState = privatecloudv1alpha1.UpdatingStorageState
	}

	return storageStatus, nil
}

type WekaStorageProvider struct {
	FsOrgPrivateClient pb.FilesystemOrgPrivateServiceClient
}

func (p *WekaStorageProvider) CreateStorage(
	ctx context.Context,
	availabilityZone string,
	fsOrgName string,
	cloudAccountId string,
	size string,
	prefix string,
	clusterType string) error {
	var storageClassFlag pb.FilesystemStorageClass
	if clusterType == superComputeClusterType {
		storageClassFlag = pb.FilesystemStorageClass_AIOptimized
	} else {
		storageClassFlag = pb.FilesystemStorageClass_GeneralPurpose
	}
	_, err := p.FsOrgPrivateClient.CreateFilesystemOrgPrivate(ctx, &pb.FilesystemOrgCreateRequestPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			Name:           fsOrgName,
			CloudAccountId: cloudAccountId,
			SkipQuotaCheck: true,
		},
		Spec: &pb.FilesystemSpecPrivate{
			AvailabilityZone: availabilityZone,
			Request: &pb.FilesystemCapacity{
				Storage: size,
			},
			FilesystemType: pb.FilesystemType_ComputeKubernetes,
			MountProtocol:  pb.FilesystemMountProtocols_Weka,
			Encrypted:      true,
			Prefix:         prefix,
			StorageClass:   storageClassFlag,
		},
	})
	if err != nil {
		return errors.Wrapf(err, "create filesystem org")
	}

	return nil
}

func (p *WekaStorageProvider) DeleteStorage(ctx context.Context, fsOrgName, cloudAccountId, prefix string) error {
	log := log.FromContext(ctx).WithName("WekaStorageProvider.DeleteStorage")
	log.V(0).Info("Deleting weka storage", logkeys.FsOrgName, fsOrgName, logkeys.CloudAccountId, cloudAccountId, logkeys.Prefix, prefix)

	_, err := p.FsOrgPrivateClient.DeleteFilesystemOrgPrivate(ctx, &pb.FilesystemOrgDeleteRequestPrivate{
		Metadata: &pb.FilesystemMetadataReference{
			NameOrId:       &pb.FilesystemMetadataReference_Name{Name: fsOrgName},
			CloudAccountId: cloudAccountId,
		},
		Prefix: prefix,
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *WekaStorageProvider) UpdateStorage(
	ctx context.Context,
	availabilityZone string,
	fsOrgName string,
	cloudAccountId string,
	size string,
	prefix string,
	clusterType string) error {

	log := log.FromContext(ctx).WithName("WekaStorageProvider.UpdateStorage")

	var storageClassFlag pb.FilesystemStorageClass
	if clusterType == superComputeClusterType {
		storageClassFlag = pb.FilesystemStorageClass_AIOptimized
	} else {
		storageClassFlag = pb.FilesystemStorageClass_GeneralPurpose
	}
	_, err := p.FsOrgPrivateClient.UpdateFilesystemOrgPrivate(ctx, &pb.FilesystemOrgUpdateRequestPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			Name:           fsOrgName,
			CloudAccountId: cloudAccountId,
			SkipQuotaCheck: true,
		},
		Spec: &pb.FilesystemSpecPrivate{
			AvailabilityZone: availabilityZone,
			Request: &pb.FilesystemCapacity{
				Storage: size,
			},
			FilesystemType: pb.FilesystemType_ComputeKubernetes,
			MountProtocol:  pb.FilesystemMountProtocols_Weka,
			Encrypted:      true,
			Prefix:         prefix,
			StorageClass:   storageClassFlag,
		},
	})
	if err != nil {
		log.Info("Error Message from Storage: ", "message", err.Error())
		if strings.Contains(err.Error(), "only size extension is allowed for file storage") {
			return nil
		}
		return errors.Wrapf(err, "Update filesystem org")
	}

	return nil
}

func (p *WekaStorageProvider) GetStorage(
	ctx context.Context,
	fsOrgName string,
	cloudAccountId string,
	storage privatecloudv1alpha1.Storage,
	oldStorageStatus privatecloudv1alpha1.StorageStatus) (privatecloudv1alpha1.StorageStatus, error) {
	log := log.FromContext(ctx).WithName("WekaStorageProvider.GetStorage")

	var storageStatus privatecloudv1alpha1.StorageStatus
	storageStatus.Provider = storage.Provider
	storageStatus.State = privatecloudv1alpha1.UpdatingStorageState
	storageStatus.NamespaceState = privatecloudv1alpha1.UpdatingStorageState
	storageStatus.LastUpdate = metav1.Now()

	resp, err := p.FsOrgPrivateClient.GetFilesystemOrgPrivate(ctx, &pb.FilesystemOrgGetRequestPrivate{
		Metadata: &pb.FilesystemMetadataReference{
			NameOrId: &pb.FilesystemMetadataReference_Name{
				Name: fsOrgName,
			},
			CloudAccountId: cloudAccountId,
		},
	})
	if err != nil {
		if grpcstatus.Code(err) == grpccodes.NotFound {
			return storageStatus, nil
		}
		return storageStatus, errors.Wrapf(err, "get filesystem org")
	}
	log.V(0).Info("Storage status", logkeys.StatusPhase, resp.Status.Phase, logkeys.StatusMessage, resp.Status.Message)

	// If storage has been created, we get the status of it.
	storageStatus.NamespaceCreated = true
	storageStatus.CreatedAt = metav1.NewTime(resp.Metadata.CreationTimestamp.AsTime())
	storageStatus.Message = resp.Status.Message
	storageStatus.Size = resp.Spec.Request.Storage

	log.Info("Current storage Size: ", "Current Storage Size", resp.Spec.Request.Storage)
	log.Info("New storage Size: ", "New Storage Size", storage.Size)

	switch resp.Status.Phase {
	case pb.FilesystemPhase_FSProvisioning:
		storageStatus.NamespaceState = privatecloudv1alpha1.UpdatingStorageState
	case pb.FilesystemPhase_FSReady:
		storageStatus.NamespaceState = privatecloudv1alpha1.ActiveStorageState
		storageStatus.NamespaceName = resp.Spec.Scheduler.Namespace.Name
		storageStatus.ClusterId = resp.Spec.Scheduler.Cluster.ClusterUUID

		// Set the time only if not set already.
		if oldStorageStatus.ActiveAt.IsZero() {
			storageStatus.ActiveAt = metav1.Now()
		} else {
			storageStatus.ActiveAt = oldStorageStatus.ActiveAt
		}
	case pb.FilesystemPhase_FSDeleting:
		storageStatus.NamespaceState = privatecloudv1alpha1.DeletingStorageState
	case pb.FilesystemPhase_FSFailed:
		storageStatus.NamespaceState = privatecloudv1alpha1.ErrorStorageState
	default:
		storageStatus.NamespaceState = privatecloudv1alpha1.UpdatingStorageState
	}

	return storageStatus, nil
}

func getStorageStatus(provider string, cluster *privatecloudv1alpha1.Cluster) *privatecloudv1alpha1.StorageStatus {
	for _, s := range cluster.Status.Storage {
		if s.Provider == provider {
			return &s
		}
	}
	return nil
}

func getStorageSpec(provider string, cluster *privatecloudv1alpha1.Cluster) *privatecloudv1alpha1.Storage {
	for _, s := range cluster.Spec.Storage {
		if s.Provider == provider {
			return &s
		}
	}
	return nil
}

// getWekaPrefix generates a prefix using 7 characters of the cluster name.
// We relay on the API to ensure those 7 characters are unique and only one cluster is named with them.
func getWekaPrefix(clusterName string) string {
	return clusterName[3:10]
}
