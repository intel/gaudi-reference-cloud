// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package postgres

import (
	// "context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func GetDeploymentCreateInput(d deployment.Deployment) (*pb.DpaiPostgresCreateRequest, error) {
	var value pb.DpaiPostgresCreateRequest
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
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("CreatePostgres:%s", ctx.ID))
	if err != nil {
		log.Printf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
		return resourceId, err
	}

	// Create and Add Tasks
	createNodeGroup := deployment.NewTask("CreateNodeGroup", CreateNodeGroup, nil, []*deployment.Task{})
	createNamespace := deployment.NewTask("CreateNamespace", CreateNamespace, nil, []*deployment.Task{})
	CreateSecret := deployment.NewTask("CreateSecret", CreateSecret, nil, []*deployment.Task{createNamespace})
	helmInstall := deployment.NewTask("HelmInstall", HelmInstall, nil, []*deployment.Task{CreateSecret, createNodeGroup})
	validate := deployment.NewTask("Validate", Validate, nil, []*deployment.Task{helmInstall})
	commitCreate := deployment.NewTask("CommitCreatePostgres", CommitCreate, nil, []*deployment.Task{validate})

	deploy.AddTasks([]*deployment.Task{
		createNodeGroup, createNamespace, CreateSecret,
		helmInstall, validate, commitCreate,
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
		// make sure the nodegroup is deleted.
		log.Printf("Deployment failed with error: %+v", deploymentError)
		err = deploy.CleanUp(&deployment.DeploymentCleanUpParams{
			NodeGroupId: result.NodeGroupId,
			Namespace:   result.Namespace,
		})

		if err != nil {
			return resourceId, err
		}
		return resourceId, deploymentError
	}

	resourceId = result.ID
	log.Printf("Created a new Postgres Instance with id %s for the deployment id %s.", resourceId, deploy.ID)
	return resourceId, err
}
