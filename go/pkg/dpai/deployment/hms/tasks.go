// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package hms

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"

	// "k8s.io/apimachinery/pkg/api/resource"

	model "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment/postgres"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/helm"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"

	// "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
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
	NodeGroupId        string
	ReleaseName        string
	SecretName         string
	RegistrySecretName string
	Release            *release.Release

	// Optional fields
	Hms                     model.Hms
	PostgresId              string
	Postgres                model.Postgres
	NumberOfInstances       int32
	NumberOfPgPoolInstances int32
	DiskSizeInGb            int32
	ServiceName             string
	Validated               bool
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

func getHelmChartReference(ctx *deployment.TaskRunContext, versionId string) (*helm.HelmChartReference, error) {

	version, err := ctx.SqlModel.GetHmsVersionById(context.Background(), versionId)
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

	version, err := ctx.SqlModel.GetHmsVersionById(context.Background(), versionId)
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
	config, err := utils.ReadConfig()
	if err != nil {
		return nil, nil, err
	}
	registry_config := config["registry"].(map[string]interface{})
	helm_config := registry_config["helm"].(map[string]interface{})

	chartRepo := repo.Entry{
		Name:               chartReference.RepoName,
		URL:                chartReference.RepoUrl, // helm_config["url"].(string),
		Username:           helm_config["username"].(string),
		Password:           os.Getenv(helm_config["passwordEnvKey"].(string)),
		PassCredentialsAll: true,
	}

	// Add a chart-repository to the client.
	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		if err != nil {
			log.Printf("Error Adding Chart Repo %+v. Error message: %+v", chartRepo, err)
			return nil, nil, err
		}
	}
	return helmClient, chartReference, nil
}

func generateHelmYamlValuesForCreate(ctx *deployment.TaskRunContext) (string, *pb.DpaiHmsCreateRequest, error) {

	var yamlString string
	yamlValues := map[string]interface{}{}

	deploymentInput, _ := GetDeploymentCreateInput(ctx.DeploymentContext)
	output, _ := parseTaskOutput(ctx.GetTaskOutput("HelmInstallHmsDb"))

	size, err := ctx.SqlModel.GetHmsSizeById(context.Background(), deploymentInput.SizeProperties.GetSizeId())

	if err != nil {
		fmt.Printf("Get HMS Size: %+v", err)
		return yamlString, nil, err
	}

	if deploymentInput.SizeProperties.NumberOfInstances == 0 {
		deploymentInput.SizeProperties.NumberOfInstances = size.NumberOfInstancesDefault
	}
	psqlDb, err := ctx.SqlModel.GetPostgresById(context.Background(), output.PostgresId)

	if err != nil {
		fmt.Printf("Get Postgres: %+v", err)
		return yamlString, nil, err
	}

	imageReference, _ := getImageReference(ctx, deploymentInput.VersionId)

	yamlValues["image"] = map[string]interface{}{
		"repository": imageReference.Repository,
		"pullPolicy": "IfNotPresent",
		"tag":        imageReference.Tag,
	}
	imagePullSecrets := make([]map[string]interface{}, 1)
	imagePullSecrets[0] = make(map[string]interface{})
	imagePullSecrets[0]["name"] = output.RegistrySecretName
	yamlValues["imagePullSecrets"] = imagePullSecrets

	yamlValues["serviceAccount"] = map[string]interface{}{
		"create": true,
		"name":   fmt.Sprintf("sa-%s", output.ReleaseName),
	}

	yamlValues["service"] = map[string]interface{}{
		"type": "ClusterIP",
		"port": 9083,
	}

	// TODO: Test without PVC. Validate whether this PVC is mandatory or not.
	// yamlValues["pvc"] = map[string]interface{}{
	// 	"warehouse": map[string]interface{}{
	// 		"size": "1Gi",
	// 	},
	// }

	psqlSecretReference, err := utils.ConvertBytesToSecretReference(psqlDb.AdminPasswordSecretReference)
	if err != nil {
		return "", nil, err
	}
	psqlSecret, _ := ctx.K8sClient.GetSecret(output.Namespace, psqlSecretReference.GetSecretName())

	yamlValues["backendDatabase"] = map[string]interface{}{
		"driver":               "postgres",
		"connectionDriverName": "org.postgresql.Driver",
		"connectionURL":        fmt.Sprintf("jdbc:postgresql://%s:5432/%s", psqlDb.ServerUrl.String, psqlDb.InitialDatabaseName.String),
		"connectionUserName":   "postgres", // psqlDb.AdminUsername,
		"connectionPassword":   string(psqlSecret[psqlSecretReference.GetSecretKeyName()]),
	}

	hmsSecret, _ := ctx.K8sClient.GetSecret(output.Namespace, output.SecretName)
	yamlValues["config"] = map[string]interface{}{
		"hiveSite": map[string]interface{}{
			"metastore.warehouse.dir": deploymentInput.ObjectStoreProperties.WarehouseDirectory,
			// "s3.endpoint":                     deploymentInput.ObjectStoreProperties.StorageEndpoint,
			"fs.s3a.access.key":               string(hmsSecret[deploymentInput.ObjectStoreProperties.GetStorageAccessKeySecretReference().GetSecretKeyName()]),
			"fs.s3a.secret.key":               string(hmsSecret[deploymentInput.ObjectStoreProperties.GetStorageAccessSecretSecretReference().GetSecretKeyName()]),
			"fs.s3a.impl":                     "org.apache.hadoop.fs.s3a.S3AFileSystem",
			"fs.s3a.aws.credentials.provider": "org.apache.hadoop.fs.s3a.SimpleAWSCredentialsProvider",
			"fs.s3a.fast.upload":              "true",
		},
	}
	yamlValues["resources"] = map[string]interface{}{
		"requests": map[string]interface{}{
			"memory": size.ResourceMemoryRequest.String,
			"cpu":    size.ResourceCpuRequest.String,
		},
		"limits": map[string]interface{}{
			"memory": size.ResourceMemoryLimit.String,
			"cpu":    size.ResourceCpuLimit.String,
		},
	}

	log.Printf("Yaml Values: %+v", yamlValues)
	// TODO: Add toleration and nodeselector
	yamlString, _ = utils.ConvertToYAMLString(yamlValues)

	return yamlString, deploymentInput, nil

}

