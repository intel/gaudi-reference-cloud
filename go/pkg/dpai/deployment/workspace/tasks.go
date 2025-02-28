// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package workspace

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/config"
	model "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/dns"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/helm"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/networking"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"

	// sig controller for custom operator ligecycle crd addition in k8s clientset schema
	"sigs.k8s.io/controller-runtime/pkg/client"
	// dns and highwire
)

// struct to pass the output between the tasks
type TaskOutput struct {
	// Make sure below fields are defined for all the deployments
	CloudAccountID       string
	WorkspaceID          string
	IksClusterName       string
	IksClusterUUID       string
	SshKeyName           string
	SecretNamespace      string
	IstioNamespace       string
	CertManagerNamespace string
	OpenEbsNamespace     string
	NodeGroupID          string

	// Nodegorups info
	Nodes                  []*pb.NodeStatus
	NodeNames              map[string]string // name --> dns-name
	NodeCount              int32
	NodeGroupName          string
	NodegorupSelectorLabel string

	// gateway-secret map // maps cert manager secret to gateway
	GatewaySecretMap map[string]string

	// use for the sequential createion from workspace to individual dpai services
	LBHwDnsARecord string
	DnsFqdn        string
}

type UpdateTaskOutput struct {
	// use for the sequential createion from workspace to individual dpai services
	CloudAccountID string
	WorkspaceID    string
	IksClusterName string
	IksClusterUUID string
	LBHwDnsARecord string
	DnsFqdn        string
}

type DeleteTaskOutput struct {
	CloudAccountID string
	WorkspaceID    string
	IksClusterName string
	IksClusterUUID string

	LBHwDnsARecord string
	DnsFqdn        string

	GatewayIstioName       string
	GatewayIstioSecretName string
	GatewayIstioLabelName  string
}

const (
	CLUSTER_ISSUER = "ClusterIssuer"
	ISSUER         = "Issuer"
	CERTIFICATE    = "Certificate"
)

// loadbalancing modes
const (
	ROUND_ROBIN_MODE              = "round-robin"
	LEAST_CONNECTIONS_MODE        = "least-connections"
	LEAST_CONNECTIONS_MEMBER_MODE = "least-connections-member"
	EEP_ALIVE_MODE                = "keep-alive"
	MONITOR_TCP                   = "tcp"
)

type DeploymentTaskOutputs interface {
	TaskOutput | UpdateTaskOutput | DeleteTaskOutput
}

func parseTaskOutput[T DeploymentTaskOutputs](output any) (T, error) {
	switch any(output).(type) {
	case TaskOutput:
		parsed, ok := output.(TaskOutput)
		if !ok {
			return *new(T), fmt.Errorf("Type Assertion failed for Create Deployment Output: %T", output)
		}
		return any(parsed).(T), nil
	case UpdateTaskOutput:
		parsed, ok := output.(UpdateTaskOutput)
		if !ok {
			return *new(T), fmt.Errorf("Type Assertion failed for Update Deployment Output: %T", output)
		}
		return any(parsed).(T), nil
	case DeleteTaskOutput:
		parsed, ok := output.(DeleteTaskOutput)
		if !ok {
			return *new(T), fmt.Errorf("Type Assertion failed for Delete Deployment Output: %T", output)
		}
		return any(parsed).(T), nil
	default:
		return *new(T), fmt.Errorf("Type Assertion failed for Deployment Output: %T", output)
	}
}

