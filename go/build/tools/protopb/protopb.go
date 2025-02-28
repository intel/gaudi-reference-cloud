// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// generate the .pb files that ope-envoy needs to parse protobuf request bodies into
// a form that can be used by .rego policies
package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"text/template"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/build/tools/util"
)

var (
	configMap  string
	pbDir      string
	includeDir string
)

func init() {
	flag.StringVar(&configMap, "configmap", "", "configmap location")
	flag.StringVar(&pbDir, "pb-dir", "", "pb directory")
	flag.StringVar(&includeDir, "I", "", "include dir")
}

func main() {
	flag.Parse()
	if pbDir == "" && configMap == "" {
		log.Fatal("specify --output-dir or --configmap")
	}

	files := map[string][]string{"all": {}}

	// Sort arguments to ensure reproducible output.
	args := flag.Args()
	sort.Strings(args)

	for _, fileName := range args {
		fd, err := util.ReadProtoFd(fileName)
		if err != nil {
			log.Fatal(err)
		}

		deploy := util.GetDeploy(fd)
		if deploy == "" {
			continue
		}
		deployFiles := files[deploy]
		files[deploy] = append(deployFiles, fileName)
		files["all"] = append(files["all"], fileName)
	}

	protos, err := readProtos(files)
	if err != nil {
		log.Fatal(err)
	}

	if configMap != "" {
		outputConfigMap(protos)
	}

	if pbDir != "" {
		outputPbs(protos)
	}
}

func readProtos(fileMap map[string][]string) (map[string][]byte, error) {
	protos := map[string][]byte{}
	opt := util.ReadProtoOpts{IncludeImports: true}
	for deploy, files := range fileMap {
		contents, err := util.ReadProto(&opt, files...)
		if err != nil {
			return nil, err
		}
		protos[deploy] = contents
	}
	return protos, nil
}

var configMapHeader = []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "idc-common.fullname" . }}-pb
  namespace: {{ include "idc-common.namespace" . }}
  annotations:
    argocd.argoproj.io/sync-options: Replace=true
`)

var configMapDataTemplate = `binaryData:
{{- range $deploy, $content := .}}
  idc.pb: "{{ $content }}"
{{- end }}
 `

// The .pb files are stored in a configmap for use by opa-envoy
func outputConfigMap(pbs map[string][]byte) {
	deploys := map[string]string{}
	for deploy, content := range pbs {
		// options are "regional", "global", "all"
		if deploy == "all" {
			deploys[deploy] = base64.StdEncoding.EncodeToString(content)
		}
	}

	tmpl, err := template.New("pb").Parse(configMapDataTemplate)
	if err != nil {
		log.Fatal(err)
	}

	if err := util.WriteFileAtomically(configMap,
		func(outf io.Writer) error {
			if _, err := outf.Write(configMapHeader); err != nil {
				return err
			}
			return tmpl.Execute(outf, deploys)
		}); err != nil {
		log.Fatal(err)
	}
}

// The .pbs can also be stored stand-alone for distribution using OPAL or other means
func outputPbs(pbs map[string][]byte) {
	if err := os.MkdirAll(pbDir, 0770); err != nil {
		log.Fatal(err)
	}

	for deploy, content := range pbs {
		fileName := fmt.Sprintf("%s/%s.pb", pbDir, deploy)
		if err := util.WriteFileAtomically(fileName,
			func(outf io.Writer) error {
				_, err := outf.Write(content)
				return err
			}); err != nil {
			log.Fatal(err)
		}
	}
}
