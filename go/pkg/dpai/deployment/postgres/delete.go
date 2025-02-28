// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package postgres

import (
	// "context"
	"encoding/json"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func GetDeploymentDeleteInput(d deployment.Deployment) (*pb.DpaiPostgresDeleteRequest, error) {
	var value pb.DpaiPostgresDeleteRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func DeleteDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("DeletePostgres:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	task1 := deployment.NewTask("DeleteNamespace", DeleteNamespace, nil, []*deployment.Task{})
	task2 := deployment.NewTask("DeleteSecret", DeleteSecret, nil, []*deployment.Task{task1})
	task3 := deployment.NewTask("CommitDelete", CommitDelete, nil, []*deployment.Task{task2})

	deploy.AddTasks([]*deployment.Task{task1, task2, task3})

	deploy.Run()

	return deploy, nil
}
