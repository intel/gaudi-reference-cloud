// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	switchclients "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()

	provisioningFEVlanId  int64 = 4008
	provisioningAccVlanId int64 = 100
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(idcnetworkv1alpha1.AddToScheme(scheme))
	utilruntime.Must(baremetalv1alpha1.AddToScheme(scheme))
	utilruntime.Must(cloudv1alpha1.AddToScheme(scheme))

}

// Example of running the tool:
// go run dataintegritycheck.go --kubeconfig="config.idc-staging-nwcp"
func main() {
	//var nwcpKubeconfig string
	var bmhKubeconfig string
	var ravenServer string
	var ravenSecretPath string
	var ravenEnv string
	var datacenter string
	var eapiSecretPath string
	var intervalSec int
	//flag.StringVar(&nwcpKubeconfig, "nwcpKubeconfig", "/", "nwcp kubeconfig file path")
	flag.StringVar(&bmhKubeconfig, "bmhKubeconfig", "", "bmh kubeconfig file path")
	flag.StringVar(&ravenServer, "raven", "raven-devcloud.app.intel.com", "Raven server")
	flag.StringVar(&ravenSecretPath, "ravenSecretPath", "/vault/secrets/raven", "Raven secret file path")
	flag.StringVar(&ravenEnv, "ravenEnv", "", "Raven environment")
	flag.StringVar(&datacenter, "datacenter", "fxhb3p3p:fxhb3p3s:fxhb3p3r:azs1101pe:azs1102pe:azs1201pe:ech2101pe:azs1401pe", "data center")
	flag.StringVar(&eapiSecretPath, "eapiSecretPath", "", "Eapi secret file path")
	flag.IntVar(&intervalSec, "interval", 3600, "refresh interval")

	flag.Parse()
	ctx := context.Background()
	log.SetDefaultLogger()

	fmt.Printf("Connecting to nwcp k8s... \n")
	// If --kubeconfig is set, will use the kubeconfig file at that location.  Otherwise will assume running
	// in cluster and use the cluster provided kubeconfig.
	nwcpK8sClient := utils.NewK8SClientWithScheme(scheme)

	var bmhK8sClient client.Client
	if len(bmhKubeconfig) != 0 {
		fmt.Printf("Connecting to bmh k8s... \n")
		bmhK8sClient = utils.NewK8SClientFromConfAndScheme(ctx, bmhKubeconfig, scheme)
	}

	fmt.Printf("Starting ticker... \n")
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	for {
		allNetworkNodeCRs, err := getAllNetworkNodes(ctx, nwcpK8sClient)
		if err != nil {
			fmt.Printf("getAllNetworkNodes failed, %v \n", err)
		}

		allSwitchPortCRs, err := GetAllSwitchPorts(ctx, nwcpK8sClient)
		if err != nil {
			fmt.Printf("GetAllSwitchPorts failed, %v \n", err)
		}

		allSwitchCRs, err := GetAllSwitches(ctx, nwcpK8sClient)
		if err != nil {
			fmt.Printf("GetAllSwitches failed, %v \n", err)
		}

		var allBMHCRs *baremetalv1alpha1.BareMetalHostList
		var allInstanceCRs *cloudv1alpha1.InstanceList
		if len(bmhKubeconfig) != 0 && bmhK8sClient != nil {
			allBMHCRs, err = GetAllBMHs(ctx, bmhK8sClient)
			if err != nil {
				fmt.Printf("GetAllBMHs failed, %v \n", err)
			}

			allInstanceCRs, err = GetAllInstances(ctx, bmhK8sClient)
			if err != nil {
				fmt.Printf("GetAllInstances failed, %v \n", err)
			}
		}

		checkNetworkNodesForDuplicateSPs(allNetworkNodeCRs)
		checkSwitchPortsOwnerRefs(allSwitchPortCRs)
		checkReciprocalRefs(allNetworkNodeCRs, allSwitchPortCRs)
		checkForEmptyVlanId(allNetworkNodeCRs, allSwitchPortCRs)
		checkSPStatusMatchesSpec(allSwitchPortCRs)
		checkNNBothFabricsProvisioning(allNetworkNodeCRs)
		if allBMHCRs != nil {
			checkBMHsMatchNNs(allNetworkNodeCRs, allBMHCRs)
			checkAvailableBMHHasProvisioningVlan(allNetworkNodeCRs, allBMHCRs)
			if allInstanceCRs != nil {
				checkNNVlanMatchesInstance(allNetworkNodeCRs, allBMHCRs, allInstanceCRs)
			}
		}
		if eapiSecretPath != "" {
			checkSwitchConnectivityViaEapi(ctx, eapiSecretPath, allSwitchCRs)
			checkSwitchportConfigViaEapi(ctx, eapiSecretPath, allSwitchCRs, allSwitchPortCRs)
		}

		select {
		case <-ticker.C:
			continue
		}
	}
}

