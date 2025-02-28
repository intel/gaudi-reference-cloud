// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	toolsk8s "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var _ = Describe("KubeConfigFromRESTConfig Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("KubeConfigFromRESTConfig Tests")
	_ = log

	It("KubeConfigFromRESTConfig Test", func() {
		By("Starting Kubernetes API Server")
		testEnv := &envtest.Environment{}
		var err error
		restConfig1, err := testEnv.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(restConfig1).NotTo(BeNil())

		By("Creating Kubernetes client with testenv rest.Config")
		clientset1, err := kubernetes.NewForConfig(restConfig1)
		Expect(err).NotTo(HaveOccurred())

		By("Listing Namespaces")
		_, err = clientset1.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Creating KubeConfig file contents")
		kubeConfigBytes, err := toolsk8s.KubeConfigFromRESTConfig(restConfig1)
		Expect(err).NotTo(HaveOccurred())

		By("Creating rest.Config from KubeConfig")
		restConfig2, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigBytes)
		Expect(err).NotTo(HaveOccurred())

		By("Creating Kubernetes client with new rest.Config")
		clientset2, err := kubernetes.NewForConfig(restConfig2)
		Expect(err).NotTo(HaveOccurred())

		By("Listing Namespaces")
		_, err = clientset2.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Stopping Kubernetes API Server")
		Eventually(func() error {
			return testEnv.Stop()
		}).ShouldNot(HaveOccurred())

	})
})

var _ = Describe("LoadKubeConfigFiles Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("LoadKubeConfigFiles Tests")
	_ = log

	It("LoadKubeConfigFiles Test", func() {
		numTestEnvs := 2
		testEnvs, _, tempDir := CreateTestEnvs(numTestEnvs, []string{})

		By("LoadKubeConfigFiles")
		restConfigs, err := toolsk8s.LoadKubeConfigFiles(ctx, os.DirFS(tempDir), "*.hconf")
		Expect(err).ToNot(HaveOccurred())
		Expect(len(restConfigs)).Should(Equal(numTestEnvs))

		for filename, restConfig := range restConfigs {
			By("Creating Kubernetes client with rest.Config for " + filename)
			clientset, err := kubernetes.NewForConfig(restConfig)
			Expect(err).NotTo(HaveOccurred())
			By("Listing Namespaces for " + filename)
			_, err = clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
		}

		StopTestEnvs(testEnvs)
	})
})
