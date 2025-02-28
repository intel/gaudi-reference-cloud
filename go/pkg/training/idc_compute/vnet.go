// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package idc_compute

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func (idcSvc *IDCServiceClient) CreateVNet(ctx context.Context, clusterId, cloudAccount, vnetName string, reqVnet *v1.VNetSpec) (*v1.VNet, error) {
	log := log.FromContext(ctx).WithName("IDCServiceProvider.CreateVNetgRPC")

	// TODO: Make this dynamic accepting it via the API body once it has been verified working.
	vnetRequest := &v1.VNetPutRequest{
		Metadata: &v1.VNetPutRequest_Metadata{
			Name:           vnetName,
			CloudAccountId: cloudAccount,
		},
		Spec: &v1.VNetSpec{
			AvailabilityZone: reqVnet.AvailabilityZone,
			Region:           reqVnet.Region,
			PrefixLength:     reqVnet.PrefixLength,
		},
	}

	vnet, err := v1.NewVNetServiceClient(idcSvc.ComputeAPIConn).Put(ctx, vnetRequest)
	if err != nil {
		log.Error(err, "error creating vnet create API")
		return nil, fmt.Errorf("error creating vnet")
	}

	log.Info("debug", "vnet create response", vnet)
	return vnet, nil
}

func (idcSvc *IDCServiceClient) IsVNetExists(ctx context.Context, cloudAccount, vnetName string) (bool, error) {
	log := log.FromContext(ctx).WithName("IDCServiceProvider.IsVNetExists")

	keyName := &v1.VNetGetRequest_Metadata_Name{
		Name: vnetName,
	}

	getReq := &v1.VNetGetRequest{
		Metadata: &v1.VNetGetRequest_Metadata{
			NameOrId:       keyName,
			CloudAccountId: cloudAccount,
		},
	}

	vnet, err := v1.NewVNetServiceClient(idcSvc.ComputeAPIConn).Get(ctx, getReq)
	if err != nil || vnet == nil {
		log.Info("error getting vnet", "error", err)
		return false, nil
	}

	log.Info("vnet found for user")
	return true, nil
}
