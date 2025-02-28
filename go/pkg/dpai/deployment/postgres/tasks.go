// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package postgres

import (
	"context"
	// "encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"

	model "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/helm"
	helmutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/helm"

	// "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
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
	NodeGroupId string
	ReleaseName string
	SecretName  string
	Release     *release.Release
	// Optional fields
	Validated               bool
	NumberOfInstances       int32
	NumberOfPgPoolInstances int32
	DiskSizeInGb            int32
	ServiceName             string
	IksClusterUUID          string
	CloudAccountID          string
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

func getHelmChartReference(ctx *deployment.TaskRunContext, versionId string) (*helmutils.HelmChartReference, error) {

	version, err := ctx.SqlModel.GetPostgresVersionById(context.Background(), versionId)
	if err != nil {
		return nil, err
	}

	chartReference, err := utils.ConvertBytesToChartReference(version.ChartReference)
	if err != nil {
		return nil, err
	}

	return &helmutils.HelmChartReference{
		RepoName:  chartReference.GetRepoName(),
		RepoUrl:   chartReference.GetRepoUrl(),
		ChartName: chartReference.GetChartName(),
		Version:   chartReference.GetVersion(),
	}, nil
}

func getHelmClient(ctx *deployment.TaskRunContext, namespace string, versionId string) (helmclient.Client, *helmutils.HelmChartReference, error) {
	helmClient, err := helmutils.GetHelmClient(ctx.K8sClient.ClientConfig, namespace)
	if err != nil {
		return nil, nil, err
	}
	chartReference, err := getHelmChartReference(ctx, versionId)
	if err != nil {
		return nil, nil, err
	}

	//This is Not required
	chartRepo := repo.Entry{
		Name: chartReference.RepoName,
		URL:  chartReference.RepoUrl, // helm_config["url"].(string),
	}

	log.Printf("Chart Repo %+v.", chartRepo)

	// Add a chart-repository to the client.
	// if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
	// 	log.Printf("Error Adding Chart Repo %+v. Error message: %+v", chartRepo, err)
	// 	return nil, nil, err
	// }
	return helmClient, chartReference, nil
}

func generateHelmYamlValuesForCreate(ctx *deployment.TaskRunContext) (string, *pb.DpaiPostgresCreateRequest, error) {
	yamlValues := map[string]interface{}{}
	deploymentInput, _ := GetDeploymentCreateInput(ctx.DeploymentContext)
	output, _ := parseTaskOutput(ctx.GetTaskOutput("CreateSecret"))

	size, err := ctx.SqlModel.GetPostgresSizeById(context.Background(), deploymentInput.SizeProperties.GetSizeId())

	if err != nil {
		return "", nil, err
	}

	var extended_configuration string
	if len(deploymentInput.GetAdvanceConfiguration()) != 0 {
		for key, value := range deploymentInput.GetAdvanceConfiguration() {
			extended_configuration += "    " + key + " = " + value + "\n"
		}
	}

	if deploymentInput.SizeProperties.NumberOfInstances == 0 {
		deploymentInput.SizeProperties.NumberOfInstances = size.NumberOfInstancesDefault
	}
	if deploymentInput.SizeProperties.NumberOfPgPoolInstances == 0 {
		deploymentInput.SizeProperties.NumberOfPgPoolInstances = size.NumberOfPgpoolInstancesDefault
	}
	if deploymentInput.SizeProperties.DiskSizeInGb == 0 {
		deploymentInput.SizeProperties.DiskSizeInGb = size.DiskSizeInGbDefault
	}

	yamlValues["serviceAccount"] = map[string]interface{}{
		"create": true,
		"name":   fmt.Sprintf("sa-%s", output.ReleaseName),
	}

	yamlValues["service"] = map[string]interface{}{
		"type": "ClusterIP",
	}

	yamlValues["postgresql"] = map[string]interface{}{
		"database": deploymentInput.OptionalProperties.InitialDatabaseName,
		// TODO: Support custom admin username
		"existingSecret": output.ReleaseName,
		"replicaCount":   deploymentInput.SizeProperties.NumberOfInstances,
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
		// FIXME: Extended configuration breaks
		// "configuration": extended_configuration,
	}

	yamlValues["pgpool"] = map[string]interface{}{
		"replicaCount":   deploymentInput.SizeProperties.NumberOfPgPoolInstances,
		"existingSecret": output.ReleaseName,
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"memory": size.ResourcePgpoolMemoryRequest.String,
				"cpu":    size.ResourcePgpoolCpuRequest.String,
			},
			"limits": map[string]interface{}{
				"memory": size.ResourcePgpoolMemoryLimit.String,
				"cpu":    size.ResourcePgpoolCpuLimit.String,
			},
		},
	}
	log.Printf("Yaml Values: %+v", yamlValues)
	// TODO: Add toleration and nodeselector
	yamlString, _ := utils.ConvertToYAMLString(yamlValues)

	return yamlString, deploymentInput, nil

}

