// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/instance_type"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/ip_resource_manager"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/machine_image"
	gittogrpcsynchronizer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/git_to_grpc_synchronizer"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog/sync"
	"google.golang.org/grpc"
)

const machineImageKind = "MachineImage"
const instanceTypeKind = "InstanceType"
const subnetKind = "Subnet"
const productKind = "Product"

type KindInfo struct {
	NewSynchronizer func(fs.FS, *grpc.ClientConn) (*gittogrpcsynchronizer.GitToGrpcSynchronizer, error)
}

func main() {
	ctx := context.Background()

	// Build map of supported kinds.
	kindMap := make(map[string]KindInfo)
	kindMap[machineImageKind] = KindInfo{machine_image.NewMachineImageSynchronizer}
	kindMap[instanceTypeKind] = KindInfo{instance_type.NewInstanceTypeSynchronizer}
	kindMap[subnetKind] = KindInfo{ip_resource_manager.NewSubnetSynchronizer}
	kindMap[productKind] = KindInfo{sync.NewProductSynchronizer}

	// Build list of supported kinds.
	kinds := make([]string, 0, len(kindMap))
	for k := range kindMap {
		kinds = append(kinds, k)
	}

	// Parse command line.
	var kind string
	var dir string
	var target string
	flag.StringVar(&kind, "kind", "", "one of: "+strings.Join(kinds, ", "))
	flag.StringVar(&dir, "dir", "", "directory that contains all yaml files")
	flag.StringVar(&target, "target", "", "address of GRPC server in format host:port")

	log.BindFlags()
	flag.Parse()

	// Initialize logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	err := func() error {
		if kind == "" {
			return fmt.Errorf("kind is required")
		}
		if dir == "" {
			return fmt.Errorf("dir is required")
		}
		if target == "" {
			return fmt.Errorf("target is required")
		}
		kindInfo, ok := kindMap[kind]
		if !ok {
			return fmt.Errorf("unsupported kind %s", kind)
		}
		log.Info("Synchronization started", "kind", kind, "dir", dir, "target", target)

		dialOptions := []grpc.DialOption{}
		clientConn, err := grpcutil.NewClient(ctx, target, dialOptions...)
		if err != nil {
			return err
		}

		synchronizer, err := kindInfo.NewSynchronizer(os.DirFS(dir), clientConn)
		if err != nil {
			return err
		}
		changed, err := synchronizer.Synchronize(ctx)
		if err != nil {
			return err
		}
		log.Info("Synchronization complete", "changed", changed)
		return nil
	}()
	if err != nil {
		log.Error(err, "fatal error")
		os.Exit(1)
	}
}
