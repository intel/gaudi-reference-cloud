// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package idletimer

import (
	"time"
)

type IdleTimer struct {
	cancel func()
	timer  *time.Timer
}

// Create an IdleTimer that calls the provided cancel function.
func New(cancel func()) *IdleTimer {
	return &IdleTimer{
		cancel: cancel,
	}
}

// Configure the IdleTimer to call the cancel function after the specified timeout.
func (t *IdleTimer) Reset(timeout time.Duration) {
	t.Stop()
	t.timer = time.AfterFunc(timeout, t.cancel)
}

// Stop the timer.
func (t *IdleTimer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
}
