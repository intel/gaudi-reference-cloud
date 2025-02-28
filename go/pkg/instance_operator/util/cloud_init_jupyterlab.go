// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"bytes"
	_ "embed"
	"text/template"
)

var (
	//go:embed idc_jupyterlab.sh.gotmpl
	idcJupyterLabShTmpl string

	tmpl = template.Must(template.New("idcJupyterLabSh").Parse(idcJupyterLabShTmpl))
)

func (c *CloudConfig) SetJupyterLab(namespace, name, rootCAPublicCertificateFileContent, defaultUsername, quickConnectHost string) error {
	buf := bytes.Buffer{}
	err := tmpl.Execute(&buf, map[string]interface{}{
		"CloudAccountId":   namespace,
		"InstanceId":       name,
		"ClientCA":         rootCAPublicCertificateFileContent,
		"User":             defaultUsername,
		"QuickConnectHost": quickConnectHost,
	})
	if err != nil {
		return err
	}
	c.AddWriteFileWithPermissions("/opt/idc/idc_jupyterlab.sh", buf.String(), "0755")
	c.AddRunCmd("/opt/idc/idc_jupyterlab.sh _setup-first-boot")
	return nil
}
