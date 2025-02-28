// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package agent

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-agent/inference"

	"github.com/go-logr/logr"

	"github.com/friendsofgo/errors"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

type Agent struct {
	dispatcher       pb.DispatcherClient
	dispatcherConn   *grpc.ClientConn
	inferenceService InferenceInterface

	model    string
	capacity uint32
	log      logr.Logger
}

type AgentConfig struct {
	Model          string
	DispatcherAddr string
	BackendAddr    string
	Capacity       uint32
}

type InferenceInterface interface {
	GenerateStream(ctx context.Context, request *pb.GenerateStreamRequest) (pb.TextGenerator_GenerateStreamClient, error)
}

func New(ctx context.Context, conf AgentConfig, log logr.Logger) (*Agent, error) {
	log = log.WithName("inference-agent").WithValues("model", conf.Model)
	log.Info("connecting to dispatcher...")

	dispatcherAddr := resolveDispatcherAddress(ctx, log, conf.DispatcherAddr)
	dialOptions := []grpc.DialOption{
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 2 * time.Second,
			Backoff: backoff.Config{
				BaseDelay:  1 * time.Second,
				Multiplier: 1.5,
				Jitter:     0.2,
				MaxDelay:   10 * time.Second,
			},
		}),
	}
	conn, err := grpcutil.NewClient(ctx, dispatcherAddr, dialOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to the service dispatcher")
	}

	inferenceClient, err := inference.New(conf.BackendAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create inference client")
	}

	return NewFromConnection(conn, inferenceClient, log, conf.Model, conf.Capacity), nil
}

func NewFromConnection(conn *grpc.ClientConn, inferenceClient inference.Client, log logr.Logger, model string, capacity uint32) *Agent {
	return &Agent{
		dispatcher:       pb.NewDispatcherClient(conn),
		inferenceService: inferenceClient,
		model:            model,
		capacity:         capacity,
		dispatcherConn:   conn,
		log:              log,
	}
}

func (a *Agent) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	a.log.Info("starting workers...", "numWorkders", a.capacity)
	for i := 0; i < int(a.capacity); i++ {
		w := &Worker{
			agent: a,
			log:   a.log.WithValues("workerID", fmt.Sprintf("worker-%d", i)),
		}
		g.Go(func() error {
			return w.Start(ctx)
		})
	}

	return g.Wait()
}

func (a *Agent) Close() error {
	return a.dispatcherConn.Close()
}

func resolveDispatcherAddress(ctx context.Context, log logr.Logger, configuredDispatcherAddr string) string {
	log = log.WithName("resolveDispatcherAddress")
	resolver := grpcutil.DnsResolver{}
	dispatcherAddr, err := resolver.Resolve(ctx, configuredDispatcherAddr)
	if err != nil {
		log.Error(err, "failed to resolve dispatcher address; will try the raw address instead")
		dispatcherAddr = configuredDispatcherAddr
	}
	log.Info("resolved dispatcher address", "address", dispatcherAddr)
	return dispatcherAddr
}
