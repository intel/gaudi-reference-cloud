// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iks

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"

	certv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"
)

const (
	clusterReadyzEndpoint         = "/readyz"
	clusterLivezEndpoint          = "/livez"
	controlplaneDesiredCount      = 3
	bootstrapTokenSecretNamespace = "kube-system"
	ipAddressesAnnotation         = "nodegroup.devcloud.io/ip-addresses"
	kubernetesSignerName          = "kubernetes.io/kubelet-serving"
	nodeProviderName              = "Harvester"
	autoRepairDisabledLabelKey    = "iks.cloud.intel.com/autorepair"

	// Force deletion of pods not managed by a controller like deployment.
	// drainForce = true
	// 0 means delete immediately
	// drainGracePeriodSeconds              = 0
	// drainIgnoreAllDaemonSets             = true
	// drainDeleteEmptyDirData              = true
	// drainTimeout                         = time.Second * 10
	// drainDisableEviction                 = true
	// drainSkipWaitForDeleteTimeoutSeconds = 1
)

type IKSProvider struct {
	ControlplaneBootstrapScript      string
	WorkerBootstrapScript            string
	ClusterDeleted                   bool
	KubernetesClient                 *kubernetes.Clientset
	CaCertExpirationPeriod           time.Duration
	ControlPlaneCertExpirationPeriod time.Duration
}

func NewIKSProvider(controlplaneBootstrapScript string, workerBootstrapScript string, clusterDeleted bool, kubernetesClient *kubernetes.Clientset, caCertExpirationPeriod time.Duration, controlPlaneCertExpirationPeriod time.Duration) (*IKSProvider, error) {
	return &IKSProvider{
		ControlplaneBootstrapScript:      controlplaneBootstrapScript,
		WorkerBootstrapScript:            workerBootstrapScript,
		ClusterDeleted:                   clusterDeleted,
		KubernetesClient:                 kubernetesClient,
		CaCertExpirationPeriod:           caCertExpirationPeriod,
		ControlPlaneCertExpirationPeriod: controlPlaneCertExpirationPeriod,
	}, nil
}

