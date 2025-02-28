// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

type CertConfig struct {
	CommonName    string
	Organizations []string
	DNSNames      []string
	IPAddresses   []net.IP
}

func CreateCa(cn string, caCertExpirationPeriod time.Duration) (string, string, error) {
	expirationPeriod := time.Now().AddDate(10, 0, 0)
	if caCertExpirationPeriod != 0 {
		expirationPeriod = time.Now().Add(caCertExpirationPeriod)
	}
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName: cn,
		},
		NotBefore: time.Now(),
		NotAfter:  expirationPeriod,
		IsCA:      true,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return "", "", err
	}

	caPEM := new(bytes.Buffer)
	if err := pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return "", "", err
	}

	caPrivateKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(caPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	}); err != nil {
		return "", "", err
	}

	return caPEM.String(), caPrivateKeyPEM.String(), nil
}

func CreateAndSignCert(ca *x509.Certificate, caPrivateKey *rsa.PrivateKey, certConfig CertConfig, expiration *time.Time) ([]byte, []byte, error) {
	today := time.Now()
	expirationDate := today.AddDate(3, 0, 0)
	if expiration != nil {
		expirationDate = *expiration
	}
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   certConfig.CommonName,
			Organization: certConfig.Organizations,
		},
		NotBefore:   today,
		NotAfter:    expirationDate,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    certConfig.DNSNames,
		IPAddresses: certConfig.IPAddresses,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		cert,
		ca,
		&privateKey.PublicKey,
		caPrivateKey)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return []byte{}, []byte{}, err
	}

	privateKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return []byte{}, []byte{}, err
	}

	return certPEM.Bytes(), privateKeyPEM.Bytes(), nil
}

func ParseCert(certb []byte) (*x509.Certificate, error) {
	certBlock, _ := pem.Decode(certb)
	if certBlock == nil {
		return nil, fmt.Errorf("cert could not be decoded")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func ParsePrivateKey(keyb []byte) (*rsa.PrivateKey, error) {
	keyBlock, _ := pem.Decode(keyb)
	if keyBlock == nil {
		return nil, fmt.Errorf("key could not be decoded")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func GetPublicKey(privateKey *rsa.PrivateKey) (string, error) {
	publicKeyB, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", err
	}

	publicKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(publicKeyPEM, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyB,
	}); err != nil {
		return "", err
	}

	return publicKeyPEM.String(), nil
}
