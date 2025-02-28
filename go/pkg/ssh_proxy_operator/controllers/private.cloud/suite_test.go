// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package privatecloud

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	idcclientset "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/clientset/versioned"
	idcinformerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/informers/externalversions"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg                      *rest.Config
	k8sClient                client.Client
	testEnv                  *envtest.Environment
	ctx                      context.Context
	cancel                   context.CancelFunc
	timeout                  time.Duration = 45 * time.Second
	poll                     time.Duration = 5 * time.Second
	ctrlConfig               cloudv1alpha1.SshProxyOperatorConfig
	authorizedKeysFilePath   string
	proxyUser                string = "guest"
	proxyAddress             string = "ssh.us-dev-1.cloud.intel.com"
	proxyPort                int    = 22
	authorizedKeysScpTargets []string
	publicKey                string
	privateKey               string
	scpTarget                string = "scp://guest@127.0.0.1:22/home/guest/.ssh/authorized_keys"
	scpTarget2               string = "scp://guest2@127.0.0.2:22/home/guest2/.ssh/authorized_keys"
	sshProxyController       *SshProxyController
	stopChannel              = make(chan struct{})
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseFlagOptions(&opts)))
	ctx, cancel = context.WithCancel(context.Background())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		// When adding CRDS, be sure to add them to the data list in BUILD.bazel.
		CRDDirectoryPaths: []string{
			"../../../k8s/config/crd/bases",
		},
		ErrorIfCRDPathMissing:    true,
		AttachControlPlaneOutput: true,
	}

	// cfg is defined in this file globally.
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = cloudv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = kubevirtv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	configFile := "../../testdata/operatorconfig.yaml"
	logf.Log.Info("CtrlTest", "configFile", configFile)

	ctrlConfig = cloudv1alpha1.SshProxyOperatorConfig{}
	options := ctrl.Options{Scheme: scheme.Scheme}
	if configFile != "" {
		var err error
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))
		if err != nil {
			panic(err)
		}
	}

	httpClient, err := rest.HTTPClientFor(cfg)
	httpClient.Timeout = 5 * time.Second
	if err != nil {
		panic(err)
	}
	logf.Log.Info("CtrlTest", "httpClient", httpClient)

	tempDir, err := os.MkdirTemp("", "sshproxytunnel_controller_test_")
	Expect(err).ToNot(HaveOccurred())
	authorizedKeysFilePath = fmt.Sprintf("%s/.ssh/authorized_keys", tempDir)
	// Passing a fake scpTarget List so as to avoid running SCP/SSH code for transferring the file to proxy server
	authorizedKeysScpTargets = []string{scpTarget}

	publicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC5GVA77uCixTzSZSfSX3dBlD9Dkg9ypbvzLefB/kxWK9BX idcuser@example.com\n"
	privateKey = `-----BEGIN OPENSSH PRIVATE KEY-----
							    b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAACFwAAAAdzc2gtcn
								NhAAAAAwEAAQAAAgEAuRlQO+7gosU80mUn0l93QZQ/Q5IPcqW78y3nwf5MVivQVxzYd/45
								trQQb90kveyWZTT3sfFmbC9Umt3+j5+D6L2bxy9mnos0Q8sVSsGFkCBW4m0HV+GrPX0BLA
								QWKz8lmR8HPNn9u56LwOGIKvWbaMFVr++SjgejX55gb6q5rNymNF80x8dz4Y2zZnAPash+
								8puxS5V1Snxrm2wMCpQq0+mc0LF6KJFZQljyHuhg4Yydrvuhi8IjatFHaqN11R/2ddD0D0
								+V/WH6wY143S3oGD64g0j/jajDUcP2urkYxquwqdSfFWzlVZKZEjotMDveGOmAgZc5mOGq
								sCY644gB/60IV/Wg+HVn4A+8f4DhXjFQmuLwdkL1HLbE0MIJ5JrWWu91i6txkN/QdCF337
								/4LuD5deSLmTBPCJ/XHprDNoE6VVnTtAbeqkmVCwacjGuepJmAc2Uci/OHhlT+X0pZ4+wJ
								/IfRcmifKo9Ha6/j595Mwyp8JUN+9QUzWWXvvf0zlx+jAzmZGif/x70CzDPi5XSuUCfgLh
								faLZ1nn8e70qQUoRy0SITbHVRaCXWj1GMZqWntsQjdLA3o/hIEv/T+cdPHy3AGESDmJ0/d
								-----END OPENSSH PRIVATE KEY-----\n`

	sshProxyConfig := SshProxyTunnelConfig{
		AuthorizedKeysFilePath:   authorizedKeysFilePath,
		ProxyUser:                proxyUser,
		ProxyAddress:             proxyAddress,
		ProxyPort:                proxyPort,
		AuthorizedKeysScpTargets: authorizedKeysScpTargets,
		PublicKey:                publicKey,
		PrivateKey:               privateKey,
	}

	kubeClientSet, err := kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(kubeClientSet).NotTo(BeNil())

	idcClientSet, err := idcclientset.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(kubeClientSet).NotTo(BeNil())

	informerFactory := idcinformerfactory.NewSharedInformerFactory(idcClientSet, 10*time.Minute)

	sshProxyController, err = NewSshProxyController(ctx, kubeClientSet, idcClientSet, informerFactory.Private().V1alpha1().SshProxyTunnels(), sshProxyConfig)
	Expect(err).NotTo(HaveOccurred())

	sshProxyController.MockScpTargetsMutex.Lock()
	sshProxyController.MockScpTargets[scpTarget] = nil
	sshProxyController.MockScpTargetsMutex.Unlock()
	stopCh := make(chan struct{})

	informerFactory.Start(stopCh)

	err = k8sManager.Add(manager.RunnableFunc(func(context.Context) error {
		return sshProxyController.Run(ctx, stopCh)
	}))
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	close(stopChannel)
	Eventually(func() error {
		return testEnv.Stop()
	}, timeout, poll).ShouldNot(HaveOccurred())
})
