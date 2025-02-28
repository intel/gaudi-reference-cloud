// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"context"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func getCommandWithLargeOutput(ctx context.Context, width int) *exec.Cmd {
	return exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("cat /dev/urandom | tr -dc 'A-Za-z0-9' | head -c %d", width))
}

func getCommandWithMultiLineOutput(ctx context.Context, width int, lines int) *exec.Cmd {
	return exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf(
		"i=1; while [ $i -le %d ]; do (cat /dev/urandom | tr -dc 'A-Za-z0-9' | head -c %d; echo); i=$((i + 1)); done",
		lines, width))
}

var _ = Describe("RunCmd Tests", func() {
	It("RunCmd with echo should succeed", func() {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, "echo", "hello", "world")
		Expect(RunCmd(ctx, cmd)).Should(Succeed())
	})

	It("RunCmd with output of size DefaultLogWidth should succeed", func() {
		ctx := context.Background()
		cmd := getCommandWithLargeOutput(ctx, DefaultLogWidth)
		Expect(RunCmd(ctx, cmd)).Should(Succeed())
	})

	It("RunCmd with output greater than DefaultLogWidth should get truncated", func() {
		ctx := context.Background()
		cmd := getCommandWithLargeOutput(ctx, DefaultLogWidth+1)
		Expect(RunCmd(ctx, cmd)).Should(Succeed())
		// TODO: Check that logged output shows truncated output.
	})

	It("RunCmd with output equal to scanner buffer should succeed", func() {
		ctx := context.Background()
		newLineSize := 1
		cmd := getCommandWithLargeOutput(ctx, MaxScanTokenSize-newLineSize)
		Expect(RunCmd(ctx, cmd)).Should(Succeed())
	})

	It("RunCmd with output greater than scanner buffer should succeed", func() {
		ctx := context.Background()
		cmd := getCommandWithLargeOutput(ctx, MaxScanTokenSize+1)
		Expect(RunCmd(ctx, cmd)).Should(Succeed())
		// TODO: Check that logged output shows "bufio.Scanner: token too long".
	})

	It("RunCmd with many lines should succeed", func() {
		ctx := context.Background()
		cmd := getCommandWithMultiLineOutput(ctx, 40, 1000)
		Expect(RunCmd(ctx, cmd)).Should(Succeed())
	})

	It("RunCmd with many lines greater than scanner buffer should succeed", func() {
		ctx := context.Background()
		cmd := getCommandWithMultiLineOutput(ctx, MaxScanTokenSize+1, 2)
		Expect(RunCmd(ctx, cmd)).Should(Succeed())
	})
})
