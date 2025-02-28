// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kfaas/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Status struct {
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Status    string `json:"status"`
}

var fieldOpts []protodb.FieldOptions = []protodb.FieldOptions{}

func NewKFService(session *sql.DB, iksClient pb.IksClient, cfg config.Config) (*Server, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}
	return &Server{
		session:   session,
		iksClient: iksClient,
		cfg:       cfg,
	}, nil
}

var (
	port = flag.Int("port", 50051, "gRPC server port")
)

type Server struct {
	pb.UnimplementedKFServiceServer
	session   *sql.DB
	iksClient pb.IksClient
	cfg       config.Config
}

func (c *Server) CreateKubeFlowDeployment(ctx context.Context, obj *pb.CreateKubeFlowDeploymentRequest) (*pb.CreateKubeFlowDeploymentResponse, error) {

	fmt.Println("Creating KF deployment")

	id, err := NewId() // cloud account id.
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("KFService.KubeFlowDeployment").WithValues(logkeys.KubeflowDeploymentId, id).Start()
	defer span.End()
	if err != nil {
		return &pb.CreateKubeFlowDeploymentResponse{}, fmt.Errorf("unable to create KF deployment id: %w", err)
	}
	fmt.Println("Inserting KF deployment record")
	//params := protodb.NewProtoToSql(obj, fieldOpts...)
	// vals := params.GetValues()
	// vals = append(vals, id)

	// query := fmt.Sprintf("INSERT INTO kfaas_deployments (%v, deployment_id) VALUES(%v, $%v)",
	// 	params.GetNamesString(), params.GetParamsString(), len(vals))
	// // fmt.Println(query)
	tx, err := c.session.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "error initiating db transaction")
	}

	query := `insert into kfaas_deployments (deployment_id, deployment_name, kf_version, k8s_cluster_i_d, k8s_cluster_name, storage_class_name, cloud_account_id, status, created_date) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, query, id, obj.DeploymentName, obj.KfVersion, obj.K8SClusterID, obj.K8SClusterName, obj.StorageClassName, obj.CloudAccountId, obj.Status, time.Now()); err != nil {
		logger.Error(err, "error inserting kf into db", logkeys.Query, query)
		return &pb.CreateKubeFlowDeploymentResponse{}, fmt.Errorf("unable to create KF deployment")
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error committing db transaction")
		return &pb.CreateKubeFlowDeploymentResponse{}, fmt.Errorf("unable to create KF deployment")

	}
	fmt.Println("Inserted KF deployment record")
	status, err := UpdateStatus(id, "Initiated", obj.CloudAccountId, c, ctx)
	if err != nil {
		logger.Error(err, "error updating kf deployment status db transaction")
		return &pb.CreateKubeFlowDeploymentResponse{}, fmt.Errorf("unable to create KF deployment")
	}
	fmt.Println(status)
	stringKubeconfig, err := c.iksClient.GetKubeConfig(ctx, &pb.ClusterID{

		Clusteruuid:    obj.K8SClusterID,
		CloudAccountId: obj.CloudAccountId,
	})

	if err != nil {
		return &pb.CreateKubeFlowDeploymentResponse{}, fmt.Errorf("unable to retrieve kubeconfig of cluster")
	}
	// fmt.Println("kubeconfig " + stringKubeconfig)
	namepaceAndAccount, err := NamespaceAndAccount(stringKubeconfig.Kubeconfig)
	if err != nil {
		return &pb.CreateKubeFlowDeploymentResponse{}, fmt.Errorf("unable to create namespace or service account")
	}
	fmt.Println(namepaceAndAccount)

	job, err := TriggerJob(stringKubeconfig.Kubeconfig, "/app/kf-install.sh", "install", c.cfg.InstallImage)

	if err != nil {
		return &pb.CreateKubeFlowDeploymentResponse{}, fmt.Errorf("unable to trigger deployment")
	}
	fmt.Println("job " + job)
	return &pb.CreateKubeFlowDeploymentResponse{DeploymentID: id}, nil

	// authToken, err := agent.CreateArgoSession()
	// if err != nil {

	// 	return &pb.CreateKubeFlowDeploymentResponse{}, fmt.Errorf("no agent connection found")
	// }
	// ca, token, server, err := agent.GetDetailsFromKubeconfig(obj.KfVersion)
	// if err != nil {
	// 	return &pb.CreateKubeFlowDeploymentResponse{}, fmt.Errorf("unable to connect to cluster")
	// }
	// registerClusterStatus, err := agent.RegisterCluster(obj.DeploymentName+"-"+obj.CloudAccountId, token, ca, server, authToken)
	// if err != nil {
	// 	return &pb.CreateKubeFlowDeploymentResponse{}, fmt.Errorf("unable to connect to cluster for deployment")
	// }
	// fmt.Println(registerClusterStatus)

	// createApplicationStatus, err := agent.CreateApplication(obj.DeploymentName+"-"+obj.CloudAccountId, server, authToken)
	// fmt.Println(createApplicationStatus)

	// if err != nil {
	// 	fmt.Println(err)
	// 	fmt.Println(status)
	// 	//logger.Error(err, "error starting db transaction")
	// 	return nil, err
	// }
	// syncApplication, err := agent.SyncApplication(obj.DeploymentName+"-"+obj.CloudAccountId, authToken)
	// if err != nil {
	// 	fmt.Println(err)

	// 	//logger.Error(err, "error starting db transaction")
	// 	return nil, err
	// }
	// fmt.Println(syncApplication)

}

func (c *Server) GetUserCredentials(ctx context.Context, obj *pb.GetUserCredentialsRequest) (*pb.GetUserCredentialsResponse, error) {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("KFService.GetUserCredentials").WithValues(logkeys.KubeflowDeploymentId, obj.DeploymentID).Start()
	defer span.End()
	logger.Info("Get KF User credentials invoked")
	getKFrecord, err := GetKFRecord(obj.DeploymentID, obj.CloudAccountId, c, ctx)
	if err != nil {
		return &pb.GetUserCredentialsResponse{}, fmt.Errorf("unable to fetch record")
	}
	stringKubeconfig, err := c.iksClient.GetKubeConfig(ctx, &pb.ClusterID{

		Clusteruuid:    getKFrecord.K8SClusterID,
		CloudAccountId: obj.CloudAccountId,
	})

	if err != nil {
		return &pb.GetUserCredentialsResponse{}, fmt.Errorf("unable to retrieve kubeconfig of cluster")
	}

	password, err := GetUserPassword(stringKubeconfig.Kubeconfig)

	if err != nil {
		return &pb.GetUserCredentialsResponse{}, fmt.Errorf("unable to retrieve password")
	}
	returnResponse := &pb.GetUserCredentialsResponse{
		Username: "user01",
		Password: password,
	}
	return returnResponse, nil
}
func (c *Server) GetExternalIP(ctx context.Context, obj *pb.GetExternalIPRequest) (*pb.GetExternalIPResponse, error) {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("KFService.GetExternalIP").WithValues(logkeys.KubeflowDeploymentId, obj.DeploymentID).Start()
	defer span.End()
	logger.Info("GetExternalIP invoked")
	getKFrecord, err := GetKFRecord(obj.DeploymentID, obj.CloudAccountId, c, ctx)
	if err != nil {
		return &pb.GetExternalIPResponse{}, fmt.Errorf("unable to fetch record")
	}
	stringKubeconfig, err := c.iksClient.GetKubeConfig(ctx, &pb.ClusterID{

		Clusteruuid:    getKFrecord.K8SClusterID,
		CloudAccountId: obj.CloudAccountId,
	})

	if err != nil {
		return &pb.GetExternalIPResponse{}, fmt.Errorf("unable to retrieve kubeconfig of cluster")
	}

	ip, err := GetIngress(stringKubeconfig.Kubeconfig)
	if err != nil {
		return &pb.GetExternalIPResponse{}, fmt.Errorf("unable to retrieve IP")
	}
	returnResponse := &pb.GetExternalIPResponse{
		Ip: ip,
	}
	return returnResponse, nil
}

func TriggerJob(kubeconfigStr string, script string, name string, image string) (string, error) {
	// Define your hardcoded kubeconfig as a string
	// fmt.Println(kubeconfigStr)
	// Write the hardcoded kubeconfig to a temporary file

	tmpfile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		// panic(err.Error())
		// fmt.Println("Error in creating temp")
		// fmt.Println(err)
		return "", err
	}
	defer tmpfile.Close()

	if _, err := tmpfile.Write([]byte(kubeconfigStr)); err != nil {
		// fmt.Println("Error in writing config")
		// fmt.Println(err)
		return "", err
	}

	// Use the temporary kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", tmpfile.Name())
	if err != nil {
		fmt.Println("Error in building config")
		fmt.Println(err)
		return "", err
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error in creating config")
		fmt.Println(err)
		return "", fmt.Errorf("unable to install Kubeflow")
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-kfaas",
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: "kfaas-cluster-admin",
					Containers: []corev1.Container{
						{
							Name:            name + "-kfaas",
							ImagePullPolicy: corev1.PullAlways,
							Image:           image,
							Command: []string{
								"/bin/sh",
								script,
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: int32Ptr(5),
		},
	}

	// Create the Job in the cluster
	createdJob, err := clientset.BatchV1().Jobs("kfaas").Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("unable to install Kubeflow")
	}

	fmt.Printf("Job %s created successfully\n", createdJob.Name)
	return createdJob.Name, nil

	// Now you can use the clientset to interact with the Kubernetes cluster
	// secret, err := clientset.CoreV1().Secrets("auth").Get(context.TODO(), "auth-secret", metav1.GetOptions{})
	// if err != nil {
	// 	// fmt.Println("Error: in fecthing password")
	// 	// fmt.Println(err)
	// 	return "", err
	// }

	// value, exists := secret.Data["password"]
	// fmt.Println("Password:" + string(value))
	// if !exists {
	// 	fmt.Printf("Key %s not found in the secret\n", "password")
	// 	return "", fmt.Errorf("key not ready for password")
	// }

	// return string(value), nil

}
func int32Ptr(i int32) *int32 {
	return &i
}
func NamespaceAndAccount(kubeconfigStr string) (string, error) {
	// Define your hardcoded kubeconfig as a string
	// fmt.Println(kubeconfigStr)
	// Write the hardcoded kubeconfig to a temporary file
	tmpfile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		// panic(err.Error())
		// fmt.Println("Error in creating temp")
		// fmt.Println(err)
		return "", err
	}
	defer tmpfile.Close()

	if _, err := tmpfile.Write([]byte(kubeconfigStr)); err != nil {
		// fmt.Println("Error in writing config")
		// fmt.Println(err)
		return "", err
	}

	// Use the temporary kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", tmpfile.Name())
	if err != nil {
		fmt.Println("Error in building config")
		fmt.Println(err)
		return "", err
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		// fmt.Println("Error in creating config")
		// fmt.Println(err)
		return "", err
	}
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kfaas",
		},
	}

	createdNamespace, err := clientset.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	// Now you can use the clientset to interact with the Kubernetes cluster

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kfaas-cluster-admin",
			Namespace: "kfaas", // Change the namespace as needed
		},
	}

	createdServiceAccount, err := clientset.CoreV1().ServiceAccounts("kfaas").Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "installclusterrolebinding",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "kfaas-cluster-admin",
				Namespace: "kfaas", // Replace with the actual namespace of the ServiceAccount
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "cluster-admin", // Replace with the actual ClusterRole name
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	createdClusterRoleBinding, err := clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("ClusterRoleBinding %s created\n", createdClusterRoleBinding.Name)
	return "Created namespace: " + createdNamespace.Name + " Created sa: " + createdServiceAccount.Name, nil

}
func GetUserPassword(kubeconfigStr string) (string, error) {
	// Define your hardcoded kubeconfig as a string
	// fmt.Println(kubeconfigStr)
	// Write the hardcoded kubeconfig to a temporary file
	tmpfile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		// panic(err.Error())
		// fmt.Println("Error in creating temp")
		// fmt.Println(err)
		return "", err
	}
	defer tmpfile.Close()

	if _, err := tmpfile.Write([]byte(kubeconfigStr)); err != nil {
		// fmt.Println("Error in writing config")
		// fmt.Println(err)
		return "", err
	}

	// Use the temporary kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", tmpfile.Name())
	if err != nil {
		fmt.Println("Error in building config")
		fmt.Println(err)
		return "", err
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		// fmt.Println("Error in creating config")
		// fmt.Println(err)
		return "", err
	}

	// Now you can use the clientset to interact with the Kubernetes cluster
	secret, err := clientset.CoreV1().Secrets("auth").Get(context.TODO(), "auth-secret", metav1.GetOptions{})
	if err != nil {
		// fmt.Println("Error: in fecthing password")
		// fmt.Println(err)
		return "", err
	}

	value, exists := secret.Data["password"]
	fmt.Println("Password:" + string(value))
	if !exists {
		fmt.Printf("Key %s not found in the secret\n", "password")
		return "", fmt.Errorf("key not ready for password")
	}

	return string(value), nil

}
func GetPodsStatus(kubeconfigStr string, namespace string) ([]Resource, error) {
	tmpfile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		// panic(err.Error())
		return nil, err
	}
	defer tmpfile.Close()

	if _, err := tmpfile.Write([]byte(kubeconfigStr)); err != nil {
		return nil, err
	}

	// Use the temporary kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", tmpfile.Name())
	if err != nil {
		return nil, err
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting pods in namespace %s: %v\n", "kfaas", err)
		return nil, err
	}

	var podData []Resource
	for _, pod := range pods.Items {
		fmt.Printf("Name: %s, Status: %s\n", pod.Name, pod.Status.Phase)
		podData = append(podData, Resource{
			Name:      pod.Name,
			Status:    string(pod.Status.Phase),
			Namespace: pod.Namespace,
		})
	}
	return podData, nil

}
func GetOverallStatus(kubeconfigStr string, configmap string) (string, error) {
	// Define your hardcoded kubeconfig as a string
	// fmt.Println(kubeconfigStr)
	// Write the hardcoded kubeconfig to a temporary file
	tmpfile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		// panic(err.Error())
		// fmt.Println("Error in creating temp")
		// fmt.Println(err)
		return "", err
	}
	defer tmpfile.Close()

	if _, err := tmpfile.Write([]byte(kubeconfigStr)); err != nil {
		// fmt.Println("Error in writing config")
		// fmt.Println(err)
		return "", err
	}

	// Use the temporary kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", tmpfile.Name())
	if err != nil {
		fmt.Println("Error in building config")
		fmt.Println(err)
		return "", err
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		// fmt.Println("Error in creating config")
		// fmt.Println(err)
		return "", err
	}

	// Now you can use the clientset to interact with the Kubernetes cluster
	configMap, err := clientset.CoreV1().ConfigMaps("kfaas").Get(context.TODO(), configmap, metav1.GetOptions{})
	if err != nil {
		// fmt.Println("Error: in fecthing password")
		// fmt.Println(err)
		return "", err
	}

	value, exists := configMap.Data["status"]
	fmt.Println("Password:" + string(value))
	if !exists {
		fmt.Printf("Key %s not found in the secret\n", "password")
		return "", fmt.Errorf("key not ready in config map")
	}

	return string(value), nil

}

func GetIngress(kubeconfigStr string) (string, error) {
	// Define your hardcoded kubeconfig as a string
	// fmt.Println(kubeconfigStr)
	// Write the hardcoded kubeconfig to a temporary file
	tmpfile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		// panic(err.Error())
		return "", err
	}
	defer tmpfile.Close()

	if _, err := tmpfile.Write([]byte(kubeconfigStr)); err != nil {
		return "", err
	}

	// Use the temporary kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", tmpfile.Name())
	if err != nil {
		return "", err
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}
	service, err := clientset.CoreV1().Services("istio-system").Get(context.TODO(), "istio-ingressgateway", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if service.Status.LoadBalancer.Ingress[0].IP != "" {
		value := service.Status.LoadBalancer.Ingress[0].IP
		return string(value), nil
	} else {
		return "", fmt.Errorf("IP not found for service")
	}

	// if !exists {
	// 	fmt.Printf("Key %s not found in the service\n", "ingress")
	// 	return "", fmt.Errorf("IP not found for service")
	// }

}

func GetKubernetesClient(kubeconfig []byte) (*kubernetes.Clientset, error) {
	rc, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func (c *Server) DeleteKubeFlowDeployment(ctx context.Context, input *pb.DeleteKubeFlowDeploymentRequest) (*pb.DeleteKubeFlowDeploymentResponse, error) {

	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("KFService.DeleteKubeFlowDeployment").WithValues(logkeys.KubeflowDeploymentId, input.DeploymentID).Start()
	defer span.End()
	logger.Info("DeleteKubeFlowDeployment invoked")
	getKFrecord, err := GetKFRecord(input.DeploymentID, input.CloudAccountId, c, ctx)
	if err != nil {
		return &pb.DeleteKubeFlowDeploymentResponse{}, fmt.Errorf("unable to fetch record")
	}
	stringKubeconfig, err := c.iksClient.GetKubeConfig(ctx, &pb.ClusterID{

		Clusteruuid:    getKFrecord.K8SClusterID,
		CloudAccountId: input.CloudAccountId,
	})

	if err != nil {
		return &pb.DeleteKubeFlowDeploymentResponse{}, fmt.Errorf("unable to retrieve kubeconfig of cluster")
	}

	job, err := TriggerJob(stringKubeconfig.Kubeconfig, "/app/kf-delete.sh", "delete", c.cfg.InstallImage)

	if err != nil {
		return &pb.DeleteKubeFlowDeploymentResponse{}, fmt.Errorf("unable to trigger deployment")
	}
	fmt.Println("job " + job)
	status, err := UpdateStatus(input.DeploymentID, "Delete Initiated", input.CloudAccountId, c, ctx)
	if err != nil {
		return &pb.DeleteKubeFlowDeploymentResponse{}, fmt.Errorf("unable to update record")
	}
	returnResponse := &pb.DeleteKubeFlowDeploymentResponse{
		Status: status,
	}

	return returnResponse, nil

}
func (c *Server) GetKFStatus(ctx context.Context, input *pb.GetKFStatusRequest) (*pb.GetKFStatusResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("KFService.GetKFStatus").WithValues(logkeys.KubeflowDeploymentId, input.DeploymentID).Start()
	defer span.End()
	logger.Info("GetKFStatus invoked")
	var componentsStatus []Resource
	returnValue := &pb.GetKFStatusResponse{
		Components: []*pb.KFStatus{},
	}
	getKFrecord, err := GetKFRecord(input.DeploymentID, input.CloudAccountId, c, ctx)
	if err != nil {
		return &pb.GetKFStatusResponse{}, fmt.Errorf("unable to fetch record")
	}
	stringKubeconfig, err := c.iksClient.GetKubeConfig(ctx, &pb.ClusterID{

		Clusteruuid:    getKFrecord.K8SClusterID,
		CloudAccountId: input.CloudAccountId,
	})

	if err != nil {
		return &pb.GetKFStatusResponse{}, fmt.Errorf("unable to retrieve kubeconfig of cluster")
	}
	// if input.Namespace == "install" {
	// 	componentsStatus, err = GetOverallStatus(stringKubeconfig, "kf-install-status")

	// } else if input.Namespace == "delete" {
	// 	componentsStatus, err = GetOverallStatus(stringKubeconfig, "kf-delete-status")
	// } else {
	componentsStatus, err = GetPodsStatus(stringKubeconfig.Kubeconfig, input.Namespace)
	//}

	if err != nil {
		return &pb.GetKFStatusResponse{}, fmt.Errorf("unable to get component status")
	}
	// returnValue.Components = componentsStatus
	for _, sourceObj := range componentsStatus {
		targetObj := &pb.KFStatus{}

		targetObj.Pod = sourceObj.Name
		targetObj.Status = sourceObj.Status
		targetObj.Namespace = "kfaas"
		returnValue.Components = append(returnValue.Components, targetObj)
	}

	return returnValue, nil
}

// 	authToken, err := agent.CreateArgoSession()
// 	if err != nil {
// 		return &pb.GetKFStatusResponse{}, fmt.Errorf("unable to connect to cluster")
// 	}
// 	componentsStatus, err := agent.GetComponentsStatus(getKFrecord.DeploymentName+"-"+getKFrecord.CloudAccountId, authToken)
// 	if err != nil {
// 		return &pb.GetKFStatusResponse{}, fmt.Errorf("unable to get component status")
// 	}
// 	// Loop through the source array and assign each struct to the target struct
// 	for _, sourceObj := range componentsStatus {
// 		targetObj := &pb.KFStatus{}

// 		targetObj.Name = sourceObj.Name
// 		targetObj.Status = sourceObj.Status
// 		// Assign other fields as needed

// 		// Append the target struct to the target array
// 		returnValue.Components = append(returnValue.Components, targetObj)
// 	}

// 	return returnValue, nil
// }

func (c *Server) ListKubeFlowDeployment(ctx context.Context, input *pb.ListKubeFlowDeploymentRequest) (*pb.ListKubeFlowDeploymentResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("KFService.ListKubeFlowDeployment").WithValues(logkeys.CloudAccountId, input.CloudAccountId).Start()
	defer span.End()
	logger.Info("List KF deployments")
	//returnError := &pb.ListKubeFlowDeploymentResponse{}
	returnValue := &pb.ListKubeFlowDeploymentResponse{
		Response: []*pb.KubeFlowDeployment{},
	}
	query := `
	SELECT deployment_id, deployment_name, kf_version, k8s_cluster_i_d, k8s_cluster_name, storage_class_name, status, created_date, cloud_account_id from kfaas_deployments 
	where status!='Deleted' and cloud_account_id = $1
	`
	args := []any{input.CloudAccountId}
	//readParams := protodb.NewSqlToProto(&obj, fieldOpts...)

	rows, err := c.session.QueryContext(ctx, query, args...)
	if err != nil {
		return &pb.ListKubeFlowDeploymentResponse{}, fmt.Errorf("unable to retrieve records")
	}

	for rows.Next() {
		obj := pb.KubeFlowDeployment{}
		err = rows.Scan(&obj.DeploymentID, &obj.DeploymentName, &obj.KfVersion,
			&obj.K8SClusterID, &obj.K8SClusterName, &obj.StorageClassName, &obj.Status, &obj.CreatedDate, &obj.CloudAccountId)

		if err != nil {
			//logger.Error(err, "error starting db transaction")
			return &pb.ListKubeFlowDeploymentResponse{}, fmt.Errorf("unable to retrieve records")
		}

		returnValue.Response = append(returnValue.Response, &obj)

	}

	return returnValue, nil

}

func GetKFRecord(id string, cloud_account_id string, c *Server, ctx context.Context) (*pb.KubeFlowDeployment, error) {

	fmt.Println("List KF deployment for a deploymentID")
	//returnError := &pb.ListKubeFlowDeploymentResponse{}
	obj := &pb.KubeFlowDeployment{}
	query := `
		select deployment_id, deployment_name, kf_version, k8s_cluster_i_d, k8s_cluster_name, storage_class_name, status, created_date, cloud_account_id 
		from kfaas_deployments 
		where  deployment_id = $1
		  and  cloud_account_id = $2
	`
	args := []any{id, cloud_account_id}
	//readParams := protodb.NewSqlToProto(&obj, fieldOpts...)

	// fmt.Println(query)

	rows, err := c.session.QueryContext(ctx, query, args...)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for rows.Next() {

		err = rows.Scan(&obj.DeploymentID, &obj.DeploymentName, &obj.KfVersion,
			&obj.K8SClusterID, &obj.K8SClusterName, &obj.StorageClassName, &obj.Status, &obj.CreatedDate, &obj.CloudAccountId)

		if err != nil {
			//logger.Error(err, "error starting db transaction")
			return nil, err
		}

	}

	return obj, nil

}

func UpdateStatus(id string, status string, cloud_account_id string, c *Server, ctx context.Context) (bool, error) {
	fmt.Println("Update KF deployment")

	query := `UPDATE kfaas_deployments set status= $1 where deployment_id= $2 and cloud_account_id= $3`
	// fmt.Println(query)
	tx, err := c.session.BeginTx(ctx, nil)
	if err != nil {
		//logger.Error(err, "error starting db transaction")
		return false, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, query, status, id, cloud_account_id); err != nil {
		//logger.Error(err, "error updating kfaas_deployments into db", "query", query)
		return false, err
	}

	if err := tx.Commit(); err != nil {
		//logger.Error(err, "error committing db transaction")
		return false, err
	}

	return true, nil
}

// non-const for testing purposes
var maxId int64 = 1_000_000_000_000

func NewId() (string, error) {
	intId, err := rand.Int(rand.Reader, big.NewInt(maxId))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%012d", intId), nil
}

func (c *Server) GetJobStatus(ctx context.Context, req *pb.GetJobStatusRequest) (*pb.GetJobStatusResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("KFService.GetJobStatus").WithValues(logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("Get KF pre-check invoked")
	getKFrecord, err := GetKFRecord(req.DeploymentID, req.CloudAccountId, c, ctx)
	if err != nil {
		return &pb.GetJobStatusResponse{}, fmt.Errorf("unable to fetch record")
	}
	stringKubeconfig, err := c.iksClient.GetKubeConfig(ctx, &pb.ClusterID{

		Clusteruuid:    getKFrecord.K8SClusterID,
		CloudAccountId: req.CloudAccountId,
	})
	if err != nil {
		return &pb.GetJobStatusResponse{}, fmt.Errorf("unable to retrieve kubeconfig of cluster")
	}
	jobStatusResult, err := GetJobStatus(stringKubeconfig.Kubeconfig, req.Job)
	if err != nil {
		return &pb.GetJobStatusResponse{}, fmt.Errorf("unable to retrieve Job status")
	}
	returnResponse := &pb.GetJobStatusResponse{
		Status: jobStatusResult,
	}
	return returnResponse, nil
}

func GetJobStatus(kubeconfigStr string, jobName string) (string, error) {
	// Define your hardcoded kubeconfig as a string
	// fmt.Println(kubeconfigStr)
	// Write the hardcoded kubeconfig to a temporary file
	tmpfile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		// panic(err.Error())
		// fmt.Println("Error in creating temp")
		// fmt.Println(err)
		return "", err
	}
	defer tmpfile.Close()

	if _, err := tmpfile.Write([]byte(kubeconfigStr)); err != nil {
		// fmt.Println("Error in writing config")
		// fmt.Println(err)
		return "", err
	}

	// Use the temporary kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", tmpfile.Name())
	if err != nil {
		fmt.Println("Error in building config")
		fmt.Println(err)
		return "", err
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		// fmt.Println("Error in creating config")
		// fmt.Println(err)
		return "", err
	}

	namespace := "kfaas"

	jobClient := clientset.BatchV1().Jobs(namespace)

	job, err := jobClient.Get(context.TODO(), jobName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("Error getting job: ", err)
		return "", err
	}

	// Check the status of the job
	if job.Status.Succeeded > 0 {
		return "Success", nil
	} else if job.Status.Failed > 0 {
		return "Failed", nil
	} else {
		return "Running", nil
	}

}

func (c *Server) ExecuteKFPreCheck(ctx context.Context, req *pb.ExecuteKFPreCheckRequest) (*pb.ExecuteKFPreCheckResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("KFService.ExecuteKFPreCheck").WithValues(logkeys.CloudAccountId, req.CloudAccountId).Start()
	defer span.End()

	logger.Info("Get KF pre-check invoked")

	stringKubeconfig, err := c.iksClient.GetKubeConfig(ctx, &pb.ClusterID{
		Clusteruuid:    req.K8SClusterID,
		CloudAccountId: req.CloudAccountId,
	})
	if err != nil {
		return &pb.ExecuteKFPreCheckResponse{}, fmt.Errorf("unable to retrieve kubeconfig of cluster")
	}

	// TODO: could be done with single block, find a way for error handling
	var statusResult bool
	if req.Check == "namespace" {
		statusResult, err = GetPreCheck(stringKubeconfig.Kubeconfig, req.Check)
		if err != nil {
			return &pb.ExecuteKFPreCheckResponse{}, fmt.Errorf("unable to retrieve namespace")
		}
	} else if req.Check == "kubectlVersion" {
		statusResult, err = GetPreCheck(stringKubeconfig.Kubeconfig, req.Check)
		if err != nil {
			return &pb.ExecuteKFPreCheckResponse{}, fmt.Errorf("unable to retrieve kubectl version")
		}
	} else if req.Check == "clusterRoleBinding" {
		statusResult, err = GetPreCheck(stringKubeconfig.Kubeconfig, req.Check)
		if err != nil {
			return &pb.ExecuteKFPreCheckResponse{}, fmt.Errorf("unable to retrieve clusterRoleBinding")
		}
	}

	returnResponse := &pb.ExecuteKFPreCheckResponse{
		Status: statusResult,
	}
	return returnResponse, nil
}

func GetPreCheck(kubeconfigStr string, checkStr string) (bool, error) {
	// Define your hardcoded kubeconfig as a string
	// fmt.Println(kubeconfigStr)
	// Write the hardcoded kubeconfig to a temporary file
	tmpfile, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		// panic(err.Error())
		// fmt.Println("Error in creating temp")
		// fmt.Println(err)
		return false, err
	}
	defer tmpfile.Close()

	if _, err := tmpfile.Write([]byte(kubeconfigStr)); err != nil {
		// fmt.Println("Error in writing config")
		// fmt.Println(err)
		return false, err
	}

	// Use the temporary kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", tmpfile.Name())
	if err != nil {
		// fmt.Println("Error in building config")
		// fmt.Println(err)
		return false, err
	}

	if checkStr == "kubectlVersion" {
		// Create a new discovery client
		discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
		if err != nil {
			fmt.Println("GetPreCheck: Error creating a new discovery client")
			fmt.Println(err)
			return false, err
		}

		serverVersion, err := discoveryClient.ServerVersion()
		if err != nil {
			fmt.Println("GetPreCheck: Error in getting server version info")
			fmt.Println(err)
			return false, err
		}
		fmt.Println("--- Kubectl version details ---")
		fmt.Println(serverVersion)
		fmt.Println(serverVersion.String())
		if serverVersion.String() == "v1.26.0" || serverVersion.String() == "v1.27.8" {
			return true, nil
		} else {
			return false, nil
		}

	} else {
		// Create a Kubernetes clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Println("GetPreCheck: Error creating a new kubernetes clintset")
			fmt.Println(err)
			return false, err
		}

		// Now you can use the clientset to interact with the Kubernetes cluster
		if checkStr == "namespace" {
			namespace, err := clientset.CoreV1().Namespaces().Get(context.TODO(), "kfaas", metav1.GetOptions{})
			if err != nil {
				fmt.Println("GetPreCheck: Error getting namespace")
				fmt.Println(err)
				return false, err
			}
			fmt.Println("--- Namespace details ---")
			fmt.Println(namespace)
			fmt.Println(namespace.Name)
			if namespace.Name == "kfaas" {
				return false, nil
			} else {
				return true, nil
			}
		} else if checkStr == "clusterRoleBinding" {
			clusterRoleBinding, err := clientset.RbacV1().ClusterRoleBindings().Get(context.TODO(), "installclusterrolebinding", metav1.GetOptions{})
			if err != nil {
				fmt.Println("GetPreCheck: Error getting clusterrolebinding")
				fmt.Println(err)
				return false, err
			}
			fmt.Println("--- ClusterRoleBinding details ---")
			fmt.Println(clusterRoleBinding)
			fmt.Println(clusterRoleBinding.Name)
			if clusterRoleBinding.Name == "installclusterrolebinding" {
				return false, nil
			} else {
				return true, nil
			}
		} else {
			return false, nil
		}
	}
}
