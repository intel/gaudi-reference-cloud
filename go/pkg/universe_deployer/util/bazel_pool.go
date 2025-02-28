// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"os"
	"path/filepath"
)

type BazelPool struct {
	Pool
}

func NewBazelPool(maxPoolDirs int, poolBaseDir string, perm os.FileMode) (*BazelPool, error) {
	return &BazelPool{
		Pool: Pool{
			MaxPoolDirs: maxPoolDirs,
			PoolBaseDir: poolBaseDir,
			Perm:        perm,
		},
	}, nil
}

func (p BazelPool) HomeDir(poolDir string) string {
	return filepath.Join(poolDir, "home")
}

func (p BazelPool) WorkspaceDir(poolDir string) string {
	return filepath.Join(poolDir, "commit")
}

func (p BazelPool) PrepareWorkspace(poolDir string) (workspaceDir string, homeDir string, err error) {
	homeDir = p.HomeDir(poolDir)
	if err = os.MkdirAll(homeDir, p.Perm); err != nil {
		return
	}
	workspaceDir = p.WorkspaceDir(poolDir)
	if err = os.MkdirAll(workspaceDir, p.Perm); err != nil {
		return
	}
	return
}
