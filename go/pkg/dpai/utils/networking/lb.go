// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package networking

import (
	"context"
	"fmt"
	"log"

	model "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (n *Network) DeleteIKSLoadbalancerResource(ctx context.Context) (*NetworkRecordsOutput, error) {

	var K8sAzClient k8s.K8sAzClient = k8s.K8sAzClient{}

	azK8sClient, err := K8sAzClient.ConfigureK8sClient()
	if err != nil {
		log.Printf("Error getting the AZ clientset for Az cluster %+v", err)
		return nil, err
	}

	client, err := GetIksAzClusterClient(azK8sClient)

	if err != nil {
		log.Printf("Error getting the LB Operator CTRL client ")
		return nil, err
	}

	// since LB is an IKS resource we dont create unwanted abstractions nad unique Id's on top of the existing cluster ID
	if err := client.Delete(ctx, &loadbalancerv1alpha1.Loadbalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", DPAI_HIGHWIRE_LB_NAME, string(n.IksClusterUUID)),
			Namespace: n.CloudAccountID,
		},
	}); err != nil {
		log.Printf("Error deleting the LB  Resource %s", fmt.Sprintf("%s-%s", DPAI_HIGHWIRE_LB_NAME, string(n.IksClusterUUID)))
		log.Printf("Error :: %+v", err)
		return nil, err
	}

	// delete all records in workspace dns gateway table matching the required workspace id
	lbResource, err := n.SqlModel.GetHwGatewaysByWorkspaceId(context.Background(), model.GetHwGatewaysByWorkspaceIdParams{
		WorkspaceID:    n.WorkspaceId,
		CloudAccountID: n.CloudAccountID,
	})

	if err != nil {
		// TODO: Need a service rollback if the service did not insert gateway as expected into the db
		log.Printf("Error getting the gateway for workspace please make sure the gateway is inserted  ")
		return nil, err
	}

	if err := n.SqlModel.DeleteHwGatewayForWorkspace(context.Background(), model.DeleteHwGatewayForWorkspaceParams{
		WorkspaceID:    n.WorkspaceId,
		CloudAccountID: n.CloudAccountID,
	}); err != nil {
		log.Printf("EError deleting the Gateway for workspace %s. Error message: %+v", n.WorkspaceId, err)
		return nil, err
	}

	log.Println("Successfully deleted the LB  Resource and cleaned the Workspace Gateway tables")

	networkRecords := &NetworkRecordsOutput{
		LBHwDnsARecord: lbResource.LbFqdn,
	}

	return networkRecords, nil
}

func (n *Network) DeleteNetworkingResourcesInDB(ctx context.Context, networkRecords *NetworkRecordsOutput) (*NetworkRecordsOutput, error) {

	// cleaning the lb hw record fqd table expect to cascade clean the service to remve all the service it runs
	err := n.SqlModel.DeleteHwGatewayForWorkspace(ctx, model.DeleteHwGatewayForWorkspaceParams{
		CloudAccountID: n.CloudAccountID,
		WorkspaceID:    n.WorkspaceId,
		LbFqdn:         networkRecords.LBHwDnsARecord,
	})

	if err != nil {
		log.Printf("Error deleting the Gateway for workspace %s. Error message: %+v", n.WorkspaceId, err)
		return nil, err
	}

	log.Printf("Successfully deleted the Gateway for workspace %s", n.WorkspaceId)
	return networkRecords, nil
}
