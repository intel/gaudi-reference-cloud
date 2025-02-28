// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test_tools

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"golang.org/x/crypto/ssh"
)

const rsaPrivateKeyPemType string = "RSA PRIVATE KEY"

// Create an SSH key pair.
// privateKey is in PEM format.
// publicKey is in the format "ssh-rsa ... comment".
func CreateSshRsaKeyPair(bitSize int, comment string) (privateKey string, publicKey string, err error) {
	// Create private key
	var rsaPrivateKey *rsa.PrivateKey
	rsaPrivateKey, err = rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return
	}
	err = rsaPrivateKey.Validate()
	if err != nil {
		return
	}

	// Create public key
	var sshPublicKey ssh.PublicKey
	sshPublicKey, err = ssh.NewPublicKey(&rsaPrivateKey.PublicKey)
	if err != nil {
		return
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)
	publicKey = strings.TrimSuffix(string(publicKeyBytes), "\n")
	if comment != "" {
		publicKey = publicKey + " " + comment
	}

	// Encode private key to PEM format.
	privateDer := x509.MarshalPKCS1PrivateKey(rsaPrivateKey)
	privateBlock := pem.Block{
		Type:    rsaPrivateKeyPemType,
		Headers: nil,
		Bytes:   privateDer,
	}
	privatePem := pem.EncodeToMemory(&privateBlock)
	privateKey = string(privatePem)
	return
}
