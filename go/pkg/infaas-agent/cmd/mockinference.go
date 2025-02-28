// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	_ "embed"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"math/rand"
	"net"
	"strings"
)

const triggerErrPrompt = "===generate mock error===" // const prompt for trigger inference error for test purposes

var (
	//go:embed random-words
	wordFile string
)

func NewMockInferenceCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "mockinference",
		Short: "Starts the mock inference server",
		Long:  `Starts the mock inference server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetDefaultLogger()
			logger := log.FromContext(cmd.Context())

			textGenerator := mockTextGenerator{
				words: strings.Split(wordFile, ", "),
				log:   logger.WithName("mock-inference"),
			}

			lis, err := net.Listen("tcp", ":50051")
			if err != nil {
				return errors.Wrap(err, "failed to listen")
			}

			grpcServer := grpc.NewServer()
			logger.Info("registering health service...")
			healthgrpc.RegisterHealthServer(grpcServer, health.NewServer())
			logger.Info("registering text generator service...")
			pb.RegisterTextGeneratorServer(grpcServer, textGenerator)
			logger.Info("registering reflection service...")
			reflection.Register(grpcServer)

			go func() {
				<-cmd.Context().Done()
				logger.Info("shutting down...")
				grpcServer.Stop()
				err := lis.Close()
				if err != nil {
					logger.Error(err, "error while closing listener")
				}
			}()

			logger.Info("starting mock server...")
			return grpcServer.Serve(lis)
		},
	}

	return cmd
}

type mockTextGenerator struct {
	pb.UnimplementedTextGeneratorServer

	words []string
	log   logr.Logger
}

func (m mockTextGenerator) GenerateStream(req *pb.GenerateStreamRequest, respStream pb.TextGenerator_GenerateStreamServer) error {
	m.log.Info("generating text", "request", req)
	if req.Prompt == triggerErrPrompt { // trigger inference err
		m.log.Info("got trigger error prompt, raising inference error")
		return status.Error(codes.Internal, "inference error")
	}

	maxTokens := 100
	if req.Params != nil {
		if req.Params.MaxNewTokens != 0 {
			maxTokens = int(req.Params.MaxNewTokens)
		}
	}
	collectedText := strings.Builder{}
	for i := 0; i < maxTokens; i++ {
		generatedText := m.words[rand.Intn(len(m.words))]
		collectedText.WriteString(generatedText)
		collectedText.WriteString(" ")

		err := respStream.Send(&pb.GenerateStreamResponse{
			Token: &pb.GenerateAPIToken{
				Id:   uint32(i),
				Text: generatedText,
			},
		})
		if err != nil {
			m.log.Error(err, "failed to send response")
		}
	}

	collectedTextStr := collectedText.String()
	return respStream.Send(&pb.GenerateStreamResponse{
		Token: &pb.GenerateAPIToken{
			Text:    "meh",
			Special: false,
		},
		GeneratedText: &collectedTextStr,
		Details: &pb.StreamDetails{
			FinishReason:    pb.FinishReason_FINISH_REASON_EOS_TOKEN,
			GeneratedTokens: uint32(maxTokens),
		},
	})
}
