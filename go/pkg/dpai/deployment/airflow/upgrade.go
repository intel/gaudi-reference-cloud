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

func GetDeploymentUpgradeInput(d deployment.Deployment) (*pb.DpaiAirflowUpgradeRequest, error) {
	var value pb.DpaiAirflowUpgradeRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// UpdateDeployment executes all the tasks that are needed to Upgrade the workspace.
func UpgradeDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("UpgradeAirflow:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	taskUpgradeAirflowDb := deployment.NewTask("UpgradeAirflowPostgres", UpgradeAirflowDb, nil, []*deployment.Task{})
	taskUpgradeAirflow := deployment.NewTask("UpgradeAirflow", HelmUpgrade, nil, []*deployment.Task{taskUpgradeAirflowDb})
	taskCommitUpgrade := deployment.NewTask("CommitUpgradeAirflow", CommitUpgrade, nil, []*deployment.Task{taskUpgradeAirflow})

	deploy.AddTasks([]*deployment.Task{taskUpgradeAirflowDb, taskUpgradeAirflow, taskCommitUpgrade})

	deploy.Run()

	return deploy, nil
}
