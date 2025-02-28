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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/errgroup"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/etcd"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/utils"
)

const (
	nodegroupControllerCertCN = "iks:nodegroup-controller"
	nodegroupControllerCertO  = "iks:nodegroup-controller"
	retryDelay                = 2 * time.Second
	nodeNotFoundReason        = "NotFound"
	storageRegisterLabelKey   = "iks.cloud.intel.com/storage-register"
	wekaDefaultMode           = "disable"
	wekaDefaultNumCores       = "4"
	labelPrefix               = "cloud.intel.com"
)

// NodegroupReconciler reconciles a Nodegroup object
type NodegroupReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	InstanceServiceClient pb.InstanceServiceClient

	InstanceTypeClient pb.InstanceTypeServiceClient
	MachineImageClient pb.MachineImageServiceClient
	*Config
	NoCacheClientReader                   client.Reader
	FilesystemStorageClusterPrivateClient pb.FilesystemStorageClusterPrivateServiceClient
}

// +kubebuilder:rbac:groups=private.cloud.intel.com,resources=nodegroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=private.cloud.intel.com,resources=nodegroups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=private.cloud.intel.com,resources=nodegroups/finalizers,verbs=update
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
func (r *NodegroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx).WithName("NodegroupReconciler.Reconcile")
	log.V(0).Info("Starting")
	defer log.V(0).Info("Stopping")

	// Get nodegroup custom resource that triggered the reconciliation.
	var nodegroup privatecloudv1alpha1.Nodegroup
	if err := r.NoCacheClientReader.Get(ctx, req.NamespacedName, &nodegroup); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Add finalizer to ensure instances or instance groups are deleted before deleting a nodegroup.
	// This is important because otherwise we would leave unmanaged instances in the node provider side.
	if nodegroup.DeletionTimestamp.IsZero() && (!controllerutil.ContainsFinalizer(&nodegroup, deleteNodesFinalizer)) {
		controllerutil.AddFinalizer(&nodegroup, deleteNodesFinalizer)
		if err := r.Update(ctx, &nodegroup); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	// Get provider in charge of managing instances.
	nodeProvider, err := newNodeProvider(nodegroup.Spec.NodeProvider, r.Config)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Get provider in charge of managing the downstream k8s cluster.
	kubernetesProvider, err := r.getKubernetesProvider(ctx, &nodegroup)
	if err != nil {
		return ctrl.Result{}, err
	}

	isInstanceGroup := isInstanceGroup(&nodegroup)

	// Delete instances, instance groups and nodegroup.
	if !nodegroup.DeletionTimestamp.IsZero() {
		if err := r.deleteNodegroup(ctx, req.NamespacedName, nodegroup, nodegroup.Status, nodeProvider, kubernetesProvider); err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "Delete nodegroup")
		}

		return ctrl.Result{}, nil
	}

	// Get the current status of the nodegroup.
	observedNodegroupStatus, err := r.observeCurrentState(ctx, isInstanceGroup, nodegroup, nodeProvider, kubernetesProvider)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "Observe current state")
	}

	// TODO: Use conditions so that we can compare previous status against current one and
	// determine if a condition changed, which means we need to reconcile otherwise stop here.

	// Compare observed nodegroup status against desired spec.
	requeueSooner, recErrors := r.reconcileStates(ctx, isInstanceGroup, nodegroup, &observedNodegroupStatus, nodeProvider, kubernetesProvider)

	// Update nodegroup status.
	if err := r.updateState(ctx, req.NamespacedName, observedNodegroupStatus); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "Update nodegroup status")
	}

	// If during reconciliation something failed we return the error to requeue.
	if recErrors != nil {
		return ctrl.Result{}, errors.Wrapf(recErrors, "Reconciling nodegroup")
	}

	// This controller manages external resources that can not be watched by informers,
	// so we requeue after some time to ensure we get the current status and make proper changes in case
	// something vary from desired spec.
	// If requeueSooner is true, we set the time to 1s to requeue sooner, this happens when a node is deleted,
	// so we reconcile sooner to create a new node in case it is missing.
	requeueAfter := r.MonitorPeriodicity
	if requeueSooner {
		requeueAfter = time.Second
	}
	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodegroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//ch := make(chan event.GenericEvent)
	//go r.monitorNodegroups(context.Background(), r.Config, ch)

	return ctrl.NewControllerManagedBy(mgr).
		For(&privatecloudv1alpha1.Nodegroup{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		//Watches(&source.Channel{Source: ch, DestBufferSize: 1024}, &handler.EnqueueRequestForObject{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.NodegroupMaxConcurrentReconciles,
		}).
		Complete(r)
}

// deleteNode properly deletes a controlplane or worker node from cluster.
func (r *NodegroupReconciler) deleteNode(ctx context.Context, node privatecloudv1alpha1.NodeStatus, nodegroup privatecloudv1alpha1.Nodegroup, nodeProvider nodeProvider, kubernetesProvider kubernetesProvider) error {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.deleteNode")

	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
		if err := removeEtcdMember(ctx, node.IpAddress, r.Client, nodegroup.Spec.EtcdLB, nodegroup.Spec.EtcdLBPort, nodegroup.Spec.ClusterName, nodegroup.Namespace, r.Config.CertExpirations.ControllerCertExpirationPeriod); err != nil {
			if !k8serrors.IsNotFound(err) {
				return errors.WithMessagef(err, "delete controlplane node %s from etcd", node.Name)
			}
		}
	}

	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.WorkerNodegroupType {
		// Deregister worker nodes if weka storage is enabled.
		if nodegroup.Spec.WekaStorage.Enable && node.WekaStorageStatus.ClientId != "" {
			log.V(0).Info("Deregistering weka agent", logkeys.NodeName, node.Name)
			if _, err := r.FilesystemStorageClusterPrivateClient.DeRegisterAgent(ctx, &pb.DeRegisterAgentRequest{
				ClusterId: nodegroup.Spec.WekaStorage.ClusterId,
				ClientId:  node.WekaStorageStatus.ClientId,
			}); err != nil {
				return errors.Wrapf(err, "Deregister weka agent for node %s", node.Name)
			}
		}

		// TODO: this is only cordoning the node, we need to work on the logic to
		// drain it.
		if err := kubernetesProvider.DrainNode(ctx, node.Name); err != nil {
			log.Error(err, "drain worker node", logkeys.NodeName, node.Name)
		}

		if err := kubernetesProvider.DeleteNode(ctx, node.Name); err != nil {
			return errors.WithMessagef(err, "deleting worker node %s from cluster", node.Name)
		}
	}

	if err := nodeProvider.DeleteNode(ctx, node.Name, nodegroup.Spec.CloudAccountId); err != nil {
		if grpcstatus.Code(err) == grpccodes.NotFound {
			log.V(0).Info("Instance not found in node provider", logkeys.NodeName, node.Name)
			return nil
		}

		return errors.WithMessagef(err, "deleting instance %s in node provider", node.Name)
	}

	return nil
}

// deleteInstanceGroupNode properly deletes a instance group worker node from cluster.
func (r *NodegroupReconciler) deleteInstanceGroupNode(ctx context.Context, node privatecloudv1alpha1.NodeStatus, nodegroup privatecloudv1alpha1.Nodegroup, nodeProvider nodeProvider, kubernetesProvider kubernetesProvider, instanceGroupName string) error {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.deleteInstanceGroupNode")

	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.WorkerNodegroupType {
		// Deregister worker nodes if weka storage is enabled.
		if nodegroup.Spec.WekaStorage.Enable && node.WekaStorageStatus.ClientId != "" {
			log.V(0).Info("Deregistering weka agent", logkeys.NodeName, node.Name)
			if _, err := r.FilesystemStorageClusterPrivateClient.DeRegisterAgent(ctx, &pb.DeRegisterAgentRequest{
				ClusterId: nodegroup.Spec.WekaStorage.ClusterId,
				ClientId:  node.WekaStorageStatus.ClientId,
			}); err != nil {
				return errors.Wrapf(err, "Deregister weka agent for node %s", node.Name)
			}
		}

		// TODO: this is only cordoning the node, we need to work on the logic to
		// drain it.
		if err := kubernetesProvider.DrainNode(ctx, node.Name); err != nil {
			log.Error(err, "drain worker node", logkeys.NodeName, node.Name)
		}

		if err := kubernetesProvider.DeleteNode(ctx, node.Name); err != nil {
			return errors.WithMessagef(err, "deleting worker node %s from cluster", node.Name)
		}
	}

	log.V(0).Info("Deleting Instance Group node", "instance group", instanceGroupName, "node", node.Name)
	if err := nodeProvider.DeleteInstanceGroupMember(ctx, node.Name, nodegroup.Spec.CloudAccountId, instanceGroupName); err != nil {
		return err
	}

	return nil
}

func (r *NodegroupReconciler) updateState(ctx context.Context, key k8stypes.NamespacedName, observedNodegroupStatus privatecloudv1alpha1.NodegroupStatus) error {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.updateState")

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var nodegroup privatecloudv1alpha1.Nodegroup
		if err := r.NoCacheClientReader.Get(ctx, key, &nodegroup); err != nil {
			return err
		}
		log.Info("Nodegroup to be updated", logkeys.NodeGroupResourceVersion, nodegroup.ResourceVersion, logkeys.NodegroupGeneration, nodegroup.Generation)

		nodegroup.Status = observedNodegroupStatus
		log.V(0).Info("Updated status", logkeys.NodeGroupStatus, nodegroup.Status)
		if err := r.Status().Update(ctx, &nodegroup); err != nil {
			return err
		}

		return nil
	})
}

func (r *NodegroupReconciler) removeNodegroupFinalizerAndUpdateResource(ctx context.Context, key k8stypes.NamespacedName) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var nodegroup privatecloudv1alpha1.Nodegroup
		if err := r.NoCacheClientReader.Get(ctx, key, &nodegroup); err != nil {
			return err
		}

		controllerutil.RemoveFinalizer(&nodegroup, deleteNodesFinalizer)
		if err := r.Update(ctx, &nodegroup); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *NodegroupReconciler) getKubernetesProvider(ctx context.Context, nodegroup *privatecloudv1alpha1.Nodegroup) (kubernetesProvider, error) {
	// If cluster secret is not found, we can't create the kubernetes client required by the kubernetes provider. This happens
	// when a cluster is deleted and its secret is deleted before the nodegroup. In this case we only delete instances from node provider
	// since cluster doesn't exist anymore.
	var clusterDeleted bool

	caCertb, caKeyb, err := getKubernetesCACertKey(ctx, r.Client, nodegroup.Spec.ClusterName, nodegroup.Namespace)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "Get CA cert and key from cluster secret")
		}

		clusterDeleted = true
	}

	var kubernetesClient *kubernetes.Clientset
	if !clusterDeleted {
		cert, key, err := getKubernetesClientCerts(caCertb, caKeyb, nodegroupControllerCertCN, nodegroupControllerCertO, r.Config.CertExpirations.ControllerCertExpirationPeriod)
		if err != nil {
			return nil, errors.Wrapf(err, "Get kubernetes client certs")
		}

		kubernetesClient, err = utils.GetKubernetesClientFromConfig(
			utils.GetKubernetesRestConfig(
				fmt.Sprintf("https://%s:%s", nodegroup.Spec.APIServerLB, nodegroup.Spec.APIServerLBPort),
				nodegroupControllerCertCN,
				caCertb,
				cert,
				key))
		if err != nil {
			return nil, errors.Wrapf(err, "Get kubernetes client")
		}
	}

	return newKubernetesProvider(nodegroup.Spec.KubernetesProvider, r.Config, clusterDeleted, kubernetesClient)
}

