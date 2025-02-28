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

func GetDeploymentResizeInput(d deployment.Deployment) (*pb.DpaiAirflowResizeRequest, error) {
	var value pb.DpaiAirflowResizeRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func ResizeDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("ResizeAirflow:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	taskResizeAirflowDb := deployment.NewTask("ResizeAirflowDb", ResizeAirflowDb, nil, []*deployment.Task{})
	taskResizeAirflow := deployment.NewTask("Resize", HelmResize, nil, []*deployment.Task{taskResizeAirflowDb})
	taskCommitResize := deployment.NewTask("CommitResize", CommitResize, nil, []*deployment.Task{taskResizeAirflow})

	deploy.AddTasks([]*deployment.Task{taskResizeAirflowDb, taskResizeAirflow, taskCommitResize})

	deploy.Run()

	return deploy, nil
}
