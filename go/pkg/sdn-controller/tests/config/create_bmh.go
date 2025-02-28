// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
)

// This tool load the bmh-rc.json and create the BMH cr.
func main() {
	loadBMHObj("bmh-cr-1.json")
	// loadBMHObj("bmh-rc-2.json")
	// loadBMHObj("bmh-cr-gaudi.json")
	// loadBMHObj("bmh-cr-gaudi-edgecore.json")
	// updateBMHStatus()
}

func loadBMHObj(file string) {
	ctx := context.Background()
	bmh := &baremetalv1alpha1.BareMetalHost{}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("ReadFile failed: %v \n", err)
		return
	}
	err = json.Unmarshal(data, bmh)
	if err != nil {
		fmt.Printf("Unmarshal failed: %v \n", err)
		return
	}
	client := utils.NewK8SClient()

	copy := bmh.DeepCopy()

	err = client.Create(ctx, bmh, &rtclient.CreateOptions{})
	if err != nil {
		fmt.Printf("Create bmh failed: %v \n", err)
		return
	}

	copy.ResourceVersion = bmh.ResourceVersion
	err = client.Status().Update(ctx, copy, &rtclient.SubResourceUpdateOptions{})
	if err != nil {
		fmt.Printf("Update status failed: %v \n", err)
		return
	}
	fmt.Printf("DONE \n")
}

func updateBMHStatus() {
	ctx := context.Background()
	client := utils.NewK8SClient()

	bmh := &baremetalv1alpha1.BareMetalHost{}
	key := types.NamespacedName{Name: "pdx04-c01-bmas051", Namespace: "metal3-1"}

	err := client.Get(ctx, key, bmh)
	if err != nil {
		fmt.Printf("Get failed: %v \n", err)
		return
	}

	// change it to OperationalStatusError
	bmh.Status.OperationalStatus = baremetalv1alpha1.OperationalStatusError
	err = client.Status().Update(ctx, bmh, &rtclient.SubResourceUpdateOptions{})
	if err != nil {
		fmt.Printf("update bmh OperationalStatusError failed: %v \n", err)
		return
	}
	fmt.Printf("update bmh OperationalStatusError success")
	time.Sleep(5 * time.Second)

	// change it back to OperationalStatusOK
	bmh.Status.OperationalStatus = baremetalv1alpha1.OperationalStatusOK
	err = client.Status().Update(ctx, bmh, &rtclient.SubResourceUpdateOptions{})
	if err != nil {
		fmt.Printf("update bmh OperationalStatusOK failed: %v \n", err)
		return
	}
	fmt.Printf("update bmh OperationalStatusOK success")
}
