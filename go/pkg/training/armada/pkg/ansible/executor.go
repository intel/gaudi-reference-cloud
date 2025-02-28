// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ansible

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/apenella/go-ansible/pkg/options"
	"github.com/apenella/go-ansible/pkg/playbook"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

//go:embed templates/hosts.tpl
var hostTmpFilepath string

type instanceConnect struct {
	SSHProxy              string
	User                  string
	JupyterHubNodeAddress string
}

func SetupJupyterHub(ctx context.Context, clusterId, privKey string, jupyterhubInstance *v1.Instance) error {
	log := log.FromContext(ctx).WithName("ansibleExecutor.SetupJupyterHub")
	connectOpt := instanceConnect{}
	if jupyterhubInstance == nil {
		return fmt.Errorf("Missing jupyterhub instance")
	}

	connectOpt.SSHProxy = fmt.Sprintf("%s@%s", jupyterhubInstance.Status.GetSshProxy().GetProxyUser(), jupyterhubInstance.Status.GetSshProxy().GetProxyAddress())
	connectOpt.User = jupyterhubInstance.Status.GetUserName()
	log.Info("connectionParser", "sshProxy", connectOpt.SSHProxy, "userName", connectOpt.User)

	labels := jupyterhubInstance.Metadata.GetLabels()
	if role, found := labels["oneapi-instance-role"]; found {
		interfaces := jupyterhubInstance.Status.GetInterfaces()
		if len(interfaces) != 0 {
			addresses := interfaces[0].GetAddresses()
			if len(addresses) == 0 {
				log.Info("no network address found for jupyterhub instance", "instanceId", jupyterhubInstance.Metadata.GetResourceId())
				return fmt.Errorf("no jupyterhub network address found")
			}

			log.Info("debug", "interface", addresses, "role", role)
			if strings.EqualFold(role, "slurm-jupyterhub-node") {
				connectOpt.JupyterHubNodeAddress = fmt.Sprintf("%s:%s", jupyterhubInstance.Metadata.Name, addresses[0])
			}
		} else {
			log.Info("no network interface found", "instanceId", jupyterhubInstance.Metadata.GetResourceId())
			return fmt.Errorf("no network interface found for jupyterhub instance")
		}
	}

	log.Info("network connection parsed successfully for jupyterhub instance", "config", connectOpt)

	// tmpInvFile, _ := ioutil.TempFile(os.TempDir(), "oneapihosts-")
	tmpInvFile, err := os.CreateTemp(os.TempDir(), "oneapihosts-")
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return err
	}

	defer os.Remove(tmpInvFile.Name())

	fmt.Println("Temporary file created successfully:", tmpInvFile.Name())

	// Verify the value returned by os.CreateTemp
	// For example, you can check if the file exists
	fileInfo, err := os.Stat(tmpInvFile.Name())
	if err != nil {
		fmt.Println("Error checking file information:", err)
		return nil
	}

	// Check if the file exists
	if fileInfo != nil && !fileInfo.IsDir() {
		fmt.Println("File verification successful.")
	} else {
		fmt.Println("File verification failed.")
	}

	jupyterhubNodeOpts := strings.Split(connectOpt.JupyterHubNodeAddress, ":")
	jupyterhubNodeOptsNodeStr := fmt.Sprintf("%s ansible_host=%s", jupyterhubNodeOpts[0], jupyterhubNodeOpts[1])
	templ := template.Must(template.New("inventoryCreate").Parse(string(hostTmpFilepath)))
	buf := bytes.Buffer{}
	// templ.Execute(&buf, map[string]interface{}{
	// 	"SlurmJupyterHubHosts": jupyterhubNodeOptsNodeStr,
	// })
	err = templ.Execute(&buf, map[string]interface{}{
		"SlurmJupyterHubHosts": jupyterhubNodeOptsNodeStr,
	})
	if err != nil {
		// handle the error here
		log.Error(err, "error reading file")
		return err
	}
	log.Info("Writing to", "hostfile", tmpInvFile.Name())

	//0644 allows the owner to read and write the file while allowing others only to read it
	//0600 sets read and write permissions only for the owner.
	if err := os.WriteFile(tmpInvFile.Name(), buf.Bytes(), 0600); err != nil {
		log.Error(err, "Error Writing to", tmpInvFile.Name())
		return err
	}

	content, err := os.ReadFile(tmpInvFile.Name())
	if err != nil {
		log.Error(err, "Read failed")
	}
	fmt.Println(string(content))

	time.Sleep(10 * time.Second)

	proxyCommand := "-o StrictHostKeyChecking=accept-new -o ProxyCommand=\"ssh -W %h:%p " +
		" -q " + connectOpt.SSHProxy + " -i " + privKey +
		" -o StrictHostKeyChecking=accept-new\""

	jupyterhubAnsiblePlaybookConnectionOptions := &options.AnsibleConnectionOptions{
		Connection:   "ssh",
		SSHExtraArgs: proxyCommand,
		User:         connectOpt.User,
		PrivateKey:   privKey,
	}

	jupyterhubPlaybookOptions := &playbook.AnsiblePlaybookOptions{
		Inventory: tmpInvFile.Name(),
	}

	jupyterhubPlaybook := &playbook.AnsiblePlaybookCmd{
		Playbooks:         []string{"/training/slurm-static/batchbeta-jupyterhub.yml"},
		ConnectionOptions: jupyterhubAnsiblePlaybookConnectionOptions,
		Options:           jupyterhubPlaybookOptions,
	}

	err = jupyterhubPlaybook.Run(ctx)
	if err != nil {
		log.Error(err, "Failed to run jupyterhub task(s)")
		return fmt.Errorf("Failed running jupyterhub ansible playbook")
	}

	return nil
}