func getAllNetworkNodes(ctx context.Context, k8sClient client.Client) (*idcnetworkv1alpha1.NetworkNodeList, error) {
	var err error
	fmt.Printf("start getting NetworkNodes... \n")
	msg := fmt.Sprintf("timestamp: %v \n", time.Now().UTC())
	fmt.Printf(msg)
	allNetworkNodeCRs := &idcnetworkv1alpha1.NetworkNodeList{}
	err = k8sClient.List(ctx, allNetworkNodeCRs)
	if err != nil {
		return allNetworkNodeCRs, err
	}
	if len(allNetworkNodeCRs.Items) == 0 {
		fmt.Printf("no NetworkNodeCRs found\n")
	}
	fmt.Printf("# allNetworkNodeCRs: [%v] \n", len(allNetworkNodeCRs.Items))
	return allNetworkNodeCRs, nil
}

func GetAllSwitchPorts(ctx context.Context, k8sClient client.Client) (*idcnetworkv1alpha1.SwitchPortList, error) {
	var err error
	fmt.Printf("start getting switchports... \n")
	msg := fmt.Sprintf("timestamp: %v \n", time.Now().UTC())
	fmt.Printf(msg)
	allSwitchPortCRs := &idcnetworkv1alpha1.SwitchPortList{}
	err = k8sClient.List(ctx, allSwitchPortCRs)
	if err != nil {
		return allSwitchPortCRs, err
	}
	if len(allSwitchPortCRs.Items) == 0 {
		fmt.Printf("no SwitchPortCRs found\n")
	}
	fmt.Printf("# allSwitchPortCRs: [%v] \n", len(allSwitchPortCRs.Items))
	return allSwitchPortCRs, nil
}

func GetAllSwitches(ctx context.Context, k8sClient client.Client) (*idcnetworkv1alpha1.SwitchList, error) {
	var err error
	fmt.Printf("start getting switches... \n")
	msg := fmt.Sprintf("timestamp: %v \n", time.Now().UTC())
	fmt.Printf(msg)
	allSwitchCRs := &idcnetworkv1alpha1.SwitchList{}
	err = k8sClient.List(ctx, allSwitchCRs)
	if err != nil {
		return allSwitchCRs, err
	}
	if len(allSwitchCRs.Items) == 0 {
		fmt.Printf("no SwitchCRs found\n")
	}
	fmt.Printf("# allSwitchCRs: [%v] \n", len(allSwitchCRs.Items))
	return allSwitchCRs, nil
}

func GetAllBMHs(ctx context.Context, k8sClient client.Client) (*baremetalv1alpha1.BareMetalHostList, error) {
	var err error
	fmt.Printf("start getting BMHs... \n")
	msg := fmt.Sprintf("timestamp: %v \n", time.Now().UTC())
	fmt.Printf(msg)
	allBMHCRs := &baremetalv1alpha1.BareMetalHostList{}
	err = k8sClient.List(ctx, allBMHCRs)
	if err != nil {
		return allBMHCRs, err
	}
	if len(allBMHCRs.Items) == 0 {
		fmt.Printf("no BMH CRs found\n")
	}
	fmt.Printf("# allBMHCRs: [%v] \n", len(allBMHCRs.Items))
	return allBMHCRs, nil
}

