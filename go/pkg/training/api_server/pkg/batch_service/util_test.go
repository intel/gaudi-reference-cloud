// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestToTimestamp(t *testing.T) {
	now := time.Now()
	timestamp := ToTimestamp(now)
	assert.NotNil(t, timestamp)
	assert.Equal(t, now.Unix(), timestamp.GetSeconds())
}

func TestToTime(t *testing.T) {
	now := timestamppb.Now()
	time := ToTime(now)
	assert.NotNil(t, time)
	assert.Equal(t, now.GetSeconds(), time.Unix())
}

func TestFNV32a(t *testing.T) {
	// hash := FNV32a("test")
	// assert.NotNil(t, hash)

	// Test case 1: Valid input
	hashValue1, err1 := FNV32a("example")
	if err1 != nil {
		fmt.Println("Error:", err1)
	} else {
		fmt.Println("Hash Value 1:", hashValue1)
	}

	// Test case 2: Invalid input causing an error
	hashValue2, err2 := FNV32a("") // empty string causing an error
	if err2 != nil {
		fmt.Println("Error:", err2)
	} else {
		fmt.Println("Hash Value 2:", hashValue2)
	}
}
