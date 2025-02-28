// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Copyright © 2023 Intel Corporation
package ipacmd

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/myssh"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets"
)

const (
	ipaImageSecretsPath = "ipaimage/sshkey/"
	sudoCommand         = "/usr/bin/sudo"
	ipaImageUser        = "sdp"
	sshConnectTimeout   = 30
	sshConnectPort      = 22
)

type ipaCmdHelper struct {
	clientConfig *ssh.ClientConfig
	sshManager   myssh.SSHManagerAccessor
	ironicIp     string
}

func NewIpaCmdHelper(ctx context.Context, vault secrets.SecretManager, sshManager myssh.SSHManagerAccessor, ironicIp string, region string) (*ipaCmdHelper, error) {
	log := log.FromContext(ctx).WithName("IpaCmd.NewIpaCmdHelper")
	log.Info("Creating SSH Configuration")

	// get vault secrets
	secretPath := fmt.Sprintf("%s/baremetal/enrollment/ipaimage/sshkey", region)
	privateKey, err := vault.GetIPAImageSSHPrivateKey(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPA Image SSH private key with error:  %v", err)
	}

	timeout := sshConnectTimeout * time.Second

	// Parse the private key
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	// Set up the SSH config
	config := &ssh.ClientConfig{
		User: ipaImageUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	log.Info("SSH Remote Connection info", "ipaImageUser", ipaImageUser, "ironicIp", ironicIp)

	helper := &ipaCmdHelper{
		clientConfig: config,
		sshManager:   sshManager,
		ironicIp:     ironicIp,
	}

	return helper, nil
}
