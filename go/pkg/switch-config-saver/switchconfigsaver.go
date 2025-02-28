// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	switchclients "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
	"github.com/rogpeppe/go-internal/diff"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(idcnetworkv1alpha1.AddToScheme(scheme))
}

// Example of running the tool:
// go run switchconfigsaver.go --kubeconfig="config.idc-staging-nwcp"
func main() {
	var eapiSecretPath string
	var intervalSec int
	flag.StringVar(&eapiSecretPath, "eapiSecretPath", "", "Eapi secret file path")
	flag.IntVar(&intervalSec, "interval", 86400, "refresh interval")

	flag.Parse()
	//var err error
	ctx := context.Background()
	log.SetDefaultLogger()

	fmt.Printf("Connecting to nwcp k8s... \n")
	// If --kubeconfig is set, will use the kubeconfig file at that location.  Otherwise will assume running
	// in cluster and use the cluster provided kubeconfig.
	nwcpK8sClient := utils.NewK8SClientWithScheme(scheme)

	if eapiSecretPath == "" {
		fmt.Printf("eapiSecretPath is not provided\n")
		os.Exit(1)
	}

	fmt.Printf("Starting ticker... \n")
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	for {
		//
		//allSwitchPortCRs, err := GetAllSwitchPorts(ctx, nwcpK8sClient)
		//if err != nil {
		//	fmt.Printf("GetAllSwitchPorts failed, %v \n", err)
		//}

		allSwitchCRs, err := GetAllSwitches(ctx, nwcpK8sClient)
		if err != nil {
			fmt.Printf("GetAllSwitches failed, %v \n", err)
		}

		saveConfigOnSwitchesWhereNeeded(ctx, allSwitchCRs, eapiSecretPath)

		select {
		case <-ticker.C:
			continue
		}
	}
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

var ignoreSwitchRegex = regexp.MustCompile("^fxhb3p3p-zal02.*$")

func saveConfigOnSwitchesWhereNeeded(ctx context.Context, allSwitchCRs *idcnetworkv1alpha1.SwitchList, switchSecretsPath string) {

	fmt.Printf("start saving config on switches... \n")

	allowedModes := []string{"access", "trunk"}

	for _, swCR := range allSwitchCRs.Items {

		// Hack to avoid pod2 in flexential which exists in Raven, but is not part of IDC.
		if ignoreSwitchRegex.MatchString(swCR.Name) {
			msg := fmt.Sprintf("ignoring %s \n", swCR.Name)
			fmt.Printf(msg)
			continue
		}
		ipToUse, err := utils.GetIp(&swCR, "")
		if err != nil {
			msg := fmt.Sprintf("error : %v", err)
			fmt.Printf(msg)
			continue
		}

		switchClient, err := switchclients.NewAristaClient(ipToUse, switchSecretsPath, 443, "https", 30*time.Second, false, []int{}, []int{}, allowedModes, nil, []int{})
		if err != nil {
			msg := fmt.Sprintf("create switch client for %s failed, %v \n", swCR.Name, err)
			fmt.Printf(msg)
			continue
		}

		runningConfig, err := switchClient.GetRunningConfig(ctx)
		if err != nil {
			msg := fmt.Sprintf("failed to get running-config for %s, %v \n", swCR.Name, err)
			fmt.Printf(msg)
			continue
		}

		startupConfig, err := switchClient.GetStartupConfig(ctx)
		if err != nil {
			msg := fmt.Sprintf("failed to get startup-config for %s, %v \n", swCR.Name, err)
			fmt.Printf(msg)
			continue
		}

		// Clean configs to remove items that should not be considered in the diff.
		var cleanRunningConfig = cleanConfigText(runningConfig)
		var cleanStartupConfig = cleanConfigText(startupConfig)

		var diffString = string(diff.Diff("startupConfig", []byte(cleanStartupConfig), "runningConfig", []byte(cleanRunningConfig)))
		if diffString != "" {
			fmt.Printf("Saving running config as startupConfig on %s. Diff: %s \n", swCR.Name, diffString)
			switchClient.SaveRunningConfigAsStartupConfig(ctx)
		}

	}

	fmt.Printf("finished saving config on switches... \n")

}

var commentRegex = regexp.MustCompile("(?m)^!.*$")
var emptyLineRegex = regexp.MustCompile("(?m)^\\s$")

func cleanConfigText(configText string) string {
	// Ignore comments
	configTextAltered := commentRegex.ReplaceAllString(configText, "")
	// Ignore empty lines
	configTextAltered = emptyLineRegex.ReplaceAllString(configTextAltered, "")

	// TODO: Ignore the VLAN for any SDN-controller controlled ports. Don't need to trigger a save if that's the only thing that changed.

	return configTextAltered
}
