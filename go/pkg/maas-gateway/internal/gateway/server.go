// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package gateway

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/contextmeta"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/metering"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
)

type Server struct {
	pb.UnimplementedMaasGatewayServer
	DispatcherClient     pb.DispatcherClient
	usageRecordGenerator *metering.RecordGenerator
	productCatalogClient pb.ProductCatalogServiceClient
	logger               logr.Logger
	config               *config.Config
	metrics              *metrics.PromMetrics
}

func NewServer(
	dispatcherClient pb.DispatcherClient,
	usageRecordGenerator *metering.RecordGenerator,
	productCatalogClient pb.ProductCatalogServiceClient,
	logger logr.Logger,
	config *config.Config,
	metrics *metrics.PromMetrics,
) (*Server, error) {

	return &Server{
		DispatcherClient:     dispatcherClient,
		usageRecordGenerator: usageRecordGenerator,
		productCatalogClient: productCatalogClient,
		logger:               logger,
		config:               config,
		metrics:              metrics,
	}, nil
}

func (s *Server) GetSupportedModels(ctx context.Context, _ *emptypb.Empty) (retModels *pb.ListSupportedModels, retErr error) {
	defer func() {
		if retErr != nil {
			s.metrics.FailedRequests.WithLabelValues("GetSupportedModels", "", retErr.Error()).Inc()
		}
	}()

	logger := s.logger.WithName("GetSupportedModels")

	logger.Info("started")
	defer logger.Info("completed")

	// extract all maas products
	filter := &pb.ProductFilter{
		FamilyId: &s.config.FamilyId,
	}
	products, err := s.productCatalogClient.AdminRead(ctx, filter)
	if err != nil {
		logger.Error(err, "can't read products")
		return nil, err
	}

	// loop over maas products and extract relevant info
	var models []*pb.Model
	for _, p := range products.GetProducts() {
		models = append(models, &pb.Model{ModelName: p.Metadata["hfModelName"], ProductId: p.Id, ProductName: p.Name})
	}

	return &pb.ListSupportedModels{
		Models: models,
	}, nil
}

func (s *Server) GenerateStream(req *pb.MaasRequest, stream pb.MaasGateway_GenerateStreamServer) (retErr error) {
	outstandingRequestsGauge := s.metrics.OutstandingRequests.WithLabelValues("GenerateStream", req.GetModel())
	outstandingRequestsGauge.Inc()
	defer outstandingRequestsGauge.Dec()

	errorLabel := ""
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		s.metrics.RequestsDurations.WithLabelValues("GenerateStream", req.Model, errorLabel).Observe(v)
	}))
	defer timer.ObserveDuration()

	// compute errorLabel - last defer executes first
	defer func() {
		if retErr != nil {
			errorLabel = retErr.Error()
			s.metrics.FailedRequests.WithLabelValues("GenerateStream", req.GetModel(), errorLabel).Inc()
		}
	}()

	ctx := stream.Context()
	requestId, err := contextmeta.GetRequestIdFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't get requestId from context")
	}
	cloudAccountId := req.CloudAccountId

	logger := s.logger.WithName("GenerateStream").WithValues("requestId", requestId).WithValues("cloudAccountId", cloudAccountId)

	logger.Info("started")
	defer logger.Info("completed")

	dispatcherRequest := &pb.DispatcherRequest{}
	//TODO add tests that prove that it works as expected, including with nil values, nil and empty slices, maps of various types, etc.
	err = copier.CopyWithOption(dispatcherRequest, req, copier.Option{DeepCopy: true})
	if err != nil {
		logger.Error(err, "couldn't copy gateway request into dispatcher request")
		return s.friendlyError(logger, err, requestId)
	}
	dispatcherRequest.RequestID = requestId

	startTime := timestamppb.Now()
	response, err := s.DispatcherClient.GenerateStream(ctx, dispatcherRequest)
	if err != nil {
		logger.Error(err, "couldn't invoke InferStream")
		return s.friendlyError(logger, err, requestId)
	}

	maasResponse := &pb.MaasResponse{}
	for {
		dispatcherResponse, err := response.Recv()
		if err == io.EOF {
			// End of stream
			break
		}
		if err != nil {
			logger.Error(err, "error receiving from InferStream")
			return s.friendlyError(logger, err, requestId)
		}

		err = copier.CopyWithOption(maasResponse, dispatcherResponse, copier.Option{DeepCopy: true})
		if err != nil {
			logger.Error(err, "couldn't copy dispatcher response into gateway response")
			return s.friendlyError(logger, err, requestId)
		}
		maasResponse.Response.RequestID = &requestId

		// Send each received response to the client
		if err := stream.Send(maasResponse); err != nil {
			logger.Error(err, "error sending response to client")
			return s.friendlyError(logger, err, requestId)
		}
	}
	endTime := timestamppb.Now()

	// perform sync call to generate a usage record
	var quantity float64 // quantity of tokens for usage record
	// expect generated tokens to be present as part of last response under details
	if maasResponse.Response != nil && maasResponse.Response.Details != nil {
		quantity = float64(maasResponse.Response.Details.GeneratedTokens)
	}

	if err := s.generateUsageRecord(ctx, requestId, cloudAccountId, req.Model, req.ProductName, quantity, startTime, endTime); err != nil {
		logger.Error(err, "couldn't create usage record", "requestId", requestId,
			"cloudAccountId", cloudAccountId)
	}

	return nil
}

func (s *Server) generateUsageRecord(ctx context.Context, requestId, cloudAccountId, model string, productName string, quantity float64, startTime, endTime *timestamppb.Timestamp) error {
	errorLabel := ""
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		s.metrics.RequestsDurations.WithLabelValues("generateUsageRecord", model, errorLabel).Observe(v)
	}))
	defer timer.ObserveDuration()

	ctxWithTO, cancel := context.WithTimeout(ctx, s.config.UsageServerTimeout)
	defer cancel()

	// Convert quantity to millions to match Usage Record rates
	var quantityInMillions float64
	quantityInMillions = quantity / 1_000_000

	if err := s.usageRecordGenerator.CreateUsageRecord(ctxWithTO, requestId, cloudAccountId, productName, quantityInMillions, startTime, endTime); err != nil {
		errorLabel = err.Error()
		s.metrics.FailedRequests.WithLabelValues("generateUsageRecord", model, errorLabel).Inc()
		return err
	}
	return nil
}

func (s *Server) friendlyError(logger logr.Logger, err error, requestId string) error {
	logger = logger.WithName("friendlyError")
	if e, ok := status.FromError(err); ok {
		v := wrapperspb.String(requestId)
		switch e.Code() {
		case codes.InvalidArgument:
			e = status.New(e.Code(), errInvalidArgument)
			break
		case codes.NotFound:
			e = status.New(e.Code(), errInvalidModel)
			break
		case codes.Unavailable:
			e = status.New(e.Code(), errUnavailable)
			break
		default:
			e = status.New(e.Code(), errGeneral)
		}
		eWithDetails, err := e.WithDetails(v)
		if err != nil {
			logger.Error(err, "couldn't add details to error")
			return e.Err()
		}
		return eWithDetails.Err()
	}
	return err
}
