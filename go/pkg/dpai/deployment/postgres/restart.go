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

func GetDeploymentRestartInput(d deployment.Deployment) (*pb.DpaiPostgresRestartRequest, error) {
	var value pb.DpaiPostgresRestartRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func RestartDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("RestartPostgres:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	task1 := deployment.NewTask("Restart", HelmRestart, nil, []*deployment.Task{})
	// task1 := deployment.NewTask("PgPoolRestart", PgPoolRestart, nil, []*deployment.Task{})
	task2 := deployment.NewTask("CommitRestart", CommitRestart, nil, []*deployment.Task{task1})

	deploy.AddTasks([]*deployment.Task{task1, task2})

	deploy.Run()

	return deploy, nil
}
