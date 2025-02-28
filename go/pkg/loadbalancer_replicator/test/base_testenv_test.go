// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"net"
	"sync"

	lbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/phayes/freeport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var watchFailureInjectionLoadBalancerName string = "WatchFailureInjectionLoadBalancerName"

// A test environment that includes an implementation of the LoadBalancerPrivate GRPC service.
// Responses to SearchStreamPrivate and Watch can be controlled by tests using go channels.
// Failures can be injected.
type BaseTestEnv struct {
	pb.UnimplementedLoadBalancerPrivateServiceServer
	LoadBalancerPrivateServiceClient pb.LoadBalancerPrivateServiceClient
	ListResponseChanChan             chan chan *pb.LoadBalancerWatchResponse
	WatchResponseChanChan            chan chan *pb.LoadBalancerWatchResponse
	GprcServerStoppable              *stoppable.Stoppable
	UpdateStatusRecorder             *sync.Map
	RemoveFinalizerRecorder          *sync.Map
}

func NewBaseTestEnv() *BaseTestEnv {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("NewBaseTestEnv")
	_ = log

	testEnv := &BaseTestEnv{
		ListResponseChanChan:    make(chan chan *pb.LoadBalancerWatchResponse, 100),
		WatchResponseChanChan:   make(chan chan *pb.LoadBalancerWatchResponse, 100),
		UpdateStatusRecorder:    &sync.Map{},
		RemoveFinalizerRecorder: &sync.Map{},
	}

	grpcListenPort := uint16(freeport.GetPort())
	grpcServer := grpc.NewServer()
	Expect(grpcServer).ShouldNot(BeNil())
	pb.RegisterLoadBalancerPrivateServiceServer(grpcServer, testEnv)
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
	testEnv.LoadBalancerPrivateServiceClient = pb.NewLoadBalancerPrivateServiceClient(clientConn)

	return testEnv
}

func (e *BaseTestEnv) Stop() {
	ctx := context.Background()
	By("Stopping")
	e.GprcServerStoppable.Stop(ctx)
	close(e.ListResponseChanChan)
	close(e.WatchResponseChanChan)
}

func (e *BaseTestEnv) NextListResponseChannel() chan *pb.LoadBalancerWatchResponse {
	watchResponseChan := make(chan *pb.LoadBalancerWatchResponse, 100)
	e.ListResponseChanChan <- watchResponseChan
	return watchResponseChan
}

func (e *BaseTestEnv) NextWatchResponseChannel() chan *pb.LoadBalancerWatchResponse {
	watchResponseChan := make(chan *pb.LoadBalancerWatchResponse, 100)
	e.WatchResponseChanChan <- watchResponseChan
	return watchResponseChan
}

func (e *BaseTestEnv) SearchStreamPrivate(req *pb.LoadBalancerSearchStreamPrivateRequest, svc pb.LoadBalancerPrivateService_SearchStreamPrivateServer) error {
	watchResponseChan := <-e.ListResponseChanChan
	for resp := range watchResponseChan {
		if resp.Object != nil && resp.Object.Metadata.Name == watchFailureInjectionLoadBalancerName {
			return fmt.Errorf("injected failure")
		}
		if err := svc.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (e *BaseTestEnv) Watch(req *pb.LoadBalancerWatchRequest, svc pb.LoadBalancerPrivateService_WatchServer) error {
	watchResponseChan := <-e.WatchResponseChanChan
	for resp := range watchResponseChan {
		if resp.Object != nil && resp.Object.Metadata.Name == watchFailureInjectionLoadBalancerName {
			return fmt.Errorf("injected failure")
		}
		if err := svc.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (e *BaseTestEnv) UpdateStatus(ctx context.Context, req *pb.LoadBalancerUpdateStatusRequest) (*emptypb.Empty, error) {
	e.UpdateStatusRecorder.Store(req.Metadata.ResourceId, req)
	return &emptypb.Empty{}, nil
}

func (e *BaseTestEnv) RemoveFinalizer(ctx context.Context, req *pb.LoadBalancerRemoveFinalizerRequest) (*emptypb.Empty, error) {
	e.RemoveFinalizerRecorder.Store(req.Metadata.ResourceId, req)
	return &emptypb.Empty{}, nil
}

func NewWatchResponse(deltaType pb.WatchDeltaType, loadbalancer *lbv1alpha1.Loadbalancer) *pb.LoadBalancerWatchResponse {
	pbLoadBalancer, err := loadbalancerConverter.K8sToPb(loadbalancer)
	Expect(err).Should(Succeed())
	response := &pb.LoadBalancerWatchResponse{
		Type:   deltaType,
		Object: pbLoadBalancer,
	}
	return response
}

func NewWatchResponseBookmark(resourceVersion string) *pb.LoadBalancerWatchResponse {
	response := &pb.LoadBalancerWatchResponse{
		Type: pb.WatchDeltaType_Bookmark,
		Object: &pb.LoadBalancerPrivate{
			Metadata: &pb.LoadBalancerMetadataPrivate{
				ResourceVersion: resourceVersion,
			},
		},
	}
	return response
}

func NewWatchResponseInjectFailure() *pb.LoadBalancerWatchResponse {
	response := &pb.LoadBalancerWatchResponse{
		Type: pb.WatchDeltaType_Updated,
		Object: &pb.LoadBalancerPrivate{
			Metadata: &pb.LoadBalancerMetadataPrivate{
				Name: watchFailureInjectionLoadBalancerName,
			},
		},
	}
	return response
}
