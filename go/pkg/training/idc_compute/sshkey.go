// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package idc_compute

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/common"
)

func (idcSvc *IDCServiceClient) CreateSSHKey(ctx context.Context, clusterId, cloudAccount string, pubKey []byte) (*v1.SshPublicKey, error) {
	log := log.FromContext(ctx).WithName("IDCServiceProvider.CreateSSHKey")

	rand.Seed(time.Now().UnixNano())

	sshKeyReq := v1.SshPublicKeyCreateRequest{
		Metadata: &v1.ResourceMetadataCreate{
			Name:           fmt.Sprintf("%s-sshkey-%d", clusterId, rand.Intn(100)),
			CloudAccountId: cloudAccount,
		},
		Spec: &v1.SshPublicKeySpec{
			SshPublicKey: string(pubKey),
		},
	}
	log.Info("sshkey", "name", sshKeyReq.Metadata.Name)
	sshkey, err := v1.NewSshPublicKeyServiceClient(idcSvc.ComputeAPIConn).Create(ctx, &sshKeyReq)
	if err != nil {
		log.Error(err, "error storing sshkey API")
		return nil, err
	}

	log.Info("Created SShkey", "key", sshkey.Metadata.Name)
	return sshkey, nil
}

func (idcSvc *IDCServiceClient) DeleteSSHKey(ctx context.Context, cloudAccount, priKeyfile, pubKeyFile string, sshKey *v1.SshPublicKey) error {
	log := log.FromContext(ctx).WithName("IDCServiceProvider.DeleteSSHKey")
	defer common.DeleteSSHKeyPair(ctx, priKeyfile, pubKeyFile)

	keyName := &v1.ResourceMetadataReference_Name{
		Name: sshKey.Metadata.Name,
	}

	deleteReq := &v1.SshPublicKeyDeleteRequest{
		Metadata: &v1.ResourceMetadataReference{
			NameOrId:       keyName,
			CloudAccountId: cloudAccount,
		},
	}

	_, err := v1.NewSshPublicKeyServiceClient(idcSvc.ComputeAPIConn).Delete(ctx, deleteReq)
	if err != nil {
		log.Error(err, "error deleting sshkey API")
		return err
	}

	log.Info("Deleted sshkey", "key", sshKey.Metadata.Name)
	return nil
}

func (idcSvc *IDCServiceClient) IsSSHKeyExists(ctx context.Context, cloudAccount, sshKeyName string) bool {
	logger := log.FromContext(ctx).WithName("IDCServiceProvider.IsSSHKeyExists")

	if pubKey, err := idcSvc.GetPublicKey(ctx, cloudAccount, sshKeyName); err != nil || pubKey == nil {
		logger.Error(err, "error getting public key", "pubkey", pubKey)
		return false
	}
	return true
}

func (idcSvc *IDCServiceClient) GetPublicKey(ctx context.Context, cloudAccount, sshKeyName string) ([]byte, error) {
	logger := log.FromContext(ctx).WithName("IDCServiceProvider.GetPublicKey")
	logger.Info("getting public sshkey by name", "ssh keyname", sshKeyName)
	keyName := &v1.ResourceMetadataReference_Name{
		Name: sshKeyName,
	}

	getReq := &v1.SshPublicKeyGetRequest{
		Metadata: &v1.ResourceMetadataReference{
			NameOrId:       keyName,
			CloudAccountId: cloudAccount,
		},
	}

	pubKey, err := v1.NewSshPublicKeyServiceClient(idcSvc.ComputeAPIConn).Get(ctx, getReq)
	if err != nil || pubKey == nil {
		logger.Error(err, "error getting sshkey API")
		return nil, err
	}
	return []byte(pubKey.Spec.GetSshPublicKey()), nil
}