func getHmsDbDeploymentId(parentDeploymentId string) string {
	if parentDeploymentId == "" {
		parentDeploymentId = "-"
	}
	return fmt.Sprintf("hmsdb-%s", strings.SplitN(parentDeploymentId, "-", 2)[1])
}

// Task Definitions - Create

func CreateNodeGroup(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Placeholder to create the nodegroup")

	return TaskOutput{NodeGroupId: "dummyNodeGroupId"}, nil
}

func CreateNamespace(ctx *deployment.TaskRunContext) (any, error) {

	deploymentInput, _ := GetDeploymentCreateInput(ctx.DeploymentContext)
	namespace, err := ctx.K8sClient.CreateNamespace(ctx.DeploymentContext.ID, false, map[string]string{})
	// namespace, err := k8s.CreateNamespace(ctx.DeploymentContext.Context.KubeClient, namespaceName, false)
	if err != nil {
		return nil, err
	}

	output := TaskOutput{
		Namespace:   namespace.Name,
		ReleaseName: deploymentInput.Name,
	}
	log.Printf("Completed: Inside CreateNamespace. Output: %v", output)

	return output, nil
}

func CreateSecret(ctx *deployment.TaskRunContext) (any, error) {
	output, _ := parseTaskOutput(ctx.GetTaskOutput("CreateNamespace"))

	secretData, err := ctx.K8sClient.GetSecret(deployment.SecretNamespace, output.Namespace)
	if err != nil {
		return nil, err
	}

	err = ctx.K8sClient.CreateSecret(output.Namespace, output.ReleaseName, secretData)
	if err != nil {
		return nil, err
	}

	// Create private docker registry secrets
	registrySecretName := deployment.DockerRegistrySecret
	config, err := utils.ReadConfig()
	if err != nil {
		return nil, err
	}
	registry_config := config["registry"].(map[string]interface{})
	docker_config := registry_config["docker"].(map[string]interface{})

	err = ctx.K8sClient.CreateSecretDockerConfigJson(output.Namespace, registrySecretName, k8s.DockerConfigSecretData{
		UserName: docker_config["username"].(string),
		Password: os.Getenv(docker_config["passwordEnvKey"].(string)),
		Server:   docker_config["url"].(string),
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

func HelmInstallHmsDb(ctx *deployment.TaskRunContext) (any, error) {

	secretOutput, _ := parseTaskOutput(ctx.GetTaskOutput("CreateSecret"))
	nodeGroupOutput, _ := parseTaskOutput(ctx.GetTaskOutput("CreateNodeGroup"))
	output := mergeTaskOutput(secretOutput, nodeGroupOutput)

	deploymentInput, _ := GetDeploymentCreateInput(ctx.DeploymentContext)
	// generate password
	deploymentId := getHmsDbDeploymentId(ctx.DeploymentContext.ID)

	version, err := ctx.SqlModel.GetHmsVersionById(context.TODO(), deploymentInput.VersionId)
	if err != nil {
		return nil, err
	}
	size, err := ctx.SqlModel.GetHmsSizeById(context.TODO(), deploymentInput.SizeProperties.SizeId)
	if err != nil {
		return nil, err
	}
	// generate secret
	password := "DummyPass"
	secretData := map[string][]byte{
		"password":        []byte(password),
		"repmgr-password": []byte(password),
		"admin-password":  []byte(password),
	}
	err = ctx.K8sClient.CreateSecret(deployment.SecretNamespace, deploymentId, secretData)
	if err != nil {
		return nil, fmt.Errorf("failed to create the secret")
	}

	input := &pb.DpaiPostgresCreateRequest{
		WorkspaceId: deploymentInput.WorkspaceId,
		Name:        fmt.Sprintf("hmsdb-%s", output.ReleaseName),
		Description: fmt.Sprintf("Backend database for the HMS %s", output.ReleaseName),
		VersionId:   version.BackendDatabaseVersionID,
		SizeProperties: &pb.DpaiPostgresSizeProperties{
			SizeId: size.BackendDatabaseSizeID.String,
		},
		AdminProperties: &pb.DpaiPostgresAdminProperties{
			AdminUsername: "hms-admin",
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
			InitialDatabaseName: "hms",
		},
	}

	// convert input payload into the bytes
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	_, err = ctx.SqlModel.CreateDeployment(context.Background(), db.CreateDeploymentParams{
		ID:                 deploymentId,
		WorkspaceID:        pgtype.Text{String: deploymentInput.WorkspaceId, Valid: true},
		ServiceType:        pb.DpaiServiceType_DPAI_POSTGRES.String(),
		ChangeIndicator:    pb.DpaiDeploymentChangeIndicator_DPAI_CREATE.String(),
		CreatedBy:          "Internal Process",
		InputPayload:       jsonData,
		ParentDeploymentID: pgtype.Text{String: ctx.DeploymentContext.ID, Valid: true},
	})

	if err != nil {
		log.Printf("Error is : %v", err)
		return nil, err
	}

	resourceId, err := postgres.CreateDeployment(deployment.DeploymentInputContext{
		ID:                 getHmsDbDeploymentId(ctx.DeploymentContext.ID), //fmt.Sprintf("hmsdb-%s", strings.SplitN(ctx.ID, "-", 2)[1]),
		SqlPool:            ctx.DeploymentContext.Context.SqlPool,
		ParentDeploymentId: ctx.DeploymentContext.ID,
	})

	if err != nil {
		return nil, err
	}

	output = mergeTaskOutput(output, TaskOutput{PostgresId: resourceId})

	return output, nil
}

func HelmInstall(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside HelmInstallOrUpgrade: %v \n", ctx.GetTaskOutput("HelmInstallHmsDb"))

	output, _ := parseTaskOutput(ctx.GetTaskOutput("HelmInstallHmsDb"))
	deploymentInput, _ := GetDeploymentCreateInput(ctx.DeploymentContext)

	helmClient, chartReference, err := getHelmClient(ctx, output.Namespace, deploymentInput.VersionId)
	if err != nil {
		return nil, err
	}

	yamlValues, deploymentInput, err := generateHelmYamlValuesForCreate(ctx)
	if err != nil {
		return nil, err
	}
	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName: output.ReleaseName,
		ChartName:   chartReference.ChartName, // fmt.Sprintf("%s/hms", helm_config["reponame"].(string)), //
		Version:     chartReference.Version,   // "0.1.0",
		Namespace:   output.Namespace,
		UpgradeCRDs: true,
		ValuesYaml:  yamlValues,
		Wait:        true,
		Atomic:      true,
		Timeout:     15 * time.Minute,
		// CreateNamespace: true,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	rel, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error while installing the service. %+v", err)
		return nil, err
	}
	output = mergeTaskOutput(output, TaskOutput{
		Release:           rel,
		NumberOfInstances: deploymentInput.GetSizeProperties().GetNumberOfInstances(),
	})
	log.Printf("Completed: Inside HelmInstall. Output: %v", output)

	return output, nil
}

func CommitCreate(ctx *deployment.TaskRunContext) (any, error) {

	log.Printf("Completed: Inside CommitCreateHmse\n")

	deploymentInput, _ := GetDeploymentCreateInput(ctx.DeploymentContext)
	output, _ := parseTaskOutput(ctx.GetTaskOutput("HelmInstall"))

	tags, err := json.Marshal(deploymentInput.GetTags())
	if err != nil {
		return nil, err
	}
	// advanceConfigutations, err := json.Marshal(deploymentInput.GetAdvanceConfiguration())
	// if err != nil {
	// 	return nil, err
	// }

	secretKey, err := utils.ConvertSecretReferenceToBytes(deploymentInput.ObjectStoreProperties.StorageAccessKeySecretReference)
	if err != nil {
		return nil, err
	}
	secretVal, err := utils.ConvertSecretReferenceToBytes(deploymentInput.ObjectStoreProperties.StorageAccessSecretSecretReference)
	if err != nil {
		return nil, err
	}
	data, err := ctx.SqlModel.CreateHms(context.Background(), model.CreateHmsParams{
		ID:                            uuid.New().String(),
		WorkspaceID:                   deploymentInput.GetWorkspaceId(),
		Name:                          deploymentInput.GetName(),
		VersionID:                     deploymentInput.GetVersionId(),
		SizeID:                        deploymentInput.SizeProperties.GetSizeId(),
		NumberOfInstances:             pgtype.Int4{Int32: output.NumberOfInstances, Valid: true},
		Description:                   pgtype.Text{String: deploymentInput.GetDescription(), Valid: true},
		Tags:                          tags,
		Endpoint:                      fmt.Sprintf("%s.%s.svc.cluster.local:9083", output.ReleaseName, output.Namespace),
		ObjectStoreStorageEndpoint:    pgtype.Text{String: deploymentInput.ObjectStoreProperties.StorageEndpoint, Valid: true},
		ObjectStoreWarehouseDirectory: deploymentInput.ObjectStoreProperties.WarehouseDirectory,
		ObjectStoreStorageAccessKeySecretReference:    secretKey,
		ObjectStoreStorageAccessSecretSecretReference: secretVal,

		DeploymentID:                ctx.DeploymentContext.ID,
		BackendDatabaseID:           output.PostgresId,
		NodeGroupID:                 output.NodeGroupId,
		DeploymentStatusState:       pb.DpaiDeploymentState_DPAI_SUCCESS.String(),
		DeploymentStatusDisplayName: pgtype.Text{String: "Success", Valid: true},
		DeploymentStatusMessage:     pgtype.Text{String: "Successfully deployed the HMS instance.", Valid: true},
		CreatedBy:                   "CallerOfTheAPI",
	})
	if err != nil {
		log.Printf("Not able to commit the status of the HMS to the backend DB for the deployment id %s. Resource might have been created.", ctx.DeploymentContext.ID)
		return nil, err
	}

	output.ID = data.ID
	return output, nil
}

func Validate(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside ValidateHms \n")
	time.Sleep(10 * time.Second)

	output, _ := parseTaskOutput(ctx.GetTaskOutput("CreateNamespace"))
	log.Printf("%T: %+v", output, output)

	output = mergeTaskOutput(output, TaskOutput{Validated: true})
	log.Printf("Completed: Inside ValidatePostgres Output: %v", output)

	return output, nil
}

// Delete tasks

func DeleteNamespace(ctx *deployment.TaskRunContext) (any, error) {
	deploymentInput, _ := GetDeploymentDeleteInput(ctx.DeploymentContext)

	data, err := ctx.SqlModel.GetHmsById(context.Background(), deploymentInput.GetId())
	if err != nil {
		return nil, err
	}

	namespaceName := data.DeploymentID
	_, err = ctx.K8sClient.DeleteNamespace(namespaceName, true)
	if err != nil {
		return nil, err
	}

	log.Printf("Completed: Inside DeleteNamespace: %v", namespaceName)

	return TaskOutput{}, nil
}

func DeleteSecret(ctx *deployment.TaskRunContext) (any, error) {
	deploymentInput, _ := GetDeploymentDeleteInput(ctx.DeploymentContext)

	data, err := ctx.SqlModel.GetHmsById(context.Background(), deploymentInput.GetId())
	if err != nil {
		return nil, err
	}

	secretName := data.DeploymentID
	err = ctx.K8sClient.DeleteSecret(deployment.SecretNamespace, secretName, true)
	if err != nil {
		return nil, err
	}

	log.Printf("Completed: DeleteSecret: %v", secretName)

	return TaskOutput{}, nil
}

func CommitDelete(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting CommitDeleteHms\n")
	deploymentInput, _ := GetDeploymentDeleteInput(ctx.DeploymentContext)

	ctx.SqlModel.DeleteHms(context.Background(), deploymentInput.GetId())
	return nil, nil
}

// Upgrade tasks
func HelmUpgrade(ctx *deployment.TaskRunContext) (any, error) {

	deploymentInput, _ := GetDeploymentUpgradeInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetHmsById(context.Background(), deploymentInput.GetId())

	namespaceName := data.DeploymentID
	helmClient, chartReference, err := getHelmClient(ctx, namespaceName, deploymentInput.VersionId)
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
	rel, err := helmClient.UpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error while installing the upgrade. %+v", err)
		return nil, err
	}
	output := TaskOutput{Release: rel}
	log.Printf("Completed: Inside HelmUpgrade. Output: %v", output)

	return output, nil
}

func UpgradeHmsDb(ctx *deployment.TaskRunContext) (any, error) {

	deploymentInput, _ := GetDeploymentUpgradeInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetHmsById(context.Background(), deploymentInput.GetId())
	existingVersion, _ := ctx.SqlModel.GetHmsVersionById(context.Background(), data.VersionID)
	newVersion, _ := ctx.SqlModel.GetHmsVersionById(context.Background(), deploymentInput.VersionId)

	if existingVersion.BackendDatabaseVersionID == newVersion.BackendDatabaseVersionID {
		log.Printf("New and the existing version of the HMS uses same backend database Version. So skipping this task.")
		return nil, nil
	}

	// Create Input payload for the Postgres Restart
	input := &pb.DpaiPostgresUpgradeRequest{
		Id:        data.BackendDatabaseID,
		VersionId: newVersion.BackendDatabaseVersionID,
	}
	// convert input payload into the bytes
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	deploymentId := getHmsDbDeploymentId(ctx.DeploymentContext.ID)
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
		log.Printf("Error is : %v", err)
		return nil, err
	}

	_, err = postgres.UpgradeDeployment(deployment.DeploymentInputContext{
		ID:                 deploymentId,
		SqlPool:            ctx.DeploymentContext.Context.SqlPool,
		ParentDeploymentId: ctx.DeploymentContext.ID,
	})

	if err != nil {
		log.Printf("Failed to upgrade the Postgres for HMS %s", data.BackendDatabaseID)
	}

	return TaskOutput{}, err
}

