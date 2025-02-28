// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"crypto/sha256"
	"fmt"
	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	// TODO metrics restore later with correct versions
	"google.golang.org/grpc/reflection"
	"io"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/friendsofgo/errors"
	"github.com/go-logr/logr"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type Dispatcher struct {
	pb.UnimplementedDispatcherServer

	model2PendingRequests        map[string]chan *pendingRequest
	grpcServer                   *grpc.Server
	randGenerator                *rand.Rand
	log                          logr.Logger
	healthService                *health.Server
	health                       gosundheit.Health
	defaultGenerateStreamTimeout *time.Duration

	metricsSDK *metrics.PromMetrics
}

type pendingRequest struct {
	startTime time.Time
	req       *pb.DispatcherRequest
	respChan  chan *pb.DispatcherResponse
	inferCtx  context.Context
}

// callbackToCalculateChanRequests returns a function that calculates
// the number of pending requests in a channel
func callbackToCalculateChanRequests(ch chan *pendingRequest) func() float64 {
	return func() float64 {
		return float64(len(ch))
	}
}

func New(_ context.Context, dispatcherConf config.DispatcherConfig, metricsSDK *metrics.PromMetrics, log logr.Logger) (*Dispatcher, error) {
	d := Dispatcher{
		log:                          log.WithName(ServiceName),
		model2PendingRequests:        make(map[string]chan *pendingRequest, len(dispatcherConf.SupportedModels)),
		randGenerator:                rand.New(rand.NewSource(time.Now().UnixNano())),
		healthService:                health.NewServer(),
		defaultGenerateStreamTimeout: dispatcherConf.DefaultGenerateStreamTimeout,
		metricsSDK:                   metricsSDK,
	}

	for _, m := range dispatcherConf.SupportedModels {
		d.model2PendingRequests[m] = make(chan *pendingRequest, dispatcherConf.BacklogSize)
		d.metricsSDK.CreateGaugeForChannel(callbackToCalculateChanRequests(d.model2PendingRequests[m]), m)
	}

	return &d, nil
}

func (d *Dispatcher) Run(ctx context.Context, ln net.Listener) error {
	d.log.Info("starting server...")

	serverOptions := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(d.metricsSDK.GrpcMetrics.UnaryServerInterceptor()),
		grpc.StreamInterceptor(d.metricsSDK.GrpcMetrics.StreamServerInterceptor()),
	}

	grpcServer, err := grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return errors.Wrap(err, "failed to create gRPC server")
	}

	// register dispatcher, health endpoints, and reflection API
	d.grpcServer = grpcServer
	d.log.Info("registering dispatch service...")
	pb.RegisterDispatcherServer(d.grpcServer, d)
	d.log.Info("registering health service...")
	healthgrpc.RegisterHealthServer(d.grpcServer, d.healthService)
	d.log.Info("registering reflection service...")
	reflection.Register(d.grpcServer)

	err = d.setupHealthChecks()
	if err != nil {
		return errors.Wrap(err, "failed to setup health checks")
	}

	d.metricsSDK.GrpcMetrics.InitializeMetrics(grpcServer)
	d.log.Info("starting server...")
	return d.grpcServer.Serve(ln)
}

func (d *Dispatcher) Stop(graceful bool) {
	if d.grpcServer == nil {
		return
	}
	if graceful {
		d.grpcServer.GracefulStop()
	} else {
		d.grpcServer.Stop()
	}
}

func (d *Dispatcher) Generate(ctx context.Context, req *pb.DispatcherRequest) (*pb.DispatcherResponse, error) {
	// singleResponseStream will keep the last response object, normally containing all the tokens
	respStream := singleResponseStream{ctx: ctx}
	err := d.GenerateStream(req, respStream)
	return respStream.response, err
}

