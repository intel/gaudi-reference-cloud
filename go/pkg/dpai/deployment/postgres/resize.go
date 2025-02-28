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

func GetDeploymentResizeInput(d deployment.Deployment) (*pb.DpaiPostgresResizeRequest, error) {
	var value pb.DpaiPostgresResizeRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func ResizeDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("ResizePostgres:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	task1 := deployment.NewTask("Resize", HelmResize, nil, []*deployment.Task{})
	task2 := deployment.NewTask("CommitResize", CommitResize, nil, []*deployment.Task{task1})

	deploy.AddTasks([]*deployment.Task{task1, task2})

	deploy.Run()

	return deploy, nil
}
