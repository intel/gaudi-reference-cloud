// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package agent

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-agent/inference"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

var (
	callOptionMatcher = func(in any) bool { _, ok := in.([]grpc.CallOption); return ok }
)

func TestDoWorkError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dispatcher := &dispatcherMock{}
	w := &Worker{
		agent: &Agent{
			dispatcher:       dispatcher,
			inferenceService: inference.Client{},
			model:            "test-model",
		},
		log: logr.Logger{},
	}

	dispatcher.
		On("DoWork", mock.AnythingOfType("*context.cancelCtx"), mock.MatchedBy(callOptionMatcher)).
		Run(func(args mock.Arguments) {
			cancel() // so we exit after the first call
		}).
		Return(&streamMock{}, fmt.Errorf("mock-failure"))

	err := w.Start(ctx)
	assert.ErrorContains(t, err, "context canceled", "context was canceled by the test in the first call")

	dispatcher.AssertExpectations(t)
}

func TestDoWorkStreamError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dispatcher := &dispatcherMock{}
	w := &Worker{
		agent: &Agent{
			dispatcher:       dispatcher,
			inferenceService: inference.Client{},
			model:            "test-model",
		},
		log: logr.Logger{},
	}

	sendErr := fmt.Errorf("mock send error")
	respStreamMock := &streamMock{sendErr: sendErr}
	dispatcher.
		On("DoWork", mock.AnythingOfType("*context.cancelCtx"), mock.MatchedBy(callOptionMatcher)).
		Run(func(args mock.Arguments) {
			cancel() // so we exit after the first call
		}).
		Return(respStreamMock, nil)

	err := w.Start(ctx)
	assert.ErrorContains(t, err, "context canceled", "context was canceled by the test in the first call")

	assert.True(t, respStreamMock.closeSendCalled, "response stream wasn't closed")
	dispatcher.AssertExpectations(t)
}

type inferenceServiceMock struct {
}

func (c inferenceServiceMock) GenerateStream(_ context.Context, _ *pb.GenerateStreamRequest) (pb.TextGenerator_GenerateStreamClient, error) {
	return nil, nil
}

type dispatcherMock struct {
	mock.Mock
}

func (m *dispatcherMock) GenerateStream(_ context.Context, _ *pb.DispatcherRequest, _ ...grpc.CallOption) (pb.Dispatcher_GenerateStreamClient, error) {
	return nil, nil
}

func (m *dispatcherMock) Generate(_ context.Context, _ *pb.DispatcherRequest, _ ...grpc.CallOption) (*pb.DispatcherResponse, error) {
	return nil, nil
}

func (m *dispatcherMock) DoWork(ctx context.Context, opts ...grpc.CallOption) (pb.Dispatcher_DoWorkClient, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(pb.Dispatcher_DoWorkClient), args.Error(1)
}

type streamMock struct {
	grpc.ClientStream

	sendErr error
	recvErr error

	closeSendCalled bool
}

func (s *streamMock) Recv() (*pb.DispatcherRequest, error) { return &pb.DispatcherRequest{}, s.recvErr }
func (s *streamMock) Send(_ *pb.DispatcherResponse) error  { return s.sendErr }
func (s *streamMock) CloseSend() error {
	s.closeSendCalled = true
	return nil
}
