// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package interceptors

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/contextmeta"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type MetaDataInterceptors struct {
	log logr.Logger
}

func NewMetaDataInterceptors(log logr.Logger) *MetaDataInterceptors {
	return &MetaDataInterceptors{
		log: log,
	}
}

func (m *MetaDataInterceptors) InjectRequestIdStreamServerInterceptor() grpc.StreamServerInterceptor {
	log := m.log.WithName("InjectMetaStreamServerInterceptor")
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		requestId := uuid.New().String()
		logger := log.WithValues("requestId", requestId)

		logger.Info("called")

		newCtx, err := contextmeta.CreateContextWithRequestId(stream.Context(), requestId)
		if err != nil {
			return errors.Wrap(err, "couldn't create context with requestId")
		}

		return handler(srv, &StreamWrapper{stream, newCtx})
	}
}

func (m *MetaDataInterceptors) InjectRequestIdUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	log := m.log.WithName("InjectMetaUnaryServerInterceptor")
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		requestId := uuid.New().String()
		logger := log.WithValues("requestId", requestId)

		logger.Info("called")

		newCtx, err := contextmeta.CreateContextWithRequestId(ctx, requestId)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't create context with requestId")
		}

		return handler(newCtx, req)
	}
}
