// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package hasher

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/google/uuid"
)

// When concatenating multiple chunks of data to hash, each chunk is delimited by these bytes to ensure
// that the boundaries between chunks is unambiguous.
// This is a fixed random UUID that is very unlikely to occur in any file.
var hasherDelimeterUuid uuid.UUID = uuid.MustParse("8ed78ece-235d-48a7-a2e8-40f8906ebec7")
var hasherDelimeter []byte = hasherDelimeterUuid[:]

// Hasher can be used to create hashes of files and in-memory bytes.
type Hasher struct {
	hash hash.Hash
}

func New() *Hasher {
	return &Hasher{
		hash: sha256.New(),
	}
}

// Add the contents of a file to the hash.
// label will be included in the hash.
// label should be used to distinguish this file from any other file
// so that files with the same content but different names do not result in the same hash.
// Generally, label should be the file's relative path.
func (h *Hasher) AddFile(ctx context.Context, label string, path string) error {
	if _, err := h.hash.Write([]byte(label)); err != nil {
		return err
	}
	if _, err := h.hash.Write(hasherDelimeter); err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(h.hash, file); err != nil {
		return err
	}
	if _, err := h.hash.Write(hasherDelimeter); err != nil {
		return err
	}
	return nil
}

// Add bytes to the hash.
// label will be included in the hash.
// label should be used to distinguish this chunk from any other chunk
// so that chunks with the same content but different names do not result in the same hash.
func (h *Hasher) AddBytes(ctx context.Context, label string, b []byte) error {
	if _, err := h.hash.Write([]byte(label)); err != nil {
		return err
	}
	if _, err := h.hash.Write(hasherDelimeter); err != nil {
		return err
	}
	if _, err := h.hash.Write(b); err != nil {
		return err
	}
	if _, err := h.hash.Write(hasherDelimeter); err != nil {
		return err
	}
	return nil
}

func (h *Hasher) AddString(ctx context.Context, label string, s string) error {
	return h.AddBytes(ctx, label, []byte(s))
}

// Return the hash.
func (h *Hasher) Sum(ctx context.Context) string {
	sum := h.hash.Sum(nil)
	return fmt.Sprintf("%x", sum)
}
