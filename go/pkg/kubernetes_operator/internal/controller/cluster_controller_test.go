// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	apiserverv1 "k8s.io/apiserver/pkg/apis/config/v1"
)

func TestRotateEtcdEncryptionKeyWithNoResources(t *testing.T) {
	encConfig := "apiVersion: apiserver.config.k8s.io/v1\nkind: EncryptionConfiguration\t\n"
	errMessage := "there are no resources in encryption config"

	_, err := rotateEtcdEncryptionKey(encConfig, "foo", "bar")
	if err == nil {
		t.Fatalf("want error, got %s", err)
	}

	if !strings.Contains(err.Error(), errMessage) {
		t.Fatalf("want %s, got %s", errMessage, err)
	}
}

func TestRotateEtcdEncryptionKeyWithNoProviders(t *testing.T) {
	encConfig := "apiVersion: apiserver.config.k8s.io/v1\nkind: EncryptionConfiguration\nresources:\n  - resources:\n      - secrets\t\n"
	errMessage := "there are no providers in encryption config"

	_, err := rotateEtcdEncryptionKey(encConfig, "foo", "bar")
	if err == nil {
		t.Fatalf("want error, got %s", err)
	}

	if !strings.Contains(err.Error(), errMessage) {
		t.Fatalf("want %s, got %s", errMessage, err)
	}
}

func TestRotateEtcdEncryptionKeyWithTwoKeys(t *testing.T) {
	encConfig := "apiVersion: apiserver.config.k8s.io/v1\nkind: EncryptionConfiguration\nresources:\n  - resources:\n      - secrets\n    providers:\n      - aescbc:\n          keys:\n            \n            - name: 2023-10-25T00:10:50Z\n              secret: foo1\n            \n            - name: 2023-10-25T01:10:50Z\n              secret: foo2\n            \n      - identity: {}\t\n"

	expectedRotatedKeys := []apiserverv1.Key{
		{

			Name:   "2023-10-25T01:10:50Z",
			Secret: "foo2",
		},
		{
			Name:   "2023-10-28T00:10:50Z",
			Secret: "rotated",
		},
		{
			Name:   "2023-10-25T00:10:50Z",
			Secret: "foo1",
		},
	}

	rotatedKeys, err := rotateEtcdEncryptionKey(encConfig, expectedRotatedKeys[1].Name, expectedRotatedKeys[1].Secret)
	if err != nil {
		t.Fatalf("want nil error, got %s", err)
	}

	for i, key := range rotatedKeys {
		if key.Name != expectedRotatedKeys[i].Name {
			t.Fatalf("want key name in position %d to be %s, got %s", i, expectedRotatedKeys[i].Name, key.Name)
		}

		if key.Secret != expectedRotatedKeys[i].Secret {
			t.Fatalf("want key secret in position %d to be %s, got %s", i, expectedRotatedKeys[i].Secret, key.Secret)
		}
	}
}

func TestRotateEtcdEncryptionKeyWithThreeKeys(t *testing.T) {
	encConfig := "apiVersion: apiserver.config.k8s.io/v1\nkind: EncryptionConfiguration\nresources:\n  - resources:\n      - secrets\n    providers:\n      - aescbc:\n          keys:\n            \n            - name: 2023-10-25T00:10:50Z\n              secret: foo2\n            \n            - name: 2023-10-25T01:10:50Z\n              secret: foo3\n            \n            - name: 2023-10-24T01:10:50Z\n              secret: foo1\n            \n      - identity: {}\t\n"

	expectedRotatedKeys := []apiserverv1.Key{
		{

			Name:   "2023-10-25T01:10:50Z",
			Secret: "foo3",
		},
		{
			Name:   "2023-10-28T00:10:50Z",
			Secret: "rotated",
		},
		{
			Name:   "2023-10-25T00:10:50Z",
			Secret: "foo2",
		},
	}

	rotatedKeys, err := rotateEtcdEncryptionKey(encConfig, expectedRotatedKeys[1].Name, expectedRotatedKeys[1].Secret)
	if err != nil {
		t.Fatalf("want nil error, got %s", err)
	}

	for i, key := range rotatedKeys {
		if key.Name != expectedRotatedKeys[i].Name {
			t.Fatalf("want key name in position %d to be %s, got %s", i, expectedRotatedKeys[i].Name, key.Name)
		}

		if key.Secret != expectedRotatedKeys[i].Secret {
			t.Fatalf("want key secret in position %d to be %s, got %s", i, expectedRotatedKeys[i].Secret, key.Secret)
		}
	}
}