// removeMissingEtcdMember deletes an etcd member that is not in the controlplane node list. Returns true if an
// etcd member was deleted.
func removeMissingEtcdMember(ctx context.Context, nodes []privatecloudv1alpha1.NodeStatus, client client.Client, etcdIP string, etcdPort string, clusterName string, namespace string, controllerCertExpirationPeriod time.Duration) (bool, error) {
	log := log.FromContext(ctx).WithName("removeMissingEtcdMember")

	expirationPeriod := time.Now().Add(controllerCertExpirationPeriod)
	etcdClientCert, etcdClientKey, etcdCACert, err := getEtcdCertificates(ctx, client, clusterName, namespace, &expirationPeriod)
	if err != nil {
		return false, err
	}

	etcdClient, err := etcd.NewClient(etcdIP, etcdPort, etcdClientCert, etcdClientKey, etcdCACert)
	if err != nil {
		return false, err
	}
	defer etcdClient.Close()

	members, err := etcdClient.ListEtcdMembers(ctx)
	if err != nil {
		return false, err
	}
	log.V(0).Info("Etcd members", logkeys.Members, members)

	for _, member := range members {
		var found bool

		if len(member.PeerURLs) == 0 {
			log.V(0).Info("Member does not have peer urls", logkeys.MemberId, member.ID, logkeys.MemberName, member.Name, logkeys.MemberPeerURL, member.PeerURLs)
			continue
		}

		for _, node := range nodes {
			if strings.Contains(member.PeerURLs[0], node.IpAddress) {
				found = true
				break
			}
		}

		if !found {
			log.V(0).Info("Deleting etcd member not found in node list", logkeys.MemberId, member.ID, logkeys.MemberName, member.Name, logkeys.MemberPeerURL, member.PeerURLs)
			return true, etcdClient.RemoveEtcdMember(ctx, member.ID)
		}
	}

	return false, nil
}

// removeEtcdMember removes the desired control plane node from the etcd members.
func removeEtcdMember(ctx context.Context, nodeIP string, client client.Client, etcdIP string, etcdPort string, clusterName string, namespace string, controllerCertExpirationPeriod time.Duration) error {
	log := log.FromContext(ctx).WithName("removeEtcdMember")

	expirationPeriod := time.Now().Add(controllerCertExpirationPeriod)
	etcdClientCert, etcdClientKey, etcdCACert, err := getEtcdCertificates(ctx, client, clusterName, namespace, &expirationPeriod)
	if err != nil {
		return err
	}

	etcdClient, err := etcd.NewClient(etcdIP, etcdPort, etcdClientCert, etcdClientKey, etcdCACert)
	if err != nil {
		return err
	}
	defer etcdClient.Close()

	members, err := etcdClient.ListEtcdMembers(ctx)
	if err != nil {
		return err
	}
	log.V(0).Info("Etcd members", logkeys.Members, members)

	for _, member := range members {
		// First we compare against the name, this works only for started members.
		// Unstarted members will have an empty name.
		if nodeIP == member.Name {
			log.V(0).Info("Delete etcd member found by name", logkeys.MemberId, member.ID, logkeys.MemberName, member.Name)
			return etcdClient.RemoveEtcdMember(ctx, member.ID)
		}

		// PeerURL should look like this https://nodeIP:2380. If member is unstarted,
		// only peerURL can be used to check if it is the member that needs to be
		// deleted.
		if len(member.PeerURLs) > 0 {
			if strings.Contains(member.PeerURLs[0], nodeIP) {
				log.V(0).Info("Delete etcd member found by peer url", logkeys.MemberId, member.ID, logkeys.MemberPeerURL, member.PeerURLs[0])
				return etcdClient.RemoveEtcdMember(ctx, member.ID)
			}
		}
	}

	log.V(0).Info("Can not delete etcd member since it was not found", logkeys.NodeIp, nodeIP)
	return nil
}

// getEtcdCertificates creates a new etcd client cert and key signed with the etcd CA cert and key.
func getEtcdCertificates(ctx context.Context, client client.Client, clusterName string, namespace string, controllerCertExpirationPeriod *time.Time) ([]byte, []byte, []byte, error) {
	etcdCACertb, etcdCAKeyb, err := getEtcdCACertKey(ctx, client, clusterName, namespace)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}

	etcdCACert, err := utils.ParseCert(etcdCACertb)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}

	etcdCAKey, err := utils.ParsePrivateKey(etcdCAKeyb)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}

	etcdClientCert, etcdClientKey, err := utils.CreateAndSignCert(etcdCACert, etcdCAKey, utils.CertConfig{
		CommonName: "operator-etcd-client",
		Organizations: []string{
			"kubernetes-operator",
		},
	}, controllerCertExpirationPeriod)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}

	return etcdClientCert, etcdClientKey, etcdCACertb, nil
}

// getEtcdCACertKey gets the CA cert and key of etcd from the cluster secret.
func getEtcdCACertKey(ctx context.Context, c client.Client, clusterName string, namespace string) ([]byte, []byte, error) {
	var clusterSecret corev1.Secret
	var caCert []byte
	var caKey []byte
	var ok bool

	if err := c.Get(ctx, k8stypes.NamespacedName{
		Namespace: namespace,
		Name:      clusterName,
	}, &clusterSecret); err != nil {
		return caCert, caKey, err
	}

	caCert, ok = clusterSecret.Data["etcd-ca.crt"]
	if !ok {
		return caCert, caKey, fmt.Errorf("can not get etcd-ca.crt from cluster secret: %s", clusterName)
	}

	caKey, ok = clusterSecret.Data["etcd-ca.key"]
	if !ok {
		return caCert, caKey, fmt.Errorf("can not get etcd-ca.key from cluster secret: %s", clusterName)
	}

	return caCert, caKey, nil
}

// getRegistrationCmd gets the registration command from the cluster secret.
func getRegistrationCmd(ctx context.Context, c client.Client, clusterName string, namespace string, nodegroupType privatecloudv1alpha1.NodegroupType) (string, error) {
	var clusterSecret corev1.Secret
	var registrationCmd []byte
	var ok bool

	if err := c.Get(ctx, k8stypes.NamespacedName{
		Namespace: namespace,
		Name:      clusterName,
	}, &clusterSecret); err != nil {
		return string(registrationCmd), err
	}

	if nodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
		if registrationCmd, ok = clusterSecret.Data["controlplane-registration-cmd"]; !ok {
			return string(registrationCmd), fmt.Errorf("can not get controlplane registration cmd from secret: %s", clusterSecret.Name)
		}
	}

	if nodegroupType == privatecloudv1alpha1.WorkerNodegroupType {
		if registrationCmd, ok = clusterSecret.Data["worker-registration-cmd"]; !ok {
			return string(registrationCmd), fmt.Errorf("can not get worker registration cmd from secret: %s", clusterSecret.Name)
		}
	}

	return string(registrationCmd), nil
}

// getWorkerNodeStatus syncs both worker status obtained from node and kubernetes providers.
func getWorkerNodeStatus(ctx context.Context, clusterName string, nodeprovidername string, nodesList []privatecloudv1alpha1.NodeStatus, nodeProvider nodeProvider, kubernetesProvider kubernetesProvider, cloudaccountid string) ([]privatecloudv1alpha1.NodeStatus, error) {
	log := log.FromContext(ctx).WithName("getWorkerNodeStatus")

	if len(nodesList) == 0 {
		return nodesList, nil
	}

	nodes := make([]privatecloudv1alpha1.NodeStatus, len(nodesList))
	for i, n := range nodesList {
		// Get node from node provider.
		node, err := nodeProvider.GetNode(ctx, n.Name, cloudaccountid)
		if err != nil {
			if nodeprovidername == ComputeProviderName && !(grpcstatus.Code(err) == grpccodes.NotFound) {
				return nil, err
			}

			if nodeprovidername == HarvesterProviderName && !k8serrors.IsNotFound(err) {
				return nil, err
			}

			// If node not found, we add it and mark it as error, since we had it in the
			// list of nodes that were created.
			log.V(0).Info("Node not found in node provider", logkeys.NodeName, n.Name)
			node = n
			node.State = privatecloudv1alpha1.ErrorNodegroupState
			node.Message = "node not found"
			node.Reason = nodeNotFoundReason
		}

		// Get node from kubernetes provider only for active nodes.
		kubernetesNode, err := kubernetesProvider.GetNode(ctx, n.Name)
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return nil, err
			}

			node.State = privatecloudv1alpha1.UpdatingNodegroupState
			node.Message = "Checking node"
			node.Reason = "WorkerNotReady"
			nodes[i] = node
			log.V(0).Info("worker node not found", logkeys.NodeName, n.Name)
			continue
		}

		if node.State == privatecloudv1alpha1.ActiveNodegroupState || node.State == privatecloudv1alpha1.ErrorNodegroupState {
			node.State = kubernetesNode.State
		}

		node.Message = kubernetesNode.Message
		node.Reason = kubernetesNode.Reason
		node.LastUpdate = kubernetesNode.LastUpdate
		node.KubeletVersion = kubernetesNode.KubeletVersion
		node.KubeProxyVersion = kubernetesNode.KubeProxyVersion
		node.Unschedulable = kubernetesNode.Unschedulable
		node.AutoRepairDisabled = kubernetesNode.AutoRepairDisabled
		nodes[i] = node
	}

	return nodes, nil
}