func GetAllInstances(ctx context.Context, k8sClient client.Client) (*cloudv1alpha1.InstanceList, error) {
	var err error
	fmt.Printf("start getting Instances... \n")
	msg := fmt.Sprintf("timestamp: %v \n", time.Now().UTC())
	fmt.Printf(msg)
	allInstanceCRs := &cloudv1alpha1.InstanceList{}
	err = k8sClient.List(ctx, allInstanceCRs)
	if err != nil {
		return allInstanceCRs, err
	}
	if len(allInstanceCRs.Items) == 0 {
		fmt.Printf("no Instance CRs found\n")
	}
	fmt.Printf("# allInstanceCRs: [%v] \n", len(allInstanceCRs.Items))
	return allInstanceCRs, nil
}

func checkNetworkNodesForDuplicateSPs(allNetworkNodeCRs *idcnetworkv1alpha1.NetworkNodeList) {

	// Check for duplicate owners (2x networknodes that claim to be connected to the same switchport)
	// This can cause the VLAN to switch back & forth between them.

	fmt.Printf("start checking for networknodes that both refer to the same switchport... \n")
	switchPortOwners := make(map[string]string, 0)
	for _, nnCR := range allNetworkNodeCRs.Items {
		sp := nnCR.Spec.FrontEndFabric.SwitchPort
		if switchPortOwners[sp] != "" {
			msg := fmt.Sprintf("both %s and %s claim to own switchport %s. \n", switchPortOwners[sp], nnCR.Name, sp)
			fmt.Printf(msg)
		} else {
			switchPortOwners[sp] = nnCR.Name
		}

		if nnCR.Spec.AcceleratorFabric != nil {
			for _, sp := range nnCR.Spec.AcceleratorFabric.SwitchPorts {
				if switchPortOwners[sp] != "" {
					msg := fmt.Sprintf("both %s and %s claim to own switchport %s. \n", switchPortOwners[sp], nnCR.Name, sp)
					fmt.Printf(msg)
				} else {
					switchPortOwners[sp] = nnCR.Name
				}
			}
		}
	}

	fmt.Printf("finished checking NetworkNodes. \n")
}

func checkSwitchPortsOwnerRefs(allSwitchPortCRs *idcnetworkv1alpha1.SwitchPortList) {

	// Check that a switchport's "SwitchPort" label matches its ownerReference.
	// because we found some issues where label was being updated, but owner was not.
	fmt.Printf("start checking switchport ownerReferences match labels... \n")

	for _, spCR := range allSwitchPortCRs.Items {
		if len(spCR.OwnerReferences) != 1 {
			msg := fmt.Sprintf("switchport %s doesn't have exactly 1 ownerReference. It has %d. \n", spCR.Name, len(spCR.OwnerReferences))
			fmt.Printf(msg)
		} else {
			if spCR.OwnerReferences[0].Name != spCR.Labels[idcnetworkv1alpha1.LabelNameNetworkNode] {
				msg := fmt.Sprintf("switchport %s has owner name %s but networkNode label %s. \n", spCR.Name, spCR.OwnerReferences[0].Name, spCR.Labels[idcnetworkv1alpha1.LabelNameNetworkNode])
				fmt.Printf(msg)
			}
		}
	}

	fmt.Printf("finished checking switchports have different owners from label. \n")
}