func CommitUpgrade(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting Version Upgrade\n")
	deploymentInput, _ := GetDeploymentUpgradeInput(ctx.DeploymentContext)

	ctx.SqlModel.UpgradeHms(context.Background(), model.UpgradeHmsParams{
		ID:        deploymentInput.GetId(),
		VersionID: deploymentInput.GetVersionId(),
	})
	return nil, nil
}

// Resize tasks

func getResizeParams(ctx *deployment.TaskRunContext) (model.ResizeHmsParams, error) {

	deploymentInput, _ := GetDeploymentResizeInput(ctx.DeploymentContext)
	data, err := ctx.SqlModel.GetHmsById(context.Background(), deploymentInput.Id)
	if err != nil {
		log.Printf("Not able to find/fetch an HMS for the id: %s", deploymentInput.Id)
		return model.ResizeHmsParams{}, err
	}

	params := model.ResizeHmsParams{
		ID:                data.ID,
		SizeID:            data.SizeID,
		NumberOfInstances: data.NumberOfInstances,
	}
	if deploymentInput.SizeId != "" {
		params.SizeID = deploymentInput.SizeId
	}
	if deploymentInput.NumberOfInstances != 0 {
		params.NumberOfInstances = pgtype.Int4{Int32: deploymentInput.NumberOfInstances, Valid: true}
	}

	return params, nil
}

