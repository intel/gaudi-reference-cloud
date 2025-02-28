// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/phayes/freeport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

var watchFailureInjectionVPCName string = "WatchFailureInjectionVPCName"

// A test environment that includes an implementation of the VPCPrivate GRPC service.
// Responses to SearchStreamPrivate and Watch can be controlled by tests using go channels.
// Failures can be injected.
type BaseTestEnv struct {
	pb.UnimplementedVPCServiceServer
	pb.UnimplementedVPCPrivateServiceServer
	VPCPrivateServiceClient  pb.VPCPrivateServiceClient
	VPCListResponseChanChan  chan chan *pb.VPCWatchResponse
	VPCWatchResponseChanChan chan chan *pb.VPCWatchResponse
	GprcServerStoppable      *stoppable.Stoppable
	UpdateStatusRecorder     *sync.Map
	RemoveFinalizerRecorder  *sync.Map
}

func NewBaseTestEnv() *BaseTestEnv {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("NewBaseTestEnv")
	_ = log

	testEnv := &BaseTestEnv{
		VPCListResponseChanChan:  make(chan chan *pb.VPCWatchResponse, 100),
		VPCWatchResponseChanChan: make(chan chan *pb.VPCWatchResponse, 100),
		UpdateStatusRecorder:     &sync.Map{},
		RemoveFinalizerRecorder:  &sync.Map{},
	}

	grpcListenPort := uint16(freeport.GetPort())
	grpcServer := grpc.NewServer()
	Expect(grpcServer).ShouldNot(BeNil())
	pb.RegisterVPCPrivateServiceServer(grpcServer, testEnv)
	listenAddr := fmt.Sprintf(":%d", grpcListenPort)
	listener, err := net.Listen("tcp", listenAddr)
	Expect(err).Should(Succeed())
	testEnv.GprcServerStoppable = stoppable.New(func(ctx context.Context) error {
		go func() {
			<-ctx.Done()
			grpcServer.Stop()
		}()
		return grpcServer.Serve(listener)
	})
	testEnv.GprcServerStoppable.Start(ctx)

	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	testEnv.VPCPrivateServiceClient = pb.NewVPCPrivateServiceClient(clientConn)

	return testEnv
}

func (e *BaseTestEnv) Stop() {
	ctx := context.Background()
	By("Stopping")
	err := e.GprcServerStoppable.Stop(ctx)
	if err != nil {
		return
	}
	close(e.VPCListResponseChanChan)
	close(e.VPCWatchResponseChanChan)
}

func (e *BaseTestEnv) NextListResponseChannel() chan *pb.VPCWatchResponse {
	watchResponseChan := make(chan *pb.VPCWatchResponse, 100)
	e.VPCListResponseChanChan <- watchResponseChan
	return watchResponseChan
}

func (e *BaseTestEnv) NextWatchResponseChannel() chan *pb.VPCWatchResponse {
	watchResponseChan := make(chan *pb.VPCWatchResponse, 100)
	e.VPCWatchResponseChanChan <- watchResponseChan
	return watchResponseChan
}

func (e *BaseTestEnv) SearchStreamPrivate(req *pb.VPCSearchStreamPrivateRequest, svc pb.VPCPrivateService_SearchStreamPrivateServer) error {
	watchResponseChan := <-e.VPCListResponseChanChan
	for resp := range watchResponseChan {
		if resp.Object != nil && resp.Object.Metadata.Name == watchFailureInjectionVPCName {
			return fmt.Errorf("injected failure")
		}
		if err := svc.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (e *BaseTestEnv) Watch(req *pb.VPCWatchRequest, svc pb.VPCPrivateService_WatchServer) error {
	watchResponseChan := <-e.VPCWatchResponseChanChan
	for resp := range watchResponseChan {
		if resp.Object != nil && resp.Object.Metadata.Name == watchFailureInjectionVPCName {
			return fmt.Errorf("injected failure")
		}
		if err := svc.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (e *BaseTestEnv) UpdateStatus(ctx context.Context, req *pb.VPCUpdateStatusRequest) (*emptypb.Empty, error) {
	e.UpdateStatusRecorder.Store(req.Metadata.ResourceId, req)
	return &emptypb.Empty{}, nil
}

// func (e *BaseTestEnv) RemoveFinalizer(ctx context.Context, req *pb.VPCRemoveFinalizerRequest) (*emptypb.Empty, error) {
// 	e.RemoveFinalizerRecorder.Store(req.Metadata.ResourceId, req)
// 	return &emptypb.Empty{}, nil
// }

func NewVPCWatchResponse(deltaType pb.WatchDeltaType, vpc *pb.VPCPrivate) (*pb.VPCWatchResponse, error) {

	marshaler := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			// When writing JSON, emit fields that have default values, including for enums.
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			// When reading JSON, ignore fields with unknown names.
			DiscardUnknown: true,
		},
	}

	// Convert the VPCPrivate into a VPCPrivateWatchResponse
	spec, err := marshaler.Marshal(vpc.Spec)
	if err != nil {
		return nil, err
	}

	status, err := marshaler.Marshal(vpc.Status)
	if err != nil {
		return nil, err
	}

	vpcPrivateWatchResponse := &pb.VPCPrivateWatchResponse{
		Metadata: vpc.Metadata,
		Spec:     string(spec),
		Status:   string(status),
	}

	response := &pb.VPCWatchResponse{
		Type:   deltaType,
		Object: vpcPrivateWatchResponse,
	}
	return response, nil
}

func NewWatchResponseBookmark(resourceVersion string) *pb.VPCWatchResponse {
	response := &pb.VPCWatchResponse{
		Type: pb.WatchDeltaType_Bookmark,
		Object: &pb.VPCPrivateWatchResponse{
			Metadata: &pb.VPCMetadataPrivate{
				ResourceVersion: resourceVersion,
			},
		},
	}
	return response
}

func NewWatchResponseInjectFailure() *pb.VPCWatchResponse {
	response := &pb.VPCWatchResponse{
		Type: pb.WatchDeltaType_Updated,
		Object: &pb.VPCPrivateWatchResponse{
			Metadata: &pb.VPCMetadataPrivate{
				Name: watchFailureInjectionVPCName,
			},
		},
	}
	return response
}