func checkReciprocalRefs(allNetworkNodeCRs *idcnetworkv1alpha1.NetworkNodeList, allSwitchPortCRs *idcnetworkv1alpha1.SwitchPortList) {

	fmt.Printf("start checking reciprocal ownerReferences... \n")

	// Look for switchPorts that refer to NetworkNodes that do not exist, or do not refer to them.
	for _, spCR := range allSwitchPortCRs.Items {

		if len(spCR.OwnerReferences) != 1 {
			msg := fmt.Sprintf("switchport %s doesn't have exactly 1 ownerReference. It has %d. \n", spCR.Name, len(spCR.OwnerReferences))
			fmt.Printf(msg)
			continue
		}

		// Find the NetworkNode that the switchPort refers to.
		foundNNs := 0
		var parentNN idcnetworkv1alpha1.NetworkNode
		for _, nnCR := range allNetworkNodeCRs.Items { // Optimization: Use a map, keyed by UID or name rather than nested loop
			if nnCR.UID == spCR.OwnerReferences[0].UID {
				foundNNs++
				parentNN = nnCR
			}
		}
		if foundNNs == 0 {
			msg := fmt.Sprintf("could not find NN with UID %s / name %s (which is supposed to be the owner of SP %s) \n", spCR.OwnerReferences[0].UID, spCR.OwnerReferences[0].Name, spCR.Name)
			fmt.Printf(msg)
		} else if foundNNs > 1 {
			msg := fmt.Sprintf("found more than 1 NN with UID %s / name %s (owner of SP %s) \n", spCR.OwnerReferences[0].UID, spCR.OwnerReferences[0].Name, spCR.Name)
			fmt.Printf(msg)
		} else {
			foundSPinNN := false
			// Look for the SwitchPort in the NN's list of switchPorts.
			if parentNN.Spec.FrontEndFabric.SwitchPort == spCR.Name {
				foundSPinNN = true
			}
			if parentNN.Spec.AcceleratorFabric != nil {
				for _, accSP := range parentNN.Spec.AcceleratorFabric.SwitchPorts {
					if accSP == spCR.Name {
						foundSPinNN = true
					}
				}
				if !foundSPinNN {
					msg := fmt.Sprintf("SP %s has owner %s, but %s doesn't contain %s in its list of switchports \n", spCR.Name, parentNN.Name, parentNN.Name, spCR.Name)
					fmt.Printf(msg)
				}
			}
		}
	}

	// Look for networkNodes that refer to switchPorts that do not refer back to them.
	for _, nnCR := range allNetworkNodeCRs.Items {
		// Find the FE SwitchPort that the networkNode refers to.
		foundSPs := 0
		var childSP idcnetworkv1alpha1.SwitchPort
		for _, spCR := range allSwitchPortCRs.Items { // Optimization: Use a map, keyed by UID or name rather than nested loop
			if spCR.Name == nnCR.Spec.FrontEndFabric.SwitchPort {
				foundSPs++
				childSP = spCR
			}
		}
		if foundSPs == 0 {
			msg := fmt.Sprintf("could not find SP with name %s (which is supposed to be FE child of NN %s) \n", nnCR.Spec.FrontEndFabric.SwitchPort, nnCR.Name)
			fmt.Printf(msg)
		} else if foundSPs > 1 {
			msg := fmt.Sprintf("found more than one SP with name %s (FE child of NN %s) \n", nnCR.Spec.FrontEndFabric.SwitchPort, nnCR.Name)
			fmt.Printf(msg)
		} else {
			if len(childSP.OwnerReferences) == 0 || childSP.OwnerReferences[0].Name != nnCR.Name {
				msg := fmt.Sprintf("NN %s has FE child %s, but %s doesn't have %s as its owner \n", nnCR.Name, childSP.Name, childSP.Name, nnCR.Name)
				fmt.Printf(msg)
			}
		}

		// Find the Accellerator SwitchPorts that the networkNode refers to.
		if nnCR.Spec.AcceleratorFabric != nil {
			for _, accSPName := range nnCR.Spec.AcceleratorFabric.SwitchPorts {
				foundSPs := 0
				var childSP idcnetworkv1alpha1.SwitchPort
				for _, spCR := range allSwitchPortCRs.Items { // Optimization: Use a map, keyed by UID or name rather than nested loop
					if spCR.Name == accSPName {
						foundSPs++
						childSP = spCR
					}
				}
				if foundSPs == 0 {
					msg := fmt.Sprintf("could not find SP with name %s (which is supposed to be accellerator child of NN %s) \n", accSPName, nnCR.Name)
					fmt.Printf(msg)
				} else if foundSPs > 1 {
					msg := fmt.Sprintf("found more than one SP with name %s (accellerator child of NN %s) \n", accSPName, nnCR.Name)
					fmt.Printf(msg)
				} else {
					if childSP.OwnerReferences[0].Name != nnCR.Name {
						msg := fmt.Sprintf("NN %s has accellerator child %s, but %s doesn't have %s as its owner \n", nnCR.Name, childSP.Name, childSP.Name, nnCR.Name)
						fmt.Printf(msg)
					}
				}
			}
		}
	}

	fmt.Printf("finished checking reciprocal ownerRefences. \n")
}