// getControlPlaneNodeStatus syncs both controlplane status obtained from node provider and etcd.
func getControlPlaneNodeStatus(ctx context.Context, client client.Client, nodeprovidername string, namespace string, etcdIP string, etcdPort string, clusterName string, nodesList []privatecloudv1alpha1.NodeStatus, nodeProvider nodeProvider, cloudaccountid string, controllerCertExpirationPeriod time.Duration) ([]privatecloudv1alpha1.NodeStatus, error) {
	log := log.FromContext(ctx).WithName("getControlPlaneNodeStatus")

	if len(nodesList) == 0 {
		return nodesList, nil
	}

	// Get etcd client.
	expirationPeriod := time.Now().Add(controllerCertExpirationPeriod)
	etcdClientCert, etcdClientKey, etcdCACert, err := getEtcdCertificates(ctx, client, clusterName, namespace, &expirationPeriod)
	if err != nil {
		return nil, errors.Wrapf(err, "get etcd client certificate")
	}

	etcdClient, err := etcd.NewClient(etcdIP, etcdPort, etcdClientCert, etcdClientKey, etcdCACert)
	if err != nil {
		return nil, errors.Wrapf(err, "create etcd client")
	}
	defer etcdClient.Close()

	// Get members from etcd.
	// We do not return the error because is expected to see
	// communication issues with etcd when adding or deleting members, so we let
	// controller to retry.
	// TODO: we should return an error because we can not determine if a member does not exist
	// using unreal data, only if call succeeds we can check if a member is part of the cluster or not.
	members, err := etcdClient.ListEtcdMembers(ctx)
	if err != nil {
		log.Error(err, "list etcd members")
	}
	log.V(0).Info("List etcd members", logkeys.Members, members)

	nodes := make([]privatecloudv1alpha1.NodeStatus, len(nodesList))
	for i, n := range nodesList {
		node, err := nodeProvider.GetNode(ctx, n.Name, cloudaccountid)
		if err != nil {
			if nodeprovidername == ComputeProviderName && !(grpcstatus.Code(err) == grpccodes.NotFound) {
				return nil, err
			}

			if nodeprovidername == HarvesterProviderName && !k8serrors.IsNotFound(err) {
				return nil, err
			}

			// If node not found, we add it and mark it as error, since we had it in the
			// list of nodes that were created.
			log.V(0).Info("Node not found in node provider", logkeys.NodeName, n.Name)
			node = n
			node.State = privatecloudv1alpha1.ErrorNodegroupState
			node.Message = "Node not found"
			node.Reason = nodeNotFoundReason
		}

		// If node is not active in node provider, we use this as the latest
		// state for the node.
		if node.State != privatecloudv1alpha1.ActiveNodegroupState {
			nodes[i] = node
			continue
		}

		// If active in node provider, we set node state to updating
		// and start checking etcd.
		node.State = privatecloudv1alpha1.UpdatingNodegroupState

		var found bool
		for _, member := range members {
			if n.IpAddress == member.Name {
				found = true
				break
			}
		}

		if !found {
			node.Message = "checking etcd membership"
			node.Reason = "EtcdNotReady"
			nodes[i] = node
			log.V(0).Info("Etcd member not found", logkeys.NodeName, n.Name)
			continue
		}

		// TODO: this has been disabled in staging because of overlay network (CIDR conflict) in cluster
		// where operator runs and compute instances.
		// if err := etcdClient.MemberStatus(ctx, member.ClientURLs[0]); err != nil {
		// 	node.Message = "node not healthy in etcd"
		// 	node.Reason = "EtcdNotReady"
		// 	nodes[i] = node
		// 	log.Error(err, "Get etcd member status", "node", n.Name)
		// 	continue
		// }

		node.State = privatecloudv1alpha1.ActiveNodegroupState
		node.Message = ""
		node.Reason = ""
		nodes[i] = node
	}

	return nodes, nil
}

// getInitialCluster returns the list of existing etcd members in the format
// expected by --initial-cluster etcd flag.
func getInitialCluster(currentNodes []privatecloudv1alpha1.NodeStatus) string {
	var initialCluster string

	for i, node := range currentNodes {
		if i == 0 {
			initialCluster = fmt.Sprintf("%s=https://%s:2380", node.IpAddress, node.IpAddress)
			continue
		}
		initialCluster = initialCluster + fmt.Sprintf(",%s=https://%s:2380", node.IpAddress, node.IpAddress)
	}

	return initialCluster
}

// getNetworkInformation reads the nodegroup annotations and get network information for creating
// a new node.
func getNetworkInformation(annotations map[string]string, currentNodes []privatecloudv1alpha1.NodeStatus) (string, string, string, error) {
	var availableIp string
	var nameserver string
	var gateway string

	// Get an available ip by checking if any of the ips in the annotation is already being used
	// by another node.
	if v, ok := annotations[ipAddressesAnnotation]; ok {
		ips := strings.Split(v, ",")
		for _, ip := range ips {
			add := true
			for _, node := range currentNodes {
				if ip == node.IpAddress {
					add = false
					break
				}
			}

			if add {
				availableIp = ip
				break
			}
		}
	}

	if len(availableIp) == 0 {
		return availableIp, nameserver, gateway, fmt.Errorf("no available ip")
	}

	nameserver, ok := annotations[nameserverAnnotation]
	if !ok {
		return availableIp, nameserver, gateway, fmt.Errorf("no nameserver")
	}

	gateway, ok = annotations[gatewayAnnotation]
	if !ok {
		return availableIp, nameserver, gateway, fmt.Errorf("no gateway")
	}

	return availableIp, nameserver, gateway, nil
}

// getBootstrapTokenData returns the required data to create a bootstrap token that
// is used when creating and joining a new worker node.
func getBootstrapTokenData() (string, string, string, error) {
	rand.NewSource(time.Now().UnixNano())
	var opts = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	b := make([]rune, 22)
	for i := range b {
		b[i] = opts[rand.Intn(len(opts))]
	}

	return string(b[:6]), string(b[6:]), time.Now().UTC().Add(time.Minute * 120).Format(time.RFC3339), nil
}

// monitoringNodegroups is meant to be used as a goroutine to watch current status
// of nodegroups and their nodes and trigger nodegroup reconciliation based on specific
// criteria.
// Trigger reconcile func criteria:
// - current count != desired count
// - node in error or updating state
// func (r *NodegroupReconciler) monitorNodegroups(ctx context.Context, config *Config, clusterEvent chan<- event.GenericEvent) {
// 	log := log.FromContext(ctx)

// 	for {
// 		// This is here to wait for controller cache to be built.
// 		time.Sleep(config.MonitorPeriodicity)

// 		var clusterList privatecloudv1alpha1.ClusterList
// 		if err := r.List(ctx, &clusterList, &client.ListOptions{}); err != nil {
// 			log.Error(err, "Get clusters")
// 			continue
// 		}

// 		for _, cluster := range clusterList.Items {
// 			nodeProvider, err := newNodeProvider(cluster.Spec.NodeProvider, config)
// 			if err != nil {
// 				log.Error(err, "Get node provider", "cluster", cluster.Name)
// 				continue
// 			}

// 			var nodegroupList privatecloudv1alpha1.NodegroupList
// 			if err := r.List(ctx, &nodegroupList, client.InNamespace(cluster.Namespace), client.MatchingFields{ownerKey: cluster.Name}); err != nil {
// 				log.Error(err, "Get nodegroups", "cluster", cluster.Name)
// 				continue
// 			}

// 			for _, nodegroup := range nodegroupList.Items {
// 				log.V(0).Info("Monitoring: checking nodegroup", "cluster", cluster.Name, "nodegroup", nodegroup.Name)

// 				// This will avoid creating more events for the controller. If nodegroup is not active, the controller
// 				// should be reconciling already, so not need to send an event.
// 				// When state is active, the controller is not reconciling a CR, that's when external changes could not
// 				// be known by controller.
// 				if nodegroup.Status.State != privatecloudv1alpha1.ActiveNodegroupState {
// 					log.V(0).Info("Monitoring: nodegroup state is not active, event will not be created")
// 					continue
// 				}

// 				caCertb, caKeyb, err := getKubernetesCACertKey(ctx, r.Client, nodegroup.Spec.ClusterName, cluster.Namespace)
// 				if err != nil {
// 					log.Error(err, "get kubernetes ca cert and key", "cluster", cluster.Name, "nodegroup", nodegroup.Name)
// 					continue
// 				}

// 				cert, key, err := getKubernetesClientCerts(caCertb, caKeyb, nodegroupControllerCertCN, nodegroupControllerCertO, r.Config.CertExpirations.ControllerCertExpirationPeriod)
// 				if err != nil {
// 					log.Error(err, "get kubernetes client certs", "cluster", cluster.Name, "nodegroup", nodegroup.Name)
// 					continue
// 				}

// 				kubernetesClient, err := utils.GetKubernetesClientFromConfig(
// 					utils.GetKubernetesRestConfig(
// 						fmt.Sprintf("https://%s:%s", nodegroup.Spec.APIServerLB, nodegroup.Spec.APIServerLBPort),
// 						nodegroupControllerCertCN,
// 						caCertb,
// 						cert,
// 						key))
// 				if err != nil {
// 					log.Error(err, "get kubernetes client", "cluster", cluster.Name, "nodegroup", nodegroup.Name)
// 					continue
// 				}

// 				kubernetesProvider, err := newKubernetesProvider(cluster.Spec.KubernetesProvider, config, false, kubernetesClient)
// 				if err != nil {
// 					log.Error(err, "get kubernetes provider", "cluster", cluster.Name, "nodegroup", nodegroup.Name)
// 					continue
// 				}

// 				var nodes []privatecloudv1alpha1.NodeStatus
// 				if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
// 					nodes, err = getControlPlaneNodeStatus(
// 						ctx,
// 						r.Client,
// 						nodegroup.Spec.NodeProvider,
// 						nodegroup.Namespace,
// 						nodegroup.Spec.EtcdLB,
// 						nodegroup.Spec.EtcdLBPort,
// 						nodegroup.Spec.ClusterName,
// 						nodegroup.Status.Nodes,
// 						nodeProvider,
// 						nodegroup.Spec.CloudAccountId,
// 						r.Config.CertExpirations.ControllerCertExpirationPeriod)
// 					if err != nil {
// 						log.Error(err, "get controlplane node status", "cluster", cluster.Name, "nodegroup", nodegroup.Name)
// 						continue
// 					}
// 				} else {
// 					nodes, err = getWorkerNodeStatus(
// 						ctx,
// 						nodegroup.Spec.ClusterName,
// 						nodegroup.Spec.NodeProvider,
// 						nodegroup.Status.Nodes, nodeProvider, kubernetesProvider, nodegroup.Spec.CloudAccountId)
// 					if err != nil {
// 						log.Error(err, "get worker node status", "cluster", cluster.Name, "nodegroup", nodegroup.Name)
// 						continue
// 					}
// 				}

// 				// Checking count
// 				if len(nodes) != nodegroup.Spec.Count {
// 					log.V(0).Info("Monitoring: current count is not equal to desired count",
// 						"cluster", cluster.Name,
// 						"nodegroup", nodegroup.Name,
// 						"current", len(nodes),
// 						"desired", nodegroup.Spec.Count)

// 					clusterEvent <- event.GenericEvent{
// 						Object: nodegroup.DeepCopy(),
// 					}
// 					continue
// 				}

// 				// Checking node state and cordoned nodes.
// 				// As soon as we find a node in a state != active, we trigger reconciler so that
// 				// we report back that state to the API.
// 				for _, node := range nodes {
// 					if node.State != privatecloudv1alpha1.ActiveNodegroupState {
// 						log.V(0).Info("Monitoring: node state has changed externally",
// 							"cluster", cluster.Name,
// 							"nodegroup", nodegroup.Name,
// 							"node", node.Name,
// 							"node state", node.State)
// 						clusterEvent <- event.GenericEvent{
// 							Object: nodegroup.DeepCopy(),
// 						}
// 						break
// 					}

