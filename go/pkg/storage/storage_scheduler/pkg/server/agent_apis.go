package server

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func (agentMgr *StorageSchedulerServiceServer) RegisterAgent(ctx context.Context, in *pb.RegisterAgentRequest) (*pb.FilesystemAgent, error) {
	logger := log.FromContext(ctx).WithName("StorageSchedulerServiceServer.RegisterAgent")
	logger.Info("register agent request on cluster", logkeys.ClusterId, in.ClusterId)

	params := storagecontroller.CreateClientRequest{
		ClusterId: in.ClusterId,
		Name:      in.Name,
		IPAddr:    in.IpAddr,
	}
	regClient, err := agentMgr.StrCntClient.RegisterClient(ctx, &params)
	if err != nil {
		logger.Error(err, "error registering weka client")
		return nil, status.Errorf(codes.Internal, "error registering the client")
	}
	return &pb.FilesystemAgent{
		ClusterId:        regClient.ClusterId,
		ClientId:         regClient.ClientId,
		Name:             regClient.Name,
		CustomStatus:     regClient.CustomStatus,
		PredefinedStatus: regClient.PredefinedStatus,
	}, nil
}

func (agentMgr *StorageSchedulerServiceServer) DeRegisterAgent(ctx context.Context, in *pb.DeRegisterAgentRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("StorageSchedulerServiceServer.DeRegisterAgent")
	logger.Info("de-register agent request on cluster", logkeys.ClusterId, in.ClusterId)

	err := agentMgr.StrCntClient.DeRegisterClient(ctx, in.ClusterId, in.ClientId)
	if err != nil {
		logger.Error(err, "error de-registering weka client")
		return nil, status.Errorf(codes.Internal, "error de-registering the client")
	}

	return &emptypb.Empty{}, nil
}

func (agentMgr *StorageSchedulerServiceServer) GetRegisteredAgent(ctx context.Context, in *pb.GetRegisterAgentRequest) (*pb.FilesystemAgent, error) {
	logger := log.FromContext(ctx).WithName("StorageSchedulerServiceServer.GetRegisteredAgent")
	logger.Info("get agent request on cluster", logkeys.ClusterId, in.ClusterId)

	regClient, err := agentMgr.StrCntClient.GetClient(ctx, in.ClusterId, in.ClientId)
	if err != nil {
		logger.Error(err, "error on get weka client")
		return nil, status.Errorf(codes.Internal, "error on get the client")
	}
	return &pb.FilesystemAgent{
		ClusterId:        regClient.ClusterId,
		ClientId:         regClient.ClientId,
		Name:             regClient.Name,
		CustomStatus:     regClient.CustomStatus,
		PredefinedStatus: regClient.PredefinedStatus,
	}, nil
}

func (agentMgr *StorageSchedulerServiceServer) ListRegisteredAgents(in *pb.ListRegisteredAgentRequest, rs pb.WekaStatefulAgentPrivateService_ListRegisteredAgentsServer) error {
	logger := log.FromContext(rs.Context()).WithName("StorageSchedulerServiceServer.ListRegisteredAgents")
	logger.Info("register agent request on cluster", logkeys.ClusterId, in.ClusterId)
	regClients, err := agentMgr.StrCntClient.ListClients(rs.Context(), in.ClusterId, in.Names)
	if err != nil {
		logger.Error(err, "error on list weka client")
		return status.Errorf(codes.Internal, "error on list the client")
	}

	for _, cl := range regClients {
		curr := pb.FilesystemAgent{
			ClusterId:        cl.ClusterId,
			ClientId:         cl.ClientId,
			Name:             cl.Name,
			CustomStatus:     cl.CustomStatus,
			PredefinedStatus: cl.PredefinedStatus,
		}
		if err := rs.Send(&curr); err != nil {
			logger.Error(err, "error sending agent info")
		}
	}
	return nil
}

func (agentMgr *StorageSchedulerServiceServer) PingWekaStatefulAgentPrivate(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("StorageSchedulerServiceServer.PingWekaStatefulAgentPrivate")

	logger.Info("entering storage weka agent private Ping  ")
	defer logger.Info("returning from storage weka agent private Ping")

	return &emptypb.Empty{}, nil
}