// Task Definitions - Create

func CreateNodeGroup(ctx *deployment.TaskRunContext) (any, error) {
	if ctx.DeploymentContext.Context.ParentDeploymentId == "" {
		log.Printf("Placeholder to create the nodegroup")
		nodeGroupId := "dummy-nodegroup-id"

		ctx.SqlModel.UpdateDeploymentNodeGroupId(context.TODO(), model.UpdateDeploymentNodeGroupIdParams{
			ID:          ctx.DeploymentContext.ID,
			NodeGroupID: pgtype.Text{String: nodeGroupId, Valid: true},
		})
	} else {
		log.Printf("Skipping CreateNodeGroup as its a subDeployment. Nodegroup from the parent deployment will be used.")
	}

	return nil, nil
}

func CreateNamespace(ctx *deployment.TaskRunContext) (any, error) {

	var namespaceName string

	deploymentInput, _ := GetDeploymentCreateInput(ctx.DeploymentContext)

	if ctx.DeploymentContext.Context.ParentDeploymentId == "" {
		labels := map[string]string{
			"istio-injection": "enabled",
		}
		namespace, err := ctx.K8sClient.CreateNamespace(ctx.DeploymentContext.ID, false, labels)
		if err != nil {
			return nil, err
		}
		namespaceName = namespace.Name
	} else {
		namespaceName = ctx.DeploymentContext.Context.ParentDeploymentId
		log.Printf("Skipping CreateNamespace as its a subDeployment. Namespace from the parent deployment will be used.")
	}

	output := TaskOutput{
		Namespace:   namespaceName,
		ReleaseName: deploymentInput.Name,
	}
	log.Printf("Completed: Inside CreateNamespace. Output: %+v", output)

	return output, nil
}

func CreateSecret(ctx *deployment.TaskRunContext) (any, error) {
	output, _ := parseTaskOutput(ctx.GetTaskOutput("CreateNamespace"))
	ctx.K8sClient.ClusterID = &pb.ClusterID{
		Clusteruuid:    output.IksClusterUUID,
		CloudAccountId: output.CloudAccountID,
	}
	err := ctx.K8sClient.GetIksClient(ctx.DeploymentContext.Context.Conf)
	if err != nil {
		return nil, err
	}

	secretData, err := ctx.K8sClient.GetSecret("secrets", ctx.DeploymentContext.ID)
	if err != nil {
		return nil, err
	}

	err = ctx.K8sClient.CreateSecret(output.Namespace, output.ReleaseName, secretData)
	if err != nil {
		return nil, err
	}
	output = mergeTaskOutput(output, TaskOutput{SecretName: output.ReleaseName})
	return output, nil
}