// 					for _, nodeInStatus := range nodegroup.Status.Nodes {
// 						if node.Name == nodeInStatus.Name && node.Unschedulable != nodeInStatus.Unschedulable {
// 							log.V(0).Info("Monitoring: node unschedulable has changed externally",
// 								"cluster", cluster.Name,
// 								"nodegroup", nodegroup.Name,
// 								"node", node.Name,
// 								"unschedulable", node.Unschedulable)
// 							clusterEvent <- event.GenericEvent{
// 								Object: nodegroup.DeepCopy(),
// 							}
// 							break
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// }

// requiredActiveNodes returns true if the given number of nodes are active.
func requiredActiveNodes(number int, nodes []privatecloudv1alpha1.NodeStatus) bool {
	activeNodes := 0
	for _, node := range nodes {
		if node.State == privatecloudv1alpha1.ActiveNodegroupState {
			activeNodes++
		}
	}

	return activeNodes >= number
}

// removeDeletedNode returns an updated slice of nodes after removing the deleted node.
func removeDeletedNode(nodes []privatecloudv1alpha1.NodeStatus, deletedNodeName string) []privatecloudv1alpha1.NodeStatus {
	updatedNodes := make([]privatecloudv1alpha1.NodeStatus, 0, len(nodes)-1)

	for _, n := range nodes {
		if n.Name != deletedNodeName {
			updatedNodes = append(updatedNodes, n)
		}
	}

	return updatedNodes
}

// isInstanceGroup determines if a nodegroup should handle instance groups instead of regular instances by checking its
// instance type. Instance types for instance groups have the following naming convention "bm-icp-gaudi2-cluster-4", where cluster
// specifies that it is an instance group and the integer the number of instances in the instance group.
func isInstanceGroup(nodegroup *privatecloudv1alpha1.Nodegroup) bool {
	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
		return false
	}

	identifier := "cluster"
	return strings.Contains(nodegroup.Spec.InstanceType, identifier)
}

// getInstanceGroupInstanceTypeAndCount extracts the count and the instance type of the instances that will be created as part of instance group.
// The instance type follows a naming convention like this one "bm-icp-gaudi2-cluster-4". In this example "bm-icp-gaudi2" is the real instance type to use and "cluster-4" the
// number of instances to create as part of instance group.
func getInstanceGroupInstanceTypeAndCount(igInstanceType string) (string, int, error) {
	igInstanceTypeSplit := strings.Split(igInstanceType, "-")

	instanceCount, err := strconv.Atoi(igInstanceTypeSplit[len(igInstanceTypeSplit)-1])
	if err != nil {
		return "", 0, errors.Wrapf(err, "parse instance group count")
	}

	instanceType := strings.Join(igInstanceTypeSplit[:len(igInstanceTypeSplit)-2], "-")

	return instanceType, instanceCount, nil
}

// getInstanceGroupName given the name of a node that belong to an instance group, the instance group name will be extracted from it since nodes are named with the instance group
// name as their prefix. Example: Instance group: ng-ntat6s264e-ig-12345, Instance: ng-ntat6s264e-ig-12345-0.
func getInstanceGroupName(nodeName string) string {
	nodeNameSplit := strings.Split(nodeName, "-")
	return strings.Join(nodeNameSplit[:len(nodeNameSplit)-1], "-")
}

// countInstanceGroups determines the number of instance groups created by parsing the name of the instances and
// obtaining the instance group name.
// When a nodegroup creates an instance group instead of regular instances, instances will use the instance group name
// as a prefix.
// Example: Instance group: ng-ntat6s264e-ig-12345, Instance: ng-ntat6s264e-ig-12345-0.
// TODO: Nodegroup CRD should have a field to identify if it is instance group or instance.
func countInstanceGroups(nodes []privatecloudv1alpha1.NodeStatus) int {
	instanceGroups := make(map[string]int, 0)

	for _, node := range nodes {
		nodeNameSplit := strings.Split(node.Name, "-")
		instanceGroups[strings.Join(nodeNameSplit[:len(nodeNameSplit)-1], "-")] = 0
	}

	return len(instanceGroups)
}

func createBootstrapToken(ctx context.Context, kubernetesProvider kubernetesProvider) (string, string, error) {
	tokenId, tokenSecret, expiration, err := getBootstrapTokenData()
	if err != nil {
		return "", "", errors.Wrapf(err, "get bootstrap token data")
	}

	bootstrapTokenSecret := corev1.Secret{}
	bootstrapTokenSecret.Name = "bootstrap-token-" + tokenId
	bootstrapTokenSecret.Type = corev1.SecretTypeBootstrapToken
	bootstrapTokenSecret.StringData = map[string]string{
		"description":                    "Bootstrap token for worker node registration",
		"token-id":                       tokenId,
		"token-secret":                   tokenSecret,
		"expiration":                     expiration,
		"usage-bootstrap-authentication": "true",
		"usage-bootstrap-signing":        "true",
		"auth-extra-groups":              "system:bootstrappers:worker",
	}

	if err := kubernetesProvider.CreateBootstrapTokenSecret(ctx, &bootstrapTokenSecret); err != nil {
		return "", "", errors.Wrapf(err, "create kubernetes bootstrap secret %s", tokenId)
	}

	return tokenId, tokenSecret, nil
}

func countMissingNodesOrInstanceGroups(isInstanceGroup bool, desiredCount int, observedNodegroupStatus privatecloudv1alpha1.NodegroupStatus) int {
	if isInstanceGroup {
		// The nodegroup CRD doesn't have in the status something that allow us to store
		// the instance groups, so we identify the possible instace groups from the instance names in the status.
		return desiredCount - countInstanceGroups(observedNodegroupStatus.Nodes)
	}

	// If regular instances must be created, we just compare existing nodes against desired nodes.
	return desiredCount - observedNodegroupStatus.Count
}

func getInstanceGroupAndMissingNodes(igInstanceType string, observedNodegroupStatus privatecloudv1alpha1.NodegroupStatus) (map[string]int, error) {
	missingInstanceGroupCounts := make(map[string]int)

	_, countPerInstanceGroup, err := getInstanceGroupInstanceTypeAndCount(igInstanceType)
	if err != nil {
		return missingInstanceGroupCounts, err
	}

	instanceGroups := make(map[string]int)
	for _, node := range observedNodegroupStatus.Nodes {
		nodeNameSplit := strings.Split(node.Name, "-")
		instanceGroups[strings.Join(nodeNameSplit[:len(nodeNameSplit)-1], "-")]++
	}

	for instanceGroup, count := range instanceGroups {
		if count < countPerInstanceGroup {
			missingInstanceGroupCounts[instanceGroup] = countPerInstanceGroup - count
		}
	}

	return missingInstanceGroupCounts, nil
}

func extraNodesOrInstanceGroups(isInstanceGroup bool, desiredCount int, observedNodegroupStatus privatecloudv1alpha1.NodegroupStatus) int {
	if isInstanceGroup {
		// The nodegroup CRD doesn't have in the status something that allow us to store
		// the instance groups, so we identify the possible instace groups from the instance names in the status.
		return countInstanceGroups(observedNodegroupStatus.Nodes) - desiredCount
	}

	// If regular instances must be created, we just compare existing nodes against desired nodes.
	return observedNodegroupStatus.Count - desiredCount
}

func (r *NodegroupReconciler) createNodeOrInstanceGroup(
	ctx context.Context,
	isInstanceGroup bool,
	nodegroup privatecloudv1alpha1.Nodegroup,
	observedNodegroupStatus *privatecloudv1alpha1.NodegroupStatus,
	nodeProvider nodeProvider,
	kubernetesProvider kubernetesProvider) error {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.createNodeOrInstanceGroup")

	// For harvester node provider, dynamic network configuration is not supported,
	// so controller has to manually set it.
	var netErr error
	var ip, nameserver, gateway string
	if nodegroup.Spec.NodeProvider == HarvesterProviderName {
		ip, nameserver, gateway, netErr = getNetworkInformation(nodegroup.Annotations, observedNodegroupStatus.Nodes)
		if netErr != nil {
			return errors.Wrapf(netErr, "Get network information")
		}
	}

	registrationCmd, err := getRegistrationCmd(ctx, r.Client, nodegroup.Spec.ClusterName, nodegroup.Namespace, nodegroup.Spec.NodegroupType)
	if err != nil {
		return errors.Wrapf(err, "Get registration command")
	}

	bootstrapScript, err := kubernetesProvider.GetBootstrapScript(nodegroup.Spec.NodegroupType)
	if err != nil {
		return errors.Wrapf(err, "Get bootstrap script")
	}

	var createErr error
	var nodeStatus privatecloudv1alpha1.NodeStatus
	var nodeStatusList []privatecloudv1alpha1.NodeStatus
	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
		log.V(0).Info("Provisioning controlplane node")

		etcdEncryptionConfig, _, err := GetEtcdEncryptionConfigFromClusterSecret(
			ctx,
			r.Client,
			k8stypes.NamespacedName{Name: nodegroup.Spec.ClusterName, Namespace: nodegroup.Namespace},
			nodegroup.Spec.InstanceIMI)
		if err != nil {
			return errors.Wrapf(err, "Get etcd encryption config")
		}

		// Since controlplane nodes are created one by one, we need to set the right values for etcd cluster state and
		// initial cluster variables. We do this by checking if this is the first node, so it will create a new etcd cluster
		// otherwise we will add controlplane nodes to an existing etcd cluster.
		clusterState := "new"
		if observedNodegroupStatus.Count > 0 {
			clusterState = "existing"
			registrationCmd = registrationCmd + fmt.Sprintf(" --etcd-initial-cluster %s", getInitialCluster(observedNodegroupStatus.Nodes))
		}

		registrationCmd = registrationCmd +
			fmt.Sprintf(" --etcd-cluster-state %s", clusterState) +
			fmt.Sprintf(" --etcd-encryption-config %s", base64.StdEncoding.EncodeToString([]byte(etcdEncryptionConfig))) +
			fmt.Sprintf(" --iptables-enabled %t", r.Config.IPTables.Enabled) +
			fmt.Sprintf(" --iptables-cidr %s", r.Config.IPTables.ControlplaneCIDR) +
			r.loggingRegistrationCmdFlags(ctx, nodegroup) + // add logging related parameters
			r.metricsRegistrationCmdParams(ctx, nodegroup) // add metrics related parameters

		nodeStatus, createErr = nodeProvider.CreateNode(ctx, ip, nameserver, gateway, registrationCmd, bootstrapScript, nodegroup)
		if createErr == nil {
			observedNodegroupStatus.Nodes = append(observedNodegroupStatus.Nodes, nodeStatus)
		}
	}

	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.WorkerNodegroupType {
		// Generate a bootstrap token secret, create the secret in the downstream cluster
		// and pass token to registration command.
		tokenId, tokenSecret, err := createBootstrapToken(ctx, kubernetesProvider)
		if err != nil {
			return errors.Wrapf(err, "Create bootstrap token secret")
		}

		// Get weka storage information.
		var numCores = wekaDefaultNumCores
		if wekaInformation, found := r.Config.Weka.InstanceTypes[nodegroup.Spec.InstanceType]; found {
			numCores = wekaInformation.NumCores
		}

		var mode = wekaDefaultMode

		if !strings.Contains(nodegroup.Spec.InstanceType, "vm") && nodegroup.Spec.WekaStorage.Enable {
			mode = "enable"
		}

		log.V(0).Info("Weka configuration", logkeys.NumCores, numCores, logkeys.Mode, mode, logkeys.InstanceType, nodegroup.Spec.InstanceType, logkeys.StorageEnabled, nodegroup.Spec.WekaStorage.Enable)

		registrationCmd = registrationCmd +
			" --bootstrap-token " + tokenId + "." + tokenSecret +
			" --storage-weka-num-cores " + numCores +
			" --storage-weka-mode " + mode +
			" --storage-weka-sw-version " + r.Config.Weka.SoftwareVersion +
			" --storage-agent-url " + r.Config.Weka.ClusterUrl

		instanceTypeMeta, machineImageMeta := r.getInstanceTypeMachineImageMeta(ctx, nodegroup)
		nodeLabels := workerNodeAdditionalLabels(ctx, instanceTypeMeta, machineImageMeta)

		if nodeLabels != "" {
			registrationCmd = registrationCmd + " --kubelet-node-labels " + nodeLabels
		}

		if isInstanceGroup {
			log.V(0).Info("Provisioning worker instance group")

			instanceType, instanceCount, err := getInstanceGroupInstanceTypeAndCount(nodegroup.Spec.InstanceType)
			if err != nil {
				return errors.Wrapf(err, "Get instance type and count for instance group")
			}

			nodeStatusList, _, createErr = nodeProvider.CreatePrivateInstanceGroup(ctx, registrationCmd, bootstrapScript, instanceType, instanceCount, nodegroup)
			if createErr == nil {
				observedNodegroupStatus.Nodes = append(observedNodegroupStatus.Nodes, nodeStatusList...)
			}
		} else {
			log.V(0).Info("Provisioning worker node")

			nodeStatus, createErr = nodeProvider.CreateNode(ctx, ip, nameserver, gateway, registrationCmd, bootstrapScript, nodegroup)
			if createErr == nil {
				observedNodegroupStatus.Nodes = append(observedNodegroupStatus.Nodes, nodeStatus)
			}
		}
	}

	if createErr != nil {
		operatorMessageString, err := json.Marshal(OperatorMessage{
			Message:   grpcstatus.Convert(createErr).Message(),
			ErrorCode: int32(grpcstatus.Code(createErr)),
		})
		if err != nil {
			log.Error(err, "Marshal create error", logkeys.Message, createErr)
		}

		observedNodegroupStatus.Message = string(operatorMessageString)

		return createErr
	}

	observedNodegroupStatus.Count = len(observedNodegroupStatus.Nodes)

	return nil
}

