// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"path"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/descriptorpb"
)

type MethInfo struct {
	FileName       string
	File           *descriptorpb.FileDescriptorProto
	FileOptions    *pb.IdcFileOptions
	Deploy         string
	ServerName     string
	Service        *descriptorpb.ServiceDescriptorProto
	ServiceOptions *pb.IdcServiceOptions
	Method         *descriptorpb.MethodDescriptorProto
	MethodOptions  *pb.IdcMethodOptions
}

type ForEachMethodFunc func(*MethInfo) error

func ForEachMethod(fileNames []string, methFunc ForEachMethodFunc) error {
	info := MethInfo{}
	var err error
	for _, info.FileName = range fileNames {
		info.File, err = ReadProtoFd(info.FileName)
		if err != nil {
			return err
		}

		info.FileOptions = GetIdcFileOptions(info.File)
		if info.FileOptions == nil {
			continue
		}

		info.Deploy = GetDeploy(info.File)
		if info.Deploy == "" {
			continue
		}

		info.ServerName = info.FileOptions.Service
		if info.ServerName == "" {
			info.ServerName = strings.TrimSuffix(path.Base(info.FileName), ".proto")
		}

		for _, info.Service = range info.File.Service {
			info.ServiceOptions = GetIdcServiceOptions(info.Service)
			for _, info.Method = range info.Service.Method {
				info.MethodOptions = GetIdcMethodOptions(info.Method)
				if err := methFunc(&info); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
