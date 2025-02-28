// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package rancher

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/utils"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	prov1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	v1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	manv3 "github.com/rancher/rancher/pkg/client/generated/management/v3"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	clusterHealthzEndpoint      = "/healthz"
	rancherManagementAPIVersion = "/v3"
	rancherBaseAPIVersion       = "/v1"
	clusterV2APIVersion         = "provisioning.cattle.io/v1"
	clusterV2APIType            = "provisioning.cattle.io.cluster"
	clusterV2Kind               = "Cluster"
	namespace                   = "fleet-default"
	nodegroupNameLabel          = "nodegroupName"
)

type RancherProvider struct {
	ManagementClient            *manv3.Client
	BaseClient                  *clientbase.APIBaseClient
	ControlplaneBootstrapScript string
	WorkerBootstrapScript       string
}

type Cluster struct {
	types.Resource
	prov1.Cluster
}

func NewRancherProvider(url, accessKey, secretKey, controlplaneBootstrapScript, workerBootstrapScript string) (*RancherProvider, error) {
	// Management client
	managementClient, err := manv3.NewClient(&clientbase.ClientOpts{
		URL:       url + rancherManagementAPIVersion,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if err != nil {
		return nil, err
	}

	// Base client
	baseClient, err := clientbase.NewAPIClient(&clientbase.ClientOpts{
		URL:       url + "/k8s/clusters/local" + rancherBaseAPIVersion,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if err != nil {
		return nil, err
	}

	return &RancherProvider{
		ManagementClient:            managementClient,
		BaseClient:                  &baseClient,
		ControlplaneBootstrapScript: controlplaneBootstrapScript,
		WorkerBootstrapScript:       workerBootstrapScript,
	}, nil
}

func (p *RancherProvider) InitCluster(ctx context.Context, secret *corev1.Secret, cluster *privatecloudv1alpha1.Cluster, etcdLB string, apiserverLB string, publicApiserverLB string, konnectivityLB string, etcdLBPort int, apiserverLBPort int, publicApiserverLBPort int) error {
	rancherCluster := &Cluster{}

	rancherCluster.TypeMeta.Kind = clusterV2Kind
	rancherCluster.TypeMeta.APIVersion = clusterV2APIVersion

	rancherCluster.ObjectMeta.Name = cluster.Name
	rancherCluster.ObjectMeta.Namespace = namespace

	rancherCluster.Spec.KubernetesVersion = cluster.Spec.KubernetesVersion + "+rke2r1"
	rancherCluster.Spec.AgentEnvVars = make([]v1.EnvVar, 0)
	// for k, v := range cluster.Spec.KubernetesProviderConfig.EnvVars {
	// 	rancherCluster.Spec.AgentEnvVars = append(rancherCluster.Spec.AgentEnvVars, v1.EnvVar{
	// 		Name:  k,
	// 		Value: v,
	// 	})
	// }
	// rancherCluster.Spec.LocalClusterAuthEndpoint.Enabled = cluster.Spec.KubernetesProviderConfig.EnableAuthEndpoint
	// rancherCluster.Spec.LocalClusterAuthEndpoint.FQDN = cluster.Spec.KubernetesProviderConfig.FqdnAuthEndpoint
	// rancherCluster.Spec.LocalClusterAuthEndpoint.CACerts = cluster.Spec.KubernetesProviderConfig.CaCertAuthEndpoint
	rancherCluster.Spec.RKEConfig = &prov1.RKEConfig{}
	rancherCluster.Spec.RKEConfig.MachineGlobalConfig.Data = make(map[string]interface{}, 0)
	rancherCluster.Spec.RKEConfig.MachineGlobalConfig.Data["cluster-cidr"] = cluster.Spec.Network.PodCIDR
	rancherCluster.Spec.RKEConfig.MachineGlobalConfig.Data["cluster-dns"] = cluster.Spec.Network.ClusterDNS
	rancherCluster.Spec.RKEConfig.MachineGlobalConfig.Data["service-cidr"] = cluster.Spec.Network.ServiceCIDR

	resp := &Cluster{}
	if err := p.BaseClient.Create(clusterV2APIType, rancherCluster, resp); err != nil {
		return err
	}

	// Get kubeconfig
	rancherClusters, err := p.ManagementClient.Cluster.List(&types.ListOpts{Filters: map[string]interface{}{"name": cluster.Name}})
	if err != nil {
		return err
	}

	// Cluster not found
	if len(rancherClusters.Data) < 1 {
		return fmt.Errorf("cluster %s not found", cluster.Name)
	}

	kubeconfig, err := p.ManagementClient.Cluster.ActionGenerateKubeconfig(&rancherClusters.Data[0])
	if err != nil {
		return errors.Wrapf(err, "generate kubeconfig")
	}
	secret.Data["admin-kubeconfig"] = []byte(kubeconfig.Config)
	secret.Data["operator-kubeconfig"] = []byte(kubeconfig.Config)

	//Get registration command
	regToken, err := p.ManagementClient.ClusterRegistrationToken.List(&types.ListOpts{Filters: map[string]interface{}{"clusterId": rancherClusters.Data[0].ID}})
	if err != nil {
		return err
	}

	if len(regToken.Data) < 1 {
		return fmt.Errorf("registration token not found")
	}

	secret.Data["controlplane-registration-cmd"] = []byte("bash /usr/local/bin/bootstrap.sh --intel-ca true --configure-os true --registration-command '" + regToken.Data[0].InsecureNodeCommand + " --etcd --controlplane'")
	secret.Data["worker-registration-cmd"] = []byte("bash /usr/local/bin/bootstrap.sh --intel-ca true --configure-os true --registration-command '" + regToken.Data[0].InsecureNodeCommand + " --worker'")

	return nil
}

// GetCluster provides all the possible information that can be obtained from a kubernetes cluster API
func (p *RancherProvider) GetCluster(ctx context.Context, kubeconfig []byte) (*privatecloudv1alpha1.ClusterStatus, error) {
	// var clusterStatus privatecloudv1alpha1.ClusterStatus

	// rancherClusters, err := p.ManagementClient.Cluster.List(&types.ListOpts{Filters: map[string]interface{}{"name": clusterName}})
	// if err != nil {
	// 	return nil, err
	// }

	// // Cluster not found
	// if len(rancherClusters.Data) < 1 {
	// 	return nil, nil
	// }
	// rancherCluster := rancherClusters.Data[0]

	// clusterStatus.Nodegroups = make([]*privatecloudv1alpha1.NodegroupStatus, 0)

	// state, lastUpdate, reason, message := getClusterState(rancherCluster.Transitioning, rancherCluster.Conditions)
	// clusterStatus.State = state
	// clusterStatus.LastUpdate = lastUpdate
	// clusterStatus.Reason = reason
	// clusterStatus.Message = message
	client, err := utils.GetKubernetesClient(kubeconfig)
	if err != nil {
		return nil, err
	}
	restClient := client.RESTClient()

	var clusterStatus privatecloudv1alpha1.ClusterStatus
	clusterStatus.Nodegroups = make([]privatecloudv1alpha1.NodegroupStatus, 0)
	clusterStatus.State = privatecloudv1alpha1.ActiveClusterState
	clusterStatus.Reason = ""
	clusterStatus.Message = ""
	clusterStatus.LastUpdate = metav1.Time{Time: time.Now()}

	result := restClient.Get().AbsPath(clusterHealthzEndpoint).Do(ctx)
	if result.Error() != nil {
		return nil, result.Error()
	}

	if err := utils.GetKubernetesAPIHealthEndpoint(ctx, clusterHealthzEndpoint, restClient); err != nil {
		clusterStatus.State = privatecloudv1alpha1.ErrorClusterState
		clusterStatus.Reason = "KubernetesNotHealthy"
		clusterStatus.Message = err.Error()
		return &clusterStatus, nil
	}

	return &clusterStatus, nil
}

func (p *RancherProvider) CleanUpCluster(ctx context.Context, clusterName string) error {
	// Get cluster
	rancherClusters, err := p.ManagementClient.Cluster.List(&types.ListOpts{Filters: map[string]interface{}{"name": clusterName}})
	if err != nil {
		return err
	}

	// Cluster not found
	if len(rancherClusters.Data) < 1 {
		return nil
	}

	err = p.ManagementClient.Cluster.Delete(&rancherClusters.Data[0])
	if err != nil {
		return err
	}

	return nil
}

func (p *RancherProvider) GetNodes(ctx context.Context, nodegroupName string, kubeconfig []byte) ([]privatecloudv1alpha1.NodeStatus, error) {
	// rancherClusters, err := p.ManagementClient.Cluster.List(
	// 	&types.ListOpts{Filters: map[string]interface{}{"name": clusterName}})
	// if err != nil {
	// 	return nil, err
	// }

	// if len(rancherClusters.Data) < 1 {
	// 	return nil, fmt.Errorf("cluster not found")
	// }
	// rancherCluster := rancherClusters.Data[0]

	// rancherNodes, err := p.ManagementClient.Node.List(
	// 	&types.ListOpts{Filters: map[string]interface{}{"clusterId": rancherCluster.ID}})
	// if err != nil {
	// 	return nil, err
	// }

	// nodes := make([]*privatecloudv1alpha1.NodeStatus, 0)

	// for _, rancherNode := range rancherNodes.Data {
	// 	if v, ok := rancherNode.Labels[nodegroupNameLabel]; ok {
	// 		if v == nodegroupName {
	// 			nodeStatus := &privatecloudv1alpha1.NodeStatus{}
	// 			nodeStatus.Name = rancherNode.Hostname
	// 			nodeStatus.IpAddress = rancherNode.IPAddress

	// 			kubeletVersion := ""
	// 			if k := rancherNode.Info.Kubernetes; k != nil {
	// 				if len(strings.Split(k.KubeletVersion, "+")) > 1 {
	// 					kubeletVersion = strings.Split(k.KubeletVersion, "+")[0]
	// 				}
	// 			}
	// 			nodeStatus.KubeletVersion = kubeletVersion
	// 			nodeStatus.KubeProxyVersion = ""

	// 			state, lastUpdate, reason, message := getNodeState(rancherNode.Transitioning, rancherNode.Conditions)
	// 			nodeStatus.State = state
	// 			nodeStatus.LastUpdate = lastUpdate
	// 			nodeStatus.Reason = reason
	// 			nodeStatus.Message = message

	// 			nodes = append(nodes, nodeStatus)
	// 		}
	// 	}
	// }
	client, err := utils.GetKubernetesClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	nodesList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]privatecloudv1alpha1.NodeStatus, 0)
	for _, node := range nodesList.Items {
		// We get the name of the nodegroup from the node name since
		// nodes are named by following nodegroup name + generated sufix.
		nodeSplit := strings.Split(node.Name, "-")
		if strings.Join(nodeSplit[:len(nodeSplit)-1], "-") == nodegroupName {
			var nodeStatus privatecloudv1alpha1.NodeStatus
			nodeStatus.Name = node.Name
			nodeStatus.CreationTime = node.ObjectMeta.CreationTimestamp

			ipAddress := ""
			for _, address := range node.Status.Addresses {
				if address.Type == corev1.NodeInternalIP {
					ipAddress = address.Address
					break
				}
			}
			nodeStatus.IpAddress = ipAddress
			nodeStatus.KubeletVersion = node.Status.NodeInfo.KubeletVersion
			nodeStatus.KubeProxyVersion = node.Status.NodeInfo.KubeProxyVersion

			state, lastUpdate, reason, message := p.getNodeState(node.Status.Conditions)
			nodeStatus.State = state
			nodeStatus.LastUpdate = *lastUpdate
			nodeStatus.Reason = reason
			nodeStatus.Message = message

			nodes = append(nodes, nodeStatus)
		}
	}

	return nodes, nil
}

func (p *RancherProvider) DeleteNode(ctx context.Context, nodeName string, kubeconfig []byte) error {
	// // Get node
	// rancherNodes, err := p.ManagementClient.Node.List(&types.ListOpts{Filters: map[string]interface{}{"name": nodeName}})
	// if err != nil {
	// 	return err
	// }

	// // Node not found
	// if len(rancherNodes.Data) < 1 {
	// 	return nil
	// }

	// err = p.ManagementClient.Node.Delete(&rancherNodes.Data[0])
	// if err != nil {
	// 	return err
	// }
	client, err := utils.GetKubernetesClient(kubeconfig)
	if err != nil {
		return err
	}

	if err := client.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (p *RancherProvider) DrainNode(ctx context.Context, nodeName string, kubeconfig []byte) error {
	return nil
}

// getNodeState defines the node state based on the Rancher node conditions and transitioning field.
// Rancher does not provide an error condition type so we need to use transitioning field that specifies
// if there is an error.
func getNodeState(transitioning string, conditions []manv3.NodeCondition) (privatecloudv1alpha1.NodegroupState, *metav1.Time, string, string) {
	lastUpdate := &metav1.Time{Time: time.Now()}
	reason := ""
	message := ""
	state := privatecloudv1alpha1.UpdatingNodegroupState

	for _, condition := range conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				state = privatecloudv1alpha1.ActiveNodegroupState
			} else {
				state = privatecloudv1alpha1.UpdatingNodegroupState
			}
			reason = condition.Reason
			message = condition.Message

			lastUpdateTime, err := time.Parse(time.RFC3339, condition.LastTransitionTime)
			if err == nil {
				lastUpdate = &metav1.Time{Time: lastUpdateTime}
			}
		}
	}

	if transitioning == "error" {
		state = privatecloudv1alpha1.ErrorNodegroupState
	}

	return state, lastUpdate, reason, message
}