// metricsRegistrationCmdParams metrics related parameters for registrationCmd
func (r *NodegroupReconciler) metricsRegistrationCmdParams(ctx context.Context, nodegroup privatecloudv1alpha1.Nodegroup) string {
	if r.Metrics == nil {
		return ""
	}

	var result string

	cluster := privatecloudv1alpha1.Cluster{}
	objKey := k8stypes.NamespacedName{Namespace: nodegroup.Namespace, Name: nodegroup.Spec.ClusterName}
	if err := r.Get(ctx, objKey, &cluster); err != nil {
		// if failed to get the cluster CR
		// disable metrics and log an error
		log.FromContext(ctx).
			WithName("NodegroupReconciler.metricsRegistrationCmdParams").
			Error(err, "failed to get cluster by name, metrics will be disabled")
		return ""
	}

	customerCloudAccountId := cluster.Spec.CustomerCloudAccountId

	result = fmt.Sprintf(" --region %s", nodegroup.Spec.Region) + fmt.Sprintf(" --cloudaccount %s", customerCloudAccountId)

	if r.Metrics.SystemMetrics != nil {
		result += fmt.Sprintf("  --system-metrics-enabled %t", true)
		result += fmt.Sprintf("  --monitoring-enabled %t", true) // // ToDo: support old version of flags in bootstrap-iks-controlplane.sh; remove this flag when bootstrap-iks-controlplane.sh that supports --system-metrics* params is deployed on all environments
		if prometheus := r.Metrics.SystemMetrics.PrometheusRemoteWrite; prometheus != nil {
			result += fmt.Sprintf("  --system-metrics-prometheus-url %s", prometheus.Url)
			result += fmt.Sprintf("  --monitoring-url %s", prometheus.Url) // ToDo: support old version of flags in bootstrap-iks-controlplane.sh; remove this flag when bootstrap-iks-controlplane.sh that supports --system-metrics* params is deployed on all environments
			if prometheus.BasicAuth != nil {
				result += fmt.Sprintf("  --system-metrics-prometheus-username %s", prometheus.BasicAuth.Username)
				result += fmt.Sprintf("  --system-metrics-prometheus-password %s", prometheus.BasicAuth.Password)

				result += fmt.Sprintf("  --monitoring-username %s", prometheus.BasicAuth.Username) // ToDo: support old version of flags in bootstrap-iks-controlplane.sh; remove this flag when bootstrap-iks-controlplane.sh that supports --system-metrics* params is deployed on all environments
				result += fmt.Sprintf("  --monitoring-password %s", prometheus.BasicAuth.Password) // ToDo: support old version of flags in bootstrap-iks-controlplane.sh; remove this flag when bootstrap-iks-controlplane.sh that supports --system-metrics* params is deployed on all environments

			}
		}
	}

	if r.Metrics.EndUserMetrics != nil {
		result += fmt.Sprintf("  --end-user-metrics-enabled %t", true)
		if prometheus := r.Metrics.EndUserMetrics.PrometheusRemoteWrite; prometheus != nil {
			result += fmt.Sprintf(" --end-user-metrics-prometheus-url %s", prometheus.Url)
			if prometheus.BearerToken != "" {
				result += fmt.Sprintf(" --end-user-metrics-prometheus-bearer-token %s", prometheus.BearerToken)
			}
		}
	}
	return result
}

func (r *NodegroupReconciler) createInstanceGroupMissingNodes(
	ctx context.Context,
	missingInstanceGroupNodesCount int,
	nodegroup privatecloudv1alpha1.Nodegroup,
	observedNodegroupStatus *privatecloudv1alpha1.NodegroupStatus,
	nodeProvider nodeProvider,
	kubernetesProvider kubernetesProvider,
	instanceGroup string) error {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.createInstanceGroupMissingNodes")

	registrationCmd, err := getRegistrationCmd(ctx, r.Client, nodegroup.Spec.ClusterName, nodegroup.Namespace, nodegroup.Spec.NodegroupType)
	if err != nil {
		return errors.Wrapf(err, "Get registration command")
	}

	bootstrapScript, err := kubernetesProvider.GetBootstrapScript(nodegroup.Spec.NodegroupType)
	if err != nil {
		return errors.Wrapf(err, "Get bootstrap script")
	}

	var createErr error
	// var nodeStatus privatecloudv1alpha1.NodeStatus
	var nodeStatusList []privatecloudv1alpha1.NodeStatus

	// Generate a bootstrap token secret, create the secret in the downstream cluster
	// and pass token to registration command.
	tokenId, tokenSecret, err := createBootstrapToken(ctx, kubernetesProvider)
	if err != nil {
		return errors.Wrapf(err, "Create bootstrap token secret")
	}

	// Get weka storage information.
	var numCores = wekaDefaultNumCores
	if wekaInformation, found := r.Config.Weka.InstanceTypes[nodegroup.Spec.InstanceType]; found {
		numCores = wekaInformation.NumCores
	}

	var mode = wekaDefaultMode
	if !strings.Contains(nodegroup.Spec.InstanceType, "vm") && nodegroup.Spec.WekaStorage.Enable {
		mode = "enable"
	}

	log.V(0).Info("Weka configuration", logkeys.NumCores, numCores, logkeys.Mode, mode, logkeys.InstanceType, nodegroup.Spec.InstanceType, logkeys.StorageEnabled, nodegroup.Spec.WekaStorage.Enable)

	registrationCmd = registrationCmd +
		" --bootstrap-token " + tokenId + "." + tokenSecret +
		" --storage-weka-num-cores " + numCores +
		" --storage-weka-mode " + mode +
		" --storage-weka-sw-version " + r.Config.Weka.SoftwareVersion +
		" --storage-agent-url " + r.Config.Weka.ClusterUrl

	instanceTypeMeta, machineImageMeta := r.getInstanceTypeMachineImageMeta(ctx, nodegroup)
	nodeLabels := workerNodeAdditionalLabels(ctx, instanceTypeMeta, machineImageMeta)

	if nodeLabels != "" {
		registrationCmd = registrationCmd + " --kubelet-node-labels " + nodeLabels
	}
	log.V(0).Info("Provisioning worker instance group")

	instanceType, instanceCount, err := getInstanceGroupInstanceTypeAndCount(nodegroup.Spec.InstanceType)
	if err != nil {
		return errors.Wrapf(err, "Get instance type and count for instance group")
	}

	// We will call Compute Scaleup Logic to add missing nodes in instance type.
	nodeStatusList, _, createErr = nodeProvider.ScaleUpInstanceGroup(ctx, registrationCmd, bootstrapScript, instanceType, instanceCount, nodegroup, instanceGroup)
	if createErr == nil {
		observedNodegroupStatus.Nodes = append(observedNodegroupStatus.Nodes, nodeStatusList...)
	}

	log.V(0).Info("Provisioning worker instance group", "Missing Instance Count in Instance Group", missingInstanceGroupNodesCount, "Instance Count of instance group after scale up", len(observedNodegroupStatus.Nodes))

	if createErr != nil {
		operatorMessageString, err := json.Marshal(OperatorMessage{
			Message:   grpcstatus.Convert(createErr).Message(),
			ErrorCode: int32(grpcstatus.Code(createErr)),
		})
		if err != nil {
			log.Error(err, "Marshal create error", "create error message", createErr)
		}

		observedNodegroupStatus.Message = string(operatorMessageString)

		return createErr
	}

	observedNodegroupStatus.Count = len(observedNodegroupStatus.Nodes)

	return nil
}