func mergeTaskOutput(t1, t2 TaskOutput) TaskOutput {
	merged := t1

	// Use reflection to iterate over the fields of the struct
	t2Value := reflect.ValueOf(t2)
	mergedValue := reflect.ValueOf(&merged).Elem()

	for i := 0; i < t2Value.NumField(); i++ {
		field := t2Value.Field(i)
		mergedField := mergedValue.Field(i)

		switch field.Kind() {
		case reflect.String:
			if field.Len() > 0 {
				mergedField.SetString(field.String())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() != 0 {
				mergedField.SetInt(field.Int())
			}
		case reflect.Bool:
			mergedField.SetBool(field.Bool())
		// Add more cases for other types if needed
		default:
			log.Printf("mergeTaskOutput not able to handle the field of type %s", field.Kind())
		}
	}

	return merged
}

func CreateIKSCluster(ctx *deployment.TaskRunContext) (any, error) {

	input, _ := GetDeploymentCreateInput(ctx.DeploymentContext)
	log.Printf("Trying to create a IKS cluster: %v", ctx.DeploymentContext.Name)
	log.Printf("Task Input: %+v", ctx.Input)
	log.Printf("Task DeploymentContext: %+v", ctx.DeploymentContext)
	log.Printf("Task DeploymentContext input type and value: %T : %+v", input, input)

	clusterName := k8s.GenerateDpaiIksClusterName(input.Name)
	request := pb.ClusterRequest{
		Name:           clusterName,
		Description:    &input.Description,
		K8Sversionname: "1.29",
		Runtimename:    "Containerd",
		InstanceType:   "iks-cluster",
		Annotations: []*pb.Annotations{
			{Key: "workspaceName", Value: input.Name},
			{Key: "cloudAccountId", Value: input.CloudAccountId},
		},
		CloudAccountId: input.CloudAccountId,
	}
	log.Printf("Created IKS cluster: %T: %+v", &request, &request)

	cluster, err := ctx.K8sClient.CreateIKSCluster(&request)

	if err != nil {
		return nil, err
	}

	log.Printf("Getting task3 output from fn: %+v", cluster)
	log.Printf("Getting task3 output from fn (not exists): %+v", ctx.GetTaskOutput("Hello5"))

	TaskOutput := TaskOutput{
		IksClusterName: clusterName,
		CloudAccountID: input.CloudAccountId,
		WorkspaceID:    ctx.DeploymentContext.Context.WorkspaceId,
		IksClusterUUID: cluster.Uuid,
	}

	_, err = ctx.SqlModel.UpdateWorkspaceDeploymentStatus(context.Background(), model.UpdateWorkspaceDeploymentStatusParams{
		ID:             pgtype.Text{String: ctx.DeploymentContext.Context.WorkspaceId, Valid: true},
		IksID:          pgtype.Text{String: cluster.Uuid, Valid: true},
		IksClusterName: pgtype.Text{String: clusterName, Valid: true},
		DeploymentID:   pgtype.Text{String: ctx.DeploymentContext.ID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("error updating the workspace deployment status. %+v", err)
	}

	return TaskOutput, nil
}

func ValidateIKSCluster(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside ValidateIKSCluster \n")
	var K8sAzClient k8s.K8sAzClient = k8s.K8sAzClient{}

	azK8sClient, err := K8sAzClient.ConfigureK8sClient()
	if err != nil {
		log.Printf("Error getting the AZ clientset for IKS AZ cluster %+v", err)
		return nil, err
	}
	log.Println("clsuter info for az cluster is ", azK8sClient)

	ns, err := azK8sClient.ClientSet.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})

	if err != nil {
		fmt.Println("Error Fetching all the namespaces")
		return nil, err
	}

	/*
		Verify IKS Lb operator is there
		Verify IKS Fw operator is there
	*/

	var iksAzOperators []string = make([]string, 0) // required and mandate to have them to proceed with workspace install
	for _, ns := range ns.Items {
		if ns.Name == "idcs-system" {
			pods := azK8sClient.ClientSet.CoreV1().Pods(ns.Name)
			pd, err := pods.List(context.Background(), metav1.ListOptions{})
			if err != nil {
				fmt.Println("Error getting pods")
			} else {
				for _, pod := range pd.Items {
					if strings.Contains(pod.Name, "loadbalancer") || strings.Contains(pod.Name, "firewall") {
						iksAzOperators = append(iksAzOperators, pod.Name)
					}
				}
			}
		}
	}

	time.Sleep(5 * time.Second)
	if len(iksAzOperators) == 2 {
		log.Printf("Verify Checks passed for IKS Cluster in AZ Zone all required IKS operators are installed")
	}
	log.Printf("Completed: Inside ValidateIKSCluster ")
	return true, nil
}

func CreateSshKey(ctx *deployment.TaskRunContext) (any, error) {

	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("CreateSecretNamespace"))
	if err != nil {
		return nil, err
	}
	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	privateKey, publicKey, err := utils.GenerateSshKey(output.IksClusterName)
	if err != nil {
		return nil, err
	}

	// TODO: Save the private key in the K8s Secrets
	log.Printf("Private key:\n%s\n", privateKey)
	secretData := map[string][]byte{
		"private-key": privateKey,
		"public-key":  publicKey,
	}
	secretName := utils.GenerateDpaiSshKeySecretName(output.IksClusterName)
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		log.Printf("Error creating the IKS client. %+v", err)
		return nil, err
	}
	err = ctx.K8sClient.CreateSecret(deployment.SecretNamespace, secretName, secretData)
	if err != nil {
		log.Printf("Error creating the secret for the sshKey. %+v", err)
		return nil, err
	}

	request, err := ctx.K8sClient.SshClient.Create(context.Background(), &pb.SshPublicKeyCreateRequest{
		Metadata: &pb.ResourceMetadataCreate{
			CloudAccountId: output.CloudAccountID,
			Name:           output.IksClusterName,
		},
		Spec: &pb.SshPublicKeySpec{
			SshPublicKey: string(publicKey),
			OwnerEmail:   fmt.Sprintf("%s@dpaiworkspace.com", output.IksClusterName),
		},
	})

	if err != nil {
		log.Printf("Error adding the sshKey. %+v", err)
		return nil, err
	}

	output.SshKeyName = request.Metadata.Name

	// _, err = ctx.SqlModel.UpdateWorkspaceDeploymentStatus(context.Background(), model.UpdateWorkspaceDeploymentStatusParams{
	// 	ID:         pgtype.Text{String: ctx.DeploymentContext.Context.WorkspaceId, Valid: true},
	// 	SshKeyName: pgtype.Text{String: request.Metadata.Name, Valid: true},
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("error updating the workspace deployment status. %+v", err)
	// }

	return output, nil
}

func CreateSecretNamespace(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside CreateSecretNamespace using CreateIKS Output: %v \n", ctx.GetTaskOutput("CreateIKSCluster"))

	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("CreateIKSCluster"))
	if err != nil {
		return nil, err
	}
	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}
	namespace, err := ctx.K8sClient.CreateNamespace(deployment.SecretNamespace, false, map[string]string{})
	if err != nil {
		return nil, err
	}
	log.Printf("Completed: Inside CreateSecretNamespace \n")
	output.SecretNamespace = namespace.Name
	return output, nil
}

func CreateNodeGroup(ctx *deployment.TaskRunContext) (any, error) {
	// know the assertion for this type is alwways taskoutput for this wqorkspace
	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("CreateSshKey"))
	if err != nil {
		return nil, err
	}
	fmt.Println("the task output passed createnodegroup ", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	description := "This node group is created and managed by DPAI -- Please do not delete this -- Deleting this will break the DPAI services running on top of it"

	nodeGroup, err := ctx.K8sClient.CreateNodeGroup(&pb.CreateNodeGroupRequest{
		CloudAccountId: output.CloudAccountID,
		Clusteruuid:    output.IksClusterUUID,
		Name:           "control-plane",
		InstanceType:   "iks-cluster",
		Description:    &description,
		Instancetypeid: "vm-spr-sml",
		Count:          1,
		Sshkeyname: []*pb.SshKey{
			{Sshkey: output.SshKeyName},
		},
	})

	if err != nil {
		return nil, err
	}

	log.Printf("Getting task output from fn: %+v", nodeGroup)

	output.NodeGroupID = nodeGroup.Nodegroupuuid

	// _, err = ctx.SqlModel.UpdateWorkspaceDeploymentStatus(context.Background(), model.UpdateWorkspaceDeploymentStatusParams{
	// 	ID:                    pgtype.Text{String: ctx.DeploymentContext.Context.WorkspaceId, Valid: true},
	// 	ManagementNodegroupID: pgtype.Text{String: nodeGroup.Nodegroupuuid, Valid: true},
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("error updating the workspace deployment status. %+v", err)
	// }

	return output, nil
}

