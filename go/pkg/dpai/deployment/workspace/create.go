// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package workspace

import (
	// "context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/deployment"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func GetDeploymentCreateInput(d deployment.Deployment) (*pb.DpaiWorkspaceCreateRequest, error) {
	var value pb.DpaiWorkspaceCreateRequest
	err := json.Unmarshal(d.RawInput, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// CreateDeployment executes all the tasks that are needed to create the workspace.
func CreateDeployment(ctx deployment.DeploymentInputContext) (*string, error) {

	// Create New Deployment
	fmt.Print("the deployment task input context for workspace creation", ctx.Conf)
	deploy, err := deployment.NewDeployment(ctx, fmt.Sprintf("CreateWorkspace:%s", ctx.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to create the DeploymentJob for Deployment: %s with error: %+v", ctx.ID, err)
	}

	// Create and Add Tasks
	createCluster := deployment.NewTask("CreateIKSCluster", CreateIKSCluster, nil, []*deployment.Task{})
	validateCluster := deployment.NewTask("ValidateIKSCluster", ValidateIKSCluster, nil, []*deployment.Task{createCluster})
	createSecretNamespace := deployment.NewTask("CreateSecretNamespace", CreateSecretNamespace, nil, []*deployment.Task{validateCluster})

	createSshKey := deployment.NewTask("CreateSshKey", CreateSshKey, nil, []*deployment.Task{createSecretNamespace})
	createNodeGroup := deployment.NewTask("CreateNodeGroup", CreateNodeGroup, nil, []*deployment.Task{createSshKey})

	// // cert manager
	certManagerNamespace := deployment.NewTask("CreateCertManagerNS", CreateCertManagerNS, nil, []*deployment.Task{createNodeGroup})
	installCertManager := deployment.NewTask("InstallCertManagerHelmChart", InstallCertManagerHelmChart, nil, []*deployment.Task{certManagerNamespace})
	createIstioGatewayCerts := deployment.NewTask("CreateIstioGatewayCerts", CreateIstioGatewayCerts, nil, []*deployment.Task{installCertManager})
	// // service mesh
	createIstioNamespace := deployment.NewTask("CreateIstioNamespace", CreateIstioNamespace, nil, []*deployment.Task{createNodeGroup})
	installIstioCrd := deployment.NewTask("InstallIstioCRD", InstallIstioCRD, nil, []*deployment.Task{createIstioNamespace})
	installIstioDaemon := deployment.NewTask("InstallIstioDaemon", InstallIstioDaemon, nil, []*deployment.Task{installIstioCrd})
	installIstioGatewayControllers := deployment.NewTask("InstallIstioIngressGateway", InstallIstioIngressGateway, nil, []*deployment.Task{installIstioDaemon})

	// load balancer operator enhanced by dpai team we dont call compute api since the enhancement is not exposed and internal
	createLoadBalancerCRD := deployment.NewTask("CreateLoadBalancerCrd", CreateLoadBalancerCrd, nil, []*deployment.Task{installIstioGatewayControllers})

	createOpenEbsNamespace := deployment.NewTask("CreateOpenEbsNamespace", CreateOpenEbsNamespace, nil, []*deployment.Task{installIstioGatewayControllers})
	installOpenEbs := deployment.NewTask("InstallOpenEbs", InstallOpenEbs, nil, []*deployment.Task{createOpenEbsNamespace})

	// commit := deployment.NewTask("CommitCreateWorkspace", CommitCreateWorkspace, nil, []*deployment.Task{createNodeGroup, createLoadBalancerCRD, installOpenEbs})

	deploy.AddTasks([]*deployment.Task{
		createCluster, validateCluster,
		createNodeGroup,
		certManagerNamespace, installCertManager, createIstioGatewayCerts,
		createIstioNamespace, installIstioCrd, installIstioDaemon, installIstioGatewayControllers,
		createLoadBalancerCRD,
		createSecretNamespace, createSshKey,
		// 	createNodeGroup,
		// 	createIstioNamespace, installIstio,
		createOpenEbsNamespace, installOpenEbs,
		// commit,
	})

	output, deploymentError := deploy.Run()

	var result TaskOutput
	err = json.Unmarshal(output, &result)
	if err != nil {
		log.Printf("Failed to convert the result to Output object. Error message: %+v", err)
		log.Printf("Incase of failure in the deployment: %s, cleanup task can not be executed.", deploy.ID)
		return nil, err
	}

	if deploymentError != nil {
		if deploy.Context.ParentDeploymentId != "" {
			log.Printf("Cleanup will be done in the parent Deployment.")
		} else {
			// TODO have the cleanup logic here
		}
		return nil, deploymentError
	}

	return &result.WorkspaceID, nil
}
