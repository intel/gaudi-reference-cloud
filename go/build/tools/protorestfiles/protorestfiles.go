// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/build/tools/util"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"log"
	"os"
	"path"
	"strings"
)

// protorestfiles filters .proto files, printing the files that contain
// http annotations. The purpose is to avoid generating grpc-rest-gateway
// code and swagger definitions for APIs that don't have http annotations.
//
// Along the way protorestfiles checks for some conditions required for
// grpc-rest-gateway and grpc-proxy to work correctly:
// 1. The package name in .proto files is required to be "proto"
// 2. Every .proto file must have an "option (idc.file).deploy" setting
// 3. Every path in the REST API must start with "/v1/"
func main() {
	errs := 0
	for _, fileName := range os.Args[1:] {
		fd, err := util.ReadProtoFd(fileName)
		if err != nil {
			log.Fatal(err)
		}

		if path.Base(fileName) != "annotations.proto" {
			// Every .proto file must have "package proto;"
			if fd.GetPackage() != "proto" {
				_, fmtErr := fmt.Fprintf(os.Stderr, "ERROR: %v: package must be \"proto\", found %v\n", fileName, fd.GetPackage())
				if fmtErr != nil {
					log.Fatal(fmtErr)
				}
				errs++
			}

			// Every .proto file must have "option (idc.file).deploy"
			deploy := util.GetDeploy(fd)
			if deploy == "" {
				_, fmtErr := fmt.Fprintf(os.Stderr, "ERROR: %v: missing option (idc.file).deploy option\n", fileName)
				if fmtErr != nil {
					log.Printf("Failed to write to STDERR")
				}
				errs++
			}
		}

		fdErrs, hasHttp := checkRestPaths(fileName, fd)
		errs += fdErrs

		if !hasHttp || errs > 0 {
			continue
		}

		fmt.Printf("%v\n", fileName)
	}
	if errs > 0 {
		os.Exit(1)
	}
}

func checkRestPaths(fileName string, fd *descriptorpb.FileDescriptorProto) (errs int, hasHttp bool) {
	errs = 0
	hasHttp = false
	for _, service := range fd.Service {
		for _, meth := range service.Method {
			opts := meth.GetOptions()
			if opts == nil {
				continue
			}
			xt := proto.GetExtension(opts, annotations.E_Http)
			rule, ok := xt.(*annotations.HttpRule)
			if !ok || rule == nil {
				continue
			}
			hasHttp = true
			errs += checkRule(fileName, service, meth, rule)
		}
	}
	return errs, hasHttp
}

func checkRule(fileName string, service *descriptorpb.ServiceDescriptorProto, meth *descriptorpb.MethodDescriptorProto,
	rule *annotations.HttpRule) int {

	errs := 0
	checkRestPath := func(path string) {
		if path == "" {
			return
		}
		// All REST paths must start with /v1/
		if !strings.HasPrefix(path, "/v1/") {
			_, fmtErr := fmt.Fprintf(os.Stderr, "ERROR: %v: %v.%v path %v: must start with /v1/\n", fileName, service.GetName(), meth.GetName(), path)
			if fmtErr != nil {
				log.Fatal(fmtErr)
			}
			errs++
		}
	}

	custom := rule.GetCustom()
	if custom != nil {
		checkRestPath(custom.Path)
	}
	checkRestPath(rule.GetDelete())
	checkRestPath(rule.GetGet())
	checkRestPath(rule.GetPatch())
	checkRestPath(rule.GetPost())
	checkRestPath(rule.GetPut())
	for _, add := range rule.GetAdditionalBindings() {
		errs += checkRule(fileName, service, meth, add)
	}
	return errs
}
