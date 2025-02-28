// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package airflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	helmclient "github.com/mittwald/go-helm-client"

	// "helm.sh/helm/v3/pkg/release"

	// "k8s.io/apimachinery/pkg/api/resource"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment/postgres"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment/workspace"

	//"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment/workspace"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/crypto"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/helm"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/networking"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/storage"

	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

// / Helper functions
// struct to pass the output between the tasks
type TaskOutput struct {
	// Make sure below fields are defined for all the deployments

	// Target resource id.
	ID string
	// This will be used to clean up the Namespace incase the deployment failed in middle. If this field,
	// is empty then it assumes the deployment failed before creating namespace itself.
	Namespace string
	// This will be used to clean up the Nodegroup incase pipeline failed in middle. If this field,
	// is empty then it assumes the deployment failed before creating nodegroup itself.
	ReleaseName        string
	SecretName         string
	RegistrySecretName string
	// Release            *release.Release

	// Optional fields Need to check
	// Airflow      db.Airflow
	ServiceName string
	Validated   bool
	PostgresId  string
	// Postgres     db.Postgres
	DiskSizeInGb int32

	//Added by Joydeep
	//Workspace
	CloudAccountID              string
	IksClusterUUID              string
	IksClusterName              string
	SshKeyName                  string
	SecretNamespace             string
	IstioNamespace              string
	OpenEbsNamespace            string
	NodeGroupID                 string
	VIPID                       int32
	WorkspaceID                 string
	ObjectstoreServicePrincipal string
	StorageSecretName           string

	// DNS Fqdn for all exposed services for a service through Service mesh, and traffic balanced via external Load balancer.
	// TODO: Make compatible for DPAI services to expose multiple services outside the cluster.
	ServiceEndpoint string
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

func parseTaskOutput(output any) (TaskOutput, error) {
	var parsed TaskOutput
	// err := json.Unmarshal(output, &parsed)
	// Type assertion to convert the interface value to MyType
	parsed, ok := output.(TaskOutput)
	if !ok {
		fmt.Println("Conversion failed. The interface value is not of type TaskOutput.")
		return TaskOutput{}, fmt.Errorf("TaskOutput conversion failed")
	}

	return parsed, nil
}

func commitOutput(ctx *deployment.TaskRunContext, params db.CommitAirflowCreateParams) error {
	_, err := ctx.SqlModel.CommitAirflowCreate(context.Background(), params)

	if err != nil {
		log.Printf("Not able to commit the status of the Airflow to the backend DB for the deployment id %s. Resource might have been created.", ctx.DeploymentContext.ID)
		return fmt.Errorf("not able to commit the status of the Airflow to the backend DB for the deployment id %s Task name: %s. Error message: %+v.", ctx.DeploymentContext.ID, ctx.Task.Name, err)
	}
	return nil
}

func getHelmChartReference(ctx *deployment.TaskRunContext, versionId string) (*helm.HelmChartReference, error) {

	version, err := ctx.SqlModel.GetAirflowVersionByName(context.Background(), versionId)
	if err != nil {
		return nil, err
	}

	chartReference, err := utils.ConvertBytesToChartReference(version.ChartReference)
	if err != nil {
		return nil, err
	}

	return &helm.HelmChartReference{
		RepoName:  chartReference.GetRepoName(),
		RepoUrl:   chartReference.GetRepoUrl(),
		ChartName: chartReference.GetChartName(),
		Version:   chartReference.GetVersion(),
		Username:  chartReference.GetUsername(),
		SecretKey: "",
	}, nil
}

func getImageReference(ctx *deployment.TaskRunContext, versionId string) (*helm.ImageReference, error) {

	version, err := ctx.SqlModel.GetAirflowVersionByName(context.Background(), versionId)
	if err != nil {
		return nil, err
	}

	imageReference, err := utils.ConvertBytesToImageReference(version.ImageReference)
	if err != nil {
		return nil, err
	}

	return &helm.ImageReference{
		Repository: imageReference.GetRepository(),
		Tag:        imageReference.GetTag(),
	}, nil
}

func getHelmClient(ctx *deployment.TaskRunContext, namespace string, versionId string) (helmclient.Client, *helm.HelmChartReference, error) {
	helmClient, err := helm.GetHelmClient(ctx.K8sClient.ClientConfig, namespace)
	if err != nil {
		return nil, nil, err
	}
	chartReference, err := getHelmChartReference(ctx, versionId)
	if err != nil {
		return nil, nil, err
	}

	// Define a Private chart repository.
	log.Printf("Start to Log in to Registry: %+v", ctx.DeploymentContext.Context.Conf.DefaultRegistry)
	err = helm.LoginRegistry(ctx.DeploymentContext.Context.Conf.DefaultRegistry)
	if err != nil {
		return nil, nil, err
	}

	return helmClient, chartReference, nil
}

func generateHelmYamlValuesForCreate(ctx *deployment.TaskRunContext) (string, *pb.DpaiAirflowCreateRequest, error) {

	var yamlString string
	yamlValues := map[string]interface{}{}

	deploymentInput, _ := GetDeploymentCreateInput(ctx.DeploymentContext)

	size, err := ctx.SqlModel.GetAirflowSizeByName(context.Background(), deploymentInput.SizeProperties.GetSize())
	if err != nil {
		fmt.Printf("Get Airflow Size: %+v", err)
		return yamlString, nil, err
	}

	if deploymentInput.SizeProperties.NumberOfNodes == 0 {
		deploymentInput.SizeProperties.NumberOfNodes = size.NumberOfNodesDefault
	}

	// Storage related properties
	// storageOutput, err := parseTaskOutput(ctx.GetTaskOutput("PrepareStorageObjects"))
	// if err != nil {
	// 	return yamlString, nil, err
	// }

	// storageSecrets, err := ctx.K8sClient.GetSecret(storageOutput.Namespace, storageOutput.StorageSecretName)
	// if err != nil {
	// 	return yamlString, nil, err
	// }
	// log.Printf("Storage Secrets: %+v", storageSecrets)
	// log.Printf("Storage Endpoint: %+v", string(storageSecrets["endpoint"]))
	// log.Printf("Storage Access Key: %+v", string(storageSecrets["accessKey"]))
	// log.Printf("Storage Secret Key: %+v", string(storageSecrets["secretKey"]))
	// log.Printf("Storage Principal: %+v", string(storageSecrets["principalName"]))

	//scheduler
	yamlValues["scheduler"] = map[string]interface{}{
		"replicas": size.SchedularCountDefault.Int32,
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    size.SchedulerCpuLimit.String,
				"memory": size.SchedulerMemoryLimit.String,
			},
			"requests": map[string]interface{}{
				"cpu":    size.SchedulerCpuRequest.String,
				"memory": size.SchedulerMemoryRequest.String,
			},
		},
	}

	//triggerer
	yamlValues["triggerer"] = map[string]interface{}{
		"replicas": size.TriggerCount.Int32,
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    size.TriggerCpuLimit.String,
				"memory": size.TriggerMemoryLimit.String,
			},
			"requests": map[string]interface{}{
				"cpu":    size.TriggerCpuRequest.String,
				"memory": size.TriggerMemoryRequest.String,
			},
		},
		"persistence": map[string]interface{}{
			"size":    "2Gi",
			"enabled": true,
			//"storageClassName": "efs-sc", //this is hard coded this is required to be derived from IKS using IKS API
		},
	}

	//workers
	yamlValues["workers"] = map[string]interface{}{
		"persistence": map[string]interface{}{
			"size":    "2Gi",
			"enabled": true,
			//"storageClassName": "efs-sc", //this is hard coded this is required to be derived from IKS using IKS API
		},
	}

	//Webserver
	yamlValues["webserver"] = map[string]interface{}{
		"replicas": size.WebserverCount.Int32,
		"defaultUser": map[string]interface{}{
			"username": deploymentInput.WebServerProperties.WebserverAdminUsername, //Input from User
			//"password" : "", //Input from User
		},
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    size.WebserverCpuLimit.String,
				"memory": size.WebserverMemoryLimit.String,
			},
			"requests": map[string]interface{}{
				"cpu":    size.WebserverCpuRequest.String,
				"memory": size.WebserverMemoryRequest.String,
			},
		},
	}

	//redis size
	yamlValues["redis"] = map[string]interface{}{
		"persistence": map[string]interface{}{
			"size":    size.RedisDiskSize.String,
			"enabled": true,
			//"storageClassName": "efs-sc", //this is hard coded this is required to be derived from IKS using IKS API
		},
	}

	//log directory size
	yamlValues["logs"] = map[string]interface{}{
		"persistence": map[string]interface{}{
			"size":    size.LogDirectoryDiskSize.String,
			"enabled": false,
			//"storageClassName": "efs-sc", //this is hard coded this is required to be derived from IKS using IKS API
		},
	}

	//Config
	// yamlValues["config"] = map[string]interface{}{
	// 	"logging": map[string]interface{}{
	// 		"remote_base_log_folder": deploymentInput,
	// 	},
	// 	//"core": map[string]interface{}{
	// 	//	"load_examples": "True",
	// 	//},
	// }

	//dags
	yamlValues["dags"] = map[string]interface{}{
		"persistence": map[string]interface{}{
			"size":    "2Gi",
			"enabled": false,
			//"storageClassName": "efs-sc", //this is hard coded this is required to be derived from IKS using IKS API
			//"accessMode":       "ReadWriteMany",
		},
		"gitSync": map[string]interface{}{
			"enabled": false,
		},
	}

	//plugin
	// yamlValues["plugins"] = map[string]interface{}{
	// 	"persistence": map[string]interface{}{
	// 		"size": "2Gi",
	// 		"enabled":          false,
	// 		//"storageClassName": "efs-sc", //this is hard coded this is required to be derived from IKS using IKS API
	// 		//"accessMode":       "ReadWriteMany",
	// 	},
	// }

	//postgresql
	yamlValues["postgresql"] = map[string]interface{}{
		"enabled": false,
	}

	//createUserJob
	yamlValues["createUserJob"] = map[string]interface{}{
		"applyCustomEnv": false,
		"useHelmHooks":   false,
		"annotations": map[string]interface{}{
			"sidecar.istio.io/inject": "'false'",
		},
	}

	//migrateDatabaseJob
	yamlValues["migrateDatabaseJob"] = map[string]interface{}{
		"useHelmHooks":   false,
		"applyCustomEnv": false,
		"annotations": map[string]interface{}{
			"sidecar.istio.io/inject": "'false'",
		},
	}

	output, _ := parseTaskOutput(ctx.GetTaskOutput("HelmInstallAirflowDb"))

	psqlDb, err := ctx.SqlModel.GetPostgresById(context.Background(), output.PostgresId)
	if err != nil {
		fmt.Printf("Get Postgres: %+v", err)
		return yamlString, nil, err
	}

	psqlSecretReference, err := utils.ConvertBytesToSecretReference(psqlDb.AdminPasswordSecretReference)
	if err != nil {
		return yamlString, nil, err
	}
	psqlSecret, _ := ctx.K8sClient.GetSecret(output.Namespace, psqlSecretReference.GetSecretName())

	yamlValues["data"] = map[string]interface{}{
		"metadataConnection": map[string]interface{}{
			"db":       "airflow",
			"host":     psqlDb.ServerUrl.String,
			"pass":     string(psqlSecret[psqlSecretReference.GetSecretKeyName()]),
			"port":     5432,
			"protocol": "postgresql",
			"sslmode":  "disable",
			"user":     "postgres",
		},
	}

	//Extra Init Container

	log.Printf("Yaml Values: %+v", yamlValues)
	// Add nodeselector for scheduling airflo only on IKS nodegroup managed for airflow, IKS Nodegroup controller add labels with ng-<uuid> for instances managed in a nodegroup
	log.Printf("Adding Node Affinity to deploy Airflow pods over IKS Nodegroup of Airflow %s", output.NodeGroupID)
	// yamlValues["nodeSelector"] = output.NodeGroupID
	yamlString, _ = utils.ConvertToYAMLString(yamlValues)
	log.Printf("Yaml String: %+v", yamlString)
	return yamlString, deploymentInput, nil
}

