// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package hms

import (
	// "context"
	"encoding/json"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func GetDeploymentResizeInput(d deployment.Deployment) (*pb.DpaiHmsResizeRequest, error) {
	var value pb.DpaiHmsResizeRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func ResizeDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("ResizeHms:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	taskResizeHmsDb := deployment.NewTask("ResizeHmsDb", ResizeHmsDb, nil, []*deployment.Task{})
	taskResizeHms := deployment.NewTask("Resize", HelmResize, nil, []*deployment.Task{taskResizeHmsDb})
	taskCommitResize := deployment.NewTask("CommitResize", CommitResize, nil, []*deployment.Task{taskResizeHms})

	deploy.AddTasks([]*deployment.Task{taskResizeHmsDb, taskResizeHms, taskCommitResize})

	deploy.Run()

	return deploy, nil
}
