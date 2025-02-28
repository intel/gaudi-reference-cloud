// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package replicator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	lbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_replicator/convert"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/idletimer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

// Implements a cache.ListerWatcher that reads updates from the GRPC LoadBalancerServiceClient.SearchStreamPrivate and Watch methods.
// See https://github.com/kubernetes/client-go/blob/master/tools/cache/listwatch.go
type LoadBalancerListerWatcher struct {
	grpcClient pb.LoadBalancerPrivateServiceClient
	converter  *convert.LoadBalancerConverter
	watcher    Watcher
	// Cancel the List or Watch method if no message is received for this duration.
	timeout time.Duration
	// Called whenever the Watch method is successful.
	// It will be successful whenever it receives any event, including a bookmark event.
	OnWatchSuccess func()
}

func NewLoadBalancerListerWatcher(grpcClient pb.LoadBalancerPrivateServiceClient, timeout time.Duration, converter *convert.LoadBalancerConverter) (*LoadBalancerListerWatcher, error) {
	return &LoadBalancerListerWatcher{
		grpcClient:     grpcClient,
		converter:      converter,
		watcher:        Watcher{},
		timeout:        timeout,
		OnWatchSuccess: func() {},
	}, nil
}

// List returns all non-deleted load balancers.
func (lw *LoadBalancerListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	ctx := context.Background()
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerListerWatcher.List").Start()
	defer span.End()
	log.V(9).Info("BEGIN", "options", options)
	defer log.V(9).Info("END")
	var loadbalancers []lbv1alpha1.Loadbalancer

	if lw.grpcClient == nil {
		return nil, fmt.Errorf("grpcClient is nil")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	logAndCancel := func() {
		log.Error(ctx.Err(), "SearchStreamPrivate response stream was idle for too long", logkeys.Timeout, lw.timeout)
		cancel()
	}
	idleTimer := idletimer.New(logAndCancel)
	defer idleTimer.Stop()
	idleTimer.Reset(lw.timeout)
	stream, err := lw.grpcClient.SearchStreamPrivate(ctx, &pb.LoadBalancerSearchStreamPrivateRequest{})
	if err != nil {
		return nil, err
	}
	if stream == nil {
		return nil, fmt.Errorf("stream is nil")
	}

	resourceVersion := ""
	for {
		log.V(9).Info("Calling Recv")
		resp, err := stream.Recv()
		log.V(9).Info("Received message", "resp", resp, "err", err)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		idleTimer.Reset(lw.timeout)
		if resp.Type == pb.WatchDeltaType_Updated {
			// Convert from Protobuf LoadBalancerPrivate to Kubernetes Loadbalancer.
			loadbalancer, err := lw.converter.PbToK8s(resp.Object)
			if err != nil {
				// An error occured converting this message.
				// Log the error and ignore the message.
				log.Error(err, "Ignoring message that cannot be converted")
			} else {
				loadbalancers = append(loadbalancers, *loadbalancer)
			}
		} else if resp.Type == pb.WatchDeltaType_Bookmark {
			resourceVersion = resp.Object.Metadata.ResourceVersion
		}
	}
	loadbalancerList := &lbv1alpha1.LoadbalancerList{
		ListMeta: metav1.ListMeta{
			ResourceVersion: resourceVersion,
		},
		Items: loadbalancers,
	}
	return loadbalancerList, nil
}

func (lw *LoadBalancerListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	ctx := context.Background()
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerListerWatcher.Watch").Start()
	defer span.End()
	log.V(9).Info("BEGIN", "options", options)
	defer log.V(9).Info("END")

	if lw.grpcClient == nil {
		return nil, fmt.Errorf("grpcClient is nil")
	}
	eventChannel := make(chan watch.Event)
	ctx, cancel := context.WithCancel(ctx)
	logAndCancel := func() {
		log.Error(ctx.Err(), "Watch response stream was idle for too long", logkeys.Timeout, lw.timeout)
		cancel()
	}
	idleTimer := idletimer.New(logAndCancel)
	idleTimer.Reset(lw.timeout)
	stream, err := lw.grpcClient.Watch(ctx, &pb.LoadBalancerWatchRequest{
		ResourceVersion: options.ResourceVersion,
	})
	if err != nil {
		cancel()
		return nil, err
	}
	if stream == nil {
		cancel()
		return nil, fmt.Errorf("stream is nil")
	}

	go func() {
		log := log.WithName("goroutine")
		log.V(9).Info("BEGIN")
		defer log.V(9).Info("END")
		for {
			err := func() error {
				log.V(9).Info("Calling Recv")
				resp, err := stream.Recv()
				log.V(9).Info("Received message", "resp", resp, "err", err)
				if err != nil {
					return err
				}
				idleTimer.Reset(lw.timeout)
				lw.OnWatchSuccess()
				if resp.Type == pb.WatchDeltaType_Updated || resp.Type == pb.WatchDeltaType_Deleted {
					// Convert from Protobuf LoadBalancerPrivate to Kubernetes LoadBalancer.
					loadbalancer, err := lw.converter.PbToK8s(resp.Object)
					if err != nil {
						// An error occured converting this message.
						// Log the error and ignore the message.
						// We do not want to terminate the watch.
						log.Error(err, "Ignoring message that cannot be converted")
						return nil
					}
					watchEventType := watch.Modified
					if resp.Type == pb.WatchDeltaType_Deleted {
						watchEventType = watch.Deleted
					}
					event := watch.Event{
						Type:   watchEventType,
						Object: loadbalancer,
					}
					eventChannel <- event
				}
				return nil
			}()
			if err != nil {
				// If an error occurs, below will cause the watch to terminate.
				// The K8s library will recover by calling List and replacing the entire cache with the result.
				log.Error(err, logkeys.Error)
				event := watch.Event{
					Type: watch.Error,
					Object: &metav1.Status{
						Status:  "Failure",
						Message: err.Error(),
					},
				}
				eventChannel <- event
				close(eventChannel)
				break
			}
		}
		idleTimer.Stop()
	}()

	return &Watcher{
		eventChannel: eventChannel,
		cancel:       cancel,
	}, nil
}