func CreateCertManagerNS(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside CreateSecretNamespace using CreateIKS Output for Cert-Manager: ")
	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("CreateNodeGroup"))
	if err != nil {
		return nil, err
	}

	fmt.Println("the task output passed create cert-manager ns ", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	ns, err := ctx.K8sClient.CreateNamespace(deployment.CertManagerNamesapce, true, map[string]string{})
	if err != nil {
		return nil, err
	}

	log.Printf("Completed: Inside Istio Operator Namespace Creation for Istio-Operator \n")
	output.CertManagerNamespace = ns.Name

	return output, nil
}

func getCertManagerHelmClient(ctx *deployment.TaskRunContext, namespace string) (helmclient.Client, error) {
	log.Printf("Fetching the Helm Chart for Cert manager")

	helmClient, err := helm.GetHelmClient(ctx.K8sClient.ClientConfig, namespace)
	if err != nil {
		log.Printf("Error in installing cert manager %s", err.Error())
		return nil, err
	}

	chartRepo := repo.Entry{
		Name: "jetstack",
		URL:  "https://charts.jetstack.io",
	}

	if err = helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		log.Fatalf("Error Adding Chart Repo %+v. Error message: %+v", chartRepo, err)
		return nil, err
	}

	return helmClient, nil
}

func InstallCertManagerHelmChart(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Installing the Helm chart for Cert Manager")

	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("CreateCertManagerNS"))
	if err != nil {
		return nil, err
	}

	fmt.Println("the task output passed create install cert manager ns ", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	fmt.Println("[x] The task output for the install cert manager helm chart is ", output)

	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	certManagerHelmClient, err := getCertManagerHelmClient(ctx, output.CertManagerNamespace)

	if err != nil {
		log.Fatal("Error Fetching the Helm client for Cert Manager")
		return nil, err
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName: "cert-manager",
		ChartName:   "jetstack/cert-manager",
		Namespace:   output.CertManagerNamespace,
		ValuesOptions: values.Options{
			Values: []string{
				"featureGates=ExperimentalCertificateSigningRequestControllers=true",
				"installCRDs=true"},
		},
		UpgradeCRDs: true,
		Wait:        true,
		Atomic:      true,
		Timeout:     5 * time.Minute,
	}

	deployedChart, err := certManagerHelmClient.InstallOrUpgradeChart(context.TODO(), &chartSpec, &helmclient.GenericHelmOptions{})
	if err != nil {
		log.Fatalf("Error installing Cert manager Helm chart %+v", err)
		return nil, err
	}

	log.Println("Waiting for the deployment to complete")
	time.Sleep(time.Second * 15)
	if deployedChart.Info.Status != release.StatusDeployed {
		log.Fatal("Error the Helm Chart for Cert Manager not successfully deployed")
		return nil, fmt.Errorf("Error the deployed status for the chart is not healthy")
	}
	return output, nil
}

func getCertManagerClient(ctx *deployment.TaskRunContext) (client.Client, error) {
	log.Printf("Registering the Controller Runtime client with cert-manager client")
	cManagerScheme := runtime.NewScheme()

	err := certmanagerv1.AddToScheme(cManagerScheme)
	if err != nil {
		return nil, fmt.Errorf("Error adding the cert-manager scheme %+v", err)
	}

	cManagerClient, err := client.New(ctx.K8sClient.ClientConfig, client.Options{Scheme: cManagerScheme})
	if err != nil {
		return nil, fmt.Errorf("Failed to create cert-manager client: %+v", err)
	}

	return cManagerClient, nil
}

// poll the clusterIssuer resource until it is ready
// TODO: Need to add Poll client go based on k8s rate limit and a deadline / cancellable deadline context
func waitClusterIssuer(kubeClient client.Client, name, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(time.Second*2, timeout, func() (bool, error) {
		log.Printf("Polling the ClusterIssuer %s for readiness of the resource", name)
		resource := &certmanagerv1.ClusterIssuer{}
		err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: name, Namespace: namespace}, resource)
		if err != nil {
			return false, err // keep retrying to poll the resource
		}

		for _, condition := range resource.Status.Conditions {
			if condition.Status == cmmeta.ConditionStatus(metav1.ConditionTrue) {
				return true, nil // ClusterIssuer is ready
			}
		}
		return false, nil // Not ready yet, retry
	})
}

// poll the certificate resource until it is ready
func waitForCertificateReady(kubeClient client.Client, name, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(time.Second*1, timeout, func() (bool, error) {
		log.Printf("Polling the Certificate %s for readiness of the resource", name)
		cert := &certmanagerv1.Certificate{}
		err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: name, Namespace: namespace}, cert)
		if err != nil {
			return false, err // Retry on error
		}

		for _, condition := range cert.Status.Conditions {
			if condition.Status == cmmeta.ConditionStatus(metav1.ConditionTrue) {
				return true, nil // Certificate is ready
			}
		}
		return false, nil // Not ready yet, retry
	})
}

// poll to check if the root secret is created for self-signed PKI
func waitRootK8sCaSecret(kubeClient *kubernetes.Clientset, name, namespace string, timeout time.Duration) error {
	/*
		Poll the resource if the root secret is not created for  signing the istio gateway certs, the entire PKI wont work
		Reprocess and create the PKI to make sure it is alywas ready
	*/
	return wait.PollImmediate(time.Second, timeout, func() (done bool, err error) {
		log.Printf("Polling the self signed PKI TLS secret %s for readiness of the resource", name)

		_, err = kubeClient.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return false, err // the root secret is yet not created
		}
		return true, nil
	})
}

