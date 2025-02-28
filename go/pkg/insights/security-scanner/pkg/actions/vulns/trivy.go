// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vulns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

const (
	// Define the command template
	scanContainerCmd = "/app/go/pkg/insights/security-scanner/cmd/main/security-scanner_image.binary.runfiles/trivy_linux_64/trivy image -f json -o {{.OutputFile}} {{.URL}}"
)

type VulnerabilityReport struct {
	Target          string `json:"Target"`
	Type            string `json:"Type"`
	Vulnerabilities []struct {
		Id               string    `json:"VulnerabilityID"`
		PkgName          string    `json:"PkgName"`
		InstalledVersion string    `json:"InstalledVersion"`
		Title            string    `json:"Title"`
		Severity         string    `json:"Severity"`
		CWEs             []string  `json:"CweIDs"`
		PublishedAt      time.Time `json:"PublishedDate"`
		LastModified     time.Time `json:"LastModifiedDate"`
		FixedVersion     string    `json:"FixedVersion"`
	} `json:"Vulnerabilities"`
}

type TrivyResult struct {
	Results []struct {
		Target          string `json:"Target"`
		Type            string `json:"Type"`
		Vulnerabilities []struct {
			Id               string    `json:"VulnerabilityID"`
			PkgName          string    `json:"PkgName"`
			InstalledVersion string    `json:"InstalledVersion"`
			Title            string    `json:"Title"`
			Severity         string    `json:"Severity"`
			CWEs             []string  `json:"CweIDs"`
			PublishedAt      time.Time `json:"PublishedDate"`
			LastModified     time.Time `json:"LastModifiedDate"`
			FixedVersion     string    `json:"FixedVersion"`
		} `json:"Vulnerabilities"`
	}
}

func ScanImage(ctx context.Context, imageURL string) ([]VulnerabilityReport, error) {
	logger := log.FromContext(ctx).WithName("trivy.ScanImage")
	vulns := []VulnerabilityReport{}

	if err := os.Setenv("https_proxy", ""); err != nil {
		logger.Error(err, "failed to set env https_proxy")
	}
	if err := os.Setenv("HTTPS_PROXY", ""); err != nil {
		logger.Error(err, "failed to set env HTTPS_PROXY")
	}

	os.Chmod("/app/go/pkg/insights/security-scanner/cmd/main/security-scanner_image.binary.runfiles/trivy_linux_64/trivy", 0755)
	cmdTmpl, err := template.New("trivyScanCmd").Parse(scanContainerCmd)
	if err != nil {
		fmt.Printf("error formating trivy scan cmd: %v", err)
	}

	type input struct {
		URL        string
		OutputFile string
	}

	outJson, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		logger.Error(err, "error creating temp file")
		return nil, err
	}
	defer outJson.Close()
	defer os.RemoveAll(outJson.Name())

	var cmdBuf bytes.Buffer
	cmdTmpl.Execute(&cmdBuf, input{URL: imageURL, OutputFile: outJson.Name()})
	logger.Info("execute command", "cmd", cmdBuf.String())

	_, err = execCmd(cmdBuf.String())
	if err != nil {
		fmt.Printf("failed to scan image: %s, err: %v\n", imageURL, err)
		return nil, fmt.Errorf("failed to scan image: %s", imageURL)
	}

	vBuf, _ := os.ReadFile(outJson.Name())

	trivyRes := TrivyResult{}
	if err := json.Unmarshal(vBuf, &trivyRes); err != nil {
		logger.Error(err, "error parsing json result")
		logger.Info("error parsing json result", "buf", string(vBuf))
		return nil, err
	}
	imgReport := VulnerabilityReport{}
	for _, t := range trivyRes.Results {
		imgReport.Vulnerabilities = append(imgReport.Vulnerabilities, t.Vulnerabilities...)
	}
	vulns = append(vulns, imgReport)

	logger.Info("debug2:", "# vulnerabilities ", len(imgReport.Vulnerabilities))
	return vulns, nil
}

func execCmd(cmdStr string) ([]byte, error) {
	cmdS := strings.Split(cmdStr, " ")
	cmd := exec.Command(cmdS[0], cmdS[1:]...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	var waitStatus syscall.WaitStatus
	exitCode := 0
	if err := cmd.Run(); err != nil {
		// Did the command fail because of an unsuccessful exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			exitCode = waitStatus.ExitStatus()
		}
	} else {
		// Command was successful
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = waitStatus.ExitStatus()
	}
	if exitCode > 1 {
		return nil, fmt.Errorf("failed to scan image")
	}

	return outb.Bytes(), nil
}
