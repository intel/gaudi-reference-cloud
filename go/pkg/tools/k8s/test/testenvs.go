// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	toolsk8s "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func CreateTestEnvs(numTestEnvs int, crdDirectoryPaths []string) (testEnvs []*envtest.Environment, restConfigList []*restclient.Config, kubeConfigDir string) {
	restConfigs := make(map[string]*restclient.Config)
	for i := 0; i < numTestEnvs; i++ {
		filename := fmt.Sprintf("cluster%d.hconf", i)
		By("Starting Kubernetes API Server for " + filename)
		testEnv := &envtest.Environment{}
		if len(crdDirectoryPaths) != 0 {
			testEnv = &envtest.Environment{
				CRDDirectoryPaths:        crdDirectoryPaths,
				ErrorIfCRDPathMissing:    true,
				AttachControlPlaneOutput: true,
			}
		}
		testEnvs = append(testEnvs, testEnv)
		restConfig, err := testEnv.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(restConfig).NotTo(BeNil())
		restConfigs[filename] = restConfig
		restConfigList = append(restConfigList, restConfig)
	}
	By("WriteKubeConfigFiles")
	var err error
	kubeConfigDir, err = os.MkdirTemp("", "")
	Expect(err).ToNot(HaveOccurred())
	Expect(toolsk8s.WriteKubeConfigFiles(context.Background(), kubeConfigDir, restConfigs)).Should(Succeed())
	return
}

func StopTestEnvs(testEnvs []*envtest.Environment) {
	By("Stopping Kubernetes API Servers")
	for _, testEnv := range testEnvs {
		Eventually(func() error {
			return testEnv.Stop()
		}).ShouldNot(HaveOccurred())
	}
}

func CreateUnstructuredHarvesterSettingObject() (unstructuredHarvesterSettingObj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	// GroupVersionResource for the Harvester 'Settings' Custom Resource
	gvr = schema.GroupVersionResource{
		Group:    "harvesterhci.io",
		Version:  "v1beta1",
		Resource: "settings",
	}
	harvesterSettingjsonString := `{
        "apiVersion": "harvesterhci.io/v1beta1",
        "default": "{\"cpu\":1600,\"memory\":150,\"storage\":200}",
        "kind": "Setting",
        "metadata": {
            "name": "overcommit-config"
        },
        "value": "{\"cpu\":100,\"memory\":100,\"storage\":100}"
    }`

	// Convert JSON string to a map
	var harvesterSettingMap map[string]interface{}
	err := json.Unmarshal([]byte(harvesterSettingjsonString), &harvesterSettingMap)
	Expect(err).NotTo(HaveOccurred())

	// Create an Unstructured object from the map
	unstructuredHarvesterSettingObj = &unstructured.Unstructured{Object: harvesterSettingMap}
	unstructuredHarvesterSettingObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "harvesterhci.io",
		Version: "v1beta1",
		Kind:    "Setting",
	})
	return
}
