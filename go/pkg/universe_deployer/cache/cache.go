// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/hasher"
)

// Cache provides a simple general-purpose cache that stores files in a local directory.
// Each cached file has a key which must consist of alpha numeric characters, plus ".", "_", and "-".
// A single Cache object can safely be used by multiple goroutines concurrently.
type Cache struct {
	CacheDir string
}

type FileResult struct {
	Path string
}

// Create a new cache.
func New(ctx context.Context, cacheDir string) (*Cache, error) {
	logger := log.FromContext(ctx).WithName("cache.New")
	if cacheDir == "" {
		return nil, fmt.Errorf("creating cache: cacheDir is required")
	}
	c := &Cache{
		CacheDir: cacheDir,
	}
	if err := os.MkdirAll(c.storageDir(ctx), 0750); err != nil {
		return nil, err
	}
	tempDir := c.tempDir(ctx)
	// Clean up temp directory whenever the cache starts.
	if err := os.RemoveAll(tempDir); err != nil {
		logger.Error(err, "Unable to delete temporary directory")
		// Continue on this error.
	}
	if err := os.MkdirAll(tempDir, 0750); err != nil {
		return nil, err
	}
	return c, nil
}

// Get the path to a temporary file that can be written by the caller and then moved to the cache with MoveFileToCache.
func (c *Cache) GetTempFilePath(ctx context.Context, key string) (string, error) {
	if err := ensureValidKey(ctx, key); err != nil {
		return "", err
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return filepath.Join(c.tempDir(ctx), id.String()+"_"+key), nil

}

// Add a file to the cache.
// This moves the file into the cache directory.
// path must have been previously returned by GetTempFilePath.
func (c *Cache) MoveFileToCache(ctx context.Context, key string, path string) (FileResult, error) {
	if err := ensureValidKey(ctx, key); err != nil {
		return FileResult{}, err
	}
	destPath, err := c.cachedFilePath(ctx, key)
	if err != nil {
		return FileResult{}, err
	}
	if err := os.Rename(path, destPath); err != nil {
		return FileResult{}, err
	}
	fileResult := FileResult{
		Path: destPath,
	}
	return fileResult, nil
}

// Get the path to a file in the cache.
// Returns ErrNotFound if key was not stored in the cache.
func (c *Cache) GetFile(ctx context.Context, key string) (FileResult, error) {
	if err := ensureValidKey(ctx, key); err != nil {
		return FileResult{}, err
	}
	path, err := c.cachedFilePath(ctx, key)
	if err != nil {
		return FileResult{}, err
	}
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return FileResult{}, &ErrNotFound{Key: key}
		} else {
			return FileResult{}, err
		}
	}
	fileResult := FileResult{
		Path: path,
	}
	return fileResult, nil
}

// Returns true if a file with the key was added to the cache.
func (c *Cache) IsCached(ctx context.Context, key string) (bool, error) {
	_, err := c.GetFile(ctx, key)
	if err != nil {
		if !errors.Is(err, &ErrNotFound{}) {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (c *Cache) cachedFilePath(ctx context.Context, key string) (string, error) {
	path := filepath.Join(c.storageDir(ctx), key)
	return path, nil
}

func (c *Cache) tempDir(_ context.Context) string {
	return filepath.Join(c.CacheDir, "tmp")
}

func (c *Cache) storageDir(_ context.Context) string {
	return filepath.Join(c.CacheDir, "storage")
}

// If desired, keys can be hashes.
// This convenience function can be used to generate hashes.
func (c *Cache) NewHasher(ctx context.Context) *hasher.Hasher {
	return hasher.New()
}

type ErrNotFound struct {
	Key string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found in cache", e.Key)
}

func (e *ErrNotFound) Is(target error) bool {
	_, ok := target.(*ErrNotFound)
	return ok
}

var reValidKey *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// Ensure the key consists of alpha numeric characters, plus ".", "_", and "-".
func ensureValidKey(_ context.Context, key string) error {
	if !reValidKey.MatchString(key) {
		return fmt.Errorf("key '%s' is invalid", key)
	}
	return nil
}