func CreateSelfSignedPKI(ctx *deployment.TaskRunContext, certManagerClient client.Client, brokenCount int, retryLimit int) error {
	if brokenCount == retryLimit {
		err := "The self signed PKI is not created, please reprocess the cluster creation or rerun the depoloyment"
		// need the same principle how reconcilation requeue works to process in controller runtime
		return fmt.Errorf("%+v", err)
	}

	var markforReque bool = false
	rootIssuer := &certmanagerv1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployment.IstioRootClusterIssuer,
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				SelfSigned: &certmanagerv1.SelfSignedIssuer{},
			},
		},
	}

	rootSelfSignedIssuerErr := certManagerClient.Create(context.Background(), rootIssuer)

	if rootSelfSignedIssuerErr != nil {
		log.Printf("Error creating the root CA Cluster Issuuer of type Self Signed for Istio %+v", rootSelfSignedIssuerErr)
		markforReque = true
	}

	log.Printf("The Root Self Signed Issuer is Ready with name %s", deployment.IstioRootClusterIssuer)
	statusPoll := waitClusterIssuer(certManagerClient, deployment.IstioRootClusterIssuer, deployment.CertManagerNamesapce, 30*time.Minute)

	if statusPoll != nil {
		log.Fatalf("Error the status while polling for ClusterIssuer readiness %+v", statusPoll)
		panic(statusPoll.Error())
	}
	rootIstioCa := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.IstioRootClusterIssuerCertificate,
			Namespace: deployment.CertManagerNamesapce,
		},
		Spec: certmanagerv1.CertificateSpec{
			IsCA:       true,
			CommonName: deployment.IstioRootClusterIssuerCertificate,
			SecretName: deployment.CLUSTER_ROOT_SECRET, // keeping this static as this is the root secret to govern all tls at edge and mtls internal in istio
			PrivateKey: &certmanagerv1.CertificatePrivateKey{
				Algorithm: certmanagerv1.ECDSAKeyAlgorithm,
				Size:      1 << 8,
			},
			IssuerRef: cmmeta.ObjectReference{
				Name:  rootIssuer.ObjectMeta.Name,
				Kind:  deployment.CLUSTER_ISSUER,
				Group: deployment.CLUSTER_ISSUER_API_VERSION,
			},
		},
	}

	rootCAIssuerErr := certManagerClient.Create(context.Background(), rootIstioCa)

	if rootCAIssuerErr != nil {
		log.Printf("Error creating the root CA Cluster Issuuer of type Self Signed for Istio %+v", rootCAIssuerErr)
		markforReque = true
	}

	statusPoll = waitForCertificateReady(certManagerClient, deployment.IstioRootClusterIssuerCertificate, deployment.CertManagerNamesapce, 30*time.Minute)

	if statusPoll != nil {
		log.Fatalf("Error the status while polling for Certificate readiness %+v", statusPoll)
		return statusPoll
	}

	log.Printf("The Root Certificate Signed Self Signed issuer is Ready %s", deployment.IstioRootClusterIssuerCertificate)
	istioClusterIssuer := &certmanagerv1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployment.IstioK8SCSR,
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				CA: &certmanagerv1.CAIssuer{
					SecretName: rootIstioCa.Spec.SecretName,
				},
			},
		},
	}

	istioCAIssuer := certManagerClient.Create(context.Background(), istioClusterIssuer)

	if istioCAIssuer != nil {
		log.Printf("Error creating the root CA Cluster Issuuer of type Self Signed for Istio %+v", rootSelfSignedIssuerErr)
		markforReque = true
	}

	statusPoll = waitClusterIssuer(certManagerClient, deployment.IstioK8SCSR, deployment.CertManagerNamesapce, 10*time.Minute)

	if statusPoll != nil {
		log.Fatalf("Error the status for cluster issuer is not ready")
		markforReque = true
	}

	rootSignerCertPoll := waitRootK8sCaSecret(ctx.K8sClient.ClientSet, deployment.CLUSTER_ROOT_SECRET, deployment.CertManagerNamesapce, 10*time.Second)

	if rootSignerCertPoll != nil {
		log.Fatalf("The PKI is not healthy, no root secret has been created to sign the Gateway Certs")
		markforReque = true
	}

	cleanUpBrokenPKI := func() error {
		// delte the broken PKI to make sure the self signed PKI is healthy to proceed further

		err := certManagerClient.Delete(context.Background(), rootIssuer)
		if err != nil {
			return fmt.Errorf("Error fixing the broken PKI please retrun the deployment")
		}

		err = certManagerClient.Delete(context.Background(), rootIstioCa)
		if err != nil {
			return fmt.Errorf("Error fixing the broken PKI please retrun the deployment")
		}

		err = certManagerClient.Delete(context.Background(), istioClusterIssuer)

		if err != nil {
			return fmt.Errorf("Error fixing the broken PKI please retrun the deployment")
		}

		return nil
	}

	if !markforReque {
		log.Println("Creating the Self Signed PKI for ingress nodegroup successeded in the cluster")
		return nil
	}
	log.Println("Creating the Self Signed PKI for ingress nodegroup failed marking  for retry ", brokenCount)
	err := cleanUpBrokenPKI()
	if err != nil {
		log.Println("Error creating the PKI the please re run the deployment")
		return err
	}

	// recursively call until the PKI is not successfully created
	// enerate process same way how the runtime queue re process the request.
	return CreateSelfSignedPKI(ctx, certManagerClient, brokenCount+1, retryLimit)
}

// defines and cretes the root CA sign for the ingress istio edge controller of the cluster
func CreateIstioGatewayCerts(ctx *deployment.TaskRunContext) (any, error) {
	log.Println("Creating the Istio Gateway Self Signed certs signed by Cert Manager self-signed issuer")
	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("InstallCertManagerHelmChart"))
	if err != nil {
		return nil, err
	}

	fmt.Println("the task output passed create self signed istio certs pki", output)
	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	certManagerClient, err := getCertManagerClient(ctx)

	if err != nil {
		log.Printf("Error in getting the cerrt manager client %+v", err)
		return false, err
	}

	err = CreateSelfSignedPKI(ctx, certManagerClient, 0, 5)
	if err != nil {
		return nil, err
	}

	log.Printf("The CA for Istio Gateway Controller  is created with Name %s", deployment.IstioK8SCSR)
	return output, nil
}

// -------------------- Service Mesh --------------------------- //
func CreateIstioNamespace(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside Create Istio Namespace Task for TaskInput: %v \n", ctx.GetTaskOutput("CreateNodeGroup"))

	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("CreateNodeGroup"))
	if err != nil {
		return nil, err
	}
	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	// create istio-system ns second
	namespace, err := ctx.K8sClient.CreateNamespace(deployment.IstioNamespace, true, map[string]string{})
	if err != nil {
		log.Fatal("Error creating namespace for Istio System")
		return nil, err
	}

	log.Printf("Completed: Istio Namespace successfully created \n")
	output.IstioNamespace = namespace.Name
	return output, nil
}

