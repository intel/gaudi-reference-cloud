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

func GetDeploymentUpgradeInput(d deployment.Deployment) (*pb.DpaiHmsUpgradeRequest, error) {
	var value pb.DpaiHmsUpgradeRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func UpgradeDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("UpgradeHms:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	taskUpgradeHmsDb := deployment.NewTask("UpgradeHmsPostgres", UpgradeHmsDb, nil, []*deployment.Task{})
	taskUpgradeHms := deployment.NewTask("UpgradeHms", HelmUpgrade, nil, []*deployment.Task{taskUpgradeHmsDb})
	taskCommitUpgrade := deployment.NewTask("CommitUpgradeHms", CommitUpgrade, nil, []*deployment.Task{taskUpgradeHms})

	deploy.AddTasks([]*deployment.Task{taskUpgradeHmsDb, taskUpgradeHms, taskCommitUpgrade})

	deploy.Run()

	return deploy, nil
}