func (d *Dispatcher) GenerateStream(req *pb.DispatcherRequest, respStream pb.Dispatcher_GenerateStreamServer) (retErr error) {
	errorLabel := ""
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		d.metricsSDK.RequestsDurations.WithLabelValues("GenerateStream", req.Model, errorLabel).Observe(v)
	}))
	defer timer.ObserveDuration()

	defer func() {
		if retErr != nil {
			d.metricsSDK.FailedRequests.WithLabelValues("GenerateStream", req.Model, errorLabel).Inc()
		}
	}()

	// compute errorLabel - last defer executes first
	defer func() {
		if retErr != nil {
			errorLabel = retErr.Error()
		}
	}()

	outstandingReqGauge := d.metricsSDK.OutstandingRequests.WithLabelValues("GenerateStream", req.Model)
	outstandingReqGauge.Inc()
	defer outstandingReqGauge.Dec()

	log := d.log.WithValues("requestID", req.RequestID, "model", req.Model)
	if req.Request == nil {
		err := errors.New("request field must not be empty")
		log.Error(err, "request is invalid", "error", err)
		return status.Error(codes.InvalidArgument, "request field must not be empty")
	}

	log.Info("got new request")

	requestsChan, ok := d.model2PendingRequests[req.Model]
	if !ok {
		return status.Errorf(codes.NotFound, "no such model %s", req.Model)
	}

	if req.RequestID == "" {
		req.RequestID = d.newID()
	}

	ctxWithTO := respStream.Context()
	if _, ok := ctxWithTO.Deadline(); !ok {
		var cancel context.CancelFunc
		ctxWithTO, cancel = context.WithTimeout(ctxWithTO, *d.defaultGenerateStreamTimeout)
		defer cancel()
	}

	pendingReq := &pendingRequest{
		startTime: time.Time{},
		req:       req,
		respChan:  make(chan *pb.DispatcherResponse),
		inferCtx:  ctxWithTO,
	}

	log.Info("enqueueing request; awaiting capacity...", "prompt-size", len(req.Request.Prompt))
	select {
	case <-ctxWithTO.Done():
		// The client request was cancelled
		log.V(1).Info("Timeout waiting to enqueue the request")
		return status.FromContextError(ctxWithTO.Err()).Err()
	case requestsChan <- pendingReq:
		log.Info("request enqueued for dispatch")
		// this blocks/wait for inference service available capacity
	}

	responseChunks := 0
	for {
		log.Info("awaiting response (chunk) from backend inference service...")
		select {
		case <-ctxWithTO.Done():
			// The client request was cancelled
			log.Error(ctxWithTO.Err(), "Timeout waiting for inference response", "responseChunks", responseChunks)
			return status.FromContextError(ctxWithTO.Err()).Err()

		case resp, ok := <-pendingReq.respChan: // this blocks/wait for inference service response
			if !ok {
				log.Info("response channel closed - stream completed", "responseChunks", responseChunks)
				return status.FromContextError(ctxWithTO.Err()).Err()
			}
			if resp.Status != nil { // check if got grpc status err from agent
				err := status.ErrorProto(resp.Status)
				log.Error(err, "inference agent returned an error status", "grpc status", resp.Status)
				return err
			}
			responseChunks++
			log.V(1).Info("returning inference response to the caller", "responseChunks", responseChunks)
			if err := respStream.Send(resp); err != nil {
				return status.Convert(errors.Wrap(err, "failed to send reply chunk to caller")).Err()
			}
		}
	}
}

func (d *Dispatcher) DoWork(stream pb.Dispatcher_DoWorkServer) (retErr error) {
	model := "unknown_model"
	defer func() {
		if retErr != nil {
			d.metricsSDK.FailedRequests.WithLabelValues("DoWork", model, retErr.Error()).Inc()
		}
	}()

	resp, err := stream.Recv()
	//if err == io.EOF { IDK....
	//	stream.
	//}
	if err != nil {
		return status.Convert(errors.Wrap(err, "failed to read work offer")).Err()
	}

	model = resp.GetModel()

	connectedAgentsGauge := d.metricsSDK.ConnectedAgents.WithLabelValues("DoWork", model)
	connectedAgentsGauge.Inc()
	defer connectedAgentsGauge.Dec()

	pendingChan, ok := d.model2PendingRequests[model]
	if !ok {
		return status.Errorf(codes.NotFound, "no such model %s", model)
	}

	log := d.log.WithValues("model", model)
	log.Info("awaiting work requests...")
	var pendingReq *pendingRequest
	select {
	case <-stream.Context().Done():
		return status.FromContextError(stream.Context().Err()).Err()
	case pendingReq = <-pendingChan: // this blocks/wait for inference requests
		deadline, ok := pendingReq.inferCtx.Deadline()
		if !ok {
			deadline = time.Now().Add(*d.defaultGenerateStreamTimeout)
		}
		if pendingReq.inferCtx.Err() != nil { // request timed out or canceled
			return pendingReq.inferCtx.Err()
		}
		pendingReq.req.Timeout = durationpb.New(time.Until(deadline)) // propagate original TO

		log = d.log.WithValues("requestID", pendingReq.req.RequestID)
		log.Info("got work request; returning work to agent...")
		err := stream.Send(pendingReq.req)
		if err != nil {
			return status.Convert(errors.Wrap(err, "failed to send work request to agent")).Err()
		}
	}

	log.Info("await inference response from agent...")
	defer close(pendingReq.respChan)
	for {
		resp, err = stream.Recv()
		if err == io.EOF {
			log.Info("agent service reply stream ended")
			return nil
		}
		if err != nil {
			return status.Convert(errors.Wrap(err, "failed to read agent response")).Err()
		}

		log.V(1).Info("enqueueing inference response for sending back...")
		select {
		case <-pendingReq.inferCtx.Done():
			return status.FromContextError(pendingReq.inferCtx.Err()).Err()
		case pendingReq.respChan <- resp: // return to caller
		}
	}
}

func (d *Dispatcher) newID() string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d", d.randGenerator.Int63())))
	return fmt.Sprintf("%x", sum)
}
