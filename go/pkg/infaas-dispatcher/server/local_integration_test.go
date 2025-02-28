// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-agent/agent"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-agent/inference"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"io"
	"net"
	"testing"
	"time"
)

const (
	supportedModel = "model1"
	numResponses   = 5
)

type closer func()

func TestGenerateStream(t *testing.T) {
	t.Setenv("IDC_SERVER_TLS_ENABLED", "false")
	t.Setenv("IDC_CLIENT_TLS_ENABLED", "false")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("TestGenerateStream")

	mockInference, stopMockFunc := startMockInference(ctx, t, log)

	dispatcherClient, dispatcherConn, dispatcherCloser := startDispatcher(ctx, t, log)
	agentCtx, agentCancel := context.WithCancel(ctx)
	go startAgent(agentCtx, t, mockInference, dispatcherConn, log)

	t.Cleanup(func() {
		log.Info("test cleanup started...")
		agentCancel()
		dispatcherCloser()
		stopMockFunc()
		cancel()
	})

	expiredCtx, toCancel := context.WithCancel(context.Background())
	toCancel()

	testCases := []struct {
		name              string
		req               *pb.DispatcherRequest
		ctx               context.Context
		expectedTokenFunc func(i int) string
		expectedStreamErr string
		expectedReqErr    string
	}{
		{
			name: "happy path",
			ctx:  ctx,
			req: &pb.DispatcherRequest{
				Model:     supportedModel,
				RequestID: "test_request",
				Request: &pb.GenerateStreamRequest{
					Prompt: "test prompt",
					Params: nil,
				},
			},
			expectedStreamErr: "",
			expectedTokenFunc: func(i int) string { return fmt.Sprintf("mock text %d", i) },
		}, {
			name: "nil req id",
			ctx:  ctx,
			req: &pb.DispatcherRequest{
				Model: supportedModel,
				Request: &pb.GenerateStreamRequest{
					Prompt: "test prompt",
					Params: nil,
				},
			},
			expectedStreamErr: "",
			expectedTokenFunc: func(i int) string { return fmt.Sprintf("mock text %d", i) },
		}, {
			name: "nil request validation error",
			ctx:  ctx,
			req: &pb.DispatcherRequest{
				Model:     supportedModel,
				RequestID: "test_request",
				Request:   nil, // should trigger validation failure
			},
			expectedStreamErr: "request field must not be empty",
			expectedTokenFunc: nil,
		}, {
			name: "timeout handling",
			ctx:  expiredCtx,
			req: &pb.DispatcherRequest{
				Model:     supportedModel,
				RequestID: "test_request",
				Request: &pb.GenerateStreamRequest{
					Prompt: "test prompt",
					Params: nil,
				},
			},
			expectedReqErr:    "context canceled",
			expectedTokenFunc: nil,
		}, {
			name: "inference err",
			ctx:  ctx,
			req: &pb.DispatcherRequest{
				Model:     supportedModel,
				RequestID: "test_request",
				Request: &pb.GenerateStreamRequest{
					Prompt: "test prompt",
					Params: &pb.GenerateRequestParameters{
						MaxNewTokens: 0, // will trigger inference error
					},
				},
			},
			expectedStreamErr: "inference error",
			expectedTokenFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generatedRespStream, err := dispatcherClient.GenerateStream(tc.ctx, tc.req)
			if tc.expectedReqErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.expectedReqErr)
				return
			}

			for i := 0; i < numResponses; i++ {
				generatedResp, err := generatedRespStream.Recv()
				if err == io.EOF {
					break
				}
				if tc.expectedStreamErr == "" {
					assert.NoError(t, err)
					log.V(-2).Info("got response", "response", generatedResp)
					assert.Equal(t, tc.expectedTokenFunc(i), generatedResp.Response.Token.Text)
				} else {
					assert.ErrorContains(t, err, tc.expectedStreamErr)
				}
			}
		})
	}
}

func startAgent(ctx context.Context, t *testing.T, inferenceClient inference.Client, dispatcherConn *grpc.ClientConn, log logr.Logger) {
	log = log.WithName("inference-agent").WithValues("model", supportedModel)
	err := agent.NewFromConnection(dispatcherConn, inferenceClient, log, supportedModel, 2).Start(ctx)
	if !errors.Is(err, context.Canceled) {
		require.NoError(t, err, "agent start error")
	}
}

func startDispatcher(ctx context.Context, t *testing.T, log logr.Logger) (pb.DispatcherClient, *grpc.ClientConn, closer) {
	defaultTO := 10 * time.Second
	conf := config.DispatcherConfig{
		ListenPort:                   0,
		SupportedModels:              []string{supportedModel},
		BacklogSize:                  10,
		DefaultGenerateStreamTimeout: &defaultTO,
	}

	metricsSDK := metrics.NewPromMetrics(log, conf, "test-service")

	dispatcher, err := New(ctx, conf, metricsSDK, log)
	require.NoError(t, err)

	ln := bufconn.Listen(1024)
	go func() {
		err := dispatcher.Run(ctx, ln)
		log.Info("dispatcher closed", "err", err)
	}()

	clientConn := createClientConnection(ctx, t, ln)
	closer := func() {
		log.Info("dispatcher closer called!")
		err := ln.Close()
		if err != nil {
			fmt.Printf("error closing listener: %v\n", err)
		}
		dispatcher.Stop(false)
		log.Info("dispatcher closer done!")
	}

	return pb.NewDispatcherClient(clientConn), clientConn, closer
}

func startMockInference(ctx context.Context, t *testing.T, log logr.Logger) (inference.Client, closer) {
	grpcServer, err := grpcutil.NewServer(ctx)
	require.NoError(t, err)

	textGenerator := mockTextGenerator{}
	pb.RegisterTextGeneratorServer(grpcServer, textGenerator)

	ln := bufconn.Listen(1024)
	go func() {
		if err := grpcServer.Serve(ln); err != nil {
			fmt.Printf("Failed to start mock server: %s\n", err)
		}
	}()

	clientConn := createClientConnection(ctx, t, ln)

	closer := func() {
		log.Info("mock textGenerator closer called!")
		err := ln.Close()
		if err != nil {
			fmt.Printf("error closing listener: %v\n", err)
		}
		grpcServer.Stop()
		log.Info("mock textGenerator closer done!")
	}

	return inference.NewFromConnection(clientConn), closer
}

type mockTextGenerator struct {
	pb.UnimplementedTextGeneratorServer
}

func (m mockTextGenerator) GenerateStream(req *pb.GenerateStreamRequest, respStream pb.TextGenerator_GenerateStreamServer) error {
	if req.Params != nil && req.Params.MaxNewTokens == 0 { // trigger inference error in such case
		return errors.New("inference error")
	}
	for i := 0; i < numResponses; i++ {
		generatedText := fmt.Sprintf("mock text %d", i)
		respStream.Send(&pb.GenerateStreamResponse{
			Token: &pb.GenerateAPIToken{
				Id:   666,
				Text: generatedText,
			},
		})
	}

	return nil
}

func createClientConnection(ctx context.Context, t *testing.T, ln *bufconn.Listener) *grpc.ClientConn {
	clientConn, err := grpc.DialContext(ctx, "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return ln.DialContext(ctx)
	}))
	require.NoError(t, err)

	return clientConn
}
