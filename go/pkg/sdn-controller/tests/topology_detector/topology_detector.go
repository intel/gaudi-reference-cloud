package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(idcnetworkv1alpha1.AddToScheme(scheme))
}

// go run topology_detector.go --nwcpKubeconfig=/home/jzhen/.kube/config.phx04-k01-nwcp
func main() {
	var nwcpKubeconfig string
	var intervalSec int
	flag.StringVar(&nwcpKubeconfig, "nwcpKubeconfig", "~/.kube/config", "nwcp kubeconfig file path")
	flag.IntVar(&intervalSec, "interval", 3600, "refresh interval")

	flag.Parse()
	var err error
	ctx := context.Background()

	if len(nwcpKubeconfig) == 0 {
		fmt.Printf("nwcpKubeconfig is not provided\n")
		os.Exit(1)
	}
	fmt.Printf("nwcpKubeconfig: [%v]\n", nwcpKubeconfig)

	k8sClient := utils.NewK8SClientFromConfAndScheme(ctx, nwcpKubeconfig, scheme)
	allNetworkNodeCRs := &idcnetworkv1alpha1.NetworkNodeList{}
	err = k8sClient.List(ctx, allNetworkNodeCRs)
	if err != nil {
		fmt.Printf("k8sClient.List failed, %v \n", err)
	}
	if len(allNetworkNodeCRs.Items) == 0 {
		fmt.Printf("no NetworkNodeCRs found\n")
		return
	}
	fmt.Printf("# allNetworkNodeCRs: [%v] \n", len(allNetworkNodeCRs.Items))

	m := make(map[string][]Node)
	for _, nn := range allNetworkNodeCRs.Items {
		accelSwitches := make([]string, 0)
		if nn.Spec.AcceleratorFabric != nil {
			hash := make(map[string]struct{})
			node := Node{
				NodeName: nn.Name,
			}
			accelPorts := nn.Spec.AcceleratorFabric.SwitchPorts
			for _, p := range accelPorts {
				idx := strings.Index(p, ".")
				if idx >= 0 {
					switchFQDN := p[idx+1:]
					if _, found := hash[switchFQDN]; !found {
						accelSwitches = append(accelSwitches, switchFQDN)
						hash[switchFQDN] = struct{}{}
					}
				}
			}
			// get the key
			sort.Strings(accelSwitches)
			var key string
			for _, s := range accelSwitches {
				key = key + "|" + s
			}
			// node.SwitchFQDNs = accelSwitches

			// add to the map
			m[key] = append(m[key], node)
		}
	}

	for k, v := range m {
		fmt.Printf("=================================================  \n")
		switches := strings.Split(k, "|")
		fmt.Printf("accelerator switches: \n")
		for _, sw := range switches {
			fmt.Printf("\t %v \n", sw)
		}
		fmt.Printf("nodes(%v) \n", len(v))
		for _, node := range v {
			fmt.Printf("\t %v \n", node)
		}
	}
}

type Node struct {
	NodeName string
	// SwitchFQDNs []string
}