func getIstioHelmClient(ctx *deployment.TaskRunContext, namespace string) (helmclient.Client, error) {
	log.Println("Fetching the istio helm chart")
	helmClient, err := helm.GetHelmClient(ctx.K8sClient.ClientConfig, namespace)
	if err != nil {
		return nil, err
	}

	chartRepo := repo.Entry{
		Name: "istio",
		URL:  "https://istio-release.storage.googleapis.com/charts",
	}

	// Add a chart-repository to the client.
	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		log.Printf("Error Adding Chart Repo %+v. Error message: %+v", chartRepo, err)
		return nil, err
	}

	return helmClient, nil
}

// install the istio crd's
func InstallIstioCRD(ctx *deployment.TaskRunContext) (any, error) {
	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("CreateIstioNamespace"))
	if err != nil {
		return nil, err
	}
	log.Println("The task output parsed for install istio crd is", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		log.Fatalf("Error getting Iks k8s clientSet, while installing Istio CRD")
		return nil, err
	}

	helmClient, err := getIstioHelmClient(ctx, output.IstioNamespace)
	if err != nil {
		log.Fatalf("Error getting Istio CRD Helm-Chart")
		return nil, err
	}
	helmValuesOptions := values.Options{
		Values: []string{"defaultRevision=default"},
	}
	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName:   "istio-base",
		ChartName:     "istio/base",
		Namespace:     output.IstioNamespace,
		UpgradeCRDs:   false,
		ValuesOptions: helmValuesOptions,
		Wait:          true,
		Atomic:        true,
		Timeout:       5 * time.Minute,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	deployedChart, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, &helmclient.GenericHelmOptions{})
	if err != nil {
		log.Fatalf("Error while installing Istio. %+v", err)
		return nil, err
	}

	fmt.Println(("[x] Waiting for the CRD to be installed "))
	time.Sleep(time.Second * 15)
	if deployedChart.Info.Status != release.StatusDeployed {
		log.Fatalf("Error the Helm Chart for Istio Base CRD not successfully deployed")
		return nil, nil
	}

	return output, nil
}

// install istio daemon
func InstallIstioDaemon(ctx *deployment.TaskRunContext) (any, error) {
	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("InstallIstioCRD"))
	if err != nil {
		return nil, err
	}
	fmt.Println("The task output parsed for install istio Daeoman is", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		log.Fatalf("Error getting Iks k8s clientSet, while installing IstioD")
		return nil, err
	}

	helmClient, err := getIstioHelmClient(ctx, output.IstioNamespace)
	if err != nil {
		log.Fatalf("Error getting IstioD Helm-Chart")
		return nil, err
	}

	helmValuesOptions := values.Options{
		Values: []string{},
	}

	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName:   "istiod",
		ChartName:     "istio/istiod",
		Namespace:     output.IstioNamespace,
		UpgradeCRDs:   false,
		ValuesOptions: helmValuesOptions,
		Wait:          true,
		Atomic:        true,
		Timeout:       5 * time.Minute,
	}

	log.Println("Installing the IstioD Helm Chart with values ", helmValuesOptions.JSONValues, helmValuesOptions.Values)
	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	deployedChart, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, &helmclient.GenericHelmOptions{})

	if err != nil {
		log.Fatalf("Error while installing Istio Daemon. %+v", err)
		return nil, err
	}

	log.Println(("Waiting for the IstioD to be installed "))
	time.Sleep(time.Second * 15)
	if deployedChart.Info.Status != release.StatusDeployed {
		log.Printf("Error the Helm Chart for Istio Base CRD not successfully deployed")
		return nil, nil
	}

	return output, nil
}

// Install Ingress Gateway
func InstallIstioIngressGateway(ctx *deployment.TaskRunContext) (any, error) {
	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("InstallIstioDaemon"))
	if err != nil {
		return nil, err
	}
	log.Println("The task output parsed for install istio Gateway  is", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		log.Fatalf("Error getting Iks k8s clientSet, while installing Istio Gateways")
		return nil, err
	}

	helmClient, err := getIstioHelmClient(ctx, output.IstioNamespace)
	if err != nil {
		log.Fatalf("Error getting Istio Ingress Gateways Helm-Chart")
		return nil, err
	}

	helmValuesOptions := values.Options{
		Values: []string{
			"service.type=NodePort",
			"kind=DaemonSet",
		},
	}

	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName:   "istio-ingressgateway",
		ChartName:     "istio/gateway",
		Namespace:     output.IstioNamespace,
		UpgradeCRDs:   false,
		ValuesOptions: helmValuesOptions,
		Wait:          true,
		Atomic:        true,
		Timeout:       5 * time.Minute,
	}

	deployedChart, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, &helmclient.GenericHelmOptions{})
	if err != nil {
		log.Printf("Error installing Istio Ingress Gateway Helm Chart %+v", err)
		return nil, err
	}

	log.Println(("Waiting for the gateway to be installed "))
	time.Sleep(time.Second * 15)
	if deployedChart.Info.Status != release.StatusDeployed {
		log.Printf("Error the Helm Chart for Istio Daemon not successfully deployed")
		return nil, nil
	}
	return output, nil
}

// -------------------- IKS LB Operator --------------------//
func generateIksLBCRDClient(azClientSet *k8s.K8sAzClient) (client.Client, error) {
	log.Println("Registering the Controller Runtime client for the IKS cluster Lb Resource private.cloud.intel.com/v1alpha1, kind: Loadbalancer")
	loadBalancerSchema := runtime.NewScheme()

	err := loadbalancerv1alpha1.AddToScheme(loadBalancerSchema)

	if err != nil {
		return nil, err
	}

	loadbalancerClient, err := client.New(azClientSet.ClientConfig, client.Options{Scheme: loadBalancerSchema})
	if err != nil {
		log.Fatalf("Failed to create load-balancer client in controller-runtime: %+v", err)
		return nil, err
	}

	return loadbalancerClient, nil
}

