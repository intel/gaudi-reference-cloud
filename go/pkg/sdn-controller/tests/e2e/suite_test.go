//go:build ginkgo_only

// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package e2e

import (
	"context"
	"fmt"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/helper"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/e2e/frontendonly"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/e2e/frontendonly-tiny"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/e2e/lacp"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/e2e/netbox"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/e2e/singleplyspineleaf"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/e2e/smallaccl2"
	testutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/test_utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	bmhns  = "../config/bmh-ns.yaml"
	bmhcrd = "../config/bmh-crd.yaml"
	scheme = runtime.NewScheme()
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SDN Test Suite", Label("sdn"))
}

var _ = BeforeSuite(func() {
	log.SetDefaultLogger()
	ctx := context.Background()
	logger := log.FromContext(ctx)
	var err error
	// TODO: to be implemented in the future:
	//	prepare the env for all the test cases. For instance, start a kind cluster, start all the necessary service or create the resources.

	logger.Info("Deleting all existing networking/bmh CRs...")
	utilruntime.Must(idcnetworkv1alpha1.AddToScheme(scheme))
	utilruntime.Must(baremetalv1alpha1.AddToScheme(scheme))
	k8sClient := utils.NewK8SClientFromConfAndScheme(ctx, "", scheme)
	sdnTestHelperWithoutTopo := helper.SDNTestHelper{
		K8sClient: k8sClient,
	}
	err = sdnTestHelperWithoutTopo.DeleteAllK8sResources()
	if err != nil {
		logger.Error(err, "Failed to delete all k8s resources")
	}

	logger.Info("Deploying SDN Controller CRDs...")
	var command string

	command = fmt.Sprintf("cd ../../../../.. && make apply-only-sdn-controller-crds")
	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Println(output)
		fmt.Printf("RunCommand failed [%v] \n", command)
	}

	// we are running SDN and BMH in the same cluster, so create the BMH related resources below.
	// create the BMH ns and CRD
	logger.Info("Creating BMH CRD and NS...")
	command = fmt.Sprintf("kubectl apply -f %s", bmhns)
	_, err = testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("RunCommand failed [%v] \n", command)
	}
	command = fmt.Sprintf("kubectl apply -f %s", bmhcrd)
	_, err = testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("RunCommand failed [%v] \n", command)
	}
	logger.Info("finished creating BMH CRD and NS")

	logger.Info("Top Level BeforeSuite Completed")
})

var _ = AfterSuite(func() {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	// remove the BMH CRD and NS(optional)
	// command := fmt.Sprintf("kubectl delete -f %s", bmhns)
	// _, err := testutils.RunCommand(command)
	// if err != nil {
	// 	fmt.Printf("RunCommand failed [%v] \n", command)
	// }
	// command = fmt.Sprintf("kubectl delete -f %s", bmhcrd)
	// _, err = testutils.RunCommand(command)
	// if err != nil {
	// 	fmt.Printf("RunCommand failed [%v] \n", command)
	// }
	// logger.Info("finished deleting BMH CRD and NS")

	logger.Info("Top Level AfterSuite Completed")
})
