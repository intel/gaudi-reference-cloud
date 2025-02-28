// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package airflow

import (
	// "context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func GetDeploymentCreateInput(d deployment.Deployment) (*pb.DpaiAirflowCreateRequest, error) {
	var value pb.DpaiAirflowCreateRequest
	log.Printf("GetDeploymentCreateInput: %+v ", d)
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// CreateDeployment executes all the tasks that are needed to create the workspace.
func CreateDeployment(ctx deployment.DeploymentInputContext) (string, error) {

	var resourceId string

	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("CreateAirflow:%s", ctx.ID))
	if err != nil {
		log.Printf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
		return resourceId, err
	}

	log.Printf("Deploy : %+v", deploy)

	//Create workspace
	createWorkspace := deployment.NewTask("CreateWorkspace", CreateWorkspace, nil, []*deployment.Task{})
	// Create and Add  Airflow Tasks
	createNodeGroup := deployment.NewTask("CreateNodeGroup", CreateNodeGroup, nil, []*deployment.Task{createWorkspace})
	createNamespace := deployment.NewTask("CreateNamespace", CreateNamespace, nil, []*deployment.Task{createWorkspace}) //This can be removed later as its not quired in PROD
	createSecret := deployment.NewTask("CreateSecret", CreateSecret, nil, []*deployment.Task{createNamespace})

	// Storage
	//prepareStorageObjects := deployment.NewTask("PrepareStorageObjects", PrepareStorageObjects, nil, []*deployment.Task{createNamespace})
	helmInstallAirflowDb := deployment.NewTask("HelmInstallAirflowDb", HelmInstallAirflowDb, nil, []*deployment.Task{createSecret, createNodeGroup})
	helmInstallAirflow := deployment.NewTask("HelmInstall", HelmInstall, nil, []*deployment.Task{helmInstallAirflowDb})
	updateAirflowUserPassword := deployment.NewTask("UpdateAirflowUserPassword", UpdateAirflowUserPassword, nil, []*deployment.Task{helmInstallAirflow})
	enableServiceMeshIngressNetwork := deployment.NewTask("EnableServiceMeshIngressNetwork", EnableServiceMeshIngressNetwork, nil, []*deployment.Task{helmInstallAirflow})
	validate := deployment.NewTask("ValidateAirflow", Validate, nil, []*deployment.Task{enableServiceMeshIngressNetwork, updateAirflowUserPassword})
	commitCreate := deployment.NewTask("CommitCreateAirflow", CommitCreate, nil, []*deployment.Task{validate})

	deploy.AddTasks([]*deployment.Task{
		createWorkspace,
		createNodeGroup, createNamespace, createSecret,
		//prepareStorageObjects,
		helmInstallAirflowDb, helmInstallAirflow, updateAirflowUserPassword,
		enableServiceMeshIngressNetwork, validate, commitCreate,
	})

	output, deploymentError := deploy.Run()

	var result TaskOutput
	err = json.Unmarshal(output, &result)
	if err != nil {
		log.Printf("Failed to convert the result to Output object. Error message: %s", err)
		log.Printf("Incase of failure in the deployment: %s, cleanup task can not be executed.", deploy.ID)
		return resourceId, err
	}

	if deploymentError != nil {
		log.Printf("Deployment failed with error: %+v", err)
		err = deploy.CleanUp(&deployment.DeploymentCleanUpParams{
			NodeGroupId: result.NodeGroupID,
			Namespace:   result.Namespace,
		})

		if err != nil {
			return resourceId, err
		} else {
			return resourceId, deploymentError
		}
	}

	resourceId = result.ID
	log.Printf("Successfully Created a new Airflow Instance with id %s for the deployment id %s.", resourceId, deploy.ID)
	return resourceId, err

}