func checkForEmptyVlanId(allNetworkNodeCRs *idcnetworkv1alpha1.NetworkNodeList, allSwitchPortCRs *idcnetworkv1alpha1.SwitchPortList) {
	fmt.Printf("started checking for empty VlanIds... \n")

	for _, nnCR := range allNetworkNodeCRs.Items {
		if nnCR.Spec.FrontEndFabric.VlanId == 0 {
			msg := fmt.Sprintf("NN %s has no frontend VlanId set \n", nnCR.Name)
			fmt.Printf(msg)
		}
		if nnCR.Spec.AcceleratorFabric != nil && len(nnCR.Spec.AcceleratorFabric.SwitchPorts) > 0 && nnCR.Spec.AcceleratorFabric.VlanId == 0 {
			msg := fmt.Sprintf("NN %s has accellerator fabric, but no VlanId set \n", nnCR.Name)
			fmt.Printf(msg)
		}
	}

	for _, spCR := range allSwitchPortCRs.Items {
		if spCR.Spec.VlanId == 0 {
			msg := fmt.Sprintf("SP %s has no VlanId set \n", spCR.Name)
			fmt.Printf(msg)
		}
	}

	fmt.Printf("finished checking for empty VlanIds. \n")
}

func checkSPStatusMatchesSpec(allSwitchPortCRs *idcnetworkv1alpha1.SwitchPortList) {
	fmt.Printf("starting checkSPStatusMatchesSpec... \n")

	for _, spCR := range allSwitchPortCRs.Items {
		if spCR.Spec.VlanId != -1 && spCR.Spec.VlanId != spCR.Status.VlanId {
			msg := fmt.Sprintf("SP %s Spec.Vlan %d doesn't match Status.Vlan %d \n", spCR.Name, spCR.Spec.VlanId, spCR.Status.VlanId)
			fmt.Printf(msg)
		}

		//if spCR.Spec.VlanId != -1 && spCR.Spec.VlanId != spCR.Status.RavenDBVlanId {
		//	msg := fmt.Sprintf("SP %s Spec.Vlan %d doesn't match Status.RavenDBVlan %d \n", spCR.Name, spCR.Spec.VlanId, spCR.Status.RavenDBVlanId)
		//	fmt.Printf(msg)
		//}
	}

	fmt.Printf("finished checkSPStatusMatchesSpec. \n")
}

func checkBMHsMatchNNs(allNetworkNodeCRs *idcnetworkv1alpha1.NetworkNodeList, allBMHCRs *baremetalv1alpha1.BareMetalHostList) {
	fmt.Printf("started checking BMHs match NetworkNodes... \n")

	if len(allNetworkNodeCRs.Items) != len(allBMHCRs.Items) {
		msg := fmt.Sprintf("There are %d BMH CRs, but %d NetworkNode CRs \n", len(allBMHCRs.Items), len(allNetworkNodeCRs.Items))
		fmt.Printf(msg)
	}

	for _, bmhCR := range allBMHCRs.Items {
		foundNNs := 0
		for _, nnCR := range allNetworkNodeCRs.Items { // Optimization: Use a map, keyed by UID or name rather than nested loop
			if nnCR.Name == bmhCR.Name {
				foundNNs++
			}
		}
		if foundNNs == 0 {
			msg := fmt.Sprintf("could not find NN with same name as BMH %s \n", bmhCR.Name)
			fmt.Printf(msg)
		} else if foundNNs > 1 {
			msg := fmt.Sprintf("Found more than one NN with same name as BMH %s \n", bmhCR.Name)
			fmt.Printf(msg)
		}
	}

	for _, nnCR := range allNetworkNodeCRs.Items {
		foundBMHs := 0
		for _, bmhCR := range allBMHCRs.Items { // Optimization: Use a map, keyed by UID or name rather than nested loop
			if bmhCR.Name == nnCR.Name {
				foundBMHs++
			}
		}
		if foundBMHs == 0 {
			msg := fmt.Sprintf("could not find BMH with same name as NN %s \n", nnCR.Name)
			fmt.Printf(msg)
		} else if foundBMHs > 1 {
			msg := fmt.Sprintf("Found more than one BMH with same name as NN %s \n", nnCR.Name)
			fmt.Printf(msg)
		}
	}

	fmt.Printf("finished checking BMHs match NetworkNodes. \n")
}

