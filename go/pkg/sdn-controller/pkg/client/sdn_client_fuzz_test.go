// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
//go:build gofuzz
// +build gofuzz

package client

import (
	"context"
	"fmt"
	"testing"

	fuzz "github.com/google/gofuzz"
)

// run `go test -tags=gofuzz` to test
func TestSDNClient_UpdateVlanFuzz(t *testing.T) {
	f := fuzz.New()
	var switchFQDN string
	var vlanID int64
	var port string
	var description string

	// Run the test 1000 times with different inputs
	for i := 0; i < 1000; i++ {
		c := &SDNClient{
			dynamicClient:       nil,
			watchTimeoutSeconds: 30,
		}
		f.Fuzz(&switchFQDN)
		f.Fuzz(&vlanID)
		f.Fuzz(&port)
		f.Fuzz(&description)
		fmt.Printf("switchFQDN [%v], port [%v], vlanID [%v], description [%v]\n", switchFQDN, port, vlanID, description)
		// fail the test if there is a panic. Since the inputs are random strings and numbers, it's ok if it returns error.
		c.UpdateVlan(context.Background(), switchFQDN, port, vlanID, description)
	}
}