func CreateNodeGroup(ctx *deployment.TaskRunContext) (any, error) {

	deploymentInput, err := GetDeploymentCreateInput(ctx.DeploymentContext)
	if err != nil {
		return nil, err
	}
	log.Printf("Deployment Inputs: %v", deploymentInput)

	// output := TaskOutput{
	// 	IksClusterUUID: "cl-oshctm7ooq",
	// 	WorkspaceID:    "workspaceid",
	// 	CloudAccountID: "513913963588",
	// 	NodeGroupId:    "ng-okq4psrnsm",
	// 	SshKeyName:     "ssh-key-dpai-airflow", //k8s.GenerateDpaiIksClusterName(deploymentInput.WorkspaceName),
	// }

	output, err := parseTaskOutput(ctx.GetTaskOutput("CreateWorkspace"))
	if err != nil {
		return nil, err
	}
	log.Printf("Inside Airflow Create Node Group output: %v", output)
	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	airflowSize, err := ctx.SqlModel.GetAirflowSizeByName(context.TODO(), deploymentInput.SizeProperties.GetSize())
	if err != nil {
		return nil, err
	}

	description := "This node group is created and managed by DPAI -- Please do not delete this -- Deleting this will break the DPAI services running on top of it"
	nodeGroup, err := ctx.K8sClient.CreateNodeGroup(&pb.CreateNodeGroupRequest{
		CloudAccountId: output.CloudAccountID,
		Clusteruuid:    output.IksClusterUUID,
		Name:           deploymentInput.Name,
		InstanceType:   "iks-cluster",
		Description:    &description,
		Instancetypeid: airflowSize.NodeSizeID,
		Count:          deploymentInput.SizeProperties.NumberOfNodes,
		Sshkeyname: []*pb.SshKey{
			{Sshkey: output.SshKeyName},
		},
	})

	if err != nil {
		return nil, err
	}

	log.Printf("Getting task output from fn: %+v", nodeGroup)

	output.NodeGroupID = nodeGroup.Nodegroupuuid

	err = commitOutput(ctx, db.CommitAirflowCreateParams{
		ID:          ctx.DeploymentContext.Context.ServiceId,
		NodeGroupID: pgtype.Text{String: nodeGroup.Nodegroupuuid, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

func CreateNamespace(ctx *deployment.TaskRunContext) (any, error) {
	output, err := parseTaskOutput(ctx.GetTaskOutput("CreateWorkspace"))
	if err != nil {
		return nil, err
	}

	log.Printf("Inside Airflow CreateNamespace. Output: %v", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	istioInjectionLabels := map[string]string{
		"istio-injection": "enabled",
	}

	namespace, err := ctx.K8sClient.CreateNamespace(ctx.DeploymentContext.ID, false, istioInjectionLabels)
	if err != nil {
		return nil, err
	}

	output = mergeTaskOutput(output, TaskOutput{
		Namespace:   namespace.Name,
		ReleaseName: "airflow",
	})

	log.Printf("Completed: Inside CreateNamespace. Output: %v", output)

	return output, nil
}

func CreateSecret(ctx *deployment.TaskRunContext) (any, error) {
	output, err := parseTaskOutput(ctx.GetTaskOutput("CreateNamespace"))
	if err != nil {
		return nil, err
	}

	log.Printf("Inside CreateSecret: %+v", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	log.Printf("Createsecret: Initialized K8s")

	// Create private docker registry secrets
	log.Printf("Createsecret: Fetching docker secret")
	registrySecretName := deployment.DockerRegistrySecret
	config := ctx.DeploymentContext.Context.Conf
	dockerPasword, err := utils.ReadSecretFile(config.DefaultRegistry.PasswordFile)
	if err != nil {
		return nil, err
	}
	log.Printf("Createsecret: Create docker secret")
	err = ctx.K8sClient.CreateSecretDockerConfigJson(output.Namespace, registrySecretName, k8s.DockerConfigSecretData{
		UserName: config.DefaultRegistry.Username,
		Password: *dockerPasword,
		Server:   config.DefaultRegistry.Host,
	})
	if err != nil {
		return nil, err
	}

	output = mergeTaskOutput(output, TaskOutput{
		SecretName:         output.ReleaseName,
		RegistrySecretName: registrySecretName,
	})
	return output, nil
}

func UpdateAirflowUserPassword(ctx *deployment.TaskRunContext) (any, error) {
	output, err := parseTaskOutput(ctx.GetTaskOutput("HelmInstall"))
	if err != nil {
		return nil, err
	}

	input, err := GetDeploymentCreateInput(ctx.DeploymentContext)
	if err != nil {
		return nil, err
	}

	username := input.WebServerProperties.WebserverAdminUsername
	secretId, err := strconv.Atoi(input.WebServerProperties.WebserverAdminPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to convert WebserverAdminPasswordSecretId to int: %v", err)
	}

	password, err := crypto.DecryptedPassword(ctx.SqlModel, *ctx.DeploymentContext.Context.Conf, int32(secretId))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt the password: %v", err)
	}

	log.Printf("Inside UpdateAirflowUserPassword: %+v", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	log.Printf("Fetching the worker pod")
	pods, err := ctx.K8sClient.ListPods(output.Namespace)
	if err != nil {
		return nil, err
	}

	workerPodName := ""
	for _, pod := range pods {
		if strings.Contains(pod.Name, "airflow-webserver-") {
			log.Printf("Worker Pod Name: %+v", pod.Name)
			workerPodName = pod.Name
			break
		}
	}
	//TODO: Read password from the secret and update the password

	resetPasswordCommand := []string{"airflow", "users", "reset-password", "-u", username, "-p", password}
	err = ctx.K8sClient.ExecInPod(output.Namespace, workerPodName, resetPasswordCommand)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func CreateWorkspace(ctx *deployment.TaskRunContext) (any, error) {
	deploymentInput, err := GetDeploymentCreateInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	deploymentId := getAirflowWorkspaceDeploymentId(ctx.DeploymentContext.ID) //Need to change this

	input := &pb.DpaiWorkspaceCreateRequest{
		CloudAccountId: deploymentInput.CloudAccountId,
		Name:           deploymentInput.GetWorkspaceName(),
		Description:    "This workspace is created and managed by DPAI -- Please do not delete this -- Deleting this will break the DPAI services running on top of it",
		Tags: map[string]string{
			"isManagedResource": "true",
			"managedBy":         deploymentInput.Name,
			"service":           "dpai",
		},
	}

	// convert input payload into the bytes
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	workspaceId, err := utils.GenerateUniqueServiceId(ctx.SqlModel, pb.DpaiServiceType_DPAI_WORKSPACE, false)
	if err != nil {
		return nil, err
	}

	//Create workspace added on 11/20
	_, err = ctx.SqlModel.CreateWorkspace(context.Background(), db.CreateWorkspaceParams{
		ID:                    workspaceId,
		CloudAccountID:        deploymentInput.CloudAccountId,
		Name:                  deploymentInput.WorkspaceName,
		Description:           pgtype.Text{String: "This workspace is created as a part of deployment. Do not delete this workspace.", Valid: true},
		DeploymentID:          deploymentId,
		DeploymentStatusState: pb.DpaiDeploymentState_DPAI_PENDING.String(),
		CreatedBy:             "Internal Process",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create the workspace. Error message: %+v", err)
	}

	_, err = ctx.SqlModel.CreateDeployment(context.Background(), db.CreateDeploymentParams{
		ID:                 deploymentId,
		CloudAccountID:     pgtype.Text{String: deploymentInput.CloudAccountId, Valid: true},
		WorkspaceID:        pgtype.Text{String: workspaceId, Valid: true},
		ServiceType:        pb.DpaiServiceType_DPAI_WORKSPACE.String(),
		ChangeIndicator:    pb.DpaiDeploymentChangeIndicator_DPAI_CREATE.String(),
		CreatedBy:          "Internal Process",
		InputPayload:       jsonData,
		ParentDeploymentID: pgtype.Text{String: ctx.DeploymentContext.ID, Valid: true},
	})
	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	_, err = workspace.CreateDeployment(deployment.DeploymentInputContext{
		ID:      getAirflowWorkspaceDeploymentId(ctx.DeploymentContext.ID),
		SqlPool: ctx.DeploymentContext.Context.SqlPool,
		//ParentDeploymentId: ctx.DeploymentContext.ID,
		Conf: ctx.DeploymentContext.Context.Conf,
	})
	if err != nil {
		return nil, err
	}

	workspace, err := ctx.SqlModel.GetWorkspace(context.Background(), db.GetWorkspaceParams{
		WorkspaceID:    workspaceId,
		CloudAccountID: pgtype.Text{String: deploymentInput.CloudAccountId, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get the workspace. Error message: %+v", err)
	}

	output := TaskOutput{
		WorkspaceID:    workspace.ID,
		CloudAccountID: deploymentInput.CloudAccountId,
		IksClusterUUID: workspace.IksID.String,
		IksClusterName: k8s.GenerateDpaiIksClusterName(deploymentInput.WorkspaceName),
		SshKeyName:     k8s.GenerateDpaiIksClusterName(deploymentInput.WorkspaceName),
	}
	log.Printf("Inside Airflow CreateWorkspace. Output: %+v", output)
	err = commitOutput(ctx, db.CommitAirflowCreateParams{
		ID:           ctx.DeploymentContext.Context.ServiceId,
		WorkspaceID:  pgtype.Text{String: workspaceId, Valid: true},
		IksClusterID: pgtype.Text{String: workspace.IksID.String, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return output, nil

}

// Storage
func PrepareStorageObjects(ctx *deployment.TaskRunContext) (any, error) {

	log.Println("Inside PrepareStorageObjects")
	output, err := parseTaskOutput(ctx.GetTaskOutput("CreateNamespace"))
	if err != nil {
		return nil, err
	}

	deploymentInput, err := GetDeploymentCreateInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	helmClient, _, err := getHelmClient(ctx, output.Namespace, deploymentInput.Version)
	if err != nil {
		return nil, err
	}

	airflowObjectStorePrincipalName := ctx.DeploymentContext.Context.ServiceId
	log.Printf("Airflow Object Store Principal Name: %+v", airflowObjectStorePrincipalName)

	st := storage.Storage{}
	log.Println("Getting the storage client")
	err = st.GetStorageClient(ctx.DeploymentContext.Context.Conf, deploymentInput.CloudAccountId)
	if err != nil {
		return nil, err
	}

	bucketId := deploymentInput.StorageProperties.BucketId
	bucket, err := st.GetBucket(bucketId)
	if err != nil {
		return nil, err
	}
	log.Printf("Bucket Name: %+v", bucket.Metadata.Name)
	log.Printf("Bucket ID: %+v", bucketId)

	user, err := st.CreateAirflowObjectUser(bucket.Metadata.Name, airflowObjectStorePrincipalName)
	if err != nil {
		return nil, err
	}

	objEndpoint, objAccessKey, objSecretKey := user.Status.Principal.Cluster.AccessEndpoint, user.Status.Principal.Credentials.AccessKey, user.Status.Principal.Credentials.SecretKey
	log.Printf("Endpoint: %+v", objEndpoint)
	log.Printf("AccessKey: %+v", objAccessKey)
	log.Printf("SecretKey: %+v", objSecretKey)

	logFolder := deploymentInput.GetStorageProperties().GetPath().GetLogFolder()
	dagsFolder := deploymentInput.GetStorageProperties().GetPath().GetDagFolderPath()
	pluginsFolder := deploymentInput.GetStorageProperties().GetPath().GetPluginFolderPath()
	requirementsFolder := deploymentInput.GetStorageProperties().GetPath().GetRequirementPath()

	folders := []string{logFolder, dagsFolder, pluginsFolder}

	//Helm chart for Storage which will check and create the folders in the S3 bucket
	//Build the Yaml Values
	var yamlString string
	yamlValues := map[string]interface{}{}

	//aws
	yamlValues["aws"] = map[string]interface{}{
		"accessKeyId":     objAccessKey,
		"secretAccessKey": objSecretKey,
		"endpointURL":     objEndpoint,
	}

	//bucket
	yamlValues["bucket"] = map[string]interface{}{
		"name": bucket.Metadata.Name,
	}

	//folders
	yamlValues["folders"] = folders

	yamlString, err = utils.ConvertToYAMLString(yamlValues)
	if err != nil {
		return nil, fmt.Errorf("unable to ConvertToYAMLString. Error message: %+v", err)
	}

	log.Printf("Yaml String: %+v", yamlString)

	registrypassword, err := utils.ReadSecretFile(ctx.DeploymentContext.Context.Conf.DefaultRegistry.PasswordFile)
	if err != nil {
		return nil, err
	}
	log.Printf("Registry Passowrd: %+v", registrypassword)
	log.Printf("Registry Username: %+v", &ctx.DeploymentContext.Context.Conf.DefaultRegistry.Username)

	storagefilepath, err := helm.DownloadChart("/tmp", "icir.cps.intel.com/dpai", "s3-folder-creator", "1.0.0", &ctx.DeploymentContext.Context.Conf.DefaultRegistry.Username, registrypassword)
	if err != nil {
		return nil, fmt.Errorf("unable to Download Chart s3-folder-creator . Error message: %+v", err)
	}

	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName: bucket.Metadata.Name,
		ChartName:   storagefilepath,
		Version:     "1.0.0",
		Namespace:   output.Namespace,
		UpgradeCRDs: true,
		//ReuseValues: true,
		ValuesYaml: yamlString,
		Wait:       true,
		Atomic:     true,
		Timeout:    30 * time.Minute,
		// CreateNamespace: true,
		WaitForJobs: true,
	}

	// Install a chart release.
	log.Printf("Debug: Starting to Install Storage \n")
	_, err = helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error Installing Storage Service. %+v", err)
		return nil, err
	}

	/*
		minio := storage.Minio{
			Endpoint:        user.Status.Principal.Cluster.AccessEndpoint,
			AccessKeyID:     user.Status.Principal.Credentials.AccessKey,
			SecretAccessKey: user.Status.Principal.Credentials.SecretKey,
		}
		err = minio.GetMinioClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get the minio client. Error message: %s", err)
		}

		// Create log folder
		log.Println("Creating the log folder")
		logFolder := deploymentInput.GetStorageProperties().GetPath().GetLogFolder()
		if logFolder == "" {
			logFolder = fmt.Sprintf("s3://%s/logs", bucket.Metadata.Name)
		}
		logFolder = strings.TrimSuffix(logFolder, fmt.Sprintf("/%s", ctx.DeploymentContext.Context.ServiceId))
		logFolder = fmt.Sprintf("%s/%s", logFolder, ctx.DeploymentContext.Context.ServiceId)
		logFolderPath, err := storage.ExtractFolderPath(logFolder)
		if err != nil {
			return nil, err
		}
		log.Printf("Creating the log folder: Bucketname %v, Logfolder: %v", bucket.Metadata.Name, logFolderPath)
		err = minio.CreateFolder(bucket.Metadata.Name, logFolderPath)
		if err != nil {
			return nil, err
		}

		// create dags folder
		log.Println("Creating the dag folder")
		dagsFolder := deploymentInput.GetStorageProperties().GetPath().GetDagFolderPath()
		if dagsFolder == "" {
			dagsFolder = fmt.Sprintf("s3://%s/dags", bucket.Metadata.Name)
		}
		dagsFolderPath, err := storage.ExtractFolderPath(dagsFolder)
		if err != nil {
			return nil, err
		}
		err = minio.CreateFolder(bucket.Metadata.Name, dagsFolderPath)
		if err != nil {
			return nil, err
		}

		// create plugins folder
		log.Println("Creating the plugin folder")
		pluginsFolder := deploymentInput.GetStorageProperties().GetPath().GetPluginFolderPath()
		if pluginsFolder == "" {
			pluginsFolder = fmt.Sprintf("s3://%s/plugins", bucket.Metadata.Name)
		}
		pluginsFolderPath, err := storage.ExtractFolderPath(pluginsFolder)
		if err != nil {
			return nil, err
		}
		err = minio.CreateFolder(bucket.Metadata.Name, pluginsFolderPath)
		if err != nil {
			return nil, err
		}

		// create requirements folder
		log.Println("Creating the requirement folder")
		requirementsFolder := deploymentInput.GetStorageProperties().GetPath().GetRequirementPath()
		if requirementsFolder == "" {
			requirementsFolder = fmt.Sprintf("s3://%s/requirements", bucket.Metadata.Name)
		}
		requirementsFolderPath, err := storage.ExtractFolderPath(requirementsFolder)
		if err != nil {
			return nil, err
		}
		err = minio.CreateFolder(bucket.Metadata.Name, requirementsFolderPath)
		if err != nil {
			return nil, err
		}
	*/

	// persist the object store principal name in k8s secret
	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}
	secretData := map[string][]byte{
		"endpoint":      []byte(objEndpoint),
		"principalName": []byte(airflowObjectStorePrincipalName),
		"accessKey":     []byte(objAccessKey),
		"secretKey":     []byte(objSecretKey),
	}

	output = mergeTaskOutput(output, TaskOutput{
		ObjectstoreServicePrincipal: user.Metadata.Name,
		StorageSecretName:           fmt.Sprintf("dpai__%s", bucket.Metadata.Name),
	})

	err = ctx.K8sClient.CreateSecret(output.Namespace, output.StorageSecretName, secretData)
	if err != nil {
		return nil, fmt.Errorf("failed to create the secret")
	}

	err = commitOutput(ctx, db.CommitAirflowCreateParams{
		ID:               ctx.DeploymentContext.Context.ServiceId,
		BucketPrincipal:  pgtype.Text{String: user.Metadata.Name, Valid: true},
		DagFolderPath:    pgtype.Text{String: dagsFolder, Valid: true},
		LogFolder:        pgtype.Text{String: logFolder, Valid: true},
		PluginFolderPath: pgtype.Text{String: pluginsFolder, Valid: true},
		RequirementPath:  pgtype.Text{String: requirementsFolder, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

func HelmInstallAirflowDb(ctx *deployment.TaskRunContext) (any, error) {

	secretOutput, err := parseTaskOutput(ctx.GetTaskOutput("CreateSecret"))
	if err != nil {
		return nil, fmt.Errorf("failed to get the CreateSecret output. Error message: %+v", err)
	}
	nodeGroupOutput, err := parseTaskOutput(ctx.GetTaskOutput("CreateNodeGroup"))
	if err != nil {
		return nil, fmt.Errorf("failed to get the CreateNodeGroup output. Error message: %+v", err)
	}
	output := mergeTaskOutput(secretOutput, nodeGroupOutput)
	log.Printf("Inside HelmInstallAirflowDb: %+v", output)

	deploymentInput, err := GetDeploymentCreateInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	deploymentId := getAirflowDbDeploymentId(ctx.DeploymentContext.ID)
	log.Printf("HelmInstallAirflowDb: DeploymentInput %+v ", deploymentInput)
	serviceId := fmt.Sprintf("db-%s", ctx.DeploymentContext.ID[4:])
	version, err := ctx.SqlModel.GetAirflowVersionByName(context.TODO(), deploymentInput.Version)
	if err != nil {
		return nil, err
	}
	size, err := ctx.SqlModel.GetAirflowSizeByName(context.TODO(), deploymentInput.SizeProperties.GetSize())
	if err != nil {
		return nil, err
	}
	// generate secret
	password, err := utils.GenerateRandomPassword(16)
	if err != nil {
		return nil, err
	}
	secretData := map[string][]byte{
		"password":        []byte(password),
		"repmgr-password": []byte(password),
		"admin-password":  []byte(password),
	}

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}
	err = ctx.K8sClient.CreateSecret(deployment.SecretNamespace, deploymentId, secretData)
	if err != nil {
		return nil, fmt.Errorf("failed to create the secret")
	}

	input := &pb.DpaiPostgresCreateRequest{
		WorkspaceId: output.WorkspaceID,
		//Name:        fmt.Sprintf("airflowdb-%s", output.ReleaseName),
		Name:        "airflowdb",
		Description: fmt.Sprintf("Backend database for the Airflow %s", output.ReleaseName),
		VersionId:   version.BackendDatabaseVersionID,
		SizeProperties: &pb.DpaiPostgresSizeProperties{
			SizeId: size.BackendDatabaseSizeID.String,
		},
		AdminProperties: &pb.DpaiPostgresAdminProperties{
			AdminUsername: "admin",
			AdminPasswordSecretReference: &pb.DpaiSecretReference{
				SecretName:    deploymentId,
				SecretKeyName: "password",
			},
		},
		Tags: map[string]string{
			"isManagedResource": "true",
			"managedBy":         deploymentInput.Name,
		},
		OptionalProperties: &pb.DpaiPostgresOptionalProperties{
			InitialDatabaseName: "airflow",
		},
	}

	// convert input payload into the bytes
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	tags, err := json.Marshal(deploymentInput.GetTags())
	if err != nil {
		return nil, err
	}

	adminSecret, err := utils.ConvertSecretReferenceToBytes(&pb.DpaiSecretReference{
		SecretName:    input.Name,
		SecretKeyName: input.GetAdminProperties().GetAdminPasswordSecretReference().GetSecretKeyName(),
	})
	if err != nil {
		return nil, err
	}
	//Create Postgres
	params := db.CreatePostgresParams{
		CloudAccountID:               output.CloudAccountID,
		WorkspaceID:                  output.WorkspaceID,
		ID:                           serviceId,
		Name:                         input.Name,
		Description:                  pgtype.Text{String: input.Description, Valid: true},
		VersionID:                    input.VersionId,
		SizeID:                       input.SizeProperties.GetSizeId(),
		AdminUsername:                input.AdminProperties.GetAdminUsername(),
		AdminPasswordSecretReference: adminSecret,
		InitialDatabaseName:          pgtype.Text{String: input.OptionalProperties.GetInitialDatabaseName(), Valid: true},
		Tags:                         tags,
		CreatedBy:                    "Internal Process",
		DeploymentID:                 deploymentId,
		DeploymentStatusState:        pb.DpaiDeploymentState_DPAI_PENDING.String(),
		DeploymentStatusDisplayName:  pgtype.Text{String: "Pending", Valid: true},
	}

	_, err = ctx.SqlModel.CreatePostgres(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("failed to create the postgres. Error message: %+v", err)
	}

	_, err = ctx.SqlModel.CreateDeployment(context.Background(), db.CreateDeploymentParams{
		ID:                 deploymentId,
		CloudAccountID:     pgtype.Text{String: output.CloudAccountID, Valid: true},
		WorkspaceID:        pgtype.Text{String: output.WorkspaceID, Valid: true},
		ServiceID:          pgtype.Text{String: serviceId, Valid: true},
		ServiceType:        pb.DpaiServiceType_DPAI_POSTGRES.String(),
		ChangeIndicator:    pb.DpaiDeploymentChangeIndicator_DPAI_CREATE.String(),
		CreatedBy:          "Internal Process",
		InputPayload:       jsonData,
		ParentDeploymentID: pgtype.Text{String: ctx.DeploymentContext.ID, Valid: true},
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	resourceId, err := postgres.CreateDeployment(deployment.DeploymentInputContext{
		ID:                 getAirflowDbDeploymentId(ctx.DeploymentContext.ID), //fmt.Sprintf("airflowdb-%s", strings.SplitN(ctx.ID, "-", 2)[1]),
		SqlPool:            ctx.DeploymentContext.Context.SqlPool,
		ParentDeploymentId: ctx.DeploymentContext.ID,
		Conf:               ctx.DeploymentContext.Context.Conf,
	})

	if err != nil {
		return nil, err
	}

	output = mergeTaskOutput(output, TaskOutput{PostgresId: resourceId})

	err = commitOutput(ctx, db.CommitAirflowCreateParams{
		ID:                ctx.DeploymentContext.Context.ServiceId,
		BackendDatabaseID: pgtype.Text{String: resourceId, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

func HelmInstall(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Started : Airflow HelmInstallOrUpgrade: %v \n", ctx.GetTaskOutput("HelmInstallAirflowDb"))

	output, err := parseTaskOutput(ctx.GetTaskOutput("HelmInstallAirflowDb"))
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

	//output, _ = parseTaskOutput(ctx.GetTaskOutput("HelmInstallAirflowDb"))
	deploymentInput, err := GetDeploymentCreateInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}
	helmClient, chartReference, err := getHelmClient(ctx, output.Namespace, deploymentInput.Version)
	if err != nil {
		return nil, err
	}

	yamlValues, _, _ := generateHelmYamlValuesForCreate(ctx)
	log.Printf("Chart Reference: %+v ", chartReference)
	log.Printf("Output: %+v ", output)
	log.Printf("Chart Name : %+v", chartReference.ChartName)
	log.Printf("Chart URL : %+v", chartReference.RepoUrl)
	log.Printf("Chart Version : %+v", chartReference.Version)

	//Start Workaround for Airflow Helm Chart
	registrypassword, err := utils.ReadSecretFile(ctx.DeploymentContext.Context.Conf.DefaultRegistry.PasswordFile)
	if err != nil {
		return nil, err
	}

	aiflowfilepath, err := helm.DownloadChart("/tmp", chartReference.RepoUrl, chartReference.ChartName, chartReference.Version, &ctx.DeploymentContext.Context.Conf.DefaultRegistry.Username, registrypassword)
	if err != nil {
		return nil, fmt.Errorf("unable to Download Chart airflow . Error message: %+v", err)
	}

	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName: output.ReleaseName, //output.ReleaseName, Need to pass airflow name in the initcontainers
		ChartName:   aiflowfilepath,     //chartReference.ChartName, Changed Chartname to Chart path
		Version:     chartReference.Version,
		Namespace:   output.Namespace,
		UpgradeCRDs: true,
		//ReuseValues: true,
		ValuesYaml: yamlValues,
		Wait:       true,
		Atomic:     true,
		Timeout:    30 * time.Minute,
		// CreateNamespace: true,
		WaitForJobs: true,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	log.Printf("Debug: Starting to Install Airflow \n")
	_, err = helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error while installing the service. %+v", err)
		return nil, err
	}
	// output = mergeTaskOutput(output, TaskOutput{
	// 	Release:     rel,
	// 	ServiceName: deploymentInput.GetName(),
	// })

	log.Printf("Completed: Inside HelmInstall. Output: %v", output)

	return output, nil
}

func EnableServiceMeshIngressNetwork(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Started: enabling service gateways, networking and security to expose %s\n", ctx.DeploymentContext.Context.ServiceId)

	helmInstallOutput := ctx.GetTaskOutput("HelmInstall")
	log.Printf("The Raw TaskInput for Installing Service mesh %+v", helmInstallOutput)

	output, err := parseTaskOutput(helmInstallOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to get the HelmInstall output. Error message: %+v", err)
	}
	log.Printf("The Parsed TaskInput for Installing Service mesh %+v", output)

	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}

	err = ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	rootCtx := context.Background()

	// For this release only expose a single service, since it internally creates DNS record, cert, istio gateway, istio vs for vhost routing
	exposedK8sServiceComponents := map[string]int{
		"airflow-webserver": 8080,
	}

	var mesh networking.NetworkResourceProvider = &networking.Network{
		ServiceType:            pb.DpaiServiceType_DPAI_AIRFLOW.String(),
		ServiceNamespace:       output.Namespace,
		ServiceId:              ctx.DeploymentContext.Context.ServiceId,
		ExposedK8sServiceNames: exposedK8sServiceComponents,
		CloudAccountID:         output.CloudAccountID,
		WorkspaceId:            output.WorkspaceID,
		IksClusterUUID:         output.IksClusterUUID,
		K8sClientConfig:        ctx.K8sClient.ClientConfig,
		SqlModel:               ctx.SqlModel,
		K8sClient:              &ctx.K8sClient,
		ServiceConfig:          ctx.DeploymentContext.Context.Conf,
	}

	log.Println("Geenrated Client Resource to provision network ", mesh)

	mesh.InitDnsService(rootCtx)                              // Initialize the DNS service Client via DNS Service proivider
	mesh.InitIstioClient(rootCtx, ctx.K8sClient.ClientConfig) // Initialize the Istio Client via Istio ClientSet provider

	networkOutput, err := mesh.CreateDnsCnameRecord(rootCtx)
	if isNetworkErr := networking.IsNetworkError(err); isNetworkErr {
		log.Printf("Error while creating the DNS Cname Record for Service. Error message: %+v %s Service Id: %s", err, "airflow", ctx.DeploymentContext.Context.ServiceId)
		return nil, err
	}
	networkOutput, err = mesh.CreateIstioGatewayCertificate(rootCtx, networkOutput)
	if isNetworkErr := networking.IsNetworkError(err); isNetworkErr {
		log.Printf("Error while creating the Istio Ingress Gatewy self-signed cert. Error message: %+v", err)
		return nil, err
	}
	networkOutput, err = mesh.CreateIstioGateway(rootCtx, networkOutput)
	if isNetworkErr := networking.IsNetworkError(err); isNetworkErr {
		log.Printf("Error while creating the Istio Gateway. Error message: %+v", err)
		return nil, err
	}
	networkOutput, err = mesh.CreateIstioVirtualService(rootCtx, networkOutput)
	if isNetworkErr := networking.IsNetworkError(err); isNetworkErr {
		log.Printf("Error while creating the Istio Virtual Service. Error message: %+v", err)
		return nil, err
	}
	networkOutput, err = mesh.CreateIstioDestinationRule(rootCtx, networkOutput)
	if isNetworkErr := networking.IsNetworkError(err); isNetworkErr {
		log.Printf("Error while creating the Istio Virtual Service. Error message: %+v", err)
		return nil, err
	}

	output = mergeTaskOutput(output, TaskOutput{
		ServiceEndpoint: fmt.Sprintf("https://%s", mesh.GetExposedServiceEndpoint(rootCtx, networkOutput)),
	})
	return output, nil
}

func Validate(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Started: Inside ValidateAirflow with Service Mesh Enabled \n")
	time.Sleep(10 * time.Second)

	output, _ := parseTaskOutput(ctx.GetTaskOutput("EnableServiceMeshIngressNetwork"))
	log.Printf("%T: %+v", output, output)

	output = mergeTaskOutput(output, TaskOutput{Validated: true})
	log.Printf("Completed: Inside ValidateAirflow Output: %v", output)

	return output, nil
}

func CommitCreate(ctx *deployment.TaskRunContext) (any, error) {

	log.Printf("Started: Inside CommitCreateAirflow\n")

	deploymentInput, err := GetDeploymentCreateInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %s", err)
	}

	output, err := parseTaskOutput(ctx.GetTaskOutput("ValidateAirflow"))
	if err != nil {
		return nil, fmt.Errorf("failed to get the HelmInstall output. Error message: %+v", err)
	}

	data, err := ctx.SqlModel.CommitAirflowCreate(context.Background(), db.CommitAirflowCreateParams{
		ID:                          ctx.DeploymentContext.Context.ServiceId,
		BucketPrincipal:             pgtype.Text{String: deploymentInput.StorageProperties.GetBucketPrincipal(), Valid: true},
		Endpoint:                    pgtype.Text{String: output.ServiceEndpoint, Valid: true},
		DeploymentStatusState:       pgtype.Text{String: pb.DpaiDeploymentState_DPAI_SUCCESS.String(), Valid: true},
		DeploymentStatusDisplayName: pgtype.Text{String: "Success", Valid: true},
		DeploymentStatusMessage:     pgtype.Text{String: "Successfully deployed the Airflow instance.", Valid: true},
	})

	if err != nil {
		log.Printf("Not able to commit the status of the Airflow to the backend DB for the deployment id %s. Resource might have been created.", ctx.DeploymentContext.ID)
		return nil, err
	}

	log.Printf("Completed: Inside CommitCreateAirflow\n")
	output.ID = data.ID
	return output, nil
}

func getAirflowDbDeploymentId(parentDeploymentId string) string {
	if parentDeploymentId == "" {
		parentDeploymentId = "-"
	}
	return fmt.Sprintf("dep-db-%s", strings.SplitN(parentDeploymentId, "-", 2)[1])
}

func getAirflowWorkspaceDeploymentId(parentDeploymentId string) string {
	if parentDeploymentId == "" {
		parentDeploymentId = "-"
	}
	return fmt.Sprintf("dep-ws-%s", strings.SplitN(parentDeploymentId, "-", 2)[1])
}

// Delete tasks

func DeleteNamespace(ctx *deployment.TaskRunContext) (any, error) {
	deploymentInput, _ := GetDeploymentDeleteInput(ctx.DeploymentContext)

	data, err := ctx.SqlModel.GetAirflowById(context.Background(), db.GetAirflowByIdParams{ID: deploymentInput.GetId()})
	if err != nil {
		return nil, err
	}

	namespaceName := data.DeploymentID
	k8sClient := ctx.K8sClient

	_, err = k8sClient.DeleteNamespace(namespaceName, true)
	if err != nil {
		return nil, err
	}

	log.Printf("Completed: Inside DeleteNamespace: %v", namespaceName)

	return TaskOutput{}, nil
}

func DeleteSecret(ctx *deployment.TaskRunContext) (any, error) {
	deploymentInput, err := GetDeploymentDeleteInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	data, err := ctx.SqlModel.GetAirflowById(context.Background(), db.GetAirflowByIdParams{ID: deploymentInput.GetId()})
	if err != nil {
		return nil, err
	}

	secretName := data.DeploymentID
	k8sClient := ctx.K8sClient
	err = k8sClient.DeleteSecret(deployment.SecretNamespace, secretName, true)
	if err != nil {
		return nil, err
	}

	log.Printf("Completed: DeleteSecret: %v", secretName)

	return TaskOutput{}, nil
}

func CommitDelete(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting CommitDeleteAirflow\n")
	deploymentInput, err := GetDeploymentDeleteInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	ctx.SqlModel.DeleteAirflow(context.Background(), deploymentInput.GetId())
	return nil, nil
}

func ResizeAirflowDb(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("%s: HelmResizeAirflowDb: started...", ctx.DeploymentContext.ID)

	deploymentInput, err := GetDeploymentResizeInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	data, _ := ctx.SqlModel.GetAirflowById(context.Background(), db.GetAirflowByIdParams{
		ID: deploymentInput.GetId(),
	})
	existingSize, _ := ctx.SqlModel.GetAirflowSizeByName(context.Background(), data.Size)
	newSize, _ := ctx.SqlModel.GetAirflowSizeByName(context.Background(), deploymentInput.GetSize())

	if deploymentInput.GetSize() == "" || existingSize.BackendDatabaseSizeID.String == newSize.BackendDatabaseSizeID.String {
		log.Printf("New and the existing size of the Airflow uses same backend database Size. So skipping this task.")
		return nil, nil
	}

	// Create Input payload for the Postgres Restart
	input := &pb.DpaiPostgresResizeRequest{
		Id: data.BackendDatabaseID,
		SizeProperties: &pb.DpaiPostgresSizeProperties{
			SizeId: newSize.BackendDatabaseSizeID.String,
		},
	}
	// convert input payload into the bytes
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	deploymentId := getAirflowDbDeploymentId(ctx.DeploymentContext.ID)
	_, err = ctx.SqlModel.CreateDeployment(context.Background(), db.CreateDeploymentParams{
		ID:                 deploymentId,
		WorkspaceID:        pgtype.Text{String: data.WorkspaceID, Valid: true},
		ServiceID:          pgtype.Text{String: data.BackendDatabaseID, Valid: true},
		ServiceType:        pb.DpaiServiceType_DPAI_POSTGRES.String(),
		ChangeIndicator:    pb.DpaiDeploymentChangeIndicator_DPAI_RESIZE.String(),
		CreatedBy:          "Internal Process",
		InputPayload:       jsonData,
		ParentDeploymentID: pgtype.Text{String: ctx.DeploymentContext.ID, Valid: true},
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	_, err = postgres.ResizeDeployment(deployment.DeploymentInputContext{
		ID:                 deploymentId,
		SqlPool:            ctx.DeploymentContext.Context.SqlPool,
		ParentDeploymentId: ctx.DeploymentContext.ID,
	})

	if err != nil {
		log.Printf("Failed to resize the Postgres for Airflow %s", data.BackendDatabaseID)
	}

	return TaskOutput{}, err
}

// Resize tasks

func getResizeParams(ctx *deployment.TaskRunContext) (db.ResizeAirflowParams, error) {

	deploymentInput, _ := GetDeploymentResizeInput(ctx.DeploymentContext)
	data, err := ctx.SqlModel.GetAirflowById(context.Background(), db.GetAirflowByIdParams{
		ID: deploymentInput.GetId(),
	})
	if err != nil {
		log.Printf("Not able to find/fetch an Airflow for the id: %s", deploymentInput.Id)
		return db.ResizeAirflowParams{}, err
	}

	params := db.ResizeAirflowParams{
		ID:            data.ID,
		Size:          data.Size,
		NumberOfNodes: data.NumberOfNodes,
	}
	if deploymentInput.Size != "" {
		params.Size = deploymentInput.Size
	}
	if deploymentInput.NumberOfNodes != 0 {
		params.NumberOfNodes = pgtype.Int4{Int32: deploymentInput.NumberOfNodes, Valid: true}
	}

	return params, nil
}

func generateHelmYamlValuesForResize(ctx *deployment.TaskRunContext, name string) (string, error) {
	params, err := getResizeParams(ctx)
	if err != nil {
		return "", err
	}

	size, err := ctx.SqlModel.GetAirflowSizeById(context.Background(), params.Size)
	if err != nil {
		return "", err
	}
	// Need to check all the required values
	yamlValues := map[string]interface{}{
		"replicaCount": params.NumberOfNodes.Int32,
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"memory": size.WebserverMemoryRequest.String,
				"cpu":    size.WebserverCpuRequest.String,
			},
			"limits": map[string]interface{}{
				"memory": size.WebserverMemoryLimit.String,
				"cpu":    size.WebserverCpuLimit.String,
			},
		},
	}

	yamlString, err := utils.ConvertToYAMLString(yamlValues)
	if err != nil {
		return "", err
	}
	return yamlString, nil

}

func HelmResize(ctx *deployment.TaskRunContext) (any, error) {

	deploymentInput, _ := GetDeploymentResizeInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetAirflowById(context.Background(), db.GetAirflowByIdParams{
		ID: deploymentInput.GetId(),
	})

	namespaceName := data.DeploymentID
	helmClient, chartReference, err := getHelmClient(ctx, namespaceName, data.Version)
	if err != nil {
		return nil, err
	}

	yamlValues, err := generateHelmYamlValuesForResize(ctx, data.Name)
	if err != nil {
		log.Printf("Error while generating YAML values from the user input for the postgres resize. %+v", err)
		return nil, err
	}

	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName: data.Name,
		ChartName:   chartReference.ChartName,
		Version:     chartReference.Version,
		Namespace:   namespaceName,
		UpgradeCRDs: true,
		ReuseValues: true,
		ValuesYaml:  yamlValues,
		Wait:        true,
		Timeout:     5 * time.Minute,
		// CreateNamespace: true,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	_, err = helmClient.UpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error while installing the upgrade. %+v", err)
		return nil, err
	}
	// output := TaskOutput{Release: rel}
	output := TaskOutput{}
	log.Printf("%s: HelmResizeAirflow: Completed. Output: %v", ctx.DeploymentContext.ID, output)

	return output, nil
}

func CommitResize(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting Airflow Size Upgrade\n")
	params, err := getResizeParams(ctx)
	if err != nil {
		return nil, err
	}

	ctx.SqlModel.ResizeAirflow(context.Background(), params)
	return nil, nil
}

//Restart

func RestartAirflowDb(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside RestartAirflowDb...")

	deploymentInput, err := GetDeploymentResizeInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	data, _ := ctx.SqlModel.GetAirflowById(context.Background(), db.GetAirflowByIdParams{
		ID: deploymentInput.GetId(),
	})

	// Create Input payload for the Postgres Restart
	input := &pb.DpaiPostgresRestartRequest{
		Id: data.BackendDatabaseID,
	}
	// convert input payload into the bytes
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	deploymentId := getAirflowDbDeploymentId(ctx.DeploymentContext.ID)
	_, err = ctx.SqlModel.CreateDeployment(context.Background(), db.CreateDeploymentParams{
		ID:                 deploymentId,
		WorkspaceID:        pgtype.Text{String: data.WorkspaceID, Valid: true},
		ServiceID:          pgtype.Text{String: data.BackendDatabaseID, Valid: true},
		ServiceType:        pb.DpaiServiceType_DPAI_POSTGRES.String(),
		ChangeIndicator:    pb.DpaiDeploymentChangeIndicator_DPAI_RESTART.String(),
		CreatedBy:          "Internal Process",
		InputPayload:       jsonData,
		ParentDeploymentID: pgtype.Text{String: ctx.DeploymentContext.ID, Valid: true},
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	_, err = postgres.RestartDeployment(deployment.DeploymentInputContext{
		ID:                 deploymentId,
		SqlPool:            ctx.DeploymentContext.Context.SqlPool,
		ParentDeploymentId: ctx.DeploymentContext.ID,
	})

	if err != nil {
		log.Printf("Failed to restart the Postgres for Airflow %s", data.BackendDatabaseID)
	}

	return TaskOutput{}, err
}

func generateHelmYamlValuesForRestart(ctx *deployment.TaskRunContext, postgresName string) (string, error) {
	yamlValues := map[string]interface{}{}

	yamlValues["podAnnotations"] = map[string]interface{}{
		"dapi.idcservice.net/restartedAt": time.Now().Format(time.RFC3339),
	}

	yamlString, _ := utils.ConvertToYAMLString(yamlValues)
	return yamlString, nil

}

func RestartAirflow(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Started: Inside Airflow Restart...")

	deploymentInput, err := GetDeploymentResizeInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	data, _ := ctx.SqlModel.GetAirflowById(context.Background(), db.GetAirflowByIdParams{
		ID: deploymentInput.GetId(),
	})

	namespaceName := data.DeploymentID
	helmClient, chartReference, err := getHelmClient(ctx, namespaceName, data.Version)
	if err != nil {
		return nil, err
	}

	yamlValues, err := generateHelmYamlValuesForRestart(ctx, data.Name)
	if err != nil {
		log.Printf("Error while generating YAML values from the user input for the postgres resize. %+v", err)
		return nil, err
	}

	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName: data.Name,
		ChartName:   chartReference.ChartName,
		Version:     chartReference.Version,
		Namespace:   namespaceName,
		UpgradeCRDs: true,
		ReuseValues: true,
		ValuesYaml:  yamlValues,
		Force:       true,
		Wait:        true,
		Timeout:     5 * time.Minute,
		// CreateNamespace: true,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	_, err = helmClient.UpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error while installing the upgrade. %+v", err)
		return nil, err
	}
	// output := TaskOutput{Release: rel}
	output := TaskOutput{}
	log.Printf("Completed: Inside Airflow Restart. Output: %v", output)

	return output, nil
}

func CommitRestart(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting Airflow Restart\n")
	deploymentInput, _ := GetDeploymentRestartInput(ctx.DeploymentContext)

	err := ctx.SqlModel.RestartAirflow(context.Background(), deploymentInput.Id)
	return TaskOutput{}, err
}

//Upgrade

func UpgradeAirflowDb(ctx *deployment.TaskRunContext) (any, error) {

	deploymentInput, err := GetDeploymentUpgradeInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	data, _ := ctx.SqlModel.GetAirflowById(context.Background(), db.GetAirflowByIdParams{ID: deploymentInput.GetId()})
	existingVersion, _ := ctx.SqlModel.GetAirflowVersionByName(context.Background(), data.Version)
	newVersion, _ := ctx.SqlModel.GetAirflowVersionByName(context.Background(), deploymentInput.Version)

	if existingVersion.PostgresVersion == newVersion.PostgresVersion {
		log.Printf("New and the existing version of the Airflow uses same backend Airflow Version. So skipping this task.")
		return nil, nil
	}

	// Create Input payload for the Postgres Restart
	input := &pb.DpaiPostgresUpgradeRequest{
		Id:        data.BackendDatabaseID,
		VersionId: newVersion.PostgresVersion,
	}
	// convert input payload into the bytes
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	deploymentId := getAirflowDbDeploymentId(ctx.DeploymentContext.ID)
	_, err = ctx.SqlModel.CreateDeployment(context.Background(), db.CreateDeploymentParams{
		ID:                 deploymentId,
		WorkspaceID:        pgtype.Text{String: data.WorkspaceID, Valid: true},
		ServiceID:          pgtype.Text{String: data.BackendDatabaseID, Valid: true},
		ServiceType:        pb.DpaiServiceType_DPAI_POSTGRES.String(),
		ChangeIndicator:    pb.DpaiDeploymentChangeIndicator_DPAI_UPGRADE.String(),
		CreatedBy:          "Internal Process",
		InputPayload:       jsonData,
		ParentDeploymentID: pgtype.Text{String: ctx.DeploymentContext.ID, Valid: true},
	})

	if err != nil {
		log.Printf("Error is : %+v", err)
		return nil, err
	}

	_, err = postgres.UpgradeDeployment(deployment.DeploymentInputContext{
		ID:                 deploymentId,
		SqlPool:            ctx.DeploymentContext.Context.SqlPool,
		ParentDeploymentId: ctx.DeploymentContext.ID,
	})

	if err != nil {
		log.Printf("Failed to upgrade the Postgres for Airflow %s", data.BackendDatabaseID)
	}

	return TaskOutput{}, err
}

func HelmUpgrade(ctx *deployment.TaskRunContext) (any, error) {

	deploymentInput, err := GetDeploymentUpgradeInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	data, _ := ctx.SqlModel.GetAirflowById(context.Background(), db.GetAirflowByIdParams{ID: deploymentInput.GetId()})

	namespaceName := data.DeploymentID
	helmClient, chartReference, err := getHelmClient(ctx, namespaceName, deploymentInput.Version)
	if err != nil {
		return nil, err
	}

	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName: data.Name,
		ChartName:   chartReference.ChartName,
		Version:     chartReference.Version,
		Namespace:   namespaceName,
		UpgradeCRDs: true,
		ReuseValues: true,
		Wait:        true,
		Timeout:     5 * time.Minute,
		// CreateNamespace: true,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	_, err = helmClient.UpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error while installing the upgrade. %+v", err)
		return nil, err
	}
	// output := TaskOutput{Release: rel}
	output := TaskOutput{}
	log.Printf("Completed: Inside HelmUpgrade. Output: %v", output)

	return output, nil
}

func CommitUpgrade(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting Version Upgrade\n")
	deploymentInput, err := GetDeploymentUpgradeInput(ctx.DeploymentContext)
	if err != nil {
		return nil, fmt.Errorf("unable to get the deploymentCreateInput. Error message: %+v", err)
	}

	ctx.SqlModel.UpgradeAirflow(context.Background(), db.UpgradeAirflowParams{
		ID:      deploymentInput.GetId(),
		Version: deploymentInput.GetVersion(),
	})
	return nil, nil
}
