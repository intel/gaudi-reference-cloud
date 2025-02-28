// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"archive/tar"
	"compress/gzip"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/utils"
	"github.com/pkg/errors"
)

const (
	deleteNodesFinalizer   = "private.cloud.intel.com/deleteNodes"
	deleteClusterFinalizer = "private.cloud.intel.com/deleteCluster"
	deleteAddonFinilizer   = "private.cloud.intel.com/deleteAddon"

	ipAddressesAnnotation = "nodegroup.devcloud.io/ip-addresses"
	nameserverAnnotation  = "nodegroup.devcloud.io/nameserver"
	gatewayAnnotation     = "nodegroup.devcloud.io/gateway"

	etcdEncryptionConfigsClusterSecretKey = "etcd-encryption-configs"
)

type OperatorMessage struct {
	ErrorCode int32  `json:"errorCode"`
	Message   string `json:"message"`
}

func createArchive(files []string, buf io.Writer) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = filename

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

func getKubernetesCACertKey(ctx context.Context, client client.Client, clusterName string, namespace string) ([]byte, []byte, error) {
	clusterSecret, err := utils.GetSecret(ctx, client, clusterName, namespace)
	if err != nil {
		return nil, nil, err
	}

	caCert, err := utils.GetDataFromSecret(clusterSecret, "ca.crt")
	if err != nil {
		return nil, nil, err
	}

	caKey, err := utils.GetDataFromSecret(clusterSecret, "ca.key")
	if err != nil {
		return nil, nil, err
	}

	return caCert, caKey, nil
}

func getKubernetesClientCerts(caCertb []byte, caKeyb []byte, commonName string, organization string, expirationPeriod time.Duration) ([]byte, []byte, error) {
	caCert, err := utils.ParseCert(caCertb)
	if err != nil {
		return nil, nil, err
	}

	caKey, err := utils.ParsePrivateKey(caKeyb)
	if err != nil {
		return nil, nil, err
	}

	expiration := time.Now().Add(time.Minute * 30)
	if expirationPeriod != 0 {
		expiration = time.Now().Add(expirationPeriod)
	}
	cert, key, err := utils.CreateAndSignCert(caCert, caKey, utils.CertConfig{
		CommonName: commonName,
		Organizations: []string{
			organization,
		},
	}, &expiration)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func GenerateRandomString(size int) ([]byte, error) {
	opts := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	random := make([]byte, size)

	for i := 0; i < size; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(opts))))
		if err != nil {
			return []byte{}, err
		}
		random[i] = opts[num.Int64()]
	}

	return random, nil
}

func GetEtcdEncryptionConfigFromClusterSecret(ctx context.Context, c client.Client, key k8stypes.NamespacedName, instanceIMI string) (string, *corev1.Secret, error) {
	var clusterSecret corev1.Secret
	if err := c.Get(ctx, key, &clusterSecret); err != nil {
		return "", nil, errors.WithMessagef(err, "get cluster secret")
	}

	etcdEncryptionConfigsB, found := clusterSecret.Data[etcdEncryptionConfigsClusterSecretKey]
	if !found {
		return "", nil, errors.Errorf("can not find %s in cluster secret %s", etcdEncryptionConfigsClusterSecretKey, clusterSecret.Name)
	}

	var etcdEncryptionConfigs map[string]string
	err := json.Unmarshal(etcdEncryptionConfigsB, &etcdEncryptionConfigs)
	if err != nil {
		return "", nil, errors.WithMessagef(err, "unmarshal etcd encription configs")
	}

	etcdEncryptionConfig, found := etcdEncryptionConfigs[instanceIMI]
	if !found {
		return "", nil, errors.Errorf("can not find etcd encryption config for instanceIMI %s in cluster secret: %s", instanceIMI, clusterSecret.Name)
	}

	return etcdEncryptionConfig, clusterSecret.DeepCopy(), nil
}

func convertToTB(size string) int64 {
	if strings.HasSuffix(size, "GB") {
		// split string to extract numeric value
		splits := strings.Split(size, "GB")
		if len(splits) != 2 {
			return -1
		}
		// convert value to bytes
		sizeInt, err := strconv.ParseInt(splits[0], 10, 64)
		if err != nil {
			return -1
		}
		return sizeInt / 1000
	} else if strings.HasSuffix(size, "TB") {
		// split string to extract numeric value
		splits := strings.Split(size, "TB")
		if len(splits) != 2 {
			return -1
		}
		// convert value to bytes
		sizeInt, err := strconv.ParseInt(splits[0], 10, 64)
		if err != nil {
			return -1
		}
		return sizeInt
	} else if size == "" {
		return 0
	} else {
		return -1
	}
}
