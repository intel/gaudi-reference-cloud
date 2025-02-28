package server

import (
	"context"
	"io"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func (fs *FilesystemServiceServer) RegisterAgent(ctx context.Context, in *pb.RegisterAgentRequest) (*pb.FilesystemAgent, error) {
	//initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.RegisterAgent").WithValues(logkeys.ClusterId, in.ClusterId).Start()
	defer span.End()

	defer logger.Info("returning from filesystem agent registration")

	return fs.wekaAgentClient.RegisterAgent(ctx, in)
}

func (fs *FilesystemServiceServer) DeRegisterAgent(ctx context.Context, in *pb.DeRegisterAgentRequest) (*emptypb.Empty, error) {
	//initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.DeRegisterAgent").WithValues(logkeys.ClusterId, in.ClusterId).Start()
	defer span.End()

	defer logger.Info("returning from filesystem agent de-registration")

	return fs.wekaAgentClient.DeRegisterAgent(ctx, in)
}

func (fs *FilesystemServiceServer) GetRegisteredAgent(ctx context.Context, in *pb.GetRegisterAgentRequest) (*pb.FilesystemAgent, error) {
	//initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.GetRegisteredAgent").WithValues(logkeys.ClusterId, in.ClusterId).Start()
	defer span.End()

	defer logger.Info("returning from filesystem agent get ")

	return fs.wekaAgentClient.GetRegisteredAgent(ctx, in)
}

func (fs *FilesystemServiceServer) ListRegisteredAgents(in *pb.ListRegisteredAgentRequest, rs pb.FilesystemStorageClusterPrivateService_ListRegisteredAgentsServer) error {
	//initialize logger and start trace span
	stream, err := fs.wekaAgentClient.ListRegisteredAgents(rs.Context(), in)
	if err != nil {
		return status.Errorf(codes.Internal, "error reading cluster list")
	}
	var streamErr error
	done := make(chan bool)
	go func() {
		for {
			resp, err := stream.Recv()

			if err != nil {

				done <- true //close(done)
				return
			}
			if err := rs.Send(resp); err != nil {

				streamErr = err
				break
			}
		}
	}()

	<-done

	return streamErr
}

func (fs *FilesystemServiceServer) ListClusters(in *pb.ListClusterRequest, rs pb.FilesystemStorageClusterPrivateService_ListClustersServer) error {
	logger := log.FromContext(rs.Context()).WithName("FilesystemServiceServer.ListClusters")

	logger.Info("entering filesystem list clusters")
	defer logger.Info("returning from filesystem list clusters")

	stream, err := fs.schedulerClient.ListClusters(rs.Context(), in)
	if err != nil {
		return status.Errorf(codes.Internal, "error reading cluster list")
	}
	var streamErr error
	done := make(chan bool)
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				done <- true //close(done)
				return
			}
			if err != nil {
				logger.Error(err, "error reading cluster stream")
			}
			if err := rs.Send(resp); err != nil {
				logger.Error(err, "error sending cluster info")
				streamErr = err
				break
			}
		}
	}()

	<-done
	return streamErr
}

func (fs *FilesystemServiceServer) PingFilesystemClusterPrivate(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("FilesystemServiceServer.PingFilesystemClusterPrivate")

	logger.Info("entering filesystem PingFilesystemClusterPrivate")
	defer logger.Info("returning from filesystem PingFilesystemClusterPrivate")

	return &emptypb.Empty{}, nil
}