func getIngressNodePort(ctx *deployment.TaskRunContext) (int, error) {
	log.Println("Fetching the NodePort running on Ingress Nodegroup for exposing Istio Ingress Gateway")

	serviceGatewayName := "istio-ingressgateway"

	ingress, err := ctx.K8sClient.ClientSet.CoreV1().Services(deployment.IstioNamespace).Get(context.Background(),
		serviceGatewayName, metav1.GetOptions{})

	// a nodeport is usually an ephemeral tcp port above 30000 which the lb operator is built to balance traffic upon
	if err != nil {
		log.Fatalf("Error fetching the Istio Ingress Gateway Service: %+v", err)
		return -1, err
	}

	https_nodePort := -1

	for _, port := range ingress.Spec.Ports {
		if port.Name == "https" {
			https_nodePort = int(port.NodePort)
		}
	}

	if https_nodePort == -1 {
		return -1, fmt.Errorf("Error the Service Has to be have a successfully NodePort running for the ingress Gateway, No matching nodeport found")
	}
	return https_nodePort, nil
}

// poll for the resource to ensure the lb resource is created as expected by IKS Lb Operator
func waitForSuccessIKSLbVIPCreation(kubeClient client.Client, name, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(time.Second*2, timeout, func() (bool, error) {
		resource := &loadbalancerv1alpha1.Loadbalancer{}
		log.Println("Started Polling the IKS LB Operator to validate the Status of cretaed External LB for DPAI Cluster ", name)
		err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: name, Namespace: namespace}, resource)
		if err != nil {
			return false, err
		}

		totalListeeners := len(resource.Status.Conditions.Listeners)
		activeVipAssociatedListeerns := 0

		for _, listener := range resource.Status.Conditions.Listeners {
			// the enhanced lb operator internally recomncile state for Highwire LB and ensure the status is updated in crd matching declared state
			if listener.VIPCreated && listener.VIPPoolLinked && listener.PoolCreated {
				activeVipAssociatedListeerns++
			}
		}

		if activeVipAssociatedListeerns != totalListeeners {
			log.Println("Some of the listener on the lb Are not healthy or the Requeired VIP Ip is not yet created retrying..,")
			return false, nil
		}
		return true, nil
	})
}

// verify correct ipv4 address is assigned to the VIP of Highwire
func isValidIPv4(ip string) (bool, error) {
	ipv4Octets := strings.Split(ip, ".")
	if len(ipv4Octets) != 4 {
		return false, fmt.Errorf("invalid IPv4 VIP format: expected 4 octets, got %d", len(ipv4Octets))
	}

	for _, ipSubmask := range ipv4Octets {
		octet, err := strconv.Atoi(ipSubmask)
		if err != nil {
			return false, fmt.Errorf("invalid IPv4 octet: %s is not a number", ipSubmask)
		}
		if octet < 0 || octet > 255 {
			return false, fmt.Errorf("invalid IPv4 octet: %d is out of range [0-255]", octet)
		}
	}
	return true, nil
}

// poll for the resource to ensure the lb resource is created as expected by IKS Lb Operator verifying the F5 VIP is proper for IPV4 IPAM CIDR
func waitForSuccessIKSLbVIPIPAMAssign(kubeClient client.Client, name, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(time.Second*2, timeout, func() (bool, error) {
		lb := &loadbalancerv1alpha1.Loadbalancer{}
		log.Printf("Polling IKS LB Operator for Highwire lb resources %s", name)
		err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: name, Namespace: namespace}, lb)
		if err != nil {
			return false, err
		}

		if len(lb.Status.Vip) == 0 {
			return false, nil
		}

		return isValidIPv4(lb.Status.Vip)
	})
}

// -------------------- Volumes --------------------------- //
func CreateOpenEbsNamespace(ctx *deployment.TaskRunContext) (any, error) {

	log.Printf("Inside CreateSecretNamespace using CreateIKS Output: %v \n", ctx.GetTaskOutput("InstallIstioIngressGateway"))

	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("InstallIstioIngressGateway"))
	if err != nil {
		return nil, err
	}
	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	labels := map[string]string{
		"istio-injection": "disabled",
	}

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}
	namespace, err := ctx.K8sClient.CreateNamespace(deployment.OpenEbsNamespace, false, labels)
	if err != nil {
		return nil, err
	}
	log.Printf("Completed: Inside CreateOpenEbsNamespace \n")
	output.OpenEbsNamespace = namespace.Name
	return output, nil
}

func getOpenEbsHelmClient(ctx *deployment.TaskRunContext, namespace string) (helmclient.Client, error) {
	helmClient, err := helm.GetHelmClient(ctx.K8sClient.ClientConfig, namespace)
	if err != nil {
		return nil, err
	}

	chartRepo := repo.Entry{
		Name: "openebs",
		URL:  "https://openebs.github.io/openebs",
	}

	// Add a chart-repository to the client.
	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		log.Printf("Error Adding Chart Repo %+v. Error message: %+v", chartRepo, err)
		return nil, err
	}
	return helmClient, nil
}