func generateHelmYamlValuesForResize(ctx *deployment.TaskRunContext, name string) (string, error) {
	params, err := getResizeParams(ctx)
	if err != nil {
		return "", err
	}

	size, err := ctx.SqlModel.GetHmsSizeById(context.Background(), params.SizeID)
	if err != nil {
		return "", err
	}

	yamlValues := map[string]interface{}{
		"replicaCount": params.NumberOfInstances.Int32,
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"memory": size.ResourceMemoryRequest.String,
				"cpu":    size.ResourceCpuRequest.String,
			},
			"limits": map[string]interface{}{
				"memory": size.ResourceMemoryLimit.String,
				"cpu":    size.ResourceCpuLimit.String,
			},
		},
	}

	yamlString, _ := utils.ConvertToYAMLString(yamlValues)
	return yamlString, nil

}

func HelmResize(ctx *deployment.TaskRunContext) (any, error) {

	deploymentInput, _ := GetDeploymentResizeInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetHmsById(context.Background(), deploymentInput.GetId())

	namespaceName := data.DeploymentID
	helmClient, chartReference, err := getHelmClient(ctx, namespaceName, data.VersionID)
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
	rel, err := helmClient.UpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error while installing the upgrade. %+v", err)
		return nil, err
	}
	output := TaskOutput{Release: rel}
	log.Printf("%s: HelmResizeHms: Completed. Output: %v", ctx.DeploymentContext.ID, output)

	return output, nil
}

