// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package idc_compute

import (
	"context"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	retry "github.com/sethvargo/go-retry"
)

func (idcSvc *IDCServiceClient) CreateInstance(ctx context.Context, createRequest InstanceCreateRequest) (*v1.InstancePrivate, error) {
	log := log.FromContext(ctx).WithName("IDCServiceClient.CreateInstance")

	networkInterfacePrivate := &v1.NetworkInterfacePrivate{
		Name: "eth0",
		VNet: createRequest.VNet,
	}
	instanceRequest := v1.InstanceCreatePrivateRequest{
		Metadata: &v1.InstanceMetadataCreatePrivate{
			Name:           createRequest.Name,
			Labels:         createRequest.Labels,
			CloudAccountId: createRequest.CloudAccountId,
			SkipQuotaCheck: true,
		},
		Spec: &v1.InstanceSpecPrivate{
			AvailabilityZone:  "us-staging-1a",
			InstanceType:      createRequest.MachineType,
			MachineImage:      createRequest.ImageName,
			SshPublicKeyNames: createRequest.SshKeyNames,
			Interfaces:        []*v1.NetworkInterfacePrivate{networkInterfacePrivate},
			UserData:          createRequest.UserData,
		},
	}

	// create instance
	inst, err := v1.NewInstancePrivateServiceClient(idcSvc.ComputeAPIConn).CreatePrivate(ctx, &instanceRequest)
	if err != nil {
		log.Error(err, "error creating instance")
		return nil, fmt.Errorf("error creating instance")
	}

	log.Info("debug", "instance create response", inst)
	return inst, nil
}

func (idcSvc *IDCServiceClient) WaitForInstanceStateReadygRPC(ctx context.Context, instanceId, cloudAccount string, timeout time.Duration) error {
	log := log.FromContext(ctx).WithName("IDCServiceClient.WaitForInstanceStateReadygRPC")
	log.Info("get and wait instace state ", "instanceId", instanceId)

	backoffTimer := retry.NewConstant(5 * time.Second)
	backoffTimer = retry.WithMaxDuration(timeout*time.Second, backoffTimer)

	instanceGetRequest := &v1.InstanceGetRequest{
		Metadata: &v1.InstanceMetadataReference{
			CloudAccountId: cloudAccount,
			NameOrId: &v1.InstanceMetadataReference_ResourceId{
				ResourceId: instanceId,
			},
		},
	}

	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		instance, err := v1.NewInstanceServiceClient(idcSvc.ComputeAPIConn).Get(ctx, instanceGetRequest)
		if err != nil {
			return fmt.Errorf("error reading instace state")
		}
		if instance == nil {
			return fmt.Errorf("error reading instace state, nil instance")
		}
		log.Info("debug", "instance-phase", instance.Status.Phase)
		if instance.Status.Phase.String() != "Ready" {
			return retry.RetryableError(fmt.Errorf("instance state not ready, retry again"))
		}
		log.Info("instance-state-ready", "instanceId", instanceId, "cloudAccountId", cloudAccount)
		return nil
	}); err != nil {
		return fmt.Errorf("instance state not ready after retries")
	}
	return nil
}

func (idcSvc *IDCServiceClient) GetInstance(ctx context.Context, instanceId, cloudAccount string) (*v1.Instance, error) {
	log := log.FromContext(ctx).WithName("IDCServiceProvider.GetInstance")

	instanceGetRequest := &v1.InstanceGetRequest{
		Metadata: &v1.InstanceMetadataReference{
			CloudAccountId: cloudAccount,
			NameOrId: &v1.InstanceMetadataReference_ResourceId{
				ResourceId: instanceId,
			},
		},
	}
	instance, err := v1.NewInstanceServiceClient(idcSvc.ComputeAPIConn).Get(ctx, instanceGetRequest)
	if err != nil {
		log.Error(err, "error instance get API")
		return nil, fmt.Errorf("error getting instance")
	}

	return instance, nil
}

func (idcSvc *IDCServiceClient) GetInstanceByName(ctx context.Context, instanaceName, cloudAccount string) (*v1.Instance, error) {
	log := log.FromContext(ctx).WithName("IDCServiceClient.GetInstanceByNamegRPC")

	instanceGetRequest := &v1.InstanceGetRequest{
		Metadata: &v1.InstanceMetadataReference{
			CloudAccountId: cloudAccount,
			NameOrId: &v1.InstanceMetadataReference_Name{
				Name: instanaceName,
			},
		},
	}

	instance, err := v1.NewInstanceServiceClient(idcSvc.ComputeAPIConn).Get(ctx, instanceGetRequest)
	if err != nil {
		log.Error(err, "error instance get API")
		return nil, fmt.Errorf("error in get instance by name")
	}

	return instance, nil
}

func (idcSvc *IDCServiceClient) DeleteInstance(ctx context.Context, instanceName, cloudAccountId string) error {
	log := log.FromContext(ctx).WithName("IDCServiceClient.DeleteInstance")

	instanceDeletePrivateRequest := &v1.InstanceDeletePrivateRequest{
		Metadata: &v1.InstanceMetadataReference{
			CloudAccountId: cloudAccountId,
			NameOrId: &v1.InstanceMetadataReference_Name{
				Name: instanceName,
			},
		},
	}

	_, err := v1.NewInstancePrivateServiceClient(idcSvc.ComputeAPIConn).DeletePrivate(ctx, instanceDeletePrivateRequest)
	if err != nil {
		log.Error(err, "error deleting instance via compute endpoint")
		return err
	}

	return nil
}
