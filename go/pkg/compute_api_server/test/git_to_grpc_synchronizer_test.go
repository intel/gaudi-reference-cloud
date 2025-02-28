// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/instance_type"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/machine_image"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ = Describe("GitToGrpcSynchronizer MachineImage Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	runSync := func(dir string) {
		By("Connecting to GRPC server")
		clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).Should(Succeed())
		By("Creating synchronizer (" + dir + ")")
		synchronizer, err := machine_image.NewMachineImageSynchronizer(
			os.DirFS("../testdata/"+dir),
			clientConn,
		)
		Expect(err).Should(Succeed())
		By("Synchronize first time should succeed (" + dir + ")")
		changed, err := synchronizer.Synchronize(ctx)
		Expect(err).Should(Succeed())
		Expect(changed).Should(BeTrue())
		By("Synchronize again should have no change (" + dir + ")")
		changed, err = synchronizer.Synchronize(ctx)
		Expect(err).Should(Succeed())
		Expect(changed).Should(BeFalse())
	}

	It("Synchronize", Serial, func() {
		runSync("MachineImage1")
		runSync("MachineImage2-delete")
		runSync("MachineImage3-update")
	})
})

var _ = Describe("GitToGrpcSynchronizer InstanceType Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	runSync := func(dir string) {
		By("Connecting to GRPC server")
		clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).Should(Succeed())
		By("Creating synchronizer (" + dir + ")")
		synchronizer, err := instance_type.NewInstanceTypeSynchronizer(
			os.DirFS("../testdata/"+dir),
			clientConn,
		)
		Expect(err).Should(Succeed())
		By("Synchronize first time should succeed (" + dir + ")")
		changed, err := synchronizer.Synchronize(ctx)
		Expect(err).Should(Succeed())
		Expect(changed).Should(BeTrue())
		By("Synchronize again should have no change (" + dir + ")")
		changed, err = synchronizer.Synchronize(ctx)
		Expect(err).Should(Succeed())
		Expect(changed).Should(BeFalse())
	}

	It("Synchronize", Serial, func() {
		runSync("InstanceType")
	})
})
