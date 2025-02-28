// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cache

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"k8s.io/apimachinery/pkg/watch"
)

type Watcher struct {
	EventChannel chan watch.Event
	Cancel       context.CancelFunc
}

func (w *Watcher) Stop() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("Watcher.Stop")
	log.V(9).Info("Stop")
	w.Cancel()
}

func (w *Watcher) ResultChan() <-chan watch.Event {
	return w.EventChannel
}