func ResizeHmsDb(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("%s: HelmResizeHmsDb: started...", ctx.DeploymentContext.ID)

	deploymentInput, _ := GetDeploymentResizeInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetHmsById(context.Background(), deploymentInput.GetId())
	existingSize, _ := ctx.SqlModel.GetHmsSizeById(context.Background(), data.SizeID)
	newSize, _ := ctx.SqlModel.GetHmsSizeById(context.Background(), deploymentInput.GetSizeId())

	if deploymentInput.GetSizeId() == "" || existingSize.BackendDatabaseSizeID.String == newSize.BackendDatabaseSizeID.String {
		log.Printf("New and the existing size of the HMS uses same backend database Size. So skipping this task.")
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

	deploymentId := getHmsDbDeploymentId(ctx.DeploymentContext.ID)
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
		log.Printf("Error is : %v", err)
		return nil, err
	}

	_, err = postgres.ResizeDeployment(deployment.DeploymentInputContext{
		ID:                 deploymentId,
		SqlPool:            ctx.DeploymentContext.Context.SqlPool,
		ParentDeploymentId: ctx.DeploymentContext.ID,
	})

	if err != nil {
		log.Printf("Failed to resize the Postgres for HMS %s", data.BackendDatabaseID)
	}

	return TaskOutput{}, err
}