func (p *IKSProvider) InitCluster(ctx context.Context, secret *corev1.Secret, cluster *privatecloudv1alpha1.Cluster, etcdLB string, apiserverLB string, publicApiserverLB string, konnectivityLB string, etcdLBPort int, apiserverLBPort int, publicApiserverLBPort int) error {
	if cluster.Spec.NodeProvider == nodeProviderName {
		cpIps, ok := cluster.ObjectMeta.Annotations[ipAddressesAnnotation]
		if !ok {
			return fmt.Errorf("unable to find controlplane ips annotation: %s", ipAddressesAnnotation)
		}

		cpIpsList := strings.Split(cpIps, ",")
		if len(cpIpsList) < controlplaneDesiredCount {
			return fmt.Errorf("controlplane ips annotation must have at least %d ips", controlplaneDesiredCount)
		}
	}

	// Create CAs
	caPEM, caPrivateKeyPEM, err := utils.CreateCa("kubernetes-ca", p.CaCertExpirationPeriod)
	if err != nil {
		return err
	}

	ca, err := utils.ParseCert([]byte(caPEM))
	if err != nil {
		return err
	}

	caPrivateKey, err := utils.ParsePrivateKey([]byte(caPrivateKeyPEM))
	if err != nil {
		return err
	}

	etcdCaPEM, etcdCaPrivateKeyPEM, err := utils.CreateCa("etcd-ca", p.CaCertExpirationPeriod)
	if err != nil {
		return err
	}

	frontProxyCaPEM, frontProxyCaPrivateKeyPEM, err := utils.CreateCa("kubernetes-front-proxy-ca", p.CaCertExpirationPeriod)
	if err != nil {
		return err
	}

	// Create sa.key and sa.pub
	saCertConfig := utils.CertConfig{
		CommonName: "service-accounts",
		Organizations: []string{
			"kubernetes",
		},
	}

	caCertExpirationPeriod := time.Now().Add(p.CaCertExpirationPeriod)
	_, saPrivateKeyPEM, err := utils.CreateAndSignCert(ca, caPrivateKey, saCertConfig, &caCertExpirationPeriod)
	if err != nil {
		return err
	}

	saPrivateKey, err := utils.ParsePrivateKey(saPrivateKeyPEM)
	if err != nil {
		return err
	}

	saPublicKeyPEM, err := utils.GetPublicKey(saPrivateKey)
	if err != nil {
		return err
	}

	cpCertExpirationPeriod := int(p.ControlPlaneCertExpirationPeriod.Hours() / 24)
	secret.Data["ca.crt"] = []byte(caPEM)
	secret.Data["ca.key"] = []byte(caPrivateKeyPEM)
	secret.Data["etcd-ca.crt"] = []byte(etcdCaPEM)
	secret.Data["etcd-ca.key"] = []byte(etcdCaPrivateKeyPEM)
	secret.Data["front-proxy-ca.crt"] = []byte(frontProxyCaPEM)
	secret.Data["front-proxy-ca.key"] = []byte(frontProxyCaPrivateKeyPEM)
	secret.Data["sa.key"] = []byte(saPrivateKeyPEM)
	secret.Data["sa.pub"] = []byte(saPublicKeyPEM)

	secret.Data["controlplane-registration-cmd"] = []byte(
		fmt.Sprintf("bash /usr/local/bin/bootstrap.sh --ca-cert %s --ca-key %s --etcd-ca-cert %s --etcd-ca-key %s --front-proxy-ca-cert %s --front-proxy-ca-key %s --sa-private-key %s --sa-public-key %s --etcd-lb %s --etcd-lb-port %d --apiserver-lb %s --apiserver-lb-port %d --public-apiserver-lb %s --public-apiserver-lb-port %d --konnectivity-lb %s --cluster-name %s --cluster-cidr %s --service-cidr %s --cp-cert-expiration-period %d",
			base64.StdEncoding.EncodeToString([]byte(caPEM)),
			base64.StdEncoding.EncodeToString([]byte(caPrivateKeyPEM)),
			base64.StdEncoding.EncodeToString([]byte(etcdCaPEM)),
			base64.StdEncoding.EncodeToString([]byte(etcdCaPrivateKeyPEM)),
			base64.StdEncoding.EncodeToString([]byte(frontProxyCaPEM)),
			base64.StdEncoding.EncodeToString([]byte(frontProxyCaPrivateKeyPEM)),
			base64.StdEncoding.EncodeToString([]byte(saPrivateKeyPEM)),
			base64.StdEncoding.EncodeToString([]byte(saPublicKeyPEM)),
			etcdLB,
			etcdLBPort,
			apiserverLB,
			apiserverLBPort,
			publicApiserverLB,
			publicApiserverLBPort,
			konnectivityLB,
			cluster.Name,
			cluster.Spec.Network.PodCIDR,
			cluster.Spec.Network.ServiceCIDR,
			cpCertExpirationPeriod,
		))

	// TODO: --containerd-envvars should be set from nodegroup controller once runtime args is supported.
	secret.Data["worker-registration-cmd"] = []byte(
		fmt.Sprintf("bash /usr/local/bin/bootstrap.sh --ca-cert %s --apiserver-lb %s --apiserver-lb-port %d --cluster-dns %s --containerd-envvars ''",
			base64.StdEncoding.EncodeToString([]byte(caPEM)),
			apiserverLB,
			apiserverLBPort,
			cluster.Spec.Network.ClusterDNS,
		))

	return nil
}

