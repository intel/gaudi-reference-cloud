package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	var action string
	var clabFolder string
	flag.StringVar(&action, "action", "add", "add/delete")
	flag.StringVar(&clabFolder, "clab", "../../../../../networking/containerlab/allscfabrics", "")
	flag.Parse()

	client := utils.NewK8SClient()
	if action == "add" {
		kubectlCmd("apply", []string{"../config/bmh-crd.yaml", "../config/bmh-ns.yaml", "switches.yaml"})
		createBMH([]string{"server1-1.json", "server1-2.json", "server2-1.json"}, client)
	} else if action == "delete" {
		kubectlCmd("delete", []string{"switches.yaml"})
		deleteBMHandNG([]string{"server1-1.json", "server1-2.json", "server2-1.json"}, client)
	}
}

func kubectlCmd(kubectlAction string, files []string) {
	cmd := "kubectl"
	for _, file := range files {
		args := []string{kubectlAction, "-f", file}
		command := exec.Command(cmd, args...)

		var stdoutBuf, stderrBuf bytes.Buffer
		command.Stdout = &stdoutBuf
		command.Stderr = &stderrBuf

		err := command.Run()
		if err != nil {
			fmt.Println("Error:", err, "\nStderr:", stderrBuf.String())
			continue
		}

		fmt.Println("Output:", stdoutBuf.String())
	}
}

func createBMH(bmhfiles []string, client rtclient.Client) {
	for _, file := range bmhfiles {
		ctx := context.Background()
		bmh := &baremetalv1alpha1.BareMetalHost{}
		data, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("ReadFile failed: %v \n", err)
			continue
		}
		err = json.Unmarshal(data, bmh)
		if err != nil {
			fmt.Printf("Unmarshal failed: %v \n", err)
			continue
		}

		bmhCopy := bmh.DeepCopy()
		err = client.Create(ctx, bmhCopy, &rtclient.CreateOptions{})
		if err != nil {
			fmt.Printf("Create bmh failed: %v \n", err)
			continue
		}

		key := types.NamespacedName{Name: bmh.Name, Namespace: bmh.Namespace}
		latestBMH := &baremetalv1alpha1.BareMetalHost{}
		err = client.Get(ctx, key, latestBMH)
		if err != nil {
			fmt.Printf("Get bmh failed: %v \n", err)
			continue
		}
		latestBMH.Status = bmh.Status
		err = client.Status().Update(ctx, latestBMH, &rtclient.SubResourceUpdateOptions{})
		if err != nil {
			fmt.Printf("Update status failed: %v \n", err)
			continue
		}
	}
}

func deleteBMHandNG(bmhfiles []string, client rtclient.Client) {
	for _, file := range bmhfiles {
		ctx := context.Background()
		bmh := &baremetalv1alpha1.BareMetalHost{}
		data, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("ReadFile failed: %v \n", err)
			continue
		}
		err = json.Unmarshal(data, bmh)
		if err != nil {
			fmt.Printf("Unmarshal failed: %v \n", err)
			continue
		}

		err = client.Delete(ctx, bmh)
		if err != nil {
			fmt.Printf("delete bmh failed: %v \n", err)
			continue
		}

		if groupID, found := bmh.Labels[idcnetworkv1alpha1.LabelBMHGroupID]; found {
			key := types.NamespacedName{Name: groupID, Namespace: "idcs-system"}
			ng := &idcnetworkv1alpha1.NodeGroup{}
			err = client.Get(ctx, key, ng)
			if err != nil {
				fmt.Printf("Get ng failed: %v \n", err)
				continue
			}
			err = client.Delete(ctx, ng)
			if err != nil {
				fmt.Printf("delete ng failed: %v \n", err)
				continue
			}
		}

	}
}
