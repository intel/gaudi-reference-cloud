// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"net"
	"sync"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var watchFailureInjectionInstanceName string = "WatchFailureInjectionInstanceName"

// A test environment that includes an implementation of the InstancePrivate GRPC service.
// Responses to SearchStreamPrivate and Watch can be controlled by tests using go channels.
// Failures can be injected.
type BaseTestEnv struct {
	pb.UnimplementedInstancePrivateServiceServer
	InstancePrivateServiceClient pb.InstancePrivateServiceClient
	ListResponseChanChan         chan chan *pb.InstanceWatchResponse
	WatchResponseChanChan        chan chan *pb.InstanceWatchResponse
	GprcServerStoppable          *stoppable.Stoppable
	UpdateStatusRecorder         *sync.Map
	RemoveFinalizerRecorder      *sync.Map
}

func NewBaseTestEnv() *BaseTestEnv {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("NewBaseTestEnv")
	_ = log

	testEnv := &BaseTestEnv{
		ListResponseChanChan:    make(chan chan *pb.InstanceWatchResponse, 100),
		WatchResponseChanChan:   make(chan chan *pb.InstanceWatchResponse, 100),
		UpdateStatusRecorder:    &sync.Map{},
		RemoveFinalizerRecorder: &sync.Map{},
	}

	grpcServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	grpcListenPort := uint16(grpcServerListener.Addr().(*net.TCPAddr).Port)
	grpcServer := grpc.NewServer()
	Expect(grpcServer).ShouldNot(BeNil())
	pb.RegisterInstancePrivateServiceServer(grpcServer, testEnv)
	Expect(err).Should(Succeed())
	testEnv.GprcServerStoppable = stoppable.New(func(ctx context.Context) error {
		go func() {
			<-ctx.Done()
			grpcServer.Stop()
		}()
		return grpcServer.Serve(grpcServerListener)
	})
	testEnv.GprcServerStoppable.Start(ctx)

	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	testEnv.InstancePrivateServiceClient = pb.NewInstancePrivateServiceClient(clientConn)

	return testEnv
}

func (e *BaseTestEnv) Stop() {
	ctx := context.Background()
	By("Stopping")
	e.GprcServerStoppable.Stop(ctx)
	close(e.ListResponseChanChan)
	close(e.WatchResponseChanChan)
}

func (e *BaseTestEnv) NextListResponseChannel() chan *pb.InstanceWatchResponse {
	watchResponseChan := make(chan *pb.InstanceWatchResponse, 100)
	e.ListResponseChanChan <- watchResponseChan
	return watchResponseChan
}

func (e *BaseTestEnv) NextWatchResponseChannel() chan *pb.InstanceWatchResponse {
	watchResponseChan := make(chan *pb.InstanceWatchResponse, 100)
	e.WatchResponseChanChan <- watchResponseChan
	return watchResponseChan
}

func (e *BaseTestEnv) SearchStreamPrivate(req *pb.InstanceSearchStreamPrivateRequest, svc pb.InstancePrivateService_SearchStreamPrivateServer) error {
	watchResponseChan := <-e.ListResponseChanChan
	for resp := range watchResponseChan {
		if resp.Object != nil && resp.Object.Metadata.Name == watchFailureInjectionInstanceName {
			return fmt.Errorf("injected failure")
		}
		if err := svc.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (e *BaseTestEnv) Watch(req *pb.InstanceWatchRequest, svc pb.InstancePrivateService_WatchServer) error {
	watchResponseChan := <-e.WatchResponseChanChan
	for resp := range watchResponseChan {
		if resp.Object != nil && resp.Object.Metadata.Name == watchFailureInjectionInstanceName {
			return fmt.Errorf("injected failure")
		}
		if err := svc.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (e *BaseTestEnv) UpdateStatus(ctx context.Context, req *pb.InstanceUpdateStatusRequest) (*emptypb.Empty, error) {
	e.UpdateStatusRecorder.Store(req.Metadata.ResourceId, req)
	return &emptypb.Empty{}, nil
}

func (e *BaseTestEnv) RemoveFinalizer(ctx context.Context, req *pb.InstanceRemoveFinalizerRequest) (*emptypb.Empty, error) {
	e.RemoveFinalizerRecorder.Store(req.Metadata.ResourceId, req)
	return &emptypb.Empty{}, nil
}

func NewWatchResponse(deltaType pb.WatchDeltaType, instance *privatecloudv1alpha1.Instance) *pb.InstanceWatchResponse {
	pbInstance, err := instanceConverter.K8sToPb(instance)
	Expect(err).Should(Succeed())
	response := &pb.InstanceWatchResponse{
		Type:   deltaType,
		Object: pbInstance,
	}
	return response
}

func NewWatchResponseBookmark(resourceVersion string) *pb.InstanceWatchResponse {
	response := &pb.InstanceWatchResponse{
		Type: pb.WatchDeltaType_Bookmark,
		Object: &pb.InstancePrivate{
			Metadata: &pb.InstanceMetadataPrivate{
				ResourceVersion: resourceVersion,
			},
		},
	}
	return response
}

func NewWatchResponseInjectFailure() *pb.InstanceWatchResponse {
	response := &pb.InstanceWatchResponse{
		Type: pb.WatchDeltaType_Updated,
		Object: &pb.InstancePrivate{
			Metadata: &pb.InstanceMetadataPrivate{
				Name: watchFailureInjectionInstanceName,
			},
		},
	}
	return response
}