func InstallOpenEbs(ctx *deployment.TaskRunContext) (any, error) {
	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("CreateOpenEbsNamespace"))
	if err != nil {
		return nil, err
	}
	log.Printf("InstallOpenEbs : Fetching docker secret")
	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	helmClient, err := getOpenEbsHelmClient(ctx, deployment.OpenEbsNamespace)
	if err != nil {
		return nil, err
	}

	helmValuesOptions := values.Options{
		Values: []string{
			"engines.replicated.mayastor.enabled=false",
			"engines.local.lvm.enabled=false",
			"engines.local.zfs.enabled=false",
		},
	}
	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName:   "openebs",
		ChartName:     "openebs/openebs",
		Namespace:     deployment.OpenEbsNamespace,
		UpgradeCRDs:   true,
		ValuesOptions: helmValuesOptions,
		Wait:          true,
		Atomic:        true,
		Timeout:       15 * time.Minute,
		// CreateNamespace: true,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	_, err = helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error while installing Istio. %+v", err)
		return nil, err
	}

	err = ctx.K8sClient.MakeDefaultStorageClass("openebs-hostpath")
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func CommitCreateWorkspace(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Started: Inside CommitCreateWorkspace\n")

	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("CreateNodeGroup"))
	if err != nil {
		return nil, fmt.Errorf("error parsing the task output. %+v", err)
	}

	_, err = ctx.SqlModel.UpdateWorkspaceDeploymentStatus(context.Background(), model.UpdateWorkspaceDeploymentStatusParams{
		DeploymentID:                pgtype.Text{String: ctx.DeploymentContext.ID, Valid: true},
		DeploymentStatusState:       pgtype.Text{String: pb.DpaiDeploymentState_DPAI_SUCCESS.String(), Valid: true},
		DeploymentStatusDisplayName: pgtype.Text{String: "Success", Valid: true},
		DeploymentStatusMessage:     pgtype.Text{String: "Successfully provisioned the DPAI Workspace", Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("error updating the workspace deployment status. %+v", err)
	}
	log.Printf("Completed: Inside CommitCreateWorkspace\n")

	return output, nil
}

func getHwTLSConfig(ctx *deployment.TaskRunContext) *config.TlsConfig {
	return &ctx.DeploymentContext.Context.Conf.Tls
}

/* ------------------------------ IKS LB Operator resource creation   ---------------------------------- */
func CreateLoadBalancerCrd(ctx *deployment.TaskRunContext) (any, error) {
	output, err := parseTaskOutput[TaskOutput](ctx.GetTaskOutput("InstallIstioIngressGateway"))
	if err != nil {
		return nil, err
	}
	log.Println("The output passed to the create Lb Operator CRD is ", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		log.Fatalf("Error getting Iks k8s clientSet, while installing Istio Gateways")
		return nil, err
	}

	log.Printf("Generating the lb crd for IKS Lb Operator")

	ingressGatewayHttpsNodePort, err := getIngressNodePort(ctx)

	if err != nil {
		log.Fatal("Error could find the deployed Ingress Gateway NodePort")
		return nil, err
	}

	log.Println("Ingress Gateway Nodeport processed ", ingressGatewayHttpsNodePort)

	var K8sAzClient k8s.K8sAzClient = k8s.K8sAzClient{}

	azK8sClient, err := K8sAzClient.ConfigureK8sClient()
	if err != nil {
		log.Printf("Error getting the AZ clientset for Az cluster %+v", err)
		return nil, err
	}

	client, err := generateIksLBCRDClient(azK8sClient)
	hwTlsSSLProfileConfig := getHwTLSConfig(ctx)

	log.Println("Fetched the HW ssl profile config from dpai config for hw ssl profile iD", hwTlsSSLProfileConfig)

	// waiiting on LB operator to finalize the required schema for the CRD and resource for deployment
	lbOperator := &loadbalancerv1alpha1.Loadbalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", networking.DPAI_HIGHWIRE_LB_NAME, string(output.IksClusterUUID)),
			Namespace: output.CloudAccountID, // use this for all ingress processing
		},
		Spec: loadbalancerv1alpha1.LoadbalancerSpec{
			Listeners: []loadbalancerv1alpha1.LoadbalancerListener{
				{
					Owner: "vedang.parasnis@intel.com",
					VIP: loadbalancerv1alpha1.VServer{
						Port:       networking.DPAI_HIGHWIRE_POOL_PORT,
						IPProtocol: "tcp",
						IPType:     "public",
						Persist:    "i_client_ip_5min", // need this to be customized
						SSLConfig: &loadbalancerv1alpha1.SSL{
							Profile: loadbalancerv1alpha1.Profile{
								Id: hwTlsSSLProfileConfig.HwSSLProfileID,
							},
						},
					},
					Pool: loadbalancerv1alpha1.VPool{
						Port:              ingressGatewayHttpsNodePort,
						LoadBalancingMode: LEAST_CONNECTIONS_MEMBER_MODE,
						MinActiveMembers:  int(output.NodeCount),
						Monitor:           "tcp",
						Members:           []loadbalancerv1alpha1.VMember{},
						InstanceSelectors: map[string]string{
							deployment.IngressIKSNodeLabelKey: output.NodeGroupID,
						},
					},
				},
			},
			Security: loadbalancerv1alpha1.LoadbalancerSecurity{
				Sourceips: []string{
					"any", // IDC FW API for IDC AZ Zone FW config, all IDC FW API are in AZ regtion and not regional services like DPAI
				},
			},
			Labels: map[string]string{
				"dpai": "dpai-istio-ingress-gateway",
			},
		},
	}

	log.Println("IKS Operator CRD Manifest Gen", lbOperator)

	if err := client.Create(context.Background(), lbOperator); err != nil {
		log.Println("The Load Balancer Resource Already Exists")
	} else {
		log.Println("The Load Balancer Resource Successfully Created")
	}

	log.Println("Waiting for the Lb Operator to flush log output during reconcilation")
	time.Sleep(time.Second * 5)

	if err := waitForSuccessIKSLbVIPCreation(client, lbOperator.Name, output.CloudAccountID, 80*time.Minute); err != nil {
		log.Println("Error after timeout the required Loadbalancer resources is not created and firewall rules added")
		log.Printf("error in VIP creation %+v ", err)
		return nil, err
	}

	if err := waitForSuccessIKSLbVIPIPAMAssign(client, lbOperator.Name, output.CloudAccountID, 80*time.Minute); err != nil {
		log.Println("The Required VIP was created but there is some issue regarding IPAM Ipv4 address assign over the VIP")
		log.Printf("error in VIP IPAM Assign %+v ", err)
		return nil, err
	}

	/// once created fetch the VIP LB IP and associated values for DNS Zone creation usacase
	// there is always going to 1 LB CRD irrespective of the number of DPAI services because of 1 Ingress Nodegroup handling all ingress traffic
	// Each service just consumes this gateway via label selector and kubernetes Gateway controller handler concurrent route and gateeway discovery for all the DPAI services adhering to Gateway API standards of K8s SIG networking
	nn := types.NamespacedName{
		Namespace: output.CloudAccountID,
		Name:      lbOperator.Name,
	}

	dpaiIngressLoadBalancer := &loadbalancerv1alpha1.Loadbalancer{}

	if err := client.Get(context.Background(), nn, dpaiIngressLoadBalancer); err != nil {
		log.Println("Error fetching the created loadbalancer with Name", nn.Name)
		return nil, err
	}

	lbHwDnsARecord := dns.GenerateFQDNfromIP(dpaiIngressLoadBalancer.Status.Vip, "idcstage")

	_, err = ctx.SqlModel.InsertGatewayForWorkspace(context.Background(), model.InsertGatewayForWorkspaceParams{
		CloudAccountID:  output.CloudAccountID,
		LbFqdn:          lbHwDnsARecord,
		LbCreated:       true,
		FwCreated:       false,
		Gatewaynodeport: int32(ingressGatewayHttpsNodePort),
		WorkspaceID:     output.WorkspaceID,
	})

	if err != nil {
		log.Printf("Error in inserting record in the DB for workspace %+v", err)
		return nil, err
	}

	return nil, nil
}