func TestAddEtcdEncryptionConfigInClusterSecretLessThanThreeConfigs(t *testing.T) {
	instanceIMI := "iks-u22-cd-cp-1-28-4-23-11-19"
	var rotatedEtcdEncryptionConfig apiserverv1.EncryptionConfiguration

	etcdEncryptionConfigs := map[string]string{
		"iks-u22-cd-cp-1-27-4-23-09-19": "",
	}

	etcdEncryptionConfigsBytes, err := json.Marshal(etcdEncryptionConfigs)
	if err != nil {
		t.Fatalf("failed to marshal etcd encryption configs, error: %s", err.Error())
	}

	clusterSecret := v1.Secret{
		Data: map[string][]byte{
			"etcd-encryption-configs": etcdEncryptionConfigsBytes,
		},
	}

	etcdEncryptionsconfigs, err := addEtcdEncryptionConfigInClusterSecret(instanceIMI, rotatedEtcdEncryptionConfig, &clusterSecret)
	if err != nil {
		t.Fatalf("want nil error, got %s", err)
	}

	for _, wantedIMI := range []string{"iks-u22-cd-cp-1-27-4-23-09-19", "iks-u22-cd-cp-1-28-4-23-11-19"} {
		if _, found := etcdEncryptionsconfigs[wantedIMI]; !found {
			t.Fatalf("want %s error, got %s", wantedIMI, getKeysFromMap(etcdEncryptionsconfigs))
		}
	}
}

func TestAddEtcdEncryptionConfigInClusterSecretMoreThanOrEqualThreeConfigs(t *testing.T) {
	instanceIMI := "iks-u22-cd-cp-1-28-4-23-11-19"
	var rotatedEtcdEncryptionConfig apiserverv1.EncryptionConfiguration

	etcdEncryptionConfigs := map[string]string{
		"iks-u22-cd-cp-1-26-9-23-09-19": "",
		"iks-u22-cd-cp-1-27-4-23-09-19": "",
		"iks-u22-cd-cp-1-25-9-23-09-18": "",
	}

	etcdEncryptionConfigsBytes, err := json.Marshal(etcdEncryptionConfigs)
	if err != nil {
		t.Fatalf("failed to marshal etcd encryption configs, error: %s", err.Error())
	}

	clusterSecret := v1.Secret{
		Data: map[string][]byte{
			"etcd-encryption-configs": etcdEncryptionConfigsBytes,
		},
	}

	etcdEncryptionsconfigs, err := addEtcdEncryptionConfigInClusterSecret(instanceIMI, rotatedEtcdEncryptionConfig, &clusterSecret)
	if err != nil {
		t.Fatalf("want nil error, got %s", err)
	}

	for _, wantedIMI := range []string{"iks-u22-cd-cp-1-26-9-23-09-19", "iks-u22-cd-cp-1-27-4-23-09-19", "iks-u22-cd-cp-1-28-4-23-11-19"} {
		if _, found := etcdEncryptionsconfigs[wantedIMI]; !found {
			t.Fatalf("want %s error, got %s", wantedIMI, getKeysFromMap(etcdEncryptionsconfigs))
		}
	}
}

func TestAddEtcdEncryptionConfigInClusterSecretConfigsNotFound(t *testing.T) {
	instanceIMI := "iks-u22-cd-cp-1-28-4-23-11-19"
	var rotatedEtcdEncryptionConfig apiserverv1.EncryptionConfiguration
	errMessage := fmt.Sprintf("can not find %s", etcdEncryptionConfigsClusterSecretKey)

	clusterSecret := v1.Secret{
		Data: map[string][]byte{},
	}

	_, err := addEtcdEncryptionConfigInClusterSecret(instanceIMI, rotatedEtcdEncryptionConfig, &clusterSecret)
	if err == nil {
		t.Fatalf("want error, got %s", err)
	}

	if !strings.Contains(err.Error(), errMessage) {
		t.Fatalf("want %s, got %s", errMessage, err)
	}
}

func getKeysFromMap(m map[string]string) []string {
	keys := make([]string, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}
