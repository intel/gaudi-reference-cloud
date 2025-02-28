// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package k8s

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	sql "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type K8sClient struct {
	ClientConfig   *rest.Config
	ClientSet      *kubernetes.Clientset
	IksClient      pb.IksClient
	ClusterID      *pb.ClusterID
	GrpcClientConn *grpc.ClientConn
	SshClient      pb.SshPublicKeyServiceClient
	VnetClient     pb.VNetServiceClient
	IstioClientSet *IstioClientSet
}

func readKubeConfig(path string) string {

	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting user home directory:", err)
			return ""
		}

		// Define the file path within the home directory
		path = fmt.Sprintf("%s/.kube/config", homeDir)
	}

	fmt.Print(path)
	// Read the file contents
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading kubeconfig file:", err)
		os.Exit(1)
	}

	// Get the kubeconfig as a string
	kubeconfigString := string(data)

	// Use the kubeconfig string as needed
	fmt.Println("Kubeconfig contents:\n", kubeconfigString)
	return kubeconfigString
}

func (k *K8sClient) GetK8sConfig() error {
	// write a function to get the kubeconfig from the secure vault
	// Provide the kubeconfig string
	// kubeconfigString := readKubeConfig("")
	if k.ClusterID.CloudAccountId == "" || k.ClusterID.Clusteruuid == "" {
		return fmt.Errorf("missing ClusterID. K8sClient must have ClusterID.CloudAccountId and ClusterID.Clusteruuid set")
	}

	kubeConfig, err := k.IksClient.GetKubeConfig(context.Background(), k.ClusterID)
	if err != nil {
		log.Fatalf("Error getting kubeconfig: %+v", err)
		return err
	}

	// Load kubeconfig from the string
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig.Kubeconfig))
	if err != nil {
		// Handle error appropriately
		log.Println("Error parsing kubeconfig:", err)
		return err
	}
	k.ClientConfig = config
	return nil
}

func (k *K8sClient) GetK8sClientSet() error {

	log.Printf("GetK8sClientSet - ClusterID %+v", k.ClusterID)
	err := k.GetK8sConfig()
	if err != nil {
		// Handle error appropriately
		log.Printf("Not able to fetch the kubeconfig for the ClusterID.CloudAccountId: %s and ClusterID.Clusteruuid: %s. Error message: %+v", k.ClusterID.CloudAccountId, k.ClusterID.Clusteruuid, err)
		return err
	}

	// Create a clientset for interacting with Kubernetes resources
	clientset, err := kubernetes.NewForConfig(k.ClientConfig)
	if err != nil {
		// Handle error appropriately
		fmt.Println("Error creating clientset:", err)
		return err
	}
	k.ClientSet = clientset

	// k.IstioClientSet =
	return nil
}

// createTempKubeconfig creates a temporary kubeconfig file from the provided kubeconfig string
func createTempKubeconfig(path, content string) error {
	log.Println("Created token")
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		return err
	}
	return nil
}

func GetIksClusterID(sqlModel *sql.Queries, workspaceID string, serviceID string) (*pb.ClusterID, error) {
	log.Printf("GetIksClusterID - workspaceid %+v", workspaceID)
	var clusterID pb.ClusterID

	if workspaceID != "" {
		workspaceData, err := sqlModel.GetWorkspace(context.TODO(), db.GetWorkspaceParams{
			WorkspaceID: workspaceID,
		})
		if err != nil {
			return nil, fmt.Errorf("no Match found for the workspace id: %s", workspaceID)
		}
		log.Printf("GetIksClusterID - workspaceData %+v", workspaceData)
		clusterID.Clusteruuid = workspaceData.IksID.String
		clusterID.CloudAccountId = workspaceData.CloudAccountID
	} else if serviceID != "" {
		workspaceData, err := sqlModel.GetClusterIdFromServiceId(context.TODO(), pgtype.Text{String: serviceID, Valid: true})
		if err != nil {
			return nil, fmt.Errorf("no Match found for the service id: %s", serviceID)
		}
		clusterID.Clusteruuid = workspaceData.IksID.String
		clusterID.CloudAccountId = workspaceData.CloudAccountID
	} else {
		log.Println("missing both WorkspaceID and ServiceID. This is allowed only when we try to create workspace.")
		// return nil, fmt.Errorf("missing input: WorkspaceID or ServiceID must be provided")
	}

	return &clusterID, nil
}

func GenerateDpaiIksClusterName(name string) string {
	return fmt.Sprintf("iks-dpai-ws-%s", name)
}
