// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package airflow

import (
	// "context"
	"encoding/json"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment/workspace"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func GetDeploymentDeleteInput(d deployment.Deployment) (*pb.DpaiAirflowDeleteRequest, error) {
	var value pb.DpaiAirflowDeleteRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// DeleteDeployment executes all the tasks that are needed to delete the workspace.
func DeleteDeployment(ctx deployment.DeploymentInputContext) (*deployment.Deployment, error) {
	// Create New Deployment
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("DeleteAirflow:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	// taskDeleteNamespace := deployment.NewTask("DeleteNamespace", DeleteNamespace, nil, []*deployment.Task{})
	// tastDeleteSecret := deployment.NewTask("DeleteSecret", DeleteSecret, nil, []*deployment.Task{})
	taskDeleteCluster := deployment.NewTask("DeleteCluster", workspace.DeleteIKSCluster, nil, []*deployment.Task{})
	taskCommitDelete := deployment.NewTask("CommitDelete", CommitDelete, nil, []*deployment.Task{taskDeleteCluster})

	deploy.AddTasks([]*deployment.Task{taskDeleteCluster, taskCommitDelete})

	deploy.Run()

	return deploy, nil
}