func (p *RancherProvider) ApproveKubeletServingCertificateSigningRequests(context.Context, string, []byte) error {
	return nil
}

func (p *RancherProvider) GetBootstrapScript(nodegroupType privatecloudv1alpha1.NodegroupType) (string, error) {
	bootstrapScript := p.WorkerBootstrapScript
	if nodegroupType == privatecloudv1alpha1.ControlplaneNodegroupType {
		bootstrapScript = p.ControlplaneBootstrapScript
	}

	data, err := os.ReadFile(bootstrapScript)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *RancherProvider) CreateBootstrapTokenSecret(context.Context, *corev1.Secret, []byte) error {
	return nil
}

func (p *RancherProvider) getNodeState(conditions []corev1.NodeCondition) (privatecloudv1alpha1.NodegroupState, *metav1.Time, string, string) {
	lastUpdate := &metav1.Time{Time: time.Now()}
	reason := ""
	message := ""
	state := privatecloudv1alpha1.UpdatingNodegroupState

	for _, condition := range conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				state = privatecloudv1alpha1.ActiveNodegroupState
			} else {
				state = privatecloudv1alpha1.UpdatingNodegroupState
			}
			reason = condition.Reason
			message = condition.Message

			lastUpdate = &condition.LastTransitionTime
		}
	}

	return state, lastUpdate, reason, message
}

// getClusterState defines the cluster state based on the Rancher cluster conditions and transitioning field.
// Rancher does not provide an error condition type so we need to use transitioning field that specifies
// if there is an error.
func getClusterState(transitioning string, conditions []manv3.ClusterCondition) (privatecloudv1alpha1.NodegroupState, *metav1.Time, string, string) {
	lastUpdate := &metav1.Time{Time: time.Now()}
	reason := ""
	message := ""
	state := privatecloudv1alpha1.UpdatingNodegroupState

	for _, condition := range conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				state = privatecloudv1alpha1.ActiveNodegroupState
			} else {
				state = privatecloudv1alpha1.UpdatingNodegroupState
			}
			reason = condition.Reason
			message = condition.Message

			lastUpdateTime, err := time.Parse(time.RFC3339, condition.LastUpdateTime)
			if err == nil {
				lastUpdate = &metav1.Time{Time: lastUpdateTime}
			}

			break
		}
	}

	if transitioning == "error" {
		state = privatecloudv1alpha1.ErrorNodegroupState
	}

	return state, lastUpdate, reason, message
}
func (p *RancherProvider) CreateClusterSecret(ctx context.Context, secret *corev1.Secret) error {
	return nil
}
