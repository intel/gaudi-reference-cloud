// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// prototempl.go processes go templates using data from
// public_api/proto. Each .proto file is parsed looking for
// service names. Information about the services is applied
// to a template. This is used to maintain the grpc-proxy,
// grpc-rest-gateway, and grpc-reflect services.
package main

import (
	"bytes"
	"flag"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/build/tools/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type Service struct {
	Name         string
	Service      string
	Cluster      string
	RegisterFunc string
	Deployment   string
}

type Method struct {
	Service       string
	Name          string
	Cluster       string
	Deployment    string
	StreamForever bool
}

type Cluster struct {
	Name       string
	Service    string
	Deployment string
}

type Data struct {
	ConfigMapNameSuffix string
	Services            []Service
	Methods             []Method
	Clusters            []Cluster
	Deployments         []string
}

var configMapNameSuffix string
var outputFile string
var tmplFile string
var includeDir string
var appClientFilter bool

func addDeployment(deployments []string, deployment string) []string {
	for _, dep := range deployments {
		if dep == deployment {
			return deployments
		}
	}
	return append(deployments, deployment)
}

func init() {
	flag.StringVar(&configMapNameSuffix, "output-configmap-name-suffix", "", "output configmap name suffix")
	flag.StringVar(&outputFile, "output-file", "", "output file")
	flag.StringVar(&tmplFile, "template-file", "", "template file")
	flag.StringVar(&includeDir, "I", "", "include dir")
	flag.BoolVar(&appClientFilter, "C", false, "app client filter")
}

func main() {
	flag.Parse()
	if outputFile == "" {
		log.Fatal("output-file is required")
	}
	if tmplFile == "" {
		log.Fatal("template-file is required")
	}
	svcs := []Service{}
	methods := []Method{}
	deployments := []string{}
	clusters := []Cluster{}

	// Sort arguments to ensure reproducible output.
	args := flag.Args()
	sort.Strings(args)

	for _, arg := range args {
		if !strings.HasSuffix(arg, ".proto") {
			log.Fatalf("%v is not a .proto file", arg)
		}
		name := strings.TrimSuffix(filepath.Base(arg), ".proto")
		cluster := name

		file, err := util.ReadProtoFd(arg)
		if err != nil {
			log.Fatal(err)
		}

		fileOpt := util.GetIdcFileOptions(file)
		if fileOpt == nil || fileOpt.Deploy == pb.DeploymentType_unspecified {
			continue
		}
		deployment := fileOpt.Deploy.String()
		if deployment == "" {
			continue
		}
		clusterService := name
		deployments = addDeployment(deployments, deployment)
		ss := fileOpt.GetService()
		if ss != "" {
			clusterService = ss
		}
		for _, service := range file.Service {
			svc := Service{
				Name:         name,
				Service:      *service.Name,
				Cluster:      cluster,
				RegisterFunc: "Register" + *service.Name + "Handler",
				Deployment:   deployment,
			}
			svcs = append(svcs, svc)
			for _, method := range service.Method {
				streamForever := false
				opts := util.GetIdcMethodOptions(method)
				if opts != nil && opts.StreamForever {
					streamForever = true
				}
				meth := Method{
					Service:       *service.Name,
					Name:          *method.Name,
					Cluster:       cluster,
					Deployment:    deployment,
					StreamForever: streamForever,
				}
				if appClientFilter {
					if opts != nil && opts.Authz != nil && opts.Authz.AppClientAccess {
						methods = append(methods, meth)
					}
				} else {
					methods = append(methods, meth)
				}
			}
		}
		clusters = append(clusters, Cluster{Name: cluster,
			Deployment: deployment,
			Service:    clusterService,
		})

		data := Data{
			ConfigMapNameSuffix: configMapNameSuffix,
			Services:            svcs,
			Methods:             methods,
			Clusters:            clusters,
			Deployments:         deployments,
		}
		tmpl, err := template.ParseFiles(tmplFile)
		if err != nil {
			log.Fatalf("parse %v: %v", tmplFile, err)
		}

		out := &bytes.Buffer{}
		comment := "#"
		if strings.HasSuffix(outputFile, ".go") {
			comment = "//"
		}
		err = util.WriteGenComment(out, "prototempl", comment)
		if err != nil {
			log.Fatalf("error writing comment: %v", err)
		}
		err = tmpl.Execute(out, data)
		if err != nil {
			log.Fatalf("error applying template: %v", err)
		}
		var outBytes []byte
		if strings.HasSuffix(outputFile, ".go") {
			outBytes, err = format.Source(out.Bytes())
			if err != nil {
				log.Fatalf("error formatting template output: %v", err)
			}
		} else {
			outBytes = out.Bytes()
		}

		tmpFile := outputFile + ".tmp"
		tmp, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0660))
		if err != nil {
			log.Fatalf("open %v: %v", tmpFile, err)
		}
		if _, err := tmp.Write(outBytes); err != nil {
			log.Fatalf("write %v: %v", tmpFile, err)
		}
		if err := tmp.Close(); err != nil {
			log.Printf("Failed to close temp file %v", tmpFile)
		}

		if err := os.Rename(tmpFile, outputFile); err != nil {
			log.Fatalf("rename %v %v: %v", tmpFile, outputFile, err)
		}
	}
}
