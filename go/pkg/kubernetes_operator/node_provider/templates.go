// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package node_provider

import (
	"text/template"
)

const (
	networkDataTemplateText = `  version: 2
  ethernets:
    enp1s0:
     dhcp4: false
     addresses: 
       - {{.IPAddressSubnetMask}}
     gateway4: {{.Gateway}}
     nameservers:
       addresses: [{{.Nameserver}}]
`

	userDataTemplateText = `#cloud-config
package_update: true
packages:
  - qemu-guest-agent
write_files:
  - path: /usr/local/bin/bootstrap.sh
    permissions: "0700"
    content: |
    {{range .BootstrapScript}}
      {{ . }}
    {{end}}
runcmd:
  - systemctl enable '--now' qemu-guest-agent
  - {{.RegistrationCmd}}
  {{.DownloadCustomBashScript}}
  {{.RunCustomBashScript}}
`
)

var (
	NetworkDataTemplate = template.Must(template.New("networkData").Parse(networkDataTemplateText))
	UserDataTemplate    = template.Must(template.New("userData").Parse(userDataTemplateText))
)

type NetworkData struct {
	IPAddressSubnetMask string
	Gateway             string
	Nameserver          string
}

type UserData struct {
	RegistrationCmd          string
	BootstrapScript          []string
	DownloadCustomBashScript string
	RunCustomBashScript      string
}
