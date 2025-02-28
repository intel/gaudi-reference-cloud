// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/danjacques/gofslock/fslock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

// Pool represents a sequence of directories where each directory can be exclusively locked using flock.
type Pool struct {
	MaxPoolDirs int
	PoolBaseDir string
	Perm        os.FileMode
}

func NewPool(maxPoolDirs int, poolBaseDir string, perm os.FileMode) (*Pool, error) {
	return &Pool{
		MaxPoolDirs: maxPoolDirs,
		PoolBaseDir: poolBaseDir,
		Perm:        perm,
	}, nil
}

func (p Pool) PoolDirFromIndex(poolIndex int) string {
	return filepath.Join(p.PoolBaseDir, fmt.Sprintf("%d", poolIndex))
}

func (p Pool) LockFileName() string {
	return "lock"
}

// Return the first directory that can be locked.
// The directory will be locked until the returned Handle is unlocked or the process terminates.
// This will retry until the context times out.
func (p Pool) GetDirectoryFromPool(ctx context.Context) (string, fslock.Handle, error) {
	log := log.FromContext(ctx).WithName("GetDirectoryFromPool")
	for {
		if err := ctx.Err(); err != nil {
			return "", nil, err
		}
		for poolIndex := 0; poolIndex < p.MaxPoolDirs; poolIndex++ {
			poolDir := p.PoolDirFromIndex(poolIndex)
			err := os.MkdirAll(poolDir, p.Perm)
			if err != nil {
				return "", nil, err
			} else {
				handle, err := fslock.Lock(filepath.Join(poolDir, p.LockFileName()))
				if err == fslock.ErrLockHeld {
					// Skip locked directory
				} else if err != nil {
					return "", nil, err
				} else {
					return poolDir, handle, nil
				}
			}
		}
		log.Info("No available pool directories. Sleeping.", "MaxPoolDirs", p.MaxPoolDirs, "PoolBaseDir", p.PoolBaseDir)
		time.Sleep(500 * time.Millisecond)
	}
}
