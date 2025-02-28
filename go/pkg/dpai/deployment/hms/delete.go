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

func GetDeploymentDeleteInput(d deployment.Deployment) (*pb.DpaiHmsDeleteRequest, error) {
	var value pb.DpaiHmsDeleteRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func DeleteDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("DeleteHms:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("Failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	taskDeleteNamespace := deployment.NewTask("DeleteNamespace", DeleteNamespace, nil, []*deployment.Task{})
	tastDeleteSecret := deployment.NewTask("DeleteSecret", DeleteSecret, nil, []*deployment.Task{})
	taskCommitDelete := deployment.NewTask("CommitDelete", CommitDelete, nil, []*deployment.Task{taskDeleteNamespace, tastDeleteSecret})

	deploy.AddTasks([]*deployment.Task{taskDeleteNamespace, tastDeleteSecret, taskCommitDelete})

	deploy.Run()

	return deploy, nil
}