func HelmInstall(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside HelmInstallOrUpgrade: %v \n", ctx.GetTaskOutput("CreateNamespace"))

	output, _ := parseTaskOutput(ctx.GetTaskOutput("CreateSecret"))
	deploymentInput, _ := GetDeploymentCreateInput(ctx.DeploymentContext)

	helmClient, chartReference, err := getHelmClient(ctx, output.Namespace, deploymentInput.VersionId)
	if err != nil {
		return nil, err
	}
	log.Printf("chartReference.RepoUrl. %+v", chartReference.RepoUrl)
	log.Printf("chartReference.RepoUrl. %+v", chartReference.ChartName)
	log.Printf("chartReference.RepoUrl. %+v", chartReference.Version)

	//Adding to for OCI Support the following lines need to uncommented
	postgresfilepath, err := helm.DownloadChart("/tmp", chartReference.RepoUrl, chartReference.ChartName, chartReference.Version, nil, nil)
	if err != nil {
		return nil, err
	}

	yamlValues, deploymentInput, _ := generateHelmYamlValuesForCreate(ctx)
	// Define the chart to be installed
	chartSpec := helmclient.ChartSpec{
		ReleaseName: output.ReleaseName,
		ChartName:   postgresfilepath, //chartReference.ChartName,
		Namespace:   output.Namespace,
		UpgradeCRDs: true,
		ValuesYaml:  yamlValues,
		Wait:        true,
		Atomic:      true,
		Timeout:     5 * time.Minute,
		// CreateNamespace: true,
	}

	log.Printf("Postgres HelmInstall- helmclient: %+v ", helmClient)
	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	rel, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil)
	if err != nil {
		log.Printf("Error while installing the upgrade. %+v", err)
		return nil, err
	}
	output = mergeTaskOutput(output, TaskOutput{
		Release:                 rel,
		NumberOfInstances:       deploymentInput.SizeProperties.NumberOfInstances,
		NumberOfPgPoolInstances: deploymentInput.SizeProperties.NumberOfPgPoolInstances,
		DiskSizeInGb:            deploymentInput.SizeProperties.DiskSizeInGb,
	})
	log.Printf("Completed: Inside Postgres. Output: %v", output)

	_, err = ctx.SqlModel.CommitPostgresCreate(context.Background(), model.CommitPostgresCreateParams{
		ID:                      ctx.DeploymentContext.Context.ServiceId,
		NumberOfInstances:       pgtype.Int4{Int32: deploymentInput.SizeProperties.NumberOfInstances, Valid: true},
		NumberOfPgpoolInstances: pgtype.Int4{Int32: deploymentInput.SizeProperties.NumberOfPgPoolInstances, Valid: true},
		DiskSizeInGb:            pgtype.Int4{Int32: deploymentInput.SizeProperties.DiskSizeInGb, Valid: true},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to commit the postgres deployment in HelmInstall. Error: %+v", err)
	}

	return output, nil
}

func CommitCreate(ctx *deployment.TaskRunContext) (any, error) {

	log.Printf("Started: Inside CommitCreatePostgres\n")

	// deploymentInput, _ := GetDeploymentCreateInput(ctx.DeploymentContext)
	output, _ := parseTaskOutput(ctx.GetTaskOutput("HelmInstall"))

	// adminSecret, err := utils.ConvertSecretReferenceToBytes(&pb.DpaiSecretReference{
	// 	SecretName:    output.SecretName,
	// 	SecretKeyName: deploymentInput.GetAdminProperties().GetAdminPasswordSecretReference().GetSecretKeyName(),
	// })
	// if err != nil {
	// 	return nil, err
	// }
	result, err := ctx.SqlModel.CommitPostgresCreate(context.Background(), model.CommitPostgresCreateParams{
		ID:                          ctx.DeploymentContext.Context.ServiceId,
		ServerUrl:                   pgtype.Text{String: fmt.Sprintf("%s-postgresql-ha-pgpool.%s.svc.cluster.local", output.ReleaseName, output.Namespace), Valid: true},
		DeploymentStatusState:       pgtype.Text{String: pb.DpaiDeploymentState_DPAI_SUCCESS.String(), Valid: true},
		DeploymentStatusDisplayName: pgtype.Text{String: "Success", Valid: true},
		DeploymentStatusMessage:     pgtype.Text{String: "Successfully deployed the postgres instance.", Valid: true},
	})
	if err != nil {
		log.Printf("Not able to commit the status of the Postgres to the backend DB. Resource might have been created.")
		return nil, err
	}

	output.ID = result.ID
	return output, nil
}

