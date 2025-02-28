// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package hms

import (
	// "context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	// "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment/postgres"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func GetDeploymentCreateInput(d deployment.Deployment) (*pb.DpaiHmsCreateRequest, error) {
	var value pb.DpaiHmsCreateRequest
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
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("CreateHms:%s", ctx.ID))
	if err != nil {
		log.Printf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
		return resourceId, err
	}

	// Create and Add Tasks
	createNodeGroup := deployment.NewTask("CreateNodeGroup", CreateNodeGroup, nil, []*deployment.Task{})
	createNamespace := deployment.NewTask("CreateNamespace", CreateNamespace, nil, []*deployment.Task{})
	createSecret := deployment.NewTask("CreateSecret", CreateSecret, nil, []*deployment.Task{createNamespace})
	helmInstallHmsDb := deployment.NewTask("HelmInstallHmsDb", HelmInstallHmsDb, nil, []*deployment.Task{createSecret, createNodeGroup})
	helmInstallHms := deployment.NewTask("HelmInstall", HelmInstall, nil, []*deployment.Task{helmInstallHmsDb})
	validate := deployment.NewTask("ValidateHms", Validate, nil, []*deployment.Task{helmInstallHms})
	commitCreate := deployment.NewTask("CommitCreateHms", CommitCreate, nil, []*deployment.Task{validate})

	deploy.AddTasks([]*deployment.Task{
		createNodeGroup, createNamespace, createSecret, helmInstallHmsDb,
		// helmInstallPostgres,
		helmInstallHms, validate, commitCreate,
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
			NodeGroupId: result.NodeGroupId,
			Namespace:   result.Namespace,
		})

		if err != nil {
			return resourceId, err
		} else {
			return resourceId, deploymentError
		}
	}

	resourceId = result.ID
	log.Printf("Created a new HMS Instance with id %s for the deployment id %s.", resourceId, deploy.ID)
	return resourceId, err

}
