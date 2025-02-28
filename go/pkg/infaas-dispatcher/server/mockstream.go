// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/metadata"
)

type singleResponseStream struct {
	response *pb.DispatcherResponse
	ctx      context.Context
}

func (s singleResponseStream) Send(resp *pb.DispatcherResponse) error {
	if s.response != nil {
		return fmt.Errorf("send() called more than once on a singleResponseStream")
	}
	s.response = resp
	return nil
}

func (s singleResponseStream) SendMsg(m any) error {
	return s.Send(m.(*pb.DispatcherResponse))
}

func (s singleResponseStream) Context() context.Context {
	return s.ctx
}
func (s singleResponseStream) SetHeader(metadata.MD) error {
	return fmt.Errorf("SetHeader not implemented...")
}

func (s singleResponseStream) SendHeader(metadata.MD) error {
	return fmt.Errorf("SendHeader not implemented...")
}

func (s singleResponseStream) SetTrailer(metadata.MD) {
}

func (s singleResponseStream) RecvMsg(m any) error {
	return fmt.Errorf("RecvMsg not implemented...")
}