func CommitResize(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting Hms Size Upgrade\n")
	params, err := getResizeParams(ctx)
	if err != nil {
		return nil, err
	}

	ctx.SqlModel.ResizeHms(context.Background(), params)
	return nil, nil
}

// restart

func generateHelmYamlValuesForRestart(ctx *deployment.TaskRunContext, postgresName string) (string, error) {
	yamlValues := map[string]interface{}{}

	yamlValues["podAnnotations"] = map[string]interface{}{
		"dapi.idcservice.net/restartedAt": time.Now().Format(time.RFC3339),
	}

	yamlString, _ := utils.ConvertToYAMLString(yamlValues)
	return yamlString, nil

}

func RestartHms(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside Restart...")

	deploymentInput, _ := GetDeploymentResizeInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetHmsById(context.Background(), deploymentInput.GetId())

	namespaceName := data.DeploymentID
	helmClient, chartReference, err := getHelmClient(ctx, namespaceName, data.VersionID)
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
	rel, err := helmClient.UpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error while installing the upgrade. %+v", err)
		return nil, err
	}
	output := TaskOutput{Release: rel}
	log.Printf("Completed: Inside HelmRestart. Output: %v", output)

	return output, nil
}

func RestartHmsDb(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside RestartHmsDb...")

	deploymentInput, _ := GetDeploymentResizeInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetHmsById(context.Background(), deploymentInput.GetId())

	// Create Input payload for the Postgres Restart
	input := &pb.DpaiPostgresRestartRequest{
		Id: data.BackendDatabaseID,
	}
	// convert input payload into the bytes
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	deploymentId := getHmsDbDeploymentId(ctx.DeploymentContext.ID)
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
		log.Printf("Error is : %v", err)
		return nil, err
	}

	_, err = postgres.RestartDeployment(deployment.DeploymentInputContext{
		ID:                 deploymentId,
		SqlPool:            ctx.DeploymentContext.Context.SqlPool,
		ParentDeploymentId: ctx.DeploymentContext.ID,
	})

	if err != nil {
		log.Printf("Failed to restart the Postgres for HMS %s", data.BackendDatabaseID)
	}

	return TaskOutput{}, err
}

func CommitRestart(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting Hms Restart\n")
	deploymentInput, _ := GetDeploymentRestartInput(ctx.DeploymentContext)

	err := ctx.SqlModel.RestartHms(context.Background(), deploymentInput.Id)
	return TaskOutput{}, err
}