func checkNNVlanMatchesInstance(allNetworkNodeCRs *idcnetworkv1alpha1.NetworkNodeList, allBMHCRs *baremetalv1alpha1.BareMetalHostList, allInstanceCRs *cloudv1alpha1.InstanceList) {
	fmt.Printf("checking NN.VlanId matches the Instance.VlanId assigned to the BMH... \n")

	for _, nnCR := range allNetworkNodeCRs.Items {
		var matchingBMH baremetalv1alpha1.BareMetalHost
		foundBMHs := 0
		for _, bmhCR := range allBMHCRs.Items { // Optimization: Use a map, keyed by UID or name rather than nested loop
			if bmhCR.Name == nnCR.Name {
				matchingBMH = bmhCR
				foundBMHs++
			}
		}

		if foundBMHs == 1 && matchingBMH.Spec.ConsumerRef != nil {
			var matchingInstance cloudv1alpha1.Instance
			foundInstances := 0
			for _, instanceCR := range allInstanceCRs.Items { // Optimization: Use a map, keyed by UID or name rather than nested loop
				if instanceCR.Name == matchingBMH.Spec.ConsumerRef.Name {
					foundInstances++
					matchingInstance = instanceCR
				}
			}

			if foundInstances != 1 {
				msg := fmt.Sprintf("could not find Instance %s corresponding to BMH %s \n", matchingBMH.Spec.ConsumerRef.Name, matchingBMH.Name)
				fmt.Printf(msg)
			} else {
				foundIfaceWithFEVlan := false
				foundIfaceWithAccVlan := false
				for _, iface := range matchingInstance.Status.Interfaces {
					if iface.VlanId == int(nnCR.Spec.FrontEndFabric.VlanId) {
						foundIfaceWithFEVlan = true
					}
					if nnCR.Spec.AcceleratorFabric != nil && iface.VlanId == int(nnCR.Spec.AcceleratorFabric.VlanId) {
						foundIfaceWithAccVlan = true
					}
				}
				if !foundIfaceWithFEVlan && nnCR.Spec.FrontEndFabric.VlanId != -1 {
					msg := fmt.Sprintf("Instance %s did not have an interface with FE VlanID %d \n", matchingInstance.Name, nnCR.Spec.FrontEndFabric.VlanId)
					fmt.Printf(msg)
				}
				if nnCR.Spec.AcceleratorFabric != nil && !foundIfaceWithAccVlan && nnCR.Spec.AcceleratorFabric.VlanId != -1 {
					msg := fmt.Sprintf("Instance %s did not have an interface with Accellerator VlanID %d \n", matchingInstance.Name, nnCR.Spec.AcceleratorFabric.VlanId)
					fmt.Printf(msg)
				}

			}
		}
	}

	fmt.Printf("finished checking NN.VlanId matches the value on the Instance. \n")
}