func (r *NodegroupReconciler) observeCurrentState(ctx context.Context, isInstanceGroup bool, nodegroup privatecloudv1alpha1.Nodegroup, nodeProvider nodeProvider, kubernetesProvider kubernetesProvider) (privatecloudv1alpha1.NodegroupStatus, error) {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.observeCurrentState")
	observedNodegroupStatus := nodegroup.Status
	observedNodegroupStatus.Name = nodegroup.Name

	var nodes []privatecloudv1alpha1.NodeStatus
	var err error
	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
		observedNodegroupStatus.Type = privatecloudv1alpha1.ControlplaneNodegroupType
		nodes, err = getControlPlaneNodeStatus(ctx, r.Client, nodegroup.Spec.NodeProvider, nodegroup.Namespace, nodegroup.Spec.EtcdLB, nodegroup.Spec.EtcdLBPort, nodegroup.Spec.ClusterName, observedNodegroupStatus.Nodes, nodeProvider, nodegroup.Spec.CloudAccountId, r.Config.CertExpirations.ControllerCertExpirationPeriod)
	} else {
		observedNodegroupStatus.Type = privatecloudv1alpha1.WorkerNodegroupType
		nodes, err = getWorkerNodeStatus(ctx, nodegroup.Spec.ClusterName, nodegroup.Spec.NodeProvider, observedNodegroupStatus.Nodes, nodeProvider, kubernetesProvider, nodegroup.Spec.CloudAccountId)
	}
	if err != nil {
		return observedNodegroupStatus, errors.Wrapf(err, "Get status of nodes")
	}

	observedNodegroupStatus.Nodes = nodes
	observedNodegroupStatus.Count = len(nodes)
	observedNodegroupStatus.State = privatecloudv1alpha1.ActiveNodegroupState
	observedNodegroupStatus.Message = "Nodegroup ready"

	// Get the weka registered agents.
	registeredAgents := make(map[string]*pb.FilesystemAgent, 0)
	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.WorkerNodegroupType && nodegroup.Spec.WekaStorage.Enable {
		client, err := r.FilesystemStorageClusterPrivateClient.ListRegisteredAgents(ctx, &pb.ListRegisteredAgentRequest{
			ClusterId: nodegroup.Spec.WekaStorage.ClusterId,
		})
		if err != nil {
			return observedNodegroupStatus, errors.Wrapf(err, "Get weka registered agents client")
		}

		for {
			registeredAgent, err := client.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				return observedNodegroupStatus, errors.Wrapf(err, "Receive weka registered agent")
			}

			registeredAgents[registeredAgent.Name] = registeredAgent
		}
	}

	for i := range observedNodegroupStatus.Nodes {
		// This will ensure that nodes that didn't change their state do not change their last update
		// time so that we keep the time since they were in that state.
		for _, existingNode := range nodegroup.Status.Nodes {
			if observedNodegroupStatus.Nodes[i].Name == existingNode.Name {
				if observedNodegroupStatus.Nodes[i].State == existingNode.State {
					log.Info("State of node has not changed",
						logkeys.NodeName, observedNodegroupStatus.Nodes[i].Name,
						logkeys.NodeState, observedNodegroupStatus.Nodes[i].State,
						logkeys.LastUpdatedTime, observedNodegroupStatus.Nodes[i].LastUpdate)
					observedNodegroupStatus.Nodes[i].LastUpdate = existingNode.LastUpdate
				}
			}
		}

		// Set weka registration status.
		if registeredAgent, found := registeredAgents[observedNodegroupStatus.Nodes[i].Name]; found {
			observedNodegroupStatus.Nodes[i].WekaStorageStatus.ClientId = registeredAgent.ClientId
			observedNodegroupStatus.Nodes[i].WekaStorageStatus.Status = registeredAgent.PredefinedStatus
			observedNodegroupStatus.Nodes[i].WekaStorageStatus.CustomStatus = registeredAgent.CustomStatus
			//observedNodegroupStatus.Nodes[i].WekaStorageStatus.Status = registeredAgent.Status
			//observedNodegroupStatus.Nodes[i].WekaStorageStatus.Message = registeredAgent.StatusMsg
		}

		// Set VAST Storage Status
		if !nodegroup.Spec.WekaStorage.Enable && nodegroup.Spec.WekaStorage.Mode == "vast" {
			observedNodegroupStatus.Nodes[i].WekaStorageStatus.Status = "Active"
			observedNodegroupStatus.Nodes[i].WekaStorageStatus.CustomStatus = "Active"
			log.Info("Setting VAST storage Status to Active for Node",
				logkeys.NodeName, observedNodegroupStatus.Nodes[i].Name,
				logkeys.NodeState, observedNodegroupStatus.Nodes[i].State)
		}

		// Set the nodegroup state based on the state of the nodes. If a node is not active, the nodegroup is set to updating.
		if observedNodegroupStatus.Nodes[i].State != privatecloudv1alpha1.ActiveNodegroupState {
			observedNodegroupStatus.State = privatecloudv1alpha1.UpdatingNodegroupState
			observedNodegroupStatus.Message = ""
		}
	}

	// Set the nodegroup state based on the number nodes.
	currentCount := observedNodegroupStatus.Count
	if isInstanceGroup {
		currentCount = countInstanceGroups(observedNodegroupStatus.Nodes)
	}
	if nodegroup.Spec.Count != currentCount {
		observedNodegroupStatus.State = privatecloudv1alpha1.UpdatingNodegroupState
		observedNodegroupStatus.Message = ""
	}

	return observedNodegroupStatus, nil
}

func (r *NodegroupReconciler) reconcileStates(
	ctx context.Context,
	isInstanceGroup bool,
	nodegroup privatecloudv1alpha1.Nodegroup,
	observedNodegroupStatus *privatecloudv1alpha1.NodegroupStatus,
	nodeProvider nodeProvider,
	kubernetesProvider kubernetesProvider) (bool, error) {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.reconcileStates")
	var requeueSooner bool
	var recErrors = &errgroup.ErrGroup{}

	// Remove etcd members that are not part of the list of nodes.
	// This happens because we rely on the list of nodes stored in the status which is not correct,
	// because controllers can read from a stale cache.
	// TODO: this could be removed if we guarantee that we read an up to date list of nodes.
	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType && requiredActiveNodes(1, observedNodegroupStatus.Nodes) {
		if _, err := removeMissingEtcdMember(
			ctx,
			observedNodegroupStatus.Nodes,
			r.Client,
			nodegroup.Spec.EtcdLB,
			nodegroup.Spec.EtcdLBPort,
			nodegroup.Spec.ClusterName,
			nodegroup.Namespace,
			r.Config.CertExpirations.ControllerCertExpirationPeriod); err != nil {
			recErrors.Add(errors.Wrapf(err, "Deleting missing etcd member"))
		}
	}

	if countMissingNodesOrInstanceGroups(isInstanceGroup, nodegroup.Spec.Count, *observedNodegroupStatus) > 0 {
		log.V(0).Info("Creating node or instance group")
		observedNodegroupStatus.Message = "Provisioning nodegroup compute"
		observedNodegroupStatus.State = privatecloudv1alpha1.UpdatingNodegroupState

		if err := r.createNodeOrInstanceGroup(ctx, isInstanceGroup, nodegroup, observedNodegroupStatus, nodeProvider, kubernetesProvider); err != nil {
			recErrors.Add(errors.Wrapf(err, "Create node or instance group"))
		}
	}

	if isInstanceGroup {
		instanceGroupMissingCount, err := getInstanceGroupAndMissingNodes(nodegroup.Spec.InstanceType, *observedNodegroupStatus)
		if err != nil {
			recErrors.Add(errors.Wrapf(err, "Get instance group missing nodes"))
		}
		if len(instanceGroupMissingCount) > 0 {
			for instanceGroup, instanceGroupMissingNodesCount := range instanceGroupMissingCount {
				log.V(0).Info("Scaling up missing nodes", "for instance group", instanceGroup, "Total Missing Nodes", instanceGroupMissingNodesCount)
				observedNodegroupStatus.Message = "Provisioning nodegroup compute"
				observedNodegroupStatus.State = privatecloudv1alpha1.UpdatingNodegroupState

				if err := r.createInstanceGroupMissingNodes(ctx, instanceGroupMissingNodesCount, nodegroup, observedNodegroupStatus, nodeProvider, kubernetesProvider, instanceGroup); err != nil {
					recErrors.Add(errors.Wrapf(err, "Create missing nodes for instance group"))
				}
			}
		}
	}

	r.detectNodeForDeletion(ctx, isInstanceGroup, nodegroup, observedNodegroupStatus)

	deleted, err := r.deleteNodeInDeletionState(ctx, isInstanceGroup, nodegroup, observedNodegroupStatus, nodeProvider, kubernetesProvider)
	if err != nil {
		observedNodegroupStatus.Message = "Failed deleting nodegroup compute"
		observedNodegroupStatus.State = privatecloudv1alpha1.UpdatingNodegroupState

		recErrors.Add(errors.Wrapf(err, "Delete node in deletion state"))
	}
	if deleted {
		requeueSooner = true
	}

	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.WorkerNodegroupType {
		// Approve the csr created by workers to request certificates for Kubelet server. These are not automatically approved
		// by any kubernetes builtin controller, so controller must take care of this.
		if err := kubernetesProvider.ApproveKubeletServingCertificateSigningRequests(ctx, nodegroup.Name); err != nil {
			recErrors.Add(errors.Wrapf(err, "Approve kubelet serving certificate signing requests"))
		}

		// Register worker nodes if weka storage is enabled.
		// TODO:
		// - Deregister nodes if weka is disabled.
		// - Check weka mode to pass correct information.

		// We can skip registering a node in the storage backend by adding
		// a label and set its value to false.
		var register = true
		if v, found := nodegroup.Labels[storageRegisterLabelKey]; found {
			if parsed, err := strconv.ParseBool(v); err == nil {
				register = parsed
			}
		}

		if nodegroup.Spec.WekaStorage.Enable && register {
			for _, node := range observedNodegroupStatus.Nodes {
				if node.WekaStorageStatus.ClientId == "" && node.State == privatecloudv1alpha1.ActiveNodegroupState {
					log.V(0).Info("Registering weka agent", logkeys.NodeName, node.Name)
					ipAddress := ""
					if node.StorageBackendIP != "" {
						ipAddress = node.StorageBackendIP
					} else {
						ipAddress = node.IpAddress
					}
					log.V(0).Info("Registering weka agent", logkeys.NodeName, node.Name, "Storage Backend IP", ipAddress)
					request := &pb.RegisterAgentRequest{
						ClusterId: nodegroup.Spec.WekaStorage.ClusterId,
						Name:      node.Name,
						IpAddr:    ipAddress,
					}
					if _, err := r.FilesystemStorageClusterPrivateClient.RegisterAgent(ctx, request); err != nil {
						recErrors.Add(errors.Wrapf(err, "Register weka agent for node %s", node.Name))
					}
				}
			}
		}
	}

	return requeueSooner, recErrors.ErrOrNil()
}

func (r *NodegroupReconciler) deleteNodegroup(ctx context.Context, namespacedName k8stypes.NamespacedName, nodegroup privatecloudv1alpha1.Nodegroup, observedNodegroupStatus privatecloudv1alpha1.NodegroupStatus, nodeProvider nodeProvider, kubernetesProvider kubernetesProvider) error {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.deleteNodegroup")
	log.V(0).Info("Deleting nodegroup")

	if controllerutil.ContainsFinalizer(&nodegroup, deleteNodesFinalizer) {
		for _, node := range observedNodegroupStatus.Nodes {
			if err := r.deleteNode(ctx, node, nodegroup, nodeProvider, kubernetesProvider); err != nil {
				return errors.Wrapf(err, "Delete node %s", node.Name)
			}
		}

		if err := r.removeNodegroupFinalizerAndUpdateResource(ctx, namespacedName); err != nil {
			return errors.Wrapf(err, "Remove Finalizer and Update Nodegroup Resource")
		}
	}

	return nil
}

