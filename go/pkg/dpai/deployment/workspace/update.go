// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package workspace

import (
	"encoding/json"
	"log"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type AddServiceDeploymentInput struct {
	WorkspaceId    string `json:"id,omitempty"`
	CloudAccountId string `json:"cloudAccountId,omitempty"`
}

// not used in release 1a
func GetDeploymentUpdateInput(d deployment.Deployment) (*pb.DpaiWorkspaceUpdateRequest, error) {
	var value pb.DpaiWorkspaceUpdateRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// use for current release due to removal of workspace concept with airflow service as single service of addition
func GetWorkspaceServiceAdditionInput(d deployment.Deployment) (*AddServiceDeploymentInput, error) {

	var addServiceInput AddServiceDeploymentInput
	err := json.Unmarshal(d.RawInput, &addServiceInput)
	if err != nil {
		return nil, err
	}
	return &addServiceInput, nil
}

// // UpdateOrchestrator executes all the activities that are needed to delete the workspace.
// // The activities executes in sequence and returns the result.
func UpdateDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	deploy, err := deployment.NewDeployment(ctx, "UpdateWorkspace")
	if err != nil {
		log.Fatalf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
		return nil, err
	}

	deploy.AddTasks([]*deployment.Task{})

	_, err = deploy.Run()
	if err != nil {
		return nil, err
	}

	return deploy, nil
}