func Validate(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside ValidatePostgres \n")
	time.Sleep(1 * time.Second)

	output, _ := parseTaskOutput(ctx.GetTaskOutput("CreateNamespace"))
	log.Printf("%T: %+v", output, output)

	output = mergeTaskOutput(output, TaskOutput{Validated: true})
	log.Printf("Completed: Inside ValidatePostgres Output: %v", output)

	return output, nil
}

// Delete tasks

func DeleteNamespace(ctx *deployment.TaskRunContext) (any, error) {
	deploymentInput, _ := GetDeploymentDeleteInput(ctx.DeploymentContext)

	data, err := ctx.SqlModel.GetPostgresById(context.Background(), deploymentInput.GetId())
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

	data, err := ctx.SqlModel.GetPostgresById(context.Background(), deploymentInput.GetId())
	if err != nil {
		return nil, err
	}

	secretName := data.DeploymentID
	err = ctx.K8sClient.DeleteSecret("secrets", secretName, true)
	if err != nil {
		return nil, err
	}

	log.Printf("Completed: DeleteSecret: %v", secretName)

	return TaskOutput{}, nil
}

func CommitDelete(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting CommitDeleteNamespace\n")
	deploymentInput, _ := GetDeploymentDeleteInput(ctx.DeploymentContext)

	ctx.SqlModel.DeletePostgres(context.Background(), deploymentInput.GetId())
	return nil, nil
}

// Upgrade tasks
func HelmUpgrade(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside HelmUpgrade...")

	deploymentInput, _ := GetDeploymentUpgradeInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetPostgresById(context.Background(), deploymentInput.GetId())

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
	log.Printf("Completed: Inside Helm. Output: %v", output)

	return output, nil
}

func CommitUpgrade(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting Version Upgrade\n")
	deploymentInput, _ := GetDeploymentUpgradeInput(ctx.DeploymentContext)

	ctx.SqlModel.UpgradePostgres(context.Background(), model.UpgradePostgresParams{
		ID:        deploymentInput.GetId(),
		VersionID: deploymentInput.GetVersionId(),
	})
	return nil, nil
}

// Upgrade tasks

func getResizeParams(ctx *deployment.TaskRunContext) (model.ResizePostgresParams, error) {

	deploymentInput, _ := GetDeploymentResizeInput(ctx.DeploymentContext)
	pg, err := ctx.SqlModel.GetPostgresById(context.Background(), deploymentInput.Id)
	if err != nil {
		log.Printf("Not able to find/fetch an postgres for the id: %s", deploymentInput.Id)
		return model.ResizePostgresParams{}, err
	}

	params := model.ResizePostgresParams{
		ID:                      pg.ID,
		SizeID:                  pg.SizeID,
		NumberOfInstances:       pg.NumberOfInstances,
		NumberOfPgpoolInstances: pg.NumberOfPgpoolInstances,
		DiskSizeInGb:            pg.DiskSizeInGb,
	}
	if deploymentInput.SizeProperties.SizeId != "" {
		params.SizeID = deploymentInput.SizeProperties.SizeId
	}
	if deploymentInput.SizeProperties.NumberOfInstances != 0 {
		params.NumberOfInstances = pgtype.Int4{Int32: deploymentInput.SizeProperties.NumberOfInstances, Valid: true}
	}
	if deploymentInput.SizeProperties.NumberOfPgPoolInstances != 0 {
		params.NumberOfPgpoolInstances = pgtype.Int4{Int32: deploymentInput.SizeProperties.NumberOfPgPoolInstances, Valid: true}
	}
	if deploymentInput.SizeProperties.DiskSizeInGb != 0 {
		params.DiskSizeInGb = pgtype.Int4{Int32: deploymentInput.SizeProperties.DiskSizeInGb, Valid: true}
	}

	return params, nil
}

