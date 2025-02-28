// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package contextmeta

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/metadata"
)

const (
	XRequestId = "x-request-id"
)

func CreateContextWithRequestId(ctx context.Context, requestId string) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if md.Get(XRequestId) == nil {
			md.Set(XRequestId, requestId)
		}
	} else {
		return nil, errors.New("couldn't get metadata from context")
	}
	return metadata.NewIncomingContext(ctx, md), nil
}

func GetRequestIdFromContext(ctx context.Context) (string, error) {
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("couldn't get metadata from context")
	}

	xRequestIds := metadata.Get(XRequestId)
	if xRequestIds == nil {
		return "", fmt.Errorf("missing required %s metadata in incoming request", XRequestId)
	}

	requestId := xRequestIds[0]
	return requestId, nil
}
