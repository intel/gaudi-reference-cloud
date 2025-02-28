// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iprm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/core/cache"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	portv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/internal/controller/helper"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/idletimer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

// Implements a cache.ListerWatcher that reads updates from the GRPC SubnetServiceClient.SearchStreamPrivate and Watch methods.
// See https://github.com/kubernetes/client-go/blob/master/tools/cache/listwatch.go
type ListerWatcher struct {
	grpcClient pb.IPRMPrivateServiceClient
	watcher    cache.Watcher
	// Cancel the List or Watch method if no message is received for this duration.
	timeout time.Duration
	// Called whenever the Watch method is successful.
	// It will be successful whenever it receives any event, including a bookmark event.
	OnWatchSuccess func()
}

func NewListerWatcher(grpcClient pb.IPRMPrivateServiceClient, timeout time.Duration) *ListerWatcher {
	return &ListerWatcher{
		grpcClient:     grpcClient,
		watcher:        cache.Watcher{},
		timeout:        timeout,
		OnWatchSuccess: func() {},
	}
}

// List returns all non-deleted subnets.
func (lw *ListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	ctx := context.Background()
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("IPRMListerWatcher.List").Start()
	defer span.End()
	log.V(9).Info("BEGIN", "options", options)
	defer log.V(9).Info("END")
	var ports []portv1alpha1.Port

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
	stream, err := lw.grpcClient.SearchStreamPrivate(ctx, &pb.PortSearchStreamPrivateRequest{})
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

			port := portv1alpha1.Port{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "private.cloud.intel.com/v1alpha1",
					Kind:       "port",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:              resp.Object.Metadata.ResourceId,
					Namespace:         resp.Object.Metadata.CloudAccountId,
					ResourceVersion:   resp.Object.Metadata.ResourceVersion,
					CreationTimestamp: helper.DerefTime(helper.FromPbTimestamp(resp.Object.Metadata.CreationTimestamp)),
					DeletionTimestamp: helper.FromPbTimestamp(resp.Object.Metadata.DeletionTimestamp),
					// Add labels used by IDC operators.
					Labels: map[string]string{
						"cloud-account-id": resp.Object.Metadata.CloudAccountId,
					},
				},
				Spec:   resp.Object.Spec,
				Status: resp.Object.Status,
			}

			ports = append(ports, port)
		} else if resp.Type == pb.WatchDeltaType_Bookmark {
			resourceVersion = resp.Object.Metadata.ResourceVersion
		}
	}

	portsList := &portv1alpha1.PortList{
		ListMeta: metav1.ListMeta{
			ResourceVersion: resourceVersion,
		},
		Items: ports,
	}
	return portsList, nil
}

func (lw *ListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	ctx := context.Background()
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("IPRMListerWatcher.Watch").Start()
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
	stream, err := lw.grpcClient.Watch(ctx, &pb.PortWatchRequest{
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
					port := &portv1alpha1.Port{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "private.cloud.intel.com/v1alpha1",
							Kind:       "port",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:              resp.Object.Metadata.ResourceId,
							Namespace:         resp.Object.Metadata.CloudAccountId,
							ResourceVersion:   resp.Object.Metadata.ResourceVersion,
							CreationTimestamp: helper.DerefTime(helper.FromPbTimestamp(resp.Object.Metadata.CreationTimestamp)),
							DeletionTimestamp: helper.FromPbTimestamp(resp.Object.Metadata.DeletionTimestamp),
							// Add labels used by IDC operators.
							Labels: map[string]string{
								"cloud-account-id": resp.Object.Metadata.CloudAccountId,
							},
						},
						Spec:   resp.Object.Spec,
						Status: resp.Object.Status,
					}

					watchEventType := watch.Modified
					if resp.Type == pb.WatchDeltaType_Deleted {
						watchEventType = watch.Deleted
					}
					event := watch.Event{
						Type:   watchEventType,
						Object: port,
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

	return &cache.Watcher{
		EventChannel: eventChannel,
		Cancel:       cancel,
	}, nil
}