func generateHelmYamlValuesForResize(ctx *deployment.TaskRunContext, postgresName string) (string, error) {
	yamlValues := map[string]interface{}{}
	params, err := getResizeParams(ctx)
	if err != nil {
		return "", err
	}

	size, err := ctx.SqlModel.GetPostgresSizeById(context.Background(), params.SizeID)
	if err != nil {
		return "", err
	}

	yamlValues["postgresql"] = map[string]interface{}{
		"existingSecret": postgresName,
		"replicaCount":   params.NumberOfInstances,
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

	yamlValues["pgpool"] = map[string]interface{}{
		"existingSecret": postgresName,
		"replicaCount":   params.NumberOfPgpoolInstances,

		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"memory": size.ResourcePgpoolMemoryRequest.String,
				"cpu":    size.ResourcePgpoolCpuRequest.String,
			},
			"limits": map[string]interface{}{
				"memory": size.ResourcePgpoolMemoryLimit.String,
				"cpu":    size.ResourcePgpoolCpuLimit.String,
			},
		},
	}

	yamlString, _ := utils.ConvertToYAMLString(yamlValues)
	return yamlString, nil

}

func HelmResize(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside HelmResize...")

	deploymentInput, _ := GetDeploymentResizeInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetPostgresById(context.Background(), deploymentInput.GetId())

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
	log.Printf("Completed: Inside CreateNamespace. Output: %v", output)

	return output, nil
}

func CommitResize(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting Version Upgrade\n")
	params, err := getResizeParams(ctx)
	if err != nil {
		return nil, err
	}

	ctx.SqlModel.ResizePostgres(context.Background(), params)
	return nil, nil
}

// restart

func generateHelmYamlValuesForRestart(ctx *deployment.TaskRunContext, postgresName string) (string, error) {
	yamlValues := map[string]interface{}{}

	yamlValues["postgresql"] = map[string]interface{}{
		"existingSecret": postgresName,
		"podAnnotations": map[string]interface{}{
			"dapi.idcservice.net/restartedAt": time.Now().Format(time.RFC3339),
		},
	}
	yamlValues["pgpool"] = map[string]interface{}{
		"existingSecret": postgresName,
		"podAnnotations": map[string]interface{}{
			"dapi.idcservice.net/restartedAt": time.Now().Format(time.RFC3339),
		},
	}

	yamlString, _ := utils.ConvertToYAMLString(yamlValues)
	return yamlString, nil

}

func HelmRestart(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside HelmRestart...")

	deploymentInput, _ := GetDeploymentResizeInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetPostgresById(context.Background(), deploymentInput.GetId())

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

func PgPoolRestart(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Inside PgPoolRestart...")

	deploymentInput, _ := GetDeploymentRestartInput(ctx.DeploymentContext)
	data, _ := ctx.SqlModel.GetPostgresById(context.Background(), deploymentInput.GetId())

	err := ctx.K8sClient.RestartAllDeployments(data.DeploymentID, 10*time.Minute)
	if err != nil {
		return nil, err
	}
	output := TaskOutput{}
	log.Printf("Completed: Inside CreateNamespace. Output: %v", output)

	return output, nil
}

func CommitRestart(ctx *deployment.TaskRunContext) (any, error) {
	log.Printf("Commiting Postgres Restart\n")
	deploymentInput, _ := GetDeploymentRestartInput(ctx.DeploymentContext)

	ctx.SqlModel.RestartPostgres(context.Background(), deploymentInput.Id)
	return nil, nil
}