func checkAvailableBMHHasProvisioningVlan(allNetworkNodeCRs *idcnetworkv1alpha1.NetworkNodeList, allBMHCRs *baremetalv1alpha1.BareMetalHostList) {
	fmt.Printf("checking available BMHs have provisioning VLAN... \n")
	for _, bmhCR := range allBMHCRs.Items {

		if bmhCR.Status.Provisioning.State != baremetalv1alpha1.StateAvailable {
			continue
		}

		foundNNs := 0
		var matchingNN idcnetworkv1alpha1.NetworkNode
		for _, nnCR := range allNetworkNodeCRs.Items { // Optimization: Use a map, keyed by UID or name rather than nested loop
			if nnCR.Name == bmhCR.Name {
				foundNNs++
				matchingNN = nnCR
			}
		}

		if foundNNs == 1 {
			if matchingNN.Spec.FrontEndFabric.VlanId != provisioningFEVlanId && matchingNN.Spec.FrontEndFabric.VlanId != -1 {
				msg := fmt.Sprintf("BMH %s is in available state, but NN %s has FE VlanId %d \n", bmhCR.Name, matchingNN.Name, matchingNN.Spec.FrontEndFabric.VlanId)
				fmt.Printf(msg)
			}
			if matchingNN.Spec.AcceleratorFabric != nil && matchingNN.Spec.AcceleratorFabric.VlanId != provisioningAccVlanId && matchingNN.Spec.AcceleratorFabric.VlanId != -1 {
				msg := fmt.Sprintf("BMH %s is in available state, but NN %s has Accellerator VlanId %d \n", bmhCR.Name, matchingNN.Name, matchingNN.Spec.AcceleratorFabric.VlanId)
				fmt.Printf(msg)
			}
		}
	}

	fmt.Printf("checking NNs with provisioning VLAN are in available state... \n")
	for _, nnCR := range allNetworkNodeCRs.Items {

		if nnCR.Spec.FrontEndFabric.VlanId != provisioningFEVlanId && (nnCR.Spec.AcceleratorFabric == nil || nnCR.Spec.AcceleratorFabric.VlanId != provisioningAccVlanId) {
			continue
		}

		foundBMHs := 0
		var matchingBMH baremetalv1alpha1.BareMetalHost
		for _, bmhCR := range allBMHCRs.Items { // Optimization: Use a map, keyed by UID or name rather than nested loop
			if bmhCR.Name == nnCR.Name {
				foundBMHs++
				matchingBMH = bmhCR
			}
		}

		if foundBMHs == 1 {
			if matchingBMH.Status.Provisioning.State != baremetalv1alpha1.StateAvailable && matchingBMH.Status.Provisioning.State != baremetalv1alpha1.StateProvisioning && matchingBMH.Status.Provisioning.State != baremetalv1alpha1.StateDeprovisioning {
				msg := fmt.Sprintf("NN %s has provisioning Vlan set, but BMH %s is in state %s. \n", nnCR.Name, matchingBMH.Name, matchingBMH.Status.Provisioning.State)
				fmt.Printf(msg)
			}
		}
	}

	fmt.Printf("finished checkAvailableBMHHasProvisioningVlan... \n")
}

// check that either BOTH fabrics are on the provisioning/default VLAN, or neither are.
func checkNNBothFabricsProvisioning(allNetworkNodeCRs *idcnetworkv1alpha1.NetworkNodeList) {
	fmt.Printf("starting checkNNBothFabricsProvisioning... \n")

	for _, nnCR := range allNetworkNodeCRs.Items {

		if nnCR.Spec.AcceleratorFabric == nil {
			continue
		}

		if nnCR.Spec.FrontEndFabric.VlanId == provisioningFEVlanId && nnCR.Spec.AcceleratorFabric.VlanId != provisioningAccVlanId && nnCR.Spec.AcceleratorFabric.VlanId != -1 {
			msg := fmt.Sprintf("NN %s has provisioning FE Vlan set (%d), but accellerator fabric is %d. \n", nnCR.Name, provisioningFEVlanId, nnCR.Spec.AcceleratorFabric.VlanId)
			fmt.Printf(msg)
		}

		if nnCR.Spec.AcceleratorFabric.VlanId == provisioningAccVlanId && nnCR.Spec.FrontEndFabric.VlanId != provisioningFEVlanId && nnCR.Spec.FrontEndFabric.VlanId != -1 {
			msg := fmt.Sprintf("NN %s has default Accelerator Vlan set (%d), but Frontend fabric is %d. \n", nnCR.Name, provisioningAccVlanId, nnCR.Spec.FrontEndFabric.VlanId)
			fmt.Printf(msg)
		}
	}

	fmt.Printf("finished checkNNBothFabricsProvisioning... \n")
}

func checkSwitchConnectivityViaEapi(ctx context.Context, switchSecretsPath string, allSwitchCRs *idcnetworkv1alpha1.SwitchList) {
	fmt.Printf("starting checkSwitchConnectivityViaEapi... \n")

	allowedModes := []string{"access", "trunk"}

	for _, swCR := range allSwitchCRs.Items {
		ipToUse, err := utils.GetIp(&swCR, "")
		if err != nil {
			msg := fmt.Sprintf("error : %v", err)
			fmt.Printf(msg)
			continue
		}

		_, err = switchclients.NewAristaClient(ipToUse, switchSecretsPath, 443, "https", 30*time.Second, true, []int{}, []int{}, allowedModes, nil, []int{})
		if err != nil {
			msg := fmt.Sprintf("create switch client for %s failed, %v \n", swCR.Name, err)
			fmt.Printf(msg)
			continue
		}
	}

	fmt.Printf("finished checkSwitchConnectivityViaEapi... \n")
}

