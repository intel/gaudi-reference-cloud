// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

// RunCmd output lines that exceed this limit will be truncated.
const DefaultLogWidth int = 4000

// RunCmd output scanner can read lines up to this amount, even though it will be later truncated.
// Lines that exceed this amount will cause cause the scanner to fail and the remainder of the output will not be logged.
const MaxScanTokenSize int = 1024 * 1024

// Run a command, logging stdout and stderr line by line during execution.
// Any line exceeding [DefaultLogWidth] will be truncated.
func RunCmd(ctx context.Context, cmd *exec.Cmd) error {
	log := log.FromContext(ctx).WithName("util.RunCmd")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	log.Info("Running", "Args", cmd.Args, "Dir", cmd.Dir, "Env", cmd.Env)
	if err := cmd.Start(); err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	streamHelper := func(name string, stream io.ReadCloser) {
		go func() {
			defer stream.Close()
			scanner := bufio.NewScanner(stream)
			scanner.Buffer(nil, MaxScanTokenSize)
			for scanner.Scan() {
				outputString := scanner.Text()
				if len(outputString) <= DefaultLogWidth {
					log.WithName(name).Info(outputString)
				} else {
					log.WithName(name).Info(outputString[0:DefaultLogWidth]+"...(truncated)", "outputLen", len(outputString))
				}
			}
			if err := scanner.Err(); err != nil {
				// This will occur if the line exceeded MaxScanTokenSize.
				log.Error(err, "scanning stream", "name", name)
			}
			// Discard any remaining output.
			// If the scanner fails, this is needed to consume the rest of the output to let the process complete.
			if _, err := io.Copy(io.Discard, stream); err != nil {
				log.Error(err, "reading stream", "name", name)
			}
			wg.Done()
		}()
	}
	streamHelper("stdout", stdout)
	streamHelper("stderr", stderr)
	wg.Wait()
	err = cmd.Wait()
	log.Info("Completed", "Args[0]", cmd.Args[0], "err", err)
	return err
}

// Run a command and return the output.
// The entire output will be returned (never truncated).
// The output will also be logged but any line exceeding [DefaultLogWidth] will be truncated when logged.
func RunCmdOutput(ctx context.Context, cmd *exec.Cmd) (string, error) {
	log := log.FromContext(ctx).WithName("RunCmdOutput")
	log.Info("Running", "Args", cmd.Args, "Dir", cmd.Dir, "Env", cmd.Env)
	output, err := cmd.Output()
	outputString := string(output)
	if len(outputString) <= DefaultLogWidth {
		log.Info("Completed", "Args[0]", cmd.Args[0], "err", err, "output", outputString)
	} else {
		log.Info("Completed", "Args[0]", cmd.Args[0], "err", err, "outputLen", len(outputString), "truncatedOutput", outputString[0:DefaultLogWidth])
	}
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Info("Stderr", "Stderr", string(exitError.Stderr))
		}
		return outputString, err
	}
	return outputString, nil
}

// Run a command without capturing stdout and stderr.
// Output will be sent directly to the caller's stdout and stderr.
func RunCmdWithoutCapture(ctx context.Context, cmd *exec.Cmd) error {
	log := log.FromContext(ctx).WithName("RunCmd")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Info("Running", "Args", cmd.Args, "Dir", cmd.Dir, "Env", cmd.Env)
	err := cmd.Run()
	log.Info("Completed", "Args", cmd.Args, "err", err)
	return err
}
