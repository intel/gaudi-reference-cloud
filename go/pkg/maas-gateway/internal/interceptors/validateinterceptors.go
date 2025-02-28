// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package interceptors

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/contextmeta"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Validator interface {
	ValidateAll() error
}

type ValidateInterceptors struct {
	log                  logr.Logger
	productCatalogClient pb.ProductCatalogServiceClient
}

func NewValidateInterceptors(log logr.Logger, productCatalogClient pb.ProductCatalogServiceClient) *ValidateInterceptors {
	return &ValidateInterceptors{
		log:                  log,
		productCatalogClient: productCatalogClient,
	}
}

func (v *ValidateInterceptors) validateRequestedProduct(ctx context.Context, logger logr.Logger, maasRequest *pb.MaasRequest) error {
	logger = logger.WithName("validateRequestedProduct")
	logger.Info("called")

	filter := &pb.ProductUserFilter{
		CloudaccountId: maasRequest.CloudAccountId,
		ProductFilter: &pb.ProductFilter{
			Id:   &maasRequest.ProductId,
			Name: &maasRequest.ProductName,
			Metadata: map[string]string{
				"hfModelName": maasRequest.Model,
			},
		},
	}
	productResp, err := v.productCatalogClient.UserRead(ctx, filter)
	if err != nil {
		return err
	}
	if len(productResp.GetProducts()) == 0 {
		err = errors.New("can't find requested product")
		logger.Error(err, "There is no product that matches all the requested parameters")
		return err
	}

	return nil
}

func (v *ValidateInterceptors) ValidateUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	log := v.log.WithName("ValidateUnaryServerInterceptor")
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		requestId, err := contextmeta.GetRequestIdFromContext(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't get requestId from context")
		}

		logger := log.WithValues("requestId", requestId)

		logger.Info("called")

		// Perform validation
		if r, ok := req.(Validator); ok {
			if err := r.ValidateAll(); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "validation error: %s", err.Error())
			}
		}

		if maasRequest, ok := req.(*pb.MaasRequest); ok {
			if err := v.validateRequestedProduct(ctx, log, maasRequest); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "validation error: %s", err.Error())
			}
		}

		// Call the handler
		return handler(ctx, req)
	}
}

func (v *ValidateInterceptors) ValidateStreamServerInterceptor() grpc.StreamServerInterceptor {
	log := v.log.WithName("ValidateStreamServerInterceptor")
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		requestId, err := contextmeta.GetRequestIdFromContext(stream.Context())
		if err != nil {
			return errors.Wrap(err, "couldn't get requestId from context")
		}

		logger := log.WithValues("requestId", requestId)

		logger.Info("called")

		wrapper := &recvWrapper{
			ServerStream:         stream,
			ValidateInterceptors: v,
			log:                  logger,
		}

		return handler(srv, wrapper)
	}
}

type recvWrapper struct {
	grpc.ServerStream
	*ValidateInterceptors
	log logr.Logger
}

func (s *recvWrapper) RecvMsg(m any) error {
	ctx := context.Background()
	logger := s.log.WithName("RecvMsg")
	logger.Info("called")

	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	//TODO write tests to ensure that infaas-request has ValidateAll method and that it's invoked successfully
	if r, ok := m.(Validator); ok {
		if err := r.ValidateAll(); err != nil {
			return status.Errorf(codes.InvalidArgument, "validation error: %s", err.Error())
		}
	}

	if maasRequest, ok := m.(*pb.MaasRequest); ok {
		if err := s.validateRequestedProduct(ctx, logger, maasRequest); err != nil {
			return status.Errorf(codes.InvalidArgument, "validation error: %s", err.Error())
		}
	}

	return nil
}
