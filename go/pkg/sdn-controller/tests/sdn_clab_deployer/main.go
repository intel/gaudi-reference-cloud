package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/clab"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/helper"
	testutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/test_utils"
)

var (
	topologyDir          = "../../../../../networking/containerlab"
	topologyFrontendonly = "../../../../../networking/containerlab/frontendonly/frontendonly.clab.yml"
	topologyAllscfabrics = "../../../../../networking/containerlab/allscfabrics/allscfabrics.clab.yml"
	testHelper           *helper.SDNTestHelper
	topology             *clab.Topology
	err                  error
	bmhns                = "../config/bmh-ns.yaml"
	bmhcrd               = "../config/bmh-crd.yaml"
)

// deploy:
// go run main.go --action=deploy --topo="../../../../../networking/containerlab/allscfabrics/allscfabrics.clab.yml"
// destroy:
// go run main.go --action=destroy --topo="../../../../../networking/containerlab/allscfabrics/allscfabrics.clab.yml"
func main() {
	var action string
	var nwcpkubeconfig string
	var clabTopologyFile string
	ctx := context.Background()
	logger := log.FromContext(ctx)
	flag.StringVar(&action, "action", "deploy", "deploy/destroy")
	flag.StringVar(&nwcpkubeconfig, "nwcpkubeconfig", "", "kubeconfig file path")
	flag.StringVar(&clabTopologyFile, "topo", topologyFrontendonly, "containerlab topology file path")
	flag.Parse()

	testHelper = helper.New(helper.Config{
		TopologyDir:    topologyDir,
		NWCPKubeConfig: nwcpkubeconfig,
		EAPISecretDir:  "/vault/secrets/eapi",
	})

	// we are running SDN and BMH in the same cluster, so create the BMH related resources below.
	// create the BMH ns and CRD
	command := fmt.Sprintf("kubectl apply -f %s", bmhns)
	_, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("RunCommand failed [%v] \n", command)
	}
	command = fmt.Sprintf("kubectl apply -f %s", bmhcrd)
	_, err = testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("RunCommand failed [%v] \n", command)
	}

	if action == "deploy" {
		logger.Info("deploying containerlab and SDN K8s resources...")
		err := Deploy(clabTopologyFile)
		if err != nil {
			fmt.Printf("Deploy failed \n")
			return
		}
		logger.Info("================ successfully deployed containerlab and K8s resources ================")
	} else if action == "destroy" {
		logger.Info("destroy containerlab and SDN K8s resources...")
		err := Destroy(clabTopologyFile)
		if err != nil {
			fmt.Printf("Destroy failed \n")
			return
		}
		logger.Info("================ successfully destroyed containerlab and K8s resources ================")
	}
}

func Deploy(topologyFile string) error {
	start := time.Now()
	err = testHelper.ContainerLabManager.Deploy(topologyFile)
	if err != nil {
		fmt.Printf("deploy failed, %v", err)
		return err
	}
	topology, err = testHelper.ContainerLabManager.Connect(topologyFile)
	if err != nil {
		fmt.Printf("Connect failed, %v", err)
		return err
	}

	err = testHelper.CreateK8sResourcesForTopology(topology)
	if err != nil {
		fmt.Printf("CreateK8sResourcesForTopology failed, %v", err)
		return err
	}
	fmt.Printf("took [%v] for the Deploy \n", time.Since(start))
	return nil
}

func Destroy(topologyFile string) error {
	err = testHelper.ContainerLabManager.Destroy(topologyFile)
	if err != nil {
		fmt.Printf("Destroy failed, %v", err)
		return err
	}

	if topology == nil {
		topology, err = testHelper.ContainerLabManager.ReadTopology(topologyFile)
		if err != nil {
			return fmt.Errorf("ReadTopology failed, %v", err)
		}
	}

	err = testHelper.DeleteK8sResourcesForClabTopology(topology)
	if err != nil {
		fmt.Printf("DeleteK8sResourcesForClabTopology failed, %v", err)
		return err
	}

	err = testutils.DeleteClabTmpFolder(".")
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		fmt.Printf("DeleteClabTmpFolder failed: %v\n", err)
	}
	return nil
}
