// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package inference

import (
	"context"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	inferenceClient pb.TextGeneratorClient
}

func New(backendAddr string) (Client, error) {
	conn, err := grpc.Dial(backendAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return Client{}, errors.Wrap(err, "failed to connect to the inference service")
	}

	return NewFromConnection(conn), nil
}

func NewFromConnection(conn grpc.ClientConnInterface) Client {
	return Client{
		inferenceClient: pb.NewTextGeneratorClient(conn),
	}
}

func (c Client) GenerateStream(ctx context.Context, request *pb.GenerateStreamRequest) (pb.TextGenerator_GenerateStreamClient, error) {
	respStream, err := c.inferenceClient.GenerateStream(ctx, request, grpc.WaitForReady(true))
	if err != nil {
		return nil, errors.Wrap(err, "failed to call inference service")
	}

	return respStream, nil
}