// detectNodeForDeletion sets the state of a node to Deleting so that it can be properly deleted later by the controller. This guarantees that
// even if the controller fails during deletion of the node, next reconcile will retry the deletion. If found, true will be returned.
//
// The criteria to mark a node for deletion is:
//   - Node in error state
//   - Node in updating state for more than grace period
//   - Node with undesired InstanceIMI
//   - If current count > desired count: node in error state, updating state, cordoned or oldest
func (r *NodegroupReconciler) detectNodeForDeletion(ctx context.Context, isInstanceGroup bool, nodegroup privatecloudv1alpha1.Nodegroup, observedNodegroupStatus *privatecloudv1alpha1.NodegroupStatus) bool {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.detectNodeForDeletion")

	// If controlplane, we want to mark a node for deletion only if there are enough quorum to continue having a healthy
	// cluster.
	if nodegroup.Spec.NodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
		if !requiredActiveNodes(1, observedNodegroupStatus.Nodes) {
			log.V(0).Info("There are not enough required active controlplane nodes")
			return false
		}
	}

	var oldestNodeIndex int
	var oldestNode privatecloudv1alpha1.NodeStatus

	errorNodeIndex := -1
	updatingNodeIndex := -1
	cordonedNodeIndex := -1

	// If instance group, we use this function to change the state to deleting in all nodes since
	// we treat the instance group as one unit, we create and delete all nodes.
	setDeletingStateInstanceGroup := func(instanceGroupName string) {
		for i, node := range observedNodegroupStatus.Nodes {
			if strings.Contains(node.Name, instanceGroupName) {
				observedNodegroupStatus.Nodes[i].State = privatecloudv1alpha1.DeletingNodegroupState
			}
		}
	}

	// We look for the Instance Type Specific Grace Period and if it is not present we set it to default Grace Period
	monitorGracePeriod := r.Config.MonitorGracePeriod
	for key, gracePeriodValue := range r.Config.MonitorGracePeriodByInstanceType {
		if strings.Contains(strings.ToLower(nodegroup.Spec.InstanceType), strings.ToLower(key)) {
			monitorGracePeriod = gracePeriodValue
		}
	}

	// We loop through the nodes to detect the first node in error state, updating state for more
	// than grace period or with undesired instanceIMI to quickly exit from here.
	for i, node := range observedNodegroupStatus.Nodes {
		if node.State == privatecloudv1alpha1.ErrorNodegroupState {
			errorNodeIndex = i

			// Check if AutoRepairDisabled is true, if it is disabled we will not set the node to delete state even when nodegroup is in error state
			if node.AutoRepairDisabled {
				log.V(0).Info("Maintenance Mode Detected", logkeys.NodeName, node.Name, logkeys.NodeState, node.State)
				continue
			} else {
				log.V(0).Info("Marking node for deletion due to error", logkeys.NodeName, node.Name)
				observedNodegroupStatus.Nodes[i].State = privatecloudv1alpha1.DeletingNodegroupState
				return true
			}
		}

		if node.State == privatecloudv1alpha1.UpdatingNodegroupState {
			updatingNodeIndex = i

			timeInUpdatingState := time.Since(node.LastUpdate.Time)
			log.V(0).Info("Grace Period for", logkeys.NodeName, node.Name, " with grace period", monitorGracePeriod)
			if timeInUpdatingState > monitorGracePeriod {
				// Check if AutoRepairDisabled is true, if it is disabled we will not set the node to delete state and instead we wait in updating state for further debug process
				if node.AutoRepairDisabled {
					log.V(0).Info("Maintenance Mode Detected, Node with autorepair label set to false", logkeys.NodeName, node.Name, logkeys.TimeInUpdatingStateSeconds, timeInUpdatingState.Seconds())
					continue
				} else {
					log.V(0).Info("Marking node for deletion due to long updating state", logkeys.NodeName, node.Name, logkeys.TimeInUpdatingStateSeconds, timeInUpdatingState.Seconds())
					observedNodegroupStatus.Nodes[i].State = privatecloudv1alpha1.DeletingNodegroupState
					return true
				}
			}
		}

		// TODO: understand how to upgrade instance groups. Currently customer would need to create a new one and delete existing one
		// with old k8s version.
		if node.InstanceIMI != nodegroup.Spec.InstanceIMI && !isInstanceGroup {
			if requiredActiveNodes(len(observedNodegroupStatus.Nodes), observedNodegroupStatus.Nodes) {
				log.V(0).Info("Marking node for deletion due to undesired instanceIMI", logkeys.NodeName, node.Name, logkeys.CurrentInstanceIMI, node.InstanceIMI, logkeys.DesiredInstanceIMI, nodegroup.Spec.InstanceIMI)
				observedNodegroupStatus.Nodes[i].State = privatecloudv1alpha1.DeletingNodegroupState
				return true
			} else {
				log.V(0).Info("There are not enough required active nodes to continue with upgrade")
				return false
			}
		}

		// Detect a cordoned node.
		if node.Unschedulable && (cordonedNodeIndex < 0) {
			cordonedNodeIndex = i
		}

		// Detect the oldest node.
		if i == 0 {
			oldestNode = node
		} else {
			if oldestNode.CreationTime.Time.After(node.CreationTime.Time) {
				oldestNode = node
				oldestNodeIndex = i
			}
		}
	}

	// update nodes deletion status accordingly to Compute API
	// oldestNodeIndex is used for keeping backward compatibility
	oldestNodeIndex = r.syncDeletedInstancesToNodegroupStatus(ctx,
		nodegroup.Spec.CloudAccountId,
		oldestNodeIndex,
		observedNodegroupStatus)

	// If we have more nodes or instance groups than desired, we select the node
	// with highest priority to be deleted.
	if extraNodesOrInstanceGroups(isInstanceGroup, nodegroup.Spec.Count, *observedNodegroupStatus) > 0 {
		var nodeIndex int

		if errorNodeIndex >= 0 {
			nodeIndex = errorNodeIndex
		} else if updatingNodeIndex >= 0 {
			nodeIndex = updatingNodeIndex
		} else if cordonedNodeIndex >= 0 {
			nodeIndex = cordonedNodeIndex
		} else {
			nodeIndex = oldestNodeIndex
		}

		if isInstanceGroup {
			instanceGroupName := getInstanceGroupName(observedNodegroupStatus.Nodes[nodeIndex].Name)
			log.V(0).Info("Marking extra instance group for deletion", logkeys.InstanceGroupName, instanceGroupName)

			setDeletingStateInstanceGroup(instanceGroupName)
		} else {
			log.V(0).Info("Marking extra node for deletion", logkeys.NodeName, observedNodegroupStatus.Nodes[nodeIndex].Name)
			observedNodegroupStatus.Nodes[nodeIndex].State = privatecloudv1alpha1.DeletingNodegroupState
		}

		return true
	}

	return false
}

// syncDeletedInstancesToNodegroupStatus mutate the observedNodegroupStatus
// by changing the state of the node
// to DeletingNodegroupState accordingly to Compute API
// return: oldestNodeIndex this is here for keeping the  backward compatibility to extraNodesOrInstanceGroups func
func (r *NodegroupReconciler) syncDeletedInstancesToNodegroupStatus(
	ctx context.Context,
	cloudAccountId string,
	oldestNodeIndex int, // this is here for keeping the backward compatibility with extraNodesOrInstanceGroups func
	observedNodegroupStatus *privatecloudv1alpha1.NodegroupStatus) int {
	l := log.FromContext(ctx).
		WithName("NodegroupReconciler.syncDeletedInstancesToNodegroupStatus")

	setDeletionStatus := func(status *privatecloudv1alpha1.NodegroupStatus, nodeIdx int) int {
		status.Nodes[nodeIdx].State = privatecloudv1alpha1.DeletingNodegroupState
		return nodeIdx
	}
	// we need to find a way to get all the instances within the nodegroup at once
	// we could use Search method, but we do not adding instance group value,
	// nor any identity labels for instances within the same node group
	// as a result, we need to make a Get for each instance individually

	// loop over all the nodes in the CR status spec
	// and find first node with none nil deletion timestamp
	for nodeIdx, node := range observedNodegroupStatus.Nodes {
		instanceGetRequest := &pb.InstanceGetRequest{
			Metadata: &pb.InstanceMetadataReference{
				CloudAccountId: cloudAccountId,
				NameOrId: &pb.InstanceMetadataReference_Name{
					Name: node.Name,
				},
			},
		}
		instance, err := r.InstanceServiceClient.Get(ctx, instanceGetRequest)
		// crucial to understand the exact error
		// in case of Not Found - assuming instance has been removed directly
		// from compute api and must be removed from the node group - that ok.
		// in any other case - situation is not clear,
		// if got an Internal or Unknown error, should requeue, if so, till when?
		// for now, in case the situation is not clear, I'll do nothing
		if err != nil {
			res, ok := grpcstatus.FromError(err)
			if !ok {
				l.Error(err, "failed to get a status code from compute service response", "node", node.Name)
			}
			if grpccodes.NotFound == res.Code() {
				oldestNodeIndex = setDeletionStatus(observedNodegroupStatus, nodeIdx)
			} else {
				// if unknown error occurs, do nothing, just continue to the next node
				l.Error(err, "unknown error, will do nothing", "node", node.Name)
			}
			continue
		}
		if instance.Metadata != nil && instance.Metadata.DeletionTimestamp != nil {
			oldestNodeIndex = setDeletionStatus(observedNodegroupStatus, nodeIdx)
		}
	}
	return oldestNodeIndex
}

