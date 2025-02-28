// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package networking

import (
	"context"
	"log"

	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	sql "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FirewallVerifyRequest struct {
	WorkspaceID    string
	CloudAccountID string
	LbResourceName string // should only inspect the lb resource, sicne firewall operator has required added finalizers
	AzK8sClientSet *k8s.K8sAzClient
}

// all added since the firewall rule take a while and lb operator has added finalizer to the resource of firewall rule crd until it get provisioned
func (n *Network) VerifyIKSFirewallCreation(ctx context.Context, verifyFirewallCreateInput *FirewallVerifyRequest,
	sqlQueryCon *sql.Queries) (bool, error) {
	resource := &loadbalancerv1alpha1.Loadbalancer{}

	loadBalancerSchema := runtime.NewScheme()

	if err := loadbalancerv1alpha1.AddToScheme(loadBalancerSchema); err != nil {
		log.Printf("Error adding loadbalancer schema to runtime %+v", err)
		return false, err
	}

	loadbalancerClient, err := client.New(verifyFirewallCreateInput.AzK8sClientSet.ClientConfig,
		client.Options{Scheme: loadBalancerSchema})
	if err != nil {
		log.Fatalf("Failed to create load-balancer client in controller-runtime: %+v", err)
		return false, err
	}

	err = loadbalancerClient.Get(context.Background(), client.ObjectKey{Name: verifyFirewallCreateInput.LbResourceName,
		Namespace: verifyFirewallCreateInput.CloudAccountID}, resource)

	if err != nil {
		return false, err
	}

	isFirewallCreated := false

	// rely on the activr lb operator state added on the lb resource manifest to determine the firewall rule creation status
	// the lb operator internally adds required finalizers over the firewall resource managed by firewall controller
	fwCreatedStatus := resource.Status.State == loadbalancerv1alpha1.READY
	if fwCreatedStatus {
		_, err := sqlQueryCon.UpdateGatewayFwStatusForWorkspace(context.Background(), db.UpdateGatewayFwStatusForWorkspaceParams{
			WorkspaceID:    verifyFirewallCreateInput.WorkspaceID,
			FwCreated:      true,
			CloudAccountID: verifyFirewallCreateInput.CloudAccountID,
		})
		if err != nil {
			log.Fatalf("Error updating the Gateway Firewall Status for the workspace: %+v", err)
			return false, err
		}

		log.Printf("The Firewall Rule for the Load Balancer %s is successfully created", verifyFirewallCreateInput.LbResourceName)

		isFirewallCreated = true
		return isFirewallCreated, nil
	}

	return isFirewallCreated, nil
}
