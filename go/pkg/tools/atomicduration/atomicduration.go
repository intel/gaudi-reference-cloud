// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package atomicduration

import (
	"sync"
	"time"
)

// A thread-safe type to calculate the duration since the last action (reset).
type AtomicDuration struct {
	lastResetTime time.Time
	mu            sync.RWMutex
}

func New() *AtomicDuration {
	return &AtomicDuration{}
}

// Store the current time.
func (t *AtomicDuration) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastResetTime = time.Now()
}

// Return the duration since Reset() was last called.
func (t *AtomicDuration) SinceReset() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return time.Since(t.lastResetTime)
}
