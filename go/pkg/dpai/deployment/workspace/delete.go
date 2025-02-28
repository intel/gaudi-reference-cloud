// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package workspace

import (
	// "context"
	"encoding/json"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func GetDeploymentDeleteInput(d deployment.Deployment) (*pb.DpaiWorkspaceDeleteRequest, error) {
	var value pb.DpaiWorkspaceDeleteRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func DeleteDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {

	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("DeleteWorkspace:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	taskDeleteService := deployment.NewTask("DeleteAllServices", DeleteAllServices, nil, []*deployment.Task{})
	taskDeleteIksCluster := deployment.NewTask("DeleteIKSCluster", DeleteIKSCluster, nil, []*deployment.Task{taskDeleteService})
	taskDeleteSshKey := deployment.NewTask("DeleteSshKey", DeleteSshKey, nil, []*deployment.Task{taskDeleteService})
	taskCommitDelete := deployment.NewTask("CommitDeleteWorkspace", CommitDeleteWorkspace, nil, []*deployment.Task{taskDeleteIksCluster, taskDeleteSshKey})

	deploy.AddTasks([]*deployment.Task{
		taskDeleteService,
		taskDeleteIksCluster, taskDeleteSshKey, taskCommitDelete})

	_, err = deploy.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	return deploy, nil
}