// deleteNodeInDeletionState loops through the list of nodes and deletes the first node in deleting state.
// If a node is deleted, removes the node from the list and returns true.
func (r *NodegroupReconciler) deleteNodeInDeletionState(ctx context.Context, isInstanceGroup bool, nodegroup privatecloudv1alpha1.Nodegroup, observedNodegroupStatus *privatecloudv1alpha1.NodegroupStatus, nodeProvider nodeProvider, kubernetesProvider kubernetesProvider) (bool, error) {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.deleteNodeInDeletionState")

	for _, node := range observedNodegroupStatus.Nodes {
		if node.State == privatecloudv1alpha1.DeletingNodegroupState {
			log.V(0).Info("Deleting node", logkeys.NodeName, node.Name)

			var instanceGroupName string
			var isInstanceGroupExists bool
			var err error
			if isInstanceGroup {
				instanceGroupName = getInstanceGroupName(node.Name)
				isInstanceGroupExists, err = nodeProvider.SearchInstanceGroup(ctx, nodegroup.Spec.CloudAccountId, instanceGroupName)
				if err != nil {
					return false, err
				}
				log.V(0).Info("Deleting node", "Instance Group Existence Check", isInstanceGroupExists)
			}

			// If Instance Group and if it exists, we are calling new endpoint to delete the nodes in instance group that are in deleted state
			if isInstanceGroup && isInstanceGroupExists {
				// We will delete the instance group members only if instance group exists
				if len(observedNodegroupStatus.Nodes) == 1 {
					log.V(0).Info("Skipping Deleting node since instance group should have atleast one node", "node", node.Name)
					return true, nil
				} else {
					if err := r.deleteInstanceGroupNode(ctx, node, nodegroup, nodeProvider, kubernetesProvider, instanceGroupName); err != nil {
						return false, err
					}
				}
			} else {
				if err := r.deleteNode(ctx, node, nodegroup, nodeProvider, kubernetesProvider); err != nil {
					return false, err
				}
			}

			updatedNodes := removeDeletedNode(observedNodegroupStatus.Nodes, node.Name)
			observedNodegroupStatus.Nodes = updatedNodes
			observedNodegroupStatus.Count = len(updatedNodes)

			return true, nil
		}
	}

	return false, nil
}

func (r *NodegroupReconciler) loggingRegistrationCmdFlags(ctx context.Context, nodegroup privatecloudv1alpha1.Nodegroup) string {
	// if logging not enabled return logging enabled false flag
	if !r.Config.Logging.Enabled {
		return " --logging-enabled false"
	}
	// to make an enrichment cluster CR data is required
	cluster := privatecloudv1alpha1.Cluster{}
	objKey := k8stypes.NamespacedName{Namespace: nodegroup.Namespace, Name: nodegroup.Spec.ClusterName}
	if err := r.Get(ctx, objKey, &cluster); err != nil {
		// if failed to get the cluster CR
		// disable logging and log an error
		log.FromContext(ctx).
			WithName("NodegroupReconciler.loggingRegistrationCmdFlags").
			Error(err, "failed to get cluster by name, logging will be disabled")
		return " --logging-enabled false"
	}
	// construct all the logging flags
	return fmt.Sprintf(" --logging-enabled %s"+
		" --logging-host %s"+
		" --logging-user %s"+
		" --logging-password %s"+
		" --logging-enrichment %s",
		strconv.FormatBool(r.Config.Logging.Enabled),
		r.Config.Logging.Host,
		r.Config.Logging.User,
		r.Config.Logging.Password,
		r.Config.Logging.EnrichString(ctx, map[string]string{
			"CLOUD_ACCOUNT_ID":          nodegroup.Spec.CloudAccountId,
			"CUSTOMER_CLOUD_ACCOUNT_ID": cluster.Spec.CustomerCloudAccountId,
			"CLUSTER_ID":                nodegroup.Spec.ClusterName,
			"CLUSTER_REGION":            nodegroup.Spec.Region,
		}))
}

// getInstanceTypeMachineImageMeta returns the InstanaceType and MachineImage information for node in node group
// this is used to add labels to worker nodes
func (r *NodegroupReconciler) getInstanceTypeMachineImageMeta(ctx context.Context, nodegroup privatecloudv1alpha1.Nodegroup) (*pb.InstanceType, *pb.MachineImage) {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.getLabelsFromMachineImageComponents")

	instanceTypeMeta, err := r.InstanceTypeClient.Get(ctx, &pb.InstanceTypeGetRequest{
		Metadata: &pb.InstanceTypeGetRequest_Metadata{
			Name: nodegroup.Spec.InstanceType,
		},
	})
	if err != nil {
		log.Error(err, "Get instance type metadata", logkeys.InstanceType, nodegroup.Spec.InstanceType)
	}

	machineImageMeta, err := r.MachineImageClient.Get(ctx, &pb.MachineImageGetRequest{
		Metadata: &pb.MachineImageGetRequest_Metadata{
			Name: nodegroup.Spec.InstanceIMI,
		},
	})
	if err != nil {
		log.Error(err, "Get machine image metadata", logkeys.MachineImage, nodegroup.Spec.InstanceType)
	}
	return instanceTypeMeta, machineImageMeta
}

// workerNodeAdditionalLabels extracts additional labels from the instance type and returns them as a comma separated string, such as "label1=value1,label2=value2".
func workerNodeAdditionalLabels(ctx context.Context, instanceType *pb.InstanceType, machineImage *pb.MachineImage) string {
	labelsMap := make(map[string]string)

	if instanceType != nil && instanceType.Spec != nil {
		// Add labels based on instance type
		labelsMap = getLabelsFromInstanceType(ctx, instanceType)
	}

	//Add labels based on machine image components
	if machineImage != nil && machineImage.Spec != nil && machineImage.Spec.Components != nil {
		componentLabels := getLabelsFromMachineImageComponents(ctx, machineImage.Spec.Components)
		for k, v := range componentLabels {
			labelsMap[k] = v
		}
	}

	// Convert labels map to string
	var labels []string
	for key, value := range labelsMap {
		labels = append(labels, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(labels, ",")
}

// getLabelsFromInstanceType extracts labels from the instance type and returns them as a map.
func getLabelsFromInstanceType(ctx context.Context, instanceType *pb.InstanceType) map[string]string {
	result := make(map[string]string)

	// Add instance type name label
	if instanceType.Metadata != nil && instanceType.Metadata.Name != "" {
		addToLabelsMap(ctx, "instance-type", instanceType.Metadata.Name, result)
	}

	// Add HBM mode label
	if instanceType.Spec.HbmMode != "" {
		addToLabelsMap(ctx, "hbm-mode", instanceType.Spec.HbmMode, result)
	}

	// Add CPU related labels
	if cpu := instanceType.Spec.Cpu; cpu != nil {
		if cpu.Cores != 0 {
			addToLabelsMap(ctx, "host-cpu-cores", strconv.Itoa(int(cpu.Cores)), result)
		}
		if cpu.Id != "" {
			addToLabelsMap(ctx, "host-cpu-id", cpu.Id, result)
		}
		if cpu.ModelName != "" {
			cpuModelName := strings.Map(removeInvalidCharactersFromLabel, cpu.ModelName)
			addToLabelsMap(ctx, "host-cpu-model-name", cpuModelName, result)
		}
		if cpu.Sockets != 0 {
			addToLabelsMap(ctx, "host-cpu-sockets", strconv.Itoa(int(cpu.Sockets)), result)
		}
		if cpu.Threads != 0 {
			addToLabelsMap(ctx, "host-cpu-threads", strconv.Itoa(int(cpu.Threads)), result)
		}
	}

	// Add GPU related labels
	if gpu := instanceType.Spec.Gpu; gpu != nil {
		if gpu.Count != 0 {
			addToLabelsMap(ctx, "host-gpu-count", strconv.Itoa(int(gpu.Count)), result)
		}
		if gpu.ModelName != "" {
			gpuModelName := strings.Map(removeInvalidCharactersFromLabel, gpu.ModelName)
			addToLabelsMap(ctx, "host-gpu-model", gpuModelName, result)
		}
	}

	// Add memory size label
	if memory := instanceType.Spec.Memory; memory != nil {
		if memory.Size != "" {
			addToLabelsMap(ctx, "host-memory-size", memory.Size, result)
		}
	}
	return result
}

// getLabelsFromMachineImageComponents extracts labels from the machine image components data and returns them as a map.
// The labels are extracted from the components of type "Firmware kit" and the label is constructed as follows:
// "<normalized kit name>.<labelPrefix>/<normalized kit version> = true".
func getLabelsFromMachineImageComponents(ctx context.Context, components []*pb.MachineImageComponent) map[string]string {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.getLabelsFromMachineImageComponents")

	result := make(map[string]string)
	const (
		firmwareKitPrefix = "fk"
		softwareKitPrefix = "sk"
	)

	var skKey, fkKey string

	for _, c := range components {
		if c.Type == "Firmware kit" && c.Version != "" && c.Name != "" {
			normalizedKitName := normalizeKitName(c.Name)
			normalizedKitVersion := strings.ToLower(c.Version)

			// Construct the label key for the firmware kit, like "<labelPrefix>/<normalized kit name>-fk-<normalized kit version> = true"
			fkKey = fmt.Sprintf("%s/%s-%s-%s", labelPrefix, normalizedKitName, firmwareKitPrefix, normalizedKitVersion)
			if errs := validation.IsQualifiedName(fkKey); len(errs) > 0 { // if the label key is still not a valid kubernetes label key, log the error and skip it
				for _, errStr := range errs {
					log.Error(fmt.Errorf(errStr), fmt.Sprintf("invalid label key for label: %s", fkKey))
				}
			}
		}

		if c.Type == "Software kit" && c.Version != "" && c.Name != "" {
			normalizedKitName := normalizeKitName(c.Name)
			normalizedKitVersion := strings.ToLower(c.Version)

			// Construct the label key for the software kit, like "<normalized kit name>.sw.<labelPrefix>/<normalized kit version> = true"
			skKey = fmt.Sprintf("%s/%s-%s-%s", labelPrefix, normalizedKitName, softwareKitPrefix, normalizedKitVersion)
			if errs := validation.IsQualifiedName(skKey); len(errs) > 0 { // if the label key is still not a valid kubernetes label key, log the error and skip it
				for _, errStr := range errs {
					log.Error(fmt.Errorf(errStr), fmt.Sprintf("invalid label key for label: %s", skKey))
				}
			}
		}

		if fkKey != "" {
			result[fkKey] = "true"
		}

		if skKey != "" {
			result[skKey] = "true"
		}
	}
	return result
}

// remove invalid characters from the sw and fw kit names
func normalizeKitName(name string) string {
	normalizedKitName := strings.ToLower(name)
	normalizedKitName = strings.ReplaceAll(normalizedKitName, " ", "-")
	normalizedKitName = strings.Map(removeInvalidCharactersFromLabel, normalizedKitName)
	return normalizedKitName
}

// addToLabelsMap adds a key-value pair to the labels map if the value is not empty and is a valid kubernetes label value.
func addToLabelsMap(ctx context.Context, key string, value string, labelsMap map[string]string) {
	if value != "" && isValidLabelValue(ctx, value) {
		labelsMap[fmt.Sprintf("%s/%s", labelPrefix, key)] = value
	}
}

// isValidLabelValue checks if the given label string is a valid kubernetes label value.
func isValidLabelValue(ctx context.Context, label string) bool {
	log := log.FromContext(ctx).WithName("NodegroupReconciler.isValidLabelValue")

	errStrings := validation.IsValidLabelValue(label) // check if the label value is a valid kubernetes label value
	if errStrings != nil {
		for _, errStr := range errStrings {
			log.Error(fmt.Errorf(errStr), fmt.Sprintf("invalid label value for label: %s", label))
		}
		return false
	}
	return true
}

// helper func to remove invalid characters from the label values
func removeInvalidCharactersFromLabel(r rune) rune {
	if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '.' || r == '_' {
		return r
	}
	return -1

}