/* ------------------------ Workspace Delete tasks ------------------------ */
/* ------------------------ Delete the Entire workspace with all the services added inside the workspace ------------------------- */
// delte for the loadbalance resource for lb resource to first start cascadingly removing all cname in menmice which are pointing the A record created by Highwire
// Deletes External IKS LB, DNS records and lb recrods in the database

func DeleteAllServices(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Marking all the services in Workspace as Deleted\n")

	deploymentInput, _ := GetDeploymentDeleteInput(ctx.DeploymentContext)
	// id := deploymentInput.GetId()
	id := ctx.DeploymentContext.Context.WorkspaceId

	tctx := context.Background()
	tx, err := ctx.SqlConn.Begin(tctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(tctx)
	qtx := ctx.SqlModel.WithTx(tx)

	// Delete Postgres DBs
	postgres, err := qtx.ListPostgres(tctx, model.ListPostgresParams{
		WorkspaceID:    pgtype.Text{String: id, Valid: true},
		CloudAccountID: deploymentInput.GetCloudAccountId(),
	})
	if err != nil {
		tx.Rollback(tctx)
		return nil, fmt.Errorf("error: Failed to the list the postgres DBs in workspace %s. Error message: %+v", id, err)
	}
	log.Printf("DeleteAllServices: Postgres %+v \n ", postgres)

	for _, pg := range postgres {
		log.Printf("DeleteAllServices: Inside Delete Postgres Loop %s \n", pg.ID)
		err = qtx.DeletePostgres(tctx, pg.ID)
		if err != nil {
			tx.Rollback(tctx)
			return nil, fmt.Errorf("error: Failed to the mark the postgres DBs in workspace %s as deleted. Error message: %+v", id, err)
		}
	}

	// Delete Airflow Services
	airflow, err := qtx.ListAirflow(tctx, model.ListAirflowParams{
		WorkspaceID:    pgtype.Text{String: id, Valid: true},
		CloudAccountID: deploymentInput.GetCloudAccountId(),
	})
	if err != nil {
		tx.Rollback(tctx)
		return nil, fmt.Errorf("error: Failed to the list the Airflow in workspace %s. Error message: %+v", id, err)
	}

	var IksClusterUUID string
	for _, af := range airflow {
		err = qtx.DeleteAirflow(tctx, af.ID)
		IksClusterUUID = af.IksClusterID
		if err != nil {
			tx.Rollback(tctx)
			return nil, fmt.Errorf("error: Failed to the mark the Airflow (id: %s) in workspace %s as deleted. Error message: %+v", af.ID, id, err)
		}
	}

	// Delete all external networking resources consumed by the workspace
	var mesh networking.NetworkResourceDelete = &networking.Network{
		CloudAccountID: deploymentInput.CloudAccountId,
		WorkspaceId:    id,
		IksClusterUUID: IksClusterUUID,
		SqlModel:       qtx,
	}

	networkRecordsOutput, err := mesh.DeleteIKSLoadbalancerResource(tctx)
	if isNetworkErr := networking.IsNetworkError(err); isNetworkErr {
		log.Printf("Error deleting IKS Lb resources. for airflow %+v", airflow)
		return nil, err
	}
	networkRecordsOutput, err = mesh.DeleteNetworkingResourcesInDB(tctx, networkRecordsOutput)
	if isNetworkErr := networking.IsNetworkError(err); isNetworkErr {
		log.Printf("Error deleting Networking resource from DPAI Database. airflow %+v", airflow)
		return nil, err
	}

	tx.Commit(tctx)

	return nil, nil
}

func DeleteIKSCluster(ctx *deployment.TaskRunContext) (any, error) {
	err := ctx.K8sClient.DeleteIKSCluster()
	return nil, err
}

func CommitDeleteWorkspace(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting CommitDeleteWorkspace\n")
	deploymentInput, _ := GetDeploymentDeleteInput(ctx.DeploymentContext)

	ctx.SqlModel.DeleteWorkspace(context.Background(), deploymentInput.GetId())
	return nil, nil
}

// delete SSH Key
func DeleteSshKey(ctx *deployment.TaskRunContext) (any, error) {

	err := ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		log.Printf("Error creating the IKS client. %+v", err)
		return nil, err
	}

	workspace, err := ctx.SqlModel.GetWorkspace(context.Background(), model.GetWorkspaceParams{
		WorkspaceID:    ctx.DeploymentContext.Context.WorkspaceId,
		CloudAccountID: pgtype.Text{String: ctx.K8sClient.ClusterID.CloudAccountId, Valid: true},
	})

	_, err = ctx.K8sClient.SshClient.Delete(context.Background(), &pb.SshPublicKeyDeleteRequest{
		Metadata: &pb.ResourceMetadataReference{
			CloudAccountId: ctx.K8sClient.ClusterID.CloudAccountId,
			NameOrId: &pb.ResourceMetadataReference_Name{
				Name: workspace.SshKeyName.String,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error deleting the sshKey %s. %+v", workspace.SshKeyName.String, err)
	}
	return nil, nil
}
