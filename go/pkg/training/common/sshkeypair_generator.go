// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package common

// This shows an example of how to generate a SSH RSA Private/Public key pair and save it locally

// source: https://gist.github.com/devinodaniel/8f9b8a4f31573f428f29ec0e884e6673

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"golang.org/x/crypto/ssh"
)

func GenerateSSHKeyPair(ctx context.Context, priKeyfile, pubKeyFile string) error {
	log := log.FromContext(ctx).WithName("GenerateSSHKeyPair")
	bitSize := 4096

	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		log.Error(err, "Failed to generate private key")
		return err
	}

	publicKeyBytes, err := generatePublicKey(ctx, &privateKey.PublicKey)
	if err != nil {
		log.Error(err, "Failed to generate public key")
		return err
	}

	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

	err = writeKeyToFile(ctx, privateKeyBytes, priKeyfile)
	if err != nil {
		log.Error(err, "Failed to write public key")
		return err
	}

	err = writeKeyToFile(ctx, []byte(publicKeyBytes), pubKeyFile)
	if err != nil {
		log.Error(err, "Failed to write public key")
		return err
	}
	return nil
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	log := log.FromContext(context.Background()).WithName("generatePrivateKey")

	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Info("Private Key generated")
	return privateKey, nil
}

// delete sshkey pair
func DeleteSSHKeyPair(ctx context.Context, priKeyfile, pubKeyFile string) error {
	log := log.FromContext(ctx).WithName("DeleteSSHKeyPair")

	sshkeyname := filepath.Base(priKeyfile)
	log.Info("Deleting sshkeypair", "key", sshkeyname)
	defer log.Info("Deleted sshkeypair", "key", sshkeyname)

	// Delete the private key file
	err := os.Remove(priKeyfile)
	if err != nil {
		log.Error(err, "Failed to delete private key file")
	}

	// Delete the public key file
	err = os.Remove(pubKeyFile)
	if err != nil {
		log.Error(err, "Failed to delete public key file")
	}

	return nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(ctx context.Context, privatekey *rsa.PublicKey) ([]byte, error) {
	log := log.FromContext(ctx).WithName("generatePublicKey")
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Info("Public key generated")
	return pubKeyBytes, nil
}

// writePemToFile writes keys to a file
func writeKeyToFile(ctx context.Context, keyBytes []byte, saveFileTo string) error {
	log := log.FromContext(ctx).WithName("generatePublicKey")
	err := ioutil.WriteFile(saveFileTo, keyBytes, 0600)
	if err != nil {
		return err
	}

	log.Info("Key saved to ", "File", saveFileTo)
	return nil
}