func checkSwitchportConfigViaEapi(ctx context.Context, switchSecretsPath string, allSwitchCRs *idcnetworkv1alpha1.SwitchList, allSwitchPortCRs *idcnetworkv1alpha1.SwitchPortList) {
	fmt.Printf("starting checkSwitchportConfigViaEapi... \n")

	allowedModes := []string{"access", "trunk"}

	for _, swCR := range allSwitchCRs.Items {
		// Find all switchports that the SDN controller manages on this switch.
		var managedPortsOnThisSwitch = make(map[string]idcnetworkv1alpha1.SwitchPort)

		for _, spCR := range allSwitchPortCRs.Items {
			portShortName, switchFQDN := utils.PortFullNameToPortNameAndSwitchFQDN(spCR.Name)

			if switchFQDN == swCR.Name {
				managedPortsOnThisSwitch[portShortName] = spCR
			}
		}

		// If there are no SDN-managed ports on this switch, ignore this switch.
		if len(managedPortsOnThisSwitch) == 0 {
			continue
		}
		ipToUse, err := utils.GetIp(&swCR, "")
		if err != nil {
			msg := fmt.Sprintf("error : %v", err)
			fmt.Printf(msg)
			continue
		}

		switchClient, err := switchclients.NewAristaClient(ipToUse, switchSecretsPath, 443, "https", 30*time.Second, true, []int{}, []int{}, allowedModes, nil, []int{})
		if err != nil {
			msg := fmt.Sprintf("create switch client for %s failed, %v \n", swCR.Name, err)
			fmt.Printf(msg)
			continue
		}

		emptySPReq := switchclients.GetSwitchPortsRequest{}
		switchPortsFromSwitch, err := switchClient.GetSwitchPorts(ctx, emptySPReq)
		if err != nil {
			msg := fmt.Sprintf("Getting ports from %s failed, %v \n", swCR.Name, err)
			fmt.Printf(msg)
			continue
		}

		for _, spFromSwitch := range switchPortsFromSwitch {
			spCR, found := managedPortsOnThisSwitch[spFromSwitch.Name]

			if !found {
				// Port is not controlled by SDN Controller
				continue
			}

			if spCR.Spec.VlanId != spFromSwitch.VlanId && spCR.Spec.VlanId != idcnetworkv1alpha1.NOOPVlanID {
				msg := fmt.Sprintf("Port %s had vlanId %d in CR spec, but %d on the switch. \n", spCR.Name, spCR.Spec.VlanId, spFromSwitch.VlanId)
				fmt.Printf(msg)
			}

			if len(spFromSwitch.TrunkGroups) != 0 {
				msg := fmt.Sprintf("Switchport %s is controlled by SDN controller, but has a TrunkGroup on the switch. \n", spCR.Name)
				fmt.Printf(msg)
			}

			if spFromSwitch.LinkStatus != "up" && spFromSwitch.LinkStatus != "connected" {
				msg := fmt.Sprintf("Managed switchport %s linkStatus is %s on the switch. \n", spCR.Name, spFromSwitch.LinkStatus)
				fmt.Printf(msg)
			}

			if spFromSwitch.Mode != "access" {
				msg := fmt.Sprintf("SDN controls SwitchPort %s but it is not set to access mode on the switch. It is '%s' \n", spCR.Name, spFromSwitch.Mode)
				fmt.Printf(msg)
			}

			if spFromSwitch.PortChannel != 0 {
				msg := fmt.Sprintf("SDN controls SwitchPort %s but it has a portChannel set, '%d' \n", spCR.Name, spFromSwitch.PortChannel)
				fmt.Printf(msg)
			}

		}
	}

	fmt.Printf("finished checkSwitchportConfigViaEapi... \n")
}
