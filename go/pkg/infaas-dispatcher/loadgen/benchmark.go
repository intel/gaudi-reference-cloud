// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/AppsFlyer/wrk3"
	"github.com/friendsofgo/errors"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"sync/atomic"
	"time"
)

const (
	//dispatcherAddr = "146.152.224.83:443" // my dev cluster
	dispatcherAddr = "146.152.225.162:443" // public preprod
	//dispatcherAddr = "146.152.224.101:443" // public
	//dispatcherAddr = "100.64.6.112:443"// private
	//dispatcherAddr = "infaas-dispatcher.idcs-system:50053"
)

func main() {
	sendFunc, cancel, err := prepareGenerateStreamFunc()
	if err != nil {
		panic(err)
	}
	defer cancel()

	//sendFunc()
	//callHealth()
	benchmark := loadBenchmark(os.Args[1], sendFunc)
	benchResult := benchmark.Run()

	wrk3.PrintBenchResult(benchmark.Throughput, benchmark.Duration, benchResult)
}

func loadBenchmark(benchmarkFile string, sendFunc wrk3.RequestFunc) *wrk3.Benchmark {
	b := &wrk3.Benchmark{
		Concurrency: 4,
		Throughput:  1,
		Duration:    120 * time.Second,
		SendRequest: sendFunc,
	}

	fmt.Printf("trying to use %q\n", benchmarkFile)
	dataFile, err := os.Open(benchmarkFile)
	if err != nil {
		_, err = fmt.Fprintf(os.Stderr, "failed to open file %q, using defaults %s\n", benchmarkFile, err)
		if err != nil {
			fmt.Printf("Fprintf error: %s\n", err)
		}
		return b
	}
	err = yaml.NewDecoder(dataFile).Decode(b)
	if err != nil {
		_, err = fmt.Fprintf(os.Stderr, "failed to parse file %q, using defaults %s\n", benchmarkFile, err)
		if err != nil {
			fmt.Printf("Fprintf error: %s\n", err)
		}
		return b
	}
	fmt.Printf("loaded: %+v\n", b)

	return b
}

func prepareGenerateStreamFunc() (wrk3.RequestFunc, context.CancelFunc, error) {
	clientCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

	conn, err := grpc.Dial(
		dispatcherAddr,
		//grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	if err != nil {
		return nil, cancel, errors.Wrap(err, "failed to connect to the service dispatcher")
	}

	var reqIDCounter int64 = 0
	client := pb.NewDispatcherClient(conn)

	return func() error {
		req := &pb.DispatcherRequest{
			RequestID: fmt.Sprintf("req-%d", reqIDCounter),
			Model:     "meta-llama/Meta-Llama-3-70B-Instruct",
			Request: &pb.GenerateStreamRequest{
				Prompt: "what is the meaning of life?",
				Params: &pb.GenerateRequestParameters{
					MaxNewTokens: 100,
				}},
		}
		atomic.AddInt64(&reqIDCounter, 1)

		respStream, err := client.GenerateStream(clientCtx, req, grpc.WaitForReady(true))
		if err != nil {
			err := errors.Wrap(err, "failed to call Generate()")
			fmt.Println("QQQ ", err)
			return err
		}
		for {
			_, err := respStream.Recv()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				fmt.Println("QQQ ", err)
				return errors.Wrap(err, "failed to receive response chunk")
			}
		}
	}, cancel, nil
}