func (p *IKSProvider) GetCluster(ctx context.Context) (*privatecloudv1alpha1.ClusterStatus, error) {
	restClient := p.KubernetesClient.RESTClient()

	var clusterStatus privatecloudv1alpha1.ClusterStatus
	clusterStatus.Nodegroups = make([]privatecloudv1alpha1.NodegroupStatus, 0)
	clusterStatus.State = privatecloudv1alpha1.ActiveClusterState
	clusterStatus.Reason = ""
	clusterStatus.Message = ""
	clusterStatus.LastUpdate = metav1.Time{Time: time.Now()}

	if err := utils.GetKubernetesAPIHealthEndpoint(ctx, clusterReadyzEndpoint, restClient); err != nil {
		clusterStatus.State = privatecloudv1alpha1.ErrorClusterState
		clusterStatus.Reason = "KubernetesNotReady"
		clusterStatus.Message = err.Error()
		return &clusterStatus, nil
	}

	if err := utils.GetKubernetesAPIHealthEndpoint(ctx, clusterLivezEndpoint, restClient); err != nil {
		clusterStatus.State = privatecloudv1alpha1.ErrorClusterState
		clusterStatus.Reason = "KubernetesNotLived"
		clusterStatus.Message = err.Error()
		return &clusterStatus, nil
	}

	return &clusterStatus, nil
}

func (p *IKSProvider) CleanUpCluster(context.Context, string) error {
	return nil
}

func (p *IKSProvider) GetNodes(ctx context.Context, nodegroupName string) ([]privatecloudv1alpha1.NodeStatus, error) {
	nodesList, err := p.KubernetesClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]privatecloudv1alpha1.NodeStatus, 0, len(nodesList.Items))
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
			nodeStatus.Unschedulable = node.Spec.Unschedulable

			nodes = append(nodes, nodeStatus)
		}
	}

	return nodes, nil
}

func (p *IKSProvider) GetNode(ctx context.Context, nodeName string) (privatecloudv1alpha1.NodeStatus, error) {
	log := log.FromContext(ctx).WithName("IKSProvider.GetNode")

	var nodeStatus privatecloudv1alpha1.NodeStatus

	node, err := p.KubernetesClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nodeStatus, err
	}
	log.V(0).Info("Node status from iks provider", logkeys.NodeName, node.Name, logkeys.NodeStatus, node.Status)

	// We get the node status information based on node name
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
	nodeStatus.Unschedulable = node.Spec.Unschedulable

	if autoRepair, ok := node.ObjectMeta.GetLabels()[autoRepairDisabledLabelKey]; ok {
		autoRepairValue, err := strconv.ParseBool(autoRepair)
		if err != nil {
			log.V(0).Error(err, logkeys.Error, logkeys.AutoRepairDisabledLabelKey, autoRepairDisabledLabelKey, logkeys.AutoRepairValue, autoRepairValue)
		} else {
			log.V(0).Info("Reading Auto Repair Value from IKS provider for node name", logkeys.NodeName, node.Name, logkeys.AutoRepairDisabledLabelKey, autoRepairDisabledLabelKey, logkeys.AutoRepairValue, autoRepairValue)
			if !autoRepairValue {
				nodeStatus.AutoRepairDisabled = true
			}
		}
	}

	return nodeStatus, nil
}

