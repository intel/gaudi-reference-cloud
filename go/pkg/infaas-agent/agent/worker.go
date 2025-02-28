// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package agent

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"io"
)

type Worker struct {
	agent *Agent
	log   logr.Logger
}

func (w *Worker) Start(ctx context.Context) error {
	w.log.Info("starting to poll dispatcher for work")

	for {
		select {
		case <-ctx.Done():
			w.log.Info("Agent loop exit due to main context cancellation")
			return ctx.Err()
		default:
			workCtx, cancel := context.WithCancel(ctx)
			w.log.Info("registering available model capacity (may block waiting for server)...")
			// TODO add ctx timeout so we don't park on the server, helping faster shutdown. TO has to be longer than infer requests
			dispatcherStream, err := w.agent.dispatcher.DoWork(workCtx, grpc.WaitForReady(true))
			if err != nil {
				w.log.Error(err, "failed to connect the dispatcher stream; backing off for 1 sec")
				cancel()
				continue
			}

			w.log.Info("got work response stream from dispatcher... sending model name...")
			err = dispatcherStream.Send(&pb.DispatcherResponse{Model: w.agent.model})
			if err != nil {
				w.log.Error(err, "failed to identify model to the dispatcher")
				w.closeSendGracefully(dispatcherStream, cancel)
				continue
			}

			w.handleInferenceRequest(workCtx, dispatcherStream)
			w.closeSendGracefully(dispatcherStream, cancel)
		}
	}
}

func (w *Worker) closeSendGracefully(dispatcherStream pb.Dispatcher_DoWorkClient, cancel context.CancelFunc) {
	if err := dispatcherStream.CloseSend(); err != nil {
		w.log.Error(err, "failed to close stream")
	}
	cancel()
}

func (w *Worker) handleInferenceRequest(ctx context.Context, dispatcherStream pb.Dispatcher_DoWorkClient) {
	req, err := dispatcherStream.Recv()
	if err != nil {
		// TODO add err counter
		w.log.V(1).Info("failed to get work from the dispatcher", "err", err)
		return
	}

	log := w.log.WithValues("requestID", req.RequestID)
	log.Info("got work payload; calling inference backend")

	// call inference service and return result
	inferenceReqCtx, cancel := context.WithTimeout(ctx, req.Timeout.AsDuration())
	defer cancel()
	generateRespStream, err := w.agent.inferenceService.GenerateStream(inferenceReqCtx, req.Request)
	if err != nil {
		log.Error(err, "failed to get inference service response")
		return
	}

	log.Info("got inference stream resp... returning to dispatcher in batches... ")
	chunksReceived := 0
	chunksSent := 0
	for {
		resp, err := generateRespStream.Recv()
		if err == io.EOF {
			w.log.V(1).Info("inference service stream ended", "chunksSent", chunksSent, "chunksReceived", chunksReceived)
			break // end of inference service response stream
		}
		if err != nil {
			log.Error(err, "failed to receive response item from inference service", "chunksSent", chunksSent, "chunksReceived", chunksReceived)
			grpcStatus, _ := status.FromError(err)
			if err = dispatcherStream.Send(&pb.DispatcherResponse{Status: grpcStatus.Proto()}); err != nil {
				log.Error(err, "failed to send error response to caller", "chunksSent", chunksSent, "chunksReceived", chunksReceived)
			}
			return
		}
		chunksReceived++

		log.V(1).Info("returning current response item to dispatcher...", "chunksSent", chunksSent, "chunksReceived", chunksReceived)
		err = dispatcherStream.Send(&pb.DispatcherResponse{
			Model:     w.agent.model,
			Response:  resp,
			RequestID: req.RequestID,
		})
		if err != nil {
			log.Error(err, "failed to send response to caller", "chunksSent", chunksSent, "chunksReceived", chunksReceived)
			return
		}
		chunksSent++
	}

	log.Info("cycle complete!", "chunksSent", chunksSent, "chunksReceived", chunksReceived)
}
