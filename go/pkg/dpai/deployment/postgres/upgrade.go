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

func GetDeploymentUpgradeInput(d deployment.Deployment) (*pb.DpaiPostgresUpgradeRequest, error) {
	var value pb.DpaiPostgresUpgradeRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func UpgradeDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("UpgradePostgres:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	task1 := deployment.NewTask("Upgrade", HelmUpgrade, nil, []*deployment.Task{})
	task2 := deployment.NewTask("CommitUpgrade", CommitUpgrade, nil, []*deployment.Task{task1})

	deploy.AddTasks([]*deployment.Task{task1, task2})

	deploy.Run()

	return deploy, nil
}
