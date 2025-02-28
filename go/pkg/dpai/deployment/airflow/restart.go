// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package airflow

import (
	// "context"
	"encoding/json"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func GetDeploymentRestartInput(d deployment.Deployment) (*pb.DpaiAirflowRestartRequest, error) {
	var value pb.DpaiAirflowRestartRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func RestartDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("RestartAirflow:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	taskRestartDb := deployment.NewTask("RestartAirflowDb", RestartAirflowDb, nil, []*deployment.Task{})
	taskRestart := deployment.NewTask("RestartAirflow", RestartAirflow, nil, []*deployment.Task{taskRestartDb})
	taskCommitRestart := deployment.NewTask("CommitRestart", CommitRestart, nil, []*deployment.Task{taskRestart})

	deploy.AddTasks([]*deployment.Task{taskRestartDb, taskRestart, taskCommitRestart})

	deploy.Run()

	return deploy, nil
}