func (p *IKSProvider) DeleteNode(ctx context.Context, nodeName string) error {
	if p.ClusterDeleted {
		return nil
	}

	if err := p.KubernetesClient.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (p *IKSProvider) DrainNode(ctx context.Context, nodeName string) error {
	if p.ClusterDeleted {
		return nil
	}

	node, err := p.KubernetesClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	// If node not found, nothing to do, so let's return.
	if k8serrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	// Cordon
	cordonHelper := drain.NewCordonHelper(node)
	if cordonHelper.UpdateIfRequired(true) {
		err, patchErr := cordonHelper.PatchOrReplaceWithContext(ctx, p.KubernetesClient, false)
		if patchErr != nil {
			return patchErr
		}

		if err != nil {
			return err
		}
	}

	// Drain
	// drainHelper := drain.Helper{
	// 	Ctx:                             ctx,
	// 	Client:                          client,
	// 	Force:                           drainForce,
	// 	GracePeriodSeconds:              drainGracePeriodSeconds,
	// 	IgnoreAllDaemonSets:             drainIgnoreAllDaemonSets,
	// 	Timeout:                         drainTimeout,
	// 	DeleteEmptyDirData:              drainDeleteEmptyDirData,
	// 	DisableEviction:                 drainDisableEviction,
	// 	SkipWaitForDeleteTimeoutSeconds: drainSkipWaitForDeleteTimeoutSeconds,
	// }

	// pods, errs := drainHelper.GetPodsForDeletion(nodeName)
	// if errs != nil {
	// 	return utilerrors.NewAggregate(errs)
	// }

	// if err := drainHelper.DeleteOrEvictPods(pods.Pods()); err != nil {
	// 	return err
	// }

	return nil
}

func (p *IKSProvider) GetBootstrapScript(nodegroupType privatecloudv1alpha1.NodegroupType) (string, error) {
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

func (p *IKSProvider) CreateBootstrapTokenSecret(ctx context.Context, secret *corev1.Secret) error {
	if _, err := p.KubernetesClient.CoreV1().Secrets(bootstrapTokenSecretNamespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func (p *IKSProvider) ApproveKubeletServingCertificateSigningRequests(ctx context.Context, nodeNamePrefix string) error {
	log := log.FromContext(ctx).WithName("IKSProvider.ApproveKubeletServingCertificateSigningRequests")

	csrList, err := p.KubernetesClient.CertificatesV1().CertificateSigningRequests().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, csr := range csrList.Items {
		// username contains the name of the node requesting the csr. The node name is
		// created using nodegroup name as a prefix, so this permits a nodegroup to only
		// approve csrs of nodes that belong to it. Example of username: system:node:ng-ntat6s264e-ig-12345-0 or system:node:ng-ntat6s264e-12345
		if strings.Contains(csr.Spec.Username, nodeNamePrefix) && csr.Spec.SignerName == kubernetesSignerName {
			if len(csr.Status.Conditions) == 0 {
				log.V(0).Info("Approving csr", logkeys.Name, csr.Name)
				csr.Status.Conditions = append(csr.Status.Conditions, certv1.CertificateSigningRequestCondition{
					Type:           certv1.CertificateApproved,
					Reason:         "operatorApproval",
					Message:        "This csr is approved by the operator",
					Status:         corev1.ConditionTrue,
					LastUpdateTime: metav1.Now(),
				})

				if _, err := p.KubernetesClient.CertificatesV1().CertificateSigningRequests().UpdateApproval(ctx, csr.Name, &csr, metav1.UpdateOptions{}); err != nil {
					return err
				}
			} else {
				log.V(0).Info("Approved csr", logkeys.Name, csr.Name)
			}
		}
	}

	return nil
}

func (p *IKSProvider) getNodeState(conditions []corev1.NodeCondition) (privatecloudv1alpha1.NodegroupState, *metav1.Time, string, string) {
	lastUpdate := &metav1.Time{Time: time.Now()}
	reason := ""
	message := ""
	state := privatecloudv1alpha1.UpdatingNodegroupState

	for _, condition := range conditions {
		if condition.Type == "Ready" {
			reason = condition.Reason

			if condition.Status == "True" {
				state = privatecloudv1alpha1.ActiveNodegroupState
				message = "Node ready"
			} else {
				state = privatecloudv1alpha1.UpdatingNodegroupState
				message = "Configuring node"
			}

			lastUpdate = &condition.LastTransitionTime
		}
	}

	return state, lastUpdate, reason, message
}

func (p *IKSProvider) CreateNamespace(ctx context.Context, namespace *corev1.Namespace) error {
	_, err := p.KubernetesClient.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (p *IKSProvider) CreateSecret(ctx context.Context, namespace string, secret *corev1.Secret) error {
	_, err := p.KubernetesClient.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (p *IKSProvider) GetSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	secret, err := p.KubernetesClient.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}
